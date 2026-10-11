package update

// reconcile_classify_test.go — SPEC-UPDATE-MIGRATION-001 M3 (card t1547):
// the managed-root classifier's behavior table (AC-UPM-001/002/003/050).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// recTarget builds one plain CleanTarget relative to the project root.
func recTarget(projectRoot, rel string) deploy.CleanTarget {
	return deploy.CleanTarget{
		DisplayPath: rel,
		FullPath:    filepath.Join(projectRoot, filepath.FromSlash(rel)),
	}
}

// TestClassifyManagedRootsFourClasses — AC-UPM-001. Every regular file under
// the managed roots lands in exactly one of the four classes; the sets are
// disjoint and exhaustive; a non-ASCII-named local-only file classifies by
// the same rules; a moai-named file the template does not carry is
// user-owned, NOT removal-eligible (the carriage gate).
func TestClassifyManagedRootsFourClasses(t *testing.T) {
	const renderedRule = "# template rule (current render)\n"
	root := newClassifyFixture(t, map[string]string{
		// template-owned: carried, healthy record, content == render.
		".claude/rules/moai/update-note.md": renderedRule,
		// user-modified: carried, content diverges.
		".claude/rules/moai/tuned.md": "# template rule with operator tuning\n",
		// user-owned: local-only.
		".claude/rules/moai/dev-only-rule.md": "local\n",
		// user-owned: non-ASCII name, same byte-semantic rules.
		".claude/rules/moai/한글-규칙.md": "로컬 규칙\n",
		// user-owned: a moai- name the template does NOT carry — the D-15
		// carriage gate keeps it out of the removal classes.
		".claude/rules/moai/moai-local-note.md": "operator note\n",
		// stale: a prior template carried it, the current one does not.
		".claude/rules/moai/old-rule.md": "dropped upstream\n",
	})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": renderedRule,
		".claude/rules/moai/tuned.md":       renderedRule,
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(".claude/rules/moai/update-note.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track update-note: %v", err)
	}
	if err := mgr.Track(".claude/rules/moai/old-rule.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track old-rule: %v", err)
	}

	targets := []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")}
	plan, err := ClassifyManagedRoots(root, targets, renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}

	want := map[string]ReconcileClass{
		".claude/rules/moai/update-note.md":     ClassTemplateOwned,
		".claude/rules/moai/tuned.md":           ClassUserModified,
		".claude/rules/moai/dev-only-rule.md":   ClassUserOwned,
		".claude/rules/moai/한글-규칙.md":           ClassUserOwned,
		".claude/rules/moai/moai-local-note.md": ClassUserOwned,
		".claude/rules/moai/old-rule.md":        ClassStale,
	}
	got := map[string]ReconcileClass{}
	for _, set := range [][]ReconcileFile{plan.TemplateOwned, plan.UserModified, plan.UserOwned, plan.Stale} {
		for _, f := range set {
			if _, dup := got[f.RelPath]; dup {
				t.Errorf("file %s classified twice (sets not disjoint)", f.RelPath)
			}
			got[f.RelPath] = f.Class
		}
	}
	if len(got) != plan.Count() {
		t.Errorf("classified %d unique paths for %d files — sets overlap", len(got), plan.Count())
	}
	for rel, wantClass := range want {
		if gotClass := got[rel]; gotClass != wantClass {
			t.Errorf("%s class = %q, want %q", rel, gotClass, wantClass)
		}
	}
	if plan.Count() != len(want) {
		t.Errorf("classified %d files, want %d (exhaustive over the fixture): %v", plan.Count(), len(want), got)
	}
	if plan.ClassOf(".claude/rules/moai/absent.md") != ReconcileClass("") {
		t.Errorf("ClassOf for an unseen path should be empty")
	}
}

// TestClassifyUserOwnedNamespaceForced — AC-UPM-002. IsUserOwnedNamespace
// paths classify user-owned regardless of manifest state, even when the
// manifest record would otherwise route them elsewhere.
func TestClassifyUserOwnedNamespaceForced(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		".claude/skills/hns-worker/SKILL.md": "# user harness skill\n",
		".claude/skills/moai-core/SKILL.md":  "# template skill\n",
	})
	carried := map[string]string{
		".claude/skills/moai-core/SKILL.md": "# template skill\n",
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// A manifest record claims the user skill is managed — the namespace
	// predicate must override it.
	if err := mgr.Track(".claude/skills/hns-worker/SKILL.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track hns skill: %v", err)
	}
	// moai-core carries a healthy managed record so its carriage decision is
	// template-owned (content equals the render).
	if err := mgr.Track(".claude/skills/moai-core/SKILL.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track moai-core: %v", err)
	}

	// Scope covers the whole skills dir (a caller-supplied target), so the
	// hns-* prefix is inside managed-root scope for this test.
	targets := []deploy.CleanTarget{recTarget(root, ".claude/skills")}
	plan, err := ClassifyManagedRoots(root, targets, renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	if got := plan.ClassOf(".claude/skills/hns-worker/SKILL.md"); got != ClassUserOwned {
		t.Errorf("hns-worker class = %q, want %q (REQ-UPM-003 force-preserve)", got, ClassUserOwned)
	}
	if got := plan.ClassOf(".claude/skills/moai-core/SKILL.md"); got != ClassTemplateOwned {
		t.Errorf("moai-core class = %q, want %q", got, ClassTemplateOwned)
	}
}

