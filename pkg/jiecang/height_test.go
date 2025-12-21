package jiecang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadHeight(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		input          []byte // Input
		expectedHeight uint8  // Expected result of function
	}{
		{
			name:           "Valid message for height",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x45, 0x7e},
			expectedHeight: 82,
		},
		{
			name:           "Correct height with incorrect checksum",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x46, 0x7e},
			expectedHeight: 82,
		},
	}

	for _, test := range tests {
		result := readHeight(test.input)
		assert.Equal(t, test.expectedHeight, result, test.name)
	}
}
