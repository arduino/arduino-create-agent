package tools

import (
	"os/exec"
)

func hideFile(path string) {

}

func TellCommandNotToSpawnShell(_ *exec.Cmd) {
}

func MessageBox(title, text string) int {
	return 6
}
