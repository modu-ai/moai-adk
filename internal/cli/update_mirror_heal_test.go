package cli

// update_mirror_heal_test.go — SPEC-UPDATE-MIRROR-HEAL-001 (card t520).
//
// `moai update` on a version-matched project returns before Deploy
// (update_template_sync.go, version-match branch), and BOTH .agents/skills
// producers live inside Deploy: the Path A symlink mirror and the Path B
// template-published SKILL.md files. A deleted mirror is therefore permanent
// for that project. These tests pin the repair pass that runs beside the
// early return — never instead of it.
//
// Isolation (REQ-UMH-009 / C-3): every fixture lives under t.TempDir(). No
// test here reads, writes, or deletes THIS repository's .agents/ state, which
// is an observation subject for sibling cards t498 and t510.
//
// Serial by construction: the fixtures that drive the real template-sync cycle
// chdir (runTemplateSyncAt), so this family carries no t.Parallel — the same
// trade update_llm_preserve_test.go already makes.

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

// mirrorHealDeployedProject builds a project under t.TempDir() by ACTUALLY
// RUNNING the deploy path, so its .moai/config/sections/system.yaml stamp,
// its .claude/skills tree, and its .agents/skills mirror are all products of
// one run. AC-UMH-001 forbids hand-writing the stamp: a hand-written stamp is
// what let an earlier SPEC revision bypass the §3.6 converse question.
func mirrorHealDeployedProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// A below-current stamp is the version-MISMATCH trigger that makes the
	// sync actually deploy; the deploy then overwrites it with {{.Version}}.
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n  template_version: \"0.0.0\"\n")
	writeTestFile(t, root, ".moai/manifest.json", "{}\n")
	runTemplateSyncAt(t, root)

	stamp, err := plan.GetProjectConfigVersion(root)
	if err != nil {
		t.Fatalf("read deployed stamp: %v", err)
	}
	if stamp != version.GetVersion() {
		t.Fatalf("fixture precondition: deployed stamp %q != binary version %q — the fixture is not a product of a real deploy", stamp, version.GetVersion())
	}
	if len(mirrorEntryNames(t, root)) == 0 {
		t.Fatal("fixture precondition: the deploy produced no .agents/skills entries — the fixture cannot show restoration")
	}
	return root
}

// mirrorHealStampedProject builds a project with a hand-written stamp and no
// deploy. Used only where the gate value itself is the subject (AC-UMH-003 /
// AC-UMH-004) or where the fixture must model a deploy that never produced a
// skill tree (AC-UMH-015 / AC-UMH-016).
func mirrorHealStampedProject(t *testing.T, stamp string) string {
	t.Helper()
	root := t.TempDir()
	if stamp != "" {
		writeTestFile(t, root, ".moai/config/sections/system.yaml",
			fmt.Sprintf("moai:\n  template_version: %q\n", stamp))
	}
	writeTestFile(t, root, ".moai/manifest.json", "{}\n")
	return root
}

// ---------------------------------------------------------------------------
// Observation helpers
// ---------------------------------------------------------------------------

