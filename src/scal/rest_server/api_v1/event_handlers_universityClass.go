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
		Base: "event/universityClass",
		Map: RouteMap{
			"AddUniversityClass": {
				"POST",
				"",
				authWrap(AddUniversityClass),
			},
			"GetUniversityClass": {
				"GET",
				"{eventId}",
				authWrap(GetUniversityClass),
			},
			"UpdateUniversityClass": {
				"PUT",
				"{eventId}",
				authWrap(UpdateUniversityClass),
			},
			"PatchUniversityClass": {
				"PATCH",
				"{eventId}",
				authWrap(PatchUniversityClass),
			},
			"MergeUniversityClass": {
				"POST",
				"{eventId}/merge",
				authWrap(MergeUniversityClass),
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibilty
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_universityClass": {
				"DELETE",
				"{eventId}",
				authWrap(DeleteEvent),
			},
			"SetEventGroupId_universityClass": {
				"PUT",
				"{eventId}/group",
				authWrap(SetEventGroupId),
			},
			"GetEventOwner_universityClass": {
				"GET",
				"{eventId}/owner",
				authWrap(GetEventOwner),
			},
			"SetEventOwner_universityClass": {
				"PUT",
				"{eventId}/owner",
				authWrap(SetEventOwner),
			},
			"GetEventMeta_universityClass": {
				"GET",
				"{eventId}/meta",
				authWrap(GetEventMeta),
			},
			"GetEventAccess_universityClass": {
				"GET",
				"{eventId}/access",
				authWrap(GetEventAccess),
			},
			"SetEventAccess_universityClass": {
				"PUT",
				"{eventId}/access",
				authWrap(SetEventAccess),
			},
			"AppendEventAccess_universityClass": {
				"POST",
				"{eventId}/access",
				authWrap(AppendEventAccess),
			},
			"JoinEvent_universityClass": {
				"GET",
				"{eventId}/join",
				authWrap(JoinEvent),
			},
			"LeaveEvent_universityClass": {
				"GET",
				"{eventId}/leave",
				authWrap(LeaveEvent),
			},
			"InviteToEvent_universityClass": {
				"POST",
				"{eventId}/invite",
				authWrap(InviteToEvent),
			},
		},
	})
}

