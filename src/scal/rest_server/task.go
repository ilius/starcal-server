package rest_server

import (
    "fmt"
    "time"
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
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }
    err = db.C(eventModel.Collection()).Insert(eventModel)
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "eventId": eventId.Hex(),
        //"eventId": string(eventId),
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







