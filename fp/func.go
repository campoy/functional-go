// Package fp reconstructs the small functional library built in "Functional
// Go?" (dotGo 2015), slides 30 through 78.
//
// The whole library is one idea applied three times: Func is a runtime-typed,
// reflection-backed function value, and each container -- List, Maybe, Many --
// knows how to Map one over itself.
//
// Slides 24 through 29 are the setup. With generics you would write
//
//	func Map(f func(a α) β, vs []α) []β
//
// but Go 1.5 had none, so every parameter collapses to interface{} and all
// type safety is lost. Func is the answer: it carries the reflect.Type
// information the interface{} signature threw away, and Compose uses it to
// reject a mismatched pair the way a compiler would.
//
// Slide 47 is why the three containers share no interface. Mapper would have
// to be declared
//
//	type Mapper interface {
//		Map(*Func) ???
//	}
//
// and Go has no way to spell the return type. The duplicated Map methods are
// the point of the talk, not an oversight.
//
// No generics are used anywhere, deliberately: the talk exists because the
// language did not have them. See NOTES.md.
package fp

import (
	"fmt"
	"reflect"
)

// Func is a function of exactly one argument and exactly one result, wrapped
// so it can be applied to an interface{} while keeping the argument and result
// types available at runtime (slide 30).
//
// The field order matters: Compose builds a Func with a positional composite
// literal, as slide 77 prints it.
type Func struct {
	in  reflect.Type
	out reflect.Type
	f   func(interface{}) interface{}
}

// Call applies f to v (slide 30).
//
// The receiver is a value, as the slide shows, even though NewFunc and Compose
// hand back a *Func.
func (f Func) Call(v interface{}) interface{} {
	return f.f(v)
}

// NewFunc wraps f, which must be a function taking exactly one argument and
// returning exactly one value (slide 31).
//
// Slide 31 elides the checking with the comment "check type of f and return an
// error if needed". Leaving it out would make the error result a lie, so this
// rejects, with a descriptive error: a nil or non-function argument, a
// variadic function, a function that does not take exactly one argument, and
// one that does not return exactly one value.
//
// Variadic functions are rejected rather than accepted as one-argument
// functions. reflect would report the input type of func(vs ...string) string
// as []string while reflect.Value.Call would happily take a single string, so
// admitting them would make Compose and Do type-check against a type the
// function never really sees.
func NewFunc(f interface{}) (*Func, error) {
	if f == nil {
		return nil, fmt.Errorf("can't build a Func from nil")
	}
	vf := reflect.ValueOf(f)
	tf := vf.Type()
	if tf.Kind() != reflect.Func {
		return nil, fmt.Errorf("%v is not a function", tf)
	}
	if vf.IsNil() {
		return nil, fmt.Errorf("%v is a nil function", tf)
	}
	if tf.IsVariadic() {
		return nil, fmt.Errorf("%v is variadic, want exactly one argument", tf)
	}
	if tf.NumIn() != 1 {
		return nil, fmt.Errorf("%v takes %d arguments, want exactly one", tf, tf.NumIn())
	}
	if tf.NumOut() != 1 {
		return nil, fmt.Errorf("%v returns %d values, want exactly one", tf, tf.NumOut())
	}

	in, out := tf.In(0), tf.Out(0)
	return &Func{
		in:  in,
		out: out,
		f: func(x interface{}) interface{} {
			res := vf.Call([]reflect.Value{argValue(in, x)})
			return res[0].Interface()
		},
	}, nil
}

// Must returns f, panicking if err is not nil (slide 32).
//
// It exists so that NewFunc can be used inline, as every usage slide does:
//
//	toUpper := Must(NewFunc(strings.ToUpper))
func Must(f *Func, err error) *Func {
	if err != nil {
		panic(err)
	}
	return f
}

// Compose returns the composition of f and g, applying g first: the result
// computes f(g(x)) (slide 77).
//
// The argument order is mathematical composition, not pipeline order. It reads
// backwards to most people, and it is kept exactly as the slide prints it.
//
// The check is the punchline of the talk: g.out != f.in is the static type
// system, reimplemented by hand at runtime because the interface{} signatures
// threw it away.
func Compose(f, g *Func) (*Func, error) {
	if g.out != f.in {
		return nil, fmt.Errorf("can't compose: %v != %v", g.out, f.in)
	}
	return &Func{
		g.in,
		f.out,
		func(x interface{}) interface{} { return f.Call(g.Call(x)) },
	}, nil
}

// argValue converts x into a reflect.Value that can be passed to a function
// whose only parameter has type in.
//
// The interesting case is the automatic dereference, which the deck needs but
// never mentions. Slide 61 chains method expressions:
//
//	Map(Must(NewFunc(Person.Address))).
//	Map(Must(NewFunc(Address.City)))
//
// Person.Address is func(Person) *Address, but Address.City is
// func(Address) *City -- a value receiver. The chain therefore hands a
// *Address to a function wanting an Address, and reflect.Value.Call would
// panic. Dereferencing here is what makes slides 61 and 63 work as printed.
// See NOTES.md.
//
// A nil pointer has nothing to dereference, and there is no honest value to
// substitute: Call has no error result, so it panics rather than invent one.
// Maybe.Map short-circuits on a nil pointer at both ends, so a Maybe chain
// never reaches this; a Many chain holding one does. See NOTES.md.
func argValue(in reflect.Type, x interface{}) reflect.Value {
	v := reflect.ValueOf(x)
	if !v.IsValid() {
		return reflect.Zero(in)
	}
	if !v.Type().AssignableTo(in) && v.Kind() == reflect.Ptr && v.Type().Elem().AssignableTo(in) {
		if v.IsNil() {
			panic(fmt.Sprintf("can't dereference a nil %v to call a function taking %v", v.Type(), in))
		}
		return v.Elem()
	}
	return v
}

// canCall reports whether a value of type t can be handed to a Func whose
// input type is in, allowing for the dereference argValue performs.
func canCall(t, in reflect.Type) bool {
	if t.AssignableTo(in) {
		return true
	}
	return t.Kind() == reflect.Ptr && t.Elem().AssignableTo(in)
}

// checkStart reports whether v, a value already sitting in a container when Do
// is called, can be handed to the first step of the chain.
//
// A nil value is not an error: Maybe short-circuits on it and argValue turns an
// invalid value into the zero value, so neither container ever passes it to
// reflect.Value.Call unchanged.
func checkStart(v interface{}, in reflect.Type) error {
	t := reflect.TypeOf(v)
	if t == nil || canCall(t, in) {
		return nil
	}
	return fmt.Errorf("can't chain: the starting value is a %v, but step 0 takes %v", t, in)
}

// newChain turns fs into Funcs and checks that each step's result can be fed
// to the next one, so that Maybe.Do and Many.Do report a type error rather
// than panicking inside reflect.Value.Call.
//
// compat decides what "can be fed to" means, because the two containers differ:
// Maybe passes the result straight along, while Many flattens a slice first.
//
// Slide 62's Do only surfaces errors from NewFunc. Type-checking the joints is
// an addition beyond the deck; see NOTES.md.
func newChain(fs []interface{}, compat func(out, in reflect.Type) bool) ([]*Func, error) {
	chain := make([]*Func, len(fs))
	for i, v := range fs {
		f, err := NewFunc(v)
		if err != nil {
			return nil, err
		}
		if i > 0 && !compat(chain[i-1].out, f.in) {
			return nil, fmt.Errorf("can't chain: step %d returns %v, but step %d takes %v",
				i-1, chain[i-1].out, i, f.in)
		}
		chain[i] = f
	}
	return chain, nil
}
