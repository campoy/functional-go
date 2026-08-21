# Claude Code prompt: reconstruct the "Functional Go" (dotGo 2015) repo

Paste everything below the line into Claude Code from an empty directory.

---

## Task

Reconstruct the source repository for the talk "Functional Go?", given by Francesc Campoy at dotGo 2015 (Paris, November 9 2015). The original repo is not publicly available. The only surviving artifact is the slide deck: https://speakerdeck.com/campoy/functional-go

Every code snippet below was transcribed from that deck. Your job is to turn those fragments into a complete, compiling, tested Go module that matches what the talk demonstrated.

## Hard constraints

1. **No generics.** The whole point of the talk is doing this with `reflect` and `interface{}` because Go had no generics in 2015. Do not "improve" the code with type parameters, and do not use `any` in place of `interface{}`. If you feel tempted, put the thought in `NOTES.md` instead.
2. **No third-party dependencies.** Standard library only.
3. Target `go 1.5` semantics but declare a modern `go` directive in `go.mod` so it builds today. Module path: `github.com/campoy/functional-go`.
4. Preserve the API surface exactly as it appears in the slides: `Func`, `NewFunc`, `Must`, `Compose`, `List`, `Maybe`, `Many`, and the `Map` / `Do` / `Each` methods. Do not rename anything for taste.
5. Everything must pass `gofmt -l .` (empty output), `go vet ./...`, and `go test ./...`.

## Repository layout to produce

```
go.mod
README.md
NOTES.md
fp/
  func.go          Func, NewFunc, Must, Compose
  func_test.go
  list.go          List and its Map method
  list_test.go
  maybe.go         Maybe, Map, Do
  maybe_test.go
  many.go          Many, NewMany, Map, Do, Each, toSlice helper
  many_test.go
  example_test.go  runnable Example functions matching the slide outputs
sum/
  sum.go           SumI, SumR, SumTR, SumTRG
  sum_test.go      correctness tests plus the four benchmarks from the deck
examples/
  weather/main.go  the Person -> Address -> City -> Weather use case
  library/main.go  the Library -> Book -> Page -> Line word-count use case
```

## The material from the slides

### Recursion and tail recursion (package `sum`)

The deck shows four implementations. Two of them have typos on the slides (`SumI` returns `v` instead of `s`; `SumTR` recurses into `Sum` rather than `SumTR`). Fix those silently in the code, and note the discrepancy in `NOTES.md`.

```go
func SumI(vs []int) int {          // iterative
	s := 0
	for _, v := range vs {
		s += v
	}
	return s
}

func SumR(vs []int) int {          // recursive
	if len(vs) == 0 {
		return 0
	}
	return vs[0] + SumR(vs[1:])
}

func SumTR(vs []int, s int) int {  // tail recursive
	if len(vs) == 0 {
		return s
	}
	return SumTR(vs[1:], s+vs[0])
}

func SumTRG(vs []int, s int) int { // tail recursion with faked TCO
	begin:
	if len(vs) == 0 {
		return s
	}
	vs, s = vs[1:], s+vs[0]
	goto begin
}
```

Write `BenchmarkSumI`, `BenchmarkSumR`, `BenchmarkSumTR`, `BenchmarkSumTRG`. The talk reported roughly 462, 4707, 5056 and 1587 ns/op on a 4-core machine; reproduce the benchmark shape (same input size, whatever that turns out to be) and record the numbers you actually measure in `README.md` alongside the historical ones.

### Representing functions (`fp/func.go`)

```go
type Func struct {
	in  reflect.Type
	out reflect.Type
	f   func(interface{}) interface{}
}

func (f Func) Call(v interface{}) interface{} { return f.f(v) }
```

`NewFunc(f interface{}) (*Func, error)` uses reflection. The slide elides the validation with a comment, so implement it properly: return a descriptive error if `f` is not a function, or does not take exactly one argument, or does not return exactly one value. The happy path from the slide is:

```go
return &Func{
	in:  tf.In(0),
	out: tf.Out(0),
	f: func(x interface{}) interface{} {
		out := vf.Call([]reflect.Value{reflect.ValueOf(x)})
		return out[0].Interface()
	},
}, nil
```

`Must(f *Func, err error) *Func` panics on error and returns `f`.

`Compose(f, g *Func) (*Func, error)` from the appendix slides:

```go
if g.out != f.in {
	return nil, fmt.Errorf("can't compose: %v != %v", g.out, f.in)
}
return &Func{
	g.in, f.out,
	func(x interface{}) interface{} { return f.Call(g.Call(x)) },
}, nil
```

### List (`fp/list.go`)

```go
type List struct {
	Head interface{}
	Tail *List
}

func (l *List) Map(f *Func) *List {
	if l == nil {
		return nil
	}
	return &List{f.Call(l.Head), l.Tail.Map(f)}
}
```

Slide usage to preserve as an Example:

```go
toUpper := Must(NewFunc(strings.ToUpper))
m := &List{"hello", &List{"bye", nil}}
res := m.Map(toUpper)   // "HELLO", "BYE"
```

Give `*List` a `String() string` so printing produces that output.

