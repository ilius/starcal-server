package event_lib

import (
	"fmt"
	"time"

	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/storage"

	"github.com/ilius/libgostarcal/cal_types"
)

type BaseEventModel struct {
	DummyType      string `bson:"-" json:"eventType"`
	Id             string `bson:"-" json:"eventId,omitempty"`
	Sha1           string `bson:"sha1" json:"sha1,omitempty"`
	TimeZone       string `bson:"timeZone,omitempty" json:"timeZone"`
	TimeZoneEnable bool   `bson:"timeZoneEnable" json:"timeZoneEnable"`
	CalType        string `bson:"calType" json:"calType"`
	Summary        string `bson:"summary" json:"summary"`
	Description    string `bson:"description,omitempty" json:"description"`
	Icon           string `bson:"icon,omitempty" json:"icon"`

	SummaryEncrypted     bool `bson:"summaryEncrypted,omitempty" json:"summaryEncrypted"`
	DescriptionEncrypted bool `bson:"descriptionEncrypted,omitempty" json:"descriptionEncrypted"`

	// NotifyBefore: seconds, default 0
	// NotifyBefore int `bson:"notifyBefore,omitempty" json:"notifyBefore"`
	// IsAllDay bool

	GroupId string `bson:"-" json:"groupId"` // FIXME
	Meta    scal.M `bson:"-" json:"meta"`
}

func (BaseEventModel) Collection() string {
	return storage.C_eventData
}

func (model BaseEventModel) UniqueM() scal.M {
	return scal.M{
		"sha1": model.Sha1,
	}
}

func LoadBaseEventModel(db storage.Database, sha1 string) (
	*BaseEventModel,
	error,
) {
	model := BaseEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type BaseEvent struct {
	id string
	// ownerEmail string
	loc          *time.Location
	locEnable    bool
	calType      cal_types.CalType
	summary      string
	description  string
	icon         string
	notifyBefore int // seconds
}

func (event BaseEvent) String() string {
	return fmt.Sprintf(
		"Event(id: %x, summary: %v, loc: %v, locEnable: %v)",
		event.id,
		event.summary,
		event.loc,
		event.locEnable,
	)
}

func (event BaseEvent) Id() string {
	return event.id
}

// func (event BaseEvent) OwnerEmail() string {
//    return event.ownerEmail
//}
func (event BaseEvent) Location() *time.Location {
	if event.locEnable && event.loc != nil {
		return event.loc
	}
	// FIXME
	// return time.Now().Location()
	return time.UTC
}

func (event BaseEvent) CalType() cal_types.CalType {
	return event.calType
}

func (event BaseEvent) Summary() string {
	return event.summary
}

func (event BaseEvent) Description() string {
	return event.description
}

func (event BaseEvent) Icon() string {
	return event.icon
}

func (event BaseEvent) NotifyBefore() int {
	return event.notifyBefore
}

func (event BaseEvent) BaseModel() BaseEventModel {
	return BaseEventModel{
		Id:             event.id,
		TimeZone:       event.loc.String(),
		TimeZoneEnable: event.locEnable,
		CalType:        event.calType.Name(),
		Summary:        event.summary,
		Description:    event.description,
		Icon:           event.icon,
		// NotifyBefore:   event.notifyBefore,
	}
}

func (model BaseEventModel) GetBaseEvent() (BaseEvent, error) {
	var loc *time.Location
	var err error
	locEnable := model.TimeZoneEnable
	if model.TimeZone == "" {
		loc = nil // FIXME
		locEnable = false
	} else {
		loc, err = time.LoadLocation(model.TimeZone)
		// does time.LoadLocation cache Location structs? FIXME
		if err != nil {
			return BaseEvent{}, err
		}
	}
	calType, err2 := cal_types.GetCalType(model.CalType)
	if err2 != nil {
		return BaseEvent{}, err2
	}
	return BaseEvent{
		id: string(model.Id),
		// ownerEmail: model.OwnerEmail,
		loc:         loc,
		locEnable:   locEnable,
		calType:     calType,
		summary:     model.Summary,
		description: model.Description,
		icon:        model.Icon,
		// notifyBefore: model.NotifyBefore,
	}, nil
}
