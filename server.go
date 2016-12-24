package main

import "fmt"
import "os"
import "scal/rest_server/api_v1"
import "scal/settings"

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
	api_v1.StartAPIv1Server()
}
