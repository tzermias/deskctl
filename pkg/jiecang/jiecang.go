package jiecang

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

var commands = map[string][]byte{
	"up":                 {0xf1, 0xf1, 0x01, 0x00, 0x01, 0x7e},
	"down":               {0xf1, 0xf1, 0x02, 0x00, 0x02, 0x7e},
	"save_memory1":       {0xf1, 0xf1, 0x03, 0x00, 0x03, 0x7e},
	"save_memory2":       {0xf1, 0xf1, 0x04, 0x00, 0x04, 0x7e},
	"goto_memory1":       {0xf1, 0xf1, 0x05, 0x00, 0x05, 0x7e},
	"goto_memory2":       {0xf1, 0xf1, 0x06, 0x00, 0x06, 0x7e},
	"fetch_height":       {0xf1, 0xf1, 0x07, 0x00, 0x07, 0x7e},
	"fetch_height_range": {0xf1, 0xf1, 0x0c, 0x00, 0x0c, 0x7e},
	"save_memory3":       {0xf1, 0xf1, 0x25, 0x00, 0x25, 0x7e},
	"goto_memory3":       {0xf1, 0xf1, 0x27, 0x00, 0x27, 0x7e},
	"stop":               {0xf1, 0xf1, 0x2b, 0x00, 0x27, 0x7e},
	"fetch_stand_time":   {0xf1, 0xf1, 0xa2, 0x00, 0xa2, 0x7e},
	"fetch_all_time":     {0xf1, 0xf1, 0xaa, 0x00, 0xaa, 0x7e},
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

	//Desk settings
	// Memory mode (One-touch mode vs constant touch)
	MemoryConstantTouchMode bool
	// Anti-collision sensitivity (1 High, 2 Medium, 3 Low)
	AntiCollisionSensitivity uint8 //TODO: Use iota
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

	j.FetchStandTime()
	j.FetchAllTime()

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

// Moves the desk up
func (j *Jiecang) Up() {
	j.sendCommand(commands["up"])
}

// Moves the desk down
func (j *Jiecang) Down() {
	j.sendCommand(commands["down"])
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
	j.sendCommand(commands["fetch_height_range"])

}

func (j *Jiecang) FetchStandTime() {
	// Implements fetch_stand_time command
	j.sendCommand(commands["fetch_stand_time"])
	j.sendCommand(commands["fetch_stand_time"])
}

func (j *Jiecang) FetchAllTime() {
	// Implements fetch_all_time command
	j.sendCommand(commands["fetch_all_time"])
	j.sendCommand(commands["fetch_all_time"])
}

func (j *Jiecang) GoToHeight(height uint8) {
	//Ensure that height is within low and high limits of the desk.
	if height > j.HighestHeight || height < j.LowestHeight {
		fmt.Printf("Height %d is out of range of the desk (Low: %d, High: %d)\n", height, j.LowestHeight, j.HighestHeight)
		return
	}
	data0 := byte((int(height) * 10) / 256)
	data1 := byte((int(height) * 10) % 256)
	command := []byte{
		0xf1,
		0xf1,
		0x1b,
		0x02,
		data0,
		data1,
		byte((int(0x1b) + int(0x02) + int(data0) + int(data1)) % 256),
		0x7e,
	}

	j.sendCommand(command)
	for (j.currentHeight - height) != 0 {
		j.sendCommand(command)
		time.Sleep(200 * time.Millisecond)
	}
	j.sendCommand(commands["stop"])
}

func (j *Jiecang) characteristicReceiver(buf []byte) {
	/* Data should always start with f2f2 (2 bytes) and end with 7e
	3rd byte is type/command?
	4th byte is length (in bytes) of returned data
	bytes 5 etc are data.
	Previous to last byte is the checksum.
	*/

	// Buffer might contain multiple messages
	msg := bytes.SplitAfter(buf, []byte{0x7e})
	for i := 0; i < len(msg)-1; i++ {
		if isValidData(msg[i]) {
			switch msg[i][2] {
			case 0x01: // Data contains height measurements
				//f2 f2 01 03 03 37 07 45 7e
				// Use mutex to set current height
				j.mu.Lock()
				j.currentHeight = readHeight(msg[i])
				j.mu.Unlock()
			case 0x07: // Data contains height range of desk
				j.mu.Lock()
				j.HighestHeight, j.LowestHeight = readHeightRange(msg[i])
				j.mu.Unlock()
			case 0x25, 0x26, 0x27, 0x28: // Data contains height for each memory preset (1-4). Memory 4 is currently 0
				memory := int(msg[i][2] % 0x24)
				memory_name := fmt.Sprintf("memory%d", memory)
				j.mu.Lock()
				j.presets[memory_name] = readMemoryPreset(msg[i])
				j.mu.Unlock()
			case 0x0e: // Data contains units setting
				fmt.Printf("Unit settings: %x\n", msg[i][3])
			case 0x17: // Unknonwn setting so far
				continue
			case 0x19: // Data contains memory mode setting
				if msg[i][3] == 0x01 {
					j.mu.Lock()
					j.MemoryConstantTouchMode = true
					j.mu.Unlock()
				}
			case 0x1b: // Data contains response from go to height command
				continue
			case 0x1d: // Data contains anti-collision sensitivity
				j.mu.Lock()
				j.AntiCollisionSensitivity = uint8(msg[i][3])
				j.mu.Unlock()
			default: // Any other case
				log.Printf("Received: %x", msg[i])
			}
		} else {
			log.Printf("Received: %x", msg[i])
		}
	}
}

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

func readHeight(buf []byte) uint8 {
	if buf[3] == 0x03 {
		height := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(height / 10.0)))
	}
	return 0
}

func readMemoryPreset(buf []byte) uint8 {
	if buf[3] == 0x02 {
		preset := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(preset / 10.0)))
	}
	return 0
}

func readHeightRange(buf []byte) (uint8, uint8) {
	if buf[3] == 0x04 {
		highestHeight := int(buf[4])*256 + int(buf[5])
		lowestHeight := int(buf[6])*256 + int(buf[7])
		return uint8(math.Round(float64(highestHeight / 10.0))),
			uint8(math.Round(float64(lowestHeight / 10.0)))
	}
	return 0, 0
}
