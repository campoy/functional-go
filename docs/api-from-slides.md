# API surface, as defined in the slides

Every type, function, method, and interface that appears in `functional_go.pdf`, with the slide it appears on. This is the conformance checklist: the reconstruction is faithful when it declares all of these, spelled exactly this way.

Each entry carries a provenance mark:

- **verbatim** — appears in the deck exactly as written here
- **corrected** — appears in the deck, but with an error that must be fixed to compile (see the deviation table at the end)
- **inferred** — *used* by the deck but never declared; the signature is a reconstruction, and is the most likely place for the repo to diverge legitimately
- **illustrative** — shown to make a point about the language, not part of the library; do not implement

---

## package `sum`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `func SumI(vs []int) int` | 8, 9 | corrected |
| `func SumR(vs []int) int` | 10 | corrected |
| `func SumTR(vs []int, s int) int` | 12 | corrected |
| `func SumTRG(vs []int, s int) int` | 14 | verbatim |

Benchmarks named on slides 11, 13, 15: `BenchmarkSumI`, `BenchmarkSumR`, `BenchmarkSumTR`, `BenchmarkSumTRG`. Input size is never stated.

Note the asymmetry to preserve: `SumI` and `SumR` take one argument, `SumTR` and `SumTRG` take an accumulator as a second.

## package `fp` — `Func`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type Func struct { in reflect.Type; out reflect.Type; f func(interface{}) interface{} }` | 30 | verbatim |
| `func (f Func) Call(v interface{}) interface{}` | 30 | verbatim |
| `func NewFunc(f interface{}) (*Func, error)` | 31 | verbatim |
| `func Must(f *Func, err error) *Func` | 32 | verbatim |
| `func Compose(f, g *Func) (*Func, error)` | 77 | verbatim |

All three `Func` fields are unexported. `Call` has a value receiver; `NewFunc` and `Compose` return `*Func`.

`Compose(f, g)` returns `g` *then* `f` — it checks `g.out != f.in` and calls `f.Call(g.Call(x))`. The argument order is mathematical composition, not pipeline order. `Compose` also constructs `&Func{g.in, f.out, ...}` positionally, which requires the field order above.

`NewFunc`'s validation is elided on slide 31 as `// check type of f and return an error if needed`. It must reject: a non-function, a function without exactly one parameter, and a function without exactly one result.

## package `fp` — `List`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type List struct { Head interface{}; Tail *List }` | 35 | verbatim |
| `func (l *List) Map(f *Func) *List` | 37 | corrected |
| `func (l *List) String() string` | 38 | inferred |

Slide 35 first shows `func Map(f *Func, l *List) *List`; slide 36 asks "Should this be a method? Of what?" and slide 37 converts it. Only the method form is part of the API.

`String` is never declared, but slide 38 does `fmt.Println(res)` and shows output `"HELLO", "BYE"` — which requires a `Stringer`.

## package `fp` — `Maybe`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type Maybe struct { Value interface{} }` | 52 | verbatim |
| `func (m Maybe) Map(f *Func) Maybe` | 52, 78 | verbatim |
| `func (m Maybe) Do(fs ...interface{}) (Maybe, error)` | 62 | verbatim |

Value receivers and value returns throughout — `Maybe` is never a pointer. `Value` is the only exported field, and callers read it directly (`res.Value`, slides 53–55, 61, 63).

Two bodies for `Map` exist: the simple nil check on slide 52, and the nil-pointer-aware version on slide 78 that inspects `reflect.ValueOf(r).Kind() == reflect.Ptr && vr.IsNil()`. Slide 78 is the one to implement; slide 52's version cannot make the weather example work.

## package `fp` — `Many`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type Many struct { Head interface{}; Tail *Many }` | 65 | verbatim |
| `func (m *Many) Map(f *Func) *Many` | 66 | corrected |
| `func NewMany(v interface{}) *Many` | 72, 73, 74 | inferred |
| `func (m *Many) Do(fs ...interface{}) (*Many, error)` | 73, 74 | inferred |
| `func (m *Many) Each(f interface{})` | 74 | inferred |
| `func toSlice(v interface{}) []interface{}` | 66 | inferred |

