//go:build ignore

// This file does not compile, on purpose. It is lesson 5, THE FIRST WALL, in
// docs/teaching-generics.md: the obvious generic transliteration of
// fp.List.Map (fp/list.go) as a method rather than a free function.
//
// fpgen/example_test.go runs `go build` on this exact file and asserts the
// diagnostic below, so the claim in the doc comments cannot silently rot.
package testdata

type List[T any] struct {
	Head T
	Tail *List[T]
}

func (l *List[A]) Map[B any](f func(A) B) *List[B] {
	if l == nil {
		return nil
	}
	return &List[B]{f(l.Head), l.Tail.Map(f)}
}
