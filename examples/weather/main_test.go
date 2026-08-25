package main

import "testing"

// impls lists the three implementations of slides 58, 61 and 63, so every test
// runs against all of them.
var impls = []struct {
	name string
	f    func(Person) string
}{
	{"WeatherImperative", WeatherImperative},
	{"WeatherMap", WeatherMap},
	{"WeatherDo", WeatherDo},
}

var people = []struct {
	name string
	p    Person
	want string
}{
	{"complete chain", Person{&Address{&City{&Weather{"sunny"}}}}, "sunny"},
	{"nil address", Person{}, noWeather},
	{"nil city", Person{&Address{}}, noWeather},
	{"nil weather", Person{&Address{&City{}}}, noWeather},
	{"empty description", Person{&Address{&City{&Weather{}}}}, ""},
}

// TestWeatherAgree is the point of the example: the three implementations must
// give the same answer, including on a chain that breaks at every depth.
//
// The nil cases are what fail if Maybe.Map uses slide 52's body rather than
// slide 78's. A Go method returns a typed nil pointer, so a plain
// m.Value == nil check never fires and the next step panics.
func TestWeatherAgree(t *testing.T) {
	for _, who := range people {
		for _, impl := range impls {
			if got := impl.f(who.p); got != who.want {
				t.Errorf("%s(%s) = %q, want %q", impl.name, who.name, got, who.want)
			}
		}
	}
}

// TestPersonWeather checks the method the deck actually declares, which
// survives as a wrapper around WeatherDo.
func TestPersonWeather(t *testing.T) {
	for _, who := range people {
		if got := who.p.Weather(); got != who.want {
			t.Errorf("Person.Weather(%s) = %q, want %q", who.name, got, who.want)
		}
	}
}

// TestMethodExpressions pins the shapes slide 60 depends on. Value receivers
// returning pointers are what make these usable as method expressions, and
// changing either half breaks WeatherMap and WeatherDo.
func TestMethodExpressions(t *testing.T) {
	var (
		_ func(Person) *Address = Person.Address
		_ func(Address) *City   = Address.City
		_ func(City) *Weather   = City.Weather
		_ func(Weather) string  = Weather.Description
	)
}
