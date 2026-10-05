package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// precommit_staged_read_e2e_test.go — the gofmt step judges the bytes being
// committed, not the bytes on disk (card t606, hooks audit H11).
//
// The REAL hook content (preCommitHookContent) runs inside a real git
// repository in t.TempDir(), driven through `git commit` so the index and the
// working tree are genuinely separate. Each case stages one version of a Go
// file and then overwrites the file on disk with the other, so a hook reading
// the wrong side reaches the opposite verdict from the right one — the two
// mismatch cases are mutually exclusive, and no hook satisfies both by
// accident.
//
// The fixture holds no go.mod on purpose: the vet step's module walk then
// finds no module and skips, leaving gofmt as the only step under test. The
// heavy gate is likewise out of scope — with no .moai/config the runner skips
// its project-wide steps under MOAI_PRECOMMIT=1.

const (
	// stagedReadFormatted is gofmt-clean; stagedReadUnformatted is the same
	// program with the spacing gofmt would rewrite.
	stagedReadFormatted   = "package main\n\nfunc main() {\n\tx := 1\n\t_ = x\n}\n"
	stagedReadUnformatted = "package main\nfunc  main( ) {\nx:=1\n_ = x\n}\n"
)

// stagedReadFailingGofmt is a gofmt that cannot run: non-zero exit, nothing on
// stdout. It stands for every way the tool can fail to produce a verdict —
// /dev/stdin missing, input unreadable, staged bytes that do not parse. NOT
// production code.
const stagedReadFailingGofmt = `#!/bin/sh
exit 1
`

// writeStagedReadRepo builds a git repo with the real hook installed, stages
// main.go holding `staged`, then overwrites the file on disk with `worktree`.
// It returns a commit func reporting the hook's combined output and exit.
func writeStagedReadRepo(t *testing.T, staged, worktree string) func() (string, error) {
	t.Helper()
	return writeStagedReadRepoWithGofmt(t, staged, worktree, "")
}

// writeStagedReadRepoWithGofmt is writeStagedReadRepo with the gofmt on the
// hook's PATH replaced by `gofmtStub` when that is non-empty.
func writeStagedReadRepoWithGofmt(t *testing.T, staged, worktree, gofmtStub string) func() (string, error) {
	t.Helper()

	var hookPath string
	if gofmtStub == "" {
		gofmtPath, err := exec.LookPath("gofmt")
		if err != nil {
			t.Skip("gofmt not on PATH; the step under test is skipped by the hook itself")
		}
		hookPath = filepath.Dir(gofmtPath) + ":/usr/bin:/bin:/usr/sbin:/sbin"
	}

	root := t.TempDir()
	if gofmtStub != "" {
		binDir := filepath.Join(root, ".stub-bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, "gofmt"), []byte(gofmtStub), 0o755); err != nil {
			t.Fatal(err)
		}
		// The stub dir comes first, and the real toolchain dir is left off
		// entirely, so nothing behind it can answer for gofmt.
		hookPath = binDir + ":/usr/bin:/bin:/usr/sbin:/sbin"
	}
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "hooks", "pre-commit"), []byte(preCommitHookContent), 0o755); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "PATH="+hookPath, "SKIP_MOAI_PRECOMMIT=")
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	mainGo := filepath.Join(root, "main.go")
	if err := os.WriteFile(mainGo, []byte(staged), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := run("add", "main.go"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	// Diverge the working tree from the index. This is the whole fixture.
	if err := os.WriteFile(mainGo, []byte(worktree), 0o644); err != nil {
		t.Fatal(err)
	}

	return func() (string, error) {
		return run("-c", "user.email=t606@example.test", "-c", "user.name=t606", "commit", "-m", "t606 fixture")
	}
}

// TestPrecommitStagedUnformattedWorktreeFormattedBlocks is the defect
// direction: the index carries unformatted bytes while the file on disk was
// already fixed. The commit would record the unformatted version, so the hook
// must block it — reading the working file lets it through.
func TestPrecommitStagedUnformattedWorktreeFormattedBlocks(t *testing.T) {
	commit := writeStagedReadRepo(t, stagedReadUnformatted, stagedReadFormatted)

	out, err := commit()
	if err == nil {
		t.Fatalf("commit succeeded although the staged bytes need formatting (the hook read the working file):\n%s", out)
	}
	if !strings.Contains(out, "need formatting") || !strings.Contains(out, "main.go") {
		t.Errorf("block did not come from the gofmt step naming the file:\n%s", out)
	}
}

// TestPrecommitStagedFormattedWorktreeUnformattedAllows is the opposite
// direction: the index is clean and only the file on disk is unformatted.
// The commit records the clean version, so the hook must allow it.
func TestPrecommitStagedFormattedWorktreeUnformattedAllows(t *testing.T) {
	commit := writeStagedReadRepo(t, stagedReadFormatted, stagedReadUnformatted)

	if out, err := commit(); err != nil {
		t.Fatalf("commit blocked although the staged bytes are formatted (the hook read the working file):\n%s", out)
	}
}

// TestPrecommitStagedFormattedNoDivergenceAllows is the plain positive
// control: index and working tree agree and are clean. Without it the
// allow-direction above could pass on a hook that never blocks anything.
func TestPrecommitStagedFormattedNoDivergenceAllows(t *testing.T) {
	commit := writeStagedReadRepo(t, stagedReadFormatted, stagedReadFormatted)

	if out, err := commit(); err != nil {
		t.Fatalf("commit blocked on a formatted file with no divergence:\n%s", out)
	}
}

// TestPrecommitStagedGofmtToolFailureBlocks pins the direction a silent gate
// fails in. The staged bytes are clean, so nothing here needs formatting —
// but gofmt itself cannot run. Deciding on empty stdout alone would report a
// pass, which reads as a verdict the gate never reached; the exit status is
// what separates "checked, clean" from "could not check".
func TestPrecommitStagedGofmtToolFailureBlocks(t *testing.T) {
	commit := writeStagedReadRepoWithGofmt(t, stagedReadFormatted, stagedReadFormatted, stagedReadFailingGofmt)

	out, err := commit()
	if err == nil {
		t.Fatalf("commit succeeded although gofmt could not run (a silent pass):\n%s", out)
	}
	if !strings.Contains(out, "could not check the staged bytes") {
		t.Errorf("block did not name the tool failure:\n%s", out)
	}
}

// TestPrecommitStagedUnformattedNoDivergenceBlocks is the plain negative
// control: index and working tree agree and are unformatted. Without it the
// block-direction above could pass on a hook that blocks everything.
func TestPrecommitStagedUnformattedNoDivergenceBlocks(t *testing.T) {
	commit := writeStagedReadRepo(t, stagedReadUnformatted, stagedReadUnformatted)

	out, err := commit()
	if err == nil {
		t.Fatalf("commit succeeded on an unformatted file with no divergence:\n%s", out)
	}
	if !strings.Contains(out, "need formatting") {
		t.Errorf("block did not come from the gofmt step:\n%s", out)
	}
}
