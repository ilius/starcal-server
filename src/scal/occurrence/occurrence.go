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

func (self JdOccurSet) GetName() string { return "jd" }
func (self JdOccurSet) GetEvent() Event { return self.Event }
func (self JdOccurSet) Len() int {
	return self.JdSet.Cardinality()
}
func (self JdOccurSet) GetStartJd() int {
	jds := self.JdSet.ToSlice()
	start := jds[0].(int)
	for _, jdI := range jds {
		jd := jdI.(int)
		if jd < start {
			start = jd
		}
	}
	return start
}
func (self JdOccurSet) GetEndJd() int {
	jds := self.JdSet.ToSlice()
	end := jds[0].(int)
	for _, jdI := range jds {
		jd := jdI.(int)
		if jd > end {
			end = jd
		}
	}
	return end
}
func (self JdOccurSet) Intersection(other OccurSet) (OccurSet, error) {
	if other.GetName() == "jd" {
		return JdOccurSet{
			self.GetEvent(),
			self.JdSet.Intersect(other.(JdOccurSet).JdSet),
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
		list, err := self.GetEpochIntervalList().Intersection(other.GetEpochIntervalList())
		return IntervalOccurSet{self.GetEvent(), list}, err
	}
}
func (self JdOccurSet) GetDaysJdList() []int {
	return IntListBySet(self.JdSet)
}
func (self JdOccurSet) GetEpochIntervalList() IntervalList {
	loc := self.GetEvent().Location()
	//self.JdSet.RLock()
	list := make(IntervalList, 0, self.Len())
	for jdI := range self.JdSet.Iter() {
		list = append(list, IntervalByJd(jdI.(int), loc))
	}
	//self.JdSet.RUnlock()
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

func (self IntervalOccurSet) GetName() string { return "interval" }
func (self IntervalOccurSet) GetEvent() Event { return self.Event }
func (self IntervalOccurSet) Len() int {
	return len(self.List)
}
func (self IntervalOccurSet) GetStartJd() int {
	loc := self.GetEvent().Location()
	startEpoch := self.List[0].Start
	for _, interval := range self.List {
		if interval.Start < startEpoch {
			startEpoch = interval.Start
		}
	}
	return GetJdByEpoch(startEpoch, loc)
}
func (self IntervalOccurSet) GetEndJd() int {
	loc := self.GetEvent().Location()
	endEpoch := self.List[0].End
	for _, interval := range self.List {
		if interval.End > endEpoch {
			endEpoch = interval.End
		}
	}
	return GetJdByEpoch(endEpoch, loc)
}
func (self IntervalOccurSet) Intersection(other OccurSet) (OccurSet, error) {
	list, err := self.List.Intersection(other.GetEpochIntervalList())
	return IntervalOccurSet{self.GetEvent(), list}, err
}
func (self IntervalOccurSet) GetDaysJdList() []int {
	loc := self.GetEvent().Location()
	/*
	   inCount = len(self.List)
	   jdCountMax := 3 + int(
	       (self.List[inCount-1].End - self.List[0].Start) / int64(24*3600)
	   )
	   if jdCountMax > inCount {
	       jdCountMax = inCount
	   }
	   jdList := make([]int, 0, jdCountMax)
	*/
	jdSet := NewSet()
	var tmpStartJd, tmpEndJd int
	for _, interval := range self.List {
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
func (self IntervalOccurSet) GetEpochIntervalList() IntervalList {
	return self.List
}
