package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"redux/pkg/model"
)

/* handleGetFolderContent is the JSON-RPC handler for the /getfoldercontent method
*  This method is the network attached version of getFolderContent
 */
func (a *APIServer) handleGetFolderContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	req, err := readAndUnmarshallTo[model.FolderContentGetRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	req.Path, err = toUserpath(user.ID, req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, invalid path accessed:", err)
		return
	}
	fmt.Println("[REDUX] reading folder", req.Path)
	content, err := getFolderContent("." + "/" + req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read folder content:", err)
		return
	}
	bin, _ := json.Marshal(content)
	fmt.Println(string(bin))
	fmt.Fprint(w, string(bin))
}

/* handleGetFileContent is the JSON-RPC handler for the /getfilecontent method
*  This method returns the content of the file specified in the 'path' field of the request
 */
func (a *APIServer) handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	req, err := readAndUnmarshallTo[model.FileContentGetRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	req.Path, err = toUserpath(user.ID, req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, invalid path accessed:", err)
		return
	}
	fmt.Println("[REDUX] reading file", req.Path)
	buff, err := ioutil.ReadFile(req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read file with error:", err)
		return
	}
	enc := base64.StdEncoding.EncodeToString(buff)
	resp := model.FileContentGetResponse{
		Blob: enc,
	}
	bin, _ := json.Marshal(resp)
	fmt.Printf("[REDUX] successfully read file %v with length %v\n", req.Path, len(buff))
	fmt.Fprint(w, string(bin))
}

/* handleFileUpload is the JSON-RPC handler for the /fileupload method
*  This method will dump the base64 encoded octet stream blob into the file
*  specified in the 'path' field and then return the folder content for the
*  folder in the 'currentDir' field as a convenience
 */
func (a *APIServer) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	req, err := readAndUnmarshallTo[model.FileUploadRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	req.Path, err = toUserpath(user.ID, req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, invalid path accessed:", err)
		return
	}
	fmt.Printf("[REDUX] writing file %v with length %v\n", "."+"/"+req.Path, len(req.Blob))
	decoded, err := base64.StdEncoding.DecodeString(req.Blob)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not decode file with error:", err)
		return
	}
	err = ioutil.WriteFile("."+"/"+req.Path, decoded, 0644)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not write file with error:", err)
		return
	}
	content, err := getFolderContent("." + "/" + req.CurrentDir)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read folder content:", err)
		return
	}
	bin, _ := json.Marshal(content)
	fmt.Fprint(w, string(bin))
}

func (a *APIServer) checkToken(w http.ResponseWriter, token string) (*model.User, bool) {
	fmt.Printf("token:%+v", token)
	var targetUser model.User
	err := a.DB.Where("token = ?", token).First(&targetUser).Error
	if err != nil {
		fmt.Println("auth: could not find user")
		w.WriteHeader(401)
		return nil, false
	}
	return &targetUser, true
}

func (a *APIServer) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	req, err := readAndUnmarshallTo[model.AuthenticationRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	// grab the user with the specified username form the database
	var targetUser model.User
	err = a.DB.Where("username = ?", req.Username).First(&targetUser).Error
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Println("[REDUX] request dropped, user not found")
		return
	}
	// check the password
	fmt.Println(req.Password, targetUser.Salt, SHA256(req.Password+targetUser.Salt))
	if targetUser.PasswordHash != SHA256(req.Password+targetUser.Salt) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Println("[REDUX] request dropped, invalid password")
		return
	}
	// prepare our response
	resp := model.AuthenticationResponse{
		Token: targetUser.Token,
	}
	send(w, resp)
}

func (a *APIServer) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// set cors header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// read the request struct from the http request body
	req, err := readAndUnmarshallTo[model.ListUsersRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	// check the provided authentication token
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	// if the user is not an admin, we drop this request
	if !user.IsAdmin {
		w.WriteHeader(401)
		fmt.Println("[REDUX] request dropped, non admin may not list users")
		return
	}
	// grab all users from the database
	var users []model.User
	err = a.DB.Find(&users).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not get users:", err)
		return
	}
	// send the response back to the client
	send(w, model.ListUsersResponse{
		Users: users,
	})
}

func (a *APIServer) handleAddUser(w http.ResponseWriter, r *http.Request) {
	// set cors header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// read the request struct from the http request body
	req, err := readAndUnmarshallTo[model.AddUserRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	// check the provided authentication token
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	// if the user is not an admin, we drop this request
	if !user.IsAdmin {
		w.WriteHeader(401)
		fmt.Println("[REDUX] request dropped, non admin may not add users")
		return
	}
	// create the user
	err = a.DB.Create(req.User).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could add user:", err)
		return
	}
	// grab all users from the database
	var users []model.User
	err = a.DB.Find(&users).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not get users:", err)
		return
	}
	// send the response back to the client
	send(w, model.AddUserResponse{
		Users: users,
	})
}

func (a *APIServer) handleRemoveUser(w http.ResponseWriter, r *http.Request) {
	// set cors header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// read the request struct from the http request body
	req, err := readAndUnmarshallTo[model.RemoveUserRequest](r.Body)
	if err != nil {
		// add proper http response codes here later
		fmt.Println("[REDUX] request dropped, invalid params:", err)
		return
	}
	// check the provided authentication token
	user, ok := a.checkToken(w, req.Token)
	if !ok {
		return
	}
	// if the user is not an admin, we drop this request
	if !user.IsAdmin {
		w.WriteHeader(401)
		fmt.Println("[REDUX] request dropped, non admin may not list users")
		return
	}
	// delete the specified user
	var targetUser model.User
	err = a.DB.Where("id = ?", req.UID).First(&targetUser).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could find target user:", err)
		return
	}
	// delete the user
	err = a.DB.Delete(targetUser).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could delete target user:", err)
		return
	}
	// grab all users from the database
	var users []model.User
	err = a.DB.Find(&users).Error
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not get users:", err)
		return
	}
	// send the response back to the client
	send(w, model.RemoveUserResponse{
		Users: users,
	})
}
