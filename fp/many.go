package fp

import (
	"fmt"
	"reflect"
	"strings"
)

// Many is a cons cell like List, but its Map flattens (slide 65).
//
// That single difference is the whole point of the type: a step returning
// []string can follow a step returning string with no per-type plumbing, which
// is what makes the library word count on slides 72 through 74 a flat chain
// rather than the four nested loops of slide 71.
type Many struct {
	Head interface{}
	Tail *Many
}

// NewMany builds a Many from v, flattening a slice or array into one cell per
// element and wrapping anything else into a single cell.
//
// The deck never declares it. Slide 72 calls NewMany(l) on a Library value and
// slides 73 and 74 call NewMany(m); slide 68's usage needs a list of strings.
// Taking interface{} and deferring to toSlice satisfies both. See NOTES.md.
func NewMany(v interface{}) *Many {
	var m *Many
	vs := toSlice(v)
	for i := len(vs) - 1; i >= 0; i-- {
		m = &Many{vs[i], m}
	}
	return m
}

// Map applies f to every element of m and flattens the results into a single
// Many (slide 66).
//
// The receiver may be nil, which is how the recursion terminates.
//
// Two departures from slide 66, both recorded in NOTES.md. The slide's loop
// body assigns to an undefined r rather than to res, discarding every result.
// And it walks each expansion forwards while prepending, which reverses the
// group: slide 68 chains strings.Fields after strings.ToUpper and prints
//
//	"HELLO", "THERE", "GOOD", "BYE"
//
// which the printed loop cannot produce. Walking the expansion backwards does.
// The printed output is ground truth; the printed loop is not.
func (m *Many) Map(f *Func) *Many {
	if m == nil {
		return nil
	}
	res := m.Tail.Map(f)
	// Prepend every element of the result, back to front, so the group ends
	// up in front of res in its original order.
	vs := toSlice(f.Call(m.Head))
	for i := len(vs) - 1; i >= 0; i-- {
		res = &Many{vs[i], res}
	}
	return res
}

// Do maps each of fs over m in turn, and returns an error if the chain does
// not type-check (slides 73 and 74).
//
// The deck never declares it either. Slide 74 does
//
//	w, err := NewMany(m).Do(...)
//
// and then w.Each(...), which pins the first result to *Many; the shape is
// otherwise Maybe.Do's, which slide 62 does show.
//
// The type check allows for the flattening Map performs, so Library.Books
// (returning []Book) chains into Book.Pages (taking a Book). As with
// Maybe.Do, checking the joints at all is an addition beyond the deck. See
// NOTES.md.
func (m *Many) Do(fs ...interface{}) (*Many, error) {
	chain, err := newChain(fs, canMap)
	if err != nil {
		return nil, err
	}
	for _, f := range chain {
		m = m.Map(f)
	}
	return m, nil
}

// Each applies f to every element of m for its side effects (slide 74).
//
// Slide 74 passes func(s string) { count[s]++ }, which returns nothing, so
// Each cannot go through NewFunc -- that requires exactly one result. It does
// its own reflection instead.
//
// It returns no error because slide 74 ignores one; a bad argument panics.
// See NOTES.md.
func (m *Many) Each(f interface{}) {
	if f == nil {
		panic("Each: f is nil")
	}
	vf := reflect.ValueOf(f)
	tf := vf.Type()
	if tf.Kind() != reflect.Func {
		panic(fmt.Sprintf("Each: %v is not a function", tf))
	}
	if tf.IsVariadic() || tf.NumIn() != 1 {
		panic(fmt.Sprintf("Each: %v must take exactly one argument", tf))
	}
	if tf.NumOut() != 0 {
		panic(fmt.Sprintf("Each: %v must return nothing", tf))
	}

	in := tf.In(0)
	for ; m != nil; m = m.Tail {
		vf.Call([]reflect.Value{argValue(in, m.Head)})
	}
}

// String renders m as its elements, comma-separated, strings quoted, the same
// way List.String does.
//
// Like List.String it is never declared in the deck, but slides 67 and 68 call
// fmt.Println on the result of Map and show
//
//	"HELLO", "THERE", "GOOD", "BYE"
//
// which requires a fmt.Stringer. See NOTES.md.
func (m *Many) String() string {
	var elems []string
	for ; m != nil; m = m.Tail {
		elems = append(elems, quote(m.Head))
	}
	return strings.Join(elems, ", ")
}

// toSlice returns the elements of v if v is a slice or an array, and a
// one-element slice holding v otherwise (slide 66).
//
// The deck calls it and never shows it. This contract is what makes Many
// flatten: strings.Fields returns []string and yields four elements, while
// strings.ToUpper returns string and yields one, and Map needs no idea which
// it is dealing with. See NOTES.md.
func toSlice(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		vs := make([]interface{}, rv.Len())
		for i := range vs {
			vs[i] = rv.Index(i).Interface()
		}
		return vs
	}
	return []interface{}{v}
}

// canMap reports whether a value of type out can be fed to a Func whose input
// type is in, once Many.Map has flattened it with toSlice.
func canMap(out, in reflect.Type) bool {
	if canCall(out, in) {
		return true
	}
	switch out.Kind() {
	case reflect.Slice, reflect.Array:
		return canCall(out.Elem(), in)
	}
	return false
}
