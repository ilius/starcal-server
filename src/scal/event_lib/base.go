package event_lib

import "time"
import "fmt"
//import "errors"

import "gopkg.in/mgo.v2-unstable/bson"
import "scal/cal_types"

type BaseEventModel struct {
    Id bson.ObjectId    `bson:"_id,omitempty"`
    OwnerId int         `bson:"ownerId"`
    TimeZone string     `bson:"timeZone,omitempty"`
    TimeZoneEnable bool `bson:"timeZoneEnable"`
    CalType string      `bson:"calType"`
    Summary string      `bson:"summary"`
    Description string  `bson:"description,omitempty"`
    Icon string         `bson:"icon,omitempty"`
    NotifyBefore int    `bson:"notifyBefore,omitempty"` // seconds, default 0
    //IsAllDay bool
}
func (self BaseEventModel) String() string {
    return fmt.Sprintf("EventModel(Id=%v, Summary=%v)", self.Id, self.Summary)
}



type BaseEvent struct {
    id string
    ownerId int
    loc *time.Location
    locEnable bool
    calType *cal_types.CalType
    summary string
    description string
    icon string
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
func (self BaseEvent) OwnerId() int {
    return self.ownerId
}
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
        Id: bson.ObjectId(self.id),
        OwnerId: self.ownerId,
        TimeZone: self.loc.String(),
        TimeZoneEnable: self.locEnable,
        CalType: self.calType.Name,
        Summary: self.summary,
        Description: self.description,
        Icon: self.icon,
        NotifyBefore: self.notifyBefore,
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
        ownerId: self.OwnerId,
        loc: loc,
        locEnable: locEnable,
        calType: calType,
        summary: self.Summary,
        description: self.Description,
        icon: self.Icon,
        notifyBefore: self.NotifyBefore,
    }, nil
}



