package rules_lib

import (
	"log"

	"github.com/ilius/libgostarcal/utils"
)

const R_ex_month = "ex_month"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_ex_month,
				value,
				value,
			)
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
