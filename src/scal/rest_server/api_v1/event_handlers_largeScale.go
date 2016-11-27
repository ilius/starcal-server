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

	"scal"
	"scal/event_lib"
	"scal/storage"
)

func init() {
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
	// functions of following operations are defined in handlers.go
	// because their definition does not depend on event type
	// but their URL still contains eventType for sake of compatibilty
	// so we will have to register their routes for each event type
	// we don't use eventType in these functions
	RegisterRoute(
		"DeleteEvent_largeScale",
		"DELETE",
		"/event/largeScale/{eventId}/",
		authenticator.Wrap(DeleteEvent),
	)
	RegisterRoute(
		"SetEventGroupId_largeScale",
		"PUT",
		"/event/largeScale/{eventId}/groupId/",
		authenticator.Wrap(SetEventGroupId),
	)
	RegisterRoute(
		"GetEventOwner_largeScale",
		"GET",
		"/event/largeScale/{eventId}/owner/",
		authenticator.Wrap(GetEventOwner),
	)
	RegisterRoute(
		"SetEventOwner_largeScale",
		"PUT",
		"/event/largeScale/{eventId}/owner/",
		authenticator.Wrap(SetEventOwner),
	)
	RegisterRoute(
		"GetEventMeta_largeScale",
		"GET",
		"/event/largeScale/{eventId}/meta/",
		authenticator.Wrap(GetEventMeta),
	)
	RegisterRoute(
		"GetEventAccess_largeScale",
		"GET",
		"/event/largeScale/{eventId}/access/",
		authenticator.Wrap(GetEventAccess),
	)
	RegisterRoute(
		"SetEventAccess_largeScale",
		"PUT",
		"/event/largeScale/{eventId}/access/",
		authenticator.Wrap(SetEventAccess),
	)
	RegisterRoute(
		"AppendEventAccess_largeScale",
		"POST",
		"/event/largeScale/{eventId}/access/",
		authenticator.Wrap(AppendEventAccess),
	)
	RegisterRoute(
		"JoinEvent_largeScale",
		"GET",
		"/event/largeScale/{eventId}/join/",
		authenticator.Wrap(JoinEvent),
	)
	RegisterRoute(
		"LeaveEvent_largeScale",
		"GET",
		"/event/largeScale/{eventId}/leave/",
		authenticator.Wrap(LeaveEvent),
	)

}

func AddLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	eventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
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
		scal.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  eventId,
			"funcName": "AddLargeScale",
			"ownerEmail": []interface{}{
				nil,
				email,
			},
		},
		scal.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  eventId,
			"funcName": "AddLargeScale",
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
	err = db.Insert(event_lib.EventRevisionModel{
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
	_, err = event_lib.LoadLargeScaleEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = db.Insert(eventModel)
			if err != nil {
				SetHttpError(w, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			SetHttpErrorInternal(w, err)
			return
		}
	}
	err = db.Insert(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"eventId": eventId.Hex(),
		"sha1":    eventModel.Sha1,
	})
}

func GetLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
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

	eventModel, err := event_lib.LoadLargeScaleEventModel(
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

func UpdateLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	eventModel := event_lib.LargeScaleEventModel{} // DYNAMIC
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
	err = db.Insert(eventRev)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadLargeScaleEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = db.Insert(eventModel)
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
func PatchLargeScale(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
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
	patchMap := scal.M{}
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
	eventModel, err := event_lib.LoadLargeScaleEventModel(
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
		rawValue, ok := patchMap["scale"]
		if ok {
			value, typeOk := rawValue.(int64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'scale'",
				)
				return
			}
			eventModel.Scale = value
			delete(patchMap, "scale")
		}
	}
	{
		rawValue, ok := patchMap["start"]
		if ok {
			value, typeOk := rawValue.(int64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'start'",
				)
				return
			}
			eventModel.Start = value
			delete(patchMap, "start")
		}
	}
	{
		rawValue, ok := patchMap["end"]
		if ok {
			value, typeOk := rawValue.(int64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'end'",
				)
				return
			}
			eventModel.End = value
			delete(patchMap, "end")
		}
	}
	{
		rawValue, ok := patchMap["durationEnable"]
		if ok {
			value, typeOk := rawValue.(bool)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'durationEnable'",
				)
				return
			}
			eventModel.DurationEnable = value
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

	err = db.Insert(event_lib.EventRevisionModel{
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
	_, err = event_lib.LoadLargeScaleEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = db.Insert(eventModel)
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
