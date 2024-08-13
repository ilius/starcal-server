// if this is a *.go file, don't modify it, it's auto-generated
// from a Go template file named `*.go.tpl` inside "templates" directory
//
//go:generate $GOPATH/pkg/scal/rest_server/gen -event-type {{.EVENT_TYPE}}
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
	rp "github.com/ilius/ripo"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/{{.EVENT_TYPE}}",
		Map: RouteMap{
			"Add{{.CAP_EVENT_TYPE}}": {
				Method: "POST",
				Pattern: "",
				Handler: Add{{.CAP_EVENT_TYPE}},
			},
			"Get{{.CAP_EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId",
				Handler: Get{{.CAP_EVENT_TYPE}},
			},
			"Update{{.CAP_EVENT_TYPE}}": {
				Method: "PUT",
				Pattern: ":eventId",
				Handler: Update{{.CAP_EVENT_TYPE}},
			},
			"Patch{{.CAP_EVENT_TYPE}}": {
				Method: "PATCH",
				Pattern: ":eventId",
				Handler: Patch{{.CAP_EVENT_TYPE}},
			},
			"Merge{{.CAP_EVENT_TYPE}}": {
				Method: "POST",
				Pattern: ":eventId/merge",
				Handler: Merge{{.CAP_EVENT_TYPE}},
			},
			// functions of following operations are defined in handlers.go
			// because their definition does not depend on event type
			// but their URL still contains eventType for sake of compatibility
			// so we will have to register their routes for each event type
			// we don't use eventType in these functions
			"DeleteEvent_{{.EVENT_TYPE}}": {
				Method: "DELETE",
				Pattern: ":eventId",
				Handler: DeleteEvent,
			},
			"SetEventGroupId_{{.EVENT_TYPE}}": {
				Method: "PUT",
				Pattern: ":eventId/group",
				Handler: SetEventGroupId,
			},
			"GetEventOwner_{{.EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId/owner",
				Handler: GetEventOwner,
			},
			"SetEventOwner_{{.EVENT_TYPE}}": {
				Method: "PUT",
				Pattern: ":eventId/owner",
				Handler: SetEventOwner,
			},
			"GetEventMeta_{{.EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId/meta",
				Handler: GetEventMeta,
			},
			"GetEventAccess_{{.EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId/access",
				Handler: GetEventAccess,
			},
			"SetEventAccess_{{.EVENT_TYPE}}": {
				Method: "PUT",
				Pattern: ":eventId/access",
				Handler: SetEventAccess,
			},
			"AppendEventAccess_{{.EVENT_TYPE}}": {
				Method: "POST",
				Pattern: ":eventId/access",
				Handler: AppendEventAccess,
			},
			"JoinEvent_{{.EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId/join",
				Handler: JoinEvent,
			},
			"LeaveEvent_{{.EVENT_TYPE}}": {
				Method: "GET",
				Pattern: ":eventId/leave",
				Handler: LeaveEvent,
			},
			"InviteToEvent_{{.EVENT_TYPE}}": {
				Method: "POST",
				Pattern: ":eventId/invite",
				Handler: InviteToEvent,
			},
		},
	})
}

