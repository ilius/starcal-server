package api_v1

import (
	"net/http"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "util",
		Map: RouteMap{
			"GetApiVersion": {
				Method:      "GET",
				Pattern:     "api-version",
				HandlerFunc: GetApiVersion,
			},
		},
	})
}

func GetApiVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte("1"))
}
