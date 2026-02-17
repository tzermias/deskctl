package jiecang

// Common functions used to decode messages from/to the controller
// checking validity etc.

// isValidData validates data received from the controller's DataOut characteristic.
//
// The Jiecang protocol uses the following message format:
//   - Bytes 0-1: Preamble (0xf2, 0xf2)
//   - Byte 2: Message type/command
//   - Byte 3: Data length (number of data bytes)
//   - Bytes 4..(4+dataLen-1): Data payload
//   - Byte (len-2): Checksum (sum of type, length, and data bytes, mod 256)
//   - Byte (len-1): Terminator (0x7e)
//
// Parameters:
//   - buf: Raw response buffer from the controller
//
// Returns true if the message has valid preamble, terminator, and checksum;
// false otherwise.
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
