// agentmemory_test.go — reconciliation core for the agent-memory drain/mirror
// (SPEC-AGENT-MEMORY-DRAIN-001).
//
// All fixtures live under t.TempDir(): a primary store directory and one or
// more worktree-like directories. No real sibling tree is ever touched, and
// no git invocation is needed for the drain core (primary resolution has its
// own seam, tested separately with a real git fixture).
package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// seedAgentStore writes agent-memory content under root/.claude/agent-memory.
// files maps agent-relative slash paths to file content; a "MEMORY.md" entry
// per agent seeds that agent's worktree index.
func seedAgentStore(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		dst := filepath.Join(root, ".claude", "agent-memory", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dst, err)
		}
		if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}
}

const feedbackXBody = "---\nname: feedback-x\ndescription: one-line summary of feedback x\ntype: feedback\n---\nbody x\n"

// seedWorktreeX is the canonical AC-AM-002 fixture: one agent topic file plus
// its worktree index line.
func seedWorktreeX(t *testing.T, tree string) {
	t.Helper()
	seedAgentStore(t, tree, map[string]string{
		"manager-spec/feedback_x.md": feedbackXBody,
		"manager-spec/MEMORY.md":     "# Memory Index\n\n- [feedback x](feedback_x.md) — wt hook line\n",
	})
}

func readFileOrFatal(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// treeChecksums fingerprints every file under root so a before/after compare
// proves the drain never mutated worktree content (AC-AM-007).
func treeChecksums(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return fmt.Errorf("rel %s: %w", path, relErr)
		}
		out[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

func requireEqualMaps(t *testing.T, want, got map[string]string) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("map size differ: want %d entries, got %d (%v)", len(want), len(got), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("key %q: want %q, got %q", k, v, got[k])
		}
	}
}

// TestDrainTreeCopiesTopicFileAndKeepsSource covers AC-AM-002 and AC-AM-007:
// the copy lands byte-identical in the primary store, the worktree copy (and
// its whole tree) is unchanged, and the worktree index is never copied.
func TestDrainTreeCopiesTopicFileAndKeepsSource(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedWorktreeX(t, tree)
	before := treeChecksums(t, tree)

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}

	got := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md"))
	if got != feedbackXBody {
		t.Errorf("primary copy differs:\nwant %q\ngot  %q", feedbackXBody, got)
	}
	requireEqualMaps(t, before, treeChecksums(t, tree))

	// The worktree index file is never copied byte-identical: the primary
	// agent index is created by index reconciliation (header + appended
	// line), never by copying the worktree's index.
	primaryIdx := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "MEMORY.md"))
	if primaryIdx == "# Memory Index\n\n- [feedback x](feedback_x.md) — wt hook line\n" {
		t.Error("the worktree MEMORY.md was copied wholesale into the primary store")
	}
	if !strings.Contains(primaryIdx, "feedback_x.md") {
		t.Errorf("primary index missing the appended line for the drained topic:\n%s", primaryIdx)
	}

	if rec.Copied != 1 || rec.Files != 1 {
		t.Errorf("record: files=%d copied=%d, want 1/1 (%+v)", rec.Files, rec.Copied, rec)
	}
}

// TestDrainTreeCollisionNeverOverwrites covers AC-AM-003: the primary file
// keeps its exact bytes and the colliding content lands under a
// tree-qualified `.wt-<tree>` name. The mutant "skip collided files
// entirely" fails the suffixed-copy assertion.
func TestDrainTreeCollisionNeverOverwrites(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedAgentStore(t, primary, map[string]string{
		"plan-auditor/feedback_dup.md": "content A",
	})
	seedAgentStore(t, tree, map[string]string{
		"plan-auditor/feedback_dup.md": "content B",
	})

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}

	if got := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "plan-auditor", "feedback_dup.md")); got != "content A" {
		t.Errorf("primary file was overwritten: %q", got)
	}
	suffixed := filepath.Join(primary, ".claude", "agent-memory", "plan-auditor", "feedback_dup.wt-"+filepath.Base(tree)+".md")
	if got := readFileOrFatal(t, suffixed); got != "content B" {
		t.Errorf("tree-qualified copy missing or wrong: %q", got)
	}
	if rec.Collided != 1 || rec.Copied != 0 {
		t.Errorf("record: copied=%d collided=%d, want 0/1 (%+v)", rec.Copied, rec.Collided, rec)
	}
}

