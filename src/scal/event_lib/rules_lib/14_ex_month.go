package rules_lib

import "scal/utils"

const R_ex_month = "ex_month"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			return false
		}
		return utils.MonthListIsValid(list)
	}
	RegisterRuleType(
		14,
		R_ex_month,
		T_int_range_list,
		&checker,
	)
}
