# Functional Go?

A reconstruction of the code from **"Functional Go?"**, a talk by [Francesc Campoy](https://github.com/campoy) at [dotGo 2015](https://www.dotgo.eu/) in Paris, November 9 2015.

The talk asks what functional programming looks like in a language with no generics and no tail-call optimisation, and answers it by rebuilding `map` — and then functors — on top of `reflect` and `interface{}`.

Go has generics now. **[`fpgen/`](fpgen/)** is the same library rebuilt on type parameters, paired with **[`docs/teaching-generics.md`](docs/teaching-generics.md)**, a curriculum that teaches Go generics through the difference between the two packages — what each concept deletes from `fp`, and the two shapes (methods can't take type parameters; there's no way to abstract over a type constructor) that generics still cannot express. Read `fp/` first for the talk as given in 2015; read `fpgen/` and the curriculum for what changes with generics, and what doesn't.

## Status

**Complete.** `sum/`, the `fp/` library, `fpgen/`, and both examples are implemented and tested.

The original repo shown in the talk has been lost. What survives is the slide deck, and this repository is an attempt to recover the code from it. That means the code here is a *reconstruction*, not the 2015 original: where the slides are ambiguous or contain typos, choices have been made. Every such choice is recorded in [`NOTES.md`](NOTES.md).

## Source material

What the reconstruction is built *from*. These describe the 2015 talk, not this repository, and are never edited to match the code:

- [`docs/functional_go.pdf`](docs/functional_go.pdf) — the slide deck, 78 slides. Ground truth.
- [`docs/api-from-slides.md`](docs/api-from-slides.md) — every type and signature the deck defines, with slide references
- [`docs/prompt.md`](docs/prompt.md) — the reconstruction spec, transcribed from the deck
- Slides online: <https://speakerdeck.com/campoy/functional-go>
- Video: <https://www.youtube.com/watch?v=ouyHp2nJl0I>

## Reconstruction notes

What the reconstruction has *learned*. These describe this repository — the choices made, and the questions that came up along the way:

- [`NOTES.md`](NOTES.md) — the deviation log: every departure from the literal slide text, and why
- [`docs/investigations/`](docs/investigations/) — long-form findings, with the evidence needed to re-derive them:
  - [`tail-recursion.md`](docs/investigations/tail-recursion.md) — does Go eliminate tail calls? (No. Here is how that was established.)
  - [`benchmark-input-size.md`](docs/investigations/benchmark-input-size.md) — how many elements do the benchmarks sum? (The deck does not say; here is the bracketing argument for 1000.)

## Layout

```
fp/       Func, NewFunc, Must, Compose, List, Maybe, Many          (reflect, no generics — the talk as given)
fpgen/    Map, Filter, Reduce, Compose, List[T], Maybe[T], Many[T] (generics — see docs/teaching-generics.md)
sum/      SumI, SumR, SumTR, SumTRG + benchmarks
examples/
  weather/   Person -> Address -> City -> Weather   (Maybe)
  library/   Library -> Book -> Page -> Line        (Many)
```

`fp` is the library the talk builds:

- **`Func`** wraps an arbitrary function as a reflection-backed value carrying its input and output `reflect.Type`. `Compose` uses those types to reject mismatched pairs at composition time.
- **`List`** is a cons cell with a `Map` method — the starting point.
- **`Maybe`** short-circuits a chain the moment a value is nil — including a typed nil pointer, whether a step returned it or the `Maybe` was built around one — which is what makes nil-safe method chaining work.
- **`Many`** maps and flattens, so a step returning `[]string` chains cleanly onto a step returning `string`.

`Maybe` and `Many` each also have a `Do(fs ...interface{})` that builds the whole chain from plain functions and returns an error on any type mismatch.

## Running it

```sh
go build ./...
go test ./...
go test ./fp -run Example              # the slide outputs, as runnable examples
go test ./sum -bench . -benchmem       # the four benchmarks from the deck
go run ./examples/weather
go run ./examples/library
```

`go run ./examples/weather` walks a chain of pointers three ways — by hand, through `Maybe.Map`, and through `Maybe.Do` — for a person who has weather and two who do not:

```
a person in a sunny city:
	imperative (slide 58): sunny
	Map chain  (slide 61): sunny
	Do         (slide 63): sunny
a person with no address:
	imperative (slide 58): no weather
	...
```

`go run ./examples/library` counts the words in a small hardcoded library twice — with slide 71's four nested loops, and with the flat `Many` chain of slides 73 and 74 — and prints the two counts side by side. They agree, which is the point.

Every slide that shows output has a matching runnable example in [`fp/example_test.go`](fp/example_test.go), so the deck's printed results are checked by `go test` rather than taken on trust.

## Benchmarks

Four ways to sum a slice of ints: iterative, recursive, tail-recursive, and tail-recursive with the tail call flattened into a `goto` by hand — because Go does not do that for you.

The deck never states the input size — it shows only benchmark output, whose leading column is `b.N` rather than a size. This reconstruction sums 1000 elements, an inference [bracketed here](docs/investigations/benchmark-input-size.md).

| Benchmark | Talk, 2015 (4 cores) | Measured, Apple M4 Pro, Go 1.26 |
| --- | ---: | ---: |
| `BenchmarkSumI` | 462 ns/op | 236 ns/op |
| `BenchmarkSumR` | 4707 ns/op | 2575 ns/op |
| `BenchmarkSumTR` | 5056 ns/op | 2098 ns/op |
| `BenchmarkSumTRG` | 1587 ns/op | 297 ns/op |

Median of five runs, `go test ./sum -bench . -count=5`. All four allocate nothing.

The headline still holds a decade later: **recursion costs about 10× the loop, and writing it in tail form does not help, because Go does not eliminate tail calls.** You have to do the elimination yourself, and `SumTRG` — the same algorithm with the tail call rewritten as a `goto` — gets essentially all of it back.

Two things have changed since 2015, though:

- **Tail recursion is no longer the slowest of the four.** The talk measured `SumTR` as *slower* than plain `SumR` (5056 vs 4707 ns/op); today it comes out faster (2098 vs 2575). The slide's implicit "and it's even worse" no longer reproduces — see below for why.
- **The hand-optimised version has caught up.** In 2015 `SumTRG` was 3.4× the cost of `SumI`; here it is 1.26×. The reslicing that used to dominate is now nearly free.

### No, Go did not add tail-call elimination

The tempting reading of that first row is that the compiler has learned to eliminate tail calls in the decade since. It has not. `SumTR` still compiles to a real recursive `CALL` in a function with a 48-byte frame; it still grows the goroutine stack by one frame per element; and it still dies with `fatal error: stack overflow` on input that `SumTRG` sums without trouble.

What changed is the **calling convention**. `SumR` adds `vs[0]` *after* the recursive call returns, so the slice pointer and index must survive the call and get spilled to the stack. `SumTR` adds *before* the call, so nothing is live across it — the accumulator stays in a register. In 2015 that tradeoff ran the other way, because Go 1.5 passed every argument on the stack and `SumTR`'s extra accumulator cost a write per call. The register ABI made it free.

The talk's actual claim is untouched: Go still does not eliminate tail calls, and `SumTRG` is still 7× faster than `SumTR` for exactly that reason.

**[`docs/investigations/tail-recursion.md`](docs/investigations/tail-recursion.md)** has the whole investigation — how each check works and why it is conclusive, the assembly, the programs, and which parts are measured rather than inferred.

Reproduce with:

```sh
go test ./sum -bench . -benchmem -count=5
```

## Constraints

The reconstruction deliberately keeps the code as it would have been written in 2015:

- no generics — `reflect` and `interface{}` throughout
- no third-party dependencies in the library or examples themselves — `testify` is used in the tests, and nowhere else
- the API surface stays exactly as the slides show it

Modern Go would write most of this with type parameters in a fraction of the space. That is rather the point of the talk, and improving the code would erase it.

## Takeaways from the talk

Functional Go is doable, but slower than normal code and dependent on reflection. Its real value is as inspiration for good APIs: functors as a way to abstract function application — and an open question about what monads could offer.
