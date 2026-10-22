package worktree

// classify_test.go - SPEC-WORKTREE-HOME-ROOT-001 M1 RED commit (card t1640).
//
// The classifier (internal/cli/worktree/classify.go) does not exist yet: this
// package fails to compile until the M1 production commits, which is the RED
// (plan.md section C, test compilation order).
//
// Acceptance mapping: AC-WHR-013 (classifier, including non-WT- trees after
// the move and the unclassified result), AC-WHR-023 (classifier clause of the
// spelling-safe equality).
//
// Surface fixed by this file (the GREEN stage implements exactly this):
//
//	type TreeClass string  // "moai-made" | "claude-native" | "codex-managed" |
//	                       // "integration" | "primary" | "unclassified"
//	type RootKind string   // "home" | "legacy-moai" | "legacy-claude" | "codex-managed" | "other"
//	type WorktreeRecord struct {
//		Path, Head, Branch string  // Branch is the short name; empty when detached or bare
//		Detached, Bare, Locked, Prunable bool
//	}
//	type ClassifyContext struct {
//		PrimaryRoot       string            // primary checkout of this repository
//		IntegrationBranch string            // configured integration branch, for example "main"
//		HomeProjectsDir   string            // <MoaiHome>/worktrees/projects
//		OwnerID           string            // this repository's owner identity
//		MarkerOwners      map[string]string // project-key folder -> owner identity of its marker
//		CodexWorktreesDir string            // the Codex-managed pool (~/.codex/worktrees)
//	}
//	type Classification struct{ Class TreeClass; Root RootKind }
//	func ParseWorktreeList(porcelain string) ([]WorktreeRecord, error)
//	func Classify(rec WorktreeRecord, ctx ClassifyContext) Classification
//
// Classify is a pure function of its arguments apart from the filesystem
// identity comparison behind spelling-safe path equality; the owner markers
// arrive as data (MarkerOwners), not as files read inside the function.
// Fixtures are real directories under t.TempDir() so path comparison does not
// depend on how a missing path is spelled. Linked trees made with
// `git worktree add` are scratch fixtures (the sanctioned exception to
// AGENTS.md section 3).

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------- fixtures

func clsReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

func clsMkdir(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
}

// clsWorld is one repository's surroundings: its primary checkout, the two
// legacy roots, the home projects directory with this repository's folder
// (marker owned by this repository), a foreign folder (marker owned by
// someone else) and a folder with no marker entry, and a Codex pool.
type clsWorld struct {
	primary, moaiRoot, claudeRoot  string
	projects, ownHome, otherHome   string
	noMarkerHome, codex, elsewhere string
	ctx                            ClassifyContext
}

func clsNewWorld(t *testing.T) clsWorld {
	t.Helper()
	base := clsReal(t, t.TempDir())
	w := clsWorld{
		primary:      filepath.Join(base, "repo"),
		projects:     filepath.Join(base, "moai-home", "worktrees", "projects"),
		codex:        filepath.Join(base, "codex-home", ".codex", "worktrees"),
		elsewhere:    filepath.Join(base, "elsewhere"),
		moaiRoot:     filepath.Join(base, "repo", ".moai", "worktrees"),
		claudeRoot:   filepath.Join(base, "repo", ".claude", "worktrees"),
		ownHome:      filepath.Join(base, "moai-home", "worktrees", "projects", "repo"),
		otherHome:    filepath.Join(base, "moai-home", "worktrees", "projects", "other-repo"),
		noMarkerHome: filepath.Join(base, "moai-home", "worktrees", "projects", "no-marker"),
	}
	clsMkdir(t, w.primary, w.moaiRoot, w.claudeRoot, w.ownHome, w.otherHome, w.noMarkerHome, w.codex, w.elsewhere)
	w.ctx = ClassifyContext{
		PrimaryRoot:       w.primary,
		IntegrationBranch: "main",
		HomeProjectsDir:   w.projects,
		OwnerID:           "owner-self",
		MarkerOwners:      map[string]string{"repo": "owner-self", "other-repo": "owner-other"},
		CodexWorktreesDir: w.codex,
	}
	return w
}

