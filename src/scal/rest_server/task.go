package rest_server

import (
    "fmt"
    "time"
    "strings"
    //"log"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "crypto/sha1"

    "scal/storage"
    "scal/event_lib"

    "gopkg.in/mgo.v2-unstable"
    "gopkg.in/mgo.v2-unstable/bson"
)


func AddTask(w http.ResponseWriter, r *http.Request) {
    eventModel := event_lib.TaskEventModel{} // DYNAMIC
    // -----------------------------------------------
    userId := 0 // FIXME
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
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
        SetHttpError(w, http.StatusInternalServerError, err.Error())
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
    eventAccess := event_lib.EventAccessModel{
        EventId: eventId,
        OwnerId: userId,
        //AccessUserIds: []int{}
    }
    err = db.C("event_access").Insert(eventAccess)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    eventRev := event_lib.EventRevisionModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        Sha1: eventModel.Sha1,
        Time: time.Now(),
    }
    err = db.C("event_revision").Insert(eventRev)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    err = db.C(eventModel.Collection()).Insert(eventModel)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "eventId": eventId.Hex(),
        "sha1": eventRev.Sha1,
    })
}

func GetTask(w http.ResponseWriter, r *http.Request) {
    eventModel := event_lib.TaskEventModel{}
    // -----------------------------------------------
    userId := 0 // FIXME
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    var ok bool
    byEventId := map[string]string{
        "eventId": "" ,
    }
    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    err = json.Unmarshal(body, &byEventId)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    db, err := storage.GetDB()
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    eventIdHex, ok := byEventId["eventId"]
    if !ok {
        SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
        return
    }
    if !bson.IsObjectIdHex(eventIdHex) {
        SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
        return
        // to avoid panic!
    }
    eventId := bson.ObjectIdHex(eventIdHex)

    eventAccess := event_lib.EventAccessModel{}
    err = db.C("event_access").Find(bson.M{
        "_id": eventId,
    }).One(&eventAccess)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    if !eventAccess.UserCanRead(userId) {
        SetHttpError(w, http.StatusUnauthorized, "you don't have access to this event")
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
            SetHttpError(w, http.StatusInternalServerError, err.Error())
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
            SetHttpError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    eventModel.Id = eventId
    json.NewEncoder(w).Encode(eventModel)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
    eventModel := event_lib.TaskEventModel{} // DYNAMIC
    sameEventModel := event_lib.TaskEventModel{} // DYNAMIC
    // -----------------------------------------------
    userId := 0 // FIXME
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
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
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    if eventModel.Id == "" {
        SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
        return
    }
    eventId := eventModel.Id

    // check if event exists, and has access to
    eventAccess := event_lib.EventAccessModel{}
    err = db.C("event_access").Find(bson.M{
        "_id": eventId,
    }).One(&eventAccess)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    if eventAccess.OwnerId != userId {
        SetHttpError(w, http.StatusUnauthorized, "you don't have write access to this event")
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
            SetHttpError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }
    */

    eventModel.Id = ""
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

