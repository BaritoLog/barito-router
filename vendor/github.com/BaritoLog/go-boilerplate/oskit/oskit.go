package oskit

import "os"

// Getenv return environment variable if available or default value
func Getenv(key, defaultValue string) (val string) {
	val = os.Getenv(key)
	if len(val) < 1 {
		val = defaultValue
	}
	return
}