// TestDrainTreeAppendsExactlyOneIndexLine covers AC-AM-004: the primary
// agent index gains exactly one line for the copied topic, derived from the
// worktree index line, and a second drain adds nothing more.
func TestDrainTreeAppendsExactlyOneIndexLine(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedAgentStore(t, primary, map[string]string{
		"manager-spec/MEMORY.md":         "# Memory Index\n\n- [other](feedback_other.md) — existing line\n",
		"manager-spec/feedback_other.md": "---\nname: other\n---\n",
	})
	seedWorktreeX(t, tree)

	indexPath := filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "MEMORY.md")
	before := strings.Split(strings.TrimRight(readFileOrFatal(t, indexPath), "\n"), "\n")

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}

	after := strings.Split(strings.TrimRight(readFileOrFatal(t, indexPath), "\n"), "\n")
	if len(after)-len(before) != 1 {
		t.Fatalf("index gained %d line(s), want exactly 1:\n%s", len(after)-len(before), strings.Join(after, "\n"))
	}
	newLine := after[len(after)-1]
	if !strings.Contains(newLine, "feedback_x.md") {
		t.Errorf("new index line does not reference feedback_x.md: %q", newLine)
	}
	// The line derives from the worktree index (its summary text survives).
	if !strings.Contains(newLine, "wt hook line") {
		t.Errorf("new index line not derived from worktree index: %q", newLine)
	}
	// The pre-existing line is untouched.
	if !strings.Contains(strings.Join(after, "\n"), "- [other](feedback_other.md) — existing line") {
		t.Errorf("pre-existing index line changed:\n%s", strings.Join(after, "\n"))
	}
	if rec.IndexLinesAdded != 1 {
		t.Errorf("record: index_lines_added=%d, want 1", rec.IndexLinesAdded)
	}

	// Idempotency: a second drain copies nothing and appends nothing.
	rec2, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree (2nd): %v", err)
	}
	after2 := strings.Split(strings.TrimRight(readFileOrFatal(t, indexPath), "\n"), "\n")
	if len(after2) != len(after) {
		t.Errorf("second drain appended again: %d → %d lines", len(after), len(after2))
	}
	if rec2.Copied != 0 || rec2.IndexLinesAdded != 0 {
		t.Errorf("second drain was not idempotent: %+v", rec2)
	}
}

// TestDrainTreeFrontmatterDerivedIndexLine covers the D.3 edge where the
// worktree index has no line for the file: the appended line derives from the
// file's frontmatter description.
func TestDrainTreeFrontmatterDerivedIndexLine(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedAgentStore(t, tree, map[string]string{
		"manager-develop/feedback_noindex.md": feedbackXBody,
	})

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}
	idx := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-develop", "MEMORY.md"))
	if !strings.Contains(idx, "feedback_noindex.md") {
		t.Fatalf("index missing line for the drained topic:\n%s", idx)
	}
	if !strings.Contains(idx, "one-line summary of feedback x") {
		t.Errorf("index line not derived from frontmatter description:\n%s", idx)
	}
	if rec.IndexLinesAdded != 1 {
		t.Errorf("record: index_lines_added=%d, want 1", rec.IndexLinesAdded)
	}
}

// TestDrainTreeArchiveSubdirCopied covers the D.3 edge for `_archive/`
// content: copied under the same collision rule, never gaining an index line.
func TestDrainTreeArchiveSubdirCopied(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedAgentStore(t, tree, map[string]string{
		"manager-spec/_archive/feedback_old.md": "archived lesson",
	})

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}
	if got := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "_archive", "feedback_old.md")); got != "archived lesson" {
		t.Errorf("archive copy missing or wrong: %q", got)
	}
	if rec.IndexLinesAdded != 0 {
		t.Errorf("archive copy gained an index line: %+v", rec)
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "MEMORY.md")); !os.IsNotExist(err) {
		t.Errorf("an index was created for an archive-only drain (err=%v)", err)
	}
}

