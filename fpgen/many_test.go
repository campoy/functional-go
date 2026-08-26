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

func TestEach(t *testing.T) {
	m := fpgen.NewMany([]string{"a", "bb", "a", "bb", "bb"})
	count := map[string]int{}
	fpgen.Each(m, func(s string) { count[s]++ })
	assert.Equal(t, map[string]int{"a": 2, "bb": 3}, count)
}
