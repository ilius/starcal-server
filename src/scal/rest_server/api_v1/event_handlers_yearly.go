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
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"

	//"github.com/gorilla/mux"

	"scal"
	"scal/event_lib"
	"scal/settings"
	"scal/storage"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/yearly",
		Map: RouteMap{
			"AddYearly": {
				"POST",
				"",
				authWrap(AddYearly),
			},
			"GetYearly": {
				"GET",
				"{eventId}",
				authWrap(GetYearly),
			},
			"UpdateYearly": {
				"PUT",
				"{eventId}",
				authWrap(UpdateYearly),
			},
			"PatchYearly": {
				"PATCH",
				"{eventId}",
				authWrap(PatchYearly),
			},
			"MergeYearly": {
				"POST",
				"{eventId}/merge",
				authWrap(MergeYearly),
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibilty
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_yearly": {
				"DELETE",
				"{eventId}",
				authWrap(DeleteEvent),
			},
			"SetEventGroupId_yearly": {
				"PUT",
				"{eventId}/group",
				authWrap(SetEventGroupId),
			},
			"GetEventOwner_yearly": {
				"GET",
				"{eventId}/owner",
				authWrap(GetEventOwner),
			},
			"SetEventOwner_yearly": {
				"PUT",
				"{eventId}/owner",
				authWrap(SetEventOwner),
			},
			"GetEventMeta_yearly": {
				"GET",
				"{eventId}/meta",
				authWrap(GetEventMeta),
			},
			"GetEventAccess_yearly": {
				"GET",
				"{eventId}/access",
				authWrap(GetEventAccess),
			},
			"SetEventAccess_yearly": {
				"PUT",
				"{eventId}/access",
				authWrap(SetEventAccess),
			},
			"AppendEventAccess_yearly": {
				"POST",
				"{eventId}/access",
				authWrap(AppendEventAccess),
			},
			"JoinEvent_yearly": {
				"GET",
				"{eventId}/join",
				authWrap(JoinEvent),
			},
			"LeaveEvent_yearly": {
				"GET",
				"{eventId}/leave",
				authWrap(LeaveEvent),
			},
			"InviteToEvent_yearly": {
				"POST",
				"{eventId}/invite",
				authWrap(InviteToEvent),
			},
		},
	})
}

