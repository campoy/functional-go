package fpgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Contains(t, string(out), "generic method requires go1.27 or later")
}

// TestWallMethodTypeParamsLiftedAtGo127 is the positive half of lesson 5's
// nuance: the exact same file that fails to compile at -lang=go1.21
// (TestWallMethodTypeParams above) compiles cleanly at -lang=go1.27. This is
// what makes lesson 10's closing argument precise rather than reused: lesson
// 5's wall is version-gated and fell in 1.27; lesson 10's is not, and
// TestWallInterfaceMethodTypeParams below is what still stands there.
func TestWallMethodTypeParamsLiftedAtGo127(t *testing.T) {
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

// TestWallComposeMismatch pins the real diagnostic from lesson 3: a
// Compose (fpgen/func.go) call whose two functions do not line up fails to
// compile, with the compiler doing the type check fp.Compose does at run
// time via g.out != f.in.
func TestWallComposeMismatch(t *testing.T) {
	cmd := exec.Command("go", "build", "-gcflags=-lang=go1.21", "testdata/compose_mismatch.go")
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "compose_mismatch.go must NOT compile: Compose's two functions don't line up")
	assert.Contains(t, string(out), "does not match inferred type")
}
