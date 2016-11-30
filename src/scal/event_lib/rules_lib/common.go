package rules_lib

import (
	"errors"
)

const (
	T_string         = "string"
	T_int            = "int"
	T_int_list       = "int_list"
	T_int_range_list = "int_range_list" // "1380-1383 1393 1396"
	T_float          = "float"
	T_HMS            = "HMS"       // format: "365 23:55:55"
	T_DHMS           = "DHMS"      // Days and HMS, format: "365 23:55:55"
	T_HMSRange       = "HMSRange"  // format: "14:30:00 15:30:00"
	T_Date           = "Date"      // format: "YYYY/MM/DD"
	T_Date_list      = "Date_list" // format: "YYYY/MM/DD YYYY/MM/DD YYYY/MM/DD"
	T_DateHMS        = "DateHMS"   // format: "YYYY/MM/DD hh:mm:ss"
	T_Duration       = "Duration"  // format: "2 s", "2.5 m", "2.5 h", "2.5 d", "2.0 w"
)

type EventRuleType struct {
	Order        int
	Name         string
	ValueDecoder func(value string) (interface{}, error)
	ValueChecker *func(value interface{}) bool
}

type EventRuleTypeList []*EventRuleType

func (list EventRuleTypeList) Names() []string {
	names := make([]string, len(list))
	for i, t := range list {
		names[i] = t.Name
	}
	return names
}

type EventRuleTypeMap map[string]*EventRuleType

var ruleTypes = EventRuleTypeMap{}

type EventRuleModel struct {
	Type  string `bson:"type" json:"type"`
	Value string `bson:"value" json:"value"`
}

func (model EventRuleModel) Decode() (EventRule, error) {
	typeObj, ok := ruleTypes[model.Type]
	if !ok {
		return EventRule{}, errors.New(
			"invalid rule type '" + model.Type + "'",
		)
	}
	newValue, err := typeObj.ValueDecoder(model.Value)
	if err != nil {
		return EventRule{}, err
	}
	return EventRule{
		Type:  typeObj,
		Value: newValue,
	}, nil
}

type EventRule struct {
	Type  *EventRuleType
	Value interface{}
}

func (rule EventRule) Check() bool {
	checker := rule.Type.ValueChecker
	if checker == nil {
		return true
	}
	return (*checker)(rule.Value)
}

/*
func (rule EventRule) Model() EventRuleModel { // Model() or Encode() FIXME
    return EventRuleModel{
        Type: rule.Type.Name,
        Value: string(rule.Value),// or rule.Type.Encode(rule.Value) FIXME
    }
}*/
