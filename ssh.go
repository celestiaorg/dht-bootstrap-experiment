package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/evan-forbes/devnet/config"
	"golang.org/x/crypto/ssh"
)

type SSHManager struct {
	Conns map[string]Connection
}

func NewSSHManager(drops map[string]config.Droplet, sshPass string) (*SSHManager, error) {
	conns := make(map[string]Connection)
	for name, drop := range drops {
		conn, err := NewConnection(drop, sshPass)
		if err != nil {
			return nil, err
		}
		conns[name] = conn
	}
	return &SSHManager{
		Conns: conns,
	}, nil
}

// CloseAll closes each established ssh session
func (s *SSHManager) CloseAll() {
	for name, c := range s.Conns {
		err := c.Close()
		if err != nil {
			log.Println(
				fmt.Errorf(
					"failure to close ssh session for %s: %w", name, err,
				),
			)
		}
	}
}

type Connection struct {
	client *ssh.Client
	drop   config.Droplet
	output *os.File
}

func NewConnection(drop config.Droplet, sshPass string) (Connection, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Connection{}, err
	}

	host, err := drop.Drop.PublicIPv4()
	if err != nil {
		return Connection{}, err
	}

	sshConfig, err := newSshClientConfig("root", host, 22, home+"/.ssh/id_rsa", sshPass)
	if err != nil {
		return Connection{}, err
	}

	// connect to the server via ssh
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, 22), sshConfig)
	if err != nil {
		return Connection{}, err
	}

	output, err := os.OpenFile(drop.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return Connection{}, err
	}

	return Connection{
		client: client,
		drop:   drop,
		output: output,
	}, nil
}

func (c Connection) DeliverPayload() error {
	ipv4, err := c.drop.Drop.PublicIPv4()
	if err != nil {
		return err
	}

	cmd := exec.Command("scp", "-r", c.drop.Payload, fmt.Sprintf("root@%s:/root/", ipv4))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failure to execute command: %w", err)
	}

	return nil
}

// Run runs a command on the ssh server and forwards the Stdout and
// Stderr of the server to the local client's output file
func (c Connection) Run(command string) error {
	sesh, err := c.NewSession()
	if err != nil {
		return err
	}
	defer sesh.Close()

	sesh.Stdout = c.output
	sesh.Stderr = c.output

	return sesh.Run(command)
}

func (c Connection) NewSession() (*ssh.Session, error) {
	return c.client.NewSession()
}

func (c Connection) Close() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	return c.output.Close()
}

func newSshClientConfig(user string, host string, port int, privateKeyPath string, privateKeyPassword string) (*ssh.ClientConfig, error) {
	// read private key file
	pemBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Reading private key file failed %v", err)
	}
	// create signer
	signer, err := signerFromPem(pemBytes, []byte(privateKeyPassword))
	if err != nil {
		return nil, err
	}
	// build SSH client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}

	return config, nil
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {
	// read pem block
	err := errors.New("Pem decode failed, no key found")
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, err
	}

	// handle encrypted key
	if x509.IsEncryptedPEMBlock(pemBlock) {
		// decrypt PEM
		pemBlock.Bytes, err = x509.DecryptPEMBlock(pemBlock, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("Decrypting PEM block failed %v", err)
		}

		// get RSA, EC or DSA key
		key, err := parsePemBlock(pemBlock)
		if err != nil {
			return nil, err
		}

		// generate signer instance from key
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, fmt.Errorf("Creating signer from encrypted key failed %v", err)
		}

		return signer, nil
	} else {
		// generate signer instance from plain key
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing plain private key failed %v", err)
		}

		return signer, nil
	}
}

func parsePemBlock(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing PKCS private key failed %v", err)
		} else {
			return key, nil
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing EC private key failed %v", err)
		} else {
			return key, nil
		}
	case "DSA PRIVATE KEY":
		key, err := ssh.ParseDSAPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing DSA private key failed %v", err)
		} else {
			return key, nil
		}
	default:
		return nil, fmt.Errorf("Parsing private key failed, unsupported key type %q", block.Type)
	}
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
