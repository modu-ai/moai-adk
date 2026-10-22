package toolpolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestMain dispatches the two-process record-lock child helper (sync-audit
// F2' regression). When MOAI_TOOLPOLICY_LOCK_CHILD is set the test binary was
// re-exec'd as the child side of the inter-process lock test: it never runs
// the suite, performs its helper role, and exits directly.
func TestMain(m *testing.M) {
	if os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD") == "guard-hold" {
		runRecordGuardChildHelper()
	}
	if os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD") != "" {
		runRecordLockChildHelper()
	}
	os.Exit(m.Run())
}

// testRecordLockPath derives the inter-process lockfile path for a project
// settings path — beside the sidecar record (the same file the production
// lock guards; the test writes it directly so the RED run compiles against a
// tree where the production lock does not exist yet).
func testRecordLockPath(t *testing.T, projectSettingsPath string) string {
	t.Helper()
	recordPath := managedAllowRecordPath(projectSettingsPath)
	return filepath.Join(filepath.Dir(recordPath), "managed-allow.lock")
}

// seedDeadPidLock spawns a short-lived child process, waits for it (so its pid
// is provably dead), and seeds a lockfile naming that pid — the takeover
// test's input.
func seedDeadPidLock(t *testing.T, lockPath string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "MOAI_TOOLPOLICY_LOCK_CHILD=exit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("spawn dead-pid source: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("pid=%d created=%s\n", cmd.Process.Pid, time.Now().UTC().Format(time.RFC3339Nano))
	if err := os.WriteFile(lockPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// runRecordLockChildHelper is the child side of the inter-process lock test.
// It never returns.
func runRecordLockChildHelper() {
	switch os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD") {
	case "exit":
		os.Exit(0)
	case "hold-release":
		settingsPath := os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD_SETTINGS")
		signalPath := os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD_SIGNAL")
		holdMs := 1500
		if v, err := strconv.Atoi(os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD_HOLD_MS")); err == nil {
			holdMs = v
		}
		lockPath := filepath.Join(filepath.Dir(managedAllowRecordPath(settingsPath)), "managed-allow.lock")
		if mkErr := os.MkdirAll(filepath.Dir(lockPath), 0o755); mkErr != nil {
			os.Exit(4)
		}
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			os.Exit(3)
		}
		_, _ = fmt.Fprintf(f, "pid=%d created=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
		_ = f.Close()
		// The child holds the lock while writing a known record state.
		_ = writeManagedAllowRecord(settingsPath, []string{"Bash(child-holds:*)"})
		if signalPath != "" {
			_ = os.WriteFile(signalPath, []byte("acquired"), 0o644)
		}
		time.Sleep(time.Duration(holdMs) * time.Millisecond)
		_ = os.Remove(lockPath)
		os.Exit(0)
	}
	os.Exit(2)
}

// SPEC-AUTONOMY-TIERS-001 M3 — tier → permission-bundle renderer (USER vs PROJECT scope).
// AC-003: defaultMode → USER scope; deny/ask → PROJECT scope.
// AC-004: deny/ask byte-identical across the 3 rendered tiers.
// AC-007: semi-auto → byte-identical to today's template (defaultMode=default).

// helperSettingsJSON builds a minimal settings.json body with a permissions
// block so the codegen region matcher has something to splice into.
func helperSettingsJSON(t *testing.T, defaultMode string) []byte {
	t.Helper()
	body := `{
  "permissions": {
    "defaultMode": "` + defaultMode + `"
  }
}`
	return []byte(body)
}

func writeSettings(t *testing.T, path, defaultMode string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, helperSettingsJSON(t, defaultMode), 0o644); err != nil {
		t.Fatal(err)
	}
}

// policyDocWith returns a minimal PolicyDocument with one deny + one ask entry
// (enough to exercise the deny/ask emission). The tier-invariance check
// (AC-004) does NOT depend on the entry count — the renderer loads doc once.
func policyDocWith(deny, ask string) *PolicyDocument {
	return &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Bash", ArgsPattern: deny, Decision: DecisionDeny, RiskTier: RiskTierIrreversible},
			{Tool: "Bash", ArgsPattern: ask, Decision: DecisionAsk, RiskTier: RiskTierWrite},
		},
	}
}

