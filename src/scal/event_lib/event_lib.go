// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program. If not, see <http://www.gnu.org/licenses/gpl.txt>.
// Also avalable in /usr/share/common-licenses/GPL on Debian systems
// or /usr/share/licenses/common/GPL3/license.txt on ArchLinux

// "timeRange" -> "intervalList" FIXME

import "os"

import . "scal/lib/mapset"
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


type Occurrence interface {
    GetName() string
    GetEvent() Event
    //String() string
    //Empty() bool
    Len() int
    GetStartJd() int
    GetEndJd() int
    Intersection(Occurrence) (Occurrence, error)
    GetDaysJdList() []int
    GetEpochIntervalList() IntervalList
    //GetJdFloatIntervalList() []FloatInterval // FIXME
}

type JdSetOccurrence struct {
    Event Event
    JdSet Set
}
func (self JdSetOccurrence) GetName() string {return "jdSet"}
func (self JdSetOccurrence) GetEvent() Event {return self.Event}
func (self JdSetOccurrence) Len() int {
    return self.JdSet.Cardinality()
}
func (self JdSetOccurrence) GetStartJd() int {
    jds := self.JdSet.ToSlice()
    start := jds[0]
    for _, jd := range jds {
        if jd < start {
            start = jd
        }
    }
    return start
}
func (self JdSetOccurrence) GetEndJd() int {
    jds := self.JdSet.ToSlice()
    end := jds[0]
    for _, jd := range jds {
        if jd > end {
            end = jd
        }
    }
    return end
}
func (self JdSetOccurrence) Intersection(other Occurrence) (Occurrence, error) {
    if other.GetName() == "jdSet" {
        return JdSetOccurrence{
            self.GetEvent(),
            self.JdSet.Intersect(other.(JdSetOccurrence).JdSet),
        }, nil
    } else {
        /*
        if other.GetName() != "intervalList" {
            return JdSetOccurrence{}, errors.New(fmp.Sprintf(
                "invalid Occurrence name '%v'",
                other.GetName(),
            ))
        }
        */
        list, err := self.GetEpochIntervalList().Intersection(other.GetEpochIntervalList())
        return IntervalListOccurrence{self.GetEvent(), list}, err
    }
}
func (self JdSetOccurrence) GetDaysJdList int[] {
    return self.JdSet.ToSlice()
}
func (self JdSetOccurrence) GetEpochIntervalList IntervalList {
    loc := self.GetEvent().GetLoc()
    self.JdSet.RLock()
    list := make(IntervalList, 0, self.Len())
    for jd := range self.JdSet.s {
        list = append(list, IntervalByJd(jd, loc))
    }
    self.JdSet.RUnlock()
    return list
}



JdSetOccurrence_CalcJdIntervalList(occur Occurrence) IntervalList {
    // occur is JdSetOccurrence
    // FIXME
}

type IntervalListOccurrence struct {
    Event Event
    List IntervalList
}
func (self IntervalListOccurrence) GetName() string {return "intervalList"}
func (self IntervalListOccurrence) GetEvent() Event {return self.Event}
func (self IntervalListOccurrence) Len() int {
    return len(self.List)
}
func (self IntervalListOccurrence) GetStartJd() int {
    loc := self.GetEvent().GetLoc()
    startEpoch := self.List[0].Start
    for _, interval := range self.List {
        if interval.Start < startEpoch {
            startEpoch = interval.Start
        }
    }
    return GetJdByEpoch(startEpoch, loc)
}
func (self IntervalListOccurrence) GetEndJd() int {
    loc := self.GetEvent().GetLoc()
    endEpoch := self.List[0].End
    for _, interval := range self.List {
        if interval.End > endEpoch {
            endEpoch = interval.End
        }
    }
    return GetJdByEpoch(endEpoch, loc)
}
func (self IntervalListOccurrence) Intersection(other Occurrence) (Occurrence, error) {
    list, err := self.List.Intersection(other.GetEpochIntervalList())
    return IntervalListOccurrence{self.GetEvent(), list}, err
}
func (self IntervalListOccurrence) GetDaysJdList() []int {
    loc := self.GetEvent().GetLoc()
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
        if interval.EndClosed {
            tmpEndJd = GetJdByEpoch(interval.End, loc)
        } else {
            tmpEndJd = GetJdByEpoch(interval.End - 1, loc)
        }
        for tmpJd := tmpStartJd ; tmpJd <= tmpEndJd ; tmpJd ++ {
            jdSet.Add(tmpJd)
        }
    }
    return jdSet.ToSlice()
}
func (self IntervalListOccurrence) GetEpochIntervalList() IntervalList {
    return self.List
}













