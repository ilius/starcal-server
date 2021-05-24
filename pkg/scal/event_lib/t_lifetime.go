package event_lib

type LifetimeEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	StartJd        int `bson:"startJd" json:"startJd"`
	EndJd          int `bson:"endJd" json:"endJd"`
}

func (LifetimeEventModel) Type() string {
	return "lifetime"
}

type LifetimeEvent struct {
	BaseEvent
	startJd int
	endJd   int
}

func (event LifetimeEvent) Type() string {
	return "lifetime"
}

func (event LifetimeEvent) StartJd() int {
	return event.startJd
}

func (event LifetimeEvent) EndJd() int {
	return event.endJd
}

func (event LifetimeEvent) Model() LifetimeEventModel {
	return LifetimeEventModel{
		BaseEventModel: event.BaseModel(),
		StartJd:        event.startJd,
		EndJd:          event.endJd,
	}
}

func (model LifetimeEventModel) GetEvent() (LifetimeEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return LifetimeEvent{}, err
	}
	return LifetimeEvent{
		BaseEvent: baseEvent,
		startJd:   model.StartJd,
		endJd:     model.EndJd,
	}, nil
}
