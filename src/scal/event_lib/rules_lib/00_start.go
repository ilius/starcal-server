package rules_lib

import (
	"log"

	lib "github.com/ilius/libgostarcal"
)

const R_start = "start"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(lib.DateHMS)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_start,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		0,
		R_start,
		T_DateHMS,
		&checker,
	)
}
