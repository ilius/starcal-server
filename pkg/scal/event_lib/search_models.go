package event_lib

import (
	"time"

	"github.com/ilius/mgo/bson"
)

type MovedEventsRow struct {
	EventId      string    `bson:"_id,objectid" json:"eventId"`
	OldGroupId   any       `json:"oldGroupId"`
	OldGroupItem [2]any    `bson:"oldGroupItem"`
	NewGroupId   any       `json:"newGroupId"`
	NewGroupItem [2]any    `bson:"newGroupItem"`
	Time         time.Time `bson:"time" json:"time"`
}

// TODO: measure memory usage of large slice of this struct
// where 1, 2 or 3 of the fields are always empty
// does it help that I use pointers?
type ListEventsRow struct {
	EventId    string  `bson:"_id,objectid" json:"eventId"`
	EventType  *string `bson:"eventType" json:"eventType,omitempty"`
	GroupId    *string `bson:"groupId,objectid" json:"groupId,omitempty"`
	OwnerEmail *string `bson:"ownerEmail" json:"ownerEmail,omitempty"`
}

type ListGroupsRow struct {
	GroupId    bson.ObjectId `bson:"_id,objectid" json:"groupId"`
	Title      string        `bson:"title" json:"title"`
	OwnerEmail string        `bson:"ownerEmail" json:"ownerEmail"`
}
