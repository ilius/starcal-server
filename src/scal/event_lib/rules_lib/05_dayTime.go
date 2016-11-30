package rules_lib

import "scal"

const R_dayTime = "dayTime"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.HMS)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		5,
		R_dayTime,
		T_HMS,
		&checker,
	)
}
