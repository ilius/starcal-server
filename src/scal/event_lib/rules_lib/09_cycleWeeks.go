package rules_lib

const R_cycleWeeks = "cycleWeeks"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(int)
		return ok && v > 0
	}
	RegisterRuleType(
		9,
		R_cycleWeeks,
		T_int,
		&checker,
	)
}
