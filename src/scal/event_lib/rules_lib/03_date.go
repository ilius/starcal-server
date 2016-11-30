package rules_lib

import "scal"

const R_date = "date"

func init() {
	checker := func(value interface{}) bool {
		v, ok := value.(scal.Date)
		return ok && v.IsValid()
	}
	RegisterRuleType(
		3,
		R_date,
		T_Date,
		&checker,
	)
}
