package cli

// migrate_ownership_preview_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M7 (card
// t1865), AC-OWR-16 and REQ-OWR-011: `moai update --dry-run` — every step it runs,
// the ownership preview included — creates, modifies and deletes nothing, and the
// removal candidates it announces are exactly the files the verdict function marks
// `removal-eligible`.
//
// The whole dispatch is driven (runUpdate through runDryRunUpdateInProject), not
// one preview step. Every directory mtime is pinned to a fixed past time first and
// folded into the tree digest, so a file created and deleted inside a directory
// (the lock file, a probe file) moves that directory off its pin and is seen.
//
// The oracle for "removal-eligible" is two independent computations: the paths the
// REAL migration removes from a copy of the fixture (switch on), and the verdict
// function run directly over the fixture. The preview's announced set must equal
// both.
//
// Time independence: one dry run per project; nothing here depends on the clock
// second a run lands in beyond the pinned past mtimes.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	upd "github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// The files of the AC-OWR-16 project, beside the migration fixture's own.
const (
	prevEligibleSkill = ".claude/skills/moai-foundation-quality/SKILL.md" // registered managed, counterpart installed
	prevEligibleAgent = ".claude/agents/moai/manager-spec.md"             // registered managed, counterpart installed
	prevTrapAgent     = ".claude/agents/moai/manager-git.md"              // registered managed, NO counterpart (optional pack)
	prevUnknownOrigin = ".agents/skills/moai-custom-x/SKILL.md"           // unregistered, unknown origin
)

// prevPastTime is the fixed past time every directory mtime is pinned to.
var prevPastTime = time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

var prevCandidateRe = regexp.MustCompile(`\[dry-run\] migration candidate: (\S+)`)

// previewFixture builds the AC-OWR-16 project: AC-OWR-02..05 files mixed with
// registered managed copies. mapped makes it a profile-mapped project whose
// managed copies carry the KeptBy mark (AC-OWR-07 shape). The removal switch is
// the caller's.
func previewFixture(t *testing.T, mapped bool) (root, home string) {
	t.Helper()
	ownProfileSandbox(t)
	home = installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root = buildMigrationFixture(t)
	cmdWriteProject(t, root, "claude", cmdOldVersion)
	for _, rel := range []string{prevEligibleSkill, prevEligibleAgent, prevTrapAgent} {
		ownAddRegisteredManaged(t, root, rel, ownEmbeddedBytes(t, rel))
	}
	ownAddUnregisteredFromTemplate(t, root, ownEmbeddedClaudeSkill)
	ownAddUnregisteredFromTemplate(t, root, ownEmbeddedAgentsSkill)
	ownAddUnregistered(t, root, prevUnknownOrigin, []byte("# unknown origin\n"))
	if mapped {
		mapProjectToWork(t, root)
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("load manifest: %v", err)
		}
		for _, rel := range []string{migIdenticalSkill, prevEligibleSkill, prevEligibleAgent, prevTrapAgent} {
			if err := mgr.SetKeptBy(rel, manifest.KeptByProfileReach); err != nil {
				t.Fatalf("mark %s: %v", rel, err)
			}
		}
		if err := mgr.Save(); err != nil {
			t.Fatalf("save manifest: %v", err)
		}
	}
	return root, home
}

// previewMigrationRemoved is the independent oracle: the paths the real migration
// removes from a copy of root (switch on), sorted. A mapped project's copy is
// mapped too, so both see the same profile reach.
func previewMigrationRemoved(t *testing.T, root, home string, mapped bool) []string {
	t.Helper()
	enableRemovalForTest(t)
	cp := t.TempDir()
	if err := copyDir(root, cp); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	if mapped {
		mapProjectToWork(t, cp)
	}
	run := runOwnershipMigration(t, cp, home)
	got := append([]string(nil), run.removed...)
	sort.Strings(got)
	return got
}

