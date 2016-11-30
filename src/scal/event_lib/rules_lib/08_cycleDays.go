package rules_lib

const R_cycleDays = "cycleDays"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(int)
		return ok && v > 0
	}
	RegisterRuleType(
		8,
		R_cycleDays,
		T_int,
		&checker,
	)
}