func TestRenderTierPermissions_ScopeSplit(t *testing.T) {
	// AC-003: defaultMode lands in USER scope; deny/ask land in PROJECT scope.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "project", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	doc := policyDocWith("git push --force:*", "git push:*")
	res, err := RenderTierPermissions(projectPath, userPath, "auto", doc)
	if err != nil {
		t.Fatalf("RenderTierPermissions: %v", err)
	}
	if res.UserPath != userPath || res.ProjectPath != projectPath {
		t.Errorf("paths mismatched: user=%q project=%q", res.UserPath, res.ProjectPath)
	}

	// USER scope MUST carry defaultMode=auto but NOT deny/ask.
	userBody, _ := os.ReadFile(userPath)
	var userPerms struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Deny        []string `json:"deny"`
			Ask         []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(userBody, &userPerms); err != nil {
		t.Fatalf("user settings unmarshal: %v\nbody: %s", err, userBody)
	}
	if userPerms.Permissions.DefaultMode != "auto" {
		t.Errorf("USER defaultMode = %q, want %q", userPerms.Permissions.DefaultMode, "auto")
	}
	if len(userPerms.Permissions.Deny) != 0 || len(userPerms.Permissions.Ask) != 0 {
		t.Errorf("USER scope MUST NOT carry deny/ask (they are PROJECT-scoped); got deny=%v ask=%v",
			userPerms.Permissions.Deny, userPerms.Permissions.Ask)
	}

	// PROJECT scope MUST carry deny/ask but NOT defaultMode.
	projBody, _ := os.ReadFile(projectPath)
	var projPerms struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Deny        []string `json:"deny"`
			Ask         []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(projBody, &projPerms); err != nil {
		t.Fatalf("project settings unmarshal: %v\nbody: %s", err, projBody)
	}
	if len(projPerms.Permissions.Deny) == 0 || len(projPerms.Permissions.Ask) == 0 {
		t.Errorf("PROJECT scope MUST carry deny/ask; got deny=%v ask=%v",
			projPerms.Permissions.Deny, projPerms.Permissions.Ask)
	}
	// AP-6: PROJECT scope MUST NOT carry auto/bypass defaultMode (v2.1.142+ makes
	// them USER-scope-only). defaultMode SHOULD be absent or "default" in PROJECT.
	if projPerms.Permissions.DefaultMode == "auto" || projPerms.Permissions.DefaultMode == "bypassPermissions" {
		t.Errorf("PROJECT defaultMode = %q — auto/bypass MUST NOT land in PROJECT scope (AP-6)", projPerms.Permissions.DefaultMode)
	}
}

