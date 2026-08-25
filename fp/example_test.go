package fp_test

import (
	"fmt"
	"sort"
	"strings"

	"github.com/campoy/functional-go/fp"
)

// Slide 38: Map over a List.
func ExampleList_Map() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	m := &fp.List{Head: "hello", Tail: &fp.List{Head: "bye"}}
	res := m.Map(toUpper)
	fmt.Println(res)
	// Output: "HELLO", "BYE"
}

// Slide 53: Map over a Maybe holding a value.
func ExampleMaybe_Map() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	m := fp.Maybe{Value: "hello"}
	res := m.Map(toUpper)
	fmt.Println(res.Value)
	// Output: HELLO
}

// Slide 54: the same chain over an empty Maybe.
func ExampleMaybe_Map_empty() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	m := fp.Maybe{}
	res := m.Map(toUpper)
	fmt.Println(res.Value)
	// Output: <nil>
}

// Slide 55: two steps chained.
func ExampleMaybe_Map_chained() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	twice := fp.Must(fp.NewFunc(func(s string) string { return s + s }))
	m := fp.Maybe{Value: "hello"}
	res := m.Map(toUpper).Map(twice)
	fmt.Println(res.Value)
	// Output: HELLOHELLO
}

// Slide 62 and 63: the same chain without naming Func, Must or NewFunc.
func ExampleMaybe_Do() {
	res, err := fp.Maybe{Value: "hello"}.Do(
		strings.ToUpper,
		func(s string) string { return s + s })
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Value)
	// Output: HELLOHELLO
}

// Do reports a type error rather than panicking, which slide 62's version
// cannot: it only surfaces errors coming out of NewFunc. See NOTES.md.
func ExampleMaybe_Do_typeError() {
	_, err := fp.Maybe{Value: "hello"}.Do(
		strings.ToUpper,
		func(n int) int { return n + 1 })
	fmt.Println(err)
	// Output: can't chain: step 0 returns string, but step 1 takes int
}

// Slide 67: Map over a Many, one result per element.
func ExampleMany_Map() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	m := fp.NewMany([]string{"hello there", "good bye"})
	res := m.Map(toUpper)
	fmt.Println(res)
	// Output: "HELLO THERE", "GOOD BYE"
}

// Slide 68, and the reason Many exists: strings.Fields returns []string, and
// Map flattens it, so two elements become four.
//
// This output is also the regression test for the loop direction in Many.Map.
// Slide 66 walks each expansion forwards while prepending, which would print
// "THERE", "HELLO", "BYE", "GOOD" instead. See NOTES.md.
func ExampleMany_Map_chained() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	fields := fp.Must(fp.NewFunc(strings.Fields))
	m := fp.NewMany([]string{"hello there", "good bye"})
	res := m.Map(toUpper).Map(fields)
	fmt.Println(res)
	// Output: "HELLO", "THERE", "GOOD", "BYE"
}

// Slides 73 and 74: the chain as a list of plain functions.
func ExampleMany_Do() {
	res, err := fp.NewMany([]string{"hello there", "good bye"}).Do(
		strings.ToUpper,
		strings.Fields)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
	// Output: "HELLO", "THERE", "GOOD", "BYE"
}

// Slide 74's last line: Each takes a function with no result, so it cannot go
// through NewFunc.
func ExampleMany_Each() {
	count := make(map[string]int)
	fp.NewMany([]string{"to be or not to be"}).
		Map(fp.Must(fp.NewFunc(strings.Fields))).
		Each(func(s string) { count[s]++ })

	words := make([]string, 0, len(count))
	for w := range count {
		words = append(words, w)
	}
	sort.Strings(words)
	for _, w := range words {
		fmt.Printf("%s: %d\n", w, count[w])
	}
	// Output:
	// be: 2
	// not: 1
	// or: 1
	// to: 2
}

// Slide 77: Compose applies g first, then f -- mathematical order, not
// pipeline order.
func ExampleCompose() {
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))
	twice := fp.Must(fp.NewFunc(func(s string) string { return s + s }))

	h, err := fp.Compose(twice, toUpper)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(h.Call("hello"))
	// Output: HELLOHELLO
}

// The check g.out != f.in is the static type system, reimplemented by hand at
// runtime because the interface{} signatures threw it away.
func ExampleCompose_typeError() {
	length := fp.Must(fp.NewFunc(func(s string) int { return len(s) }))
	toUpper := fp.Must(fp.NewFunc(strings.ToUpper))

	_, err := fp.Compose(toUpper, length)
	fmt.Println(err)
	// Output: can't compose: int != string
}
