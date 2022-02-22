package api

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"redux/pkg/model"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/sqlite"
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
		instance.initx509()
		fmt.Println("[REDUX][INIT] x509 keypair initialized successfully")
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
	if !a.CheckJWTHeader(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
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

/* handleAuthenticate is the JSON-RPC handler for the /authenticate method
*  This method will fetch the user with the specified username and check
*  its password hash against the specified password hash. If the authentication
*  was successful, a Json-Web-Token (JWT) is set as a cookie for the user, which can
*  now be used to call other methods.
 */
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
	token, expiry, err := instance.authJWT(req.Username, req.Password)
	if err != nil {
		fmt.Println("[REDUX] request dropped, cannot auth", err)
		return
	}
	// set a JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expiry,
	})
	fmt.Printf("[REDUX] user %v autheticated, JWT cookie was set", req.Username)
	resp := model.AuthenticationResponse{Token: token}
	bin, _ := json.Marshal(resp)
	fmt.Fprint(w, string(bin))
}

// CheckJWTHeader returns wether or not the request has a valid JWT cookie
func (a *APIServer) CheckJWTHeader(r *http.Request) bool {
	// read our cookie
	c, err := r.Cookie("token")
	if err != nil {
		return false
	}
	// parse the jwt
	tok, err := jwt.Parse(c.Value, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}
		return a.Keypair.Public(), nil
	})
	// check for parsing errors
	if err != nil {
		return false
	}
	// check for validity
	_, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return false
	}
	return true
}

// authJWT performs the actual authentication and JWT generation
func (a *APIServer) authJWT(username string, password string) (string, time.Time, error) {
	// check if there is a user with this username and password
	var targetUser model.User
	err := a.DB.Where("username = ?", username).First(&targetUser).Error
	if err != nil {
		return "", time.Now(), fmt.Errorf("user not found")
	}
	// hash the password
	hashBytes := sha256.Sum256([]byte(password + targetUser.Salt))
	// hexencode the hash so we can store/compare in the database
	encodedHash := []byte{}
	hex.Encode(encodedHash, hashBytes[:])
	hashStr := string(encodedHash)
	// check password
	if hashStr != targetUser.PasswordHash {
		return "", time.Now(), fmt.Errorf("authentication failed")
	}
	// create credentials for the JWT
	creds := model.Credentials{Username: username, Password: password}
	// create a JWT expiration time
	expirationTime := time.Now().Add(24 * time.Hour)
	// Create JWT claims
	claims := &model.Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(), // maybe we need millis here?
		},
	}
	// create our JWT
	token := jwt.NewWithClaims(&jwt.SigningMethodRSA{Hash: crypto.BLAKE2b_512}, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(a.Keypair)
	if err != nil {
		return "", time.Now(), fmt.Errorf("could not sign JWT")
	}
	return tokenString, expirationTime, nil
}

/* initx509 reads or creates the server's rsa keypair
*  this method should be called once during the initialization of the api server,
*  but before the changeroot environment has been entered, since the keypair is
*  outside the changeroot environment
 */
func (a *APIServer) initx509() {
	var keypair *rsa.PrivateKey
	// check if our keypair exists
	pemBuffer, err := ioutil.ReadFile("./private.pem")
	if err != nil {
		// if it does not exists, generate it
		kp, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatal("could not generate x509 keypair with error", err)
		}
		keypair = kp
	} else {
		decoded, _ := pem.Decode(pemBuffer)
		if err != nil {
			log.Fatal("could not decode PEM with error", err)
		}
		kp, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
		if err != nil {
			log.Fatal("could not parse x509 key with error", err)
		}
		keypair = kp
	}
	a.Keypair = *keypair
}

// Serve will begin the operation of the api server, binding to the specified port
func (a *APIServer) Serve() {
	fmt.Println("[Redux] now serving on 8080")
	http.HandleFunc("/getfoldercontent", a.handleGetFolderContent)
	http.HandleFunc("/getfilecontent", a.handleGetFileContent)
	http.HandleFunc("/fileupload", a.handleFileUpload)
	http.ListenAndServe(":"+a.APIPort, nil)
}
