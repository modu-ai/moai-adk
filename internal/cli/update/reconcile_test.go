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
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

// TestMergePhaseSkipsVanishedPending — the merge phase's not-exist arm: a
// pending user-modified file the deploy stage did not carry (removed, not
// rewritten) leaves the operator's bytes as they are — nothing to reconcile,
// no error.
func TestMergePhaseSkipsVanishedPending(t *testing.T) {
	const rel = ".claude/rules/moai/tuned.md"
	root := newClassifyFixture(t, map[string]string{rel: "user edit\n"})
	carried := map[string]string{".claude/rules/moai/note.md": "render\n"}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	// The deploy did not carry this path and it vanished in between.
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("remove: %v", err)
	}
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}
	if len(summary.Merged) != 0 || len(summary.Conflicts) != 0 {
		t.Errorf("vanished pending produced an outcome: %+v", summary)
	}
}

// TestExclusiveWriteFileFileParentIsAFile — the helper's MkdirAll arm: a
// target whose parent chain terminates at a regular file errors instead of
// creating through it.
func TestExclusiveWriteFileFileParentIsAFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude", "rules"), []byte("a file, not a dir\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := exclusiveWriteFile(root, ".claude/rules/moai/policy.json.moai-new", []byte("x\n")); err == nil {
		t.Fatal("exclusiveWriteFile under a file-parent must error")
	}
}

// TestMergePhaseAbortsOnLinkedParentRestore — the conflict-restore error
// arm: when the pending file's PARENT chain is swapped to a link before the
// merge phase, the conflict disposition refuses to write (and the whole
// phase aborts with the error) — it never writes through the link.
func TestMergePhaseAbortsOnLinkedParentRestore(t *testing.T) {
	external := t.TempDir()
	const rel = ".claude/rules/moai/tuned.md"
	root := newClassifyFixture(t, map[string]string{rel: "user edit\n"})
	carried := map[string]string{rel: "render NEW\n", ".claude/rules/moai/note.md": "render\n"}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	// Swap the PARENT directory (.claude/rules/moai) to an external link.
	// The external dir carries a same-named file so the pending read
	// SUCCEEDS through the link (the render bytes are external) — pushing
	// the disposition into its write, which the chain gate must refuse.
	if err := os.RemoveAll(filepath.Join(root, ".claude", "rules", "moai")); err != nil {
		t.Fatalf("remove dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(external, "tuned.md"), []byte("render NEW\n"), 0o644); err != nil {
		t.Fatalf("external twin: %v", err)
	}
	if err := os.Symlink(external, filepath.Join(root, ".claude", "rules", "moai")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{}, pending, summary)
	if err == nil {
		t.Fatal("conflict restore through a linked parent must be refused")
	}
	// The external twin survives — the write was refused, not followed. The
	// conflict-restore content must not have landed in the external dir.
	if data, readErr := os.ReadFile(filepath.Join(external, "tuned.md")); readErr != nil || string(data) != "render NEW\n" {
		t.Errorf("external twin altered by the refused restore: %q (err=%v)", data, readErr)
	}
	entries, readErr := os.ReadDir(external)
	if readErr != nil || len(entries) != 1 {
		t.Errorf("external directory polluted: %v (err=%v)", entries, readErr)
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

// TestConflictPreservedFileStaysUserModifiedNextRun — gate round 9, finding
// 1's safety property: a conflict-preserved file whose manifest record still
// reads template_managed with the RENDER's hash (the state the deploy's own
// tracking left behind) classifies user-modified on the NEXT run, because
// the recorded hash no longer matches the operator's restored bytes. The cli
// wiring's exclusion of conflict paths from the retrack is what keeps this
// property — re-tracking one would flip it to template-owned and the next
// update would overwrite the operator's content (the R-2 recurrence path).
func TestConflictPreservedFileStaysUserModifiedNextRun(t *testing.T) {
	const rel = ".claude/rules/moai/policy.json"
	const ours = "{\"shared\": \"user-value\"}\n"
	const render = "{\"shared\": \"template-value\"}\n"
	root := newClassifyFixture(t, map[string]string{rel: ours})
	carried := map[string]string{rel: render}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	// The deploy's tracking state: template_managed with the render's hash.
	entry := mgr.Manifest().Files[rel]
	entry.CurrentHash = manifest.HashBytes([]byte(render))
	mgr.Manifest().Files[rel] = entry

	plan, err := ClassifyManagedRoots(root, []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")},
		renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	if got := plan.ClassOf(rel); got != ClassUserModified {
		t.Errorf("conflict-preserved file class = %q, want %q (template-owned would let the next update overwrite it)", got, ClassUserModified)
	}
}

// TestUniqueArchiveRunDirDistinct — gate round 9, finding 2: two claims in
// the same second land in distinct directories (the atomic Mkdir + numbered
// suffix), so a same-stamped re-run can never write over a previous run's
// recovery copy.
func TestUniqueArchiveRunDirDistinct(t *testing.T) {
	root := t.TempDir()
	first, err := uniqueArchiveRunDir(root)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	second, err := uniqueArchiveRunDir(root)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if first == second {
		t.Fatalf("two claims returned the same run dir %q — a re-run would overwrite the first recovery copies", first)
	}
	// Both exist as real directories.
	for _, d := range []string{first, second} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(d)))
		if err != nil || !info.IsDir() {
			t.Errorf("claimed run dir %s missing or not a directory (err=%v)", d, err)
		}
	}
}

// TestArchiveRootCreationRefusesExternalLink — gate round 10, finding 6:
// .moai/archive as an external symlink is refused BEFORE any creation — the
// MkdirAll must never follow the link and create the archive tree outside
// the project (the round 11 refinement: the check precedes creation).
func TestArchiveRootCreationRefusesExternalLink(t *testing.T) {
	external := t.TempDir()
	const rel = ".claude/rules/moai/old-rule.md"
	root := newClassifyFixture(t, map[string]string{rel: "stale\n"})
	archiveAbs := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()))
	if err := os.MkdirAll(filepath.Dir(archiveAbs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, archiveAbs); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	carried := map[string]string{".claude/rules/moai/note.md": "render\n"}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}

	if _, _, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), mgr.Manifest(), ReconcileOptions{}); err == nil {
		t.Fatal("archive root on an external link must be refused")
	}
	// Nothing was created outside the project through the link.
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatalf("read external: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("external directory polluted through the link: %v", entries)
	}
	// The stale file is still in place.
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Errorf("stale file missing after the refused run: %v", err)
	}
}

