# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

A reconstruction of the source repo for Francesc Campoy's talk **"Functional Go?"** (dotGo 2015, Paris). The code in the talk came from a repo the author has since lost; this repository recovers it from the slides.

### The documents, and which kind each one is

The docs fall into two kinds, and the difference matters when they disagree with each other or with the code. **Inputs** are what the reconstruction is built *from*; they describe the 2015 talk. **Outputs** are what the reconstruction has *learned*; they describe this repository. Never edit an input to match the code — that erases the evidence the project exists to preserve. Record the discrepancy in an output instead.

**Inputs**, in priority order:

1. **`docs/functional_go.pdf`** — the actual 78-slide deck. This is the ground truth, and it is immutable: do not regenerate, annotate, or "fix" it.
2. **`docs/api-from-slides.md`** — every type and signature the deck defines, with slide numbers, provenance marks, and the full list of slide errors. Derived mechanically from the deck. Use it as the conformance checklist when writing or reviewing code.
3. **`docs/prompt.md`** — a spec derived from transcribing the deck, with the intended repo layout and working method.

Where the deck and the derived documents disagree, the deck wins, but read `docs/prompt.md` too: it resolves ambiguities the deck leaves open (`NewMany`/`Each` signatures, benchmark shape) and specifies the layout.

**Outputs**, written as the work proceeds:

- **`NOTES.md`** — the deviation log. Every departure from the literal slide text, and why. One entry per decision.
- **`docs/investigations/`** — long-form findings that are too big for a `NOTES.md` bullet: a question that came up, how it was checked, and what the answer turned out to be. `tail-recursion.md` is the first (does Go eliminate tail calls? — no, and here is the proof).

The rule of thumb: if a `NOTES.md` entry needs more than a paragraph, or if someone would reasonably want to re-run your evidence, it belongs in `docs/investigations/` with a one-line summary and a link left behind in `NOTES.md`.

**Status:** complete. `sum/`, `fp/` and both examples are implemented and tested.

### Reading the deck

No PDF tooling is installed on this machine (`pdftotext`, `pdftoppm`, `mutool`, `qpdf`, `gs` are all absent), so the Read tool cannot render the PDF. The deck extracts cleanly as text via `pypdf` in a throwaway venv:

```sh
python3 -m venv /tmp/pdfvenv && /tmp/pdfvenv/bin/pip install -q pypdf
/tmp/pdfvenv/bin/python -c "
from pypdf import PdfReader
for i,p in enumerate(PdfReader('docs/functional_go.pdf').pages,1):
    print('===== SLIDE %d ====='%i); print(p.extract_text())"
```

Code slides lose their indentation in extraction but keep every token, which is enough. Don't install poppler system-wide without asking.

### Slide map

| Slides | Topic | Lands in |
| --- | --- | --- |
| 1–7 | Why functional, no mutable state | — |
| 8–16 | `SumI`, `SumR`, `SumTR`, `SumTRG` + benchmarks | `sum/` |
| 17–29 | First-class functions, `Map` on concrete types, why generics are missing | — (motivation) |
| 30–34 | `Func`, `NewFunc`, `Must` | `fp/func.go` |
| 35–38 | `List` and its `Map` | `fp/list.go` |
| 39–50 | Typeclasses interlude, `Mapper`, Functor | — (motivation) |
| 51–55 | `Maybe`, chaining | `fp/maybe.go` |
| 56–63 | Weather use case, method expressions, `Maybe.Do` | `examples/weather` |
| 64–68 | `Many`, `toSlice` flattening | `fp/many.go` |
| 69–74 | Library word count, `Many.Do`, `Each` | `examples/library` |
| 75–76 | Conclusions | `README.md` |
| 77–78 | Appendix: `Compose`, nil-pointer-aware `Maybe.Map` | `fp/func.go`, `fp/maybe.go` |

Slides 77–78 are appendix material but hold the **real** implementations — use them, not the simplified bodies from the talk proper.

## Hard constraints (these are the point of the project, not preferences)

- **No generics, ever.** The talk exists because Go 1.5 had none. Use `reflect` and `interface{}`. Do not write type parameters. Do not spell `interface{}` as `any` — the 2015 spelling is deliberate.
- **Standard library only in non-test code.** `fp/`, `sum/` and the examples' non-test code must not import anything outside the standard library. Tests may use `github.com/stretchr/testify` (`require` for preconditions that must abort, `assert` for independent checks); it is the only third-party dependency in `go.mod`, and it is test-only.
- Module path is `github.com/campoy/functional-go`. Target Go 1.5 *semantics*, but declare a modern `go` directive so it builds on current toolchains.
- **Do not rename anything.** The API surface (`Func`, `NewFunc`, `Must`, `Compose`, `List`, `Maybe`, `Many`, and the `Map`/`Do`/`Each` methods) must match the slides exactly, even where a modern name would read better.
- Improvements you're tempted to make belong in `NOTES.md`, not in the code.

## Known slide errors

