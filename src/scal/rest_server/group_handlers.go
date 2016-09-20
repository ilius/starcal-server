package rest_server

import (
    //"fmt"
    "net/http"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/mgo.v2-unstable/bson"
    //"gopkg.in/mgo.v2-unstable"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"
    "scal/storage"
    "scal/event_lib"
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
    type resultModel struct {
        GroupId bson.ObjectId       `bson:"_id" json:"groupId"`
        Title string                `bson:"title" json:"title"`
        OwnerEmail string           `bson:"ownerEmail" json:"ownerEmail"`
    }
    var results []resultModel
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
        results = make([]resultModel, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "groups": results,
    })
}

func AddGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }

    groupModel := event_lib.EventGroupModel{}

    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    err = json.Unmarshal(body, &groupModel)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    if groupModel.Id != "" {
        SetHttpError(
            w,
            http.StatusBadRequest,
            "can not specify 'groupId'",
        )
        return
    }
    if groupModel.OwnerEmail != "" {
        SetHttpError(
            w,
            http.StatusBadRequest,
            "can not specify 'ownerEmail'",
        )
        return
    }

    groupId := bson.NewObjectId()
    groupModel.Id = groupId
    groupModel.OwnerEmail = email
    err = db.C("event_group").Insert(groupModel)

    json.NewEncoder(w).Encode(map[string]string{
        "groupId": groupId.Hex(),
    })
}

func GetGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    parts := SplitURL(r.URL)
    groupIdHex := parts[len(parts)-1]
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    if groupIdHex == "" {
        SetHttpError(w, http.StatusBadRequest, "missing 'groupId'")
        return
    }
    if !bson.IsObjectIdHex(groupIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
        // to avoid panic!
    }
    groupId := bson.ObjectIdHex(groupIdHex)

    var groupModel *event_lib.EventGroupModel
    db.C("event_group").Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }
    if !groupModel.EmailCanRead(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have access to this event group",
        )
        return
    }
    json.NewEncoder(w).Encode(groupModel)
}

func GetGroupEventList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    parts := SplitURL(r.URL)
    groupIdHex := parts[len(parts)-2]
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    if groupIdHex == "" {
        SetHttpError(w, http.StatusBadRequest, "missing 'groupId'")
        return
    }
    if !bson.IsObjectIdHex(groupIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
        // to avoid panic!
    }
    groupId := bson.ObjectIdHex(groupIdHex)
    var groupModel *event_lib.EventGroupModel
    db.C("event_group").Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }

    type resultModel struct {
        EventId bson.ObjectId       `bson:"_id" json:"eventId"`
        EventType string            `bson:"eventType" json:"eventType"`
        //OwnerEmail string         `bson:"ownerEmail" json:"ownerEmail"`
        //AccessEmails []string     `bson:"accessEmails"`
        //GroupId *bson.ObjectId    `bson:"groupId" json:"groupId"`
    }
    var results []resultModel

    db.C("event_access").Find(bson.M{
        "groupId": groupId,
        "$or": [2]bson.M{
            bson.M{
                "ownerEmail": email,
            },
            bson.M{
                "accessEmails": email,// works :D
            },
        },
    }).All(&results)
    if results == nil {
        results = make([]resultModel, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "events": results,
    })
}
