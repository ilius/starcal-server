package api_v1

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/event_lib"
	"github.com/ilius/starcal-server/pkg/scal/storage"
	"github.com/ilius/starcal-server/pkg/scal/user_lib"

	"github.com/ilius/mgo/bson"
	rp "github.com/ilius/ripo"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/copy",
		Map: RouteMap{
			"CopyEvent": {
				Method:  "POST",
				Pattern: ":eventId",
				Handler: CopyEvent,
			},
		},
	})
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/ungrouped",
		Map: RouteMap{
			"GetUngroupedEvents": {
				Method:  "GET",
				Pattern: "",
				Handler: GetUngroupedEvents,
			},
		},
	})
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/my",
		Map: RouteMap{
			"GetMyEventList": {
				Method:  "GET",
				Pattern: "events",
				Handler: GetMyEventList,
			},
			"GetMyEventsFull": {
				Method:  "GET",
				Pattern: "events-full",
				Handler: GetMyEventsFull,
			},
		},
	})
}

func DeleteEvent(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}

	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}

	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", nil)
	}
	now := time.Now()
	metaChangeLog := event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "DeleteEvent",
		OwnerEmail: &[2]*string{
			&eventMeta.OwnerEmail,
			nil,
		},
	}
	if eventMeta.GroupId != nil {
		metaChangeLog.GroupId = &[2]*string{
			eventMeta.GroupId,
			nil,
		}
	}

	if len(eventMeta.AccessEmails) > 0 {
		metaChangeLog.AccessEmails = &[2][]string{
			eventMeta.AccessEmails,
			{},
		}
	}
	err = db.Insert(metaChangeLog)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = db.Insert(event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventMeta.EventType,
		Sha1:      "",
		Time:      now,
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = db.Remove(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{}, nil
}

func CopyEvent(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	oldEventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, oldEventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", nil)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	eventRev, err := event_lib.LoadLastRevisionModel(db, oldEventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", nil)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	newEventId := bson.NewObjectId().Hex()

	newGroupId := userModel.DefaultGroupId
	if eventMeta.GroupModel != nil {
		if eventMeta.GroupModel.OwnerEmail == email {
			newGroupId = &eventMeta.GroupModel.Id // == eventMeta.GroupId
		}
	}
	now := time.Now()
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  newEventId,
		FuncName: "CopyEvent",
		OwnerEmail: &[2]*string{
			nil,
			&email,
		},
		GroupId: &[2]*string{
			nil,
			newGroupId,
		},
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = db.Insert(event_lib.EventMetaModel{
		EventId:      newEventId,
		EventType:    eventMeta.EventType,
		CreationTime: time.Now(),
		OwnerEmail:   email,
		GroupId:      newGroupId,
		// AccessEmails: []string{}// must not copy AccessEmails
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventRev.EventId = newEventId
	err = db.Insert(eventRev)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}

	return &rp.Response{
		Data: map[string]string{
			"eventType": eventRev.EventType,
			"eventId":   newEventId,
			"sha1":      eventRev.Sha1,
		},
	}, nil
}

func SetEventGroupId(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", err)
	}
	newGroupId, err := req.GetString("newGroupId")
	if err != nil {
		return nil, err
	}
	if !bson.IsObjectIdHex(*newGroupId) {
		return nil, rp.NewError(rp.InvalidArgument, "invalid 'newGroupId'", nil)
		// to avoid panic!
	}
	newGroupModel := event_lib.EventGroupModel{
		Id: *newGroupId,
	}
	err = db.Get(&newGroupModel)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if !newGroupModel.EmailCanAdd(email) {
		return nil, ForbiddenError("you don't have write access to this group", nil)
	}

	now := time.Now()
	metaChangeLog := event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "SetEventGroupId",
		GroupId: &[2]*string{
			eventMeta.GroupId,
			newGroupId,
		},
	}

	/*
		addedAccessEmails := Set(
			eventMeta.GroupModel.ReadAccessEmails,
		).Difference(newGroupModel.ReadAccessEmails)
		if addedAccessEmails {
			metaChangeLog.AddedAccessEmails = addedAccessEmails
		}
	*/
	err = db.Insert(metaChangeLog)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	/*userModel := UserModelByEmail(email, db)
	  if userModel == nil {
	      SetHttpErrorUserNotFound(w, email)
	      return
	  }*/
	eventMeta.GroupId = newGroupId
	err = db.Update(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{}, nil
}

