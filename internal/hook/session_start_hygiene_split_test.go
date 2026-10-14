package hook

// SPEC-CAPACITY-RECLAIM-001 AC-CAP-012: the SessionStart auto path applies
// only the non-destructive hygiene categories. The fixture project's
// workflow.yaml IS the shipped template's file, so the pass's mode follows
// the template default — which is what couples this test's green path to the
// M1 flip (AC-CAP-008): against the pre-flip template (mode: report) the
// apply-row assertions fail, and after the flip they pass.
//
// Clause 1 — with the configured mode, the pass writes "mode":"apply" audit
// rows for the rotator unit and at least one "unit":"gc" / "mode":"apply"
// row for a seeded `verify` residue-class target (the GC writes apply rows
// only per decision and nothing when none fires, so the seeded target is
// what makes the cache/GC apply half observable).
// Clause 2 — no row, and no filesystem change, names a path outside the
// project's .moai/logs and .moai/state, and a fixture home seeded with old
// backups/removed-* directories and a non-current release is byte-identical
// after the pass: no destructive category, no backups or releases path.
// Clause 3 (added at M3) — the profile-cap advisory check fires in the pass.

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hygiene"
	"github.com/modu-ai/moai-adk/internal/template"
)

// splitTestSessionKey is the fixture session key: the 36-char hyphenated
// UUID shape the GC's sessionKeyShape matches.
const splitTestSessionKey = "01234567-89ab-cdef-0123-456789abcdef"

// writeSplitFixtureProject seeds a fixture project whose workflow.yaml is the
// shipped template's file and that carries one over-threshold registered sink
// and one seeded `verify` residue-class target.
func writeSplitFixtureProject(t *testing.T) string {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	workflow, err := fs.ReadFile(fsys, ".moai/config/sections/workflow.yaml")
	if err != nil {
		t.Fatalf("read embedded workflow.yaml: %v", err)
	}

	project := t.TempDir()
	sections := filepath.Join(project, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), workflow, 0o644); err != nil {
		t.Fatalf("write fixture workflow.yaml: %v", err)
	}

	// One registered sink at/above the template's audit_log_max_bytes
	// (10485760), so the rotator has a decision to make.
	logs := filepath.Join(project, ".moai", "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logs, "hook-runtime.log"), bytesOfLen(10485760+1), 0o644); err != nil {
		t.Fatalf("seed over-threshold sink: %v", err)
	}

	// One seeded `verify` scratch target: a session-keyed directory holding a
	// datable entry (recorded_at is the class's pinned JSON dating field).
	verifyDir := filepath.Join(project, ".moai", "state", "verify", splitTestSessionKey)
	if err := os.MkdirAll(verifyDir, 0o755); err != nil {
		t.Fatalf("mkdir verify target: %v", err)
	}
	entry := `{"recorded_at":"2020-01-01T00:00:00Z","note":"seeded verify scratch"}`
	if err := os.WriteFile(filepath.Join(verifyDir, "probe.json"), []byte(entry), 0o644); err != nil {
		t.Fatalf("seed verify entry: %v", err)
	}
	return project
}

// writeSplitFixtureHome seeds a fixture HOME whose .moai carries exactly the
// categories the auto path must never touch: backups/removed-* directories
// and a non-current release binary. It also seeds the transcript the
// liveness evaluator reads for the fixture key (old mtime — a negative
// signal), so the seeded verify target's verdict is measured, not unmeasured.
func writeSplitFixtureHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()

	transcripts := filepath.Join(home, ".claude", "projects")
	if err := os.MkdirAll(transcripts, 0o755); err != nil {
		t.Fatalf("mkdir transcripts: %v", err)
	}
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	transcript := filepath.Join(transcripts, splitTestSessionKey+".jsonl")
	if err := os.WriteFile(transcript, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("seed transcript: %v", err)
	}
	if err := os.Chtimes(transcript, old, old); err != nil {
		t.Fatalf("backdate transcript: %v", err)
	}

	moai := filepath.Join(home, ".moai")
	for _, dir := range []string{
		filepath.Join(moai, "backups", "removed-20260101T000000"),
		filepath.Join(moai, "releases"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir home fixture %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(moai, "backups", "removed-20260101T000000", "notes.txt"), []byte("backup payload"), 0o644); err != nil {
		t.Fatalf("seed backup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moai, "releases", "moai-v0.0.1"), []byte("old release binary"), 0o755); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	return home
}

// hashTree returns a digest over every file under root (rel path + content),
// for the byte-identical clause-2 comparison.
func hashTree(t *testing.T, root string) []byte {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if _, writeErr := io.WriteString(h, filepath.ToSlash(rel)+"\n"); writeErr != nil {
			return writeErr
		}
		if d.IsDir() {
			return nil
		}
		blob, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		_, _ = h.Write(blob)
		return nil
	})
	if err != nil {
		t.Fatalf("hash tree %s: %v", root, err)
	}
	return h.Sum(nil)
}

func bytesOfLen(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return b
}

type splitAuditRow struct {
	Unit    string `json:"unit"`
	Mode    string `json:"mode"`
	Outcome string `json:"outcome"`
	Path    string `json:"path"`
}

