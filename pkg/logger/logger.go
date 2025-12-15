package logger

import "fmt"

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(message string) {
	// Implementation for info level logging
	fmt.Println("love you bin")
}
