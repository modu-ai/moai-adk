// cutover_fixture_test.go: shared scratch fixtures for the cutover guards
// (SPEC-GITHUB-FLOW-DEFAULT-001 M6: AC-GFD-018 rehearsal, AC-GFD-019 precheck,
// AC-GFD-020 protection comparison, AC-GFD-023 runbook shape).
//
// The scripts under test read git repositories and, through injectable commands,
// the moai queue/window/slot/session readers and the GitHub CLI. None of them may
// reach this repository's remotes, the real GitHub origin or the real moai state,
// so every test builds throwaway repositories under t.TempDir(), points every
// reader at a stub script, and puts a poisoned `moai` and `gh` first on PATH: a
// script that bypasses a seam trips the poison (exit 99 plus a log line), and the
// test asserts the poison log is empty.
//
// All helpers carry the `cvo` prefix: internal/template/ holds many sibling test
// files in one package. The release fixtures (rls*) are reused for process
// execution and hermetic git environments.
package template_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	cvoPrecheckRel   = "scripts/cutover-precheck.sh"
	cvoRehearsalRel  = "scripts/cutover-rehearsal.sh"
	cvoProtectionRel = "scripts/cutover-protection-compare.sh"
	cvoRunbookRel    = ".moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/cutover-runbook.md"
)

// cvoScript resolves a project-relative script path against the real repository.
func cvoScript(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rel))
}

// cvoPoison is a PATH directory whose `moai`, `gh`, `ssh` and `curl` fail loudly
// and log every call. A script that reaches a real reader or the network instead
// of its injected seam trips it.
type cvoPoison struct {
	bin string
	log string
}

func cvoNewPoison(t *testing.T) *cvoPoison {
	t.Helper()
	dir := t.TempDir()
	p := &cvoPoison{bin: filepath.Join(dir, "poison-bin"), log: filepath.Join(dir, "poison.log")}
	if err := os.MkdirAll(p.bin, 0o755); err != nil {
		t.Fatalf("mkdir poison bin: %v", err)
	}
	if err := os.WriteFile(p.log, nil, 0o644); err != nil {
		t.Fatalf("create poison log: %v", err)
	}
	body := "#!/bin/sh\necho \"$(basename \"$0\") $*\" >> '" + p.log + "'\necho 'poisoned tool called; the script bypassed its seam' >&2\nexit 99\n"
	for _, name := range []string{"moai", "gh", "ssh", "curl"} {
		rlsWriteScript(t, p.bin, name, body)
	}
	return p
}

// assertUntouched fails the test when any poisoned tool was called.
func (p *cvoPoison) assertUntouched(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile(p.log)
	if err != nil {
		t.Fatalf("read poison log: %v", err)
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("a real moai/gh/ssh/curl was called instead of the injected seam:\n%s", data)
	}
}

// env returns a hermetic environment with the poison first on PATH.
func (p *cvoPoison) env(extra ...string) []string {
	base := rlsEnv("PATH=" + p.bin + string(os.PathListSeparator) + os.Getenv("PATH"))
	return append(base, extra...)
}

// cvoStubCmd writes a stub reader: it prints body to stdout and exits with code.
// The body is stored in a side file so it needs no shell quoting.
func cvoStubCmd(t *testing.T, dir, name, body string, code int) string {
	t.Helper()
	data := rlsWriteScript(t, dir, name+".out", body)
	script := "#!/bin/sh\ncat '" + data + "'\nexit " + strconv.Itoa(code) + "\n"
	return rlsWriteScript(t, dir, name, script)
}
