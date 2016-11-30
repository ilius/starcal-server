package rules_lib

import "scal"

const R_cycleLen = "cycleLen"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.DHMS)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		7,
		R_cycleLen,
		T_DHMS,
		&checker,
	)
}
