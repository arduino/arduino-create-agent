// +build cli

// Systray_fake gets compiled when the tag 'cli' is present. This is useful to build an agent without trayicon functionalities
package systray

import "os"

func (s *Systray) Start() {
	select {}
}

func (s *Systray) Quit() {
	os.Exit(0)
}
