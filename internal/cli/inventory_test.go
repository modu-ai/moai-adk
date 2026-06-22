package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	wtroot "github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/core/git"
)

// ---------------------------------------------------------------------------
// M1 — Report data shapes (REQ-INV-004, REQ-INV-005)
// ---------------------------------------------------------------------------

// TestInventory_JSONShape marshals a UnifiedInventoryReport fixture and asserts
// the three top-level keys (sessions/worktrees/harnesses) plus the per-row key
// fields round-trip via json.MarshalIndent. (AC-INV-003)
func TestInventory_JSONShape(t *testing.T) {
	report := UnifiedInventoryReport{
		Sessions: SessionInventory{
			Count: 1,
			Entries: []SessionSummaryRow{
				{SessionID: "abcd1234", SpecID: "SPEC-FOO-001", Phase: "run"},
			},
		},
		Worktrees: WorktreeInventory{
			Count: 1,
			Entries: []WorktreeSummaryRow{
				{Branch: "feat/foo", Path: "/tmp/wt", HEAD: "deadbeef"},
			},
		},
		Harnesses: HarnessInventory{
			Count: 1,
			Entries: []HarnessSummaryRow{
				{Name: "release", Domain: "ci", ManifestMissing: false},
			},
		},
	}

	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Round-trip into a generic map and assert the three top-level keys.
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(out, &generic); err != nil {
		t.Fatalf("unmarshal generic: %v", err)
	}
	for _, key := range []string{"sessions", "worktrees", "harnesses"} {
		if _, ok := generic[key]; !ok {
			t.Errorf("top-level key %q missing from JSON output: %s", key, out)
		}
	}

	// Per-row key field presence (snake_case json tags).
	s := string(out)
	for _, field := range []string{
		`"session_id"`, `"spec_id"`, `"phase"`,
		`"branch"`, `"path"`, `"head"`,
		`"name"`, `"domain"`, `"manifest_missing"`,
		`"count"`, `"entries"`,
	} {
		if !strings.Contains(s, field) {
			t.Errorf("expected JSON to contain field %s; got:\n%s", field, s)
		}
	}

	// Round-trip back into the typed struct to assert structural fidelity.
	var rt UnifiedInventoryReport
	if err := json.Unmarshal(out, &rt); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if rt.Sessions.Count != 1 || rt.Worktrees.Count != 1 || rt.Harnesses.Count != 1 {
		t.Errorf("round-trip count mismatch: %+v", rt)
	}
	if rt.Sessions.Entries[0].SpecID != "SPEC-FOO-001" {
		t.Errorf("round-trip session spec_id mismatch: %q", rt.Sessions.Entries[0].SpecID)
	}
}

