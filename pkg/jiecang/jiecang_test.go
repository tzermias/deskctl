package jiecang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidData(t *testing.T) {
	tests := []struct {
		// Test case struct
		name           string // Name of the testcase
		input          []byte // Input
		expectedResult bool   // Expected result of function
	}{
		{
			name:           "Valid message for height",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x45, 0x7e},
			expectedResult: true,
		},
		{
			name:           "Valid message for height range",
			input:          []byte{0xf2, 0xf2, 0x07, 0x04, 0x04, 0xf8, 0x02, 0x6c, 0x75, 0x7e},
			expectedResult: true,
		},
		{
			name:           "Valid message for memory presets",
			input:          []byte{0xf2, 0xf2, 0x25, 0x02, 0x04, 0x4e, 0x79, 0x7e},
			expectedResult: true,
		},
		{
			name:           "Valid message for unit settings",
			input:          []byte{0xf2, 0xf2, 0x0e, 0x01, 0x00, 0x0f, 0x7e},
			expectedResult: true,
		},
		{
			name:           "Message for height range with incorrect checksum",
			input:          []byte{0xf2, 0xf2, 0x07, 0x04, 0x04, 0xf8, 0x02, 0x6c, 0x48, 0x7e},
			expectedResult: false,
		},
		{
			name:           "Message that doesn't start with f2f2",
			input:          []byte{0xde, 0xad, 0xbe, 0xef, 0x7e},
			expectedResult: false,
		},
		{
			name:           "Message that doesn't end with 7e",
			input:          []byte{0xf2, 0xf2, 0xca, 0xfe},
			expectedResult: false,
		},
		{
			name:           "Message with length less than 6 bytes",
			input:          []byte{0xf2, 0xf2, 0x01, 0x00, 0x02, 0x7e},
			expectedResult: false,
		},
		{
			name:           "Message with incorrect length parameter",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x04, 0x7e},
			expectedResult: false,
		},
	}

	for _, test := range tests {
		result := isValidData(test.input)
		assert.Equal(t, test.expectedResult, result, test.name)
	}
}

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
