package main

import . "fmt"
import "time"
import "math/rand"
import "os"
import "strings"

import "scal"
import _ "scal/cal_types/loader"
import . "scal/cal_types"
import . "scal/utils"
import . "scal/interval"

import "scal/cal_types/gregorian"
//import "scal/cal_types/jalali"
//import "scal/cal_types/indian_national"

/*
func testJalali() {
    jalali.TestIsLeap(2000, 2020)
    jalali.TestToJd(2000, 2010)
    jalali.TestConvert(2000, 2010)
}
func testGregorian() {
    gregorian.TestIsLeap(2000, 2020)
    gregorian.TestToJd(2000, 2020)
    gregorian.TestConvert(2000, 2020)
}
func testIndian() {
    indian_national.TestIsLeap(1930, 1950)
    indian_national.TestToJd(1930, 1950)
    indian_national.TestConvert(1930, 1950)
}
*/

func ShuffleIntervals(a []Interval) {
    rand.Seed(time.Now().UnixNano())
    for i := range a {
        j := rand.Intn(i + 1)
        a[i], a[j] = a[j], a[i]
    }
}



func testConvert() {
    Println(CalTypesMap["gregorian"])
    gdate := scal.Date{2016, 1, 1}
    jdate, err := Convert(gdate, "gregorian", "indian_national")
    Printf("%v => %v   error=%v\n", gdate, jdate, err)
}
func testToJd() {
    gdate := scal.Date{2016, 1, 1}
    jd, err := ToJd(gdate, "gregorian")
    Println(jd, err)
    gdate2, err2 := JdTo(jd, "gregorian")
    Println(gdate2, err2)
}

func testTimeZone() {
    //t := time.Now()// .In(loc)
    //loc := t.Location()
    loc, err := time.LoadLocation("Asia/Tehran")
    if err != nil {Println("err: ", err.Error())}
    Println("Location:", loc)
    //utc, err := time.LoadLocation("UTC")
    //if err != nil {Println("err: ", err.Error())}
    //Println("Location:", utc, "    Time:", t.In(utc))
    //epoch := int(t.Unix())
    //jd := GetJdByEpoch(epoch, loc) // OK
    //Println("jd =", jd)
    //date := CalTypesMap["gregorian"].JdTo(jd)
    date, _ := GetCurrentDate("gregorian")
    Println("Date:", date)
    startJd := gregorian.ToJd(scal.Date{2016, 1, 1})
    endJd := gregorian.ToJd(scal.Date{2020, 1, 1})
    for jd:=startJd; jd < endJd; jd++ {
        epoch1 := GetEpochByJd(jd, loc)
        gdate := gregorian.JdTo(jd)
        t := time.Date(
            gdate.Year,
            time.Month(gdate.Month), // date.Month is int
            gdate.Day,
            0, // hour
            0, // min
            0, // sec
            0, // nsec
            loc, // location
        )
        epoch2 := t.Unix()
        if epoch1 == epoch2 {
            //Println("Epoch OK")
        } else {
            Println("EPOCH MISATCH, delta =", epoch2-epoch1, "  , gdate =", gdate)
        }
        floatJd := GetFloatJdByEpoch(epoch1, loc)
        //Printf("floatJd = %f\n", floatJd)
        dayPortion := floatJd - float64(int(floatJd))
        if dayPortion > 0 {
            Printf("gdate=%v, dayPortion = %f\n", gdate, dayPortion)
            floatJd2 := GetFloatJdByEpoch(epoch1-1, loc)
            dayPortion2 := floatJd2 - float64(int(floatJd2))
            Printf("gdate=%v, dayPortion2 = %f\n\n", gdate, dayPortion2)
        }
    }
}

func testGetJdListFromEpochRange() {
    t := time.Now()
    nowEpoch := t.Unix()
    loc := t.Location()
    startJd, endJd := GetJdRangeFromEpochRange(nowEpoch-10*24*3600, nowEpoch, loc)
    for jd := startJd; jd <= endJd; jd++ {
        Println("jd =", jd)
    }
}

func testGetJhmsByEpoch() {
    t := time.Now()
    epoch := t.Unix()
    loc := t.Location()
    jd, hms := GetJhmsByEpoch(epoch, loc)
    Println("GetJhmsByEpoch", jd, hms)
    epoch2 := GetEpochByJhms(jd, hms, loc)
    Println(epoch2 - epoch)
    Println(GetJdAndSecondsFromEpoch(epoch, loc))
}

func testHMS_FloatHour(){
    hms := scal.HMS{12, 59, 5}
    fh := hms.GetFloatHour()
    hms2 := scal.FloatHourToHMS(fh)
    Println("hms =", hms)
    Println("fh =", fh)
    Println("hms2 =", hms2)
}

func testDecodeHMS() {
    s := "12:01:01"
    hms, err := scal.DecodeHMS(s)
    Println(hms, err)
}

