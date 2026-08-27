package fpgen_test

import (
	"strings"
	"testing"

	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fromSlice[T any](vs []T) *fpgen.List[T] {
	var l *fpgen.List[T]
	for i := len(vs) - 1; i >= 0; i-- {
		l = &fpgen.List[T]{Head: vs[i], Tail: l}
	}
	return l
}

func TestListMap(t *testing.T) {
	l := fromSlice([]string{"hello", "bye"})
	got := fpgen.ListMap(l, strings.ToUpper)
	require.NotNil(t, got)
	assert.Equal(t, `"HELLO", "BYE"`, got.String())
}

func TestListMapNil(t *testing.T) {
	var l *fpgen.List[string]
	got := fpgen.ListMap(l, strings.ToUpper)
	assert.Nil(t, got)
}

func TestListMapChangesType(t *testing.T) {
	l := fromSlice([]string{"a", "bb", "ccc"})
	got := fpgen.ListMap(l, func(s string) int { return len(s) })
	assert.Equal(t, "1, 2, 3", got.String())
}

func TestListReverse(t *testing.T) {
	l := fromSlice([]int{1, 2, 3})
	assert.Equal(t, "3, 2, 1", l.Reverse().String())
}
