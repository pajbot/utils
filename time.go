package utils

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type relTimeMagnitude struct {
	D     time.Duration
	Name  string
	DivBy time.Duration
	ModBy time.Duration
}

const (
	Day   = time.Hour * 24
	Week  = Day * 7
	Month = Day * 30
	Year  = Month * 12
)

var magnitudes = []relTimeMagnitude{
	{time.Minute, "second", time.Second, 60},
	{time.Hour, "minute", time.Minute, 60},
	{Day, "hour", time.Hour, 24},
	{Week, "day", Day, 7},
	{Month, "week", Week, 7},
	{Year, "month", Month, 12},
	{math.MaxInt64, "year", Year, -1},
}

func CustomDurationString(diff time.Duration, numParts int, glue string) string {
	if diff < time.Second {
		return "now"
	}

	n := sort.Search(len(magnitudes), func(i int) bool {
		return magnitudes[i].D > diff
	})

	if n >= len(magnitudes) {
		n--
	}

	var parts []string

	partIndex := 0
	for i := 0; partIndex < numParts && n-i >= 0; i++ {
		mag := magnitudes[n-i]

		value := diff
		if mag.DivBy != -1 {
			value /= mag.DivBy
		}
		if mag.ModBy != -1 {
			value %= mag.ModBy
		}
		if value > 0 {
			part := fmt.Sprintf("%d %s", value, mag.Name)
			if !(value == 1 || value == -1) {
				part += "s"
			}

			diff -= value * mag.DivBy

			parts = append(parts, part)
			partIndex++
		}
	}

	return strings.Join(parts, glue)
}

func CustomRelTime(t1, t2 time.Time, numParts int, glue string) string {
	var diff time.Duration

	if t1.After(t2) {
		diff = t1.Sub(t2)
	} else {
		diff = t2.Sub(t1)
	}

	return CustomDurationString(diff, numParts, glue)
}

func RelTime(t1, t2 time.Time) string {
	return CustomRelTime(t1, t2, 2, " ")
}

func TimeSince(t2 time.Time) string {
	t1 := time.Now()
	return CustomRelTime(t1, t2, 2, " ")
}

func DurationString(diff time.Duration) string {
	return CustomDurationString(diff, 2, " ")
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var unitMap = map[string]int64{
	"ns": int64(time.Nanosecond),
	"us": int64(time.Microsecond),
	"µs": int64(time.Microsecond), // U+00B5 = micro symbol
	"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": int64(time.Millisecond),
	"s":  int64(time.Second),
	"m":  int64(time.Minute),
	"h":  int64(time.Hour),
	"d":  int64(24 * time.Hour),
	"w":  int64(7 * 24 * time.Hour),
}

func consumeUnit(orig string) (unit int64, s string, err error) {
	s = orig

	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' || c >= '0' && c <= '9' {
			break
		}
	}
	if i == 0 {
		err = errors.New("time: missing unit in duration ")
		return
	}
	u := s[:i]
	s = s[i:]
	unit, ok := unitMap[u]
	if !ok {
		err = errUnknownUnit
		return
	}

	return
}

var (
	errA           = errors.New("no digits")
	errLeadingInt  = errors.New("time: bad [0-9]*")
	errUnknownUnit = errors.New("unknown unit")
)

func parseDurationPart(part string) (v, f, unit int64, scale float64, s string, err error) {
	scale = 1.0
	s = part
	// The next character must be [0-9.]
	if !(s[0] == '.' || s[0] >= '0' && s[0] <= '9') {
		err = errors.New("bad first character")
		return
	}

	// Consume leading integer [0-9]*
	pl := len(s)
	v, s, err = leadingInt(s)
	if err != nil {
		return
	}

	// Consume fraction part of string (\.[0-9]*)?
	post := false
	if s != "" && s[0] == '.' {
		s = s[1:]
		pl = len(s)
		f, scale, s = leadingFraction(s)
		post = pl != len(s)
	}
	pre := pl != len(s) // whether we consumed anything before a period
	if !pre && !post {
		// no digits (e.g. ".s" or "-.s")
		err = errA
		return
	}

	// Consume unit part of string
	unit, s, err = consumeUnit(s)
	if err != nil {
		return
	}

	if v > (1<<63-1)/unit {
		// overflow
		err = errors.New("time: invalid duration ")
		return
	}
	v *= unit

	return
}

func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d int64
	neg := false

	if s == "" {
		return 0, errors.New("time: invalid duration " + orig)
	}

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}

	for s != "" {
		var (
			v, f, unit int64
			scale      float64
			err        error
		)

		v, f, unit, scale, s, err = parseDurationPart(s)
		if err != nil {
			return 0, err
		}

		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += int64(float64(f) * (float64(unit) / scale))
			if v < 0 {
				// overflow
				return 0, errors.New("time: invalid duration " + orig)
			}
		}
		d += v
		if d < 0 {
			// overflow
			return 0, errors.New("time: invalid duration " + orig)
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}
