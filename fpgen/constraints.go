package fpgen

import (
	"cmp"
	"fmt"
)

// Quote renders v the way fp.quote does: a string in double quotes,
// everything else however fmt would show it (lesson 6; compare fp's
// unexported quote, fp/list.go).
//
// fp.quote is a runtime type switch on interface{}:
//
//	func quote(v interface{}) string {
//		if s, ok := v.(string); ok {
//			return fmt.Sprintf("%q", s)
//		}
//		return fmt.Sprintf("%v", v)
//	}
//
// Quote[T any] keeps that exact shape, deliberately -- T any is the
// constraint that promises nothing, so the body still needs a dynamic
// assertion (any(v).(string)) to find out whether T happens to be string.
// This is the honest half of lesson 6: an unconstrained type parameter does
// not remove a run-time check, it just spells interface{} as any. Max below
// is the other half -- what changes when the constraint actually promises
// something.
func Quote[T any](v T) string {
	if s, ok := any(v).(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}

// Max returns the larger of a and b, using cmp.Ordered (stdlib, go1.21).
//
// fp has nothing like Max: writing it over interface{} would need a type
// switch enumerating every orderable Kind (Int, Int64, Float64, String, ...)
// reflect can report, and would still be wrong for a defined type built on
// one of them. cmp.Ordered promises that > is defined for T at compile time,
// so Max needs no such switch and no reflect import -- the check that would
// have been a run-time type switch in fp is instead the compiler refusing to
// instantiate Max[T] for a T that is not ordered. This is the case Quote
// above is not: a constraint that actually buys something.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Number is a union constraint: any type whose underlying type is one of
// these (the ~ is the approximation element, so a defined type like
// type Cents int also satisfies it).
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

// Sum adds up vs. Motivated the same way Max is -- fp would need a Kind
// switch over reflect.Int, reflect.Int64, reflect.Float64 and so on, plus a
// case for every defined type sharing one of those underlying types; the
// Number constraint states the same set once, at compile time, and Sum's
// body never inspects a type at all.
func Sum[T Number](vs []T) T {
	var total T
	for _, v := range vs {
		total += v
	}
	return total
}

// Named is a method constraint: any type with a Name() string method.
type Named interface {
	Name() string
}

// Describe returns v.Name(). A method constraint is checked at the call
// site, at compile time -- a T without a Name() method fails to instantiate
// Describe[T], the same way a T without > fails to instantiate Max[T]. fp's
// closest equivalent, Many.Each (fp/many.go), instead reflects on f at
// Each's call and panics if f is not a plain func(T) with no results; there
// is no equivalent panic here because there is nothing left to check once
// the program compiles.
func Describe[T Named](v T) string {
	return v.Name()
}
