package event_lib

/*
startYear := start * scale
endYear := end * scale
*/

type LargeScaleEventModel struct {
    BaseEventModel          `bson:",inline"`
    Scale int64             `bson:"scale"`
    Start int64             `bson:"start"`
    End int64               `bson:"end"`
    DurationEnable bool     `bson:"durationEnable"`
}

type LargeScaleEvent struct {
    BaseEvent
    scale int64
    start int64
    end int64
    durationEnable bool
}
func (self LargeScaleEvent) Type() string {
    return "largeScale"
}
func (self LargeScaleEvent) Scale() int64 {
    return self.scale
}
func (self LargeScaleEvent) Start() int64 {
    return self.start
}
func (self LargeScaleEvent) End() int64 {
    return self.end
}
func (self LargeScaleEvent) DurationEnable() bool {
    return self.durationEnable
}
func (self LargeScaleEvent) StartYear() int64 {
    return self.start * self.scale
}
func (self LargeScaleEvent) EndYear() int64 {
    return self.end * self.scale
}
/*
func (self LargeScaleEvent) StartJd() int64 {
    return int64(self.calType.ToJd(scal.Date{
        int(self.start * self.scale),
        1,
        1,
    }))
}
func (self LargeScaleEvent) EndJd() int64 {
    return int64(self.calType.ToJd(scal.Date{
        int(self.end * self.scale),
        1,
        1,
    }))
}*/




func (self LargeScaleEvent) Model() LargeScaleEventModel {
    return LargeScaleEventModel{
        BaseEventModel: self.BaseModel(),
        Scale: self.scale,
        Start: self.start,
        End: self.end,
        DurationEnable: self.durationEnable,
    }
}
func (self LargeScaleEventModel) GetEvent() (LargeScaleEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return LargeScaleEvent{}, err
    }
    return LargeScaleEvent{
        BaseEvent: baseEvent,
        scale: self.Scale,
        start: self.Start,
        end: self.End,
        durationEnable: self.DurationEnable,
    }, nil
}





