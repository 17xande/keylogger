# Key Logger

### **0 Dependency** Library to capture events from **multiple** global input devices on Linux systems.

_Heavily based on [MarinX's](https://github.com/MarinX) [keylogger](https://github.com/MarinX/keylogger)._

### Why not just use [MarinX's](https://github.com/MarinX) [keylogger](https://github.com/MarinX/keylogger)?
Some devices, like keyboards, show up as multiple devices on Linux. To be able to listen to all key events, this version of `keyloger` listens to all devices that match a provided substring using multiple goroutines.

## Example

```go
func main() {
	device := flag.String("device", "keyboard", "device name to listen to")
	flag.Parse()

  // Wait for Ctrl + C to exit
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

func listenLoop(ctx context.Context, device string) {
  // We can use an infinite loop to start listening again if something
  // goes wrong, eg: a device is disconnected.
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

	ds := keylogger.ScanDevices(dev)
	if len(ds) <= 0 {
		return fmt.Errorf("device '%s' not found", dev)
	}

	for _, d := range ds {
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
```
