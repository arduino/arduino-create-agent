//go:build !windows

package updater

import (
	"os"
	"os/exec"
)

func elevate() error {
	args := append([]string{os.Args[0]}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}