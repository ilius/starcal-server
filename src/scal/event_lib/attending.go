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
    EventId bson.ObjectId       `bson:"eventId"`
    Email string                `bson:"email"`
    Attending string            `bson:"attending"`// YES, NO, MAYBE, UNKNOWN
    ModifiedTime time.Time      `bson:"modifiedTime"`
}
func (self EventAttendingModel) UniqueM() bson.M {
    return bson.M{
        "eventId": self.EventId,
        "email": self.Email,
    }
}
func (self EventAttendingModel) Collection() string {
    return storage.C_attending
}
func (self *EventAttendingModel) Save(db *mgo.Database) error {
    if self.Attending == UNKNOWN {
        return storage.Remove(db, self)
    }
    return storage.Upsert(db, self)
    /*if self.Id == nil {
        return storage.Insert(db, self)
    } else {
        return storage.Update(db, self)
    }*/
}

func LoadEventAttendingModel(
    db *mgo.Database,
    eventId bson.ObjectId,
    email string,
) (EventAttendingModel, error) {
    attendingModel := EventAttendingModel{}
    err := db.C(storage.C_attending).Find(bson.M{
        "eventId": eventId,
        "email": email,
    }).One(&attendingModel)
    if err == mgo.ErrNotFound {
        attendingModel.EventId = eventId
        attendingModel.Email = email
        attendingModel.Attending = UNKNOWN
        //attendingModel.ModifiedTime = time.Now()
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

