package worktree

// classify.go — SPEC-WORKTREE-HOME-ROOT-001 M1 (REQ-WHR-013, decision D-2).
//
// One shared classifier over `git worktree list --porcelain` records, built
// here so that every later consumer (M3 tier classification, M4 `moai
// worktree list`, M6 migration selection, M7 retirement) calls a function
// that already exists; list and migrate cannot disagree.
//
// Classify is a pure function of its arguments apart from the filesystem
// identity comparison behind spelling-safe path equality (worktreeroot.Same
// and the containment walk below); the owner markers arrive as data
// (ClassifyContext.MarkerOwners), not as files read inside the function.
// unclassified is a first-class result: such trees are listed and never
// moved. Witness (c) keeps a moved non-WT- tree (rc30-preview,
// release-v314, t1515-absorb, t1868-publish) moai-made through its owner
// marker; witness (b) is the WT- branch; witness (a) is the legacy-moai
// root itself. A Claude-native tree is any agent-* directory under
// .claude/worktrees (0.6.0 ruling).

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// TreeClass is the creation class of one worktree (REQ-WHR-013).
type TreeClass string

// The classes. moai-made trees are the migration's selection; every other
// class is listed and never moved.
const (
	ClassMoaiMade     TreeClass = "moai-made"
	ClassClaudeNative TreeClass = "claude-native"
	ClassCodexManaged TreeClass = "codex-managed"
	ClassIntegration  TreeClass = "integration"
	ClassPrimary      TreeClass = "primary"
	ClassUnclassified TreeClass = "unclassified"
)

// RootKind names which root the tree sits under (location fact, independent
// of the class).
type RootKind string

// The root kinds. "other" is any path under none of the known roots.
const (
	RootHome         RootKind = "home"
	RootLegacyMoai   RootKind = "legacy-moai"
	RootLegacyClaude RootKind = "legacy-claude"
	RootCodexManaged RootKind = "codex-managed"
	RootOther        RootKind = "other"
)

// WorktreeRecord is one stanza of `git worktree list --porcelain`.
type WorktreeRecord struct {
	Path     string // absolute path of the tree
	Head     string // commit id, display-only
	Branch   string // short branch name; empty when detached or bare
	Detached bool
	Bare     bool
	Locked   bool
	Prunable bool
}

// ClassifyContext carries the caller's surroundings: everything that is not
// the record itself.
type ClassifyContext struct {
	PrimaryRoot       string            // primary checkout of this repository
	IntegrationBranch string            // configured integration branch, for example "main"
	HomeProjectsDir   string            // <MoaiHome>/worktrees/projects
	OwnerID           string            // this repository's owner identity
	MarkerOwners      map[string]string // project-key folder -> owner identity of its marker
	CodexWorktreesDir string            // the Codex-managed pool (~/.codex/worktrees)
}

// Classification is the classifier's verdict for one record.
type Classification struct {
	Class TreeClass
	Root  RootKind
}

