package jiecang

// Common functions used to decode messages from/to the controller
// checking validity etc.

// Function that checks whether data received from DataIn are valid.
// They should start with "f2f2", end with "7e" and te previous to last byte (which is a checksum) should not fail.
func isValidData(buf []byte) bool {
	// Check length first to prevent index out of bounds
	if len(buf) < 6 {
		return false
	}

	// Check preamble and last byte
	if buf[0] != 0xf2 || buf[1] != 0xf2 || buf[len(buf)-1] != 0x7e {
		return false
	}

	// Calculate checksum and verify if its correct
	dataType := int(buf[2])
	dataLen := int(buf[3])
	// Length of the data should not exceed the length of the payload.
	// Last two bytes should always be the checksum and EoM (Ox7e)
	if dataLen+3 >= len(buf)-2 {
		return false
	}
	receivedChecksum := int(buf[len(buf)-2])

	calcChecksum := dataType + dataLen
	for i := 0; i < dataLen; i++ {
		calcChecksum += int(buf[4+i])
	}
	return (calcChecksum % 256) == receivedChecksum
}
