package event_lib

import "gopkg.in/mgo.v2"
import "scal/storage"

/*
bson/bson.go:
    type M map[string]interface{}


Current rule value types:
(Must convert all values into string)


cycleDays       int
cycleWeeks      int

duration        string
date            string              %Y-%m-%d
dayTime         string              %H:%M:%S
weekNumMode:    string              "any| "odd" | "even"

year:           []int
month:          []int
day:            []int
ex_year         []int
ex_month        []int
weekDay         []int

dayTimeRange    [2]string           ["%H:%M:%S", "%H:%M:%S"]
ex_dates        []string            ["%Y-%m-%d" ...]

weekMonth       map[string][int]    keys: month, wmIndex, weekDay
start           map[string]string   keys: date, time
end             map[string]string   keys: date, time
cycleLen        map                 keys: days(int), extraTime(HMS)

*/


type EventRuleModel struct {
    Name string
    Value string
}

type EventRuleModelList []EventRuleModel

type CustomEventModel struct {
    BaseEventModel              `bson:",inline" json:",inline"`
    Rules EventRuleModelList    `bson:"rules" json:"rules"`
}
func (self CustomEventModel) Type() string {
    return "custom"
}


func LoadCustomEventModel(db *mgo.Database, sha1 string) (
    *CustomEventModel,
    error,
) {
    model := CustomEventModel{}
    model.Sha1 = sha1
    err := storage.Get(db, &model)
    return &model, err
}

// Modular mode:
type EventRule interface {
    Name() string
    Value() interface{}
    ValueString() string
    Model() EventRuleModel
}

type EventRuleList []EventRule
//func (rules EventRuleList) Model() EventRuleModelList {
//}



/*


// Non-Modular mode:
type EventRule struct {
    name string
    value string
}
func (self EventRule) Name() string {
    return self.name
}
func (self EventRule) Value() string {
    return self.value
}
//func (self EventRule) ParseValue() interface{} {
//    return self.value
//}


type CustomEvent struct {
    BaseEvent
    rules EventRuleList
}
func (self CustomEvent) Type() string {
    return "custom"
}
func (self CustomEvent) Rules() EventRuleList {
    return self.rules
}

func (self EventRule) Model() EventRuleModel {
    return EventRuleModel{
        Name: self.name,
        Value: self.value,
    }
}

func (self EventRuleModel) GetRule() EventRule {
    return EventRule{
        name: self.Name,
        value: self.Value,
    }
}

func (self CustomEvent) Model() CustomEventModel {
    return CustomEventModel{
        BaseEventModel: self.BaseModel("custom"),
        //Rules: self.rules,
    }
}
func (self CustomEventModel) GetEvent() (CustomEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent("custom")
    if err != nil {
        return CustomEvent{}, err
    }
    return CustomEvent{
        BaseEvent: baseEvent,
        //rules: self.Rules,
    }, nil
}
*/
