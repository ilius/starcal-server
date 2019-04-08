package rules_lib

import (
	"log"

	lib "github.com/ilius/libgostarcal"
)

const R_date = "date"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(lib.Date)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_date,
				value,
				value,
			)
		}
		return ok && v.IsValid()
	}
	RegisterRuleType(
		3,
		R_date,
		T_Date,
		&checker,
	)
}
