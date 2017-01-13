package rules_lib

import "log"

const R_cycleDays = "cycleDays"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(int)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_cycleDays,
				value,
				value,
			)
		}
		return ok && v > 0
	}
	RegisterRuleType(
		8,
		R_cycleDays,
		T_int,
		&checker,
	)
}
