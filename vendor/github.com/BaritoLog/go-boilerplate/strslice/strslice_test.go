package strslice

import "testing"

func TestContain(t *testing.T) {

	testcases := []struct {
		slice    []string
		s        string
		expected bool
	}{
		{[]string{"abcd", "fghij", "klmn"}, "abcd", true},
		{[]string{"abcd", "fghij", "klmn"}, "fghij", true},
		{[]string{"abcd", "fghij", "klmn"}, "klmn", true},
		{[]string{"abcd", "fghij", "klmn"}, "xyz", false},
	}

	for _, tt := range testcases {
		get := Contain(tt.slice, tt.s)
		if get != tt.expected {
			t.Fatalf("Slice=%v\tString=%s\tGet '%t' while expect '%t'", tt.slice, tt.s, get, tt.expected)
		}
	}

}
