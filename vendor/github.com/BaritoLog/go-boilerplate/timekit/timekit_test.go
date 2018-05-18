package timekit

import (
	"testing"
	"time"
)

func TestSleep(t *testing.T) {
	testcases := []struct {
		Duration string
		seconds  int64
	}{
		{"1s", 1},
		{"2s", 2},
	}

	for _, tt := range testcases {
		start := time.Now()
		Sleep(tt.Duration)

		seconds := time.Now().Unix() - start.Unix()
		if tt.seconds != seconds {
			t.Fatalf("Sleep with duration '%s' running for '%d' seconds", tt.Duration, tt.seconds)
		}
	}
}

func TestEqualUTC(t *testing.T) {
	testcases := []struct {
		t        time.Time
		s        string
		expected bool
	}{
		{UTC("2015-11-10T23:00:00Z"), "2015-11-10T23:00:00Z", true},
		{UTC("2015-12-10T23:00:00Z"), "2015-11-10T23:00:00Z", false},
	}

	for _, tt := range testcases {
		if EqualUTC(tt.t, tt.s) != tt.expected {
			t.Fatalf("Equal %v with %s is expect to %t", tt.t, tt.s, tt.expected)
		}
	}
}

func TestEqualString(t *testing.T) {
	testcases := []struct {
		t        time.Time
		format   string
		s        string
		expected bool
	}{
		{Parse("2006-01-02T15:04:05", "2018-10-10T20:04:30"), "2006-01-02T15:04:05", "2018-10-10T20:04:30", true},
		{Parse("2006-01-02T15:04:05", "2018-10-10T20:04:30"), "2006-01-02T15:04:05", "2018-10-11T20:04:30", false},
	}

	for _, tt := range testcases {
		if EqualString(tt.t, tt.format, tt.s) != tt.expected {
			t.Fatalf("Equal %v with %s (format: %s) is expect to %t",
				tt.t, tt.s, tt.format, tt.expected)
		}
	}

}
