package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/17xande/keylogger"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(c)
	}()

	go listenLoop(ctx)
	fmt.Println("listening: ")

	<-c
	cancel()
}

func listenLoop(ctx context.Context) {
	for {
		if err := Listen(ctx, "keyboard"); err != nil {
			fmt.Println("error listening to device:", err)
			time.Sleep(1 * time.Second)
		}
	}
}

// Listen to all the Input Devices supplied.
// Return an error if there is a problem, or if one of the devices disconnects.
func Listen(ctx context.Context, dev string) error {
	var e keylogger.InputEvent
	var err error
	var open bool

	ds := keylogger.ScanDevices(dev)
	if len(ds) <= 0 {
		return fmt.Errorf("device '%s' not found", dev)
	}

	cie := make(chan keylogger.InputEvent)
	cer := make(chan error)
	c, cancel := context.WithCancel(ctx)

	defer func() {
		// Cancel all the goroutines reading the device files.
		cancel()
		close(cie)
		close(cer)
	}()

	for _, d := range ds {
		go d.Read(c, cie, cer)
	}

	for {
		select {
		case e, open = <-cie:
			if !open {
				return errors.New("event channel closed")
			}
			fmt.Printf("%2x\t%s\n", e.Type, e.KeyString())
		case err, open = <-cer:
			if !open {
				return errors.New("error channel closed")
			}
			return err
		}
	}
}