func GetEventOwner(req rp.Request) (*rp.Response, error) {
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
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	return &rp.Response{
		Data: scal.M{
			//"eventId": eventId.Hex(),
			"ownerEmail": eventMeta.OwnerEmail,
		},
	}, nil
}

func SetEventOwner(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	newOwnerEmail, err := req.GetString("newOwnerEmail")
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't own this event", err)
	}
	// should check if user with `newOwnerEmail` exists?
	newOwnerUserModel := user_lib.UserModelByEmail(*newOwnerEmail, db)
	if newOwnerUserModel == nil {
		return nil, rp.NewError(
			rp.InvalidArgument,
			fmt.Sprintf(
				"user with email '%s' does not exist",
				*newOwnerEmail,
			),
			nil,
		)
	}
	now := time.Now()
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "SetEventOwner",
		OwnerEmail: &[2]*string{
			&eventMeta.OwnerEmail,
			newOwnerEmail,
		},
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventMeta.OwnerEmail = *newOwnerEmail
	err = db.Update(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	// send an E-Mail to `newOwnerEmail` FIXME
	return &rp.Response{}, nil
}

func GetEventMetaModelFromRequest(req rp.Request, email string) (*event_lib.EventMetaModel, error) {
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.InvalidArgument, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	if !eventMeta.CanReadFull(email) {
		return nil, ForbiddenError("you can't meta information of this event", nil)
	}
	return eventMeta, nil
}

func GetEventMeta(req rp.Request) (*rp.Response, error) {
	// includes owner, creation time, groupId, access info, attendings info
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventMeta, err := GetEventMetaModelFromRequest(req, email)
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}

	return &rp.Response{
		Data: scal.M{
			//"eventId": eventMeta.EventId.Hex(),
			"ownerEmail":           eventMeta.OwnerEmail,
			"creationTime":         eventMeta.CreationTime,
			"fieldsMtime":          eventMeta.FieldsMtime,
			"groupId":              eventMeta.GroupIdHex(),
			"isPublic":             eventMeta.IsPublic,
			"accessEmails":         eventMeta.AccessEmails,
			"publicJoinOpen":       eventMeta.PublicJoinOpen,
			"maxAttendees":         eventMeta.MaxAttendees,
			"attendingEmails":      eventMeta.GetAttendingEmails(db),
			"notAttendingEmails":   eventMeta.GetNotAttendingEmails(db),
			"maybeAttendingEmails": eventMeta.GetMaybeAttendingEmails(db),
		},
	}, nil
}

func GetEventAccess(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventMeta, err := GetEventMetaModelFromRequest(req, email)
	if err != nil {
		return nil, err
	}
	return &rp.Response{
		Data: scal.M{
			//"eventId": eventMeta.EventId.Hex(),
			"isPublic":       eventMeta.IsPublic,
			"accessEmails":   eventMeta.AccessEmails,
			"publicJoinOpen": eventMeta.PublicJoinOpen,
			"maxAttendees":   eventMeta.MaxAttendees,
		},
	}, nil
}

