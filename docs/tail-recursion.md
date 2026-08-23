# Does Go eliminate tail calls?

No. This document records how that was checked, because the reconstruction's benchmark
numbers make it look like the answer might have changed since 2015.

## The question

Slides 12 through 15 of the talk rest on a single claim: Go does not perform tail-call
elimination, so writing a recursive function in tail form buys you nothing, and if you want
the optimisation you have to do it by hand. `SumTRG` — `SumTR` with the tail call rewritten
as a `goto` — is the punchline.

The talk's measurements back this up, and add a detail: `SumTR` came out not just no faster
than `SumR` but actually *slower*.

| Benchmark | Talk, 2015 (4 cores) | Measured, Apple M4 Pro, Go 1.26 |
| --- | ---: | ---: |
| `BenchmarkSumI` | 462 ns/op | 236 ns/op |
| `BenchmarkSumR` | 4707 ns/op | 2575 ns/op |
| `BenchmarkSumTR` | 5056 ns/op | 2098 ns/op |
| `BenchmarkSumTRG` | 1587 ns/op | 297 ns/op |

Reproducing the benchmarks a decade later, that last detail flips: `SumTR` is now *faster*
than `SumR`, 2098 against 2575 ns/op.

Which invites an obvious explanation — that the compiler has learned to eliminate tail
calls in the meantime, and slide 15's premise is stale. If that were true it would not be a
footnote. It would mean the reconstruction was faithfully reproducing code whose entire
motivation had expired, and `NOTES.md` would need to say so prominently.

So it was worth more than an assumption either way.

## What we checked, and why each check is conclusive

Three independent checks. The first inspects what the compiler emits; the other two observe
what the program does at runtime. They agree.

### 1. The generated assembly

If the compiler eliminated the tail call in `SumTR`, it would turn the call into a jump.
The function would then make no calls at all — it would be a *leaf* — and would need no
stack frame. That is a visible, unambiguous property of the emitted code.

```sh
go build -gcflags=-S ./sum 2>&1 | grep -E 'TEXT.*sum\.Sum'
```

```
TEXT sum.SumI(SB),   LEAF|NOFRAME|ABIInternal, $0-24
TEXT sum.SumR(SB),   ABIInternal,              $48-24
TEXT sum.SumTR(SB),  ABIInternal,              $48-32
TEXT sum.SumTRG(SB), LEAF|NOFRAME|ABIInternal, $0-32
```

`SumI` and `SumTRG` are `LEAF|NOFRAME` with a `$0` frame. `SumR` and `SumTR` are neither:
both carry a 48-byte frame. Counting the recursive calls confirms it directly — `SumR` and
`SumTR` each contain one `CALL` to themselves, `SumI` and `SumTRG` contain zero.

`SumTR`'s call site is a plain call-and-return, at `sum.go:57`:

```
0x0068 00104 (sum.go:57)  CALL  sum.SumTR(SB)
0x006c 00108 (sum.go:57)  MOVD  R0, R5
...
0x007c 00124 (sum.go:57)  RET   (R30)
```

Both functions also call `runtime.morestack_noctxt` in their prologue, which is how Go
functions participate in stack growth. A function compiled into a loop would not need to.

**Conclusion: the tail call is still a call.** `SumTRG` shows what elimination would look
like, and only the hand-written version gets it.

### 2. Stack consumption

The assembly says each recursive step should cost a stack frame. That is observable: a
tail-call-eliminated function runs in constant stack no matter the input size, while a real
recursive one grows the goroutine stack linearly.

The following program measures it. It runs each implementation on its own goroutine and
samples `runtime.MemStats.StackInuse` either side of the call.

```go
// Command stackprobe reports how much goroutine stack each sum implementation
// consumes. A tail-call-eliminated function runs in constant stack; a real
// recursive call burns one frame per element.
package main

import (
	"fmt"
	"runtime"

	"github.com/campoy/functional-go/sum"
)

func seq(n int) []int {
	vs := make([]int, n)
	for i := range vs {
		vs[i] = i
	}
	return vs
}

// stackDelta runs f on its own goroutine and reports the StackInuse growth
// attributable to it, in KB.
func stackDelta(n int, f func([]int)) int64 {
	vs := seq(n)
	done := make(chan int64)
	go func() {
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		f(vs)
		runtime.ReadMemStats(&after)
		done <- (int64(after.StackInuse) - int64(before.StackInuse)) / 1024
	}()
	return <-done
}

func main() {
	for _, n := range []int{100000, 1000000} {
		fmt.Printf("n = %d\n", n)
		fmt.Printf("  SumR   %6d KB\n", stackDelta(n, func(vs []int) { sum.SumR(vs) }))
		fmt.Printf("  SumTR  %6d KB\n", stackDelta(n, func(vs []int) { sum.SumTR(vs, 0) }))
		fmt.Printf("  SumTRG %6d KB\n", stackDelta(n, func(vs []int) { sum.SumTRG(vs, 0) }))
	}
}
```

