package main

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/iand/logfmtr"
)

type E struct {
	str string
}

func (e E) Error() string {
	return e.str
}

func main() {
	logfmtr.SetVerbosity(1)

	demo(logfmtr.New())

	opts := logfmtr.DefaultOptions()
	opts.Humanize = true
	opts.Colorize = true
	demo(logfmtr.NewWithOptions(opts))
}

func demo(base logr.Logger) {
	log := base.WithName("MyName").WithValues("user", "you")
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(1).Info("you should see this")
	log.V(1).V(1).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(fmt.Errorf("an error occurred"), "goodbye", "code", -1)

}
