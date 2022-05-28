package api

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"redux/pkg/model"
	"strings"
	"sync"
	"syscall"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var instance *APIServer

var once sync.Once

func GetInstance() *APIServer {
	return instance
}

type APIServer struct {
	APIPort string
	DB      *gorm.DB
	Keypair rsa.PrivateKey
}

/* Init will initialize the api package.
*  This method should be called IMMEDIATELY after parsing commandline arguments,
*  as it takes care of securing the process from unwanted interference using a
*  changeroot environment and privildege dropping.
 */
func Init(fsroot string, apiPort string, userlandUID int) {
	once.Do(func() {
		db, err := gorm.Open(sqlite.Open(fsroot+"/redux_db.sqlite"), &gorm.Config{})
		if err != nil {
			log.Fatal("failed to open sqlite database", err)
		}
		fmt.Println("[REDUX][INIT] sqlite database handle opened")
		instance = &APIServer{APIPort: apiPort, DB: db}
		if err := os.Chdir(fsroot + "/files"); err != nil {
			log.Fatal("Failed to change to new root", err)
		}
		fmt.Println("[REDUX][INIT] changed active directory into cloud root directory")
		if err := syscall.Chroot(fsroot + "/files"); err != nil {
			log.Fatal("Failed to chroot", err)
		}
		fmt.Printf("[REDUX][INIT] changeroot into %v\n", fsroot+"/files")
		if err := syscall.Setresuid(userlandUID, userlandUID, userlandUID); err != nil {
			log.Fatal("Failed to call setresuid", err)
		}
		fmt.Println("[REDUX][INIT] successfully dropped root permissions with setresuid, new uid:", os.Geteuid())
	})
}

// getFolderContent returns the Files and Folders int the folder in the specified path or an error
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

/* handleGetFolderContent is the JSON-RPC handler for the /getfoldercontent method
*  This method is the network attached version of getFolderContent
 */
func (a *APIServer) handleGetFolderContent(w http.ResponseWriter, r *http.Request) {
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

/* handleFileUpload is the JSON-RPC handler for the /fileupload method
*  This method will dump the base64 encoded octet stream blob into the file
*  specified in the 'path' field and then return the folder content for the
*  folder in the 'currentDir' field as a convenience
 */
func (a *APIServer) handleFileUpload(w http.ResponseWriter, r *http.Request) {
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

func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// Serve will begin the operation of the api server, binding to the specified port
func (a *APIServer) Serve() {
	fmt.Println("[Redux] now serving on", a.APIPort)
	http.HandleFunc("/getfoldercontent", a.handleGetFolderContent)
	http.HandleFunc("/getfilecontent", a.handleGetFileContent)
	http.HandleFunc("/fileupload", a.handleFileUpload)
	http.ListenAndServe(":"+a.APIPort, nil)
}
