package fpgen_test

import (
	"strings"
	"testing"

	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
)

func TestManyMap(t *testing.T) {
	m := fpgen.NewMany([]string{"hello", "bye"})
	got := fpgen.ManyMap(m, strings.ToUpper)
	assert.Equal(t, `"HELLO", "BYE"`, got.String())
}

func TestFlatMap(t *testing.T) {
	// Slide 68's chain: ToUpper (string -> string, one cell) then Fields
	// (string -> []string, many cells). The generic version has to name each
	// step's shape.
	m := fpgen.NewMany([]string{"hello there", "good bye"})
	upper := fpgen.ManyMap(m, strings.ToUpper)
	words := fpgen.FlatMap(upper, strings.Fields)
	assert.Equal(t, `"HELLO", "THERE", "GOOD", "BYE"`, words.String())
}

// TestFlatMapAppliesInListOrder pins the order f is *called* in, not just
// the order of the result. A side-effecting f -- a counter, an appender,
// anything logging -- must see the head first, the way a reader of
// FlatMap's signature would assume.
func TestFlatMapAppliesInListOrder(t *testing.T) {
	m := fpgen.NewMany([]string{"a", "b", "c"})

	var seen []string
	words := fpgen.FlatMap(m, func(s string) []string {
		seen = append(seen, s)
		return []string{s, s}
	})

	assert.Equal(t, []string{"a", "b", "c"}, seen)
	assert.Equal(t, `"a", "a", "b", "b", "c", "c"`, words.String())
}

func TestEach(t *testing.T) {
	m := fpgen.NewMany([]string{"a", "bb", "a", "bb", "bb"})
	count := map[string]int{}
	m.Each(func(s string) { count[s]++ })
	assert.Equal(t, map[string]int{"a": 2, "bb": 3}, count)
}
