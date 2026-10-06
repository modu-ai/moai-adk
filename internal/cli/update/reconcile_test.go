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
	"testing"
	"testing/fstest"

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
	const baseJSON = "{\"keep\": \"user\", \"shared\": \"user-value\"}\n"
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
	const baseJSON = "{\"keep\": \"user\", \"shared\": \"user-value\"}\n"
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
	// Archived byte-identical under the migration archive root, layout kept.
	archivePath := filepath.Join(root, filepath.FromSlash(ArchiveFilesRoot()), filepath.FromSlash(rel))
	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("stale file not archived at %s: %v", archivePath, err)
	}
	if string(data) != staleContent {
		t.Errorf("archive copy = %q, want byte-identical %q", data, staleContent)
	}
	// The summary lists the removal.
	if !containsPath(summary.ArchivedRemoved, rel) {
		t.Errorf("summary.ArchivedRemoved = %v, want it to list %s", summary.ArchivedRemoved, rel)
	}
}
