# Teaching generics through fp and fpgen

`fp` (`fp/`) is the reflection-based library the 2015 talk built because Go
1.5 had no generics. `fpgen` (`fpgen/`) is the same library rebuilt on Go's
type parameters. Neither package is complete on its own as a lesson; the
lesson is the *difference* between them. Every entry below points at a
specific before-in-`fp` / after-in-`fpgen` pair, not a generics feature in
the abstract.

Two of the talk's shapes still cannot be written, even with generics. They
are lessons 5 and 9-10 below, and they are not softened: the honest limits of
Go's generics are as much the point of this document as the features are.

`fp` is unchanged by this work — see `NOTES.md`'s "fpgen" section for
`fpgen`'s own departures from a literal transliteration.

---

## 0. The setup

**Before:** slides 24-29 of the deck. With generics you would write

```go
func Map(f func(a α) β, vs []α) []β
```

but Go 1.5 had none, so every parameter collapses to `interface{}`:

```go
func Map(f interface{}, vs []interface{}) []interface{}
```

and all static type safety is gone — nothing stops a caller passing a
`func(int) string` and a `[]bool`. Nothing is taught yet; this is the
problem the rest of the document answers, piece by piece.

## 1. Type parameters on functions

**Before:** `fp.Func`, `fp.NewFunc`, `fp.Must` (`fp/func.go`). Because
`interface{}` erases a function's argument and result types, `fp` has to
reconstruct them at run time: `NewFunc` uses `reflect` to inspect `f`, checks
its arity, and stores its `in`/`out` `reflect.Type`s in a `Func` so later
code (`Compose`, `List.Map`, `Do`) can check them again. `Must` exists purely
so `NewFunc`'s `error` can be unwrapped inline at a call site — an entire
error-returning constructor and its panic-wrapper, both there only to carry
information a compiler with type parameters carries for free.

**After:** `fpgen.Map[A, B any](vs []A, f func(A) B) []B`
(`fpgen/func.go`). One line. `Func`, `NewFunc`, and `Must` do not exist in
`fpgen` at all — there is nothing left to reconstruct, and nothing to
validate, because a mismatched `f` is a compile error before `Map` ever
runs.

**Concept:** a type parameter list on a function (`[A, B any]`) lets the
signature refer to types supplied at the call site, so the compiler enforces
what `NewFunc`'s reflection used to check by hand, before the program runs
rather than the first time it's called.

## 2. Type inference

**Before:** nothing to infer — `interface{}` has no type argument to leave
out.

**After:** `fpgen.Map(vs, strings.ToUpper)` compiles with no type arguments
written anywhere; the compiler infers `A` and `B` from `vs`'s and
`strings.ToUpper`'s types. `fpgen.None[int]()` (`fpgen/maybe.go`) is the
boundary: `None` takes no arguments at all, so there is nothing for
inference to look at, and the type argument must be given explicitly.
Inference works from the types of arguments (and, since Go 1.21, from
context in some assignments); it never works from a return type alone with
no other clue. That is the rule that trips people up in practice — not "does
inference work," but "does *this* call give the compiler enough to look at."

## 3. Compose, and the check moving to compile time

**Before:** `fp.Compose` (`fp/func.go`), the talk's punchline. Its entire
body is one comparison:

```go
if g.out != f.in {
	return nil, fmt.Errorf("can't compose: %v != %v", g.out, f.in)
}
```

— two `reflect.Type` values, reconstructed by hand, compared at the moment
`Compose` is *called*. It is the static type system, reimplemented in
`reflect.Type` equality, because the `interface{}` signature threw the real
one away.

**After:** `fpgen.Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C`
(`fpgen/func.go`). There is no error return, because there is no way to call
it wrong:

```go
Compose(strings.ToUpper, func(n int) int { return n })
// fpgen/testdata/compose_mismatch.go:18:27: in call to Compose, type func(n int) int
// of func(n int) int {…} does not match inferred type func(int) string for func(A) B
```

That is the real compiler output — `TestWallComposeMismatch` in
`fpgen/wall_test.go` runs `go build` on
`fpgen/testdata/compose_mismatch.go` and pins it, so this document can't
drift from what the compiler actually says. Put next to `fp.Compose`'s
`fmt.Errorf("can't compose: %v != %v", ...)`, this is the talk's whole
argument landing as two side-by-side diagnostics: one written by a person at
run time, one produced by the compiler before the program exists.

## 4. Generic types

**Before:** `fp.List` (`fp/list.go`) — `Head interface{}`, `Tail *List` —
holds anything, and gives back `interface{}`, so every read needs a type
assertion.

**After:** `fpgen.List[T any]` (`fpgen/list.go`):

```go
type List[T any] struct {
	Head T
	Tail *List[T]
}
```