func Add{{.CAP_EVENT_TYPE}}(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.{{.CAP_EVENT_TYPE}}EventModel{} // DYNAMIC
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
		return nil, rp.NewError(rp.InvalidArgument, err.Error(), err) // FIXME: correct msg?
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	if eventModel.Id != "" {
		return nil, rp.NewError(rp.InvalidArgument, "you can't specify 'eventId'", nil)
	}
	eventModel.Sha1 = ""
	jsonByte, _ := json.Marshal(eventModel)
	eventModel.Sha1 = fmt.Sprintf("%x", sha1.Sum(jsonByte))
	eventId := bson.NewObjectId().Hex()
	eventModel.Id = eventId
	groupId := userModel.DefaultGroupId
	if eventModel.GroupId != "" {
		if !bson.IsObjectIdHex(eventModel.GroupId) {
			return nil, rp.NewError(rp.InvalidArgument, "invalid 'groupId'", nil)
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
			return nil, ForbiddenError("you don't have write access this event group",nil)
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
		Time:     now,
		Email:    email,

		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,

		EventId:  eventId,
		FuncName: "Add{{.CAP_EVENT_TYPE}}",
		OwnerEmail: &[2]*string{
			nil,
			&email,
		},
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = db.Insert(event_lib.EventRevisionModel{
		EventId:   eventId,
		EventType: eventModel.Type(),
		Sha1:      eventModel.Sha1,
		Time:      time.Now(),
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			err = db.Insert(eventModel)
			if err != nil {
				return nil, rp.NewError(rp.Internal, "", err)
			}
		} else {
			return nil, rp.NewError(rp.Internal, "", err)
		}
	}

	eventMeta.FieldsMtime = map[string]time.Time{
		{{ range .EVENT_PATCH_PARAMS }}
			"{{.PARAM}}": now,
		{{ end }}
	}
	{{ if eq .EVENT_TYPE "custom" }}
		for _, ruleModel := range eventModel.Rules {
			eventMeta.FieldsMtime["rules:"+ruleModel.Type] = now
		}
	{{ end }}
	err = db.Insert(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{
		Data: map[string]string{
			"eventId": eventId,
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func Get{{.CAP_EVENT_TYPE}}(req rp.Request) (*rp.Response, error) {
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
		return nil, rp.NewError(rp.Unavailable, "", err)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err).Add(
				"msg", "event meta not found",
			)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	if !settings.ALLOW_MISMATCH_EVENT_TYPE {
		if eventMeta.EventType != "{{.EVENT_TYPE}}" {
			return nil, rp.NewError(rp.
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
			return nil, rp.NewError(rp.NotFound, "event not found", err).Add(
				"msg", "event revision not found",
			)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}

	eventModel, err := event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err).Add(
				"msg", "event data not found",
			)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}

	eventModel.Id = *eventId
	eventModel.DummyType = eventMeta.EventType  // not "{{.EVENT_TYPE}}"
	eventModel.GroupId = eventMeta.GroupIdHex() // FIXME
	if eventMeta.CanReadFull(email) {
		eventModel.Meta = eventMeta.JsonM()
	}
	return &rp.Response{
		Data: eventModel,
	}, nil
}

func Update{{.CAP_EVENT_TYPE}}(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventModel := event_lib.{{.CAP_EVENT_TYPE}}EventModel{} // DYNAMIC
	// -----------------------------------------------
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Event(*eventId)
	if failed {
		return nil, rp.NewError(rp.ResourceLocked, "event is locked by another request", nil)
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
	{{ if eq .EVENT_TYPE "custom" }}
		event, err := eventModel.GetEvent() // for validation and comparing rules
	{{ else }}
		_, err = eventModel.GetEvent() // for validation
	{{ end }}
	if err != nil {
		return nil, rp.NewError(rp.InvalidArgument, err.Error(), err)
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}

	// check if event exists, and has access to
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", nil)
	}

	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	lastEventModel, err := event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event snapshot not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}

	if eventModel.Id != "" {
		return nil, rp.NewError(rp.InvalidArgument, "'eventId' must not be present in JSON", nil)
	}
	if eventModel.GroupId != "" {
		return nil, rp.NewError(rp.InvalidArgument, "'groupId' must not be present in JSON", nil)
	}
	if len(eventModel.Meta) != 0 {
		return nil, rp.NewError(rp.InvalidArgument, "'meta' must not be present in JSON", nil)
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
		return nil, rp.NewError(rp.Internal, "", err)
	}

	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		eventRev.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			err = db.Insert(eventModel)
			if err != nil {
				// FIXME: BadRequest or Internal error?
				return nil, rp.NewError(rp.Internal, "", err)
			}
		} else {
			return nil, rp.NewError(rp.Internal, "", err)
		}
	}

	{{ range .EVENT_PATCH_PARAMS }}
		// PARAM="{{.PARAM}}", PARAM_TYPE="{{.PARAM_TYPE}}", PARAM_INT={{.PARAM_INT}}
		{{ if eq .PARAM "rules" }}
			lastEvent, err := lastEventModel.GetEvent() // for comparing rules
			if err != nil {
				return nil, rp.NewError(rp.Internal, "", err)
			}
			modRuleTypes := event.GetModifiedRuleTypes(&lastEvent)
			if len(modRuleTypes) > 0 {
				eventMeta.FieldsMtime["rules"] = now
				for _, ruleType := range modRuleTypes {
					eventMeta.FieldsMtime["rules:"+ruleType.Name] = now
				}
			}
		{{ else }}
			if !reflect.DeepEqual(
				eventModel.{{.CAP_PARAM}},
				lastEventModel.{{.CAP_PARAM}},
			) {
				eventMeta.FieldsMtime["{{.PARAM}}"] = now
			}
		{{ end }}
	{{ end }}
	err = db.Update(eventMeta) // just for FieldsMtime
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}

	return &rp.Response{
		Data: map[string]string{
			"eventId": *eventId,
			"sha1":    eventRev.Sha1,
		},
	}, nil
}

func Patch{{.CAP_EVENT_TYPE}}(req rp.Request) (*rp.Response, error) {
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
		return nil, rp.NewError(rp.ResourceLocked, "event is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	patchMap := map[string]any{}
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
		return nil, rp.NewError(rp.Unavailable, "", err)
	}

	// check if event exists, and has access to
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, false)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", nil)
	}

	// do we need the last revision? to compare or what?
	lastEventRev, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventModel, err := event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		lastEventRev.Sha1,
	)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	{{ if eq .EVENT_TYPE "custom" }}
		lastEvent, err := eventModel.GetEvent()// for comparing rules
		if err != nil {
			return nil, rp.NewError(rp.Internal, "", err)
		}
		rulesModified := false
	{{ end }}
	now := time.Now()
	{{ range .EVENT_PATCH_PARAMS }}
	{
		rawValue, ok := patchMap["{{.PARAM}}"]
		if ok {
			{{ if eq .PARAM "rules" }}
				value, err := event_lib.DecodeMapEventRuleModelList(rawValue)
				if err != nil {
					return nil, rp.NewError(rp.
						InvalidArgument,
						"bad type or value for parameter 'rules': "+err.Error(),
						err,
					)
				}
			{{ else }}
				{{ if .PARAM_INT }}
					// json Unmarshal converts {{.PARAM_TYPE}} to float64
					value, typeOk := rawValue.(float64)
				{{ else if eq .PARAM_TYPE "*time.Time" }}
					value, typeOk := rawValue.(string)
				{{ else }}
					value, typeOk := rawValue.({{.PARAM_TYPE}})
				{{ end }}
				if !typeOk {
					return nil, rp.NewError(rp.
						InvalidArgument,
						"bad type for parameter '{{.PARAM}}'",
						nil,
					)
				}
			{{ end }}
			{{ if .PARAM_INT }}
				eventModel.{{.CAP_PARAM}} = {{.PARAM_TYPE}}(value)
			{{ else if eq .PARAM_TYPE "*time.Time" }}
				timeValue, err := time.Parse(time.RFC3339, value)
				if err != nil {
					// FIXME: give a better msg
					return nil, rp.NewError(rp.InvalidArgument, err.Error(), err)
				}
				eventModel.{{.CAP_PARAM}} = &timeValue
			{{ else }}
				eventModel.{{.CAP_PARAM}} = value
			{{ end }}
			delete(patchMap, "{{.PARAM}}")
			eventMeta.FieldsMtime["{{.PARAM}}"] = now
			{{ if eq .PARAM "rules" }}
				rulesModified = true
			{{ end }}
		}
	}{{ end }}
	if len(patchMap) > 0 {
		extraNames := []string{}
		for param := range patchMap {
			extraNames = append(extraNames, param)
		}
		return nil, rp.NewError(rp.
			InvalidArgument,
			fmt.Sprintf(
				"extra parameters: %v",
				strings.Join(extraNames, ", "),
			),
			nil,
		)
	}
	{{ if eq .EVENT_TYPE "custom" }}
		event, err := eventModel.GetEvent() // for validation and comparing rules
	{{ else }}
		_, err = eventModel.GetEvent() // for validation
	{{ end }}
	if err != nil {
		return nil, rp.NewError(rp.InvalidArgument, err.Error(), err)
	}
	{{ if eq .EVENT_TYPE "custom" }}
		if rulesModified {
			for _, ruleType := range event.GetModifiedRuleTypes(&lastEvent) {
				eventMeta.FieldsMtime["rules:"+ruleType.Name] = now
			}
		}
	{{ end }}

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
		return nil, rp.NewError(rp.Internal, "", err)
	}
	// don't store duplicate eventModel, even if it was added by another user
	// the (underlying) eventModel does not belong to anyone
	// like git's blobs and trees
	_, err = event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(
		db,
		eventModel.Sha1,
	)
	if err != nil {
		if db.IsNotFound(err) {
			err = db.Insert(eventModel)
			if err != nil {
				// FIXME: BadRequest or Internal error?
				return nil, rp.NewError(rp.Internal, "", err)
			}
		} else {
			return nil, rp.NewError(rp.Internal, "", err)
		}
	}
	err = db.Update(eventMeta) // just for FieldsMtime
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{
		Data: map[string]string{
			"eventId": *eventId,
			"sha1":    eventModel.Sha1,
		},
	}, nil
}