// snapshotTree records every path under dir as
// "relpath|type|size|symlink-target". It is the portable filesystem witness
// this SPEC uses in place of git, which cannot see .agents/ at all because
// that path is gitignored (acceptance.md §D.0 rule 1).
func mirrorHealSnapshot(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && p == dir {
				return nil
			}
			return err
		}
		rel, relErr := filepath.Rel(dir, p)
		if relErr != nil {
			return relErr
		}
		info, lerr := os.Lstat(p)
		if lerr != nil {
			return lerr
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			out = append(out, fmt.Sprintf("%s|symlink|0|%s", rel, target))
		case info.IsDir():
			out = append(out, fmt.Sprintf("%s|dir|0|", rel))
		default:
			// ModTime is load-bearing, not decoration. atomicWriteFile
			// write-then-renames, so a file rewritten with IDENTICAL bytes
			// is invisible to path+size alone — and "rewrote a file it was
			// supposed to leave alone" is exactly the forbidden act
			// AC-UMH-006 guards. Portable (no syscall.Stat_t), so the guard
			// still compiles on windows.
			out = append(out, fmt.Sprintf("%s|file|%d|mtime=%d", rel, info.Size(), info.ModTime().UnixNano()))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

func mirrorHealSnapshotDiff(before, after []string) []string {
	seen := map[string]int{}
	for _, b := range before {
		seen[b]++
	}
	for _, a := range after {
		seen[a]--
	}
	var diff []string
	for k, v := range seen {
		if v != 0 {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return diff
}

// mirrorEntryNames lists the entry names directly under .agents/skills.
func mirrorEntryNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, ".agents", "skills"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read mirror dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// runRepairAt is the call-site seam under test.
func runRepairAt(t *testing.T, root string) (stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	repairSkillMirrorBestEffortAt(root, &out, &errOut)
	return out.String(), errOut.String()
}

// embeddedPublishedSKILLs returns the deploy-relative published SKILL.md paths
// derived from the PRODUCTION name set — never hand-typed (AC-UMH-002).
func embeddedPublishedSKILLs(t *testing.T) map[string][]byte {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	got := map[string][]byte{}
	for _, name := range template.PublishedSkillNames() {
		rel := ".agents/skills/" + name + "/SKILL.md"
		data, readErr := fs.ReadFile(fsys, rel)
		if readErr != nil {
			t.Fatalf("embedded %s: %v", rel, readErr)
		}
		got[rel] = data
	}
	return got
}

// ---------------------------------------------------------------------------
// AC-UMH-001 — Path A restoration
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_RestoresPathA(t *testing.T) {
	root := mirrorHealDeployedProject(t)

	// The set the deploy produced — the equality target. Path B publishes
	// real directories into the same namespace, so Path A's own set is the
	// symlink/copy subset: everything the deploy mirrored.
	before := mirrorEntryNames(t, root)

	if err := os.RemoveAll(filepath.Join(root, ".agents")); err != nil {
		t.Fatalf("delete mirror: %v", err)
	}

	runRepairAt(t, root)

	after := mirrorEntryNames(t, root)
	if diff := mirrorHealSnapshotDiff(before, after); len(diff) != 0 {
		t.Errorf("restored mirror entry set != pre-deletion set; symmetric difference: %v", diff)
	}
	if len(after) == 0 {
		t.Fatal("repair restored no entries at all")
	}
	// No dangling entries: every entry must resolve.
	for _, name := range after {
		p := filepath.Join(root, ".agents", "skills", name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("entry %s does not resolve (dangling): %v", name, err)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-002 — Path B restoration
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_RestoresPathB(t *testing.T) {
	root := mirrorHealDeployedProject(t)
	want := embeddedPublishedSKILLs(t)
	if len(want) != 16 {
		t.Fatalf("production published set is %d names, expected 16", len(want))
	}

	if err := os.RemoveAll(filepath.Join(root, ".agents")); err != nil {
		t.Fatalf("delete mirror: %v", err)
	}
	runRepairAt(t, root)

	for rel, wantBytes := range want {
		gotBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("published artifact not restored: %s: %v", rel, err)
			continue
		}
		if !bytes.Equal(gotBytes, wantBytes) {
			t.Errorf("published artifact %s restored with %d bytes, embedded has %d", rel, len(gotBytes), len(wantBytes))
		}
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-003 / AC-UMH-004 — the gate-closed no-ops
// ---------------------------------------------------------------------------

// assertSnapshotHelperSeesChange is the mandatory control for the two
// absence-as-PASS criteria below: an identical snapshot proves nothing unless
// the same helper is shown able to report a difference at all.
func assertSnapshotHelperSeesChange(t *testing.T) {
	t.Helper()
	root := mirrorHealDeployedProject(t)
	if err := os.RemoveAll(filepath.Join(root, ".agents")); err != nil {
		t.Fatalf("delete mirror: %v", err)
	}
	before := mirrorHealSnapshot(t, root)
	runRepairAt(t, root)
	after := mirrorHealSnapshot(t, root)
	if len(mirrorHealSnapshotDiff(before, after)) == 0 {
		t.Fatal("CONTROL FAILED: the snapshot helper reported no difference across a run that demonstrably repairs — an identical snapshot elsewhere would be meaningless")
	}
}

func TestUpdateMirrorHeal_NoCreateBelowStamp(t *testing.T) {
	assertSnapshotHelperSeesChange(t)

	root := mirrorHealStampedProject(t, "3.1.2")
	writeTestFile(t, root, ".claude/skills/moai-workflow-tdd/SKILL.md", "x\n")

	before := mirrorHealSnapshot(t, root)
	runRepairAt(t, root)
	after := mirrorHealSnapshot(t, root)

	if _, err := os.Stat(filepath.Join(root, ".agents")); !os.IsNotExist(err) {
		t.Errorf(".agents exists after a below-stamp run (err=%v) — the pass created it in a project the feature never targeted", err)
	}
	if diff := mirrorHealSnapshotDiff(before, after); len(diff) != 0 {
		t.Errorf("below-stamp run changed the project; diff: %v", diff)
	}
}

func TestUpdateMirrorHeal_NoCreateWithoutStamp(t *testing.T) {
	assertSnapshotHelperSeesChange(t)

	root := mirrorHealStampedProject(t, "") // no system.yaml at all → "0.0.0"
	writeTestFile(t, root, ".claude/skills/moai-workflow-tdd/SKILL.md", "x\n")

	stamp, err := plan.GetProjectConfigVersion(root)
	if err != nil {
		t.Fatalf("GetProjectConfigVersion: %v", err)
	}
	if stamp != "0.0.0" {
		t.Fatalf("fixture precondition: expected the absent-stamp degradation %q, got %q", "0.0.0", stamp)
	}

	before := mirrorHealSnapshot(t, root)
	runRepairAt(t, root)
	after := mirrorHealSnapshot(t, root)

	if _, err := os.Stat(filepath.Join(root, ".agents")); !os.IsNotExist(err) {
		t.Errorf(".agents exists after an unstamped run (err=%v)", err)
	}
	if diff := mirrorHealSnapshotDiff(before, after); len(diff) != 0 {
		t.Errorf("unstamped run changed the project; diff: %v", diff)
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-005 — the early return is preserved
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_EarlyReturnPreserved(t *testing.T) {
	// Source half: the optimization is still there, in its present shape.
	src, err := os.ReadFile("update_template_sync.go")
	if err != nil {
		t.Fatalf("read update_template_sync.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error)") {
		t.Error("runTemplateSyncWithProgress lost its (skipped bool, err error) contract (C-2 / REQ-UMH-003)")
	}
	sigIdx := strings.Index(body, "packageVersion == projectVersion && !forceUpdate")
	if sigIdx < 0 {
		t.Fatal("version-match guard not found — the early return was removed or rewritten (C-2)")
	}
	tail := body[sigIdx:]
	retIdx := strings.Index(tail, "return true, nil")
	callIdx := strings.Index(tail, "runTemplateSyncWithReporter(")
	if retIdx < 0 {
		t.Fatal("version-match branch no longer returns (true, nil)")
	}
	if callIdx >= 0 && callIdx < retIdx {
		t.Error("runTemplateSyncWithReporter is now reached before the version-match return — the optimization was weakened")
	}

	// Behavioral half: a version-matched run writes no template file AND the
	// repair pass still restores the mirror. The source read alone would pass
	// on a tree where the guard exists but is unreachable.
	root := mirrorHealDeployedProject(t)
	if err := os.RemoveAll(filepath.Join(root, ".agents")); err != nil {
		t.Fatalf("delete mirror: %v", err)
	}
	beforeSync := mirrorHealSnapshot(t, root)

	origDir, _ := os.Getwd()
	if chErr := os.Chdir(root); chErr != nil {
		t.Fatalf("chdir: %v", chErr)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	cmd := &cobra.Command{Use: "update-test"}
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("force", false, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	skipped, syncErr := runTemplateSyncWithProgress(cmd)
	if syncErr != nil {
		t.Fatalf("version-matched sync errored: %v", syncErr)
	}
	if !skipped {
		t.Fatal("fixture precondition: the sync did not take the version-match branch")
	}
	if diff := mirrorHealSnapshotDiff(beforeSync, mirrorHealSnapshot(t, root)); len(diff) != 0 {
		t.Errorf("the version-matched sync wrote files (the optimization is gone); diff: %v", diff)
	}

	runRepairAt(t, root)
	if len(mirrorEntryNames(t, root)) == 0 {
		t.Error("the repair pass restored nothing after a version-matched sync — repair does not run beside the optimization")
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-006 — healthy project is a no-op
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_HealthyIsNoop(t *testing.T) {
	root := mirrorHealDeployedProject(t)
	agents := filepath.Join(root, ".agents")

	before := mirrorHealSnapshot(t, agents)
	if len(before) == 0 {
		t.Fatal("fixture precondition: .agents is empty, so an identical snapshot would assert nothing")
	}
	runRepairAt(t, root)
	after := mirrorHealSnapshot(t, agents)

	if diff := mirrorHealSnapshotDiff(before, after); len(diff) != 0 {
		t.Errorf("repair changed a healthy mirror; diff: %v", diff)
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-007 — a non-symlink occupant is left untouched
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_SkipsForeignOccupant(t *testing.T) {
	root := mirrorHealDeployedProject(t)

	// Pick a canonical (Path A) entry and replace it with the user's bytes.
	var victim string
	for _, name := range mirrorEntryNames(t, root) {
		p := filepath.Join(root, ".agents", "skills", name)
		if info, err := os.Lstat(p); err == nil && info.Mode()&os.ModeSymlink != 0 {
			victim = name
			break
		}
	}
	if victim == "" {
		t.Skip("no symlink mirror entry on this host (copy fallback) — the occupancy branch is unreachable here")
	}
	occupied := filepath.Join(root, ".agents", "skills", victim)
	if err := os.RemoveAll(occupied); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(occupied, 0o755); err != nil {
		t.Fatal(err)
	}
	const userBytes = "user-owned content, must survive\n"
	if err := os.WriteFile(filepath.Join(occupied, "MINE.md"), []byte(userBytes), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr := runRepairAt(t, root)

	got, err := os.ReadFile(filepath.Join(occupied, "MINE.md"))
	if err != nil {
		t.Fatalf("user file gone after repair: %v", err)
	}
	if string(got) != userBytes {
		t.Errorf("user bytes rewritten: %q", string(got))
	}
	info, err := os.Lstat(occupied)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("the repair replaced a non-symlink occupant with a symlink")
	}
	if !strings.Contains(stderr, victim) {
		t.Errorf("the skip was not reported; stderr: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-008 — repair scope is template-shipped skills only
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_ScopeExcludesLocalSkills(t *testing.T) {
	root := mirrorHealDeployedProject(t)
	writeTestFile(t, root, ".claude/skills/local-only-skill/SKILL.md", "local\n")

	if err := os.RemoveAll(filepath.Join(root, ".agents")); err != nil {
		t.Fatalf("delete mirror: %v", err)
	}
	runRepairAt(t, root)

	names := mirrorEntryNames(t, root)
	for _, n := range names {
		if n == "local-only-skill" {
			t.Error("the repair mirrored a locally-authored skill no deploy produced (scope S1, not S2)")
		}
	}
	// Positive control: a pass that mirrors nothing at all must not satisfy
	// the negative half above.
	if len(names) == 0 {
		t.Fatal("CONTROL FAILED: the repair created no entries, so the absence of local-only-skill asserts nothing")
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-009 — fail-open
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_FailOpen(t *testing.T) {
	root := mirrorHealDeployedProject(t)
	agents := filepath.Join(root, ".agents")
	if err := os.RemoveAll(agents); err != nil {
		t.Fatal(err)
	}
	// A regular FILE at .agents makes every MkdirAll under it fail — the
	// OS-level injection, no production seam required.
	if err := os.WriteFile(agents, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr := runRepairAt(t, root)

	if stderr == "" {
		t.Error("mirror creation failed silently — no warning reached stderr (REQ-UMH-006)")
	}
	info, err := os.Lstat(agents)
	if err != nil {
		t.Fatalf(".agents disappeared: %v", err)
	}
	if info.IsDir() {
		t.Error("the repair removed the user's .agents file to make room for itself")
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-010 — the gate constant is grounded
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_GateConstantGrounded(t *testing.T) {
	if template.MirrorIntroducedVersion != "3.1.3" {
		t.Errorf("MirrorIntroducedVersion = %q, want %q (CHANGELOG [3.1.3] - 2026-08-24)", template.MirrorIntroducedVersion, "3.1.3")
	}
	// The constant lives next to the producer whose behavior it describes
	// (plan.md M1 decision), so the citation check reads that file.
	src, err := os.ReadFile(filepath.Join("..", "template", "skill_mirror_repair.go"))
	if err != nil {
		t.Fatalf("read constant file: %v", err)
	}
	body := string(src)
	// Anchor on the DECLARATION, not the first mention: the doc comment opens
	// with the constant's own name, so anchoring on the first occurrence
	// would cut the head before the citations it is supposed to check.
	idx := strings.Index(body, "const MirrorIntroducedVersion")
	if idx < 0 {
		t.Fatal("MirrorIntroducedVersion not declared in skill_mirror_repair.go")
	}
	head := body[:idx]
	for _, cite := range []string{"CHANGELOG.md", "[3.1.3] - 2026-08-24", "9c94c6b7a"} {
		if !strings.Contains(head, cite) {
			t.Errorf("the constant's doc comment does not cite %q — the value would be a guess on the record", cite)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-011 — doctor guidance reconciled
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_DoctorGuidance(t *testing.T) {
	problems, _ := codexMirrorObservations(skillMirrorState{dirPresent: false})
	if len(problems) != 1 {
		t.Fatalf("absent-mirror state produced %d findings, want 1", len(problems))
	}
	detail := problems[0].detail
	if strings.Contains(detail, "does not restore it") {
		t.Errorf("the doctor still asserts a routine update cannot restore the mirror — false after this card (REQ-UMH-008):\n%s", detail)
	}
	// Controls: the detail is still the mirror-absent detail, so the absence
	// above is a measurement rather than a selector aimed at the wrong text.
	if !strings.Contains(detail, "does not scan .claude/skills") {
		t.Errorf("the mirror-absent detail lost its explaining clause:\n%s", detail)
	}
	if !strings.Contains(problems[0].summary, "mirror absent") {
		t.Errorf("the mirror-absent finding lost its summary:\n%s", problems[0].summary)
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-012 — the doctor stays read-only
// ---------------------------------------------------------------------------

var mirrorHealWriteCalls = []string{
	"os.Mkdir", "os.MkdirAll", "os.WriteFile", "os.Remove", "os.Symlink", "os.Create",
}

func countWriteCalls(body string) int {
	n := 0
	for _, call := range mirrorHealWriteCalls {
		n += strings.Count(body, call)
	}
	return n
}

func TestUpdateMirrorHeal_DoctorRemainsReadOnly(t *testing.T) {
	src, err := os.ReadFile("doctor_codex.go")
	if err != nil {
		t.Fatalf("read doctor_codex.go: %v", err)
	}
	body := string(src)
	start := strings.Index(body, "func inspectSkillMirror(")
	if start < 0 {
		t.Fatal("inspectSkillMirror not found — test premise stale")
	}
	end := strings.Index(body[start:], "\n// codexMirrorObservations")
	if end < 0 {
		t.Fatal("codexMirrorObservations not found after inspectSkillMirror — test premise stale")
	}
	span := body[start : start+end]
	if n := countWriteCalls(span); n != 0 {
		t.Errorf("the doctor mirror span contains %d write call(s) — C-4 says the doctor reports and never repairs", n)
	}

	// Control: the same selector, over the repair pass's own source, must be
	// non-zero — otherwise the zero above could be a selector that matches
	// nothing anywhere.
	repairSrc, err := os.ReadFile(filepath.Join("..", "template", "skill_mirror_repair.go"))
	if err != nil {
		t.Fatalf("read repair pass source: %v", err)
	}
	if n := countWriteCalls(string(repairSrc)); n == 0 {
		t.Fatal("CONTROL FAILED: the write-call selector matches nothing in the repair pass either, so the doctor's zero asserts nothing")
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-015 / AC-UMH-016 — the stamped-but-partial boundary (spec §3.6)
// ---------------------------------------------------------------------------

// mirrorHealPartialProject models the state §3.6 accepts as residual risk: a
// deploy that wrote the stamp inside the walk and then FAILED, leaving no
// .claude/skills and no mirror. It is built by running a real deploy whose
// renderer fails on a file sorting after the stamp — not by hand-assembling
// the end state — so the reachability of the state is itself demonstrated.
// mirrorHealFailingRenderer fails every render, which turns any .tmpl file in
// the fixture FS into a walk-aborting deploy error.
type mirrorHealFailingRenderer struct{}

func (mirrorHealFailingRenderer) Render(string, any) ([]byte, error) {
	return nil, fmt.Errorf("injected render failure (partial-deploy fixture)")
}

func mirrorHealPartialProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	stamp := version.GetVersion()
	// fs.WalkDir visits in lexical order, so ".zz-late-file.tmpl" is reached
	// AFTER the stamp is written — the walk fails with the stamp already on
	// disk, which is exactly the §3.6 state.
	fsys := fstest.MapFS{
		".moai/config/sections/system.yaml": &fstest.MapFile{
			Data: []byte(fmt.Sprintf("moai:\n  template_version: %q\n", stamp)),
		},
		".zz-late-file.tmpl": &fstest.MapFile{Data: []byte("late")},
	}
	dep := template.NewDeployerWithRenderer(fsys, mirrorHealFailingRenderer{})
	err := dep.Deploy(t.Context(), root, newManifestManagerForTest(t, root), &template.TemplateContext{})
	if err == nil {
		t.Fatal("fixture precondition: the deploy was meant to FAIL after writing the stamp — it succeeded, so this fixture does not model a partial deploy")
	}

	got, verr := plan.GetProjectConfigVersion(root)
	if verr != nil || got != stamp {
		t.Fatalf("fixture precondition: stamp not written by the failed deploy (got %q, err %v)", got, verr)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".claude", "skills")); !os.IsNotExist(statErr) {
		t.Fatalf("fixture precondition: .claude/skills exists (err=%v) — this is not a partial deploy", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".agents")); !os.IsNotExist(statErr) {
		t.Fatalf("fixture precondition: .agents exists — this is not a partial deploy")
	}
	return root
}

func TestUpdateMirrorHeal_PartialDeployPathA(t *testing.T) {
	root := mirrorHealPartialProject(t)
	runRepairAt(t, root)

	// Zero Path A entries: every candidate's canonical target is absent, so
	// REQ-UMH-010's filter drops all of them. Without that filter the
	// unmodified producer would create dangling symlinks here (os.Symlink
	// succeeds against a missing target).
	symlinks := 0
	dangling := 0
	for _, name := range mirrorEntryNames(t, root) {
		p := filepath.Join(root, ".agents", "skills", name)
		info, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			symlinks++
			if _, statErr := os.Stat(p); statErr != nil {
				dangling++
			}
		}
	}
	if symlinks != 0 {
		t.Errorf("the repair created %d Path A entries on a project with no .claude/skills", symlinks)
	}
	if dangling != 0 {
		t.Errorf("the repair created %d dangling entries — damage, not repair (a moai doctor finding in its own right)", dangling)
	}
}

func TestUpdateMirrorHeal_PartialDeployPathB(t *testing.T) {
	root := mirrorHealPartialProject(t)
	stdout, _ := runRepairAt(t, root)

	want := embeddedPublishedSKILLs(t)
	for rel := range want {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("accepted residual risk not delivered: %s missing: %v", rel, err)
		}
	}
	if stdout == "" {
		t.Error("the run restored 16 published files and reported nothing (AC-UMH-016: the behavior is asserted, not silent)")
	}
}

// ---------------------------------------------------------------------------
// AC-UMH-017 — version-string degradation matrix
// ---------------------------------------------------------------------------

func TestUpdateMirrorHeal_VersionMatrix(t *testing.T) {
	cases := []struct {
		stamp string
		open  bool
		why   string
	}{
		{"3.1.3", true, "equal to the constant"},
		{"v3.1.3", true, "v prefix stripped"},
		{"3.2.0-rc.0", true, "pre-release suffix truncated"},
		{"3.1.2", false, "below the constant"},
		{"dev", false, "parses to [0,0,0]"},
	}
	for _, tc := range cases {
		t.Run(tc.stamp, func(t *testing.T) {
			if got := mirrorRepairGateOpen(tc.stamp); got != tc.open {
				t.Errorf("gate for %q = %v, want %v (%s)", tc.stamp, got, tc.open, tc.why)
			}
		})
	}
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		stamp, err := plan.GetProjectConfigVersion(root)
		if err != nil {
			t.Fatalf("GetProjectConfigVersion: %v", err)
		}
		if stamp != "0.0.0" {
			t.Fatalf("absent-stamp degradation = %q, want 0.0.0", stamp)
		}
		if mirrorRepairGateOpen(stamp) {
			t.Error("gate open on an absent stamp — the pass would run outside a MoAI project")
		}
	})
}