// TestInventory_ErrorFieldOmitEmpty asserts the per-surface Error field is
// omitted from JSON when empty (REQ-INV-008 degradation field is opt-in).
func TestInventory_ErrorFieldOmitEmpty(t *testing.T) {
	report := UnifiedInventoryReport{}
	out, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), `"error"`) {
		t.Errorf("empty Error field must be omitted (omitempty); got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// M2 — Surface collectors (REQ-INV-003, REQ-INV-007, REQ-INV-008)
// ---------------------------------------------------------------------------

// stubWorktreeProvider is a NON-NIL git.WorktreeManager whose List() result is
// fully controlled by the test. It lets AC-INV-005c inject a List() error
// (the real out-of-git-repo path) without touching internal/cli/worktree.
type stubWorktreeProvider struct {
	worktrees []git.Worktree
	listErr   error
}

func (s *stubWorktreeProvider) List() ([]git.Worktree, error) {
	return s.worktrees, s.listErr
}

// The remaining git.WorktreeManager methods are no-ops; the inventory command
// only ever calls List().
func (s *stubWorktreeProvider) Add(path, branch string) error               { return nil }
func (s *stubWorktreeProvider) Remove(path string, force bool) error        { return nil }
func (s *stubWorktreeProvider) Prune() error                                { return nil }
func (s *stubWorktreeProvider) Repair() error                               { return nil }
func (s *stubWorktreeProvider) Root() string                                { return "/stub" }
func (s *stubWorktreeProvider) Sync(wt, base, strategy string) error        { return nil }
func (s *stubWorktreeProvider) DeleteBranch(name string) error              { return nil }
func (s *stubWorktreeProvider) IsBranchMerged(b, base string) (bool, error) { return false, nil }

// withWorktreeProvider temporarily installs a worktree provider for the test
// by overriding the package-level worktree.WorktreeProvider var, restoring it
// on cleanup. Tests using it MUST NOT call t.Parallel() (mutates package state).
func withWorktreeProvider(t *testing.T, p git.WorktreeManager) {
	t.Helper()
	orig := wtroot.WorktreeProvider
	wtroot.WorktreeProvider = p
	t.Cleanup(func() { wtroot.WorktreeProvider = orig })
}

// TestInventory_ComposesThreeSurfaces verifies the collectors call the three
// existing exported functions and produce key-field summary rows. (AC-INV-002)
func TestInventory_ComposesThreeSurfaces(t *testing.T) {
	// Isolate cwd to a non-MoAI, git-free temp dir so sessions + harnesses
	// degrade-to-empty deterministically.
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Non-nil worktree provider returning two worktrees → exercises the mapping.
	withWorktreeProvider(t, &stubWorktreeProvider{
		worktrees: []git.Worktree{
			{Branch: "main", Path: "/repo", HEAD: "0123456789abcdef"},
			{Branch: "feat/x", Path: "/repo-wt", HEAD: "fedcba9876543210"},
		},
	})

	rep := collectInventory(tempDir)

	if rep.Sessions.Count != 0 {
		t.Errorf("sessions: want count 0 in git-free temp dir, got %d (err=%q)", rep.Sessions.Count, rep.Sessions.Error)
	}
	if rep.Harnesses.Count != 0 {
		t.Errorf("harnesses: want count 0 in harness-free temp dir, got %d (err=%q)", rep.Harnesses.Count, rep.Harnesses.Error)
	}
	if rep.Worktrees.Count != 2 {
		t.Fatalf("worktrees: want count 2, got %d (err=%q)", rep.Worktrees.Count, rep.Worktrees.Error)
	}
	// HEAD must be shortened to first 8 chars (REQ-INV-004 key fields).
	if got := rep.Worktrees.Entries[0].HEAD; got != "01234567" {
		t.Errorf("worktree HEAD not shortened: want 01234567, got %q", got)
	}
	if rep.Worktrees.Entries[0].Branch != "main" {
		t.Errorf("worktree branch mismatch: %q", rep.Worktrees.Entries[0].Branch)
	}
}

// TestInventory_EmptySurfacesGraceful verifies absent backing data yields
// count 0 (not an error) for sessions and harnesses. (AC-INV-005a, E1+E2)
func TestInventory_EmptySurfacesGraceful(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{worktrees: nil})

	rep := collectInventory(tempDir)

	if rep.Sessions.Count != 0 || rep.Sessions.Error != "" {
		t.Errorf("sessions: want count 0 no-error, got count=%d err=%q", rep.Sessions.Count, rep.Sessions.Error)
	}
	if rep.Harnesses.Count != 0 || rep.Harnesses.Error != "" {
		t.Errorf("harnesses: want count 0 no-error, got count=%d err=%q", rep.Harnesses.Count, rep.Harnesses.Error)
	}
	if rep.Worktrees.Count != 0 || rep.Worktrees.Error != "" {
		t.Errorf("worktrees: want count 0 no-error, got count=%d err=%q", rep.Worktrees.Count, rep.Worktrees.Error)
	}
}

// TestInventory_PerSurfaceErrorDegrades verifies a nil WorktreeProvider (git
// module unavailable) populates worktrees.error while other surfaces still
// render. (AC-INV-005b, E4)
func TestInventory_PerSurfaceErrorDegrades(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, nil) // nil provider — git module unavailable

	rep := collectInventory(tempDir)

	if rep.Worktrees.Error == "" {
		t.Errorf("worktrees.error must be set when provider is nil")
	}
	if rep.Worktrees.Count != 0 {
		t.Errorf("worktrees.count must be 0 when provider is nil, got %d", rep.Worktrees.Count)
	}
	// The other two surfaces still render (degrade-to-empty).
	if rep.Sessions.Error != "" {
		t.Errorf("sessions must still render (no error), got err=%q", rep.Sessions.Error)
	}
	if rep.Harnesses.Error != "" {
		t.Errorf("harnesses must still render (no error), got err=%q", rep.Harnesses.Error)
	}
}

