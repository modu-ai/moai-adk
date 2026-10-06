package update

// reconcile_test.go — SPEC-UPDATE-MIGRATION-001 M2 (card t1547): the RED
// hazard tests the fix must satisfy. Both release-blocking ACs bind to the
// exact test names below (acceptance.md:31 / :37); their RED captures land
// in progress.md §E.2 with all four §2.1 elements BEFORE M4.
//
// Each test drives the update pipeline's two phases over a t.TempDir fixture
// project:
//
//	ReconcileManagedPaths (the pre-deploy clean replacement) →
//	recSimulateDeploy (mirrors the deploy stage: the render is written over
//	every template-carried path — the clobber the merge phase must answer) →
//	ReconcileMerges (the post-deploy merge phase).
//
// Until M4 lands the real bodies, ReconcileManagedPaths delegates to the
// current wipe-first clean and ReconcileMerges is a no-op, so every test
// here runs RED against today's flow.

import (
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"gopkg.in/yaml.v3"
)

// recTmplFS builds an embedded-template FS stand-in carrying exactly the
// given relative slash paths.
func recTmplFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for rel, content := range files {
		fsys[rel] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

// recSimulateDeploy mirrors the deploy stage of the update flow: the render
// for every template-carried path is written over the project tree. The
// pipeline under test does not own the deployer (the cli wiring keeps the
// real one); this helper reproduces exactly the write the deployer makes so
// the merge phase is tested against the clobber it must answer.
func recSimulateDeploy(t *testing.T, root string, carried map[string]string) {
	t.Helper()
	for rel, content := range carried {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("deploy mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("deploy write %s: %v", rel, err)
		}
	}
}

// recRunUpdate drives both reconciliation phases the way the cli wiring
// sequences them, returning the completed summary.
func recRunUpdate(t *testing.T, root string, tmplFS fs.FS, render TemplateRender, mf *manifest.Manifest, carried map[string]string) ReconciliationSummary {
	t.Helper()
	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, render, mf, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, render, mf, ReconcileMergeOptions{}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}
	return summary
}

// summaryPreserves reports whether rel appears in the summary's preserved
// list (AC-UPM-020: the preserved file is listed, not just left on disk).
func summaryPreserves(s ReconciliationSummary, rel string) bool {
	for _, p := range s.Preserved {
		if p == rel {
			return true
		}
	}
	return false
}

// TestUpdate_LocalOnlyFileSurvives — AC-UPM-020 (RELEASE-BLOCKING). A
// tracked local-only file under a managed root survives an update over the
// fixture project byte-for-byte, listed as preserved, with no
// backup-restore step.
func TestUpdate_LocalOnlyFileSurvives(t *testing.T) {
	const localOnly = "# operator's dev-only rule — exists in no template\n"
	const rel = ".claude/rules/moai/dev-only-rule.md"
	root := newClassifyFixture(t, map[string]string{
		rel: localOnly,
	})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "# template render\n",
	}
	tmplFS := recTmplFS(carried)

	summary := recRunUpdate(t, root, tmplFS, renderWith(carried), nil, carried)

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("local-only file did not survive update: %v", err)
	}
	if string(data) != localOnly {
		t.Errorf("local-only file content = %q, want byte-identical %q", data, localOnly)
	}
	if !summaryPreserves(summary, rel) {
		t.Errorf("summary.Preserved = %v, want it to list %s", summary.Preserved, rel)
	}
}

