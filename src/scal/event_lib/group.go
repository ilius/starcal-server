package event_lib

import "gopkg.in/mgo.v2/bson"


type EventGroupModel struct {
    Id bson.ObjectId            `bson:"_id" json:"groupId"`
    Title string                `bson:"title" json:"title"`
    OwnerEmail string           `bson:"ownerEmail" json:"ownerEmail"`
    AddAccessEmails []string    `bson:"addAccessEmails,omitempty" json:"addAccessEmails,omitempty"`
    ReadAccessEmails []string   `bson:"readAccessEmails,omitempty" json:"readAccessEmails,omitempty"`
}
func (self EventGroupModel) EmailCanAdd(email string) bool {
    if email == self.OwnerEmail {
        return true
    }
    for _, aEmail := range self.AddAccessEmails {
        if email == aEmail {
            return true
        }
    }
    return false
}
func (self EventGroupModel) EmailCanRead(email string) bool {
    if email == self.OwnerEmail {
        return true
    }
    for _, aEmail := range self.ReadAccessEmails {
        if email == aEmail {
            return true
        }
    }
    return false
}

