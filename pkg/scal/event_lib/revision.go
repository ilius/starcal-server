package event_lib

import (
	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/storage"
	"time"

	"github.com/ilius/mgo/bson"
)

type EventRevisionModel struct {
	EventId   string    `bson:"eventId,objectid" json:"eventId"`
	EventType string    `bson:"eventType" json:"eventType"`
	Sha1      string    `bson:"sha1" json:"sha1"`
	Time      time.Time `bson:"time" json:"time"`
	// InvitedEmails []string    `bson:"invitedEmails" json:"invitedEmails"`
}

func (model EventRevisionModel) Collection() string {
	return storage.C_revision
}

func LoadLastRevisionModel(db storage.Database, eventIdHex *string) (
	*EventRevisionModel,
	error,
) {
	eventRev := EventRevisionModel{}
	eventId := bson.ObjectIdHex(*eventIdHex)
	err := db.First(
		scal.M{
			"eventId": eventId,
		},
		"-time",
		&eventRev,
	)
	return &eventRev, err
}
