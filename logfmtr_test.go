package logfmtr_test

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/iand/logfmtr"
)

var _ logr.Logger = (*logfmtr.Logger)(nil)
var _ logr.Logger = logfmtr.Null

func ExampleNewWithOptions() {
	opts := logfmtr.DefaultOptions()
	opts.Humanize = true
	logger := logfmtr.NewWithOptions(opts)
	logger.Info("the sun is shining")
}

func ExampleNew() {
	// Create the logger with deferred options
	logger := logfmtr.New()

	// Later on set options that all loggers will be based on
	opts := logfmtr.DefaultOptions()
	opts.Humanize = true
	logfmtr.UseOptions(opts)

	// Logger is instantiated with the options
	logger.Info("the sun is shining")
}

func ExampleNewNamed() {
	// Create the logger with a name
	logger := logfmtr.NewNamed("europa")

	logger.Info("the sun is shining")
}

func ExampleFromContext() {
	// Create a logger
	root := logfmtr.New().WithName("root").V(2)

	// Embed the logger in a context
	loggerCtx := logfmtr.NewContext(context.Background(), root)

	// A function that uses a context
	other := func(ctx context.Context) {
		// Retrieve the logger from the context
		logger := logfmtr.FromContext(ctx)
		logger.Info("the sun is shining")
	}

	// Pass the context to the other function
	other(loggerCtx)

}