// rec creates the tree directory and returns its record.
func (w clsWorld) rec(t *testing.T, path, branch string) WorktreeRecord {
	t.Helper()
	clsMkdir(t, path)
	return WorktreeRecord{Path: path, Head: "0123456789abcdef0123456789abcdef01234567", Branch: branch}
}

// -------------------------------------------------------- AC-WHR-013 classifier

// TestClassify_MeasuredShapes is the table of AC-WHR-013: each measured shape
// of design.md D-2 and the two post-move home-root shapes, with the class and
// the root kind REQ-WHR-010 names. Root "" means the root kind is not pinned
// for that row (the primary checkout and a foreign project's folder).
func TestClassify_MeasuredShapes(t *testing.T) {
	w := clsNewWorld(t)
	type row struct {
		name     string
		path     string
		branch   string
		detached bool
		class    string
		root     string
	}
	rows := []row{
		// moai-made by creation-class witness
		{"home tree with a WT- branch", filepath.Join(w.ownHome, "t1"), "WT-fix-x", false, "moai-made", "home"},
		{"home tree, non-WT- branch, own marker (witness c)", filepath.Join(w.ownHome, "release-v314"), "release/v3.1.4", false, "moai-made", "home"},
		{"home tree, detached, own marker (witness c)", filepath.Join(w.ownHome, "detached-one"), "", true, "moai-made", "home"},
		{"legacy-moai tree with a WT- branch", filepath.Join(w.moaiRoot, "t2"), "WT-fix-y", false, "moai-made", "legacy-moai"},
		{"legacy-moai rc30-preview, non-WT- (witness a)", filepath.Join(w.moaiRoot, "rc30-preview"), "rc30", false, "moai-made", "legacy-moai"},
		{"legacy-moai release-v314, non-WT- (witness a)", filepath.Join(w.moaiRoot, "release-v314"), "release/v3.1.4", false, "moai-made", "legacy-moai"},
		{"legacy-moai t1515-absorb, non-WT- (witness a)", filepath.Join(w.moaiRoot, "t1515-absorb"), "feature/absorb", false, "moai-made", "legacy-moai"},
		{"legacy-moai t1868-publish, non-WT- (witness a)", filepath.Join(w.moaiRoot, "t1868-publish"), "publish-x", false, "moai-made", "legacy-moai"},
		{"legacy-moai detached (witness a)", filepath.Join(w.moaiRoot, "detached-two"), "", true, "moai-made", "legacy-moai"},
		{"legacy-claude moai-made t1497 (witness b)", filepath.Join(w.claudeRoot, "t1497"), "WT-t1497-fix", false, "moai-made", "legacy-claude"},
		{"legacy-claude name that merely starts with agent (no hyphen)", filepath.Join(w.claudeRoot, "agentless"), "WT-agentless", false, "moai-made", "legacy-claude"},
		{"elsewhere tree with a WT- branch (witness b)", filepath.Join(w.elsewhere, "wt-tree"), "WT-elsewhere", false, "moai-made", "other"},
		// claude-native: any agent-* directory under .claude/worktrees
		{"agent-<hex> tree", filepath.Join(w.claudeRoot, "agent-a1b2c3d4"), "worktree-agent-a1b2c3d4", false, "claude-native", "legacy-claude"},
		{"agent-<hex> tree, detached", filepath.Join(w.claudeRoot, "agent-0f9e8d7c"), "", true, "claude-native", "legacy-claude"},
		{"agent-task tree with a WT- branch", filepath.Join(w.claudeRoot, "agent-task"), "WT-agent-task", false, "claude-native", "legacy-claude"},
		// integration
		{"integration worktree under .claude/worktrees", filepath.Join(w.claudeRoot, "main"), "main", false, "integration", "legacy-claude"},
		{"integration branch beats the directory witness", filepath.Join(w.moaiRoot, "holds-main"), "main", false, "integration", "legacy-moai"},
		// codex-managed
		{"codex pool tree", filepath.Join(w.codex, "ab12cd", "repo"), "WT-codex-tree", false, "codex-managed", "codex-managed"},
		{"codex pool tree, detached", filepath.Join(w.codex, "ef34ab", "repo"), "", true, "codex-managed", "codex-managed"},
		// unclassified: fits no class, listed and never moved
		{"claude tree with a worktree- branch (audit-gate-async-worktree-doc)", filepath.Join(w.claudeRoot, "audit-gate-async-worktree-doc"), "worktree-audit-gate-async-worktree-doc", false, "unclassified", "legacy-claude"},
		{"home tree, non-WT- branch, foreign marker", filepath.Join(w.otherHome, "rc30-preview"), "rc30", false, "unclassified", ""},
		{"home tree, non-WT- branch, folder with no marker", filepath.Join(w.noMarkerHome, "rc30-preview"), "rc30", false, "unclassified", ""},
		{"elsewhere tree, non-WT- branch", filepath.Join(w.elsewhere, "feature-tree"), "feature/x", false, "unclassified", "other"},
		// primary
		{"primary checkout", w.primary, "main", false, "primary", ""},
		{"primary checkout on a feature branch", w.primary, "feature/y", false, "primary", ""},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			rec := w.rec(t, r.path, r.branch)
			rec.Detached = r.detached
			got := Classify(rec, w.ctx)
			if string(got.Class) != r.class {
				t.Errorf("class = %q, want %q", got.Class, r.class)
			}
			if r.root != "" && string(got.Root) != r.root {
				t.Errorf("root kind = %q, want %q", got.Root, r.root)
			}
		})
	}
}

