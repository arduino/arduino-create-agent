package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
)

var (
	serialListWindowsWg sync.WaitGroup
)

func associateVidPidWithPort(ports []OsSerialPort) []OsSerialPort {
	ports, _ = getList()
	return ports
}

func getList() ([]OsSerialPort, os.SyscallError) {
	// use a queue to do this to avoid conflicts
	// we've been getting crashes when this getList is requested
	// too many times too fast. i think it's something to do with
	// the unsafe syscalls overwriting memory

	// this will only block if waitgroupctr > 0. so first time
	// in shouldn't block
	serialListWindowsWg.Wait()

	serialListWindowsWg.Add(1)
	arr, sysCallErr := getListViaWmiPnpEntity()
	serialListWindowsWg.Done()
	//arr = make([]OsSerialPort, 0)

	// see if array has any data, if not fallback to the traditional
	// com port list model
	/*
		if len(arr) == 0 {
			// assume it failed
			arr, sysCallErr = getListViaOpen()
		}
	*/

	// see if array has any data, if not fallback to looking at
	// the registry list
	/*
		arr = make([]OsSerialPort, 0)
		if len(arr) == 0 {
			// assume it failed
			arr, sysCallErr = getListViaRegistry()
		}
	*/

	return arr, sysCallErr
}

func getListSynchronously() {

}

