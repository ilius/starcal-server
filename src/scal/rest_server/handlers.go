package rest_server

import (
    "fmt"
    "time"
    "net"
    "net/http"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"
    "scal/storage"
    "scal/event_lib"
)

func init() {
    RegisterRoute(
        "CopyEvent",
        "POST",
        "/event/copy/",
        authenticator.Wrap(CopyEvent),
    )
    RegisterRoute(
        "GetUngroupedEvents",
        "GET",
        "/event/ungrouped/",
        authenticator.Wrap(GetUngroupedEvents),
    )
}



func DeleteEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventId := ObjectIdFromURL(w, r, "eventId", 0)
    if eventId==nil { return }

    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventAccess.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't have write access to this event")
        return
    }
    now := time.Now()
    accessChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "DeleteEvent",
        "ownerEmail": []interface{}{
            eventAccess.OwnerEmail,
            nil,
        },
    }
    if eventAccess.GroupId != nil {
        accessChangeLog["groupId"] = []interface{}{
            eventAccess.GroupId,
            nil,
        }
    }
    if len(eventAccess.AccessEmails) > 0 {
        accessChangeLog["accessEmails"] = []interface{}{
            eventAccess.AccessEmails,
            nil,
        }
    }
    err = db.C(storage.C_accessChangeLog).Insert(accessChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = storage.Insert(db, event_lib.EventRevisionModel{
        EventId: *eventId,
        EventType: eventAccess.EventType,
        Sha1: "",
        Time: now,
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = storage.Remove(db, eventAccess)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
}

func CopyEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    var ok bool
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    inputMap := map[string]string{}
    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &inputMap)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    oldEventIdHex, ok := inputMap["eventId"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
        return
    }
    if !bson.IsObjectIdHex(oldEventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    oldEventId := bson.ObjectIdHex(oldEventIdHex)

    eventAccess, err := event_lib.LoadEventAccessModel(db, &oldEventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if !eventAccess.EmailCanRead(email) {
        SetHttpError(w, http.StatusForbidden, "you don't have access to this event")
        return
    }

    eventRev := event_lib.EventRevisionModel{}
    err = db.C(storage.C_revision).Find(bson.M{
        "eventId": oldEventId,
    }).Sort("-time").One(&eventRev)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }

    newEventId := bson.NewObjectId()

    userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }

    newGroupId := userModel.DefaultGroupId
    if eventAccess.GroupModel != nil {
        if eventAccess.GroupModel.OwnerEmail == email {
            newGroupId = &eventAccess.GroupModel.Id // == eventAccess.GroupId
        }
    }

    now := time.Now()
    err = db.C(storage.C_accessChangeLog).Insert(
        bson.M{
            "time": now,
            "email": email,
            "remoteIp": remoteIp,
            "eventId": newEventId,
            "funcName": "CopyEvent",
            "ownerEmail": []interface{}{
                nil,
                email,
            },
        },
        bson.M{
            "time": now,
            "email": email,
            "remoteIp": remoteIp,
            "eventId": newEventId,
            "funcName": "CopyEvent",
            "groupId": []interface{}{
                nil,
                newGroupId,
            },
        },
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    err = storage.Insert(db, event_lib.EventAccessModel{
        EventId: newEventId,
        EventType: eventAccess.EventType,
        OwnerEmail: email,
        GroupId: newGroupId,
        //AccessEmails: []string{}
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventRev.EventId = newEventId
    err = storage.Insert(db, eventRev)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "eventType": eventRev.EventType,
        "eventId": newEventId.Hex(),
        "sha1": eventRev.Sha1,
    })

}

func SetEventGroupId(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    var ok bool
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    inputMap := map[string]string{}
    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &inputMap)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventAccess.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't have write access to this event")
        return
    }

    newGroupIdHex, ok := inputMap["newGroupId"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'newGroupId'")
        return
    }
    if !bson.IsObjectIdHex(newGroupIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'newGroupId'")
        return
        // to avoid panic!
    }
    newGroupId := bson.ObjectIdHex(newGroupIdHex)
    newGroupModel := event_lib.EventGroupModel{}
    err = db.C(storage.C_group).Find(bson.M{
        "_id": newGroupId,
    }).One(&newGroupModel)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if !newGroupModel.EmailCanAdd(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have write access to this group",
        )
        return
    }

    now := time.Now()
    accessChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventGroupId",
        "groupId": []interface{}{
            eventAccess.GroupId,
            newGroupId,
        },
    }
    /*
    addedAccessEmails := Set(
        eventAccess.GroupModel.ReadAccessEmails,
    ).Difference(newGroupModel.ReadAccessEmails)
    if addedAccessEmails {
        accessChangeLog["addedAccessEmails"] = addedAccessEmails
    }
    */
    err = db.C(storage.C_accessChangeLog).Insert(accessChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    /*userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }*/
    eventAccess.GroupId = &newGroupId
    err = storage.Update(db, eventAccess)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
}


