package event_lib

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"

	"scal/cal_types"
	"scal/storage"
)

type BaseEventModel struct {
	Id             bson.ObjectId `bson:"-" json:"eventId,omitempty"`
	Sha1           string        `bson:"sha1" json:"sha1,omitempty"`
	TimeZone       string        `bson:"timeZone,omitempty" json:"timeZone"`
	TimeZoneEnable bool          `bson:"timeZoneEnable" json:"timeZoneEnable"`
	CalType        string        `bson:"calType" json:"calType"`
	Summary        string        `bson:"summary" json:"summary"`
	Description    string        `bson:"description,omitempty" json:"description"`
	Icon           string        `bson:"icon,omitempty" json:"icon"`
	NotifyBefore   int           `bson:"notifyBefore,omitempty" json:"notifyBefore"` // seconds, default 0
	//IsAllDay bool
	GroupId string `bson:"-" json:"groupId"` // FIXME
	Meta    bson.M `bson:"-" json:"meta"`
}

func (self BaseEventModel) Collection() string {
	return storage.C_eventData
}
func (self BaseEventModel) UniqueM() bson.M {
	return bson.M{
		"sha1": self.Sha1,
	}
}

type BaseEvent struct {
	id string
	//ownerEmail string
	loc          *time.Location
	locEnable    bool
	calType      *cal_types.CalType
	summary      string
	description  string
	icon         string
	notifyBefore int // seconds
}

func (self BaseEvent) String() string {
	return fmt.Sprintf(
		"Event(id: %x, summary: %v, loc: %v, locEnable: %v)",
		self.id,
		self.summary,
		self.loc,
		self.locEnable,
	)
}
func (self BaseEvent) Id() string {
	return self.id
}

//func (self BaseEvent) OwnerEmail() string {
//    return self.ownerEmail
//}
func (self BaseEvent) Location() *time.Location {
	if self.locEnable && self.loc != nil {
		return self.loc
	}
	// FIXME
	//return time.Now().Location()
	return time.UTC
}
func (self BaseEvent) CalType() *cal_types.CalType {
	return self.calType
}
func (self BaseEvent) Summary() string {
	return self.summary
}
func (self BaseEvent) Description() string {
	return self.description
}
func (self BaseEvent) Icon() string {
	return self.icon
}
func (self BaseEvent) NotifyBefore() int {
	return self.notifyBefore
}

func (self BaseEvent) BaseModel() BaseEventModel {
	return BaseEventModel{
		Id:             bson.ObjectId(self.id),
		TimeZone:       self.loc.String(),
		TimeZoneEnable: self.locEnable,
		CalType:        self.calType.Name,
		Summary:        self.summary,
		Description:    self.description,
		Icon:           self.icon,
		NotifyBefore:   self.notifyBefore,
	}
}
func (self BaseEventModel) GetBaseEvent() (BaseEvent, error) {
	var loc *time.Location
	var err error
	locEnable := self.TimeZoneEnable
	if self.TimeZone == "" {
		loc = nil // FIXME
		locEnable = false
	} else {
		loc, err = time.LoadLocation(self.TimeZone)
		// does time.LoadLocation cache Location structs? FIXME
		if err != nil {
			return BaseEvent{}, err
		}
	}
	calType, err2 := cal_types.GetCalType(self.CalType)
	if err2 != nil {
		return BaseEvent{}, err2
	}
	return BaseEvent{
		id: string(self.Id),
		//ownerEmail: self.OwnerEmail,
		loc:          loc,
		locEnable:    locEnable,
		calType:      calType,
		summary:      self.Summary,
		description:  self.Description,
		icon:         self.Icon,
		notifyBefore: self.NotifyBefore,
	}, nil
}
