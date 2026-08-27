//go:build ignore

// This file does not compile, on purpose -- but note WHERE it fails. It is
// lesson 9, THE SECOND WALL, in docs/teaching-generics.md.
//
// Unlike lessons 5 and 10, whose walls are compiler rejections of a
// DECLARATION, Do below declares fine: `fs ...func(T) T` is a perfectly
// legal generic signature. The wall only appears at the CALL SITE, when
// that legal signature is handed slide 61's chain, whose element type
// changes at every step.
//
// fpgen/wall_test.go runs `go build -gcflags=-lang=go1.21` on this exact
// file and asserts the diagnostic, so the claim in the doc comments cannot
// silently rot.
package main

// The weather chain's types, mirroring examples/weather/main.go: value
// receivers, pointer returns, a different type at every step.
type (
	Person  struct{ address *Address }
	Address struct{ city *City }
	City    struct{ weather *Weather }
	Weather struct{ desc string }
)

func (p Person) Address() *Address    { return p.address }
func (a Address) City() *City         { return a.city }
func (c City) Weather() *Weather      { return c.weather }
func (w Weather) Description() string { return w.desc }

// Do is the honest same-type shape. It compiles: this declaration is not
// the wall.
func Do[T any](v T, fs ...func(T) T) T {
	for _, f := range fs {
		v = f(v)
	}
	return v
}

func main() {
	var p Person
	// This is the wall. Person.Address is func(Person) *Address, so T is
	// asked to be both Person and *Address at once.
	_ = Do(p, Person.Address, Address.City, City.Weather, Weather.Description)
}
