# Reconstruction notes

Where this repository departs from the literal text of the slides, and why.

The deck (`docs/functional_go.pdf`) was written to be projected, not compiled. Several snippets do not build as printed, and several functions are called but never defined. This file records every such decision so the reconstruction can be audited against the source. `docs/api-from-slides.md` holds the signature-by-signature checklist; this file holds the reasoning.

**Status:** `sum/`, `fp/` and both examples are implemented. Everything not marked as an implementation note was decided by reading the deck.

## Deviations from the literal slide text

| Slide | Deck | Reconstruction | Why |
| --- | --- | --- | --- |
| 8, 9 | `SumI` ends `return v` | `return s` | `v` is the loop variable and is out of scope after the `range`; `s` is the accumulator |
| 10 | `return vs[0] + Sum(vs[1:])` | `SumR(vs[1:])` | `Sum` is never defined anywhere in the deck |
| 12 | `return Sum(vs[1:], s+vs[0])` | `SumTR(vs[1:], s+vs[0])` | same; the two-argument shape only matches `SumTR` |
| 31 | `// check type of f and return an error if needed` | real validation | `NewFunc` returns an `error`; leaving the check out makes the signature a lie. Rejects a non-function, a function without exactly one parameter, and one without exactly one result |
| 37 | method body calls `Map(f, l.Tail)` | `l.Tail.Map(f)` | leftover from the package-level form on slide 35, which slide 36 explicitly converts to a method |
| 57 | `func (w Weather) Desc() string` | `func (w Weather) Description() string` | the deck contradicts itself: slides 58, 61, and 63 all call `Weather.Description`. Three uses beat one declaration |
| 66 | `r = &Many{v, res}` | `res = &Many{v, res}` | `r` is undefined; the loop has to accumulate into `res` or the results are discarded |
| 66 | `for _, v := range toSlice(...)` | walk the expansion backwards | see [Many.Map walks each expansion backwards](#manymap-walks-each-expansion-backwards) below |
| 67, 68 | `Many{"hello there", "good bye"}` | `NewMany([]string{...})` | `Many`'s second field is `*Many`, so a two-string composite literal cannot compile |
| 67, 68 | `fmt.Println(res.Value)` | print the `*Many` itself | `Many` has `Head` and `Tail`; it has no `Value` field. Slides 67–68 were adapted from the `Maybe` slides and kept `.Value` by accident |
| 73, 74 | `strings.Field` | `strings.Fields` | `strings.Field` does not exist in the standard library |

Slides 8/9 and 10/12 are the two typos the original spec flagged. The rest were found by reading the deck directly.

### `Many.Map` walks each expansion backwards

Slide 66 prints the loop as

```go
res := m.Tail.Map(f)
for _, v := range toSlice(f.Call(m.Head)) {
	res = &Many{v, res}
}
```

which walks the expansion forwards while *prepending*, and therefore reverses every group. Slide 68 chains `strings.Fields` after `strings.ToUpper` over `"hello there", "good bye"` and prints

```
"HELLO", "THERE", "GOOD", "BYE"
```

The printed loop yields `"THERE", "HELLO", "BYE", "GOOD"` instead. Walking the expansion back to front reproduces the slide, so that is what `Many.Map` does.

The printed output is ground truth; the printed loop is not. `fp/example_test.go` asserts the slide 68 order, so reverting the loop direction fails the build rather than quietly reordering everybody's results.

## Signatures the slides use but never declare

These are reconstructions. They are the most likely places for this repository to diverge from the lost original, and each is a judgement call:

- **`func NewMany(v interface{}) *Many`** — called as `NewMany(l)` on a `Library` value (slide 72) and `NewMany(m)` (slides 73–74). Slide 68's usage requires it to accept a slice as well, so it takes `interface{}` and flattens a slice or array into cells, wrapping anything else as a single cell. This mirrors `toSlice`.
- **`func (m *Many) Do(fs ...interface{}) (*Many, error)`** — modelled on `Maybe.Do` (slide 62), which *is* shown. Slide 74 does `w, err := NewMany(m).Do(...)` and then `w.Each(...)`, which pins the first result to `*Many`.
- **`func (m *Many) Each(f interface{})`** — slide 74 passes `func(s string) { count[s]++ }`, a function with no return value. It therefore cannot go through `NewFunc`, which requires exactly one result, so `Each` does its own reflection. No error is returned because the slide ignores one; a bad argument panics.
- **`func toSlice(v interface{}) []interface{}`** — called on slide 66, never shown. Contract: a slice or array yields its elements, anything else yields a one-element slice. This single behaviour is what lets `strings.Fields` (returning `[]string`) chain after `strings.ToUpper` (returning `string`), which is the whole point of `Many`.
- **`func (l *List) String() string`** — never declared, but slide 38 calls `fmt.Println(res)` and shows `"HELLO", "BYE"`, which requires a `Stringer`. Format chosen to reproduce that output: quoted elements, comma-separated. Strings are quoted with `%q`; anything else is printed with `%v`, since `%q` on an `int` would render it as a rune.
- **`func (m *Many) String() string`** — the same argument for `Many`: slides 67 and 68 call `fmt.Println` on the result of `Map` and show quoted, comma-separated output. It is the one signature missing from `docs/api-from-slides.md`, which has been amended to list it as **inferred** alongside `List.String`. That is an error in the checklist rather than a disagreement with the deck, which is why the checklist could be edited and the deck could not.

## Choices the deck left open

Both were flagged before the code was written; both were settled as follows.

- **`Compose` argument order.** Slide 77 checks `g.out != f.in` and returns `f.Call(g.Call(x))`, so `Compose(f, g)` applies `g` first. This is mathematical order, not pipeline order, and it reads backwards to most people. Kept as printed.
- **Whether `Maybe.Map` uses the slide 52 or slide 78 body.** Slide 78 (the appendix), because the simple nil check on slide 52 cannot handle the typed nil pointers that Go methods return, and the weather example depends on exactly that.

## Implementation notes

### `sum` (slides 8–15)

**Benchmark input size: 1000 elements.** The deck reports 462/4707/5056/1587 ns/op (slides 11, 13, 15) but never says how many elements are summed, and never shows the benchmark source at all — only its output. The leading column on those slides (3000000, 300000, 1000000) is `b.N`, the iteration count, not the input size; it is fully determined by `ns/op` and the default one-second `-benchtime`, so it constrains the size not at all.

1000 is inference, not evidence, but it is bracketed: the `SumR`−`SumI` gap implies 4.25 ns per recursive call at that size, against 2.34 ns measured here on current hardware. 500 would require an implausibly slow call and 2000 an implausibly fast one. Recorded in `sum_test.go` as `benchSize`; if the real figure ever surfaces, it is the one number to change. Full working in [`docs/investigations/benchmark-input-size.md`](docs/investigations/benchmark-input-size.md).

**The talk's ordering does not fully reproduce.** On an M4 Pro under Go 1.26, `SumTR` (2098 ns/op) comes out *faster* than `SumR` (2575), where the talk had it slower (5056 vs 4707). The conclusion the slides draw is unaffected, but the specific "tail recursion is the worst of the three" reading of slide 13 is now wrong. `README.md` reports both columns rather than quietly substituting today's numbers.

**The reversal is not tail-call elimination.** The obvious explanation — that Go has since learned to eliminate tail calls — was checked and is false. `SumTR` still compiles to a real recursive `CALL`, still burns a stack frame per element, and still overflows the stack on input that `SumTRG` handles. What changed is the calling convention: Go 1.5 passed arguments on the stack, making `SumTR`'s extra accumulator cost a write per call, while the register ABI made it free and left `SumTR`'s no-spill-across-the-call property as a net win. Slide 15's premise is untouched.

The full investigation — the three checks, the commands and programs to re-run them, the assembly, and what is measured versus inferred — is in [`docs/investigations/tail-recursion.md`](docs/investigations/tail-recursion.md).

`SumTRG` has also closed most of its gap to `SumI` since 2015: 1.26× rather than 3.4×.

**Callers seed the accumulator with 0.** `SumTR` and `SumTRG` take `(vs []int, s int)` as the slides show, so unlike `SumI` and `SumR` they cannot be called with just a slice. The tests wrap them to compare all four uniformly. Keeping the two-argument form matters — it is what makes the recursive call a tail call, which is the whole subject of slides 12–15.

**Tests cover the three slide typos directly.** `TestSumAgree` compares all four implementations on random input; reintroducing any of the slide's errors makes it fail rather than silently returning a wrong answer.

### `fp` (slides 30–38, 51–55, 62, 65–68, 77–78)

**`NewFunc` rejects variadic functions.** Slide 31 elides the validation entirely, so this is a free choice. `reflect` reports the input type of `func(vs ...string) string` as `[]string`, while `reflect.Value.Call` will happily accept a single `string` — so admitting variadics would leave `Compose` and `Do` type-checking against a type the function never actually sees, and `Func.in` would be a small lie. Rejecting them costs nothing: no function in the deck is variadic. The alternative, treating the variadic parameter as the single argument and setting `in` to its element type, was considered and dropped as more machinery than the talk's material justifies.

The same paragraph of validation also rejects a nil `interface{}` and a typed nil function value (`var f func(string) string`), both of which pass the kind and arity checks and then panic inside `reflect.Value.Call`. Slide 31 says nothing about either; returning the error the signature already promises is cheaper than the panic.

**`Func.Call` dereferences a pointer argument when the function wants a value.** This is a deviation the deck needs but never mentions. Slide 61 chains

```go
Map(Must(NewFunc(Person.Address))).
Map(Must(NewFunc(Address.City)))
```

`Person.Address` is `func(Person) *Address`, but `Address.City` is `func(Address) *City` — a value receiver. The chain therefore hands a `*Address` to a function wanting an `Address`, and `reflect.Value.Call` panics on exactly that. Every step of the weather example after the first has this shape, because slide 57 declares value receivers returning pointers.

The alternatives were worse. Declaring pointer receivers instead would contradict slide 57 and break the method expressions slide 60 is specifically about; dereferencing inside `Maybe.Map` would contradict slide 78, which stores `r` — the pointer — unchanged. So `argValue` in `fp/func.go` dereferences a pointer whose element type is assignable to the parameter type, and `Compose`'s and `Do`'s type checks allow for it. A nil pointer becomes the zero value rather than a panic, though `Maybe.Map` short-circuits before that can ever be reached.

`Compose` is the exception: slide 77's `g.out != f.in` is printed in full and is kept exactly as an identity check, since it is appendix material that nothing else calls. That asymmetry is deliberate.

**`Maybe.Do` and `Many.Do` type-check the whole chain before applying any of it.** Slide 62 builds the chain one step at a time and returns only the errors coming out of `NewFunc`, so a step whose result does not fit the next step's argument panics inside `reflect.Value.Call` rather than returning the `error` the signature promises. Both `Do` methods here build every `Func` first, compare each step's `out` against the next step's `in`, and return a descriptive error on a mismatch. The check also covers the *starting* value — `Maybe.Value`, and every cell already in a `Many` — against the first step's `in`, since a seed that does not fit step 0 panics in exactly the same place a bad joint does. A nil starting value is not an error: `Maybe` short-circuits on it, and `argValue` turns an invalid value into the zero value.

This is an addition beyond the deck, not a correction of it — the slide compiles and works for the chains it shows. It is justified by the signature: a `Do` that returns an `error` and then panics on the most likely error is not much of a `Do`. `fp/maybe_test.go` and `fp/many_test.go` each feed a deliberately mismatched chain and assert an error rather than a panic.

The two containers need different notions of "fits", because their `Map`s differ: `Maybe` passes the result straight along, while `Many` flattens it with `toSlice` first. So `Many.Do` also accepts a step returning `[]Book` feeding a step taking a `Book`, which is what makes the library chain on slide 74 type-check at all.

**`Each` panics rather than returning an error.** Slide 74 calls `w.Each(func(s string) { count[s]++ })` and ignores any result, which fixes the signature. The panics are descriptive, and the validation mirrors `NewFunc`'s except that exactly *zero* results are required.

### `examples/weather` (slides 56–63)

**Three implementations, three invented names.** The deck writes all three as `func (p Person) Weather() string`, on slides 58, 61 and 63. Go will not accept three methods of the same name on `Person`, and comparing the three is the entire point of that run of slides, so they are `WeatherImperative`, `WeatherMap` and `WeatherDo` here. All three return `"no weather"` when the chain breaks.

They are plain functions taking a `Person` rather than methods with invented names, which keeps `Person`'s method set to exactly what slide 57 declares — `Address`, and the `Weather` below. `func (p Person) Weather() string` survives as a one-line wrapper around `WeatherDo`, so the checklist's verbatim entry stays satisfied.

### `examples/library` (slides 69–74)

**`WordCountImperative` and `WordCountFunctional` are invented names.** Slides 71 and 74 show the two bodies as bare statements, with no enclosing function at all. The names exist so `main` can print both counts side by side and a test can assert they agree.

**Slide 73's extra step is dropped.** Slide 73 ends the chain with `func(s string) bool { count[s]++; return true }` — a counting step disguised as a mapped function, returning a `bool` nobody reads, purely because `Map` requires exactly one result. Slide 74 then introduces `Each` to do the same thing honestly. The example implements slide 74's version; slide 73 reads as the intermediate step that motivates it.

## Continuous integration

`.github/workflows/ci.yml` runs on pull requests targeting `main` and on pushes to `main`: `go build`, `go test`, `go test -race`, `gofmt -l`, `go vet`, and `staticcheck`. Build and test run on both Go 1.21 — the module's own `go` directive, kept as the floor — and current stable, so the reconstruction is known to build on the oldest toolchain it claims and the newest one available.

**Linter: `staticcheck`, pinned, with no suppressions.** Chosen over `golangci-lint` because it is a single tool with a single pinned version and needs no configuration file, which suits a repository whose defining constraint is that it has no third-party dependencies. It is installed as a CI tool only; it never enters `go.mod`.

The expectation was that a modern linter would object to the deliberate 2015-era style — `reflect` and `interface{}` throughout, no generics, slide-faithful names. It does not. staticcheck's default check set reports nothing, on `sum/` and on the reflection-heavy `fp/` code alike, so nothing is disabled and no `staticcheck.conf` exists. If a future check does fire on deliberate style, the exclusion belongs in a `staticcheck.conf` scoped as narrowly as the tool allows, with a comment naming the constraint it serves — not a blanket `//lint:ignore`.

Enabling the non-default `ST*` checks was tried and rejected: the only thing it surfaced was a doc-comment phrasing nit, which is not worth the maintenance.

## Modernizations deliberately refused

The talk exists because Go 1.5 had no generics. Rewriting any of this with type parameters would erase the subject matter, so:

- no type parameters anywhere, however much `Map` is asking for them
- `interface{}` is never spelled `any` — the 2015 spelling is part of the artifact
- no third-party dependencies
- no renaming for taste, even where the slide names are awkward (`Func`, `Many`, `Do`)

A generic version of this library would be perhaps a fifth of the size and fully type-safe at compile time. That comparison is the point of preserving this one, not a reason to change it.
