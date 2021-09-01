package logfmtr_test

import (
	"testing"

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

func ExampleNewNamed() {
	// Create the logger with a name
	logger := logfmtr.NewNamed("europa")

	logger.Info("the sun is shining")
}

func ExampleDisableLogger() {
	// Create the logger with a name
	logger := logfmtr.NewNamed("europa")

	// It logs normally
	logger.Info("the sun is shining")

	// Disable the logger
	logfmtr.DisableLogger("europa")

	logger.Info("the sun not shining now")
}

func TestIssue3(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	logfmtr.SetVerbosity(1)
	opts := logfmtr.DefaultOptions()
	base := logfmtr.NewWithOptions(opts)

	log := base.WithValues("user", "you")

	// Should not panic
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
}
