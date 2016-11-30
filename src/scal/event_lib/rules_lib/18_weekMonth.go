/*
	the value for this rule is a json string
	containing a map/dict with 3 keys: "month", "weekIndex", "weekDay"
	where all 3 values are non-negative integers, with different upper bounds
		key="weekIndex"		0 <= value <= 4
		key="weekDay"		0 <= value <= 6
		key="month"			0 <= value <= 12
	Meaning of these integer values:
		weekIndex=0 means the first week of month
		weekIndex=4 means the last week of month
		weekDay=0 means Sunday
		weekDay=6 means Saturday
		month=0  means all months
		month=1  means the first month (January in Gregorian calendar)
		month=12 means the last month (December in Gregorian calendar)
	For example, this is a valid value for this rule:
		{"weekIndex": 4, "weekDay": 6, "month": 12}
	And in Gregorian calendar, it means: Last Saturday of December

	Another example:
		{"weekIndex": 1, "weekDay": 1, "month": 0}
	Means:
		Second Monday of all months

	The type of calendar (Gregorian, Jalali, etc) is specified is the event
*/

package rules_lib

import "encoding/json"

const R_weekMonth = "weekMonth"
const T_weekMonth = "WeekMonth"

type WeekMonth struct {
	WeekIndex int `json:"weekIndex"` // 0..4
	WeekDay   int `json:"weekDay"`   // 0..6
	Month     int `json:"month"`     // 0..12   0 means every month
}

func (wm WeekMonth) IsValid() bool {
	return (wm.Month >= 0 && wm.Month <= 12 &&
		wm.WeekIndex >= 0 && wm.WeekIndex <= 4 &&
		wm.WeekDay >= 0 && wm.WeekDay <= 6)
}

func init() {
	RegisterValueDecoder(T_weekMonth, func(value string) (interface{}, error) {
		valueBytes := []byte(value)
		obj := WeekMonth{}
		err := json.Unmarshal(valueBytes, &obj)
		return obj, err
	})
	checker := func(value interface{}) bool {
		wm, ok := value.(WeekMonth)
		return ok && wm.IsValid()
	}
	RegisterRuleType(
		18,
		R_weekMonth,
		T_weekMonth,
		&checker,
	)
}
