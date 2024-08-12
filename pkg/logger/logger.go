package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(level slog.Leveler) *slog.Logger {
	return NewWithProjectDir(level, "zephyr/")
}

func NewWithProjectDir(level slog.Leveler, projectDir string) *slog.Logger {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// Remove time.
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			if index := strings.Index(source.File, projectDir); index != -1 {
				source.File = source.File[index+len(projectDir):]
			} else {
				source.File = filepath.Base(source.File)
			}
		}
		return a
	}
	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replace,
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &opts))
}

func NewWithRotated(level *slog.LevelVar, serviceName string) *slog.Logger {
	dir := "/var/log/" + serviceName
	if _, err1 := os.Stat(dir); os.IsNotExist(err1) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	w := &lumberjack.Logger{
		Filename: dir + "/json.log",
		MaxSize:  10,
		MaxAge:   60,
		//MaxBackups: 10,
		LocalTime: true,
		Compress:  false,
	}

	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}
	return slog.New(slog.NewJSONHandler(w, &opts))
}
