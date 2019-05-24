package api_v1

import (
	"net/http"

	// to load / register calendar types
	_ "scal/init/cal_types_init"
)

var port = "9001"

func StartAPIv1Server() {
	go ErrorSaverLoop()
	router := GetRouter()
	log.Info("Starting to serve api v1 on port %v\n", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
