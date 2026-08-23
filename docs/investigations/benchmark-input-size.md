# How many elements do the benchmarks sum?

The deck does not say. This document records what it does show, why the obvious
place to look is a dead end, and how `benchSize = 1000` was arrived at — so that
the next person to wonder does not have to redo it.

## What the deck actually shows

Slides 11, 13 and 15 are the same slide accumulating rows. By slide 15:

```
PASS
BenchmarkSumI-4    3000000   462 ns/op
BenchmarkSumR-4     300000  4707 ns/op
BenchmarkSumTR-4    300000  5056 ns/op
BenchmarkSumTRG-4  1000000  1587 ns/op
```

That is the whole of it. Searching all 78 slides for `b.N`, `testing.B`, `Benchmark`,
`make([]int` or `rand.` turns up these three output slides and nothing else — **the deck
never shows the benchmark source**. The four function bodies are on slides 8, 10, 12 and
14; the code that drives them was never projected.

So the input size is not recorded anywhere in the surviving material.

## The first column is `b.N`, not the input size

It is tempting to read `3000000` as the number of elements summed. It is not: it is the
iteration count `go test -bench` chose, which it prints to the left of `ns/op`.

That column also carries no hidden information about the input, because it is entirely
determined by `ns/op`. Go's benchmark driver raises `b.N` until the run takes about a
second. Multiplying the two columns back out:

| Benchmark | `b.N` | ns/op | total |
| --- | ---: | ---: | ---: |
| `SumI` | 3,000,000 | 462 | 1.386 s |
| `SumR` | 300,000 | 4707 | 1.412 s |
| `SumTR` | 300,000 | 5056 | 1.517 s |
| `SumTRG` | 1,000,000 | 1587 | 1.587 s |

All four land just over one second, which is the default `-benchtime`. Every `b.N` here is
exactly what the driver would pick given that `ns/op` — so the column is a consequence of
the timing, not an independent measurement, and it constrains the input size not at all.

(The `-4` suffix is `GOMAXPROCS`, i.e. the 4-core machine of 2015. Also not a size.)

## Bracketing the size from the numbers that are there

The size is unrecoverable, but it is not unconstrained. `SumR` differs from `SumI` in
almost exactly one respect: one function call per element. So the gap between them,
divided by the input size, is the per-call overhead — and per-call overhead on 2015
hardware is something we can sanity-check.

The deck's gap is `4707 − 462 = 4245 ns`. Dividing by candidate sizes:

| Assumed size | Implied ns per recursive call | Verdict |
| ---: | ---: | --- |
| 500 | 8.49 | Implausibly slow — several times a 2015 call/return |
| **1000** | **4.25** | Plausible |
| 2000 | 2.12 | Implausibly fast — quicker than an M4 Pro manages today |

For comparison, this reconstruction measures `2575 − 236 = 2339 ns` at 1000 elements, or
**2.34 ns per recursive call** on an Apple M4 Pro under Go 1.26. Against that, 4.25 ns on a
2015 4-core machine is a believable generational gap of roughly 1.8×. The 2000-element
reading would require 2015 hardware to have been *faster* than current hardware at the one
operation being measured, which rules it out; the 500-element reading needs a call/return
several times more expensive than anything plausible for the era.

A cross-check on the loop itself: 462 ns for 1000 elements is 0.46 ns per element for
`SumI`, or roughly one element per clock at 2 GHz. That is the right order for a tight
integer sum loop, and it is consistent rather than merely non-contradictory.

## Conclusion

**`benchSize = 1000`**, recorded in `sum_test.go`. It is inference from the reported
timings, not evidence — the deck simply does not contain the answer. But it is bracketed
inference rather than a guess: 500 and 2000 both imply per-call costs that do not survive
contact with what the hardware of either era can do.

## How to overturn this

Any of the following would settle it, and all beat re-deriving the arithmetic above:

- The author remembering the number.
- The original repo resurfacing — it is lost, which is why this reconstruction exists.
- A recording of the talk where the size is stated aloud rather than projected. The video
  is linked from `README.md`; the benchmark slides go by quickly, and it has not been
  checked.

If the real figure turns up, `benchSize` in `sum_test.go` is the single constant to change,
and the measured column in `README.md` should be re-run against it.
