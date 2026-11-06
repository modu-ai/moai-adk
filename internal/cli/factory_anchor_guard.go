package cli

// factory_anchor_guard.go — SPEC-LAUNCHER-ARGV-001 M2: the leader-launch anchor
// guard (REQ-LA-004..009).
//
// A factory leader launches from the primary checkout, not from a card
// worktree. The guard refuses a leader launch whose tree resolves to a card
// worktree (a branch with the WT- prefix) — BEFORE any session mutation (the
// role marker, the run record, the settings preparation, the worktree
// materialization all wait behind it).
//
// What the guard judges is the launch VIEW (launch_view.go), never raw
// arguments: the final backend argv after the launcher's own consumption, with
// the backend's own option model deciding which worktree option wins. The
// guard owns no argv reader. That is the redesign: card t1604 re-read the raw
// argv with hand-written readers that predicted what the launcher would consume
// and what claude would select, and each audit round found one more spelling on
// which the prediction was wrong.
//
// The non-argv semantics below (override flag with an audit row, audit-log
// hygiene, lane/non-factory exemption, fail-open, env-label independence) are
// ported from the preserved tree of card t1604.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// allowCardAnchorFlag is the documented explicit override of the leader anchor
// guard (REQ-LA-005): a leader launch anchored in a card worktree proceeds only
// when the operator passes it, and the bypass is recorded in the audit log.
// Silent bypass is prohibited.
const allowCardAnchorFlag = "--allow-card-anchor"

// leaderAnchorRefusalFormat is the REQ-LA-004 refusal: it names the detected
// anchor tree, its checked-out branch, and the correction (launch from the
// primary checkout), and offers the documented override.
const leaderAnchorRefusalFormat = "factory leader launch refused: %s is anchored on card branch %q (WT- prefix) — a leader session launches from the primary checkout, not a card worktree; pass %s to override"

// leaderAnchorVerdict is a detected card-tree anchor: the tree and the branch
// that made it one.
type leaderAnchorVerdict struct {
	tree   string
	branch string
}

// anchorBranchOfTree reads the checked-out branch of dir; an unreadable tree
// (not a git repository, a failed git call, a detached HEAD) reads as no
// branch — the guard refuses on positive evidence only (REQ-LA-008).
//
// The git child runs with the repository-scoping variables scrubbed
// (gitenv.Env): an inherited GIT_DIR outranks `git -C`, so without the scrub the
// lookup would read the branch of whatever checkout that variable names, not of
// dir (card t1604).
func anchorBranchOfTree(dir string) string {
	cmd := exec.Command("git", "-C", dir, "branch", "--show-current")
	cmd.Env = gitenv.Env()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// shortWorktreeRoot is the project root a short -w name resolves under at
// launch: the repository top level the launch directory belongs to, where
// claude resolves the same name (card t1604, codex-517). A nested .moai below
// the working directory used to pull findProjectRootFn's nearest-root lookup
// into a subdirectory, pointing the check at a tree claude never enters. When
// git cannot name a top level, the older lookups apply in order:
// findProjectRootFn's root, then root — the fallback that keeps a non-repo
// launch directory working.
func shortWorktreeRoot(root string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Env = gitenv.Env()
	if out, err := cmd.Output(); err == nil {
		if top := strings.TrimSpace(string(out)); top != "" {
			return top
		}
	}
	if found, err := findProjectRootFn(); err == nil && found != "" {
		return found
	}
	return root
}

// resolveWorktreeLaunchTarget maps a worktree option value to the tree path the
// backend would launch into: an absolute path names the tree directly; a short
// name resolves across the three roots (SPEC-WORKTREE-HOME-ROOT-001 M3,
// REQ-WHR-007) and maps to its single tree, while a name at no tree keeps the
// legacy .claude spelling and an ambiguous name falls back to it (the anchor
// guard reports what it can; REQ-WHR-022's no-guessing rule binds the
// launching lookups). A value naming no existing tree cannot be an anchor —
// anchorBranchOfTree reads it as empty.
func resolveWorktreeLaunchTarget(root, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	if cands, err := worktreeroot.Candidates(root, value); err == nil && len(cands) == 1 {
		return cands[0].Path
	}
	return filepath.Join(shortWorktreeRoot(root), ".claude", "worktrees", value)
}

// leaderLaunchDir is the directory a leader backend starts in: the process
// working directory. The launcher never changes it — the backend replaces the
// process and inherits it — so the anchor guard reads its branch. It does not
// read CLAUDE_PROJECT_DIR (launchProjectRoot), which names the project and can
// name a different checkout from the one the backend launches into.
// launchProjectRoot is the fallback only when the working directory cannot be
// read.
func leaderLaunchDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return launchProjectRoot()
}

