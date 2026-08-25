package fp

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFunc(t *testing.T) {
	f, err := NewFunc(strings.ToUpper)
	require.NoError(t, err, "NewFunc(strings.ToUpper)")

	assert.Equal(t, reflect.TypeOf(""), f.in, "in")
	assert.Equal(t, reflect.TypeOf(""), f.out, "out")
	assert.Equal(t, "HELLO", f.Call("hello"), `Call("hello")`)
}

// TestNewFuncErrors covers the validation slide 31 elides behind the comment
// "check type of f and return an error if needed".
func TestNewFuncErrors(t *testing.T) {
	tests := []struct {
		name string
		f    interface{}
	}{
		{"nil", nil},
		// A typed nil function has a non-nil type word, so it passes every
		// other check and only panics later, inside Call.
		{"typed nil function", (func(string) string)(nil)},
		{"not a function", 42},
		{"string", "hello"},
		{"no arguments", func() int { return 0 }},
		{"two arguments", func(a, b int) int { return a + b }},
		{"no results", func(int) {}},
		{"two results", func(int) (int, error) { return 0, nil }},
		// Rejected deliberately: reflect reports the input type as []string
		// while Call would take a single string, so Compose and Do would
		// type-check against a type the function never sees. See NOTES.md.
		{"variadic", func(vs ...string) string { return strings.Join(vs, "") }},
	}

	for _, tt := range tests {
		f, err := NewFunc(tt.f)
		assert.Errorf(t, err, "NewFunc(%s) = %v, want an error", tt.name, f)
	}
}

// TestNewFuncDereferences covers the automatic dereference in argValue, which
// the deck needs but never mentions: slide 61 feeds a *Address, returned by
// the method expression Person.Address, to Address.City, which takes a value.
func TestNewFuncDereferences(t *testing.T) {
	type point struct{ x int }

	f := Must(NewFunc(func(p point) int { return p.x }))
	assert.Equal(t, 7, f.Call(&point{7}), "Call(&point{7})")
	assert.Equal(t, 7, f.Call(point{7}), "Call(point{7})")
}

func TestMust(t *testing.T) {
	f := Must(NewFunc(strings.ToUpper))
	assert.Equal(t, "BYE", f.Call("bye"), `Call("bye")`)
}

func TestMustPanics(t *testing.T) {
	assert.Panics(t, func() { Must(NewFunc(42)) }, "Must(NewFunc(42))")
}

// TestCompose checks the argument order, which is mathematical rather than
// pipeline order: Compose(f, g) applies g first.
func TestCompose(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	twice := Must(NewFunc(func(s string) string { return s + s }))

	h, err := Compose(twice, toUpper)
	require.NoError(t, err, "Compose(twice, toUpper)")

	assert.Equal(t, "ABAB", h.Call("ab"), `Call("ab")`)
	assert.Equal(t, reflect.TypeOf(""), h.in, "in")
	assert.Equal(t, reflect.TypeOf(""), h.out, "out")
}

// TestComposeMismatch is the talk's punchline: the check g.out != f.in is the
// static type system, reimplemented by hand at runtime.
func TestComposeMismatch(t *testing.T) {
	length := Must(NewFunc(func(s string) int { return len(s) }))
	toUpper := Must(NewFunc(strings.ToUpper))

	h, err := Compose(toUpper, length)
	require.Errorf(t, err, "Compose(toUpper, length) = %v, want an error", h)
	assert.EqualError(t, err, "can't compose: int != string")
}
