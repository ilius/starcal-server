package main

import (
	"fmt"
	"os"
	"scal/rest_server/api_v1"
	"scal/settings"
	"scal/storage"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--show-settings":
			settings.PrintSettings()
			os.Exit(0)
		default:
			fmt.Printf("invalid argument %v\n", os.Args[1])
			os.Exit(1)
		}
	}
	storage.InitDB()
	storage.EnsureIndexes()
	api_v1.SetMongoErrorDispatcher()
	api_v1.StartAPIv1Server()
}
