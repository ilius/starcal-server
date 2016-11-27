package event_lib

import "scal/storage"

type DailyNoteEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	Jd             int `bson:"jd" json:"jd"`
}

func (self DailyNoteEventModel) Type() string {
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

func (self DailyNoteEvent) Type() string {
	return "dailyNote"
}
func (self DailyNoteEvent) Jd() int {
	return self.jd
}

func (self DailyNoteEvent) Model() DailyNoteEventModel {
	return DailyNoteEventModel{
		BaseEventModel: self.BaseModel(),
		Jd:             self.jd,
	}
}
func (self DailyNoteEventModel) GetEvent() (DailyNoteEvent, error) {
	baseEvent, err := self.BaseEventModel.GetBaseEvent()
	if err != nil {
		return DailyNoteEvent{}, err
	}
	return DailyNoteEvent{
		BaseEvent: baseEvent,
		jd:        self.Jd,
	}, nil
}
