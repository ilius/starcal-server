package rules_lib

import "scal"

const R_ex_dates = "ex_dates"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]scal.Date)
		if !ok {
			return false
		}
		for _, date := range list {
			if !date.IsValid() {
				return false
			}
		}
		return true
	}
	RegisterRuleType(
		4,
		R_ex_dates,
		T_Date_list,
		&checker,
	)
}
