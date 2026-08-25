package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/campoy/functional-go/fp"
)

// The shapes slides 70 and 72 depend on. Value receivers returning slices are
// what make these usable as method expressions, and the slices are what
// Many.Map flattens. These are declarations rather than a test because there
// is nothing to run: the compiler is the whole assertion.
var (
	_ func(Library) []Book = Library.Books
	_ func(Book) []Page    = Book.Pages
	_ func(Page) []Line    = Page.Lines
	_ func(Line) string    = Line.Text
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
		if !assert.NoErrorf(t, err, "%s: WordCountFunctional failed", lib.name) {
			continue
		}
		assert.Equalf(t, want, got, "%s: functional and imperative disagree", lib.name)
	}
}

// TestShelfWordCount pins a few counts from the hardcoded library, so a change
// that breaks both implementations the same way still shows up.
func TestShelfWordCount(t *testing.T) {
	words := WordCountImperative(shelf)

	for word, want := range map[string]int{
		"the": 5, "fox": 2, "dog": 2, "be": 2, "to": 2, "quick": 1, "amused": 1,
	} {
		assert.Equalf(t, want, words[word], "count[%q]", word)
	}
	assert.NotContains(t, words, "", "an empty line contributes no words")
}

// TestChainTypeError is slide 74's "// type error" branch. Do checks every
// joint before applying anything, so a step that does not fit comes back as an
// error rather than a panic inside reflect.Value.Call. See NOTES.md.
//
// The error is pinned exactly: the test is worth nothing if it passes on any
// error at all, since then a chain broken somewhere else entirely still looks
// like a success.
func TestChainTypeError(t *testing.T) {
	var (
		w   *fp.Many
		err error
	)
	require.NotPanics(t, func() {
		// Page.Lines is left out, so Book.Pages feeds a Page to Line.Text.
		w, err = fp.NewMany(shelf).Do(
			Library.Books,
			Book.Pages,
			Line.Text,
			strings.Fields)
	}, "Do must report a type error rather than panicking")

	require.Errorf(t, err, "Do(...) = %v, want an error: a Page is not a Line", w)
	require.EqualError(t, err,
		"can't chain: step 1 returns []main.Page, but step 2 takes main.Line")
}
