package rules_lib

import "scal"

const R_end = "end"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.DateHMS)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		1,
		R_end,
		T_DateHMS,
		&checker,
	)
}
