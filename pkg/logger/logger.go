package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Logger defines the logging interface
type Logger interface {
	Debug(message string, fields map[string]any)
	Info(message string, fields map[string]any)
	Warn(message string, fields map[string]any)
	Error(message string, fields map[string]any)
}

// Level represents the logging level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// logger implements structured logging
type logger struct {
	level  Level
	format string
	output *log.Logger
}

// New creates a new logger
func New(config Config) Logger {
	return &logger{
		level:  Level(config.Level),
		format: config.Format,
		output: log.New(os.Stdout, "", 0),
	}
}

// Debug logs a debug message
func (l *logger) Debug(message string, fields map[string]any) {
	if !l.shouldLog(LevelDebug) {
		return
	}
	l.log(LevelDebug, message, fields)
}

// Info logs an info message
func (l *logger) Info(message string, fields map[string]any) {
	if !l.shouldLog(LevelInfo) {
		return
	}
	l.log(LevelInfo, message, fields)
}

// Warn logs a warning message
func (l *logger) Warn(message string, fields map[string]any) {
	if !l.shouldLog(LevelWarn) {
		return
	}
	l.log(LevelWarn, message, fields)
}

// Error logs an error message
func (l *logger) Error(message string, fields map[string]any) {
	if !l.shouldLog(LevelError) {
		return
	}
	l.log(LevelError, message, fields)
}

// log handles the actual logging
func (l *logger) log(level Level, message string, fields map[string]any) {
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	for k, v := range fields {
		entry[k] = v
	}

	if l.format == "json" {
		data, _ := json.Marshal(entry)
		l.output.Println(string(data))
	} else {
		// Text format
		output := fmt.Sprintf("[%s] %s: %s", entry["timestamp"], level, message)
		if len(fields) > 0 {
			fieldsJSON, _ := json.Marshal(fields)
			output += " " + string(fieldsJSON)
		}
		l.output.Println(output)
	}
}

// shouldLog checks if the message should be logged based on level
func (l *logger) shouldLog(level Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return levels[level] >= levels[l.level]
}