func TestRenderTierPermissions_DenyAskInvariantAcrossTiers(t *testing.T) {
	// AC-004: deny/ask arrays byte-identical across the 3 rendered tiers.
	dir := t.TempDir()
	doc := policyDocWith("git push --force:*", "git push:*")
	modes := []string{"default", "auto", "bypassPermissions"}
	var denyBlobs, askBlobs [][]byte
	for _, mode := range modes {
		projectPath := filepath.Join(dir, mode, ".claude", "settings.json")
		writeSettings(t, projectPath, "default")
		if _, err := RenderTierPermissions(projectPath, projectPath, mode, doc); err != nil {
			t.Fatalf("RenderTierPermissions(%q): %v", mode, err)
		}
		body, _ := os.ReadFile(projectPath)
		var perms struct {
			Permissions struct {
				Deny []string `json:"deny"`
				Ask  []string `json:"ask"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(body, &perms); err != nil {
			t.Fatalf("unmarshal %q: %v", mode, err)
		}
		d, _ := json.Marshal(perms.Permissions.Deny)
		a, _ := json.Marshal(perms.Permissions.Ask)
		denyBlobs = append(denyBlobs, d)
		askBlobs = append(askBlobs, a)
	}
	for i := 1; i < len(modes); i++ {
		if string(denyBlobs[i]) != string(denyBlobs[0]) {
			t.Errorf("AC-004 violation: deny array differs between %q and %q:\n  %s\n  %s",
				modes[0], modes[i], denyBlobs[0], denyBlobs[i])
		}
		if string(askBlobs[i]) != string(askBlobs[0]) {
			t.Errorf("AC-004 violation: ask array differs between %q and %q:\n  %s\n  %s",
				modes[0], modes[i], askBlobs[0], askBlobs[i])
		}
	}
}

func TestRenderTierPermissions_SemiAutoDefaultMode(t *testing.T) {
	// AC-007: semi-auto → defaultMode="default" in USER scope (today's value).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "u", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "p", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	doc := policyDocWith("rm -rf /:*", "git push:*")
	if _, err := RenderTierPermissions(projectPath, userPath, "default", doc); err != nil {
		t.Fatalf("RenderTierPermissions: %v", err)
	}
	body, _ := os.ReadFile(userPath)
	var perms struct {
		Permissions struct {
			DefaultMode string `json:"defaultMode"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(body, &perms); err != nil {
		t.Fatalf("unmarshal user perms: %v", err)
	}
	if perms.Permissions.DefaultMode != "default" {
		t.Errorf("semi-auto USER defaultMode = %q, want %q", perms.Permissions.DefaultMode, "default")
	}
}

func TestRenderTierPermissions_Idempotent(t *testing.T) {
	// Two consecutive renders produce byte-identical output (mirrors AC-TPS-013).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "u", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "p", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")
	doc := policyDocWith("git push --force:*", "git push:*")

	if _, err := RenderTierPermissions(projectPath, userPath, "auto", doc); err != nil {
		t.Fatal(err)
	}
	firstUser, _ := os.ReadFile(userPath)
	firstProj, _ := os.ReadFile(projectPath)

	if _, err := RenderTierPermissions(projectPath, userPath, "auto", doc); err != nil {
		t.Fatal(err)
	}
	secondUser, _ := os.ReadFile(userPath)
	secondProj, _ := os.ReadFile(projectPath)

	if string(firstUser) != string(secondUser) {
		t.Errorf("USER scope not idempotent:\nfirst:  %s\nsecond: %s", firstUser, secondUser)
	}
	if string(firstProj) != string(secondProj) {
		t.Errorf("PROJECT scope not idempotent:\nfirst:  %s\nsecond: %s", firstProj, secondProj)
	}
}

// --- card t1630 card-review repair (P1) — the last_generated sidecar writer.
// The read side (mergeProjectAllow) already classifies against
// .moai/state/tool-policy/managed-allow.json, but nothing wrote the record, so
// an allow entry later removed from the policy document persisted forever
// (revoked permission resurrection). These tests pin the writer: a successful
// PROJECT-scope render persists full.Allow as last_generated; a removed
// generated entry is then dropped at the next render. (Review round 2 revised
// the failure-path contract — a failed render may advance the record to the
// staged union, which is the safe direction; the round-2 tests below pin that.)

// helperAllowDoc returns a PolicyDocument whose entries generate exactly the
// given allow specifiers ("Bash(x:*)" form) plus one deny and one ask entry so
// both scopes carry content. Allow entries sort deterministically (the codegen
// sorts by specifier), so expected allow lists are compared sorted.
func helperAllowDoc(allow ...string) *PolicyDocument {
	entries := []PolicyEntry{
		{Tool: "Bash", ArgsPattern: "git push --force:*", Decision: DecisionDeny, RiskTier: RiskTierIrreversible},
		{Tool: "Bash", ArgsPattern: "git push:*", Decision: DecisionAsk, RiskTier: RiskTierWrite},
	}
	for _, spec := range allow {
		inner := strings.TrimSuffix(strings.TrimPrefix(spec, "Bash("), ")")
		entries = append(entries, PolicyEntry{Tool: "Bash", ArgsPattern: inner, Decision: DecisionAllow, RiskTier: RiskTierRead})
	}
	return &PolicyDocument{Entries: entries}
}

// readManagedAllowRecord reads and parses the sidecar record for a project
// settings path. Fails the test when the record file is absent.
func readManagedAllowRecord(t *testing.T, projectSettingsPath string) managedAllowRecord {
	t.Helper()
	root := filepath.Dir(filepath.Dir(projectSettingsPath))
	body, err := os.ReadFile(filepath.Join(root, ".moai", "state", "tool-policy", "managed-allow.json"))
	if err != nil {
		t.Fatalf("managed-allow record not written: %v", err)
	}
	var rec managedAllowRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		t.Fatalf("managed-allow record is not valid JSON: %v\n%s", err, body)
	}
	return rec
}

func TestRenderTierPermissions_PersistsManagedAllowRecord(t *testing.T) {
	// (a) a successful render writes the record with the generated managed
	// allow list, and a later render with an updated document refreshes it.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")); err != nil {
		t.Fatalf("first render: %v", err)
	}
	rec := readManagedAllowRecord(t, projectPath)
	want := []string{"Bash(managed-a:*)", "Bash(managed-b:*)"}
	sort.Strings(want)
	if !reflect.DeepEqual(rec.LastGenerated, want) {
		t.Errorf("last_generated = %v, want %v", rec.LastGenerated, want)
	}

	// Refresh: the record must be overwritten, not appended to.
	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-c:*)")); err != nil {
		t.Fatalf("second render: %v", err)
	}
	rec = readManagedAllowRecord(t, projectPath)
	if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(managed-c:*)"}) {
		t.Errorf("last_generation refresh failed: last_generated = %v, want [Bash(managed-c:*)]", rec.LastGenerated)
	}
}

func TestRenderTierPermissions_RemovedGeneratedEntryIsDropped(t *testing.T) {
	// (b) the deciding test: an entry the document generated and later stops
	// generating must be dropped from the PROJECT managed block at the next
	// render (it is managed, not user-added). Without the record writer the
	// removed entry survives forever — revoked permission resurrection.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")); err != nil {
		t.Fatalf("first render: %v", err)
	}

	// The document now stops generating Bash(managed-b:*).
	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err != nil {
		t.Fatalf("second render: %v", err)
	}
	body, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, _, _, _, _ := ac001ReadPerms(t, body)
	want := []string{"Bash(managed-a:*)"}
	if !reflect.DeepEqual(allow, want) {
		t.Errorf("removed generated entry not dropped: allow = %v, want %v\nbody=%s", allow, want, body)
	}
}

