package api_v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"

	"scal"
	"scal/event_lib"
	//"scal/settings"
	"scal/storage"
	. "scal/user_lib"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "user",
		Map: RouteMap{
			"SetUserFullName": {
				"PUT",
				"full-name",
				authWrap(SetUserFullName),
			},
			"UnsetUserFullName": {
				"DELETE",
				"full-name",
				authWrap(UnsetUserFullName),
			},
			"GetUserInfo": {
				"GET",
				"info",
				authWrap(GetUserInfo),
			},
			"SetUserDefaultGroupId": {
				"PUT",
				"default-group",
				authWrap(SetUserDefaultGroupId),
			},
			"UnsetUserDefaultGroupId": {
				"DELETE",
				"default-group",
				authWrap(UnsetUserDefaultGroupId),
			},
		},
	})
}

func SetUserAttrInput(
	w http.ResponseWriter,
	email string,
	body []byte,
	attrName string,
) string {
	var err error
	attrMap := map[string]string{
		attrName: "",
	}
	err = json.Unmarshal(body, &attrMap)
	if err != nil {
		SetHttpError(w, http.StatusBadRequest, err.Error())
		return ""
	}
	attrValue, ok := attrMap[attrName]
	if !ok || attrValue == "" {
		SetHttpError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("missing '%s'", attrName),
		)
		return ""
	}
	//fmt.Println("attrValue =", attrValue)
	return attrValue
}

func SetUserFullName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	const attrName = "fullName"
	// -----------------------------------------------
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	attrValue := SetUserAttrInput(
		w,
		email,
		body,
		attrName,
	)
	if attrValue == "" {
		return
	}

	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "SetUserFullName",
		FullName: &[2]*string{
			&userModel.FullName,
			&attrValue,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel.FullName = attrValue
	err = db.Update(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(scal.M{})
}

func UnsetUserFullName(w http.ResponseWriter, r *http.Request) {
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

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "UnsetUserFullName",
		FullName: &[2]*string{
			&userModel.FullName,
			nil,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel.FullName = ""
	err = db.Update(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(scal.M{})
}

func SetUserDefaultGroupId(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	const attrName = "defaultGroupId"
	// -----------------------------------------------
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	attrValue := SetUserAttrInput(
		w,
		email,
		body,
		attrName,
	)
	if attrValue == "" {
		return
	}

	groupModel, err, internalErr := event_lib.LoadGroupModelByIdHex(
		"defaultGroupId",
		db,
		attrValue,
	)
	if err != nil {
		if internalErr {
			SetHttpErrorInternal(w, err)
		} else {
			SetHttpError(w, http.StatusBadRequest, err.Error())
		}
	}
	groupId := groupModel.Id

	if groupModel.OwnerEmail != email {
		SetHttpError(w, http.StatusBadRequest, "invalid 'defaultGroupId'")
		return
	}

	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "SetUserDefaultGroupId",
		DefaultGroupId: &[2]*bson.ObjectId{
			userModel.DefaultGroupId,
			&groupId,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel.DefaultGroupId = &groupId
	err = db.Update(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(scal.M{})
}

func UnsetUserDefaultGroupId(w http.ResponseWriter, r *http.Request) {
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

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "UnsetUserDefaultGroupId",
		DefaultGroupId: &[2]*bson.ObjectId{
			userModel.DefaultGroupId,
			nil,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel.DefaultGroupId = nil
	err = db.Update(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	json.NewEncoder(w).Encode(scal.M{})
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email
	// -----------------------------------------------
	json.NewEncoder(w).Encode(scal.M{
		"email":          email,
		"fullName":       userModel.FullName,
		"defaultGroupId": userModel.DefaultGroupId,
		//"locked": userModel.Locked,
	})
}
