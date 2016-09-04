package event_lib

// DurationEnable is just a matter of UI

type AllDayTaskEventModel struct {
    BaseEventModel  `bson:",inline"`
    StartJd int     `bson:"startJd"`
    EndJd int       `bson:"endJd"`
    DurationEnable bool     `bson:"durationEnable"`
}


type AllDayTaskEvent struct {
    BaseEvent
    startJd int
    endJd int
    durationEnable bool
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
        BaseEventModel: self.BaseModel("allDayTask"),
        StartJd: self.startJd,
        EndJd: self.endJd,
        DurationEnable: self.durationEnable,
    }
}
func (self AllDayTaskEventModel) GetEvent() (AllDayTaskEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent("allDayTask")
    if err != nil {
        return AllDayTaskEvent{}, err
    }
    return AllDayTaskEvent{
        BaseEvent: baseEvent,
        startJd: self.StartJd,
        endJd: self.EndJd,
        durationEnable: self.DurationEnable,
    }, nil
}




