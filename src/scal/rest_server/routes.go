package rest_server

import (
    "net/http"

    "github.com/gorilla/mux"
)

type Route struct {
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type RouteMap map[string]Route

func NewRouter() *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for name, route := range routeMap {
        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(name).
            Handler(route.HandlerFunc)
    }

    return router
}

func RegisterRoute(
    name string,
    method string,
    pattern string,
    handler http.HandlerFunc,
){
    routeMap[name] = Route{
        method,
        pattern,
        handler,
    }
}

var routeMap = RouteMap{
}



