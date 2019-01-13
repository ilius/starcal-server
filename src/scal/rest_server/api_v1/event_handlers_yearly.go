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
	"reflect"

	"github.com/globalsign/mgo/bson"
	. "github.com/ilius/ripo"

	"scal/event_lib"
	"scal/settings"
	"scal/storage"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/yearly",
		Map: RouteMap{
			"AddYearly": {
				Method:  "POST",
				Pattern: "",
				Handler: AddYearly,
			},
			"GetYearly": {
				Method:  "GET",
				Pattern: ":eventId",
				Handler: GetYearly,
			},
			"UpdateYearly": {
				Method:  "PUT",
				Pattern: ":eventId",
				Handler: UpdateYearly,
			},
			"PatchYearly": {
				Method:  "PATCH",
				Pattern: ":eventId",
				Handler: PatchYearly,
			},
			"MergeYearly": {
				Method:  "POST",
				Pattern: ":eventId/merge",
				Handler: MergeYearly,
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibility
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_yearly": {
				Method:  "DELETE",
				Pattern: ":eventId",
				Handler: DeleteEvent,
			},
			"SetEventGroupId_yearly": {
				Method:  "PUT",
				Pattern: ":eventId/group",
				Handler: SetEventGroupId,
			},
			"GetEventOwner_yearly": {
				Method:  "GET",
				Pattern: ":eventId/owner",
				Handler: GetEventOwner,
			},
			"SetEventOwner_yearly": {
				Method:  "PUT",
				Pattern: ":eventId/owner",
				Handler: SetEventOwner,
			},
			"GetEventMeta_yearly": {
				Method:  "GET",
				Pattern: ":eventId/meta",
				Handler: GetEventMeta,
			},
			"GetEventAccess_yearly": {
				Method:  "GET",
				Pattern: ":eventId/access",
				Handler: GetEventAccess,
			},
			"SetEventAccess_yearly": {
				Method:  "PUT",
				Pattern: ":eventId/access",
				Handler: SetEventAccess,
			},
			"AppendEventAccess_yearly": {
				Method:  "POST",
				Pattern: ":eventId/access",
				Handler: AppendEventAccess,
			},
			"JoinEvent_yearly": {
				Method:  "GET",
				Pattern: ":eventId/join",
				Handler: JoinEvent,
			},
			"LeaveEvent_yearly": {
				Method:  "GET",
				Pattern: ":eventId/leave",
				Handler: LeaveEvent,
			},
			"InviteToEvent_yearly": {
				Method:  "POST",
				Pattern: ":eventId/invite",
				Handler: InviteToEvent,
			},
		},
	})
}

func AddYearly(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.YearlyEventModel{} // DYNAMIC
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
	eventId := bson.NewObjectId()
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
		//AccessEmails: []string{}
	}
	now := time.Now()
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:  now,
		Email: email,

		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,

		EventId:  eventId,
		FuncName: "AddYearly",
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
	_, err = event_lib.LoadYearlyEventModel(
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
		return nil, NewError(Internal, "", err)
	}
	return &Response{
		Data: map[string]string{
			"eventId": eventId.Hex(),
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func GetYearly(req Request) (*Response, error) {
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
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	if !settings.ALLOW_MISMATCH_EVENT_TYPE {
		if eventMeta.EventType != "yearly" {
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
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}

	eventModel, err := event_lib.LoadYearlyEventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}

	eventModel.Id = *eventId
	eventModel.DummyType = eventMeta.EventType  // not "yearly"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	return &Response{
		Data: eventModel,
	}, nil
}

func UpdateYearly(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.YearlyEventModel{} // DYNAMIC
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
	lastEventModel, err := event_lib.LoadYearlyEventModel(
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
	_, err = event_lib.LoadYearlyEventModel(
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
	// PARAM="month", PARAM_TYPE="uint8", PARAM_INT=true
	if !reflect.DeepEqual(
		eventModel.Month,
		lastEventModel.Month,
	) {
		eventMeta.FieldsMtime["month"] = now
	}
	// PARAM="day", PARAM_TYPE="uint8", PARAM_INT=true
	if !reflect.DeepEqual(
		eventModel.Day,
		lastEventModel.Day,
	) {
		eventMeta.FieldsMtime["day"] = now
	}
	// PARAM="startYear", PARAM_TYPE="int", PARAM_INT=true
	if !reflect.DeepEqual(
		eventModel.StartYear,
		lastEventModel.StartYear,
	) {
		eventMeta.FieldsMtime["startYear"] = now
	}
	// PARAM="startYearEnable", PARAM_TYPE="bool", PARAM_INT=false
	if !reflect.DeepEqual(
		eventModel.StartYearEnable,
		lastEventModel.StartYearEnable,
	) {
		eventMeta.FieldsMtime["startYearEnable"] = now
	}
	err = db.Update(eventMeta) // just for FieldsMtime
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{
		Data: map[string]string{
			"eventId": eventId.Hex(),
			"sha1":    eventRev.Sha1,
		},
	}, nil
}
func PatchYearly(req Request) (*Response, error) {
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
	eventModel, err := event_lib.LoadYearlyEventModel(
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
		rawValue, ok := patchMap["month"]
		if ok {
			// json Unmarshal converts uint8 to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'month'",
					nil,
				)
			}
			eventModel.Month = uint8(value)
			delete(patchMap, "month")
			eventMeta.FieldsMtime["month"] = now
		}
	}
	{
		rawValue, ok := patchMap["day"]
		if ok {
			// json Unmarshal converts uint8 to float64
			value, typeOk := rawValue.(float64)
			if !typeOk {
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'day'",
					nil,
				)
			}
			eventModel.Day = uint8(value)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'startYear'",
					nil,
				)
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
				return nil, NewError(
					InvalidArgument,
					"bad type for parameter 'startYearEnable'",
					nil,
				)
			}
			eventModel.StartYearEnable = value
			delete(patchMap, "startYearEnable")
			eventMeta.FieldsMtime["startYearEnable"] = now
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
	_, err = event_lib.LoadYearlyEventModel(
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
			"eventId": eventId.Hex(),
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func MergeYearly(req Request) (*Response, error) {
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
		Event event_lib.YearlyEventModel `json:"event"` // DYNAMIC

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
	if inputEventModel.Id.Hex() != eventId.Hex() {
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
	parentEventModel, err := event_lib.LoadYearlyEventModel(db, inputStruct.LastMergeSha1)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(InvalidArgument, "invalid lastMergeSha1: revision not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	lastEventModel, err := event_lib.LoadYearlyEventModel(db, lastRevModel.Sha1)
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
	func() {
		// PARAM="month", PARAM_TYPE="uint8", PARAM_INT=true
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
		// PARAM="day", PARAM_TYPE="uint8", PARAM_INT=true
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
		// PARAM="startYear", PARAM_TYPE="int", PARAM_INT=true
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
		// PARAM="startYearEnable", PARAM_TYPE="bool", PARAM_INT=false
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
	// err = db.Update(eventMeta) // just for FieldsMtime
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }

	return &Response{}, nil
}
