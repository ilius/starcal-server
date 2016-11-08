package event_lib

import (
    "errors"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "scal/storage"
)

/*
// PublicJoinPolicy
const (
    FreeToJoin = "FreeToJoin"
    NoJoin = "NoJoin"
    //JoinNeedsVerify = "JoinNeedsVerify" // makes implementation more complex
)
*/

type EventAccessModel struct {
    EventId bson.ObjectId           `bson:"_id"`
    EventType string                `bson:"eventType"`
    OwnerEmail string               `bson:"ownerEmail"`
    IsPublic bool                   `bson:"isPublic"`
    AccessEmails []string           `bson:"accessEmails"`
    GroupId *bson.ObjectId          `bson:"groupId"`
    GroupModel *EventGroupModel     `bson:"-"`

    AttendingEmails []string        `bson:"attendingEmails"`
    NotAttendingEmails []string     `bson:"notAttendingEmails"`
    //PublicJoinPolicy string         `bson:"publicJoinPolicy"` // not indexed
    PublicJoinOpen bool             `bson:"publicJoinOpen"`
    MaxAttendees int                `bson:"maxAttendees"`
}
func (self EventAccessModel) UniqueM() bson.M {
    return bson.M{
        "_id": self.EventId,
    }
}
func (self EventAccessModel) Collection() string {
    return storage.C_access
}
func (self EventAccessModel) GroupIdHex() string {
    if self.GroupId != nil {
        return self.GroupId.Hex()
    }
    return ""
}
func (self *EventAccessModel) CanReadFull(email string) bool {
    if email == self.OwnerEmail {
        return true
    }
    for _, aEmail := range self.AccessEmails {
        if email == aEmail {
            return true
        }
    }
    if self.GroupModel != nil {
        for _, aEmail := range self.GroupModel.ReadAccessEmails {
            if email == aEmail {
                return true
            }
        }
    }
    return false
}
func (self *EventAccessModel) CanRead(email string) bool {
    if self.IsPublic {
        return true
    }
    return self.CanReadFull(email)
}
func (self *EventAccessModel) IsAttending(email string) bool {
    for _, aEmail := range self.AttendingEmails {
        if email == aEmail {
            return true
        }
    }
    return false
}
func (self *EventAccessModel) IsNotAttending(email string) bool {
    for _, aEmail := range self.NotAttendingEmails {
        if email == aEmail {
            return true
        }
    }
    return false
}
func (self *EventAccessModel) RemoveAttending(email string) bool {
    // return true if found and removed, false if not found
    for index, aEmail := range self.AttendingEmails {
        if email == aEmail {
            self.AttendingEmails = append(
                self.AttendingEmails[:index],
                self.AttendingEmails[index+1:]...
            )
            return true
        }
    }
    return false
}
func (self *EventAccessModel) RemoveNotAttending(email string) bool {
    // return true if found and removed, false if not found
    for index, aEmail := range self.NotAttendingEmails {
        if email == aEmail {
            self.NotAttendingEmails = append(
                self.NotAttendingEmails[:index],
                self.NotAttendingEmails[index+1:]...
            )
            return true
        }
    }
    return false
}
func (self *EventAccessModel) Join(email string) error {
    if self.IsAttending(email) {
        return errors.New("you have already joined this event")
    }
    if !self.CanReadFull(email) {
        if self.IsPublic {
            if !self.PublicJoinOpen {
                return errors.New("this public event is not open for joining")
            }
        } else {
            return errors.New("no access, no join")
        }
    }
    if self.MaxAttendees > 0 && len(self.AttendingEmails) >= self.MaxAttendees {
        return errors.New("maximum attendees exceeded, can not join event")
    }
    self.RemoveNotAttending(email)
    self.AttendingEmails = append(self.AttendingEmails, email)
    return nil
}
func (self *EventAccessModel) Leave(email string) error {
    if !self.RemoveAttending(email) {
        if self.CanReadFull(email) {
            return errors.New("you are not attending for this event")
        }
    }
    self.NotAttendingEmails = append(self.NotAttendingEmails, email)
    return nil
}



func LoadEventAccessModel(
    db *mgo.Database,
    eventId *bson.ObjectId,
    loadGroup bool,
) (*EventAccessModel, error) {
    var err error
    accessModel := EventAccessModel{}
    err = db.C(storage.C_access).Find(bson.M{
        "_id": eventId,
    }).One(&accessModel)
    if err != nil {
        return nil, err
    }
    if loadGroup && accessModel.GroupId != nil {
        groupModel := EventGroupModel{}
        err = db.C(storage.C_group).Find(bson.M{
            "_id": accessModel.GroupId,
        }).One(&groupModel)
        if err != nil {
            return nil, err
        }
    }
    return &accessModel, nil
}


