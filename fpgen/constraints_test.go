package fpgen_test

import (
	"testing"

	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	assert.Equal(t, `"hello"`, fpgen.Quote("hello"))
	assert.Equal(t, "42", fpgen.Quote(42))
}

func TestMax(t *testing.T) {
	assert.Equal(t, 5, fpgen.Max(3, 5))
	assert.Equal(t, "bye", fpgen.Max("bye", "ahoy"))
}

func TestSum(t *testing.T) {
	assert.Equal(t, 10, fpgen.Sum([]int{1, 2, 3, 4}))
	assert.Equal(t, 6.5, fpgen.Sum([]float64{1.5, 2, 3}))
}

type person struct{ name string }

func (p person) Name() string { return p.name }

func TestDescribe(t *testing.T) {
	assert.Equal(t, "Ada", fpgen.Describe(person{name: "Ada"}))
}