func GetEventOwner(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if !eventAccess.EmailCanRead(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have access to this event",
        )
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        //"eventId": eventId.Hex(),
        "ownerEmail": eventAccess.OwnerEmail,
    })
}

func SetEventOwner(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    var ok bool
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    inputMap := map[string]string{}
    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &inputMap)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventAccess.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't own this event")
        return
    }

    newOwnerEmail, ok := inputMap["newOwnerEmail"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'newOwnerEmail'")
        return
    }
    // should check if user with `newOwnerEmail` exists?
    userModel := UserModelByEmail(newOwnerEmail, db)
    if userModel == nil {
        SetHttpError(
            w,
            http.StatusBadRequest,
            fmt.Sprintf(
                "user with email '%s' does not exist",
                newOwnerEmail,
            ),
        )
        return
    }
    now := time.Now()
    err = db.C(storage.C_accessChangeLog).Insert(bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventOwner",
        "ownerEmail": []interface{}{
            eventAccess.OwnerEmail,
            newOwnerEmail,
        },
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess.OwnerEmail = newOwnerEmail
    err = storage.Update(db, eventAccess)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    // send an E-Mail to `newOwnerEmail` FIXME
}

func GetEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if !eventAccess.EmailCanRead(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have access to this event",
        )
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        //"eventId": eventId.Hex(),
        "accessEmails": eventAccess.AccessEmails,
    })
}

func SetEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    inputStruct := struct {
        AccessEmails *[]string
    }{
        nil,
    }

    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &inputStruct)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventAccess.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't own this event")
        return
    }

    newAccessEmails := inputStruct.AccessEmails
    if newAccessEmails==nil {
        SetHttpError(w, http.StatusBadRequest, "missing 'accessEmails'")
        return
    }

    now := time.Now()
    err = db.C(storage.C_accessChangeLog).Insert(bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventAccess",
        "accessEmails": []interface{}{
            eventAccess.AccessEmails,
            *newAccessEmails,
        },
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess.AccessEmails = *newAccessEmails
    err = storage.Update(db, eventAccess)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
}

func AppendEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    var ok bool
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    inputMap := map[string]string{}
    body, _ := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &inputMap)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventAccess.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't own this event")
        return
    }

    toAddEmail, ok := inputMap["toAddEmail"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'toAddEmail'")
        return
    }
    newAccessEmails := append(eventAccess.AccessEmails, toAddEmail)
    now := time.Now()
    err = db.C(storage.C_accessChangeLog).Insert(bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "AppendEventAccess",
        "accessEmails": []interface{}{
            eventAccess.AccessEmails,
            newAccessEmails,
        },
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess.AccessEmails = newAccessEmails
    err = storage.Update(db, eventAccess)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
}


func GetUngroupedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    type eventModel struct {
        EventId bson.ObjectId   `bson:"_id" json:"eventId"`
        EventType string        `bson:"eventType" json:"eventType"`
    }
    var events []eventModel
    err = db.C(storage.C_access).Find(bson.M{
        "ownerEmail": email,
        "groupId": nil,
    }).All(&events)
    if events == nil {
        events = make([]eventModel, 0)
    }
    json.NewEncoder(w).Encode(bson.M{
        "events": events,
    })
}
