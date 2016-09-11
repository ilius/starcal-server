package event_lib

import "scal"
import . "scal/utils"

type UniversityExamEventModel struct {
    BaseEventModel          `bson:",inline"`
    Jd int                  `bson:"jd"`
    DayStartSeconds int     `bson:"dayStartSeconds"`
    DayEndSeconds int       `bson:"dayEndSeconds"`
    CourseId int            `bson:"courseId"`
}
func (self UniversityExamEventModel) Type() string {
    return "universityExam"
}
func (self UniversityExamEventModel) Collection() string {
    return "events_universityExam"
}


type UniversityExamEvent struct {
    BaseEvent
    jd int
    dayStartSeconds int
    dayEndSeconds int
    courseId int
}
func (self UniversityExamEvent) Type() string {
    return "universityExam"
}
func (self UniversityExamEvent) DayStartSeconds() int {
    return self.dayStartSeconds
}
func (self UniversityExamEvent) DayEndSeconds() int {
    return self.dayEndSeconds
}
func (self UniversityExamEvent) DayStartHMS() scal.HMS {
    return GetHmsBySeconds(self.dayStartSeconds)
}
func (self UniversityExamEvent) DayEndHMS() scal.HMS {
    return GetHmsBySeconds(self.dayEndSeconds)
}



func (self UniversityExamEvent) Model() UniversityExamEventModel {
    return UniversityExamEventModel{
        BaseEventModel: self.BaseModel(),
        Jd: self.jd,
        DayStartSeconds: self.dayStartSeconds,
        DayEndSeconds: self.dayEndSeconds,
        CourseId: self.courseId,
    }
}
func (self UniversityExamEventModel) GetEvent() (UniversityExamEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return UniversityExamEvent{}, err
    }
    return UniversityExamEvent{
        BaseEvent: baseEvent,
        jd: self.Jd,
        dayStartSeconds: self.DayStartSeconds,
        dayEndSeconds: self.DayEndSeconds,
        courseId: self.CourseId,
    }, nil
}


