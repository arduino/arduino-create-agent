package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
)

func associateVidPidWithPort(ports []OsSerialPort) []OsSerialPort {

	for index, _ := range ports {
		ueventcmd := exec.Command("cat", "/sys/class/tty/"+filepath.Base(ports[index].Name)+"/device/uevent")
		grep3cmd := exec.Command("grep", "PRODUCT=")

		cmdOutput2, _ := utilities.PipeCommands(ueventcmd, grep3cmd)
		cmdOutput2S := string(cmdOutput2)

		if len(cmdOutput2S) == 0 {
			continue
		}

		infos := strings.Split(cmdOutput2S, "=")

		vid_pid := strings.Split(infos[1], "/")

		vid, _ := strconv.ParseInt(vid_pid[0], 16, 32)
		pid, _ := strconv.ParseInt(vid_pid[1], 16, 32)
		ports[index].IdVendor = fmt.Sprintf("0x%04x", vid)
		ports[index].IdProduct = fmt.Sprintf("0x%04x", pid)
	}
	return ports
}

func hideFile(path string) {
}

func tellCommandNotToSpawnShell(_ *exec.Cmd) {
}
