package indian_national

import "testing"

import "scal"

func TestIsLeap(t *testing.T) {
    testMap := map[int]bool{
        1920: false,
        1921: false,
        1922: true,
        1923: false,
        1924: false,
        1925: false,
        1926: true,
        1927: false,
        1928: false,
        1929: false,
        1930: true,
        1931: false,
        1932: false,
        1933: false,
        1934: true,
        1935: false,
        1936: false,
        1937: false,
        1938: true,
        1939: false,
        1940: false,
        1941: false,
        1942: true,
        1943: false,
        1944: false,
        1945: false,
        1946: true,
        1947: false,
        1948: false,
        1949: false,
    }
    for year, isLeap := range testMap {
        if isLeap != IsLeap(year) {
            t.Errorf("Wrong: year=%v, isLeap=%v", year, isLeap)
        }
    }
}


func TestToJd(t *testing.T) {
    testMap := map[scal.Date]int{
        scal.Date{1936, 1, 1}: 2456739,
        scal.Date{1936, 2, 1}: 2456769,
        scal.Date{1936, 3, 1}: 2456800,
        scal.Date{1936, 4, 1}: 2456831,
        scal.Date{1936, 5, 1}: 2456862,
        scal.Date{1936, 6, 1}: 2456893,
        scal.Date{1936, 7, 1}: 2456924,
        scal.Date{1936, 8, 1}: 2456954,
        scal.Date{1936, 9, 1}: 2456984,
        scal.Date{1936, 10, 1}: 2457014,
        scal.Date{1936, 11, 1}: 2457044,
        scal.Date{1936, 12, 1}: 2457074,
        scal.Date{1937, 1, 1}: 2457104,
        scal.Date{1937, 2, 1}: 2457134,
        scal.Date{1937, 3, 1}: 2457165,
        scal.Date{1937, 4, 1}: 2457196,
        scal.Date{1937, 5, 1}: 2457227,
        scal.Date{1937, 6, 1}: 2457258,
        scal.Date{1937, 7, 1}: 2457289,
        scal.Date{1937, 8, 1}: 2457319,
        scal.Date{1937, 9, 1}: 2457349,
        scal.Date{1937, 10, 1}: 2457379,
        scal.Date{1937, 11, 1}: 2457409,
        scal.Date{1937, 12, 1}: 2457439,
        scal.Date{1938, 1, 1}: 2457469,
        scal.Date{1938, 2, 1}: 2457500,
        scal.Date{1938, 3, 1}: 2457531,
        scal.Date{1938, 4, 1}: 2457562,
        scal.Date{1938, 5, 1}: 2457593,
        scal.Date{1938, 6, 1}: 2457624,
        scal.Date{1938, 7, 1}: 2457655,
        scal.Date{1938, 8, 1}: 2457685,
        scal.Date{1938, 9, 1}: 2457715,
        scal.Date{1938, 10, 1}: 2457745,
        scal.Date{1938, 11, 1}: 2457775,
        scal.Date{1938, 12, 1}: 2457805,
    }
    for date, jd := range testMap {
        if jd != ToJd(date) {
            t.Errorf("Wrong: date=%v, jd=%v", date, jd)
        }
    }
}


func TestConvert(t *testing.T) {
    start_year := 1920
    end_year := 2950
    for year:=start_year; year < end_year; year++ {
        for month:=1; month <= 12; month++ {
            var monthLen = GetMonthLen(year, month)
            for day:=1; day <= monthLen; day++ {
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
                    );
                }
            }
        }
    }
}



