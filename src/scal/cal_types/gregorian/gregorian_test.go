package gregorian

import "testing"

import "scal"

func TestIsLeap(t *testing.T) {
	testMap := map[int]bool{
		1990: false,
		1991: false,
		1992: true,
		1993: false,
		1994: false,
		1995: false,
		1996: true,
		1997: false,
		1998: false,
		1999: false,
		2000: true,
		2001: false,
		2002: false,
		2003: false,
		2004: true,
		2005: false,
		2006: false,
		2007: false,
		2008: true,
		2009: false,
		2010: false,
		2011: false,
		2012: true,
		2013: false,
		2014: false,
		2015: false,
		2016: true,
		2017: false,
		2018: false,
		2019: false,
		2020: true,
		2021: false,
		2022: false,
		2023: false,
		2024: true,
		2025: false,
		2026: false,
		2027: false,
		2028: true,
		2029: false,
	}
	for year, isLeap := range testMap {
		if isLeap != IsLeap(year) {
			t.Errorf("Wrong: year=%v, isLeap=%v", year, isLeap)
		}
	}
}

func TestToJd(t *testing.T) {
	testMap := map[scal.Date]int{
		scal.Date{2015, 1, 1}:  2457024,
		scal.Date{2015, 2, 1}:  2457055,
		scal.Date{2015, 3, 1}:  2457083,
		scal.Date{2015, 4, 1}:  2457114,
		scal.Date{2015, 5, 1}:  2457144,
		scal.Date{2015, 6, 1}:  2457175,
		scal.Date{2015, 7, 1}:  2457205,
		scal.Date{2015, 8, 1}:  2457236,
		scal.Date{2015, 9, 1}:  2457267,
		scal.Date{2015, 10, 1}: 2457297,
		scal.Date{2015, 11, 1}: 2457328,
		scal.Date{2015, 12, 1}: 2457358,
		scal.Date{2016, 1, 1}:  2457389,
		scal.Date{2016, 2, 1}:  2457420,
		scal.Date{2016, 3, 1}:  2457449,
		scal.Date{2016, 4, 1}:  2457480,
		scal.Date{2016, 5, 1}:  2457510,
		scal.Date{2016, 6, 1}:  2457541,
		scal.Date{2016, 7, 1}:  2457571,
		scal.Date{2016, 8, 1}:  2457602,
		scal.Date{2016, 9, 1}:  2457633,
		scal.Date{2016, 10, 1}: 2457663,
		scal.Date{2016, 11, 1}: 2457694,
		scal.Date{2016, 12, 1}: 2457724,
		scal.Date{2017, 1, 1}:  2457755,
		scal.Date{2017, 2, 1}:  2457786,
		scal.Date{2017, 3, 1}:  2457814,
		scal.Date{2017, 4, 1}:  2457845,
		scal.Date{2017, 5, 1}:  2457875,
		scal.Date{2017, 6, 1}:  2457906,
		scal.Date{2017, 7, 1}:  2457936,
		scal.Date{2017, 8, 1}:  2457967,
		scal.Date{2017, 9, 1}:  2457998,
		scal.Date{2017, 10, 1}: 2458028,
		scal.Date{2017, 11, 1}: 2458059,
		scal.Date{2017, 12, 1}: 2458089,
	}
	for date, jd := range testMap {
		if jd != ToJd(date) {
			t.Errorf("Wrong: date=%v, jd=%v", date, jd)
		}
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
