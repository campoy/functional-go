package fpgen

// Map applies f to every element of vs and returns the results (lesson 1).
//
// This single signature replaces fp.Func, fp.NewFunc and fp.Must outright.
// fp.Map does not exist as a free function at all -- slide 24's
//
//	func Map(f interface{}, vs []interface{}) []interface{}
//
// cannot express that f's argument type matches vs's element type, or that
// the result type matches f's return type, so fp routes every call through
// NewFunc (which reconstructs those types at run time via reflect) and Must
// (which turns NewFunc's error into a panic so callers can write it inline).
// Both exist purely to recover information a type parameter carries for
// free. With one:
//
//	func Map[A, B any](vs []A, f func(A) B) []B
//
// there is nothing left to reconstruct and nothing left to validate -- a
// mismatched f fails to compile, so there is no error for a Must to unwrap.
func Map[A, B any](vs []A, f func(A) B) []B {
	out := make([]B, len(vs))
	for i, v := range vs {
		out[i] = f(v)
	}
	return out
}

// Filter returns the elements of vs for which f reports true.
//
// fp has no equivalent: expressing "same element type in and out, one bool
// predicate" through interface{} bought nothing over a concrete
// func([]interface{}, func(interface{}) bool) []interface{}, so slide 24's
// motivating example stuck to Map. Filter costs nothing extra with type
// parameters, which is itself the lesson -- generics make the small,
// obviously-useful combinator cheap enough to write.
func Filter[A any](vs []A, f func(A) bool) []A {
	var out []A
	for _, v := range vs {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce folds f over vs left to right, starting from init.
//
// Also absent from fp for the same reason as Filter. Reduce is the one that
// makes the case hardest for reflect: a reflection-based version would need
// to know not just vs's element type but the accumulator's type too, and
// with two independent type parameters (A for elements, B for the
// accumulator) there is no single reflect.Type to reconstruct it from --
// NewFunc only ever handles one-argument, one-result functions (slide 31).
func Reduce[A, B any](vs []A, init B, f func(B, A) B) B {
	acc := init
	for _, v := range vs {
		acc = f(acc, v)
	}
	return acc
}

// Compose returns the composition of f and g, applying g first: the result
// computes f(g(x)) (lesson 3; compare fp.Compose, fp/func.go).
//
// fp.Compose returns (*Func, error) and its whole body is one line:
//
//	if g.out != f.in {
//		return nil, fmt.Errorf("can't compose: %v != %v", g.out, f.in)
//	}
//
// -- comparing two reflect.Type values it reconstructed by hand, because the
// interface{} signature threw the real types away. That comparison is the
// talk's punchline: a compiler with generics performs it during type
// checking, before the program runs, so there is no error path to return.
// Composing g's actual output type against a function expecting something
// else does not type-check:
//
//	Compose(strings.ToUpper, func(n int) int { return n })
//	// testdata/compose_mismatch.go:18:27: in call to Compose, type func(n int) int of func(n int) int {…} does not match inferred type func(int) string for func(A) B
//
// (the real diagnostic, under this module's go1.21 floor, as printed from
// fpgen/ where the test runs -- see fpgen/testdata/compose_mismatch.go and
// TestWallComposeMismatch in fpgen/wall_test.go, which builds that file and
// always asserts the call is rejected, plus pins the line above character
// for character on a go1.27 or later toolchain. Go's type-inference wording
// is toolchain-dependent, so an older toolchain phrases the same rejection
// differently). Argument
// order matches fp.Compose: mathematical
// composition, g first, which still reads backwards to most people and is
// kept for the same reason fp kept it -- it is what slide 77 prints.
func Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	return func(x A) C { return f(g(x)) }
}
