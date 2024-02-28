package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountExcluding(t *testing.T) {
	result := CountExcluding([]string{"a", "b", "c", "d", "e"}, []string{"a", "b"}...)
	assert.Equal(t, 3, result)
}