// judgeLeaderAnchor judges a leader launch view and returns the detected card
// anchor, or nil. The discriminator is the branch prefix only
// (cardScopeFromBranch) — environment labels are absent from the input by
// construction (REQ-LA-006).
//
// The effective branch (REQ-LA-004) is, in order:
//   - the launch directory's checked-out branch;
//   - for the worktree the backend selects (view.Selection — the LAST option of
//     the final argv): an existing tree's checked-out branch, never the value of
//     a --branch flag (the t295 path reuses an existing tree unchanged); a tree
//     that does not exist yet and that the --branch path will create takes the
//     annotation's branch value, because that is the branch it will be checked
//     out at;
//   - anything unverifiable allows (REQ-LA-008).
func judgeLeaderAnchor(view launchView, launchDir, root string) *leaderAnchorVerdict {
	if branch := anchorBranchOfTree(launchDir); cardScopeFromBranch(branch) {
		return &leaderAnchorVerdict{tree: launchDir, branch: branch}
	}
	sel := view.Selection
	if !sel.Present || !sel.HasValue {
		return nil
	}
	target := resolveWorktreeLaunchTarget(root, sel.Value)
	if branch := anchorBranchOfTree(target); branch != "" {
		if cardScopeFromBranch(branch) {
			return &leaderAnchorVerdict{tree: target, branch: branch}
		}
		return nil
	}
	if _, err := os.Stat(target); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil // the tree exists (or cannot be read) but names no branch: unverifiable
	}
	// The tree does not exist: only the --branch path can create it, and only
	// under the short name it materializes.
	ann := view.Branch
	if ann.Present && ann.Err == nil && !filepath.IsAbs(sel.Value) &&
		view.StageB.Has && view.StageB.Value == sel.Value && cardScopeFromBranch(ann.Branch) {
		return &leaderAnchorVerdict{tree: target, branch: ann.Branch}
	}
	return nil
}

// leaderAnchorAuditOpenRace is a test seam: openLeaderAnchorAuditLog calls it
// between the Lstat check and the os.OpenFile call, the window in which the path
// can be swapped for a link or a FIFO (card t1604, codex-690). Production leaves
// it a no-op; the race test replaces it to swap the path at exactly that point.
var leaderAnchorAuditOpenRace = func(path string) {}

// openLeaderAnchorAuditLog opens the override audit log for an append, refusing
// a path that exists and is not a regular file (REQ-LA-009). os.OpenFile follows
// a symbolic link, so a link would send the row to another file, and it blocks
// on a FIFO until a reader appears, which would hang the leader launch. Lstat
// reads the path without following it, and the refusal happens before any open.
// The Lstat leaves a window before the open, so the open itself carries
// auditLogOpenNoFollow: O_NOFOLLOW makes the kernel refuse a link swapped in at
// the final path component, and O_NONBLOCK keeps a FIFO swapped in from blocking
// the open. O_NOFOLLOW covers the final component only; the directory
// components are checked before MkdirAll in appendLeaderAnchorAuditRow. Off
// Unix auditLogOpenNoFollow is zero, so the Lstat check is the only guard
// there. A reader-attached FIFO swapped into the window passes the Lstat and
// survives O_NOFOLLOW|O_NONBLOCK, so the open re-reads the type of the
// descriptor it would write through and refuses a non-regular file.
//
// @MX:DEBT: [AUTO] the parent directories (.moai and .moai/logs) are verified by path in appendLeaderAnchorAuditRow and the log file is then opened by path, so a swap of .moai/logs between that check and this open can redirect the append
// @MX:CEILING: REQ-LA-009's three repairs only — a link or FIFO at the final path component and the descriptor's file type; a directory component swapped inside the check-to-open window is not covered
// @MX:UPGRADE: open .moai/logs once with no-follow and use an openat-style relative open of the log file, when the audit log is judged to need protection against a local attacker who can write the project's .moai directory
func openLeaderAnchorAuditLog(path string) (*os.File, error) {
	fi, err := os.Lstat(path)
	switch {
	case err == nil:
		if !fi.Mode().IsRegular() {
			return nil, &os.PathError{Op: "open", Path: path, Err: fmt.Errorf("not a regular file (mode %s)", fi.Mode())}
		}
	case !errors.Is(err, os.ErrNotExist):
		return nil, err
	}
	leaderAnchorAuditOpenRace(path)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|auditLogOpenNoFollow, 0o644)
	if err != nil {
		return nil, err
	}
	if info, statErr := f.Stat(); statErr != nil || !info.Mode().IsRegular() {
		_ = f.Close()
		if statErr != nil {
			return nil, statErr
		}
		return nil, &os.PathError{Op: "open", Path: path, Err: fmt.Errorf("not a regular file (mode %s)", info.Mode())}
	}
	return f, nil
}

