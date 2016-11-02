package rest_server

import (
    "fmt"
    "time"
    "net"
    "net/http"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/mgo.v2/bson"
    //"gopkg.in/mgo.v2"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"
    "scal/storage"
    "scal/event_lib"
)

const ALLOW_DELETE_DEFAULT_GROUP = true

// time.RFC3339 == "2006-01-02T15:04:05Z07:00"

func init() {
    RegisterRoute(
        "GetGroupList",
        "GET",
        "/event/groups/",
        authenticator.Wrap(GetGroupList),
    )
    RegisterRoute(
        "AddGroup",
        "POST",
        "/event/groups/",
        authenticator.Wrap(AddGroup),
    )
    RegisterRoute(
        "UpdateGroup",
        "PUT",
        "/event/groups/{groupId}/",
        authenticator.Wrap(UpdateGroup),
    )
    RegisterRoute(
        "GetGroup",
        "GET",
        "/event/groups/{groupId}/",
        authenticator.Wrap(GetGroup),
    )
    RegisterRoute(
        "DeleteGroup",
        "DELETE",
        "/event/groups/{groupId}/",
        authenticator.Wrap(DeleteGroup),
    )
    RegisterRoute(
        "GetGroupEventList",
        "GET",
        "/event/groups/{groupId}/events/",
        authenticator.Wrap(GetGroupEventList),
    )
    RegisterRoute(
        "GetGroupEventsFull",
        "GET",
        "/event/groups/{groupId}/events-full/",
        authenticator.Wrap(GetGroupEventsFull),
    )
    RegisterRoute(
        "GetGroupModifiedEvents",
        "GET",
        "/event/groups/{groupId}/modified-events/{sinceDateTime}/",
        authenticator.Wrap(GetGroupModifiedEvents),
    )
    RegisterRoute(
        "GetGroupMovedEvents",
        "GET",
        "/event/groups/{groupId}/moved-events/{sinceDateTime}/",
        authenticator.Wrap(GetGroupMovedEvents),
    )
}

func GetGroupList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    type resultModel struct {
        GroupId bson.ObjectId       `bson:"_id" json:"groupId"`
        Title string                `bson:"title" json:"title"`
        OwnerEmail string           `bson:"ownerEmail" json:"ownerEmail"`
    }
    var results []resultModel
    err = db.C(storage.C_group).Find(bson.M{
        "$or": []bson.M{
            bson.M{
                "ownerEmail": email,
            },
            bson.M{
                "readAccessEmails": email,// works :D
            },
        },
    }).All(&results)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if results == nil {
        results = make([]resultModel, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "groups": results,
    })
}

func AddGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    groupModel := event_lib.EventGroupModel{}

    body, _ := ioutil.ReadAll(r.Body)
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
    err = storage.Insert(db, groupModel)

    json.NewEncoder(w).Encode(map[string]string{
        "groupId": groupId.Hex(),
    })
}

func UpdateGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    groupId := ObjectIdFromURL(w, r, "groupId", 0)
    if groupId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    newGroupModel := event_lib.EventGroupModel{}

    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &newGroupModel)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    if newGroupModel.Id != "" {
        SetHttpError(
            w,
            http.StatusBadRequest,
            "can not specify 'groupId'",
        )
        return
    }
    if newGroupModel.OwnerEmail != "" {
        SetHttpError(
            w,
            http.StatusBadRequest,
            "can not specify 'ownerEmail'",
        )
        return
    }
    if newGroupModel.Title == "" {
        SetHttpError(
            w,
            http.StatusBadRequest,
            "missing or empty 'title'",
        )
        return
    }

    var oldGroupModel *event_lib.EventGroupModel
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&oldGroupModel)
    if oldGroupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }
    if oldGroupModel.OwnerEmail != email {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have write access to this event group",
        )
        return
    }
    oldGroupModel.Title             = newGroupModel.Title
    oldGroupModel.AddAccessEmails   = newGroupModel.AddAccessEmails
    oldGroupModel.ReadAccessEmails  = newGroupModel.ReadAccessEmails
    err = storage.Update(db, oldGroupModel)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
}

func GetGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    groupId := ObjectIdFromURL(w, r, "groupId", 0)
    if groupId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    var groupModel *event_lib.EventGroupModel
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
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

func DeleteGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    groupId := ObjectIdFromURL(w, r, "groupId", 0)
    if groupId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    var groupModel *event_lib.EventGroupModel
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }
    if groupModel.OwnerEmail != email {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you are not allowed to delete this event group",
        )
        return
    }

    userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }
    if *userModel.DefaultGroupId == *groupId {
        if !ALLOW_DELETE_DEFAULT_GROUP {
            SetHttpError(
                w,
                http.StatusForbidden,
                "you can not delete your default event group",
            )
            return
        }
        userModel.DefaultGroupId = nil
        err = storage.Update(db, userModel)
        if err != nil {
            SetHttpErrorInternal(w, err)
            return
        }
    }

    eventAccessCol := db.C(storage.C_access)

    var eventAccessModels []event_lib.EventAccessModel
    err = eventAccessCol.Find(bson.M{
        "groupId": groupId,
    }).All(&eventAccessModels)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    for _, eventAccessModel := range eventAccessModels {
        if eventAccessModel.OwnerEmail != email {
            // send an Email to {eventAccessModel.OwnerEmail}
            // to inform the event owner, and let him move this
            // (ungrouped) event into his default (or any other) group
            // FIXME
        }
        // insert a new record to storage.C_accessChangeLog // FIXME
        eventAccessModel.GroupId = nil
        err = eventAccessCol.Update(
            bson.M{"_id": eventAccessModel.EventId},
            eventAccessModel,
        )
        if err != nil {
            SetHttpErrorInternal(w, err)
            return
        }
        now := time.Now()
        err = db.C(storage.C_accessChangeLog).Insert(
            bson.M{
                "time": now,
                "email": email,
                "remoteIp": remoteIp,
                "eventId": eventAccessModel.EventId,
                "funcName": "DeleteGroup",
                "groupId": []interface{}{
                    groupId,
                    nil,
                },
            },
        )
        if err != nil {
            SetHttpErrorInternal(w, err)
            return
        }
    }
    err = storage.Remove(db, groupModel)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
}

func GetGroupEventList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    groupId := ObjectIdFromURL(w, r, "groupId", 1)
    if groupId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    var groupModel *event_lib.EventGroupModel
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
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
    var cond bson.M
    if groupModel.EmailCanRead(email) {
        cond = bson.M{
            "groupId": groupId,
        }
    } else {
        cond = bson.M{
            "groupId": groupId,
            "$or": [2]bson.M{
                bson.M{
                    "ownerEmail": email,
                },
                bson.M{
                    "accessEmails": email,// works :D
                },
            },
        }
    }
    err = db.C(storage.C_access).Find(cond).All(&results)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if results == nil {
        results = make([]resultModel, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "events": results,
    })
}


func GetGroupEventsFull(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    groupId := ObjectIdFromURL(w, r, "groupId", 1)
    if groupId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    var groupModel *event_lib.EventGroupModel
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }
    var results []bson.M
    var pipeline []bson.M
    if groupModel.EmailCanRead(email) {
        pipeline = []bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$lookup": bson.M{
                "from": "event_revision",
                "localField": "_id",
                "foreignField": "eventId",
                "as": "revision",
            }},
            {"$unwind": "$revision"},
            {"$group": bson.M{
                "_id": "$_id",
                "eventType": bson.M{"$first": "$eventType"},
                "ownerEmail": bson.M{"$first": "$ownerEmail"},
                "accessEmails": bson.M{"$first": "$accessEmails"},
                "lastModifiedTime": bson.M{"$first": "$revision.time"},
                "lastSha1": bson.M{"$first": "$revision.sha1"},
            }},
            {"$lookup": bson.M{
                "from": "event_data",
                "localField": "lastSha1",
                "foreignField": "sha1",
                "as": "data",
            }},
            {"$unwind": "$data"},
        }
    } else {
        pipeline = []bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$match": bson.M{
                "$or": []bson.M{
                    bson.M{"ownerEmail": email},
                    bson.M{"accessEmails": email},
                },
            }},
            {"$lookup": bson.M{
                "from": "event_revision",
                "localField": "_id",
                "foreignField": "eventId",
                "as": "revision",
            }},
            {"$unwind": "$revision"},
            {"$group": bson.M{
                "_id": "$_id",
                "eventType": bson.M{"$first": "$eventType"},
                "ownerEmail": bson.M{"$first": "$ownerEmail"},
                "accessEmails": bson.M{"$first": "$accessEmails"},
                "lastModifiedTime": bson.M{"$first": "$revision.time"},
                "lastSha1": bson.M{"$first": "$revision.sha1"},
            }},
            {"$lookup": bson.M{
                "from": "event_data",
                "localField": "lastSha1",
                "foreignField": "sha1",
                "as": "data",
            }},
            {"$unwind": "$data"},
        }
    }
    err = db.C(storage.C_access).Pipe(pipeline).All(&results)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if results == nil {
        results = make([]bson.M, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "events": results,
    })
}


func GetGroupModifiedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    //groupId := ObjectIdFromURL(w, r, "groupId", 2)
    //if groupId==nil { return }
    parts := SplitURL(r.URL)
    if len(parts) < 3 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    groupIdHex := parts[len(parts)-3]
    sinceStr := parts[len(parts)-1] // datetime string
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
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
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }

    since, err := time.Parse(time.RFC3339, sinceStr)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    //json.NewEncoder(w).Encode(bson.M{"sinceDateTime": since})

    results := []bson.M{}
    if groupModel.EmailCanRead(email) {
        err = db.C(storage.C_access).Pipe([]bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$lookup": bson.M{
                "from": "event_revision",
                "localField": "_id",
                "foreignField": "eventId",
                "as": "revision",
            }},
            {"$unwind": "$revision"},
            {"$match": bson.M{
                "revision.time": bson.M{
                    "$gt": since,
                },
            }},
            {"$sort": bson.M{"revision.time": -1}},
            {"$group": bson.M{
                "_id": "$_id",
                "eventType": bson.M{"$first": "$eventType"},
                "ownerEmail": bson.M{"$first": "$ownerEmail"},
                "lastModifiedTime": bson.M{"$first": "$revision.time"},
                "lastSha1": bson.M{"$first": "$revision.sha1"},
            }},
            {"$lookup": bson.M{
                "from": "event_data",
                "localField": "lastSha1",
                "foreignField": "sha1",
                "as": "data",
            }},
            {"$unwind": "$data"},
        }).All(&results)
    } else {
        err = db.C(storage.C_access).Pipe([]bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$match": bson.M{
                "$or": []bson.M{
                    bson.M{"ownerEmail": email},
                    bson.M{"accessEmails": email},
                },
            }},
            {"$lookup": bson.M{
                "from": "event_revision",
                "localField": "_id",
                "foreignField": "eventId",
                "as": "revision",
            }},
            {"$unwind": "$revision"},
            {"$match": bson.M{
                "revision.time": bson.M{
                    "$gt": since,
                },
            }},
            {"$sort": bson.M{"revision.time": -1}},
            {"$group": bson.M{
                "_id": "$_id",
                "eventType": bson.M{"$first": "$eventType"},
                "ownerEmail": bson.M{"$first": "$ownerEmail"},
                "lastModifiedTime": bson.M{"$first": "$revision.time"},
                "lastSha1": bson.M{"$first": "$revision.sha1"},
            }},
            {"$lookup": bson.M{
                "from": "event_data",
                "localField": "lastSha1",
                "foreignField": "sha1",
                "as": "data",
            }},
            {"$unwind": "$data"},
        }).All(&results)
    }
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        "groupId": groupId,
        "since_datetime": since,
        "modified_events": results,
    })

}


func GetGroupMovedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    //groupId := ObjectIdFromURL(w, r, "groupId", 2)
    //if groupId==nil { return }
    parts := SplitURL(r.URL)
    if len(parts) < 3 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    groupIdHex := parts[len(parts)-3]
    sinceStr := parts[len(parts)-1] // datetime string
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
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
    db.C(storage.C_group).Find(bson.M{"_id": groupId}).One(&groupModel)
    if groupModel == nil {
        SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
        return
    }

    since, err := time.Parse(time.RFC3339, sinceStr)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    //json.NewEncoder(w).Encode(bson.M{"sinceDateTime": since})

    results := []bson.M{}
    if groupModel.EmailCanRead(email) {
        err = db.C(storage.C_accessChangeLog).Pipe([]bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$match": bson.M{
                "time": bson.M{
                    "$gt": since,
                },
            }},
            {"$sort": bson.M{"time": -1}},
            {"$group": bson.M{
                "_id": "$eventId",
                "time": bson.M{"$first": "$time"},
                "groupId": bson.M{"$first": "$groupId"},
            }},
        }).All(&results)
    } else {
        err = db.C(storage.C_access).Pipe([]bson.M{
            {"$match": bson.M{
                "groupId": groupId,
            }},
            {"$match": bson.M{
                "time": bson.M{
                    "$gt": since,
                },
            }},
            {"$sort": bson.M{"time": -1}},
            {"$lookup": bson.M{
                "from": "event_access",
                "localField": "eventId",
                "foreignField": "_id",
                "as": "access",
            }},
            {"$unwind": "$access"},
            {"$match": bson.M{
                "$or": []bson.M{
                    bson.M{"access.ownerEmail": email},
                    bson.M{"access.accessEmails": email},
                },
            }},
            {"$group": bson.M{
                "_id": "$eventId",
                "time": bson.M{"$first": "$time"},
                "groupId": bson.M{"$first": "$groupId"},
            }},
        }).All(&results)
    }
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        "groupId": groupId,
        "since_datetime": since,
        "moved_events": results,
    })

}

