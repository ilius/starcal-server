// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along
// with this program. If not, see <https://www.gnu.org/licenses/agpl.txt>.

// "jdSet" -> "jd"
// "timeRange" -> "interval" FIXME

package occurrence

import . "scal-lib/mapset"
import . "scal/utils"
import . "scal/interval"

/*
func getEventUID(event Event) string {
    event_st = core.compressLongInt(hash(str(event.getData())))
    time_st = core.getCompactTime()
    host, _ := os.Hostname()
    return event_st + '_' + time_st + '@' + host
}
*/

type OccurSet interface {
	GetName() string
	GetEvent() Event
	//String() string
	//Empty() bool
	Len() int
	GetStartJd() int
	GetEndJd() int
	Intersection(OccurSet) (OccurSet, error)
	GetDaysJdList() []int
	GetEpochIntervalList() IntervalList
	//GetJdFloatIntervalList() []FloatInterval // FIXME
}

type JdOccurSet struct {
	Event Event
	JdSet Set
}

func (occur JdOccurSet) GetName() string { return "jd" }
func (occur JdOccurSet) GetEvent() Event { return occur.Event }
func (occur JdOccurSet) Len() int {
	return occur.JdSet.Cardinality()
}
func (occur JdOccurSet) GetStartJd() int {
	jds := occur.JdSet.ToSlice()
	start := jds[0].(int)
	for _, jdI := range jds {
		jd := jdI.(int)
		if jd < start {
			start = jd
		}
	}
	return start
}
func (occur JdOccurSet) GetEndJd() int {
	jds := occur.JdSet.ToSlice()
	end := jds[0].(int)
	for _, jdI := range jds {
		jd := jdI.(int)
		if jd > end {
			end = jd
		}
	}
	return end
}
func (occur JdOccurSet) Intersection(other OccurSet) (OccurSet, error) {
	if other.GetName() == "jd" {
		return JdOccurSet{
			occur.GetEvent(),
			occur.JdSet.Intersect(other.(JdOccurSet).JdSet),
		}, nil
	} else {
		/*
		   if other.GetName() != "interval" {
		       return JdOccurSet{}, errors.New(fmp.Sprintf(
		           "invalid OccurSet name '%v'",
		           other.GetName(),
		       ))
		   }
		*/
		list, err := occur.GetEpochIntervalList().Intersection(other.GetEpochIntervalList())
		return IntervalOccurSet{occur.GetEvent(), list}, err
	}
}
func (occur JdOccurSet) GetDaysJdList() []int {
	return IntListBySet(occur.JdSet)
}
func (occur JdOccurSet) GetEpochIntervalList() IntervalList {
	loc := occur.GetEvent().Location()
	//occur.JdSet.RLock()
	list := make(IntervalList, 0, occur.Len())
	for jdI := range occur.JdSet.Iter() {
		list = append(list, IntervalByJd(jdI.(int), loc))
	}
	//occur.JdSet.RUnlock()
	return list
}

func JdOccurSet_CalcJdIntervalList(occur OccurSet) IntervalList {
	// occur is JdOccurSet
	// FIXME
	return IntervalList{}
}

type IntervalOccurSet struct {
	Event Event
	List  IntervalList
}

func (occur IntervalOccurSet) GetName() string { return "interval" }
func (occur IntervalOccurSet) GetEvent() Event { return occur.Event }
func (occur IntervalOccurSet) Len() int {
	return len(occur.List)
}
func (occur IntervalOccurSet) GetStartJd() int {
	loc := occur.GetEvent().Location()
	startEpoch := occur.List[0].Start
	for _, interval := range occur.List {
		if interval.Start < startEpoch {
			startEpoch = interval.Start
		}
	}
	return GetJdByEpoch(startEpoch, loc)
}
func (occur IntervalOccurSet) GetEndJd() int {
	loc := occur.GetEvent().Location()
	endEpoch := occur.List[0].End
	for _, interval := range occur.List {
		if interval.End > endEpoch {
			endEpoch = interval.End
		}
	}
	return GetJdByEpoch(endEpoch, loc)
}
func (occur IntervalOccurSet) Intersection(other OccurSet) (OccurSet, error) {
	list, err := occur.List.Intersection(other.GetEpochIntervalList())
	return IntervalOccurSet{occur.GetEvent(), list}, err
}
func (occur IntervalOccurSet) GetDaysJdList() []int {
	loc := occur.GetEvent().Location()
	/*
	   inCount = len(occur.List)
	   jdCountMax := 3 + int(
	       (occur.List[inCount-1].End - occur.List[0].Start) / int64(24*3600)
	   )
	   if jdCountMax > inCount {
	       jdCountMax = inCount
	   }
	   jdList := make([]int, 0, jdCountMax)
	*/
	jdSet := NewSet()
	var tmpStartJd, tmpEndJd int
	for _, interval := range occur.List {
		tmpStartJd = GetJdByEpoch(interval.Start, loc)
		if interval.ClosedEnd {
			tmpEndJd = GetJdByEpoch(interval.End, loc)
		} else {
			tmpEndJd = GetJdByEpoch(interval.End-1, loc)
		}
		for tmpJd := tmpStartJd; tmpJd <= tmpEndJd; tmpJd++ {
			jdSet.Add(tmpJd)
		}
	}
	return IntListBySet(jdSet)
}
func (occur IntervalOccurSet) GetEpochIntervalList() IntervalList {
	return occur.List
}
