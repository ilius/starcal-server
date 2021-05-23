package event_lib

/*
startYear := start * scale
endYear := end * scale
*/

type LargeScaleEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	Scale          int64 `bson:"scale" json:"scale"`
	Start          int64 `bson:"start" json:"start"`
	End            int64 `bson:"end" json:"end"`
	DurationEnable bool  `bson:"durationEnable" json:"durationEnable"`
}

func (LargeScaleEventModel) Type() string {
	return "largeScale"
}

type LargeScaleEvent struct {
	BaseEvent
	scale          int64
	start          int64
	end            int64
	durationEnable bool
}

func (LargeScaleEvent) Type() string {
	return "largeScale"
}

func (event LargeScaleEvent) Scale() int64 {
	return event.scale
}

func (event LargeScaleEvent) Start() int64 {
	return event.start
}

func (event LargeScaleEvent) End() int64 {
	return event.end
}

func (event LargeScaleEvent) DurationEnable() bool {
	return event.durationEnable
}

func (event LargeScaleEvent) StartYear() int64 {
	return event.start * event.scale
}

func (event LargeScaleEvent) EndYear() int64 {
	return event.end * event.scale
}

/*
func (event LargeScaleEvent) StartJd() int64 {
    return int64(event.calType.ToJd(lib.Date{
        int(event.start * event.scale),
        1,
        1,
    }))
}
func (event LargeScaleEvent) EndJd() int64 {
    return int64(event.calType.ToJd(lib.Date{
        int(event.end * event.scale),
        1,
        1,
    }))
}*/

func (event LargeScaleEvent) Model() LargeScaleEventModel {
	return LargeScaleEventModel{
		BaseEventModel: event.BaseModel(),
		Scale:          event.scale,
		Start:          event.start,
		End:            event.end,
		DurationEnable: event.durationEnable,
	}
}

func (model LargeScaleEventModel) GetEvent() (LargeScaleEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return LargeScaleEvent{}, err
	}
	return LargeScaleEvent{
		BaseEvent:      baseEvent,
		scale:          model.Scale,
		start:          model.Start,
		end:            model.End,
		durationEnable: model.DurationEnable,
	}, nil
}
