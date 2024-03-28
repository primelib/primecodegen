package util

// Ternary evaluates a boolean and returns one of two values based on the condition.
func Ternary[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}

	return falseValue
}
