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
        Index,
    },
    Route{
        "CopyEvent",
        "POST",
        "/events/copy/",
        CopyEvent,
    },
    Route{
        "AddTask",
        "POST",
        "/events/task/add/",
        AddTask,
    },
    Route{
        "GetTask",
        "POST",
        "/events/task/get/",
        GetTask,
    },
    Route{
        "UpdateTask",
        "POST",
        "/events/task/update/",
        UpdateTask,
    },
}



