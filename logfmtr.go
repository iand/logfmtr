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

var (
	goptionsmu sync.Mutex
	goptions   = DefaultOptions()

	disabledLoggersMu sync.Mutex // synchronises writes to disabledLoggers map
	disabledLoggers   atomic.Value
	anyDisabled       int32 = 0 // an atomicly accessed variable that is set to 1 if there are any loggers that have been manually disabled
)

func init() {
	disabledLoggers.Store(map[string]bool{})
}

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
func New() logr.Logger {
	return logr.New(&sink{})
}

// NewNamed returns a deferred logger with the given name that writes in logfmt using the default options.
func NewNamed(name string) logr.Logger {
	return New().WithName(name)
}

// NewWithOptions returns an instantiated logger that writes in logfmt using the supplied options. Panics if
// no writer is supplied in the options.
func NewWithOptions(opts Options) logr.Logger {
	s := &sink{}
	s.applyOptions(opts)
	return logr.New(s)
}

// DefaultOptions returns the default options used by New unless overridden by a call to UseOptions.
// Override the option field to customise behaviour and then pass to UseOptions or NewWithOptions.
func DefaultOptions() Options {
	return Options{
		Writer:          os.Stdout,
		TimestampFormat: "2006-01-02T15:04:05.000000000Z07:00",
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

var _ logr.LogSink = (*sink)(nil)

// sink is a logger sink that writes messages in the logfmt style.
// See https://www.brandur.org/logfmt for more information.
type sink struct {
	core        *core
	init        sync.Once
	parent      *sink
	runtimeInfo logr.RuntimeInfo
	dfn         func(*core)
}

func (l *sink) instantiate() {
	if l.core != nil {
		return
	}
	if l.parent == nil {
		goptionsmu.Lock()
		l.core = &core{
			runtimeInfo: l.runtimeInfo,
		}
		l.core.applyOptions(goptions)
		goptionsmu.Unlock()
		if l.dfn != nil {
			l.dfn(l.core)
		}
		return
	}

	l.core = l.parent.copyCore(l.dfn)
}

func (l *sink) copyCore(dfn func(*core)) *core {
	l.init.Do(l.instantiate)

	c := *l.core
	dfn(&c)
	return &c
}

func (l *sink) applyOptions(opts Options) {
	l.core = &core{}
	l.core.applyOptions(opts)
}

func (l *sink) Init(info logr.RuntimeInfo) {
	l.runtimeInfo = info
}

// Enabled reports whether this Logger is enabled with respect to the current global log level.
func (l *sink) Enabled(level int) bool {
	l.init.Do(l.instantiate)
	if level > int(atomic.LoadInt32(&gv)) {
		return false
	}
	if l.core.name == "" || atomic.LoadInt32(&anyDisabled) == 0 {
		return true
	}
	disabled := disabledLoggers.Load().(map[string]bool)
	return !disabled[l.core.name]
}

// Info logs a non-error message with the given key/value pairs as context.
func (l *sink) Info(level int, msg string, kvs ...interface{}) {
	l.init.Do(l.instantiate)
	l.core.write(level, "info", msg, l.core.flatten(kvs...))
}

// Error logs an error, with the given message and key/value pairs as context.
func (l *sink) Error(err error, msg string, kvs ...interface{}) {
	l.init.Do(l.instantiate)
	l.core.write(0, "error", msg, l.core.flatten(kvs...), "error", err)
}

// WithName returns a logger with a new element added to the logger's name.
func (l *sink) WithName(name string) logr.LogSink {
	return &sink{
		parent: l,
		dfn: func(c *core) {
			c.appendName(name)
		},
	}
}

// WithValues returns a logger with additional key-value pairs of context.
func (l *sink) WithValues(kvs ...interface{}) logr.LogSink {
	return &sink{
		parent: l,
		dfn: func(c *core) {
			values := c.flatten(kvs...)
			c.appendValues(values)
		},
	}
}

func (l *sink) WithCallDepth(depth int) logr.LogSink {
	return &sink{
		parent: l,
		dfn: func(c *core) {
			c.callerSkip += depth
		},
	}
}

type core struct {
	w           io.Writer
	name        string
	values      string
	humanize    bool
	tsFormat    string
	nameDelim   string
	colorize    bool
	addCaller   bool
	callerSkip  int
	runtimeInfo logr.RuntimeInfo
}

func (c *core) write(level int, humanprefix, msg string, values string, extras ...interface{}) {
	var b bytes.Buffer
	if c.humanize {
		if c.colorize {
			if humanprefix == "error" {
				humanprefix = colorRed + humanprefix + colorDefault
			} else {
				humanprefix = colorGreen + humanprefix + " " + colorDefault
			}
		}

		b.WriteString(fmt.Sprintf("%d %-5s | %15s | %-30s", level, humanprefix, time.Now().UTC().Format("15:04:05.000000"), msg))
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
			b.WriteString(c.caller(1))
		}
	} else {
		b.WriteString("level=")
		b.WriteString(strconv.Itoa(level))
		if c.name != "" {
			b.WriteRune(' ')
			b.WriteString("logger=")
			b.WriteString(quote(c.name))
		}
		if c.tsFormat != "" {
			b.WriteRune(' ')
			b.WriteString("ts=")
			b.WriteString(quote(time.Now().UTC().Format(c.tsFormat)))
		}
		b.WriteRune(' ')
		b.WriteString("msg=")
		b.WriteString(quote(msg))
		if c.addCaller {
			b.WriteRune(' ')
			b.WriteString("caller=")
			b.WriteString(c.caller(1))
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
		_, file, line, ok := runtime.Caller(c.runtimeInfo.CallDepth + skip + c.callerSkip + i)
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

func (c *core) appendName(name string) {
	if name == "" {
		return
	}
	if c.name != "" {
		c.name = c.name + c.nameDelim + name
	} else {
		c.name = name
	}
}

func (c *core) appendValues(values string) {
	if values == "" {
		return
	}
	if len(c.values) > 0 {
		c.values = c.values + " " + values
	} else {
		c.values = values
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

func DisableLogger(name string) {
	setLoggerDisabledStatus(name, true)
}

func EnableLogger(name string) {
	setLoggerDisabledStatus(name, false)
}

func setLoggerDisabledStatus(name string, disabled bool) {
	disabledLoggersMu.Lock()
	defer disabledLoggersMu.Unlock()
	current := disabledLoggers.Load().(map[string]bool)
	next := make(map[string]bool, len(current))
	for k, v := range current {
		if k == name {
			continue
		}
		next[k] = v
	}
	if disabled {
		next[name] = disabled
	}
	disabledLoggers.Store(next)
	if len(next) == 0 {
		atomic.StoreInt32(&anyDisabled, 0)
	} else {
		atomic.StoreInt32(&anyDisabled, 1)
	}
}
