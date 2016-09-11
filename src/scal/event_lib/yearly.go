package event_lib


type YearlyEventModel struct {
    BaseEventModel  `bson:",inline"`
    Month int       `bson:"month"`
    Day int         `bson:"day"`
    StartYear int   `bson:"startYear"`
    StartYearEnable bool    `bson:"startYearEnable"`
}
func (self YearlyEventModel) Type() string {
    return "yearly"
}
func (self YearlyEventModel) Collection() string {
    return "events_yearly"
}


type YearlyEvent struct {
    BaseEvent
    month int
    day int
    startYear int
    startYearEnable bool
}
func (self YearlyEvent) Type() string {
    return "yearly"
}
func (self YearlyEvent) Month() int {
    return self.month
}
func (self YearlyEvent) Day() int {
    return self.day
}
func (self YearlyEvent) StartYear() int {
    return self.startYear
}
func (self YearlyEvent) StartYearEnable() bool {
    return self.startYearEnable
}


func (self YearlyEvent) Model() YearlyEventModel {
    return YearlyEventModel{
        BaseEventModel: self.BaseModel(),
        Month: self.month,
        Day: self.day,
        StartYear: self.startYear,
        StartYearEnable: self.startYearEnable,
    }
}
func (self YearlyEventModel) GetEvent() (YearlyEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return YearlyEvent{}, err
    }
    return YearlyEvent{
        BaseEvent: baseEvent,
        month: self.Month,
        day: self.Day,
        startYear: self.StartYear,
        startYearEnable: self.StartYearEnable,
    }, nil
}