func TestRenderTierPermissions_FailedProjectRenderLeavesSettingsUntouchedRecordSafe(t *testing.T) {
	// (d, revised by review round 2, re-revised by sync-audit F3) when the
	// PROJECT-scope render fails AFTER the record stage, the PROJECT settings
	// stay byte-identical and the record is ROLLED BACK to its pre-render
	// content — a failed render leaves no trace of its staged generation,
	// because the settings write that would have landed it never succeeded
	// (upgrading entries to managed off a failed write would let a later
	// policy revocation delete user-authored entries).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	recordPath := filepath.Join(dir, "proj", ".moai", "state", "tool-policy", "managed-allow.json")
	writeSettings(t, userPath, "default")
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid JSON without a permissions key — the region matcher fails.
	brokenBody := `{"env":{"FOO":"1"}}`
	if err := os.WriteFile(projectPath, []byte(brokenBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(recordPath), 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := `{"last_generated":["Bash(sentinel:*)"]}`
	if err := os.WriteFile(recordPath, []byte(sentinel), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
		t.Fatalf("render over a permissions-less file must fail")
	}
	afterBody, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterBody) != brokenBody {
		t.Errorf("PROJECT settings changed on failed render:\nbefore=%s\nafter=%s", brokenBody, afterBody)
	}
	rec := readManagedAllowRecord(t, projectPath)
	if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(sentinel:*)"}) {
		t.Errorf("record after failed render = %v, want the rolled-back pre-render [Bash(sentinel:*)]", rec.LastGenerated)
	}

	// Failed render with no record yet: the rollback removes the staged file
	// — the record-absent state is restored exactly.
	freshProject := filepath.Join(dir, "fresh", ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(freshProject), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(freshProject, []byte(brokenBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderTierPermissions(freshProject, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
		t.Fatalf("render over a permissions-less file must fail")
	}
	freshAfter, err := os.ReadFile(freshProject)
	if err != nil {
		t.Fatal(err)
	}
	if string(freshAfter) != brokenBody {
		t.Errorf("fresh PROJECT settings changed on failed render:\nbefore=%s\nafter=%s", brokenBody, freshAfter)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "fresh", ".moai", "state", "tool-policy", "managed-allow.json")); !os.IsNotExist(statErr) {
		t.Errorf("failed render must roll back to record-absent on a fresh project; stat err=%v", statErr)
	}
}

// --- card t1630 card-review round 2 (P2-1, P2-2) — failure-ordering and the
// BuildInto generation path. P2-1: the record must be staged BEFORE the
// PROJECT settings mutation, so a settings write can never land entries the
// record does not know about, and a failed settings write cannot corrupt the
// next classification. P2-2: the `moai tool-policy build` path (BuildInto /
// BuildIntoAuto, the TargetJSON project-settings target) is a
// policy-generation path too and must maintain the same record.

func TestRenderTierPermissions_RecordStageFailureLeavesProjectSettingsUntouched(t *testing.T) {
	// (P2-1a) when the record cannot be staged, the render errors out BEFORE
	// any PROJECT settings mutation — settings entries must never exist
	// without the record knowing they are managed.
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")
	before, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}

	// Block the record path: a regular file where the record's directory must
	// be, so the stage's MkdirAll fails.
	if err := os.MkdirAll(filepath.Join(dir, "proj", ".moai", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "proj", ".moai", "state", "tool-policy"), []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
		t.Fatalf("render must error when the record cannot be staged")
	}
	after, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("PROJECT settings mutated although the record stage failed:\nbefore=%s\nafter=%s", before, after)
	}
}

func TestRenderTierPermissions_SettingsWriteFailureAfterRecordStage_ClassifiesCorrectlyNextRender(t *testing.T) {
	// (P2-1b) a failed PROJECT settings write after a successful record stage
	// must leave the record carrying the new generation's entries (and the
	// managed entries the un-updated project file still holds), so the next
	// successful render classifies correctly: still-generated entries are
	// re-added as managed, and an entry the policy revoked is dropped rather
	// than resurrected as "user-added".
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	docAB := helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docAB); err != nil {
		t.Fatalf("render 1: %v", err)
	}

	// The policy changes between renders: drops managed-b, adds managed-c.
	docAC := helperAllowDoc("Bash(managed-a:*)", "Bash(managed-c:*)")

	// Break the PROJECT settings write: a directory where the settings file
	// must be makes renderIntoFile's read fail before any write.
	if err := os.Remove(projectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docAC); err == nil {
		t.Fatalf("render 2 must fail: the settings path is a directory")
	}
	// Rollback (sync-audit F3): the failed render must leave the record at
	// its pre-render content — the staged generation belongs to a write that
	// never landed, so no entry may read as managed off this render.
	rec := readManagedAllowRecord(t, projectPath)
	got := append([]string{}, rec.LastGenerated...)
	sort.Strings(got)
	want := []string{"Bash(managed-a:*)", "Bash(managed-b:*)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("record after failed settings write = %v, want the rolled-back pre-render %v", got, want)
	}

	// Recovery: restore the settings file to its pre-failure content (render
	// 1's output: managed-a + managed-b).
	if err := os.RemoveAll(projectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatal(err)
	}
	recovered := `{"permissions":{"allow":["Bash(managed-a:*)","Bash(managed-b:*)"]}}`
	if err := os.WriteFile(projectPath, []byte(recovered), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RenderTierPermissions(projectPath, userPath, "default", docAC); err != nil {
		t.Fatalf("render 3: %v", err)
	}
	afterBody, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, _, _, _, _ := ac001ReadPerms(t, afterBody)
	wantAllow := []string{"Bash(managed-a:*)", "Bash(managed-c:*)"}
	if !reflect.DeepEqual(allow, wantAllow) {
		t.Errorf("post-recovery classification: allow = %v, want %v (managed-b dropped, managed-c re-added as managed)\nbody=%s", allow, wantAllow, afterBody)
	}
}

