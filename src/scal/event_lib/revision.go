package event_lib

import "time"
import "gopkg.in/mgo.v2-unstable/bson"

type EventRevisionModel struct {
    EventId bson.ObjectId   `bson:"eventId" json:"eventId"`
    EventType string        `bson:"eventType" json:"eventType"`
    Sha1 string             `bson:"sha1" json:"sha1"`
    Time time.Time          `bson:"time" json:"time"`
    //InvitedEmails []string    `bson:"invitedEmails" json:"invitedEmails"`
}

