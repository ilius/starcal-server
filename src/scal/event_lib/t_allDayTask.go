package event_lib

import "scal/storage"

// DurationEnable is just a matter of UI

type AllDayTaskEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	StartJd        int  `bson:"startJd" json:"startJd"`
	EndJd          int  `bson:"endJd" json:"endJd"`
	DurationEnable bool `bson:"durationEnable" json:"durationEnable"`
}

func (self AllDayTaskEventModel) Type() string {
	return "allDayTask"
}

func LoadAllDayTaskEventModel(db *storage.MongoDatabase, sha1 string) (
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

func (self AllDayTaskEvent) Type() string {
	return "allDayTask"
}
func (self AllDayTaskEvent) StartJd() int {
	return self.startJd
}
func (self AllDayTaskEvent) EndJd() int {
	return self.endJd
}
func (self AllDayTaskEvent) DurationEnable() bool {
	return self.durationEnable
}

func (self AllDayTaskEvent) Model() AllDayTaskEventModel {
	return AllDayTaskEventModel{
		BaseEventModel: self.BaseModel(),
		StartJd:        self.startJd,
		EndJd:          self.endJd,
		DurationEnable: self.durationEnable,
	}
}
func (self AllDayTaskEventModel) GetEvent() (AllDayTaskEvent, error) {
	baseEvent, err := self.BaseEventModel.GetBaseEvent()
	if err != nil {
		return AllDayTaskEvent{}, err
	}
	return AllDayTaskEvent{
		BaseEvent:      baseEvent,
		startJd:        self.StartJd,
		endJd:          self.EndJd,
		durationEnable: self.DurationEnable,
	}, nil
}
