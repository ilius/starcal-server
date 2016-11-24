package api_v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"github.com/gorilla/mux"

	"scal-lib/go-http-auth"
	"scal/event_lib"
	"scal/storage"
)

func init() {
	RegisterRoute(
		"CopyEvent",
		"POST",
		"/event/copy/",
		authenticator.Wrap(CopyEvent),
	)
	RegisterRoute(
		"GetUngroupedEvents",
		"GET",
		"/event/ungrouped/",
		authenticator.Wrap(GetUngroupedEvents),
	)
	RegisterRoute(
		"GetMyEventList",
		"GET",
		"/event/my/events/",
		authenticator.Wrap(GetMyEventList),
	)
	RegisterRoute(
		"GetMyEventsFull",
		"GET",
		"/event/my/events-full/",
		authenticator.Wrap(GetMyEventsFull),
	)
	RegisterRoute(
		"GetMyLastCreatedEvents",
		"GET",
		"/event/my/last-created-events/{count}/",
		authenticator.Wrap(GetMyLastCreatedEvents),
	)
}

func DeleteEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventId := ObjectIdFromURL(w, r, "eventId", 0)
	if eventId == nil {
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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
	now := time.Now()
	metaChangeLog := bson.M{
		"time":     now,
		"email":    email,
		"remoteIp": remoteIp,
		"eventId":  eventId,
		"funcName": "DeleteEvent",
		"ownerEmail": []interface{}{
			eventMeta.OwnerEmail,
			nil,
		},
	}
	if eventMeta.GroupId != nil {
		metaChangeLog["groupId"] = []interface{}{
			eventMeta.GroupId,
			nil,
		}
	}
	if len(eventMeta.AccessEmails) > 0 {
		metaChangeLog["accessEmails"] = []interface{}{
			eventMeta.AccessEmails,
			nil,
		}
	}
	err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	err = storage.Insert(db, event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventMeta.EventType,
		Sha1:      "",
		Time:      now,
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	err = storage.Remove(db, eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func CopyEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	var ok bool
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	inputMap := map[string]string{}
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputMap)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	oldEventIdHex, ok := inputMap["eventId"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'eventId'")
		return
	}
	if !bson.IsObjectIdHex(oldEventIdHex) {
		SetHttpError(w, http.StatusBadRequest, "invalid 'eventId'")
		return
		// to avoid panic!
	}
	oldEventId := bson.ObjectIdHex(oldEventIdHex)

	eventMeta, err := event_lib.LoadEventMetaModel(db, &oldEventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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

	eventRev, err := event_lib.LoadLastRevisionModel(db, &oldEventId)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}

	newEventId := bson.NewObjectId()

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
		return
	}

	newGroupId := userModel.DefaultGroupId
	if eventMeta.GroupModel != nil {
		if eventMeta.GroupModel.OwnerEmail == email {
			newGroupId = &eventMeta.GroupModel.Id // == eventMeta.GroupId
		}
	}

	now := time.Now()
	err = db.C(storage.C_eventMetaChangeLog).Insert(
		bson.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  newEventId,
			"funcName": "CopyEvent",
			"ownerEmail": []interface{}{
				nil,
				email,
			},
		},
		bson.M{
			"time":     now,
			"email":    email,
			"remoteIp": remoteIp,
			"eventId":  newEventId,
			"funcName": "CopyEvent",
			"groupId": []interface{}{
				nil,
				newGroupId,
			},
		},
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	err = storage.Insert(db, event_lib.EventMetaModel{
		EventId:      newEventId,
		EventType:    eventMeta.EventType,
		CreationTime: time.Now(),
		OwnerEmail:   email,
		GroupId:      newGroupId,
		//AccessEmails: []string{}// must not copy AccessEmails
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventRev.EventId = newEventId
	err = storage.Insert(db, eventRev)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"eventType": eventRev.EventType,
		"eventId":   newEventId.Hex(),
		"sha1":      eventRev.Sha1,
	})

}

func SetEventGroupId(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	var ok bool
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	inputMap := map[string]string{}
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputMap)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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

	newGroupIdHex, ok := inputMap["newGroupId"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'newGroupId'")
		return
	}
	if !bson.IsObjectIdHex(newGroupIdHex) {
		SetHttpError(w, http.StatusBadRequest, "invalid 'newGroupId'")
		return
		// to avoid panic!
	}
	newGroupId := bson.ObjectIdHex(newGroupIdHex)
	newGroupModel := event_lib.EventGroupModel{
		Id: newGroupId,
	}
	err = storage.Get(db, &newGroupModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	if !newGroupModel.EmailCanAdd(email) {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you don't have write access to this group",
		)
		return
	}

	now := time.Now()
	metaChangeLog := bson.M{
		"time":     now,
		"email":    email,
		"remoteIp": remoteIp,
		"eventId":  eventId,
		"funcName": "SetEventGroupId",
		"groupId": []interface{}{
			eventMeta.GroupId,
			newGroupId,
		},
	}
	/*
	   addedAccessEmails := Set(
	       eventMeta.GroupModel.ReadAccessEmails,
	   ).Difference(newGroupModel.ReadAccessEmails)
	   if addedAccessEmails {
	       metaChangeLog["addedAccessEmails"] = addedAccessEmails
	   }
	*/
	err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	/*userModel := UserModelByEmail(email, db)
	  if userModel == nil {
	      SetHttpErrorUserNotFound(w, email)
	      return
	  }*/
	eventMeta.GroupId = &newGroupId
	err = storage.Update(db, eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func GetEventOwner(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	if !eventMeta.CanRead(email) {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you don't have access to this event",
		)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		//"eventId": eventId.Hex(),
		"ownerEmail": eventMeta.OwnerEmail,
	})
}

