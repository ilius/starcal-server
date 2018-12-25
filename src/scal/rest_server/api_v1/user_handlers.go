package api_v1

import (
	"scal"
	"scal/event_lib"
	"time"

	. "github.com/ilius/ripo"
	"gopkg.in/mgo.v2/bson"

	//"scal/settings"
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
		return nil, NewError(ResourceLocked, "user is being modified by another request", err)
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
		Time:         time.Now(),
		RequestEmail: email,
		RemoteIp:     remoteIp,
		FuncName:     "SetUserFullName",
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
		return nil, NewError(ResourceLocked, "user is being modified by another request", err)
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
		return nil, NewError(ResourceLocked, "user is being modified by another request", err)
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
		return nil, NewError(ResourceLocked, "user is being modified by another request", err)
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
		return nil, NewError(Internal, "", err)
	}

	userModel.DefaultGroupId = nil
	err = db.Update(userModel)
	if err != nil {
		return nil, NewError(Internal, "", err)
	}

	return &Response{}, nil
}

func GetUserInfo(req Request) (*Response, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email
	// -----------------------------------------------
	return &Response{
		Data: scal.M{
			"email":          email,
			"fullName":       userModel.FullName,
			"defaultGroupId": userModel.DefaultGroupId,
			//"locked": userModel.Locked,
			"lastLogoutTime": userModel.LastLogoutTime,
			"emailConfirmed": userModel.EmailConfirmed,
		},
	}, nil
}
