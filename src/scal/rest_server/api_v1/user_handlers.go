package api_v1

import (
	"scal"
	"scal/event_lib"
	"scal/user_lib"
	"time"

	. "github.com/ilius/ripo"
	"github.com/globalsign/mgo/bson"

	"scal/settings"
	"scal/storage"
	. "scal/user_lib"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "user",
		Map: RouteMap{
			"SetUserFullName": {
				Method:  "PUT",
				Pattern: "full-name",
				Handler: SetUserFullName,
			},
			"UnsetUserFullName": {
				Method:  "DELETE",
				Pattern: "full-name",
				Handler: UnsetUserFullName,
			},
			"GetUserInfo": {
				Method:  "GET",
				Pattern: "info",
				Handler: GetUserInfo,
			},
			"GetUserLoginHistory": {
				Method:  "GET",
				Pattern: "login-history",
				Handler: GetUserLoginHistory,
			},
			"SetUserDefaultGroupId": {
				Method:  "PUT",
				Pattern: "default-group",
				Handler: SetUserDefaultGroupId,
			},
			"UnsetUserDefaultGroupId": {
				Method:  "DELETE",
				Pattern: "default-group",
				Handler: UnsetUserDefaultGroupId,
			},
		},
	})
}

func SetUserFullName(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	failed, unlock := resLock.User(email)
	if failed {
		return nil, NewError(ResourceLocked, "user is being modified by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	const attrName = "fullName"
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	attrValue, err := req.GetString(attrName)
	if err != nil {
		return nil, err
	}

	err = db.Insert(UserChangeLogModel{
		Time:          time.Now(),
		RequestEmail:  email,
		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,
		FuncName:      "SetUserFullName",
		FullName: &[2]*string{
			&userModel.FullName,
			attrValue,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	userModel.FullName = *attrValue
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{}, nil
}

func UnsetUserFullName(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	failed, unlock := resLock.User(email)
	if failed {
		return nil, NewError(ResourceLocked, "user is being modified by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	err = db.Insert(UserChangeLogModel{
		Time:          time.Now(),
		RequestEmail:  email,
		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,
		FuncName:      "UnsetUserFullName",
		FullName: &[2]*string{
			&userModel.FullName,
			nil,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	userModel.FullName = ""
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{}, nil
}

func SetUserDefaultGroupId(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	failed, unlock := resLock.User(email)
	if failed {
		return nil, NewError(ResourceLocked, "user is being modified by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	const attrName = "defaultGroupId"
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	attrValue, err := req.GetString(attrName)
	if err != nil {
		return nil, err
	}

	groupModel, err := event_lib.LoadGroupModelByIdHex(
		"defaultGroupId",
		db,
		*attrValue,
	)
	if err != nil {
		return nil, err
	}
	groupId := groupModel.Id

	if groupModel.OwnerEmail != email {
		return nil, NewError(InvalidArgument, "invalid 'defaultGroupId'", nil)
	}

	err = db.Insert(UserChangeLogModel{
		Time:          time.Now(),
		RequestEmail:  email,
		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,
		FuncName:      "SetUserDefaultGroupId",
		DefaultGroupId: &[2]*bson.ObjectId{
			userModel.DefaultGroupId,
			&groupId,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	userModel.DefaultGroupId = &groupId
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{}, nil
}

func UnsetUserDefaultGroupId(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	failed, unlock := resLock.User(email)
	if failed {
		return nil, NewError(ResourceLocked, "user is being modified by another request", nil)
	}
	defer unlock()
	// -----------------------------------------------
	remoteIp, err := req.RemoteIP()
	if err != nil {
		return nil, err
	}

	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}

	err = db.Insert(UserChangeLogModel{
		Time:          time.Now(),
		RequestEmail:  email,
		RemoteIp:      remoteIp,
		TokenIssuedAt: *userModel.TokenIssuedAt,
		FuncName:      "UnsetUserDefaultGroupId",
		DefaultGroupId: &[2]*bson.ObjectId{
			userModel.DefaultGroupId,
			nil,
		},
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	userModel.DefaultGroupId = nil
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{}, nil
}

func loadLastLogins(req Request, userModel *user_lib.UserModel, limit int) ([]*user_lib.UserLoginAttemptModel, error) {
	result := []*user_lib.UserLoginAttemptModel{}
	db, err := storage.GetDB()
	if err != nil {
		return nil, NewError(Unavailable, "", err)
	}
	err = db.FindAll(&result, &storage.FindInput{
		Collection: storage.C_userLogins,
		Conditions: scal.M{
			"userId": userModel.Id,
		},
		SortBy: "-time",
		Limit:  limit,
	})
	if err != nil {
		return nil, NewError(Internal, "", err)
	}
	return result, nil
}

func GetUserInfo(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	lastLogins, err := loadLastLogins(req, userModel, settings.USER_INFO_LAST_LOGINS_LIMIT)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: scal.M{
			"email":          email,
			"fullName":       userModel.FullName,
			"defaultGroupId": userModel.DefaultGroupId,
			"emailConfirmed": userModel.EmailConfirmed,
			//"locked": userModel.Locked,

			"lastLogoutTime": userModel.LastLogoutTime,

			"lastLogins": lastLogins,
		},
	}, nil
}

func GetUserLoginHistory(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	// -----------------------------------------------
	const defaultLimit = settings.USER_LOGIN_HISTORY_DEFAULT_LIMIT
	limit, err := req.GetIntDefault("limit", defaultLimit)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultLimit
		// otherwise db will return everything
	}
	lastLogins, err := loadLastLogins(req, userModel, limit)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: scal.M{
			"lastLogins": lastLogins,
		},
	}, nil
}
