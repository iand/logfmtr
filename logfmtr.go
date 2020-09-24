package logfmtr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

var goptionsmu sync.Mutex
var goptions = DefaultOptions()

// UseOptions sets options that new loggers will use when it they are instantiated.
func UseOptions(opts Options) {
	goptionsmu.Lock()
	goptions = opts
	goptionsmu.Unlock()
}

// New returns a deferred logger that writes in logfmt using the default options.
// The logger defers configuring its options until it is instantiated with the first call to Info, Error
// or Enabled or the first call to those function on any child loggers created via V, WithName or
// WithValues.
func New() *Logger {
	return &Logger{}
}

// NewNamed returns a deferred logger with the given name that writes in logfmt using the default options.
func NewNamed(name string) *Logger {
	return &Logger{
		dfn: func(c *core) {
			if c.name != "" {
				c.name = c.name + c.nameDelim + name
			} else {
				c.name = name
			}
		},
	}
}

// NewWithOptions returns an instantiated logger that writes in logfmt using the supplied options. Panics if
// no writer is supplied in the options.
func NewWithOptions(opts Options) *Logger {
	l := &Logger{}
	l.applyOptions(opts)
	return l
}

// DefaultOptions returns the default options used by New unless overridden by a call to UseOptions.
// Override the option field to customise behaviour and then pass to UseOptions or NewWithOptions.
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

	// AddCaller indicates that log messages should include the file and line number of the caller of the logger.
	AddCaller bool

	// CallerSkip adds frames to skip when determining the caller of the logger. Useful when the logger is wrapped
	// by another logger.
	CallerSkip int
}

// Logger is a logger that writes messages in the logfmt style.
// See https://www.brandur.org/logfmt for more information.
type Logger struct {
	core   *core
	init   sync.Once
	parent *Logger
	dfn    func(*core)
}

func (l *Logger) instantiate() {
	if l.core != nil {
		return
	}
	if l.parent == nil {
		goptionsmu.Lock()
		l.core = &core{}
		l.core.applyOptions(goptions)
		goptionsmu.Unlock()
		if l.dfn != nil {
			l.dfn(l.core)
		}
		return
	}

	l.core = l.parent.copyCore(l.dfn)
}

func (l *Logger) copyCore(dfn func(*core)) *core {
	l.init.Do(l.instantiate)

	c := *l.core
	dfn(&c)
	return &c
}

func (l *Logger) applyOptions(opts Options) {
	l.core = &core{}
	l.core.applyOptions(opts)
}

// Enabled repoorts whether this Logger is enabled with respect to the current global log level.
func (l *Logger) Enabled() bool {
	l.init.Do(l.instantiate)
	return l.core.level <= int(atomic.LoadInt32(&gv))
}

// Info logs a non-error message with the given key/value pairs as context.
func (l *Logger) Info(msg string, kvs ...interface{}) {
	if l.Enabled() {
		l.core.write("info", msg, l.core.flatten(kvs...))
	}
}

// Error logs an error, with the given message and key/value pairs as context.
func (l *Logger) Error(err error, msg string, kvs ...interface{}) {
	if l.Enabled() {
		l.core.write("error", msg, l.core.flatten(kvs...), "error", err)
	}
}

// V returns a logger for a specific verbosity level, relative to this Logger.
func (l *Logger) V(level int) logr.Logger {
	return &Logger{
		parent: l,
		dfn: func(c *core) {
			c.level += level
		},
	}
}

// WithName returns a logger with a new element added to the logger's name.
func (l *Logger) WithName(name string) logr.Logger {
	return &Logger{
		parent: l,
		dfn: func(c *core) {
			if c.name != "" {
				c.name = c.name + c.nameDelim + name
			} else {
				c.name = name
			}
		},
	}
}

// WithValues returns a logger with additional key-value pairs of context.
func (l *Logger) WithValues(kvs ...interface{}) logr.Logger {
	return &Logger{
		parent: l,
		dfn: func(c *core) {
			values := c.flatten(kvs...)
			if len(c.values) > 0 {
				c.values = c.values + " " + values
			} else {
				c.values = values
			}
		},
	}
}

