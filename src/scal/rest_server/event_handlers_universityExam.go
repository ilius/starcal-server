// if this is a *.go file, don't modify it, it's auto-generated
// from a Django template file named `*.got` inside "templates" directory
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

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    //"github.com/gorilla/mux"

    "scal-lib/go-http-auth"

    "scal/storage"
    "scal/event_lib"
)

func init(){
    RegisterRoute(
        "AddUniversityExam",
        "POST",
        "/event/universityExam/",
        authenticator.Wrap(AddUniversityExam),
    )
    RegisterRoute(
        "GetUniversityExam",
        "GET",
        "/event/universityExam/{eventId}/",
        authenticator.Wrap(GetUniversityExam),
    )
    RegisterRoute(
        "UpdateUniversityExam",
        "PUT",
        "/event/universityExam/{eventId}/",
        authenticator.Wrap(UpdateUniversityExam),
    )
    RegisterRoute(
        "PatchUniversityExam",
        "PATCH",
        "/event/universityExam/{eventId}/",
        authenticator.Wrap(PatchUniversityExam),
    )
    // functions of following operations are defined in handlers.go
    // because their definition does not depend on event type
    // but their URL still contains eventType for sake of compatibilty
    // so we will have to register their routes for each event type
    // we don't use eventType in these functions
    RegisterRoute(
        "DeleteEvent_universityExam",
        "DELETE",
        "/event/universityExam/{eventId}/",
        authenticator.Wrap(DeleteEvent),
    )
    RegisterRoute(
        "SetEventGroupId_universityExam",
        "PUT",
        "/event/universityExam/{eventId}/groupId/",
        authenticator.Wrap(SetEventGroupId),
    )
    RegisterRoute(
        "GetEventOwner_universityExam",
        "GET",
        "/event/universityExam/{eventId}/owner/",
        authenticator.Wrap(GetEventOwner),
    )
    RegisterRoute(
        "SetEventOwner_universityExam",
        "PUT",
        "/event/universityExam/{eventId}/owner/",
        authenticator.Wrap(SetEventOwner),
    )
    RegisterRoute(
        "GetEventMeta_universityExam",
        "GET",
        "/event/universityExam/{eventId}/meta/",
        authenticator.Wrap(GetEventMeta),
    )
    RegisterRoute(
        "GetEventAccess_universityExam",
        "GET",
        "/event/universityExam/{eventId}/access/",
        authenticator.Wrap(GetEventAccess),
    )
    RegisterRoute(
        "SetEventAccess_universityExam",
        "PUT",
        "/event/universityExam/{eventId}/access/",
        authenticator.Wrap(SetEventAccess),
    )
    RegisterRoute(
        "AppendEventAccess_universityExam",
        "POST",
        "/event/universityExam/{eventId}/access/",
        authenticator.Wrap(AppendEventAccess),
    )
    RegisterRoute(
        "JoinEvent_universityExam",
        "GET",
        "/event/universityExam/{eventId}/join/",
        authenticator.Wrap(JoinEvent),
    )
    RegisterRoute(
        "LeaveEvent_universityExam",
        "GET",
        "/event/universityExam/{eventId}/leave/",
        authenticator.Wrap(LeaveEvent),
    )

}

