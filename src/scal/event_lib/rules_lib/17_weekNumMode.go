package rules_lib

import "log"
import "scal"

const R_weekNumMode = "weekNumMode"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(string)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_weekNumMode,
				value,
				value,
			)
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
