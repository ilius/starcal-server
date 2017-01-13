package rules_lib

import "log"
import "scal/utils"

const R_month = "month"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_month,
				value,
				value,
			)
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
