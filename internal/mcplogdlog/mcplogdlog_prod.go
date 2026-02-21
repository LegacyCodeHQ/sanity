//go:build !dev

package mcplogdlog

func Info(message string, metadata map[string]any) {
	_ = message
	_ = metadata
}

func Debug(message string, metadata map[string]any) {
	_ = message
	_ = metadata
}

func Warn(message string, metadata map[string]any) {
	_ = message
	_ = metadata
}

func Error(message string, metadata map[string]any) {
	_ = message
	_ = metadata
}