func getListViaWmiPnpEntity() ([]OsSerialPort, os.SyscallError) {

	//log.Println("Doing getListViaWmiPnpEntity()")

	// this method panics a lot and i'm not sure why, just catch
	// the panic and return empty list
	defer func() {
		if e := recover(); e != nil {
			// e is the interface{} typed-value we passed to panic()
			log.Println("Got panic: ", e) // Prints "Whoops: boom!"
		}
	}()

	var err os.SyscallError

	//var friendlyName string

	// init COM, oh yeah
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// result is a SWBemObjectSet
	//pname := syscall.StringToUTF16("SELECT * FROM Win32_PnPEntity where Name like '%" + "COM35" + "%'")
	pname := "SELECT * FROM Win32_PnPEntity WHERE ConfigManagerErrorCode = 0 and Name like '%(COM%'"
	//pname := "SELECT * FROM Win32_PnPEntity WHERE ConfigManagerErrorCode = 0"
	resultRaw, err2 := oleutil.CallMethod(service, "ExecQuery", pname)
	//log.Println("Got result from oleutil.CallMethod")
	if err2 != nil {
		// we got back an error or empty list
		log.Printf("Got an error back from oleutil.CallMethod. err:%v", err2)
		return nil, err
	}

	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)

	list := make([]OsSerialPort, count)

	for i := 0; i < count; i++ {

		// items we're looping thru look like below and
		// thus we can query for any of these names
		/*
					__GENUS                     : 2
			__CLASS                     : Win32_PnPEntity
			__SUPERCLASS                : CIM_LogicalDevice
			__DYNASTY                   : CIM_ManagedSystemElement
			__RELPATH                   : Win32_PnPEntity.DeviceID="USB\\VID_1D50&PID_606D&MI_02\\6&2F09EA14&0&0002"
			__PROPERTY_COUNT            : 24
			__DERIVATION                : {CIM_LogicalDevice, CIM_LogicalElement, CIM_ManagedSystemElement}
			__SERVER                    : JOHN-ATIV
			__NAMESPACE                 : root\cimv2
			__PATH                      : \\JOHN-ATIV\root\cimv2:Win32_PnPEntity.DeviceID="USB\\VID_1D50&PID_606D&MI_02\\6&2F09EA14
			                              &0&0002"
			Availability                :
			Caption                     : TinyG v2 (Data Channel) (COM12)
			ClassGuid                   : {4d36e978-e325-11ce-bfc1-08002be10318}
			CompatibleID                : {USB\Class_02&SubClass_02&Prot_01, USB\Class_02&SubClass_02, USB\Class_02}
			ConfigManagerErrorCode      : 0
			ConfigManagerUserConfig     : False
			CreationClassName           : Win32_PnPEntity
			Description                 : TinyG v2 (Data Channel)
			DeviceID                    : USB\VID_1D50&PID_606D&MI_02\6&2F09EA14&0&0002
			ErrorCleared                :
			ErrorDescription            :
			HardwareID                  : {USB\VID_1D50&PID_606D&REV_0097&MI_02, USB\VID_1D50&PID_606D&MI_02}
			InstallDate                 :
			LastErrorCode               :
			Manufacturer                : Synthetos (www.synthetos.com)
			Name                        : TinyG v2 (Data Channel) (COM12)
			PNPDeviceID                 : USB\VID_1D50&PID_606D&MI_02\6&2F09EA14&0&0002
			PowerManagementCapabilities :
			PowerManagementSupported    :
			Service                     : usbser
			Status                      : OK
			StatusInfo                  :
			SystemCreationClassName     : Win32_ComputerSystem
			SystemName                  : JOHN-ATIV
			PSComputerName              : JOHN-ATIV
		*/

		// item is a SWbemObject, but really a Win32_Process
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", i)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		asString, _ := oleutil.GetProperty(item, "Name")

		//log.Println(asString.ToString())

		// get the com port
		//if false {
		s := strings.Split(asString.ToString(), "(COM")[1]
		s = "COM" + s
		s = strings.Split(s, ")")[0]
		list[i].Name = s
		//}

		// get the deviceid so we can figure out related ports
		// it will look similar to
		// USB\VID_1D50&PID_606D&MI_00\6&2F09EA14&0&0000
		deviceIdStr, _ := oleutil.GetProperty(item, "DeviceID")
		devIdItems := strings.Split(deviceIdStr.ToString(), "&")
		log.Printf("DeviceId elements:%v", devIdItems)
		if len(devIdItems) > 3 {
			list[i].SerialNumber = devIdItems[3]
			list[i].IdProduct = strings.Replace(devIdItems[1], "PID_", "", 1)
			list[i].IdVendor = strings.Replace(devIdItems[0], "USB\\VID_", "", 1)
		} else {
			list[i].SerialNumber = deviceIdStr.ToString()
			pidMatch := regexp.MustCompile("PID_(\\d+)").FindAllStringSubmatch(deviceIdStr.ToString(), -1)
			if len(pidMatch) > 0 {
				if len(pidMatch[0]) > 1 {
					list[i].IdProduct = pidMatch[0][1]
				}
			}
			vidMatch := regexp.MustCompile("VID_(\\d+)").FindAllStringSubmatch(deviceIdStr.ToString(), -1)
			if len(vidMatch) > 0 {
				if len(vidMatch[0]) > 1 {
					list[i].IdVendor = vidMatch[0][1]
				}
			}
		}

		list[i].IdVendor = "0x" + list[i].IdVendor
		list[i].IdProduct = "0x" + list[i].IdProduct

		manufStr, _ := oleutil.GetProperty(item, "Manufacturer")
		list[i].Manufacturer = manufStr.ToString()
		descStr, _ := oleutil.GetProperty(item, "Description")
		list[i].Product = descStr.ToString()
		//classStr, _ := oleutil.GetProperty(item, "CreationClassName")
		//list[i].DeviceClass = classStr.ToString()

	}

	return list, err
}

func convertByteArrayToUint16Array(b []byte, mylen uint32) []uint16 {

	log.Println("converting. len:", mylen)
	var i uint32
	ret := make([]uint16, mylen/2)
	for i = 0; i < mylen; i += 2 {
		//ret[i/2] = binary.LittleEndian.Uint16(b[i : i+1])
		ret[i/2] = uint16(b[i]) | uint16(b[i+1])<<8
	}
	return ret
}

func tellCommandNotToSpawnShell(oscmd *exec.Cmd) {
	oscmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
