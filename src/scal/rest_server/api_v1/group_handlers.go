package api_v1

import (
	"time"

	"github.com/globalsign/mgo/bson"
	. "github.com/ilius/ripo"

	"scal"
	"scal/event_lib"
	"scal/storage"
)

const ALLOW_DELETE_DEFAULT_GROUP = true

// time.RFC3339 == "2006-01-02T15:04:05Z07:00"

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/groups",
		Map: RouteMap{
			"GetGroupList": {
				Method:  "GET",
				Pattern: "",
				Handler: GetGroupList,
			},
			"AddGroup": {
				Method:  "POST",
				Pattern: "",
				Handler: AddGroup,
			},
			"UpdateGroup": {
				Method:  "PUT",
				Pattern: ":groupId",
				Handler: UpdateGroup,
			},
			"GetGroup": {
				Method:  "GET",
				Pattern: ":groupId",
				Handler: GetGroup,
			},
			"DeleteGroup": {
				Method:  "DELETE",
				Pattern: ":groupId",
				Handler: DeleteGroup,
			},
			"GetGroupEventList": {
				Method:  "GET",
				Pattern: ":groupId/events",
				Handler: GetGroupEventList,
			},
			"GetGroupEventListWithSha1": {
				Method:  "GET",
				Pattern: ":groupId/events-sha1",
				Handler: GetGroupEventListWithSha1,
			},
			"GetGroupModifiedEvents": {
				Method:  "GET",
				Pattern: ":groupId/modified-events/:sinceDateTime",
				Handler: GetGroupModifiedEvents,
			},
			"GetGroupMovedEvents": {
				Method:  "GET",
				Pattern: ":groupId/moved-events/:sinceDateTime",
				Handler: GetGroupMovedEvents,
			},
		},
	})
}

