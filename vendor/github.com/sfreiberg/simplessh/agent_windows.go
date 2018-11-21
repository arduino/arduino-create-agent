package simplessh

import (
	"fmt"
	"time"

	"github.com/davidmz/go-pageant"
	"golang.org/x/crypto/ssh"
)

// Connect with a ssh agent with a custom timeout. If username is empty simplessh will attempt to get the current user.
func connectWithAgentTimeout(host, username string, timeout time.Duration) (*Client, error) {
	if !pageant.Available() {
		return nil, fmt.Errorf("Pageant is unavailable")
	}

	authMethod := ssh.PublicKeysCallback(pageant.New().Signers)
	return connect(username, host, authMethod, timeout)
}