Pointer receivers throughout, unlike `Maybe`. `Map` must tolerate a nil receiver — it recurses into `m.Tail.Map(f)` and returns `nil` when `m == nil`.

The four inferred entries are the weakest part of the reconstruction:

- `NewMany` is called as `NewMany(l)` on a `Library` value (slide 72) and `NewMany(m)` (slides 73–74). It must accept both a plain value and a slice.
- `Do` is assigned two ways: `_, err :=` (slide 73) and `w, err :=` followed by `w.Each(...)` (slide 74), which fixes the first result as `*Many`.
- `Each` takes `func(s string)` — a function with no return — so it cannot reuse `NewFunc`, which requires exactly one result.
- `toSlice` is called but never shown. Its contract: slice or array in, elements out; anything else in, a one-element slice out. This is what lets `strings.Fields` (`[]string`) follow `strings.ToUpper` (`string`).

## `examples/weather`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type Person struct{ address *Address }` | 57 | verbatim |
| `type Address struct{ city *City }` | 57 | verbatim |
| `type City struct{ weather *Weather }` | 57 | verbatim |
| `type Weather struct{ desc string }` | 57 | verbatim |
| `func (p Person) Address() *Address` | 57 | verbatim |
| `func (a Address) City() *City` | 57 | verbatim |
| `func (c City) Weather() *Weather` | 57 | verbatim |
| `func (w Weather) Description() string` | 57, 58, 61, 63 | corrected |
| `func (p Person) Weather() string` | 58, 61, 63 | verbatim |

Value receivers returning pointers — that combination is what makes `Person.Address` usable as a method expression of type `func(Person) *Address` (slide 60), and what makes `Maybe.Map`'s typed-nil check necessary.

`Person.Weather() string` has three implementations across slides 58 (imperative, three nil checks), 61 (`Map` chain), and 63 (`Do`). All return `"no weather"` when the chain breaks.

## `examples/library`

| Signature | Slide | Provenance |
| --- | --- | --- |
| `type Library struct{ books []Book }` | 70 | verbatim |
| `type Book struct{ pages []Page }` | 70 | verbatim |
| `type Page struct{ lines []Line }` | 70 | verbatim |
| `type Line struct{ text string }` | 70 | verbatim |
| `func (l Library) Books() []Book` | 70 | verbatim |
| `func (b Book) Pages() []Page` | 70 | verbatim |
| `func (p Page) Lines() []Line` | 70 | verbatim |
| `func (l Line) Text() string` | 70 | verbatim |

Value receivers returning slices, so each is a method expression `func(Library) []Book` and so on. Slide 71 gives the four-deep nested loop these replace.

## Illustrative — do not implement

| Signature | Slide | Note |
| --- | --- | --- |
| `func Map(f func(v int) bool, vs []int) []bool` | 21, 22 | the concrete-type version, shown twice with different recursion order |
| `func Map(f interface{}, vs interface{}) interface{}` | 26, 27, 29, 33 | rejected: "embrace the interface, but not too much" |
| `func Map(f *Func, vs interface{}) interface{}` | 33, 34 | intermediate step toward `List.Map` |
| `type Stringer interface { String() string }` | 41 | Go's answer to the `Show` typeclass |
| `type Equatable interface { Equals(e Equatable) bool }` | 43 | shown *not* to work: `func (i Integer) Equals(j Integer) bool` does not satisfy it |
| `type Mapper interface { Map(*Func) ??? }` | 47 | not expressible in Go — the reason `List`, `Maybe`, and `Many` share no interface |

Slide 47's `???` is the talk's central admission: Go cannot express Functor, so the three containers duplicate `Map` by hand.

---

## Deviations from the literal slide text

Every entry marked **corrected** above departs from what the slide prints. The full table — deck text, replacement, and reasoning for each — is in [`../NOTES.md`](../NOTES.md), which is the canonical deviation log.
