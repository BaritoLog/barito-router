package envkit

import (
	"os"
	"strconv"
	"strings"
)

func GetString(key, defaultValue string) string {
	s := os.Getenv(key)
	if len(s) > 0 {
		return s
	}

	return defaultValue
}

func GetInt(key string, defaultValue int) int {
	s := os.Getenv(key)
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}

	return defaultValue
}

func GetSlice(key, separator string, defaultSlice []string) []string {
	s := os.Getenv(key)

	if len(s) > 0 {
		return strings.Split(s, separator)
	}

	return defaultSlice

}
