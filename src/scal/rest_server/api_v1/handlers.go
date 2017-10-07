package api_v1

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	. "github.com/ilius/restpc"
	"gopkg.in/mgo.v2/bson"

	"scal"
	"scal/event_lib"
	"scal/storage"
	. "scal/user_lib"
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
			"GetMyLastCreatedEvents": {
				Method:  "GET",
				Pattern: "last-created-events",
				Handler: GetMyLastCreatedEvents,
			},
		},
	})
}

func DeleteEvent(req Request) (*Response, error) {
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

	eventId, err := ObjectIdFromURL(req, "eventId", 0)
	if err != nil {
		return nil, err
	}

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
		metaChangeLog.GroupId = &[2]*bson.ObjectId{
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
		return nil, NewError(Internal, "", err)
	}
	err = db.Insert(event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventMeta.EventType,
		Sha1:      "",
		Time:      now,
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Remove(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func CopyEvent(req Request) (*Response, error) {
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
	oldEventId, err := ObjectIdFromURL(req, "eventId", 0)
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, oldEventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", nil)
		}
		return nil, NewError(Internal, "", err)
	}
	if !eventMeta.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event", nil)
	}
	eventRev, err := event_lib.LoadLastRevisionModel(db, oldEventId)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", nil)
		}
		return nil, NewError(Internal, "", err)
	}
	newEventId := bson.NewObjectId()

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
		GroupId: &[2]*bson.ObjectId{
			nil,
			newGroupId,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Insert(event_lib.EventMetaModel{
		EventId:      newEventId,
		EventType:    eventMeta.EventType,
		CreationTime: time.Now(),
		OwnerEmail:   email,
		GroupId:      newGroupId,
		//AccessEmails: []string{}// must not copy AccessEmails
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	eventRev.EventId = newEventId
	err = db.Insert(eventRev)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{
		Data: map[string]string{
			"eventType": eventRev.EventType,
			"eventId":   newEventId.Hex(),
			"sha1":      eventRev.Sha1,
		},
	}, nil
}

func SetEventGroupId(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}
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
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event", err)
	}
	newGroupIdHex, err := req.GetString("newGroupId")
	if err != nil {
		return nil, err
	}
	if !bson.IsObjectIdHex(*newGroupIdHex) {
		return nil, NewError(InvalidArgument, "invalid 'newGroupId'", nil)
		// to avoid panic!
	}
	newGroupId := bson.ObjectIdHex(*newGroupIdHex)
	newGroupModel := event_lib.EventGroupModel{
		Id: newGroupId,
	}
	err = db.Get(&newGroupModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
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
		GroupId: &[2]*bson.ObjectId{
			eventMeta.GroupId,
			&newGroupId,
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
		return nil, NewError(Internal, "", err)
	}
	/*userModel := UserModelByEmail(email, db)
	  if userModel == nil {
	      SetHttpErrorUserNotFound(w, email)
	      return
	  }*/
	eventMeta.GroupId = &newGroupId
	err = db.Update(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func GetEventOwner(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}
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
	return &Response{
		Data: scal.M{
			//"eventId": eventId.Hex(),
			"ownerEmail": eventMeta.OwnerEmail,
		},
	}, nil
}

func SetEventOwner(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}
	newOwnerEmail, err := req.GetString("newOwnerEmail")
	if err != nil {
		return nil, err
	}
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
	if eventMeta.OwnerEmail != email {
		return nil, ForbiddenError("you don't own this event", err)
	}
	// should check if user with `newOwnerEmail` exists?
	newOwnerUserModel := UserModelByEmail(*newOwnerEmail, db)
	if newOwnerUserModel == nil {
		return nil, NewError(
			InvalidArgument,
			fmt.Sprintf(
				"user with email '%s' does not exist",
				newOwnerEmail,
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
		return nil, NewError(Internal, "", err)
	}
	eventMeta.OwnerEmail = *newOwnerEmail
	err = db.Update(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	// send an E-Mail to `newOwnerEmail` FIXME
	return &Response{}, nil
}

func GetEventMetaModelFromRequest(req Request, email string) (*event_lib.EventMetaModel, error) {
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(InvalidArgument, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if !eventMeta.CanReadFull(email) {
		return nil, ForbiddenError("you can't meta information of this event", nil)
	}
	return eventMeta, nil
}

func GetEventMeta(req Request) (*Response, error) {
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
		return nil, NewError(Internal, "", err)
	}

	return &Response{
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

func GetEventAccess(req Request) (*Response, error) {
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
	return &Response{
		Data: scal.M{
			//"eventId": eventMeta.EventId.Hex(),
			"isPublic":       eventMeta.IsPublic,
			"accessEmails":   eventMeta.AccessEmails,
			"publicJoinOpen": eventMeta.PublicJoinOpen,
			"maxAttendees":   eventMeta.MaxAttendees,
		},
	}, nil
}

func SetEventAccess(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}

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
		return nil, NewError(Internal, "", err)
	}
	err = db.Update(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func AppendEventAccess(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Internal, "", err)
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
		return nil, NewError(Internal, "", err)
	}
	eventMeta.AccessEmails = newAccessEmails
	err = db.Update(eventMeta)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func joinEventWithToken(req Request, tokenStr string, eventId *bson.ObjectId) (*Response, error) {
	email, err := event_lib.CheckEventInvitationToken(tokenStr, eventId)
	if err != nil {
		return nil, ForbiddenError("invalid event invitation token", err)
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
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
		eventMeta.EventId.Hex(),
		values.Encode(),
	)
	fmt.Println("eventPath:", eventPath)
	return &Response{
		RedirectPath: eventPath,
	}, nil
}

func JoinEvent(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		tokenPtr, _ := req.GetString("token")
		if tokenPtr != nil {
			eventId, err := ObjectIdFromURL(req, "eventId", 1)
			if err != nil {
				return nil, err
			}
			return joinEventWithToken(req, *tokenPtr, eventId)
		}
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	// remoteIp, err := req.RemoteIP()
	// if err != nil {
	// 	return nil, err
	// }
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}

	if !userModel.EmailConfirmed {
		return nil, ForbiddenError("you need to confirm your email address before joining an event", nil)
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	err = eventMeta.Join(db, email)
	if err != nil {
		return nil, ForbiddenError(err.Error(), err)
	}
	return &Response{}, nil
}

func LeaveEvent(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
	}
	err = eventMeta.Leave(db, email)
	if err != nil {
		return nil, ForbiddenError(err.Error(), err)
	}
	return &Response{}, nil
}

func InviteToEvent(req Request) (*Response, error) {
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
	eventId, err := ObjectIdFromURL(req, "eventId", 1)
	if err != nil {
		return nil, err
	}

	inviteEmails, err := req.GetStringList("inviteEmails")
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, NewError(NotFound, "event not found", err)
		}
		return nil, NewError(Internal, "", err)
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
	return &Response{}, nil
}

func GetUngroupedEvents(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	type eventModel struct {
		EventId   bson.ObjectId `bson:"_id" json:"eventId"`
		EventType string        `bson:"eventType" json:"eventType"`
	}
	var events []eventModel
	err = db.FindAll(
		storage.C_eventMeta,
		scal.M{
			"ownerEmail": email,
			"groupId":    nil,
		},
		&events,
	)
	if events == nil {
		events = make([]eventModel, 0)
	}
	return &Response{
		Data: scal.M{
			"events": events,
		},
	}, nil
}

func GetMyEventList(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	type resultModel struct {
		EventId   bson.ObjectId `bson:"_id" json:"eventId"`
		EventType string        `bson:"eventType" json:"eventType"`
		//GroupId *bson.ObjectId    `bson:"groupId" json:"groupId"`
	}

	var results []resultModel
	err = db.FindAll(
		storage.C_eventMeta,
		scal.M{
			"ownerEmail": email,
		},
		&results,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	if results == nil {
		results = make([]resultModel, 0)
	}
	return &Response{
		Data: scal.M{
			"events": results,
		},
	}, nil
}

func GetMyEventsFull(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	pipeline := []scal.M{
		{"$match": scal.M{
			"ownerEmail": email,
		}},
		{"$lookup": scal.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$group": scal.M{
			"_id":       "$_id",
			"eventType": scal.M{"$first": "$eventType"},
			"groupId":   scal.M{"$first": "$groupId"},
			"meta": scal.M{
				"$first": scal.M{
					"ownerEmail":     "$ownerEmail",
					"isPublic":       "$isPublic",
					"creationTime":   "$creationTime",
					"accessEmails":   "$accessEmails",
					"publicJoinOpen": "$publicJoinOpen",
					"maxAttendees":   "$maxAttendees",
				},
			},
			"lastModifiedTime": scal.M{"$first": "$revision.time"},
			"lastSha1":         scal.M{"$first": "$revision.sha1"},
		}},
		{"$lookup": scal.M{
			"from":         storage.C_eventData,
			"localField":   "lastSha1",
			"foreignField": "sha1",
			"as":           "data",
		}},
		{"$unwind": "$data"},
	}

	results, err := event_lib.GetEventMetaPipeResults(db, &pipeline)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{
		Data: scal.M{
			"eventsFull": results,
		},
	}, nil
}

func GetMyLastCreatedEvents(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	maxCount, err := req.GetIntDefault("maxCount", 100)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	pipeline := []scal.M{
		{"$match": scal.M{
			"ownerEmail": email,
		}},
		{"$sort": scal.M{"creationTime": -1}},
		{"$limit": maxCount},
		{"$lookup": scal.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$group": scal.M{
			"_id":       "$_id",
			"eventType": scal.M{"$first": "$eventType"},
			"groupId":   scal.M{"$first": "$groupId"},
			"meta": scal.M{
				"$first": scal.M{
					"ownerEmail":     "$ownerEmail",
					"isPublic":       "$isPublic",
					"creationTime":   "$creationTime",
					"accessEmails":   "$accessEmails",
					"publicJoinOpen": "$publicJoinOpen",
					"maxAttendees":   "$maxAttendees",
				},
			},
			"lastModifiedTime": scal.M{"$first": "$revision.time"},
			"lastSha1":         scal.M{"$first": "$revision.sha1"},
		}},
		{"$lookup": scal.M{
			"from":         storage.C_eventData,
			"localField":   "lastSha1",
			"foreignField": "sha1",
			"as":           "data",
		}},
		{"$unwind": "$data"},
		{"$sort": scal.M{"meta.creationTime": -1}},
	}

	results, err := event_lib.GetEventMetaPipeResults(db, &pipeline)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{
		Data: scal.M{
			"maxCount":          maxCount,
			"lastCreatedEvents": results,
		},
	}, nil

}
