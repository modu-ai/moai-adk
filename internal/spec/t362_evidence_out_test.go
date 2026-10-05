package spec

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// Output-location resolution for the t362 measurement harnesses
// (SPEC-HARNESS-EVIDENCE-WRITE-001 REQ-007).
//
// The default is the per-run t.TempDir() directory, so a harness re-run can
// never touch the repository working tree. An operator who wants a durable
// capture sets MOAI_T362_EVIDENCE_OUT=<dir> BEFORE the run — t.TempDir()
// directories are deleted when the test completes, so a post-run copy from
// the t.Logf-announced path is impossible and the destination must be
// selected up front. The override is no-clobber: an existing target file
// fails the run loudly and nothing is written. Durable capture always goes
// to a NEW file, because a silent overwrite would recreate the exact defect
// this SPEC exists to kill.
const t362EvidenceOutEnv = "MOAI_T362_EVIDENCE_OUT"

// t362ReportPath resolves the output file path for a t362 harness report.
// filename is a bare name (e.g. "m1-corpus-measurement.txt"), never a path.
func t362ReportPath(t *testing.T, filename string) string {
	t.Helper()
	dir := os.Getenv(t362EvidenceOutEnv)
	if dir == "" {
		return filepath.Join(t.TempDir(), filename)
	}
	target := filepath.Join(dir, filename)
	if _, err := os.Stat(target); err == nil {
		t.Fatalf("%s no-clobber: target %q already exists; refusing to overwrite — durable capture must write a NEW file (rename or remove the old report first)", t362EvidenceOutEnv, target)
	} else if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("stat %s: %v", target, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return target
}
