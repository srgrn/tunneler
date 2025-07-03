package tunnel

import (
	"net"

	"golang.org/x/crypto/ssh"
)

// SSHTunnel represents a tunnel over an SSH connection.
type SSHTunnel struct {
	client *ssh.Client
}

// NewSSHTunnel creates a new SSH client from an existing network connection.
func NewSSHTunnel(conn net.Conn, user string, privateKey ssh.Signer) (*SSHTunnel, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	c, _, _, err := ssh.NewClientConn(conn, "", config)
	if err != nil {
		return nil, err
	}

	return &SSHTunnel{
		client: ssh.NewClient(c, nil, nil),
	}, nil
}

// Dial opens a connection to the remote address through the SSH tunnel.
func (t *SSHTunnel) Dial(network, address string) (net.Conn, error) {
	return t.client.Dial(network, address)
}

// Close closes the underlying SSH client and network connection.
func (t *SSHTunnel) Close() error {
	return t.client.Close()
}
