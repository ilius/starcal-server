package rules_lib

import "log"

const R_cycleWeeks = "cycleWeeks"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_cycleWeeks,
				value,
				value,
			)
		}
		return ok && v > 0
	}
	RegisterRuleType(
		9,
		R_cycleWeeks,
		T_int,
		&checker,
	)
}