// TestUpdate_GitStrategyValuesSurvive — AC-UPM-021 (RELEASE-BLOCKING).
// Operator-set VALUES on keys the template carries with neutral defaults
// survive an update verbatim: worktree_base_branch develop (template ""), and
// git_strategy.manual.workflow git-flow (template github-flow), with zero
// manual steps. The hazard is VALUE reversion — these assertions check the
// values, never key presence.
func TestUpdate_GitStrategyValuesSurvive(t *testing.T) {
	const userConfig = `git_strategy:
  worktree_base_branch: develop
  manual:
    workflow: git-flow
`
	const templateConfig = `git_strategy:
  worktree_base_branch: ""
  manual:
    workflow: github-flow
`
	const rel = ".moai/config/sections/git-strategy.yaml"
	root := newClassifyFixture(t, map[string]string{rel: userConfig})
	carried := map[string]string{rel: templateConfig}
	tmplFS := recTmplFS(carried)

	// The merge base is the PRIOR TEMPLATE RENDER (the template carried the
	// same neutral defaults before this update — nothing changed on theirs).
	base := func(baseRel string) ([]byte, bool) {
		if baseRel == rel {
			return []byte(templateConfig), true
		}
		return nil, false
	}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{Base: base}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("git-strategy.yaml did not survive update: %v", err)
	}
	var doc struct {
		GitStrategy struct {
			WorktreeBaseBranch string `yaml:"worktree_base_branch"`
			Manual             struct {
				Workflow string `yaml:"workflow"`
			} `yaml:"manual"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse post-update git-strategy.yaml: %v", err)
	}
	if doc.GitStrategy.WorktreeBaseBranch != "develop" {
		t.Errorf("worktree_base_branch = %q, want %q (VALUE reversion — the 2026-09-24 hazard)", doc.GitStrategy.WorktreeBaseBranch, "develop")
	}
	if doc.GitStrategy.Manual.Workflow != "git-flow" {
		t.Errorf("manual.workflow = %q, want %q (VALUE reversion — the 2026-09-24 hazard)", doc.GitStrategy.Manual.Workflow, "git-flow")
	}
	if !summaryPreserves(summary, rel) && !containsPath(summary.Merged, rel) {
		t.Errorf("summary neither preserved nor merged %s: %+v", rel, summary)
	}
}

func containsPath(paths []string, rel string) bool {
	for _, p := range paths {
		if p == rel {
			return true
		}
	}
	return false
}

// recRunGit runs one git command inside the fixture repo and returns stdout.
func recRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// TestUpdate_NoGitDeletionsForLocalOnlyFiles — AC-UPM-040. In a git-tracked
// fixture project, an update leaves no ` D` (deletion) status lines for
// local-only files: the manual §2.3 check, automated.
func TestUpdate_NoGitDeletionsForLocalOnlyFiles(t *testing.T) {
	const rel = ".claude/rules/moai/dev-only-rule.md"
	root := newClassifyFixture(t, map[string]string{
		rel: "local content\n",
	})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "# template render\n",
	}
	tmplFS := recTmplFS(carried)

	recRunGit(t, root, "init", "-q")
	recRunGit(t, root, "-c", "user.email=t@local", "-c", "user.name=t", "add", "-A")
	recRunGit(t, root, "-c", "user.email=t@local", "-c", "user.name=t", "commit", "-q", "-m", "fixture")

	summary := recRunUpdate(t, root, tmplFS, renderWith(carried), nil, carried)

	status := recRunGit(t, root, "status", "--porcelain")
	for _, line := range splitLines2(status) {
		// Porcelain deletion form: " D path" or "D  path" (the latter cannot
		// result from an update that never stages, but check both forms).
		if len(line) > 2 && (line[0] == 'D' || line[1] == 'D') {
			t.Errorf("update deleted a tracked file: porcelain line %q (summary: %+v)", line, summary)
		}
	}
}

// splitLines2 splits s into non-empty lines.
func splitLines2(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// TestUpdate_ConflictPreservesFileAndWritesSidecar — AC-UPM-030 (conflict
// arm). A user-modified mergeable file whose edit conflicts with the new
// render: the on-disk file stays byte-identical to pre-run, the render lands
// in a <path>.moai-new sidecar, and the summary carries the conflict.
func TestUpdate_ConflictPreservesFileAndWritesSidecar(t *testing.T) {
	const rel = ".claude/rules/moai/policy.json"
	const ours = "{\"keep\": \"user\", \"shared\": \"user-value\"}\n"
	const rendered = "{\"keep\": \"user\", \"shared\": \"template-value\"}\n"
	// The merge base holds a THIRD value: ours changed from it (user-value)
	// AND theirs changed from it (template-value) — the both-changed branch
	// is what produces the conflict.
	const baseJSON = "{\"keep\": \"user\", \"shared\": \"base-value\"}\n"
	root := newClassifyFixture(t, map[string]string{rel: ours})
	carried := map[string]string{rel: rendered}
	tmplFS := recTmplFS(carried)
	base := func(rel string) ([]byte, bool) {
		if rel == ".claude/rules/moai/policy.json" {
			return []byte(baseJSON), true
		}
		return nil, false
	}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{Base: base}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}

	// The on-disk file is byte-identical to the user's pre-run content.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("conflicted file did not survive: %v", err)
	}
	if string(data) != ours {
		t.Errorf("conflicted file content = %q, want the untouched user bytes %q", data, ours)
	}
	// The render landed in the .moai-new sidecar.
	sidecarPath := rel + ".moai-new"
	sideData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sidecarPath)))
	if err != nil {
		t.Fatalf("conflict sidecar missing: %v", err)
	}
	if string(sideData) != rendered {
		t.Errorf("sidecar content = %q, want the render %q", sideData, rendered)
	}
	// The summary names the conflict with both paths.
	if len(summary.Conflicts) != 1 {
		t.Fatalf("summary.Conflicts = %+v, want exactly 1", summary.Conflicts)
	}
	if summary.Conflicts[0].Path != rel || summary.Conflicts[0].Sidecar != sidecarPath || summary.Conflicts[0].Collision {
		t.Errorf("conflict record = %+v, want {Path: %s, Sidecar: %s, Collision: false}", summary.Conflicts[0], rel, sidecarPath)
	}
}

// TestUpdate_ConflictSidecarCollisionUsesFirstUnusedNumber — AC-UPM-030
// (collision arm). A pre-existing <path>.moai-new sibling is never
// overwritten: the sidecar lands at the first unused numbered sibling and
// the summary reports the collision.
func TestUpdate_ConflictSidecarCollisionUsesFirstUnusedNumber(t *testing.T) {
	const rel = ".claude/rules/moai/policy.json"
	const ours = "{\"keep\": \"user\", \"shared\": \"user-value\"}\n"
	const rendered = "{\"keep\": \"user\", \"shared\": \"template-value\"}\n"
	const baseJSON = "{\"keep\": \"user\", \"shared\": \"base-value\"}\n"
	const occupiedSibling = "<<leftover sidecar from a prior unresolved conflict>>\n"
	root := newClassifyFixture(t, map[string]string{
		rel:                 ours,
		rel + ".moai-new":   occupiedSibling,
		rel + ".moai-new.2": "<<also occupied>>\n",
	})
	carried := map[string]string{rel: rendered}
	tmplFS := recTmplFS(carried)
	base := func(rel string) ([]byte, bool) {
		if rel == ".claude/rules/moai/policy.json" {
			return []byte(baseJSON), true
		}
		return nil, false
	}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{Base: base}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}

	// The pre-existing siblings are byte-identical — never overwritten.
	for _, sibling := range []string{rel + ".moai-new", rel + ".moai-new.2"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sibling)))
		if err != nil {
			t.Fatalf("pre-existing sibling %s missing: %v", sibling, err)
		}
		want := occupiedSibling
		if sibling == rel+".moai-new.2" {
			want = "<<also occupied>>\n"
		}
		if string(data) != want {
			t.Errorf("sibling %s content = %q, want untouched %q", sibling, data, want)
		}
	}
	// The sidecar landed at the first unused number.
	sidecarPath := rel + ".moai-new.3"
	sideData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sidecarPath)))
	if err != nil {
		t.Fatalf("collision sidecar %s missing: %v", sidecarPath, err)
	}
	if string(sideData) != rendered {
		t.Errorf("collision sidecar content = %q, want the render", sideData)
	}
	if len(summary.Conflicts) != 1 || !summary.Conflicts[0].Collision || summary.Conflicts[0].Sidecar != sidecarPath {
		t.Errorf("conflict record = %+v, want collision at %s", summary.Conflicts, sidecarPath)
	}
}

// TestUpdate_StaleFileArchivedAndRemoved — AC-UPM-031. A file a prior
// template carried and the current one does not: copied into the migration
// archive root (layout preserved), removed from place, archive copy
// byte-identical, removal listed in the summary.
func TestUpdate_StaleFileArchivedAndRemoved(t *testing.T) {
	const rel = ".claude/rules/moai/old-rule.md"
	const staleContent = "# carried by a prior template, dropped by the current one\n"
	root := newClassifyFixture(t, map[string]string{rel: staleContent})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "# template render\n",
	}
	tmplFS := recTmplFS(carried)

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track stale file: %v", err)
	}

	summary := recRunUpdate(t, root, tmplFS, renderWith(carried), mgr.Manifest(), carried)

	// Removed from place.
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Errorf("stale file still in place (err=%v)", err)
	}
	// Archived byte-identical under the run-scoped reconciliation archive
	// root (tag/<timestamp>/<rel> — review finding 5: a re-update must never
	// overwrite the recovery copy a previous run took). The timestamp run
	// scope is located by glob under the tag root.
	tagRoot := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()))
	matches, globErr := filepath.Glob(filepath.Join(tagRoot, "*", filepath.FromSlash(rel)))
	if globErr != nil {
		t.Fatalf("glob archive: %v", globErr)
	}
	if len(matches) != 1 {
		t.Fatalf("archive copies under %s = %v, want exactly 1", tagRoot, matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read archive copy: %v", err)
	}
	if string(data) != staleContent {
		t.Errorf("archive copy = %q, want byte-identical %q", data, staleContent)
	}
	// The summary lists the removal.
	if !containsPath(summary.ArchivedRemoved, rel) {
		t.Errorf("summary.ArchivedRemoved = %v, want it to list %s", summary.ArchivedRemoved, rel)
	}
}

// TestClassifyHealthyRecordPristineContentRefreshes — card t1547 review
// finding 1: a file pristine as last deployed (healthy manifest record,
// content equals the TRACKED state) whose PRIOR render differs from the NEW
// render is template-owned — the template's own update must land. Routing it
// user-modified would send the pristine file through the merge path where a
// missing base reverts the template's update.
func TestClassifyHealthyRecordPristineContentRefreshes(t *testing.T) {
	const priorRender = "rules_dir: .moai/previous-template-rules\n"
	const newRender = "rules_dir: .moai/config/astgrep-rules\n"
	const rel = ".moai/config/sections/cache.yaml"
	root := newClassifyFixture(t, map[string]string{rel: priorRender})
	carried := map[string]string{rel: newRender}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	// The record's CurrentHash matches the on-disk (prior-render) bytes —
	// pristine as deployed.
	entry := mgr.Manifest().Files[rel]
	entry.CurrentHash = manifest.HashBytes([]byte(priorRender))
	mgr.Manifest().Files[rel] = entry

	targets := []deploy.CleanTarget{recTarget(root, ".moai/config")}
	plan, err := ClassifyManagedRoots(root, targets, renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	if got := plan.ClassOf(rel); got != ClassTemplateOwned {
		t.Errorf("pristine file with healthy record = %q, want %q (the template's update must land)", got, ClassTemplateOwned)
	}

	// The operator-edit case still routes user-modified: the recorded hash
	// no longer matches the edited bytes.
	const userEdit = "rules_dir: .moai/previous-template-rules\nextra: operator\n"
	root2 := newClassifyFixture(t, map[string]string{rel: userEdit})
	// A fresh manager bound to root2 — the first manager's record belongs to
	// the first fixture tree.
	mgr2 := manifest.NewManager()
	if _, err := mgr2.Load(root2); err != nil {
		t.Fatalf("load manifest 2: %v", err)
	}
	if err := mgr2.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track 2: %v", err)
	}
	entry2 := mgr2.Manifest().Files[rel]
	entry2.CurrentHash = manifest.HashBytes([]byte(priorRender)) // stale: file edited since
	mgr2.Manifest().Files[rel] = entry2
	targets2 := []deploy.CleanTarget{recTarget(root2, ".moai/config")}
	plan2, err := ClassifyManagedRoots(root2, targets2, renderWith(carried), mgr2.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots 2: %v", err)
	}
	if got := plan2.ClassOf(rel); got != ClassUserModified {
		t.Errorf("edited file with stale hash = %q, want %q", got, ClassUserModified)
	}
}

// TestUpdate_SymlinkedRootDisposedBeforeDeploy — card t1547 review finding
// 3: a symlinked managed root is disposed (the link removed, the target
// untouched) BEFORE the deploy stage, so the deploy's writes land in a real
// directory inside the project instead of following the link outside it.
// The wholesale clean this pipeline replaces removed link entries; leaving
// them would be a regression with an external-write blast radius.
func TestUpdate_SymlinkedRootDisposedBeforeDeploy(t *testing.T) {
	external := t.TempDir() // OUTSIDE the fixture project
	sentinel := filepath.Join(external, "canary.txt")
	if err := os.WriteFile(sentinel, []byte("untouched\n"), 0o644); err != nil {
		t.Fatalf("sentinel: %v", err)
	}
	root := newClassifyFixture(t, map[string]string{})
	// .claude/rules/moai is a symlink to the external directory.
	if err := os.MkdirAll(filepath.Join(root, ".claude", "rules"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, filepath.Join(root, ".claude", "rules", "moai")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "# template render\n",
	}
	tmplFS := recTmplFS(carried)

	recRunUpdate(t, root, tmplFS, renderWith(carried), nil, carried)

	// The path is no longer a link (the deploy recreated a real directory in
	// its place) — and the external target was never written through.
	info, err := os.Lstat(filepath.Join(root, ".claude", "rules", "moai"))
	if err != nil {
		t.Fatalf("post-deploy root missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("symlinked root survived the reconcile — the deploy would write through it")
	}
	data, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("external sentinel unreadable: %v", err)
	}
	if string(data) != "untouched\n" {
		t.Errorf("external sentinel polluted: %q — deploy wrote through the link", data)
	}
}

// TestUpdate_ArchiveRefusesSymlinkDestination — card t1547 review finding 4:
// an archive destination on (or under) a symlink is refused before any
// write, so the archive copy can neither pollute an external file nor strip
// the operator's file after a hijacked copy.
func TestUpdate_ArchiveRefusesSymlinkDestination(t *testing.T) {
	external := t.TempDir() // outside the fixture project
	sentinelPath := filepath.Join(external, "sentinel.txt")
	const sentinel = "external file — must stay byte-identical\n"
	if err := os.WriteFile(sentinelPath, []byte(sentinel), 0o644); err != nil {
		t.Fatalf("sentinel: %v", err)
	}
	const rel = ".claude/rules/moai/old-rule.md"
	root := newClassifyFixture(t, map[string]string{rel: "stale content\n"})
	// Stale requires manifest evidence: track the file as prior-template-
	// carried so the classifier routes it to the archive step at all.
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	// .moai/archive/files/update-migration is a symlink to the external dir:
	// the archive copy would land outside the project.
	archiveAbs := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()))
	if err := os.MkdirAll(filepath.Dir(archiveAbs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, archiveAbs); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	carried := map[string]string{".claude/rules/moai/update-note.md": "render\n"}

	summary, _, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), mgr.Manifest(), ReconcileOptions{})
	if err == nil {
		t.Fatalf("archive through a symlinked destination must be refused (summary: %+v)", summary)
	}
	// The operator's file is still in place — no hijacked write, no removal.
	if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); statErr != nil {
		t.Errorf("stale file vanished despite the refused archive: %v", statErr)
	}
	// The external sentinel is byte-identical.
	data, readErr := os.ReadFile(sentinelPath)
	if readErr != nil || string(data) != sentinel {
		t.Errorf("external sentinel polluted (%q, err=%v) — the archive write escaped the project", data, readErr)
	}
}

// TestUpdate_ArchiveRunScopedNoOverwrite — card t1547 review finding 5: a
// second update over the same fixture must not overwrite the first run's
// recovery copy — each run's archive lives under its own timestamp scope.
func TestUpdate_ArchiveRunScopedNoOverwrite(t *testing.T) {
	const rel = ".claude/rules/moai/old-rule.md"
	carried := map[string]string{".claude/rules/moai/update-note.md": "render\n"}
	tmplFS := recTmplFS(carried)

	// Run 1: stale v1 content archived.
	root := newClassifyFixture(t, map[string]string{rel: "stale v1\n"})
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	recRunUpdate(t, root, tmplFS, renderWith(carried), mgr.Manifest(), carried)

	// Run 2: a NEW stale file with different content is archived; the
	// run-1 copy must survive byte-identical.
	const rel2 = ".claude/rules/moai/old-rule-2.md"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel2)), []byte("stale v2\n"), 0o644); err != nil {
		t.Fatalf("seed run-2 stale file: %v", err)
	}
	if err := mgr.Track(rel2, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track 2: %v", err)
	}
	recRunUpdate(t, root, tmplFS, renderWith(carried), mgr.Manifest(), carried)

	firstCopy := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()), "*", filepath.FromSlash(rel))
	matches, err := filepath.Glob(firstCopy)
	if err != nil || len(matches) != 1 {
		t.Fatalf("run-1 archive copies = %v (err=%v), want exactly 1 untouched", matches, err)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil || string(data) != "stale v1\n" {
		t.Errorf("run-1 archive copy overwritten or altered: %q (err=%v)", data, err)
	}
	secondCopy := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()), "*", filepath.FromSlash(rel2))
	matches2, err := filepath.Glob(secondCopy)
	if err != nil || len(matches2) != 1 {
		t.Fatalf("run-2 archive copies = %v (err=%v), want exactly 1", matches2, err)
	}
}

// TestUpdate_ExcludedPathsNotPending — card t1547 review finding 2: paths
// another update-flow step already reconciles (the config sections restore,
// the mergeable-file merge) are never captured for the merge phase, so the
// phase cannot diff the operator's pre-deploy bytes against the other step's
// merged output and revert its delivered template updates.
func TestUpdate_ExcludedPathsNotPending(t *testing.T) {
	const rel = ".moai/config/sections/git-strategy.yaml"
	root := newClassifyFixture(t, map[string]string{
		rel: "git_strategy:\n  worktree_base_branch: develop\n",
	})
	carried := map[string]string{rel: "git_strategy:\n  worktree_base_branch: \"\"\n"}
	tmplFS := recTmplFS(carried)

	_, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), nil, ReconcileOptions{
		// The production exclude shape (the cli wiring composes it with the
		// mergeable-file set): restore-handled config sections are never
		// captured for reprocessing.
		Exclude: func(rel string) bool { return strings.HasPrefix(rel, ".moai/config/sections/") },
	})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	for _, p := range pending {
		if p.RelPath == rel {
			t.Errorf("restore-handled path %s captured for reprocessing: %+v", rel, pending)
		}
	}
}

// TestUpdate_SummaryHonesty — AC-UPM-032. One fixture run that refreshes,
// merges, conflicts, preserves, and archive-removes at least one file each:
// the summary reports all five categories with per-path lists (including
// every deletion), and the report renderer surfaces every archived removal.
// Structure assertions only — no output-format vocabulary that pre-commits
// t1527.
func TestUpdate_SummaryHonesty(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		// refreshed: healthy tracked record, content == render.
		".claude/rules/moai/note.md": "render\n",
		// merged: user changed one key, template changed another → clean.
		".claude/rules/moai/tuning.json": "{\"a\": \"user\", \"b\": \"base\"}\n",
		// conflicted: both sides changed the same key.
		".claude/rules/moai/policy.json": "{\"shared\": \"user-value\"}\n",
		// preserved: local-only.
		".claude/rules/moai/local-note.md": "operator note\n",
		// archived-removed: prior template carried it, current does not.
		".claude/rules/moai/old-rule.md": "stale\n",
	})
	carried := map[string]string{
		".claude/rules/moai/note.md":     "render\n",
		".claude/rules/moai/tuning.json": "{\"a\": \"user\", \"b\": \"template\"}\n",
		".claude/rules/moai/policy.json": "{\"shared\": \"template-value\"}\n",
	}
	tmplFS := recTmplFS(carried)

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for rel, content := range map[string]string{
		".claude/rules/moai/note.md":     "render\n",
		".claude/rules/moai/old-rule.md": "stale\n",
	} {
		if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
			t.Fatalf("track %s: %v", rel, err)
		}
		entry := mgr.Manifest().Files[rel]
		entry.CurrentHash = manifest.HashBytes([]byte(content))
		mgr.Manifest().Files[rel] = entry
	}

	base := func(rel string) ([]byte, bool) {
		if rel == ".claude/rules/moai/tuning.json" {
			return []byte("{\"a\": \"user\", \"b\": \"base\"}\n"), true
		}
		if rel == ".claude/rules/moai/policy.json" {
			return []byte("{\"shared\": \"base-value\"}\n"), true
		}
		return nil, false
	}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), mgr.Manifest(), ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), mgr.Manifest(), ReconcileMergeOptions{Base: base}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}

	// All five categories carry at least one path.
	if !containsPath(summary.Refreshed, ".claude/rules/moai/note.md") {
		t.Errorf("refreshed set missing the healthy file: %+v", summary)
	}
	if !containsPath(summary.Merged, ".claude/rules/moai/tuning.json") {
		t.Errorf("merged set missing the clean merge: %+v", summary)
	}
	if len(summary.Conflicts) != 1 || summary.Conflicts[0].Path != ".claude/rules/moai/policy.json" {
		t.Errorf("conflicts = %+v, want exactly the policy.json conflict", summary.Conflicts)
	}
	if !containsPath(summary.Preserved, ".claude/rules/moai/local-note.md") {
		t.Errorf("preserved set missing the local-only file: %+v", summary)
	}
	if !containsPath(summary.ArchivedRemoved, ".claude/rules/moai/old-rule.md") {
		t.Errorf("archived-removed set missing the stale file: %+v", summary)
	}

	// The renderer names every deletion (REQ-UPM-031) and lists the
	// preserved set — structure only, no t1527 vocabulary.
	counts := report.ReconciliationCounts{
		Refreshed:       len(summary.Refreshed),
		Merged:          len(summary.Merged),
		Conflicts:       len(summary.Conflicts),
		Preserved:       len(summary.Preserved),
		ArchivedRemoved: len(summary.ArchivedRemoved),
	}
	rendered := report.RenderReconciliation(counts,
		[]string{summary.Conflicts[0].Path}, summary.Preserved, summary.ArchivedRemoved)
	if !strings.Contains(rendered, ".claude/rules/moai/old-rule.md") {
		t.Errorf("rendered summary must name every archived removal, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, ".claude/rules/moai/local-note.md") {
		t.Errorf("rendered summary must list the preserved set, got:\n%s", rendered)
	}
}