**Concept:** a type parameter on a type declaration, instantiated at use
(`List[string]`, `*List[List[int]]`). Inside the type's own body, `T` is
just an ordinary in-scope identifier — the recursive field is `Tail
*List[T]`, not `Tail *List`, because a bare `*List` would be missing a type
argument.

## 5. THE FIRST WALL — methods cannot have type parameters

**Before:** slide 36 asks, of `fp.List.Map`, "Should this be a method? Of
what?" and slide 37 answers by making it one:
`func (l *List) Map(f *Func) *List`.

**After — the wall:** the direct generic transliteration does not compile:

```go
func (l *List[A]) Map[B any](f func(A) B) *List[B] { ... }
```

Building `fpgen/testdata/wall_method_type_param.go` under this module's
declared `go 1.21` (`go.mod`) gives:

```
fpgen/testdata/wall_method_type_param.go:16:23: generic method requires go1.27 or later (-lang was set to go1.21; check go.mod)
```

`TestWallMethodTypeParams` in `fpgen/wall_test.go` runs `go build
-gcflags=-lang=go1.21` on that exact file and pins that exact diagnostic, so
this claim is a checked fact, not an assertion.

**The nuance, stated precisely because it changes the shape of the lesson:**
Go 1.27 *did* lift this restriction — generic methods are real Go as of that
release, gated behind declaring `go 1.27` or later in `go.mod`. This
repository's `go.mod` deliberately stays at `go 1.21` (a standing constraint
— see `AGENTS.md`), so the wall is real *for this repository*, and for
every module that has not raised its floor past 1.27. It is not, as of this
writing, a permanent feature of the Go language — it is a feature of *this
module's declared floor*, which is exactly why `-lang=go1.21` is pinned
explicitly in the test rather than left to whatever toolchain happens to be
installed: the day the ambient default catches up to 1.27, an unpinned test
would silently stop demonstrating anything.

Slide 36's question — "should this be a method? of what?" — turns out to
have a precise, checkable answer even where the wall applies: **a method,
only when the element type does not change.** `fpgen.List[T].Reverse()`
(`fpgen/list.go`) needs no second type parameter and is a legal method;
`ListMap` (`A` in, `B` out) does, and is a free function. `fpgen.List`,
`fpgen.Maybe`, and `fpgen.Many` therefore still share no common `Map`
interface, for the same reason `fp`'s three containers don't (slide 47 —
see lesson 10) — just for a related but distinct reason now.

A smaller, secondary consequence worth naming: `fp.List.Map`, `fp.Maybe.Map`
and `fp.Many.Map` are three methods sharing one name because a method lives
in its receiver type's namespace. Free functions share one flat package
namespace with no overloading, so `fpgen` cannot give all three the name
`Map` — `Map` itself already names the slice function from lesson 1. They
are `ListMap`, `MaybeMap` (`fpgen/maybe.go`), and `ManyMap`
(`fpgen/many.go`).

## 6. Constraints

**Before:** `fp`'s unexported `quote` (`fp/list.go`) — a runtime type switch
on `interface{}`:

```go
func quote(v interface{}) string {
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}
```

**After, and the half-lesson people miss:** `fpgen.Quote[T any]`
(`fpgen/constraints.go`) keeps the *exact same shape*, dynamic assertion
included. `T any` is the constraint that promises nothing about `T`, so
there is nothing for the compiler to check and the run-time `any(v).(string)`
stays. **An unconstrained type parameter does not remove a run-time check —
it just spells `interface{}` as `any`.**

The other half is `fpgen.Max[T cmp.Ordered]` in the same file: `cmp.Ordered`
(stdlib, `go 1.21`) promises `>` is defined for `T`, so `Max`'s body needs no
type switch and no `reflect` import at all — a `T` that doesn't satisfy
`cmp.Ordered` fails to *instantiate* `Max[T]`, at the call site, before the
program runs. The same file also has `Sum[T Number]` (a union/approximation
constraint, `~int | ~int32 | ... `, so a defined type like `type Cents int`
still qualifies) and `Describe[T Named]` (a method constraint, `interface {
Name() string }`). Each is the same move: state, as a constraint, the one
fact the function actually needs, and let the compiler enforce it instead of
inspecting a `reflect.Kind` at run time. `Quote` is the honest counter-case
that keeps the lesson from overclaiming: a constraint only buys what it
actually promises, and `any` promises nothing.

## 7. The zero value

**Before, and this is the strongest lesson because it is a real bug on this
repo's `main` branch, not a hypothetical:** `fp.Maybe` (`fp/maybe.go`) has
no "present" field — it distinguishes present from missing by *inspecting*
`Value`: `nil` means missing, and (per the slide-78 fix) so does a typed nil
pointer, found via `reflect.ValueOf(x).IsNil()`. Its current `Map` is:

