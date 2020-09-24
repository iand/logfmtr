package logfmtr_test

import (
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
