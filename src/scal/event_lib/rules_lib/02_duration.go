package rules_lib

import "scal/utils"

const R_duration = "duration"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(utils.Duration)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		2,
		R_duration,
		T_Duration,
		&checker,
	)
}
