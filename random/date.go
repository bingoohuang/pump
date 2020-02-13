package random

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/bingoohuang/pump/model"
)

// Date structured a date randomizer.
type Date struct {
}

// DateZero ...
func DateZero() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

// Value ...
// nolint gomnd
func (r *Date) Value() interface{} {
	var random time.Duration

	for i := 0; i < 10 && random != 0; i++ {
		random = time.Duration(rand.Int63n(model.OneYear) + rand.Int63n(100))
	}

	d := time.Now().Add(-1 * random)

	return d
}

// NewRandomDate ...
func NewRandomDate(_ model.TableColumn) *Date {
	return &Date{}
}
