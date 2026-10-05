// integration_merge_collision.go — the added-path collision check
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001 REQ-MWQ-017 cause 13).
//
// The check refuses a merge whose candidate would overwrite an ignored or
// untracked byte in the integration worktree — the data-loss path the plan
// audit reproduced (a candidate that starts tracking `runtime.local`
// silently destroyed the integration worktree's ignored `runtime.local`
// file with exit 0 and a clean status, because every porcelain gate reads
// ignored files as nothing). It refuses BEFORE any `git merge` call and
// leaves every colliding byte untouched.
//
// WHAT COUNTS, per spec.md:188 as amended v0.8.0/v0.9.0: a "path" is a
// LEAF entry — a file, symlink or submodule entry as `git ls-tree -r`
// lists it, never a tree (directory) entry — and an "added path" is a leaf
// present in the pinned SHA's leaf set and absent from the integration
// branch tip's leaf set. A collision is an added path, any ANCESTOR of an
// added path (path components only, D8 — `runtime` is not an ancestor of
// `runtime.local`), or any path BENEATH an added path, that already exists
// in the integration worktree as an ignored or untracked file or
// directory.
package factory

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/factorylane"
)

// GitlinkMode is the git mode of a submodule entry — a leaf the leaf-set
// reader must keep: dropping it would empty the added-path set exactly
// where the gitlink conversions hide their data loss (codex-P1).
const GitlinkMode = "160000"

// leafSets reads both sides' leaf sets: `git ls-tree -r` lists every file,
// symlink and submodule entry and never a tree entry, so the whole output
// is leaves. The value kept is the MODE CLASS — "gitlink" for 160000,
// "leaf" for everything else — because the type-change detection (P1)
// needs to see a gitlink flipping to a regular entry even when the PATH
// stays a leaf on both sides and the added-path set therefore comes back
// EMPTY.
func readLeafSet(git factorylane.GitRunner, sha string) (map[string]string, error) {
	out, err := git.Git([]string{"ls-tree", "-r", sha}...)
	if err != nil {
		return nil, fmt.Errorf("ls-tree %s: %w", sha, err)
	}
	leaves := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// "<mode> <type> <sha>\t<path>" — the path is everything after the
		// first tab.
		tab := strings.IndexByte(line, '\t')
		if tab < 0 {
			continue
		}
		meta, path := line[:tab], line[tab+1:]
		mode := strings.Fields(meta)
		if len(mode) == 0 {
			continue
		}
		if mode[0] == GitlinkMode {
			leaves[path] = "gitlink"
		} else {
			leaves[path] = "leaf"
		}
	}
	return leaves, nil
}

// worktreeBytesExist reports whether the worktree holds any ignored or
// untracked byte AT path (not beneath it): the path exists on disk and git
// does not track it. lstat, not os.Stat — a symlink at the path is bytes
// the merge may overwrite, and a case-insensitive volume answers an exact
// case probe conservatively (O5: the refusal direction is safe; the
// recorded decision is that a case-insensitive FS refuses MORE, never
// less).
//
// A NOT-A-DIRECTORY lstat reads as absence: the path's ancestor is a file,
// so no byte exists AT the path — the ancestor itself is what the
// ancestor-of-an-added-path rule examines.
func worktreeBytesExist(worktree, path string) (bool, error) {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR) {
			return false, nil
		}
		return false, fmt.Errorf("lstat %s: %w", path, err)
	}
	out, err := gitLsFiles(worktree, path)
	if err != nil {
		return false, err
	}
	// A tracked path (file, symlink, or submodule whose gitlink git
	// tracks) is not worktree debris: the merge manages it. Anything
	// else on disk is ignored or untracked.
	return strings.TrimSpace(out) == "", nil
}

// gitLsFiles answers whether git tracks path, exactly.
func gitLsFiles(worktree, path string) (string, error) {
	runner := factorylane.ExecGitRunner{Dir: worktree}
	return runner.Git("ls-files", "--", path)
}

// untrackedBytesBeneath reports whether the worktree holds ignored or
// untracked bytes BENEATH dir (which exists as a directory): `git clean
// -nd -x` dry-runs the deletion git itself would perform and lists every
// byte it would remove — tracked files are never listed, so a non-empty
// listing is exactly "bytes the merge's directory replacement would
// destroy".
func untrackedBytesBeneath(worktree, dir string) (bool, error) {
	runner := factorylane.ExecGitRunner{Dir: worktree}
	out, err := runner.Git("clean", "-nd", "-x", "--", dir)
	if err != nil {
		return false, fmt.Errorf("git clean -nd -x %s: %w", dir, err)
	}
	return strings.TrimSpace(out) != "", nil
}

// pathComponents returns dir's ancestor components of p ("a/b/c" → "a",
// "a/b") — the PATH-COMPONENT ancestors D8 names. `a/bc` shares a string
// prefix with `a/b` and is deliberately not here.
func pathComponents(p string) []string {
	var components []string
	parts := strings.Split(p, "/")
	for i := 1; i < len(parts); i++ {
		components = append(components, strings.Join(parts[:i], "/"))
	}
	return components
}