// writeLeaderAnchorAuditRow writes one formatted row to the audit log. It is a
// variable so a test can force a write failure: a regular file cannot produce a
// write error on every OS, and a link or FIFO is refused before the open.
var writeLeaderAnchorAuditRow = func(f *os.File, row string) error {
	_, err := io.WriteString(f, row)
	return err
}

// appendLeaderAnchorAuditRow records one override bypass (REQ-LA-005) and
// returns the write error, so the caller reports a bypass that was NOT recorded
// instead of claiming it was. The trail gates the card-tree override only: a
// failed write refuses the bypass, and the caller decides that.
func appendLeaderAnchorAuditRow(root, backend string, verdict *leaderAnchorVerdict) error {
	path := filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath)
	dir := filepath.Dir(path)
	// MkdirAll follows a symbolic link at an existing ancestor and creates the
	// missing directories through it, so the ancestors are refused before it runs.
	// A path that does not exist yet is not a link.
	for _, ancestor := range []string{filepath.Dir(dir), dir} {
		fi, err := os.Lstat(ancestor)
		switch {
		case err == nil:
			if fi.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to write the audit row: %s is a symbolic link", ancestor)
			}
		case !errors.Is(err, os.ErrNotExist):
			return err
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write the audit row: %s is a symbolic link", dir)
	}
	f, err := openLeaderAnchorAuditLog(path)
	if err != nil {
		return err
	}
	werr := writeLeaderAnchorAuditRow(f, fmt.Sprintf("[%s] backend=%s tree=%q branch=%q override=%s\n",
		time.Now().UTC().Format(time.RFC3339), backend, verdict.tree, verdict.branch, allowCardAnchorFlag))
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// leaderAnchorGuardProbe is a test seam called as the guard starts. REQ-LA-004
// places the guard before every session mutation; the placement test reads the
// environment, the run record and the settings directory at exactly this point,
// so a guard moved behind a mutation shows the mutation. Production leaves it a
// no-op.
var leaderAnchorGuardProbe = func() {}

// enforceLeaderAnchorGuard refuses a leader launch anchored in a card worktree
// BEFORE any session mutation (REQ-LA-004). With the documented override flag
// present (already consumed by consumeAllowCardAnchor), an anchored verdict
// proceeds with a stderr warning and an audit row in the project root
// (REQ-LA-005). The success text is printed only after the row is written. A
// failed write refuses the launch (fail-closed): the bypass must be recorded, so
// a bypass that cannot be recorded does not proceed unrecorded.
// launchDir is the directory the backend starts in; root is the project root a
// short -w name resolves under, as the fallback behind shortWorktreeRoot.
func enforceLeaderAnchorGuard(backend, launchDir, root string, view launchView, override bool, warn io.Writer) error {
	leaderAnchorGuardProbe()
	verdict := judgeLeaderAnchor(view, launchDir, root)
	if verdict == nil {
		return nil
	}
	if !override {
		return fmt.Errorf(leaderAnchorRefusalFormat, verdict.tree, verdict.branch, allowCardAnchorFlag)
	}
	_, _ = fmt.Fprintf(warn, "factory leader launch: card-tree anchor override — launching from %s (branch %q)\n",
		verdict.tree, verdict.branch)
	if err := appendLeaderAnchorAuditRow(root, backend, verdict); err != nil {
		_, _ = fmt.Fprintf(warn, "factory leader launch: REFUSED — the bypass was NOT recorded in %s (%v); the launch is refused without an audit row\n",
			config.DefaultLeaderAnchorAuditRelPath, err)
		return fmt.Errorf("factory leader launch refused: the card-tree anchor bypass could not be recorded in %s (%w)",
			filepath.Join(root, config.DefaultLeaderAnchorAuditRelPath), err)
	}
	_, _ = fmt.Fprintf(warn, "factory leader launch: bypass recorded in %s\n", config.DefaultLeaderAnchorAuditRelPath)
	return nil
}
