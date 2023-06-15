package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Lysander66/zephyr/internal/service"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Hosts    []string
	Commands []string
}

func readConfig(name string) *Config {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	config := &Config{}
	scanner := bufio.NewScanner(file)
	isCommand := false // 标记是否为Commands部分的标志位
	for scanner.Scan() {
		line := scanner.Text()
		if line == "// Command" {
			isCommand = true
		}
		if line == "" || line[0] == '/' {
			continue
		}
		if isCommand {
			config.Commands = append(config.Commands, line)
		} else {
			config.Hosts = append(config.Hosts, line)
		}
	}

	return config
}

func run(m map[string]*service.SSHConfig, host, command string, timeout time.Duration) (string, error) {
	sshConfig, ok := m[host]
	if !ok {
		return "", fmt.Errorf("host not found: %s", host)
	}
	output, err := service.RunCommandWithTimeout(sshConfig, replaceVars(command, host), timeout)
	if err != nil {
		return "", fmt.Errorf("%s: %v", host, err)
	}
	return fmt.Sprintf("%s:\n%s\n", host, output), nil
}

func main() {
	var lFlag bool
	flag.BoolVar(&lFlag, "l", false, "run on local machine")
	commandFile := flag.String("c", "cmd.txt", "command file path")
	sshConfigFile := flag.String("s", "~/.ssh/config", "ssh config file path")
	numParallel := flag.Int("p", 10, "number of parallel hosts to process")
	timeoutSeconds := flag.Int("t", 30, "timeout in seconds for each command")
	flag.Parse()

	config := readConfig(*commandFile)
	log.Println("start...")

	// local
	if lFlag {
		runLocal(config, *numParallel)
		return
	}

	command := strings.Join(config.Commands, " && ")
	sshConfigs := service.LoadSSHConfig(*sshConfigFile)
	m := make(map[string]*service.SSHConfig)
	for _, v := range sshConfigs {
		m[v.Host] = v
	}

	wg := sync.WaitGroup{}
	semaphoreChan := make(chan struct{}, *numParallel)

	for _, host := range config.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			semaphoreChan <- struct{}{}

			output, err := run(m, host, command, time.Duration(*timeoutSeconds)*time.Second)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			} else {
				fmt.Printf("%s:\n%s\n", host, output)
			}

			<-semaphoreChan
		}(host)
	}

	wg.Wait()
}

func runLocal(config *Config, numParallel int) {
	wg := sync.WaitGroup{}
	semaphoreChan := make(chan struct{}, numParallel)

	command := strings.Join(config.Commands, ";")
	for _, host := range config.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			semaphoreChan <- struct{}{}

			cmd := exec.Command("/bin/sh", "-c", replaceVars(command, host))
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("[ERROR] %v", err)
			} else {
				fmt.Print(output)
			}

			<-semaphoreChan
		}(host)
	}

	wg.Wait()
}

func replaceVars(command, host string) string {
	return strings.ReplaceAll(command, "${HOST}", host)
}
