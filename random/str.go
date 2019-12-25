package random

import (
	"math/rand"
	"reflect"

	"github.com/bingoohuang/pump/model"
)

// Str ...
type Str struct {
	maxSize   int64
	allowNull bool
	rr        *RuneRandom
}

// StrZero ...
func StrZero() reflect.Type {
	return reflect.TypeOf("")
}

// Value ...
func (r *Str) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < model.NilFrequency {
		return nil
	}

	maxSize := uint64(r.maxSize)

	if maxSize == 0 {
		maxSize = uint64(rand.Int63n(100))
	}

	return r.rr.Rune(int(maxSize))
}

// NewRandomStr ...
func NewRandomStr(column model.TableColumn) *Str {
	return &Str{maxSize: column.GetMaxSize().Int64, allowNull: column.IsNullable(), rr: MakeRuneRandom()}
}
