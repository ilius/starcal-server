package rest_server

import (
    "fmt"
    "time"
    "strings"
    //"log"
    "io/ioutil"
    "net"
    "net/http"
    "encoding/json"
    "crypto/sha1"

    "gopkg.in/mgo.v2-unstable"
    "gopkg.in/mgo.v2-unstable/bson"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"

    "scal/storage"
    "scal/event_lib"
)

func init(){
    RegisterRoute(
        "AddAllDayTask",
        "POST",
        "/event/allDayTask/",
        authenticator.Wrap(AddAllDayTask),
    )
    RegisterRoute(
        "GetAllDayTask",
        "GET",
        "/event/allDayTask/{eventId}/",
        authenticator.Wrap(GetAllDayTask),
    )
    RegisterRoute(
        "UpdateAllDayTask",
        "PUT",
        "/event/allDayTask/{eventId}/",
        authenticator.Wrap(UpdateAllDayTask),
    )
    
}

func AddAllDayTask(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
    sameEventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
    // -----------------------------------------------
    email := r.Username
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    err = json.Unmarshal(body, &eventModel)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    _, err = eventModel.GetEvent() // (event, err), just for validation
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if eventModel.Id != "" {
        SetHttpError(w, http.StatusBadRequest, "you can't specify 'eventId'")
        return
    }
    eventModel.Sha1 = ""
    jsonByte, _ := json.Marshal(eventModel)
    eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))
    eventId := bson.NewObjectId()
    eventModel.Id = eventId
    userModel := UserModelByEmail(email, db)
    if userModel == nil {
        SetHttpErrorUserNotFound(w, email)
        return
    }
    groupId := userModel.DefaultGroupId
    if eventModel.GroupId != "" {
        if !bson.IsObjectIdHex(eventModel.GroupId) {
            SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
            return
            // to avoid panic!
        }
        var groupModel *event_lib.EventGroupModel
        db.C("event_group").Find(bson.M{
            "_id": bson.ObjectIdHex(eventModel.GroupId),
        }).One(&groupModel)
        if groupModel == nil {
            SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
            return
        }
        if groupModel.OwnerEmail != email {
            SetHttpError(
                w,
                http.StatusForbidden,
                "you don't have write access this event group",
            )
            return
        }
        groupId = &groupModel.Id
    }
    eventAccess := event_lib.EventAccessModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        OwnerEmail: email,
        GroupId: groupId,
        //AccessEmails: []string{}
    }
    err = db.C("event_access").Insert(eventAccess)
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
            "eventId": eventId,
            "ownerEmail": []interface{}{
                nil,
                email,
            },
        },
        bson.M{
            "time": now,
            "email": email,
            "remoteIp": remoteIp,
            "eventId": eventId,
            "groupId": []interface{}{
                nil,
                groupId,
            },
        },
    )
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    err = db.C("event_revision").Insert(event_lib.EventRevisionModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        Sha1: eventModel.Sha1,
        Time: time.Now(),
    })
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    // don't store duplicate eventModel, even if it was added by another user
    // the (underlying) eventModel does not belong to anyone
    // like git's blobs and trees
    err = db.C(eventModel.Collection()).Find(bson.M{
        "sha1": eventModel.Sha1,
    }).One(&sameEventModel)
    if err == mgo.ErrNotFound {
        err = db.C(eventModel.Collection()).Insert(eventModel)
        if err != nil {
            SetHttpError(w, http.StatusBadRequest, err.Error())
            return
        }
    }
    json.NewEncoder(w).Encode(map[string]string{
        "eventId": eventId.Hex(),
        "sha1": eventModel.Sha1,
    })
}

func GetAllDayTask(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.AllDayTaskEventModel{}
    // -----------------------------------------------
    email := r.Username
    //vars := mux.Vars(&r.Request) // vars == map[] // FIXME
    //eventIdHex := vars["eventId"]
    parts := SplitURL(r.URL)
    if len(parts) < 1 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-1]
    fmt.Printf("eventIdHex = %v\n", eventIdHex)
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    if eventIdHex == "" {
        SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
        return
    }
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)

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
        SetHttpError(w, http.StatusForbidden, "you don't have access to this event")
        return
    }

    eventRev := event_lib.EventRevisionModel{}
    err = db.C("event_revision").Find(bson.M{
        "eventId": eventId,
    }).Sort("-time").One(&eventRev)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }

    err = db.C(eventModel.Collection()).Find(bson.M{
        "sha1": eventRev.Sha1,
    }).One(&eventModel)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusInternalServerError, "event snapshot not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }

    eventModel.Id = eventId
    eventModel.GroupId = eventAccess.GroupId.Hex()
    json.NewEncoder(w).Encode(eventModel)
}

func UpdateAllDayTask(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
    sameEventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
    // -----------------------------------------------
    email := r.Username
    //vars := mux.Vars(&r.Request) // vars == map[] // FIXME
    //eventIdHex := vars["eventId"]
    parts := SplitURL(r.URL)
    if len(parts) < 1 {
        SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
        return
    }
    eventIdHex := parts[len(parts)-1]
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    if eventIdHex == "" {
        SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
        return
    }
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)
    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    err = json.Unmarshal(body, &eventModel)
    if err != nil {
        msg := err.Error()
        if strings.Contains(msg, "invalid ObjectId in JSON") {
            msg = "invalid 'eventId'"
        }
        SetHttpError(w, http.StatusBadRequest, msg)
        return
    }
    _, err = eventModel.GetEvent() // (event, err), just for validation
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }

    // check if event exists, and has access to
    eventAccess, err := event_lib.LoadEventAccessModel(db, eventId, false)
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

    /*
    // do we need the last revision? to compare or what?
    lastEventRev := event_lib.EventRevisionModel{}
    err = db.C("event_revision").Find(bson.M{
        "eventId": eventId,
    }).Sort("-time").One(&lastEventRev)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpErrorInternal(w, err)
        }
        return
    }
    */

    if eventModel.Id != "" {
        SetHttpError(w, http.StatusBadRequest, "'eventId' must not be present in JSON")
        return
    }
    if eventModel.GroupId != "" {
        SetHttpError(w, http.StatusBadRequest, "'groupId' must not be present in JSON")
        return
    }
    eventModel.Sha1 = ""
    jsonByte, _ := json.Marshal(eventModel)
    eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))

    eventRev := event_lib.EventRevisionModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        Sha1: eventModel.Sha1,
        Time: time.Now(),
    }
    err = db.C("event_revision").Insert(eventRev)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }

    // don't store duplicate eventModel, even if it was added by another user
    // the (underlying) eventModel does not belong to anyone
    // like git's blobs and trees
    err = db.C(eventModel.Collection()).Find(bson.M{
        "sha1": eventRev.Sha1,
    }).One(&sameEventModel)
    if err == mgo.ErrNotFound {
        err = db.C(eventModel.Collection()).Insert(eventModel)
        if err != nil {
            SetHttpError(w, http.StatusBadRequest, err.Error())
            return
        }
    }

    json.NewEncoder(w).Encode(map[string]string{
        "eventId": eventId.Hex(),
        "sha1": eventRev.Sha1,
    })
}