func testIntervalListByNumList() {
    nums := []int64{1, 2, 3, 4, 5, 7, 9, 10, 14, 16, 17, 18, 19, 21, 22, 23, 24}
    Println(nums)
    Println(IntervalListByNumList(nums, 3))
}

func testIntervalListNormalize() {
    testMap := map[string]string{
        "10-20 20": "10-20 20",
        "10-20 20 20": "10-20 20",
        "10-20 20-30": "10-30",
        "10-20 20-30 30-40": "10-40",
        "10-20 20-30 25-40": "10-40",
        "1-10 14 2-5 9-13 16-18 17-20 15-16 25-30": "1-13 14 15-20 25-30",
        "60-70 0-40 10-50 20-30 80-90 70-80 85-100 110 55": "0-50 55 60-100 110",
    }
    testCount := len(testMap)
    succeedCount := 0
    for testStr, answerStr := range testMap {
        //Println(testStr, "=>", answerStr)
        testList, testErr := ParseIntervalList(testStr)
        if testErr != nil {
            Println(testErr)
            continue
        }
        answerList, answerErr := ParseIntervalList(answerStr)
        if answerErr != nil {
            Println(answerErr)
            continue
        }
        ShuffleIntervals(testList) // FIXME
        testList, testErr = testList.Normalize()
        if testErr != nil {
            Println(testErr)
            continue
        }
        testList = testList.Humanize()
        if testList.String() != answerList.String() {// FIXME
            Println("test failed, result doesn't match the answer")
            Println("  testList =", testList)
            Println("answerList =", answerList)
        }
        Println(testStr, "=>", answerList)
        succeedCount ++
    }
    failedCount := testCount - succeedCount
    Printf("%d tests out of %d succeeded, %d failed\n", succeedCount, testCount, failedCount)
    
    argc := len(os.Args)
    if argc > 1 {
        listStr := strings.Join(os.Args[1:argc], " ")
        //listStr := 
        list, err := ParseIntervalList(listStr)
        if err != nil {
            Println(err)
            return
        }
        Println("Original Intervals:", list)
        ShuffleIntervals(list) // FIXME
        Println("Shuffled Intervals:", list)
        list, err = list.Normalize()
        if err != nil {
            Println(err)
            return
        }
        Println("    New Intervals:", list)
        //Printf("    New Intervals: %v\n", list)
    }
}

func getIntersectionString(list1Str string, list2Str string) string {
    list1, err1 := ParseIntervalList(list1Str)
    list2, err2 := ParseIntervalList(list2Str)
    if err1 != nil {
        Println(err1)
        return ""
    }
    if err2 != nil {
        Println(err2)
        return ""
    }
    ShuffleIntervals(list1)
    ShuffleIntervals(list2)

    result, err := list1.Intersection(list2)
    if err != nil {
        Println(err)
        return ""
    }
    result = result.Humanize()
    return result.String()
}

func testIntervalListIntersection() {
    type p [2]string

    testMap := map[[2]string]string{
        p{
            "0-20",
            "10-30",
        }:  "10-20",

        p{
            "10-30 40-50 60-80",
            "25-45",
        }:  "25-30 40-45",

        p{
            "10-30 40-50 60-80",
            "25-45 50-60",
        }:  "25-30 40-45",

        p{
            "10-30 40-50 60-80",
            "25-45 50-60 60",
        }:  "25-30 40-45 60",

        p{
            "10-30 40-50 60-80",
            "25-45 48-70 60",
        }:  "25-30 40-45 48-50 60-70",

        p{
            "10-30 40-50 60-80",
            "25-45 48-70",
        }:  "25-30 40-45 48-50 60-70",

        p{
            "0-10 20-30 40-50 60-70",
            "1-2 6-7 11-12 16-17 21-22 26-27 27",
        }:  "1-2 6-7 21-22 26-27 27",


        /*
        p{
            "",
            "",
        }:  "",
        */


    }

    for testPair, answerStr := range testMap {
        resultStr := getIntersectionString(testPair[0], testPair[1])
        if resultStr != answerStr {
            Println("test failed:")
            Println("resultStr =", resultStr)
            Println("answerStr =", answerStr)
        }
    }

    if len(os.Args) > 2 {
        list1Str := os.Args[1]
        list2Str := os.Args[2]
        resultStr := getIntersectionString(list1Str, list2Str)
        Println("")
        Println(" First:", list1Str)
        Println("Second:", list2Str)
        Println("Result:", resultStr)
    }
}

func main() {
    //Println(len(CalTypesMap))
    //testTimeZone()
    //testGetJdListFromEpochRange()
    //testGetJhmsByEpoch()
    //testHMS_FloatHour()
    //testDecodeHMS()
    //testToJd()
    //testIntervalListByNumList()
    //testIntervalListNormalize()
    testIntervalListIntersection()
}

