package rules_lib

import "scal/utils"

const R_weekDay = "weekDay"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			return false
		}
		return utils.WeekDayListIsValid(list)
	}
	RegisterRuleType(
		10,
		R_weekDay,
		T_int_list,
		&checker,
	)
}