// previewVerdictSet is the second oracle: the verdict function run over every
// file under the four common-asset roots with the inputs the migration reads,
// assembled here independently of the preview code. Sorted.
func previewVerdictSet(t *testing.T, root, home string) []string {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	um, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatalf("load user manifest: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("absolute root: %v", err)
	}
	keysByDir := manifestKeysByDir(mgr)
	env := ownershipEnv{
		entry:    mgr.GetEntry,
		evidence: newProvenanceSource(root),
		counterpartConfirmed: func(rel string) bool {
			return counterpartConfirmedCaseAware(um, cat, embedded, home, rel)
		},
		notInCatalog:   func(rel string) bool { return skillNotInCatalog(cat, rel) },
		profileMapped:  profile.ResolveLaunchProfileForProject(rootAbs, "") != "",
		removalEnabled: removalEnabled,
		rootResolved:   rootResolved,
		counterpartStatus: func(rel string) counterpartStatus {
			return userCounterpartStatus(um, cat, embedded, home, rel)
		},
		caseAlias: func(rel string) (string, bool) {
			return resolveCaseAlias(hostCaseFS, root, rel, keysByDir[path.Dir(rel)])
		},
		shippedSpelling: func(rel string) string { return shippedCaseSpelling(embedded, rel) },
	}
	var got []string
	for _, rootRel := range projectCommonAssetRels {
		_ = filepath.WalkDir(filepath.Join(root, filepath.FromSlash(rootRel)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(root, p)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if classifyOwnership(env, rel, p).removalEligible {
				got = append(got, rel)
			}
			return nil
		})
	}
	sort.Strings(got)
	return got
}

// previewAnnounced extracts the announced candidate paths from the dry-run output.
func previewAnnounced(out string) []string {
	var got []string
	for _, m := range prevCandidateRe.FindAllStringSubmatch(ansiStrip.ReplaceAllString(out, ""), -1) {
		got = append(got, m[1])
	}
	sort.Strings(got)
	return got
}

// candidateSetProblems is the pure comparison: one problem per path in one set
// and not the other.
func candidateSetProblems(announced, want []string) []string {
	var problems []string
	for _, p := range announced {
		if !slices.Contains(want, p) {
			problems = append(problems, fmt.Sprintf("announced as a migration candidate but not removal-eligible: %s", p))
		}
	}
	for _, p := range want {
		if !slices.Contains(announced, p) {
			problems = append(problems, fmt.Sprintf("removal-eligible but not announced: %s", p))
		}
	}
	return problems
}

// legacyPreviewCandidate is the preview's former, private judgment, kept here
// only as the wrong implementation the comparison must catch: a template_managed
// record whose recorded template hash matches the bytes on disk, nothing else.
func legacyPreviewCandidate(root string, mf *manifest.Manifest, rel string) bool {
	if mf == nil {
		return false
	}
	entry, ok := mf.Files[rel]
	if !ok || entry.Provenance != manifest.TemplateManaged || entry.TemplateHash == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil && entry.TemplateHash == manifest.HashBytes(data)
}

// previewLegacySet applies the legacy judgment over the files a dry run would
// look at: every file under the common-asset roots.
func previewLegacySet(t *testing.T, root string) []string {
	t.Helper()
	mf := update.LoadManifestReadOnly(root)
	var got []string
	for _, rootRel := range projectCommonAssetRels {
		_ = filepath.WalkDir(filepath.Join(root, filepath.FromSlash(rootRel)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(root, p)
			if relErr == nil && legacyPreviewCandidate(root, mf, filepath.ToSlash(rel)) {
				got = append(got, filepath.ToSlash(rel))
			}
			return nil
		})
	}
	sort.Strings(got)
	return got
}

// previewUserOwned returns the reconcile plan's user-owned paths for root.
func previewUserOwned(t *testing.T, root string) map[string]bool {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	plan, err := update.ClassifyManagedRoots(root,
		reconcileTargets(root, computeRunCleanTargets(root, resolveUpdateDeployMode(root, false), nil)),
		templateRenderCarriage(embedded, nil, nil), update.LoadManifestReadOnly(root))
	if err != nil {
		t.Fatalf("classify managed roots: %v", err)
	}
	out := map[string]bool{}
	for _, f := range plan.UserOwned {
		out[f.RelPath] = true
	}
	return out
}

