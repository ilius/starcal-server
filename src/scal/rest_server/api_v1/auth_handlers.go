package api_v1

import (
	"bytes"
	"fmt"
	"scal"
	"scal/event_lib"
	"scal/settings"
	"scal/storage"
	. "scal/user_lib"
	"scal/utils"
	"text/template"
	"time"

	. "github.com/ilius/restpc"

	"gopkg.in/mgo.v2/bson"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "auth",
		Map: RouteMap{
			"RegisterUser": {
				Method:  "POST",
				Pattern: "register",
				Handler: RegisterUser,
			},
			"Login": {
				Method:  "POST",
				Pattern: "login",
				Handler: Login,
			},
			"Logout": {
				Method:  "POST",
				Pattern: "logout",
				Handler: Logout,
			},
			"ChangePassword": {
				Method:  "POST",
				Pattern: "change-password",
				Handler: ChangePassword,
			},
			"ResetPasswordRequest": {
				Method:  "POST",
				Pattern: "reset-password-request",
				Handler: ResetPasswordRequest,
			},
			"ResetPasswordAction": {
				Method:  "POST",
				Pattern: "reset-password-action",
				Handler: ResetPasswordAction,
			},
		},
	})
}

func RegisterUser(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	email, err := req.GetString("email", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	password, err := req.GetString("password", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}

	userModel := UserModel{
		Email:    *email,
		Password: *password,
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	anotherUserModel := UserModelByEmail(userModel.Email, db)
	if anotherUserModel != nil {
		return nil, NewError(
			AlreadyExists, // FIXME: right code?
			"email is already registered",
			nil,
		)
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
		return nil, NewError(Internal, "", err)
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
		return nil, NewError(Internal, "", err)
	}

	err = db.Insert(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	signedToken := NewSignedToken(&userModel)
	return &Response{
		Data: map[string]interface{}{
			"token": signedToken,
		},
	}, nil
}

func Login(req Request) (*Response, error) {
	// Expires the token and cookie in 30 days
	//expireToken := time.Now().Add(30 * time.Day)
	//expireCookie := time.Now().Add(30 * time.Day)
	email, err := req.GetString("email", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	password, err := req.GetString("password", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	userModel := UserModelByEmail(*email, db)
	if userModel == nil {
		return nil, AuthError(fmt.Errorf("no user was found with this email"))
	}
	if GetPasswordHash(*email, *password) != userModel.Password {
		return nil, AuthError(fmt.Errorf("wrong password"))
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
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

	return &Response{
		Data: scal.M{
			"token": signedToken,
		},
	}, nil
}

func Logout(req Request) (*Response, error) {
	if req.GetHeader("Authorization") == "" {
		return &Response{}, nil
	}
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
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
		return nil, NewError(Internal, "", err)
	}
	userModel.LastLogoutTime = &now
	db.Update(userModel)
	return &Response{}, nil
}

func ChangePassword(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	email, err := req.GetString("email", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	password, err := req.GetString("password", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	newPassword, err := req.GetString("newPassword", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	userModel := UserModelByEmail(*email, db)
	if userModel == nil {
		return nil, AuthError(fmt.Errorf("no user was found with this email"))
	}
	if GetPasswordHash(*email, *password) != userModel.Password {
		return nil, AuthError(fmt.Errorf("wrong password"))
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	newPasswordHash := GetPasswordHash(
		userModel.Email,
		*newPassword,
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
		return nil, NewError(Internal, "", err)
	}
	userModel.Password = newPasswordHash
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func ResetPasswordRequest(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	email, err := req.GetString("email", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	userModel := UserModelByEmail(*email, db)
	if userModel == nil {
		// FIXME: should we let them know this email is not registered?
		return nil, NewError(InvalidArgument, "bad 'email'", nil)
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
			return nil, NewError(
				PermissionDenied,
				"There has been a Reset Password request for this email recently. Check your email inbox or wait more",
				nil,
			)
		}
	} else if !db.IsNotFound(err) {
		return nil, NewError(Internal, "", err)
	}

	expTime := now.Add(expDuration)
	token, err := utils.GenerateRandomBase64String(
		settings.RESET_PASSWORD_TOKEN_LENGTH,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	tokenModel := ResetPasswordTokenModel{
		Token:         token,
		Email:         *email,
		IssueTime:     now,
		ExpireTime:    expTime, // not reliable
		IssueRemoteIp: remoteIp,
	}
	err = db.Insert(tokenModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	// send an email containing this `token`, (and some instructions to use it)
	tplText := settings.RESET_PASSWORD_TOKEN_EMAIL_TEMPLATE
	tpl, err := template.New("ResetPassword " + *email).Parse(tplText)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	buf := bytes.NewBufferString("")
	err = tpl.Execute(buf, tokenModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	emailBody := buf.String()
	scal.SendEmail(
		*email,
		"StarCalendar Password Reset",
		false, // isHtml
		emailBody,
	)
	return &Response{
		Data: scal.M{
			"description": "Reset Password Token is sent to your email",
		},
	}, nil
}

func ResetPasswordAction(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	email, err := req.GetString("email", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	resetPasswordToken, err := req.GetString("resetPasswordToken", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	newPassword, err := req.GetString("newPassword", NotFromForm) // only from json
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	tokenModel := ResetPasswordTokenModel{
		Token: *resetPasswordToken,
	}
	err = db.Get(&tokenModel)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ForbiddenError("invalid 'resetPasswordToken'", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if tokenModel.Email != *email {
		return nil, ForbiddenError("invalid 'resetPasswordToken'", nil)
	}
	userModel := UserModelByEmail(*email, db)
	if userModel == nil {
		return nil, ForbiddenError("invalid 'resetPasswordToken'", nil)
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	now := time.Now()
	logModel := ResetPasswordLogModel{
		ResetPasswordTokenModel: tokenModel,
		UsedTime:                now,
		UsedRemoteIp:            remoteIp,
	}
	err = db.Insert(logModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Remove(tokenModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	newPasswordHash := GetPasswordHash(
		userModel.Email,
		*newPassword,
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
		return nil, NewError(Internal, "", err)
	}
	userModel.Password = newPasswordHash
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	tplText := settings.RESET_PASSWORD_DONE_EMAIL_TEMPLATE
	tpl, err := template.New("ResetPasswordAction " + *email).Parse(tplText)
	if err != nil {
		return nil, NewError(Internal, "", err)
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
		return nil, NewError(Internal, "", err)
	}
	emailBody := buf.String()
	err = scal.SendEmail(
		*email,
		"StarCalendar Password Reset",
		false, // isHtml
		emailBody,
	)
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	return &Response{}, nil
}
