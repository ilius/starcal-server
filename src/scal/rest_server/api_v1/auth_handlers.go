package api_v1

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"net/http"
	"scal"
	"scal/event_lib"
	"scal/storage"
	. "scal/user_lib"
	"time"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "auth",
		Map: RouteMap{
			"RegisterUser": {
				"POST",
				"register",
				RegisterUser,
			},
			"Login": {
				"POST",
				"login",
				Login,
			},
			"ChangePassword": {
				"POST",
				"change-password",
				ChangePassword,
			},
		},
	})
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	userModel := UserModel{}
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
		//	msg = ""
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

	// add new field userModel.PasswordHash, FIXME
	userModel.Password = GetPasswordHash(
		userModel.Email,
		userModel.Password,
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
		//	nil
		//	&userModel.FullName,
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

func Login(w http.ResponseWriter, r *http.Request) {
	// Expires the token and cookie in 30 days
	//expireToken := time.Now().Add(30 * time.Day)
	//expireCookie := time.Now().Add(30 * time.Day)

	inputModel := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &inputModel)
	if err != nil {
		msg := err.Error()
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	if inputModel.Email == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'email'")
		return
	}
	if inputModel.Password == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'password'")
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel := UserModelByEmail(inputModel.Email, db)
	if userModel == nil {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"authentication failed",
		)
		return
	}

	if GetPasswordHash(inputModel.Email, inputModel.Password) != userModel.Password {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"authentication failed",
		)
		return
	}

	if userModel.Locked {
		SetHttpError(
			w,
			http.StatusForbidden,
			"user is locked",
		)
		return
	}

	signedToken := NewSignedToken(userModel)

	/*
		// Place the token in the client's cookie
		cookie := http.Cookie{
			Name:  "Auth",
			Value: signedToken,
			//Expires: expireCookie,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
	*/

	json.NewEncoder(w).Encode(scal.M{
		"token": signedToken,
	})
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	inputModel := struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		NewPassword string `json:"newPassword"`
	}{}
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &inputModel)
	if err != nil {
		msg := err.Error()
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	if inputModel.Email == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'email'")
		return
	}
	if inputModel.Password == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'password'")
		return
	}
	if inputModel.NewPassword == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'newPassword'")
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel := UserModelByEmail(inputModel.Email, db)
	if userModel == nil {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"authentication failed",
		)
		return
	}

	if GetPasswordHash(inputModel.Email, inputModel.Password) != userModel.Password {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"authentication failed",
		)
		return
	}

	if userModel.Locked {
		SetHttpError(
			w,
			http.StatusForbidden,
			"user is locked",
		)
		return
	}

	newPasswordHash := GetPasswordHash(
		userModel.Email,
		inputModel.NewPassword,
	)

	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: "", // FIXME
		RemoteIp:     remoteIp,
		FuncName:     "ChangePassword",
		Password: &[2]string{
			userModel.Password,
			newPasswordHash,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel.Password = newPasswordHash
	err = db.Update(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}
