package event_lib

import (
	//"errors"
	"fmt"
	"time"

	"scal"
	"scal/storage"
)

type EventAttendingModel struct {
	//Id *string           `bson:"_id,objectid"` // omitempty??
	EventId      string    `bson:"eventId,objectid"`
	Email        string    `bson:"email"`
	Attending    string    `bson:"attending"` // YES, NO, MAYBE, UNKNOWN
	ModifiedTime time.Time `bson:"modifiedTime"`
}

func (model EventAttendingModel) UniqueM() scal.M {
	return scal.M{
		"eventId": model.EventId,
		"email":   model.Email,
	}
}
func (EventAttendingModel) Collection() string {
	return storage.C_attending
}
func (model *EventAttendingModel) Save(db storage.Database) error {
	if model.Attending == UNKNOWN {
		return db.Remove(model)
	}
	return db.Upsert(model)
	// if model.Id == nil {
	// 	return db.Insert(model)
	// } else {
	// 	return db.Update(model)
	// }
}

func LoadEventAttendingModel(
	db storage.Database,
	eventId string,
	email string,
) (EventAttendingModel, error) {
	attendingModel := EventAttendingModel{
		EventId: eventId,
		Email:   email,
	}
	err := db.Get(&attendingModel)
	if db.IsNotFound(err) {
		attendingModel.Attending = UNKNOWN
		attendingModel.ModifiedTime = time.Now()
		err = nil
	}
	if err != nil {
		log.Error(fmt.Sprintf(
			"Internal Error: fetching EventAttendingModel{%v, %v}: %s\n",
			eventId,
			email,
			err.Error(),
		))
	}
	return attendingModel, err
}
