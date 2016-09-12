package rest_server

import (
    "fmt"
    "html"
    "log"
    "net/http"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/mgo.v2-unstable/bson"
    "gopkg.in/mgo.v2-unstable"
    //"github.com/gorilla/mux"

    "scal/storage"
    "scal/event_lib"
)


func StartRestServer() {
    router := NewRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
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


func CopyEvent(w http.ResponseWriter, r *http.Request) {
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
    oldEventIdHex, ok := byEventId["eventId"]
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

    eventAccess := event_lib.EventAccessModel{}
    err = db.C("event_access").Find(bson.M{
        "_id": oldEventId,
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
        "eventId": oldEventId,
    }).Sort("-time").One(&eventRev)
    if err != nil {
        if err == mgo.ErrNotFound {
            SetHttpError(w, http.StatusBadRequest, "event not found")
        } else {
            SetHttpError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    newEventId := bson.NewObjectId()

    newEventAccess := event_lib.EventAccessModel{
        EventId: newEventId,
        OwnerId: userId,
        //AccessUserIds: []int{}
    }
    err = db.C("event_access").Insert(newEventAccess)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    eventRev.EventId = newEventId
    err = db.C("event_revision").Insert(eventRev)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "eventType": eventRev.EventType,
        "eventId": newEventId.Hex(),
        "sha1": eventRev.Sha1,
    })

}
