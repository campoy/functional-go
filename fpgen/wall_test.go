package fpgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// goMinorVersion reports the minor version of the `go` command on PATH --
// the toolchain that actually compiles the testdata files below, which is
// not necessarily the one that built this test binary. It returns -1 when
// the version cannot be parsed (a devel or otherwise unusual toolchain).
func goMinorVersion(t *testing.T) int {
	t.Helper()
	out, err := exec.Command("go", "version").Output()
	require.NoError(t, err, "could not run `go version`")

	// "go version go1.27.0 darwin/arm64"
	fields := strings.Fields(string(out))
	if len(fields) < 3 || !strings.HasPrefix(fields[2], "go1.") {
		return -1
	}
	parts := strings.SplitN(strings.TrimPrefix(fields[2], "go1."), ".", 2)
	minor, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return minor
}

// skipUnlessGo127 skips a test that needs a go1.27 or later toolchain.
// -lang=go1.27 is not merely ignored by an older toolchain, it is rejected
// outright ("invalid value \"go1.27\" for -lang: max known version is ..."),
// and the go1.27-specific diagnostics these tests pin do not exist there --
// so on CI's go1.21 matrix leg there is nothing to assert, only a toolchain
// that cannot be asked the question.
func skipUnlessGo127(t *testing.T) {
	t.Helper()
	switch minor := goMinorVersion(t); {
	case minor < 0:
		t.Skip("could not determine the version of the `go` command on PATH; this test needs go1.27 or later")
	case minor < 27:
		t.Skipf("needs a go1.27 or later toolchain: go1.%d rejects -gcflags=-lang=go1.27 and does not emit the go1.27 diagnostics this test pins", minor)
	}
}

// wantGo121Diagnostic is the exact line a go1.27 or later toolchain prints
// for testdata/wall_method_type_param.go at -lang=go1.21. It is quoted
// verbatim by fpgen/list.go's ListMap doc comment and by lesson 5 of
// docs/teaching-generics.md; TestWallMethodTypeParams below is what keeps
// all three copies honest.
const wantGo121Diagnostic = "testdata/wall_method_type_param.go:17:23: generic method requires go1.27 or later (-lang was set to go1.21; check go.mod)"

// TestWallMethodTypeParams runs the real go build on
// testdata/wall_method_type_param.go and pins the real compiler diagnostic,
// so lesson 5's claim in docs/teaching-generics.md (methods cannot have type
// parameters) cannot silently rot into a stale claim about the language.
//
// -lang=go1.21 matches this module's go.mod floor exactly, on purpose: Go
// 1.27 lifted this restriction (see the "generic method requires go1.27 or
// later" text below), so building the file with the ambient toolchain's
// default language version -- rather than the version this module actually
// declares -- would silently stop demonstrating the wall the day the
// installed toolchain default caught up. Pinning -lang to the module's own
// floor is what keeps the test honest about what THIS repository can
// express, independent of which toolchain happens to be installed.
func TestWallMethodTypeParams(t *testing.T) {
	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.21", "testdata/wall_method_type_param.go")
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "wall_method_type_param.go must NOT compile under go1.21 language semantics")

	// Two spellings, one wall. A go1.27 or later toolchain knows the
	// restriction was lifted and reports the version gate; a toolchain older
	// than 1.27 has no such gate to report and rejects the type parameter
	// list in the parser instead. Either diagnostic proves the same thing
	// about what THIS module, pinned at go1.21, can express.
	assert.Regexp(t,
		`generic method requires go1\.27 or later|method must have no type parameters`,
		string(out),
		"expected the compiler to reject a type parameter list on a method")

	// On a toolchain that can produce it, pin the whole line -- position
	// included. fpgen/list.go and docs/teaching-generics.md quote this text
	// verbatim, and without the position pinned an edit to the testdata
	// file's header silently shifts the line number out from under both.
	if goMinorVersion(t) >= 27 {
		assert.Contains(t, string(out), wantGo121Diagnostic,
			"the diagnostic quoted verbatim in fpgen/list.go and docs/teaching-generics.md must match what the compiler actually prints")
	}
}

