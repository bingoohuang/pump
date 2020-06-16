package random

import (
	"math/rand"
	"reflect"

	"github.com/bingoohuang/pump/model"
)

// Int ...
type Int struct {
	mask      int64
	allowNull bool
}

var _ model.Randomizer = (*Int)(nil)

// IntZero ...
func IntZero() reflect.Type {
	return reflect.TypeOf(int64(1))
}

// Value ...
func (r Int) Value() interface{} {
	return rand.Int63n(r.mask)
}

// NewRandomInt ...
func NewRandomInt(col model.TableColumn, mask int64) *Int {
	return &Int{mask: mask, allowNull: col.IsNullable()}
}
