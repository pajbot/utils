package utils

import "os"

// GetEnv returns an environment key from the system, with an optional value if the environment key isn't set
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
