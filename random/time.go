package random

import (
	"fmt"
	"math/rand"
	"reflect"

	"github.com/bingoohuang/pump/model"
)

// Time ...
type Time struct {
	allowNull bool
}

// TimeZero ...
func TimeZero() reflect.Type {
	return reflect.TypeOf("")
}

// Value ...
func (r *Time) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < model.NilFrequency {
		return nil
	}

	h := rand.Int63n(24)
	m := rand.Int63n(60)
	s := rand.Int63n(60)

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// NewRandomTime ...
func NewRandomTime(column model.TableColumn) *Time {
	return &Time{allowNull: column.IsAllowNull()}
}
