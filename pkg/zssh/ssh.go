package zssh

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func RunCommandWithTimeout(c *SSHConfig, cmd string, timeout time.Duration) ([]byte, error) {
	pemBytes, err := os.ReadFile(findHomePath(c.IdentityFile))
	if err != nil {
		log.Fatalf("Failed to read private key file: %s", err)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %s", err)
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Create SSH client
	client, err := ssh.Dial("tcp", net.JoinHostPort(c.HostName, c.Port), config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
		return nil, err
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
		return nil, err
	}
	defer session.Close()

	//output, err := session.CombinedOutput(cmd)
	//if err != nil {
	//	log.Fatalf("Failed to run command on %s: %s", hostName, err)
	//	return nil, err
	//}
	//fmt.Printf("Output of command '%s' on %s:\n%s\n", cmd, hostName, output)

	// Set standard input/output and error
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Start command
	if err := session.Start(cmd); err != nil {
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	// Wait for command to finish or timeout
	done := make(chan error)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-time.After(timeout):
		session.Signal(ssh.SIGKILL) // Kill the command if timeout exceeded
		return nil, fmt.Errorf("command timed out after %v", timeout)
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("command failed: %v, output: %s, error: %s", err, stdoutBuf.String(), stderrBuf.String())
		}
		return stdoutBuf.Bytes(), nil
	}
}
