package scal

type M = map[string]interface{}

type MErr struct {
	M
	Err error
}

type D []DItem

type DItem struct {
	Name  string
	Value interface{}
}

func (d D) Map() (m M) {
	m = make(M, len(d))
	for _, item := range d {
		m[item.Name] = item.Value
	}
	return m
}
