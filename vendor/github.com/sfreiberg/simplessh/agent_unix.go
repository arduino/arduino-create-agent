// +build !windows

package simplessh

import (
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Connect with a ssh agent with a custom timeout. If username is empty simplessh will attempt to get the current user.
func connectWithAgentTimeout(host, username string, timeout time.Duration) (*Client, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	authMethod := ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)

	return connect(username, host, authMethod, timeout)
}
