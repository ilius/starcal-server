package api_v1

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ilius/restpc"
)

type Route struct {
	Method      string
	Pattern     string
	Handler     restpc.Handler
	HandlerFunc http.HandlerFunc // used only in util_handlers.go
}

func (route Route) GetHandlerFunc() http.HandlerFunc {
	if route.Handler != nil {
		return restpc.TranslateHandler(route.Handler)
	}
	if route.HandlerFunc != nil {
		return route.HandlerFunc
	}
	panic(fmt.Sprintf(
		"GetHandlerFunc: not route.Handler nor route.HandlerFunc is set, Method=%v",
		route.Method,
	))
}

type RouteMap map[string]Route

type RouteGroup struct {
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
				Handler(route.GetHandlerFunc())
		}
	}
	return router
}