func TestBuildIntoPersistsManagedAllowRecord_RemovedEntryNotKeptAsUserAdded(t *testing.T) {
	// (P2-2) the `moai tool-policy build` path (BuildInto, TargetJSON — the
	// project settings target) is a policy-generation path: it must maintain
	// the same sidecar record, so an entry it generated that the policy later
	// drops is DROPPED by the next render, not kept as user-added.
	dir := t.TempDir()
	projectPath := filepath.Join(dir, ".claude", "settings.json")
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	docAB := helperAllowDoc("Bash(managed-a:*)", "Bash(managed-b:*)")
	if _, err := BuildInto(projectPath, docAB, TargetJSON, ""); err != nil {
		t.Fatalf("BuildInto: %v", err)
	}
	// BuildInto must have maintained the record.
	recBody, recErr := os.ReadFile(managedAllowRecordPath(projectPath))
	if recErr != nil {
		t.Errorf("BuildInto did not maintain the managed-allow record: %v", recErr)
	} else {
		var rec managedAllowRecord
		if err := json.Unmarshal(recBody, &rec); err != nil {
			t.Fatalf("record is not valid JSON: %v", err)
		}
		got := append([]string{}, rec.LastGenerated...)
		sort.Strings(got)
		if !reflect.DeepEqual(got, []string{"Bash(managed-a:*)", "Bash(managed-b:*)"}) {
			t.Errorf("record after BuildInto = %v, want [Bash(managed-a:*) Bash(managed-b:*)]", got)
		}
	}

	// The policy document drops managed-b; the next render must drop it
	// (managed), not keep it (user-added).
	docA := helperAllowDoc("Bash(managed-a:*)")
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docA); err != nil {
		t.Fatalf("render after build: %v", err)
	}
	body, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, _, _, _, _ := ac001ReadPerms(t, body)
	if !reflect.DeepEqual(allow, []string{"Bash(managed-a:*)"}) {
		t.Errorf("entry built by BuildInto then removed from the policy was kept as user-added: allow = %v, want [Bash(managed-a:*)]\nbody=%s", allow, body)
	}
}

// --- card t1630 sync-audit repairs (F2, F3) — record-protocol concurrency
// and failure ownership.

func TestConcurrentRendersKeepRecordConsistentWithSettings(t *testing.T) {
	// (F2) concurrent renders of the same project must serialize the record
	// protocol (read→classify→stage→settings-write→converge). After a burst,
	// the sidecar record must name exactly the managed entries the settings
	// file carries: an inverted record (a converge landing after a later
	// render's completion) is the resurrection hazard — an entry the policy
	// has since revoked reads as never-generated and persists as
	// "user-added".
	docA := helperAllowDoc("Bash(managed-a:*)", "Bash(managed-a2:*)")
	docB := helperAllowDoc("Bash(managed-b:*)", "Bash(managed-b2:*)")
	for round := 0; round < 30; round++ {
		dir := t.TempDir()
		userPath := filepath.Join(dir, "user", ".claude", "settings.json")
		projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
		writeSettings(t, userPath, "default")
		writeSettings(t, projectPath, "default")

		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			doc := docA
			if i%2 == 1 {
				doc = docB
			}
			wg.Add(1)
			go func(d *PolicyDocument) {
				defer wg.Done()
				_, _ = RenderTierPermissions(projectPath, userPath, "default", d)
			}(doc)
		}
		wg.Wait()

		rec := readManagedAllowRecord(t, projectPath)
		body, err := os.ReadFile(projectPath)
		if err != nil {
			t.Fatal(err)
		}
		allow, _, _, _, _ := ac001ReadPerms(t, body)
		gotRec := append([]string{}, rec.LastGenerated...)
		gotAllow := append([]string{}, allow...)
		sort.Strings(gotRec)
		sort.Strings(gotAllow)
		if !reflect.DeepEqual(gotRec, gotAllow) {
			t.Fatalf("round %d: record inverted against settings: record=%v allow=%v", round, gotRec, gotAllow)
		}
	}
}

