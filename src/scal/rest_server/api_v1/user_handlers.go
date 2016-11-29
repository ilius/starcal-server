package api_v1

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"

	"scal"
	"scal-lib/go-http-auth"
	"scal/event_lib"
	"scal/storage"
)

const REALM = "starcalendar.net"

var globalDb, globalDbErr = storage.GetDB()

var authenticator = auth.NewDigestAuthenticator(REALM, Secret)

func init() {
	if globalDbErr != nil {
		panic(globalDbErr)
	}
	RegisterRoute(
		"RegisterUser",
		"POST",
		"/user/register/",
		RegisterUser,
	)
	RegisterRoute(
		"SetUserFullName",
		"PUT",
		"/user/full-name/",
		authenticator.Wrap(SetUserFullName),
	)
	RegisterRoute(
		"UnsetUserFullName",
		"DELETE",
		"/user/full-name/",
		authenticator.Wrap(UnsetUserFullName),
	)
	RegisterRoute(
		"GetUserInfo",
		"GET",
		"/user/info/",
		authenticator.Wrap(GetUserInfo),
	)
	RegisterRoute(
		"SetUserDefaultGroupId",
		"PUT",
		"/user/default-group-id/",
		authenticator.Wrap(SetUserDefaultGroupId),
	)
	RegisterRoute(
		"UnsetUserDefaultGroupId",
		"DELETE",
		"/user/default-group-id/",
		authenticator.Wrap(UnsetUserDefaultGroupId),
	)
}

//type Request http.Request
/*
type Request auth.AuthenticatedRequest
func (r Request) Email() string {
    return r.Username
}
*/

func Secret(email string, realm string) string {
	userModel := UserModelByEmail(email, globalDb)
	if userModel == nil {
		return ""
	}
	if userModel.Locked {
		return "" // FIXME
	}
	return userModel.Password
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userModel := UserModel{}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &userModel)
	if err != nil {
		msg := err.Error()
		//if strings.Contains(msg, "") {
		//    msg = ""
		//}
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	if userModel.Email == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'email'")
		return
	}
	if userModel.Password == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'password'")
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	anotherUserModel := UserModelByEmail(userModel.Email, db)
	if anotherUserModel != nil {
		SetHttpError(w, http.StatusBadRequest, "duplicate 'email'")
		return
	}
	userModel.Password = fmt.Sprintf(
		"%x",
		md5.Sum(
			[]byte(
				fmt.Sprintf(
					"%s:%s:%s",
					userModel.Email,
					REALM,
					userModel.Password,
				),
			),
		),
	)
	defaultGroup := event_lib.EventGroupModel{
		Id:         bson.NewObjectId(),
		Title:      userModel.Email,
		OwnerEmail: userModel.Email,
	}
	err = db.Insert(defaultGroup)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	userModel.DefaultGroupId = &defaultGroup.Id
	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: "", // FIXME
		RemoteIp:     remoteIp,
		FuncName:     "RegisterUser",
		Email: &[2]*string{
			nil,
			&userModel.Email,
		},
		DefaultGroupId: &[2]*bson.ObjectId{
			nil,
			userModel.DefaultGroupId,
		},
		//FullName: &[2]*string{
		//    nil
		//    &userModel.FullName,
		//},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	err = db.Insert(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
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

func SetUserFullName(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	const attrName = "fullName"
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
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

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
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

func UnsetUserFullName(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
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

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
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

func SetUserDefaultGroupId(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	const attrName = "defaultGroupId"
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
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

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
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

func UnsetUserDefaultGroupId(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
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

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
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

func GetUserInfo(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	defer r.Body.Close()
	email := r.Username
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpErrorUserNotFound(w, email)
		return
	}

	json.NewEncoder(w).Encode(scal.M{
		"email":          userModel.Email,
		"fullName":       userModel.FullName,
		"defaultGroupId": userModel.DefaultGroupId,
		//"locked": userModel.Locked,
	})
}
