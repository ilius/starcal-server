package event_lib

import "scal/storage"

type LifeTimeEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	StartJd        int `bson:"startJd" json:"startJd"`
	EndJd          int `bson:"endJd" json:"endJd"`
}

func (LifeTimeEventModel) Type() string {
	return "lifeTime"
}

func LoadLifeTimeEventModel(db storage.Database, sha1 string) (
	*LifeTimeEventModel,
	error,
) {
	model := LifeTimeEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type LifeTimeEvent struct {
	BaseEvent
	startJd int
	endJd   int
}

func (event LifeTimeEvent) Type() string {
	return "lifeTime"
}

func (event LifeTimeEvent) StartJd() int {
	return event.startJd
}

func (event LifeTimeEvent) EndJd() int {
	return event.endJd
}

func (event LifeTimeEvent) Model() LifeTimeEventModel {
	return LifeTimeEventModel{
		BaseEventModel: event.BaseModel(),
		StartJd:        event.startJd,
		EndJd:          event.endJd,
	}
}

func (model LifeTimeEventModel) GetEvent() (LifeTimeEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return LifeTimeEvent{}, err
	}
	return LifeTimeEvent{
		BaseEvent: baseEvent,
		startJd:   model.StartJd,
		endJd:     model.EndJd,
	}, nil
}
