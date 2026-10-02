package statusline

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SPEC-QUOTA-RECORD-WORKTREES-001 M3 (REQ-QWR-001..005, -008): the
// multi-directory quota reading.
//
// A session working in a linked worktree writes its telemetry record under that
// worktree's own state directory, so a reading that exists only there is
// invisible to a read of the primary checkout's directory alone. QuotaStateDirs
// names the directories to read — found through file reads of the git metadata
// only, never a git process — and AggregateQuotaDirs applies the single-directory
// rules of quota.go across all of them. Like quota.go this is a pure reader: no
// network, no process spawn, no write, no recursive walk.

// quotaGitdirReadBytes bounds how much of a worktree's `gitdir` file is read; the
// file holds one absolute path, so only its first line is ever used.
const quotaGitdirReadBytes = 4096

// QuotaStateDirs returns the state directories the quota reading covers for the
// repository rooted at root: the primary checkout's own first, then the state
// directory of every linked worktree the repository's git metadata registers,
// in sorted metadata-entry-name order. At most maxDirs metadata entries are
// examined, a usable one or not (an entry that cannot be used consumes the bound
// like any other). A root whose `.git` is absent, a regular file, or holds no
// readable `worktrees` directory yields the primary directory alone, as before
// worktrees were read at all.
//
// Each entry's `gitdir` file holds the absolute path of its worktree's `.git`
// file on its first line; an entry whose first line is empty, not absolute, or
// names a path that no longer exists contributes no directory. A locked entry is
// read like any other.
func QuotaStateDirs(root string, maxDirs int) []string {
	dirs := []string{filepath.Join(root, ".moai", "state")}
	gitDir := filepath.Join(root, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		return dirs
	}
	metaRoot := filepath.Join(gitDir, "worktrees")
	entries, err := os.ReadDir(metaRoot) // sorted by name
	if err != nil {
		return dirs
	}
	for i, e := range entries {
		if i >= maxDirs {
			break
		}
		if stateDir, ok := worktreeStateDir(filepath.Join(metaRoot, e.Name())); ok {
			dirs = append(dirs, stateDir)
		}
	}
	return dirs
}

// worktreeStateDir resolves one metadata entry (a `<common>/worktrees/<name>`
// directory) to the state directory of the worktree it registers.
func worktreeStateDir(metaDir string) (string, bool) {
	f, err := os.Open(filepath.Join(metaDir, "gitdir"))
	if err != nil {
		return "", false
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, quotaGitdirReadBytes)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", false
	}
	first, _, _ := bytes.Cut(buf[:n], []byte{'\n'})
	gitFile := strings.TrimSuffix(string(first), "\r")
	if !filepath.IsAbs(gitFile) {
		return "", false
	}
	if _, err := os.Stat(gitFile); err != nil {
		return "", false // the worktree was removed or moved: a stale entry
	}
	return filepath.Join(filepath.Dir(gitFile), ".moai", "state"), true
}

// AggregateQuotaDirs reads the session telemetry records under every state
// directory in stateDirs and reports each window as AggregateQuota does for one
// directory: the freshness, clock-skew, max-age, and reset rules apply to every
// record alike wherever it was found, and per window the record with the latest
// capture time across all directories wins. A tie on capture time keeps the
// directory listed first. An empty stateDirs, or no usable record in any
// directory, yields unknown.
func AggregateQuotaDirs(stateDirs []string, now time.Time, maxAge time.Duration) QuotaAggregate {
	var fiveHour, sevenDay winner
	for _, dir := range stateDirs {
		scanQuotaDir(dir, now, maxAge, &fiveHour, &sevenDay)
	}
	return QuotaAggregate{
		FiveHour: fiveHour.reading(now),
		SevenDay: sevenDay.reading(now),
	}
}
