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
	Username     string `gorm:"not null"`
	PasswordHash string `gorm:"not null"`
	Salt         string `gorm:"not null"`
	Token        string `gorm:"not null"`
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
	Error string
}
