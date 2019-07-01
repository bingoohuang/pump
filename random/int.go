package random

import (
	"math/rand"
	"reflect"

	"github.com/bingoohuang/pump/model"
)

type Int struct {
	mask      int64
	allowNull bool
}

var _ ColumnRandomizer = (*Int)(nil)

func IntZero() reflect.Type {
	return reflect.TypeOf(int64(1))
}

func (r Int) Value() interface{} {
	return rand.Int63n(r.mask)
}

func NewRandomInt(col model.TableColumn, mask int64) *Int {
	return &Int{mask: mask, allowNull: col.IsAllowNull()}
}
