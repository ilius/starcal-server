package event_lib

import (
	lib "github.com/ilius/libgostarcal"

	"github.com/ilius/libgostarcal/utils"
)

type MonthlyEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`

	StartJd int `bson:"startJd" json:"startJd"`
	EndJd   int `bson:"endJd" json:"endJd"`

	Day uint8 `bson:"day" json:"day"`

	DayStartSeconds uint32 `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   uint32 `bson:"dayEndSeconds" json:"dayEndSeconds"`
}

func (MonthlyEventModel) Type() string {
	return "monthly"
}

type MonthlyEvent struct {
	BaseEvent
	startJd         int
	endJd           int
	day             uint8
	dayStartSeconds uint32
	dayEndSeconds   uint32
}

func (MonthlyEvent) Type() string {
	return "monthly"
}

func (event MonthlyEvent) StartJd() int {
	return event.startJd
}

func (event MonthlyEvent) EndJd() int {
	return event.endJd
}

func (event MonthlyEvent) Day() uint8 {
	return event.day
}

func (event MonthlyEvent) DayStartSeconds() uint32 {
	return event.dayStartSeconds
}

func (event MonthlyEvent) DayEndSeconds() uint32 {
	return event.dayEndSeconds
}

func (event MonthlyEvent) DayStartHMS() lib.HMS {
	return utils.GetHmsBySeconds(uint(event.dayStartSeconds))
}

func (event MonthlyEvent) DayEndHMS() lib.HMS {
	return utils.GetHmsBySeconds(uint(event.dayEndSeconds))
}

func (event MonthlyEvent) Model() MonthlyEventModel {
	return MonthlyEventModel{
		BaseEventModel:  event.BaseModel(),
		StartJd:         event.startJd,
		EndJd:           event.endJd,
		Day:             event.day,
		DayStartSeconds: event.dayStartSeconds,
		DayEndSeconds:   event.dayEndSeconds,
	}
}

func (model MonthlyEventModel) GetEvent() (MonthlyEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return MonthlyEvent{}, err
	}
	return MonthlyEvent{
		BaseEvent:       baseEvent,
		startJd:         model.StartJd,
		endJd:           model.EndJd,
		day:             model.Day,
		dayStartSeconds: model.DayStartSeconds,
		dayEndSeconds:   model.DayEndSeconds,
	}, nil
}
