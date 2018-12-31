package ethiopian

import "testing"

import "scal"

func TestIsLeap(t *testing.T) {
	testMap := map[int]bool{
		1990: false,
		1991: true,
		1992: false,
		1993: false,
		1994: false,
		1995: true,
		1996: false,
		1997: false,
		1998: false,
		1999: true,
		2000: false,
		2001: false,
		2002: false,
		2003: true,
		2004: false,
		2005: false,
		2006: false,
		2007: true,
		2008: false,
		2009: false,
		2010: false,
		2011: true,
		2012: false,
		2013: false,
		2014: false,
		2015: true,
		2016: false,
		2017: false,
		2018: false,
		2019: true,
		2020: false,
		2021: false,
		2022: false,
		2023: true,
		2024: false,
		2025: false,
		2026: false,
		2027: true,
		2028: false,
		2029: false,
	}
	for year, isLeap := range testMap {
		if isLeap != IsLeap(year) {
			t.Errorf("Wrong: year=%v, isLeap=%v", year, isLeap)
		}
		//t.Log(year, isLeap)
	}
}

func TestToJd(t *testing.T) {
	testMap := map[scal.Date]int{
		{2015, 1, 1}:  2459834,
		{2015, 2, 1}:  2459864,
		{2015, 3, 1}:  2459894,
		{2015, 4, 1}:  2459924,
		{2015, 5, 1}:  2459954,
		{2015, 6, 1}:  2459984,
		{2015, 7, 1}:  2460014,
		{2015, 8, 1}:  2460044,
		{2015, 9, 1}:  2460074,
		{2015, 10, 1}: 2460104,
		{2015, 11, 1}: 2460134,
		{2015, 12, 1}: 2460164,
		{2016, 1, 1}:  2460200,
		{2016, 2, 1}:  2460230,
		{2016, 3, 1}:  2460260,
		{2016, 4, 1}:  2460290,
		{2016, 5, 1}:  2460320,
		{2016, 6, 1}:  2460350,
		{2016, 7, 1}:  2460380,
		{2016, 8, 1}:  2460410,
		{2016, 9, 1}:  2460440,
		{2016, 10, 1}: 2460470,
		{2016, 11, 1}: 2460500,
		{2016, 12, 1}: 2460530,
		{2017, 1, 1}:  2460565,
		{2017, 2, 1}:  2460595,
		{2017, 3, 1}:  2460625,
		{2017, 4, 1}:  2460655,
		{2017, 5, 1}:  2460685,
		{2017, 6, 1}:  2460715,
		{2017, 7, 1}:  2460745,
		{2017, 8, 1}:  2460775,
		{2017, 9, 1}:  2460805,
		{2017, 10, 1}: 2460835,
		{2017, 11, 1}: 2460865,
		{2017, 12, 1}: 2460895,
	}
	for date, jd := range testMap {
		if jd != ToJd(date) {
			t.Errorf("Wrong: date=%v, jd=%v", date, jd)
		}
		//t.Log(date, jd)
	}
}

func TestConvert(t *testing.T) {
	start_year := 1970
	end_year := 2050
	for year := start_year; year < end_year; year++ {
		for month := 1; month <= 12; month++ {
			var monthLen = GetMonthLen(year, month)
			for day := 1; day <= monthLen; day++ {
				var date = scal.Date{year, month, day}
				var jd = ToJd(date)
				var ndate = JdTo(jd)
				if date == ndate {
					//t.Logf("%v  OK\n", date);
				} else {
					t.Errorf(
						"Wrong: %v  =>  jd=%d  =>  %v\n",
						date,
						jd,
						ndate,
					)
				}
			}
		}
	}
}
