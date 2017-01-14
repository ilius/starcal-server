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
	"reflect"

	"gopkg.in/mgo.v2/bson"
	//"github.com/gorilla/mux"

	"scal"
	"scal/event_lib"
	"scal/settings"
	"scal/storage"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/allDayTask",
		Map: RouteMap{
			"AddAllDayTask": {
				"POST",
				"",
				authWrap(AddAllDayTask),
			},
			"GetAllDayTask": {
				"GET",
				"{eventId}",
				authWrap(GetAllDayTask),
			},
			"UpdateAllDayTask": {
				"PUT",
				"{eventId}",
				authWrap(UpdateAllDayTask),
			},
			"PatchAllDayTask": {
				"PATCH",
				"{eventId}",
				authWrap(PatchAllDayTask),
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibilty
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_allDayTask": {
				"DELETE",
				"{eventId}",
				authWrap(DeleteEvent),
			},
			"SetEventGroupId_allDayTask": {
				"PUT",
				"{eventId}/group",
				authWrap(SetEventGroupId),
			},
			"GetEventOwner_allDayTask": {
				"GET",
				"{eventId}/owner",
				authWrap(GetEventOwner),
			},
			"SetEventOwner_allDayTask": {
				"PUT",
				"{eventId}/owner",
				authWrap(SetEventOwner),
			},
			"GetEventMeta_allDayTask": {
				"GET",
				"{eventId}/meta",
				authWrap(GetEventMeta),
			},
			"GetEventAccess_allDayTask": {
				"GET",
				"{eventId}/access",
				authWrap(GetEventAccess),
			},
			"SetEventAccess_allDayTask": {
				"PUT",
				"{eventId}/access",
				authWrap(SetEventAccess),
			},
			"AppendEventAccess_allDayTask": {
				"POST",
				"{eventId}/access",
				authWrap(AppendEventAccess),
			},
			"JoinEvent_allDayTask": {
				"GET",
				"{eventId}/join",
				authWrap(JoinEvent),
			},
			"LeaveEvent_allDayTask": {
				"GET",
				"{eventId}/leave",
				authWrap(LeaveEvent),
			},
			"InviteToEvent_allDayTask": {
				"POST",
				"{eventId}/invite",
				authWrap(InviteToEvent),
			},
		},
	})
}

func AddAllDayTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
	// -----------------------------------------------
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
	_, err = eventModel.GetEvent() // for validation
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
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  eventId,
		FuncName: "AddAllDayTask",
		OwnerEmail: &[2]*string{
			nil,
			&email,
		},
	})
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
	_, err = event_lib.LoadAllDayTaskEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
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

	eventMeta.FieldsMtime = map[string]time.Time{
		"timeZone":       now,
		"timeZoneEnable": now,
		"calType":        now,
		"summary":        now,
		"description":    now,
		"icon":           now,
		"startJd":        now,
		"endJd":          now,
		"durationEnable": now,
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

func GetAllDayTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
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
	if !settings.ALLOW_MISMATCH_EVENT_TYPE {
		if eventMeta.EventType != "allDayTask" {
			SetHttpError(
				w,
				http.StatusBadRequest,
				fmt.Sprintf(
					"mismatch {eventType}, must be '%s'",
					eventMeta.EventType,
				),
			)
			return
		}
	}

	eventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

	eventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusInternalServerError, "event snapshot not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

	eventModel.Id = *eventId
	eventModel.DummyType = eventMeta.EventType  // not "allDayTask"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	json.NewEncoder(w).Encode(eventModel)
}

func UpdateAllDayTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
	// -----------------------------------------------
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
	// -----------------------------------------------
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &eventModel)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "invalid ObjectId in JSON") {
			msg = "invalid 'eventId'"
		}
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	_, err = eventModel.GetEvent() // for validation
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
		if db.IsNotFound(err) {
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

	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	lastEventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusInternalServerError, "event snapshot not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

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

	now := time.Now()

	eventRev := event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      now,
	}
	err = db.Insert(eventRev)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadAllDayTaskEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
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
	// PARAM="timeZone", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.TimeZone,
		lastEventModel.TimeZone,
	) {
		eventMeta.FieldsMtime["timeZone"] = now
	}
	// PARAM="timeZoneEnable", PARAM_TYPE="bool"
	if !reflect.DeepEqual(
		eventModel.TimeZoneEnable,
		lastEventModel.TimeZoneEnable,
	) {
		eventMeta.FieldsMtime["timeZoneEnable"] = now
	}
	// PARAM="calType", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.CalType,
		lastEventModel.CalType,
	) {
		eventMeta.FieldsMtime["calType"] = now
	}
	// PARAM="summary", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.Summary,
		lastEventModel.Summary,
	) {
		eventMeta.FieldsMtime["summary"] = now
	}
	// PARAM="description", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.Description,
		lastEventModel.Description,
	) {
		eventMeta.FieldsMtime["description"] = now
	}
	// PARAM="icon", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.Icon,
		lastEventModel.Icon,
	) {
		eventMeta.FieldsMtime["icon"] = now
	}
	// PARAM="startJd", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.StartJd,
		lastEventModel.StartJd,
	) {
		eventMeta.FieldsMtime["startJd"] = now
	}
	// PARAM="endJd", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.EndJd,
		lastEventModel.EndJd,
	) {
		eventMeta.FieldsMtime["endJd"] = now
	}
	// PARAM="durationEnable", PARAM_TYPE="bool"
	if !reflect.DeepEqual(
		eventModel.DurationEnable,
		lastEventModel.DurationEnable,
	) {
		eventMeta.FieldsMtime["durationEnable"] = now
	}
	err = db.Update(eventMeta) // just for FieldsMtime, is it safe? FIXME
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"eventId": eventId.Hex(),
		"sha1":    eventRev.Sha1,
	})
}
func PatchAllDayTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
	// -----------------------------------------------
	body, _ := ioutil.ReadAll(r.Body)
	patchMap := scal.M{}
	err := json.Unmarshal(body, &patchMap)
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
		if db.IsNotFound(err) {
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
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	eventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	now := time.Now()
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
			eventMeta.FieldsMtime["timeZone"] = now
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
			eventMeta.FieldsMtime["timeZoneEnable"] = now
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
			eventMeta.FieldsMtime["calType"] = now
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
			eventMeta.FieldsMtime["summary"] = now
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
			eventMeta.FieldsMtime["description"] = now
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
			eventMeta.FieldsMtime["icon"] = now
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
			eventMeta.FieldsMtime["startJd"] = now
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
			eventMeta.FieldsMtime["endJd"] = now
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
			eventMeta.FieldsMtime["durationEnable"] = now
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
	_, err = eventModel.GetEvent() // for validation
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
		Time:      now,
	})
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.LoadAllDayTaskEventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
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
	err = db.Update(eventMeta) // just for FieldsMtime, is it safe? FIXME
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"eventId": eventId.Hex(),
		"sha1":    eventModel.Sha1,
	})
}
