package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/campoy/functional-go/fp"
)

var libraries = []struct {
	name string
	l    Library
}{
	{"the hardcoded shelf", shelf},
	{"empty library", Library{}},
	{"a book with no pages", Library{[]Book{{}}}},
	{"a page with no lines", Library{[]Book{{[]Page{{}}}}}},
	{"a line with no text", Library{[]Book{{[]Page{{[]Line{{""}}}}}}}},
	{"one word", Library{[]Book{{[]Page{{[]Line{{"hello"}}}}}}}},
	{"repeated words", Library{[]Book{
		{[]Page{{[]Line{{"a b a"}}}}},
		{[]Page{{[]Line{{"b a"}}}}},
	}}},
}

// TestWordCountAgree is the point of the example: the flat chain of slides 73
// and 74 must count exactly what the four nested loops of slide 71 count.
//
// The empty cases matter as much as the full one. A page with no lines, or a
// line with no words, makes strings.Fields return an empty slice, and Many.Map
// has to drop the element rather than keep an empty one.
func TestWordCountAgree(t *testing.T) {
	for _, lib := range libraries {
		want := WordCountImperative(lib.l)
		got, err := WordCountFunctional(lib.l)
		if err != nil {
			t.Errorf("%s: WordCountFunctional failed: %v", lib.name, err)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: functional = %v, imperative = %v", lib.name, got, want)
		}
	}
}

// TestShelfWordCount pins a few counts from the hardcoded library, so a change
// that breaks both implementations the same way still shows up.
func TestShelfWordCount(t *testing.T) {
	words := WordCountImperative(shelf)

	for word, want := range map[string]int{
		"the": 5, "fox": 2, "dog": 2, "be": 2, "to": 2, "quick": 1, "amused": 1,
	} {
		if got := words[word]; got != want {
			t.Errorf("count[%q] = %d, want %d", word, got, want)
		}
	}
	if got, ok := words[""]; ok {
		t.Errorf(`count[""] = %d, want no entry: an empty line contributes no words`, got)
	}
}

// TestMethodExpressions pins the shapes slides 70 and 72 depend on. Value
// receivers returning slices are what make these usable as method
// expressions, and the slices are what Many.Map flattens.
func TestMethodExpressions(t *testing.T) {
	var (
		_ func(Library) []Book = Library.Books
		_ func(Book) []Page    = Book.Pages
		_ func(Page) []Line    = Page.Lines
		_ func(Line) string    = Line.Text
	)
}

// TestChainTypeError is slide 74's "// type error" branch. Do checks every
// joint before applying anything, so a step that does not fit comes back as an
// error rather than a panic inside reflect.Value.Call. See NOTES.md.
func TestChainTypeError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Do panicked with %v, want an error", r)
		}
	}()

	// Page.Lines is left out, so Book.Pages feeds a Page to Line.Text.
	w, err := fp.NewMany(shelf).Do(
		Library.Books,
		Book.Pages,
		Line.Text,
		strings.Fields)
	if err == nil {
		t.Fatalf("Do(...) = %v, want an error: a Page is not a Line", w)
	}
}
