package fpgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
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
