package discovery

import (
	"fmt"
	"time"

	"github.com/facchinm/go-serial-native"
	"github.com/juju/errors"
)

// Merge updates the device with the new one, returning false if the operation
// didn't change anything
func (d *SerialDevice) merge(dev SerialDevice) bool {
	changed := false
	if d.Port != dev.Port {
		changed = true
		d.Port = dev.Port
	}
	if d.SerialNumber != dev.SerialNumber {
		changed = true
		d.SerialNumber = dev.SerialNumber
	}
	if d.ProductID != dev.ProductID {
		changed = true
		d.ProductID = dev.ProductID
	}
	if d.VendorID != dev.VendorID {
		changed = true
		d.VendorID = dev.VendorID
	}

	if d.Serial != dev.Serial {
		d.Serial = dev.Serial
	}
	return changed
}

func (m *Monitor) serialDiscover() error {
	ports, err := serial.ListPorts()
	if err != nil {
		return errors.Annotatef(err, "while listing the serial ports")
	}

	for _, port := range ports {
		m.addSerial(port)

	}
	m.pruneSerial(ports)

	time.Sleep(m.Interval)
	return nil
}

func (m *Monitor) addSerial(port *serial.Info) {
	vid, pid, _ := port.USBVIDPID()
	if vid == 0 || pid == 0 {
		return
	}

	device := SerialDevice{
		Port:         port.Name(),
		SerialNumber: port.USBSerialNumber(),
		ProductID:    fmt.Sprintf("0x%04X", pid),
		VendorID:     fmt.Sprintf("0x%04X", vid),
		Serial:       port,
	}
	for port, dev := range m.serial {
		if port == device.Port {
			changed := dev.merge(device)
			if changed {
				m.Events <- Event{Name: "change", SerialDevice: dev}
			}
			return
		}
	}

	m.serial[device.Port] = &device
	m.Events <- Event{Name: "add", SerialDevice: &device}
}

func (m *Monitor) pruneSerial(ports []*serial.Info) {
	toPrune := []string{}
	for port := range m.serial {
		found := false
		for _, p := range ports {
			if port == p.Name() {
				found = true
			}
		}
		if !found {
			toPrune = append(toPrune, port)
		}
	}

	for _, port := range toPrune {
		m.Events <- Event{Name: "remove", SerialDevice: m.serial[port]}
		delete(m.serial, port)
	}
}