```go
func (m Maybe) Map(f *Func) Maybe {
	if m.Value == nil || isNilPtr(m.Value) {
		return Maybe{}
	}
	...
}
```

This short-circuits on **any** typed nil pointer, unconditionally — it never
asks whether the step it's about to skip could have handled one. `fpgen`'s
own test, `TestMaybeNilRegressionContrast` (`fpgen/maybe_test.go`),
demonstrates it directly: a `*box` with a nil-safe pointer-receiver method
(`func (b *box) Get() int { if b == nil { return -1 }; return b.n }`) holding
a nil receiver has a real, well-defined answer. `fp.Maybe{Value:
(*box)(nil)}.Do((*box).Get)` never calls it — the answer is discarded before
the step runs, and no error is reported; the call just returns `Maybe{}`. A
green pipeline did not catch this because no test had ever named the case;
it is tracked separately as `functional-go-maybe-nil` and is not fixed by
this package.

**After:** `fpgen.Maybe[T]` (`fpgen/maybe.go`):

```go
type Maybe[T any] struct {
	value T
	ok    bool
}
```

`ok` is the only thing `MaybeMap` and `Get` ever consult. There is no nil to
be typed or untyped, no pointer to dereference, no `reflect.Value.IsNil` to
call — the entire bug class is **unrepresentable**, not merely untested. A
present `*box` that happens to be nil is still present (`ok == true`); its
nil-safe method runs and its answer comes through, which is exactly what the
second half of `TestMaybeNilRegressionContrast` checks.

**Concept:** `var zero T` is a real value of `T` — `0`, `""`, a nil slice, a
zero-valued struct, whatever `T`'s zero value is — that a generic function
can produce without knowing anything else about `T`. `Maybe[T]`'s `value`
field holds exactly that zero value whenever `ok` is `false`, and nothing is
ever allowed to read it in that state. Compare to `fp.Maybe`, which has no
such value to fall back to and instead has to *interrogate* whatever is in
`Value` to guess whether it counts as "there."

## 8. FlatMap

