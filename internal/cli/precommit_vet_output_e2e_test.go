package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// precommit_vet_output_e2e_test.go — the vet step shows WHY it blocked
// (card t768, GH #1679).
//
// The REAL hook content (preCommitHookContent) runs inside a real git
// repository in t.TempDir(), driven through `git commit`. A hook that
// discards vet's diagnostics still blocks the commit, so an exit-code-only
// assertion passes either way: the two cases below pin the OUTPUT, which is
// the thing that was missing.
//
// Both directions are asserted, because a fix that prints on failure but also
// turns a passing commit noisy trades one defect for another.
//
// The fixture carries a go.mod so the vet step's module walk finds a module
// and actually runs. The heavy gate stays out of scope: the hook's PATH is
// built from the toolchain directory plus the system directories, so `moai`
// is not reachable and the `command -v moai` guard skips that block.

const (
	// vetOutputClean passes go vet; vetOutputBroken compiles but hands
	// Printf a string for a %d verb, which vet reports.
	vetOutputClean  = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Printf(\"%s\\n\", \"ok\")\n}\n"
	vetOutputBroken = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Printf(\"%d\\n\", \"not a number\")\n}\n"

	// vetOutputDiagnostic is a fragment only `go vet` itself emits — the
	// hook's own failure message never contains it, so asserting on it
	// cannot be satisfied by the wrapper text alone.
	vetOutputDiagnostic = "wrong type string"
)

// writeVetOutputRepo builds a git repo holding a go.mod and a main.go with
// `src`, stages it, and returns a commit func reporting the hook's combined
// output and exit.
func writeVetOutputRepo(t *testing.T, src string) func() (string, error) {
	t.Helper()

	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH; the step under test is skipped by the hook itself")
	}
	hookPath := filepath.Dir(goPath) + ":/usr/bin:/bin:/usr/sbin:/sbin"

	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "hooks", "pre-commit"), []byte(preCommitHookContent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module vetoutput\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "PATH="+hookPath, "SKIP_MOAI_PRECOMMIT=")
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	if out, err := run("add", "go.mod", "main.go"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	return func() (string, error) {
		return run("-c", "user.email=t768@example.test", "-c", "user.name=t768", "commit", "-m", "t768 fixture")
	}
}

// TestPrecommitVetFailureShowsDiagnostics is the defect direction: the hook
// blocks on a vet finding, and the user must be able to read what the finding
// was. Without it the only available move is to turn the gate off.
func TestPrecommitVetFailureShowsDiagnostics(t *testing.T) {
	commit := writeVetOutputRepo(t, vetOutputBroken)

	out, err := commit()
	if err == nil {
		t.Fatalf("commit succeeded; the hook must block a vet finding\n%s", out)
	}
	if !strings.Contains(out, "go vet reported") {
		t.Errorf("hook did not report a vet failure\n%s", out)
	}
	if !strings.Contains(out, vetOutputDiagnostic) {
		t.Errorf("vet diagnostics were swallowed: output lacks %q\n%s", vetOutputDiagnostic, out)
	}
}

// TestPrecommitVetPassStaysQuiet is the control: a vet-clean commit must go
// through without the step announcing itself. Showing the failure is only a
// fix if the passing path stays silent.
func TestPrecommitVetPassStaysQuiet(t *testing.T) {
	commit := writeVetOutputRepo(t, vetOutputClean)

	out, err := commit()
	if err != nil {
		t.Fatalf("commit blocked on vet-clean source: %v\n%s", err, out)
	}
	if strings.Contains(out, "[pre-commit]") {
		t.Errorf("passing commit emitted hook output\n%s", out)
	}
	if strings.Contains(out, vetOutputDiagnostic) {
		t.Errorf("passing commit emitted vet diagnostics\n%s", out)
	}
}
