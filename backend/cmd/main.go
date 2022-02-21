package main

import (
	"flag"
	"redux/pkg/api"
)

func main() {
	// declare our cli variables
	fsRoot := flag.String("fs-root", "./reduxfs/", "the path to the root directory of the cloud's filesystem")
	apiPort := flag.String("api-port", "8050", "the port on which the json-rpc api will listen")
	userlandID := flag.Int("uid", -1, "the uid to revert to after the changeroot environment has been entered")

	flag.Parse()

	// initialize our api server with the config we got from CLI
	api.Init(*fsRoot, *apiPort, *userlandID)
	api.GetInstance().Serve()

	select {}
}
