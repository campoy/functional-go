# Functional Go?

A reconstruction of the code from **"Functional Go?"**, a talk by [Francesc Campoy](https://github.com/campoy) at [dotGo 2015](https://www.dotgo.eu/) in Paris, November 9 2015.

The talk asks what functional programming looks like in a language with no generics and no tail-call optimisation, and answers it by rebuilding `map` — and then functors — on top of `reflect` and `interface{}`.

## Status

**Reconstruction in progress.** `sum/` is implemented and tested. The `fp/` library and the two examples are not written yet.

The original repo shown in the talk has been lost. What survives is the slide deck, and this repository is an attempt to recover the code from it. That means the code here is a *reconstruction*, not the 2015 original: where the slides are ambiguous or contain typos, choices have been made. Every such choice is recorded in [`NOTES.md`](NOTES.md).

## Source material

- [`docs/functional_go.pdf`](docs/functional_go.pdf) — the slide deck, 78 slides
- [`docs/api-from-slides.md`](docs/api-from-slides.md) — every type and signature the deck defines, with slide references
- [`docs/prompt.md`](docs/prompt.md) — the reconstruction spec, transcribed from the deck
- Slides online: <https://speakerdeck.com/campoy/functional-go>
- Video: <https://www.youtube.com/watch?v=ouyHp2nJl0I>

## Planned layout

```
fp/       Func, NewFunc, Must, Compose, List, Maybe, Many
sum/      SumI, SumR, SumTR, SumTRG + benchmarks
examples/
  weather/   Person -> Address -> City -> Weather   (Maybe)
  library/   Library -> Book -> Page -> Line        (Many)
```

`fp` is the library the talk builds:

- **`Func`** wraps an arbitrary function as a reflection-backed value carrying its input and output `reflect.Type`. `Compose` uses those types to reject mismatched pairs at composition time.
- **`List`** is a cons cell with a `Map` method — the starting point.
- **`Maybe`** short-circuits a chain the moment a step yields nil, including a typed nil pointer, which is what makes nil-safe method chaining work.
- **`Many`** maps and flattens, so a step returning `[]string` chains cleanly onto a step returning `string`.

`Maybe` and `Many` each also have a `Do(fs ...interface{})` that builds the whole chain from plain functions and returns an error on any type mismatch.

## Running it

```sh
go test ./...                          # works today
go test ./sum -bench . -benchmem       # works today
go run ./examples/weather              # not written yet
go run ./examples/library              # not written yet
```

## Benchmarks

Four ways to sum a slice of ints: iterative, recursive, tail-recursive, and tail-recursive with the tail call flattened into a `goto` by hand — because Go does not do that for you.

The deck never states the input size. This reconstruction sums 1000 elements, which is consistent with the 462 ns/op reported for `SumI` on 2015 hardware.

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

**[`docs/tail-recursion.md`](docs/tail-recursion.md)** has the whole investigation — how each check works and why it is conclusive, the assembly, the programs, and which parts are measured rather than inferred.

Reproduce with:

```sh
go test ./sum -bench . -benchmem -count=5
```

## Constraints

The reconstruction deliberately keeps the code as it would have been written in 2015:

- no generics — `reflect` and `interface{}` throughout
- no third-party dependencies
- the API surface stays exactly as the slides show it

Modern Go would write most of this with type parameters in a fraction of the space. That is rather the point of the talk, and improving the code would erase it.

## Takeaways from the talk

Functional Go is doable, but slower than normal code and dependent on reflection. Its real value is as inspiration for good APIs: functors as a way to abstract function application — and an open question about what monads could offer.