// previewDryRun pins every directory mtime, runs the whole dry-run dispatch, and
// returns its output and the digest differences (paths whose descriptor moved).
func previewDryRun(t *testing.T, root string) (out string, diff []string) {
	t.Helper()
	pinTreeDirMtimes(t, root, prevPastTime)
	opts := ownTreeOpts{DirMtime: true}
	before := ownTreeSnapshot(t, root, opts)
	out = runDryRunUpdateInProject(t, root)
	after := ownTreeSnapshot(t, root, opts)
	return out, ownTreeDiff(before, after)
}

// TestDryRunChangesNothingAndMatchesVerdict is AC-OWR-16.
func TestDryRunChangesNothingAndMatchesVerdict(t *testing.T) {
	t.Run("UnmappedProjectAnnouncesExactlyTheEligibleSet", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := previewFixture(t, false)
		removed := previewMigrationRemoved(t, root, home, false)
		verdict := previewVerdictSet(t, root, home)
		// Reach (rule 8 shape): the oracles agree and are not empty, and every
		// eligible path is one the dry run's reconcile plan lists as user-owned —
		// the only files the preview looks at. A gap here is a measurement, not a
		// reason to narrow the comparison.
		if len(verdict) < 3 {
			t.Fatalf("the verdict function marks %v removal-eligible; the fixture should make at least three files eligible", verdict)
		}
		if !slices.Equal(removed, verdict) {
			t.Fatalf("the real migration removed %v but the verdict function marks %v removal-eligible; the oracles disagree", removed, verdict)
		}
		owned := previewUserOwned(t, root)
		for _, rel := range verdict {
			if !owned[rel] {
				t.Errorf("removal-eligible path %s is not in the dry run's reconcile plan as user-owned: the preview cannot announce it", rel)
			}
		}
		if len(owned) == 0 {
			t.Fatalf("the reconcile plan lists no user-owned file; the preview would print nothing")
		}

		out, diff := previewDryRun(t, root)
		if !strings.Contains(out, "Dry-run: reconciliation preview") {
			t.Fatalf("the dry run printed no reconciliation preview; the comparison would be vacuous:\n%s", out)
		}
		for _, p := range candidateSetProblems(previewAnnounced(out), verdict) {
			t.Errorf("preview vs verdict: %s", p)
		}
		if len(diff) != 0 {
			t.Errorf("the dry run changed the tree (paths whose bytes, mode or directory mtime moved): %v", diff)
		}
		if _, err := os.Lstat(filepath.Join(root, ".moai-backups")); err == nil {
			t.Errorf("the dry run created .moai-backups")
		}
	})

	t.Run("ProfileMappedKeptByProjectAnnouncesNothing", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := previewFixture(t, true)
		removed := previewMigrationRemoved(t, root, home, true)
		verdict := previewVerdictSet(t, root, home)
		if len(removed) != 0 || len(verdict) != 0 {
			t.Fatalf("a profile-mapped project keeps its shipped copies: migration removed %v, verdict marks %v", removed, verdict)
		}
		// The trap is live: the preview's former judgment announces these.
		legacy := previewLegacySet(t, root)
		if !slices.Contains(legacy, prevEligibleSkill) {
			t.Fatalf("the former judgment would not announce %s; the case would prove nothing (legacy set %v)", prevEligibleSkill, legacy)
		}
		out, diff := previewDryRun(t, root)
		if !strings.Contains(out, "Dry-run: reconciliation preview") {
			t.Fatalf("the dry run printed no reconciliation preview; the comparison would be vacuous:\n%s", out)
		}
		for _, p := range candidateSetProblems(previewAnnounced(out), verdict) {
			t.Errorf("preview vs verdict: %s", p)
		}
		if len(diff) != 0 {
			t.Errorf("the dry run changed the tree: %v", diff)
		}
	})

	t.Run("SwitchOffAnnouncesNothing", func(t *testing.T) {
		root, _ := previewFixture(t, false)
		out, diff := previewDryRun(t, root)
		if got := previewAnnounced(out); len(got) != 0 {
			t.Errorf("the removal switch is off, yet the preview announced %v", got)
		}
		if len(diff) != 0 {
			t.Errorf("the dry run changed the tree: %v", diff)
		}
	})

	t.Run("MutantOwnJudgmentIsCaught", func(t *testing.T) {
		enableRemovalForTest(t)
		root, home := previewFixture(t, false)
		verdict := previewVerdictSet(t, root, home)
		got := candidateSetProblems(previewLegacySet(t, root), verdict)
		if !slices.ContainsFunc(got, func(p string) bool { return strings.Contains(p, prevTrapAgent) }) {
			t.Errorf("the comparison did not report the file whose counterpart is unconfirmed (%s); problems: %v", prevTrapAgent, got)
		}
		mroot, mhome := previewFixture(t, true)
		mgot := candidateSetProblems(previewLegacySet(t, mroot), previewVerdictSet(t, mroot, mhome))
		if !slices.ContainsFunc(mgot, func(p string) bool { return strings.Contains(p, prevEligibleSkill) }) {
			t.Errorf("the comparison did not report the profile-mapped copy %s; problems: %v", prevEligibleSkill, mgot)
		}
	})

	t.Run("MutantCreateThenDeleteIsSeenByTheDigest", func(t *testing.T) {
		root, _ := previewFixture(t, false)
		pinTreeDirMtimes(t, root, prevPastTime)
		opts := ownTreeOpts{DirMtime: true}
		before := ownTreeSnapshot(t, root, opts)
		tmp := filepath.Join(root, ".claude", "skills", ".probe")
		if err := os.WriteFile(tmp, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(tmp); err != nil {
			t.Fatal(err)
		}
		diff := ownTreeDiff(before, ownTreeSnapshot(t, root, opts))
		if !slices.Contains(diff, ".claude/skills/") {
			t.Errorf("a file created and deleted in .claude/skills left the digest unchanged; diff %v", diff)
		}
	})
}

