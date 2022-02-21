package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"redux/pkg/model"
	"strings"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var instance *APIServer

var once sync.Once

func Init(fsroot string, apiPort string) {
	once.Do(func() {
		db, err := gorm.Open(sqlite.Open(fsroot+"/redux_db.sqlite"), &gorm.Config{})
		if err != nil {
			log.Fatalf("[REDUX] failes to open sqlite database")
		}
		instance = &APIServer{FSRoot: fsroot + "/files", APIPort: apiPort, DB: db}
	})
}

func GetInstance() *APIServer {
	return instance
}

type APIServer struct {
	FSRoot  string
	APIPort string
	DB      *gorm.DB
}

func getFolderContent(path string) (*model.FolderContent, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	files := []model.File{}
	folders := []model.Folder{}
	for _, entry := range infos {
		if entry.IsDir() {
			folders = append(folders, model.Folder{Name: entry.Name(), Path: path + "/" + entry.Name() + "/"})
		} else {
			files = append(files, model.File{Name: entry.Name(), Extension: strings.Split(entry.Name(), ".")[1], Path: path + "/" + entry.Name()})
		}
	}
	return &model.FolderContent{Files: files, Folders: folders}, nil
}

func handleGetFolderContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot read body:", err)
		return
	}
	fmt.Println(">>", string(buff))
	var req model.FolderContentGetRequest
	err = json.Unmarshal(buff, &req)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot unmarshall request:", err)
		return
	}
	fmt.Println("[REDUX] reading folder", req.Path)
	content, err := getFolderContent(instance.FSRoot + "/" + req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read folder content:", err)
		return
	}
	bin, _ := json.Marshal(content)
	fmt.Println(string(bin))
	fmt.Fprint(w, string(bin))
}

func handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot read body:", err)
		return
	}
	fmt.Println(">>", string(buff))
	var req model.FileContentGetRequest
	err = json.Unmarshal(buff, &req)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot unmarshall request:", err)
		return
	}
	fmt.Println("[REDUX] reading file", req.Path)
	buff, err = ioutil.ReadFile(req.Path)
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

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot read body:", err)
		return
	}
	var req model.FileUploadRequest
	err = json.Unmarshal(buff, &req)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot unmarshall request:", err)
		return
	}
	fmt.Printf("[REDUX] writing file %v with length %v\n", instance.FSRoot+"/"+req.Path, len(req.Blob))
	decoded, err := base64.StdEncoding.DecodeString(req.Blob)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not decode file with error:", err)
		return
	}
	err = ioutil.WriteFile(instance.FSRoot+"/"+req.Path, decoded, 0644)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not write file with error:", err)
		return
	}
	content, err := getFolderContent(instance.FSRoot + "/" + req.CurrentDir)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read folder content:", err)
		return
	}
	bin, _ := json.Marshal(content)
	fmt.Fprint(w, string(bin))
}

func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot read body:", err)
		return
	}
	fmt.Println(">>", string(buff))
	var req model.AuthenticationRequest
	err = json.Unmarshal(buff, &req)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot unmarshall request:", err)
		return
	}
	token, err := instance.Authenticate(req.Username, req.Password)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot auth", err)
		return
	}
	fmt.Printf("[REDUX] user %v autheticated", req.Username)
	resp := model.AuthenticationResponse{Token: token}
	bin, _ := json.Marshal(resp)
	fmt.Fprint(w, string(bin))
}

func (a *APIServer) Authenticate(username string, password string) (string, error) {
	// hash the password
	hashBytes := sha256.Sum256([]byte(password))
	// hexencode the hash so we can store/compare in the database
	encodedHash := []byte{}
	hex.Encode(encodedHash, hashBytes[:])
	hashStr := string(encodedHash)
	// check if there is a user with this username and password
	var targetUser model.User
	err := a.DB.Where("username = ? AND password_hash = ?", username, hashStr).First(&targetUser).Error
	if err != nil {
		return "", fmt.Errorf("error: authentication failed")
	}
	// generate a token for this user
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("error: cannot generate token")
	}
	// check if the user already has a token
	if len(targetUser.Token) == 0 {
		// update the users stored token
		encodedToken := []byte{}
		hex.Encode(encodedToken, tokenBytes[:])
		targetUser.Token = string(encodedToken)
		err = a.DB.Save(&encodedToken).Error
		if err != nil {
			return "", fmt.Errorf("error: cannot update token")
		}
	}
	// yield the users token
	return targetUser.Token, nil
}

func (a *APIServer) GetUserFromToken(token string) (*model.User, error) {
	if len(token) == 0 {
		return nil, fmt.Errorf("error: token cannot be empty")
	}
	// try to query a user for the specified token
	var targetUser model.User
	err := a.DB.Where("token = ?", token).First(&targetUser).Error
	if err != nil {
		return nil, fmt.Errorf("error: could not find a user for the specified token")
	}
	// yield the user
	return &targetUser, nil
}

func (a *APIServer) Serve() {
	fmt.Println("[Redux] now serving on 8080")
	http.HandleFunc("/getfoldercontent", handleGetFolderContent)
	http.HandleFunc("/getfilecontent", handleGetFileContent)
	http.HandleFunc("/fileupload", handleFileUpload)
	http.ListenAndServe(":8080", nil)
}
