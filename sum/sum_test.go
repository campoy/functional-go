package sum

import (
	"math/rand"
	"testing"
)

// impls lists the four implementations behind a common signature, so that
// every correctness test runs against all of them. SumTR and SumTRG take an
// accumulator, which callers seed with 0.
var impls = []struct {
	name string
	sum  func([]int) int
}{
	{"SumI", SumI},
	{"SumR", SumR},
	{"SumTR", func(vs []int) int { return SumTR(vs, 0) }},
	{"SumTRG", func(vs []int) int { return SumTRG(vs, 0) }},
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		vs   []int
		want int
	}{
		{"nil", nil, 0},
		{"empty", []int{}, 0},
		{"single", []int{42}, 42},
		{"several", []int{1, 2, 3, 4, 5}, 15},
		{"negatives", []int{-1, -2, -3}, -6},
		{"mixed signs", []int{-10, 5, 5}, 0},
		{"zeroes", []int{0, 0, 0}, 0},
		{"long", seq(1000), 499500},
	}

	for _, impl := range impls {
		for _, tt := range tests {
			got := impl.sum(tt.vs)
			if got != tt.want {
				t.Errorf("%s(%s) = %d, want %d", impl.name, tt.name, got, tt.want)
			}
		}
	}
}

// TestSumAgree is the regression test for the slide typos: SumI returning the
// wrong variable, and SumR and SumTR recursing into a function that does not
// exist. Any of those reintroduced would show up as a disagreement here.
func TestSumAgree(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for n := 0; n < 100; n++ {
		vs := make([]int, r.Intn(200))
		for i := range vs {
			vs[i] = r.Intn(2000) - 1000
		}

		want := impls[0].sum(vs)
		for _, impl := range impls[1:] {
			if got := impl.sum(vs); got != want {
				t.Fatalf("%s(%v) = %d, want %d (%s)", impl.name, vs, got, want, impls[0].name)
			}
		}
	}
}

// TestSumAccumulator checks that the two tail-recursive forms add to the
// accumulator they are given rather than ignoring it, which is what makes
// them tail recursive in the first place.
func TestSumAccumulator(t *testing.T) {
	accs := []struct {
		name string
		sum  func([]int, int) int
	}{
		{"SumTR", SumTR},
		{"SumTRG", SumTRG},
	}

	for _, acc := range accs {
		if got := acc.sum([]int{1, 2, 3}, 10); got != 16 {
			t.Errorf("%s([1 2 3], 10) = %d, want 16", acc.name, got)
		}
		if got := acc.sum(nil, 7); got != 7 {
			t.Errorf("%s(nil, 7) = %d, want 7", acc.name, got)
		}
	}
}

// TestSumDoesNotModify guards the reslicing in SumR, SumTR and SumTRG: they
// walk vs by taking vs[1:], which shares the backing array with the caller's
// slice, so an implementation that wrote through it would corrupt the input.
func TestSumDoesNotModify(t *testing.T) {
	for _, impl := range impls {
		vs := []int{1, 2, 3, 4, 5}
		impl.sum(vs)
		for i, v := range vs {
			if v != i+1 {
				t.Errorf("%s modified its input: vs[%d] = %d, want %d", impl.name, i, v, i+1)
			}
		}
	}
}

// benchSize is how many elements the benchmarks sum.
//
// The slides report the timings but never state the input size. 1000 is the
// value this reconstruction picked, and is consistent with the 462 ns/op the
// talk reports for SumI on 2015 hardware. See NOTES.md.
const benchSize = 1000

var benchInput = seq(benchSize)

// sink keeps the compiler from proving the benchmarked calls are dead and
// eliminating them.
var sink int

func BenchmarkSumI(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s = SumI(benchInput)
	}
	sink = s
}

func BenchmarkSumR(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s = SumR(benchInput)
	}
	sink = s
}

func BenchmarkSumTR(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s = SumTR(benchInput, 0)
	}
	sink = s
}

func BenchmarkSumTRG(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s = SumTRG(benchInput, 0)
	}
	sink = s
}

// seq returns []int{0, 1, ..., n-1}.
func seq(n int) []int {
	vs := make([]int, n)
	for i := range vs {
		vs[i] = i
	}
	return vs
}
