package fp

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFunc(t *testing.T) {
	f, err := NewFunc(strings.ToUpper)
	if err != nil {
		t.Fatalf("NewFunc(strings.ToUpper) failed: %v", err)
	}
	if want := reflect.TypeOf(""); f.in != want {
		t.Errorf("in = %v, want %v", f.in, want)
	}
	if want := reflect.TypeOf(""); f.out != want {
		t.Errorf("out = %v, want %v", f.out, want)
	}
	if got := f.Call("hello"); got != "HELLO" {
		t.Errorf(`Call("hello") = %v, want "HELLO"`, got)
	}
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
		if err == nil {
			t.Errorf("NewFunc(%s) = %v, want an error", tt.name, f)
		}
	}
}

// TestNewFuncDereferences covers the automatic dereference in argValue, which
// the deck needs but never mentions: slide 61 feeds a *Address, returned by
// the method expression Person.Address, to Address.City, which takes a value.
func TestNewFuncDereferences(t *testing.T) {
	type point struct{ x int }

	f := Must(NewFunc(func(p point) int { return p.x }))
	if got := f.Call(&point{7}); got != 7 {
		t.Errorf("Call(&point{7}) = %v, want 7", got)
	}
	if got := f.Call(point{7}); got != 7 {
		t.Errorf("Call(point{7}) = %v, want 7", got)
	}
}

func TestMust(t *testing.T) {
	f := Must(NewFunc(strings.ToUpper))
	if got := f.Call("bye"); got != "BYE" {
		t.Errorf(`Call("bye") = %v, want "BYE"`, got)
	}
}

func TestMustPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Must(NewFunc(42)) returned, want a panic")
		}
	}()
	Must(NewFunc(42))
}

// TestCompose checks the argument order, which is mathematical rather than
// pipeline order: Compose(f, g) applies g first.
func TestCompose(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	twice := Must(NewFunc(func(s string) string { return s + s }))

	h, err := Compose(twice, toUpper)
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if got := h.Call("ab"); got != "ABAB" {
		t.Errorf(`Call("ab") = %v, want "ABAB"`, got)
	}
	if want := reflect.TypeOf(""); h.in != want || h.out != want {
		t.Errorf("in, out = %v, %v, want %v, %v", h.in, h.out, want, want)
	}
}

// TestComposeMismatch is the talk's punchline: the check g.out != f.in is the
// static type system, reimplemented by hand at runtime.
func TestComposeMismatch(t *testing.T) {
	length := Must(NewFunc(func(s string) int { return len(s) }))
	toUpper := Must(NewFunc(strings.ToUpper))

	if h, err := Compose(toUpper, length); err == nil {
		t.Errorf("Compose(toUpper, length) = %v, want an error: int is not a string", h)
	}
}
