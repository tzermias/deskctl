package jiecang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandChecksums(t *testing.T) {
	for name, cmd := range commands {
		if !assert.GreaterOrEqual(t, len(cmd), 6, name) {
			continue
		}
		assert.Equal(t, []byte{0xf1, 0xf1}, cmd[:2], name+": preamble")
		assert.Equal(t, byte(0x7e), cmd[len(cmd)-1], name+": terminator")
		dataLen := int(cmd[3])
		if !assert.Equal(t, dataLen+6, len(cmd), name+": length") {
			continue
		}
		sum := 0
		for _, b := range cmd[2 : 4+dataLen] {
			sum += int(b)
		}
		assert.Equal(t, byte(sum%256), cmd[len(cmd)-2], name+": checksum")
	}
}
