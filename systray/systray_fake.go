// +build cli

package systray

import "os"

func (s *Systray) Start() {
	select {}
}

func (s *Systray) Quit() {
	os.Exit(0)
}
