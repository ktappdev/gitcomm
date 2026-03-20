package diag

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	stateMu    sync.RWMutex
	state      = loggerState{logger: log.New(ioDiscard{}, "", 0)}
	bearerExpr = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._\-]+`)
	keyExpr    = regexp.MustCompile(`(?i)(open[_-]?router[_-]?api[_-]?key\s*[=:]\s*)([^\s,;]+)`)
)

type loggerState struct {
	logger *log.Logger
	path   string
	debug  bool
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func DefaultPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".gitcomm", "logs", "diagnostics.log"), nil
}

func Init(debug bool) (string, error) {
	path, err := DefaultPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}

	stateMu.Lock()
	state = loggerState{logger: log.New(file, "", 0), path: path, debug: debug}
	stateMu.Unlock()

	Info("diag", "diagnostics initialized", "path", path, "debug", debug)
	return path, nil
}

func Path() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return state.path
}

func DebugEnabled() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return state.debug
}

func Debug(component, msg string, kv ...any) {
	if !DebugEnabled() {
		return
	}
	write("DEBUG", component, msg, kv...)
}

func Info(component, msg string, kv ...any)  { write("INFO", component, msg, kv...) }
func Warn(component, msg string, kv ...any)  { write("WARN", component, msg, kv...) }
func Error(component, msg string, kv ...any) { write("ERROR", component, msg, kv...) }

func write(level, component, msg string, kv ...any) {
	stateMu.RLock()
	logger := state.logger
	stateMu.RUnlock()

	fields := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprintf("field_%d", i)
		if i+1 < len(kv) {
			key = fmt.Sprintf("%v", kv[i])
		}
		val := "<missing>"
		if i+1 < len(kv) {
			val = fmt.Sprintf("%v", kv[i+1])
		}
		fields = append(fields, fmt.Sprintf("%s=%q", sanitize(key), sanitize(val)))
	}

	line := fmt.Sprintf("%s level=%s component=%s msg=%q", time.Now().Format(time.RFC3339), level, sanitize(component), sanitize(msg))
	if len(fields) > 0 {
		line += " " + strings.Join(fields, " ")
	}
	logger.Println(line)
}

func Sanitize(value string) string {
	return sanitize(value)
}

func sanitize(value string) string {
	value = bearerExpr.ReplaceAllString(value, "Bearer [REDACTED]")
	value = keyExpr.ReplaceAllString(value, `${1}[REDACTED]`)
	return value
}

func Snippet(value string, max int) string {
	value = sanitize(strings.TrimSpace(value))
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "... (truncated)"
}
