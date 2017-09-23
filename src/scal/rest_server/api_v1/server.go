package api_v1

import (
	"log"
	"net/http"

	_ "scal/init/cal_types_init"
)

func StartAPIv1Server() {
	router := GetRouter()
	log.Fatal(http.ListenAndServe(":9001", router))
}
