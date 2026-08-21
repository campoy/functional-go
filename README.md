# Functional Go?

A reconstruction of the code from **"Functional Go?"**, a talk by [Francesc Campoy](https://github.com/campoy) at [dotGo 2015](https://www.dotgo.eu/) in Paris, November 9 2015.

The talk asks what functional programming looks like in a language with no generics and no tail-call optimisation, and answers it by rebuilding `map` — and then functors — on top of `reflect` and `interface{}`.

## Status

**Reconstruction in progress.** This repository currently contains the source material only; the Go code is not written yet.

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

Once the code exists:

```sh
go test ./...
go run ./examples/weather
go run ./examples/library
go test ./sum -bench . -benchmem
```

## Benchmarks

Four ways to sum a slice of ints: iterative, recursive, tail-recursive, and tail-recursive with the tail call flattened into a `goto` by hand — because Go does not do that for you.

The numbers reported in the talk, on a 4-core machine in 2015:

| Benchmark | Historical (ns/op) | Measured here |
| --- | --- | --- |
| `BenchmarkSumI` | 462 | _pending_ |
| `BenchmarkSumR` | 4707 | _pending_ |
| `BenchmarkSumTR` | 5056 | _pending_ |
| `BenchmarkSumTRG` | 1587 | _pending_ |

The shape is the interesting part: tail recursion is *slower* than plain recursion in Go, and faking the optimisation by hand recovers most, but not all, of the gap to the loop.

The deck does not state the input size, so the reconstruction picks one and records it in [`NOTES.md`](NOTES.md). The measured column will be filled in against the same benchmark shape once `sum/` is written.

## Constraints

The reconstruction deliberately keeps the code as it would have been written in 2015:

- no generics — `reflect` and `interface{}` throughout
- no third-party dependencies
- the API surface stays exactly as the slides show it

Modern Go would write most of this with type parameters in a fraction of the space. That is rather the point of the talk, and improving the code would erase it.

## Takeaways from the talk

Functional Go is doable, but slower than normal code and dependent on reflection. Its real value is as inspiration for good APIs: functors as a way to abstract function application — and an open question about what monads could offer.
