package event_lib

import "gopkg.in/mgo.v2/bson"

import "scal/storage"

type EventGroupModel struct {
    Id bson.ObjectId            `bson:"_id" json:"groupId"`
    Title string                `bson:"title" json:"title"`
    OwnerEmail string           `bson:"ownerEmail" json:"ownerEmail"`
    AddAccessEmails []string    `bson:"addAccessEmails,omitempty" json:"addAccessEmails,omitempty"`
    ReadAccessEmails []string   `bson:"readAccessEmails,omitempty" json:"readAccessEmails,omitempty"`
}
func (self EventGroupModel) UniqueM() bson.M {
    return bson.M{
        "_id": self.Id,
    }
}
func (self EventGroupModel) Collection() string {
    return storage.C_group
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
func (self EventGroupModel) CanRead(email string) bool {
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

