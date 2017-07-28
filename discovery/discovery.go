// Package discovery keeps an updated list of the devices connected to the
// computer, via serial ports or found in the network
//
// Usage:
// 	monitor := discovery.New(time.Millisecond)
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	monitor.Start(ctx)
// 	time.Sleep(10 * time.Second)
// 	fmt.Println(monitor.Serial())
// 	fmt.Println(monitor.Network())
//
// Output:
// 	map[/dev/ttyACM0:0x2341/0x8036]
// 	map[192.168.1.107:YunShield]
//
// You may also decide to subscribe to the Events channel of the Monitor:
//
// 	monitor := discovery.New(time.Millisecond)
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	monitor.Start(ctx)
// 	for ev := range monitor.Events {
// 		fmt.Println(ev)
// 	}
//
// Output:
// {add 0x2341/0x8036 <nil>}
// {add <nil> YunShield}
package discovery

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	serial "github.com/facchinm/go-serial-native"
)

// SerialDevice is something connected to the Serial Ports
type SerialDevice struct {
	Port         string       `json:"port"`
	SerialNumber string       `json:"serial_number"`
	ProductID    string       `json:"pid"`
	VendorID     string       `json:"vid"`
	Serial       *serial.Info `json:"-"`
}

func (d SerialDevice) String() string {
	if d.SerialNumber != "" {
		return fmt.Sprintf(`%s/%s/%s`, d.VendorID, d.ProductID, d.SerialNumber)
	}
	return fmt.Sprintf(`%s/%s`, d.VendorID, d.ProductID)
}

//SerialDevices is a list of currently connected devices to the computer
type SerialDevices map[string]*SerialDevice

// NetworkDevice is something connected to the Network Ports
type NetworkDevice struct {
	Address string `json:"address"`
	Info    string `json:"info"`
	Name    string `json:"name"`
	Port    int    `json:"port"`
}

func (d NetworkDevice) String() string {
	return d.Name
}

//NetworkDevices is a list of currently connected devices to the computer
type NetworkDevices map[string]*NetworkDevice

// Event tells you that something has changed in the list of connected devices.
// Name can be one of ["Add", "Change", "Remove"]
// SerialDevice or NetworkDevice can be present and they refer to the device
// that has been added, changed, or removed
type Event struct {
	Name          string         `json:"name"`
	SerialDevice  *SerialDevice  `json:"serial_device,omitempty"`
	NetworkDevice *NetworkDevice `json:"network_device,omitempty"`
}

// Monitor periodically checks the serial ports and the network in order to have
// an updated list of Serial and Network ports.
//
// You can subscribe to the Events channel to get realtime notification of what's changed
type Monitor struct {
	Interval time.Duration
	Events   chan Event

	serial  SerialDevices
	network NetworkDevices
}

// New Creates a new monitor that can start querying the serial ports and
// the local network for devices
func New(interval time.Duration) *Monitor {
	m := Monitor{
		serial:   SerialDevices{},
		network:  NetworkDevices{},
		Interval: interval,
	}
	return &m
}

// Start begins the loop that queries the serial ports and the local network.
// It accepts a cancelable context
func (m *Monitor) Start(ctx context.Context) {
	m.Events = make(chan (Event))

	var done chan bool
	var stop = false

	go func() {
		<-ctx.Done()
		stop = true
	}()

	go func() {
		for {
			if stop {
				break
			}
			m.serialDiscover()
		}
		done <- true
	}()
	go func() {
		for {
			if stop {
				break
			}
			m.networkDiscover()
		}
		done <- true
	}()

	go func() {
		// We need to wait until both goroutines have finished
		<-done
		<-done
		close(m.Events)
	}()
}

// Serial returns a cached list of devices connected to the serial ports
func (m *Monitor) Serial() SerialDevices {
	return m.serial
}

// Network returns a cached list of devices found on the local network
func (m *Monitor) Network() NetworkDevices {
	return m.network
}