// TestWallMethodTypeParamsLiftedAtGo127 is the positive half of lesson 5's
// nuance: the exact same file that fails to compile at -lang=go1.21
// (TestWallMethodTypeParams above) compiles cleanly at -lang=go1.27. This is
// what makes lesson 10's closing argument precise rather than reused: lesson
// 5's wall is version-gated and fell in 1.27; lesson 10's is not, and
// TestWallInterfaceMethodTypeParams below is what still stands there.
func TestWallMethodTypeParamsLiftedAtGo127(t *testing.T) {
	skipUnlessGo127(t)

	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.27", "testdata/wall_method_type_param.go")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "a generic method on a concrete type must compile at go1.27: %s", out)
}

// TestWallInterfaceMethodTypeParams pins lesson 10's real diagnostic: unlike
// a generic method on a concrete type (lifted in 1.27, see above), a type
// parameter on a method declared INSIDE an interface is still rejected at
// go1.27 -- a separate, narrower rule than the one lesson 5's wall relied
// on, and not gated by a language version the way that one was. This is why
// fp's Mapper (slide 47) and fpgen's still cannot be written: the first
// failure reason changed out from under the talk's complaint, but the
// complaint itself survives on a sturdier rule.
//
// The source lives in testdata/wall_interface_method_type_param.go.txt, not
// .go: go/parser (what gofmt uses) treats a type-parameter list on an
// interface method as a syntax error rather than a type-check error, which
// would make `gofmt -l .` fail on this repository. This test copies the
// text into a temporary .go file so `go build` still sees and rejects it.
func TestWallInterfaceMethodTypeParams(t *testing.T) {
	skipUnlessGo127(t)

	src, err := os.ReadFile("testdata/wall_interface_method_type_param.go.txt")
	require.NoError(t, err)

	dir := t.TempDir()
	tmpGo := filepath.Join(dir, "wall_interface_method_type_param.go")
	require.NoError(t, os.WriteFile(tmpGo, src, 0o644))

	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.27", tmpGo)
	out, buildErr := cmd.CombinedOutput()
	require.Error(t, buildErr, "wall_interface_method_type_param.go.txt must NOT compile: interface methods can't have type parameters")
	assert.Contains(t, string(out), "interface method must have no type parameters")
}

// wantComposeDiagnostic is the exact line a go1.27 or later toolchain prints
// for testdata/compose_mismatch.go. Like wantGo121Diagnostic above it is
// quoted verbatim -- by fpgen/func.go's Compose doc comment and by lesson 3
// of docs/teaching-generics.md -- so pinning it here is what stops those
// copies drifting from what the compiler really says. Type inference error
// text is toolchain-dependent and not stabilised by -lang, which is why the
// pin is guarded and the unconditional assertion below is the looser one.
const wantComposeDiagnostic = "testdata/compose_mismatch.go:18:27: in call to Compose, type func(n int) int of func(n int) int {…} does not match inferred type func(int) string for func(A) B"

// wantVariadicChainDiagnostic is the exact line a go1.27 or later toolchain
// prints for testdata/wall_variadic_chain.go. Quoted verbatim by
// fpgen/chain.go and by lesson 9 of docs/teaching-generics.md.
const wantVariadicChainDiagnostic = "testdata/wall_variadic_chain.go:44:12: in call to Do, type func(Person) *Address of Person.Address does not match inferred type func(Person) Person for func(T) T"

// wantVariadicChainDiagnosticPre127 is the same rejection as go1.21 words
// it. The two differ only by the "in call to Do, " clause, which go1.21's
// inference path does not emit; position, types and reason are identical.
const wantVariadicChainDiagnosticPre127 = "testdata/wall_variadic_chain.go:44:12: type func(Person) *Address of Person.Address does not match inferred type func(Person) Person for func(T) T"

