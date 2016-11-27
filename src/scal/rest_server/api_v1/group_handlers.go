package api_v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
	//"gopkg.in/mgo.v2"
	//"github.com/gorilla/mux"

	"scal-lib/go-http-auth"
	"scal/event_lib"
	"scal/storage"
)

const ALLOW_DELETE_DEFAULT_GROUP = true

// time.RFC3339 == "2006-01-02T15:04:05Z07:00"

func init() {
	RegisterRoute(
		"GetGroupList",
		"GET",
		"/event/groups/",
		authenticator.Wrap(GetGroupList),
	)
	RegisterRoute(
		"AddGroup",
		"POST",
		"/event/groups/",
		authenticator.Wrap(AddGroup),
	)
	RegisterRoute(
		"UpdateGroup",
		"PUT",
		"/event/groups/{groupId}/",
		authenticator.Wrap(UpdateGroup),
	)
	RegisterRoute(
		"GetGroup",
		"GET",
		"/event/groups/{groupId}/",
		authenticator.Wrap(GetGroup),
	)
	RegisterRoute(
		"DeleteGroup",
		"DELETE",
		"/event/groups/{groupId}/",
		authenticator.Wrap(DeleteGroup),
	)
	RegisterRoute(
		"GetGroupEventList",
		"GET",
		"/event/groups/{groupId}/events/",
		authenticator.Wrap(GetGroupEventList),
	)
	RegisterRoute(
		"GetGroupEventsFull",
		"GET",
		"/event/groups/{groupId}/events-full/",
		authenticator.Wrap(GetGroupEventsFull),
	)
	RegisterRoute(
		"GetGroupModifiedEvents",
		"GET",
		"/event/groups/{groupId}/modified-events/{sinceDateTime}/",
		authenticator.Wrap(GetGroupModifiedEvents),
	)
	RegisterRoute(
		"GetGroupMovedEvents",
		"GET",
		"/event/groups/{groupId}/moved-events/{sinceDateTime}/",
		authenticator.Wrap(GetGroupMovedEvents),
	)
	RegisterRoute(
		"GetGroupLastCreatedEvents",
		"GET",
		"/event/groups/{groupId}/last-created-events/{count}/",
		authenticator.Wrap(GetGroupLastCreatedEvents),
	)
}

func GetGroupList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	type resultModel struct {
		GroupId    bson.ObjectId `bson:"_id" json:"groupId"`
		Title      string        `bson:"title" json:"title"`
		OwnerEmail string        `bson:"ownerEmail" json:"ownerEmail"`
	}
	var results []resultModel
	err = db.FindAll(
		storage.C_group,
		bson.M{
			"$or": []bson.M{
				bson.M{"ownerEmail": email},
				bson.M{"readAccessEmails": email},
			},
		},
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if results == nil {
		results = make([]resultModel, 0)
	}
	json.NewEncoder(w).Encode(bson.M{
		"groups": results,
	})
}

func AddGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	groupModel := event_lib.EventGroupModel{}

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &groupModel)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if groupModel.Id != "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			"can not specify 'groupId'",
		)
		return
	}
	if groupModel.OwnerEmail != "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			"can not specify 'ownerEmail'",
		)
		return
	}

	groupId := bson.NewObjectId()
	groupModel.Id = groupId
	groupModel.OwnerEmail = email
	err = db.Insert(groupModel)

	json.NewEncoder(w).Encode(map[string]string{
		"groupId": groupId.Hex(),
	})
}

func UpdateGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	groupId := ObjectIdFromURL(w, r, "groupId", 0)
	if groupId == nil {
		return
	}
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	newGroupModel := event_lib.EventGroupModel{}

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &newGroupModel)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if newGroupModel.Id != "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			"can not specify 'groupId'",
		)
		return
	}
	if newGroupModel.OwnerEmail != "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			"can not specify 'ownerEmail'",
		)
		return
	}
	if newGroupModel.Title == "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			"missing or empty 'title'",
		)
		return
	}

	oldGroupModel, err, internalErr := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	if oldGroupModel.OwnerEmail != email {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you don't have write access to this event group",
		)
		return
	}
	oldGroupModel.Title = newGroupModel.Title
	oldGroupModel.AddAccessEmails = newGroupModel.AddAccessEmails
	oldGroupModel.ReadAccessEmails = newGroupModel.ReadAccessEmails
	err = db.Update(oldGroupModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func GetGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	groupId := ObjectIdFromURL(w, r, "groupId", 0)
	if groupId == nil {
		return
	}
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	if !groupModel.CanRead(email) {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you don't have access to this event group",
		)
		return
	}
	json.NewEncoder(w).Encode(groupModel)
}

