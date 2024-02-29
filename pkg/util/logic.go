package util

// ConditionalValue evaluates a conditional expression and returns one of two values based on the condition.
func ConditionalValue(condition bool, left interface{}, right interface{}) interface{} {
	if condition {
		return left
	}

	return right
}
