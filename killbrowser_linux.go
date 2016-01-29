package main

import (
	"log"
	"os/exec"
	"strings"
)

func findBrowser(process string) ([]byte, error) {
	ps := exec.Command("ps", "-A", "-o", "command")
	grep := exec.Command("grep", process)
	head := exec.Command("head", "-n", "1")

	log.Println("ps command:")
	log.Printf("%+v", ps)

	log.Println("grep command:")
	log.Printf("%+v", grep)

	log.Println("head command:")
	log.Printf("%+v", head)

	return pipe_commands(ps, grep, head)
}

func killBrowser(process string) ([]byte, error) {
	cmd := exec.Command("pkill", "-9", process)

	log.Println("kill command:")
	log.Printf("%+v", cmd)

	return cmd.Output()
}

func startBrowser(command []byte, url string) ([]byte, error) {
	parts := strings.Split(string(command), " ")
	cmd := exec.Command(parts[0], url)

	log.Println("start command:")
	log.Printf("%+v", cmd)

	return cmd.Output()
}
