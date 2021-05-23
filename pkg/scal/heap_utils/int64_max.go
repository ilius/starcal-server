package heap_utils

// An Int64MaxHeap is a max-heap of int64 elements
type Int64MaxHeap []int64

func (h Int64MaxHeap) Len() int           { return len(h) }
func (h Int64MaxHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h Int64MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *Int64MaxHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int64))
}

func (h *Int64MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
