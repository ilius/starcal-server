package rules_lib

import (
	"log"

	lib "github.com/ilius/libgostarcal"
)

const R_dayTimeRange = "dayTimeRange"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(lib.HMSRange)
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
