package event_lib

import "scal"
import . "scal/utils"
import "scal/storage"

type MonthlyEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`

	StartJd int `bson:"startJd" json:"startJd"`
	EndJd   int `bson:"endJd" json:"endJd"`

	Day uint8 `bson:"day" json:"day"`

	DayStartSeconds uint32 `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   uint32 `bson:"dayEndSeconds" json:"dayEndSeconds"`
}

func (self MonthlyEventModel) Type() string {
	return "monthly"
}

func LoadMonthlyEventModel(db storage.Database, sha1 string) (
	*MonthlyEventModel,
	error,
) {
	model := MonthlyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type MonthlyEvent struct {
	BaseEvent
	startJd         int
	endJd           int
	day             uint8
	dayStartSeconds uint32
	dayEndSeconds   uint32
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
func (self MonthlyEvent) Day() uint8 {
	return self.day
}
func (self MonthlyEvent) DayStartSeconds() uint32 {
	return self.dayStartSeconds
}
func (self MonthlyEvent) DayEndSeconds() uint32 {
	return self.dayEndSeconds
}
func (self MonthlyEvent) DayStartHMS() scal.HMS {
	return GetHmsBySeconds(int(self.dayStartSeconds))
}
func (self MonthlyEvent) DayEndHMS() scal.HMS {
	return GetHmsBySeconds(int(self.dayEndSeconds))
}

func (self MonthlyEvent) Model() MonthlyEventModel {
	return MonthlyEventModel{
		BaseEventModel:  self.BaseModel(),
		StartJd:         self.startJd,
		EndJd:           self.endJd,
		Day:             self.day,
		DayStartSeconds: self.dayStartSeconds,
		DayEndSeconds:   self.dayEndSeconds,
	}
}
func (self MonthlyEventModel) GetEvent() (MonthlyEvent, error) {
	baseEvent, err := self.BaseEventModel.GetBaseEvent()
	if err != nil {
		return MonthlyEvent{}, err
	}
	return MonthlyEvent{
		BaseEvent:       baseEvent,
		startJd:         self.StartJd,
		endJd:           self.EndJd,
		day:             self.Day,
		dayStartSeconds: self.DayStartSeconds,
		dayEndSeconds:   self.DayEndSeconds,
	}, nil
}