// TestInventory_WorktreeListErrorDegrades verifies a NON-NIL provider whose
// List() returns an error (the real out-of-git-repo path) degrades only the
// worktree surface. (AC-INV-005c, E6)
func TestInventory_WorktreeListErrorDegrades(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{
		listErr: errors.New("fatal: not a git repository"),
	})

	rep := collectInventory(tempDir)

	if rep.Worktrees.Error == "" {
		t.Errorf("worktrees.error must be set when List() returns an error")
	}
	if !strings.Contains(rep.Worktrees.Error, "not a git repository") {
		t.Errorf("worktrees.error should carry the git error, got %q", rep.Worktrees.Error)
	}
	if rep.Worktrees.Count != 0 {
		t.Errorf("worktrees.count must be 0 on List() error, got %d", rep.Worktrees.Count)
	}
	// Sessions + harnesses degrade-to-empty (count 0, no error).
	if rep.Sessions.Count != 0 || rep.Sessions.Error != "" {
		t.Errorf("sessions: want count 0 no-error, got count=%d err=%q", rep.Sessions.Count, rep.Sessions.Error)
	}
	if rep.Harnesses.Count != 0 || rep.Harnesses.Error != "" {
		t.Errorf("harnesses: want count 0 no-error, got count=%d err=%q", rep.Harnesses.Count, rep.Harnesses.Error)
	}
}

// ---------------------------------------------------------------------------
// M3 — Command factory + registration (REQ-INV-001, REQ-INV-002, REQ-INV-009)
// ---------------------------------------------------------------------------

// TestNewInventory_CommandShape verifies newInventoryCmd builds a single
// top-level command named "inventory" with NoArgs and a --json flag.
// (AC-INV-001 — REQ-INV-001, REQ-INV-002)
func TestNewInventory_CommandShape(t *testing.T) {
	cmd := newInventoryCmd()
	if cmd.Use != "inventory" {
		t.Errorf("Use: want %q, got %q", "inventory", cmd.Use)
	}
	if cmd.GroupID != "tools" {
		t.Errorf("GroupID: want %q, got %q", "tools", cmd.GroupID)
	}
	if !cmd.HasSubCommands() {
		// inventory is NOT a subcommand group — it must have zero subcommands.
	} else {
		t.Errorf("inventory must be a single command, not a subcommand group")
	}
	if flag := cmd.Flags().Lookup("json"); flag == nil {
		t.Error("--json flag is not registered")
	} else if flag.Value.Type() != "bool" {
		t.Errorf("--json flag type: want bool, got %s", flag.Value.Type())
	}
	if flag := cmd.Flags().Lookup("project-root"); flag == nil {
		t.Error("--project-root flag is not registered")
	}
}

// TestInventory_RegisteredOnRoot verifies the command is wired onto the root
// command tree via rootCmd.AddCommand (AC-INV-001).
func TestInventory_RegisteredOnRoot(t *testing.T) {
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Name() == "inventory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("inventory command is not registered on rootCmd")
	}
}

// TestInventory_ProjectRootResolution verifies the command resolves the project
// root via the resolveProjectRoot helper (--project-root flag honored).
// (AC-INV-006a — REQ-INV-009)
func TestInventory_ProjectRootResolution(t *testing.T) {
	// A temp dir with a .claude/commands/harness/ directory holding one harness
	// command file → ListHarnesses(projectRoot) must find it when --project-root
	// points there, proving the resolved root is used.
	tempDir := t.TempDir()
	harnessDir := filepath.Join(tempDir, ".claude", "commands", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(harnessDir, "release.md"), []byte("# release\n"), 0o644); err != nil {
		t.Fatalf("write harness cmd: %v", err)
	}

	// cwd is a DIFFERENT, harness-free dir — proving --project-root (not cwd)
	// is what resolveProjectRoot returns.
	otherDir := t.TempDir()
	t.Chdir(otherDir)
	withWorktreeProvider(t, &stubWorktreeProvider{worktrees: nil})

	cmd := newInventoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--json", "--project-root", tempDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var rep UnifiedInventoryReport
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatalf("unmarshal output: %v\nraw: %s", err, out.String())
	}
	if rep.Harnesses.Count != 1 {
		t.Errorf("harnesses.count: want 1 (from --project-root dir), got %d", rep.Harnesses.Count)
	}
	if rep.Harnesses.Entries[0].Name != "release" {
		t.Errorf("harness name: want release, got %q", rep.Harnesses.Entries[0].Name)
	}
}

// ---------------------------------------------------------------------------
// M4 — Output rendering (REQ-INV-005, REQ-INV-006, REQ-INV-010)
// ---------------------------------------------------------------------------

