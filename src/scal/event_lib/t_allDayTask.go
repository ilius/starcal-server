package event_lib

import "scal/storage"

// DurationEnable is just a matter of UI

type AllDayTaskEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	StartJd        int  `bson:"startJd" json:"startJd"`
	EndJd          int  `bson:"endJd" json:"endJd"`
	DurationEnable bool `bson:"durationEnable" json:"durationEnable"`
}

func (AllDayTaskEventModel) Type() string {
	return "allDayTask"
}

func LoadAllDayTaskEventModel(db storage.Database, sha1 string) (
	*AllDayTaskEventModel,
	error,
) {
	model := AllDayTaskEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type AllDayTaskEvent struct {
	BaseEvent
	startJd        int
	endJd          int
	durationEnable bool
}

func (AllDayTaskEvent) Type() string {
	return "allDayTask"
}

func (event AllDayTaskEvent) StartJd() int {
	return event.startJd
}

func (event AllDayTaskEvent) EndJd() int {
	return event.endJd
}

func (event AllDayTaskEvent) DurationEnable() bool {
	return event.durationEnable
}

func (event AllDayTaskEvent) Model() AllDayTaskEventModel {
	return AllDayTaskEventModel{
		BaseEventModel: event.BaseModel(),
		StartJd:        event.startJd,
		EndJd:          event.endJd,
		DurationEnable: event.durationEnable,
	}
}

func (model AllDayTaskEventModel) GetEvent() (AllDayTaskEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return AllDayTaskEvent{}, err
	}
	return AllDayTaskEvent{
		BaseEvent:      baseEvent,
		startJd:        model.StartJd,
		endJd:          model.EndJd,
		durationEnable: model.DurationEnable,
	}, nil
}