func SetEventOwner(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	var ok bool
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	inputMap := map[string]string{}
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputMap)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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

	newOwnerEmail, ok := inputMap["newOwnerEmail"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'newOwnerEmail'")
		return
	}
	// should check if user with `newOwnerEmail` exists?
	userModel := UserModelByEmail(newOwnerEmail, db)
	if userModel == nil {
		SetHttpError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"user with email '%s' does not exist",
				newOwnerEmail,
			),
		)
		return
	}
	now := time.Now()
	err = db.C(storage.C_eventMetaChangeLog).Insert(bson.M{
		"time":     now,
		"email":    email,
		"remoteIp": remoteIp,
		"eventId":  eventId,
		"funcName": "SetEventOwner",
		"ownerEmail": []interface{}{
			eventMeta.OwnerEmail,
			newOwnerEmail,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta.OwnerEmail = newOwnerEmail
	err = storage.Update(db, eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	// send an E-Mail to `newOwnerEmail` FIXME
	json.NewEncoder(w).Encode(bson.M{})
}

func GetEventMetaModelFromRequest(
	w http.ResponseWriter,
	r *auth.AuthenticatedRequest,
) *event_lib.EventMetaModel {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return nil
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return nil
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return nil
	}
	if !eventMeta.CanReadFull(email) {
		SetHttpError(
			w,
			http.StatusForbidden,
			"you can't meta information of this event",
		)
		return nil
	}
	return eventMeta
}

func GetEventMeta(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	// includes owner, creation time, groupId, access info, attendings info
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta := GetEventMetaModelFromRequest(w, r)
	if eventMeta == nil {
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		//"eventId": eventMeta.EventId.Hex(),
		"ownerEmail":           eventMeta.OwnerEmail,
		"creationTime":         eventMeta.CreationTime,
		"groupId":              eventMeta.GroupIdHex(),
		"isPublic":             eventMeta.IsPublic,
		"accessEmails":         eventMeta.AccessEmails,
		"publicJoinOpen":       eventMeta.PublicJoinOpen,
		"maxAttendees":         eventMeta.MaxAttendees,
		"attendingEmails":      eventMeta.GetAttendingEmails(db),
		"notAttendingEmails":   eventMeta.GetNotAttendingEmails(db),
		"maybeAttendingEmails": eventMeta.GetMaybeAttendingEmails(db),
	})
}

func GetEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	eventMeta := GetEventMetaModelFromRequest(w, r)
	if eventMeta == nil {
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		//"eventId": eventMeta.EventId.Hex(),
		"isPublic":       eventMeta.IsPublic,
		"accessEmails":   eventMeta.AccessEmails,
		"publicJoinOpen": eventMeta.PublicJoinOpen,
		"maxAttendees":   eventMeta.MaxAttendees,
	})
}

func SetEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	inputStruct := struct {
		IsPublic       *bool     `json:"isPublic"`
		AccessEmails   *[]string `json:"accessEmails"`
		PublicJoinOpen *bool     `json:"publicJoinOpen"`
		MaxAttendees   *int      `json:"maxAttendees"`
	}{
		nil,
		nil,
		nil,
		nil,
	}

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputStruct)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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

	newIsPublic := inputStruct.IsPublic
	if newIsPublic == nil {
		SetHttpError(w, http.StatusBadRequest, "missing 'isPublic'")
		return
	}
	newAccessEmails := inputStruct.AccessEmails
	if newAccessEmails == nil {
		SetHttpError(w, http.StatusBadRequest, "missing 'accessEmails'")
		return
	}
	newPublicJoinOpen := inputStruct.PublicJoinOpen
	if newPublicJoinOpen == nil {
		SetHttpError(w, http.StatusBadRequest, "missing 'publicJoinOpen'")
		return
	}
	newMaxAttendees := inputStruct.MaxAttendees
	if newMaxAttendees == nil {
		SetHttpError(w, http.StatusBadRequest, "missing 'maxAttendees'")
		return
	}

	now := time.Now()
	metaChangeLog := bson.M{
		"time":     now,
		"email":    email,
		"remoteIp": remoteIp,
		"eventId":  eventId,
		"funcName": "SetEventAccess",
	}
	if *newIsPublic != eventMeta.IsPublic {
		metaChangeLog["isPublic"] = []interface{}{
			eventMeta.IsPublic,
			newIsPublic,
		}
		eventMeta.IsPublic = *newIsPublic
	}
	if !reflect.DeepEqual(*newAccessEmails, eventMeta.AccessEmails) {
		metaChangeLog["accessEmails"] = []interface{}{
			eventMeta.AccessEmails,
			newAccessEmails,
		}
		eventMeta.AccessEmails = *newAccessEmails
	}
	if *newPublicJoinOpen != eventMeta.PublicJoinOpen {
		metaChangeLog["publicJoinOpen"] = []interface{}{
			eventMeta.PublicJoinOpen,
			newPublicJoinOpen,
		}
		eventMeta.PublicJoinOpen = *newPublicJoinOpen

	}
	if *newMaxAttendees != eventMeta.MaxAttendees {
		metaChangeLog["maxAttendees"] = []interface{}{
			eventMeta.MaxAttendees,
			newMaxAttendees,
		}
		eventMeta.MaxAttendees = *newMaxAttendees
	}
	err = db.C(storage.C_eventMetaChangeLog).Insert(metaChangeLog)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	err = storage.Update(db, eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func AppendEventAccess(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	var ok bool
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	inputMap := map[string]string{}
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputMap)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
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

	toAddEmail, ok := inputMap["toAddEmail"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'toAddEmail'")
		return
	}
	newAccessEmails := append(eventMeta.AccessEmails, toAddEmail)
	now := time.Now()
	err = db.C(storage.C_eventMetaChangeLog).Insert(bson.M{
		"time":     now,
		"email":    email,
		"remoteIp": remoteIp,
		"eventId":  eventId,
		"funcName": "AppendEventAccess",
		"accessEmails": []interface{}{
			eventMeta.AccessEmails,
			newAccessEmails,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta.AccessEmails = newAccessEmails
	err = storage.Update(db, eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func JoinEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	/*remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	  if err != nil {
	      SetHttpErrorInternal(w, err)
	      return
	  }*/
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	err = eventMeta.Join(db, email)
	if err != nil {
		SetHttpError(w, http.StatusForbidden, err.Error())
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func LeaveEvent(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	/*remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	  if err != nil {
	      SetHttpErrorInternal(w, err)
	      return
	  }*/
	eventId := ObjectIdFromURL(w, r, "eventId", 1)
	if eventId == nil {
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta, err := event_lib.LoadEventMetaModel(db, eventId, true)
	if err != nil {
		if err == mgo.ErrNotFound {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	err = eventMeta.Leave(db, email)
	if err != nil {
		SetHttpError(w, http.StatusForbidden, err.Error())
		return
	}
	json.NewEncoder(w).Encode(bson.M{})
}

func GetUngroupedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	type eventModel struct {
		EventId   bson.ObjectId `bson:"_id" json:"eventId"`
		EventType string        `bson:"eventType" json:"eventType"`
	}
	var events []eventModel
	err = db.C(storage.C_eventMeta).Find(bson.M{
		"ownerEmail": email,
		"groupId":    nil,
	}).All(&events)
	if events == nil {
		events = make([]eventModel, 0)
	}
	json.NewEncoder(w).Encode(bson.M{
		"events": events,
	})
}

func GetMyEventList(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	type resultModel struct {
		EventId   bson.ObjectId `bson:"_id" json:"eventId"`
		EventType string        `bson:"eventType" json:"eventType"`
		//GroupId *bson.ObjectId    `bson:"groupId" json:"groupId"`
	}

	var results []resultModel
	err = db.C(storage.C_eventMeta).Find(bson.M{
		"ownerEmail": email,
	}).All(&results)
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

func GetMyEventsFull(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"ownerEmail": email,
		}},
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
			"groupId":   bson.M{"$first": "$groupId"},
			"meta": bson.M{
				"$first": bson.M{
					"ownerEmail":     "$ownerEmail",
					"isPublic":       "$isPublic",
					"creationTime":   "$creationTime",
					"accessEmails":   "$accessEmails",
					"publicJoinOpen": "$publicJoinOpen",
					"maxAttendees":   "$maxAttendees",
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
	}

	results := []bson.M{}
	err = db.C(storage.C_eventMeta).Pipe(pipeline).All(&results)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		"events_full": results,
	})
}

func GetMyLastCreatedEvents(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	parts := SplitURL(r.URL)
	if len(parts) < 2 {
		SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
		return
	}
	countStr := parts[len(parts)-1] // int string
	// -----------------------------------------------
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	count, err := strconv.ParseInt(countStr, 10, 0)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, "invalid 'count', must be integer")
		return
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"ownerEmail": email,
		}},
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
			"groupId":   bson.M{"$first": "$groupId"},
			"meta": bson.M{
				"$first": bson.M{
					"ownerEmail":     "$ownerEmail",
					"isPublic":       "$isPublic",
					"creationTime":   "$creationTime",
					"accessEmails":   "$accessEmails",
					"publicJoinOpen": "$publicJoinOpen",
					"maxAttendees":   "$maxAttendees",
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
	}

	results := []bson.M{}
	err = db.C(storage.C_eventMeta).Pipe(pipeline).All(&results)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(bson.M{
		"max_count":           count,
		"last_created_events": results,
	})
}