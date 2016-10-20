package programmer

import (
	"path/filepath"
	"strings"
)

// differ returns the first item that differ between the two input slices
func differ(slice1 []string, slice2 []string) string {
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			return mKey
		}
	}

	return ""
}

// resolve replaces some symbols in the commandline with the appropriate values
func resolve(port, board, file, commandline string, extra Extra) string {
	commandline = strings.Replace(commandline, "{build.path}", filepath.ToSlash(filepath.Dir(file)), -1)
	commandline = strings.Replace(commandline, "{build.project_name}", strings.TrimSuffix(filepath.Base(file), filepath.Ext(filepath.Base(file))), -1)
	commandline = strings.Replace(commandline, "{serial.port}", port, -1)
	commandline = strings.Replace(commandline, "{serial.port.file}", filepath.Base(port), -1)

	if extra.Verbose == true {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsVerbose, -1)
	} else {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsQuiet, -1)
	}
	return commandline
}
