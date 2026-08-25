package fp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// elems flattens a List so tests can compare against a plain slice.
func elems(l *List) []interface{} {
	var vs []interface{}
	for ; l != nil; l = l.Tail {
		vs = append(vs, l.Head)
	}
	return vs
}

func TestListMap(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	l := &List{"hello", &List{"bye", nil}}

	want := []interface{}{"HELLO", "BYE"}
	assert.Equal(t, want, elems(l.Map(toUpper)), "Map(toUpper)")
}

// TestListMapDoesNotModify guards the structural recursion: Map builds a new
// list rather than writing through the old one.
func TestListMapDoesNotModify(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	l := &List{"hello", &List{"bye", nil}}

	l.Map(toUpper)
	assert.Equal(t, `"hello", "bye"`, l.String(), "original list")
}

func TestListMapNil(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	assert.Nil(t, (*List)(nil).Map(toUpper), "(*List)(nil).Map(toUpper)")
}

// TestListMapChangesType checks that Map is not restricted to endofunctions:
// the whole reason Func carries in and out separately is that they can differ.
func TestListMapChangesType(t *testing.T) {
	length := Must(NewFunc(func(s string) int { return len(s) }))
	l := &List{"hello", &List{"bye", nil}}

	assert.Equal(t, "5, 3", l.Map(length).String(), "Map(length)")
}

func TestListString(t *testing.T) {
	tests := []struct {
		name string
		l    *List
		want string
	}{
		{"empty", nil, ""},
		{"single", &List{"hello", nil}, `"hello"`},
		// The output slide 38 prints, once Map has run.
		{"slide 38", &List{"HELLO", &List{"BYE", nil}}, `"HELLO", "BYE"`},
		{"not strings", &List{1, &List{2, nil}}, "1, 2"},
	}

	for _, tt := range tests {
		assert.Equalf(t, tt.want, tt.l.String(), "%s: String()", tt.name)
	}
}
