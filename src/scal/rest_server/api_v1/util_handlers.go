package api_v1

import (
	//"encoding/json"
	//"gopkg.in/mgo.v2/bson"
	"net/http"
)

func init() {
	RegisterRoute(
		"GetApiVersion",
		"GET",
		"/util/api-version/",
		GetApiVersion,
	)
}

func GetApiVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte("1"))
}
