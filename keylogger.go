// Package keylogger is a simple 0 dependency keylogger package
package keylogger

import (
	"fmt"
	"io/ioutil"
)

const (
	inputPath  = "/sys/class/input/event%d/device/uevent"
	deviceFile = "/dev/input/event%d"
)

// KeyLogger keeps a reference to the InputDevice that it's listening to
type KeyLogger struct {
	inputDevices []*InputDevice
}

// GetDevices gets the desired input device, or returns all of them if no device name is sent
func GetDevices(deviceName string) []*InputDevice {
	var devs []*InputDevice

	for i := 0; i < 255; i++ {
		// TODO check if file exists first
		buff, err := ioutil.ReadFile(fmt.Sprintf(inputPath, i))
		if err != nil {
			break
		}
		dev := newInputDevice(buff, i)
		if deviceName == "" || deviceName != "" && deviceName == dev.Name {
			devs = append(devs, dev)
		}
	}

	return devs
}

// NewKeyLogger creates a new keylogger for a device, based on it's name
func NewKeyLogger(deviceName string) *KeyLogger {
	devs := GetDevices(deviceName)
	return &KeyLogger{
		inputDevices: devs,
	}
}