func SetEventAccess(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
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

	newIsPublic, err := req.GetBool("isPublic")
	if err != nil {
		return nil, err
	}
	newAccessEmails, err := req.GetStringList("accessEmails")
	if err != nil {
		return nil, err
	}
	newPublicJoinOpen, err := req.GetBool("publicJoinOpen")
	if err != nil {
		return nil, err
	}
	newMaxAttendees, err := req.GetInt("maxAttendees")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	metaChangeLog := event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "SetEventAccess",
	}
	if *newIsPublic != eventMeta.IsPublic {
		metaChangeLog.IsPublic = &[2]bool{
			eventMeta.IsPublic,
			*newIsPublic,
		}
		eventMeta.IsPublic = *newIsPublic
	}
	if !reflect.DeepEqual(newAccessEmails, eventMeta.AccessEmails) {
		metaChangeLog.AccessEmails = &[2][]string{
			eventMeta.AccessEmails,
			newAccessEmails,
		}
		eventMeta.AccessEmails = newAccessEmails
	}
	if *newPublicJoinOpen != eventMeta.PublicJoinOpen {
		metaChangeLog.PublicJoinOpen = &[2]bool{
			eventMeta.PublicJoinOpen,
			*newPublicJoinOpen,
		}
		eventMeta.PublicJoinOpen = *newPublicJoinOpen

	}
	if *newMaxAttendees != eventMeta.MaxAttendees {
		metaChangeLog.MaxAttendees = &[2]int{
			eventMeta.MaxAttendees,
			*newMaxAttendees,
		}
		eventMeta.MaxAttendees = *newMaxAttendees
	}
	err = db.Insert(metaChangeLog)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = db.Update(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{}, nil
}

func AppendEventAccess(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
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

	toAddEmail, err := req.GetString("toAddEmail")
	if err != nil {
		return nil, err
	}

	newAccessEmails := append(eventMeta.AccessEmails, *toAddEmail)
	now := time.Now()
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "AppendEventAccess",
		AccessEmails: &[2][]string{
			eventMeta.AccessEmails,
			newAccessEmails,
		},
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventMeta.AccessEmails = newAccessEmails
	err = db.Update(eventMeta)
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	return &rp.Response{}, nil
}

func joinEventWithToken(tokenStr string, eventId *string) (*rp.Response, error) {
	email, err := event_lib.CheckEventInvitationToken(tokenStr, eventId)
	if err != nil {
		return nil, ForbiddenError("invalid event invitation token", err)
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = eventMeta.Join(db, *email)
	if err != nil {
		return nil, ForbiddenError(err.Error(), err)
	}

	values := url.Values{}
	values.Add("token", tokenStr)
	eventPath := fmt.Sprintf(
		"/event/%v/%v/?%v",
		eventMeta.EventType,
		eventMeta.EventId,
		values.Encode(),
	)
	fmt.Println("eventPath:", eventPath)
	return &rp.Response{
		RedirectPath: eventPath,
	}, nil
}

func JoinEvent(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		tokenPtr, _ := req.GetString("token")
		if tokenPtr != nil {
			eventId, err := ObjectIdFromRequest(req, "eventId")
			if err != nil {
				return nil, err
			}
			return joinEventWithToken(*tokenPtr, eventId)
		}
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	// remoteIp, err := req.RemoteIP()
	// if err != nil {
	// 	return nil, err
	// }
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	if !userModel.EmailConfirmed {
		return nil, ForbiddenError("you need to confirm your email address before joining an event", nil)
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = eventMeta.Join(db, email)
	if err != nil {
		return nil, ForbiddenError(err.Error(), err)
	}
	return &rp.Response{}, nil
}

func LeaveEvent(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	// remoteIp, err := req.RemoteIP()
	// if err != nil {
	// 	return nil, err
	// }
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = eventMeta.Leave(db, email)
	if err != nil {
		return nil, ForbiddenError(err.Error(), err)
	}
	return &rp.Response{}, nil
}

func InviteToEvent(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	eventId, err := ObjectIdFromRequest(req, "eventId")
	if err != nil {
		return nil, err
	}

	inviteEmails, err := req.GetStringList("inviteEmails")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, rp.NewError(rp.NotFound, "event not found", err)
		}
		return nil, rp.NewError(rp.Internal, "", err)
	}
	err = eventMeta.Invite(
		db,
		email,
		inviteEmails,
		remoteIp,
		"http://"+req.Host(), // FIXME
	)
	if err != nil {
		return nil, err
	}
	return &rp.Response{}, nil
}

