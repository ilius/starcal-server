package rest_server

import (
	"fmt"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"scal/storage"
)

type UserModel struct {
	Id             bson.ObjectId  `bson:"_id,omitempty" json:"-"` // FIXME
	Email          string         `bson:"email" json:"email"`
	FullName       string         `bson:"fullName" json:"fullName"`
	Password       string         `bson:"password" json:"password"`
	Locked         bool           `bson:"locked" json:"-"`
	DefaultGroupId *bson.ObjectId `bson:"defaultGroupId" json:"defaultGroupId"`
}

func (self UserModel) UniqueM() bson.M {
	return bson.M{
		"email": self.Email,
	}
}
func (self UserModel) Collection() string {
	return storage.C_user
}

func UserModelByEmail(email string, db *mgo.Database) *UserModel {
	user := UserModel{
		Email: email,
	}
	err := storage.Get(db, &user)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Print("unkown error in fetching user model:", err)
		}
		return nil
	}
	return &user
}

func SetHttpErrorUserNotFound(w http.ResponseWriter, email string) {
	SetHttpError(
		w,
		http.StatusInternalServerError,
		fmt.Sprintf(
			"user with email '%s' not found",
			email,
		),
	)
}
