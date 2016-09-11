package event_lib

import "time"
import "gopkg.in/mgo.v2-unstable/bson"

type EventRevisionModel struct {
    UserId int              `bson:"userId"`
    EventId bson.ObjectId   `bson:"eventId"`
    EventType string        `bson:"eventType"`
    Sha1 string             `bson:"sha1"`
    Time time.Time          `bson:"time"`
    //InvitedUserIds []int  `bson:"invitedUserIds"`
}





