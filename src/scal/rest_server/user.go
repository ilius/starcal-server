package rest_server

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "crypto/md5"



    "gopkg.in/mgo.v2-unstable/bson"
    "gopkg.in/mgo.v2-unstable"

    "scal/lib/go-http-auth"
    "scal/storage"
    "scal/event_lib"
)

const REALM = "starcalendar.net"

var globalDb, globalDbErr = storage.GetDB()

var authenticator = auth.NewDigestAuthenticator(REALM, Secret)

func init() {
    if globalDbErr != nil {
        panic(globalDbErr)
    }
}


type UserModel struct {
    Id bson.ObjectId    `bson:"_id,omitempty" json:"-"`
    Email string        `bson:"email" json:"email"`
    FullName string     `bson:"fullName" json:"fullName"`
    Password string     `bson:"password" json:"password"`
    Locked bool         `bson:"locked" json:"-"`
    DefaultGroupId bson.ObjectId    `bson:"defaultGroupId" json:"defaultGroupId"`
}

func UserModelByEmail(email string, db *mgo.Database) *UserModel {
    user := UserModel{}
    err := db.C("users").Find(bson.M{
        "email": email,
    }).One(&user)
    if err != nil {
        if err != mgo.ErrNotFound {
            log.Print("unkown error in fetching user model:", err)
        }
        return nil
    }
    return &user
}

//type Request http.Request
/*
type Request auth.AuthenticatedRequest
func (r Request) Email() string {
    return r.Username
}
*/


func Secret(email string, realm string) string {
    userModel := UserModelByEmail(email, globalDb)
    if userModel == nil {
        return ""
    }
    if userModel.Locked {
        return "" // FIXME
    }
    return userModel.Password
}


func RegisterUser(w http.ResponseWriter, r *http.Request) {
    userModel := UserModel{}
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    var err error
    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    err = json.Unmarshal(body, &userModel)
    if err != nil {
        msg := err.Error()
        //if strings.Contains(msg, "") {
        //    msg = ""
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
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    anotherUserModel := UserModelByEmail(userModel.Email, db)
    if anotherUserModel != nil {
        SetHttpError(w, http.StatusBadRequest, "duplicate 'email'")
        return
    }
    userModel.Password = fmt.Sprintf(
        "%x",
        md5.Sum(
            []byte(
                fmt.Sprintf(
                    "%s:%s:%s",
                    userModel.Email,
                    REALM,
                    userModel.Password,
                ),
            ),
        ),
    )
    defaultGroup := event_lib.EventGroupModel{
        Id: bson.NewObjectId(),
        Title: userModel.Email,
        OwnerEmail: userModel.Email,
    }
    err = db.C("event_group").Insert(defaultGroup)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    userModel.DefaultGroupId = defaultGroup.Id
    err = db.C("users").Insert(userModel)
    if err != nil {
        SetHttpError(w, http.StatusInternalServerError, err.Error())
        return
    }
    json.NewEncoder(w).Encode(map[string]string{
        "successful": "true",
        "defaultGroupId": defaultGroup.Id.Hex(),
    })
}








