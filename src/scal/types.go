package scal

type M map[string]interface{}

type MErr struct {
	M
	Err error
}
