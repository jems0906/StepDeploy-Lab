package logging

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	prefix string
}

// New creates a new logger with the given prefix
func New(prefix string) *Logger {
	levelStr := os.Getenv("LOG_LEVEL")
	level := INFO

	switch levelStr {
	case "debug":
		level = DEBUG
	case "info":
		level = INFO
	case "warn":
		level = WARN
	case "error":
		level = ERROR
	}

	return &Logger{
		level:  level,
		prefix: prefix,
	}
}

func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	levelName := "UNKNOWN"
	switch level {
	case DEBUG:
		levelName = "DEBUG"
	case INFO:
		levelName = "INFO"
	case WARN:
		levelName = "WARN"
	case ERROR:
		levelName = "ERROR"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] [%s]", timestamp, levelName, l.prefix)

	if len(args) > 0 {
		log.Printf("%s %s", prefix, fmt.Sprintf(msg, args...))
	} else {
		log.Printf("%s %s", prefix, msg)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}
