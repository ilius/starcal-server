package interval

import (
    "time"
    "fmt"
    "errors"
    "strings"
    "strconv"
    "sort"
    //"container/heap"
)

import . "scal/utils"
//import . "scal/heap_utils"
import . "scal/lib/stack"


type FloatInterval struct {
    Start float64
    End float64
}


type Interval struct {
    Start int64
    End int64
    ClosedEnd bool
}
func (interval Interval) String() string {
    if interval.Start == interval.End {
        if interval.ClosedEnd {
            return fmt.Sprintf("%d", interval.Start)
        } else {
            return "()"
        }
    } else {
        if interval.ClosedEnd {
            return fmt.Sprintf("[%d-%d]", interval.Start, interval.End)
        } else {
            return fmt.Sprintf("[%d-%d)", interval.Start, interval.End)
        }
    }
}
func IntervalByJd (jd int, loc *time.Location) Interval {
    return Interval{
        GetEpochByJd(jd, loc),
        GetEpochByJd(jd+1, loc),
        false,
    }
}
func ParseInterval(str string) (Interval, error) {
    parts := strings.Split(str, "-")
    if len(parts) > 2 {
        return Interval{}, errors.New("invalid Interval string '" + str + "'")
    }
    start, startErr := strconv.ParseInt(parts[0], 10, 0)
    if startErr != nil {
        return Interval{}, startErr
    }
    var end int64
    if len(parts) == 1 {
        end = start
    } else {
        var endErr error
        end, endErr = strconv.ParseInt(parts[1], 10, 0)
        if endErr != nil {
            return Interval{}, endErr
        }
        if end < start {
            return Interval{}, errors.New("invalid interval: end < start")
        }
    }
    return Interval{int64(start), int64(end), start==end}, nil
}

type IntervalPoint struct {
    Pos int64
    IsEnd bool
    Closed bool
    ListId int
    // ListId is index of IntervalList that interval belongs to
    // used for `intersection` only, otherwise set to zero
}
func (p IntervalPoint) String() string {
    var format string
    if p.IsEnd {
        if p.Closed {
            format = "%v%v]"
        } else {
            format = "%v%v)"
        }
    } else {
        if p.Closed {
            format = "[%v%v"
        } else {
            format = "(%v%v"
        }
    }
    /*
    return fmt.Sprintf(
        "\n    (Pos=%v, IsEnd=%v, Closed=%v, ListId=%v)",
        p.Pos,
        p.IsEnd,
        p.Closed,
        p.ListId,
    )*/
    symbol := string('A'+byte(p.ListId))
    return fmt.Sprintf(" " + format + " ", symbol, p.Pos)
}

type IntervalPointList []IntervalPoint
func (p IntervalPointList) Len() int { return len(p) }
func (p IntervalPointList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p IntervalPointList) Less(i, j int) bool {
    a := p[i]
    b := p[j]
    if a.Pos != b.Pos {
        return a.Pos < b.Pos
    } else if a.IsEnd != b.IsEnd {
        // Start always come before End
        return b.IsEnd
    } else if a.Closed != b.Closed {
        // ClosedStart < OpenStart < OpenEnd < ClosedEnd
        if a.IsEnd {// && b.IsEnd
            return b.Closed // && !a.Closed
        } else{
            return a.Closed // && !b.Closed
        }
    } else if a.ListId != b.ListId {
        return a.ListId < b.ListId
    }
    return false
}
func (a IntervalPointList) Sort() {
    sort.Sort(a)
}
func (points IntervalPointList) GetIntervalList() (IntervalList, error) {
    pcount := len(points)
    list := make(IntervalList, 0, pcount/2)// we need no more than `count` spaces
    startedStack := make(Int64Stack, 0, pcount)
    var start int64
    for _, point := range points {
        if point.IsEnd {
            if len(startedStack)==0 {
                return IntervalList{}, errors.New(fmt.Sprintf(
                    "point='%v', startedStack=[]\n",
                    point,
                ))
            }
            startedStack, start = startedStack.Pop()
            //fmt.Println("pop:", start, ", new len:", len(startedStack))
            if len(startedStack)==0 {
                list = append(list, Interval{start, point.Pos, point.Closed})
                // We will replace closed ends (with 2 intervals) after the final operation (intersection)
                // By: list = list.Humanize()
                // If we do it here, something breaks, because it's not mathematical,
                // and we won't have a fully normalized IntervalList
            }
        } else {
            //fmt.Println("push:", point.Pos)
            startedStack = startedStack.Push(point.Pos)
        }
    }
    return list, nil
}