func TestFailedSettingsWriteRollsRecordBack_UserAddedEntrySurvivesRevocation(t *testing.T) {
	// (F3) a stage-ahead that records an entry and is then followed by a
	// FAILED settings write must not upgrade that entry to managed: the
	// record rolls back to its pre-render content, so a later policy
	// revision that drops the entry cannot delete the user's own allow
	// entry (a tool-initiated removal of a user-authored entry).
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	writeSettings(t, userPath, "default")
	writeSettings(t, projectPath, "default")

	// Render 1: the policy generates managed-m; record [managed-m].
	docM := helperAllowDoc("Bash(managed-m:*)")
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docM); err != nil {
		t.Fatalf("render 1: %v", err)
	}

	// The user manually adds user-x alongside the managed entry.
	userBody := `{"permissions":{"allow":["Bash(managed-m:*)","Bash(user-x:*)"]}}`
	if err := os.WriteFile(projectPath, []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}

	// Render 2: the policy now ALSO generates user-x (it coincides with the
	// user's own entry). Classification keeps it (absent from the record);
	// the stage-ahead records it — and then the PROJECT write fails.
	docMX := helperAllowDoc("Bash(managed-m:*)", "Bash(user-x:*)")
	if err := os.Remove(projectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docMX); err == nil {
		t.Fatalf("render 2 must fail: the settings path is a directory")
	}
	// Rollback: the record must be back to its pre-render content —
	// user-x must not read as managed off a failed write.
	rec := readManagedAllowRecord(t, projectPath)
	if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(managed-m:*)"}) {
		t.Errorf("record after failed settings write = %v, want the pre-render [Bash(managed-m:*)] (a failed render must not upgrade entries to managed)", rec.LastGenerated)
	}

	// Recovery: restore the settings file as it was; the policy then REVOKES
	// user-x. The next healthy render must KEEP it (user-added).
	if err := os.RemoveAll(projectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectPath, []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderTierPermissions(projectPath, userPath, "default", docM); err != nil {
		t.Fatalf("render 3: %v", err)
	}
	afterBody, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, _, _, _, _ := ac001ReadPerms(t, afterBody)
	want := []string{"Bash(managed-m:*)", "Bash(user-x:*)"}
	if !reflect.DeepEqual(allow, want) {
		t.Errorf("policy revoked user-x but the user's own entry was deleted: allow = %v, want %v (user-added survives policy revocation)\nbody=%s", allow, want, afterBody)
	}
}

// --- card t1630 sync-audit repair F2' — project-scoped inter-process record
// lock. The in-process recordProtocolMu does not cover two concurrent moai OS
// processes: a child process holding the record lock while it writes a known
// record state must block a concurrent parent render until release (and a
// dead holder must be taken over, never wedged).

func TestInterProcessRecordLock_TwoProcessRepro(t *testing.T) {
	t.Run("blocks-until-child-releases-and-converges", func(t *testing.T) {
		dir := t.TempDir()
		userPath := filepath.Join(dir, "user", ".claude", "settings.json")
		projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
		writeSettings(t, userPath, "default")
		writeSettings(t, projectPath, "default")
		lockPath := testRecordLockPath(t, projectPath)
		signalPath := filepath.Join(dir, "child-acquired")

		// Child: acquires the lock, writes its sentinel record, signals
		// readiness, holds, then releases and exits.
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(),
			"MOAI_TOOLPOLICY_LOCK_CHILD=hold-release",
			"MOAI_TOOLPOLICY_LOCK_CHILD_SETTINGS="+projectPath,
			"MOAI_TOOLPOLICY_LOCK_CHILD_SIGNAL="+signalPath,
			"MOAI_TOOLPOLICY_LOCK_CHILD_HOLD_MS=1500",
		)
		if err := cmd.Start(); err != nil {
			t.Fatalf("reexec child: %v", err)
		}
		waited := false
		t.Cleanup(func() {
			if !waited {
				_ = cmd.Process.Kill()
				_, _ = cmd.Process.Wait()
			}
		})

		// Wait until the child provably holds the lock.
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(signalPath); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("child never acquired the record lock")
			}
			time.Sleep(20 * time.Millisecond)
		}

		// The parent render must BLOCK until the child releases — the render
		// may not return (or mutate the record) while the lock is held.
		doc := helperAllowDoc("Bash(managed-a:*)")
		if _, err := RenderTierPermissions(projectPath, userPath, "default", doc); err != nil {
			t.Fatalf("render must proceed after the child releases: %v", err)
		}
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Errorf("render returned while the inter-process lock was still held: lockfile present (err=%v) — the render did not respect the cross-process lock", err)
		}
		if err := cmd.Wait(); err != nil {
			t.Fatalf("child: %v", err)
		}
		waited = true

		// The parent's converge replaced the child's sentinel record.
		rec := readManagedAllowRecord(t, projectPath)
		if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(managed-a:*)"}) {
			t.Errorf("record after contended render = %v, want the parent's converged [Bash(managed-a:*)]", rec.LastGenerated)
		}
	})

	t.Run("takes-over-a-dead-pid-lockfile", func(t *testing.T) {
		dir := t.TempDir()
		userPath := filepath.Join(dir, "user", ".claude", "settings.json")
		projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
		writeSettings(t, userPath, "default")
		writeSettings(t, projectPath, "default")
		lockPath := testRecordLockPath(t, projectPath)
		seedDeadPidLock(t, lockPath)

		// A dead holder must never wedge a render: after the bounded retry
		// budget the lock is taken over, the render proceeds, and the
		// lockfile is released.
		doc := helperAllowDoc("Bash(managed-a:*)")
		if _, err := RenderTierPermissions(projectPath, userPath, "default", doc); err != nil {
			t.Fatalf("render must take over a dead-pid lock: %v", err)
		}
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Errorf("dead-pid lockfile was not taken over and released; stat err=%v", err)
		}
		rec := readManagedAllowRecord(t, projectPath)
		if !reflect.DeepEqual(rec.LastGenerated, []string{"Bash(managed-a:*)"}) {
			t.Errorf("record after takeover render = %v, want [Bash(managed-a:*)]", rec.LastGenerated)
		}
	})

	t.Run("errors-after-budget-on-a-live-holder", func(t *testing.T) {
		// A LIVE holder beyond the retry budget must surface an error, never
		// an infinite wait (bounded retry then error) — and never a takeover
		// of a provably-live pid while the lockfile is fresh.
		dir := t.TempDir()
		userPath := filepath.Join(dir, "user", ".claude", "settings.json")
		projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
		writeSettings(t, userPath, "default")
		writeSettings(t, projectPath, "default")
		lockPath := testRecordLockPath(t, projectPath)

		// Seed a lockfile naming THIS (live) process, freshly created.
		if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf("pid=%d created=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
		if err := os.WriteFile(lockPath, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}

		if _, err := RenderTierPermissions(projectPath, userPath, "default", helperAllowDoc("Bash(managed-a:*)")); err == nil {
			t.Fatalf("render must error when a live holder keeps the lock beyond the retry budget")
		}
	})
}

