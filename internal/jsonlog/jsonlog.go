package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "Error"
	case LevelInfo:
		return "Info"
	case LevelFatal:
		return "Fatal"
	default:
		return ""
	}
}

type Logger struct {
	out      []io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out []io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}
func (l *Logger) print(level Level, message string, prop map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: prop,
	}
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}
	var line []byte
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": Unable to marshal log:" + err.Error())
	}
	// Lock the mutex so that no two writes to the output destination cannot happen
	// concurrently. If we don't do this, it's possible that the text for two or more
	// log entries will be intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, writer := range l.out {
		writer.Write(append(line, '\n'))
	}
	return 0, nil
}

func (l *Logger) PrintError(err error, prop map[string]string) {
	l.print(LevelError, err.Error(), prop)
}
func (l *Logger) PrintInfo(msg string, prop map[string]string) {
	l.print(LevelInfo, msg, prop)
}
func (l *Logger) PrintFatal(err error, prop map[string]string) {
	l.print(LevelFatal, err.Error(), prop)
	os.Exit(1)
}

// We also implement a Write() method on our Logger type so that it satisfies the
// io.Writer interface. This writes a log entry at the ERROR level with no additional
// properties.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
