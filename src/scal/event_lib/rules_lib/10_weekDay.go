package rules_lib

import (
	"log"

	"github.com/ilius/libgostarcal/utils"
)

const R_weekDay = "weekDay"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_weekDay,
				value,
				value,
			)
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