// TestDryRunRunsNoBinaryUpdateStep: REQ-OWR-011 covers every step the command
// runs. On a release build the binary self-update step comes before the dry-run
// branch; a dry run must not reach the installer. The fake orchestrator refuses
// to install (so nothing can re-exec) and counts calls.
func TestDryRunRunsNoBinaryUpdateStep(t *testing.T) {
	t.Setenv("MOAI_SKIP_BINARY_UPDATE", "")
	_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")
	ownProfileSandbox(t)
	root := buildMigrationFixture(t)
	cmdWriteProject(t, root, "claude", cmdBinaryVersion)
	cmdPinVersion(t, cmdBinaryVersion)
	installs := 0
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{isUpdateAvailFunc: func(string) (bool, *upd.VersionInfo, error) {
			return true, &upd.VersionInfo{Version: "v100.0.0"}, nil
		}},
		UpdateOrch: &mockUpdateOrchestrator{updateFunc: func(context.Context) (*upd.UpdateResult, error) {
			installs++
			return nil, errors.New("test: refuse to install")
		}},
	}
	snapChdir(t, root)
	var buf bytes.Buffer
	cmd := newDryRunUpdateCmd(&buf)
	cmd.SetContext(context.Background())
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("runUpdate --dry-run: %v\n%s", err, buf.String())
	}
	if installs != 0 {
		t.Errorf("a dry run called the binary installer %d time(s)", installs)
	}
}

// previewWritableOS are the os functions that create, modify or delete a file or
// directory.
var previewWritableOS = map[string]bool{
	"WriteFile": true, "Create": true, "CreateTemp": true, "Mkdir": true, "MkdirAll": true,
	"MkdirTemp": true, "Remove": true, "RemoveAll": true, "Rename": true, "Truncate": true,
	"Chmod": true, "Chtimes": true, "Symlink": true, "Link": true, "OpenFile": true,
}

// previewWritableMethods are manifest-manager methods that change state.
var previewWritableMethods = map[string]bool{"Save": true, "Track": true, "SetKeptBy": true, "SetDeployingVersion": true}

