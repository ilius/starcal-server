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
        "PUT",
        "/user/full-name/",
        authenticator.Wrap(SetUserFullName),
    },
    Route{
        "UnsetUserFullName",
        "DELETE",
        "/user/full-name/",
        authenticator.Wrap(UnsetUserFullName),
    },
    Route{
        "SetUserDefaultGroupId",
        "PUT",
        "/user/default-group-id/",
        authenticator.Wrap(SetUserDefaultGroupId),
    },
    Route{
        "UnsetUserDefaultGroupId",
        "DELETE",
        "/user/default-group-id/",
        authenticator.Wrap(UnsetUserDefaultGroupId),
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
        "/events/task/",
        authenticator.Wrap(AddTask),
    },
    Route{
        "GetTask",
        "GET",
        "/events/task/{eventId}/",
        authenticator.Wrap(GetTask),
    },
    Route{
        "UpdateTask",
        "PUT",
        "/events/task/{eventId}/",
        authenticator.Wrap(UpdateTask),
    },
}



