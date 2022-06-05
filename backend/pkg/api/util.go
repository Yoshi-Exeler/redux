package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"redux/pkg/model"
	"strings"
)

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

func SHA256(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

// toUserpath converts an unsafe path sepecified by a user to a path that is guaranteed to be in his user directory
func toUserpath(uid uint64, requestedPath string) (string, error) {
	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(absPath, "/files/"+fmt.Sprint(uid)) {
		absPath = strings.TrimPrefix(absPath, "/files/"+fmt.Sprint(uid))
	}
	return filepath.Join("/files/"+fmt.Sprint(uid), absPath), nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func send(writer http.ResponseWriter, value any) error {
	bin, err := json.Marshal(value)
	if err != nil {
		return err
	}
	fmt.Fprint(writer, string(bin))
	return nil
}

// readAndUnmarshallTo reads the reader until completion and then json unmarshalls into a variable of type T
func readAndUnmarshallTo[T any](reader io.Reader) (T, error) {
	var req T
	buff, err := ioutil.ReadAll(reader)
	if err != nil {
		return req, fmt.Errorf("[REDUX] type assertion failed")
	}

	err = json.Unmarshal(buff, &req)
	if err != nil {
		return req, fmt.Errorf("[REDUX] type assertion failed")
	}
	return req, nil
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
