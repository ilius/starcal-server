package rules_lib

import "scal"

const R_dayTimeRange = "dayTimeRange"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.HMSRange)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		6,
		R_dayTimeRange,
		T_HMSRange,
		&checker,
	)
}
