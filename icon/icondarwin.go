//go:build darwin

package icon

import (
	_ "embed" // import embed to embed the icon
	"os/exec"
	"strings"
)

// isDarkMode will return if the system is in darkmode
func isDarkMode() bool {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, _ := cmd.Output()
	return strings.Contains(string(output), "Dark")
}

// GetIcon will return the icon
func GetIcon() []byte {
	if isDarkMode() {
		return data
	}
	return dataLight
}

// GetIconHiber will return the hibernated icon
func GetIconHiber() []byte {
	if isDarkMode() {
		return dataDarkHibernate
	}
	return dataLightHibernate
}

// dataLight represents the icon
//
//go:embed icon_mac_light.png
var dataLight []byte

// dataLightHibernate represents the light icon hibernated
//
//go:embed icon_mac_light_hiber.png
var dataLightHibernate []byte

// data represents the icon
//
//go:embed icon_mac.png
var data []byte

// dataDarkHibernate represents the dark icon hibernated
//
//go:embed icon_mac_hiber.png
var dataDarkHibernate []byte
