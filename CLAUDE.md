# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

A reconstruction of the source repo for Francesc Campoy's talk **"Functional Go?"** (dotGo 2015, Paris). The original repo is lost; the only surviving artifact is the slide deck (https://speakerdeck.com/campoy/functional-go).

`prompt.md` is the full specification: every code fragment transcribed from the deck, the target layout, and the known slide typos to correct. **Read `prompt.md` before writing code** — it is the ground truth, not a historical note.

As of the initial commit the repo contains only `prompt.md`. There is no Go code, no `go.mod`, no tests yet.

## Hard constraints (these are the point of the project, not preferences)

- **No generics, ever.** The talk exists because Go 1.5 had none. Use `reflect` and `interface{}`. Do not write type parameters. Do not spell `interface{}` as `any` — the 2015 spelling is deliberate.
- **Standard library only.** No third-party dependencies.
- Module path is `github.com/campoy/functional-go`. Target Go 1.5 *semantics*, but declare a modern `go` directive so it builds on current toolchains.
- **Do not rename anything.** The API surface (`Func`, `NewFunc`, `Must`, `Compose`, `List`, `Maybe`, `Many`, and the `Map`/`Do`/`Each` methods) must match the slides exactly, even where a modern name would read better.
- Improvements you're tempted to make belong in `NOTES.md`, not in the code.

`NOTES.md` is the required log of every deviation from the literal slide text (the `SumI`/`SumTR`/`Many.Map` typos, `strings.Field` → `strings.Fields`, validation the slides elided, and any ambiguity resolved by choice). Append to it whenever you deviate.

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

- `fp/func.go` — `Func` captures a function's `in`/`out` `reflect.Type` plus a closure that unwraps/rewraps `interface{}`. `NewFunc` validates via reflection (must be a func, exactly one in, exactly one out) and is the only place reflection on the *function* happens; everything downstream just calls `Func.Call`. `Compose` is the type check `g.out != f.in` — this is the static type system reimplemented at runtime, which is the talk's punchline.
- `fp/list.go`, `fp/maybe.go`, `fp/many.go` — three containers, each with the same `Map(f *Func)` shape:
  - `List` is a cons cell; `Map` recurses structurally.
  - `Maybe` short-circuits on `nil`, and critically also on a **typed nil pointer** (`vr.Kind() == reflect.Ptr && vr.IsNil()`). Without that reflection check the weather example breaks, since Go methods return typed nil pointers, not `nil` interfaces.
  - `Many` is a cons cell that *flattens*: `toSlice` turns a slice/array result into elements and any other value into a one-element slice. That single behaviour is what lets `strings.Fields` (`[]string`) chain after `strings.ToUpper` (`string`) without any per-type plumbing.
- `Do(fs ...interface{})` on `Maybe` and `Many` is the ergonomic layer: it calls `NewFunc` per step and folds `Map` across them, so type errors surface as a returned `error` at chain-build time rather than a panic.
- `sum/` is unrelated to `fp/` — it is the opening act on recursion vs. tail recursion, ending in `SumTRG`'s `goto`-faked TCO. Its benchmarks exist to be compared against the numbers the talk reported (~462/4707/5056/1587 ns/op); `README.md` carries both those and the numbers actually measured here.
- `examples/weather` and `examples/library` are the payoff. Both rely on **method expressions** (`Person.Address`, `Library.Books`) to feed `Do` — that's why the receiver types are values and the returns are pointers/slices. Each example shows the imperative version and the functional version side by side, and each has a test asserting the two agree.

## Working method

Build in dependency order, running `go test ./...` after each step: `sum` → `fp/func.go` → `List` → `Maybe` → `Many` → the two examples → the docs. Keep commits small, messaged by slide topic.

When a slide is ambiguous, pick the reading that makes the deck's usage snippets compile unchanged — those snippets are the ground truth, and they are quoted verbatim in `prompt.md`.
