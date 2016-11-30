package rules_lib

const R_ex_year = "ex_year"

func init() {
	RegisterRuleType(
		12,
		R_ex_year,
		T_int_range_list,
		nil,
	)
}
