package v2

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogAdapter struct {
	*logrus.Logger
}

func (a LogAdapter) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
	}
	var fm bytes.Buffer
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
	return nil
}
