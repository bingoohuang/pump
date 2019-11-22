package random

// Const ...
type Const struct {
	con interface{}
}

// NewConst ...
func NewConst(con interface{}) *Const {
	return &Const{con: con}
}

// Value ...
func (r *Const) Value() interface{} {
	return r.con
}
