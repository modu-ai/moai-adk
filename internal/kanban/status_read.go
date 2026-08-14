// status_read.go — the branch-side status read
// (SPEC-KANBAN-BOARD-001 REQ-KB-020/024, M2).
//
// The board reads a card's frontmatter `status` from the card's BRANCH,
// without checking it out: a blob read (`git show <ref>:<path>`) needs no
// worktree and no fetch, because refs are shared across every checkout of
// one repository. Which branch: the one the card's WORKTREE reports —
// recognized by SPEC-KANBAN-WORKTREE-001 REQ-KW-003's exact-token rule (the
// segment after the type prefix begins with the SPEC identifier; the next
// character is either absent or a hyphen) — never a name re-derived here and
// never a SEARCH of the branch set, since cards keep every branch they ever
// carried and a search is what REQ-KW-019 refuses on more than one match.
//
// The selector for the source is worktree LIVENESS, not branch existence
// (AP-23/28/28a): branch existence essentially never changes, while the
// worktree spans creation to both-pull-requests-merged — exactly the
// interval in which the primary-side copy is stale. Where no worktree is
// live, the primary checkout supplies the value (a card not yet cut, a
// disposed card whose content has merged, a card whose branch was deleted
// after merge). Committed state only: a transition written but not committed
// is not yet a transition.
package kanban

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Status source tags recorded on CardStatus.Source.
const (
	StatusSourceBranch     = "branch"
	StatusSourcePrimary    = "primary"
	StatusSourceUnresolved = "unresolved"
)

// WorktreeBases names the two scanned bases beneath which a card's worktree
// lives (REQ-KW-003's "either scanned base" and the path-identity rule: the
// path's final segment IS the card's SPEC identifier).
type WorktreeBases struct {
	// Claude is the project-local base (<primaryRoot>/.claude/worktrees).
	Claude string
	// MoAI is the persistent user-level base (<home>/.moai/worktrees).
	MoAI string
}

// DefaultWorktreeBases resolves the scanned bases for production callers.
func DefaultWorktreeBases(primaryRoot string) WorktreeBases {
	bases := WorktreeBases{Claude: filepath.Join(primaryRoot, ".claude", "worktrees")}
	if home, err := os.UserHomeDir(); err == nil {
		bases.MoAI = filepath.Join(home, ".moai", "worktrees")
	}
	return bases
}

// CardStatus is the source-resolved status of one card.
type CardStatus struct {
	// Status is a canonical enum value, or StatusUnresolved where the source
	// resolved to no single tree. Empty means the card has no spec.md at all
	// (the backlog admission case, legal in backlog and plan).
	Status string
	// Source names where Status came from: the branch a live worktree
	// reports, the primary checkout, or unresolved.
	Source string
	// SpecFilePresent is false where the read source carries no spec.md for
	// the card.
	SpecFilePresent bool
	// WorktreePath names the observed live worktree, when one exists.
	WorktreePath string
	// WorktreeBranch names the branch that worktree reported, when it
	// reported one.
	WorktreeBranch string
	// Candidates surfaces every candidate the resolution found, when the
	// source is unresolved (REQ-KB-024). The observation route reads only
	// what a live worktree reports, so it is empty there; the field exists
	// so a caller that reaches a refused resolution can surface what it saw.
	Candidates []string
}

// @MX:ANCHOR: [AUTO] ReadCardStatus — the branch-vs-primary status source selection (REQ-KB-020)
// @MX:REASON: expected fan_in >= 3 (ReconcileCard, the review column's verify gate, any board rendering); selecting the primary while a worktree is live reports a stale draft for the whole run interval — the v0.2.0 defect
//
// ReadCardStatus resolves a card's status source and reads the value
// (REQ-KB-020), reporting unresolved where the source resolves to no single
// tree (REQ-KB-024).
func ReadCardStatus(primaryRoot, specID string, bases WorktreeBases) (*CardStatus, error) {
	if specID == "" {
		return nil, fmt.Errorf("read card status: empty spec id")
	}

	// The live-worktree observation, over both scanned bases. The path's
	// final segment is the card's identity, so the candidate worktree for a
	// card is deterministic; its REPORTED branch decides the source.
	for _, base := range []string{bases.MoAI, bases.Claude} {
		if base == "" {
			continue
		}
		wtPath := filepath.Join(base, specID)
		if !isGitWorktreeDir(wtPath) {
			continue
		}
		branch, err := reportedBranch(wtPath)
		if err != nil {
			// The worktree exists but its HEAD cannot be read as a branch
			// reference — reporting no branch, the unresolved case.
			return &CardStatus{
				Status:          StatusUnresolved,
				Source:          StatusSourceUnresolved,
				SpecFilePresent: true,
				WorktreePath:    wtPath,
			}, nil
		}
		if branch == "" {
			// Detached HEAD: a worktree existing but reporting no branch.
			return &CardStatus{
				Status:          StatusUnresolved,
				Source:          StatusSourceUnresolved,
				SpecFilePresent: true,
				WorktreePath:    wtPath,
			}, nil
		}
		if !branchNamesSpec(branch, specID) {
			// The worktree's branch does not name this card ("when and only
			// when", REQ-KW-003) — this tree is not this card's worktree;
			// keep looking, and fall back to the primary.
			continue
		}
		return readBranchStatus(primaryRoot, specID, branch, wtPath)
	}

	// No live worktree: the primary checkout supplies the value.
	return readPrimaryStatus(primaryRoot, specID)
}

