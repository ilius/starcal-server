package rules_lib

var RulesRequire = map[string][]string{
	R_duration:    {R_start},
	R_cycleLen:    {R_start},
	R_cycleDays:   {R_start},
	R_cycleWeeks:  {R_start},
	R_weekNumMode: {R_start},
}

var RulesConflictWith = map[string][]string{
	R_duration:     {R_end},
	R_date:         {R_start, R_end, R_duration},
	R_dayTimeRange: {R_dayTime},
	R_cycleLen: {
		R_date,
		R_dayTime,
		R_dayTimeRange,
	},
	R_cycleDays: {
		R_date,
		R_cycleLen,
	},
	R_cycleWeeks: {
		R_date,
		R_cycleLen,
		R_cycleDays,
	},
	R_weekDay:     {R_date},
	R_year:        {R_date},
	R_ex_year:     {R_date},
	R_month:       {R_date},
	R_ex_month:    {R_date, R_month},
	R_day:         {R_date},
	R_ex_day:      {R_date},
	R_weekNumMode: {R_date},
	R_weekMonth: {
		R_date,
		R_cycleLen,
		R_cycleDays,
		R_cycleWeeks,
		R_weekDay,
		R_month,
		R_ex_month,
		R_weekNumMode,
	},
}