type IntervalList []Interval
func (list IntervalList) String() string {
    parts := make([]string, 0, len(list))
    for _, interval := range list {
        var part string
        if interval.ClosedEnd {
            if interval.Start == interval.End {
                part = fmt.Sprintf("%d", interval.Start)
            } else {
                part = fmt.Sprintf("[%d-%d]", interval.Start, interval.End)
            }
        } else {
            if interval.End > interval.Start {
                part = fmt.Sprintf("%d-%d", interval.Start, interval.End)
            } else {
                part = "None"
            }
        }
        parts = append(parts, part)
    }
    return strings.Join(parts, " ")
}
func (list IntervalList) GetPointList(listId int) IntervalPointList {
    count := len(list)
    points := make(IntervalPointList, 2*count)// we need exactly `2*count` spaces
    for ii, interval := range list {
        // if interval.Start > interval.End // what? FIXME
        points[2 * ii] = IntervalPoint{
            interval.Start,
            false,// IsEnd
            true,// Closed
            listId,
        }
        points[2 * ii + 1] = IntervalPoint{
            interval.End,
            true,// IsEnd
            interval.ClosedEnd,// Closed
            listId,
        }
    }
    return points
}
func (list IntervalList) Humanize() IntervalList {
    /*
    Replace Closed Ends
    Replace all [a, b] intervals with two new intervals: [a, b) and [b, b]
    */
    closedEndCount := 0
    for _, interval := range list {
        if interval.ClosedEnd && interval.End > interval.Start {
            closedEndCount ++
        }
    }
    if closedEndCount == 0 {
        return list
    }
    newLen := len(list) + closedEndCount
    //if cap(list) >= newLen
    // we need to insert to slice, can't do it in-place
    newList := make(IntervalList, 0, newLen)
    for _, interval := range list {
        if interval.ClosedEnd && interval.End > interval.Start {
            newList = append(newList, Interval{interval.Start, interval.End, false})
            newList = append(newList, Interval{interval.End, interval.End, true})
        } else {
            newList = append(newList, interval)
        }
    }
    return newList
}


func ParseIntervalList(str string) (IntervalList, error) {
    parts := strings.Split(str, " ")
    count := len(parts)
    list := make(IntervalList, 0, count)
    var interval Interval
    var err error
    for _, intervalStr := range parts {
        interval, err = ParseInterval(intervalStr)
        if err != nil {
            return list, err
        }
        list = append(list, interval)
    }
    return list, nil
}

func (list IntervalList) Normalize() (IntervalList, error) {
    points := list.GetPointList(0)
    points.Sort()
    return points.GetIntervalList()
}

func (list IntervalList) Intersection(list2 IntervalList) (IntervalList, error) {
    return IntersectionOfSomeIntervalLists(list, list2)
}
func IntersectionOfSomeIntervalLists(lists ...IntervalList) (IntervalList, error) {
    var err error
    listCount := len(lists)
    // assert listCount == 2
    intervalCount := 0
    for listId, list := range lists {
        list, err = list.Normalize()
        if err != nil {
            return IntervalList{}, err
        }
        lists[listId] = list
        intervalCount += len(list)
    }
    points := make(IntervalPointList, 0, 2*intervalCount)// exactly `2*intervalCount` spaces
    for listId, list := range lists {
        for _, point := range list.GetPointList(listId) {
            points = append(points, point)
        }
    }
    points.Sort()
    result := make(IntervalList, 0, intervalCount)// smaller capacity? FIXME

    openStartList := make([]int64, listCount)
    for i:=0 ; i<listCount; i++ {
        openStartList[i] = MIN_INT64
    }
    var hasNil bool
    var start int64
    //fmt.Printf("points = %v\n\n", points)
    for _, point := range points {
        //fmt.Printf("point:    %v\n", point)
        if point.IsEnd { // end (closed or open)
            // end == point.Pos
            hasNil = false
            start = MIN_INT64
            for _, tmpStart := range openStartList {
                if tmpStart == MIN_INT64 {
                    hasNil = true
                    //break // FIXME
                }
                if tmpStart > start {
                    start = tmpStart
                }
            }
            if ! hasNil {
                if start > point.Pos {
                    return IntervalList{}, errors.New(fmt.Sprintf(
                        "Internal Error: start - point.Pos = %d",
                        start - point.Pos,
                    ))
                }
                if point.Pos > start || point.Closed {
                    //fmt.Println("adding", Interval{start, point.Pos, point.Closed}, "  point  ", point)
                    result = append(result, Interval{start, point.Pos, point.Closed})
                }
            }
            //if start == point.Pos:## FIXME
            //    print('start = point.Pos = %s, IsEnd=%s'%(start%(24*3600)/3600.0, point.IsEnd))
            openStartList[point.ListId] = MIN_INT64
            //fmt.Printf("openStartList[%v] = %v\n", point.ListId, MIN_INT64)
        } else { // start
            // start == point.Pos
            if openStartList[point.ListId] != MIN_INT64 {
                //for _, list := range lists { fmt.Println(list) }
                return IntervalList{}, errors.New(fmt.Sprintf(
                    "Internal Error: point:  %v   openStartList: %v",
                    point,
                    openStartList,
                ))
            }
            openStartList[point.ListId] = point.Pos
            //fmt.Printf("openStartList[%v] = %v\n", point.ListId, point.Pos)

        }
    }

    return result, nil
}



func IntervalListByNumList(nums []int64, minCount int) IntervalList {
    // typically minCount=3
    // nums must be sorted, minCount >= 2
    list := make(IntervalList, 0, len(nums))
    tmpNums := make([]int64, 0, len(nums))
    for _, num := range nums {
        if len(tmpNums) > 0 && num - tmpNums[len(tmpNums)-1] != 1 {
            if len(tmpNums) > minCount {
                list = append(list, Interval{tmpNums[0], tmpNums[len(tmpNums)-1], true})
            } else {
                for _, x := range tmpNums {
                    list = append(list, Interval{x, x, true})
                }
            }
            tmpNums = nil
        }
        tmpNums = append(tmpNums, num)
    }
    if len(tmpNums) > 0 {
        if len(tmpNums) > minCount {
            list = append(list, Interval{tmpNums[0], tmpNums[len(tmpNums)-1], true})
        } else {
            for _, num := range tmpNums {
                list = append(list, Interval{num, num, true})
            }
        }
    }
    return list
}