// TestArchiveNestedStaleFileArchived — gate round 16: a stale file under a
// NESTED managed path archives into the matching nested layout (the run's
// destination subdirectories are created first) and is removed from place.
func TestArchiveNestedStaleFileArchived(t *testing.T) {
	const rel = ".claude/rules/moai/sub/deep-rule.md"
	root := newClassifyFixture(t, map[string]string{rel: "nested stale\n"})
	carried := map[string]string{".claude/rules/moai/note.md": "render\n"}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}

	summary := recRunUpdate(t, root, recTmplFS(carried), renderWith(carried), mgr.Manifest(), carried)

	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Errorf("nested stale file still in place (err=%v)", err)
	}
	tagRoot := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()))
	matches, err := filepath.Glob(filepath.Join(tagRoot, "*", filepath.FromSlash(rel)))
	if err != nil || len(matches) != 1 {
		t.Fatalf("nested archive copies = %v (err=%v), want exactly 1", matches, err)
	}
	if data, readErr := os.ReadFile(matches[0]); readErr != nil || string(data) != "nested stale\n" {
		t.Errorf("nested archive copy = %q (err=%v)", data, readErr)
	}
	if !containsPath(summary.ArchivedRemoved, rel) {
		t.Errorf("summary.ArchivedRemoved = %v, want it to list %s", summary.ArchivedRemoved, rel)
	}
}

