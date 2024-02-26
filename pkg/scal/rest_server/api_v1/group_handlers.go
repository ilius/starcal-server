package api_v1

import (
	"fmt"
	"time"

	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/event_lib"
	"github.com/ilius/starcal-server/pkg/scal/storage"

	"github.com/ilius/mgo/bson"
	. "github.com/ilius/ripo"
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

	cond := db.NewCondition(storage.OR).
		Equals("ownerEmail", email).
		Includes("readAccessEmails", email)

	var results []event_lib.ListGroupsRow
	err = db.FindAll(&results, &storage.FindInput{
		Collection: storage.C_group,
		Condition:  cond,
		SortBy:     "_id",
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

	groupId := bson.NewObjectId().Hex()
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
			"groupId": groupId,
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

	cond := db.NewCondition(storage.AND)
	cond.IdEquals("groupId", *groupId)

	var eventMetaModels []event_lib.EventMetaModel
	err = db.FindAll(&eventMetaModels, &storage.FindInput{
		Collection: storage.C_eventMeta,
		Condition:  cond,
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
			GroupId: &[2]*string{
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
	fmt.Println("groupModel.Id:", groupModel.Id)
	fmt.Println("groupId:", *groupId)

	cond := db.NewCondition(storage.AND)
	cond.IdEquals("groupId", *groupId)
	if !groupModel.CanRead(email) {
		cond.NewSubCondition(storage.OR).
			Equals("ownerEmail", email).
			Equals("isPublic", true).
			Includes("accessEmails", email)
	}
	cond.SetPageOptions(pageOpts)
	{
		// b, _ := json.MarshalIndent(cond.Prepare(), "", "    ")
		// fmt.Println(string(b))
		fmt.Println(cond.Prepare())
	}

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

	pipeline := NewPipelines(db, storage.C_eventMeta)
	pipeline.MatchValue("groupId", groupId)
	if !groupModel.CanRead(email) {
		pipeline.AddEventGroupAccess(email)
	}
	pipeline.SetPageOptions(pageOpts)
	pipeline.Lookup(storage.C_revision, "_id", "eventId", "revision")
	pipeline.Unwind("revision")

	pipeline.GroupBy("_id").
		AddFromFirst("eventType", "eventType").
		AddFromFirst("revision.sha1", "lastSha1")

	results, err := GetEventMetaPipeResults(db, pipeline, nil)
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

	pipeline := NewPipelines(db, storage.C_eventMeta)
	pipeline.MatchValue("groupId", groupId)
	if !groupModel.CanRead(email) {
		pipeline.AddEventGroupAccess(email)
	}
	pipeline.Lookup(storage.C_revision, "_id", "eventId", "revision")
	pipeline.Unwind("revision")
	pipeline.NewMatchGreaterThan("revision.time", since)
	pipeline.Sort("revision.time", false)
	pipeline.AppendLimit(limit)
	pipeline.GroupBy("_id").
		AddFromFirst("eventType", "eventType").
		AddFromFirst("revision.sha1", "lastSha1").
		AddFromFirst("revision.time", "lastModifiedTime").
		AddFromFirst("ownerEmail", "ownerEmail").
		AddFromFirst("isPublic", "isPublic").
		AddFromFirst("creationTime", "creationTime")

	pipeline.Sort("lastModifiedTime", false)
	pipeline.Lookup(storage.C_eventData, "lastSha1", "sha1", "data")
	pipeline.Unwind("data")

	results, err := GetEventMetaPipeResults(db, pipeline, []string{
		"ownerEmail",
		"isPublic",
		"creationTime",
	})
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

	pipeline := NewPipelines(db, storage.C_eventMetaChangeLog)
	pipeline.MatchValue("groupId", groupId)
	pipeline.MatchGreaterThan("time", since)
	pipeline.Sort("time", false)
	pipeline.Limit(limit)
	if !groupModel.CanRead(email) {
		pipeline.AddEventLookupMetaAccess(
			email,
			"eventId", // localField for storage.C_eventMetaChangeLog
		)
	}
	pipeline.GroupBy("eventId").
		AddFromFirst("time", "time").
		AddFromLast("groupId", "oldGroupItem"). // want the index 0 of {old, new}
		AddFromFirst("groupId", "newGroupItem") // want the index 1 of {old, new}

	pipeline.Sort("time", false)

	results := []event_lib.MovedEventsRow{}
	err = pipeline.All(&results)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	// set OldGroupId and NewGroupId, while also converting nil ObjectId values to empty strings
	for i := range len(results) {
		results[i].OldGroupId = storage.Hex(results[i].OldGroupItem[0])
		results[i].NewGroupId = storage.Hex(results[i].NewGroupItem[1])
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
