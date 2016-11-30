package rules_lib

import "scal"

const R_weekNumMode = "weekNumMode"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(string)
		if !ok {
			return false
		}
		switch v {
		case scal.ODD:
			return true
		case scal.EVEN:
			return true
		case scal.ANY:
			return true
		}
		return false
	}
	RegisterRuleType(
		17,
		R_weekNumMode,
		T_string,
		&checker,
	)
}