func AddUniversityClass(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.UniversityClassEventModel{} // DYNAMIC
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
		FuncName: "AddUniversityClass",
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
	_, err = event_lib.LoadUniversityClassEventModel(
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
		"weekNumMode":     now,
		"weekDayList":     now,
		"dayStartSeconds": now,
		"dayEndSeconds":   now,
		"courseId":        now,
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

func GetUniversityClass(w http.ResponseWriter, r *http.Request) {
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
		if eventMeta.EventType != "universityClass" {
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

	eventModel, err := event_lib.LoadUniversityClassEventModel(
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
	eventModel.DummyType = eventMeta.EventType  // not "universityClass"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	json.NewEncoder(w).Encode(eventModel)
}

func UpdateUniversityClass(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.UniversityClassEventModel{} // DYNAMIC
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
	lastEventModel, err := event_lib.LoadUniversityClassEventModel(
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
	_, err = event_lib.LoadUniversityClassEventModel(
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
	// PARAM="weekNumMode", PARAM_TYPE="string"
	if !reflect.DeepEqual(
		eventModel.WeekNumMode,
		lastEventModel.WeekNumMode,
	) {
		eventMeta.FieldsMtime["weekNumMode"] = now
	}
	// PARAM="weekDayList", PARAM_TYPE="[]int"
	if !reflect.DeepEqual(
		eventModel.WeekDayList,
		lastEventModel.WeekDayList,
	) {
		eventMeta.FieldsMtime["weekDayList"] = now
	}
	// PARAM="dayStartSeconds", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.DayStartSeconds,
		lastEventModel.DayStartSeconds,
	) {
		eventMeta.FieldsMtime["dayStartSeconds"] = now
	}
	// PARAM="dayEndSeconds", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.DayEndSeconds,
		lastEventModel.DayEndSeconds,
	) {
		eventMeta.FieldsMtime["dayEndSeconds"] = now
	}
	// PARAM="courseId", PARAM_TYPE="int"
	if !reflect.DeepEqual(
		eventModel.CourseId,
		lastEventModel.CourseId,
	) {
		eventMeta.FieldsMtime["courseId"] = now
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
func PatchUniversityClass(w http.ResponseWriter, r *http.Request) {
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
	eventModel, err := event_lib.LoadUniversityClassEventModel(
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
		rawValue, ok := patchMap["weekNumMode"]
		if ok {
			value, typeOk := rawValue.(string)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'weekNumMode'",
				)
				return
			}
			eventModel.WeekNumMode = value
			delete(patchMap, "weekNumMode")
			eventMeta.FieldsMtime["weekNumMode"] = now
		}
	}
	{
		rawValue, ok := patchMap["weekDayList"]
		if ok {
			value, typeOk := rawValue.([]int)
			if !typeOk {
				SetHttpError(
					w,
					http.StatusBadRequest,
					"bad type for parameter 'weekDayList'",
				)
				return
			}
			eventModel.WeekDayList = value
			delete(patchMap, "weekDayList")
			eventMeta.FieldsMtime["weekDayList"] = now
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
			eventMeta.FieldsMtime["dayStartSeconds"] = now
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
			eventMeta.FieldsMtime["dayEndSeconds"] = now
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
			eventMeta.FieldsMtime["courseId"] = now
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
	_, err = event_lib.LoadUniversityClassEventModel(
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

func MergeUniversityClass(w http.ResponseWriter, r *http.Request) {
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
		Event event_lib.UniversityClassEventModel `json:"event"` // DYNAMIC

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
	parentEventModel, err := event_lib.LoadUniversityClassEventModel(db, inputStruct.LastMergeSha1)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "invalid lastMergeSha1: revision not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	lastEventModel, err := event_lib.LoadUniversityClassEventModel(db, lastRevModel.Sha1)
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
		// PARAM="weekNumMode", PARAM_TYPE="string"
		inputValue := inputEventModel.WeekNumMode
		lastValue := lastEventModel.WeekNumMode
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.WeekNumMode
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.WeekNumMode = inputValue
			eventMeta.FieldsMtime["weekNumMode"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["weekNumMode"].After(eventMeta.FieldsMtime["weekNumMode"]) {
			lastEventModel.WeekNumMode = inputValue
			eventMeta.FieldsMtime["weekNumMode"] = now
			return
		}
	}()
	func() {
		// PARAM="weekDayList", PARAM_TYPE="[]int"
		inputValue := inputEventModel.WeekDayList
		lastValue := lastEventModel.WeekDayList
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.WeekDayList
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.WeekDayList = inputValue
			eventMeta.FieldsMtime["weekDayList"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["weekDayList"].After(eventMeta.FieldsMtime["weekDayList"]) {
			lastEventModel.WeekDayList = inputValue
			eventMeta.FieldsMtime["weekDayList"] = now
			return
		}
	}()
	func() {
		// PARAM="dayStartSeconds", PARAM_TYPE="int"
		inputValue := inputEventModel.DayStartSeconds
		lastValue := lastEventModel.DayStartSeconds
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.DayStartSeconds
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.DayStartSeconds = inputValue
			eventMeta.FieldsMtime["dayStartSeconds"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["dayStartSeconds"].After(eventMeta.FieldsMtime["dayStartSeconds"]) {
			lastEventModel.DayStartSeconds = inputValue
			eventMeta.FieldsMtime["dayStartSeconds"] = now
			return
		}
	}()
	func() {
		// PARAM="dayEndSeconds", PARAM_TYPE="int"
		inputValue := inputEventModel.DayEndSeconds
		lastValue := lastEventModel.DayEndSeconds
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.DayEndSeconds
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.DayEndSeconds = inputValue
			eventMeta.FieldsMtime["dayEndSeconds"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["dayEndSeconds"].After(eventMeta.FieldsMtime["dayEndSeconds"]) {
			lastEventModel.DayEndSeconds = inputValue
			eventMeta.FieldsMtime["dayEndSeconds"] = now
			return
		}
	}()
	func() {
		// PARAM="courseId", PARAM_TYPE="int"
		inputValue := inputEventModel.CourseId
		lastValue := lastEventModel.CourseId
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.CourseId
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.CourseId = inputValue
			eventMeta.FieldsMtime["courseId"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["courseId"].After(eventMeta.FieldsMtime["courseId"]) {
			lastEventModel.CourseId = inputValue
			eventMeta.FieldsMtime["courseId"] = now
			return
		}
	}()
	// err = db.Update(eventMeta) // just for FieldsMtime, is it safe? FIXME
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }
}
