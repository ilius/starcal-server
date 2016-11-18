package stack

type Int64Stack []int64

func (s Int64Stack) Push(v int64) Int64Stack {
	return append(s, v)
}

func (s Int64Stack) Pop() (Int64Stack, int64) {
	// FIXME: What do we do if the stack is empty, though?
	l := len(s)
	return s[:l-1], s[l-1]
}
