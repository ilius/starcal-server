package rules_lib

import (
	"testing"
)

type RuleTestCase struct {
	Type     string
	Value    string
	DecodeOk bool
	CheckOk  bool
}

func TestRules(t *testing.T) {
	testCases := []RuleTestCase{
		{"cycleDays", "10", true, true},
		{"cycleDays", "-1", true, false},
		{"cycleDays", "2f", false, false},
		{"cycleLen", "90 23:55:55", true, true},
		{"cycleLen", "-1 23:55:55", true, false},
		{"cycleLen", "10 -1:55:55", true, false},
		{"cycleLen", "10", false, false},
		{"cycleLen", "10 a:b:c", false, false},
		{"date", "2000/1/1", true, true},
		{"date", "-1000/1/1", true, true},
		{"date", "-1000/12/31", true, true},
		{"date", "2000/13/31", true, false},
		{"date", "2000/12/-1", true, false},
		{"date", "2000-1-1", false, false},
		{"date", "2000", false, false},
		{"ex_dates", "2000/1/1 2010/12/1 -1000/1/31", true, true},
		{"ex_dates", "aa/1/1 2010/12/1 -1000/1/31", false, false},
		{"day", "1", true, true},
		{"day", "1 5 10 15 30", true, true},
		{"day", "31", true, true},
		{"day", "0", true, false},
		{"day", "50", true, false},
		{"day", "1 50 5 10 15 30", true, false},
		{"day", "ff", false, false},
		{"ex_day", "1", true, true},
		{"ex_day", "1 5 10 15 30", true, true},
		{"ex_day", "31", true, true},
		{"ex_day", "0", true, false},
		{"ex_day", "50", true, false},
		{"ex_day", "1 50 5 10 15 30", true, false},
		{"ex_day", "ff", false, false},
		{"dayTime", "22:30:55", true, true},
		{"dayTime", "24:30:55", true, false},
		{"dayTime", "20:60:55", true, false},
		{"dayTime", "20:55:61", true, false},
		{"dayTime", "20:55:-1", true, false},
		{"dayTime", "205501", false, false},
		{"dayTime", "ab", false, false},
		{"dayTimeRange", "22:30:55 23:30:55", true, true},
		{"dayTimeRange", "22:30:55 24:30:55", true, false},
		{"dayTimeRange", "22:30:55-23:30:55", false, false},
		{"duration", "3.1 d", true, true},
		{"duration", "-1 d", true, false},
		{"duration", "-1.5 w", true, false},
		{"duration", "1d", false, false},
		{"end", "2016/12/31 23:55:55", true, true},
		{"end", "2016/12/31 23:55:61", true, false},
		{"end", "2016/12/31-23:55:55", false, false},
		{"month", "1", true, true},
		{"month", "12", true, true},
		{"month", "0", true, false},
		{"month", "bb", false, false},
		{"ex_month", "1", true, true},
		{"ex_month", "12", true, true},
		{"ex_month", "0", true, false},
		{"ex_month", "bb", false, false},
		{"start", "-1/11/30 22:50:00", true, true},
		{"start", "-1/0/30 22:50:00", true, false},
		{"start", "-1/11/30", false, false},
		{"weekDay", "0 2 4 6", true, true},
		{"weekDay", "-1 2 4 6", true, false},
		{"weekDay", "0 2 4 7", true, false},
		{"weekDay", "a 2 4 7", false, false},
		{"weekNumMode", "odd", true, true},
		{"weekNumMode", "even", true, true},
		{"weekNumMode", "any", true, true},
		{"weekNumMode", "foo", true, false},
		{"year", "-1000 100 0 1", true, true},
		{"year", "ff -1000 100 0 1", false, false},
		{"ex_year", "-1000 100 0 1", true, true},
		{"ex_year", "ff -1000 100 0 1", false, false},
		{
			"weekMonth",
			"{\"weekIndex\": 4, \"weekDay\": 6, \"month\": 12}",
			true,
			true,
		},
		{
			"weekMonth",
			"{\"weekIndex\": 1, \"weekDay\": 1, \"month\": 0}",
			true,
			true,
		},
		{
			"weekMonth",
			"{\"weekIndex\": 5, \"weekDay\": 6, \"month\": 12}",
			true,
			false,
		},
		{
			"weekMonth",
			"{\"weekIndex\": 0, \"weekDay\": 7, \"month\": 12}",
			true,
			false,
		},
		{
			"weekMonth",
			"\"weekIndex\": 0, \"weekDay\": 0, \"month\": 0",
			false,
			false,
		},
	}
	for _, tc := range testCases {
		model := EventRuleModel{
			Type:  tc.Type,
			Value: tc.Value,
		}
		rule, err := model.Decode()
		if (err == nil) != tc.DecodeOk {
			t.Errorf("mismatch DecodeOk: tc=%v, err=%v\n", tc, err)
		}
		if err != nil {
			continue
		}
		checkOk := rule.Check()
		if checkOk != tc.CheckOk {
			t.Errorf(
				"mismatch CheckOk: tc=%v, checkOk=%v, rule.Value=%v\n",
				tc,
				checkOk,
				rule.Value,
			)
		}
	}
}
