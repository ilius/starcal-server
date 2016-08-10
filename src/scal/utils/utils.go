package utils

import "fmt"
import "sort"
import "time"
import "math"
import "errors"

import "scal"
import "scal/cal_types"
import "scal/cal_types/gregorian"

var MIN_INT64 = int64(-9223372036854775808)
var J1970 = 2440588

var DurationUnitByValue = map[int]string{
    1: "second",
    60: "minute",
    3600: "hour",
    86400: "day",
    604800: "week",
}

//DurationUnitList=[(1, "second"), (60, "minute"), (3600, "hour"), (86400, "day"), (604800, "week")]


func init() {
    fmt.Printf("")
    //fmt.Println("DurationUnitByValue =", DurationUnitByValue)
}

func IntMin(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func bisectLeftRange(a []int, v int, lo, hi int) int {  
    s := a[lo:hi]
    return sort.Search(len(s), func(i int) bool {
        return s[i] >= v
    })
}

func BisectLeft(a []int, v int) int {  
    return bisectLeftRange(a, v, 0, len(a))
}

// tested
func GetUtcOffsetByGDate(gdate scal.Date, loc *time.Location) int {
    t := time.Date(
        gdate.Year,
        time.Month(gdate.Month), // gdate.Month is int
        gdate.Day,
        0, // hour
        0, // min
        0, // sec
        0, // nsec
        loc, // location
    )
    _, offset := t.Zone() // zoneName, offset
    return offset
}

// tested
func GetUtcOffsetByEpoch(epoch int64, loc *time.Location) int {
    // is this working perfectly? FIXME
    // python code is too tricky
    t := time.Unix(epoch, 0).In(loc) // .In useful? FIXME
    _, offset := t.Zone() // zoneName, offset
    return offset
}

func GetUtcOffsetCurrent(loc *time.Location) int {
    t := time.Now().In(loc)
    _, offset := t.Zone() // zoneName, offset
    return offset
}

// tested
func GetEpochByGDate(gdate scal.Date, loc *time.Location) int64 {
    t := time.Date(
        gdate.Year,
        time.Month(gdate.Month), // gdate.Month is int
        gdate.Day,
        0, // hour
        0, // min
        0, // sec
        0, // nsec
        loc, // location
    )
    return t.Unix()
}

// tested
func GetEpochByJd(jd int, loc *time.Location) int64 {
    return GetEpochByGDate(gregorian.JdTo(jd), loc)
}
/*
func GetEpochByJd2(jd int, loc *time.Location) int64 {
    localEpoch := int64((jd-J1970) * 86400)
    offset := GetUtcOffsetByGDate(gdate, loc)
    epoch := localEpoch - offset
    offset2 := GetUtcOffsetByEpoch(epoch, loc)
    if offset2 != offset {
        fmt.Println("Warning: GetEpochByJd: offset mistmatch: delta =", offset2-offset, ", gdate =", gdate)
        epoch = localEpoch - offset2
        //3600 seconds error in days when DST is just changed
        //gdate = {2016 9 21}
        //gdate = {2017 9 22}
        //gdate = {2018 9 22}
        //gdate = {2019 9 22}
    }
    return epoch
}*/


func GetFloatJdByEpoch(epoch int64, loc *time.Location) float64 {
    offset := GetUtcOffsetByEpoch(epoch, loc)
    return float64(J1970) + float64(epoch + int64(offset)) / 86400.0
}

func GetJdByEpoch(epoch int64, loc *time.Location) int {
    return int(math.Floor(GetFloatJdByEpoch(epoch, loc)))
}

//RoundEpochToDay // not useful

func GetJdRangeFromEpochRange(startEpoch int64, endEpoch int64, loc *time.Location) (int, int) {
    startJd := GetJdByEpoch(startEpoch, loc)
    endJd := GetJdByEpoch(endEpoch - 1, loc) + 1
    return startJd, endJd
}

func GetHmsBySeconds(second int) scal.HMS {
    return scal.HMS{second/3600, second/60, second%60}
}

func GetJhmsByEpoch(epoch int64, loc *time.Location) (int, scal.HMS) {
    // return (jd, hour, minute, second)
    t := time.Unix(epoch, 0).In(loc) // .In useful? FIXME
    return gregorian.ToJd(scal.Date{
        t.Year(),
        int(t.Month()),
        t.Day(),
    }), scal.HMS{t.Hour(), t.Minute(), t.Second()}
}

func GetEpochByJhms(jd int, hms scal.HMS, loc *time.Location) int64 {
    gdate := gregorian.JdTo(jd)
    t := time.Date(
        gdate.Year,
        time.Month(gdate.Month), // gdate.Month is int
        gdate.Day,
        hms.Hour,
        hms.Minute,
        hms.Second,
        0, // nsec
        loc, // location
    )
    return t.Unix()
}

func GetJdAndSecondsFromEpoch(epoch int64, loc *time.Location) (int, int) {
    // return a tuple (julain_day, extra_seconds) from epoch
    jd, hms := GetJhmsByEpoch(epoch, loc)
    return jd, hms.GetTotalSeconds()
}


func GetCurrentDate(calTypeName string) (scal.Date, error) {
    t := time.Now() // .In(loc)
    if calTypeName == "gregorian" {
        return scal.Date{t.Year(), int(t.Month()), t.Day()}, nil
    }
    calType, ok := cal_types.CalTypesMap[calTypeName]
    if !ok {
        return scal.Date{},
            errors.New("invalid calendar type '" + calTypeName + "'")
    }
    loc := t.Location() // FIXME
    jd := GetJdByEpoch(t.Unix(), loc)
    return calType.JdTo(jd), nil
}