// --- SPEC-USER-SETTINGS-PROTECT-001 (card t1630) — repository-authored
// assertions for AC-001, AC-002, and AC-003 part A. The overlay probes under
// .moai/reports/t1630/evidence/ carried the RED observation on the pinned
// tree; these are the same assertions authored in the repository
// (acceptance.md: "run-phase: the same assertion authored in the repository").

// ac001ReadPerms unmarshals a settings body into the fields AC-001 asserts on.
func ac001ReadPerms(t *testing.T, body []byte) (allow, ask, deny, additional []string, env map[string]string) {
	t.Helper()
	var wrap struct {
		Permissions struct {
			Allow                 []string `json:"allow"`
			Ask                   []string `json:"ask"`
			Deny                  []string `json:"deny"`
			AdditionalDirectories []string `json:"additionalDirectories"`
		} `json:"permissions"`
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		t.Fatalf("after is not valid JSON: %v\n%s", err, body)
	}
	return wrap.Permissions.Allow, wrap.Permissions.Ask, wrap.Permissions.Deny, wrap.Permissions.AdditionalDirectories, wrap.Env
}

// TestAC001PreservesListsAndUnmodelledKeys: the USER-scope writer keeps the
// user's allow/ask/deny entries, every unmodelled permissions key, and every
// sibling region (REQ-001). The EC-1..EC-4 edge cases of acceptance.md are
// bound to this criterion and asserted as subtests.
func TestAC001PreservesListsAndUnmodelledKeys(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	before := `{"permissions": {
      "defaultMode": "plan",
      "allow": ["Bash(make:*)"],
      "ask": ["Bash(rm:*)"],
      "deny": ["Read(./secret)"],
      "additionalDirectories": ["/tmp/x"]
    },"env":{"FOO":"1"}}`
	if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteUserDefaultMode(userPath, "acceptEdits"); err != nil {
		t.Fatalf("WriteUserDefaultMode: %v", err)
	}
	after, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatal(err)
	}
	allow, ask, deny, additional, env := ac001ReadPerms(t, after)
	if !reflect.DeepEqual(allow, []string{"Bash(make:*)"}) {
		t.Errorf("permissions.allow was removed; after=%s", after)
	}
	if !reflect.DeepEqual(ask, []string{"Bash(rm:*)"}) {
		t.Errorf("permissions.ask was removed; after=%s", after)
	}
	if !reflect.DeepEqual(deny, []string{"Read(./secret)"}) {
		t.Errorf("permissions.deny was removed; after=%s", after)
	}
	if !reflect.DeepEqual(additional, []string{"/tmp/x"}) {
		t.Errorf("permissions.additionalDirectories (unmodelled key) was removed; after=%s", after)
	}
	if !reflect.DeepEqual(env, map[string]string{"FOO": "1"}) {
		t.Errorf("sibling env region was removed; after=%s", after)
	}

	t.Run("ec1-absent-file", func(t *testing.T) {
		// EC-1: a settings file that does not exist yet — the writer creates
		// the minimal permissions skeleton with the tier default, no lists.
		absent := filepath.Join(t.TempDir(), ".claude", "settings.json")
		if err := WriteUserDefaultMode(absent, "acceptEdits"); err != nil {
			t.Fatalf("WriteUserDefaultMode on absent file: %v", err)
		}
		body, err := os.ReadFile(absent)
		if err != nil {
			t.Fatal(err)
		}
		allow, ask, deny, _, _ := ac001ReadPerms(t, body)
		if len(allow) != 0 || len(ask) != 0 || len(deny) != 0 {
			t.Errorf("created file MUST NOT invent lists; body=%s", body)
		}
	})
	t.Run("ec2-no-permissions-key", func(t *testing.T) {
		// EC-2: no "permissions" key — the writer reports the region error and
		// leaves the file unchanged.
		path := filepath.Join(t.TempDir(), "settings.json")
		body := `{"env":{"FOO":"1"}}`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err == nil {
			t.Errorf("missing permissions key MUST return the region error")
		}
		afterBody, _ := os.ReadFile(path)
		if string(afterBody) != body {
			t.Errorf("file MUST stay unchanged on the error path:\nbefore=%s\nafter=%s", body, afterBody)
		}
	})
	t.Run("ec3-invalid-json", func(t *testing.T) {
		// EC-3: invalid JSON — the writer reports the region error and leaves
		// the file unchanged.
		path := filepath.Join(t.TempDir(), "settings.json")
		body := `{not json`
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err == nil {
			t.Errorf("invalid JSON MUST return the region error")
		}
		afterBody, _ := os.ReadFile(path)
		if string(afterBody) != body {
			t.Errorf("file MUST stay unchanged on the error path:\nbefore=%s\nafter=%s", body, afterBody)
		}
	})
	t.Run("ec4-duplicate-allow-entries", func(t *testing.T) {
		// EC-4: duplicate allow entries are preserved as written (no
		// de-duplication, no removal) on the write path.
		path := filepath.Join(t.TempDir(), "settings.json")
		before := `{"permissions":{"allow":["Bash(a:*)","Bash(a:*)"]}}`
		if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := WriteUserDefaultMode(path, "acceptEdits"); err != nil {
			t.Fatalf("WriteUserDefaultMode: %v", err)
		}
		afterBody, _ := os.ReadFile(path)
		allow, _, _, _, _ := ac001ReadPerms(t, afterBody)
		if !reflect.DeepEqual(allow, []string{"Bash(a:*)", "Bash(a:*)"}) {
			t.Errorf("duplicate allow entries MUST be preserved as written; after=%s", afterBody)
		}
	})
}

