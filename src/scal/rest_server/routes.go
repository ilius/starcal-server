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
    "Index": Route{
        "GET",
        "/",
        authenticator.Wrap(Index),
    },
    "GetUngroupedEvents": Route{
        "GET",
        "/event/ungrouped/",
        authenticator.Wrap(GetUngroupedEvents),
    },
    "DeleteEvent": Route{
        "DELETE",
        "/event/{eventType}/{eventId}/",// we ignore {eventType}
        authenticator.Wrap(DeleteEvent),
    },
    "CopyEvent": Route{
        "POST",
        "/event/copy/",
        authenticator.Wrap(CopyEvent),
    },
    "SetEventGroupId": Route{
        "PUT",
        "/event/{eventType}/{eventId}/groupId/",// we ignore {eventType}
        authenticator.Wrap(SetEventGroupId),
    },
}



