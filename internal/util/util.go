package util

import (
	"time"
)

// MaxTime returns the later of the given times.
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

// MinTime returns the earlier of the given times.
func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}

func MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
