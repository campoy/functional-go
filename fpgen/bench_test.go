package fpgen_test

import (
	"strings"
	"testing"

	"github.com/campoy/functional-go/fp"
	"github.com/campoy/functional-go/fpgen"
)

// benchSize matches sum/sum_test.go's benchSize: 1000 elements, so the two
// packages' costs are compared on the same shape of work sum/ already
// benchmarks fp on. See docs/investigations/generics-vs-reflection.md.
const benchSize = 1000

func fpBenchList() *fp.List {
	var l *fp.List
	for i := benchSize - 1; i >= 0; i-- {
		l = &fp.List{Head: "hello", Tail: l}
	}
	return l
}

func fpgenBenchList() *fpgen.List[string] {
	var l *fpgen.List[string]
	for i := benchSize - 1; i >= 0; i-- {
		l = &fpgen.List[string]{Head: "hello", Tail: l}
	}
	return l
}

var sink interface{}

func BenchmarkFPListMap(b *testing.B) {
	l := fpBenchList()
	f := fp.Must(fp.NewFunc(strings.ToUpper))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = l.Map(f)
	}
}

func BenchmarkFPGenListMap(b *testing.B) {
	l := fpgenBenchList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = fpgen.ListMap(l, strings.ToUpper)
	}
}
