package fpgen_test

import (
	"fmt"
	"strings"

	"github.com/campoy/functional-go/fpgen"
)

func ExampleMap() {
	fmt.Println(fpgen.Map([]string{"hello", "bye"}, strings.ToUpper))
	// Output: [HELLO BYE]
}

func ExampleListMap() {
	l := &fpgen.List[string]{Head: "hello", Tail: &fpgen.List[string]{Head: "bye"}}
	fmt.Println(fpgen.ListMap(l, strings.ToUpper))
	// Output: "HELLO", "BYE"
}

func ExampleMaybeMap() {
	m := fpgen.MaybeMap(fpgen.Some(3), func(n int) int { return n * n })
	fmt.Println(m.Get())
	m = fpgen.MaybeMap(fpgen.None[int](), func(n int) int { return n * n })
	fmt.Println(m.Get())
	// Output:
	// 9 true
	// 0 false
}

func ExampleFlatMap() {
	// The generic answer to slide 68's chain: ToUpper keeps one cell per
	// element (ManyMap), Fields multiplies each into several (FlatMap).
	m := fpgen.NewMany([]string{"hello there", "good bye"})
	upper := fpgen.ManyMap(m, strings.ToUpper)
	words := fpgen.FlatMap(upper, strings.Fields)
	fmt.Println(words)
	// Output: "HELLO", "THERE", "GOOD", "BYE"
}

func ExampleChain3() {
	// Person -> Address -> City -> Weather, the shape slide 61 chains with
	// method expressions (see examples/weather). Chain3 needs each step's
	// type named up front, in exchange for no reflection at all.
	type address struct{ city string }
	type person struct{ addr address }

	weather := map[string]string{"Paris": "cloudy"}
	got := fpgen.Chain3(
		person{addr: address{city: "Paris"}},
		func(p person) address { return p.addr },
		func(a address) string { return a.city },
		func(city string) string { return weather[city] },
	)
	fmt.Println(got)
	// Output: cloudy
}

func ExampleChain3_pointerStep() {
	// The shape examples/weather actually has: Person.Address returns
	// *Address, but Address.City has a value receiver, so the chain hands a
	// *Address to a step wanting an Address. fp repairs that mismatch at run
	// time inside argValue (fp/func.go), with no line of fp's source showing
	// it happen. fpgen cannot: feeding the pointer step straight into the
	// value step is a compile error, so the dereference is written out as its
	// own step below, where a reader sees it.
	type address struct{ city string }
	type person struct{ addr *address }

	weather := map[string]string{"Paris": "cloudy"}
	got := fpgen.Chain3(
		person{addr: &address{city: "Paris"}},
		func(p person) *address { return p.addr },
		func(a *address) address { return *a },
		func(a address) string { return weather[a.city] },
	)
	fmt.Println(got)
	// Output: cloudy
}

// The weather chain of slides 57 through 63, mirrored here because
// examples/weather is package main and cannot be imported. Value receivers
// returning pointers, exactly as the deck declares them.
type (
	chainPerson  struct{ addr *chainAddress }
	chainAddress struct{ city *chainCity }
	chainCity    struct{ weather *chainWeather }
	chainWeather struct{ desc string }
)

func (p chainPerson) Address() *chainAddress { return p.addr }
func (a chainAddress) City() *chainCity      { return a.city }
func (c chainCity) Weather() *chainWeather   { return c.weather }
func (w chainWeather) Description() string   { return w.desc }

func ExampleChain4() {
	// The talk's own chain, run end to end: four functions, which is one more
	// than Chain3 takes. Reaching it needed a whole new function in
	// fpgen/chain.go, and a five-step chain would need another -- that is
	// lesson 9's wall, met at the deck's own length.
	//
	// Only the first step can be a bare method expression: every later one
	// receives the previous step's pointer and its method has a value
	// receiver, so the dereference is written out, three times, where a
	// reader sees it. fp.Maybe.Do does the same three conversions invisibly
	// inside argValue (fp/func.go).
	p := chainPerson{addr: &chainAddress{city: &chainCity{weather: &chainWeather{desc: "cloudy"}}}}

	got := fpgen.Chain4(p,
		chainPerson.Address,
		func(a *chainAddress) *chainCity { return a.City() },
		func(c *chainCity) *chainWeather { return c.Weather() },
		func(w *chainWeather) string { return w.Description() },
	)
	fmt.Println(got)
	// Output: cloudy
}
