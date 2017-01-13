package rules_lib

import "log"
import "scal/utils"

const R_ex_day = "ex_day"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_ex_day,
				value,
				value,
			)
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