func Merge{{.CAP_EVENT_TYPE}}(req rp.Request) (*rp.Response, error) {
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
		return nil, rp.NewError(rp.ResourceLocked, "event is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	inputStruct := struct {
		Event event_lib.{{.CAP_EVENT_TYPE}}EventModel `json:"event"` // DYNAMIC

		LastMergeSha1 string               `json:"lastMergeSha1"`
		FieldsMtime   map[string]time.Time `json:"fieldsMtime"`
	}{}

	err = req.BodyTo(&inputStruct)
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	// if inputStruct.Event.DummyType == "" {
	//	return nil, rp.NewError(rp.MissingArgument, "missing 'eventType'", nil)
	// }
	if inputStruct.Event.Id == "" {
		return nil, rp.NewError(rp.MissingArgument, "missing 'eventId'", nil)
	}
	// FIXME: LastMergeSha1 can be empty?
	if inputStruct.LastMergeSha1 == "" {
		return nil, rp.NewError(rp.MissingArgument, "missing 'lastMergeSha1'", nil)
	}
	inputEventModel := &inputStruct.Event
	if inputEventModel.Id != *eventId {
		return nil, rp.NewError(rp.InvalidArgument, "mismatch 'event.id'", nil)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't own this event", nil)
	}

	lastRevModel, err := event_lib.LoadLastRevisionModel(db, eventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	parentEventModel, err := event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(db, inputStruct.LastMergeSha1)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.InvalidArgument, "invalid lastMergeSha1: revision not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	lastEventModel, err := event_lib.Load{{.CAP_EVENT_TYPE}}EventModel(db, lastRevModel.Sha1)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	fmt.Println(parentEventModel)
	fmt.Println(lastEventModel)

	// performing a 3-way merge
	// C <== parentEventModel	<== The common ancestor (last merge for this client)
	// A <== inputEventModel	<== The input (client's latest) data
	// B <== lastEventModel		<== The current (server's latest) data
	now := time.Now()
	{{ range .EVENT_PATCH_PARAMS }}
	// define a func because we want to return from it to avoid nested code
	func() {
		// PARAM="{{.PARAM}}", PARAM_TYPE="{{.PARAM_TYPE}}", PARAM_INT={{.PARAM_INT}}
		inputValue := inputEventModel.{{.CAP_PARAM}}
		lastValue := lastEventModel.{{.CAP_PARAM}}
		if reflect.DeepEqual(inputValue, lastValue) {
			return
		}
		parentValue := parentEventModel.{{.CAP_PARAM}}
		if reflect.DeepEqual(parentValue, lastValue) {
			return
		}
		if reflect.DeepEqual(parentValue, inputValue) {
			lastEventModel.{{.CAP_PARAM}} = inputValue
			eventMeta.FieldsMtime["{{.PARAM}}"] = now
			return
		}
		// Now we have a conflict
		if inputStruct.FieldsMtime["{{.PARAM}}"].After(eventMeta.FieldsMtime["{{.PARAM}}"]) {
			lastEventModel.{{.CAP_PARAM}} = inputValue
			eventMeta.FieldsMtime["{{.PARAM}}"] = now
			return
		}
	}()
	{{ end }}
	// err = db.Update(eventMeta) // just for FieldsMtime
	// if err != nil {
	// 	SetHttpErrorInternal(w, err)
	// 	return
	// }

	return &rp.Response{}, nil
}
