package utils

import "sort"

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



