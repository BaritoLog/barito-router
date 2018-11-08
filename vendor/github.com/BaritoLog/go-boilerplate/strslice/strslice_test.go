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

func TestEqual(t *testing.T) {
	tts := []struct {
		slice1 []string
		slice2 []string
		result bool
	}{
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{}, false},
		{[]string{"asdf"}, []string{"asdf"}, true},
		{[]string{"asdf", "qwer"}, []string{"asdf", "qwer"}, true},
		{[]string{"asdf", "qwex"}, []string{"asdf", "qwer"}, false},
	}

	for _, tt := range tts {
		get := Equal(tt.slice1, tt.slice2)
		if get != tt.result {
			t.Fatalf("Slice1=%v\tSlice2=%v\tGet '%t' while expect '%t'", tt.slice1, tt.slice2, get, tt.result)
		}
	}

}
