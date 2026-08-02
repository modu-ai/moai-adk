package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- helpers (SPEC-PROFILE-MEMORY-001 REQ-PM-021: every test sandboxes the
// profile base via BaseDirOverride + t.TempDir(); the real home is never read
// or written) ---

func pmSandboxBase(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := BaseDirOverride
	BaseDirOverride = dir
	t.Cleanup(func() { BaseDirOverride = orig })
	return dir
}

func pmMkProfile(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create profile dir %q: %v", name, err)
	}
	return dir
}

func pmWriteLedger(t *testing.T, base, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(base, "launch.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ledger: %v", err)
	}
}

func pmReadLedger(t *testing.T, base string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, "launch.yaml"))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	m := map[string]any{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal ledger: %v", err)
	}
	return m
}

func pmProjectsEntry(t *testing.T, ledger map[string]any, root string) (string, bool) {
	t.Helper()
	raw, ok := ledger["projects"]
	if !ok {
		return "", false
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("projects key is %T, want map[string]any", raw)
	}
	v, ok := m[normalizeProjectKey(root)]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("projects entry is %T, want string", v)
	}
	return s, true
}

// --- AC-PM-001 (REQ-PM-001, 002, 005 legacy-key clause) ---

func TestRecordForProject_PreservesLegacyKeys(t *testing.T) {
	base := pmSandboxBase(t)
	root := t.TempDir()
	pmMkProfile(t, base, "new")
	pmWriteLedger(t, base, "bypass: true\nmodel: claude-opus-4-6\nlast_profile: old\n")

	if err := RecordLastUsedProfileForProject(root, "new"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject: %v", err)
	}

	ledger := pmReadLedger(t, base)
	if got := ledger["bypass"]; got != true {
		t.Errorf("bypass = %v, want true (legacy key must survive read-modify-write)", got)
	}
	if got := ledger["model"]; got != "claude-opus-4-6" {
		t.Errorf("model = %v, want claude-opus-4-6", got)
	}
	if got := ledger["last_profile"]; got != "new" {
		t.Errorf("last_profile = %v, want new", got)
	}
	got, ok := pmProjectsEntry(t, ledger, root)
	if !ok {
		t.Fatalf("projects entry for %q missing; ledger = %#v", root, ledger)
	}
	if got != "new" {
		t.Errorf("projects[%q] = %q, want new", root, got)
	}
}

// --- AC-PM-002 (REQ-PM-003, function layer) ---

// TestResolveForProject_ProjectScopedIsolated pins the project-scoped-only
// resolution contract: project A resolves via its own projects[] entry, and a
// DIFFERENT project B (not in the map) resolves to "" (default) — the global
// last_profile key no longer participates in resolution, so it cannot bleed
// across projects. (This is the function-layer twin of the user's
// cross-project-bleed bug; TestResolveLaunchProfile_NoGlobalBleedAcrossProjects
// is the named regression guard.)
func TestResolveForProject_ProjectScopedIsolated(t *testing.T) {
	base := pmSandboxBase(t)
	projA := t.TempDir()
	projB := t.TempDir()
	pmMkProfile(t, base, "proj-one")
	pmMkProfile(t, base, "global-one")
	pmWriteLedger(t, base, "last_profile: global-one\nprojects:\n  "+normalizeProjectKey(projA)+": proj-one\n")

	if got := ResolveLaunchProfileForProject(projA, ""); got != "proj-one" {
		t.Errorf("ResolveLaunchProfileForProject(projA, \"\") = %q, want proj-one", got)
	}
	if got := ResolveLaunchProfileForProject(projB, ""); got != "" {
		t.Errorf("ResolveLaunchProfileForProject(projB, \"\") = %q, want \"\" (global last_profile must not bleed into a project with no projects[] entry)", got)
	}
}

// --- AC-PM-003 (REQ-PM-004) ---

// TestResolveForProject_LegacyLedgerUnchanged verifies that resolving from a
// legacy-shape ledger (last_profile only, no projects map) does not mutate the
// file — the read side is read-only. The resolution itself returns "" now
// (global last_profile is write-only on this binary; root has no projects[]
// entry), but the no-mutation contract is the actual point of this test.
func TestResolveForProject_LegacyLedgerUnchanged(t *testing.T) {
	base := pmSandboxBase(t)
	root := t.TempDir()
	pmMkProfile(t, base, "legacy")
	pmWriteLedger(t, base, "last_profile: legacy\n")

	ledgerPath := filepath.Join(base, "launch.yaml")
	before, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}

	if got := ResolveLaunchProfileForProject(root, ""); got != "" {
		t.Errorf("ResolveLaunchProfileForProject = %q, want \"\" (global last_profile no longer read)", got)
	}
	if got := ResolveLaunchProfile(""); got != "" {
		t.Errorf("ResolveLaunchProfile = %q, want \"\" (global last_profile no longer read)", got)
	}

	after, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("re-read ledger: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("resolution mutated the ledger:\nbefore=%q\nafter =%q", before, after)
	}
}

// --- AC-PM-004 (REQ-PM-006) ---