**Before:** `fp.Many.Map` (`fp/many.go`) flattens implicitly via `toSlice`,
which inspects the *result*'s `reflect.Kind` — slice or array means
"multiply," anything else means "one cell." It can't tell `func(string)
string` from `func(string) []string` apart as *declared types*, and doesn't
need to, because `interface{}` erased that distinction before `toSlice` ever
ran.

**After:** `fpgen.ManyMap[A, B any](m *Many[A], f func(A) B) *Many[B]` and
`fpgen.FlatMap[A, B any](m *Many[A], f func(A) []B) *Many[B]`
(`fpgen/many.go`) are two different functions, because `func(A) B` and
`func(A) []B` are two different, fully distinct Go types — there is no
single signature that accepts either the way `fp.Many.Map`'s `interface{}`
parameter does. `TestFlatMap` (`fpgen/many_test.go`) and `ExampleFlatMap`
(`fpgen/example_test.go`) rebuild slide 68's chain — `strings.ToUpper`
(`func(string) string`, one cell, so `ManyMap`) then `strings.Fields`
(`func(string) []string`, many cells, so `FlatMap`) — and it still prints
`"HELLO", "THERE", "GOOD", "BYE"`.

**Concept:** a distinction the old code could stay silent about (does this
step expand its element into several?) becomes one the type system forces a
caller to *state*, once, at the call site — traded for having to know and
name which of the two shapes each step in a chain actually is.

## 9. THE SECOND WALL — heterogeneous chains

**Before:** `fp.Maybe.Do` and `fp.Many.Do` (`fp/maybe.go`, `fp/many.go`)
take `fs ...interface{}`, a chain where every step can have a *different*
type pair:

```go
w, err := Maybe{p}.Do(Person.Address, Address.City, City.Weather, Weather.Description)
```

`Person.Address` is `func(Person) *Address`; `Address.City` is
`func(*Address) City`; `City.Weather` is `func(City) *Weather` — three
distinct signatures, one variadic parameter.

**The wall:** there is no generic signature for that, because Go's
variadics are homogeneous by construction — `fs ...T` needs one `T` for
every element. None of the tempting shapes work:

```go
func Do[T any](v T, fs ...func(T) T) T           // forces every step to keep the same type
func Do[T, U any](v T, fs ...func(any) any) U    // back to interface{} -- the whole point was to avoid it
func Do[T, U any](v T, fs ...func(T) U) U        // every step must be the SAME func(T) U -- only a chain of length 1
```

**The honest alternatives, in `fpgen/chain.go`:**

- **Nested calls** — `MaybeMap(MaybeMap(MaybeMap(m, step1), step2), step3)`
  — correct, and unreadable past three steps.
- **A fixed-arity family** — `fpgen.Chain2[A, B, C any]` and
  `fpgen.Chain3[A, B, C, D any]`, readable, but the library has to predict
  the longest chain anyone wants and provide a `Chain` for it. `ExampleChain3`
  (`fpgen/example_test.go`) rebuilds the weather chain's shape this way. A
  fifth step needs a `Chain4` nobody has written yet — this does not
  generalise, and the doc comment on `Chain2`/`Chain3` says so.
- **Keep reflection** — `fp.Maybe.Do` itself, unchanged: arbitrary length,
  back to run-time type errors instead of compile errors.

This is where "Go's generics are not Haskell's" stops being an abstract
warning and becomes a concrete, checkable fact: monomorphic, homogeneous
variadics are a real design choice with a real cost, paid exactly at the
point `fp.Maybe.Do`'s heterogeneous chain lives.

## 10. THE THIRD WALL — no higher-kinded types

**Before:** slide 47. `fp`'s three containers can't share a `Mapper`
interface because the return type of `Map` can't be spelled:

```go
type Mapper interface {
	Map(*Func) ???
}
```

**After — still unspellable, in a new way:** with generics, the naive
attempt is

```go
type Mapper[A any] interface {
	Map[B any](f func(A) B) Mapper[B]
}
```

and it fails **twice over**: methods cannot have type parameters (lesson
5's wall, so `Map` can't be a method here either), and even granting that,
Go has no way to abstract over "some container type constructor `F`" so
that `Mapper[F, A]` could say `Map(func(A) B) F[B]` for whichever `F` — a
*higher-kinded type parameter*, a type parameter that is itself generic.
Go's type parameters range over types, never over generic type constructors.

So `fpgen.List`, `fpgen.Maybe`, and `fpgen.Many` share no interface, exactly
as `fp.List`, `fp.Maybe`, and `fp.Many` don't. Generics answers lesson 5's
version of the question precisely (a method, only when the element type
doesn't change) and leaves this one exactly where the talk found it. This is
the closing lesson and it is not hedged: **the talk's central complaint about
Go survives generics.**

## 11. What reflection still buys, and what it costs

**What `fp` gets that `fpgen` cannot, from the walls above:** arbitrary-arity
dispatch (`fp.Maybe.Do`'s variadic chain, lesson 9), chains built at run time
from data the program didn't know about at compile time (a config file
naming which methods to call, say), and the automatic pointer dereference in
`argValue` (`fp/func.go`) — which silently adapts `Person.Address`
(returning `*Address`) into `Address.City` (taking `Address`). Generics
forces that adapter to be written out and *seen*: `fpgen`'s
`Chain3`-based weather example (`ExampleChain3`) has an explicit
`func(a address) string { return a.city }` step where `fp`'s version has
nothing at all, because `argValue` papered over the exact same gap.

**What it costs, measured, not asserted:** `fpgen/bench_test.go` maps
`strings.ToUpper` over a 1000-element list both ways — `fp.List.Map` with a
`Must(NewFunc(...))`-built `*Func`, and `fpgen.ListMap` — matching `sum/`'s
own `benchSize = 1000` (`sum/sum_test.go`) so this repo compares reflection
against generics on the same shape of work it already compares recursion
against iteration on.

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkFPListMap` (`fp`, reflection) | 118,483 | 72,000 | 4,000 |
| `BenchmarkFPGenListMap` (`fpgen`, generics) | 25,174 | 32,000 | 2,000 |

Median of five runs, `go test ./fpgen -bench . -benchmem -count=5`, on an
Apple M4 Pro, Go 1.27 (see `docs/investigations/generics-vs-reflection.md`
for the full numbers and reproduction command, following the format
`docs/investigations/benchmark-input-size.md` set). `fpgen` is roughly 4.7×
faster and does half the allocations per element (`fp` allocates a `*List`
cell and boxes the `interface{}` conversion in `Func.Call`'s two directions;
`fpgen` allocates only the cell).

Against that speed and allocation win: no compile-time safety on the paths
`fp` covers and `fpgen` cannot (lessons 9-10), a panic instead of a returned
`error` when a chain built from `interface{}` doesn't type-check, and the
`reflect` package's own overhead on every call. Both are real; neither one
is "obviously better" in the abstract — which of the two costs matters more
depends entirely on whether the chain you need to express is one lessons 9
and 10 say generics can reach.

---

## Status

All twelve lessons above are implemented, tested (`go test ./fpgen/...`),
and pass `go vet` / `gofmt -l`. The benchmark in lesson 11 has been run and
its numbers recorded both here and in
`docs/investigations/generics-vs-reflection.md`.
