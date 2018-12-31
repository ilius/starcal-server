package event_lib

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type MovedEventsRow struct {
	EventId    bson.ObjectId `bson:"_id" json:"eventId"`
	OldGroupId interface{}   `bson:"oldGroupId" json:"oldGroupId"`
	NewGroupId interface{}   `bson:"newGroupId" json:"newGroupId"`
	Time       time.Time     `bson:"time" json:"time"`
}

// TODO: measure memory usage of large slice of this struct
// where 1, 2 or 3 of the fields are always empty
// does it help that I use pointers?
type ListEventsRow struct {
	EventId    bson.ObjectId  `bson:"_id" json:"eventId"`
	EventType  *string        `bson:"eventType" json:"eventType,omitempty"`
	GroupId    *bson.ObjectId `bson:"groupId" json:"groupId,omitempty"`
	OwnerEmail *string        `bson:"ownerEmail" json:"ownerEmail,omitempty"`
}

type ListGroupsRow struct {
	GroupId    bson.ObjectId `bson:"_id" json:"groupId"`
	Title      string        `bson:"title" json:"title"`
	OwnerEmail string        `bson:"ownerEmail" json:"ownerEmail"`
}
