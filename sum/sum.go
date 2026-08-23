// Package sum implements the four ways of adding a slice of ints shown at the
// start of the "Functional Go?" talk (dotGo 2015), slides 8 through 15.
//
// The progression runs from a loop, to recursion, to tail recursion, to tail
// recursion with the tail call flattened into a goto by hand. That last step
// is necessary because Go does not eliminate tail calls, which is the point
// the talk is making: writing in a recursive style costs you real time, and
// the compiler will not give it back.
//
// The benchmarks in sum_test.go reproduce the comparison.
package sum

// SumI adds vs with a loop.
//
// Slides 8 and 9 print "return v", which does not compile: v is the range
// variable and is out of scope by then. The accumulator is s. See NOTES.md.
func SumI(vs []int) int {
	s := 0
	for _, v := range vs {
		s += v
	}
	return s
}

// SumR adds vs by recursing on its tail.
//
// Each call keeps a stack frame alive until the whole slice has been walked,
// so this is both slower than SumI and bounded by stack size.
//
// Slide 10 recurses into Sum, which is never defined anywhere in the deck.
func SumR(vs []int) int {
	if len(vs) == 0 {
		return 0
	}
	return vs[0] + SumR(vs[1:])
}

// SumTR adds vs to s by tail recursion, carrying the running total in s so
// that the recursive call is the last thing the function does.
//
// A language with tail-call elimination would compile this into a loop. Go
// does not: this still emits a recursive CALL and burns a stack frame per
// element, which is the whole reason SumTRG below exists.
//
// The talk measured this as slower than SumR. On a modern toolchain it comes
// out slightly faster instead, because the addition happens before the call
// rather than after it, so nothing has to be spilled across the call. That is
// a calling-convention difference, not tail-call elimination. The evidence is
// in docs/tail-recursion.md.
//
// Callers start the recursion with s set to 0.
//
// Slide 12 recurses into Sum, as slide 10 does.
func SumTR(vs []int, s int) int {
	if len(vs) == 0 {
		return s
	}
	return SumTR(vs[1:], s+vs[0])
}

// SumTRG is SumTR with the tail call flattened into a goto by hand, which is
// what a compiler performing tail-call elimination would have done.
//
// It recovers most of the gap to SumI, but not all of it: reslicing vs on
// every step costs more than ranging over it once.
//
// Callers start with s set to 0, as with SumTR.
func SumTRG(vs []int, s int) int {
begin:
	if len(vs) == 0 {
		return s
	}
	vs, s = vs[1:], s+vs[0]
	goto begin
}
