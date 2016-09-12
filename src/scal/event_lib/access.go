package event_lib

import "gopkg.in/mgo.v2-unstable/bson"

type EventAccessModel struct {
    EventId bson.ObjectId   `bson:"_id"`
    OwnerId int             `bson:"ownerId"`
    AccessUserIds []int     `bson:"accessUserIds"`
}
func (self EventAccessModel) UserCanRead(userId int) bool {
    if userId == self.OwnerId {
        return true
    }
    for _, aUserId := range self.AccessUserIds {
        if userId == aUserId {
            return true
        }
    }
    return false
}