func TestResolveForProject_OptOutDisablesBothLookups(t *testing.T) {
	base := pmSandboxBase(t)
	projA := t.TempDir()
	pmMkProfile(t, base, "proj-one")
	pmMkProfile(t, base, "global-one")
	pmWriteLedger(t, base, "last_profile: global-one\nprojects:\n  "+normalizeProjectKey(projA)+": proj-one\n")

	t.Setenv("MOAI_NO_PROFILE_FALLBACK", "1")

	if got := ResolveLaunchProfileForProject(projA, ""); got != "" {
		t.Errorf("with opt-out set, ResolveLaunchProfileForProject = %q, want \"\"", got)
	}
	if got := ResolveLaunchProfile(""); got != "" {
		t.Errorf("with opt-out set, ResolveLaunchProfile = %q, want \"\"", got)
	}
}

// --- AC-PM-005 (REQ-PM-007) ---

// TestForProject_EmptyRootResolvesDefault verifies that an empty projectRoot
// skips the project-scoped lookup and resolves to "" (default). The global
// last_profile key is no longer read, so even when it holds a USABLE named
// profile ("global-one") an empty root does not bleed it in. The recording
// half below is unchanged: RecordLastUsedProfileForProject("", ...) still
// writes last_profile (legacy/downgrade compat) and leaves the projects map
// untouched.
func TestForProject_EmptyRootResolvesDefault(t *testing.T) {
	base := pmSandboxBase(t)
	projA := t.TempDir()
	pmMkProfile(t, base, "proj-one")
	pmMkProfile(t, base, "global-one")
	pmWriteLedger(t, base, "last_profile: global-one\nprojects:\n  "+normalizeProjectKey(projA)+": proj-one\n")
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	if got := ResolveLaunchProfileForProject("", ""); got != "" {
		t.Errorf("ResolveLaunchProfileForProject(\"\", \"\") = %q, want \"\" (empty root + no global read → default)", got)
	}
	if got := GetCurrentNameForProject(""); got != "default" {
		t.Errorf("GetCurrentNameForProject(\"\") = %q, want %q", got, "default")
	}

	// Recording with an empty root writes last_profile only (legacy compat);
	// the projects map is left untouched.
	if err := RecordLastUsedProfileForProject("", "proj-one"); err != nil {
		t.Fatalf("RecordLastUsedProfileForProject(\"\", ...) = %v, want nil", err)
	}
	ledger := pmReadLedger(t, base)
	if got := ledger["last_profile"]; got != "proj-one" {
		t.Errorf("last_profile = %v, want proj-one", got)
	}
	got, ok := pmProjectsEntry(t, ledger, projA)
	if !ok || got != "proj-one" {
		t.Errorf("projects[projA] = %q (present=%v); the pre-existing entry must be unchanged", got, ok)
	}
	projects, _ := ledger["projects"].(map[string]any)
	if len(projects) != 1 {
		t.Errorf("projects map has %d entries, want 1 — an empty root must not add a key", len(projects))
	}
}

// --- AC-PM-006 (REQ-PM-008) ---

// TestResolveForProject_StaleProjectEntrySkipped verifies that a projects[]
// entry whose target directory is gone is skipped, and resolution falls
// through to "" (default). The global last_profile key is no longer read,
// so the fallthrough lands at default rather than the (formerly global)
// "alive" entry — the stale-skip behavior itself is unchanged.
func TestResolveForProject_StaleProjectEntrySkipped(t *testing.T) {
	base := pmSandboxBase(t)
	projA := t.TempDir()
	pmMkProfile(t, base, "alive")
	pmWriteLedger(t, base, "last_profile: alive\nprojects:\n  "+normalizeProjectKey(projA)+": ghost\n")

	if got := ResolveLaunchProfileForProject(projA, ""); got != "" {
		t.Errorf("ResolveLaunchProfileForProject = %q, want \"\" (stale project entry must fall through to default)", got)
	}
}

// --- AC-PM-007 (REQ-PM-009) ---

func TestProjectKey_NormalizationSymmetric(t *testing.T) {
	base := pmSandboxBase(t)
	// On macOS t.TempDir() hands back a /var/folders/... path that is a symlink
	// to /private/var/folders/... — the exact asymmetry this test pins.
	root := t.TempDir()
	pmMkProfile(t, base, "sym")

	if err := RecordLastUsedProfileForProject(root, "sym"); err != nil {
		t.Fatalf("record: %v", err)
	}

	if got := ResolveLaunchProfileForProject(root, ""); got != "sym" {
		t.Errorf("resolve with the recording path = %q, want sym", got)
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Skipf("EvalSymlinks(%q): %v — symmetry across the resolved form is unverifiable here", root, err)
	}
	if got := ResolveLaunchProfileForProject(resolvedRoot, ""); got != "sym" {
		t.Errorf("resolve with the symlink-resolved path %q = %q, want sym", resolvedRoot, got)
	}
}

// --- AC-PM-008 (REQ-PM-011) ---

