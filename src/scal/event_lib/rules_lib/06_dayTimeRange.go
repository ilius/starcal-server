package rules_lib

import "log"
import "scal"

const R_dayTimeRange = "dayTimeRange"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.HMSRange)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_dayTimeRange,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		6,
		R_dayTimeRange,
		T_HMSRange,
		&checker,
	)
}
