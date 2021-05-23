//go:generate $GOPATH/pkg/scal/rest_server/gen -event-type allDayTask
// if this is a *.go file, don't modify it, it's auto-generated
// from a Go template file named `*.go.tpl` inside "templates" directory
package api_v1

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ilius/starcal-server/pkg/scal/event_lib"
	"github.com/ilius/starcal-server/pkg/scal/settings"
	"github.com/ilius/starcal-server/pkg/scal/storage"

	"github.com/ilius/mgo/bson"
	. "github.com/ilius/ripo"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/allDayTask",
		Map: RouteMap{
			"AddAllDayTask": {
				Method:  "POST",
				Pattern: "",
				Handler: AddAllDayTask,
			},
			"GetAllDayTask": {
				Method:  "GET",
				Pattern: ":eventId",
				Handler: GetAllDayTask,
			},
			"UpdateAllDayTask": {
				Method:  "PUT",
				Pattern: ":eventId",
				Handler: UpdateAllDayTask,
			},
			"PatchAllDayTask": {
				Method:  "PATCH",
				Pattern: ":eventId",
				Handler: PatchAllDayTask,
			},
			"MergeAllDayTask": {
				Method:  "POST",
				Pattern: ":eventId/merge",
				Handler: MergeAllDayTask,
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibility
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_allDayTask": {
				Method:  "DELETE",
				Pattern: ":eventId",
				Handler: DeleteEvent,
			},
			"SetEventGroupId_allDayTask": {
				Method:  "PUT",
				Pattern: ":eventId/group",
				Handler: SetEventGroupId,
			},
			"GetEventOwner_allDayTask": {
				Method:  "GET",
				Pattern: ":eventId/owner",
				Handler: GetEventOwner,
			},
			"SetEventOwner_allDayTask": {
				Method:  "PUT",
				Pattern: ":eventId/owner",
				Handler: SetEventOwner,
			},
			"GetEventMeta_allDayTask": {
				Method:  "GET",
				Pattern: ":eventId/meta",
				Handler: GetEventMeta,
			},
			"GetEventAccess_allDayTask": {
				Method:  "GET",
				Pattern: ":eventId/access",
				Handler: GetEventAccess,
			},
			"SetEventAccess_allDayTask": {
				Method:  "PUT",
				Pattern: ":eventId/access",
				Handler: SetEventAccess,
			},
			"AppendEventAccess_allDayTask": {
				Method:  "POST",
				Pattern: ":eventId/access",
				Handler: AppendEventAccess,
			},
			"JoinEvent_allDayTask": {
				Method:  "GET",
				Pattern: ":eventId/join",
				Handler: JoinEvent,
			},
			"LeaveEvent_allDayTask": {
				Method:  "GET",
				Pattern: ":eventId/leave",
				Handler: LeaveEvent,
			},
			"InviteToEvent_allDayTask": {
				Method:  "POST",
				Pattern: ":eventId/invite",
				Handler: InviteToEvent,
			},
		},
	})
}

