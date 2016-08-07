package indian_national

import "fmt"

import "scal"

func TestIsLeap(start_year int, end_year int) {
    // 1930 to 1940
    for year:=start_year; year < end_year; year++ {
        if IsLeap(year) {
            fmt.Printf("%d is leap\n", year)
        } else {
            fmt.Printf("%d is not leap\n", year)
        }
    }
}

func TestToJd(start_year int, end_year int) {
    // 1930 to 1940
    for year:=start_year; year < end_year; year++ {
        for month:=1; month <= 12; month++ {
            var day = 1
            var jd = ToJd(scal.Date{year, month, day})
            fmt.Printf(
                "%.4d/%.2d/%.2d   jd=%d\n",
                year,
                month,
                day,
                jd,
            );
        }
    }
}

func TestConvert(start_year int, end_year int) {
    // 1930 to 1940
    for year:=start_year; year < end_year; year++ {
        for month:=1; month <= 12; month++ {
            var monthLen = GetMonthLen(year, month)
            for day:=1; day <= monthLen; day++ {
                var date = scal.Date{year, month, day}
                var jd = ToJd(date)
                var ndate = JdTo(jd)
                if date == ndate {
                    fmt.Printf("%.4d/%.2d/%.2d  OK\n", year, month, day);
                } else {
                    fmt.Printf(
                        "%.4d/%.2d/%.2d  =>  jd=%d  =>  %.4d/%.2d/%.2d\n",
                        year,
                        month,
                        day,
                        jd,
                        ndate.Year,
                        ndate.Month,
                        ndate.Day,
                    );
                }
            }
        }
    }
}


