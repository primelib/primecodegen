package util

import (
	"slices"
)

// CountExcluding counts the number of occurrences of elements in a slice, excluding specified values.
func CountExcluding[S ~[]E, E comparable](s S, exclude ...E) int {
	count := 0
	for i := range s {
		if !slices.Contains(exclude, s[i]) {
			count = count + 1
		}
	}
	return count
}