func AddAllDayTask(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	err = req.BodyTo(&eventModel)
	if err != nil {
		return nil, err
	}
	_, err = eventModel.GetEvent() // for validation
	if err != nil {
		return nil, NewError(InvalidArgument, err.Error(), err) // FIXME: correct msg?
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	if eventModel.Id != "" {
		return nil, NewError(InvalidArgument, "you can't specify 'eventId'", nil)
	}
	eventModel.Sha1 = ""
	jsonByte, _ := json.Marshal(eventModel)
	eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))
	eventId := bson.NewObjectId().Hex()
	eventModel.Id = eventId
	groupId := userModel.DefaultGroupId
	if eventModel.GroupId != "" {
		if !bson.IsObjectIdHex(eventModel.GroupId) {
			return nil, NewError(InvalidArgument, "invalid 'groupId'", nil)
		}
		groupModel, err := event_lib.LoadGroupModelByIdHex(
			"groupId",
			db,
			eventModel.GroupId,
		)
		if err != nil {
			return nil, err
		}
		if groupModel.OwnerEmail != email {
			return nil, ForbiddenError("you don't have write access this event group", nil)
		}
		groupId = &groupModel.Id
	}
	eventMeta := event_lib.EventMetaModel{
		EventId:      eventId,
		EventType:    eventModel.Type(),
		CreationTime: time.Now(),
		OwnerEmail:   email,
		GroupId:      groupId,
		// AccessEmails: []string{}
	}
	now := time.Now()
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:  now,
		Email: email,

		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,

		EventId:  eventId,
		FuncName: "AddAllDayTask",
		OwnerEmail: &[2]*string{
			nil,
			&email,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Insert(event_lib.EventRevisionModel{
		EventId:   eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      time.Now(),
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
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
				return nil, NewError(Internal, "", err)
			}
		} else {
			return nil, NewError(Internal, "", err)
		}
	}

	eventMeta.FieldsMtime = map[string]time.Time{
		"timeZone":             now,
		"timeZoneEnable":       now,
		"calType":              now,
		"summary":              now,
		"description":          now,
		"icon":                 now,
		"summaryEncrypted":     now,
		"descriptionEncrypted": now,
		"startJd":              now,
		"endJd":                now,
		"durationEnable":       now,
	}
	err = db.Insert(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{
		Data: map[string]string{
			"eventId": eventId,
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func GetAllDayTask(req Request) (*Response, error) {
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	var email string
	userModel, err := CheckAuth(req)
	if err == nil {
		email = userModel.Email
	} else {
		tokenStr, _ := req.GetString("token")
		if tokenStr == nil || *tokenStr == "" {
			return nil, err
		}
		emailPtr, err := event_lib.CheckEventInvitationToken(*tokenStr, eventId)
		if err != nil || emailPtr == nil {
			return nil, ForbiddenError("invalid event invitation token", err)
		}
		email = *emailPtr
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err).Add(
				"msg", "event meta not found",
			)
		}
		return nil, NewError(Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	if !settings.ALLOW_MISMATCH_EVENT_TYPE {
		if eventMeta.EventType != "allDayTask" {
			return nil, NewError(
				InvalidArgument,
				fmt.Sprintf(
					"mismatch {eventType}, must be '%s'",
					eventMeta.EventType,
				),
				nil,
			)
		}
	}

	eventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err).Add(
				"msg", "event revision not found",
			)
		}
		return nil, NewError(Internal, "", err)
	}

	eventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err).Add(
				"msg", "event data not found",
			)
		}
		return nil, NewError(Internal, "", err)
	}

	eventModel.Id = *eventId
	eventModel.DummyType = eventMeta.EventType  // not "allDayTask"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	return &Response{
		Data: eventModel,
	}, nil
}

func UpdateAllDayTask(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.AllDayTaskEventModel{} // DYNAMIC
	// -----------------------------------------------
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Event(*eventId)
	if failed {
		return nil, NewError(ResourceLocked, "event is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	err = req.BodyTo(&eventModel)
	if err != nil {
		// msg := err.Error()
		// if strings.Contains(msg, "invalid ObjectId in JSON") {
		// 	msg = "invalid 'eventId'"
		// }
		return nil, err
	}
	_, err = eventModel.GetEvent() // for validation
	if err != nil {
		return nil, NewError(InvalidArgument, err.Error(), err)
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	// check if event exists, and has access to
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", nil)
	}

	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	lastEventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event snapshot not found", err)
		}
		return nil, NewError(Internal, "", err)
	}

	if eventModel.Id != "" {
		return nil, NewError(InvalidArgument, "'eventId' must not be present in JSON", nil)
	}
	if eventModel.GroupId != "" {
		return nil, NewError(InvalidArgument, "'groupId' must not be present in JSON", nil)
	}
	if len(eventModel.Meta) != 0 {
		return nil, NewError(InvalidArgument, "'meta' must not be present in JSON", nil)
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
		// FIXME: BadRequest or Internal error?
		return nil, NewError(Internal, "", err)
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
				// FIXME: BadRequest or Internal error?
				return nil, NewError(Internal, "", err)
			}
		} else {
			return nil, NewError(Internal, "", err)
		}
	}
	// PARAM="timeZone", PARAM_TYPE="string", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.TimeZone,
		lastEventModel.TimeZone,
	) {
		eventMeta.FieldsMtime["timeZone"] = now
	}
	// PARAM="timeZoneEnable", PARAM_TYPE="bool", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.TimeZoneEnable,
		lastEventModel.TimeZoneEnable,
	) {
		eventMeta.FieldsMtime["timeZoneEnable"] = now
	}
	// PARAM="calType", PARAM_TYPE="string", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.CalType,
		lastEventModel.CalType,
	) {
		eventMeta.FieldsMtime["calType"] = now
	}
	// PARAM="summary", PARAM_TYPE="string", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.Summary,
		lastEventModel.Summary,
	) {
		eventMeta.FieldsMtime["summary"] = now
	}
	// PARAM="description", PARAM_TYPE="string", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.Description,
		lastEventModel.Description,
	) {
		eventMeta.FieldsMtime["description"] = now
	}
	// PARAM="icon", PARAM_TYPE="string", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.Icon,
		lastEventModel.Icon,
	) {
		eventMeta.FieldsMtime["icon"] = now
	}
	// PARAM="summaryEncrypted", PARAM_TYPE="bool", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.SummaryEncrypted,
		lastEventModel.SummaryEncrypted,
	) {
		eventMeta.FieldsMtime["summaryEncrypted"] = now
	}
	// PARAM="descriptionEncrypted", PARAM_TYPE="bool", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.DescriptionEncrypted,
		lastEventModel.DescriptionEncrypted,
	) {
		eventMeta.FieldsMtime["descriptionEncrypted"] = now
	}
	// PARAM="startJd", PARAM_TYPE="int", PARAM_INT=true
	if !reflect.DeepEqual(
		eventModel.StartJd,
		lastEventModel.StartJd,
	) {
		eventMeta.FieldsMtime["startJd"] = now
	}
	// PARAM="endJd", PARAM_TYPE="int", PARAM_INT=true
	if !reflect.DeepEqual(
		eventModel.EndJd,
		lastEventModel.EndJd,
	) {
		eventMeta.FieldsMtime["endJd"] = now
	}
	// PARAM="durationEnable", PARAM_TYPE="bool", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.DurationEnable,
		lastEventModel.DurationEnable,
	) {
		eventMeta.FieldsMtime["durationEnable"] = now
	}
	err = db.Update(eventMeta) // just for FieldsMtime
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{
		Data: map[string]string{
			"eventId": *eventId,
			"sha1":    eventRev.Sha1,
		},
	}, nil
}