// TestMergePhaseRefusesSwappedSymlinkTarget — gate round 10, finding 2: a
// user-modified file swapped to an EXTERNAL symlink between the reconcile
// and the merge phase must not route the merge writes outside the project.
// The safe write replaces the link (rename never follows) with the
// operator's bytes; the external sentinel is untouched.
func TestMergePhaseRefusesSwappedSymlinkTarget(t *testing.T) {
	external := t.TempDir()
	sentinel := filepath.Join(external, "sentinel.txt")
	const sentinelBytes = "external — byte-identical\n"
	if err := os.WriteFile(sentinel, []byte(sentinelBytes), 0o644); err != nil {
		t.Fatalf("sentinel: %v", err)
	}
	const rel = ".claude/rules/moai/tuned.md"
	const ours = "operator tuning\n"
	const rendered = "template render NEW\n"
	root := newClassifyFixture(t, map[string]string{rel: ours})
	carried := map[string]string{rel: rendered, ".claude/rules/moai/note.md": "render\n"}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	// The swap: the deployed file becomes a link to the external sentinel.
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.Remove(abs); err != nil {
		t.Fatalf("remove deployed file: %v", err)
	}
	if err := os.Symlink(sentinel, abs); err != nil {
		t.Fatalf("swap to symlink: %v", err)
	}

	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}

	// The external sentinel is byte-identical — nothing followed the link.
	if data, readErr := os.ReadFile(sentinel); readErr != nil || string(data) != sentinelBytes {
		t.Errorf("external sentinel polluted: %q (err=%v)", data, readErr)
	}
	// The path no longer carries a link, and the disposition ran (merged or
	// conflict — either way the write landed INSIDE the project).
	info, err := os.Lstat(abs)
	if err != nil {
		t.Fatalf("reconciled path missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("path still a symlink — the safe write did not replace it")
	}
	_ = summary
}

// TestClassifyNonRegularTargetDoesNotHang — gate round 10, finding 5: a
// FIFO at a managed target path must not hang the classifier's read (the
// dry-run preview included). Non-regular plain targets are skipped.
func TestClassifyNonRegularTargetDoesNotHang(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFOs are a POSIX fixture")
	}
	root := newClassifyFixture(t, map[string]string{})
	fifo := filepath.Join(root, ".claude", "rules", "moai", "pipe")
	if err := os.MkdirAll(filepath.Dir(fifo), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	makeFifo(t, fifo)

	done := make(chan error, 1)
	go func() {
		_, err := ClassifyManagedRoots(root, []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")},
			renderWith(map[string]string{}), nil)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ClassifyManagedRoots: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("classifier blocked reading a FIFO — the IsRegular guard is missing")
	}
}

// TestArchiveThenRemoveRefusesLinkedDestinationDir — gate round 11's
// parent-swap variant, exercised through its deterministic shape: an archive
// root whose PARENT chain holds a link is refused at the pre-check, and the
// operator's file stays in place (the removal is gated on a verified-clean
// chain, so a mid-copy swap can never destroy the original).
func TestArchiveThenRemoveRefusesLinkedDestinationDir(t *testing.T) {
	external := t.TempDir()
	const rel = ".claude/rules/moai/old-rule.md"
	root := newClassifyFixture(t, map[string]string{rel: "stale\n"})
	// The claimed run dir is created by the reconcile; here the deterministic
	// variant drives archiveThenRemove directly with a root whose parent is a
	// symlink to the external directory.
	linkedRoot := filepath.Join(ReconcileArchiveFilesRoot(), "linked-parent")
	abs := filepath.Join(root, filepath.FromSlash(linkedRoot))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, abs); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	err := archiveThenRemove(root, rel, linkedRoot)
	if err == nil {
		t.Fatal("archive through a linked destination parent must be refused")
	}
	if data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); readErr != nil || string(data) != "stale\n" {
		t.Errorf("operator's file altered by the refused archive: %q (err=%v)", data, readErr)
	}
	entries, err := os.ReadDir(external)
	if err != nil || len(entries) != 0 {
		t.Errorf("external directory polluted: %v (err=%v)", entries, err)
	}
}

