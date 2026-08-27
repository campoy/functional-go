//go:build ignore

// This file does not compile, on purpose: it feeds Compose (fpgen/func.go)
// two functions whose types do not line up, the same mismatch fp.Compose
// (fp/func.go) catches at run time with "can't compose: %v != %v". Here the
// compiler catches it instead, before the program runs -- that shift is
// lesson 3 in docs/teaching-generics.md, and fpgen/wall_test.go runs
// `go build` on this file to pin the real diagnostic.
package main

import "strings"

func Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	return func(x A) C { return f(g(x)) }
}

func main() {
	Compose(strings.ToUpper, func(n int) int { return n })
}
