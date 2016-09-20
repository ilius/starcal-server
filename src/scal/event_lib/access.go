package event_lib

import "gopkg.in/mgo.v2-unstable"
import "gopkg.in/mgo.v2-unstable/bson"

type EventAccessModel struct {
    EventId bson.ObjectId           `bson:"_id"`
    OwnerEmail string               `bson:"ownerEmail"`
    AccessEmails []string           `bson:"accessEmails"`
    GroupId *bson.ObjectId          `bson:"groupId"`
    GroupModel *EventGroupModel     `bson:"-"`
}
func (self EventAccessModel) EmailCanRead(email string) bool {
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
    eventId bson.ObjectId,
    loadGroup bool,
) (*EventAccessModel, error) {
    var err error
    accessModel := EventAccessModel{}
    err = db.C("event_access").Find(bson.M{
        "_id": eventId,
    }).One(&accessModel)
    if err != nil {
        return nil, err
    }
    if loadGroup && accessModel.GroupId != nil {
        groupModel := EventGroupModel{}
        err = db.C("event_group").Find(bson.M{
            "_id": accessModel.GroupId,
        }).One(&groupModel)
        if err != nil {
            return nil, err
        }
    }
    return &accessModel, nil
}


