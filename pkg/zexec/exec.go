package zexec

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"sync"
)

var processRegistry = sync.Map{}

func StartProcess(uniqueID, command string, args ...string) error {
	os.Setenv("LD_LIBRARY_PATH", path.Join(os.Getenv("HOME"), "/ffmpeg_build/lib"))

	if _, exists := processRegistry.Load(uniqueID); exists {
		slog.Debug("Process already exists", "uniqueID", uniqueID)
		return nil
	}

	cmd := exec.Command(command, args...)
	fmt.Println(cmd)

	go func() {
		processRegistry.Store(uniqueID, 1)
		slog.Info("Process started", "uniqueID", uniqueID)
		cmd.Run()
		slog.Info("Process exited", "uniqueID", uniqueID)
		processRegistry.Delete(uniqueID)
	}()

	return nil
}
