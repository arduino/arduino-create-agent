package main

import (
	"os/exec"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
)

func findBrowser(process string) ([]byte, error) {
	ps := exec.Command("ps", "-A", "-o", "command")
	grep := exec.Command("grep", process)
	head := exec.Command("head", "-n", "1")

	return utilities.PipeCommands(ps, grep, head)
}

func killBrowser(process string) ([]byte, error) {
	cmd := exec.Command("pkill", "-9", process)
	return cmd.Output()
}

func startBrowser(command []byte, url string) ([]byte, error) {
	parts := strings.Split(string(command), " ")
	cmd := exec.Command(parts[0], url)
	return cmd.Output()
}
