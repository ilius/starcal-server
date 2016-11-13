package rest_server

import (
    "reflect"
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventMeta.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't have write access to this event")
        return
    }
    now := time.Now()
    metaChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "DeleteEvent",
        "ownerEmail": []interface{}{
            eventMeta.OwnerEmail,
            nil,
        },
    }
    if eventMeta.GroupId != nil {
        metaChangeLog["groupId"] = []interface{}{
            eventMeta.GroupId,
            nil,
        }
    }
    if len(eventMeta.AccessEmails) > 0 {
        metaChangeLog["accessEmails"] = []interface{}{
            eventMeta.AccessEmails,
            nil,
        }
    }
    err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = storage.Insert(db, event_lib.EventRevisionModel{
        EventId: *eventId,
        EventType: eventMeta.EventType,
        Sha1: "",
        Time: now,
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = storage.Remove(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, &oldEventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if !eventMeta.CanRead(email) {
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
    if eventMeta.GroupModel != nil {
        if eventMeta.GroupModel.OwnerEmail == email {
            newGroupId = &eventMeta.GroupModel.Id // == eventMeta.GroupId
        }
    }

    now := time.Now()
    err = db.C(storage.C_eventMetaChangeLog).Insert(
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

    err = storage.Insert(db, event_lib.EventMetaModel{
        EventId: newEventId,
        EventType: eventMeta.EventType,
        CreationTime: time.Now(),
        OwnerEmail: email,
        GroupId: newGroupId,
        //AccessEmails: []string{}// must not copy AccessEmails
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventMeta.OwnerEmail != email {
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
    metaChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventGroupId",
        "groupId": []interface{}{
            eventMeta.GroupId,
            newGroupId,
        },
    }
    /*
    addedAccessEmails := Set(
        eventMeta.GroupModel.ReadAccessEmails,
    ).Difference(newGroupModel.ReadAccessEmails)
    if addedAccessEmails {
        metaChangeLog["addedAccessEmails"] = addedAccessEmails
    }
    */
    err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    /*userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }*/
    eventMeta.GroupId = &newGroupId
    err = storage.Update(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
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
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if !eventMeta.CanRead(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you don't have access to this event",
        )
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        //"eventId": eventId.Hex(),
        "ownerEmail": eventMeta.OwnerEmail,
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventMeta.OwnerEmail != email {
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
    err = db.C(storage.C_eventMetaChangeLog).Insert(bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventOwner",
        "ownerEmail": []interface{}{
            eventMeta.OwnerEmail,
            newOwnerEmail,
        },
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventMeta.OwnerEmail = newOwnerEmail
    err = storage.Update(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    // send an E-Mail to `newOwnerEmail` FIXME
    json.NewEncoder(w).Encode(bson.M{})
}

func GetEventMetaModelFromRequest(
    w http.ResponseWriter,
    r *auth.AuthenticatedRequest,
) *event_lib.EventMetaModel {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return nil }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return nil
    }
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return nil
    }
    if !eventMeta.CanReadFull(email) {
        SetHttpError(
            w,
            http.StatusForbidden,
            "you can't meta information of this event",
        )
        return nil
    }
    return eventMeta
}

func GetEventMeta(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    // includes owner, creation time, groupId, access info, attendings info
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventMeta := GetEventMetaModelFromRequest(w, r)
    if eventMeta == nil {
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        //"eventId": eventMeta.EventId.Hex(),
        "ownerEmail": eventMeta.OwnerEmail,
        "creationTime": eventMeta.CreationTime,
        "groupId": eventMeta.GroupIdHex(),
        "isPublic": eventMeta.IsPublic,
        "accessEmails": eventMeta.AccessEmails,
        "publicJoinOpen": eventMeta.PublicJoinOpen,
        "maxAttendees": eventMeta.MaxAttendees,
        "attendingEmails": eventMeta.GetAttendingEmails(db),
        "notAttendingEmails": eventMeta.GetNotAttendingEmails(db),
        "maybeAttendingEmails": eventMeta.GetMaybeAttendingEmails(db),
    })
}

func GetEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventMeta := GetEventMetaModelFromRequest(w, r)
    if eventMeta == nil {
        return
    }
    json.NewEncoder(w).Encode(bson.M{
        //"eventId": eventMeta.EventId.Hex(),
        "isPublic": eventMeta.IsPublic,
        "accessEmails": eventMeta.AccessEmails,
        "publicJoinOpen": eventMeta.PublicJoinOpen,
        "maxAttendees": eventMeta.MaxAttendees,
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
        IsPublic *bool                   `json:"isPublic"`
        AccessEmails *[]string           `json:"accessEmails"`
        PublicJoinOpen *bool             `json:"publicJoinOpen"`
        MaxAttendees *int                `json:"maxAttendees"`
    }{
        nil,
        nil,
        nil,
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventMeta.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't own this event")
        return
    }

    newIsPublic := inputStruct.IsPublic
    if newIsPublic==nil {
        SetHttpError(w, http.StatusBadRequest, "missing 'isPublic'")
        return
    }
    newAccessEmails := inputStruct.AccessEmails
    if newAccessEmails==nil {
        SetHttpError(w, http.StatusBadRequest, "missing 'accessEmails'")
        return
    }
    newPublicJoinOpen := inputStruct.PublicJoinOpen
    if newPublicJoinOpen == nil {
        SetHttpError(w, http.StatusBadRequest, "missing 'publicJoinOpen'")
        return
    }
    newMaxAttendees := inputStruct.MaxAttendees
    if newMaxAttendees == nil {
        SetHttpError(w, http.StatusBadRequest, "missing 'maxAttendees'")
        return
    }

    now := time.Now()
    metaChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "SetEventAccess",
    }
    if *newIsPublic != eventMeta.IsPublic {
        metaChangeLog["isPublic"] = []interface{}{
            eventMeta.IsPublic,
            newIsPublic,
        }
        eventMeta.IsPublic = *newIsPublic
    }
    if !reflect.DeepEqual(*newAccessEmails, eventMeta.AccessEmails) {
        metaChangeLog["accessEmails"] = []interface{}{
            eventMeta.AccessEmails,
            newAccessEmails,
        }
        eventMeta.AccessEmails = *newAccessEmails
    }
    if *newPublicJoinOpen != eventMeta.PublicJoinOpen {
        metaChangeLog["publicJoinOpen"] = []interface{}{
            eventMeta.PublicJoinOpen,
            newPublicJoinOpen,
        }
        eventMeta.PublicJoinOpen = *newPublicJoinOpen

    }
    if *newMaxAttendees != eventMeta.MaxAttendees {
        metaChangeLog["maxAttendees"] = []interface{}{
            eventMeta.MaxAttendees,
            newMaxAttendees,
        }
        eventMeta.MaxAttendees = *newMaxAttendees
    }
    err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = storage.Update(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
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

    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    if eventMeta.OwnerEmail != email {
        SetHttpError(w, http.StatusForbidden, "you don't own this event")
        return
    }

    toAddEmail, ok := inputMap["toAddEmail"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'toAddEmail'")
        return
    }
    newAccessEmails := append(eventMeta.AccessEmails, toAddEmail)
    now := time.Now()
    err = db.C(storage.C_eventMetaChangeLog).Insert(bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "funcName": "AppendEventAccess",
        "accessEmails": []interface{}{
            eventMeta.AccessEmails,
            newAccessEmails,
        },
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventMeta.AccessEmails = newAccessEmails
    err = storage.Update(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
}

func JoinEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    /*remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }*/
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    err = eventMeta.Join(db, email)
    if err != nil {
        SetHttpError(w, http.StatusForbidden, err.Error())
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
}

func LeaveEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    /*remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }*/
    eventId := ObjectIdFromURL(w, r, "eventId", 1)
    if eventId==nil { return }

    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    err = eventMeta.Leave(db, email)
    if err != nil {
        SetHttpError(w, http.StatusForbidden, err.Error())
        return
    }
    json.NewEncoder(w).Encode(bson.M{})
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
    err = db.C(storage.C_eventMeta).Find(bson.M{
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
