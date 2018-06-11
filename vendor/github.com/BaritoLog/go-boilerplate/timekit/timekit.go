package timekit

import "time"

// Sleep pause current goroutine according duration format
func Sleep(durationFormat string) {
	sleepDuration, _ := time.ParseDuration(durationFormat)
	time.Sleep(sleepDuration)
}

// EqualUTC return whether time is same with UTC string
func EqualUTC(t1 time.Time, s string) bool {
	return t1.Equal(UTC(s))
}

// EqualString return whether time is same with string
func EqualString(t1 time.Time, format, s string) bool {
	t2, _ := time.Parse(format, s)
	return t1.Equal(t2)
}

// UTC is errorless time.Parse() using RFC3339
func UTC(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// Parse is errorless time.Parse()
func Parse(format, s string) time.Time {
	t, _ := time.Parse(format, s)
	return t
}

func Duration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d

}
