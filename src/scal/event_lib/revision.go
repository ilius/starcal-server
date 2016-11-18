package event_lib

import "time"
import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"

import "scal/storage"

type EventRevisionModel struct {
	EventId   bson.ObjectId `bson:"eventId" json:"eventId"`
	EventType string        `bson:"eventType" json:"eventType"`
	Sha1      string        `bson:"sha1" json:"sha1"`
	Time      time.Time     `bson:"time" json:"time"`
	//InvitedEmails []string    `bson:"invitedEmails" json:"invitedEmails"`
}

func (self EventRevisionModel) Collection() string {
	return storage.C_revision
}

func LoadLastRevisionModel(db *mgo.Database, eventId *bson.ObjectId) (
	*EventRevisionModel,
	error,
) {
	eventRev := EventRevisionModel{}
	err := db.C(storage.C_revision).Find(bson.M{
		"eventId": eventId,
	}).Sort("-time").One(&eventRev)
	return &eventRev, err
}
