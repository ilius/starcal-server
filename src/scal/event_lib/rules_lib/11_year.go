package rules_lib

const R_year = "year"

func init() {
	RegisterRuleType(
		11,
		R_year,
		T_int_range_list,
		nil,
	)
}
