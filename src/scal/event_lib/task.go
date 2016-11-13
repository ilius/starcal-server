package event_lib

import "time"
import "fmt"
import "errors"


// DurationUnit is just a matter of UI
// DurationUnit=0       ==> shows End datetime in UI
// DurationUnit=1       ==> seconds
// DurationUnit=60      ==> minutes
// DurationUnit=3600    ==> hours
// DurationUnit=86400   ==> days
// DurationUnit=604800  ==> weeks

type TaskEventModel struct {
    BaseEventModel          `bson:",inline" json:",inline"`
    StartTime *time.Time    `bson:"startTime" json:"startTime"`
    EndTime *time.Time      `bson:"endTime" json:"endTime"`
    DurationUnit int        `bson:"durationUnit" json:"durationUnit"`
}
func (self TaskEventModel) Type() string {
    return "task"
}


type TaskEvent struct {
    BaseEvent
    startTime time.Time
    endTime time.Time
    durationUnit int
}
func (self TaskEvent) Type() string {
    return "task"
}
func (self TaskEvent) StartTime() time.Time {
    if self.locEnable && self.loc != nil {
        return self.startTime.In(self.loc)
    }
    return self.startTime
}
func (self TaskEvent) EndTime() time.Time {
    if self.locEnable && self.loc != nil {
        return self.endTime.In(self.loc)
    }
    return self.endTime
}
func (self TaskEvent) DurationUnit() int {
    return self.durationUnit
}
func (self TaskEvent) String() string {
    const time_format = "2006-01-02 15:04:05"
    return fmt.Sprintf(
        "Task: %v - %v",
        self.StartTime().Format(time_format),
        self.EndTime().Format(time_format),
    )
}





func (self TaskEvent) Model() TaskEventModel {
    startTime := self.startTime
    endTime := self.endTime
    return TaskEventModel{
        BaseEventModel: self.BaseModel(),
        StartTime: &startTime,
        EndTime: &endTime,
        DurationUnit: self.durationUnit,
    }
}
func (self TaskEventModel) GetEvent() (TaskEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return TaskEvent{}, err
    }
    if self.StartTime == nil {
        return TaskEvent{}, errors.New("missing 'startTime'")
    }
    if self.EndTime == nil {
        return TaskEvent{}, errors.New("missing 'endTime'")
    }
    return TaskEvent{
        BaseEvent: baseEvent,
        startTime: *self.StartTime,
        endTime: *self.EndTime,
        durationUnit: self.DurationUnit,
    }, nil
}



