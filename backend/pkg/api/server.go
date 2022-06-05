package api

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"redux/pkg/model"
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
		db.AutoMigrate(&model.User{})
		db.Save(&model.User{
			Username:     "yoshi.exeler",
			PasswordHash: "Qh1DSyIpvzQvHInbwbkYnXCszKBq64yb7OhHO/vi9SQ=", // testcool_salt
			Token:        "cool_auth_token",
			Salt:         "cool_salt",
			IsAdmin:      true,
		})
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

// Serve will begin the operation of the api server, binding to the specified port
func (a *APIServer) Serve() {
	fmt.Println("[Redux] now serving on", a.APIPort)
	http.HandleFunc("/getfoldercontent", a.handleGetFolderContent)
	http.HandleFunc("/getfilecontent", a.handleGetFileContent)
	http.HandleFunc("/fileupload", a.handleFileUpload)
	http.HandleFunc("/authenticate", a.handleAuthenticate)
	http.HandleFunc("/listusers", a.handleListUsers)
	http.HandleFunc("/adduser", a.handleAddUser)
	http.HandleFunc("/removeuser", a.handleRemoveUser)
	http.ListenAndServe(":"+a.APIPort, nil)
}