// TestClassify_NeverMovedClassesAreNotMoaiMade pins "unclassified is listed
// and never counted as moai-made": over the whole table above, no
// unclassified, claude-native, codex-managed, integration or primary result
// may read as moai-made. The check re-runs a mixed set and counts.
func TestClassify_NeverMovedClassesAreNotMoaiMade(t *testing.T) {
	w := clsNewWorld(t)
	recs := []WorktreeRecord{
		w.rec(t, filepath.Join(w.claudeRoot, "audit-gate-async-worktree-doc"), "worktree-audit-gate-async-worktree-doc"),
		w.rec(t, filepath.Join(w.claudeRoot, "agent-a1b2c3d4"), "worktree-agent-a1b2c3d4"),
		w.rec(t, filepath.Join(w.codex, "ab12cd", "repo"), "WT-x"),
		w.rec(t, filepath.Join(w.claudeRoot, "main"), "main"),
		w.rec(t, w.primary, "main"),
		w.rec(t, filepath.Join(w.otherHome, "rc30-preview"), "rc30"),
	}
	madeBy := 0
	for _, r := range recs {
		if Classify(r, w.ctx).Class == "moai-made" {
			madeBy++
			t.Errorf("%s classified moai-made, want a never-moved class", r.Path)
		}
	}
	if madeBy != 0 {
		t.Errorf("%d of %d never-moved shapes read as moai-made", madeBy, len(recs))
	}
}

// TestClassify_SpellingSafeRootsAndPrimary pins the classifier clause of
// AC-WHR-023: a record spelled with different case than the primary root the
// context carries is still the same tree on a case-insensitive volume.
func TestClassify_SpellingSafeRootsAndPrimary(t *testing.T) {
	if !clsCaseInsensitive(t) {
		t.Skip("volume is case-sensitive; two directories differing only by case stay distinct there (see the next test)")
	}
	base := clsReal(t, t.TempDir())
	primary := filepath.Join(base, "Repo")
	clsMkdir(t, filepath.Join(primary, ".moai", "worktrees", "t1"))
	alias := filepath.Join(base, "repo")
	ctx := ClassifyContext{PrimaryRoot: primary, IntegrationBranch: "main"}

	tree := WorktreeRecord{Path: filepath.Join(alias, ".moai", "worktrees", "t1"), Branch: "feature/z"}
	got := Classify(tree, ctx)
	if string(got.Class) != "moai-made" || string(got.Root) != "legacy-moai" {
		t.Errorf("case-aliased legacy-moai tree = %q/%q, want moai-made/legacy-moai", got.Class, got.Root)
	}
	prim := Classify(WorktreeRecord{Path: alias, Branch: "main"}, ctx)
	if string(prim.Class) != "primary" {
		t.Errorf("case-aliased primary = %q, want primary", prim.Class)
	}
}