// runInventoryCmd is a small helper that builds + executes the inventory
// command with the given args, returning captured stdout and the error.
func runInventoryCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newInventoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// TestInventory_HumanReadableSummary verifies the default (non-JSON) output
// contains all three surface labels with counts, rendered via render utilities.
// (AC-INV-004 — REQ-INV-006)
func TestInventory_HumanReadableSummary(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "main", Path: "/repo", HEAD: "abcdef0123"}},
	})

	out, err := runInventoryCmd(t, "--project-root", tempDir)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, label := range []string{"Sessions", "Worktrees", "Harnesses"} {
		if !strings.Contains(out, label) {
			t.Errorf("human-readable output missing surface label %q:\n%s", label, out)
		}
	}
	// Counts are present in the card headers.
	if !strings.Contains(out, "Sessions (0)") {
		t.Errorf("expected 'Sessions (0)' header in output:\n%s", out)
	}
	if !strings.Contains(out, "Worktrees (1)") {
		t.Errorf("expected 'Worktrees (1)' header in output:\n%s", out)
	}
	// The worktree branch key field should appear.
	if !strings.Contains(out, "main") {
		t.Errorf("expected worktree branch 'main' in output:\n%s", out)
	}
}

// TestInventory_JSONOutputUnmarshals verifies --json stdout unmarshals into the
// UnifiedInventoryReport with the expected counts. (AC-INV-003 — REQ-INV-005)
func TestInventory_JSONOutputUnmarshals(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{
		worktrees: []git.Worktree{
			{Branch: "main", Path: "/repo", HEAD: "abcdef0123"},
			{Branch: "feat/y", Path: "/repo-wt", HEAD: "9876543210"},
		},
	})

	out, err := runInventoryCmd(t, "--json", "--project-root", tempDir)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var rep UnifiedInventoryReport
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, out)
	}
	if rep.Worktrees.Count != 2 {
		t.Errorf("worktrees.count: want 2, got %d", rep.Worktrees.Count)
	}
	if rep.Sessions.Count != 0 || rep.Harnesses.Count != 0 {
		t.Errorf("sessions/harnesses should be 0 in temp dir: %+v", rep)
	}
}

// TestInventory_ConsistentCountsAcrossModes verifies --json and the default
// human-readable mode report identical surface counts (E7). (AC-INV-003 + AC-INV-004)
func TestInventory_ConsistentCountsAcrossModes(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{
		worktrees: []git.Worktree{
			{Branch: "main", Path: "/repo", HEAD: "abcdef0123"},
			{Branch: "feat/z", Path: "/repo-wt", HEAD: "1111222233"},
		},
	})

	jsonOut, err := runInventoryCmd(t, "--json", "--project-root", tempDir)
	if err != nil {
		t.Fatalf("json Execute: %v", err)
	}
	var rep UnifiedInventoryReport
	if err := json.Unmarshal([]byte(jsonOut), &rep); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	textOut, err := runInventoryCmd(t, "--project-root", tempDir)
	if err != nil {
		t.Fatalf("text Execute: %v", err)
	}

	// The text mode must report the same worktree count in its card header.
	wantHeader := "Worktrees (" + itoa(rep.Worktrees.Count) + ")"
	if !strings.Contains(textOut, wantHeader) {
		t.Errorf("text mode count %q not consistent with JSON count %d:\n%s",
			wantHeader, rep.Worktrees.Count, textOut)
	}
}

// TestInventory_OutOfProjectExitsZero verifies the out-of-project invocation
// (REQ-INV-010): sessions/harnesses degrade-to-empty, the worktree surface
// errors (List() returns "fatal: not a git repository"), and the command still
// exits 0 because ≥1 surface rendered. (E6 — AC-INV-005a + AC-INV-005c)
func TestInventory_OutOfProjectExitsZero(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	withWorktreeProvider(t, &stubWorktreeProvider{
		listErr: errors.New("fatal: not a git repository (or any of the parent directories): .git"),
	})

	out, err := runInventoryCmd(t, "--json", "--project-root", tempDir)
	if err != nil {
		t.Fatalf("out-of-project invocation must exit 0 (>=1 surface rendered), got err: %v", err)
	}
	var rep UnifiedInventoryReport
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, out)
	}
	if rep.Sessions.Count != 0 || rep.Harnesses.Count != 0 {
		t.Errorf("sessions/harnesses must be 0-count out-of-project: %+v", rep)
	}
	if rep.Worktrees.Error == "" {
		t.Errorf("worktrees.error must be populated out-of-project")
	}
}

