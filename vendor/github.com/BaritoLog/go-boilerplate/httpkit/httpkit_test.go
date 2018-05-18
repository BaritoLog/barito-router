package httpkit

import (
	"testing"
)

func TestPathParameter(t *testing.T) {
	testcases := []struct {
		path     string
		key      string
		expected string
	}{
		{"/users/1", "users", "1"},
		{"/users/1/item/12", "users", "1"},
		{"/users/1/item/12", "shop", ""},
	}

	for _, tt := range testcases {
		get := PathParameter(tt.path, tt.key)
		if get != tt.expected {
			t.Fatalf("get '%s' instead of '%s'", get, tt.expected)
		}
	}
}

func TestHost(t *testing.T) {
	testcases := []struct {
		rawurl string
		host   string
		port   int
	}{
		{"http://localhost:8088", "localhost", 8088},
		{"wrong-url", "", 0},
		{"http://other-host:wrong", "other-host", 0},
		{"http://more-host", "more-host", 0},
	}

	for _, tt := range testcases {
		host, port := Host(tt.rawurl)
		if host != tt.host {
			t.Fatalf("wrong host: get '%s' instead of '%s'", host, tt.host)
		}

		if port != tt.port {
			t.Fatalf("wrong port: get '%d' instead of '%d'", port, tt.port)
		}
	}

}