```
n = 100000
  SumR     4096 KB
  SumTR    4096 KB
  SumTRG      0 KB
n = 1000000
  SumR    32768 KB
  SumTR   32768 KB
  SumTRG      0 KB
```

`SumR` and `SumTR` are indistinguishable, and both scale with the input: ten times the
elements, eight times the stack. `SumTRG` never leaves its initial stack at all.

**Conclusion: `SumTR` allocates a frame per element, exactly like plain recursion.**

### 3. Stack overflow

The strongest form of the same observation. Push the input past what the stack can hold and
a recursive implementation dies where a genuinely iterative one returns.

```go
debug.SetMaxStack(64 << 20) // 64 MB
vs := seq(5000000)
// ... call one of the three
```

```
--- SumR ---
runtime: goroutine stack exceeds 67108864-byte limit
runtime: sp=0x13d781e00380 stack=[0x13d781e00000, 0x13d785e00000]
fatal error: stack overflow

--- SumTR ---
runtime: goroutine stack exceeds 67108864-byte limit
runtime: sp=0x4bd3c3102380 stack=[0x4bd3c3102000, 0x4bd3c7102000]
fatal error: stack overflow

--- SumTRG ---
SumTRG = 12499997500000
```

**Conclusion: `SumTR` overflows the stack on input that `SumTRG` handles.** There is no
reading of that result compatible with the tail call having been eliminated.

## So why did the ordering flip?

Not tail calls — the **calling convention**.

Look at what each function does around its recursive call. `SumR` computes
`vs[0] + SumR(vs[1:])`. The addition cannot happen until the callee returns, so the slice
pointer and the index have to survive the call. The compiler spills both to the stack
beforehand and reloads them after:

```
MOVD  R3, ..autotmp_9-8(SP)   ; spill index
...
CALL  sum.SumR(SB)
MOVD  ..autotmp_9-8(SP), R3   ; reload index
MOVD  sum.vs(FP), R4          ; reload slice pointer
MOVD  (R4)(R3), R3            ; load vs[0] from memory
ADD   R0, R3, R3              ; and only now, the addition
```

`SumTR` computes `SumTR(vs[1:], s+vs[0])`. The addition happens *before* the call, so
nothing is live across it. The accumulator stays in a register, and the callee's return
value passes straight through to `RET`:

```
ADD   R5, R4, R3              ; accumulator, in a register
MOVD  R6, R0
CALL  sum.SumTR(SB)
MOVD  R0, R5                  ; pass the result through
MOVD  R5, R0
RET   (R30)
```

Zero spills, zero reloads. That is worth roughly 0.5 ns per element, which accounts for the
entire measured gap between the two.

In 2015 the same tradeoff ran the other way. Go 1.5 passed every argument on the stack, so
`SumTR`'s extra accumulator argument meant an extra stack write on every single call — a
cost `SumR` did not pay. That is the most likely reason the talk measured tail recursion as
the slowest of the three. The register-based ABI, introduced in Go 1.17 for amd64 and Go
1.18 for arm64, made the extra argument free and left `SumTR`'s no-spill property as a net
win.

## What is measured and what is inferred

Worth keeping separate:

- **Measured, on this machine:** everything in the three checks above. The absence of
  tail-call elimination is established, not surmised.
- **Inferred:** the explanation for the 2015 ordering. It follows from the assembly and
  from Go's documented switch to a register ABI, but it was not re-measured on a Go 1.5
  toolchain. If someone wants to close that gap, building `sum/` under Go 1.5 and
  re-reading the assembly would do it.

## What is deliberately not a test

None of this is checked in as a test, for a specific reason: **stack overflow in Go is a
fatal runtime error, not a recoverable panic.** `recover` does not catch it. A test
asserting that `SumTR` overflows would terminate the test binary rather than pass, taking
the rest of the suite with it. The stack-growth probe is merely slow and allocation-heavy
rather than fatal, but it measures the runtime rather than the package, which is not what
`sum`'s tests are for.

Hence this document: the commands and programs are here so the result can be re-derived on
demand, without `go test ./...` having to pay for it on every run.

## Bottom line

Slide 15's premise holds exactly as the talk stated it a decade ago. Go does not eliminate
tail calls. `SumTRG` is seven times faster than `SumTR` for precisely that reason, and the
only thing that has changed is *which* of the two recursive forms is marginally less bad.

## Environment

- Apple M4 Pro, `darwin/arm64`
- `go version go1.26.6 darwin/arm64`
- Benchmarks: medians of five runs, `go test ./sum -bench . -benchmem -count=5`, summing
  1000 elements (see `NOTES.md` for where that size comes from)
