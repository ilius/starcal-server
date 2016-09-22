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
        "GetUngroupedEvents",
        "GET",
        "/event/ungrouped/",
        authenticator.Wrap(GetUngroupedEvents),
    },
    Route{
        "GetGroupList",
        "GET",
        "/event/groups/",
        authenticator.Wrap(GetGroupList),
    },
    Route{
        "AddGroup",
        "POST",
        "/event/groups/",
        authenticator.Wrap(AddGroup),
    },
    Route{
        "UpdateGroup",
        "PUT",
        "/event/groups/{groupId}/",
        authenticator.Wrap(UpdateGroup),
    },
    Route{
        "GetGroup",
        "GET",
        "/event/groups/{groupId}/",
        authenticator.Wrap(GetGroup),
    },
    Route{
        "GetGroupEventList",
        "GET",
        "/event/groups/{groupId}/events/",
        authenticator.Wrap(GetGroupEventList),
    },
    Route{
        "CopyEvent",
        "POST",
        "/event/copy/",
        authenticator.Wrap(CopyEvent),
    },
    Route{
        "AddTask",
        "POST",
        "/event/task/",
        authenticator.Wrap(AddTask),
    },
    Route{
        "GetTask",
        "GET",
        "/event/task/{eventId}/",
        authenticator.Wrap(GetTask),
    },
    Route{
        "UpdateTask",
        "PUT",
        "/event/task/{eventId}/",
        authenticator.Wrap(UpdateTask),
    },
}



