package rules_lib

import (
	"log"

	lib "github.com/ilius/libgostarcal"
)

const R_cycleLen = "cycleLen"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(lib.DHMS)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_cycleLen,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		7,
		R_cycleLen,
		T_DHMS,
		&checker,
	)
}
