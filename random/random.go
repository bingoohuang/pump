package random

import (
	"math/rand"
	"time"
)

type RuneRandom struct {
	stdChars []rune
	numbers  []rune
}

func MakeRuneRandom() *RuneRandom {
	return &RuneRandom{
		stdChars: []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"),
		numbers:  []rune("一二三四五六七八九十1234567890"),
	}
}

func (rr *RuneRandom) Rune(size int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randRune := make([]rune, size)
	stdCharsLen := int32(len(rr.stdChars))
	numbersLen := int32(len(rr.numbers))

	for i := range randRune {
		j := r.Int31n(3)
		switch j {
		case 0:
			randRune[i] = rr.stdChars[r.Int31n(stdCharsLen)]
		case 1:
			randRune[i] = rr.numbers[r.Int31n(numbersLen)]
		default:
			randRune[i] = RandInt(r, 19968, 40869)
		}
	}
	return string(randRune)
}

func RandInt(r *rand.Rand, min, max int32) int32 {
	return min + r.Int31n(max-min)
}