// writableCalls returns one entry per call in src (filename for messages) that
// can write: an os function above, a manager mutation, or probeCaseSensitiveFS.
// funcs limits the scan to the named top-level functions; nil scans the file.
func writableCalls(filename, src string, funcs map[string]bool) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		return nil, err
	}
	var hits []string
	scan := func(n ast.Node, within string) {
		ast.Inspect(n, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				if id, ok := fn.X.(*ast.Ident); ok && id.Name == "os" && previewWritableOS[fn.Sel.Name] {
					hits = append(hits, fmt.Sprintf("%s: %s calls os.%s", filename, within, fn.Sel.Name))
				} else if previewWritableMethods[fn.Sel.Name] {
					hits = append(hits, fmt.Sprintf("%s: %s calls .%s", filename, within, fn.Sel.Name))
				}
			case *ast.Ident:
				if fn.Name == "probeCaseSensitiveFS" {
					hits = append(hits, fmt.Sprintf("%s: %s calls probeCaseSensitiveFS", filename, within))
				}
			}
			return true
		})
	}
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		if funcs != nil && !funcs[fd.Name.Name] {
			continue
		}
		scan(fd.Body, fd.Name.Name)
	}
	return hits, nil
}

// TestVerdictAndPreviewCodeHasNoWritableCalls: the code this SPEC adds for the
// verdict and the preview reads and never writes. Whole files where the file is
// read-only by purpose; named functions where the file also holds a writer.
func TestVerdictAndPreviewCodeHasNoWritableCalls(t *testing.T) {
	wholeFiles := []string{
		"migrate_ownership.go", "migrate_ownership_evidence.go", "migrate_ownership_case.go",
		"migrate_ownership_env.go", "migrate_ownership_remaining.go", "update_dryrun_preview.go",
	}
	named := map[string]map[string]bool{
		"migrate_project_assets.go":     {"newOwnershipEnv": true},
		"migrate_ownership_emptydir.go": {"dirManagedProvenance": true, "startOfRunRecords": true, "emptyDirTree": true},
	}
	scanned := 0
	for _, name := range wholeFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		hits, err := writableCalls(name, string(data), nil)
		if err != nil {
			t.Errorf("parse %s: %v", name, err)
		}
		scanned++
		for _, h := range hits {
			t.Errorf("writable call in verdict/preview code: %s", h)
		}
	}
	for name, funcs := range named {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		hits, err := writableCalls(name, string(data), funcs)
		if err != nil {
			t.Errorf("parse %s: %v", name, err)
		}
		// The named functions must exist, or the scan covers nothing.
		f, _ := parser.ParseFile(token.NewFileSet(), name, data, 0)
		found := 0
		if f != nil {
			for _, decl := range f.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok && funcs[fd.Name.Name] {
					found++
				}
			}
		}
		if found != len(funcs) {
			t.Errorf("%s: found %d of the %d functions the scan names", name, found, len(funcs))
		}
		scanned++
		for _, h := range hits {
			t.Errorf("writable call in verdict/preview code: %s", h)
		}
	}
	if scanned != len(wholeFiles)+len(named) {
		t.Errorf("scanned %d files, want %d", scanned, len(wholeFiles)+len(named))
	}

	// The scan is shown red in front of each kind of write.
	for name, src := range map[string]string{
		"os.WriteFile":         `package p; import "os"; func f() { _ = os.WriteFile("x", nil, 0) }`,
		"os.Create":            `package p; import "os"; func f() { _, _ = os.Create("x") }`,
		"os.MkdirAll":          `package p; import "os"; func f() { _ = os.MkdirAll("x", 0) }`,
		"os.Remove":            `package p; import "os"; func f() { _ = os.Remove("x") }`,
		"os.Rename":            `package p; import "os"; func f() { _ = os.Rename("x", "y") }`,
		"manager Save":         `package p; func f(m interface{ Save() error }) { _ = m.Save() }`,
		"probeCaseSensitiveFS": `package p; func f() { probeCaseSensitiveFS("x") }`,
	} {
		hits, err := writableCalls("mutant.go", src, nil)
		if err != nil || len(hits) == 0 {
			t.Errorf("the scan stayed green in front of %s (hits %v, err %v)", name, hits, err)
		}
	}
	if hits, _ := writableCalls("clean.go", `package p; import "os"; func f() { _, _ = os.ReadFile("x"); _, _ = os.Stat("x") }`, nil); len(hits) != 0 {
		t.Errorf("the scan flags reads: %v", hits)
	}
}