### Maybe (`fp/maybe.go`)

Basic version from the body of the talk:

```go
type Maybe struct {
	Value interface{}
}

func (m Maybe) Map(f *Func) Maybe {
	if m.Value == nil {
		return Maybe{}
	}
	return Maybe{f.Call(m.Value)}
}
```

Use the appendix version as the real implementation, because the nil-pointer handling is what makes the weather example work:

```go
func (m Maybe) Map(f *Func) Maybe {
	if m.Value == nil {
		return Maybe{}
	}
	r := f.Call(m.Value)
	vr := reflect.ValueOf(r)
	if vr.Kind() == reflect.Ptr && vr.IsNil() {
		return Maybe{}
	}
	return Maybe{r}
}
```

Chaining helper:

```go
func (m Maybe) Do(fs ...interface{}) (Maybe, error) {
	if len(fs) == 0 {
		return m, nil
	}
	f, err := NewFunc(fs[0])
	if err != nil {
		return Maybe{}, err
	}
	return m.Map(f).Do(fs[1:]...)
}
```

### Many (`fp/many.go`)

```go
type Many struct {
	Head interface{}
	Tail *Many
}

func (m *Many) Map(f *Func) *Many {
	if m == nil {
		return nil
	}
	res := m.Tail.Map(f)
	for _, v := range toSlice(f.Call(m.Head)) {
		res = &Many{v, res}
	}
	return res
}
```

The slide has `r = &Many{v, res}` where it clearly means `res = &Many{v, res}`; use the corrected form and note it. `toSlice(v interface{}) []interface{}` uses reflection: if the value is a slice or array, return its elements; otherwise return a one-element slice containing the value. That single behaviour is what lets `strings.Fields` (returning `[]string`) chain after `strings.ToUpper` (returning `string`).

Also implement:

- `NewMany(v interface{}) *Many`, which builds a `*Many` from a value or a slice.
- `Do(fs ...interface{}) (*Many, error)`, the same shape as `Maybe.Do`.
- `Each(f interface{})`, which applies a one-argument function to every element for its side effects, as used in the final slide.

Slide usage:

```go
toUpper := Must(NewFunc(strings.ToUpper))
fields  := Must(NewFunc(strings.Fields))
m := NewMany([]string{"hello there", "good bye"})
res := m.Map(toUpper).Map(fields)  // "HELLO", "THERE", "GOOD", "BYE"
```

### The Maybe use case (`examples/weather`)

```go
type Person struct{ address *Address }
type Address struct{ city *City }
type City struct{ weather *Weather }
type Weather struct{ desc string }

func (p Person) Address() *Address  { return p.address }
func (a Address) City() *City       { return a.city }
func (c City) Weather() *Weather    { return c.weather }
func (w Weather) Description() string { return w.desc }
```

Show three implementations of `Person.Weather() string` side by side, all returning `"no weather"` when the chain breaks:

1. The imperative version with three nil checks.
2. The `Maybe{p}.Map(Must(NewFunc(Person.Address)))...` chain, which relies on method expressions.
3. The `Maybe{p}.Do(Person.Address, Address.City, City.Weather, Weather.Description)` version.

`main` should run all three against a fully populated person and a person with a nil address, and print the results so they can be compared. Add a test asserting all three agree on both inputs.

### The Many use case (`examples/library`)

```go
type Library struct{ books []Book }
type Book struct{ pages []Page }
type Page struct{ lines []Line }
type Line struct{ text string }

func (l Library) Books() []Book { return l.books }
func (b Book) Pages() []Page    { return b.pages }
func (p Page) Lines() []Line    { return p.lines }
func (l Line) Text() string     { return l.text }
```

Show the nested four-loop word count, then the equivalent:

```go
w, err := NewMany(lib).Do(
	Library.Books, Book.Pages, Page.Lines, Line.Text, strings.Fields)
if err != nil {
	// type error
}
w.Each(func(s string) { count[s]++ })
```

Note that the slide writes `strings.Field`, which does not exist; it is `strings.Fields`. Include a small hardcoded library so `go run ./examples/library` prints a deterministic word count, and a test that the two approaches produce identical maps.

## README.md

Cover: what the talk was, links to the deck (https://speakerdeck.com/campoy/functional-go) and the video (https://www.youtube.com/watch?v=ouyHp2nJl0I, unverified), a note that this repo is a reconstruction from the slides rather than the original 2015 source, package overview, how to run each example, and the benchmark table.

## NOTES.md

List every place where you deviated from the literal slide text and why: the `SumI` return value, `SumTR` recursing into `Sum`, `Many.Map` assigning to `r`, `strings.Field`, anything the slides left as `// check type of f and return an error if needed`, and any ambiguity you had to resolve by choosing (for example the exact signature of `Each` or `NewMany`).

## Working method

Build it in this order, running `go test ./...` after each step: `sum`, then `fp/func.go`, then `List`, `Maybe`, `Many`, then the two examples, then the docs. Keep commits small and message them by slide topic. When a slide is ambiguous, pick the reading that makes the final usage snippets compile unchanged, since those snippets are the ground truth.