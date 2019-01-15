package event_lib

import "scal/storage"

type DailyNoteEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	Jd             int `bson:"jd" json:"jd"`
}

func (DailyNoteEventModel) Type() string {
	return "dailyNote"
}

func LoadDailyNoteEventModel(db storage.Database, sha1 string) (
	*DailyNoteEventModel,
	error,
) {
	model := DailyNoteEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
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