The deck was written to be presented, not compiled. `docs/api-from-slides.md` has the full table; the short version is that `SumI` returns the wrong variable, `SumR`/`SumTR` recurse into a nonexistent `Sum`, `List.Map` calls the function form instead of the method, `Weather.Desc` is declared but `Weather.Description` is called, `Many.Map` assigns to an undefined `r`, `Many` is built with a struct literal that cannot compile, and `strings.Field` should be `strings.Fields`.

Fix each silently in code and append it to `NOTES.md`, the deviation log. It is already seeded with everything decidable from the deck alone; add to it as implementation forces further choices.

## Commands

```sh
go build ./...
go test ./...
go test ./fp -run TestMaybe          # single test
go test ./fp -run Example            # runnable examples in fp/example_test.go
go test ./sum -bench . -benchmem     # the four benchmarks from the deck
gofmt -l .                           # must print nothing
go vet ./...
go run ./examples/weather
go run ./examples/library
```

`gofmt -l .` (empty output), `go vet ./...`, and `go test ./...` all passing is the definition of done for any change.

## Architecture

The whole library is one idea applied three times: **`Func` is a runtime-typed, reflection-backed function value, and each container knows how to `Map` one over itself.**

Slides 24–29 are the setup: with generics you'd write `Map(f func(a α) β, vs []α) []β`; without them every parameter collapses to `interface{}` and all type safety is lost. `Func` is the answer — it carries the `reflect.Type` information that the signature threw away.

- `fp/func.go` — `Func` captures a function's `in`/`out` `reflect.Type` plus a closure that unwraps/rewraps `interface{}`. `NewFunc` validates via reflection and is the only place reflection on the *function* happens; everything downstream just calls `Func.Call`. `Compose` (slide 77) is the check `g.out != f.in` — the static type system reimplemented at runtime, which is the talk's punchline.
- `fp/list.go`, `fp/maybe.go`, `fp/many.go` — three containers, each with the same `Map(f *Func)` shape. Slide 47 is the reason they can't share an interface: `Map(*Func) ???` has no expressible return type in Go, so `Mapper` cannot be written. The duplication is the point, not an oversight.
  - `List` is a cons cell; `Map` recurses structurally.
  - `Maybe` short-circuits on `nil`, and critically also on a **typed nil pointer** (`vr.Kind() == reflect.Ptr && vr.IsNil()`, slide 78). Without that check the weather example breaks, since Go methods return typed nil pointers, not `nil` interfaces.
  - `Many` is a cons cell that *flattens*: `toSlice` turns a slice/array result into elements and any other value into a one-element slice. That single behaviour is what lets `strings.Fields` (`[]string`) chain after `strings.ToUpper` (`string`) without per-type plumbing.
- `Do(fs ...interface{})` on `Maybe` and `Many` is the ergonomic layer: it builds every step's `Func` up front, checks each step's `out` against the next step's `in`, then folds `Map` across them, so type errors surface as a returned `error` rather than a panic. Slide 62 only checks the former; the joint check is an addition, recorded in `NOTES.md`. The two containers need different notions of "fits" — `Many.Map` flattens with `toSlice` in between, `Maybe.Map` does not — which is why `canCall` and `canMap` are separate.
- **`Func.Call` dereferences a pointer argument when the wrapped function wants a value** (`argValue` in `fp/func.go`). The deck needs this and never mentions it: slide 61 feeds a `*Address` from `Person.Address` into `Address.City`, which has a value receiver, and `reflect.Value.Call` would panic. Do not remove it — the whole weather example depends on it. `Compose` deliberately keeps slide 77's stricter `g.out != f.in` identity check.
- **`Many.Map` walks each expansion backwards.** Slide 66's printed loop prepends while iterating forwards, which reverses every group and does not reproduce slide 68's printed output. The output is ground truth, the loop is not; `fp/example_test.go` pins the slide 68 order.
- `sum/` is unrelated to `fp/` — it is the opening act on recursion vs. tail recursion, ending in `SumTRG`'s `goto`-faked TCO. The deck never states the input size; ~1000 elements is consistent with the reported 462 ns/op for `SumI`, but that is an inference, so pick a size, record it, and report measured numbers next to the historical ones.
- `examples/weather` and `examples/library` are the payoff. Both rely on **method expressions** (`Person.Address`, `Library.Books` — slide 60, which the deck mislabels "Method values") to feed `Do`; that's why receivers are values and returns are pointers/slices. Each example shows the imperative version and the functional version side by side, with a test asserting the two agree.

## Working method

Run `go test ./...` after each step, and keep commits small, messaged by slide topic.

When a slide is ambiguous, pick the reading that makes the deck's usage snippets compile unchanged — those snippets (slides 38, 53–55, 61, 63, 67–68, 72–74) are the ground truth. Every one of them is pinned as a runnable `Example` in `fp/example_test.go`, so changing library behaviour that contradicts a slide fails the build.

## Maintaining this file

Keep this file for knowledge useful to almost every future agent session in this project.
Do not repeat what the codebase already shows; point to the authoritative file or command instead.
Prefer rewriting or pruning existing entries over appending new ones.
When updating this file, preserve this bar for all agents and keep entries concise.
