package envkit

import (
	"os"
	"strconv"
	"strings"
)

func GetString(key, defaultValue string) (val string, success bool) {
	s := os.Getenv(key)
	if len(s) > 0 {
		return s, true
	}

	return defaultValue, false
}

func GetInt(key string, defaultValue int) (val int, success bool) {
	s := os.Getenv(key)
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, true
	}

	return defaultValue, false
}

func GetSlice(key, separator string, defaultSlice []string) (val []string, success bool) {
	s := os.Getenv(key)

	if len(s) > 0 {
		return strings.Split(s, separator), true
	}

	return defaultSlice, false

}
