package browser

import (
	"os/exec"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
)

func Find(process string) ([]byte, error) {
	ps := exec.Command("ps", "-A", "-o", "command")
	grep := exec.Command("grep", process)
	head := exec.Command("head", "-n", "1")

	return utilities.PipeCommands(ps, grep, head)
}

func Kill(process string) ([]byte, error) {
	cmd := exec.Command("pkill", "-9", process)
	return cmd.Output()
}

func Start(command []byte, url string) ([]byte, error) {
	parts := strings.Split(string(command), " ")
	cmd := exec.Command(parts[0], url)
	return cmd.Output()
}
