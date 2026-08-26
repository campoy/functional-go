package fpgen

// Chain2 and Chain3 answer lesson 9, THE SECOND WALL, honestly rather than
// solving it.
//
// fp.Maybe.Do and fp.Many.Do (fp/maybe.go, fp/many.go) take a variadic chain
// where every step can have a different type pair:
//
//	w, err := Maybe{p}.Do(Person.Address, Address.City, City.Weather, Weather.Description)
//
// Person.Address is func(Person) *Address; Address.City is func(*Address)
// City; City.Weather is func(City) *Weather. Three different signatures,
// one variadic parameter. There is no way to write that as a generic
// signature, because Go's variadics are homogeneous by construction --
// `fs ...T` needs a single T for every element, and generics gives you more
// ways to parameterise T, not a way to make one parameter list stand for a
// sequence of different types. Concretely, none of the shapes that look like
// they should work do:
//
//	func Do[T any](v T, fs ...func(T) T) T                     // forces every step to keep the same type
//	func Do[T, U any](v T, fs ...func(any) any) U               // back to interface{}; the whole point was to avoid it
//	func Do[T, U any](v T, fs ...func(T) U) U                   // every step must be func(T) U -- the SAME T and U, so it only accepts a chain of length 1
//
// The honest alternatives are: nested calls (MaybeMap(MaybeMap(MaybeMap(m,
// step1), step2), step3) -- correct, and unreadable past three steps), a
// fixed-arity family like Chain2/Chain3 below (readable, but the library
// has to predict the longest chain anyone will want and provide a Chain for
// it), or keeping reflect, i.e. fp.Maybe.Do itself, unchanged (arbitrary
// length, back to run-time type errors). This package takes the middle
// option for exactly the lengths the deck's own examples need -- the weather
// chain is four steps long counting the starting value, i.e. three function
// arguments -- and states plainly that it does not generalise: a fifth
// step needs a Chain4 nobody has written yet.
//
// Chain2 threads v through f then g: g(f(v)).
func Chain2[A, B, C any](v A, f func(A) B, g func(B) C) C {
	return g(f(v))
}

// Chain3 threads v through f, g, then h: h(g(f(v))).
func Chain3[A, B, C, D any](v A, f func(A) B, g func(B) C, h func(C) D) D {
	return h(g(f(v)))
}
