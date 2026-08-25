// Command library is the Many use case from "Functional Go?", slides 69
// through 74: counting the words in a library, four levels down.
//
// It prints the word count computed two ways -- the four nested loops of slide
// 71, and the flat chain of slides 73 and 74 -- so the two can be compared.
//
// This is what Many is for. Each step returns a slice, and Map flattens it, so
// Library.Books ([]Book), Book.Pages ([]Page), Page.Lines ([]Line), Line.Text
// (string) and strings.Fields ([]string) chain without a single loop and
// without any per-type plumbing.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/campoy/functional-go/fp"
)

// The chain from slide 70: library, books, pages, lines.
type (
	Library struct{ books []Book }
	Book    struct{ pages []Page }
	Page    struct{ lines []Line }
	Line    struct{ text string }
)

// Value receivers returning slices, so each of these is a method expression:
// Library.Books has type func(Library) []Book, and so on (slide 60).
func (l Library) Books() []Book { return l.books }
func (b Book) Pages() []Page    { return b.pages }
func (p Page) Lines() []Line    { return p.lines }
func (l Line) Text() string     { return l.text }

// WordCountImperative is slide 71: four nested loops, one per level.
//
// The deck does not name it; the name is invented so the two versions can sit
// side by side. See NOTES.md.
func WordCountImperative(l Library) map[string]int {
	words := make(map[string]int)
	for _, b := range l.Books() {
		for _, p := range b.Pages() {
			for _, l := range p.Lines() {
				for _, word := range strings.Fields(l.Text()) {
					words[word]++
				}
			}
		}
	}
	return words
}

// WordCountFunctional is slides 73 and 74: the same count as one chain, with
// Many.Do flattening between every step and Many.Each doing the tallying.
//
// Slide 73 writes strings.Field, which does not exist in the standard library.
// It is strings.Fields. See NOTES.md.
//
// The error is the one slide 74 comments as "// type error": it comes back if
// a step does not fit the next, which Do checks before applying anything.
func WordCountFunctional(l Library) (map[string]int, error) {
	w, err := fp.NewMany(l).Do(
		Library.Books,
		Book.Pages,
		Page.Lines,
		Line.Text,
		strings.Fields)
	if err != nil {
		return nil, err
	}

	words := make(map[string]int)
	w.Each(func(s string) { words[s]++ })
	return words, nil
}

// shelf is a small hardcoded library, so that running this prints a
// deterministic count.
var shelf = Library{[]Book{
	{[]Page{
		{[]Line{
			{"the quick brown fox"},
			{"jumps over the lazy dog"},
		}},
		{[]Line{
			{"the dog was not amused"},
			{""},
			{"the fox was"},
		}},
	}},
	{[]Page{
		{[]Line{
			{"to be or not to be"},
			{"that is the question"},
		}},
	}},
}}

func main() {
	imperative := WordCountImperative(shelf)
	functional, err := WordCountFunctional(shelf)
	if err != nil {
		fmt.Println("chain does not type-check:", err)
		return
	}

	words := make([]string, 0, len(imperative))
	for w := range imperative {
		words = append(words, w)
	}
	sort.Strings(words)

	fmt.Printf("%-10s %11s %11s\n", "word", "slide 71", "slides 73-74")
	for _, w := range words {
		fmt.Printf("%-10s %11d %11d\n", w, imperative[w], functional[w])
	}
}
