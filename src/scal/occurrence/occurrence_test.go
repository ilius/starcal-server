package occurrence

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"scal"
	. "scal-lib/mapset"
	"scal/cal_types/gregorian"
	. "scal/interval"
	. "scal/utils"
)

//const time_format = "2006-01-02 15:04:05";
const time_format = "2006-01-02 15:04"

func (occur JdOccurSet) String() string {
	strList := make([]string, 0, occur.Len())
	for jdI := range occur.JdSet.Iter() {
		jd := jdI.(int)
		date := gregorian.JdTo(jd)
		strList = append(strList, date.String())
	}
	return fmt.Sprintf(
		"JdOccurSet{EventId:%v, Dates: %v}",
		occur.Event.Id(),
		strings.Join(strList, ", "),
	)
}

func (occur IntervalOccurSet) String() string {
	strList := make([]string, 0, occur.Len())
	for _, interval := range occur.List {
		strList = append(strList, fmt.Sprintf(
			"\n    %v - %v",
			time.Unix(interval.Start, 0).Format(time_format),
			time.Unix(interval.End, 0).Format(time_format),
		))
	}
	return fmt.Sprintf(
		"IntervalOccurSet{EventId:%v, Dates: %v}",
		occur.Event.Id(),
		strings.Join(strList, ","),
	)
}

func makeEpoch(str string) int64 {
	tm, _ := time.Parse(time_format, str)
	return tm.Unix()
}

func makeInterval(startStr string, endStr string) Interval {
	return Interval{makeEpoch(startStr), makeEpoch(endStr), false}
}

func makeJdSet(strList ...string) Set {
	set := NewSet()
	for _, str := range strList {
		date, _ := scal.ParseDate(str)
		set.Add(gregorian.ToJd(date))
	}
	return set
}

func TestOccurSet_1(t *testing.T) {
	//t.Logf("%v => %v\n", tmStr, makeEpoch("2016-04-06 12:23:00"))
	event := NilEvent{}
	i := makeInterval
	occur1 := IntervalOccurSet{event, IntervalList{
		i("2016-01-01 12:00", "2016-01-01 13:00"),
		i("2016-01-01 15:00", "2016-01-01 15:30"),
		i("2016-01-05 12:00", "2016-01-07 12:00"),
	}}
	occur2 := JdOccurSet{event, makeJdSet(
		"2016/01/02",
		"2016/01/04",
		"2016/01/01",
		"2016/01/06",
	)}
	inter, err := occur1.Intersection(occur2)
	if err != nil {
		t.Error(err)
	}
	t.Log(occur1)
	t.Log("\n")
	t.Log(occur2)
	t.Log("\n")
	t.Log(inter)
}
