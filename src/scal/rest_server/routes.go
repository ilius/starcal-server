package rest_server

import (
    "net/http"

    "github.com/gorilla/mux"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for _, route := range routes {
        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(route.HandlerFunc)
    }

    return router
}

var routes = Routes{
    Route{
        "Index",
        "GET",
        "/",
        authenticator.Wrap(Index),
    },
    Route{
        "RegisterUser",
        "POST",
        "/user/register/",
        RegisterUser,
    },
    Route{
        "SetUserFullName",
        "POST",
        "/user/full-name/set",
        authenticator.Wrap(SetUserFullName),
    },
    Route{
        "UnsetUserFullName",
        "POST",
        "/user/full-name/unset",
        authenticator.Wrap(UnsetUserFullName),
    },
    Route{
        "SetUserDefaultGroupId",
        "POST",
        "/user/default-group-id/set",
        authenticator.Wrap(SetUserDefaultGroupId),
    },
    Route{
        "CopyEvent",
        "POST",
        "/events/copy/",
        authenticator.Wrap(CopyEvent),
    },
    Route{
        "AddTask",
        "POST",
        "/events/task/add/",
        authenticator.Wrap(AddTask),
    },
    Route{
        "GetTask",
        "POST",
        "/events/task/get/",
        authenticator.Wrap(GetTask),
    },
    Route{
        "UpdateTask",
        "POST",
        "/events/task/update/",
        authenticator.Wrap(UpdateTask),
    },
}



