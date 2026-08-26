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
// the appendix, adds a check for a nil pointer, and that is the one
// implemented here: a Go method returns a typed nil pointer, not a nil
// interface, so slide 52's version sees a non-nil m.Value and keeps calling
// down a chain that has already broken. The weather example does not work
// without this. See NOTES.md.
//
// Slide 78 applies that check to the value coming out of f. It is applied to
// the value going in as well, since a Maybe can be handed a typed nil to begin
// with -- Maybe{p.Address()} starts one link down slide 61's chain -- and
// nothing about that nil is different from one a step produced. See NOTES.md.
func (m Maybe) Map(f *Func) Maybe {
	if m.Value == nil || isNilPtr(m.Value) {
		return Maybe{}
	}
	r := f.Call(m.Value)
	if isNilPtr(r) {
		return Maybe{}
	}
	return Maybe{r}
}

// isNilPtr reports whether x is a nil pointer wrapped in a non-nil interface,
// which is what a Go method returns when it has nothing to give back.
func isNilPtr(x interface{}) bool {
	v := reflect.ValueOf(x)
	return v.Kind() == reflect.Ptr && v.IsNil()
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
	if len(chain) > 0 {
		if err := checkStart(m.Value, chain[0].in); err != nil {
			return Maybe{}, err
		}
	}
	for _, f := range chain {
		m = m.Map(f)
	}
	return m, nil
}