func TestClassify_CaseDifferingDirectoriesStayDistinctOnACaseSensitiveVolume(t *testing.T) {
	if clsCaseInsensitive(t) {
		t.Skip("volume is case-insensitive; the case-aliased test above covers it")
	}
	base := clsReal(t, t.TempDir())
	clsMkdir(t, filepath.Join(base, "Repo"), filepath.Join(base, "repo", ".moai", "worktrees", "t1"))
	ctx := ClassifyContext{PrimaryRoot: filepath.Join(base, "Repo"), IntegrationBranch: "main"}
	got := Classify(WorktreeRecord{Path: filepath.Join(base, "repo", ".moai", "worktrees", "t1"), Branch: "feature/z"}, ctx)
	if string(got.Class) == "moai-made" {
		t.Errorf("a tree under repo/ was classified under Repo/'s legacy root on a case-sensitive volume")
	}
}

func clsCaseInsensitive(t *testing.T) bool {
	t.Helper()
	dir := clsReal(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(dir, "CaseProbe"), nil, 0o644); err != nil {
		t.Fatalf("probe file: %v", err)
	}
	_, err := os.Stat(filepath.Join(dir, "caseprobe"))
	return err == nil
}

// ---------------------------------------------------- porcelain record parsing

// TestParseWorktreeList_Porcelain pins the parser over a verbatim
// `git worktree list --porcelain` document: stanzas separated by a blank line,
// the short branch name, detached, bare, locked (with and without a reason)
// and prunable.
func TestParseWorktreeList_Porcelain(t *testing.T) {
	const (
		h1 = "1111111111111111111111111111111111111111"
		h2 = "2222222222222222222222222222222222222222"
		h3 = "3333333333333333333333333333333333333333"
		h4 = "4444444444444444444444444444444444444444"
		h5 = "5555555555555555555555555555555555555555"
	)
	doc := "worktree /r/repo\nHEAD " + h1 + "\nbranch refs/heads/main\n\n" +
		"worktree /r/repo/.moai/worktrees/t1\nHEAD " + h2 + "\nbranch refs/heads/WT-fix\n\n" +
		"worktree /r/repo/.claude/worktrees/agent-ab12\nHEAD " + h3 + "\ndetached\n\n" +
		"worktree /r/repo/.moai/worktrees/t2\nHEAD " + h4 + "\nbranch refs/heads/WT-locked\nlocked kept for the card\n\n" +
		"worktree /r/repo/.moai/worktrees/gone\nHEAD " + h5 + "\nbranch refs/heads/WT-gone\nprunable gitdir file points to non-existent location\n\n" +
		"worktree /r/bare\nbare\n"

	want := []WorktreeRecord{
		{Path: "/r/repo", Head: h1, Branch: "main"},
		{Path: "/r/repo/.moai/worktrees/t1", Head: h2, Branch: "WT-fix"},
		{Path: "/r/repo/.claude/worktrees/agent-ab12", Head: h3, Detached: true},
		{Path: "/r/repo/.moai/worktrees/t2", Head: h4, Branch: "WT-locked", Locked: true},
		{Path: "/r/repo/.moai/worktrees/gone", Head: h5, Branch: "WT-gone", Prunable: true},
		{Path: "/r/bare", Bare: true},
	}
	// Git's porcelain spells paths with slashes on every OS. The parser's
	// filesystem-facing record uses the host's native separators.
	for i := range want {
		want[i].Path = filepath.FromSlash(want[i].Path)
	}
	for name, input := range map[string]string{"with trailing blank line": doc + "\n", "without trailing blank line": doc} {
		t.Run(name, func(t *testing.T) {
			got, err := ParseWorktreeList(input)
			if err != nil {
				t.Fatalf("ParseWorktreeList: %v", err)
			}
			if len(got) != len(want) {
				t.Fatalf("parsed %d records, want %d: %+v", len(got), len(want), got)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("record %d = %+v, want %+v", i, got[i], want[i])
				}
			}
		})
	}

	t.Run("empty input", func(t *testing.T) {
		got, err := ParseWorktreeList("")
		if err != nil || len(got) != 0 {
			t.Errorf("ParseWorktreeList(\"\") = %v, %v, want no records and no error", got, err)
		}
	})
}