type core struct {
	w          io.Writer
	level      int
	name       string
	values     string
	humanize   bool
	tsFormat   string
	nameDelim  string
	colorize   bool
	addCaller  bool
	callerSkip int
}

func (c *core) write(humanprefix, msg string, values string, extras ...interface{}) {
	var b bytes.Buffer
	if c.humanize {
		if c.colorize {
			if humanprefix == "error" {
				humanprefix = colorRed + humanprefix + colorDefault
			} else {
				humanprefix = colorGreen + humanprefix + " " + colorDefault
			}
		}

		b.WriteString(fmt.Sprintf("%d %-5s | %15s | %-30s", c.level, humanprefix, time.Now().UTC().Format("15:04:05.000000"), msg))
		if c.name != "" {
			b.WriteRune(' ')
			b.WriteString(c.key("logger"))
			b.WriteString("=")
			b.WriteString(c.name)
		}
		if c.addCaller {
			b.WriteRune(' ')
			b.WriteString(c.key("caller"))
			b.WriteString("=")
			b.WriteString(c.caller(2))
		}
	} else {
		b.WriteString("level=")
		b.WriteString(strconv.Itoa(c.level))
		if c.name != "" {
			b.WriteRune(' ')
			b.WriteString("logger=")
			b.WriteString(quote(c.name))
		}
		b.WriteRune(' ')
		b.WriteString("msg=")
		b.WriteString(quote(msg))
		if c.tsFormat != "" {
			b.WriteRune(' ')
			b.WriteString("ts=")
			b.WriteString(quote(time.Now().UTC().Format(c.tsFormat)))
		}
		if c.addCaller {
			b.WriteRune(' ')
			b.WriteString("caller=")
			b.WriteString(c.caller(2))
		}
	}
	if len(extras) > 0 {
		b.WriteRune(' ')
		b.WriteString(c.flatten(extras...))
	}

	if c.values != "" {
		b.WriteRune(' ')
		b.WriteString(c.values)
	}
	if values != "" {
		b.WriteRune(' ')
		b.WriteString(values)
	}
	b.WriteRune('\n')
	_, _ = c.w.Write(b.Bytes())
}

func (c *core) caller(skip int) string {
	for i := 1; i < 3; i++ {
		_, file, line, ok := runtime.Caller(skip + c.callerSkip + i)
		if ok && file != "<autogenerated>" {
			return path.Base(file) + ":" + strconv.Itoa(line)
		}
	}
	return "unknown"
}

func (c *core) applyOptions(opts Options) {
	if opts.Writer == nil {
		panic("logger was supplied with nil writer")
	}
	c.w = opts.Writer
	c.humanize = opts.Humanize
	c.tsFormat = opts.TimestampFormat
	c.nameDelim = opts.NameDelim
	c.colorize = opts.Colorize && opts.Humanize
	c.addCaller = opts.AddCaller
	c.callerSkip = opts.CallerSkip
}

func (c *core) flatten(kvs ...interface{}) string {
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
		b.WriteString(c.key(stringify(k)))
		b.WriteRune('=')
		b.WriteString(stringify(v))
	}

	return b.String()
}

func (c *core) key(s string) string {
	if !c.colorize {
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
	if strings.ContainsAny(s, " ") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

const (
	colorDefault = "\x1b[0m"
	colorRed     = "\x1b[1;31m"
	colorGreen   = "\x1b[1;32m"
	colorYellow  = "\x1b[1;33m"
	colorBlue    = "\x1b[1;34m"
)

// Null is a non-functional logger that may be used as a placeholder or to disable logging with zero overhead
var Null nullLogger

type nullLogger struct{}

// Enabled always reports false
func (n nullLogger) Enabled() bool { return false }

// Info is a no-op
func (n nullLogger) Info(string, ...interface{}) {}

// Error is a no-op
func (n nullLogger) Error(error, string, ...interface{}) {}

// V is not supported and panics if called
func (n nullLogger) V(int) logr.Logger { panic("V is not supported by null logger") }

// WithName is not supported and panics if called
func (n nullLogger) WithName(string) logr.Logger {
	panic("WithName is not supported by null logger")
}

// WithValues is not supported and panics if called
func (n nullLogger) WithValues(...interface{}) logr.Logger {
	panic("WithValues is not supported by null logger")
}
