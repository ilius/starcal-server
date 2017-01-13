package rules_lib

import "log"
import "scal"

const R_ex_dates = "ex_dates"

func init() {
	checker := func(value interface{}) bool {
		list, ok := value.([]scal.Date)
		if !ok {
			log.Printf(
				"%s rule value checker: type conversion failed, value=%v with type %T\n",
				R_ex_dates,
				value,
				value,
			)
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
