package random

import (
	"math/rand"
	"time"
	"unicode/utf8"
)

// RuneRandom ...
type RuneRandom struct {
	stdChars []rune
	numbers  []rune
}

// MakeRuneRandom ...
func MakeRuneRandom() *RuneRandom {
	return &RuneRandom{
		stdChars: []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"),
		numbers:  []rune("一二三四五六七八九十1234567890"),
	}
}

// Rune ...
// nolint gomnd
func (rr *RuneRandom) Rune(maxSize int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randRune := make([]rune, maxSize)
	stdCharsLen := int32(len(rr.stdChars))
	numbersLen := int32(len(rr.numbers))

	bytesLen := 0

	for i := range randRune {
		j := r.Int31n(3)

		var ru rune

		switch j {
		case 0:
			ru = rr.stdChars[r.Int31n(stdCharsLen)]
		case 1:
			ru = rr.numbers[r.Int31n(numbersLen)]
		default:
			ru = RandInt(r, 19968, 40869)
		}

		if bytesLen += utf8.RuneLen(ru); bytesLen >= maxSize {
			randRune = randRune[0:i]
			break
		}

		randRune[i] = ru
	}

	return string(randRune)
}

// RandInt ...
func RandInt(r *rand.Rand, min, max int32) int32 {
	return min + r.Int31n(max-min)
}
