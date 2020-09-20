package logfmtr_test

import (
	"github.com/iand/logfmtr"
)

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