// TestAC002RegionBytesUnchangedWhenDefaultModeMatches: when the resolved tier
// default already equals the file's defaultMode, the writer does not rewrite
// the permissions region — the file bytes are identical (REQ-002).
func TestAC002RegionBytesUnchangedWhenDefaultModeMatches(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Non-canonical layout: the bytes must survive untouched, not be re-rendered.
	before := "{\n  \"permissions\": {\n      \"allow\": [\"Bash(make:*)\"],\n      \"defaultMode\":   \"acceptEdits\"\n  },\n  \"env\": {\"FOO\": \"1\"}\n}\n"
	if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteUserDefaultMode(userPath, "acceptEdits"); err != nil {
		t.Fatalf("WriteUserDefaultMode: %v", err)
	}
	after, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Errorf("file bytes changed\nbefore=%q\nafter=%q", before, string(after))
	}
}

// TestAC003ListsPreservedUnderEitherDisposition: the lists survive whether
// the existing defaultMode differs from or matches the resolved tier default
// (AC-003 part A).
func TestAC003ListsPreservedUnderEitherDisposition(t *testing.T) {
	cases := []struct {
		name     string
		existing string
		resolved string
	}{
		{"differing", "plan", "acceptEdits"},
		{"matching", "acceptEdits", "acceptEdits"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			userPath := filepath.Join(dir, ".claude", "settings.json")
			if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
				t.Fatal(err)
			}
			before := `{"permissions": {
      "defaultMode": "` + c.existing + `",
      "allow": ["Bash(make:*)"],
      "ask": ["Bash(rm:*)"],
      "deny": ["Read(./secret)"],
      "additionalDirectories": ["/tmp/x"]
    },"env":{"FOO":"1"}}`
			if err := os.WriteFile(userPath, []byte(before), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := WriteUserDefaultMode(userPath, c.resolved); err != nil {
				t.Fatalf("WriteUserDefaultMode: %v", err)
			}
			after, err := os.ReadFile(userPath)
			if err != nil {
				t.Fatal(err)
			}
			allow, ask, deny, additional, _ := ac001ReadPerms(t, after)
			if !reflect.DeepEqual(allow, []string{"Bash(make:*)"}) {
				t.Errorf("permissions.allow was removed; after=%s", after)
			}
			if !reflect.DeepEqual(ask, []string{"Bash(rm:*)"}) {
				t.Errorf("permissions.ask was removed; after=%s", after)
			}
			if !reflect.DeepEqual(deny, []string{"Read(./secret)"}) {
				t.Errorf("permissions.deny was removed; after=%s", after)
			}
			if !reflect.DeepEqual(additional, []string{"/tmp/x"}) {
				t.Errorf("permissions.additionalDirectories was removed; after=%s", after)
			}
		})
	}
}
