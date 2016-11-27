package event_lib

import "scal"
import . "scal/utils"
import "scal/storage"

type WeeklyEventModel struct {
	BaseEventModel  `bson:",inline" json:",inline"`
	StartJd         int `bson:"startJd" json:"startJd"`
	EndJd           int `bson:"endJd" json:"endJd"`
	CycleWeeks      int `bson:"cycleWeeks" json:"cycleWeeks"`
	DayStartSeconds int `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   int `bson:"dayEndSeconds" json:"dayEndSeconds"`
}

func (self WeeklyEventModel) Type() string {
	return "weekly"
}

func LoadWeeklyEventModel(db storage.Database, sha1 string) (
	*WeeklyEventModel,
	error,
) {
	model := WeeklyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type WeeklyEvent struct {
	BaseEvent
	startJd         int
	endJd           int
	cycleWeeks      int
	dayStartSeconds int
	dayEndSeconds   int
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
		BaseEventModel:  self.BaseModel(),
		StartJd:         self.startJd,
		EndJd:           self.endJd,
		CycleWeeks:      self.cycleWeeks,
		DayStartSeconds: self.dayStartSeconds,
		DayEndSeconds:   self.dayEndSeconds,
	}
}
func (self WeeklyEventModel) GetEvent() (WeeklyEvent, error) {
	baseEvent, err := self.BaseEventModel.GetBaseEvent()
	if err != nil {
		return WeeklyEvent{}, err
	}
	return WeeklyEvent{
		BaseEvent:       baseEvent,
		startJd:         self.StartJd,
		endJd:           self.EndJd,
		cycleWeeks:      self.CycleWeeks,
		dayStartSeconds: self.DayStartSeconds,
		dayEndSeconds:   self.DayEndSeconds,
	}, nil
}
