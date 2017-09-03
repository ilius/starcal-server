package api_v1

import (
	"myrpc"
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc // TO REMOVE LATER
	Handler     myrpc.Handler
}
type RouteMap map[string]Route

type RouteGroup struct {
	//NeedsAuth bool
	Base string
	Map  RouteMap
}

var routeGroups = []RouteGroup{}

func GetRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	for _, routeGroup := range routeGroups {
		for name, route := range routeGroup.Map {
			path := "/" + routeGroup.Base + "/"
			if route.Pattern != "" {
				path += route.Pattern + "/"
			}
			router.
				Methods(route.Method).
				Path(path).
				Name(name).
				Handler(route.HandlerFunc)
		}
	}
	return router
}
