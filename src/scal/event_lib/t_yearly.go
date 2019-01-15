package event_lib

import "scal/storage"

type YearlyEventModel struct {
	BaseEventModel  `bson:",inline" json:",inline"`
	Month           uint8 `bson:"month" json:"month"`
	Day             uint8 `bson:"day" json:"day"`
	StartYear       int   `bson:"startYear" json:"startYear"`
	StartYearEnable bool  `bson:"startYearEnable" json:"startYearEnable"`
}

func (YearlyEventModel) Type() string {
	return "yearly"
}

func LoadYearlyEventModel(db storage.Database, sha1 string) (
	*YearlyEventModel,
	error,
) {
	model := YearlyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

type YearlyEvent struct {
	BaseEvent
	month           uint8
	day             uint8
	startYear       int
	startYearEnable bool
}

func (YearlyEvent) Type() string {
	return "yearly"
}
func (event YearlyEvent) Month() uint8 {
	return event.month
}
func (event YearlyEvent) Day() uint8 {
	return event.day
}
func (event YearlyEvent) StartYear() int {
	return event.startYear
}
func (event YearlyEvent) StartYearEnable() bool {
	return event.startYearEnable
}

func (event YearlyEvent) Model() YearlyEventModel {
	return YearlyEventModel{
		BaseEventModel:  event.BaseModel(),
		Month:           event.month,
		Day:             event.day,
		StartYear:       event.startYear,
		StartYearEnable: event.startYearEnable,
	}
}

func (model YearlyEventModel) GetEvent() (YearlyEvent, error) {
	baseEvent, err := model.BaseEventModel.GetBaseEvent()
	if err != nil {
		return YearlyEvent{}, err
	}
	return YearlyEvent{
		BaseEvent:       baseEvent,
		month:           model.Month,
		day:             model.Day,
		startYear:       model.StartYear,
		startYearEnable: model.StartYearEnable,
	}, nil
}
