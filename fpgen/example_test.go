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
