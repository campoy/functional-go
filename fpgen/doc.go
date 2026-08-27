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
// code and the fpgen code shown side by side. Three of the talk's shapes
// still do not fit, and each is pinned to real compiler output in the
// curriculum rather than asserted:
//
//   - Map, ListMap, MaybeMap and ManyMap are free functions because a method
//     cannot take type parameters of its own at the go 1.21 this module
//     declares. Go 1.27 lifts that for a concrete type's methods, so this
//     wall is real for this repository rather than for the language
//     (lesson 5).
//   - fp.Maybe.Do's heterogeneous variadic chain has no generic signature,
//     because Go's variadics are homogeneous. Chain2, Chain3 and Chain4 in
//     chain.go are the honest fixed-arity substitute, and they do not
//     generalise (lesson 9).
//   - fpgen's List, Maybe and Many still cannot share a Mapper interface,
//     same as fp (slide 47), on two independent grounds: a type parameter on
//     a method declared inside an interface is rejected, still at go 1.27
//     and with no version gate in sight, and Go cannot abstract over a type
//     constructor, so the higher-kinded shape has no declaration to reject
//     at all (lesson 10).
//
// Those walls are the lessons, not omissions from this package.
//
// fpgen is stdlib-only in non-test code, the same standing constraint as fp.
// go.mod stays at go 1.21 deliberately -- see NOTES.md.
package fpgen
