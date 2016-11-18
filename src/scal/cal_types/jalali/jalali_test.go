package jalali

import "testing"

import "scal"

func TestIsLeap(t *testing.T) {
	testMap := map[int]bool{
		1360: false,
		1361: false,
		1362: true,
		1363: false,
		1364: false,
		1365: false,
		1366: true,
		1367: false,
		1368: false,
		1369: false,
		1370: true,
		1371: false,
		1372: false,
		1373: false,
		1374: false,
		1375: true,
		1376: false,
		1377: false,
		1378: false,
		1379: true,
		1380: false,
		1381: false,
		1382: false,
		1383: true,
		1384: false,
		1385: false,
		1386: false,
		1387: true,
		1388: false,
		1389: false,
		1390: false,
		1391: true,
		1392: false,
		1393: false,
		1394: false,
		1395: true,
		1396: false,
		1397: false,
		1398: false,
		1399: true,
	}
	for year, isLeap := range testMap {
		if isLeap != IsLeap(year) {
			t.Errorf("Wrong: year=%v, isLeap=%v", year, isLeap)
		}
	}
}

func TestToJd(t *testing.T) {
	testMap := map[scal.Date]int{
		scal.Date{1394, 1, 1}:  2457103,
		scal.Date{1394, 2, 1}:  2457134,
		scal.Date{1394, 3, 1}:  2457165,
		scal.Date{1394, 4, 1}:  2457196,
		scal.Date{1394, 5, 1}:  2457227,
		scal.Date{1394, 6, 1}:  2457258,
		scal.Date{1394, 7, 1}:  2457289,
		scal.Date{1394, 8, 1}:  2457319,
		scal.Date{1394, 9, 1}:  2457349,
		scal.Date{1394, 10, 1}: 2457379,
		scal.Date{1394, 11, 1}: 2457409,
		scal.Date{1394, 12, 1}: 2457439,
		scal.Date{1395, 1, 1}:  2457468,
		scal.Date{1395, 2, 1}:  2457499,
		scal.Date{1395, 3, 1}:  2457530,
		scal.Date{1395, 4, 1}:  2457561,
		scal.Date{1395, 5, 1}:  2457592,
		scal.Date{1395, 6, 1}:  2457623,
		scal.Date{1395, 7, 1}:  2457654,
		scal.Date{1395, 8, 1}:  2457684,
		scal.Date{1395, 9, 1}:  2457714,
		scal.Date{1395, 10, 1}: 2457744,
		scal.Date{1395, 11, 1}: 2457774,
		scal.Date{1395, 12, 1}: 2457804,
		scal.Date{1396, 1, 1}:  2457834,
		scal.Date{1396, 2, 1}:  2457865,
		scal.Date{1396, 3, 1}:  2457896,
		scal.Date{1396, 4, 1}:  2457927,
		scal.Date{1396, 5, 1}:  2457958,
		scal.Date{1396, 6, 1}:  2457989,
		scal.Date{1396, 7, 1}:  2458020,
		scal.Date{1396, 8, 1}:  2458050,
		scal.Date{1396, 9, 1}:  2458080,
		scal.Date{1396, 10, 1}: 2458110,
		scal.Date{1396, 11, 1}: 2458140,
		scal.Date{1396, 12, 1}: 2458170,
	}
	for date, jd := range testMap {
		if jd != ToJd(date) {
			t.Errorf("Wrong: date=%v, jd=%v", date, jd)
		}
	}
}

func TestConvert(t *testing.T) {
	start_year := 1350
	end_year := 1450
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
