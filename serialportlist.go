package main

import (
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/arduino/arduino-create-agent/tools"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/sirupsen/logrus"
)

type serialPortList struct {
	tools *tools.Tools

	Ports     []*SpPortItem
	portsLock sync.Mutex

	OnList func([]byte) `json:"-"`
	OnErr  func(string) `json:"-"`
}

// SpPortItem is the serial port item
type SpPortItem struct {
	Name            string
	SerialNumber    string
	DeviceClass     string
	IsOpen          bool
	IsPrimary       bool
	Baud            int
	BufferAlgorithm string
	Ver             string
	VendorID        string
	ProductID       string
}

func newSerialPortList(tools *tools.Tools) *serialPortList {
	return &serialPortList{tools: tools}
}

// List broadcasts a Json representation of the ports found
func (sp *serialPortList) List() {
	sp.portsLock.Lock()
	ls, err := json.MarshalIndent(sp, "", "\t")
	sp.portsLock.Unlock()

	if err != nil {
		sp.OnErr("Error creating json on port list " + err.Error())
	} else {
		sp.OnList(ls)
	}
}

// MarkPortAsOpened marks a port as opened by the user
func (sp *serialPortList) MarkPortAsOpened(portname string) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	port := sp.getPortByName(portname)
	if port != nil {
		port.IsOpen = true
	}
}

// MarkPortAsClosed marks a port as no more opened by the user
func (sp *serialPortList) MarkPortAsClosed(portname string) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	port := sp.getPortByName(portname)
	if port != nil {
		port.IsOpen = false
	}
}

// Run is the main loop for port discovery and management
func (sp *serialPortList) Run() {
	for retries := 0; retries < 10; retries++ {
		sp.runSerialDiscovery()

		logrus.Errorf("Serial discovery stopped working, restarting it in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
	logrus.Errorf("Failed restarting serial discovery. Giving up...")
}

func (sp *serialPortList) runSerialDiscovery() {
	// First ensure that all the discoveries are available
	noOpProgress := func(msg string) {}
	if err := sp.tools.Download("builtin", "serial-discovery", "latest", "keep", noOpProgress); err != nil {
		logrus.Errorf("Error downloading serial-discovery: %s", err)
		panic(err)
	}
	sd, err := sp.tools.GetLocation("serial-discovery")
	if err != nil {
		logrus.Errorf("Error downloading serial-discovery: %s", err)
		panic(err)
	}
	d := discovery.NewClient("serial", sd+"/serial-discovery")
	dLogger := logrus.WithField("discovery", "serial")
	if *verbose {
		d.SetLogger(dLogger)
	}
	d.SetUserAgent("arduino-create-agent/" + version)
	if err := d.Run(); err != nil {
		logrus.Errorf("Error running serial-discovery: %s", err)
		panic(err)
	}
	defer d.Quit()

	events, err := d.StartSync(10)
	if err != nil {
		logrus.Errorf("Error starting event watcher on serial-discovery: %s", err)
		panic(err)
	}

	logrus.Infof("Serial discovery started, watching for events")
	for ev := range events {
		logrus.WithField("event", ev).Debugf("Serial discovery event")
		switch ev.Type {
		case "add":
			sp.add(ev.Port)
		case "remove":
			sp.remove(ev.Port)
		}
	}

	sp.reset()
	logrus.Errorf("Serial discovery stopped.")
}

func (sp *serialPortList) reset() {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	sp.Ports = []*SpPortItem{}
}

func (sp *serialPortList) add(addedPort *discovery.Port) {
	if addedPort.Protocol != "serial" {
		return
	}
	props := addedPort.Properties
	if !props.ContainsKey("vid") {
		return
	}
	vid, pid := props.Get("vid"), props.Get("pid")
	if vid == "0x0000" || pid == "0x0000" {
		return
	}
	if portsFilter != nil && !portsFilter.MatchString(addedPort.Address) {
		logrus.Debugf("ignoring port not matching filter. port: %v\n", addedPort.Address)
		return
	}

	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()

	// If the port is already in the list, just update the metadata...
	for _, oldPort := range sp.Ports {
		if oldPort.Name == addedPort.Address {
			oldPort.SerialNumber = props.Get("serialNumber")
			oldPort.VendorID = vid
			oldPort.ProductID = pid
			return
		}
	}
	// ...otherwise, add it to the list
	sp.Ports = append(sp.Ports, &SpPortItem{
		Name:            addedPort.Address,
		SerialNumber:    props.Get("serialNumber"),
		VendorID:        vid,
		ProductID:       pid,
		Ver:             version,
		IsOpen:          false,
		IsPrimary:       false,
		Baud:            0,
		BufferAlgorithm: "",
	})
}

func (sp *serialPortList) remove(removedPort *discovery.Port) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()

	// Remove the port from the list
	sp.Ports = slices.DeleteFunc(sp.Ports, func(oldPort *SpPortItem) bool {
		return oldPort.Name == removedPort.Address
	})
}

func (sp *serialPortList) getPortByName(portname string) *SpPortItem {
	for _, port := range sp.Ports {
		if port.Name == portname {
			return port
		}
	}
	return nil
}
