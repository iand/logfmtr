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