// TestParseWorktreeList_RealGit feeds the parser real git output from a
// scratch repository: a branch tree, a detached tree and a locked tree.
func TestParseWorktreeList_RealGit(t *testing.T) {
	base := clsReal(t, t.TempDir())
	repo := filepath.Join(base, "repo")
	clsMkdir(t, repo)
	clsGit(t, repo, "init", "-q", "--template=", "-b", "main")
	clsGit(t, repo, "config", "user.email", "cls-test@example.com")
	clsGit(t, repo, "config", "user.name", "Cls Test")
	clsGit(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	clsGit(t, repo, "add", "README.md")
	clsGit(t, repo, "commit", "-q", "-m", "seed")

	branchTree := filepath.Join(base, "trees", "branch-tree")
	detachedTree := filepath.Join(base, "trees", "detached-tree")
	lockedTree := filepath.Join(base, "trees", "locked-tree")
	clsGit(t, repo, "worktree", "add", "-q", "-b", "WT-real", "--", branchTree)
	clsGit(t, repo, "worktree", "add", "-q", "--detach", "--", detachedTree)
	clsGit(t, repo, "worktree", "add", "-q", "-b", "WT-held", "--", lockedTree)
	clsGit(t, repo, "worktree", "lock", "--reason", "kept for the card", lockedTree)

	out := clsGit(t, repo, "worktree", "list", "--porcelain")
	got, err := ParseWorktreeList(out)
	if err != nil {
		t.Fatalf("ParseWorktreeList: %v\n%s", err, out)
	}
	if len(got) != 4 {
		t.Fatalf("parsed %d records, want 4: %+v\n%s", len(got), got, out)
	}
	byPath := map[string]WorktreeRecord{}
	for _, r := range got {
		byPath[clsReal(t, r.Path)] = r
	}
	check := func(path, branch string, detached, locked bool) {
		t.Helper()
		r, ok := byPath[clsReal(t, path)]
		if !ok {
			t.Errorf("no record for %s in %+v", path, got)
			return
		}
		if r.Branch != branch || r.Detached != detached || r.Locked != locked {
			t.Errorf("%s = branch %q detached %v locked %v, want %q %v %v", path, r.Branch, r.Detached, r.Locked, branch, detached, locked)
		}
		if len(r.Head) != 40 {
			t.Errorf("%s HEAD = %q, want a 40-character commit id", path, r.Head)
		}
	}
	check(repo, "main", false, false)
	check(branchTree, "WT-real", false, false)
	check(detachedTree, "", true, false)
	check(lockedTree, "WT-held", false, true)
}

// TestClassify_ListingEndToEnd runs the parser and the classifier together
// over a listing assembled from the world's real directories, the way
// `moai worktree list` and `moai worktree migrate` will (one shared function).
func TestClassify_ListingEndToEnd(t *testing.T) {
	w := clsNewWorld(t)
	type shape struct{ path, branch, class string }
	shapes := []shape{
		{w.primary, "main", "primary"},
		{filepath.Join(w.moaiRoot, "t1"), "WT-a", "moai-made"},
		{filepath.Join(w.moaiRoot, "release-v314"), "release/v3.1.4", "moai-made"},
		{filepath.Join(w.claudeRoot, "agent-ab12cd"), "worktree-agent-ab12cd", "claude-native"},
		{filepath.Join(w.claudeRoot, "main"), "main", "integration"},
		{filepath.Join(w.claudeRoot, "audit-gate-async-worktree-doc"), "worktree-audit-gate-async-worktree-doc", "unclassified"},
		{filepath.Join(w.codex, "ab12", "repo"), "WT-c", "codex-managed"},
		{filepath.Join(w.ownHome, "rc30-preview"), "rc30", "moai-made"},
	}
	var doc strings.Builder
	for i, s := range shapes {
		clsMkdir(t, s.path)
		fmt.Fprintf(&doc, "worktree %s\nHEAD %040d\nbranch refs/heads/%s\n\n", s.path, i+1, s.branch)
	}
	recs, err := ParseWorktreeList(doc.String())
	if err != nil {
		t.Fatalf("ParseWorktreeList: %v", err)
	}
	if len(recs) != len(shapes) {
		t.Fatalf("parsed %d records, want %d", len(recs), len(shapes))
	}
	for i, r := range recs {
		if got := Classify(r, w.ctx); string(got.Class) != shapes[i].class {
			t.Errorf("%s = %q, want %q", shapes[i].path, got.Class, shapes[i].class)
		}
	}
}

// clsGit runs one git command in dir and fails the test on error.
func clsGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}
