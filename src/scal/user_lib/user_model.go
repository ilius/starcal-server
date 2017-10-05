package user_lib

import (
	"log"
	"time"

	"gopkg.in/mgo.v2/bson"

	"scal"
	"scal/storage"
)

type UserModel struct {
	Id             bson.ObjectId  `bson:"_id,omitempty" json:"-"` // FIXME
	Email          string         `bson:"email" json:"email"`
	EmailConfirmed bool           `bson:"emailConfirmed" json:"emailConfirmed"`
	Password       string         `bson:"password" json:"password"`
	FullName       string         `bson:"fullName" json:"fullName"`
	Locked         bool           `bson:"locked" json:"-"`
	DefaultGroupId *bson.ObjectId `bson:"defaultGroupId" json:"defaultGroupId"`
	LastLogoutTime *time.Time     `bson:"lastLogoutTime" json:"-"`
}

func (self UserModel) UniqueM() scal.M {
	return scal.M{
		"email": self.Email,
	}
}
func (self UserModel) Collection() string {
	return storage.C_user
}

func UserModelByEmail(email string, db storage.Database) *UserModel {
	user := UserModel{
		Email: email,
	}
	err := db.Get(&user)
	if err != nil {
		if !db.IsNotFound(err) {
			log.Print("unkown error in fetching user model:", err)
		}
		return nil
	}
	return &user
}

type UserChangeLogModel struct {
	Time         time.Time `bson:"time"`
	RequestEmail string    `bson:"requestEmail"`
	RemoteIp     string    `bson:"remoteIp"`
	FuncName     string    `bson:"funcName"`

	Email          *[2]*string        `bson:"email,omitempty"`
	EmailConfirmed *[2]bool           `bson:"emailConfirmed,omitempty"`
	Password       *[2]string         `bson:"password,omitempty"`
	FullName       *[2]*string        `bson:"fullName,omitempty"`
	Locked         *[2]bool           `bson:"locked,omitempty"`
	DefaultGroupId *[2]*bson.ObjectId `bson:"defaultGroupId,omitempty"`
	LastLogoutTime *[2]*time.Time     `bson:"lastLogoutTime,omitempty"`
}

func (model UserChangeLogModel) Collection() string {
	return storage.C_userChangeLog
}

type ResetPasswordTokenModel struct {
	Token         string    `bson:"token"`
	Email         string    `bson:"email"`
	IssueTime     time.Time `bson:"issueTime"`
	ExpireTime    time.Time `bson:"expireTime"` // not reliable
	IssueRemoteIp string    `bson:"issueRemoteIp"`
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
	ResetPasswordTokenModel `bson:",inline"`
	UsedTime                time.Time `bson:"usedTime"`
	UsedRemoteIp            string    `bson:"usedRemoteIp"`
}

func (model ResetPasswordLogModel) Collection() string {
	return storage.C_resetPwLog
}
