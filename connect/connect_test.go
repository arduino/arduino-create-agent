package connect_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/arduino/arduino-create-agent/connect"
	"golang.org/x/net/context"
)

func TestUsage(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	input, output, err := connect.Open(ctx, "/dev/ttyACM0", 9600)
	if err != nil {
		t.Fatalf(err.Error())
	}

	go func() {
		for msg := range output {
			fmt.Print(string(msg))
		}
	}()

	input <- []byte("some message")

	time.Sleep(11 * time.Second)
}
