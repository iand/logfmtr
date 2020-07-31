package logfmtr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
)

// The global verbosity level.
var gv int32 = 0

// SetVerbosity sets the global log level. Only loggers with a V level less than
// or equal to this value will be enabled.
func SetVerbosity(v int) int {
	old := atomic.SwapInt32(&gv, int32(v))
	return int(old)
}

// New returns a logger that writes in logfmt using the default options.
func New() *Logger {
	return NewWithOptions(DefaultOptions())
}

// New returns a logger that writes in logfmt using the supplied options. Panics if
// no writer is supplied in the options.
func NewWithOptions(opts Options) *Logger {
	if opts.Writer == nil {
		panic("logger was supplied with nil writer")
	}
	l := &Logger{
		w:         opts.Writer,
		humanize:  opts.Humanize,
		tsFormat:  opts.TimestampFormat,
		nameDelim: opts.NameDelim,
		colorize:  opts.Colorize && opts.Humanize,
	}
	return l
}

// DefaultOptions returns the default options used by New. Override the option fields
// to customise behaviour.
func DefaultOptions() Options {
	return Options{
		Writer:          os.Stdout,
		TimestampFormat: time.RFC3339Nano,
		NameDelim:       ".",
	}
}

// Options contains fields and flags for customizing logger behaviour
type Options struct {
	// Writer is where logs will be written to
	Writer io.Writer

	// Humanize changes the log output to a human friendly format
	Humanize bool

	// Colorize adds color to the log output. Only applies if Humanize is also true.
	Colorize bool

	// TimestampFormat sets the format for log timestamps. Set to empty to disable timestamping
	// of log messages. Humanize uses a fixed short timestamp format.
	TimestampFormat string

	// NameDelim is the delimiter character used when appending names of loggers.
	NameDelim string
}

var _ logr.Logger = (*Logger)(nil)

// Logger is a logger that writes messages in the logfmt style.
// See https://www.brandur.org/logfmt for more information.
type Logger struct {
	w         io.Writer
	level     int
	name      string
	values    string
	humanize  bool
	tsFormat  string
	nameDelim string
	colorize  bool
}

// Enabled repoorts whether this Logger is enabled with respect to the current global log level.
func (l *Logger) Enabled() bool {
	return l.level <= int(atomic.LoadInt32(&gv))
}

// Info logs a non-error message with the given key/value pairs as context.
func (l *Logger) Info(msg string, kvs ...interface{}) {
	if l.Enabled() {
		l.write("info", msg, l.flatten(kvs...))
	}
}

// Error logs an error, with the given message and key/value pairs as context.
func (l *Logger) Error(err error, msg string, kvs ...interface{}) {
	if l.Enabled() {
		l.write("error", msg, l.flatten(kvs...), "error", err)
	}
}

func (l *Logger) write(humanprefix, msg string, values string, extras ...interface{}) {
	var b bytes.Buffer
	if l.humanize {
		if l.colorize {
			if humanprefix == "error" {
				humanprefix = colorRed + humanprefix + colorDefault
			} else {
				humanprefix = colorGreen + humanprefix + " " + colorDefault
			}
		}

		b.WriteString(fmt.Sprintf("%d %-5s | %15s | %-30s", l.level, humanprefix, time.Now().UTC().Format("15:04:05.000000"), msg))
		if l.name != "" {
			b.WriteRune(' ')
			b.WriteString(l.key("logger"))
			b.WriteString("=")
			b.WriteString(l.name)
		}
	} else {
		b.WriteString("level=")
		b.WriteString(strconv.Itoa(l.level))
		if l.name != "" {
			b.WriteRune(' ')
			b.WriteString("logger=")
			b.WriteString(quote(l.name))
		}
		b.WriteRune(' ')
		b.WriteString("msg=")
		b.WriteString(quote(msg))
		if l.tsFormat != "" {
			b.WriteRune(' ')
			b.WriteString("ts=")
			b.WriteString(quote(time.Now().UTC().Format(l.tsFormat)))
		}
	}
	if len(extras) > 0 {
		b.WriteRune(' ')
		b.WriteString(l.flatten(extras...))
	}
	if l.values != "" {
		b.WriteRune(' ')
		b.WriteString(l.values)
	}
	if values != "" {
		b.WriteRune(' ')
		b.WriteString(values)
	}
	b.WriteRune('\n')
	l.w.Write(b.Bytes())
}

// V returns a logger for a specific verbosity level, relative to this Logger.
func (l *Logger) V(level int) logr.Logger {
	l2 := *l
	l2.level += level
	return &l2
}

// WithName returns a logger with a new element added to the logger's name.
func (l *Logger) WithName(name string) logr.Logger {
	l2 := *l
	if l.name != "" {
		l2.name = l.name + l.nameDelim + name
	} else {
		l2.name = name
	}
	return &l2
}

// WithValues returns a logger with additional key-value pairs of context.
func (l *Logger) WithValues(kvs ...interface{}) logr.Logger {
	l2 := *l
	values := l.flatten(kvs...)
	if len(l.values) > 0 {
		l2.values = l.values + " " + values
	} else {
		l2.values = values
	}
	return &l2
}

func (l *Logger) flatten(kvs ...interface{}) string {
	if len(kvs) == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(kvs); i += 2 {
		if i > 0 {
			b.WriteRune(' ')
		}

		k := kvs[i]
		var v interface{}
		if i+1 < len(kvs) {
			v = kvs[i+1]
		} else {
			v = ""
		}
		b.WriteString(l.key(stringify(k)))
		b.WriteRune('=')
		b.WriteString(stringify(v))
	}

	return b.String()
}

func (l *Logger) key(s string) string {
	if !l.colorize {
		return s
	}

	switch s {
	case "error":
		return colorRed + s + colorDefault
	case "logger", "caller":
		return colorBlue + s + colorDefault
	default:
		return colorYellow + s + colorDefault
	}

}

func stringify(v interface{}) string {
	var s string
	switch vv := v.(type) {
	case string:
		s = vv
	case fmt.Stringer:
		s = vv.String()
	case error:
		s = vv.Error()
	default:
		s = fmt.Sprint(v)
	}
	return quote(s)
}

func quote(s string) string {
	if strings.ContainsAny(s, " .") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

const (
	colorDefault = "\x1b[0m"
	colorBlack   = "\x1b[1;30m"
	colorRed     = "\x1b[1;31m"
	colorGreen   = "\x1b[1;32m"
	colorYellow  = "\x1b[1;33m"
	colorBlue    = "\x1b[1;34m"
	colorMagenta = "\x1b[1;35m"
	colorCyan    = "\x1b[1;36m"
	colorWhite   = "\x1b[1;37m"
)
