package scal

type M = map[string]any

type MErr struct {
	M
	Err error
}

type D []DItem

type DItem struct {
	Name  string
	Value any
}

func (d D) Map() (m M) {
	m = make(M, len(d))
	for _, item := range d {
		m[item.Name] = item.Value
	}
	return m
}
