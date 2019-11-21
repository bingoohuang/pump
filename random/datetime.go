package random

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/bingoohuang/pump/model"
)

// DateTimeInRange ...
type DateTimeInRange struct{}

// DateTimeInRangeZero ...
func DateTimeInRangeZero() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

// Value returns a random time.Time in the range specified by the New method
func (r *DateTimeInRange) Value() interface{} {
	rand.Seed(time.Now().UnixNano())
	randomSeconds := rand.Int63n(model.OneYear)
	d := time.Now().Add(-1 * time.Duration(randomSeconds) * time.Second)

	return d
}

// NewRandomDateTime returns a new random datetime between Now() and Now() - 1 year
func NewRandomDateTime() *DateTimeInRange {
	return &DateTimeInRange{}
}
