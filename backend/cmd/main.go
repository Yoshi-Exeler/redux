package main

import "redux/pkg/api"

func main() {
	api.Init("./cloudstorage/")
	api.GetInstance().Serve()
}
