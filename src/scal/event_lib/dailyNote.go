package event_lib


type DailyNoteEventModel struct {
    BaseEventModel  `bson:",inline"`
    Jd int          `bson:"jd"`
}

type DailyNoteEvent struct {
    BaseEvent
    jd int
}
func (self DailyNoteEvent) Jd() int {
    return self.jd
}



func (self DailyNoteEvent) Model() DailyNoteEventModel {
    return DailyNoteEventModel{
        BaseEventModel: self.BaseModel("dailyNote"),
        Jd: self.jd,
    }
}
func (self DailyNoteEventModel) GetEvent() (DailyNoteEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent("dailyNote")
    if err != nil {
        return DailyNoteEvent{}, err
    }
    return DailyNoteEvent{
        BaseEvent: baseEvent,
        jd: self.Jd,
    }, nil
}
