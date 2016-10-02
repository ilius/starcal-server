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

const ALLOW_DELETE_DEFAULT_GROUP = true

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

    err = db.C("event_access").Find(bson.M{
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
