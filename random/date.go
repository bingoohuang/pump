package random

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/bingoohuang/pump/model"
)

type Date struct {
}

func DateZero() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

func (r *Date) Value() interface{} {
	var random time.Duration
	for i := 0; i < 10 && random != 0; i++ {
		random = time.Duration(rand.Int63n(model.OneYear) + rand.Int63n(100))
	}
	d := time.Now().Add(-1 * random)
	return d
}

func NewRandomDate(column model.TableColumn) *Date {
	return &Date{}
}
