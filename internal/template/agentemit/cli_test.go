// cli_test.go — the agents-emit command body: `make agents-emit` regenerates,
// `make agents-emit-check` compares and prints the summary line only on a match.
package agentemit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

func runCLI(t *testing.T, root string, lay agentemit.Layout, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = agentemit.RunCLI(root, lay, mustManifest(t), args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunCLIRegeneratesThenChecks(t *testing.T) {
	root, lay := fixtureRoot(t, twoSources())

	code, out, errOut := runCLI(t, root, lay)
	if code != 0 || out != "agents-emit: specs=2 copies=4\n" || errOut != "" {
		t.Fatalf("regenerate: code=%d stdout=%q stderr=%q", code, out, errOut)
	}
	code, out, errOut = runCLI(t, root, lay, "-check")
	if code != 0 || out != "agents-emit-check: specs=2 copies=4 diff=0\n" || errOut != "" {
		t.Fatalf("check on a fresh tree: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	if err := os.WriteFile(filepath.Join(root, "out/c2/beta.md"), []byte("hand edit"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errOut = runCLI(t, root, lay, "-check")
	if code != 1 {
		t.Errorf("check after a hand edit: exit code %d, want 1", code)
	}
	if out != "" {
		t.Errorf("a failing check must not print the summary line on stdout, got %q", out)
	}
	if !strings.Contains(errOut, "out/c2/beta.md") {
		t.Errorf("stderr must name the differing file, got %q", errOut)
	}
}

func TestRunCLIRefusals(t *testing.T) {
	root, lay := fixtureRoot(t, twoSources())
	if code, _, errOut := runCLI(t, root, lay, "--bogus"); code != 2 || !strings.Contains(errOut, "--bogus") {
		t.Errorf("unknown argument: code=%d stderr=%q, want 2 naming it", code, errOut)
	}
	bad := map[string]string{"bad.md": fixtureMD("bad", "Bad.\n", "Read, Bogus", "inherit", "low", nil, "\nbody\n")}
	badRoot, badLay := fixtureRoot(t, bad)
	badLay.Targets = badLay.Targets[:2]
	for _, args := range [][]string{nil, {"-check"}} {
		if code, out, errOut := runCLI(t, badRoot, badLay, args...); code != 1 || out != "" || !strings.Contains(errOut, "Bogus") {
			t.Errorf("args %v over an unmappable source: code=%d stdout=%q stderr=%q", args, code, out, errOut)
		}
	}
}
