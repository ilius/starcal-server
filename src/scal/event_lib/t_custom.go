package event_lib

import (
	"errors"
	"fmt"
	"reflect"

	rlib "github.com/ilius/libgostarcal/event/rules_lib"
)

/*
bson/bson.go:
    type M map[string]interface{}


Current rule value types:

start           "%Y/%m/%d %H:%M:%S"
end             "%Y/%m/%d %H:%M:%S"
duration        examples: "2 s", "2.5 m", "2.5 h", "2.5 d", "2.0 w"
date            "%Y/%m/%d"
ex_dates        "%Y/%m/%d %Y/%m/%d %Y/%m/%d"
dayTime         "%H:%M:%S"
dayTimeRange    "%H:%M:%S %H:%M:%S"
cycleLen        "%{days} %H:%M:%S"
cycleDays       int
cycleWeeks      int
weekDay         space-separated integers (0 to 6)
year            space-separated ranges of integers
                "1380-1383 1393 1396" = "1380 1381 1382 1383 1390 1393 1396"
ex_year         space-separated ranges of integers, like `year`
month           space-separated ranges of integers (1 to 12)
                example: "1-6 10 12"
ex_month        space-separated ranges of integers (1 to 12), like `month`
day             space-separated ranges of integers (1 to 39)
                example: "1-10 20 30-33"
ex_day          space-separated ranges of integers (1 to 39), like `day`
weekNumMode     "any | "odd" | "even"
weekMonth       json: `{"weekIndex": 4, "weekDay": 6, "month": 12}`

*/

type EventRuleModelList []rlib.EventRuleModel

type CustomEventModel struct {
	BaseEventModel `bson:",inline" json:",inline"`
	Rules          EventRuleModelList `bson:"rules" json:"rules"`
}

func (CustomEventModel) Type() string {
	return "custom"
}

type EventRuleMap map[string]rlib.EventRule

type CustomEvent struct {
	BaseEvent
	ruleMap   EventRuleMap
	ruleTypes rlib.EventRuleTypeList
}

func (CustomEvent) Type() string {
	return "custom"
}

// func (event CustomEvent) RuleMap() EventRuleMap {
//	return event.ruleMap
//}
func (event CustomEvent) GetRule(typeName string) (rlib.EventRule, bool) {
	typeObj, ok := event.ruleMap[typeName]
	return typeObj, ok
}

func (event CustomEvent) RuleTypes() rlib.EventRuleTypeList {
	return event.ruleTypes
}

func (event CustomEvent) IterRules() <-chan rlib.EventRule {
	ch := make(chan rlib.EventRule)
	go func() {
		defer close(ch)
		for _, ruleType := range event.ruleTypes {
			rule, ok := event.ruleMap[ruleType.Name]
			if !ok {
				log.Error(fmt.Sprintf(
					"IterRules: rule type %s not found, eventId=%s\n",
					ruleType.Name,
					event.Id(),
				))
				continue
			}
			ch <- rule
		}
	}()
	return ch
}

func (event CustomEvent) CheckRuleTypes() error {
	for _, ruleType := range event.ruleTypes {
		//_, ok := event.ruleMap[ruleType.Name]
		// if !ok {
		//	return errors.New(
		//		"rule type " + ruleType.Name + " not found",
		//	)
		//}
		requiredTypes, hasRequired := rlib.RulesRequire[ruleType.Name]
		if hasRequired {
			for _, requiredType := range requiredTypes {
				_, ok := event.ruleMap[requiredType]
				if !ok {
					return errors.New(
						"rule type '" + requiredType +
							"' is required by '" + ruleType.Name + "'")
				}
			}
		}
		conflictTypes, hasConflicts := rlib.RulesConflictWith[ruleType.Name]
		if hasConflicts {
			for _, conflictType := range conflictTypes {
				_, nok := event.ruleMap[conflictType]
				if nok {
					return errors.New(
						"rule type '" + ruleType.Name +
							"' conflicts with '" + conflictType + "'")
				}
			}
		}
	}
	return nil
}

func (event *CustomEvent) GetModifiedRuleTypes(oldEvent *CustomEvent) rlib.EventRuleTypeList {
	modTypes := make(
		rlib.EventRuleTypeList,
		0,
		len(event.ruleTypes)+len(oldEvent.ruleTypes),
	)
	for _, ruleType := range event.ruleTypes {
		newRule, ok := event.ruleMap[ruleType.Name]
		if !ok {
			log.Error(fmt.Sprintf(
				"GetModifiedRuleTypes: rule type %s not found, eventId=%s\n",
				ruleType.Name,
				event.Id(),
			))
			continue
		}
		oldRule, hasOld := oldEvent.ruleMap[ruleType.Name]
		if !(hasOld && reflect.DeepEqual(oldRule.Value, newRule.Value)) {
			modTypes = append(modTypes, ruleType)
		}
	}
	for _, oldRuleType := range oldEvent.ruleTypes {
		_, hasNew := event.ruleMap[oldRuleType.Name]
		if !hasNew {
			// rule has been deleted
			modTypes = append(modTypes, oldRuleType)
		}
	}
	return modTypes
}

/*
	func (event CustomEvent) Model() CustomEventModel {
		ruleModels := make(EventRuleModelList, 0, len(event.ruleTypes))
		for rule := range event.IterRules() {
			ruleModels = append(ruleModels, rule.Model())
		}
		return CustomEventModel{
			BaseEventModel: event.BaseModel("custom"),
			Rules: ruleModels,
		}
	}
*/

func (eventModel CustomEventModel) GetEvent() (CustomEvent, error) {
	baseEvent, err := eventModel.BaseEventModel.GetBaseEvent()
	if err != nil {
		return CustomEvent{}, err
	}
	ruleMap := EventRuleMap{}
	ruleTypes := make(rlib.EventRuleTypeList, len(eventModel.Rules))
	for index, ruleModel := range eventModel.Rules {
		rule, err := ruleModel.Decode()
		if err != nil {
			return CustomEvent{}, err
		}
		if !rule.Check() {
			return CustomEvent{}, errors.New(
				"bad value for event rule '" + ruleModel.Type + "'",
			)
		}
		ruleMap[rule.Type.Name] = rule
		ruleTypes[index] = rule.Type
	}
	event := CustomEvent{
		BaseEvent: baseEvent,
		ruleMap:   ruleMap,
		ruleTypes: ruleTypes,
	}
	// whether or not check rule types (dependencies)
	// pass a bool argument, or use a settings bool flag? FIXME
	err = event.CheckRuleTypes()
	return event, err
}

func DecodeMapEventRuleModelList(rawMapList interface{}) (EventRuleModelList, error) {
	rawList, ok := rawMapList.([]interface{})
	if !ok {
		return EventRuleModelList{}, errors.New(
			"could not convert to rawList",
		)
	}
	modelList := make(EventRuleModelList, len(rawList))
	for i, raw := range rawList {
		m, ok := raw.(map[string]interface{})
		if !ok {
			return EventRuleModelList{}, fmt.Errorf(
				"could not convert %v with type %T to M",
				raw,
				raw,
			)
		}
		typeName, ok := m["type"].(string)
		if !ok {
			return EventRuleModelList{}, errors.New(
				"missing or bad parameter 'type'",
			)
		}
		value, ok := m["value"].(string)
		if !ok {
			return EventRuleModelList{}, errors.New(
				"missing or bad parameter 'value'",
			)
		}
		modelList[i] = rlib.EventRuleModel{
			Type:  typeName,
			Value: value,
		}
	}
	return modelList, nil
}
