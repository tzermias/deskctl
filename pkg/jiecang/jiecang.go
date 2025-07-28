package jiecang

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

var commands = map[string][]byte{
	"up":                 []byte{0xf1, 0xf1, 0x01, 0x00, 0x01, 0x7e},
	"down":               []byte{0xf1, 0xf1, 0x02, 0x00, 0x02, 0x7e},
	"save_memory1":       []byte{0xf1, 0xf1, 0x03, 0x00, 0x03, 0x7e},
	"save_memory2":       []byte{0xf1, 0xf1, 0x04, 0x00, 0x04, 0x7e},
	"goto_memory1":       []byte{0xf1, 0xf1, 0x05, 0x00, 0x05, 0x7e},
	"goto_memory2":       []byte{0xf1, 0xf1, 0x06, 0x00, 0x06, 0x7e},
	"fetch_height":       []byte{0xf1, 0xf1, 0x07, 0x00, 0x07, 0x7e},
	"fetch_height_range": []byte{0xf1, 0xf1, 0x0c, 0x00, 0x0c, 0x7e},
}

const (
	BLEDeviceId      = 0xFE60
	BLECharDataInId  = 0xFE61
	BLECharDataOutId = 0xFE62
)

type Jiecang struct {
	device  bluetooth.Device
	dataIn  bluetooth.DeviceCharacteristic
	dataOut bluetooth.DeviceCharacteristic

	//currentHeight in cm
	currentHeight uint8
	mu            sync.RWMutex

	//Memory presets (memory 1-4)
	presets map[string]uint8

	//Highest and lowest height of desk
	LowestHeight  uint8
	HighestHeight uint8
}

func Init(a *bluetooth.Adapter, addr bluetooth.Address) *Jiecang {
	j := new(Jiecang)

	// Connect to BLE Device
	d, err := a.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		log.Println("Error connecting to device:", err.Error())
	}

	// Scan Services and characteristics
	services, err := d.DiscoverServices([]bluetooth.UUID{
		bluetooth.New16BitUUID(BLEDeviceId),
	})

	if err != nil {
		log.Println("error getting services:", err.Error())
	}
	for _, service := range services {
		if service.UUID() != bluetooth.New16BitUUID(BLEDeviceId) {
			// Wrong service
			continue
		}
		// Found the correct service
		// Get a list of characteristics below the service
		characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{
			bluetooth.New16BitUUID(BLECharDataInId),
			bluetooth.New16BitUUID(BLECharDataOutId),
		})
		if err != nil {
			log.Println("error getting characteristics:", err.Error())
		}

		j.dataIn = characteristics[0]
		j.dataOut = characteristics[1]
		// Enable notifications on dataOut
		err = j.dataOut.EnableNotifications(j.characteristicReceiver)
		// If error, bail out
		if err != nil {
			log.Println("error enabling notifications:", err.Error())
		}
	}

	// Read desk current height
	result := []byte{}
	_, err = j.dataOut.Read(result)
	if err != nil {
		log.Println("error reading from DataIn characteristic", err.Error())
	}
	log.Printf("Initial height: %d mm", j.currentHeight)

	//Fetch height memory presets
	j.presets = make(map[string]uint8)
	j.FetchHeight()

	// Fetch desk low and high height
	j.FetchHeightRange()

	j.device = d
	return j
}

func (j *Jiecang) Disconnect() error {
	return j.device.Disconnect()
}

// simple wrapper to bluetooth.WriteWithoutResponse
func (j *Jiecang) sendCommand(buf []byte) {
	_, _ = j.dataIn.WriteWithoutResponse(buf)
}

func (j *Jiecang) GoToMemory1() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory1"])
	for (j.currentHeight - j.presets["memory1"]) != 0 {
		j.sendCommand(commands["goto_memory1"])

		//log.Printf("Height: %d cm Preset1: %d", j.currentHeight, j.presets["memory1"])
		time.Sleep(200 * time.Millisecond)
	}
}

func (j *Jiecang) GoToMemory2() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory2"])
	for (j.currentHeight - j.presets["memory2"]) != 0 {
		j.sendCommand(commands["goto_memory2"])
		//log.Printf("Height: %d cm", j.currentHeight)
		time.Sleep(200 * time.Millisecond)
	}
}

