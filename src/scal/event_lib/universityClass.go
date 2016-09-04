package event_lib

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
    BaseEventModel      `bson:",inline"`
    WeekNumMode string  `bson:"weekNumMode"`
    WeekDayList []int   `bson:"weekDayList"`
}


type UniversityClassEvent struct {
    BaseEvent
    weekNumMode string
    weekDayList []int
}
func (self UniversityClassEvent) WeekNumMode() string {
    return self.weekNumMode
}
func (self UniversityClassEvent) WeekDayList() []int {
    return self.weekDayList
}



func (self UniversityClassEvent) Model() UniversityClassEventModel {
    return UniversityClassEventModel{
        BaseEventModel: self.BaseModel("universityClass"),
        WeekNumMode: self.weekNumMode,
        WeekDayList: self.weekDayList,
    }
}
func (self UniversityClassEventModel) GetEvent() (UniversityClassEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent("universityClass")
    if err != nil {
        return UniversityClassEvent{}, err
    }
    return UniversityClassEvent{
        BaseEvent: baseEvent,
        weekNumMode: self.WeekNumMode,
        weekDayList: self.WeekDayList,
    }, nil
}




