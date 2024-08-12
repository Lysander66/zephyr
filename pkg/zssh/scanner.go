package zssh

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	ProxyJump    string //since OpenSSH 7.3
	ProxyCommand string //ssh -W %h:%p host1
}

func LoadSSHConfig(name string) []*SSHConfig {
	file, err := os.Open(findHomePath(name))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var configs []*SSHConfig
	var current *SSHConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			// Ignore empty or comment lines
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "Host":
			// Start a new config block
			current = &SSHConfig{Host: parts[1]}
			if current.Host == "*" {
				// Skip this block if Host is *
				current = nil
			} else {
				configs = append(configs, current)
			}
		case "HostName":
			current.HostName = parts[1]
		case "User":
			current.User = parts[1]
		case "Port":
			fmt.Sscan(parts[1], &current.Port)
		case "IdentityFile":
			current.IdentityFile = parts[1]
		case "ProxyJump":
			current.ProxyJump = parts[1]
		default:
			// Ignore unknown options
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return configs
}

func findHomePath(s string) string {
	if strings.HasPrefix(s, "~/") {
		return strings.ReplaceAll(s, "~", os.Getenv("HOME"))
	}
	return s
}
