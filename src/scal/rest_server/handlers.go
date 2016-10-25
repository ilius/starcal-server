package rest_server

import (
    "fmt"
    "time"
    "strings"
    "log"
    "net"
    "net/url"
    "net/http"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/mgo.v2-unstable/bson"
    "gopkg.in/mgo.v2-unstable"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"
    "scal/storage"
    "scal/event_lib"
)

func init() {
    RegisterRoute(
        "Index",
        "GET",
        "/",
        authenticator.Wrap(Index),
    )
    RegisterRoute(
        "DeleteEvent",
        "DELETE",
        "/event/{eventType}/{eventId}/",// we ignore {eventType}
        authenticator.Wrap(DeleteEvent),
    )
    RegisterRoute(
        "CopyEvent",
        "POST",
        "/event/copy/",
        authenticator.Wrap(CopyEvent),
    )
    RegisterRoute(
        "SetEventGroupId",
        "PUT",
        "/event/{eventType}/{eventId}/groupId/",// we ignore {eventType}
        authenticator.Wrap(SetEventGroupId),
    )
    RegisterRoute(
        "GetEventOwner",
        "GET",
        "/event/{eventType}/{eventId}/owner/",// we ignore {eventType}
        authenticator.Wrap(GetEventOwner),
    )
    RegisterRoute(
        "SetEventOwner",
        "PUT",
        "/event/{eventType}/{eventId}/owner/",// we ignore {eventType}
        authenticator.Wrap(SetEventOwner),
    )
    RegisterRoute(
        "GetUngroupedEvents",
        "GET",
        "/event/ungrouped/",
        authenticator.Wrap(GetUngroupedEvents),
    )
}

func SplitURL(u *url.URL) []string {
    return strings.Split(strings.Trim(u.Path, "/"), "/")
}


func Index(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    fmt.Fprintf(w, "Hello, %q", r.Username)
}


func SetHttpError(w http.ResponseWriter, code int, msg string){
    if code == http.StatusInternalServerError {
        log.Print("Internal Server Error: ", msg)
        // don't expose internal error messages to outsiders
        msg = "Internal Server Error"
    }
    jsonByte, _ := json.Marshal(map[string]string{
        "error": msg,
    })
    http.Error(
        w,
        string(jsonByte),
        code,
    )
}

func SetHttpErrorInternalMsg(w http.ResponseWriter, errMsg string) {
    SetHttpError(w, http.StatusInternalServerError, errMsg)
}

func SetHttpErrorInternal(w http.ResponseWriter, err error) {
    SetHttpError(w, http.StatusInternalServerError, err.Error())
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
    parts := SplitURL(r.URL)
    if len(parts) < 1 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-1]
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)

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
    err = db.C("event_access").Remove(
        bson.M{"_id": eventId},
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    now := time.Now()
    accessChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
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
    err = db.C("event_access_change_log").Insert(accessChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = db.C("event_revision").Insert(bson.M{
        "eventId": eventId,
        "eventType": eventAccess.EventType,
        "sha1": nil,
        "time": now,
    })
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
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
    inputMap := map[string]string{
        "eventId": "" ,
    }
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

    eventAccess, err := event_lib.LoadEventAccessModel(db, oldEventId, true)
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
    err = db.C("event_revision").Find(bson.M{
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

    newEventAccess := event_lib.EventAccessModel{
        EventId: newEventId,
        EventType: eventAccess.EventType,
        OwnerEmail: email,
        GroupId: newGroupId,
        //AccessEmails: []string{}
    }
    err = db.C("event_access").Insert(newEventAccess)
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
            "eventId": newEventId,
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
    eventRev.EventId = newEventId
    err = db.C("event_revision").Insert(eventRev)
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
    parts := SplitURL(r.URL)
    if len(parts) < 2 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-2]
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)

    inputMap := map[string]string{
        "newGroupId": "",
    }
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
    err = db.C("event_group").Find(bson.M{
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

    /*userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }*/
    eventAccess.GroupId = &newGroupId
    err = db.C("event_access").Update(
        bson.M{"_id": eventId},
        eventAccess,
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    now := time.Now()
    accessChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "groupId": []interface{}{
            nil,
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
    err = db.C("event_access_change_log").Insert(accessChangeLog)
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
    parts := SplitURL(r.URL)
    if len(parts) < 2 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-2]
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)
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
        //"eventId": eventIdHex,
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
    parts := SplitURL(r.URL)
    if len(parts) < 2 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-2]
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)

    inputMap := map[string]string{
        "newOwnerEmail": "",
    }
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
        SetHttpErrorUserNotFound(w, newOwnerEmail)
        return
    }
    now := time.Now()
    accessChangeLog := bson.M{
        "time": now,
        "email": email,
        "remoteIp": remoteIp,
        "eventId": eventId,
        "ownerEmail": []interface{}{
            eventAccess.OwnerEmail,
            newOwnerEmail,
        },
    }
    err = db.C("event_access_change_log").Insert(accessChangeLog)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    eventAccess.OwnerEmail = newOwnerEmail
    err = db.C("event_access").Update(
        bson.M{"_id": eventId},
        eventAccess,
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    // send an E-Mail to `newOwnerEmail` FIXME
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
    err = db.C("event_access").Find(bson.M{
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
