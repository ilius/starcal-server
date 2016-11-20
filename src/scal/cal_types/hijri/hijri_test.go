package hijri

import "testing"

import "scal"

func TestIsLeap(t *testing.T) {
	testMap := map[int]bool{
		1410: false,
		1411: false,
		1412: true,
		1413: false,
		1414: false,
		1415: true,
		1416: false,
		1417: true,
		1418: false,
		1419: false,
		1420: true,
		1421: false,
		1422: false,
		1423: true,
		1424: false,
		1425: false,
		1426: true,
		1427: false,
		1428: true,
		1429: false,
		1430: false,
		1431: true,
		1432: false,
		1433: false,
		1434: true,
		1435: false,
		1436: true,
		1437: false,
		1438: false,
		1439: true,
		1440: false,
		1441: false,
		1442: true,
		1443: false,
		1444: false,
		1445: true,
		1446: false,
		1447: true,
		1448: false,
		1449: false,
	}
	for year, isLeap := range testMap {
		if isLeap != IsLeap(year) {
			t.Errorf("Wrong: year=%v, isLeap=%v", year, isLeap)
		}
	}
}

func TestToJd(t *testing.T) {
	testMap := map[scal.Date]int{
		scal.Date{1436, 1, 1}:  2456957,
		scal.Date{1436, 2, 1}:  2456987,
		scal.Date{1436, 3, 1}:  2457016,
		scal.Date{1436, 4, 1}:  2457046,
		scal.Date{1436, 5, 1}:  2457075,
		scal.Date{1436, 6, 1}:  2457105,
		scal.Date{1436, 7, 1}:  2457134,
		scal.Date{1436, 8, 1}:  2457164,
		scal.Date{1436, 9, 1}:  2457193,
		scal.Date{1436, 10, 1}: 2457223,
		scal.Date{1436, 11, 1}: 2457252,
		scal.Date{1436, 12, 1}: 2457282,
		scal.Date{1437, 1, 1}:  2457312,
		scal.Date{1437, 2, 1}:  2457342,
		scal.Date{1437, 3, 1}:  2457371,
		scal.Date{1437, 4, 1}:  2457401,
		scal.Date{1437, 5, 1}:  2457430,
		scal.Date{1437, 6, 1}:  2457460,
		scal.Date{1437, 7, 1}:  2457489,
		scal.Date{1437, 8, 1}:  2457519,
		scal.Date{1437, 9, 1}:  2457548,
		scal.Date{1437, 10, 1}: 2457578,
		scal.Date{1437, 11, 1}: 2457607,
		scal.Date{1437, 12, 1}: 2457637,
		scal.Date{1438, 1, 1}:  2457666,
		scal.Date{1438, 2, 1}:  2457696,
		scal.Date{1438, 3, 1}:  2457725,
		scal.Date{1438, 4, 1}:  2457755,
		scal.Date{1438, 5, 1}:  2457784,
		scal.Date{1438, 6, 1}:  2457814,
		scal.Date{1438, 7, 1}:  2457843,
		scal.Date{1438, 8, 1}:  2457873,
		scal.Date{1438, 9, 1}:  2457902,
		scal.Date{1438, 10, 1}: 2457932,
		scal.Date{1438, 11, 1}: 2457961,
		scal.Date{1438, 12, 1}: 2457991,
		scal.Date{1439, 1, 1}:  2458020,
		scal.Date{1439, 2, 1}:  2458050,
		scal.Date{1439, 3, 1}:  2458079,
		scal.Date{1439, 4, 1}:  2458109,
		scal.Date{1439, 5, 1}:  2458138,
		scal.Date{1439, 6, 1}:  2458168,
		scal.Date{1439, 7, 1}:  2458197,
		scal.Date{1439, 8, 1}:  2458227,
		scal.Date{1439, 9, 1}:  2458256,
		scal.Date{1439, 10, 1}: 2458286,
		scal.Date{1439, 11, 1}: 2458315,
		scal.Date{1439, 12, 1}: 2458345,
	}
	for date, jd := range testMap {
		if jd != ToJd(date) {
			t.Errorf("Wrong: date=%v, jd=%v", date, jd)
		}
	}
}

func TestMonthLen(t *testing.T) {
	testMap := map[[2]int]int{
		{1436, 1}:  30,
		{1436, 2}:  29,
		{1436, 3}:  30,
		{1436, 4}:  29,
		{1436, 5}:  30,
		{1436, 6}:  29,
		{1436, 7}:  30,
		{1436, 8}:  29,
		{1436, 9}:  30,
		{1436, 10}: 29,
		{1436, 11}: 30,
		{1436, 12}: 30,
		{1437, 1}:  30,
		{1437, 2}:  29,
		{1437, 3}:  30,
		{1437, 4}:  29,
		{1437, 5}:  30,
		{1437, 6}:  29,
		{1437, 7}:  30,
		{1437, 8}:  29,
		{1437, 9}:  30,
		{1437, 10}: 29,
		{1437, 11}: 30,
		{1437, 12}: 29,
		{1438, 1}:  30,
		{1438, 2}:  29,
		{1438, 3}:  30,
		{1438, 4}:  29,
		{1438, 5}:  30,
		{1438, 6}:  29,
		{1438, 7}:  30,
		{1438, 8}:  29,
		{1438, 9}:  30,
		{1438, 10}: 29,
		{1438, 11}: 30,
		{1438, 12}: 29,
		{1439, 1}:  30,
		{1439, 2}:  29,
		{1439, 3}:  30,
		{1439, 4}:  29,
		{1439, 5}:  30,
		{1439, 6}:  29,
		{1439, 7}:  30,
		{1439, 8}:  29,
		{1439, 9}:  30,
		{1439, 10}: 29,
		{1439, 11}: 30,
		{1439, 12}: 30,
	}
	for ym, mLen := range testMap {
		if mLen != GetMonthLen(ym[0], ym[1]) {
			t.Errorf("Wrong: mLen=%v, year=%v, month=%v", mLen, ym[0], ym[1])
		}
	}
}

func TestConvert(t *testing.T) {
	start_year := 1390
	end_year := 1480
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
