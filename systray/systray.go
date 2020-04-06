package systray

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kardianos/osext"
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
}

// Restart restarts the program
// it works by finding the executable path and launching it before quitting
func (s *Systray) Restart() {

	fmt.Println(s.path)
	fmt.Println(osext.Executable())
	if s.path == "" {
		var err error
		s.path, err = osext.Executable()
		if err != nil {
			fmt.Printf("Error getting exe path using osext lib. err: %v\n", err)
		}
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
		fmt.Printf("Error restarting process: %v\n", err)
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

// Pause restarts the program with the hibernate flag set to false
func (s *Systray) Resume() {
	s.Hibernate = false
	s.Restart()
}

// Update restarts the program with the given path
func (s *Systray) Update(path string) {
	s.path = path
	s.Restart()
}
