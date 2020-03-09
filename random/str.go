package random

import (
	"math/rand"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/pump/model"
	"github.com/gdamore/encoding"
)

// Str ...
type Str struct {
	maxSize      int
	allowNull    bool
	rr           *RuneRandom
	characterSet string
}

// StrZero ...
func StrZero() reflect.Type {
	return reflect.TypeOf("")
}

// Value ...
// nolint gomnd
func (r *Str) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < model.NilFrequency {
		return nil
	}

	maxSize := uint64(r.maxSize)

	if maxSize == 0 {
		maxSize = uint64(rand.Int63n(100))
	}

	s := r.rr.Rune(int(maxSize))

	var err error

	if FoldContains(r.characterSet, "latin1") {
		latin1Encoder := encoding.ISO8859_1.NewEncoder()
		s, err = latin1Encoder.String(s)
		if err != nil {
			logrus.Panicf("failed to encode to latin1, error %v", err)
		}
	}

	bytes := []byte(s)
	if len(bytes) <= r.maxSize {
		return s
	}

	return string(bytes[0:r.maxSize])
}

// FoldContains tells if s contains sub in fold mode.
func FoldContains(s, sub string) bool {
	us := strings.ToLower(s)
	usub := strings.ToLower(sub)

	return strings.Contains(us, usub)
}

// NewRandomStr ...
func NewRandomStr(column model.TableColumn) *Str {
	return &Str{
		maxSize:      column.GetMaxSize(),
		allowNull:    column.IsNullable(),
		rr:           MakeRuneRandom(),
		characterSet: column.GetCharacterSet(),
	}
}
