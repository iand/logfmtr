package logfmtr

import (
	"context"

	"github.com/go-logr/logr"
)

type contextKey struct{}

// FromContext returns a logger constructed from the context or a new logger if no logger details are found in the context.
// Deprecated: overlaps with logr functionality, use logr.FromContext instead
func FromContext(ctx context.Context) *Logger {
	l := logr.FromContext(ctx)
	if ll, ok := l.(*Logger); ok {
		return ll
	}
	return New()
}

// NewContext returns a new context that embeds the logger's name, values, level and other options.
// Deprecated: overlaps with logr functionality, use logr.NewContext instead
func NewContext(ctx context.Context, l logr.Logger) context.Context {
	return logr.NewContext(ctx, l)
}
