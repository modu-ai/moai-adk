package homestate

// location_memo.go — remembers where a repository is (Issue #1948).
//
// Every card operation resolves the repository location, and each resolution was
// several git processes (`rev-parse --git-dir --git-common-dir`, `--show-toplevel`,
// `--absolute-git-dir --is-inside-git-dir`). A factory test cell spent about 96 of
// its 104 git processes on answers about the same directory, and a long-lived
// process (the MCP server, the web console) repeats them for as long as it runs.
//
// The memo keeps an answer for as long as the filesystem it was read from looks
// the same. It does NOT rely on callers to say when a repository changed: tests,
// other processes and the user run git themselves, and no in-process hook sees
// that. Each entry carries a fingerprint of everything the answer depends on, read
// BEFORE the answer is computed, and the entry is used only while a fresh
// fingerprint equals it.
//
// What an answer depends on, and so what the fingerprint covers, for every
// directory from the one asked about up to the filesystem root (git discovers a
// repository by walking up):
//
//   - whether `.git` exists there, and as what: a directory (then whether it has a
//     `commondir`, and the size and modification time of its `config`, which can
//     carry core.worktree and core.bare), a file (its content, which names the git
//     directory), or a symlink (its target);
//   - whether the directory is itself a git directory (`HEAD`, `objects` and `refs`
//     all present), and then the size and modification time of its `config`.
//
// The modification time of a `.git` directory is deliberately NOT part of it: git
// touches that directory on almost every command, and including it would turn the
// memo into a cache that never hits in an active repository.
//
// The key also carries the environment that decides which repository git acts on
// (gitenv.RepoScopingVars). The callers scrub those variables before spawning git,
// so today they cannot change an answer; the key keeps the dependence explicit so
// that a caller that stops scrubbing cannot be served from another environment's
// entry.
//
// An answer that swallowed an error (git missing, a timeout, an unexpected exit) is
// a degraded answer, not a fact about the repository, and is never remembered.
// "Not a git repository" is a fact and is remembered.
//
// Known limit: git also reads the user's and the system's configuration
// (safe.directory, for one). A change there is not in the fingerprint.
//
// @MX:DEBT: user/system git config is not part of the location fingerprint
// @MX:CEILING: valid while no run changes safe.directory or similar mid-process
// @MX:UPGRADE: add the global/system config stamps if a long-lived process must
// see such a change without restarting

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// outsideRepository reports whether err says git found no repository: the
// question has a definite answer ("no") and the answer may be remembered.
//
// ResolveGitDirs does not return the *exec.ExitError (execerr.StatusDetail cuts
// that chain on purpose, so it is not mistaken for an intentional exit code at the
// cmd/moai seam), only its text: "... exited with status 128 (fatal: not a git
// repository ...)". notGitRepository, which reads the ExitError, cannot see it. The
// test is the text git prints, together with the status 128 it exits with, and
// nothing looser: any other failure stays a degraded answer.
func outsideRepository(err error) bool {
	if err == nil {
		return false
	}
	if notGitRepository(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "status 128") && strings.Contains(msg, "fatal: not a git repository")
}

// maxLocationEntries bounds the memo. A process that resolves more directories
// than this (a test binary that makes a repository per test) starts over rather
// than growing without limit.
const maxLocationEntries = 1024

type locationKey struct{ op, dir, env string }

type locationEntry struct{ value, fingerprint string }

var locationMemo = struct {
	sync.Mutex
	m map[locationKey]locationEntry
}{m: map[locationKey]locationEntry{}}

// resetLocationMemo forgets every remembered answer.
func resetLocationMemo() {
	locationMemo.Lock()
	locationMemo.m = map[locationKey]locationEntry{}
	locationMemo.Unlock()
}

// discoveryVars change how git WALKS UP to find a repository but are not in gitenv.RepoScopingVars
// (the callers do not scrub them, so a child sees them): audit F19.
var discoveryVars = []string{"GIT_DISCOVERY_ACROSS_FILESYSTEM"}

