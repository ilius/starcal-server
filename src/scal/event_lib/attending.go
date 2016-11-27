package event_lib

import (
	//"errors"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"scal/storage"
)

type EventAttendingModel struct {
	//Id *bson.ObjectId           `bson:"_id,omitempty"`
	EventId      bson.ObjectId `bson:"eventId"`
	Email        string        `bson:"email"`
	Attending    string        `bson:"attending"` // YES, NO, MAYBE, UNKNOWN
	ModifiedTime time.Time     `bson:"modifiedTime"`
}

func (self EventAttendingModel) UniqueM() bson.M {
	return bson.M{
		"eventId": self.EventId,
		"email":   self.Email,
	}
}
func (self EventAttendingModel) Collection() string {
	return storage.C_attending
}
func (self *EventAttendingModel) Save(db *storage.MongoDatabase) error {
	if self.Attending == UNKNOWN {
		return db.Remove(self)
	}
	return db.Upsert(self)
	/*if self.Id == nil {
	      return db.Insert(self)
	  } else {
	      return db.Update(self)
	  }*/
}

func LoadEventAttendingModel(
	db *storage.MongoDatabase,
	eventId bson.ObjectId,
	email string,
) (EventAttendingModel, error) {
	attendingModel := EventAttendingModel{
		EventId: eventId,
		Email:   email,
	}
	err := db.Get(&attendingModel)
	if err == mgo.ErrNotFound {
		attendingModel.Attending = UNKNOWN
		attendingModel.ModifiedTime = time.Now()
		err = nil
	}
	if err != nil {
		log.Printf(
			"Internal Error: fetching EventAttendingModel{%v, %v}: %s\n",
			eventId,
			email,
			err.Error(),
		)
	}
	return attendingModel, err
}
