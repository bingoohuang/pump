package util

import "io"

func Closeq(c io.Closer) {
	_ = c.Close()
}

func Repeat(s, sep string, times int) string {
	str := ""
	for i := 0; i < times; i++ {
		if i > 0 {
			str += sep
		}

		str += s
	}

	return str
}
