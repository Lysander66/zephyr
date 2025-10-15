package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type ConsoleHandler struct {
	level      slog.Leveler
	projectDir string
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	// 时间
	buf.WriteString(r.Time.Format(time.RFC3339))
	buf.WriteString(" ")

	// 级别
	switch r.Level {
	case slog.LevelDebug:
		buf.WriteString("DBG")
	case slog.LevelInfo:
		buf.WriteString("INF")
	case slog.LevelWarn:
		buf.WriteString("WRN")
	case slog.LevelError:
		buf.WriteString("ERR")
	default:
		buf.WriteString(r.Level.String())
	}
	buf.WriteString(" ")

	// 源文件信息
	if r.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		file := frame.File

		// 简化路径 - 保留最后3级目录
		parts := strings.Split(file, "/")
		if len(parts) >= 3 {
			file = strings.Join(parts[len(parts)-3:], "/")
		} else {
			file = filepath.Base(file)
		}

		buf.WriteString(file)
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%d", frame.Line))
		buf.WriteString(" > ")
	}

	// 消息
	buf.WriteString(r.Message)

	// 属性
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		buf.WriteString(a.Key)
		buf.WriteString("=")
		buf.WriteString(formatValue(a.Value))
		return true
	})

	buf.WriteString("\n")

	_, err := os.Stdout.WriteString(buf.String())
	return err
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return h
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindFloat64:
		return fmt.Sprintf("%.2f", v.Float64())
	case slog.KindBool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case slog.KindAny:
		return fmt.Sprintf("%v", v.Any())
	default:
		return v.String()
	}
}

// NewConsole 创建控制台日志器
func NewConsole(level slog.Leveler) *slog.Logger {
	handler := &ConsoleHandler{
		level: level,
	}
	return slog.New(handler)
}

// NewFile 创建文件日志器，简化source信息
func NewFile(level slog.Leveler, serviceName string) *slog.Logger {
	logDir := "/var/log/" + serviceName
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	writer := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    100, // MB
		MaxAge:     30,  // days
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   true,
	}

	replace := func(groups []string, a slog.Attr) slog.Attr {
		// 简化source信息
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			file := source.File

			// 简化文件路径 - 保留最后3级目录
			parts := strings.Split(file, "/")
			if len(parts) >= 3 {
				file = strings.Join(parts[len(parts)-3:], "/")
			} else {
				file = filepath.Base(file)
			}

			// 只保留文件路径和行号
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", file, source.Line))
		}
		return a
	}

	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replace,
	}

	return slog.New(slog.NewJSONHandler(writer, &opts))
}

// New 默认创建控制台日志器
func New(level slog.Leveler) *slog.Logger {
	return NewConsole(level)
}

// NewDual 根据环境创建合适的日志器
func NewDual(level slog.Leveler, serviceName string) *slog.Logger {
	if os.Getenv("ENV") == "production" {
		return NewFile(level, serviceName)
	}
	return NewConsole(level)
}
