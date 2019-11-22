package random

// Fn ...
type Fn struct {
	fn func() interface{}
}

// NewFn ...
func NewFn(fn func() interface{}) *Fn {
	return &Fn{fn: fn}
}

// Value ...
func (r *Fn) Value() interface{} {
	return r.fn()
}
