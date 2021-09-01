package logfmtr_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/iand/logfmtr"
)

// Benchmarks copied from github.com/go-logr/logr/benchmark

//go:noinline
func doInfoOneArg(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("this is", "a", "string")
	}
}

//go:noinline
func doInfoSeveralArgs(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doV0Info(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.V(0).Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doV9Info(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.V(9).Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doError(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	err := fmt.Errorf("error message")
	for i := 0; i < b.N; i++ {
		log.Error(err, "multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doWithValues(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := log.WithValues("k1", "v1", "k2", "v2")
		_ = l
	}
}

//go:noinline
func doWithName(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := log.WithName("name")
		_ = l
	}
}

//go:noinline
func doWithCallDepth(b *testing.B, log logr.Logger) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := log.WithCallDepth(1)
		_ = l
	}
}

func newLogger() logr.Logger {
	return logfmtr.NewWithOptions(discard())
}

func BenchmarkLogfmtrInfoOneArg(b *testing.B) {
	doInfoOneArg(b, newLogger())
}

func BenchmarkLogfmtrInfoSeveralArgs(b *testing.B) {
	doInfoSeveralArgs(b, newLogger())
}

func BenchmarkLogfmtrV0Info(b *testing.B) {
	doV0Info(b, newLogger())
}

func BenchmarkLogfmtrV9Info(b *testing.B) {
	doV9Info(b, newLogger())
}

func BenchmarkLogfmtrError(b *testing.B) {
	doError(b, newLogger())
}

func BenchmarkLogfmtrWithValues(b *testing.B) {
	doWithValues(b, newLogger())
}

func BenchmarkLogfmtrWithName(b *testing.B) {
	doWithName(b, newLogger())
}

func BenchmarkLogfmtrWithCallDepth(b *testing.B) {
	doWithCallDepth(b, newLogger())
}
