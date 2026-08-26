package fp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equalf(t, tt.want, NewMany(tt.v).String(), "%s: NewMany(%v)", tt.name, tt.v)
	}
}

// TestManyMap is slide 67: mapping strings.ToUpper over two strings leaves two
// elements, because toSlice wraps a plain string into one cell.
func TestManyMap(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	m := NewMany([]string{"hello there", "good bye"})

	assert.Equal(t, `"HELLO THERE", "GOOD BYE"`, m.Map(toUpper).String(), "Map(toUpper)")
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
	assert.Equal(t, `"HELLO", "THERE", "GOOD", "BYE"`, got, "Map(toUpper).Map(fields)")
}

// TestManyMapFlattens exercises toSlice with expansions of every size,
// including the empty one, which drops an element entirely.
func TestManyMapFlattens(t *testing.T) {
	fields := Must(NewFunc(strings.Fields))
	m := NewMany([]string{"a b c", "", "d"})

	assert.Equal(t, `"a", "b", "c", "d"`, m.Map(fields).String(), "Map(fields)")
}

func TestManyMapNil(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	assert.Nil(t, (*Many)(nil).Map(toUpper), "(*Many)(nil).Map(toUpper)")
}

func TestManyDo(t *testing.T) {
	m, err := NewMany([]string{"hello there", "good bye"}).Do(strings.ToUpper, strings.Fields)
	require.NoError(t, err, "Do")
	assert.Equal(t, `"HELLO", "THERE", "GOOD", "BYE"`, m.String(), "Do(toUpper, fields)")
}

func TestManyDoEmpty(t *testing.T) {
	m, err := NewMany([]string{"hello"}).Do()
	require.NoError(t, err, "Do()")
	assert.Equal(t, `"hello"`, m.String(), "Do()")
}

func TestManyDoBadFunc(t *testing.T) {
	m, err := NewMany([]string{"hello"}).Do(strings.ToUpper, 42)
	assert.Errorf(t, err, "Do(toUpper, 42) = %v, want an error", m)
}

// TestManyDoTypeMismatch is the Many half of the addition described in
// NOTES.md: without it, a mismatched joint panics inside reflect.Value.Call.
func TestManyDoTypeMismatch(t *testing.T) {
	var (
		m   *Many
		err error
	)
	require.NotPanics(t, func() {
		m, err = NewMany([]string{"hello"}).Do(strings.ToUpper, func(n int) int { return n + 1 })
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(toUpper, incr) = %v, want an error", m)
	assert.EqualError(t, err, "can't chain: step 0 returns string, but step 1 takes int")
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
	require.NoError(t, err, "Do")
	assert.Equal(t, `"a", "b", "c"`, m.String(), "Do(...)")
}

// TestManyDoRejectsUnflattenedSlice guards the flattening contract in canMap:
// Map always passes a step's result through toSlice, so a step returning
// []string feeds strings to the next step, never a []string. That joint is a
// type error, and Do must report it rather than panic inside reflect.Call.
func TestManyDoRejectsUnflattenedSlice(t *testing.T) {
	var (
		m   *Many
		err error
	)
	require.NotPanics(t, func() {
		m, err = NewMany([]string{"a b"}).Do(
			strings.Fields,
			func(ss []string) int { return len(ss) })
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(Fields, len) = %v, want an error", m)
	assert.EqualError(t, err, "can't chain: step 0 returns []string, but step 1 takes []string")
}

// TestManyDoStartMismatch is the Many half of the starting-value check: every
// cell already in the container is passed straight to step 0, so a cell that
// does not fit it panics inside reflect.Value.Call.
func TestManyDoStartMismatch(t *testing.T) {
	var (
		m   *Many
		err error
	)
	require.NotPanics(t, func() {
		m, err = NewMany([]int{1}).Do(strings.ToUpper)
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(toUpper) over ints = %v, want an error", m)
	assert.EqualError(t, err, "can't chain: the starting value is a int, but step 0 takes string")
}

// TestManyDoStartAllowed covers what that check must NOT reject: a nil
// receiver, an empty chain, and a nil cell head, which reflect.TypeOf reports
// as a nil Type.
func TestManyDoStartAllowed(t *testing.T) {
	m, err := (*Many)(nil).Do(strings.ToUpper)
	if assert.NoError(t, err, "Do on a nil Many") {
		assert.Nil(t, m, "Do on a nil Many")
	}

	m, err = NewMany([]int{1}).Do()
	if assert.NoError(t, err, "Do()") {
		assert.Equal(t, "1", m.String(), "Do()")
	}

	m, err = (&Many{nil, nil}).Do(strings.ToUpper)
	if assert.NoError(t, err, "Do over a nil head") {
		assert.Equal(t, `""`, m.String(), "Do over a nil head")
	}
}

// TestManyEach is slide 74's final step.
func TestManyEach(t *testing.T) {
	count := make(map[string]int)
	NewMany([]string{"a", "b", "a"}).Each(func(s string) { count[s]++ })

	assert.Equal(t, map[string]int{"a": 2, "b": 1}, count, "count")
}

func TestManyEachNil(t *testing.T) {
	calls := 0
	(*Many)(nil).Each(func(s string) { calls++ })
	assert.Zero(t, calls, "Each on a nil Many made calls")
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
		assert.Panicsf(t, func() { NewMany([]string{"a"}).Each(tt.f) }, "Each(%s)", tt.name)
	}
}

// TestManyMapTypedNilPanics pins the other half of the typed-nil fix. Many
// does not short-circuit -- it has no notion of an absent element -- so a step
// returning a typed nil leaves a nil *box where the next step wants a box.
// argValue used to substitute the zero box and report the answer as a success;
// now it says so.
func TestManyMapTypedNilPanics(t *testing.T) {
	type box struct{ n int }

	none := Must(NewFunc(func(b box) *box { return nil }))
	get := Must(NewFunc(func(b box) int { return b.n }))

	assert.PanicsWithValue(t,
		"can't dereference a nil *fp.box to call a function taking fp.box",
		func() { NewMany([]box{{7}}).Map(none).Map(get) },
		"Many chain over a step returning a typed nil")
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
		assert.Equalf(t, tt.want, toSlice(tt.v), "%s: toSlice(%v)", tt.name, tt.v)
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
		assert.Equalf(t, tt.want, tt.m.String(), "%s: String()", tt.name)
	}
}
