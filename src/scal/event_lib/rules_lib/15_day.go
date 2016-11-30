package rules_lib

import "scal/utils"

const R_day = "day"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			return false
		}
		return utils.DayListIsValid(list)
	}
	RegisterRuleType(
		15,
		R_day,
		T_int_range_list,
		&checker,
	)
}