// locEnvKey is the value of every repository-locating variable, in a fixed order.
func locEnvKey() string {
	var b strings.Builder
	for _, name := range append(append([]string(nil), gitenv.RepoScopingVars...), discoveryVars...) {
		b.WriteString(name)
		b.WriteByte('=')
		b.WriteString(os.Getenv(name))
		b.WriteByte(0)
	}
	return b.String()
}

// memoLocation returns the remembered value of op for dir while the filesystem
// still looks as it did when the value was computed, and otherwise runs compute.
// compute reports whether its answer may be remembered.
func memoLocation(op, dir string, compute func() (value string, remember bool)) string {
	key := locationKey{op: op, dir: dir, env: locEnvKey()}
	// Read before computing: a change between the read and the answer leaves an
	// entry whose fingerprint no longer matches, so the next lookup recomputes.
	fp := locationFingerprint(dir)
	locationMemo.Lock()
	e, ok := locationMemo.m[key]
	locationMemo.Unlock()
	if ok && e.fingerprint == fp {
		return e.value
	}
	value, remember := compute()
	if remember {
		locationMemo.Lock()
		if len(locationMemo.m) >= maxLocationEntries {
			locationMemo.m = map[locationKey]locationEntry{}
		}
		locationMemo.m[key] = locationEntry{value: value, fingerprint: fp}
		locationMemo.Unlock()
	}
	return value
}

// locationFingerprint describes what a location answer for dir depends on; see the
// package comment above.
func locationFingerprint(dir string) string {
	var b strings.Builder
	p := filepath.Clean(dir)
	for {
		b.WriteString(p)
		b.WriteByte('|')
		stampDotGit(&b, filepath.Join(p, ".git"))
		bare := 0
		for _, n := range []string{"HEAD", "objects", "refs"} {
			if _, err := os.Lstat(filepath.Join(p, n)); err == nil {
				bare++
			}
		}
		if bare == 3 {
			b.WriteString("G")
			stampFile(&b, filepath.Join(p, "config"))
		}
		b.WriteByte('\n')
		parent := filepath.Dir(p)
		if parent == p {
			return b.String()
		}
		p = parent
	}
}

// stampDotGit records what the `.git` entry at path is.
func stampDotGit(b *strings.Builder, path string) {
	info, err := os.Lstat(path)
	if err != nil {
		b.WriteString("-")
		return
	}
	switch {
	case info.IsDir():
		b.WriteString("d")
		stampFile(b, filepath.Join(path, "config"))
		if _, err := os.Lstat(filepath.Join(path, "commondir")); err == nil {
			b.WriteString("c")
		}
	case info.Mode()&os.ModeSymlink != 0:
		b.WriteString("l")
		if target, err := os.Readlink(path); err == nil {
			b.WriteString(target)
		}
	default:
		// A `.git` file names the git directory ("gitdir: <path>"); its content is
		// the answer, so the content is the fingerprint. It is small by construction.
		b.WriteString("f")
		if f, err := os.Open(path); err == nil {
			content, _ := io.ReadAll(io.LimitReader(f, 4096))
			_ = f.Close()
			b.Write(content)
			stampGitdirTarget(b, path, content)
		}
	}
}

// stampGitdirTarget records whether the git directory a `.git` file names still exists. The file is
// unchanged when the primary repository drops a linked worktree's administrative directory (`git
// worktree prune`), yet git then no longer finds a repository there (audit F19).
func stampGitdirTarget(b *strings.Builder, dotGit string, content []byte) {
	target, ok := strings.CutPrefix(strings.TrimSpace(string(content)), "gitdir:")
	if !ok {
		return
	}
	target = strings.TrimSpace(target)
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(dotGit), target)
	}
	b.WriteString("{")
	for _, n := range []string{"HEAD", "commondir"} {
		if _, err := os.Lstat(filepath.Join(target, n)); err == nil {
			b.WriteString(n[:1])
		} else {
			b.WriteString("-")
		}
	}
	b.WriteString("}")
}

// stampFile records the size and modification time of path, or its absence.
func stampFile(b *strings.Builder, path string) {
	info, err := os.Lstat(path)
	if err != nil {
		b.WriteString("[-]")
		return
	}
	fmt.Fprintf(b, "[%d:%d]", info.Size(), info.ModTime().UnixNano())
}
