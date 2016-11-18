package event_lib

import (
	"errors"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"scal/storage"
)

/*
// PublicJoinPolicy
const (
    FreeToJoin = "FreeToJoin"
    NoJoin = "NoJoin"
    //JoinNeedsVerify = "JoinNeedsVerify" // makes implementation more complex
)
*/

type EventMetaModel struct {
	EventId      bson.ObjectId    `bson:"_id"`
	EventType    string           `bson:"eventType"`
	CreationTime time.Time        `bson:"creationTime"`
	OwnerEmail   string           `bson:"ownerEmail"`
	IsPublic     bool             `bson:"isPublic"`
	AccessEmails []string         `bson:"accessEmails"`
	GroupId      *bson.ObjectId   `bson:"groupId"`
	GroupModel   *EventGroupModel `bson:"-"`

	//PublicJoinPolicy string         `bson:"publicJoinPolicy"` // not indexed
	PublicJoinOpen bool `bson:"publicJoinOpen"`
	MaxAttendees   int  `bson:"maxAttendees"`
}

func (self EventMetaModel) UniqueM() bson.M {
	return bson.M{
		"_id": self.EventId,
	}
}
func (self EventMetaModel) Collection() string {
	return storage.C_eventMeta
}
func (self EventMetaModel) GroupIdHex() string {
	if self.GroupId != nil {
		return self.GroupId.Hex()
	}
	return ""
}
func (self EventMetaModel) JsonM() bson.M {
	return bson.M{
		"creationTime":   self.CreationTime,
		"ownerEmail":     self.OwnerEmail,
		"isPublic":       self.IsPublic,
		"accessEmails":   self.AccessEmails,
		"groupId":        self.GroupIdHex(),
		"publicJoinOpen": self.PublicJoinOpen,
		"maxAttendees":   self.MaxAttendees,
	}
}
func (self *EventMetaModel) CanReadFull(email string) bool {
	if email == self.OwnerEmail {
		return true
	}
	for _, aEmail := range self.AccessEmails {
		if email == aEmail {
			return true
		}
	}
	if self.GroupModel != nil {
		for _, aEmail := range self.GroupModel.ReadAccessEmails {
			if email == aEmail {
				return true
			}
		}
	}
	return false
}
func (self *EventMetaModel) CanRead(email string) bool {
	if self.IsPublic {
		return true
	}
	return self.CanReadFull(email)
}
func (self *EventMetaModel) GetAttending(
	db *mgo.Database,
	email string,
) string {
	// returns YES, NO, or MAYBE
	attendingModel, _ := LoadEventAttendingModel(db, self.EventId, email)
	return attendingModel.Attending
}
func (self *EventMetaModel) SetAttending(
	db *mgo.Database,
	email string,
	attending string,
) error {
	// attending: YES, NO, or MAYBE
	attendingModel, err := LoadEventAttendingModel(db, self.EventId, email)
	if err != nil {
		return err
	}
	attendingModel.Attending = attending
	attendingModel.ModifiedTime = time.Now()
	err = attendingModel.Save(db)
	return err
}
func (self *EventMetaModel) AttendingStatusCount(
	db *mgo.Database,
	attending string,
) (int, error) {
	return db.C(storage.C_attending).Find(bson.M{
		"eventId":   self.EventId,
		"attending": attending,
	}).Count()
}
func (self *EventMetaModel) Join(db *mgo.Database, email string) error {
	// does not make any changes on self
	if self.GetAttending(db, email) == YES {
		return errors.New("you have already joined this event")
	}
	if !self.CanReadFull(email) {
		if self.IsPublic {
			if !self.PublicJoinOpen {
				return errors.New("this public event is not open for joining")
			}
		} else {
			return errors.New("no access, no join")
		}
	}
	if self.MaxAttendees > 0 {
		attendingCount, err := self.AttendingStatusCount(db, YES)
		if err != nil {
			return err
		}
		if attendingCount >= self.MaxAttendees {
			return errors.New("maximum attendees exceeded, can not join event")
		}
	}
	self.SetAttending(db, email, YES)
	return nil
}
func (self *EventMetaModel) Leave(db *mgo.Database, email string) error {
	// does not make any changes on self
	if self.GetAttending(db, email) == NO {
		if self.CanReadFull(email) {
			return errors.New("you are not attending for this event")
		}
	}
	self.SetAttending(db, email, NO)
	return nil
}
func (self *EventMetaModel) GetEmailsByAttendingStatus(
	db *mgo.Database,
	attending string,
) []string {
	emailStructs := []struct {
		Email string `bson:"email"`
	}{}
	err := db.C(storage.C_attending).Find(bson.M{
		"eventId":   self.EventId,
		"attending": attending,
	}).All(&emailStructs)
	if err != nil {
		log.Printf(
			"Internal Error: GetAttendingEmails: eventId=%v: %s\n",
			self.EventId,
			err.Error(),
		)
	}
	emails := make([]string, len(emailStructs))
	for i, m := range emailStructs {
		emails[i] = m.Email
	}
	return emails
}
func (self *EventMetaModel) GetAttendingEmails(db *mgo.Database) []string {
	return self.GetEmailsByAttendingStatus(db, YES)
}
func (self *EventMetaModel) GetNotAttendingEmails(db *mgo.Database) []string {
	return self.GetEmailsByAttendingStatus(db, NO)
}
func (self *EventMetaModel) GetMaybeAttendingEmails(db *mgo.Database) []string {
	return self.GetEmailsByAttendingStatus(db, MAYBE)
}

func LoadEventMetaModel(
	db *mgo.Database,
	eventId *bson.ObjectId,
	loadGroup bool,
) (*EventMetaModel, error) {
	var err error
	eventMeta := EventMetaModel{
		EventId: *eventId,
	}
	err = storage.Get(db, &eventMeta)
	if err != nil {
		return nil, err
	}
	if loadGroup && eventMeta.GroupId != nil {
		// groupModel, err, internalErr
		groupModel, err, _ := LoadGroupModelById(
			"groupId",
			db,
			eventMeta.GroupId,
		)
		if err != nil {
			return nil, err
		}
		eventMeta.GroupModel = groupModel
	}
	return &eventMeta, nil
}
