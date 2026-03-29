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

func SliceToMapWithKeyFunc[T any, K comparable](items []T, keyFunc func(T) K) map[K]T {
	m := make(map[K]T, len(items))
	for _, item := range items {
		m[keyFunc(item)] = item
	}
	return m
}

// AppendUnique appends elements from the source slice to the target slice, ensuring that only unique elements are added.
func AppendUnique(target []string, source []string) []string {
	if len(source) == 0 {
		return target
	}

	// Build a lookup set for the target
	exists := make(map[string]struct{}, len(target))
	for _, item := range target {
		exists[item] = struct{}{}
	}

	for _, item := range source {
		if _, ok := exists[item]; !ok {
			target = append(target, item)
			exists[item] = struct{}{} // Prevent duplicates within the source itself
		}
	}
	return target
}
