package event_lib

import (
	"errors"
	"fmt"
	"scal/storage"
	"time"
)

// DurationUnit is just a matter of UI
// DurationUnit=0       ==> shows End datetime in UI
// DurationUnit=1       ==> seconds
// DurationUnit=60      ==> minutes
// DurationUnit=3600    ==> hours
// DurationUnit=86400   ==> days
// DurationUnit=604800  ==> weeks

type TaskEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	StartTime      *time.Time `bson:"startTime" json:"startTime"`
	EndTime        *time.Time `bson:"endTime" json:"endTime"`
	DurationUnit   uint       `bson:"durationUnit" json:"durationUnit"`
}

func (TaskEventModel) Type() string {
	return "task"
}

func LoadTaskEventModel(db storage.Database, sha1 string) (
	*TaskEventModel,
	error,
) {
	model := TaskEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type TaskEvent struct {
	BaseEvent
	startTime    time.Time
	endTime      time.Time
	durationUnit uint
}

func (TaskEvent) Type() string {
	return "task"
}
func (event TaskEvent) StartTime() time.Time {
	if event.locEnable && event.loc != nil {
		return event.startTime.In(event.loc)
	}
	return event.startTime
}
func (event TaskEvent) EndTime() time.Time {
	if event.locEnable && event.loc != nil {
		return event.endTime.In(event.loc)
	}
	return event.endTime
}
func (event TaskEvent) DurationUnit() uint {
	return event.durationUnit
}
func (event TaskEvent) String() string {
	const time_format = "2006-01-02 15:04:05"
	return fmt.Sprintf(
		"Task: %v - %v",
		event.StartTime().Format(time_format),
		event.EndTime().Format(time_format),
	)
}

func (event TaskEvent) Model() TaskEventModel {
	startTime := event.startTime
	endTime := event.endTime
	return TaskEventModel{
		BaseEventModel: event.BaseModel(),
		StartTime:      &startTime,
		EndTime:        &endTime,
		DurationUnit:   event.durationUnit,
	}
}

func (model TaskEventModel) GetEvent() (TaskEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return TaskEvent{}, err
	}
	if model.StartTime == nil {
		return TaskEvent{}, errors.New("missing 'startTime'")
	}
	if model.EndTime == nil {
		return TaskEvent{}, errors.New("missing 'endTime'")
	}
	return TaskEvent{
		BaseEvent:    baseEvent,
		startTime:    *model.StartTime,
		endTime:      *model.EndTime,
		durationUnit: model.DurationUnit,
	}, nil
}