func AddUniversityExam(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    eventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
    sameEventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
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
        db.C(storage.C_group).Find(bson.M{
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
    eventMeta := event_lib.EventMetaModel{
        EventId: eventId,
        EventType: eventModel.Type(),
        CreationTime: time.Now(),
        OwnerEmail: email,
        GroupId: groupId,
        //AccessEmails: []string{}
    }
    now := time.Now()
    err = db.C(storage.C_eventMetaChangeLog).Insert(
        bson.M{
            "time": now,
            "email": email,
            "remoteIp": remoteIp,
            "eventId": eventId,
            "funcName": "AddUniversityExam",
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
            "funcName": "AddUniversityExam",
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
    err = storage.Insert(db, event_lib.EventRevisionModel{
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
        err = storage.Insert(db, eventModel)
        if err != nil {
            SetHttpError(w, http.StatusBadRequest, err.Error())
            return
        }
    }
    err = storage.Insert(db, eventMeta)
    if err != nil {
        SetHttpErrorInternal(w, err)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{
        "eventId": eventId.Hex(),
        "sha1": eventModel.Sha1,
    })
}

func GetUniversityExam(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    eventModel := event_lib.UniversityExamEventModel{}
    // -----------------------------------------------
    email := r.Username
    //vars := mux.Vars(&r.Request) // vars == map[] // FIXME
    eventId := ObjectIdFromURL(w, r, "eventId", 0)
    if eventId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
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
        SetHttpError(w, http.StatusForbidden, "you don't have access to this event")
        return
    }

    eventRev := event_lib.EventRevisionModel{}
    err = db.C(storage.C_revision).Find(bson.M{
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

    eventModel.Id = *eventId
    eventModel.GroupId = eventMeta.GroupIdHex()
    json.NewEncoder(w).Encode(eventModel)
}

func UpdateUniversityExam(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    eventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
    sameEventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
    // -----------------------------------------------
    email := r.Username
    //vars := mux.Vars(&r.Request) // vars == map[] // FIXME
    eventId := ObjectIdFromURL(w, r, "eventId", 0)
    if eventId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    body, _ := ioutil.ReadAll(r.Body)
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
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
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

    /*
    // do we need the last revision? to compare or what?
    lastEventRev := event_lib.EventRevisionModel{}
    err = db.C(storage.C_revision).Find(bson.M{
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
        EventId: *eventId,
        EventType: eventModel.Type(),
        Sha1: eventModel.Sha1,
        Time: time.Now(),
    }
    err = storage.Insert(db, eventRev)
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
        err = storage.Insert(db, eventModel)
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
func PatchUniversityExam(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
    defer r.Body.Close()
    eventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
    sameEventModel := event_lib.UniversityExamEventModel{} // DYNAMIC
    // -----------------------------------------------
    email := r.Username
    //vars := mux.Vars(&r.Request) // vars == map[] // FIXME
    eventId := ObjectIdFromURL(w, r, "eventId", 0)
    if eventId==nil { return }
    // -----------------------------------------------
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    body, _ := ioutil.ReadAll(r.Body)
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
    eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
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

    // do we need the last revision? to compare or what?
    lastEventRev := event_lib.EventRevisionModel{}
    err = db.C(storage.C_revision).Find(bson.M{
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
        rawValue, ok := patchMap["timeZone"]
        if ok {
              value, typeOk := rawValue.(string)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'timeZone'",
                )
                return
            }
              eventModel.TimeZone = value
            delete(patchMap, "timeZone")
        }
    }
    {
        rawValue, ok := patchMap["timeZoneEnable"]
        if ok {
              value, typeOk := rawValue.(bool)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'timeZoneEnable'",
                )
                return
            }
              eventModel.TimeZoneEnable = value
            delete(patchMap, "timeZoneEnable")
        }
    }
    {
        rawValue, ok := patchMap["calType"]
        if ok {
              value, typeOk := rawValue.(string)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'calType'",
                )
                return
            }
              eventModel.CalType = value
            delete(patchMap, "calType")
        }
    }
    {
        rawValue, ok := patchMap["summary"]
        if ok {
              value, typeOk := rawValue.(string)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'summary'",
                )
                return
            }
              eventModel.Summary = value
            delete(patchMap, "summary")
        }
    }
    {
        rawValue, ok := patchMap["description"]
        if ok {
              value, typeOk := rawValue.(string)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'description'",
                )
                return
            }
              eventModel.Description = value
            delete(patchMap, "description")
        }
    }
    {
        rawValue, ok := patchMap["icon"]
        if ok {
              value, typeOk := rawValue.(string)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'icon'",
                )
                return
            }
              eventModel.Icon = value
            delete(patchMap, "icon")
        }
    }
    {
        rawValue, ok := patchMap["notifyBefore"]
        if ok {
              // json Unmarshal converts int to float64
              value, typeOk := rawValue.(float64)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'notifyBefore'",
                )
                return
            }
              eventModel.NotifyBefore = int(value)
            delete(patchMap, "notifyBefore")
        }
    }
    {
        rawValue, ok := patchMap["jd"]
        if ok {
              // json Unmarshal converts int to float64
              value, typeOk := rawValue.(float64)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'jd'",
                )
                return
            }
              eventModel.Jd = int(value)
            delete(patchMap, "jd")
        }
    }
    {
        rawValue, ok := patchMap["dayStartSeconds"]
        if ok {
              // json Unmarshal converts int to float64
              value, typeOk := rawValue.(float64)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'dayStartSeconds'",
                )
                return
            }
              eventModel.DayStartSeconds = int(value)
            delete(patchMap, "dayStartSeconds")
        }
    }
    {
        rawValue, ok := patchMap["dayEndSeconds"]
        if ok {
              // json Unmarshal converts int to float64
              value, typeOk := rawValue.(float64)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'dayEndSeconds'",
                )
                return
            }
              eventModel.DayEndSeconds = int(value)
            delete(patchMap, "dayEndSeconds")
        }
    }
    {
        rawValue, ok := patchMap["courseId"]
        if ok {
              // json Unmarshal converts int to float64
              value, typeOk := rawValue.(float64)
            if !typeOk {
                SetHttpError(
                    w,
                    http.StatusBadRequest,
                    "bad type for parameter 'courseId'",
                )
                return
            }
              eventModel.CourseId = int(value)
            delete(patchMap, "courseId")
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

    err = storage.Insert(db, event_lib.EventRevisionModel{
        EventId: *eventId,
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
        err = storage.Insert(db, eventModel)
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
