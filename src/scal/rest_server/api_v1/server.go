package api_v1

import (
	"log"
	"net/http"

	// to load / register calendar types
	_ "scal/init/cal_types_init"
)

var port = "9001"

func StartAPIv1Server() {
	router := GetRouter()
	log.Printf("Starting to serve api v1 on port %v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