func DeleteGroup(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	groupId := ObjectIdFromURL(w, r, "groupId", 0)
	if groupId == nil {
		return
	}
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	if groupModel.OwnerEmail != email {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you are not allowed to delete this event group",
		)
		return
	}

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
		return
	}
	if userModel.DefaultGroupId != nil && *userModel.DefaultGroupId == *groupId {
		if !ALLOW_DELETE_DEFAULT_GROUP {
			SetHttpError(
				w,
				http.StatusForbidden,
				"you can not delete your default event group",
			)
			return
		}
		userModel.DefaultGroupId = nil
		err = db.Update(userModel)
		if err != nil {
			SetHttpErrorInternal(w, err)
			return
		}
	}

	eventMetaCol := db.C(storage.C_eventMeta)

	var eventMetaModels []event_lib.EventMetaModel
	err = db.FindAll(
		storage.C_eventMeta,
		bson.M{
			"groupId": groupId,
		},
		&eventMetaModels,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	for _, eventMetaModel := range eventMetaModels {
		if eventMetaModel.OwnerEmail != email {
			// send an Email to {eventMetaModel.OwnerEmail}
			// to inform the event owner, and let him move this
			// (ungrouped) event into his default (or any other) group
			// FIXME
		}
		// insert a new record to storage.C_eventMetaChangeLog // FIXME
		eventMetaModel.GroupId = nil
		err = eventMetaCol.Update(
			bson.M{"_id": eventMetaModel.EventId},
			eventMetaModel,
		)
		if err != nil {
			SetHttpErrorInternal(w, err)
			return
		}
		now := time.Now()
		err = db.C(storage.C_eventMetaChangeLog).Insert(
			bson.M{
				"time":     now,
				"email":    email,
				"remoteIp": remoteIp,
				"eventId":  eventMetaModel.EventId,
				"funcName": "DeleteGroup",
				"groupId": []interface{}{
					groupId,
					nil,
				},
			},
		)
		if err != nil {
			SetHttpErrorInternal(w, err)
			return
		}
	}
	err = db.Remove(groupModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func GetGroupEventList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	groupId := ObjectIdFromURL(w, r, "groupId", 1)
	if groupId == nil {
		return
	}
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}

	type resultModel struct {
		EventId   bson.ObjectId `bson:"_id" json:"eventId"`
		EventType string        `bson:"eventType" json:"eventType"`
		//OwnerEmail string         `bson:"ownerEmail" json:"ownerEmail"`
		//GroupId *bson.ObjectId    `bson:"groupId" json:"groupId"`
	}

	cond := groupModel.GetAccessCond(email)
	cond["groupId"] = groupId
	var results []resultModel
	err = db.FindAll(
		storage.C_eventMeta,
		cond,
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if results == nil {
		results = make([]resultModel, 0)
	}
	json.NewEncoder(w).Encode(bson.M{
		"events": results,
	})
}

func GetGroupEventsFull(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	groupId := ObjectIdFromURL(w, r, "groupId", 1)
	if groupId == nil {
		return
	}
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelById(
		"groupId",
		db,
		groupId,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"groupId": groupId,
		}},
	}
	aCond := groupModel.GetAccessCond(email)
	if len(aCond) > 0 {
		pipeline = append(pipeline, bson.M{"$match": aCond})
	}
	pipeline = append(pipeline, []bson.M{
		{"$lookup": bson.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$group": bson.M{
			"_id":       "$_id",
			"eventType": bson.M{"$first": "$eventType"},
			"meta": bson.M{
				"$first": bson.M{
					"ownerEmail":   "$ownerEmail",
					"isPublic":     "$isPublic",
					"creationTime": "$creationTime",
				},
			},
			"lastModifiedTime": bson.M{"$first": "$revision.time"},
			"lastSha1":         bson.M{"$first": "$revision.sha1"},
		}},
		{"$lookup": bson.M{
			"from":         storage.C_eventData,
			"localField":   "lastSha1",
			"foreignField": "sha1",
			"as":           "data",
		}},
		{"$unwind": "$data"},
	}...)
	var results []bson.M
	err = db.PipeAll(
		storage.C_eventMeta,
		pipeline,
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if results == nil {
		results = make([]bson.M, 0)
	}
	json.NewEncoder(w).Encode(bson.M{
		"events": results,
	})
}

func GetGroupModifiedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	//groupId := ObjectIdFromURL(w, r, "groupId", 2)
	//if groupId==nil { return }
	parts := SplitURL(r.URL)
	if len(parts) < 3 {
		SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
		return
	}
	groupIdHex := parts[len(parts)-3]
	sinceStr := parts[len(parts)-1] // datetime string
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if groupIdHex == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'groupId'")
		return
	}
	if !bson.IsObjectIdHex(groupIdHex) {
		SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
		return
		// to avoid panic!
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelByIdHex(
		"groupId",
		db,
		groupIdHex,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	groupId := groupModel.Id

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	//json.NewEncoder(w).Encode(bson.M{"sinceDateTime": since})

	pipeline := []bson.M{
		{"$match": bson.M{
			"groupId": groupId,
		}},
	}
	aCond := groupModel.GetAccessCond(email)
	if len(aCond) > 0 {
		pipeline = append(pipeline, bson.M{"$match": aCond})
	}
	pipeline = append(pipeline, []bson.M{
		{"$lookup": bson.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$match": bson.M{
			"revision.time": bson.M{
				"$gt": since,
			},
		}},
		{"$sort": bson.M{"revision.time": -1}},
		{"$group": bson.M{
			"_id":       "$_id",
			"eventType": bson.M{"$first": "$eventType"},
			"meta": bson.M{
				"$first": bson.M{
					"ownerEmail":   "$ownerEmail",
					"isPublic":     "$isPublic",
					"creationTime": "$creationTime",
				},
			},
			"lastModifiedTime": bson.M{"$first": "$revision.time"},
			"lastSha1":         bson.M{"$first": "$revision.sha1"},
		}},
		{"$lookup": bson.M{
			"from":         storage.C_eventData,
			"localField":   "lastSha1",
			"foreignField": "sha1",
			"as":           "data",
		}},
		{"$unwind": "$data"},
	}...)

	results := []bson.M{}
	err = db.PipeAll(
		storage.C_eventMeta,
		pipeline,
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		"groupId":         groupModel.Id,
		"since_datetime":  since,
		"modified_events": results,
	})

}

func GetGroupMovedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	//groupId := ObjectIdFromURL(w, r, "groupId", 2)
	//if groupId==nil { return }
	parts := SplitURL(r.URL)
	if len(parts) < 3 {
		SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
		return
	}
	groupIdHex := parts[len(parts)-3]
	sinceStr := parts[len(parts)-1] // datetime string
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if groupIdHex == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'groupId'")
		return
	}
	if !bson.IsObjectIdHex(groupIdHex) {
		SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
		return
		// to avoid panic!
	}

	groupModel, err, internalErr := event_lib.LoadGroupModelByIdHex(
		"groupId",
		db,
		groupIdHex,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	groupId := groupModel.Id

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"groupId": groupId,
		}},
		{"$match": bson.M{
			"time": bson.M{
				"$gt": since,
			},
		}},
		{"$sort": bson.M{"time": -1}},
	}
	accessPl := groupModel.GetLookupMetaAccessPipeline(
		email,
		"eventId", // localField for storage.C_eventMetaChangeLog
	)
	if len(accessPl) > 0 {
		pipeline = append(pipeline, accessPl...)
	}
	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":     "$eventId",
			"time":    bson.M{"$first": "$time"},
			"groupId": bson.M{"$first": "$groupId"},
		},
	})

	results := []bson.M{}
	err = db.PipeAll(
		storage.C_eventMetaChangeLog,
		pipeline,
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		"groupId":        groupModel.Id,
		"since_datetime": since,
		"moved_events":   results,
	})

}

func GetGroupLastCreatedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	//groupId := ObjectIdFromURL(w, r, "groupId", 2)
	//if groupId==nil { return }
	parts := SplitURL(r.URL)
	if len(parts) < 3 {
		SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
		return
	}
	groupIdHex := parts[len(parts)-3]
	countStr := parts[len(parts)-1] // int string
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if groupIdHex == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'groupId'")
		return
	}
	if !bson.IsObjectIdHex(groupIdHex) {
		SetHttpError(w, http.StatusBadRequest, "invalid 'groupId'")
		return
		// to avoid panic!
	}
	groupModel, err, internalErr := event_lib.LoadGroupModelByIdHex(
		"groupId",
		db,
		groupIdHex,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	groupId := groupModel.Id

	count, err := strconv.ParseInt(countStr, 10, 0)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, "invalid 'count', must be integer")
		return
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"groupId": groupId,
		}},
	}
	aCond := groupModel.GetAccessCond(email)
	if len(aCond) > 0 {
		pipeline = append(pipeline, bson.M{"$match": aCond})
	}
	pipeline = append(pipeline, []bson.M{
		{"$sort": bson.M{"creationTime": -1}},
		{"$limit": count},
		{"$lookup": bson.M{
			"from":         storage.C_revision,
			"localField":   "_id",
			"foreignField": "eventId",
			"as":           "revision",
		}},
		{"$unwind": "$revision"},
		{"$group": bson.M{
			"_id":       "$_id",
			"eventType": bson.M{"$first": "$eventType"},
			"meta": bson.M{
				"$first": bson.M{
					"ownerEmail":   "$ownerEmail",
					"isPublic":     "$isPublic",
					"creationTime": "$creationTime",
				},
			},
			"lastModifiedTime": bson.M{"$first": "$revision.time"},
			"lastSha1":         bson.M{"$first": "$revision.sha1"},
		}},
		{"$lookup": bson.M{
			"from":         storage.C_eventData,
			"localField":   "lastSha1",
			"foreignField": "sha1",
			"as":           "data",
		}},
		{"$unwind": "$data"},
		{"$sort": bson.M{"meta.creationTime": -1}},
	}...)

	results := []bson.M{}
	err = db.PipeAll(
		storage.C_eventMeta,
		pipeline,
		&results,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		"groupId":             groupModel.Id,
		"max_count":           count,
		"last_created_events": results,
	})

}
