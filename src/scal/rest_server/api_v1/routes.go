package api_v1

import (
	"fmt"
	"net/http"

	"github.com/ilius/ripo"
	"github.com/julienschmidt/httprouter"
)

type Route struct {
	Method      string
	Pattern     string
	Handler     ripo.Handler
	HandlerFunc http.HandlerFunc // used only in util_handlers.go
}

func (route Route) GetHandlerFunc() http.HandlerFunc {
	if route.Handler != nil {
		return ripo.TranslateHandler(route.Handler)
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
	router := httprouter.New()
	for _, routeGroup := range routeGroups {
		for _, route := range routeGroup.Map {
			path := "/" + routeGroup.Base + "/"
			if route.Pattern != "" {
				path += route.Pattern + "/"
			}
			handlerFunc := route.GetHandlerFunc()
			router.Handle(
				route.Method,
				path,
				func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
					r.ParseForm()
					for _, p := range params {
						r.Form.Add(p.Key, p.Value)
					}
					handlerFunc(w, r)
				},
			)
		}
	}
	return router
}
