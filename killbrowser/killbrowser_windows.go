package browser

import "os/exec"

func Find(process string) ([]byte, error) {
	return []byte(process), nil
}

func Kill(process string) ([]byte, error) {
	cmd := exec.Command("Taskkill", "/F", "/IM", process+".exe")
	return cmd.Output()
}

func Start(command []byte, url string) ([]byte, error) {
	cmd := exec.Command("cmd", "/C", "start", string(command), url)
	return cmd.Output()
}
