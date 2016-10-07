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
        "AddLargeScale",
        "POST",
        "/event/largeScale/",
        authenticator.Wrap(AddLargeScale),
    )
    RegisterRoute(
        "GetLargeScale",
        "GET",
        "/event/largeScale/{eventId}/",
        authenticator.Wrap(GetLargeScale),
    )
    RegisterRoute(
        "UpdateLargeScale",
        "PUT",
        "/event/largeScale/{eventId}/",
        authenticator.Wrap(UpdateLargeScale),
    )
    RegisterRoute(
        "PatchLargeScale",
        "PATCH",
        "/event/largeScale/{eventId}/",
        authenticator.Wrap(PatchLargeScale),
    )
}

func AddLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
    sameEventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
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

func GetLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.LargeScaleEventModel{}
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

func UpdateLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
    sameEventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
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


func PatchLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    eventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
    sameEventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
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
    patchMap := bson.M{}
    err = json.Unmarshal(body, &patchMap)
    if err != nil {
        msg := err.Error()
        if strings.Contains(msg, "invalid ObjectId in JSON") {
            msg = "invalid 'eventId'"
        }
        SetHttpError(w, http.StatusBadRequest, msg)
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
    err = db.C(eventModel.Collection()).Find(bson.M{
        "sha1": lastEventRev.Sha1,
    }).One(&eventModel)

    
    {
        value, ok := patchMap["timeZone"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'timeZone'",
                )
                return
            }
            
            eventModel.TimeZone = newValue
            
            delete(patchMap, "timeZone")
        }
    }
    {
        value, ok := patchMap["timeZoneEnable"]
        if ok {
            
            newValue, typeOk := value.(bool)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'timeZoneEnable'",
                )
                return
            }
            
            eventModel.TimeZoneEnable = newValue
            
            delete(patchMap, "timeZoneEnable")
        }
    }
    {
        value, ok := patchMap["calType"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'calType'",
                )
                return
            }
            
            eventModel.CalType = newValue
            
            delete(patchMap, "calType")
        }
    }
    {
        value, ok := patchMap["summary"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'summary'",
                )
                return
            }
            
            eventModel.Summary = newValue
            
            delete(patchMap, "summary")
        }
    }
    {
        value, ok := patchMap["description"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'description'",
                )
                return
            }
            
            eventModel.Description = newValue
            
            delete(patchMap, "description")
        }
    }
    {
        value, ok := patchMap["icon"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'icon'",
                )
                return
            }
            
            eventModel.Icon = newValue
            
            delete(patchMap, "icon")
        }
    }
    {
        value, ok := patchMap["notifyBefore"]
        if ok {
            
            // json Unmarshal converts int to float64
            newValue, typeOk := value.(float64)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'notifyBefore'",
                )
                return
            }
            
            eventModel.NotifyBefore = int(newValue)
            
            delete(patchMap, "notifyBefore")
        }
    }
    {
        value, ok := patchMap["groupId"]
        if ok {
            
            newValue, typeOk := value.(string)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'groupId'",
                )
                return
            }
            
            eventModel.GroupId = newValue
            
            delete(patchMap, "groupId")
        }
    }
    {
        value, ok := patchMap["scale"]
        if ok {
            
            newValue, typeOk := value.(int64)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'scale'",
                )
                return
            }
            
            eventModel.Scale = newValue
            
            delete(patchMap, "scale")
        }
    }
    {
        value, ok := patchMap["start"]
        if ok {
            
            newValue, typeOk := value.(int64)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'start'",
                )
                return
            }
            
            eventModel.Start = newValue
            
            delete(patchMap, "start")
        }
    }
    {
        value, ok := patchMap["end"]
        if ok {
            
            newValue, typeOk := value.(int64)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'end'",
                )
                return
            }
            
            eventModel.End = newValue
            
            delete(patchMap, "end")
        }
    }
    {
        value, ok := patchMap["durationEnable"]
        if ok {
            
            newValue, typeOk := value.(bool)
            
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'durationEnable'",
                )
                return
            }
            
            eventModel.DurationEnable = newValue
            
            delete(patchMap, "durationEnable")
        }
    }
    if len(patchMap) > 0 {
        for param, _ := range patchMap {
            SetHttpError(
                w,
                http.StatusBadRequest,
                fmt.Sprintf(
                    "extra parameter '%s'",
                    param,
                ),
            )
        }
        return
    }
    _, err = eventModel.GetEvent() // (event, err), for validation
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
        return
    }

    eventModel.Sha1 = ""
    jsonByte, _ := json.Marshal(eventModel)
    eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))

    err = db.C("event_revision").Insert(event_lib.EventRevisionModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        Sha1: eventModel.Sha1,
        Time: time.Now(),
    })
    if err != nil {
        SetHttpError(w, http.StatusBadRequest, err.Error())
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
