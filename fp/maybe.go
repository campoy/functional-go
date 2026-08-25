package fp

import "reflect"

// Maybe holds a value that may be missing (slide 52).
//
// Unlike List and Many it is used as a value, never a pointer, and callers
// read Value directly -- slides 53 through 55, 61 and 63 all end with
// res.Value.
type Maybe struct {
	Value interface{}
}

// Map returns Maybe{f(m.Value)}, or the empty Maybe if there is nothing to
// apply f to (slide 78).
//
// The deck gives two bodies. Slide 52 checks only m.Value == nil. Slide 78, in
// the appendix, adds the check below for a nil pointer, and that is the one
// implemented here: a Go method returns a typed nil pointer, not a nil
// interface, so slide 52's version sees a non-nil m.Value and keeps calling
// down a chain that has already broken. The weather example does not work
// without this. See NOTES.md.
func (m Maybe) Map(f *Func) Maybe {
	if m.Value == nil {
		return Maybe{}
	}
	r := f.Call(m.Value)
	vr := reflect.ValueOf(r)
	if vr.Kind() == reflect.Ptr && vr.IsNil() {
		return Maybe{}
	}
	return Maybe{r}
}

// Do maps each of fs over m in turn, short-circuiting as Map does, and returns
// an error if the chain does not type-check (slides 62 and 63).
//
// It is the ergonomic layer: callers pass plain functions and never mention
// Func, Must or NewFunc.
//
//	w, err := Maybe{p}.Do(Person.Address, Address.City, City.Weather, Weather.Description)
//
// Slide 62 builds the chain one step at a time and only reports errors coming
// out of NewFunc, so a step whose result type does not fit the next step's
// argument panics inside reflect.Value.Call instead. Here the whole chain is
// built and checked before anything is applied, so a mismatch comes back as an
// error. That is an addition beyond the deck; see NOTES.md.
func (m Maybe) Do(fs ...interface{}) (Maybe, error) {
	chain, err := newChain(fs, canCall)
	if err != nil {
		return Maybe{}, err
	}
	for _, f := range chain {
		m = m.Map(f)
	}
	return m, nil
}
