# Reconstruction notes

Where this repository departs from the literal text of the slides, and why.

The deck (`docs/functional_go.pdf`) was written to be projected, not compiled. Several snippets do not build as printed, and several functions are called but never defined. This file records every such decision so the reconstruction can be audited against the source. `docs/api-from-slides.md` holds the signature-by-signature checklist; this file holds the reasoning.

**Status:** `sum/` is implemented; `fp/` and the examples are not. Everything not marked as an implementation note was decided by reading the deck.

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
| 67, 68 | `Many{"hello there", "good bye"}` | `NewMany([]string{...})` | `Many`'s second field is `*Many`, so a two-string composite literal cannot compile |
| 67, 68 | `fmt.Println(res.Value)` | print the `*Many` itself | `Many` has `Head` and `Tail`; it has no `Value` field. Slides 67–68 were adapted from the `Maybe` slides and kept `.Value` by accident |
| 73, 74 | `strings.Field` | `strings.Fields` | `strings.Field` does not exist in the standard library |

Slides 8/9 and 10/12 are the two typos the original spec flagged. The rest were found by reading the deck directly.

## Signatures the slides use but never declare

These are reconstructions. They are the most likely places for this repository to diverge from the lost original, and each is a judgement call:

- **`func NewMany(v interface{}) *Many`** — called as `NewMany(l)` on a `Library` value (slide 72) and `NewMany(m)` (slides 73–74). Slide 68's usage requires it to accept a slice as well, so it takes `interface{}` and flattens a slice or array into cells, wrapping anything else as a single cell. This mirrors `toSlice`.
- **`func (m *Many) Do(fs ...interface{}) (*Many, error)`** — modelled on `Maybe.Do` (slide 62), which *is* shown. Slide 74 does `w, err := NewMany(m).Do(...)` and then `w.Each(...)`, which pins the first result to `*Many`.
- **`func (m *Many) Each(f interface{})`** — slide 74 passes `func(s string) { count[s]++ }`, a function with no return value. It therefore cannot go through `NewFunc`, which requires exactly one result, so `Each` does its own reflection. No error is returned because the slide ignores one; a bad argument panics.
- **`func toSlice(v interface{}) []interface{}`** — called on slide 66, never shown. Contract: a slice or array yields its elements, anything else yields a one-element slice. This single behaviour is what lets `strings.Fields` (returning `[]string`) chain after `strings.ToUpper` (returning `string`), which is the whole point of `Many`.
- **`func (l *List) String() string`** — never declared, but slide 38 calls `fmt.Println(res)` and shows `"HELLO", "BYE"`, which requires a `Stringer`. Format chosen to reproduce that output: quoted elements, comma-separated.

## Choices still open

To be resolved when the code is written, and recorded here:

- **`Compose` argument order.** Slide 77 checks `g.out != f.in` and returns `f.Call(g.Call(x))`, so `Compose(f, g)` applies `g` first. This is mathematical order, not pipeline order, and it reads backwards to most people. Kept as printed.
- **Whether `Maybe.Map` uses the slide 52 or slide 78 body.** Slide 78 (the appendix), because the simple nil check on slide 52 cannot handle the typed nil pointers that Go methods return, and the weather example depends on exactly that.

## Implementation notes

### `sum` (slides 8–15)

**Benchmark input size: 1000 elements.** The deck reports 462/4707/5056/1587 ns/op (slides 11, 13, 15) but never says how many elements are summed. 1000 is inference, not evidence: it is the size at which `SumI` costs about 462 ns on 2015 hardware. Recorded in `sum_test.go` as `benchSize`. If the real figure ever surfaces, this is the one number to change.

**The talk's ordering does not fully reproduce.** On an M4 Pro under Go 1.26, `SumTR` (2098 ns/op) comes out *faster* than `SumR` (2575), where the talk had it slower (5056 vs 4707). The conclusion the slides draw is unaffected, but the specific "tail recursion is the worst of the three" reading of slide 13 is now wrong. `README.md` reports both columns rather than quietly substituting today's numbers.

**The reversal is not tail-call elimination.** The obvious explanation — that Go has since learned to eliminate tail calls — was checked and is false. `SumTR` still compiles to a real recursive `CALL`, still burns a stack frame per element, and still overflows the stack on input that `SumTRG` handles. What changed is the calling convention: Go 1.5 passed arguments on the stack, making `SumTR`'s extra accumulator cost a write per call, while the register ABI made it free and left `SumTR`'s no-spill-across-the-call property as a net win. Slide 15's premise is untouched.

The full investigation — the three checks, the commands and programs to re-run them, the assembly, and what is measured versus inferred — is in [`docs/tail-recursion.md`](docs/tail-recursion.md).

`SumTRG` has also closed most of its gap to `SumI` since 2015: 1.26× rather than 3.4×.

**Callers seed the accumulator with 0.** `SumTR` and `SumTRG` take `(vs []int, s int)` as the slides show, so unlike `SumI` and `SumR` they cannot be called with just a slice. The tests wrap them to compare all four uniformly. Keeping the two-argument form matters — it is what makes the recursive call a tail call, which is the whole subject of slides 12–15.

**Tests cover the three slide typos directly.** `TestSumAgree` compares all four implementations on random input; reintroducing any of the slide's errors makes it fail rather than silently returning a wrong answer.

## Modernizations deliberately refused

The talk exists because Go 1.5 had no generics. Rewriting any of this with type parameters would erase the subject matter, so:

- no type parameters anywhere, however much `Map` is asking for them
- `interface{}` is never spelled `any` — the 2015 spelling is part of the artifact
- no third-party dependencies
- no renaming for taste, even where the slide names are awkward (`Func`, `Many`, `Do`)

A generic version of this library would be perhaps a fifth of the size and fully type-safe at compile time. That comparison is the point of preserving this one, not a reason to change it.