// itoa is a tiny strconv.Itoa substitute kept local to avoid widening imports.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// ---------------------------------------------------------------------------
// M5 — Boundary guard + invariant verification (REQ-INV-008, REQ-INV-013)
// ---------------------------------------------------------------------------

// TestNewInventory_NoAskUserQuestion: static check — internal/cli/inventory.go
// MUST NOT import or invoke AskUserQuestion. The orchestrator owns user
// interaction; the CLI returns exit codes / JSON / text only
// (agent-common-protocol C-HRA-008). (AC-INV-007 — REQ-INV-013)
//
// This test does NOT t.Chdir — it reads the source file relative to the test's
// package directory (internal/cli/), mirroring worktree/new_test.go.
func TestNewInventory_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("inventory.go")
	if err != nil {
		t.Fatalf("read inventory.go: %v", err)
	}
	if strings.Contains(string(src), "AskUserQuestion") {
		t.Error("internal/cli/inventory.go must NOT reference AskUserQuestion (orchestrator-only HARD, C-HRA-008)")
	}
	if strings.Contains(string(src), "mcp__askuser") {
		t.Error("internal/cli/inventory.go must NOT reference mcp__askuser (orchestrator-only HARD, C-HRA-008)")
	}
}

// TestInventory_AllSurfacesFailExitsNonZero verifies the exit-code discipline
// (REQ-INV-008): when NO surface could render (all three error), the command
// returns a non-nil error so cobra exits non-zero. This is the only condition
// that aborts the whole command.
func TestInventory_AllSurfacesFailExitsNonZero(t *testing.T) {
	rep := UnifiedInventoryReport{
		Sessions:  SessionInventory{Error: "session read failed"},
		Worktrees: WorktreeInventory{Error: "worktree provider unavailable"},
		Harnesses: HarnessInventory{Error: "harness read failed"},
	}
	if got := renderedSurfaceCount(rep); got != 0 {
		t.Errorf("renderedSurfaceCount: want 0 when all surfaces error, got %d", got)
	}

	rep2 := UnifiedInventoryReport{
		Sessions:  SessionInventory{Count: 0}, // rendered (no error)
		Worktrees: WorktreeInventory{Error: "worktree provider unavailable"},
		Harnesses: HarnessInventory{Error: "harness read failed"},
	}
	if got := renderedSurfaceCount(rep2); got != 1 {
		t.Errorf("renderedSurfaceCount: want 1 when sessions rendered, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Coverage — populated card bodies + error-branch rendering
// ---------------------------------------------------------------------------

// TestInventory_RenderTextPopulated exercises renderInventoryText with a fully
// populated report (sessions + worktrees + harnesses with rows, one harness
// with manifest_missing), covering the non-empty card-body builders and the
// error-branch of surfaceErrorOrEmpty. (REQ-INV-004, REQ-INV-006, REQ-INV-008)
func TestInventory_RenderTextPopulated(t *testing.T) {
	report := UnifiedInventoryReport{
		Sessions: SessionInventory{
			Count: 2,
			Entries: []SessionSummaryRow{
				{SessionID: "abc12345", SpecID: "SPEC-A-001", Phase: "run"},
				{SessionID: "def67890", SpecID: "SPEC-B-002", Phase: "sync"},
			},
		},
		Worktrees: WorktreeInventory{
			// Worktree surface in genuine error → exercises the error branch.
			Error: "fatal: not a git repository",
		},
		Harnesses: HarnessInventory{
			Count: 2,
			Entries: []HarnessSummaryRow{
				{Name: "release", Domain: "ci", ManifestMissing: false},
				{Name: "broken", Domain: "", ManifestMissing: true},
			},
		},
	}

	cmd := newInventoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := renderInventoryText(cmd, report); err != nil {
		t.Fatalf("renderInventoryText: %v", err)
	}
	s := out.String()

	// Session rows rendered with their key fields.
	for _, want := range []string{"abc12345", "SPEC-A-001", "run", "def67890", "SPEC-B-002"} {
		if !strings.Contains(s, want) {
			t.Errorf("session field %q missing from rendered text:\n%s", want, s)
		}
	}
	// Worktree surface rendered its error.
	if !strings.Contains(s, "not a git repository") {
		t.Errorf("worktree error not rendered:\n%s", s)
	}
	// Harness rows rendered, including the manifest_missing marker.
	for _, want := range []string{"release", "domain=ci", "broken", "manifest_missing=true"} {
		if !strings.Contains(s, want) {
			t.Errorf("harness field %q missing from rendered text:\n%s", want, s)
		}
	}
	// Counts in the card headers.
	if !strings.Contains(s, "Sessions (2)") || !strings.Contains(s, "Harnesses (2)") {
		t.Errorf("expected populated counts in headers:\n%s", s)
	}
}

// TestInventory_CollectSessionsRows exercises the real collectSessions path by
// writing a fixture active-sessions registry in a temp cwd, covering the
// non-empty session collector branch (shortID row mapping). (REQ-INV-003,
// REQ-INV-004, REQ-INV-007)
func TestInventory_CollectSessionsRows(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	stateDir := filepath.Join(tempDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// QueryActiveWork reads .moai/state/active-sessions.json relative to cwd.
	fixture := `[{"session_id":"0123456789abcdef0123456789abcdef","spec_id":"SPEC-X-001","phase":"run","started_at":"2026-06-23T00:00:00Z","last_heartbeat":"2026-06-23T00:00:00Z","pid":123,"host":"h","cwd":"/x"}]`
	if err := os.WriteFile(filepath.Join(stateDir, "active-sessions.json"), []byte(fixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	inv := collectSessions()
	if inv.Error != "" {
		t.Fatalf("collectSessions error: %q", inv.Error)
	}
	if inv.Count != 1 {
		t.Fatalf("sessions count: want 1, got %d", inv.Count)
	}
	// session_id must be shortened to 8 chars.
	if inv.Entries[0].SessionID != "01234567" {
		t.Errorf("session_id not shortened to 8 chars: %q", inv.Entries[0].SessionID)
	}
	if inv.Entries[0].SpecID != "SPEC-X-001" || inv.Entries[0].Phase != "run" {
		t.Errorf("session row fields mismatch: %+v", inv.Entries[0])
	}
}

// TestInventory_CollectHarnessesRows exercises the real collectHarnesses path
// with two harness command files (one manifest-missing), covering the
// non-empty harness collector branch. (REQ-INV-003, REQ-INV-004)
func TestInventory_CollectHarnessesRows(t *testing.T) {
	tempDir := t.TempDir()
	harnessDir := filepath.Join(tempDir, ".claude", "commands", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(harnessDir, "release.md"), []byte("# release\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(harnessDir, "broken.md"), []byte("# broken\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	inv := collectHarnesses(tempDir)
	if inv.Error != "" {
		t.Fatalf("collectHarnesses error: %q", inv.Error)
	}
	if inv.Count != 2 {
		t.Fatalf("harness count: want 2, got %d", inv.Count)
	}
	// Both harnesses have no manifest.json → ManifestMissing true; rows are
	// sorted by name, so "broken" precedes "release".
	if inv.Entries[0].Name != "broken" || !inv.Entries[0].ManifestMissing {
		t.Errorf("first harness row mismatch: %+v", inv.Entries[0])
	}
}

// TestInventory_EnsureWorktreeProviderShortCircuits verifies
// ensureWorktreeProvider leaves an already-set provider unchanged (the test
// stub installed via withWorktreeProvider must survive the call). (REQ-INV-011)
func TestInventory_EnsureWorktreeProviderShortCircuits(t *testing.T) {
	stub := &stubWorktreeProvider{worktrees: []git.Worktree{{Branch: "main", Path: "/x", HEAD: "abc"}}}
	withWorktreeProvider(t, stub)

	ensureWorktreeProvider(t.TempDir())

	if wtroot.WorktreeProvider != git.WorktreeManager(stub) {
		t.Error("ensureWorktreeProvider must NOT overwrite an already-set provider")
	}
}

// TestInventory_EnsureWorktreeProviderNonGitRepo verifies that when the project
// root is NOT a git repository, ensureWorktreeProvider leaves the provider nil
// (EnsureGit fails) so the worktree surface degrades. (REQ-INV-008, REQ-INV-010)
func TestInventory_EnsureWorktreeProviderNonGitRepo(t *testing.T) {
	withWorktreeProvider(t, nil) // start from nil
	// deps may be nil in the unit-test process (InitDependencies not called);
	// either way the provider must remain nil for a non-git temp dir.
	ensureWorktreeProvider(t.TempDir())
	if wtroot.WorktreeProvider != nil {
		t.Errorf("ensureWorktreeProvider must leave provider nil for a non-git dir, got %v", wtroot.WorktreeProvider)
	}
}

var _ = os.Getenv
var _ = bytes.NewBuffer
