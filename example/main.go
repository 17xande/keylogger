package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"github.com/17xande/keylogger"
)

func main() {
	l := flag.Bool("list", false, "list devices connected to system")
	device := flag.String("device", "keyboard", "device name to listen to")
	flag.Parse()

	if *l {
		list()
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(c)
	}()

	go listenLoop(ctx, *device)
	fmt.Println("listening: ")

	<-c
	cancel()
}

func list() {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintf(w, "Id\tName\t\n")
	kl := keylogger.NewKeyLogger("")
	for i, d := range kl.GetDevices() {
		fmt.Fprintf(w, "%d\t%s\n", i, d.Name)
	}
	w.Flush()
}

func listenLoop(ctx context.Context, device string) {
	for {
		if err := Listen(ctx, device); err != nil {
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

	kl := keylogger.NewKeyLogger(dev)
	if len(kl.GetDevices()) <= 0 {
		return fmt.Errorf("device '%s' not found", dev)
	}

	for _, d := range kl.GetDevices() {
		fmt.Printf("Listening to device %s\n", d.Name)
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

	for _, d := range kl.GetDevices() {
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
