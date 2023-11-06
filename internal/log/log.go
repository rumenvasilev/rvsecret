package log

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
)

// These are a consistent set of error codes instead of using random non-zero integers
const (
	FATAL     = 5
	ERROR     = 4
	WARN      = 3
	IMPORTANT = 2
	INFO      = 1
	DEBUG     = 0
)

// LogColors sets the color for each type of logging output
var LogColors = map[int]*color.Color{
	FATAL:     color.New(color.FgRed).Add(color.Bold),
	ERROR:     color.New(color.FgRed),
	WARN:      color.New(color.FgYellow),
	IMPORTANT: color.New(color.Bold),
	DEBUG:     color.New(color.FgCyan).Add(color.Faint),
}

// Logger holds specific configuration data for the logging
type logger struct {
	sync.Mutex

	debug  bool
	silent bool
}

// Log is the program's logging package. It is self-initialized (with init() func)
// so we don't need to do it manually in the client code. It's a dependency
// the program cannot function without.
// It's a global variable, because it's a small program, hence it's easier to
// manage it from a central place. And we don't mutate it's state more than once,
// only at startup, to set debug and silent modes.
var Log *logger

func init() {
	Log = new(logger)
}

// NewNoopLogger will mutate existing logger, ensuring no output will be sent to stdout
// useful for testing.
func NewNoopLogger() {
	Log.SetSilent(true)
}

func SetSilent(s bool) *logger {
	return Log.SetSilent(s)
}

// SetSilent will configure the logger to not display any realtime output to stdout
func (l *logger) SetSilent(s bool) *logger {
	l.silent = s
	return l
}

func SetDebug(d bool) *logger {
	return Log.SetDebug(d)
}

// SetDebug will configure the logger to enable debug output to be set to stdout
func (l *logger) SetDebug(d bool) *logger {
	l.debug = d
	return l
}

// Log is a generic printer for sending data to stdout. It does not do traditional syslog logging
func (l *logger) Log(level int, format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	if level == DEBUG && !l.debug {
		return
	} else if level < ERROR && l.silent {
		return
	}

	if c, ok := LogColors[level]; ok {
		_, _ = c.Printf(format, args...)
	} else {
		fmt.Printf(format, args...)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// Fatal prints a fatal level log message to stdout
func (l *logger) Fatal(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(FATAL, reformat, args...)
}

// Error prints an error level log message to stdout
func (l *logger) Error(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(ERROR, reformat, args...)
}

// Warn prints a warn level log message to stdout
func (l *logger) Warn(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(WARN, reformat, args...)
}

// Important prints an important level log message to stdout
func (l *logger) Important(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(IMPORTANT, reformat, args...)
}

// Info prints an info level log message to stdout
func (l *logger) Info(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(INFO, reformat, args...)
}

// Debug prints a debug level log message to stdout
func (l *logger) Debug(format string, args ...interface{}) {
	reformat := fmt.Sprintf("%s\n", format)
	l.Log(DEBUG, reformat, args...)
}
