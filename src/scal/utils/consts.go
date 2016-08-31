package utils

var MIN_INT64 = int64(-9223372036854775808)
var J1970 = 2440588

var DayLen = 24*3600

var IcsMinStartYear = 1970
var IcsMaxEndYear = 2050

var DurationUnitByValue = map[int]string{
    1: "second",
    60: "minute",
    3600: "hour",
    86400: "day",
    604800: "week",
}

//DurationUnitList=[(1, "second"), (60, "minute"), (3600, "hour"), (86400, "day"), (604800, "week")]


func init() {
    //fmt.Printf("")
    //fmt.Println("DurationUnitByValue =", DurationUnitByValue)
}