// FindAddedPathCollisions returns every colliding path for the merge of
// pinnedSHA into the worktree whose integration tip is tipSHA, or nil when
// the merge touches no ignored or untracked byte.
//
// Four shapes are examined, per spec.md as amended through v0.9.0:
//
//   - the added path ITSELF exists as ignored/untracked (13a);
//   - an ANCESTOR of an added path exists as ignored/untracked (13b) —
//     path components only, D8: `runtime` is not an ancestor of
//     `runtime.local`, so a sibling sharing a string prefix never refuses;
//   - the added path is a DIRECTORY in the worktree holding ignored or
//     untracked bytes beneath it (13c, and 13d's directory the candidate
//     replaces with a file);
//   - the TYPE-CHANGE TARGET (codex-P1): a path that is a gitlink leaf in
//     the tip and a regular leaf in the pinned SHA (gitlink→file — the
//     added-path set comes back EMPTY because the path is a leaf on both
//     sides), or a gitlink leaf in the tip that becomes a directory in the
//     pinned SHA (gitlink→directory — the added paths are the directory's
//     new leaves, and the target itself is what the submodule's local
//     files hide under). The target and its local files beneath it are
//     examined exactly as an added path is.
//
// A leaf-to-file-class REVERSE change (tip leaf, pinned gitlink) is also a
// type-change target and reads conservatively: bytes at the path refuse
// rather than risk them (D9's gitlink-fixture safety).
func FindAddedPathCollisions(worktree, tipSHA, pinnedSHA string) ([]string, error) {
	git := factorylane.ExecGitRunner{Dir: worktree}
	tipLeaves, err := readLeafSet(git, tipSHA)
	if err != nil {
		return nil, err
	}
	pinnedLeaves, err := readLeafSet(git, pinnedSHA)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var colliding []string
	add := func(path string) {
		if !seen[path] {
			seen[path] = true
			colliding = append(colliding, path)
		}
	}

	// One worktree probe for a path: itself, then its ancestor components
	// (D8's path-component ancestors), then — when the path is a directory
	// on disk — the ignored/untracked bytes beneath it.
	examine := func(path string) error {
		exists, err := worktreeBytesExist(worktree, worktreePath(worktree, path))
		if err != nil {
			return err
		}
		if exists {
			add(path)
		}
		for _, ancestor := range pathComponents(path) {
			exists, err := worktreeBytesExist(worktree, worktreePath(worktree, ancestor))
			if err != nil {
				return err
			}
			if exists {
				add(ancestor)
				break // the deeper components live under a colliding ancestor
			}
		}
		info, err := os.Lstat(worktreePath(worktree, path))
		if err == nil && info.IsDir() {
			beneath, err := untrackedBytesBeneath(worktree, worktreePath(worktree, path))
			if err != nil {
				return err
			}
			if beneath {
				add(path)
			}
		}
		return nil
	}

	for path := range pinnedLeaves {
		if _, inTip := tipLeaves[path]; inTip {
			continue
		}
		if err := examine(path); err != nil {
			return nil, err
		}
	}

	// The type-change targets (codex-P1): a gitlink in the tip that the
	// candidate turns into a regular leaf, or into a directory.
	for path, tipClass := range tipLeaves {
		if tipClass != "gitlink" {
			continue
		}
		pinnedClass, stillLeaf := pinnedLeaves[path]
		switch {
		case stillLeaf && pinnedClass != "gitlink":
			// gitlink -> file: the added-path set is empty; the target is
			// the conversion that deletes the submodule's local bytes.
		case !stillLeaf && hasPrefixLeaf(pinnedLeaves, path+"/"):
			// gitlink -> directory: the target itself is not a leaf in the
			// pinned SHA, but the conversion unlinks the submodule tree.
		default:
			continue
		}
		// A gitlink target's local bytes hide inside its working directory,
		// which `git clean` will not enter — the path is tracked (the
		// gitlink IS the tracked entry), so the clean probe lists nothing.
		// The conversion examines the target the direct way instead: a
		// non-empty directory at the path carries local files (checked-out
		// submodule content, ignored files, untracked scratch) that the
		// unlink destroys. An empty directory — a submodule that was never
		// checked out — is safe to convert.
		info, statErr := os.Lstat(worktreePath(worktree, path))
		if statErr != nil {
			continue // nothing at the path; nothing to lose
		}
		if info.IsDir() {
			empty, err := dirIsEmpty(worktreePath(worktree, path))
			if err != nil {
				return nil, err
			}
			if !empty {
				add(path)
			}
			continue
		}
		if err := examine(path); err != nil {
			return nil, err
		}
	}
	return colliding, nil
}

// dirIsEmpty reports whether dir carries any entry at all (files,
// subdirectories, the submodule's own .git file — anything). Read errors
// read as not-empty: a directory we cannot enumerate may hold bytes, and
// the refusal direction is the safe one.
func dirIsEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, nil
	}
	return len(entries) == 0, nil
}

// hasPrefixLeaf reports whether any leaf lives under the prefix (a
// directory-to-leaf change's opposite leg: the candidate turns the gitlink
// into a directory whose new leaves all carry the prefix).
func hasPrefixLeaf(leaves map[string]string, prefix string) bool {
	for path := range leaves {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// worktreePath joins the worktree and a slash-separated leaf path with
// filepath semantics — git prints forward slashes on every platform, and
// the probe runs on the caller's filesystem.
func worktreePath(worktree, path string) string {
	return strings.Join([]string{strings.TrimRight(worktree, "/"), path}, "/")
}
