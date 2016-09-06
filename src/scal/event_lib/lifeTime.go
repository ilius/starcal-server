package event_lib

type LifeTimeEventModel struct {
    BaseEventModel  `bson:",inline"`
    StartJd int     `bson:"startJd"`
    EndJd int       `bson:"endJd"`
}

type LifeTimeEvent struct {
    BaseEvent
    startJd int
    endJd int
}
func (self LifeTimeEvent) Type() string {
    return "lifeTime"
}
func (self LifeTimeEvent) StartJd() int {
    return self.startJd
}
func (self LifeTimeEvent) EndJd() int {
    return self.endJd
}


func (self LifeTimeEvent) Model() LifeTimeEventModel {
    return LifeTimeEventModel{
        BaseEventModel: self.BaseModel(),
        StartJd: self.startJd,
        EndJd: self.endJd,
    }
}
func (self LifeTimeEventModel) GetEvent() (LifeTimeEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return LifeTimeEvent{}, err
    }
    return LifeTimeEvent{
        BaseEvent: baseEvent,
        startJd: self.StartJd,
        endJd: self.EndJd,
    }, nil
}


