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

var neverReady = make(chan struct{}) // never closed

func main() {
	l := flag.Bool("list", false, "list devices connected to system")
	device := flag.String("device", "keyboard", "device name to listen to")
	flag.Parse()

	if *l {
		list()
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		stop()
	}()

	go listenLoop(ctx, *device)
	fmt.Println("listening: ")

	select {
	case <-neverReady:
		fmt.Println("ready")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context canceled"
		stop()                 // stop receiving signal notifications as soon as possible.
	}
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
func Listen(ctx context.Context, dev string) []error {
	var e keylogger.InputEvent
	var err error
	var open bool
	errs := make([]error, 3)

	kl := keylogger.NewKeyLogger(dev)
	if len(kl.GetDevices()) <= 0 {
		return []error{fmt.Errorf("device '%s' not found", dev)}
	}

	for _, d := range kl.GetDevices() {
		fmt.Printf("Listening to device %s\n", d.Name)
	}

	cie := make(chan keylogger.InputEvent)
	cer := make(chan error)
	cwait := make(chan struct{})

	go kl.Read(ctx, cwait, cie, cer)

	for {
		select {
		// Wait for all goroutines to finish, otherwise they'll get stuck trying
		// to write to a channel without a listener.
		case <-cwait:
			// All goroutines have returned. Safe to break out.
			return errs
		case e, open = <-cie:
			if !open {
				errs = append(errs, errors.New("event channel closed"))
			}

			// Handle key input here.
			fmt.Printf("%2x\t%s\n", e.Type, e.KeyString())

		case err, open = <-cer:
			if !open {
				errs = append(errs, errors.New("error channel closed"))
			}
			errs = append(errs, err)
		}
	}
}
