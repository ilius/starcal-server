package interval

import (
    "testing"
    "time"
    "os"
    //"strings"
    "math/rand"
)

func ShuffleIntervals(a []Interval) {
    rand.Seed(time.Now().UnixNano())
    for i := range a {
        j := rand.Intn(i + 1)
        a[i], a[j] = a[j], a[i]
    }
}

func TestIntervalListByNumList(t *testing.T) {
    nums := []int64{1, 2, 3, 4, 5, 7, 9, 10, 14, 16, 17, 18, 19, 21, 22, 23, 24}
    t.Log(nums)
    t.Log(IntervalListByNumList(nums, 3))
}

func TestIntervalListNormalize(t *testing.T) {
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
        //t.Log(testStr, "=>", answerStr)
        testList, testErr := ParseIntervalList(testStr)
        if testErr != nil {
            t.Log(testErr)
            continue
        }
        answerList, answerErr := ParseIntervalList(answerStr)
        if answerErr != nil {
            t.Log(answerErr)
            continue
        }
        ShuffleIntervals(testList) // FIXME
        testList, testErr = testList.Normalize()
        if testErr != nil {
            t.Log(testErr)
            continue
        }
        testList = testList.Humanize()
        if testList.String() != answerList.String() {// FIXME
            t.Error("test failed, result doesn't match the answer")
            t.Error("  testList =", testList)
            t.Error("answerList =", answerList)
        }
        t.Log(testStr, "=>", answerList)
        succeedCount ++
    }
    failedCount := testCount - succeedCount
    t.Logf("%d tests out of %d succeeded, %d failed\n", succeedCount, testCount, failedCount)

    t.Log("len(os.Args) =", len(os.Args))
    /*
    argc := len(os.Args)
    if argc > 1 {
        listStr := strings.Join(os.Args[1:argc], " ")
        //listStr := 
        list, err := ParseIntervalList(listStr)
        if err != nil {
            t.Error(err)
            return
        }
        t.Log("Original Intervals:", list)
        ShuffleIntervals(list) // FIXME
        t.Log("Shuffled Intervals:", list)
        list, err = list.Normalize()
        if err != nil {
            t.Error(err)
            return
        }
        t.Log("    New Intervals:", list)
        //t.Logf("    New Intervals: %v\n", list)
    }*/
}

func getIntersectionString(t *testing.T, list1Str string, list2Str string) string {
    list1, err1 := ParseIntervalList(list1Str)
    list2, err2 := ParseIntervalList(list2Str)
    if err1 != nil {
        t.Error(err1)
        return ""
    }
    if err2 != nil {
        t.Error(err2)
        return ""
    }
    ShuffleIntervals(list1)
    ShuffleIntervals(list2)

    result, err := list1.Intersection(list2)
    if err != nil {
        t.Error(err)
        return ""
    }
    result = result.Humanize()
    return result.String()
}

func TestIntervalListIntersection(t *testing.T) {
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
        resultStr := getIntersectionString(t, testPair[0], testPair[1])
        if resultStr != answerStr {
            t.Error("test failed:")
            t.Error("resultStr =", resultStr)
            t.Error("answerStr =", answerStr)
        }
    }

    if len(os.Args) > 2 {
        list1Str := os.Args[1]
        list2Str := os.Args[2]
        resultStr := getIntersectionString(t, list1Str, list2Str)
        t.Log("")
        t.Log(" First:", list1Str)
        t.Log("Second:", list2Str)
        t.Log("Result:", resultStr)
    }
}

