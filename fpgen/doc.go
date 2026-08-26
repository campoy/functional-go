// Package fpgen is the generics counterpart to fp: the same library, rebuilt
// on Go's type parameters (Go 1.18+) instead of reflect.
//
// fp exists because Go 1.5 had no generics: Func carries by hand, at run
// time, the type information a compiler with type parameters would carry
// for free, and Compose's g.out != f.in check is the static type system
// reimplemented in reflect.Type comparisons. fpgen asks the question the
// talk could not: what does this library look like once the language can
// answer it?
//
// Read docs/teaching-generics.md alongside this package. It is not a
// generics tutorial that happens to use lists -- every concept it introduces
// is motivated by a specific thing fp had to do the hard way, with the fp
// code and the fpgen code shown side by side. Two of the talk's shapes still
// do not fit: Go's generics cannot parameterise a method (fpgen's List, Maybe
// and Many still cannot share a Mapper interface, same as fp) and cannot
// express fp.Maybe.Do's heterogeneous variadic chain. Those walls are lessons
// 5, 9 and 10 in the curriculum, not omissions from this package.
//
// fpgen is stdlib-only in non-test code, the same standing constraint as fp.
// go.mod stays at go 1.21 deliberately -- see NOTES.md.
package fpgen
