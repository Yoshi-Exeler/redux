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
}

type FolderContent struct {
	Files   []File
	Folders []Folder
}

type FolderContentGetRequest struct {
	Path string
}