// variadicChainDiagnosticRE matches whichever of the two spellings above the
// running toolchain produces, and nothing else. It is built from the two
// constants rather than restating them, so each remains pinned character for
// character and there is only one copy of each to keep true.
var variadicChainDiagnosticRE = regexp.MustCompile(
	regexp.QuoteMeta(wantVariadicChainDiagnostic) + "|" +
		regexp.QuoteMeta(wantVariadicChainDiagnosticPre127))

// TestWallVariadicChain pins lesson 9's wall, THE SECOND WALL -- and pins it
// at the place it actually stands. Lessons 5 and 10 are compiler rejections
// of a DECLARATION; this one is not. `func Do[T any](v T, fs ...func(T) T) T`
// declares perfectly legally, which is exactly why the wall is easy to miss:
// it appears only at the CALL SITE, when that legal signature is handed
// slide 61's chain, whose element type changes at every step. Nothing about
// the file fails to parse or fails to declare; the call is what the compiler
// rejects.
func TestWallVariadicChain(t *testing.T) {
	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.21", "testdata/wall_variadic_chain.go")
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "wall_variadic_chain.go must NOT compile: a homogeneous variadic can't take a heterogeneous chain")

	// Two spellings, one wall: go1.27 prefixes the offending argument with
	// "in call to Do, " and go1.21 does not. Everything that carries the
	// lesson -- the position, that Person.Address is func(Person) *Address,
	// and that a single T cannot be both -- is identical, so this holds on
	// any toolchain without accepting an unrelated failure.
	assert.Regexp(t, variadicChainDiagnosticRE, string(out),
		"expected the compiler to reject the heterogeneous chain at the call to Do")

	// On a toolchain that can produce it, pin the line fpgen/chain.go and
	// docs/teaching-generics.md quote, character for character.
	if goMinorVersion(t) >= 27 {
		assert.Contains(t, string(out), wantVariadicChainDiagnostic,
			"the diagnostic quoted verbatim in fpgen/chain.go and docs/teaching-generics.md must match what the compiler actually prints")
	}
}

// TestWallVariadicChainSignatureItselfCompiles is the other half of lesson
// 9's point, and the reason the wall needed a call site to show up at all:
// the same generic signature the file above declares is legal Go. Chain2
// and Chain3 below it are legal too. What cannot be written is a single
// variadic signature that accepts a chain of changing types.
func TestWallVariadicChainSignatureItselfCompiles(t *testing.T) {
	dir := t.TempDir()
	src := "package main\n\nfunc Do[T any](v T, fs ...func(T) T) T {\n\tfor _, f := range fs {\n\t\tv = f(v)\n\t}\n\treturn v\n}\n\nfunc main() { _ = Do(1, func(n int) int { return n + 1 }) }\n"
	tmpGo := filepath.Join(dir, "variadic_chain_decl.go")
	require.NoError(t, os.WriteFile(tmpGo, []byte(src), 0o644))

	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "out"), "-gcflags=-lang=go1.21", tmpGo)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "the homogeneous variadic signature must compile on its own; the wall is at the call site, not the declaration: %s", out)
}

// TestWallComposeMismatch pins the real diagnostic from lesson 3: a
// Compose (fpgen/func.go) call whose two functions do not line up fails to
// compile, with the compiler doing the type check fp.Compose does at run
// time via g.out != f.in.
func TestWallComposeMismatch(t *testing.T) {
	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.21", "testdata/compose_mismatch.go")
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "compose_mismatch.go must NOT compile: Compose's two functions don't line up")
	assert.Contains(t, string(out), "does not match inferred type")

	if goMinorVersion(t) >= 27 {
		assert.Contains(t, string(out), wantComposeDiagnostic,
			"the diagnostic quoted verbatim in fpgen/func.go and docs/teaching-generics.md must match what the compiler actually prints")
	}
}