func (j *Jiecang) GoToMemory3() {
	// Send the connand twice the first time
	j.sendCommand(commands["goto_memory3"])
	for (j.currentHeight - j.presets["memory3"]) != 0 {
		j.sendCommand(commands["goto_memory3"])
		//log.Printf("Height: %d cm", j.currentHeight)
		time.Sleep(200 * time.Millisecond)
	}
}

func (j *Jiecang) SaveMemory1() {
	//Save memory
	j.sendCommand(commands["save_memory1"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

func (j *Jiecang) FetchHeight() {
	//Implements fetch_height command

	// Returns
	/*
		f2f2 25 02 044e 79 7e //044e =1100 in decimal. Memory 1
		f2f2 25 02 044e 79 7e
		f2f2 26 02 030c 37 7e //030c = 780 in dec. Memory 2
		f2f2 26 02 030c 37 7e
		f2f2 27 02 0372 9e 7e //0372 = 882 in dec. Memory 3
		f2f2 27 02 0372 9e 7e
		f2f2 28 02 0000 2a 7e // Memory 4?
		f2f2 28 02 0000 2a 7e

	*/
	j.sendCommand(commands["fetch_height"])
	j.sendCommand(commands["fetch_height"])

}

func (j *Jiecang) FetchHeightRange() {
	// Implements fetch_height_range command

	//Retuns
	/*
			  f2f2 07 04 04f8 026c 75 7e
		            LEN HGH LOW  CSUM
	*/
	j.sendCommand(commands["fetch_height_range"])

}

func (j *Jiecang) characteristicReceiver(buf []byte) {
	/* Data should always start with f2f2 (2 bytes) and end with 7e
	3rd byte is type/command?
	4th byte is length (in bytes) of returned data
	bytes 5 etc are data.
	Previous to last byte is the checksum.
	*/

	if isValidData(buf) {
		switch buf[2] {
		case 0x01: // Data contains height measurements
			//f2 f2 01 03 03 37 07 45 7e
			// Use mutex to set current height
			j.mu.Lock()
			j.currentHeight = readHeight(buf)
			j.mu.Unlock()
		case 0x07: // Data contains height range of desk
			j.mu.Lock()
			j.HighestHeight, j.LowestHeight = readHeightRange(buf)
			j.mu.Unlock()
		case 0x25, 0x26, 0x27, 0x28: // Data contains height for each memory preset (1-4). Memory 4 is currently 0
			memory, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[2]%0x24), 16, 32)
			memory_name := fmt.Sprintf("memory%d", memory)
			j.mu.Lock()
			j.presets[memory_name] = readMemoryPreset(buf)
			j.mu.Unlock()
		default: // Any other case
			log.Printf("Received: %x", buf)
		}
	} else {
		log.Printf("Received: %x", buf)
	}
}

// Function that checks whether data received from DataIn are valid.
// They should start with "f2f2", end with "7e" and te previous to last byte (which is a checksum) should not fail.
func isValidData(buf []byte) bool {
	// Check preamble and last byte
	if buf[0] != 0xf2 || buf[1] != 0xf2 || buf[len(buf)-1] != 0x7e {
		return false
	}

	// Calculate checksum and verify if its correct
	data_type, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[2]), 16, 32)
	data_len, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[3]), 16, 32)
	received_checksum, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[len(buf)-2]), 16, 32)

	calc_checksum := data_type + data_len
	for i := 0; i < int(data_len); i++ {
		tmp, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[4+i]), 16, 32)
		calc_checksum += tmp
	}
	return (calc_checksum % 256) == received_checksum
}

func readHeight(buf []byte) uint8 {
	if buf[3] == 0x03 {
		height, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[4:6]), 16, 32)
		// Hack to round the value
		return uint8(math.Round(float64(height / 10.0)))
	}
	return 0
}

func readMemoryPreset(buf []byte) uint8 {
	if buf[3] == 0x02 {
		preset, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[4:6]), 16, 32)
		// Hack to round the value
		return uint8(math.Round(float64(preset / 10.0)))
	}
	return 0
}

func readHeightRange(buf []byte) (uint8, uint8) {
	if buf[3] == 0x04 {
		highestHeight, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[4:6]), 16, 32)
		lowestHeight, _ := strconv.ParseInt(fmt.Sprintf("%x", buf[6:8]), 16, 32)
		return uint8(math.Round(float64(highestHeight / 10.0))),
			uint8(math.Round(float64(lowestHeight / 10.0)))
	}
	return 0, 0
}
