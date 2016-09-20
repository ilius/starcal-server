package rest_server

import (
    //"fmt"
    "net/http"
    "encoding/json"
    //"io/ioutil"

    "gopkg.in/mgo.v2-unstable/bson"
    //"gopkg.in/mgo.v2-unstable"
    //"github.com/gorilla/mux"

    "scal/lib/go-http-auth"
    "scal/storage"
    //"scal/event_lib"
)

func GetGroupList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    type groupResult struct {    
        Id bson.ObjectId            `bson:"_id" json:"groupId"`
        Title string                `bson:"title" json:"title"`
        OwnerEmail string           `bson:"ownerEmail" json:"ownerEmail"`
    }
    var results []groupResult
    db.C("event_group").Find(bson.M{
        "$or": []bson.M{
            bson.M{
                "ownerEmail": email,
            },
            bson.M{
                "readAccessEmails": email,// works :D
            },
        },
    }).All(&results)
    if results == nil {
        results = make([]groupResult, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "groups": results,
    })
}
