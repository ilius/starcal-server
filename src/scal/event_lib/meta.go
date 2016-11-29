package event_lib

import (
	"errors"
	"log"
	"time"

	"gopkg.in/mgo.v2/bson"

	"scal"
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

func (self EventMetaModel) UniqueM() scal.M {
	return scal.M{
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
func (self EventMetaModel) JsonM() scal.M {
	return scal.M{
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
	db storage.Database,
	email string,
) string {
	// returns YES, NO, or MAYBE
	attendingModel, _ := LoadEventAttendingModel(db, self.EventId, email)
	return attendingModel.Attending
}
func (self *EventMetaModel) SetAttending(
	db storage.Database,
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
	db storage.Database,
	attending string,
) (int, error) {
	return db.FindCount(
		storage.C_attending,
		scal.M{
			"eventId":   self.EventId,
			"attending": attending,
		},
	)
}
func (self *EventMetaModel) Join(db storage.Database, email string) error {
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
func (self *EventMetaModel) Leave(db storage.Database, email string) error {
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
	db storage.Database,
	attending string,
) []string {
	emailStructs := []struct {
		Email string `bson:"email"`
	}{}
	err := db.FindAll(
		storage.C_attending,
		scal.M{
			"eventId":   self.EventId,
			"attending": attending,
		},
		&emailStructs,
	)
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
func (self *EventMetaModel) GetAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, YES)
}
func (self *EventMetaModel) GetNotAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, NO)
}
func (self *EventMetaModel) GetMaybeAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, MAYBE)
}

func LoadEventMetaModel(
	db storage.Database,
	eventId *bson.ObjectId,
	loadGroup bool,
) (*EventMetaModel, error) {
	var err error
	eventMeta := EventMetaModel{
		EventId: *eventId,
	}
	err = db.Get(&eventMeta)
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

type EventMetaChangeLogModel struct {
	Time     time.Time     `bson:"time"`
	Email    string        `bson:"email"`
	RemoteIp string        `bson:"remoteIp"`
	EventId  bson.ObjectId `bson:"eventId"`
	FuncName string        `bson:"funcName"`

	GroupId        *[2]*bson.ObjectId `bson:"groupId,omitempty"`
	OwnerEmail     *[2]*string        `bson:"ownerEmail,omitempty"`
	IsPublic       *[2]bool           `bson:"isPublic,omitempty"`
	AccessEmails   *[2][]string       `bson:"accessEmails,omitempty"`
	PublicJoinOpen *[2]bool           `bson:"publicJoinOpen,omitempty"`
	MaxAttendees   *[2]int            `bson:"maxAttendees,omitempty"`
}

func (model EventMetaChangeLogModel) Collection() string {
	return storage.C_eventMetaChangeLog
}

func GetEventMetaPipeResults(
	db storage.Database,
	pipeline *[]scal.M,
) (*[]scal.M, error) {
	results := []scal.M{}
	for res := range db.PipeIter(storage.C_eventMeta, pipeline) {
		if err := res.Err; err != nil {
			return nil, err
		}
		if eventId, ok := res.M["_id"]; ok {
			res.M["eventId"] = eventId
			delete(res.M, "_id")
		}
		if dataI, ok := res.M["data"]; ok {
			data := dataI.(scal.M)
			delete(data, "_id")
			res.M["data"] = data
		}
		results = append(results, res.M)
	}
	return &results, nil
}
