package jiecang

import (
	"log"
	"math"
	"time"
)

// GoToMemoryX functions
func (j *Jiecang) GoToMemory1() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory1"])
	for (j.currentHeight - j.presets["memory1"]) != 0 {
		j.sendCommand(commands["goto_memory1"])

		time.Sleep(200 * time.Millisecond)
	}
}

func (j *Jiecang) GoToMemory2() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory2"])
	for (j.currentHeight - j.presets["memory2"]) != 0 {
		j.sendCommand(commands["goto_memory2"])
		time.Sleep(200 * time.Millisecond)
	}
}

func (j *Jiecang) GoToMemory3() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory3"])
	for (j.currentHeight - j.presets["memory3"]) != 0 {
		j.sendCommand(commands["goto_memory3"])
		time.Sleep(200 * time.Millisecond)
	}
}

// Save memory commands

func (j *Jiecang) SaveMemory1() {
	//Save memory
	j.sendCommand(commands["save_memory1"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

func (j *Jiecang) SaveMemory2() {
	//Save memory
	j.sendCommand(commands["save_memory2"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

func (j *Jiecang) SaveMemory3() {
	//Save memory
	j.sendCommand(commands["save_memory3"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

// Reads the response of the controller containing height settings of
// memory presets.
func readMemoryPreset(buf []byte) uint8 {
	if buf[3] == 0x02 {
		preset := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(preset / 10.0)))
	}
	return 0
}
