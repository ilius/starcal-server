package rules_lib

import "scal"

const R_start = "start"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.DateHMS)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		0,
		R_start,
		T_DateHMS,
		&checker,
	)
}
