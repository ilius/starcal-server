package event_lib

import "scal/storage"

// WeekNumMode: "any" | "odd" | "even"
/*
   WeekDayList: a slice of integers,
   each int represents a WeekDay:
       0: Sunday
       1: Monday
       2: Tuesday
       3: Wednesday
       4: Thursday
       5: Friday
       6: Saturday
*/

type UniversityClassEventModel struct {
	BaseEventModel  `bson:",inline" json:",inline"`
	WeekNumMode     string `bson:"weekNumMode" json:"weekNumMode"`
	WeekDayList     []int  `bson:"weekDayList" json:"weekDayList"`
	DayStartSeconds uint32 `bson:"dayStartSeconds" json:"dayStartSeconds"`
	DayEndSeconds   uint32 `bson:"dayEndSeconds" json:"dayEndSeconds"`
	CourseId        int    `bson:"courseId" json:"courseId"`
}

func (UniversityClassEventModel) Type() string {
	return "universityClass"
}

func LoadUniversityClassEventModel(db storage.Database, sha1 string) (
	*UniversityClassEventModel,
	error,
) {
	model := UniversityClassEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type UniversityClassEvent struct {
	BaseEvent
	weekNumMode string
	weekDayList []int
}

func (UniversityClassEvent) Type() string {
	return "universityClass"
}
func (event UniversityClassEvent) WeekNumMode() string {
	return event.weekNumMode
}
func (event UniversityClassEvent) WeekDayList() []int {
	return event.weekDayList
}

func (event UniversityClassEvent) Model() UniversityClassEventModel {
	return UniversityClassEventModel{
		BaseEventModel: event.BaseModel(),
		WeekNumMode:    event.weekNumMode,
		WeekDayList:    event.weekDayList,
	}
}

func (model UniversityClassEventModel) GetEvent() (UniversityClassEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return UniversityClassEvent{}, err
	}
	return UniversityClassEvent{
		BaseEvent:   baseEvent,
		weekNumMode: model.WeekNumMode,
		weekDayList: model.WeekDayList,
	}, nil
}
