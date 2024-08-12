package zexec

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	mutex           sync.Mutex
	processRegistry = make(map[string]int)
)

// StartProcess 启动一个新进程并将其与 uniqueID 关联
func StartProcess(uniqueID, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}

	mutex.Lock()
	processRegistry[uniqueID] = pgid
	mutex.Unlock()

	fmt.Println(cmd)
	slog.Info("StartProcess", "uniqueID", uniqueID, "pid", cmd.Process.Pid, "pgid", pgid)
	return nil
}

// StopProcess 停止与 uniqueID 关联的进程
func StopProcess(uniqueID, keyword string) error {
	mutex.Lock()
	pgid, exists := processRegistry[uniqueID]
	mutex.Unlock()

	if !exists {
		slog.Info("findProcessByKeyword", "uniqueID", uniqueID, "keyword", keyword)
		pid, err := findProcessByKeyword(keyword)
		if err != nil {
			return err
		}
		pgid = pid
	}

	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		return err
	}

	slog.Info("Process killed successfully", "uniqueID", uniqueID)

	mutex.Lock()
	delete(processRegistry, uniqueID)
	mutex.Unlock()

	return nil
}

// findProcessByKeyword 使用 ps 和 grep 命令查找包含关键字的进程
func findProcessByKeyword(keyword string) (int, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ps -eo pid,args | grep '%s' | grep -v grep", keyword))
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(&out)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		return pid, nil
	}

	return 0, fmt.Errorf("no process found with keyword: %s", keyword)
}
