package fp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaybeMap(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))

	// Slide 53: with a value.
	assert.Equal(t, "HELLO", (Maybe{"hello"}).Map(toUpper).Value, `Maybe{"hello"}.Map(toUpper)`)
	// Slide 54: without one.
	assert.Nil(t, (Maybe{}).Map(toUpper).Value, "Maybe{}.Map(toUpper)")
}

// TestMaybeMapChained is slide 55.
func TestMaybeMapChained(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	twice := Must(NewFunc(func(s string) string { return s + s }))

	got := (Maybe{"hello"}).Map(toUpper).Map(twice).Value
	assert.Equal(t, "HELLOHELLO", got, "chained")
}

// TestMaybeMapTypedNil is the reason slide 78 exists rather than slide 52. A
// Go method returns a typed nil pointer, not a nil interface, so a plain
// m.Value == nil check sees something non-nil and keeps calling down a chain
// that has already broken.
func TestMaybeMapTypedNil(t *testing.T) {
	type box struct{ n int }

	// The step that breaks the chain: it returns (*box)(nil), which is not a
	// nil interface value.
	broken := Must(NewFunc(func(s string) *box { return nil }))
	require.Nil(t, (Maybe{"hello"}).Map(broken).Value,
		"Map(broken).Value: a typed nil must short-circuit")

	// And the chain must stay broken rather than panicking on the next step.
	next := Must(NewFunc(func(b box) int { return b.n }))
	assert.Nil(t, (Maybe{"hello"}).Map(broken).Map(next).Value, "Map(broken).Map(next)")
}

func TestMaybeDo(t *testing.T) {
	m, err := (Maybe{"hello"}).Do(strings.ToUpper, func(s string) string { return s + s })
	require.NoError(t, err, "Do")
	assert.Equal(t, "HELLOHELLO", m.Value, "Do(...).Value")
}

func TestMaybeDoEmpty(t *testing.T) {
	m, err := (Maybe{"hello"}).Do()
	require.NoError(t, err, "Do()")
	assert.Equal(t, "hello", m.Value, "Do().Value")
}

// TestMaybeDoBadFunc covers the errors slide 62 does report, the ones coming
// out of NewFunc.
func TestMaybeDoBadFunc(t *testing.T) {
	m, err := (Maybe{"hello"}).Do(strings.ToUpper, 42)
	assert.Errorf(t, err, "Do(toUpper, 42) = %v, want an error", m)
}

// TestMaybeDoTypeMismatch covers the errors slide 62 does not report. Its Do
// builds the chain one step at a time, so a step whose result does not fit the
// next step's argument panics inside reflect.Value.Call. Checking the joints
// up front turns that into an error; see NOTES.md.
func TestMaybeDoTypeMismatch(t *testing.T) {
	var (
		m   Maybe
		err error
	)
	require.NotPanics(t, func() {
		// strings.ToUpper returns a string; the next step wants an int.
		m, err = (Maybe{"hello"}).Do(strings.ToUpper, func(n int) int { return n + 1 })
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(toUpper, incr) = %v, want an error", m)
	assert.EqualError(t, err, "can't chain: step 0 returns string, but step 1 takes int")
}

// TestMaybeDoStartMismatch guards the other end of the chain: the value the
// Maybe already holds has to fit step 0, or Map hands it straight to
// reflect.Value.Call and panics there.
func TestMaybeDoStartMismatch(t *testing.T) {
	var (
		m   Maybe
		err error
	)
	require.NotPanics(t, func() {
		m, err = (Maybe{42}).Do(strings.ToUpper)
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(toUpper) on a Maybe{42} = %v, want an error", m)
	assert.EqualError(t, err, "can't chain: the starting value is a int, but step 0 takes string")
}

// TestMaybeDoStartAllowed covers what must NOT be rejected by that check: a nil
// value short-circuits legitimately, and an empty chain is a no-op.
func TestMaybeDoStartAllowed(t *testing.T) {
	m, err := (Maybe{}).Do(strings.ToUpper)
	if assert.NoError(t, err, "Do on an empty Maybe") {
		assert.Nil(t, m.Value, "Do on an empty Maybe")
	}

	m, err = (Maybe{42}).Do()
	if assert.NoError(t, err, "Do()") {
		assert.Equal(t, 42, m.Value, "Do()")
	}
}

// TestMaybeDoStopsEarly checks that Do short-circuits like Map: once the chain
// is empty, no later step is applied.
func TestMaybeDoStopsEarly(t *testing.T) {
	type box struct{ n int }

	called := false
	m, err := (Maybe{"hello"}).Do(
		func(s string) *box { return nil },
		func(b box) int { called = true; return b.n },
	)
	require.NoError(t, err, "Do")
	assert.Nil(t, m.Value, "Do(...).Value")
	assert.False(t, called, "the step after the broken one was applied, want it skipped")
}