// TestDrainTreeSkeletonOnlyReportedNothingCopied covers the 62-of-88 case:
// a tree with empty agent directories is reported, nothing is written.
func TestDrainTreeSkeletonOnlyReportedNothingCopied(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	// Skeleton: agent directories exist, all empty.
	if err := os.MkdirAll(filepath.Join(tree, ".claude", "agent-memory", "manager-spec"), 0o755); err != nil {
		t.Fatalf("mkdir skeleton agent dir: %v", err)
	}

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree: %v", err)
	}
	if rec.Files != 0 || rec.Copied != 0 || rec.Collided != 0 {
		t.Errorf("skeleton tree drained something: %+v", rec)
	}
	if len(rec.Agents) == 0 {
		t.Errorf("skeleton agents not reported: %+v", rec)
	}
	// Nothing may appear under the primary store.
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory")); !os.IsNotExist(err) {
		t.Errorf("drain wrote into the primary store for a skeleton-only tree (err=%v)", err)
	}
}

// TestDrainTreeMissingTreeReported covers a registered-but-gone worktree.
func TestDrainTreeMissingTreeReported(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := filepath.Join(t.TempDir(), "gone")

	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree on a missing tree must not error: %v", err)
	}
	if !rec.Missing {
		t.Errorf("missing tree not flagged: %+v", rec)
	}
}

// TestDrainTreePreviewWritesNothing covers AC-AM-001a's core invariant: the
// preview classifies the copy set and writes nothing.
func TestDrainTreePreviewWritesNothing(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedWorktreeX(t, tree)

	rec, err := DrainTree(primary, tree, false)
	if err != nil {
		t.Fatalf("DrainTree preview: %v", err)
	}
	if rec.Copied != 1 || rec.IndexLinesAdded != 1 {
		t.Errorf("preview did not classify the copy set: %+v", rec)
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md")); !os.IsNotExist(err) {
		t.Errorf("preview wrote into the primary store (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory")); !os.IsNotExist(err) {
		t.Errorf("preview created the primary store (err=%v)", err)
	}
}

// TestDrainTreeTwoWorktreesSameLesson covers the D.3 edge: two trees writing
// the same lesson both land (two tree-qualified copies, two index lines).
func TestDrainTreeTwoWorktreesSameLesson(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	treeA := t.TempDir()
	treeB := t.TempDir()
	seedAgentStore(t, primary, map[string]string{
		"sync-auditor/feedback_dup.md": "primary content",
	})
	for _, tree := range []string{treeA, treeB} {
		seedAgentStore(t, tree, map[string]string{
			"sync-auditor/feedback_dup.md": "tree content " + filepath.Base(tree),
		})
	}

	for _, tree := range []string{treeA, treeB} {
		if _, err := DrainTree(primary, tree, true); err != nil {
			t.Fatalf("DrainTree(%s): %v", tree, err)
		}
	}
	am := filepath.Join(primary, ".claude", "agent-memory", "sync-auditor")
	for _, tree := range []string{treeA, treeB} {
		p := filepath.Join(am, "feedback_dup.wt-"+filepath.Base(tree)+".md")
		if got := readFileOrFatal(t, p); got != "tree content "+filepath.Base(tree) {
			t.Errorf("tree-qualified copy for %s missing or wrong: %q", tree, got)
		}
	}
	idx := readFileOrFatal(t, filepath.Join(am, "MEMORY.md"))
	for _, tree := range []string{treeA, treeB} {
		if !strings.Contains(idx, "feedback_dup.wt-"+filepath.Base(tree)+".md") {
			t.Errorf("index missing line for %s:\n%s", tree, idx)
		}
	}
}

