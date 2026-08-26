# What does reflection cost, next to generics?

`fp` and `fpgen` do the same job — map a function over a cons-cell list —
two different ways. This records the comparison, the same way
`benchmark-input-size.md` records `sum/`'s.

## What is measured

`BenchmarkFPListMap` and `BenchmarkFPGenListMap`, both in
`fpgen/bench_test.go`, apply `strings.ToUpper` to every element of a
1000-element list of the string `"hello"`:

- `BenchmarkFPListMap` builds the list as `*fp.List`, wraps `strings.ToUpper`
  once with `fp.Must(fp.NewFunc(strings.ToUpper))` outside the timed loop
  (matching how a real caller would use it — the `*Func` is built once, not
  per call), and calls `l.Map(f)` inside `b.N`.
- `BenchmarkFPGenListMap` builds the list as `*fpgen.List[string]` and calls
  `fpgen.ListMap(l, strings.ToUpper)` inside `b.N`, with no wrapping step at
  all — there is nothing to wrap.

1000 elements is not a new inference: it is `sum/sum_test.go`'s own
`benchSize`, reused so this comparison sits on the same input size `sum/`
already benchmarks reflection-adjacent code against, rather than picking a
fresh number this document would have to separately justify.

## Machine and command

Apple M4 Pro, Go 1.27, darwin/arm64. Reproduce with:

```sh
go test ./fpgen -bench . -benchmem -count=5
```

## Results

Median of five runs:

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkFPListMap` | 118,483 | 72,000 | 4,000 |
| `BenchmarkFPGenListMap` | 25,174 | 32,000 | 2,000 |

All five runs of each, for the record:

```
BenchmarkFPListMap-14       10117    116346 ns/op   72000 B/op   4000 allocs/op
BenchmarkFPListMap-14       10000    119464 ns/op   72000 B/op   4000 allocs/op
BenchmarkFPListMap-14       10000    118776 ns/op   72000 B/op   4000 allocs/op
BenchmarkFPListMap-14       10000    118063 ns/op   72000 B/op   4000 allocs/op
BenchmarkFPListMap-14        9870    118483 ns/op   72000 B/op   4000 allocs/op
BenchmarkFPGenListMap-14    46492     25035 ns/op   32000 B/op   2000 allocs/op
BenchmarkFPGenListMap-14    48031     25845 ns/op   32000 B/op   2000 allocs/op
BenchmarkFPGenListMap-14    46815     25231 ns/op   32000 B/op   2000 allocs/op
BenchmarkFPGenListMap-14    47385     25219 ns/op   32000 B/op   2000 allocs/op
BenchmarkFPGenListMap-14    47544     25174 ns/op   32000 B/op   2000 allocs/op
```

## Reading it

`fpgen` is about **4.7× faster** and allocates **half** as much per element.

Both allocate: 1000 elements means 1000 new `*List` cells either way, which
neither version can avoid — `List[T].Map`/`fpgen.ListMap` build a new list
rather than mutating in place, same as `fp.List.Map`. The difference is the
*second* allocation per element that `fp` pays and `fpgen` doesn't:
`fp.Func.Call` takes and returns `interface{}`, so passing `l.Head` in and
`f.Call(l.Head)`'s result back out both box a `string` into an `interface{}`
— two boxing allocations per element on top of the cell, versus `fpgen`'s
one. That accounts for the roughly 2× allocation gap (4000 vs 2000) directly;
the larger time gap (4.7×) also includes `reflect.Value.Call`'s own dispatch
overhead, which has no generics-side equivalent to compare against because
there is no reflection happening at all.

This is one workload — a `Map` over a list of strings, not a `Do` chain, not
`Many`'s flattening, not a program that builds its chain from data at run
time. See `docs/teaching-generics.md`, lesson 11, for what that workload
does and doesn't tell you about `fp.Maybe.Do`'s reflection-only chain
(lesson 9) or a program that genuinely needs to decide its call sequence
after compilation.
