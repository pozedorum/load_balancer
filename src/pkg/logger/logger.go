package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	*log.Logger
	file *os.File
}

// New создает новый логгер с записью в файл
func New(name, serverID, port string) (*Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", name, port))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	l := log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	fmt.Fprintf(file, "\n\n=== Server %s started at %s ===\n", serverID, time.Now().Format(time.RFC3339))

	return &Logger{
		Logger: l,
		file:   file,
	}, nil
}

// Close закрывает файл лога
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetGlobal устанавливает этот логгер как глобальный
func (l *Logger) SetGlobal() {
	log.SetOutput(l.file)
}