func GetUngroupedEvents(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	pageOpts, err := GetPageOptions(req)
	if err != nil {
		return nil, err
	}

	cond := db.NewCondition(storage.AND)
	cond.Equals("ownerEmail", email)
	cond.Equals("groupId", nil)
	cond.SetPageOptions(pageOpts)

	var results []*event_lib.ListEventsRow
	err = db.FindAll(&results, &storage.FindInput{
		Collection:   storage.C_eventMeta,
		Condition:    cond,
		SortBy:       "_id",
		ReverseOrder: pageOpts.ReverseOrder,
		Limit:        pageOpts.Limit,
		Fields:       []string{"_id", "eventType"},
	})
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = make([]*event_lib.ListEventsRow, 0)
	}
	output := scal.M{
		"events": results,
	}
	if len(results) > 0 {
		output["lastId"] = results[len(results)-1].EventId
	}
	return &rp.Response{
		Data: output,
	}, nil
}

func GetMyEventList(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	pageOpts, err := GetPageOptions(req)
	if err != nil {
		return nil, err
	}

	cond := db.NewCondition(storage.AND)
	cond.Equals("ownerEmail", email)
	cond.SetPageOptions(pageOpts)

	var results []*event_lib.ListEventsRow
	err = db.FindAll(&results, &storage.FindInput{
		Collection:   storage.C_eventMeta,
		Condition:    cond,
		SortBy:       "_id",
		ReverseOrder: pageOpts.ReverseOrder,
		Limit:        pageOpts.Limit,
		Fields: []string{
			"_id",
			"eventType",
			// "groupId",
		},
	})
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}

	if results == nil {
		results = make([]*event_lib.ListEventsRow, 0)
	}
	output := scal.M{
		"events": results,
	}
	if len(results) > 0 {
		output["lastId"] = results[len(results)-1].EventId
	}

	return &rp.Response{
		Data: output,
	}, nil
}

func getMyEventsFullData(db storage.Database, email string, pageOpts *scal.PageOptions) ([]scal.M, error) {
	pipeline := NewPipelines(db, storage.C_eventMeta)
	pipeline.MatchValue("ownerEmail", email)
	if pageOpts != nil {
		pipeline.SetPageOptions(pageOpts)
	}
	pipeline.Lookup(storage.C_revision, "_id", "eventId", "revision")
	pipeline.Unwind("revision")
	pipeline.GroupBy("_id").
		AddFromFirst("groupId", "groupId").
		AddFromFirst("eventType", "eventType").
		AddFromFirst("revision.sha1", "lastSha1").
		AddFromFirst("revision.time", "lastModifiedTime").
		AddFromFirst("ownerEmail", "ownerEmail").
		AddFromFirst("isPublic", "isPublic").
		AddFromFirst("creationTime", "creationTime").
		AddFromFirst("accessEmails", "accessEmails").
		AddFromFirst("publicJoinOpen", "publicJoinOpen").
		AddFromFirst("maxAttendees", "maxAttendees")

	pipeline.Lookup(storage.C_eventData, "lastSha1", "sha1", "data")
	pipeline.Unwind("data")

	return GetEventMetaPipeResults(pipeline, []string{
		"ownerEmail",
		"isPublic",
		"creationTime",
		"accessEmails",
		"publicJoinOpen",
		"maxAttendees",
	})
}

func GetMyEventsFull(req rp.Request) (*rp.Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Unavailable, "", err)
	}
	pageOpts, err := GetPageOptions(req)
	if err != nil {
		return nil, err
	}
	results, err := getMyEventsFullData(db, email, pageOpts)
	if err != nil {
		return nil, err
	}
	output := scal.M{
		"eventsFull": results,
	}
	if len(results) > 0 {
		output["lastId"] = results[len(results)-1]["eventId"]
	}
	return &rp.Response{
		Data: output,
	}, nil
}
