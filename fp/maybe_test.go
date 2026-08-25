package fp

import (
	"strings"
	"testing"
)

func TestMaybeMap(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))

	// Slide 53: with a value.
	if got := (Maybe{"hello"}).Map(toUpper).Value; got != "HELLO" {
		t.Errorf(`Maybe{"hello"}.Map(toUpper).Value = %v, want "HELLO"`, got)
	}
	// Slide 54: without one.
	if got := (Maybe{}).Map(toUpper).Value; got != nil {
		t.Errorf("Maybe{}.Map(toUpper).Value = %v, want nil", got)
	}
}

// TestMaybeMapChained is slide 55.
func TestMaybeMapChained(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	twice := Must(NewFunc(func(s string) string { return s + s }))

	if got := (Maybe{"hello"}).Map(toUpper).Map(twice).Value; got != "HELLOHELLO" {
		t.Errorf("chained = %v, want HELLOHELLO", got)
	}
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
	if got := (Maybe{"hello"}).Map(broken).Value; got != nil {
		t.Fatalf("Map(broken).Value = %v, want nil: a typed nil must short-circuit", got)
	}

	// And the chain must stay broken rather than panicking on the next step.
	next := Must(NewFunc(func(b box) int { return b.n }))
	if got := (Maybe{"hello"}).Map(broken).Map(next).Value; got != nil {
		t.Errorf("Map(broken).Map(next).Value = %v, want nil", got)
	}
}

func TestMaybeDo(t *testing.T) {
	m, err := (Maybe{"hello"}).Do(strings.ToUpper, func(s string) string { return s + s })
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if m.Value != "HELLOHELLO" {
		t.Errorf("Do(...).Value = %v, want HELLOHELLO", m.Value)
	}
}

func TestMaybeDoEmpty(t *testing.T) {
	m, err := (Maybe{"hello"}).Do()
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}
	if m.Value != "hello" {
		t.Errorf("Do().Value = %v, want hello", m.Value)
	}
}

// TestMaybeDoBadFunc covers the errors slide 62 does report, the ones coming
// out of NewFunc.
func TestMaybeDoBadFunc(t *testing.T) {
	if m, err := (Maybe{"hello"}).Do(strings.ToUpper, 42); err == nil {
		t.Errorf("Do(toUpper, 42) = %v, want an error", m)
	}
}

// TestMaybeDoTypeMismatch covers the errors slide 62 does not report. Its Do
// builds the chain one step at a time, so a step whose result does not fit the
// next step's argument panics inside reflect.Value.Call. Checking the joints
// up front turns that into an error; see NOTES.md.
func TestMaybeDoTypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Do panicked with %v, want an error", r)
		}
	}()

	// strings.ToUpper returns a string; the next step wants an int.
	m, err := (Maybe{"hello"}).Do(strings.ToUpper, func(n int) int { return n + 1 })
	if err == nil {
		t.Fatalf("Do(toUpper, incr) = %v, want an error: string is not an int", m)
	}
	if !strings.Contains(err.Error(), "string") || !strings.Contains(err.Error(), "int") {
		t.Errorf("error = %q, want it to name both types", err)
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
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if m.Value != nil {
		t.Errorf("Do(...).Value = %v, want nil", m.Value)
	}
	if called {
		t.Error("the step after the broken one was applied, want it skipped")
	}
}
