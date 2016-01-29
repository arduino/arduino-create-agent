package main

import "os/exec"

func findBrowser(process string) ([]byte, error) {
	return []byte(process), nil
}

func killBrowser(process string) ([]byte, error) {
	cmd := exec.Command("Taskkill", "/F", "/IM", process+".exe")
	return cmd.Output()
}

func startBrowser(command []byte, url string) ([]byte, error) {
	cmd := exec.Command("cmd", "/C", "start", string(command), url)
	return cmd.Output()
}
