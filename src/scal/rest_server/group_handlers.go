package rest_server

import (
    //"fmt"
    "time"
    "net"
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

const ALLOW_DELETE_DEFAULT_GROUP = true

// time.RFC3339 == "2006-01-02T15:04:05Z07:00"

func GetGroupList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
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
    err = db.C("event_group").Find(bson.M{
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

func UpdateGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    parts := SplitURL(r.URL)
    groupIdHex := parts[len(parts)-1]
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

    newGroupModel := event_lib.EventGroupModel{}

    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
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
    db.C("event_group").Find(bson.M{"_id": groupId}).One(&oldGroupModel)
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
    err = db.C("event_group").Update(
        bson.M{"_id": groupId},
        oldGroupModel,
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
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

func DeleteGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    email := r.Username
    parts := SplitURL(r.URL)
    groupIdHex := parts[len(parts)-1]
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
    if *userModel.DefaultGroupId == groupId {
        if !ALLOW_DELETE_DEFAULT_GROUP {
            SetHttpError(
                w,
                http.StatusForbidden,
                "you can not delete your default event group",
            )
            return
        }
        userModel.DefaultGroupId = nil
        err = db.C("users").Update(
            bson.M{"email": email},
            userModel,
        )
        if err != nil {
            SetHttpErrorInternal(w, err)
            return
        }
    }

    eventAccessCol := db.C("event_access")

    var eventAccessModels []event_lib.EventAccessModel
    err = eventAccessCol.Find(bson.M{
        "groupId": groupId,
    }).All(&eventAccessModels)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if eventAccessModels != nil {
        for _, eventAccessModel := range eventAccessModels {
            if eventAccessModel.OwnerEmail != email {
                // send an Email to {eventAccessModel.OwnerEmail}
                // to inform the event owner, and let him move this
                // (ungrouped) event into his default (or any other) group
                // FIXME
            }
            // insert a new record to "event_access_change_log" // FIXME
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
            err = db.C("event_access_change_log").Insert(
                bson.M{
                    "time": now,
                    "email": email,
                    "remoteIp": remoteIp,
                    "eventId": eventAccessModel.EventId,
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
    }
    err = db.C("event_group").Remove(bson.M{"_id": groupId})
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
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
    err = db.C("event_access").Find(cond).All(&results)
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
    email := r.Username
    parts := SplitURL(r.URL)
    groupIdHex := parts[len(parts)-2]
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
    db.C("event_group").Find(bson.M{"_id": groupId}).One(&groupModel)
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
    err = db.C("event_access").Pipe(pipeline).All(&results)
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
    email := r.Username
    parts := SplitURL(r.URL)
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
    db.C("event_group").Find(bson.M{"_id": groupId}).One(&groupModel)
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
        err = db.C("event_access").Pipe([]bson.M{
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
        err = db.C("event_access").Pipe([]bson.M{
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