func PatchAllDayTask(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Event(*eventId)
	if failed {
		return nil, NewError(ResourceLocked, "event is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	patchMap := map[string]interface{}{}
	err = req.BodyTo(&patchMap)
	if err != nil {
		// msg := err.Error()
		// if strings.Contains(msg, "invalid ObjectId in JSON") {
		// 	msg = "invalid 'eventId'"
		// }
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	// check if event exists, and has access to
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", nil)
	}

	// do we need the last revision? to compare or what?
	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	eventModel, err := event_lib.LoadAllDayTaskEventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	now := time.Now()
	{
		rawValue, ok := patchMap["timeZone"]
		if ok {
			value, typeOk := rawValue.(string)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'timeZone'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'timeZoneEnable'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'calType'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'summary'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'description'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'icon'",
					nil,
				)
			}
			eventModel.Icon = value
			delete(patchMap, "icon")
			eventMeta.FieldsMtime["icon"] = now
		}
	}
	{
		rawValue, ok := patchMap["summaryEncrypted"]
		if ok {
			value, typeOk := rawValue.(bool)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'summaryEncrypted'",
					nil,
				)
			}
			eventModel.SummaryEncrypted = value
			delete(patchMap, "summaryEncrypted")
			eventMeta.FieldsMtime["summaryEncrypted"] = now
		}
	}
	{
		rawValue, ok := patchMap["descriptionEncrypted"]
		if ok {
			value, typeOk := rawValue.(bool)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'descriptionEncrypted'",
					nil,
				)
			}
			eventModel.DescriptionEncrypted = value
			delete(patchMap, "descriptionEncrypted")
			eventMeta.FieldsMtime["descriptionEncrypted"] = now
		}
	}
	{
		rawValue, ok := patchMap["startJd"]
		if ok {
			// json Unmarshal converts int to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'startJd'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'endJd'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'durationEnable'",
					nil,
				)
			}
			eventModel.DurationEnable = value
			delete(patchMap, "durationEnable")
			eventMeta.FieldsMtime["durationEnable"] = now
		}
	}
	if len(patchMap) > 0 {
		extraNames := []string{}
		for param := range patchMap {
			extraNames = append(extraNames, param)
		}
		return nil, NewError(
			InvalidArgument,
			fmt.Sprintf(
				"extra parameters: %v",
				strings.Join(extraNames, ", "),
			),
			nil,
		)
	}
	_, err = eventModel.GetEvent() // for validation
	if err != nil {
		return nil, NewError(InvalidArgument, err.Error(), err)
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
		// FIXME: BadRequest or Internal error?
		return nil, NewError(Internal, "", err)
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
				// FIXME: BadRequest or Internal error?
				return nil, NewError(Internal, "", err)
			}
		} else {
			return nil, NewError(Internal, "", err)
		}
	}
	err = db.Update(eventMeta) // just for FieldsMtime
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{
		Data: map[string]string{
			"eventId": *eventId,
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func MergeAllDayTask(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Event(*eventId)
	if failed {
		return nil, NewError(ResourceLocked, "event is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	inputStruct := struct {
		Event event_lib.AllDayTaskEventModel `json:"event"` // DYNAMIC

		LastMergeSha1 string               `json:"lastMergeSha1"`
		FieldsMtime   map[string]time.Time `json:"fieldsMtime"`
	}{}

	err = req.BodyTo(&inputStruct)
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	// if inputStruct.Event.DummyType == "" {
	//	return nil, NewError(MissingArgument, "missing 'eventType'", nil)
	// }
	if inputStruct.Event.Id == "" {
		return nil, NewError(MissingArgument, "missing 'eventId'", nil)
	}
	// FIXME: LastMergeSha1 can be empty?
	if inputStruct.LastMergeSha1 == "" {
		return nil, NewError(MissingArgument, "missing 'lastMergeSha1'", nil)
	}
	inputEventModel := &inputStruct.Event
	if inputEventModel.Id != *eventId {
		return nil, NewError(InvalidArgument, "mismatch 'event.id'", nil)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't own this event", nil)
	}

	lastRevModel, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	parentEventModel, err := event_lib.LoadAllDayTaskEventModel(db, inputStruct.LastMergeSha1)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(InvalidArgument, "invalid lastMergeSha1: revision not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	lastEventModel, err := event_lib.LoadAllDayTaskEventModel(db, lastRevModel.Sha1)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	fmt.Println(parentEventModel)
	fmt.Println(lastEventModel)

	// performing a 3-way merge
	// C <== parentEventModel	<== The common ancestor (last merge for this client)
	// A <== inputEventModel	<== The input (client's latest) data
	// B <== lastEventModel		<== The current (server's latest) data
	now := time.Now()
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="timeZone", PARAM_TYPE="string", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="timeZoneEnable", PARAM_TYPE="bool", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="calType", PARAM_TYPE="string", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="summary", PARAM_TYPE="string", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="description", PARAM_TYPE="string", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="icon", PARAM_TYPE="string", PARAM_INT=false
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
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="summaryEncrypted", PARAM_TYPE="bool", PARAM_INT=false
		inputValue := inputEventModel.SummaryEncrypted
		lastValue := lastEventModel.SummaryEncrypted
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.SummaryEncrypted
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.SummaryEncrypted = inputValue
			eventMeta.FieldsMtime["summaryEncrypted"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["summaryEncrypted"].After(eventMeta.FieldsMtime["summaryEncrypted"]) {
			lastEventModel.SummaryEncrypted = inputValue
			eventMeta.FieldsMtime["summaryEncrypted"] = now
			return
		}
	}()
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="descriptionEncrypted", PARAM_TYPE="bool", PARAM_INT=false
		inputValue := inputEventModel.DescriptionEncrypted
		lastValue := lastEventModel.DescriptionEncrypted
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.DescriptionEncrypted
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.DescriptionEncrypted = inputValue
			eventMeta.FieldsMtime["descriptionEncrypted"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["descriptionEncrypted"].After(eventMeta.FieldsMtime["descriptionEncrypted"]) {
			lastEventModel.DescriptionEncrypted = inputValue
			eventMeta.FieldsMtime["descriptionEncrypted"] = now
			return
		}
	}()
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="startJd", PARAM_TYPE="int", PARAM_INT=true
		inputValue := inputEventModel.StartJd
		lastValue := lastEventModel.StartJd
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.StartJd
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.StartJd = inputValue
			eventMeta.FieldsMtime["startJd"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["startJd"].After(eventMeta.FieldsMtime["startJd"]) {
			lastEventModel.StartJd = inputValue
			eventMeta.FieldsMtime["startJd"] = now
			return
		}
	}()
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="endJd", PARAM_TYPE="int", PARAM_INT=true
		inputValue := inputEventModel.EndJd
		lastValue := lastEventModel.EndJd
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.EndJd
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.EndJd = inputValue
			eventMeta.FieldsMtime["endJd"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["endJd"].After(eventMeta.FieldsMtime["endJd"]) {
			lastEventModel.EndJd = inputValue
			eventMeta.FieldsMtime["endJd"] = now
			return
		}
	}()
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="durationEnable", PARAM_TYPE="bool", PARAM_INT=false
		inputValue := inputEventModel.DurationEnable
		lastValue := lastEventModel.DurationEnable
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.DurationEnable
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.DurationEnable = inputValue
			eventMeta.FieldsMtime["durationEnable"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["durationEnable"].After(eventMeta.FieldsMtime["durationEnable"]) {
			lastEventModel.DurationEnable = inputValue
			eventMeta.FieldsMtime["durationEnable"] = now
			return
		}
	}()
	// err = db.Update(eventMeta) // just for FieldsMtime
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }

	return &Response{}, nil
}
