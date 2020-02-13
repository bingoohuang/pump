package random

import (
	"math"
	"math/rand"
	"reflect"

	"github.com/bingoohuang/pump/model"
)

// Decimal holds unexported data for decimal values
type Decimal struct {
	size      int64
	allowNull bool
}

// DecimalZero ...
func DecimalZero() reflect.Type {
	return reflect.TypeOf(0.0) // nolint gomnd
}

// Value ...
func (r *Decimal) Value() interface{} {
	return rand.Float64() * float64(rand.Int63n(int64(math.Pow10(int(r.size)))))
}

// NewRandomDecimal ...
func NewRandomDecimal(column model.TableColumn, size int64) *Decimal {
	return &Decimal{size: size, allowNull: column.IsNullable()}
}