func GetGroupList(req Request) (*Response, error) {
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

	var results []event_lib.ListGroupsRow
	err = db.FindAll(&results, &storage.FindInput{
		Collection: storage.C_group,
		Conditions: scal.M{
			"$or": []scal.M{
				{"ownerEmail": email},
				{"readAccessEmails": email},
			},
		},
		SortBy: "_id",
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	if results == nil {
		results = make([]event_lib.ListGroupsRow, 0)
	}
	return &Response{
		Data: scal.M{
			"groups": results,
		},
	}, nil
}

func AddGroup(req Request) (*Response, error) {
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

	title, err := req.GetString("title")
	if err != nil {
		return nil, err
	}

	addAccessEmails, _ := req.GetStringList("addAccessEmails", FromBody, FromForm, FromEmpty)
	readAccessEmails, _ := req.GetStringList("readAccessEmails", FromBody, FromForm, FromEmpty)

	groupModel := event_lib.EventGroupModel{}

	groupId := bson.NewObjectId()
	groupModel.Id = groupId
	groupModel.OwnerEmail = email
	groupModel.Title = *title
	groupModel.AddAccessEmails = addAccessEmails
	groupModel.ReadAccessEmails = readAccessEmails

	err = db.Insert(groupModel)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: map[string]string{
			"groupId": groupId.Hex(),
		},
	}, nil
}

func UpdateGroup(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	groupId, err := ObjectIdFromRequest(req, "groupId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Group(*groupId)
	if failed {
		return nil, NewError(ResourceLocked, "group is locked by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	newTitle, err := req.GetString("title")
	if err != nil {
		return nil, err
	}

	newAddAccessEmails, _ := req.GetStringList("addAccessEmails", FromBody, FromForm, FromEmpty)
	newReadAccessEmails, _ := req.GetStringList("readAccessEmails", FromBody, FromForm, FromEmpty)

	groupModel, err := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		return nil, err
	}
	if groupModel.OwnerEmail != email {
		return nil, ForbiddenError("you don't have write access to this event group", nil)
	}
	groupModel.Title = *newTitle
	groupModel.AddAccessEmails = newAddAccessEmails
	groupModel.ReadAccessEmails = newReadAccessEmails
	err = db.Update(groupModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func GetGroup(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	groupId, err := ObjectIdFromRequest(req, "groupId")
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	groupModel, err := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		return nil, err
	}
	if !groupModel.CanRead(email) {
		return nil, ForbiddenError("you don't have access to this event group", nil)
	}
	return &Response{
		Data: groupModel,
	}, nil
}

func DeleteGroup(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	groupId, err := ObjectIdFromRequest(req, "groupId")
	if err != nil {
		return nil, err
	}
	failed, unlock := resLock.Group(*groupId)
	if failed {
		return nil, NewError(ResourceLocked, "group is locked by another request", nil)
	}
	defer unlock()
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	groupModel, err := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		return nil, err
	}
	if groupModel.OwnerEmail != email {
		return nil, ForbiddenError("you are not allowed to delete this event group", nil)
	}

	if userModel.DefaultGroupId != nil && *userModel.DefaultGroupId == *groupId {
		if !ALLOW_DELETE_DEFAULT_GROUP {
			return nil, ForbiddenError("you can not delete your default event group", nil)
		}
		userModel.DefaultGroupId = nil
		err = db.Update(userModel)
		if err != nil {
			return nil, NewError(Internal, "", err)
		}
	}

	var eventMetaModels []event_lib.EventMetaModel
	err = db.FindAll(&eventMetaModels, &storage.FindInput{
		Collection: storage.C_eventMeta,
		Conditions: scal.M{
			"groupId": groupId,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	for _, eventMetaModel := range eventMetaModels {
		if eventMetaModel.OwnerEmail != email {
			// send an Email to {eventMetaModel.OwnerEmail}
			// to inform the event owner, and let him move this
			// (ungrouped) event into his default (or any other) group
			// FIXME
		}
		now := time.Now()
		err = db.Insert(event_lib.EventMetaChangeLogModel{
			Time:     now,
			Email:    email,
			RemoteIp: remoteIp,
			EventId:  eventMetaModel.EventId,
			FuncName: "DeleteGroup",
			GroupId: &[2]*bson.ObjectId{
				groupId,
				nil,
			},
		})
		if err != nil {
			return nil, NewError(Internal, "", err)
		}
		eventMetaModel.GroupId = nil
		err = db.Update(eventMetaModel)
		if err != nil {
			return nil, NewError(Internal, "", err)
		}
	}
	err = db.Remove(groupModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func GetGroupEventList(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	groupId, err := ObjectIdFromRequest(req, "groupId")
	if err != nil {
		return nil, err
	}
	pageOpts, err := GetPageOptions(req)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	// -----------------------------------------------
	groupModel, err := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		return nil, err
	}

	cond := groupModel.GetAccessCond(email)
	cond["groupId"] = groupId
	pageOpts.AddStartIdCond(cond)

	var results []*event_lib.ListEventsRow
	err = db.FindAll(&results, &storage.FindInput{
		Collection: storage.C_eventMeta,
		Conditions: cond,
		SortBy:     pageOpts.SortBy(),
		Limit:      pageOpts.Limit,
		Fields: []string{
			"_id",
			"eventType",
			// "ownerEmail",
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
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
	return &Response{
		Data: output,
	}, nil
}

func GetGroupEventListWithSha1(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	groupId, err := ObjectIdFromRequest(req, "groupId")
	if err != nil {
		return nil, err
	}
	pageOpts, err := GetPageOptions(req)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	groupModel, err := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		return nil, err
	}

	cond := groupModel.GetAccessCond(email)
	cond["groupId"] = groupId
	pageOpts.AddStartIdCond(cond)

	sortMap := pageOpts.SortByMap()

	pipeline := []scal.M{
		{"$match": cond},
		sortMap,
		{"$limit": pageOpts.Limit},
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
			"lastSha1":  scal.M{"$first": "$revision.sha1"},
		}},
		sortMap,
	}

	results, err := event_lib.GetEventMetaPipeResults(db, &pipeline)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	output := scal.M{
		"events": results,
	}
	if len(results) > 0 {
		output["lastId"] = results[len(results)-1]["eventId"]
	}
	return &Response{
		Data: output,
	}, nil
}

func GetGroupModifiedEvents(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	// groupId, err := ObjectIdFromRequest(req, "groupId")
	// if err != nil {
	// 	return nil, err
	// }
	// if groupId==nil { return }
	groupIdHex, err := req.GetString("groupId")
	if err != nil {
		return nil, err
	}
	since, err := req.GetTime("sinceDateTime")
	if err != nil {
		return nil, err
	}
	limit, err := GetPageLimit(req)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	if !bson.IsObjectIdHex(*groupIdHex) {
		return nil, NewError(InvalidArgument, "invalid 'groupId'", nil)
		// to avoid panic!
	}
	groupModel, err := event_lib.LoadGroupModelByIdHex(
		"groupId",
		db,
		*groupIdHex,
	)
	if err != nil {
		return nil, err
	}
	groupId := groupModel.Id

	cond := groupModel.GetAccessCond(email)
	cond["groupId"] = groupId
	pipeline := []scal.M{
		{"$match": cond},
		{"$lookup": scal.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$match": scal.M{
			"revision.time": scal.M{
				"$gt": since,
			},
		}},
		{"$sort": scal.M{"revision.time": -1}},
		{"$limit": limit},
		{"$group": scal.M{
			"_id":       "$_id",
			"eventType": scal.M{"$first": "$eventType"},
			"meta": scal.M{
				"$first": scal.M{
					"ownerEmail":   "$ownerEmail",
					"isPublic":     "$isPublic",
					"creationTime": "$creationTime",
				},
			},
			"lastModifiedTime": scal.M{"$first": "$revision.time"},
			"lastSha1":         scal.M{"$first": "$revision.sha1"},
		}},
		{"$sort": scal.M{"lastModifiedTime": -1}},
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

	output := scal.M{
		"groupId":        groupModel.Id,
		"sinceDatetime":  since,
		"limit":          limit,
		"modifiedEvents": results,
	}
	return &Response{
		Data: output,
	}, nil
}

func GetGroupMovedEvents(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	// groupId, err := ObjectIdFromRequest(req, "groupId)
	// if err != nil {
	// 	return nil, err
	// }
	// if groupId==nil { return }
	groupIdHex, err := req.GetString("groupId")
	if err != nil {
		return nil, err
	}
	since, err := req.GetTime("sinceDateTime")
	if err != nil {
		return nil, err
	}
	limit, err := GetPageLimit(req)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	if *groupIdHex == "" {
		return nil, NewError(MissingArgument, "missing 'groupId'", nil)
	}
	if !bson.IsObjectIdHex(*groupIdHex) {
		return nil, NewError(InvalidArgument, "invalid 'groupId'", nil)
		// to avoid panic!
	}

	groupModel, err := event_lib.LoadGroupModelByIdHex(
		"groupId",
		db,
		*groupIdHex,
	)
	if err != nil {
		return nil, err
	}
	groupId := groupModel.Id

	pipeline := []scal.M{
		{"$match": scal.M{
			"groupId": groupId,
			"time": scal.M{
				"$gt": since,
			},
		}},
		{"$sort": scal.M{"time": -1}},
		{"$limit": limit},
	}
	accessPipeline := groupModel.GetLookupMetaAccessPipeline(
		email,
		"eventId", // localField for storage.C_eventMetaChangeLog
	)
	if len(accessPipeline) > 0 {
		pipeline = append(pipeline, accessPipeline...)
	}
	pipeline = append(pipeline, scal.M{
		"$group": scal.M{
			"_id":  "$eventId",
			"time": scal.M{"$first": "$time"},
			"oldGroupId": scal.M{"$last": scal.M{
				"$arrayElemAt": []interface{}{"$groupId", 0},
			}},
			"newGroupId": scal.M{"$first": scal.M{
				"$arrayElemAt": []interface{}{"$groupId", 1},
			}},
		},
	})
	pipeline = append(pipeline, scal.M{"$sort": scal.M{"time": -1}})

	results := []event_lib.MovedEventsRow{}
	err = storage.PipeAll(db,
		storage.C_eventMetaChangeLog,
		&pipeline,
		&results,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	// convert nil values to empty strings
	for i := 0; i < len(results); i++ {
		results[i].OldGroupId = storage.Hex(results[i].OldGroupId)
		results[i].NewGroupId = storage.Hex(results[i].NewGroupId)
	}

	return &Response{
		Data: scal.M{
			"groupId":       groupModel.Id,
			"sinceDatetime": since,
			"limit":         limit,
			"movedEvents":   results,
		},
	}, nil
}
