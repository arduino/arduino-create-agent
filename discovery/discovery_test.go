package discovery_test

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/arduino/arduino-create-agent/discovery"
)

// TestUsage doesn't really test anything, since we don't have (yet) a way to reproduce hardware. It's useful to test by hand though
func TestUsage(t *testing.T) {
	monitor := discovery.New(time.Millisecond)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	monitor.Start(ctx)

	time.Sleep(10 * time.Second)

	fmt.Println(monitor.Serial())
	fmt.Println(monitor.Network())
}

// TestEvent doesn't really test anything, since we don't have (yet) a way to reproduce hardware. It's useful to test by hand though
func TestEvents(t *testing.T) {
	monitor := discovery.New(time.Millisecond)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	monitor.Start(ctx)

	for ev := range monitor.Events {
		fmt.Println(ev)
	}
}
