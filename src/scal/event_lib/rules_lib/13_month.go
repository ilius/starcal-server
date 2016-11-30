package rules_lib

import "scal/utils"

const R_month = "month"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			return false
		}
		return utils.MonthListIsValid(list)
	}
	RegisterRuleType(
		13,
		R_month,
		T_int_range_list,
		&checker,
	)
}
