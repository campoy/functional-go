package fpgen_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	got := fpgen.Map([]string{"hello", "bye"}, strings.ToUpper)
	require.Len(t, got, 2)
	assert.Equal(t, []string{"HELLO", "BYE"}, got)
}

func TestMapChangesType(t *testing.T) {
	got := fpgen.Map([]string{"1", "22", "333"}, func(s string) int { return len(s) })
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestFilter(t *testing.T) {
	got := fpgen.Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
	assert.Equal(t, []int{2, 4}, got)
}

func TestReduce(t *testing.T) {
	sum := fpgen.Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
	assert.Equal(t, 10, sum)

	joined := fpgen.Reduce([]string{"a", "b", "c"}, "", func(acc, s string) string { return acc + s })
	assert.Equal(t, "abc", joined)
}

func TestCompose(t *testing.T) {
	f := fpgen.Compose(strconv.Itoa, func(s string) int { return len(s) })
	assert.Equal(t, "5", f("hello"))
}