// TestClassifySymlinksNotDereferenced — AC-UPM-003 (classification half).
// Live and dangling links under a managed root are recorded in Symlinks,
// never classified, and never followed.
func TestClassifySymlinksNotDereferenced(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		".claude/rules/moai/real-rule.md": "real\n",
	})
	// Live file link.
	if err := os.Symlink(
		filepath.Join(root, ".claude", "rules", "moai", "real-rule.md"),
		filepath.Join(root, ".claude", "rules", "moai", "link-rule.md"),
	); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	// Dangling link.
	if err := os.Symlink(
		filepath.Join(root, ".claude", "rules", "moai", "no-such-target.md"),
		filepath.Join(root, ".claude", "rules", "moai", "dangling-rule.md"),
	); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	targets := []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")}
	plan, err := ClassifyManagedRoots(root, targets, renderWith(map[string]string{}), nil)
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	if got := plan.ClassOf(".claude/rules/moai/link-rule.md"); got != "" {
		t.Errorf("live link was classified %q — links are never classified", got)
	}
	if got := plan.ClassOf(".claude/rules/moai/dangling-rule.md"); got != "" {
		t.Errorf("dangling link was classified %q — links are never classified", got)
	}
	want := map[string]bool{
		".claude/rules/moai/link-rule.md":     true,
		".claude/rules/moai/dangling-rule.md": true,
	}
	if len(plan.Symlinks) != len(want) {
		t.Fatalf("Symlinks = %v, want exactly %v", plan.Symlinks, want)
	}
	for _, s := range plan.Symlinks {
		if !want[s] {
			t.Errorf("unexpected symlink record %q", s)
		}
	}
}

// TestClassifyDeterministic — AC-UPM-050. Two runs over the same tree and
// template set produce byte-identical (JSON-serialized) output.
func TestClassifyDeterministic(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		".claude/rules/moai/update-note.md":   "render\n",
		".claude/rules/moai/tuned.md":         "user edit\n",
		".claude/rules/moai/dev-only-rule.md": "local\n",
		".claude/rules/moai/old-rule.md":      "stale\n",
	})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "render\n",
		".claude/rules/moai/tuned.md":       "render\n",
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(".claude/rules/moai/old-rule.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track stale: %v", err)
	}
	targets := []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")}

	first, err := ClassifyManagedRoots(root, targets, renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("first classify: %v", err)
	}
	second, err := ClassifyManagedRoots(root, targets, renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("second classify: %v", err)
	}
	a, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	b, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(a) != string(b) {
		t.Errorf("classification not deterministic:\nfirst:  %s\nsecond: %s", a, b)
	}
}

// TestClassifyManifestAbsentProject — spec R-1 conservative route: no
// manifest → every template-carried file reads user-modified, every
// template-absent file reads user-owned (never removal-eligible).
func TestClassifyManifestAbsentProject(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		".claude/rules/moai/update-note.md":   "render\n",
		".claude/rules/moai/dev-only-rule.md": "local\n",
	})
	carried := map[string]string{
		".claude/rules/moai/update-note.md": "render\n",
	}
	targets := []deploy.CleanTarget{recTarget(root, ".claude/rules/moai")}
	plan, err := ClassifyManagedRoots(root, targets, renderWith(carried), nil)
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	if got := plan.ClassOf(".claude/rules/moai/update-note.md"); got != ClassUserModified {
		t.Errorf("template-carried file without manifest = %q, want %q (R-1 conservative)", got, ClassUserModified)
	}
	if got := plan.ClassOf(".claude/rules/moai/dev-only-rule.md"); got != ClassUserOwned {
		t.Errorf("local-only file without manifest = %q, want %q", got, ClassUserOwned)
	}
	if len(plan.Stale) != 0 {
		t.Errorf("manifest-absent project produced stale entries %v — stale requires manifest evidence", plan.Stale)
	}
}

// TestClassifyHashlessRecordTakesTheMergeRoute — gate round 16 (card t1547
// repair round): a template_managed record whose CurrentHash is EMPTY must
// not short-circuit to template-owned — the empty-hash arm verified nothing
// about the disk bytes, so a user-modified file carrying a hash-less record
// was refreshed (overwritten) in place. Hash-less records take the
// ClassUserModified route: the file becomes a merge candidate, and a
// conflict preserves the operator's bytes with a sidecar.
func TestClassifyHashlessRecordTakesTheMergeRoute(t *testing.T) {
	const renderedRule = "# template rule (current render)\n"
	const userEdit = "# template rule (current render)\n<!-- the operator's own note -->\n"
	root := newClassifyFixture(t, map[string]string{
		".claude/rules/moai/hashless.md": userEdit,
	})
	carried := map[string]string{
		".claude/rules/moai/hashless.md": renderedRule,
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// The hash-less record: template_managed with every hash field empty —
	// the shape a hand-edited or legacy-schema manifest carries (manifest
	// Track always fills CurrentHash, so the empty form only arises here).
	mgr.Manifest().Files[".claude/rules/moai/hashless.md"] = manifest.FileEntry{
		Provenance: manifest.TemplateManaged,
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	plan, err := ClassifyManagedRoots(root,
		[]deploy.CleanTarget{recTarget(root, ".claude/rules/moai")},
		renderWith(carried), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyManagedRoots: %v", err)
	}
	const rel = ".claude/rules/moai/hashless.md"
	if got := plan.ClassOf(rel); got != ClassUserModified {
		t.Errorf("a hash-less record's user-modified file classified %q — the empty-hash arm refreshed it over the operator's bytes", got)
	}
}
