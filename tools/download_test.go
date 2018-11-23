package tools

import (
	"fmt"
	"testing"
)

func Test_findBaseDir(t *testing.T) {
	cases := []struct {
		dirList []string
		want    string
	}{
		{[]string{"bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac", "example"}, ""},
		{[]string{"avrdude/bin/avrdude", "avrdude/etc/avrdude.conf"}, "avrdude/"},
		{[]string{"pax_global_header","bin/", "bin/bossac"}, "bin/"},

	}
	for _, tt := range cases {
		t.Run(fmt.Sprintln(tt.dirList), func(t *testing.T) {
			if got := findBaseDir(tt.dirList); got != tt.want {
				t.Errorf("findBaseDir() = got %v, want %v", got, tt.want)
			}
		})
	}
}
