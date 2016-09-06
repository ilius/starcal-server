package event_lib

import "scal"
import . "scal/utils"

type MonthlyEventModel struct {
    BaseEventModel          `bson:",inline"`
    StartJd int             `bson:"startJd"`
    EndJd int               `bson:"endJd"`
    Day int                 `bson:"day"`
    DayStartSeconds int     `bson:"dayStartSeconds"`
    DayEndSeconds   int     `bson:"dayEndSeconds"`
}

type MonthlyEvent struct {
    BaseEvent
    startJd int
    endJd int
    day int
    dayStartSeconds int
    dayEndSeconds int
}
func (self MonthlyEvent) Type() string {
    return "monthly"
}
func (self MonthlyEvent) StartJd() int {
    return self.startJd
}
func (self MonthlyEvent) EndJd() int {
    return self.endJd
}
func (self MonthlyEvent) Day() int {
    return self.day
}
func (self MonthlyEvent) DayStartSeconds() int {
    return self.dayStartSeconds
}
func (self MonthlyEvent) DayEndSeconds() int {
    return self.dayEndSeconds
}
func (self MonthlyEvent) DayStartHMS() scal.HMS {
    return GetHmsBySeconds(self.dayStartSeconds)
}
func (self MonthlyEvent) DayEndHMS() scal.HMS {
    return GetHmsBySeconds(self.dayEndSeconds)
}





func (self MonthlyEvent) Model() MonthlyEventModel {
    return MonthlyEventModel{
        BaseEventModel: self.BaseModel(),
        StartJd: self.startJd,
        EndJd: self.endJd,
        Day: self.day,
        DayStartSeconds: self.dayStartSeconds,
        DayEndSeconds: self.dayEndSeconds,
    }
}
func (self MonthlyEventModel) GetEvent() (MonthlyEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return MonthlyEvent{}, err
    }
    return MonthlyEvent{
        BaseEvent: baseEvent,
        startJd: self.StartJd,
        endJd: self.EndJd,
        day: self.Day,
        dayStartSeconds: self.DayStartSeconds,
        dayEndSeconds: self.DayEndSeconds,
    }, nil
}





