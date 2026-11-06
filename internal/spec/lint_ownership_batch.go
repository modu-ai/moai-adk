package spec

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// lint_ownership_batch.go — batched replacement for the per-SPEC
// `git log --follow -p -- <spec.md>` walk (card t2087).
//
// OwnershipTransitionRule and StatusTransitionValidityRule both ask, for every
// SPEC, "what is the most recent `status:` transition of this spec.md within its
// last gitLogWindowSize commits?". Answering that with one subprocess per SPEC
// made a corpus lint run O(n) in process spawns (~1138 serial `git` calls, >25 min
// on the real repository). Here ONE `git log -p` over every `<root>/*/spec.md`
// is parsed into per-file segments, and each SPEC's answer is resolved from that
// in-memory list, so the subprocess count is independent of the SPEC count.
//
// What is reproduced exactly: the 50-commit lookback per file, the newest-first
// scan for the first commit whose diff carries a `status:` delta (same line
// parser as the per-SPEC path), and `--follow` rename lineage (a rename segment
// moves the lookup to the old path). What is only approximated is documented at
// batchedOwnershipTransition.

// ownershipSeg is one commit's touch of one spec.md, newest-first in log order.
type ownershipSeg struct {
	oldPath string                     // pre-rename path; "" when the file was not renamed here
	rec     *ownershipTransitionRecord // non-nil iff this diff carries a `status:` delta
}

// ownershipBatch is the parsed result of one batched read for a specs root.
type ownershipBatch struct {
	err    error
	segs   []ownershipSeg
	byPath map[string][]int // post-commit path -> indices into segs, ascending (newest first)
}

// repoLocation is where the process sits inside its repository.
type repoLocation struct {
	top    string // absolute repository top-level (symlinks resolved)
	prefix string // cwd relative to top, slash-separated with trailing "/" (or "")
	err    error
}

// batchedOwnershipTransition answers the per-SPEC transition question from the
// run's batched read. ok is false when the batch cannot answer faithfully — no
// per-run cache, git unavailable, a path that cannot be mapped into the
// repository, or a failed read — and the caller then runs the per-SPEC walk.
//
// Residual divergence from the per-SPEC walk (it cannot be reproduced by one
// pathspec-limited read; measured on this repository in the t2087 report):
//   - history simplification runs over the whole `<root>/*/spec.md` pathspec, not
//     per file, so a merge that is TREESAME to one parent for one file but not for
//     the set may expose side-branch commits the per-file walk would prune;
//   - a rename whose source lies outside `<root>/*/spec.md` is seen as an addition.
//
// @MX:NOTE: [AUTO] batchedOwnershipTransition — the constant-subprocess seam of the ownership walk
// @MX:REASON: card t2087 — per-SPEC `git log --follow -p` made corpus lint exceed the CI time limit; verdicts must stay equal to lookupOwnershipTransitionPerSpec
func batchedOwnershipTransition(specPath string) (*ownershipTransitionRecord, bool) {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()
	if c == nil {
		return nil, false
	}
	// Takes the cache lock itself, so it must run before we hold it.
	if !cachedGitDirAvailable() {
		return nil, false
	}

	gitQueryCacheMu.Lock()
	defer gitQueryCacheMu.Unlock()

	rel, ok := c.repoRelative(specPath)
	if !ok || path.Base(rel) != "spec.md" || !batchSafePath(rel) {
		return nil, false
	}
	root := path.Dir(path.Dir(rel))
	if root != "." && !batchSafePath(root) {
		return nil, false
	}

	b, built := c.ownershipBatches[root]
	if !built {
		b = readOwnershipBatch(root)
		c.ownershipBatches[root] = b
	}
	if b.err != nil {
		return nil, false
	}
	return b.resolve(rel), true
}

// batchSafePath rejects paths git would quote, split ambiguously in a
// `diff --git a/<old> b/<new>` header, or read as pathspec magic. Such SPEC paths
// take the per-SPEC walk.
func batchSafePath(p string) bool {
	if p == "" || strings.HasPrefix(p, ":") || strings.Contains(p, " b/") {
		return false
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f || strings.ContainsRune("\"\\*?[]", r) {
			return false
		}
	}
	return true
}

