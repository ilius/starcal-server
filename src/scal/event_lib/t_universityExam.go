package event_lib

import (
	"scal/storage"

	lib "github.com/ilius/libgostarcal"
	"github.com/ilius/libgostarcal/utils"
)

type UniversityExamEventModel struct {
	BaseEventModel  `bson:",inline" json:",inline"`
	Jd              int    `bson:"jd" json:"jd"`
	DayStartSeconds uint32 `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   uint32 `bson:"dayEndSeconds" json:"dayEndSeconds"`
	CourseId        int    `bson:"courseId" json:"courseId"`
}

func (UniversityExamEventModel) Type() string {
	return "universityExam"
}

func LoadUniversityExamEventModel(db storage.Database, sha1 string) (
	*UniversityExamEventModel,
	error,
) {
	model := UniversityExamEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type UniversityExamEvent struct {
	BaseEvent
	jd              int
	dayStartSeconds uint32
	dayEndSeconds   uint32
	courseId        int
}

func (UniversityExamEvent) Type() string {
	return "universityExam"
}

func (event UniversityExamEvent) DayStartSeconds() uint32 {
	return event.dayStartSeconds
}

func (event UniversityExamEvent) DayEndSeconds() uint32 {
	return event.dayEndSeconds
}

func (event UniversityExamEvent) DayStartHMS() lib.HMS {
	return utils.GetHmsBySeconds(uint(event.dayStartSeconds))
}

func (event UniversityExamEvent) DayEndHMS() lib.HMS {
	return utils.GetHmsBySeconds(uint(event.dayEndSeconds))
}

func (event UniversityExamEvent) Model() UniversityExamEventModel {
	return UniversityExamEventModel{
		BaseEventModel:  event.BaseModel(),
		Jd:              event.jd,
		DayStartSeconds: event.dayStartSeconds,
		DayEndSeconds:   event.dayEndSeconds,
		CourseId:        event.courseId,
	}
}

func (model UniversityExamEventModel) GetEvent() (UniversityExamEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return UniversityExamEvent{}, err
	}
	return UniversityExamEvent{
		BaseEvent:       baseEvent,
		jd:              model.Jd,
		dayStartSeconds: model.DayStartSeconds,
		dayEndSeconds:   model.DayEndSeconds,
		courseId:        model.CourseId,
	}, nil
}
