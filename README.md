# Key Logger

### **0 Dependency** Library to capture events from **multiple** global input devices on Linux systems.

_Based on [MarinX's](https://github.com/MarinX) [keylogger](https://github.com/MarinX/keylogger)._

### With added features:
- Log/Listen to multiple devices simultaneously.
- Added support for media keys as well as other keys.
- Use a map for key lookup internally, for simplicity.

## Example

```go
import "github.com/17xande/keylogger"

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

// Listen to all the Input Devices supplied. This blocks while listening.
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
		// Close opened channels.
		close(cie)
		close(cer)
	}()

	// Star reading each found device.
	kl.Read(c, cie, cer)

	for {
		select {
		// Listen for device events.
		case e, open = <-cie:
			if !open {
				return errors.New("event channel closed")
			}
			// Handle device event.
			fmt.Printf("%2x\t%s\n", e.Type, e.KeyString())
		// Listen for reading errors.
		case err, open = <-cer:
			if !open {
				return errors.New("error channel closed")
			}
			return err
		}
	}
}
```
