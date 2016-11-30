package rules_lib

import "scal/utils"

const R_ex_day = "ex_day"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			return false
		}
		return utils.DayListIsValid(list)
	}
	RegisterRuleType(
		16,
		R_ex_day,
		T_int_range_list,
		&checker,
	)
}