// TestArchiveThenRemoveRefusesLinkedSourceDir — gate round 17, finding 1's
// deterministic shape: the SOURCE's parent chain swapped to an external link
// refuses the archive outright — no read through the link, and above all no
// os.Remove at the swapped path (the gate's reproduction deleted the
// external sentinel through it).
func TestArchiveThenRemoveRefusesLinkedSourceDir(t *testing.T) {
	external := t.TempDir()
	sentinel := filepath.Join(external, "sentinel.txt")
	const sentinelBytes = "external file — must survive\n"
	if err := os.WriteFile(sentinel, []byte(sentinelBytes), 0o644); err != nil {
		t.Fatalf("sentinel: %v", err)
	}
	const rel = ".claude/rules/moai/old-rule.md"
	root := newClassifyFixture(t, map[string]string{})
	// The source's parent (.claude/rules/moai) is a symlink to the external
	// directory, and the source path resolves INSIDE it.
	externalRule := filepath.Join(external, "old-rule.md")
	if err := os.WriteFile(externalRule, []byte("external twin\n"), 0o644); err != nil {
		t.Fatalf("external twin: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude", "rules"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, filepath.Join(root, ".claude", "rules", "moai")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	err := archiveThenRemove(root, rel, filepath.Join(ReconcileArchiveFilesRoot(), "run"))
	if err == nil {
		t.Fatal("archive through a linked source parent must be refused")
	}
	if data, readErr := os.ReadFile(sentinel); readErr != nil || string(data) != sentinelBytes {
		t.Errorf("external sentinel deleted or altered: %q (err=%v)", data, readErr)
	}
	// The external twin (the entry the swapped path would resolve to) also
	// survives — the removal never fired.
	if _, statErr := os.Stat(externalRule); statErr != nil {
		t.Errorf("external twin deleted: %v", statErr)
	}
}

// TestExclusiveSidecarClaimNeverClobbers — gate round 17, finding 2: the
// sidecar name is claimed with an exclusive create. An occupied candidate is
// refused (EEXIST), never followed or replaced; the pre-existing sibling's
// bytes are byte-identical after the conflict disposition ran through the
// numbered retries.
func TestExclusiveSidecarClaimNeverClobbers(t *testing.T) {
	root := t.TempDir()
	occupied := filepath.Join(root, ".claude", "rules", "moai", "policy.json.moai-new")
	if err := os.MkdirAll(filepath.Dir(occupied), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(occupied, []byte("32MiB recovery file from a prior run\n"), 0o644); err != nil {
		t.Fatalf("seed sibling: %v", err)
	}

	err := exclusiveWriteFile(root, ".claude/rules/moai/policy.json.moai-new", []byte("payload\n"))
	if !os.IsExist(err) {
		t.Fatalf("exclusive create on an occupied name: err=%v, want EEXIST", err)
	}
	// The occupied sibling is byte-identical — never clobbered.
	if data, readErr := os.ReadFile(occupied); readErr != nil || string(data) != "32MiB recovery file from a prior run\n" {
		t.Errorf("occupied sibling clobbered: %q (err=%v)", data, readErr)
	}
	// A fresh name claims exclusively and lands the payload.
	if err := exclusiveWriteFile(root, ".claude/rules/moai/policy.json.moai-new.2", []byte("payload\n")); err != nil {
		t.Fatalf("exclusive create on a fresh name: %v", err)
	}
	if data, readErr := os.ReadFile(filepath.Join(root, ".claude", "rules", "moai", "policy.json.moai-new.2")); readErr != nil || string(data) != "payload\n" {
		t.Errorf("claimed sidecar content = %q (err=%v)", data, readErr)
	}
}

// TestCleanTreeEndStateEqualsWipeRedeploy — AC-UPM-010 / NFR-UPM-003. A
// clean fixture (no operator modifications, no local-only files) run through
// the pipeline reaches the same end state as the wipe-and-redeploy it
// replaces: directory-diff equal. The preservation pipeline is additive —
// a behavior change for no clean tree.
func TestCleanTreeEndStateEqualsWipeRedeploy(t *testing.T) {
	files := map[string]string{
		".claude/settings.json":             "{\n  \"permissions\": {}\n}\n",
		".claude/rules/moai/rule.md":        "rule render\n",
		".moai/config/sections/system.yaml": "moai:\n  template_version: 0.0.0\n",
		".moai/config/sections/llm.yaml":    "llm:\n  harness: claude\n",
	}
	carried := map[string]string{
		".claude/settings.json":             "{\n  \"permissions\": {}\n}\n",
		".claude/rules/moai/rule.md":        "rule render NEW\n",
		".moai/config/sections/system.yaml": "moai:\n  template_version: 9.9.9\n",
		".moai/config/sections/llm.yaml":    "llm:\n  harness: claude\n",
	}

	// Arm A — the wipe-and-redeploy end state (the M1-characterized flow):
	// clean removes everything, deploy writes the renders.
	rootA := newClassifyFixture(t, files)
	if err := deploy.CleanMoaiManagedPaths(rootA, io.Discard, recTmplFS(carried)); err != nil {
		t.Fatalf("arm A clean: %v", err)
	}
	recSimulateDeploy(t, rootA, carried)

	// Arm B — the pipeline: reconcile → [deploy] → merge. The clean-project
	// state of the AC's Given: a moai-installed tree whose manifest is
	// healthy (every carried file pristine as last deployed). The
	// manifest-ABSENT divergence case deliberately takes the R-1
	// conservative route (conflict-preserve) instead — an untracked file
	// whose bytes differ from the render is indistinguishable from an
	// operator edit, and the pipeline never guesses.
	rootB := newClassifyFixture(t, files)
	mgrB := manifest.NewManager()
	if _, err := mgrB.Load(rootB); err != nil {
		t.Fatalf("arm B manifest: %v", err)
	}
	for rel, prior := range files {
		if err := mgrB.Track(rel, manifest.TemplateManaged, ""); err != nil {
			t.Fatalf("arm B track %s: %v", rel, err)
		}
		entry := mgrB.Manifest().Files[rel]
		entry.CurrentHash = manifest.HashBytes([]byte(prior))
		mgrB.Manifest().Files[rel] = entry
	}
	summary, pending, err := ReconcileManagedPaths(rootB, io.Discard, recTmplFS(carried), renderWith(carried), mgrB.Manifest(), ReconcileOptions{})
	if err != nil {
		t.Fatalf("arm B reconcile: %v", err)
	}
	recSimulateDeploy(t, rootB, carried)
	summary, err = ReconcileMerges(rootB, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{}, pending, summary)
	if err != nil {
		t.Fatalf("arm B merges: %v", err)
	}

	// Directory-diff equal: every fixture path ends byte-identical across
	// the two arms.
	for rel := range files {
		absA := filepath.Join(rootA, filepath.FromSlash(rel))
		absB := filepath.Join(rootB, filepath.FromSlash(rel))
		dataA, errA := os.ReadFile(absA)
		dataB, errB := os.ReadFile(absB)
		if (errA == nil) != (errB == nil) {
			t.Errorf("%s existence differs: arm A err=%v, arm B err=%v", rel, errA, errB)
			continue
		}
		if errA == nil && string(dataA) != string(dataB) {
			t.Errorf("%s content differs across arms:\nA: %q\nB: %q", rel, dataA, dataB)
		}
	}
	_ = summary
}

// TestAbortLeavesTreeIntact — AC-UPM-041 (archive arm). An archive
// machinery failure before any copy aborts the reconcile with the whole
// tree byte-identical: no stale file was removed, no pending merge was
// captured-and-lost (the run returns an error and nothing else happened).
func TestAbortLeavesTreeIntact(t *testing.T) {
	const staleRel = ".claude/rules/moai/old-rule.md"
	const otherRel = ".claude/rules/moai/keep.md"
	// .moai/archive exists as a regular FILE: the run-dir claim fails
	// before any archive write.
	root := newClassifyFixture(t, map[string]string{
		staleRel:        "stale\n",
		otherRel:        "keep me\n",
		".moai/archive": "not a directory\n",
	})
	carried := map[string]string{".claude/rules/moai/note.md": "render\n"}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(staleRel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}

	_, _, err := ReconcileManagedPaths(root, io.Discard, recTmplFS(carried), renderWith(carried), mgr.Manifest(), ReconcileOptions{})
	if err == nil {
		t.Fatal("archive claim failure must abort the reconcile")
	}
	// Both files are byte-identical to pre-run: the failure preceded every
	// destructive step.
	for rel, want := range map[string]string{staleRel: "stale\n", otherRel: "keep me\n"} {
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if readErr != nil || string(data) != want {
			t.Errorf("%s altered by the aborted run: %q (err=%v)", rel, data, readErr)
		}
	}
}
