package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// LogEntry is one structured log line written to file as JSON.
type LogEntry struct {
	ClientIP  string `json:"client_ip"`
	Method    string `json:"method"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
	Decision  string `json:"decision"`
	Status    int    `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

// Logger writes newline-delimited JSON to a file, thread-safe.
type Logger struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
}

func NewLogger(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &Logger{file: f, enc: json.NewEncoder(f)}, nil
}

func (l *Logger) Log(meta RequestMeta, decision string, status int, reason string) {
	entry := LogEntry{
		ClientIP:  meta.ClientIP,
		Method:    meta.Method,
		Host:      meta.Host,
		Path:      meta.Path,
		Timestamp: meta.Timestamp,
		Decision:  decision,
		Status:    status,
		Reason:    reason,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enc.Encode(entry)
}

func (l *Logger) Close() error { return l.file.Close() }
