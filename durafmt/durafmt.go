// Package durafmt formats time.Duration into a human readable format.
package durafmt

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Durafmt holds the parsed duration and the original input duration.
type Durafmt struct {
	duration time.Duration
	input    string // Used as reference.
}

// Parse creates a new *Durafmt struct, returns error if input is invalid.
func Format(dinput time.Duration) string {
	input := dinput.String()
	return Durafmt{dinput, input}.format()
}

// ParseString creates a new *Durafmt struct from a string.
// returns an error if input is invalid.
func FormatString(input string) (string, error) {
	if input == "0" || input == "-0" {
		return "", errors.New("durafmt: missing unit in duration " + input)
	}
	duration, err := time.ParseDuration(input)
	if err != nil {
		return "", err
	}
	return Durafmt{duration, input}.format(), nil
}

// format parses d *Durafmt into a human readable duration.
func (d Durafmt) format() string {
	var duration string

	// Check for minus durations.
	if string(d.input[0]) == "-" {
		duration += "-"
		d.duration = -d.duration
	}

	// Convert duration.
	seconds := int64(d.duration.Seconds()) % 60
	minutes := int64(d.duration.Minutes()) % 60
	hours := int64(d.duration.Hours()) % 24
	days := int64(d.duration/(24*time.Hour)) % 365 % 7

	// Edge case between 364 and 365 days.
	// We need to calculate weeks from what is left from years
	leftYearDays := int64(d.duration/(24*time.Hour)) % 365
	weeks := leftYearDays / 7
	if leftYearDays >= 364 && leftYearDays < 365 {
		weeks = 52
	}

	years := int64(d.duration/(24*time.Hour)) / 365
	milliseconds := int64(d.duration/time.Millisecond) -
		(seconds * 1000) - (minutes * 60000) - (hours * 3600000) -
		(days * 86400000) - (weeks * 604800000) - (years * 31536000000)

	// Create a map of the converted duration time.
	durationMap := map[string]int64{
		"ms": milliseconds,
		"s":  seconds,
		"m":  minutes,
		"h":  hours,
		"d":  days,
		"w":  weeks,
		"y":  years,
	}

	// Construct duration string.
	for _, u := range []string{"y", "w", "d", "h", "m", "s", "ms"} {
		v := durationMap[u]
		strval := strconv.FormatInt(v, 10)
		switch {
		// add to the duration string if v > 1.
		case v > 1:
			duration += strval + u
		// remove the plural 's', if v is 1.
		case v == 1:
			duration += strval + strings.TrimRight(u, "s")
		// omit any value with 0s or 0.
		case d.duration.String() == "0" || d.duration.String() == "0s":
			// note: milliseconds and minutes have the same suffix (m)
			// so we have to check if the units match with the suffix.

			// check for a suffix that is NOT the milliseconds suffix.
			if strings.HasSuffix(d.input, string(u[0])) && !strings.Contains(d.input, "ms") {
				// if it happens that the units are milliseconds, skip.
				if u == "ms" {
					continue
				}
				duration += strval + u
			}
			// process milliseconds here.
			if u == "ms" {
				if strings.Contains(d.input, "ms") {
					duration += strval + u
					break
				}
			}
		// omit any value with 0.
		case v == 0:
			continue
		}
	}
	// trim any remaining spaces.
	return strings.TrimSpace(duration)
}
