package util

import (
	"testing"
)

func TestConditionalValue(t *testing.T) {
	tests := []struct {
		cond     bool
		left     interface{}
		right    interface{}
		expected interface{}
	}{
		{true, 10, 20, 10},                 // Condition is true, expect left value
		{false, 10, 20, 20},                // Condition is false, expect right value
		{true, "hello", "world", "hello"},  // Condition is true, expect left value (string)
		{false, "hello", "world", "world"}, // Condition is false, expect right value (string)
		{true, true, false, true},          // Condition is true, expect left value (bool)
		{false, true, false, false},        // Condition is false, expect right value (bool)
		{true, 3.14, 2.71, 3.14},           // Condition is true, expect left value (float64)
		{false, 3.14, 2.71, 2.71},          // Condition is false, expect right value (float64)
		{true, nil, "default", nil},        // Condition is true, expect left value (nil)
		{false, nil, "default", "default"}, // Condition is false, expect right value (string)
	}

	for _, test := range tests {
		result := ConditionalValue(test.cond, test.left, test.right)
		if result != test.expected {
			t.Errorf("ConditionalValue(%t, %v, %v) returned %v, expected %v", test.cond, test.left, test.right, result, test.expected)
		}
	}
}