func AddYearly(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.YearlyEventModel{} // DYNAMIC
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
		FuncName: "AddYearly",
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
	_, err = event_lib.LoadYearlyEventModel(
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
		"timeZone":        now,
		"timeZoneEnable":  now,
		"calType":         now,
		"summary":         now,
		"description":     now,
		"icon":            now,
		"month":           now,
		"day":             now,
		"startYear":       now,
		"startYearEnable": now,
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

func GetYearly(w http.ResponseWriter, r *http.Request) {
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
		if eventMeta.EventType != "yearly" {
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

	eventModel, err := event_lib.LoadYearlyEventModel(
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
	eventModel.DummyType = eventMeta.EventType  // not "yearly"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	json.NewEncoder(w).Encode(eventModel)
}

func UpdateYearly(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.YearlyEventModel{} // DYNAMIC
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
	lastEventModel, err := event_lib.LoadYearlyEventModel(
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
	_, err = event_lib.LoadYearlyEventModel(
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
	// PARAM="month", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.Month,
		lastEventModel.Month,
	) {
		eventMeta.FieldsMtime["month"] = now
	}
	// PARAM="day", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.Day,
		lastEventModel.Day,
	) {
		eventMeta.FieldsMtime["day"] = now
	}
	// PARAM="startYear", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.StartYear,
		lastEventModel.StartYear,
	) {
		eventMeta.FieldsMtime["startYear"] = now
	}
	// PARAM="startYearEnable", PARAM_TYPE="bool"
	if !reflect.DeepEqual(
		eventModel.StartYearEnable,
		lastEventModel.StartYearEnable,
	) {
		eventMeta.FieldsMtime["startYearEnable"] = now
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
func PatchYearly(w http.ResponseWriter, r *http.Request) {
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
	eventModel, err := event_lib.LoadYearlyEventModel(
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
		rawValue, ok := patchMap["month"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'month'",
				)
				return
			}
			eventModel.Month = int(value)
			delete(patchMap, "month")
			eventMeta.FieldsMtime["month"] = now
		}
	}
	{
		rawValue, ok := patchMap["day"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'day'",
				)
				return
			}
			eventModel.Day = int(value)
			delete(patchMap, "day")
			eventMeta.FieldsMtime["day"] = now
		}
	}
	{
		rawValue, ok := patchMap["startYear"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'startYear'",
				)
				return
			}
			eventModel.StartYear = int(value)
			delete(patchMap, "startYear")
			eventMeta.FieldsMtime["startYear"] = now
		}
	}
	{
		rawValue, ok := patchMap["startYearEnable"]
		if ok {
			value, typeOk := rawValue.(bool)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'startYearEnable'",
				)
				return
			}
			eventModel.StartYearEnable = value
			delete(patchMap, "startYearEnable")
			eventMeta.FieldsMtime["startYearEnable"] = now
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
	_, err = event_lib.LoadYearlyEventModel(
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

func MergeYearly(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }
	// -----------------------------------------------
	//vars := mux.Vars(&r.Request) // vars == map[] // FIXME
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}
	// -----------------------------------------------
	inputStruct := struct {
		Event event_lib.YearlyEventModel `json:"event"` // DYNAMIC

		LastMergeSha1 string               `json:"lastMergeSha1"`
		FieldsMtime   map[string]time.Time `json:"fieldsMtime"`
	}{}

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &inputStruct)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	// if inputStruct.Event.DummyType == "" {
	// 	SetHttpError(w, http.StatusBadRequest, "missing 'eventType'")
	// 	return
	// }
	if inputStruct.Event.Id == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
		return
	}
	// FIXME: LastMergeSha1 can be empty?
	if inputStruct.LastMergeSha1 == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'lastMergeSha1'")
		return
	}
	inputEventModel := &inputStruct.Event
	if inputEventModel.Id.Hex() != eventId.Hex() {
		SetHttpError(w, http.StatusBadRequest, "mismatch 'event.id'")
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
	if eventMeta.OwnerEmail != email {
		SetHttpError(w, http.StatusForbidden, "you don't own this event")
		return
	}

	lastRevModel, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	parentEventModel, err := event_lib.LoadYearlyEventModel(db, inputStruct.LastMergeSha1)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "invalid lastMergeSha1: revision not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	lastEventModel, err := event_lib.LoadYearlyEventModel(db, lastRevModel.Sha1)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	fmt.Println(parentEventModel)
	fmt.Println(lastEventModel)

	// performing a 3-way merge
	// C <== parentEventModel	<== The common ancestor (last merge for this client)
	// A <== inputEventModel	<== The input (client's latest) data
	// B <== lastEventModel		<== The current (server's latest) data
	now := time.Now()
	func() {
		// PARAM="timeZone", PARAM_TYPE="string"
		inputValue := inputEventModel.TimeZone
		lastValue := lastEventModel.TimeZone
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.TimeZone
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.TimeZone = inputValue
			eventMeta.FieldsMtime["timeZone"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["timeZone"].After(eventMeta.FieldsMtime["timeZone"]) {
			lastEventModel.TimeZone = inputValue
			eventMeta.FieldsMtime["timeZone"] = now
			return
		}
	}()
	func() {
		// PARAM="timeZoneEnable", PARAM_TYPE="bool"
		inputValue := inputEventModel.TimeZoneEnable
		lastValue := lastEventModel.TimeZoneEnable
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.TimeZoneEnable
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.TimeZoneEnable = inputValue
			eventMeta.FieldsMtime["timeZoneEnable"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["timeZoneEnable"].After(eventMeta.FieldsMtime["timeZoneEnable"]) {
			lastEventModel.TimeZoneEnable = inputValue
			eventMeta.FieldsMtime["timeZoneEnable"] = now
			return
		}
	}()
	func() {
		// PARAM="calType", PARAM_TYPE="string"
		inputValue := inputEventModel.CalType
		lastValue := lastEventModel.CalType
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.CalType
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.CalType = inputValue
			eventMeta.FieldsMtime["calType"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["calType"].After(eventMeta.FieldsMtime["calType"]) {
			lastEventModel.CalType = inputValue
			eventMeta.FieldsMtime["calType"] = now
			return
		}
	}()
	func() {
		// PARAM="summary", PARAM_TYPE="string"
		inputValue := inputEventModel.Summary
		lastValue := lastEventModel.Summary
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.Summary
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.Summary = inputValue
			eventMeta.FieldsMtime["summary"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["summary"].After(eventMeta.FieldsMtime["summary"]) {
			lastEventModel.Summary = inputValue
			eventMeta.FieldsMtime["summary"] = now
			return
		}
	}()
	func() {
		// PARAM="description", PARAM_TYPE="string"
		inputValue := inputEventModel.Description
		lastValue := lastEventModel.Description
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.Description
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.Description = inputValue
			eventMeta.FieldsMtime["description"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["description"].After(eventMeta.FieldsMtime["description"]) {
			lastEventModel.Description = inputValue
			eventMeta.FieldsMtime["description"] = now
			return
		}
	}()
	func() {
		// PARAM="icon", PARAM_TYPE="string"
		inputValue := inputEventModel.Icon
		lastValue := lastEventModel.Icon
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.Icon
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.Icon = inputValue
			eventMeta.FieldsMtime["icon"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["icon"].After(eventMeta.FieldsMtime["icon"]) {
			lastEventModel.Icon = inputValue
			eventMeta.FieldsMtime["icon"] = now
			return
		}
	}()
	func() {
		// PARAM="month", PARAM_TYPE="int"
		inputValue := inputEventModel.Month
		lastValue := lastEventModel.Month
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.Month
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.Month = inputValue
			eventMeta.FieldsMtime["month"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["month"].After(eventMeta.FieldsMtime["month"]) {
			lastEventModel.Month = inputValue
			eventMeta.FieldsMtime["month"] = now
			return
		}
	}()
	func() {
		// PARAM="day", PARAM_TYPE="int"
		inputValue := inputEventModel.Day
		lastValue := lastEventModel.Day
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.Day
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.Day = inputValue
			eventMeta.FieldsMtime["day"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["day"].After(eventMeta.FieldsMtime["day"]) {
			lastEventModel.Day = inputValue
			eventMeta.FieldsMtime["day"] = now
			return
		}
	}()
	func() {
		// PARAM="startYear", PARAM_TYPE="int"
		inputValue := inputEventModel.StartYear
		lastValue := lastEventModel.StartYear
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.StartYear
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.StartYear = inputValue
			eventMeta.FieldsMtime["startYear"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["startYear"].After(eventMeta.FieldsMtime["startYear"]) {
			lastEventModel.StartYear = inputValue
			eventMeta.FieldsMtime["startYear"] = now
			return
		}
	}()
	func() {
		// PARAM="startYearEnable", PARAM_TYPE="bool"
		inputValue := inputEventModel.StartYearEnable
		lastValue := lastEventModel.StartYearEnable
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.StartYearEnable
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.StartYearEnable = inputValue
			eventMeta.FieldsMtime["startYearEnable"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["startYearEnable"].After(eventMeta.FieldsMtime["startYearEnable"]) {
			lastEventModel.StartYearEnable = inputValue
			eventMeta.FieldsMtime["startYearEnable"] = now
			return
		}
	}()
	// err = db.Update(eventMeta) // just for FieldsMtime, is it safe? FIXME
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }
}
