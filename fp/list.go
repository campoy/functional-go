package fp

import (
	"fmt"
	"strings"
)

// List is a cons cell: a value and the rest of the list, with nil for the
// empty list (slide 35).
//
// It is the first of the three containers, and the simplest -- Map just
// recurses structurally.
type List struct {
	Head interface{}
	Tail *List
}

// Map returns the list of f applied to every element of l (slide 37).
//
// Slide 35 first shows this as a package-level func Map(f *Func, l *List)
// *List; slide 36 asks "Should this be a method? Of what?" and slide 37
// converts it. Only the method form is part of the API.
//
// The conversion on slide 37 is incomplete: the body still recurses into
// Map(f, l.Tail), the package-level form, which no longer exists. It has to be
// l.Tail.Map(f). See NOTES.md.
//
// The nil check is what makes recursing into l.Tail safe on the last cell.
func (l *List) Map(f *Func) *List {
	if l == nil {
		return nil
	}
	return &List{f.Call(l.Head), l.Tail.Map(f)}
}

// String renders l as its elements, comma-separated, strings quoted.
//
// The deck never declares this, but slide 38 does fmt.Println(res) on a *List
// and shows the output
//
//	"HELLO", "BYE"
//
// which only happens if *List is a fmt.Stringer. The format is chosen to
// reproduce that line exactly. See NOTES.md.
func (l *List) String() string {
	var elems []string
	for ; l != nil; l = l.Tail {
		elems = append(elems, quote(l.Head))
	}
	return strings.Join(elems, ", ")
}

// quote renders a single element the way slides 38 and 68 print them: strings
// in double quotes, everything else however fmt would show it.
func quote(v interface{}) string {
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}
