package jiecang

// Common functions used to decode messages from/to the controller
// checking validity etc.

// Function that checks whether data received from DataIn are valid.
// They should start with "f2f2", end with "7e" and te previous to last byte (which is a checksum) should not fail.
func isValidData(buf []byte) bool {
	// Check preamble and last byte
	if buf[0] != 0xf2 || buf[1] != 0xf2 || buf[len(buf)-1] != 0x7e || len(buf) < 6 {
		return false
	}

	// Calculate checksum and verify if its correct
	data_type := int(buf[2])
	data_len := int(buf[3])
	// Length of the data should not exceed the length of the payload.
	// Last two bytes should always be the checksum and EoM (Ox7e)
	if data_len+3 >= len(buf)-2 {
		return false
	}
	received_checksum := int(buf[len(buf)-2])

	calc_checksum := data_type + data_len
	for i := 0; i < data_len; i++ {
		calc_checksum += int(buf[4+i])
	}
	return (calc_checksum % 256) == received_checksum
}
