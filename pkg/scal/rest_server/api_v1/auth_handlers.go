package api_v1

import (
	"bytes"
	"fmt"
	"net/url"
	"text/template"
	"time"

	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/event_lib"
	"github.com/ilius/starcal-server/pkg/scal/settings"
	"github.com/ilius/starcal-server/pkg/scal/storage"
	"github.com/ilius/starcal-server/pkg/scal/user_lib"
	. "github.com/ilius/starcal-server/pkg/scal/user_lib"

	"github.com/ilius/libgostarcal/utils"

	jwt "github.com/golang-jwt/jwt/v5"
	. "github.com/ilius/ripo"

	"github.com/ilius/mgo/bson"
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
			"ConfirmEmailRequest": {
				Method:  "POST",
				Pattern: "confirm-email-request",
				Handler: ConfirmEmailRequest,
			},
			"ConfirmEmailAction": {
				Method:  "GET",
				Pattern: "confirm-email-action",
				Handler: ConfirmEmailAction,
			},
		},
	})
}

func RegisterUser(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	email, err := req.GetString("email", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	password, err := req.GetString("password", FromBody) // only from json
	if err != nil {
		return nil, err
	}

	userModel := &UserModel{
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

	passwordHash, err := GetPasswordHash(
		userModel.Email,
		userModel.Password,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	// add new field userModel.PasswordHash, FIXME
	userModel.Password = passwordHash

	userModel.Id = bson.NewObjectId().Hex()

	defaultGroup := event_lib.EventGroupModel{
		Id:         bson.NewObjectId().Hex(),
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
		DefaultGroupId: &[2]*string{
			nil,
			userModel.DefaultGroupId,
		},
		// FullName: &[2]*string{
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

	err = sendEmailConfirmation(req, userModel, remoteIp)
	if err != nil {
		// FIXME: call error dispatcher (to save to mongo), but don't return error
		log.Error(err)
	}

	signedToken, exp := NewSignedToken(userModel)
	return &Response{
		Data: map[string]any{
			"token":      signedToken,
			"expiration": exp.Format(time.RFC3339),
			"message":    "an email confirmation is sent to your email address",
		},
	}, nil
}

func failedLogin(req Request, db storage.Database, userModel *user_lib.UserModel, remoteIP string) {
	if settings.STORE_FAILED_LOGINS {
		now := time.Now()
		err := db.Insert(UserLoginAttemptModel{
			Time:       now,
			UserId:     userModel.Id,
			Email:      userModel.Email,
			RemoteIp:   remoteIP,
			Successful: false,
		})
		if err != nil {
			DispatchError(req, NewError(Internal, "", err))
		}
	}
}

func successfulLogin(req Request, db storage.Database, userModel *user_lib.UserModel, remoteIP string) {
	if settings.STORE_SUCCESSFUL_LOGINS {
		now := time.Now()
		err := db.Insert(UserLoginAttemptModel{
			Time:       now,
			UserId:     userModel.Id,
			Email:      userModel.Email,
			RemoteIp:   remoteIP,
			Successful: true,
		})
		if err != nil {
			DispatchError(req, NewError(Internal, "", err))
		}
	}
}

func lockedSuccessfulLogin(req Request, db storage.Database, userModel *user_lib.UserModel, remoteIP string) {
	if settings.STORE_LOCKED_SUCCESSFUL_LOGINS {
		now := time.Now()
		err := db.Insert(UserLoginAttemptModel{
			Time:       now,
			UserId:     userModel.Id,
			Email:      userModel.Email,
			RemoteIp:   remoteIP,
			Successful: true,
			Locked:     true,
		})
		if err != nil {
			DispatchError(req, NewError(Internal, "", err))
		}
	}
}

func Login(req Request) (*Response, error) {
	// Expires the token and cookie in 30 days
	// expireToken := time.Now().Add(30 * time.Day)
	// expireCookie := time.Now().Add(30 * time.Day)
	email, err := req.GetString("email", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	// ----------------------
	remoteIP, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	// ----------------------
	failed, unlock := resLock.UserLogin(*email, remoteIP)
	if failed {
		time.Sleep(1 * time.Second)
		return nil, NewError(ResourceLocked, "someone else with your IP is trying to login", nil)
	}
	defer unlock()
	// ----------------------
	password, err := req.GetString("password", FromBody) // only from json
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
	if !CheckPasswordHash(*email, *password, userModel.Password) {
		failedLogin(req, db, userModel, remoteIP)
		return nil, AuthError(fmt.Errorf("wrong password"))
	}

	if userModel.Locked {
		lockedSuccessfulLogin(req, db, userModel, remoteIP)
		return nil, ForbiddenError("user is locked", nil)
	}

	successfulLogin(req, db, userModel, remoteIP)

	signedToken, exp := NewSignedToken(userModel)

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
			"token":      signedToken,
			"expiration": exp.Format(time.RFC3339),
		},
	}, nil
}

func Logout(req Request) (*Response, error) {
	if req.Header("Authorization") == "" {
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
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return &Response{}, nil
}

func ChangePassword(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	email, err := req.GetString("email", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	password, err := req.GetString("password", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	newPassword, err := req.GetString("newPassword", FromBody) // only from json
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
	if !CheckPasswordHash(*email, *password, userModel.Password) {
		return nil, AuthError(fmt.Errorf("wrong password"))
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	newPasswordHash, err := GetPasswordHash(
		userModel.Email,
		*newPassword,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
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

type ResetPasswordRequestTemplateParams struct {
	ResetPasswordTokenModel
	Host string
}

func ResetPasswordRequest(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	email, err := req.GetString("email", FromBody) // only from json
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
	lastToken := ResetPasswordTokenModel{}
	err = db.First(
		scal.M{
			"email": email,
		}, // cond
		"-issueTime", // sortBy
		&lastToken,
	)
	if err == nil {
		if now.Sub(lastToken.IssueTime) < settings.RESET_PASSWORD_REJECT_SECONDS*time.Second {
			return nil, NewError(
				PermissionDenied,
				"There has been a Reset Password request for this email recently."+
					" Check your email inbox, or wait a little bit and re-send this request.",
				nil,
			)
		}
	} else if !db.IsNotFound(err) {
		return nil, NewError(Internal, "", err)
	}

	expTime := now.Add(settings.RESET_PASSWORD_EXP_SECONDS * time.Second)
	token, err := utils.GenerateRandomBase64String(
		settings.RESET_PASSWORD_TOKEN_LENGTH,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	tokenModel := ResetPasswordTokenModel{
		SpecialUserTokenModel: SpecialUserTokenModel{
			Token:         token,
			Email:         *email,
			IssueTime:     now,
			ExpireTime:    expTime, // not reliable
			IssueRemoteIp: remoteIp,
		},
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
	err = tpl.Execute(buf, ResetPasswordRequestTemplateParams{
		ResetPasswordTokenModel: tokenModel,
		Host:                    settings.HOST,
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	emailBody := buf.String()
	err = scal.SendEmail(&scal.SendEmailInput{
		To:      *email,
		Subject: "StarCalendar Password Reset",
		IsHtml:  false,
		Body:    emailBody,
	})
	if err != nil {
		log.Error("Failed to send email:\n", emailBody)
		return nil, NewError(Unavailable, "error in sending email", err)
	}
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
	emailPtr, err := req.GetString("email", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	email := *emailPtr
	resetPasswordToken, err := req.GetString("resetPasswordToken", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	newPassword, err := req.GetString("newPassword", FromBody) // only from json
	if err != nil {
		return nil, err
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	tokenModel := ResetPasswordTokenModel{
		SpecialUserTokenModel: SpecialUserTokenModel{
			Token: *resetPasswordToken,
		},
	}
	err = db.Get(&tokenModel)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ForbiddenError("invalid 'resetPasswordToken'", err)
		}
		return nil, NewError(Internal, "", err)
	}
	if tokenModel.Email != email {
		return nil, ForbiddenError(
			"invalid 'resetPasswordToken'",
			fmt.Errorf("MISMATCH Email: %#v != %#v", tokenModel.Email, email),
		)
	}
	if tokenModel.ExpireTime.Before(time.Now()) {
		return nil, ForbiddenError(
			"invalid 'resetPasswordToken'",
			fmt.Errorf("token expired, ExpireTime=%v", tokenModel.ExpireTime),
		)
	}
	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		return nil, ForbiddenError(
			"invalid 'resetPasswordToken'",
			fmt.Errorf("no user found with email=%#v", email),
		)
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	now := time.Now()
	logModel := ResetPasswordLogModel{
		TokenModel:     tokenModel.SpecialUserTokenModel,
		ActionTime:     now,
		ActionRemoteIp: remoteIp,
	}
	err = db.Insert(logModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Remove(tokenModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	newPasswordHash, err := GetPasswordHash(
		userModel.Email,
		*newPassword,
	)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
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
	tpl, err := template.New("ResetPasswordAction " + email).Parse(tplText)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	buf := bytes.NewBufferString("")
	tplParams := struct {
		Name     string
		RemoteIp string
		Host     string
	}{
		Name:     userModel.FullName,
		RemoteIp: remoteIp,
		Host:     settings.HOST,
	}
	err = tpl.Execute(buf, tplParams)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	emailBody := buf.String()
	err = scal.SendEmail(&scal.SendEmailInput{
		To:      email,
		Subject: "StarCalendar Password Reset",
		IsHtml:  false,
		Body:    emailBody,
	})
	if err != nil {
		log.Error("Failed to send email:\n", emailBody)
		return nil, NewError(Unavailable, "", err)
	}
	return &Response{}, nil
}

func sendEmailConfirmation(req Request, userModel *UserModel, remoteIp string) error {
	email := userModel.Email

	now := time.Now()
	exp := now.Add(time.Duration(60) * time.Minute)
	tokenStr, _ := jwt.NewWithClaims(
		jwt.GetSigningMethod(settings.JWT_ALG),
		jwt.MapClaims{
			"email":    email,
			"remoteIp": remoteIp,
			"iat":      now.Unix(),
			"exp":      exp.Unix(),
		},
	).SignedString([]byte(
		settings.CONFIRM_EMAIL_SECRET,
	))

	values := url.Values{}
	values.Add("token", tokenStr)
	confirmationURL := "http://" + req.Host() + "/auth/confirm-email-action?" + values.Encode()

	tplText := settings.CONFIRM_EMAIL_EMAIL_TEMPLATE
	tpl, err := template.New("ConfirmEmailRequest " + email).Parse(tplText)
	if err != nil {
		return NewError(Internal, "", err)
	}
	buf := bytes.NewBufferString("")
	tplParams := struct {
		Name            string
		ConfirmationURL string
		ExpirationTime  string
		Host            string
	}{
		Name:            userModel.FullName,
		ConfirmationURL: confirmationURL,
		ExpirationTime:  exp.Format(time.RFC1123),
		Host:            settings.HOST,
	}
	err = tpl.Execute(buf, tplParams)
	if err != nil {
		return NewError(Internal, "", err)
	}
	emailBody := buf.String()
	fmt.Println(emailBody)
	err = scal.SendEmail(&scal.SendEmailInput{
		To:      email,
		Subject: "StarCalendar Email Confirmation",
		IsHtml:  false,
		Body:    emailBody,
	})
	if err != nil {
		log.Error("Failed to send email:\n", emailBody)
		return NewError(Unavailable, "", err)
	}
	return nil
}

func ConfirmEmailRequest(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	if userModel.EmailConfirmed {
		return &Response{
			Data: scal.M{
				"message": "Your email address has been ALREADY CONFIRMED",
			},
		}, nil
	}
	err = sendEmailConfirmation(req, userModel, remoteIp)
	if err != nil {
		return nil, err
	}
	return &Response{}, nil
}

func ConfirmEmailAction(req Request) (*Response, error) {
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}
	tokenStr, err := req.GetString("token")
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse(
		*tokenStr,
		func(token *jwt.Token) (any, error) {
			return []byte(settings.CONFIRM_EMAIL_SECRET), nil
		},
	)
	if err != nil {
		return nil, ForbiddenError("invalid email confirmation token", err)
	}

	expectedAlg := settings.JWT_ALG
	tokenAlg := token.Header["alg"]
	if expectedAlg != tokenAlg {
		return nil, ForbiddenError("invalid email confirmation token", fmt.Errorf(
			"Expected %s signing method but token specified %s",
			expectedAlg,
			tokenAlg,
		))
	}

	if !token.Valid {
		return nil, ForbiddenError("invalid email confirmation token", fmt.Errorf("token.Valid == false"))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ForbiddenError("invalid email confirmation token", errClaimsNotFound)
	}
	tokenEmail := claims["email"]
	tokenRemoteIp := claims["remoteIp"]

	if tokenRemoteIp != remoteIp {
		return nil, ForbiddenError(
			"invalid email confirmation token",
			fmt.Errorf("MISMATCH REMOTE IP %#v != %#v", tokenRemoteIp, remoteIp),
		)
	}
	email, ok := tokenEmail.(string)
	if !ok {
		return nil, ForbiddenError("invalid email confirmation token", fmt.Errorf("tokenEmail = %#v", tokenEmail))
	}
	if email == "" {
		return nil, ForbiddenError("invalid email confirmation token", nil)
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	userModel := UserModelByEmail(email, db)
	if userModel == nil {
		return nil, ForbiddenError(
			"invalid email confirmation token",
			fmt.Errorf("no user found with email %#v", email),
		)
	}
	if userModel.EmailConfirmed {
		return &Response{
			Data: scal.M{"message": "Your email address is already confirmed."},
		}, nil
	}
	userModel.EmailConfirmed = true

	err = db.Insert(UserChangeLogModel{
		Time:           time.Now(),
		RequestEmail:   email,
		RemoteIp:       remoteIp,
		FuncName:       "ConfirmEmailAction",
		EmailConfirmed: &[2]bool{false, true},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{
		Data: scal.M{"message": "Your email address is now confirmed."},
	}, nil
}
