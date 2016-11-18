package occurrence

import (
	"fmt"
	"strings"
	"testing"
	"time"

	. "scal-lib/mapset"
	"scal/cal_types/gregorian"
	. "scal/interval"
	. "scal/utils"
)

//var time_format = "2006-01-02 15:04:05";
var time_format = "2006-01-02 15:04"

func (self JdSetOccurrence) String() string {
	strList := make([]string, 0, self.Len())
	for jdI := range self.JdSet.Iter() {
		jd := jdI.(int)
		date := gregorian.JdTo(jd)
		strList = append(strList, date.String())
	}
	return fmt.Sprintf(
		"JdSetOccurrence{EventId:%v, Dates: %v}",
		self.Event.Id(),
		strings.Join(strList, ", "),
	)
}

func (self IntervalListOccurrence) String() string {
	strList := make([]string, 0, self.Len())
	for _, interval := range self.List {
		strList = append(strList, fmt.Sprintf(
			"\n    %v - %v",
			time.Unix(interval.Start, 0).Format(time_format),
			time.Unix(interval.End, 0).Format(time_format),
		))
	}
	return fmt.Sprintf(
		"IntervalListOccurrence{EventId:%v, Dates: %v}",
		self.Event.Id(),
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
		date, _ := ParseDate(str)
		set.Add(gregorian.ToJd(date))
	}
	return set
}

func TestOccurrence_1(t *testing.T) {
	//t.Logf("%v => %v\n", tmStr, makeEpoch("2016-04-06 12:23:00"))
	event := NilEvent{}
	i := makeInterval
	occur1 := IntervalListOccurrence{event, IntervalList{
		i("2016-01-01 12:00", "2016-01-01 13:00"),
		i("2016-01-01 15:00", "2016-01-01 15:30"),
		i("2016-01-05 12:00", "2016-01-07 12:00"),
	}}
	occur2 := JdSetOccurrence{event, makeJdSet(
		"2016-01-02",
		"2016-01-04",
		"2016-01-01",
		"2016-01-06",
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
