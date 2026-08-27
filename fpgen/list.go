package fpgen

import (
	"fmt"
	"strings"
)

// List is a cons cell: a value and the rest of the list, with nil for the
// empty list (lesson 4; compare fp.List, fp/list.go).
//
// Type parameters go on the type declaration, and List[T] instantiates like
// any other generic type: List[string], List[int], *List[List[int]]. The
// recursive field is spelled the same way the non-generic version was --
// Tail *List[T], not Tail *List -- because inside the type's own body T is
// just an ordinary in-scope identifier.
type List[T any] struct {
	Head T
	Tail *List[T]
}

// ListMap returns the list of f applied to every element of l (lesson 5, THE
// FIRST WALL; compare fp.List.Map, fp/list.go).
//
// fp.List.Map is a method: func (l *List) Map(f *Func) *List. The obvious
// generic transliteration is not legal Go:
//
//	func (l *List[A]) Map[B any](f func(A) B) *List[B] { ... }
//	// testdata/wall_method_type_param.go:16:23: generic method requires go1.27 or later (-lang was set to go1.21; check go.mod)
//
// (verified against this module's own floor -- see
// fpgen/testdata/wall_method_type_param.go and fpgen/wall_test.go, which
// runs `go build -gcflags=-lang=go1.21` on it and checks that exact text).
// A method's receiver fixes every type parameter the method can use;
// List[A]'s Map needs a second one, B, that the receiver does not supply,
// and under the go1.21 semantics this module declares there is no syntax
// for a method to introduce one of its own. Go 1.27 lifted that for methods
// on concrete types, which is why the diagnostic reads as a version gate --
// but go.mod stays at go1.21, so the wall is real here. So
// Map has to be a free function here, the same shape fp.List.Map itself
// started as on slide 35 before slide 36 asked "Should this be a method? Of
// what?" and slide 37 turned it into one. Generics answers that question
// precisely: a method, only when the element type does not change.
//
// Reverse below is the contrast -- A to A, no second type parameter needed,
// legal as a method.
//
// A second, smaller consequence of the same wall: fp.List.Map, fp.Maybe.Map
// and fp.Many.Map are three methods sharing one name, because a method
// lives in its receiver type's namespace. Free functions all live in one
// flat package namespace with no overloading, so none of the three can just
// be called Map here -- Map itself already names the slice function from
// lesson 1 (fpgen/func.go), so List's is ListMap, Maybe's is MaybeMap
// (fpgen/maybe.go), Many's is ManyMap (fpgen/many.go).
func ListMap[A, B any](l *List[A], f func(A) B) *List[B] {
	if l == nil {
		return nil
	}
	return &List[B]{f(l.Head), ListMap(l.Tail, f)}
}

// Reverse returns l reversed. Unlike Map it changes no element type, so it
// takes no second type parameter and is a legal method -- the case the wall
// in Map does not block.
func (l *List[T]) Reverse() *List[T] {
	var out *List[T]
	for ; l != nil; l = l.Tail {
		out = &List[T]{l.Head, out}
	}
	return out
}

// String renders l as its elements, comma-separated, strings quoted -- the
// same format fp.List.String uses, via Quote (fpgen/constraints.go).
func (l *List[T]) String() string {
	var elems []string
	for ; l != nil; l = l.Tail {
		elems = append(elems, Quote(l.Head))
	}
	return strings.Join(elems, ", ")
}

var _ fmt.Stringer = (*List[int])(nil)
