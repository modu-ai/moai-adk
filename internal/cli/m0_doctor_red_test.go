// m0_doctor_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, cli doctor
// family: AC-012 (unregistered mirror copy classified+reported), AC-019a/b
// (load-failure honesty + byte/path preservation), AC-020 (vacuous lock
// match), AC-021 (workflow root by required file), AC-022 (fallback keeps
// project scope L1).
//
// M0 discipline: observation only — no production change.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestMigrationClassifiesUnregisteredMirrorCopy — AC-012 (ledger 4,
// REQ-SRF-001). Given a .agents/skills mirror copy with NO manifest entry
// plus a confirmed user-side counterpart, the migration must classify and
// REPORT the copy — removal stays behind provenance proof. RED-now reason:
// migrate_project_assets.go:119-122 counts an untracked file "untouched"
// and returns without any per-file report — the copy is permanently
// invisible.
func TestMigrationClassifiesUnregisteredMirrorCopy(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// An unregistered mirror copy under the migration's walk roots: no
	// project-manifest record, a same-named user-side counterpart present.
	const unregRel = ".agents/skills/moai-unregistered/SKILL.md"
	unregAbs := filepath.Join(root, filepath.FromSlash(unregRel))
	if err := os.MkdirAll(filepath.Dir(unregAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unregAbs, []byte("# an unregistered mirror copy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var report strings.Builder
	reportFn := func(format string, args ...interface{}) {
		report.WriteString(fmt.Sprintf(format, args...) + "\n")
	}
	if _, err := migrateProjectCommonAssets(root, home, true, nil, reportFn); err != nil {
		t.Fatalf("migrateProjectCommonAssets: %v", err)
	}

	out := report.String()
	if !strings.Contains(out, unregRel) && !strings.Contains(out, "moai-unregistered") {
		t.Fatalf("RED (intended): the migration report never names the unregistered mirror copy %q — it is counted as untouched with no classification or visibility. Report was:\n%s", unregRel, out)
	}
	if !strings.Contains(out, "migration:") {
		t.Fatalf("the report carried no migration lines at all:\n%s", out)
	}
}

// TestDoctorProjectManifestLoadFailureNotDisguised — AC-019a arm 1 (ledger 2,
// REQ-DOC-001). Given a corrupt PROJECT manifest, checkProjectVsLock must
// report the load failure at a failure grade — never disguise it as
// CheckOK "no project manifest (nothing to compare)".
func TestDoctorProjectManifestLoadFailureNotDisguised(t *testing.T) {
	root := t.TempDir()
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("{ this is not valid json for a manifest\n")
	if err := os.WriteFile(filepath.Join(moai, "manifest.json"), original, 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkProjectVsLock(root, false)
	if check.Status == "ok" {
		t.Fatalf("RED (intended): a corrupt project manifest load produced CheckOK %q — the load failure is disguised as nothing-to-compare (doctor_user_install.go:134-137)", check.Message)
	}
}

// TestDoctorProjectManifestPreservedOnLoadFailure — AC-019a arm 2 (ledger 2,
// REQ-DOC-001, the byte+path preservation probe from the verification
// two-cell discipline). The corrupt ORIGINAL must stay at its original path
// with its original bytes, and a pre-existing .corrupt recovery copy must
// not be overwritten. RED-now reason: manifest.Load's corrupt arm does
// os.Rename(original, original+".corrupt") (internal/manifest/manifest.go:86)
// — renaming the original AWAY and silently destroying the prior .corrupt.
//
// Note: this arm currently fails through the doctor path's own Load call;
// the assertion keeps the preservation probe inside the same observation the
// AC demands (both sub-claims, one Given).
func TestDoctorProjectManifestPreservedOnLoadFailure(t *testing.T) {
	root := t.TempDir()
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("{ corrupt original bytes v2 — must stay put\n")
	manifestPath := filepath.Join(moai, "manifest.json")
	if err := os.WriteFile(manifestPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	priorCorrupt := []byte("{ prior corrupt recovery copy from an earlier incident\n")
	corruptPath := manifestPath + ".corrupt"
	if err := os.WriteFile(corruptPath, priorCorrupt, 0o644); err != nil {
		t.Fatal(err)
	}

	_ = checkProjectVsLock(root, false) // the observation target; arm 1 judges its verdict

	got, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("RED (intended): the corrupt original was moved away from %s during a read-only doctor check: %v", manifestPath, err)
	}
	if string(got) != string(original) {
		t.Fatalf("RED (intended): the corrupt original's bytes changed at %s during a read-only doctor check (Load renamed it aside)", manifestPath)
	}
	priorGot, err := os.ReadFile(corruptPath)
	if err != nil {
		t.Fatalf("RED (intended): the prior .corrupt recovery copy disappeared: %v", err)
	}
	if string(priorGot) != string(priorCorrupt) {
		t.Fatalf("RED (intended): the prior .corrupt recovery copy was overwritten (os.Rename at internal/manifest/manifest.go:86 destroys it) — now holds: %q", priorGot)
	}
}

// TestDoctorUserInstallHonestFailureAndReadOnly — AC-019b (ledger 2a,
// REQ-DOC-001, born-green regression guard). Given a corrupt USER manifest,
// checkUserInstallIntegrity reports the failure honestly and never mutates
// the file. RED at HEAD would be a regression finding for the leader.
func TestDoctorUserInstallHonestFailureAndReadOnly(t *testing.T) {
	home := t.TempDir()
	moai := filepath.Join(home, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	corrupt := []byte("{ corrupt user manifest\n")
	userManifest := filepath.Join(moai, "user-assets.json")
	if err := os.WriteFile(userManifest, corrupt, 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkUserInstallIntegrity(home, false)
	if check.Status != "warn" && check.Status != "fail" {
		t.Fatalf("RED-guard regression: a corrupt user manifest produced %s %q — the honest-failure invariant broke", check.Status, check.Message)
	}
	got, err := os.ReadFile(userManifest)
	if err != nil {
		t.Fatalf("RED-guard regression: the doctor check mutated the user manifest path: %v", err)
	}
	if string(got) != string(corrupt) {
		t.Fatalf("RED-guard regression: the user manifest's bytes changed during a read-only check")
	}
}

// TestDoctorLockCheckVacuousNotReportedMatch — AC-020 (ledger 2b,
// REQ-DOC-002). Given a project tree with NO lock-comparable surface (only
// .claude/settings.json + .moai/config), "project tree matches the lock
// file" must not be reported — the vacuous condition must be named.
// RED-now reason: with 0 files passing the root gate (:165), the switch
// falls to default OK (:184).
func TestDoctorLockCheckVacuousNotReportedMatch(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude", "settings.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	secDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(secDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secDir, "llm.yaml"), []byte("llm:\n  harness: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A valid, loadable lock manifest so the :134 interception doesn't fire —
	// the vacuous arm is the one under observation.
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load fresh manifest: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("seed lock manifest: %v", err)
	}

	check := checkProjectVsLock(root, false)
	if check.Status == "ok" && strings.Contains(check.Message, "matches the lock file") {
		t.Fatalf("RED (intended): a tree with zero lock-comparable files reported %q — the vacuous comparison is disguised as a match", check.Message)
	}
}

// TestDoctorWorkflowRootSelectedByRequiredFile — AC-021 (ledger 9b,
// REQ-DOC-003). Given an EMPTY project workflows dir and a healthy USER
// workflow tree, the workflow root must be selected by required-file
// presence and L4 must pass. RED-now reason: the selection gate checks the
// directory's EXISTENCE (doctor_harness.go:57), so the empty project dir
// wins and L4 reads nothing.
func TestDoctorWorkflowRootSelectedByRequiredFile(t *testing.T) {
	root, home := seedHarnessDoctorProject(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// EMPTY project workflows dir: existence-only evidence.
	projWf := filepath.Join(root, ".claude", "skills", "moai", "workflows")
	if err := os.MkdirAll(projWf, 0o755); err != nil {
		t.Fatal(err)
	}
	// Healthy USER workflow tree.
	writeDoctorUserWorkflows(t, home)

	check := runHarnessCheck(root)
	if strings.Contains(check.Message, "L4:FAIL") || strings.Contains(check.Detail, "L4 ") {
		t.Fatalf("RED (intended): L4 failed although the USER workflow tree is healthy — the empty project workflows dir won the existence-only selection gate (:57). Message: %q Detail: %q", check.Message, check.Detail)
	}
	if !strings.Contains(check.Message, "L4:PASS") {
		t.Fatalf("L4 did not report PASS (message %q, detail %q)", check.Message, check.Detail)
	}
}

// TestDoctorFallbackKeepsProjectScopeL1 — AC-022 (ledger 7b, REQ-DOC-004).
// Given the fallback state (project workflows ABSENT, user workflows
// present), L1 must still examine the PROJECT skillsDir — the wholesale
// skillsDir replacement (:58) must not flip L1 to the user scope.
// RED-now reason: the replacement hands L1 the user dir, so a broken
// project-side harness skill reads PASS (the original L6:FAIL·L1:PASS shape).
func TestDoctorFallbackKeepsProjectScopeL1(t *testing.T) {
	root, home := seedHarnessDoctorProject(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// A broken PROJECT-side harness skill (no triggers section) that L1 must
	// catch — and a missing project workflows dir so the fallback fires.
	broken := filepath.Join(root, ".claude", "skills", "hns-probe-broken", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(broken), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(broken, []byte("---\nname: hns-probe-broken\n---\nno triggers here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeDoctorUserWorkflows(t, home)

	check := runHarnessCheck(root)
	if !strings.Contains(check.Message, "L1:FAIL") {
		t.Fatalf("RED (intended): L1 reported %q with a broken project-side harness skill present — the wholesale skillsDir swap handed L1 the USER scope (doctor_harness.go:58), so the project defect is invisible", check.Message)
	}
}

// seedHarnessDoctorProject builds a minimal configured project for
// runHarnessCheck (the .moai/harness configured gate must pass to reach the
// L-battery) plus a healthy USER home skeleton.
func seedHarnessDoctorProject(t *testing.T) (root, home string) {
	t.Helper()
	root = t.TempDir()
	home = t.TempDir()
	harnessDir := filepath.Join(root, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(harnessDir, "main.md"), []byte("# baseline\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, home
}

// writeDoctorUserWorkflows writes the user workflow tree with the four
// required files, each carrying the L4 import line.
func writeDoctorUserWorkflows(t *testing.T, home string) {
	t.Helper()
	wf := filepath.Join(home, ".claude", "skills", "moai", "workflows")
	if err := os.MkdirAll(wf, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"plan.md", "run.md", "sync.md", "design.md"} {
		if err := os.WriteFile(filepath.Join(wf, name), []byte("import @.moai/harness/run-extension.md\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
