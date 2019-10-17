package user_lib

import (
	"time"

	"scal"
	"scal/storage"
)

var log = scal.Log

type UserModel struct {
	Id             string     `bson:"_id,objectid" json:"-"` // FIXME: why omitempty
	Email          string     `bson:"email" json:"email"`
	EmailConfirmed bool       `bson:"emailConfirmed" json:"emailConfirmed"`
	Password       string     `bson:"password" json:"password"`
	FullName       string     `bson:"fullName" json:"fullName"`
	Locked         bool       `bson:"locked" json:"-"`
	DefaultGroupId *string    `bson:"defaultGroupId,objectid" json:"defaultGroupId"`
	LastLogoutTime *time.Time `bson:"lastLogoutTime" json:"-"`

	TokenIssuedAt *time.Time `bson:"-" json:"-"`
}

func (m UserModel) UniqueM() scal.M {
	return scal.M{
		"email": m.Email,
	}
}

func (UserModel) Collection() string {
	return storage.C_user
}

func UserModelByEmail(email string, db storage.Database) *UserModel {
	user := UserModel{
		Email: email,
	}
	err := db.Get(&user)
	if err != nil {
		if !db.IsNotFound(err) {
			log.Error("unknown error in fetching user model: ", err)
		}
		return nil
	}
	return &user
}

type UserChangeLogModel struct {
	Time          time.Time `bson:"time"`
	RequestEmail  string    `bson:"requestEmail"`
	RemoteIp      string    `bson:"remoteIp"`
	TokenIssuedAt time.Time `bson:"tokenIssuedAt"`
	FuncName      string    `bson:"funcName"`

	Email          *[2]*string    `bson:"email,omitempty"`
	EmailConfirmed *[2]bool       `bson:"emailConfirmed,omitempty"`
	Password       *[2]string     `bson:"password,omitempty"`
	FullName       *[2]*string    `bson:"fullName,omitempty"`
	Locked         *[2]bool       `bson:"locked,omitempty"`
	DefaultGroupId *[2]*string    `bson:"defaultGroupId,omitempty"`
	LastLogoutTime *[2]*time.Time `bson:"lastLogoutTime,omitempty"`
}

func (model UserChangeLogModel) Collection() string {
	return storage.C_userChangeLog
}

type SpecialUserTokenModel struct {
	Token         string    `bson:"token"`
	Email         string    `bson:"email"`
	IssueTime     time.Time `bson:"issueTime"`
	ExpireTime    time.Time `bson:"expireTime"` // not reliable
	IssueRemoteIp string    `bson:"issueRemoteIp"`
}

type ResetPasswordTokenModel struct {
	SpecialUserTokenModel `bson:",inline"`
}

func (model ResetPasswordTokenModel) Collection() string {
	return storage.C_resetPwToken
}

func (model ResetPasswordTokenModel) UniqueM() scal.M {
	return scal.M{
		"token": model.Token,
	}
}

type ResetPasswordLogModel struct {
	TokenModel     SpecialUserTokenModel `bson:",inline"`
	ActionTime     time.Time             `bson:"actionTime"`
	ActionRemoteIp string                `bson:"actionRemoteIp"`
}

func (model ResetPasswordLogModel) Collection() string {
	return storage.C_resetPwLog
}

type UserLoginAttemptModel struct {
	Time time.Time `bson:"time" json:"time"`

	UserId string `bson:"userId,objectid" json:"userId"`
	Email  string `bson:"email" json:"email"`

	RemoteIp string `bson:"remoteIp" json:"remoteIp"`

	// password was correct
	Successful bool `bson:"successful" json:"successful"`

	// password was correct but login was rejected because user was locked
	Locked bool `bson:"locked" json:"locked,omitempty"`
}

func (model UserLoginAttemptModel) Collection() string {
	return storage.C_userLogins
}