// branchNamesSpec implements REQ-KW-003's recognition rule: a branch names
// the card when and only when the segment following its type prefix BEGINS
// with the card's SPEC identifier and the character immediately after it is
// either absent or a hyphen — so a phase-suffixed branch is recognized and a
// branch merely embedding the identifier as a longer token is not.
func branchNamesSpec(branch, specID string) bool {
	if branch == "" || specID == "" {
		return false
	}
	seg := branch
	if i := strings.LastIndex(branch, "/"); i >= 0 {
		seg = branch[i+1:]
	}
	if seg == specID {
		return true
	}
	return strings.HasPrefix(seg, specID) && len(seg) > len(specID) && seg[len(specID)] == '-'
}

// reportedBranch returns the branch the worktree at wtPath reports, or ""
// for a detached HEAD. An unreadable HEAD reference is an error.
func reportedBranch(wtPath string) (string, error) {
	cmd := exec.Command("git", "-C", wtPath, "symbolic-ref", "--quiet", "--short", "HEAD")
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		// symbolic-ref exits non-zero on a detached HEAD: the worktree
		// reports no branch, which is not an error condition.
		if strings.Contains(errBuf.String(), "not a symbolic ref") || strings.Contains(strings.ToLower(err.Error()), "symbolic") {
			return "", nil
		}
		return "", fmt.Errorf("reported branch of %s: %w (%s)", wtPath, err, strings.TrimSpace(errBuf.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// readBranchStatus performs the blob read against the branch, without
// checkout and without fetch: refs are shared across every checkout of the
// repository, so the primary root can read any branch a worktree holds.
func readBranchStatus(primaryRoot, specID, branch, wtPath string) (*CardStatus, error) {
	rel := fmt.Sprintf(".moai/specs/%s/spec.md", specID)
	cmd := exec.Command("git", "-C", primaryRoot, "show", fmt.Sprintf("%s:%s", branch, rel))
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		// The branch exists but carries no spec.md for the card: the
		// no-file case read from the branch side.
		return &CardStatus{
			Status:          "",
			Source:          StatusSourceBranch,
			SpecFilePresent: false,
			WorktreePath:    wtPath,
			WorktreeBranch:  branch,
		}, nil
	}
	status, ok := parseFrontmatterStatus(out.Bytes())
	return &CardStatus{
		Status:          status,
		Source:          StatusSourceBranch,
		SpecFilePresent: ok,
		WorktreePath:    wtPath,
		WorktreeBranch:  branch,
	}, nil
}

// readPrimaryStatus reads the primary checkout's copy of the card's spec.md.
func readPrimaryStatus(primaryRoot, specID string) (*CardStatus, error) {
	path := filepath.Join(primaryRoot, ".moai", "specs", specID, "spec.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CardStatus{Status: "", Source: StatusSourcePrimary, SpecFilePresent: false}, nil
		}
		return nil, fmt.Errorf("read card status: primary copy: %w", err)
	}
	status, ok := parseFrontmatterStatus(raw)
	return &CardStatus{Status: status, Source: StatusSourcePrimary, SpecFilePresent: ok}, nil
}

// frontmatterStatusLine matches the frontmatter `status:` key.
var frontmatterStatusLine = regexp.MustCompile(`(?m)^status:[ \t]*(\S+)[ \t]*$`)

// parseFrontmatterStatus extracts the frontmatter status value; ok is false
// when the document carries no status key.
func parseFrontmatterStatus(raw []byte) (string, bool) {
	m := frontmatterStatusLine.FindSubmatch(raw)
	if m == nil {
		return "", false
	}
	return string(m[1]), true
}

// isGitWorktreeDir reports whether path exists and holds a .git file (a
// linked worktree's git pointer) or directory (a primary checkout shape).
func isGitWorktreeDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	gitEntry := filepath.Join(path, ".git")
	if _, err := os.Stat(gitEntry); err == nil {
		return true
	}
	return false
}
