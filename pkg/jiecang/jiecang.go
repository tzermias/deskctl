// Package jiecang provides control functionality for Jiecang standing desk controllers
// via Bluetooth Low Energy (BLE) communication using the Lierda LSD4BT-E95ASTD001 module.
//
// The package implements the Jiecang UART protocol over BLE, supporting:
//   - Height control (up/down, go to specific height)
//   - Memory presets (save and recall positions)
//   - Height range queries
//   - Desk settings (memory mode, anti-collision sensitivity)
//
// Example usage:
//
//	ctx := context.Background()
//	adapter := bluetooth.DefaultAdapter
//	adapter.Enable()
//
//	address := bluetooth.MustParseMAC("AA:BB:CC:DD:EE:FF")
//	desk, err := jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: address}})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer desk.Disconnect()
//
//	// Move desk to 100cm height with timeout
//	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
//	defer cancel()
//	desk.GoToHeight(ctx, 100)
package jiecang

import (
	"bytes"
	"fmt"
	"log"
	"sync"

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

// Jiecang represents a connection to a Jiecang desk controller.
// It manages BLE communication and maintains the current state of the desk.
//
// The struct uses a mutex to protect concurrent access to shared state,
// allowing safe use from multiple goroutines. Height values are stored
// in centimeters for convenience.
type Jiecang struct {
	device  bluetooth.Device               // BLE device connection
	dataIn  bluetooth.DeviceCharacteristic // Write characteristic for sending commands
	dataOut bluetooth.DeviceCharacteristic // Read characteristic for receiving responses

	currentHeight uint8        // Current height in centimeters
	mu            sync.RWMutex // Protects concurrent access to shared state

	presets map[string]uint8 // Memory presets (memory1-4) in centimeters

	// LowestHeight is the minimum height limit of the desk in centimeters.
	// Set during initialization from the controller.
	LowestHeight uint8

	// HighestHeight is the maximum height limit of the desk in centimeters.
	// Set during initialization from the controller.
	HighestHeight uint8

	// MemoryConstantTouchMode indicates if memory mode requires constant touch.
	// false = one-touch mode, true = constant touch mode.
	MemoryConstantTouchMode bool

	// AntiCollisionSensitivity indicates the anti-collision sensitivity level.
	// Valid values: 1 = High, 2 = Medium, 3 = Low
	AntiCollisionSensitivity uint8
}

// Init initializes a connection to a Jiecang desk controller via Bluetooth.
//
// The function performs the following steps:
//  1. Connects to the BLE device at the specified address
//  2. Discovers the Jiecang service (0xFE60)
//  3. Discovers data input/output characteristics (0xFE61, 0xFE62)
//  4. Enables notifications for receiving responses
//  5. Queries the desk for current height, height range, and memory presets
//
// Returns an error if any step fails (connection, service discovery,
// characteristic discovery, or initial queries).
//
// Example:
//
//	adapter := bluetooth.DefaultAdapter
//	adapter.Enable()
//	address := bluetooth.MustParseMAC("AA:BB:CC:DD:EE:FF")
//	desk, err := jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: address}})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer desk.Disconnect()
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

// Disconnect closes the BLE connection to the desk controller.
// Should be called when done using the controller to free resources.
// Safe to call even if the connection is already closed.
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

// FetchStandTime requests the desk's standing time statistics from the controller.
// The command is sent twice as required by the protocol for reliability.
// Returns an error if the command transmission fails.
func (j *Jiecang) FetchStandTime() error {
	if err := j.sendCommand(commands["fetch_stand_time"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_stand_time"])
}

// FetchAllTime requests the desk's total usage time statistics from the controller.
// The command is sent twice as required by the protocol for reliability.
// Returns an error if the command transmission fails.
func (j *Jiecang) FetchAllTime() error {
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
				memory := int(msg[i][2] % 0x24)
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
