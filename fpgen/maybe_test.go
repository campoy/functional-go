package fpgen_test

import (
	"testing"

	"github.com/campoy/functional-go/fp"
	"github.com/campoy/functional-go/fpgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaybeSomeMap(t *testing.T) {
	m := fpgen.MaybeMap(fpgen.Some(3), func(n int) int { return n * n })
	v, ok := m.Get()
	require.True(t, ok)
	assert.Equal(t, 9, v)
}

func TestMaybeNoneMap(t *testing.T) {
	m := fpgen.MaybeMap(fpgen.None[int](), func(n int) int { return n * n })
	_, ok := m.Get()
	assert.False(t, ok)
}

// box has a nil-safe pointer-receiver method: it returns a real answer even
// when the receiver itself is nil. fp.Maybe.Map cannot reach that answer --
// this is functional-go-maybe-nil, the live regression on fp's main branch
// that lesson 7 in docs/teaching-generics.md is built around.
type box struct{ n int }

func (b *box) Get() int {
	if b == nil {
		return -1
	}
	return b.n
}

func TestMaybeNilRegressionContrast(t *testing.T) {
	var b *box // typed nil, but Get is nil-safe and has a real answer: -1

	// fp.Maybe: short-circuits on ANY typed nil pointer, unconditionally.
	// The nil-safe method is never even called.
	fm, err := fp.Maybe{Value: b}.Do((*box).Get)
	require.NoError(t, err)
	assert.Nil(t, fm.Value, "fp.Maybe discards the nil-safe method's real answer -- see functional-go-maybe-nil")

	// fpgen.Maybe[*box]: ok is the only thing that decides "missing". A
	// present *box that happens to be nil is still present, so its nil-safe
	// method runs and its answer comes through.
	gm := fpgen.Some[*box](b)
	got := fpgen.MaybeMap(gm, (*box).Get)
	v, ok := got.Get()
	require.True(t, ok, "a present nil pointer is still present in fpgen.Maybe")
	assert.Equal(t, -1, v)
}
