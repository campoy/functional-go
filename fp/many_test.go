package fp

import (
	"strings"
	"testing"
)

func TestNewMany(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
	}{
		{"slice", []string{"hello there", "good bye"}, `"hello there", "good bye"`},
		{"array", [2]int{1, 2}, "1, 2"},
		{"empty slice", []string{}, ""},
		// Anything that is not a slice or array becomes one cell, which is
		// what slide 72's NewMany(l) on a Library value needs.
		{"single value", "hello", `"hello"`},
		{"struct", struct{ n int }{3}, "{3}"},
	}

	for _, tt := range tests {
		if got := NewMany(tt.v).String(); got != tt.want {
			t.Errorf("%s: NewMany(%v) = %s, want %s", tt.name, tt.v, got, tt.want)
		}
	}
}

// TestManyMap is slide 67: mapping strings.ToUpper over two strings leaves two
// elements, because toSlice wraps a plain string into one cell.
func TestManyMap(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	m := NewMany([]string{"hello there", "good bye"})

	if got, want := m.Map(toUpper).String(), `"HELLO THERE", "GOOD BYE"`; got != want {
		t.Errorf("Map(toUpper) = %s, want %s", got, want)
	}
}

// TestManyMapChained is slide 68, and the regression test for the loop
// direction in Many.Map. Slide 66 walks each expansion forwards while
// prepending, which reverses every group and yields
//
//	"THERE", "HELLO", "BYE", "GOOD"
//
// The order below is what slide 68 prints. See NOTES.md.
func TestManyMapChained(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	fields := Must(NewFunc(strings.Fields))
	m := NewMany([]string{"hello there", "good bye"})

	got := m.Map(toUpper).Map(fields).String()
	want := `"HELLO", "THERE", "GOOD", "BYE"`
	if got != want {
		t.Errorf("Map(toUpper).Map(fields) = %s, want %s", got, want)
	}
}

// TestManyMapFlattens exercises toSlice with expansions of every size,
// including the empty one, which drops an element entirely.
func TestManyMapFlattens(t *testing.T) {
	fields := Must(NewFunc(strings.Fields))
	m := NewMany([]string{"a b c", "", "d"})

	if got, want := m.Map(fields).String(), `"a", "b", "c", "d"`; got != want {
		t.Errorf("Map(fields) = %s, want %s", got, want)
	}
}

func TestManyMapNil(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	if got := (*Many)(nil).Map(toUpper); got != nil {
		t.Errorf("(*Many)(nil).Map(toUpper) = %v, want nil", got)
	}
}

func TestManyDo(t *testing.T) {
	m, err := NewMany([]string{"hello there", "good bye"}).Do(strings.ToUpper, strings.Fields)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if got, want := m.String(), `"HELLO", "THERE", "GOOD", "BYE"`; got != want {
		t.Errorf("Do(toUpper, fields) = %s, want %s", got, want)
	}
}

func TestManyDoEmpty(t *testing.T) {
	m, err := NewMany([]string{"hello"}).Do()
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}
	if got, want := m.String(), `"hello"`; got != want {
		t.Errorf("Do() = %s, want %s", got, want)
	}
}

func TestManyDoBadFunc(t *testing.T) {
	if m, err := NewMany([]string{"hello"}).Do(strings.ToUpper, 42); err == nil {
		t.Errorf("Do(toUpper, 42) = %v, want an error", m)
	}
}

// TestManyDoTypeMismatch is the Many half of the addition described in
// NOTES.md: without it, a mismatched joint panics inside reflect.Value.Call.
func TestManyDoTypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Do panicked with %v, want an error", r)
		}
	}()

	m, err := NewMany([]string{"hello"}).Do(strings.ToUpper, func(n int) int { return n + 1 })
	if err == nil {
		t.Fatalf("Do(toUpper, incr) = %v, want an error: string is not an int", m)
	}
	if !strings.Contains(err.Error(), "string") || !strings.Contains(err.Error(), "int") {
		t.Errorf("error = %q, want it to name both types", err)
	}
}

