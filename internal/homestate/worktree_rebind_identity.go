package homestate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorktreeRebindIdentity is captured before a move. Its private fields require
// callers to capture complete Git and filesystem identity through the helper.
type WorktreeRebindIdentity struct {
	commonDir, gitDir, head, branch string
	commonInfo, gitInfo             os.FileInfo
}

// CaptureWorktreeRebindIdentity records the registered source's repository,
// administrative directory, HEAD and branch before a forward or reverse move.
func CaptureWorktreeRebindIdentity(path string) (WorktreeRebindIdentity, error) {
	id, err := inspectWorktreeRebindIdentity(path)
	if err != nil {
		return id, err
	}
	ok, err := defaultCardWorktreeTargetProbe(path, id)
	if err != nil {
		return id, err
	}
	if !ok {
		return id, fmt.Errorf("%s is not a registered worktree", path)
	}
	return id, nil
}

func inspectWorktreeRebindIdentity(path string) (WorktreeRebindIdentity, error) {
	var id WorktreeRebindIdentity
	probes := []struct {
		target *string
		args   []string
	}{
		{&id.commonDir, []string{"rev-parse", "--path-format=absolute", "--git-common-dir"}},
		{&id.gitDir, []string{"rev-parse", "--absolute-git-dir"}},
		{&id.head, []string{"rev-parse", "--verify", "HEAD"}},
		{&id.branch, []string{"rev-parse", "--symbolic-full-name", "HEAD"}},
	}
	for _, probe := range probes {
		out, err := scrubbedGit(path, probe.args...).CombinedOutput()
		if err != nil {
			return id, fmt.Errorf("inspect rebind source %s (%v): %w: %s", path, probe.args, err, strings.TrimSpace(string(out)))
		}
		*probe.target = strings.TrimSpace(string(out))
		if *probe.target == "" {
			return id, fmt.Errorf("empty rebind identity for %s (%v)", path, probe.args)
		}
	}
	var err error
	id.commonInfo, err = rebindDirectoryInfo(id.commonDir)
	if err != nil {
		return id, err
	}
	id.gitInfo, err = rebindDirectoryInfo(id.gitDir)
	if err != nil {
		return id, err
	}
	return id, nil
}

// File.Stat captures Windows file IDs now, instead of deferring identity
// lookup until a later path could have been replaced.
func rebindDirectoryInfo(path string) (os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, statErr := f.Stat()
	closeErr := f.Close()
	if statErr != nil {
		return nil, statErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("rebind metadata path is not a directory: %s", path)
	}
	return info, nil
}

func sameWorktreeFile(a, b string) bool {
	left, err := os.Stat(a)
	if err != nil {
		return false
	}
	right, err := os.Stat(b)
	return err == nil && os.SameFile(left, right)
}

// defaultCardWorktreeTargetProbe verifies both the source witness and Git's
// registration at call time. A copied .git pointer is not a registration.
func defaultCardWorktreeTargetProbe(path string, expected WorktreeRebindIdentity) (bool, error) {
	if expected.commonDir == "" || expected.gitDir == "" || expected.head == "" || expected.branch == "" {
		return false, fmt.Errorf("rebind requires a pre-move source identity")
	}
	current, err := inspectWorktreeRebindIdentity(path)
	if err != nil {
		return false, err
	}
	if expected.commonInfo == nil || expected.gitInfo == nil || !os.SameFile(current.commonInfo, expected.commonInfo) || !os.SameFile(current.gitInfo, expected.gitInfo) || current.head != expected.head || current.branch != expected.branch {
		return false, nil
	}
	out, err := scrubbedGit(path, "worktree", "list", "--porcelain", "-z").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("list rebind target registration: %w: %s", err, strings.TrimSpace(string(out)))
	}
	var registered, head, branch string
	matches := func() bool {
		return registered != "" && sameWorktreeFile(filepath.FromSlash(registered), path) && head == expected.head && branch == expected.branch
	}
	for _, field := range strings.Split(string(out), "\x00") {
		if field == "" {
			if matches() {
				return true, nil
			}
			registered, head, branch = "", "", ""
			continue
		}
		key, value, _ := strings.Cut(field, " ")
		switch key {
		case "worktree":
			registered = value
		case "HEAD":
			head = value
		case "branch":
			branch = value
		case "detached":
			branch = "HEAD"
		}
	}
	return matches(), nil
}
