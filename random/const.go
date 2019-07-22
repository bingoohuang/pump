package random

type Const struct {
	con interface{}
}

func NewConst(con interface{}) *Const {
	return &Const{con: con}
}

// Value returns a random time.Time in the range specified by the New method
func (r *Const) Value() interface{} {
	return r.con
}