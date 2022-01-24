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
	Username     string
	PasswordHash string
}

type FolderContent struct {
	Files   []File
	Folders []Folder
}

type FolderContentGetRequest struct {
	Path string
}
