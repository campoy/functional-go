package fp

import (
	"strings"
	"testing"
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

	got := elems(l.Map(toUpper))
	want := []interface{}{"HELLO", "BYE"}
	if len(got) != len(want) {
		t.Fatalf("Map(toUpper) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("element %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestListMapDoesNotModify guards the structural recursion: Map builds a new
// list rather than writing through the old one.
func TestListMapDoesNotModify(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	l := &List{"hello", &List{"bye", nil}}

	l.Map(toUpper)
	if got, want := l.String(), `"hello", "bye"`; got != want {
		t.Errorf("original list = %s, want %s", got, want)
	}
}

func TestListMapNil(t *testing.T) {
	toUpper := Must(NewFunc(strings.ToUpper))
	if got := (*List)(nil).Map(toUpper); got != nil {
		t.Errorf("(*List)(nil).Map(toUpper) = %v, want nil", got)
	}
}

// TestListMapChangesType checks that Map is not restricted to endofunctions:
// the whole reason Func carries in and out separately is that they can differ.
func TestListMapChangesType(t *testing.T) {
	length := Must(NewFunc(func(s string) int { return len(s) }))
	l := &List{"hello", &List{"bye", nil}}

	if got, want := l.Map(length).String(), "5, 3"; got != want {
		t.Errorf("Map(length) = %s, want %s", got, want)
	}
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
		if got := tt.l.String(); got != tt.want {
			t.Errorf("%s: String() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