// TestManyDoAcceptsFlattening checks that the type check allows for what Map
// actually does. Library.Books returns []Book and Book.Pages takes a Book;
// that joint only type-checks because Map flattens between them.
func TestManyDoAcceptsFlattening(t *testing.T) {
	type page struct{ text string }
	type book struct{ pages []page }

	m, err := NewMany([]book{{[]page{{"a b"}, {"c"}}}}).Do(
		func(b book) []page { return b.pages },
		func(p page) string { return p.text },
		strings.Fields,
	)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if got, want := m.String(), `"a", "b", "c"`; got != want {
		t.Errorf("Do(...) = %s, want %s", got, want)
	}
}

// TestManyDoRejectsUnflattenedSlice guards the flattening contract in canMap:
// Map always passes a step's result through toSlice, so a step returning
// []string feeds strings to the next step, never a []string. That joint is a
// type error, and Do must report it rather than panic inside reflect.Call.
func TestManyDoRejectsUnflattenedSlice(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Do panicked with %v, want an error", r)
		}
	}()

	m, err := NewMany([]string{"a b"}).Do(
		strings.Fields,
		func(ss []string) int { return len(ss) })
	if err == nil {
		t.Fatalf("Do(Fields, len) = %v, want an error: Map flattens []string into strings", m)
	}
	if !strings.Contains(err.Error(), "[]string") {
		t.Errorf("error = %q, want it to name []string", err)
	}
}

// TestManyEach is slide 74's final step.
func TestManyEach(t *testing.T) {
	count := make(map[string]int)
	NewMany([]string{"a", "b", "a"}).Each(func(s string) { count[s]++ })

	if count["a"] != 2 || count["b"] != 1 {
		t.Errorf("count = %v, want map[a:2 b:1]", count)
	}
}

func TestManyEachNil(t *testing.T) {
	calls := 0
	(*Many)(nil).Each(func(s string) { calls++ })
	if calls != 0 {
		t.Errorf("Each on a nil Many made %d calls, want 0", calls)
	}
}

// TestManyEachPanics covers the validation Each does for itself. It cannot
// reuse NewFunc, because slide 74 passes a function with no result at all.
func TestManyEachPanics(t *testing.T) {
	tests := []struct {
		name string
		f    interface{}
	}{
		{"nil", nil},
		{"not a function", 42},
		{"no arguments", func() {}},
		{"two arguments", func(a, b string) {}},
		{"returns a value", func(s string) string { return s }},
	}

	for _, tt := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Each(%s) returned, want a panic", tt.name)
				}
			}()
			NewMany([]string{"a"}).Each(tt.f)
		}()
	}
}

func TestToSlice(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want []interface{}
	}{
		{"slice", []string{"a", "b"}, []interface{}{"a", "b"}},
		{"array", [2]int{1, 2}, []interface{}{1, 2}},
		{"empty slice", []string{}, []interface{}{}},
		{"nil slice", []string(nil), []interface{}{}},
		{"string", "a b", []interface{}{"a b"}},
		{"int", 3, []interface{}{3}},
	}

	for _, tt := range tests {
		got := toSlice(tt.v)
		if len(got) != len(tt.want) {
			t.Errorf("%s: toSlice(%v) = %v, want %v", tt.name, tt.v, got, tt.want)
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("%s: element %d = %v, want %v", tt.name, i, got[i], tt.want[i])
			}
		}
	}
}

func TestManyString(t *testing.T) {
	tests := []struct {
		name string
		m    *Many
		want string
	}{
		{"empty", nil, ""},
		{"single", &Many{"hello", nil}, `"hello"`},
		{"not strings", &Many{1, &Many{2, nil}}, "1, 2"},
	}

	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("%s: String() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
