// Package logger provides structured JSON logging for all QuantLab services.
package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Level represents a log severity level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Entry is a single structured log entry.
type Entry struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Service   string `json:"service"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// Logger provides structured JSON logging.
type Logger struct {
	service string
	logger  *log.Logger
	level   Level
}

// New creates a new Logger for the given service name.
func New(service string) *Logger {
	return &Logger{
		service: service,
		logger:  log.New(os.Stdout, "", 0),
		level:   DEBUG,
	}
}

// WithLevel sets the minimum log level.
func (l *Logger) WithLevel(level Level) *Logger {
	l.level = level
	return l
}

func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}
	entry := Entry{
		Level:     level.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   msg,
		Service:   l.service,
		Extra:     fields,
	}
	data, _ := json.Marshal(entry)
	l.logger.Println(string(data))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.log(FATAL, msg, fields)
	os.Exit(1)
}
