package fpgen

// Maybe holds a value that may be missing (lesson 7; compare fp.Maybe,
// fp/maybe.go).
//
// This is the strongest lesson in the package, because it is backed by a
// real bug on this repo's main branch, not a hypothetical. fp.Maybe has no
// "missing" field -- it distinguishes present from missing by inspecting
// Value itself: nil means missing, and (as of the slide-78 fix) so does a
// typed nil pointer, detected by reflect.ValueOf(x).IsNil(). That inspection
// is fp.Maybe's entire mechanism, and fp/maybe.go's current Map is:
//
//	func (m Maybe) Map(f *Func) Maybe {
//		if m.Value == nil || isNilPtr(m.Value) {
//			return Maybe{}
//		}
//		...
//	}
//
// which short-circuits on ANY typed nil pointer, unconditionally -- it never
// asks whether the step it is about to skip could actually have handled a
// nil pointer, for instance a pointer-receiver method with a nil-safe body.
// A valid answer is silently discarded whenever the value in hand happens to
// be a nil pointer of a type the next step would have accepted. That is a
// live regression, not a design choice: it was merged, a green pipeline did
// not catch it because no test had ever named the case, and it went
// unnoticed until the code was read again. See functional-go-maybe-nil.
//
// The regression exists because "missing" and "present" share one field, and
// telling them apart means inspecting the value -- a Kind check, a nil
// check, an IsNil call -- and every one of those is a place to get the
// condition slightly wrong. Maybe[T] does not inspect anything:
//
//	type Maybe[T any] struct {
//		value T
//		ok    bool
//	}
//
// ok is the only thing Map, Get and everything else here ever consult.
// There is no nil to be typed or untyped, no pointer to dereference, no
// reflect.Value to call IsNil on -- the entire bug class fp.Maybe fell into
// is unrepresentable, not merely untested. That is what "zero value" buys:
// var zero T is a real value of T (0, "", a nil slice, a zero-valued
// struct -- whatever T's zero value is), never inspected, sitting in the
// value field precisely when ok is false and nobody is allowed to look at it.
type Maybe[T any] struct {
	value T
	ok    bool
}

// Some returns a Maybe holding v.
func Some[T any](v T) Maybe[T] {
	return Maybe[T]{value: v, ok: true}
}

// None returns a Maybe holding nothing, of type T.
//
// Go cannot infer T from zero arguments, so a call site that cannot infer it
// from context needs an explicit type argument: None[int](). Lesson 2
// covers exactly this boundary -- inference works from arguments and
// results, never from thin air.
func None[T any]() Maybe[T] {
	return Maybe[T]{}
}

// Get returns m's value and whether it is present.
func (m Maybe[T]) Get() (T, bool) {
	return m.value, m.ok
}

// Map returns Some(f(v)) if m holds v, or None if m is empty (lesson 7;
// compare fp.Maybe.Map, fp/maybe.go).
//
// A free function, for the same reason fpgen's List.Map is (lesson 5): T and
// U differ, and a method cannot introduce U on its own.
//
// Note what this Map does NOT need: fp.Maybe.Map checks isNilPtr at both
// ends, because a Go method handed a typed nil pointer and a Go method
// handed nil both have to be told apart from a genuine value by inspecting
// the interface{} they arrived in. Map here only ever asks m.ok. If T
// happens to be a pointer type and f legitimately wants to receive a nil
// *U -- fp.Maybe.Map's regression is exactly that it refuses to let this
// happen -- Some[*U](nil) represents that on purpose: present, and the
// present value happens to be nil. Nothing here conflates that with absent.
func MaybeMap[T, U any](m Maybe[T], f func(T) U) Maybe[U] {
	if !m.ok {
		return None[U]()
	}
	return Some(f(m.value))
}