func TestRecordForProject_RejectsMissingDirectory(t *testing.T) {
	base := pmSandboxBase(t)
	root := t.TempDir()
	ledgerPath := filepath.Join(base, "launch.yaml")

	if _, err := os.Stat(ledgerPath); !os.IsNotExist(err) {
		t.Fatalf("precondition: ledger must be absent, stat err = %v", err)
	}

	err := RecordLastUsedProfileForProject(root, "nope")
	if err == nil {
		t.Fatal("RecordLastUsedProfileForProject with a missing profile directory returned nil, want an error")
	}

	if _, statErr := os.Stat(ledgerPath); !os.IsNotExist(statErr) {
		t.Errorf("refused record still created the ledger (stat err = %v)", statErr)
	}
}

// --- AC-PM-011 predicate half (REQ-PM-016, 017, 018) ---

func TestHasClaudeConfig_DecidesOnClaudeJSONAlone(t *testing.T) {
	base := pmSandboxBase(t)
	dirA := pmMkProfile(t, base, "fresh")
	dirB := pmMkProfile(t, base, "populated")

	if err := os.WriteFile(filepath.Join(dirB, ".claude.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("seed populated profile: %v", err)
	}

	if HasClaudeConfig("fresh") {
		t.Error("HasClaudeConfig(fresh) = true, want false")
	}
	if !HasClaudeConfig("populated") {
		t.Error("HasClaudeConfig(populated) = false, want true")
	}

	// A credential file alone must NOT flip the predicate — the decision is
	// the presence of .claude.json and nothing else (REQ-PM-018).
	if err := os.WriteFile(filepath.Join(dirA, ".credentials.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("seed credentials: %v", err)
	}
	if HasClaudeConfig("fresh") {
		t.Error("HasClaudeConfig(fresh) = true after adding .credentials.json only, want false")
	}

	// Adding .claude.json alone flips it.
	if err := os.WriteFile(filepath.Join(dirA, ".claude.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("seed claude.json: %v", err)
	}
	if !HasClaudeConfig("fresh") {
		t.Error("HasClaudeConfig(fresh) = false after adding .claude.json, want true")
	}
}

// --- AC-PM-019 (REQ-PM-024) ---

// TestGetCurrentNameForProject_ProjectScoped verifies that the project-scoped
// read names the project's own recorded profile, while the project-less
// wrapper (GetCurrentName, empty root) resolves to "default" — the global
// last_profile key no longer participates, so it cannot supply a value for a
// caller that passes no project root.
func TestGetCurrentNameForProject_ProjectScoped(t *testing.T) {
	base := pmSandboxBase(t)
	projA := t.TempDir()
	pmMkProfile(t, base, "proj-one")
	pmMkProfile(t, base, "global-one")
	pmWriteLedger(t, base, "last_profile: global-one\nprojects:\n  "+normalizeProjectKey(projA)+": proj-one\n")
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	if got := GetCurrentNameForProject(projA); got != "proj-one" {
		t.Errorf("GetCurrentNameForProject(projA) = %q, want proj-one", got)
	}
	if got := GetCurrentName(); got != "default" {
		t.Errorf("GetCurrentName() = %q, want %q (empty-root wrapper no longer reads the global key)", got, "default")
	}
}

// --- AC-PM-020 (REQ-PM-005 atomicity clause) ---

// TestRecordForProject_NoPartialStateOnFailure induces a failure at the
// os.Rename step — AFTER os.CreateTemp — by making the ledger path a non-empty
// directory. A guard that rejects before CreateTemp (e.g. the missing-directory
// guard) would prove nothing about atomicity, so the recipe is fixed here.
func TestRecordForProject_NoPartialStateOnFailure(t *testing.T) {
	base := pmSandboxBase(t)
	root := t.TempDir()
	pmMkProfile(t, base, "ok")

	// Ledger path is a NON-EMPTY directory: os.CreateTemp still succeeds, and
	// os.Rename onto it fails with ENOTEMPTY/EISDIR.
	ledgerPath := filepath.Join(base, "launch.yaml")
	if err := os.MkdirAll(filepath.Join(ledgerPath, "occupant"), 0o755); err != nil {
		t.Fatalf("stage ledger-as-directory: %v", err)
	}

	err := RecordLastUsedProfileForProject(root, "ok")
	if err == nil {
		t.Fatal("record onto a directory ledger path returned nil, want an error")
	}
	if strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("failure came from the directory-existence guard (%v); this test must fail at os.Rename, after os.CreateTemp", err)
	}

	// (c) the deferred os.Remove reclaimed the temp file.
	entries, readErr := os.ReadDir(base)
	if readErr != nil {
		t.Fatalf("read base: %v", readErr)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".launch-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp residue left behind: %s", e.Name())
		}
	}

	// (d) the ledger path's directory content is unchanged.
	inner, readErr := os.ReadDir(ledgerPath)
	if readErr != nil {
		t.Fatalf("read ledger dir: %v", readErr)
	}
	if len(inner) != 1 || inner[0].Name() != "occupant" {
		t.Errorf("ledger directory content changed: %v", inner)
	}
}
