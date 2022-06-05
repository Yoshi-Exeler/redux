package model

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
	Username     string `gorm:"not null; unique"`
	PasswordHash string `gorm:"not null"`
	Salt         string `gorm:"not null"`
	Token        string `gorm:"not null"`
	IsAdmin      bool   `gorm:"not null"`
}
type FolderContent struct {
	Files   []File
	Folders []Folder
}

type FileUploadRequest struct {
	Token      string
	Path       string
	Blob       string
	CurrentDir string
}

type FolderContentGetRequest struct {
	Token string
	Path  string
}

type FileContentGetRequest struct {
	Token string
	Path  string
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

type ListUsersRequest struct {
	Token string
}

type ListUsersResponse struct {
	Users []User
}

type AddUserRequest struct {
	Token string
	User  User
}

type AddUserResponse struct {
	Users []User
}

type RemoveUserRequest struct {
	Token string
	UID   uint64
}

type RemoveUserResponse struct {
	Users []User
}
