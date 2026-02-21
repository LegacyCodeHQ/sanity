//go:build dev

package mcplogdlog

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const defaultSocket = "/tmp/mcplogd.sock"
const appName = "clarity"

const (
	levelInfo  = "info"
	levelDebug = "debug"
	levelWarn  = "warn"
	levelError = "error"
)

type entry struct {
	App       string         `json:"app"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func Info(message string, metadata map[string]any) {
	log(levelInfo, message, metadata)
}

func Debug(message string, metadata map[string]any) {
	log(levelDebug, message, metadata)
}

func Warn(message string, metadata map[string]any) {
	log(levelWarn, message, metadata)
}

func Error(message string, metadata map[string]any) {
	log(levelError, message, metadata)
}

func log(level, message string, metadata map[string]any) {
	conn, err := net.Dial("unix", defaultSocket)
	if err != nil {
		return
	}
	defer conn.Close()

	e := entry{
		App:       appName,
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Metadata:  metadata,
	}
	data, _ := json.Marshal(e)
	fmt.Fprintf(conn, "%s\n", data)
}
