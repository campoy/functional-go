package fpgen

import "strings"

// Many is a cons cell like List (lesson 8; compare fp.Many, fp/many.go).
//
// fp.Many is List with one difference: its Map flattens, via toSlice, so a
// step returning []string can follow a step returning string with no
// per-type plumbing. Many[T] here is List[T] with no difference at all --
// ManyMap below is exactly List's Map with a new name (see fpgen/list.go).
// The flattening fp.Many.Map does implicitly is FlatMap's job instead; see
// FlatMap for why that has to be a second function rather than a build-in
// behaviour of Map.
type Many[T any] struct {
	Head T
	Tail *Many[T]
}

// NewMany builds a Many from vs, one cell per element.
//
// fp.NewMany takes interface{} and flattens a slice or array into cells
// because it has no way to say "a slice of T" -- interface{} forgets that
// the argument was ever a slice at all, so toSlice has to rediscover it with
// reflect.ValueOf(v).Kind(). NewMany[T any](vs []T) says exactly what it
// takes and needs no such rediscovery.
func NewMany[T any](vs []T) *Many[T] {
	var m *Many[T]
	for i := len(vs) - 1; i >= 0; i-- {
		m = &Many[T]{vs[i], m}
	}
	return m
}

// ManyMap applies f to every element of m, one result per element (lesson 8;
// compare fp.Many.Map, fp/many.go). See fpgen/list.go for why Many's Map is
// named ManyMap here, not Map.
func ManyMap[A, B any](m *Many[A], f func(A) B) *Many[B] {
	if m == nil {
		return nil
	}
	return &Many[B]{f(m.Head), ManyMap(m.Tail, f)}
}

// FlatMap applies f to every element of m and concatenates the results
// (lesson 8; compare fp.Many.Map, fp/many.go).
//
// fp.Many.Map does both jobs ManyMap and FlatMap do here, under one name,
// because toSlice cannot tell func(string) string from func(string)
// []string apart -- both go through the same
//
//	interface{} -> reflect.ValueOf(...).Kind()
//
// check, which sees "is the result a slice" and nothing about what f's
// declared return type was. A one-result f flattens to one cell; a
// slice-result f flattens to many; fp.Many.Map cannot express that these are
// two different shapes because interface{} erased the difference before
// toSlice ever ran.
//
// func(A) B and func(A) []B are two different, fully distinct Go types, so
// ManyMap and FlatMap have to be two different generic functions -- there is
// no single signature that accepts either the way fp.Many.Map's interface{}
// parameter does. Losing that convenience is real: NewMany(strs) chaining
// strings.ToUpper (func(string) string, wants ManyMap) into strings.Fields
// (func(string) []string, wants FlatMap) has to name the right one at each
// step, where fp.Many.Do just calls Map throughout (see fp/many_test.go and
// fpgen/many_test.go for the same chain, written both ways). What is gained
// is that the call site states, once and in the type, whether a step is
// expected to multiply its element -- fp's version can only be found out by
// running it.
func FlatMap[A, B any](m *Many[A], f func(A) []B) *Many[B] {
	if m == nil {
		return nil
	}
	vs := f(m.Head)
	rest := FlatMap(m.Tail, f)
	for i := len(vs) - 1; i >= 0; i-- {
		rest = &Many[B]{vs[i], rest}
	}
	return rest
}

// Each applies f to every element of m for its side effects (compare
// fp.Many.Each, fp/many.go).
//
// fp.Many.Each does its own reflection because a func(T) with no result
// cannot go through NewFunc, which requires exactly one. Each[T any](m
// *Many[T], f func(T)) needs no such carve-out -- there was never a NewFunc
// to route through in the first place, so there is nothing to work around.
func Each[T any](m *Many[T], f func(T)) {
	for ; m != nil; m = m.Tail {
		f(m.Head)
	}
}

// String renders m as its elements, comma-separated, strings quoted, the
// same format fp.Many.String uses, via Quote (fpgen/constraints.go).
func (m *Many[T]) String() string {
	var elems []string
	for ; m != nil; m = m.Tail {
		elems = append(elems, Quote(m.Head))
	}
	return strings.Join(elems, ", ")
}
