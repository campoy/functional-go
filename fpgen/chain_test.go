package fpgen_test

import (
	"strconv"
	"testing"

	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
)

func TestChain2(t *testing.T) {
	got := fpgen.Chain2(5, strconv.Itoa, func(s string) int { return len(s) })
	assert.Equal(t, 1, got)
}

func TestChain3(t *testing.T) {
	got := fpgen.Chain3(12345,
		strconv.Itoa,
		func(s string) int { return len(s) },
		func(n int) bool { return n%2 == 1 },
	)
	assert.True(t, got)
}

func TestChain4(t *testing.T) {
	got := fpgen.Chain4(12345,
		strconv.Itoa,
		func(s string) int { return len(s) },
		func(n int) bool { return n%2 == 1 },
		strconv.FormatBool,
	)
	assert.Equal(t, "true", got)
}