// TestIsAgentMemoryMDPathAnchor pins the D7 trigger anchoring: the predicate
// matches the literal `.claude/agent-memory/` segment only. The audit's
// unanchored substring form must NOT be inherited here.
func TestIsAgentMemoryMDPathAnchor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want bool
	}{
		{"/repo/.claude/agent-memory/manager-spec/feedback_x.md", true},
		{"/repo/.claude/agent-memory/manager-spec/MEMORY.md", true},
		{".claude/agent-memory/a/f.md", true},
		{"/repo/docs/agent-memory/x.md", false},         // unanchored substring — must NOT match
		{"/repo/agent-memory/x.md", false},              // missing .claude segment
		{"/repo/.claude/agent-memory/notes.txt", false}, // not markdown
		{"/repo/.claude/agent-memory-x/f.md", false},    // adjacent name, not the segment
	}
	for _, c := range cases {
		if got := IsAgentMemoryMDPath(c.path); got != c.want {
			t.Errorf("IsAgentMemoryMDPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// TestSplitAgentMemoryPath pins the tree-root/agent/rel decomposition the
// mirror and drain both build on.
func TestSplitAgentMemoryPath(t *testing.T) {
	t.Parallel()
	root, agent, rel, ok := SplitAgentMemoryPath("/repo/wt/.claude/agent-memory/manager-spec/_archive/f.md")
	if !ok {
		t.Fatal("SplitAgentMemoryPath failed on a valid path")
	}
	if root != "/repo/wt" || agent != "manager-spec" || rel != "_archive/f.md" {
		t.Errorf("got (%q,%q,%q), want (/repo/wt, manager-spec, _archive/f.md)", root, agent, rel)
	}
	if _, _, _, ok := SplitAgentMemoryPath("/repo/docs/agent-memory/f.md"); ok {
		t.Error("SplitAgentMemoryPath accepted an unanchored path")
	}
}

// runGitIn runs a git command in dir, failing the test on error. (Named
// runGitIn because the package already carries a runGit test helper.)
func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
	}
}

// TestPrimaryRootOf exercises the real git-based primary resolution against a
// throwaway repository: the primary resolves to itself, and a linked worktree
// resolves back to the primary with inWorktree=true.
func TestPrimaryRootOf(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	primary := t.TempDir()
	// git resolves /var → /private/var on macOS, so compare against the
	// symlink-resolved spelling of the fixture root.
	wantPrimary, evalErr := filepath.EvalSymlinks(primary)
	if evalErr != nil {
		t.Fatalf("eval symlinks: %v", evalErr)
	}
	runGitIn(t, primary, "init", "-q")
	runGitIn(t, primary, "config", "user.email", "t@example.com")
	runGitIn(t, primary, "config", "user.name", "t")
	runGitIn(t, primary, "commit", "-q", "--allow-empty", "-m", "init")

	// Primary itself.
	p1, inWt, err := PrimaryRootOf(primary)
	if err != nil {
		t.Fatalf("PrimaryRootOf(primary): %v", err)
	}
	if inWt {
		t.Errorf("primary reported as worktree")
	}
	if filepath.Clean(p1) != filepath.Clean(wantPrimary) {
		t.Errorf("primary resolved to %q, want %q", p1, primary)
	}

	// A linked worktree resolves back to the primary.
	wt := filepath.Join(t.TempDir(), "wt")
	runGitIn(t, primary, "worktree", "add", "-q", "--detach", wt)
	p2, inWt2, err2 := PrimaryRootOf(wt)
	if err2 != nil {
		t.Fatalf("PrimaryRootOf(worktree): %v", err2)
	}
	if !inWt2 {
		t.Error("linked worktree not detected")
	}
	if filepath.Clean(p2) != filepath.Clean(wantPrimary) {
		t.Errorf("worktree primary resolved to %q, want %q", p2, primary)
	}

	// A non-repo directory is an error, not a guess.
	if _, _, err3 := PrimaryRootOf(t.TempDir()); err3 == nil {
		t.Error("non-repo directory resolved without error")
	}
}
