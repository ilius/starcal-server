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

	"gopkg.in/mgo.v2/bson"
	//"github.com/gorilla/mux"

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
				"POST",
				"{eventId}",
				authWrap(CopyEvent),
			},
		},
	})
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/ungrouped",
		Map: RouteMap{
			"GetUngroupedEvents": {
				"GET",
				"",
				authWrap(GetUngroupedEvents),
			},
		},
	})
	routeGroups = append(routeGroups, RouteGroup{
		Base: "event/my",
		Map: RouteMap{
			"GetMyEventList": {
				"GET",
				"events",
				authWrap(GetMyEventList),
			},
			"GetMyEventsFull": {
				"GET",
				"events-full",
				authWrap(GetMyEventsFull),
			},
			"GetMyLastCreatedEvents": {
				"GET",
				"last-created-events/{count}",
				authWrap(GetMyLastCreatedEvents),
			},
		},
	})
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
		SetHttpErrorInternal(w, err)
		return
	}
	err = db.Insert(event_lib.EventRevisionModel{
		EventId:   *eventId,
		EventType: eventMeta.EventType,
		Sha1:      "",
		Time:      now,
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	err = db.Remove(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func CopyEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	oldEventId := ObjectIdFromURL(w, r, "eventId", 0)
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	eventMeta, err := event_lib.LoadEventMetaModel(db, oldEventId, true)
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

	eventRev, err := event_lib.LoadLastRevisionModel(db, oldEventId)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
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
		SetHttpErrorInternal(w, err)
		return
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
		SetHttpErrorInternal(w, err)
		return
	}

	eventRev.EventId = newEventId
	err = db.Insert(eventRev)
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

func SetEventGroupId(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
	err = db.Get(&newGroupModel)
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
		SetHttpErrorInternal(w, err)
		return
	}

	/*userModel := UserModelByEmail(email, db)
	  if userModel == nil {
	      SetHttpErrorUserNotFound(w, email)
	      return
	  }*/
	eventMeta.GroupId = &newGroupId
	err = db.Update(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func GetEventOwner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
		if db.IsNotFound(err) {
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
	json.NewEncoder(w).Encode(scal.M{
		//"eventId": eventId.Hex(),
		"ownerEmail": eventMeta.OwnerEmail,
	})
}

func SetEventOwner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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

	newOwnerEmail, ok := inputMap["newOwnerEmail"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'newOwnerEmail'")
		return
	}
	// should check if user with `newOwnerEmail` exists?
	newOwnerUserModel := UserModelByEmail(newOwnerEmail, db)
	if newOwnerUserModel == nil {
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
	err = db.Insert(event_lib.EventMetaChangeLogModel{
		Time:     now,
		Email:    email,
		RemoteIp: remoteIp,
		EventId:  *eventId,
		FuncName: "SetEventOwner",
		OwnerEmail: &[2]*string{
			&eventMeta.OwnerEmail,
			&newOwnerEmail,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta.OwnerEmail = newOwnerEmail
	err = db.Update(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	// send an E-Mail to `newOwnerEmail` FIXME
	json.NewEncoder(w).Encode(scal.M{})
}

func GetEventMetaModelFromRequest(
	w http.ResponseWriter,
	r *http.Request,
	email string,
) *event_lib.EventMetaModel {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// -----------------------------------------------
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
		if db.IsNotFound(err) {
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

func GetEventMeta(w http.ResponseWriter, r *http.Request) {
	// includes owner, creation time, groupId, access info, attendings info
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventMeta := GetEventMetaModelFromRequest(w, r, email)
	if eventMeta == nil {
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{
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

func GetEventAccess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	eventMeta := GetEventMetaModelFromRequest(w, r, email)
	if eventMeta == nil {
		return
	}
	json.NewEncoder(w).Encode(scal.M{
		//"eventId": eventMeta.EventId.Hex(),
		"isPublic":       eventMeta.IsPublic,
		"accessEmails":   eventMeta.AccessEmails,
		"publicJoinOpen": eventMeta.PublicJoinOpen,
		"maxAttendees":   eventMeta.MaxAttendees,
	})
}

func SetEventAccess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
	if !reflect.DeepEqual(*newAccessEmails, eventMeta.AccessEmails) {
		metaChangeLog.AccessEmails = &[2][]string{
			eventMeta.AccessEmails,
			*newAccessEmails,
		}
		eventMeta.AccessEmails = *newAccessEmails
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
		SetHttpErrorInternal(w, err)
		return
	}
	err = db.Update(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func AppendEventAccess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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

	toAddEmail, ok := inputMap["toAddEmail"]
	if !ok {
		SetHttpError(w, http.StatusBadRequest, "missing 'toAddEmail'")
		return
	}
	newAccessEmails := append(eventMeta.AccessEmails, toAddEmail)
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
		SetHttpErrorInternal(w, err)
		return
	}
	eventMeta.AccessEmails = newAccessEmails
	err = db.Update(eventMeta)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func JoinEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
		if db.IsNotFound(err) {
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
	json.NewEncoder(w).Encode(scal.M{})
}

func LeaveEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
		if db.IsNotFound(err) {
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
	json.NewEncoder(w).Encode(scal.M{})
}

func InviteToEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
		InviteEmails *[]string `json:"inviteEmails"`
	}{
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
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusBadRequest, "event not found")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	err, errCode := eventMeta.Invite(
		db,
		email,
		inputStruct.InviteEmails,
		remoteIp,
		"http://"+r.Host, // FIXME
	)
	if err != nil {
		if errCode == scal.InternalServerError {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, errCode, err.Error())
		}
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func GetUngroupedEvents(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
	json.NewEncoder(w).Encode(scal.M{
		"events": events,
	})
}

func GetMyEventList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
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
	err = db.FindAll(
		storage.C_eventMeta,
		scal.M{
			"ownerEmail": email,
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
	json.NewEncoder(w).Encode(scal.M{
		"events": results,
	})
}

func GetMyEventsFull(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
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
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{
		"eventsFull": results,
	})
}

func GetMyLastCreatedEvents(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	parts := SplitURL(r.URL)
	if len(parts) < 2 {
		SetHttpErrorInternalMsg(w, fmt.Sprintf("Unexpected URL: %s", r.URL))
		return
	}
	countStr := parts[len(parts)-1] // int string
	// -----------------------------------------------
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

	pipeline := []scal.M{
		{"$match": scal.M{
			"ownerEmail": email,
		}},
		{"$sort": scal.M{"creationTime": -1}},
		{"$limit": count},
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
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{
		"maxCount":          count,
		"lastCreatedEvents": results,
	})
}
