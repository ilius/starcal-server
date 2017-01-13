package rules_lib

import "log"
import "scal"

const R_end = "end"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.DateHMS)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_end,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		1,
		R_end,
		T_DateHMS,
		&checker,
	)
}
