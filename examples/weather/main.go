// Command weather is the Maybe use case from "Functional Go?", slides 56
// through 63: walking a chain of pointers that any link may break.
//
// It prints all three implementations of the walk side by side -- the
// imperative one from slide 58, the Map chain from slide 61, and the Do
// version from slide 63 -- against a person who has weather and one who does
// not, so the three can be compared.
//
// The Maybe use case is the reason Maybe.Map has to check for a typed nil
// pointer (slide 78). Every method here has a value receiver and returns a
// pointer, which is what makes Person.Address usable as a method expression of
// type func(Person) *Address (slide 60), and is also exactly what produces the
// typed nils.
package main

import (
	"fmt"

	"github.com/campoy/functional-go/fp"
)

// The chain from slide 57: person, address, city, weather.
type (
	Person  struct{ address *Address }
	Address struct{ city *City }
	City    struct{ weather *Weather }
	Weather struct{ desc string }
)

func (p Person) Address() *Address { return p.address }
func (a Address) City() *City      { return a.city }
func (c City) Weather() *Weather   { return c.weather }

// Description returns the weather.
//
// Slide 57 declares this as Desc, but slides 58, 61 and 63 all call
// Weather.Description. Three uses beat one declaration. See NOTES.md.
func (w Weather) Description() string { return w.desc }

// noWeather is what all three implementations return when the chain breaks.
const noWeather = "no weather"

// WeatherImperative is slide 58: the chain walked by hand, with a nil check
// after every step.
//
// The deck writes all three implementations as func (p Person) Weather()
// string. Go will not accept three methods of the same name, and comparing
// them is the point of slides 58 through 63, so they are three functions with
// invented names here. Person.Weather survives below as a wrapper. See
// NOTES.md.
func WeatherImperative(p Person) string {
	a := p.Address()
	if a == nil {
		return noWeather
	}
	c := a.City()
	if c == nil {
		return noWeather
	}
	w := c.Weather()
	if w == nil {
		return noWeather
	}
	return w.Description()
}

// WeatherMap is slide 61: the same walk as a chain of Maps over a Maybe, with
// the nil checks folded into Maybe.Map.
//
// Every step is a method expression, which is what slide 60 sets up:
// Person.Address has type func(Person) *Address, so NewFunc can wrap it like
// any other one-argument function.
func WeatherMap(p Person) string {
	w := fp.Maybe{Value: p}.
		Map(fp.Must(fp.NewFunc(Person.Address))).
		Map(fp.Must(fp.NewFunc(Address.City))).
		Map(fp.Must(fp.NewFunc(City.Weather))).
		Map(fp.Must(fp.NewFunc(Weather.Description)))
	if w.Value == nil {
		return noWeather
	}
	return w.Value.(string)
}

// WeatherDo is slide 63: the same chain again, with Maybe.Do doing the
// wrapping, so neither Must nor NewFunc appears.
//
// The error can only be a type error in the chain, which is fixed here at
// compile time in everything but name.
func WeatherDo(p Person) string {
	w, err := fp.Maybe{Value: p}.Do(
		Person.Address,
		Address.City,
		City.Weather,
		Weather.Description)
	if err != nil || w.Value == nil {
		return noWeather
	}
	return w.Value.(string)
}

// Weather returns the description of the weather where p lives, or "no
// weather" if any link in the chain is missing (slides 58, 61 and 63).
func (p Person) Weather() string { return WeatherDo(p) }

func main() {
	sunny := Person{&Address{&City{&Weather{"sunny"}}}}
	nobody := Person{}
	homeless := Person{&Address{}}

	people := []struct {
		name string
		p    Person
	}{
		{"a person in a sunny city", sunny},
		{"a person with no address", nobody},
		{"a person whose address has no city", homeless},
	}

	for _, who := range people {
		fmt.Printf("%s:\n", who.name)
		fmt.Printf("\timperative (slide 58): %s\n", WeatherImperative(who.p))
		fmt.Printf("\tMap chain  (slide 61): %s\n", WeatherMap(who.p))
		fmt.Printf("\tDo         (slide 63): %s\n", WeatherDo(who.p))
	}
}
