package event_lib

import "time"

import "scal"
import "scal/storage"

type EventRevisionModel struct {
	EventId   string    `bson:"eventId,objectid" json:"eventId"`
	EventType string    `bson:"eventType" json:"eventType"`
	Sha1      string    `bson:"sha1" json:"sha1"`
	Time      time.Time `bson:"time" json:"time"`
	//InvitedEmails []string    `bson:"invitedEmails" json:"invitedEmails"`
}

func (model EventRevisionModel) Collection() string {
	return storage.C_revision
}

func LoadLastRevisionModel(db storage.Database, eventId *string) (
	*EventRevisionModel,
	error,
) {
	eventRev := EventRevisionModel{}
	err := db.First(
		scal.M{
			"eventId": eventId,
		},
		"-time",
		&eventRev,
	)
	return &eventRev, err
}
