// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package systray

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/arduino/go-paths-helper"
	log "github.com/sirupsen/logrus"
)

// Systray manages the systray icon with its menu and actions. It also handles the pause/resume behaviour of the agent
type Systray struct {
	// Whether the Agent is in Pause mode
	Hibernate bool
	// The version of the Agent, displayed in the trayicon menu
	Version string
	// The url of the debug page. It's a function because it could change port
	DebugURL func() string
	// The active configuration file
	AdditionalConfig string
	// The path of the exe (only used in update)
	path string
	// The path of the configuration file
	configPath *paths.Path
}

// Restart restarts the program
// it works by finding the executable path and launching it before quitting
func (s *Systray) Restart() {

	if s.path == "" {
		log.Println("Update binary path not set")
		var err error
		s.path, err = os.Executable()
		if err != nil {
			log.Printf("Error getting exe path using os lib. err: %v\n", err)
		}
	} else {
		log.Println("Starting updated binary: ", s.path)
	}

	// Trim newlines (needed on osx)
	s.path = strings.Trim(s.path, "\n")

	// Build args
	args := []string{"-ls", fmt.Sprintf("--hibernate=%v", s.Hibernate)}

	if s.AdditionalConfig != "" {
		args = append(args, fmt.Sprintf("--additional-config=%s", s.AdditionalConfig))
	}

	// Launch executable
	cmd := exec.Command(s.path, args...)
	err := cmd.Start()
	if err != nil {
		log.Printf("Error restarting process: %v\n", err)
		return
	}

	// If everything was fine, quit
	s.Quit()
}

// Pause restarts the program with the hibernate flag set to true
func (s *Systray) Pause() {
	s.Hibernate = true
	s.Restart()
}

// Resume restarts the program with the hibernate flag set to false
func (s *Systray) Resume() {
	s.Hibernate = false
	s.Restart()
}

// RestartWith restarts the program with the given path
func (s *Systray) RestartWith(path string) {
	s.path = path
	s.Restart()
}

// SetConfig allows to specify the path of the configuration the agent is using.
// The tray menu with this info can display an "open config file" option
func (s *Systray) SetConfig(configPath *paths.Path) {
	s.configPath = configPath
}