func readSplitAuditRows(t *testing.T, path string) []splitAuditRow {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log %s: %v", path, err)
	}
	var rows []splitAuditRow
	for _, line := range strings.Split(string(blob), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row splitAuditRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("parse audit row: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}

// captureSlog swaps the default logger for one writing to a buffer and
// returns the buffer plus a restore func. The calling test must be serial
// (no t.Parallel): slog's default is process-global.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

// TestSessionStartAutoPathSplit drives the SessionStart hygiene pass through
// the hygieneRunFn seam and asserts the three clauses of AC-CAP-012.
func TestSessionStartAutoPathSplit(t *testing.T) {
	project := writeSplitFixtureProject(t)
	home := writeSplitFixtureHome(t)

	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", "")
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	moaiRoot := filepath.Join(project, ".moai")
	hygiene.RegisterTestRoot(moaiRoot)

	homeBefore := hashTree(t, home)

	// Clause 3, fixture-2 arm — the home carries no state.profile_max_bytes
	// key, so the pass emits no profile-cap advisory (no compiled fallback).
	logs1 := captureSlog(t)
	if err := hygieneRunFn(project); err != nil {
		t.Fatalf("hygieneRunFn: %v", err)
	}
	if strings.Contains(logs1.String(), "profile_max_bytes") {
		t.Fatalf("absent-key home produced a profile-cap advisory: %s", logs1.String())
	}

	rows := readSplitAuditRows(t, filepath.Join(moaiRoot, "logs", "hygiene-audit.jsonl"))
	if len(rows) == 0 {
		t.Fatalf("the pass wrote no audit rows")
	}

	// Clause 1 — apply rows for both units.
	rotatorApply := 0
	for _, row := range rows {
		if row.Unit == "rotator" && row.Mode == "apply" {
			rotatorApply++
		}
	}
	if rotatorApply == 0 {
		t.Fatalf("no rotator row carries mode=apply (shipped template default did not reach the pass); rows=%+v", rows)
	}
	entryRel := "state/verify/" + splitTestSessionKey + "/probe.json"
	gcApply := 0
	for _, row := range rows {
		if row.Unit == "gc" && row.Mode == "apply" && filepath.ToSlash(row.Path) == entryRel {
			gcApply++
		}
	}
	if gcApply == 0 {
		t.Fatalf("no gc row carries mode=apply for the seeded verify target %s (the GC writes apply rows only per decision); rows=%+v", entryRel, rows)
	}

	// Clause 2 — no row names a path outside .moai/logs and .moai/state.
	// Apply-mode rows carry absolute sink paths (rotate.go), so absolute
	// paths are resolved against the fixture root; relative paths are
	// checked by their slash-normalized prefix.
	logsRoot := filepath.ToSlash(filepath.Join(moaiRoot, "logs")) + string(filepath.Separator)
	stateRoot := filepath.ToSlash(filepath.Join(moaiRoot, "state")) + string(filepath.Separator)
	for _, row := range rows {
		if row.Path == "" {
			continue
		}
		slashed := filepath.ToSlash(row.Path)
		if filepath.IsAbs(row.Path) {
			if !strings.HasPrefix(slashed, logsRoot) && !strings.HasPrefix(slashed, stateRoot) {
				t.Fatalf("audit row names a path outside .moai/logs and .moai/state: %+v", row)
			}
			continue
		}
		if !strings.HasPrefix(slashed, "logs/") && !strings.HasPrefix(slashed, "state/") {
			t.Fatalf("audit row names a path outside .moai/logs and .moai/state: %+v", row)
		}
	}
	if homeAfter := hashTree(t, home); string(homeAfter) != string(homeBefore) {
		t.Fatalf("the pass mutated the fixture home (backups/releases must stay untouched)")
	}

	// Clause 3, fixture-1 arm — a home whose state.yaml sets
	// state.profile_max_bytes: 1048576 and whose profile is above it: the
	// pass emits the warning naming the profile and the key, and deletes
	// nothing (the home is byte-identical after the pass).
	home2 := t.TempDir()
	sections := filepath.Join(home2, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir home2 sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "state.yaml"),
		[]byte("state:\n    profile_max_bytes: 1048576\n"), 0o644); err != nil {
		t.Fatalf("seed home2 state.yaml: %v", err)
	}
	bigProjects := filepath.Join(home2, ".moai", "claude-profiles", "p-big", "projects", "session-x")
	if err := os.MkdirAll(bigProjects, 0o755); err != nil {
		t.Fatalf("mkdir home2 profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bigProjects, "chunk.bin"), bytesOfLen(2*1024*1024), 0o644); err != nil {
		t.Fatalf("seed home2 profile payload: %v", err)
	}
	t.Setenv("HOME", home2)
	t.Setenv("MOAI_HOME", "")

	home2Before := hashTree(t, home2)
	logs2 := captureSlog(t)
	if err := hygieneRunFn(project); err != nil {
		t.Fatalf("hygieneRunFn (fixture-1 home): %v", err)
	}
	if !strings.Contains(logs2.String(), "profile_max_bytes") || !strings.Contains(logs2.String(), "p-big") {
		t.Fatalf("profile-cap advisory did not fire for the over-cap profile; logs=%s", logs2.String())
	}
	if home2After := hashTree(t, home2); string(home2After) != string(home2Before) {
		t.Fatalf("the pass mutated the over-cap fixture home (advisory must delete nothing)")
	}
}
