package fpgen_test

import (
	"os/exec"
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
