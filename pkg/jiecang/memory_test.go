package jiecang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMemoryPreset(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		input          []byte // Input
		expectedHeight uint8  // Expected result of function
	}{
		{
			name:           "Correct height with incorrect checksum",
			input:          []byte{0xf2, 0xf2, 0x25, 0x02, 0x04, 0x4e, 0x79, 0x7e},
			expectedHeight: 110,
		},
	}

	for _, test := range tests {
		result := readMemoryPreset(test.input)
		assert.Equal(t, test.expectedHeight, result, test.name)
	}
}
