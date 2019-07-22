package random

type Fn struct {
	fn func() interface{}
}

func NewFn(fn func() interface{}) *Fn {
	return &Fn{fn: fn}
}

// Value returns a random time.Time in the range specified by the New method
func (r *Fn) Value() interface{} {
	return r.fn()
}