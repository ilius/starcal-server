package api_v1

import (
	"bytes"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"net/http"
	"scal"
	"scal/event_lib"
	"scal/settings"
	"scal/storage"
	. "scal/user_lib"
	"scal/utils"
	"text/template"
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
			"Logout": {
				"POST",
				"logout",
				Logout,
			},
			"ChangePassword": {
				"POST",
				"change-password",
				ChangePassword,
			},
			"ResetPasswordRequest": {
				"POST",
				"reset-password-request",
				ResetPasswordRequest,
			},
			"ResetPasswordAction": {
				"POST",
				"reset-password-action",
				ResetPasswordAction,
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
		SetHttpError(w, http.StatusBadRequest, "email is already registered")
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

func Logout(w http.ResponseWriter, r *http.Request) {
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
	now := time.Now()
	err = db.Insert(UserChangeLogModel{
		Time:         now,
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "Logout",
		LastLogoutTime: &[2]*time.Time{
			userModel.LastLogoutTime,
			&now,
		},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	userModel.LastLogoutTime = &now
	db.Update(userModel)
	json.NewEncoder(w).Encode(scal.M{})
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

func ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// -----------------------------------------------
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	inputModel := struct {
		Email string `json:"email"`
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
	email := inputModel.Email
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpError(w, http.StatusBadRequest, "bad 'email'")
		return
	}
	now := time.Now()
	expDuration := settings.RESET_PASSWORD_EXP_SECONDS * time.Second
	lastToken := ResetPasswordTokenModel{}
	err = db.First(
		scal.M{
			"email": email,
		}, // cond
		"-issueTime", // sortBy
		&lastToken,
	)
	if err == nil {
		if now.Sub(lastToken.IssueTime) < expDuration {
			SetHttpError(
				w,
				http.StatusForbidden,
				"There has been a Reset Password request for this email recently. Check your email inbox or wait more",
			)
			return
		}
	} else if !db.IsNotFound(err) {
		SetHttpErrorInternal(w, err)
		return
	}
	expTime := now.Add(expDuration)
	token, err := utils.GenerateRandomBase64String(
		settings.RESET_PASSWORD_TOKEN_LENGTH,
	)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	tokenModel := ResetPasswordTokenModel{
		Token:         token,
		Email:         email,
		IssueTime:     now,
		ExpireTime:    expTime, // not reliable
		IssueRemoteIp: remoteIp,
	}
	err = db.Insert(tokenModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	// send an email containing this `token`, (and some instructions to use it)
	tplText := settings.RESET_PASSWORD_TOKEN_EMAIL_TEMPLATE
	tpl, err := template.New("ResetPassword " + email).Parse(tplText)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	buf := bytes.NewBufferString("")
	err = tpl.Execute(buf, tokenModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	emailBody := buf.String()
	scal.SendEmail(
		email,
		"StarCalendar Password Reset",
		false, // isHtml
		emailBody,
	)
	json.NewEncoder(w).Encode(scal.M{
		"description": "Reset Password Token is sent to your email",
	})
}

func ResetPasswordAction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// -----------------------------------------------
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	inputModel := struct {
		Email              string `json:"email"`
		ResetPasswordToken string `json:"resetPasswordToken"`
		NewPassword        string `json:"newPassword"`
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
	if inputModel.ResetPasswordToken == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'resetPasswordToken'")
		return
	}
	if inputModel.NewPassword == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'newPassword'")
		return
	}
	email := inputModel.Email
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	tokenModel := ResetPasswordTokenModel{
		Token: inputModel.ResetPasswordToken,
	}
	err = db.Get(&tokenModel)
	if err != nil {
		if db.IsNotFound(err) {
			SetHttpError(w, http.StatusForbidden, "invalid 'resetPasswordToken'")
		} else {
			SetHttpErrorInternal(w, err)
		}
		return
	}
	if tokenModel.Email != email {
		SetHttpError(w, http.StatusForbidden, "invalid 'resetPasswordToken'")
		return
	}
	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		SetHttpError(w, http.StatusBadRequest, "bad 'email'")
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
	now := time.Now()
	logModel := ResetPasswordLogModel{
		ResetPasswordTokenModel: tokenModel,
		UsedTime:                now,
		UsedRemoteIp:            remoteIp,
	}
	err = db.Insert(logModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	db.Remove(tokenModel)
	newPasswordHash := GetPasswordHash(
		userModel.Email,
		inputModel.NewPassword,
	)
	err = db.Insert(UserChangeLogModel{
		Time:         now,
		RequestEmail: "", // FIXME
		RemoteIp:     remoteIp,
		FuncName:     "ResetPasswordAction",
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
	tplText := settings.RESET_PASSWORD_DONE_EMAIL_TEMPLATE
	tpl, err := template.New("ResetPasswordAction " + email).Parse(tplText)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	buf := bytes.NewBufferString("")
	tplParams := struct {
		Name     string
		RemoteIp string
	}{
		Name:     userModel.FullName,
		RemoteIp: remoteIp,
	}
	err = tpl.Execute(buf, tplParams)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	emailBody := buf.String()
	scal.SendEmail(
		email,
		"StarCalendar Password Reset",
		false, // isHtml
		emailBody,
	)
	json.NewEncoder(w).Encode(scal.M{})
}
