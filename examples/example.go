package main

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/iand/logfmtr"
)

func main() {
	logfmtr.SetVerbosity(1)

	demo(logfmtr.New())

	opts := logfmtr.DefaultOptions()
	opts.Humanize = true
	opts.Colorize = true
	opts.CallerSkip = 0
	demo(logfmtr.NewWithOptions(opts))

	deferred()

	disableDemo()
}

func demo(base logr.Logger) {
	log := base.WithName("MyName").WithValues("user", "you")
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(1).Info("you should see this")
	log.V(1).V(1).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(fmt.Errorf("an error occurred"), "goodbye", "code", -1)
}

func deferred() {
	l1 := logfmtr.New().WithName("before")
	l2 := logfmtr.NewNamed("after")
	l3 := l2.WithValues("some", "value")

	l1.Info("this should be logged with default options")
	opts := logfmtr.DefaultOptions()
	opts.Humanize = true
	opts.Colorize = true
	logfmtr.UseOptions(opts)

	l2.Info("this should be logged with global options since instatiation was deferred until first write")
	l3.Info("this should also be logged with new options")
	l1.Info("this should be logged with the old options since first write was before we set global options")
}

func disableDemo() {
	log := logfmtr.NewNamed("europa")
	log.Info("hello, this logger is enabled")

	logfmtr.DisableLogger("europa")
	log.Info("you should NOT see this, the logger is disabled")

	log2 := log.WithName("moon")
	log2.Info("you should see this, a child logger does not inherit from its parent")

	logfmtr.EnableLogger("europa")
	log.Info("you should see this now, the logger was enabled again")
}
