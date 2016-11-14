package event_lib

import "gopkg.in/mgo.v2"
import "scal"
import . "scal/utils"
import "scal/storage"

type MonthlyEventModel struct {
    BaseEventModel          `bson:",inline" json:",inline"`
    StartJd int             `bson:"startJd" json:"startJd"`
    EndJd int               `bson:"endJd" json:"endJd"`
    Day int                 `bson:"day" json:"day"`
    DayStartSeconds int     `bson:"dayStartSeconds" json:"dayStartSeconds"`
    DayEndSeconds   int     `bson:"dayEndSeconds" json:"dayEndSeconds"`
}
func (self MonthlyEventModel) Type() string {
    return "monthly"
}

func LoadMonthlyEventModel(db *mgo.Database, sha1 string) (
    *MonthlyEventModel,
    error,
) {
    model := MonthlyEventModel{}
    model.Sha1 = sha1
    err := storage.Get(db, &model)
    return &model, err
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




