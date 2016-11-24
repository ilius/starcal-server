// if this is a *.go file, don't modify it, it's auto-generated
// from a Django template file named `*.got` inside "templates" directory
package api_v1

import (
	"fmt"
	"strings"
	"time"
	//"log"
	"crypto/sha1"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"github.com/gorilla/mux"

	"scal-lib/go-http-auth"

	"scal/event_lib"
	"scal/storage"
)

func init() {
	RegisterRoute(
		"AddLifeTime",
		"POST",
		"/event/lifeTime/",
		authenticator.Wrap(AddLifeTime),
	)
	RegisterRoute(
		"GetLifeTime",
		"GET",
		"/event/lifeTime/{eventId}/",
		authenticator.Wrap(GetLifeTime),
	)
	RegisterRoute(
		"UpdateLifeTime",
		"PUT",
		"/event/lifeTime/{eventId}/",
		authenticator.Wrap(UpdateLifeTime),
	)
	RegisterRoute(
		"PatchLifeTime",
		"PATCH",
		"/event/lifeTime/{eventId}/",
		authenticator.Wrap(PatchLifeTime),
	)
	// functions of following operations are defined in handlers.go
	// because their definition does not depend on event type
	// but their URL still contains eventType for sake of compatibilty
	// so we will have to register their routes for each event type
	// we don't use eventType in these functions
	RegisterRoute(
		"DeleteEvent_lifeTime",
		"DELETE",
		"/event/lifeTime/{eventId}/",
		authenticator.Wrap(DeleteEvent),
	)
	RegisterRoute(
		"SetEventGroupId_lifeTime",
		"PUT",
		"/event/lifeTime/{eventId}/groupId/",
		authenticator.Wrap(SetEventGroupId),
	)
	RegisterRoute(
		"GetEventOwner_lifeTime",
		"GET",
		"/event/lifeTime/{eventId}/owner/",
		authenticator.Wrap(GetEventOwner),
	)
	RegisterRoute(
		"SetEventOwner_lifeTime",
		"PUT",
		"/event/lifeTime/{eventId}/owner/",
		authenticator.Wrap(SetEventOwner),
	)
	RegisterRoute(
		"GetEventMeta_lifeTime",
		"GET",
		"/event/lifeTime/{eventId}/meta/",
		authenticator.Wrap(GetEventMeta),
	)
	RegisterRoute(
		"GetEventAccess_lifeTime",
		"GET",
		"/event/lifeTime/{eventId}/access/",
		authenticator.Wrap(GetEventAccess),
	)
	RegisterRoute(
		"SetEventAccess_lifeTime",
		"PUT",
		"/event/lifeTime/{eventId}/access/",
		authenticator.Wrap(SetEventAccess),
	)
	RegisterRoute(
		"AppendEventAccess_lifeTime",
		"POST",
		"/event/lifeTime/{eventId}/access/",
		authenticator.Wrap(AppendEventAccess),
	)
	RegisterRoute(
		"JoinEvent_lifeTime",
		"GET",
		"/event/lifeTime/{eventId}/join/",
		authenticator.Wrap(JoinEvent),
	)
	RegisterRoute(
		"LeaveEvent_lifeTime",
		"GET",
		"/event/lifeTime/{eventId}/leave/",
		authenticator.Wrap(LeaveEvent),
	)

}

func AddLifeTime(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	eventModel := event_lib.LifeTimeEventModel{} // DYNAMIC
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
		groupModel, err, internalErr := event_lib.LoadGroupModelByIdHex(
			"groupId",
			db,
			eventModel.GroupId,
		)
		if err != nil {
			if internalErr {
				SetHttpErrorInternal(w, err)
			} else {
				SetHttpError(w, http.StatusBadRequest, err.Error())
			}
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
		EventId:      eventId,
		EventType:    eventModel.Type(),
		CreationTime: time.Now(),
		OwnerEmail:   email,
		GroupId:      groupId,
		//AccessEmails: []string{}
	}
	now := time.Now()
	err = db.C(storage.C_eventMetaChangeLog).Insert(
		bson.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  eventId,
			"funcName": "AddLifeTime",
			"ownerEmail": []interface{}{
				nil,
				email,
			},
		},
		bson.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  eventId,
			"funcName": "AddLifeTime",
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
		EventId:   eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      time.Now(),
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadLifeTimeEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.Insert(db, eventModel)
			if err != nil {
				SetHttpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			SetHttpErrorInternal(w, err)
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
		"sha1":    eventModel.Sha1,
	})
}

func GetLifeTime(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	// -----------------------------------------------
	email := r.Username
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
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

	eventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

	eventModel, err := event_lib.LoadLifeTimeEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusInternalServerError, "event snapshot not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

	eventModel.Id = *eventId
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	json.NewEncoder(w).Encode(eventModel)
}

func UpdateLifeTime(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	eventModel := event_lib.LifeTimeEventModel{} // DYNAMIC
	// -----------------------------------------------
	email := r.Username
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
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
		lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
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
	if len(eventModel.Meta) != 0 {
		SetHttpError(w, http.StatusBadRequest, "'meta' must not be present in JSON")
		return
	}
	eventModel.Sha1 = ""
	jsonByte, _ := json.Marshal(eventModel)
	eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))

	eventRev := event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      time.Now(),
	}
	err = storage.Insert(db, eventRev)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadLifeTimeEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.Insert(db, eventModel)
			if err != nil {
				SetHttpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			SetHttpErrorInternal(w, err)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"eventId": eventId.Hex(),
		"sha1":    eventRev.Sha1,
	})
}
func PatchLifeTime(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	// -----------------------------------------------
	email := r.Username
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
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
	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	eventModel, err := event_lib.LoadLifeTimeEventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
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
		rawValue, ok := patchMap["startJd"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'startJd'",
				)
				return
			}
			eventModel.StartJd = int(value)
			delete(patchMap, "startJd")
		}
	}
	{
		rawValue, ok := patchMap["endJd"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'endJd'",
				)
				return
			}
			eventModel.EndJd = int(value)
			delete(patchMap, "endJd")
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
		EventId:   *eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      time.Now(),
	})
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadLifeTimeEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.Insert(db, eventModel)
			if err != nil {
				SetHttpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			SetHttpErrorInternal(w, err)
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]string{
		"eventId": eventId.Hex(),
		"sha1":    eventModel.Sha1,
	})
}
