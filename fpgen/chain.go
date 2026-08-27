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
// sequence of different types.
//
// This wall stands in a different place from the other two, and where it
// stands is part of the lesson. Lesson 5's wall and lesson 10's are
// rejections of a DECLARATION: write the declaration, the compiler refuses
// it, done. Every shape below DECLARES fine. The wall shows up only at the
// CALL SITE, when one of these legal signatures is handed a chain whose type
// changes at every step:
//
//	func Do[T any](v T, fs ...func(T) T) T                     // legal; forces every step to keep the same type
//	func Do[T, U any](v T, fs ...func(any) any) U               // legal; back to interface{}, the whole point was to avoid it
//	func Do[T, U any](v T, fs ...func(T) U) U                   // legal; every step must be the SAME func(T) U, so a chain of length 1
//
// Only the first is a wall rather than a compromise -- the other two compile
// AND accept a call, they just give up type safety and generality
// respectively. Feed the weather chain to the first and the compiler says
// so, at the call:
//
//	Do(p, Person.Address, Address.City, City.Weather, Weather.Description)
//	// testdata/wall_variadic_chain.go:44:12: in call to Do, type func(Person) *Address of Person.Address does not match inferred type func(Person) Person for func(T) T
//
// (the real diagnostic, as printed from fpgen/ where the test runs -- see
// fpgen/testdata/wall_variadic_chain.go and TestWallVariadicChain in
// fpgen/wall_test.go, which builds that file, always asserts the call is
// rejected, and pins the line above character for character on a go1.27 or
// later toolchain. The other half -- that the bare Do declaration itself
// compiles -- is pinned by a package-level declaration in that same file,
// var _ func(wallPerson, ...func(wallPerson) wallPerson) wallPerson =
// wallDo[wallPerson], checked by every ordinary go build, go test and go
// vet rather than by a runtime assertion, since "the declaration is fine,
// the call is not" is the whole distinction.)
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
