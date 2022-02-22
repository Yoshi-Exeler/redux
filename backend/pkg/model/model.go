package model

import "github.com/dgrijalva/jwt-go"

type File struct {
	Name      string
	Extension string
	Path      string
}

type Folder struct {
	Name string
	Path string
}

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"not null"`
	PasswordHash string `gorm:"not null"`
	Salt         string `gorm:"not null"`
	Token        string `gorm:"not null"`
}

// JWT Credentials
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type FolderContent struct {
	Files   []File
	Folders []Folder
}

type FileUploadRequest struct {
	Path       string
	Blob       string
	CurrentDir string
}

type FolderContentGetRequest struct {
	Path string
}

type FileContentGetRequest struct {
	Path string
}

type FileContentGetResponse struct {
	Blob string
}

type AuthenticationRequest struct {
	Username string
	Password string
}

type AuthenticationResponse struct {
	Token string
	Error string
}
