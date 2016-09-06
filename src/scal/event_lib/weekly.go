package event_lib

import "scal"
import . "scal/utils"

type WeeklyEventModel struct {
    BaseEventModel          `bson:",inline"`
    StartJd int             `bson:"startJd"`
    EndJd int               `bson:"endJd"`
    CycleWeeks int          `bson:"cycleWeeks"`
    DayStartSeconds int     `bson:"dayStartSeconds"`
    DayEndSeconds int       `bson:"dayEndSeconds"`
}

type WeeklyEvent struct {
    BaseEvent
    startJd int
    endJd int
    cycleWeeks int
    dayStartSeconds int
    dayEndSeconds int
}
func (self WeeklyEvent) Type() string {
    return "weekly"
}
func (self WeeklyEvent) StartJd() int {
    return self.startJd
}
func (self WeeklyEvent) EndJd() int {
    return self.endJd
}
func (self WeeklyEvent) CycleWeeks() int {
    return self.cycleWeeks
}
func (self WeeklyEvent) DayStartSeconds() int {
    return self.dayStartSeconds
}
func (self WeeklyEvent) DayEndSeconds() int {
    return self.dayEndSeconds
}
func (self WeeklyEvent) DayStartHMS() scal.HMS {
    return GetHmsBySeconds(self.dayStartSeconds)
}
func (self WeeklyEvent) DayEndHMS() scal.HMS {
    return GetHmsBySeconds(self.dayEndSeconds)
}





func (self WeeklyEvent) Model() WeeklyEventModel {
    return WeeklyEventModel{
        BaseEventModel: self.BaseModel(),
        StartJd: self.startJd,
        EndJd: self.endJd,
        CycleWeeks: self.cycleWeeks,
        DayStartSeconds: self.dayStartSeconds,
        DayEndSeconds: self.dayEndSeconds,
    }
}
func (self WeeklyEventModel) GetEvent() (WeeklyEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return WeeklyEvent{}, err
    }
    return WeeklyEvent{
        BaseEvent: baseEvent,
        startJd: self.StartJd,
        endJd: self.EndJd,
        cycleWeeks: self.CycleWeeks,
        dayStartSeconds: self.DayStartSeconds,
        dayEndSeconds: self.DayEndSeconds,
    }, nil
}





