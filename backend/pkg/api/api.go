package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
			folders = append(folders, model.Folder{Name: entry.Name(), Path: path + entry.Name()})
		} else {
			files = append(files, model.File{Name: entry.Name(), Extension: strings.Split(entry.Name(), ".")[1], Path: path + entry.Name()})
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
	content, err := getFolderContent(instance.FSRoot + req.Path)
	if err != nil {
		fmt.Println("[REDUX] request dropped, could not read folder content:", err)
		return
	}
	// files := []model.File{
	// 	model.File{Name: "Shrek", Extension: "mp4", Path: "./Shrek.mp4"},
	// 	model.File{Name: "Finanzen", Extension: "txt", Path: "./Finanzen.txt"},
	// }
	// folders := []model.Folder{
	// 	model.Folder{Name: "Arbeit", Path: "./Arbeit/"},
	// 	model.Folder{Name: "Projekte", Path: "./test/Projekte"},
	// }
	// content := model.FolderContent{
	// 	Files:   files,
	// 	Folders: folders,
	// }
	bin, _ := json.Marshal(content)
	fmt.Println(string(bin))
	fmt.Fprint(w, string(bin))
}

func (a *APIServer) Serve() {
	http.HandleFunc("/getfoldercontent", handleGetFolderContent)
	http.ListenAndServe(":8080", nil)
}
