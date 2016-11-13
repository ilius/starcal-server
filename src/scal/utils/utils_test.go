package utils

import (
    "testing"
    "time"
    "math/rand"

    "scal"
    "scal/cal_types/gregorian"
)

func TestTimeZone(t *testing.T) {
    //t := time.Now()// .In(loc)
    //loc := tm.Location()
    //loc, err := time.LoadLocation("Asia/Tehran")
    loc, err := time.LoadLocation("UTC")
    if err != nil {t.Error(err)}
    t.Log("Location:", loc)
    //date := CalTypesMap["gregorian"].JdTo(jd)
    date, _ := GetCurrentDate("gregorian")
    t.Log("Date:", date)
    startJd := gregorian.ToJd(scal.Date{2016, 1, 1})
    endJd := gregorian.ToJd(scal.Date{2020, 1, 1})
    for jd:=startJd; jd < endJd; jd++ {
        epoch1 := GetEpochByJd(jd, loc)
        gdate := gregorian.JdTo(jd)
        tm := time.Date(
            gdate.Year,
            time.Month(gdate.Month), // date.Month is int
            gdate.Day,
            0, // hour
            0, // min
            0, // sec
            0, // nsec
            loc, // location
        )
        epoch2 := tm.Unix()
        if epoch1 == epoch2 {
            //t.Log("Epoch OK")
        } else {
            t.Error("EPOCH MISATCH, delta =", epoch2-epoch1, "  , gdate =", gdate)
        }
        floatJd := GetFloatJdByEpoch(epoch1, loc)
        //t.Log("floatJd = %f\n", floatJd)
        dayPortion := floatJd - float64(int(floatJd))
        if dayPortion > 0 {
            t.Logf("Warning: gdate=%v, dayPortion = %f\n", gdate, dayPortion)
            floatJd2 := GetFloatJdByEpoch(epoch1-1, loc)
            dayPortion2 := floatJd2 - float64(int(floatJd2))
            t.Logf("Warning: gdate=%v, dayPortion2 = %f\n\n", gdate, dayPortion2)
        }
    }
}


func TestGetJdListFromEpochRange(t *testing.T) {
    rand.Seed(time.Now().UnixNano())
    randSec1 := int64(rand.Int() % (24*3600))
    rand.Seed(time.Now().UnixNano())
    randSec2 := int64(rand.Int() % (24*3600))
    //tm := time.Now()
    //loc := tm.Location()
    loc, err := time.LoadLocation("Asia/Tehran")
    if err != nil {
        t.Error(err)
    }
    startDate := scal.Date{2015, 03, 12}
    endDate := scal.Date{2016, 07, 02}
    startEpoch := GetEpochByGDate(startDate, loc) + randSec1
    endEpoch := GetEpochByGDate(endDate, loc) + randSec2
    startJd, endJd := GetJdRangeFromEpochRange(startEpoch, endEpoch, loc)
    startJdDelta := startJd - gregorian.ToJd(startDate)
    endJdDelta := endJd - (gregorian.ToJd(endDate) + 1)
    if startJdDelta != 0 {
        t.Error("non-zero: startJdDelta: ", startJdDelta)
    }
    if endJdDelta != 0 {
        t.Error("non-zero: endJdDelta =", endJdDelta)
    }
}

func TestGetJhmsByEpoch(t *testing.T) {
    tm := time.Now()
    epoch := tm.Unix()
    loc := tm.Location()
    jd, hms := GetJhmsByEpoch(epoch, loc)
    //t.Log("GetJhmsByEpoch =", jd, hms)
    epoch2 := GetEpochByJhms(jd, hms, loc)
    //t.Log("epoch2 - epoch =", epoch2 - epoch)
    if epoch2 != epoch {
        t.Errorf("%v != %v", epoch2, epoch)
    }
    jd2, _ := GetJdAndSecondsFromEpoch(epoch, loc)
    if jd2 != jd {
        t.Errorf("%v != %v", jd2, jd)
    }
    //t.Log(jd2, sec)
}


