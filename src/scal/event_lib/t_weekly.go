package event_lib

import (
	"scal/storage"

	lib "github.com/ilius/libgostarcal"
	"github.com/ilius/libgostarcal/utils"
)

type WeeklyEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`

	StartJd int `bson:"startJd" json:"startJd"`
	EndJd   int `bson:"endJd" json:"endJd"`

	CycleWeeks      uint   `bson:"cycleWeeks" json:"cycleWeeks"`
	DayStartSeconds uint32 `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   uint32 `bson:"dayEndSeconds" json:"dayEndSeconds"`
}

func (WeeklyEventModel) Type() string {
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
	cycleWeeks      uint
	dayStartSeconds uint32
	dayEndSeconds   uint32
}

func (WeeklyEvent) Type() string {
	return "weekly"
}

func (event WeeklyEvent) StartJd() int {
	return event.startJd
}

func (event WeeklyEvent) EndJd() int {
	return event.endJd
}

func (event WeeklyEvent) CycleWeeks() uint {
	return event.cycleWeeks
}

func (event WeeklyEvent) DayStartSeconds() uint32 {
	return event.dayStartSeconds
}

func (event WeeklyEvent) DayEndSeconds() uint32 {
	return event.dayEndSeconds
}

func (event WeeklyEvent) DayStartHMS() lib.HMS {
	return utils.GetHmsBySeconds(int(event.dayStartSeconds))
}

func (event WeeklyEvent) DayEndHMS() lib.HMS {
	return utils.GetHmsBySeconds(int(event.dayEndSeconds))
}

func (event WeeklyEvent) Model() WeeklyEventModel {
	return WeeklyEventModel{
		BaseEventModel:  event.BaseModel(),
		StartJd:         event.startJd,
		EndJd:           event.endJd,
		CycleWeeks:      event.cycleWeeks,
		DayStartSeconds: event.dayStartSeconds,
		DayEndSeconds:   event.dayEndSeconds,
	}
}

func (model WeeklyEventModel) GetEvent() (WeeklyEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return WeeklyEvent{}, err
	}
	return WeeklyEvent{
		BaseEvent:       baseEvent,
		startJd:         model.StartJd,
		endJd:           model.EndJd,
		cycleWeeks:      model.CycleWeeks,
		dayStartSeconds: model.DayStartSeconds,
		dayEndSeconds:   model.DayEndSeconds,
	}, nil
}
