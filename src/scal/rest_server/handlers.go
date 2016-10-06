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

func SplitURL(u *url.URL) []string {
    return strings.Split(strings.Trim(u.Path, "/"), "/")
}

func StartRestServer() {
    router := NewRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
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

func SetHttpErrorInternal(w http.ResponseWriter, err error) {
    SetHttpError(w, http.StatusInternalServerError, err.Error())
}

func CopyEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
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
    r.Body.Close()
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


func GetUngroupedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
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
