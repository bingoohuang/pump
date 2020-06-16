package util

import "github.com/Masterminds/goutils"

// Abbr abbreviated the s by verbose to maxWidth.
// verbose == 0 return ""
// verbose == 1 return max(s, maxWidth)
// verbose == 2 return full s
// others return "".
func Abbr(s string, verbose, maxWidth int) string {
	switch verbose {
	case 2: // nolint:gomnd
		return s
	case 1:
		v, _ := goutils.Abbreviate(s, maxWidth)
		return v
	default:
		return ""
	}
}