// repoRelative maps a SPEC path (absolute, or relative to the process cwd) to
// the repository-relative, slash-separated form git prints. Caller holds the lock.
func (c *gitEnvCache) repoRelative(specPath string) (string, bool) {
	if !c.repoLocSet {
		c.repoLoc = locateRepo()
		c.repoLocSet = true
	}
	if c.repoLoc.err != nil {
		return "", false
	}

	var rel string
	if filepath.IsAbs(specPath) {
		dir, err := filepath.EvalSymlinks(filepath.Dir(specPath))
		if err != nil {
			return "", false
		}
		r, err := filepath.Rel(c.repoLoc.top, filepath.Join(dir, filepath.Base(specPath)))
		if err != nil {
			return "", false
		}
		rel = filepath.ToSlash(r)
	} else {
		rel = path.Clean(c.repoLoc.prefix + filepath.ToSlash(specPath))
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}

// locateRepo resolves the repository top-level and the cwd prefix in one spawn.
func locateRepo() repoLocation {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel", "--show-prefix").Output()
	if err != nil {
		return repoLocation{err: err}
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) < 1 || lines[0] == "" {
		return repoLocation{err: fmt.Errorf("git rev-parse printed no top-level")}
	}
	top, err := filepath.EvalSymlinks(lines[0])
	if err != nil {
		return repoLocation{err: err}
	}
	loc := repoLocation{top: top}
	if len(lines) > 1 {
		loc.prefix = lines[1]
	}
	return loc
}

// ownershipBatchFormat starts each commit with RS and ends the header with a
// third US, so a multi-line body cannot be mistaken for diff content.
const ownershipBatchFormat = "--format=%x1e%H%x1f%s%x1f%b%x1f"

// readOwnershipBatch runs the one batched read for specs under root and parses it.
func readOwnershipBatch(root string) *ownershipBatch {
	pathspec := ":(top,glob)" + root + "/*/spec.md"
	if root == "." {
		pathspec = ":(top,glob)*/spec.md"
	}
	// core.quotepath=false keeps non-ASCII names unquoted; --no-color,
	// --no-ext-diff and the explicit prefixes pin the output shape against
	// user/repo diff configuration. --no-relative keeps patch paths
	// repository-relative when diff.relative=true and lint runs from a
	// subdirectory; a git too old to know it fails the batch, which then falls
	// back to the per-SPEC walk.
	cmd := exec.Command("git", "-c", "core.quotepath=false", "log", "-M",
		"--no-color", "--no-ext-diff", "--no-relative", "--src-prefix=a/", "--dst-prefix=b/",
		ownershipBatchFormat, "-p", "--", pathspec)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &ownershipBatch{err: err}
	}
	if err := cmd.Start(); err != nil {
		return &ownershipBatch{err: err}
	}

	b, perr := parseOwnershipBatch(stdout)
	if perr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return &ownershipBatch{err: perr}
	}
	if werr := cmd.Wait(); werr != nil {
		return &ownershipBatch{err: werr}
	}
	return b
}

// parseOwnershipBatch streams `git log -p` output into per-file segments.
func parseOwnershipBatch(r io.Reader) (*ownershipBatch, error) {
	b := &ownershipBatch{byPath: make(map[string][]int)}

	sc := bufio.NewScanner(r)
	// Same 1 MiB line cap the per-SPEC parser uses; a longer line fails the
	// batch and the caller falls back to the per-SPEC walk.
	sc.Buffer(make([]byte, 64*1024), 1024*1024)

	type commitMeta struct{ hash, subject, body string }
	var (
		meta               commitMeta
		inHeader           bool
		hdr                strings.Builder
		hdrSeps            int
		segPath, segOld    string
		inSeg              bool
		oldStatus, newStat string
	)

	flush := func() {
		if !inSeg {
			return
		}
		seg := ownershipSeg{oldPath: segOld}
		if newStat != "" {
			seg.rec = &ownershipTransitionRecord{
				PreviousStatus:  oldStatus,
				CurrentStatus:   newStat,
				CommitSubject:   meta.subject,
				CommitSHA:       meta.hash,
				AuthoredByAgent: parseAuthoredByAgent(meta.body),
			}
		}
		b.byPath[segPath] = append(b.byPath[segPath], len(b.segs))
		b.segs = append(b.segs, seg)
		inSeg = false
	}

	finishHeader := func() error {
		parts := strings.Split(hdr.String(), "\x1f")
		if len(parts) < 4 {
			return fmt.Errorf("malformed commit header in git log output")
		}
		meta = commitMeta{hash: parts[0], subject: parts[1], body: parts[2]}
		inHeader = false
		return nil
	}

	for sc.Scan() {
		line := sc.Text()

		if inHeader {
			hdr.WriteString("\n" + line)
			hdrSeps += strings.Count(line, "\x1f")
			if hdrSeps >= 3 {
				if err := finishHeader(); err != nil {
					return nil, err
				}
			}
			continue
		}

		if strings.HasPrefix(line, "\x1e") {
			flush()
			inHeader = true
			hdr.Reset()
			hdr.WriteString(line[1:])
			hdrSeps = strings.Count(line, "\x1f")
			if hdrSeps >= 3 {
				if err := finishHeader(); err != nil {
					return nil, err
				}
			}
			continue
		}

		if fileHdr, ok := strings.CutPrefix(line, "diff --git "); ok {
			// Any file header closes the previous segment. A quoted path
			// (`diff --git "a/..."`) opens no segment of its own, so its lines
			// are not credited to the file before it.
			flush()
			oldStatus, newStat = "", ""
			segPath, segOld, inSeg = "", "", false
			if rest, ok := strings.CutPrefix(fileHdr, "a/"); ok {
				if i := strings.Index(rest, " b/"); i >= 0 {
					segPath, segOld = rest[i+3:], rest[:i]
					if segOld == segPath {
						segOld = ""
					}
					inSeg = true
				}
			}
			continue
		}

		if inSeg {
			// Same per-line rule as extractStatusDelta: the last `-status:` is the
			// old value, the last `+status:` the new one.
			if v, ok := parseStatusDiffLine(line, "-"); ok {
				oldStatus = v
			} else if v, ok := parseStatusDiffLine(line, "+"); ok {
				newStat = v
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if inHeader {
		return nil, fmt.Errorf("truncated commit header in git log output")
	}
	flush()
	return b, nil
}

// resolve returns the most recent `status:` transition for rel within the
// gitLogWindowSize newest commits of its rename lineage, or nil.
func (b *ownershipBatch) resolve(rel string) *ownershipTransitionRecord {
	cur, pos := rel, -1
	for count := 0; count < gitLogWindowSize; count++ {
		idxs := b.byPath[cur]
		k := sort.SearchInts(idxs, pos+1)
		if k == len(idxs) {
			return nil
		}
		pos = idxs[k]
		seg := b.segs[pos]
		if seg.rec != nil {
			rec := *seg.rec
			return &rec
		}
		if seg.oldPath != "" {
			cur = seg.oldPath
		}
	}
	return nil
}
