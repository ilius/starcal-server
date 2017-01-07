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
	Password       string         `bson:"password" json:"password"`
	FullName       string         `bson:"fullName" json:"fullName"`
	Locked         bool           `bson:"locked" json:"-"`
	DefaultGroupId *bson.ObjectId `bson:"defaultGroupId" json:"defaultGroupId"`
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
	Password       *[2]string         `bson:"password,omitempty"`
	FullName       *[2]*string        `bson:"fullName,omitempty"`
	Locked         *[2]bool           `bson:"locked,omitempty"`
	DefaultGroupId *[2]*bson.ObjectId `bson:"defaultGroupId,omitempty"`
}

func (model UserChangeLogModel) Collection() string {
	return storage.C_userChangeLog
}
