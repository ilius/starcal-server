package rules_lib

import (
	"log"

	"github.com/ilius/libgostarcal/utils"
)

const R_duration = "duration"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(utils.Duration)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_duration,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		2,
		R_duration,
		T_Duration,
		&checker,
	)
}
