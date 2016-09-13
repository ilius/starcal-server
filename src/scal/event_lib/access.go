package event_lib

import "gopkg.in/mgo.v2-unstable/bson"

type EventAccessModel struct {
    EventId bson.ObjectId   `bson:"_id"`
    OwnerEmail string       `bson:"ownerEmail"`
    AccessEmails []string   `bson:"accessEmails"`
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
    return false
}

