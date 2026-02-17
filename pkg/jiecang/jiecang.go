package jiecang

import (
	"bytes"
	"fmt"
	"log"
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
	// BLE service and characteristic IDs
	BLEDeviceId      = 0xFE60
	BLECharDataInId  = 0xFE61
	BLECharDataOutId = 0xFE62

	// Protocol constants
	ProtocolPreamble1  = 0xf1 // Command message preamble byte 1
	ProtocolPreamble2  = 0xf1 // Command message preamble byte 2
	ProtocolResponse1  = 0xf2 // Response message preamble byte 1
	ProtocolResponse2  = 0xf2 // Response message preamble byte 2
	ProtocolTerminator = 0x7e // Message terminator

	// Conversion factors
	HeightConversionFactor = 10 // Multiply height by 10 for protocol

	// Memory preset constants
	MemoryPresetModulo = 0x24 // Modulo for memory preset calculation

	// Timing constants
	PollingInterval     = 200 * time.Millisecond // Polling interval for height operations
	SaveMemoryDelay     = 200 * time.Millisecond // Delay after saving memory preset
	OperationTimeout    = 60 * time.Second       // Default timeout for operations
	InitializationDelay = 200 * time.Millisecond // Delay during initialization
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

func Init(a *bluetooth.Adapter, addr bluetooth.Address) (*Jiecang, error) {
	j := new(Jiecang)

	// Connect to BLE Device
	d, err := a.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}

	// Scan Services and characteristics
	services, err := d.DiscoverServices([]bluetooth.UUID{
		bluetooth.New16BitUUID(BLEDeviceId),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	serviceFound := false
	for _, service := range services {
		if service.UUID() != bluetooth.New16BitUUID(BLEDeviceId) {
			// Wrong service
			continue
		}
		serviceFound = true

		// Found the correct service
		// Get a list of characteristics below the service
		characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{
			bluetooth.New16BitUUID(BLECharDataInId),
			bluetooth.New16BitUUID(BLECharDataOutId),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to discover characteristics: %w", err)
		}

		if len(characteristics) < 2 {
			return nil, fmt.Errorf("expected 2 characteristics, got %d", len(characteristics))
		}

		j.dataIn = characteristics[0]
		j.dataOut = characteristics[1]
		// Enable notifications on dataOut
		err = j.dataOut.EnableNotifications(j.characteristicReceiver)
		if err != nil {
			return nil, fmt.Errorf("failed to enable notifications: %w", err)
		}
	}

	if !serviceFound {
		return nil, fmt.Errorf("BLE service %04x not found", BLEDeviceId)
	}

	// Read desk current height
	result := []byte{}
	_, err = j.dataOut.Read(result)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial height: %w", err)
	}
	log.Printf("Initial height: %d mm", j.currentHeight)

	//Fetch height memory presets
	j.presets = make(map[string]uint8)
	if err := j.FetchHeight(); err != nil {
		return nil, fmt.Errorf("failed to fetch height: %w", err)
	}

	// Fetch desk low and high height
	if err := j.FetchHeightRange(); err != nil {
		return nil, fmt.Errorf("failed to fetch height range: %w", err)
	}

	if err := j.FetchStandTime(); err != nil {
		return nil, fmt.Errorf("failed to fetch stand time: %w", err)
	}

	if err := j.FetchAllTime(); err != nil {
		return nil, fmt.Errorf("failed to fetch all time: %w", err)
	}

	j.device = d
	return j, nil
}

func (j *Jiecang) Disconnect() error {
	return j.device.Disconnect()
}

// simple wrapper to bluetooth.WriteWithoutResponse
func (j *Jiecang) sendCommand(buf []byte) error {
	_, err := j.dataIn.WriteWithoutResponse(buf)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	return nil
}

func (j *Jiecang) FetchStandTime() error {
	// Implements fetch_stand_time command
	if err := j.sendCommand(commands["fetch_stand_time"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_stand_time"])
}

func (j *Jiecang) FetchAllTime() error {
	// Implements fetch_all_time command
	if err := j.sendCommand(commands["fetch_all_time"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_all_time"])
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
		// Check that message has minimum required length before validation
		if len(msg[i]) < 3 {
			continue
		}

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
				memory := int(msg[i][2] % MemoryPresetModulo)
				memoryName := fmt.Sprintf("memory%d", memory)
				j.mu.Lock()
				j.presets[memoryName] = readMemoryPreset(msg[i])
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
