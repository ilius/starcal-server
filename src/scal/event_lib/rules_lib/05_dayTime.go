package rules_lib

import "log"
import "scal"

const R_dayTime = "dayTime"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.HMS)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_dayTime,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		5,
		R_dayTime,
		T_HMS,
		&checker,
	)
}
