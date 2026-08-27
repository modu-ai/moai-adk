package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRunStep_RunsInConfiguredProjectDir pins the working directory of every
// quality-gate subprocess to the configured project dir.
//
// runStep used to build its exec.Cmd without setting cmd.Dir, so each step
// ("go vet ./...", "golangci-lint run", "go test ./...", and their non-Go
// counterparts — every one of which takes a cwd-relative target) ran in
// whatever directory the calling process happened to occupy, not in
// cfg.ProjectDir. In production the moai binary's cwd usually equals the
// project root, which kept the defect latent; under `go test` the cwd is the
// package directory of the test being run, so a gate pointed at a temp fixture
// silently graded the real repository instead — and, when the step was
// "go test ./...", re-executed the very suite that had invoked it. That
// recursion is what filled TMPDIR with thousands of orphaned t.TempDir()
// fixtures: each recursive test binary was killed before its cleanups ran.
//
// The fixture below carries a printf verb/arg mismatch that `go vet` reliably
// flags. A step honouring cfg.ProjectDir fails on it and names the fixture
// path; a step running anywhere else never sees bad.go at all.
func TestRunStep_RunsInConfiguredProjectDir(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH — the vet step cannot run")
	}

	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	vetBad := "package sample\n\nimport \"fmt\"\n\n// VetBad triggers go vet's printf verb/arg check.\nfunc VetBad() { fmt.Printf(\"%d\", \"string-arg\") }\n"
	if err := os.WriteFile(filepath.Join(repo, "gate_cwd_probe.go"), []byte(vetBad), 0o644); err != nil {
		t.Fatalf("write gate_cwd_probe.go: %v", err)
	}

	cfg := DefaultGateConfig()
	cfg.ProjectDir = repo
	g := NewQualityGate(cfg)

	ok, output := g.runStep(context.Background(), "go vet", "", 60*time.Second, "go", "vet", "./...")
	if ok {
		t.Fatalf("runStep passed on a fixture whose only package fails `go vet` — the step did not run in cfg.ProjectDir (%s) but in the calling process's cwd", repo)
	}

	// `go vet` reports diagnostics relative to its own cwd, so the fixture dir
	// never appears in the output — the fixture's uniquely-named file does, and
	// only if the step actually ran there.
	if !strings.Contains(output, "gate_cwd_probe.go") {
		t.Fatalf("vet diagnostics do not name the fixture file — the step failed for some other reason, not on the fixture (%s).\noutput:\n%s", repo, output)
	}
}
