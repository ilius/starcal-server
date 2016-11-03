package event_lib

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "scal/storage"
)

type EventAccessModel struct {
    EventId bson.ObjectId           `bson:"_id"`
    EventType string                `bson:"eventType"`
    OwnerEmail string               `bson:"ownerEmail"`
    AccessEmails []string           `bson:"accessEmails"`
    GroupId *bson.ObjectId          `bson:"groupId"`
    GroupModel *EventGroupModel     `bson:"-"`
}
func (self EventAccessModel) UniqueM() bson.M {
    return bson.M{
        "_id": self.EventId,
    }
}
func (self EventAccessModel) Collection() string {
    return storage.C_access
}
func (self *EventAccessModel) CanRead(email string) bool {
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


