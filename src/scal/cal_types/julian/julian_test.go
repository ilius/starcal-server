package julian

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
		scal.Date{2000, 1, 1}: 2451558,
		scal.Date{2001, 1, 1}: 2451924,
		scal.Date{2002, 1, 1}: 2452289,
		scal.Date{2003, 1, 1}: 2452654,
		scal.Date{2004, 1, 1}: 2453019,
		scal.Date{2005, 1, 1}: 2453385,
		scal.Date{2006, 1, 1}: 2453750,
		scal.Date{2007, 1, 1}: 2454115,
		scal.Date{2008, 1, 1}: 2454480,
		scal.Date{2009, 1, 1}: 2454846,
		scal.Date{2010, 1, 1}: 2455211,
		scal.Date{2011, 1, 1}: 2455576,
		scal.Date{2012, 1, 1}: 2455941,
		scal.Date{2013, 1, 1}: 2456307,
		scal.Date{2014, 1, 1}: 2456672,
		scal.Date{2015, 1, 1}: 2457037,
		scal.Date{2016, 1, 1}: 2457402,
		scal.Date{2017, 1, 1}: 2457768,
		scal.Date{2018, 1, 1}: 2458133,
		scal.Date{2019, 1, 1}: 2458498,
		scal.Date{2020, 1, 1}: 2458863,
		scal.Date{2021, 1, 1}: 2459229,
		scal.Date{2022, 1, 1}: 2459594,
		scal.Date{2023, 1, 1}: 2459959,
		scal.Date{2024, 1, 1}: 2460324,
		scal.Date{2025, 1, 1}: 2460690,
		scal.Date{2026, 1, 1}: 2461055,
		scal.Date{2027, 1, 1}: 2461420,
		scal.Date{2028, 1, 1}: 2461785,
		scal.Date{2029, 1, 1}: 2462151,
		//scal.Date{2015, 1, 1}: 2457037,
		scal.Date{2015, 2, 1}:  2457068,
		scal.Date{2015, 3, 1}:  2457096,
		scal.Date{2015, 4, 1}:  2457127,
		scal.Date{2015, 5, 1}:  2457157,
		scal.Date{2015, 6, 1}:  2457188,
		scal.Date{2015, 7, 1}:  2457218,
		scal.Date{2015, 8, 1}:  2457249,
		scal.Date{2015, 9, 1}:  2457280,
		scal.Date{2015, 10, 1}: 2457310,
		scal.Date{2015, 11, 1}: 2457341,
		scal.Date{2015, 12, 1}: 2457371,
		//scal.Date{2016, 1, 1}: 2457402,
		scal.Date{2016, 2, 1}:  2457433,
		scal.Date{2016, 3, 1}:  2457462,
		scal.Date{2016, 4, 1}:  2457493,
		scal.Date{2016, 5, 1}:  2457523,
		scal.Date{2016, 6, 1}:  2457554,
		scal.Date{2016, 7, 1}:  2457584,
		scal.Date{2016, 8, 1}:  2457615,
		scal.Date{2016, 9, 1}:  2457646,
		scal.Date{2016, 10, 1}: 2457676,
		scal.Date{2016, 11, 1}: 2457707,
		scal.Date{2016, 12, 1}: 2457737,
		//scal.Date{2017, 1, 1}: 2457768,
		scal.Date{2017, 2, 1}:  2457799,
		scal.Date{2017, 3, 1}:  2457827,
		scal.Date{2017, 4, 1}:  2457858,
		scal.Date{2017, 5, 1}:  2457888,
		scal.Date{2017, 6, 1}:  2457919,
		scal.Date{2017, 7, 1}:  2457949,
		scal.Date{2017, 8, 1}:  2457980,
		scal.Date{2017, 9, 1}:  2458011,
		scal.Date{2017, 10, 1}: 2458041,
		scal.Date{2017, 11, 1}: 2458072,
		scal.Date{2017, 12, 1}: 2458102,
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