// ParseWorktreeList parses one `git worktree list --porcelain` document:
// stanzas separated by blank lines, each carrying a worktree path and any of
// HEAD, branch (recorded as the short name), detached, bare, locked (with or
// without a reason) and prunable. An empty document parses to no records and
// no error.
func ParseWorktreeList(porcelain string) ([]WorktreeRecord, error) {
	var out []WorktreeRecord
	var cur *WorktreeRecord
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		head, rest, _ := strings.Cut(line, " ")
		rest = strings.TrimSpace(rest)
		switch head {
		case "worktree":
			flush()
			cur = &WorktreeRecord{Path: filepath.FromSlash(rest)}
		case "HEAD":
			if cur != nil {
				cur.Head = rest
			}
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(rest, "refs/heads/")
			}
		case "detached":
			if cur != nil {
				cur.Detached = true
			}
		case "bare":
			if cur != nil {
				cur.Bare = true
			}
		case "locked":
			if cur != nil {
				cur.Locked = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	flush()
	return out, nil
}

// @MX:ANCHOR: [AUTO] shared tree classifier — list, migrate, tier and retirement must never disagree
// @MX:REASON: [AUTO] fan_in spans M3 done.go/sweep.go, M4 list, M6 selection and M7 retirement; a diverging second classifier re-splits one tree class into two verdicts (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-013)
func Classify(rec WorktreeRecord, ctx ClassifyContext) Classification {
	if ctx.PrimaryRoot != "" && worktreeroot.Same(rec.Path, ctx.PrimaryRoot) {
		return Classification{Class: ClassPrimary}
	}
	if underPath(rec.Path, ctx.CodexWorktreesDir) {
		return Classification{Class: ClassCodexManaged, Root: RootCodexManaged}
	}
	if ctx.PrimaryRoot != "" {
		claudeRoot := filepath.Join(ctx.PrimaryRoot, ".claude", "worktrees")
		if underPath(rec.Path, claudeRoot) && strings.HasPrefix(filepath.Base(rec.Path), "agent-") {
			return Classification{Class: ClassClaudeNative, Root: RootLegacyClaude}
		}
	}
	if ctx.IntegrationBranch != "" && rec.Branch == ctx.IntegrationBranch {
		return Classification{Class: ClassIntegration, Root: rootKindOf(rec.Path, ctx)}
	}
	// Witness (a): everything under the legacy-moai root is moai-made.
	if ctx.PrimaryRoot != "" && underPath(rec.Path, filepath.Join(ctx.PrimaryRoot, ".moai", "worktrees")) {
		return Classification{Class: ClassMoaiMade, Root: RootLegacyMoai}
	}
	// Witness (b): the WT- branch prefix.
	if strings.HasPrefix(rec.Branch, "WT-") {
		return Classification{Class: ClassMoaiMade, Root: rootKindOf(rec.Path, ctx)}
	}
	// Witness (c): a home-root tree whose project-folder marker is ours —
	// this keeps a moved non-WT- tree moai-made.
	if underPath(rec.Path, ctx.HomeProjectsDir) && ownerOf(ctx, homeKeyOf(rec.Path, ctx)) == ctx.OwnerID {
		return Classification{Class: ClassMoaiMade, Root: RootHome}
	}
	return Classification{Class: ClassUnclassified, Root: rootKindOf(rec.Path, ctx)}
}

// rootKindOf reports the root the path sits under, a location fact
// independent of the class.
func rootKindOf(path string, ctx ClassifyContext) RootKind {
	switch {
	case underPath(path, ctx.HomeProjectsDir):
		return RootHome
	case ctx.PrimaryRoot != "" && underPath(path, filepath.Join(ctx.PrimaryRoot, ".moai", "worktrees")):
		return RootLegacyMoai
	case ctx.PrimaryRoot != "" && underPath(path, filepath.Join(ctx.PrimaryRoot, ".claude", "worktrees")):
		return RootLegacyClaude
	case underPath(path, ctx.CodexWorktreesDir):
		return RootCodexManaged
	default:
		return RootOther
	}
}

// ownerOf looks a project key up in the marker-owner map, exact first and
// case-insensitively second (spelling-safe, REQ-WHR-023).
func ownerOf(ctx ClassifyContext, key string) string {
	if key == "" {
		return ""
	}
	if v, ok := ctx.MarkerOwners[key]; ok {
		return v
	}
	for k, v := range ctx.MarkerOwners {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

// homeKeyOf returns the project-key segment of a path under the home
// projects directory (the first segment below it), "" when there is none.
func homeKeyOf(path string, ctx ClassifyContext) string {
	if ctx.HomeProjectsDir == "" {
		return ""
	}
	rp := evalOrClean(path)
	rr := evalOrClean(ctx.HomeProjectsDir)
	rel, err := filepath.Rel(rr, rp)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		// Spelling skew between the two spellings of one directory
		// (case-aliased volume): retry on the case-folded spellings.
		lrp, lrr := strings.ToLower(rp), strings.ToLower(rr)
		if !strings.HasPrefix(lrp, lrr+string(filepath.Separator)) {
			return ""
		}
		rel = rp[len(rr)+1:]
	}
	return strings.SplitN(filepath.ToSlash(rel), "/", 2)[0]
}

// underPath reports whether path lies strictly inside root, spelling-safe:
// containment is decided by directory identity (same-file ancestors), so a
// case-aliased spelling of the root still contains its trees on a
// case-insensitive volume while two case-differing directories stay distinct
// on a case-sensitive one (REQ-WHR-023).
func underPath(path, root string) bool {
	if path == "" || root == "" {
		return false
	}
	rr := evalOrClean(root)
	rp := evalOrClean(path)
	if rp == rr {
		return false
	}
	if strings.HasPrefix(rp, rr+string(filepath.Separator)) {
		return true
	}
	stRoot, err := os.Stat(rr)
	if err != nil {
		return false
	}
	// Walk the ancestors of path upward; any ancestor denoting the same
	// directory as root proves containment regardless of spelling.
	for {
		parent := filepath.Dir(rp)
		if parent == rp {
			return false
		}
		rp = parent
		fi, err := os.Stat(rp)
		if err != nil {
			if os.IsNotExist(err) {
				return false
			}
			continue
		}
		if os.SameFile(fi, stRoot) {
			return true
		}
	}
}

func evalOrClean(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}
