package event_lib

type DailyNoteEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	Jd             int `bson:"jd" json:"jd"`
}

func (DailyNoteEventModel) Type() string {
	return "dailyNote"
}

type DailyNoteEvent struct {
	BaseEvent
	jd int
}

func (DailyNoteEvent) Type() string {
	return "dailyNote"
}

func (event DailyNoteEvent) Jd() int {
	return event.jd
}

func (event DailyNoteEvent) Model() DailyNoteEventModel {
	return DailyNoteEventModel{
		BaseEventModel: event.BaseModel(),
		Jd:             event.jd,
	}
}

func (model DailyNoteEventModel) GetEvent() (DailyNoteEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return DailyNoteEvent{}, err
	}
	return DailyNoteEvent{
		BaseEvent: baseEvent,
		jd:        model.Jd,
	}, nil
}
