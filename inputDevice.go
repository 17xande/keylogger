package keylogger

import (
	"bufio"
	"bytes"
	"strings"
	"syscall"
)

// InputDevice is a reference to a physical device plugged into the machine
type InputDevice struct {
	ID   int
	Name string
}

// InputEvent is an event invoked by and InputDevice
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

func newInputDevice(buff []byte, id int) *InputDevice {
	rd := bufio.NewReader(bytes.NewReader(buff))
	rd.ReadLine()              // not sure why we have to read one line first?
	dev, _, _ := rd.ReadLine() // not sure why we're ignoring things?
	splt := strings.Split(string(dev), "=")

	return &InputDevice{
		ID:   id,
		Name: splt[1],
	}
}
