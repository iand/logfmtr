package logfmtr

import (
	"context"

	"github.com/go-logr/logr"
)

type contextKey struct{}

// FromContext returns a logger constructed from the context or a new logger if no logger details are found in the context.
func FromContext(ctx context.Context) *Logger {
	if v, ok := ctx.Value(contextKey{}).(core); ok {
		return &Logger{
			dfn: func(c *core) {
				*c = v
			},
		}
	}

	return New()
}

// NewContext returns a new context that embeds the logger's name, values, level and other options.
func NewContext(ctx context.Context, l logr.Logger) context.Context {
	if lf, ok := l.(*Logger); ok {
		return context.WithValue(ctx, contextKey{}, lf.getCore())
	}

	return ctx
}
