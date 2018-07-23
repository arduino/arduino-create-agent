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
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintln(tt.dirList), func(t *testing.T) {
			if got := findBaseDir(tt.dirList); got != tt.want {
				t.Errorf("findBaseDir() = %v, want %v", got, tt.want)
			}
		})
	}
}
