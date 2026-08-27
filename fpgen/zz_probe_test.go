package fpgen_test

import (
	"strings"
	"testing"

	"github.com/campoy/functional-go/fp"
)

func BenchmarkProbeFPIdent(b *testing.B) {
	l := fpBenchList()
	f := fp.Must(fp.NewFunc(func(s string) string { return s }))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = l.Map(f)
	}
}

func BenchmarkProbeFPUpper(b *testing.B) {
	l := fpBenchList()
	f := fp.Must(fp.NewFunc(strings.ToUpper))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = l.Map(f)
	}
}
