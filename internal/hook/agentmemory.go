// agentmemory.go — reconciliation core for the agent-memory drain/mirror
// (SPEC-AGENT-MEMORY-DRAIN-001).
//
// A worktree is its own project root and `.claude/agent-memory/` is
// gitignored, so memory written inside a worktree never reaches the primary
// checkout through git. This file owns the shared reconciliation rules both
// consumers build on — the `moai memory drain` backfill (internal/cli) and
// the write-time mirror (post_tool.go):
//
//   - topic files are COPIED, never moved; a worktree is never mutated;
//   - an existing primary file is never overwritten — a collision lands as
//     `<topic>.wt-<worktree-name>.md` (the tree-qualified slot is that tree's
//     own, so a later write from the same tree may refresh it);
//   - the per-agent `MEMORY.md` index is never copied; each newly-landed
//     topic gains exactly one appended index line, derived from the
//     worktree's index line or the file's frontmatter;
//   - `_archive/` subdirectories copy under the same collision rule and
//     never gain index lines (archived topics carry none).
package hook

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// agentMemorySegment is the literal repo-local agent-memory path segment, in
// slash form. The mirror's trigger predicate is anchored to this segment
// (plan D7): a file at e.g. docs/agent-memory/x.md must NOT match.
const agentMemorySegment = ".claude/agent-memory/"

// agentMemoryIndexName is the per-agent memory index. It is never copied by
// the drain or mirror; only appended to in the primary store.
const agentMemoryIndexName = "MEMORY.md"

// agentMemoryPrimaryRootFn is the seam over PrimaryRootOf the write-time
// mirror resolves through — the drain CLI keeps the direct call (tests
// inject fixture primaries here instead of building git repositories).
var agentMemoryPrimaryRootFn = PrimaryRootOf

// IsAgentMemoryMDPath reports whether path is a markdown file under the
// literal `.claude/agent-memory/` path segment. This is the STRICT form —
// the memory audit's unanchored `agent-memory/` substring predicate is a
// looser scan and must not be reused here.
//
// @MX:ANCHOR: [AUTO] strict agent-memory path predicate shared by drain + mirror
// @MX:REASON: plan D7 — an unanchored predicate mirrors docs/agent-memory/x.md too; both consumers must anchor identically
func IsAgentMemoryMDPath(path string) bool {
	if !strings.HasSuffix(path, ".md") {
		return false
	}
	normalized := filepath.ToSlash(path)
	return strings.Contains(normalized, "/"+agentMemorySegment) ||
		strings.HasPrefix(normalized, agentMemorySegment)
}

// SplitAgentMemoryPath splits a path containing the `.claude/agent-memory/`
// segment into the tree root (the checkout the memory was written in), the
// agent name, and the agent-relative slash path (which may traverse
// subdirectories such as `_archive/`). ok is false when the segment is
// absent or no agent/rel follows it.
func SplitAgentMemoryPath(path string) (treeRoot, agent, rel string, ok bool) {
	if !IsAgentMemoryMDPath(path) {
		return "", "", "", false
	}
	normalized := filepath.ToSlash(filepath.Clean(path))
	idx := strings.Index(normalized, "/"+agentMemorySegment)
	leading := false
	if idx < 0 {
		idx = 0
		leading = true
	}
	var after string
	if leading {
		after = strings.TrimPrefix(normalized, agentMemorySegment)
	} else {
		treeRoot = normalized[:idx]
		after = normalized[idx+len(agentMemorySegment)+1:]
	}
	if treeRoot == "" && !leading {
		return "", "", "", false
	}
	parts := strings.SplitN(after, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", false
	}
	return treeRoot, parts[0], parts[1], true
}

// PrimaryRootOf resolves the primary checkout root of the repository
// containing dir. inWorktree reports whether dir lives in a linked worktree
// (rather than the primary itself or a plain subdirectory of it).
//
// `git rev-parse --git-dir --git-common-dir` distinguishes the two: the
// values are equal in the primary checkout and differ inside a linked
// worktree, and the parent of the absolute common dir is the primary root in
// both cases (plan §D).
func PrimaryRootOf(dir string) (primary string, inWorktree bool, err error) {
	gitDir, commonDir, err := revParseDirs(dir)
	if err != nil {
		return "", false, fmt.Errorf("agent memory: resolve primary root: %w", err)
	}
	if filepath.Clean(gitDir) != filepath.Clean(commonDir) {
		return filepath.Dir(commonDir), true, nil
	}
	return filepath.Dir(commonDir), false, nil
}

// revParseDirs returns the absolute --git-dir and --git-common-dir values.
// --path-format=absolute needs git >= 2.31; on older git the flag itself
// fails and the values are re-resolved without it, made absolute against
// dir (git prints them relative to the working directory when relative).
func revParseDirs(dir string) (string, string, error) {
	out, err := gitOutput(dir, "rev-parse", "--path-format=absolute",
		"--git-dir", "--git-common-dir")
	if err != nil {
		plain, plainErr := gitOutput(dir, "rev-parse", "--git-dir", "--git-common-dir")
		if plainErr != nil {
			return "", "", fmt.Errorf("agent memory: %w", plainErr)
		}
		out = plain
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || lines[0] == "" || lines[1] == "" {
		return "", "", fmt.Errorf("agent memory: git rev-parse: unexpected output %q", out)
	}
	return absPathAgainst(dir, lines[0]), absPathAgainst(dir, lines[1]), nil
}

// gitOutput runs `git -C dir <args>` and returns its stdout.
func gitOutput(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).Output()
	if err != nil {
		return "", fmt.Errorf("git %v: %w", args, err)
	}
	return string(out), nil
}

// absPathAgainst makes p absolute, resolving a relative value against base.
func absPathAgainst(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	abs, err := filepath.Abs(filepath.Join(base, p))
	if err != nil {
		return filepath.Clean(p)
	}
	return abs
}

// AgentSummary is the per-agent content summary of one drained tree.
type AgentSummary struct {
	Name     string `json:"name"`
	Files    int    `json:"files"`
	HasIndex bool   `json:"has_index"`
	Copied   int    `json:"copied"`
	Collided int    `json:"collided"`
	Skipped  int    `json:"skipped"`
}

// TreeDrainRecord is the drain result for one worktree (REQ-AM-010's
// machine-readable record; Actions carries the per-file human trace).
type TreeDrainRecord struct {
	Path            string         `json:"path"`
	Missing         bool           `json:"missing,omitempty"`
	Agents          []AgentSummary `json:"agents"`
	Files           int            `json:"files"`
	Copied          int            `json:"copied"`
	Collided        int            `json:"collided"`
	Skipped         int            `json:"skipped"`
	IndexLinesAdded int            `json:"index_lines_added"`
	Actions         []string       `json:"actions,omitempty"`
}

// syncOutcome classifies what happened to one topic file.
type syncOutcome int

const (
	syncNone syncOutcome = iota
	syncCopied
	syncCollided
	syncSkipped
)

// DrainTree scans treeRoot's agent-memory store and (apply=true) copies its
// topic files into primaryRoot under the reconciliation rules. The worktree
// is never mutated; the primary store is never read for anything but
// collision/idempotency checks. A missing tree or store is reported, not an
// error.
//
// @MX:ANCHOR: [AUTO] DrainTree is the shared drain/mirror reconciliation entry (backfill CLI + mirror)
// @MX:REASON: fan_in >= 3 — moai memory drain, the PostToolUse mirror, and the doctor regression guard all consume these rules; drift here desyncs all three
func DrainTree(primaryRoot, treeRoot string, apply bool) (TreeDrainRecord, error) {
	rec := TreeDrainRecord{Path: treeRoot}
	src := filepath.Join(treeRoot, ".claude", "agent-memory")
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		if err != nil && !os.IsNotExist(err) {
			return rec, fmt.Errorf("agent memory: stat %s: %w", src, err)
		}
		rec.Missing = true
		return rec, nil
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return rec, fmt.Errorf("agent memory: read %s: %w", src, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		agent := e.Name()
		summary := AgentSummary{Name: agent}
		agentDir := filepath.Join(src, agent)
		if _, statErr := os.Stat(filepath.Join(agentDir, agentMemoryIndexName)); statErr == nil {
			summary.HasIndex = true
		}

		var files []string
		walkErr := filepath.WalkDir(agentDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("agent memory: walk %s: %w", path, walkErr)
			}
			// Skip the per-agent index wherever it sits: only topic files
			// drain, and an index is reconciled by appending, never by
			// copying.
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") || d.Name() == agentMemoryIndexName {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if walkErr != nil {
			return rec, walkErr
		}

		for _, f := range files {
			rel, relErr := filepath.Rel(agentDir, f)
			if relErr != nil {
				return rec, fmt.Errorf("agent memory: rel %s: %w", f, relErr)
			}
			relSlash := filepath.ToSlash(rel)
			outcome, destName, syncErr := syncAgentMemoryFile(primaryRoot, treeRoot, f, agent, relSlash, apply)
			if syncErr != nil {
				return rec, syncErr
			}
			summary.Files++
			rec.Files++
			switch outcome {
			case syncCopied:
				summary.Copied++
				rec.Copied++
				rec.Actions = append(rec.Actions, fmt.Sprintf("copy %s/%s", agent, relSlash))
			case syncCollided:
				summary.Collided++
				rec.Collided++
				rec.Actions = append(rec.Actions, fmt.Sprintf("collide %s/%s → %s", agent, relSlash, destName))
			case syncSkipped:
				summary.Skipped++
				rec.Skipped++
				rec.Actions = append(rec.Actions, fmt.Sprintf("skip %s/%s (already present)", agent, relSlash))
			}

			// Index reconciliation: only newly-landed topics directly under
			// the agent directory gain an index line — archived subdirectory
			// content carries none by the archive rule.
			if outcome == syncCopied || outcome == syncCollided {
				if !strings.Contains(relSlash, "/") {
					line := indexLineFor(agentDir, f, destName)
					added, addErr := ensureIndexLine(primaryRoot, agent, line, destName, apply)
					if addErr != nil {
						return rec, addErr
					}
					if added {
						rec.IndexLinesAdded++
					}
				}
			}
		}
		rec.Agents = append(rec.Agents, summary)
	}
	sort.Slice(rec.Agents, func(i, j int) bool { return rec.Agents[i].Name < rec.Agents[j].Name })
	return rec, nil
}

// syncAgentMemoryFile copies srcFile (inside treeRoot's store) into
// primaryRoot's store at agent/rel under the collision rule. The returned
// destName is the file name the content occupies in the primary store.
//
// Never-overwrite rule (REQ-AM-003): a plain-name destination that exists
// with different content is left untouched and the content lands in the
// tree-qualified slot `<base>.wt-<tree>.md`. That slot is namespaced to this
// exact tree, so a re-sync from the same tree may refresh it — the repeated
// write-time mirror of an edited topic must not fan out into new names.
func syncAgentMemoryFile(primaryRoot, treeRoot, srcFile, agent, relSlash string, apply bool) (syncOutcome, string, error) {
	srcData, err := os.ReadFile(srcFile)
	if err != nil {
		return syncNone, "", fmt.Errorf("agent memory: read %s: %w", srcFile, err)
	}
	base := filepath.Base(relSlash)
	destDir := filepath.Join(primaryRoot, ".claude", "agent-memory", agent,
		filepath.FromSlash(filepath.Dir(relSlash)))
	destPath := filepath.Join(destDir, base)

	if destData, readErr := os.ReadFile(destPath); readErr == nil {
		if bytes.Equal(destData, srcData) {
			return syncSkipped, base, nil
		}
		suffixed := treeQualifiedCopyName(base, filepath.Base(treeRoot))
		suffixedPath := filepath.Join(destDir, suffixed)
		if existing, readErr := os.ReadFile(suffixedPath); readErr == nil && bytes.Equal(existing, srcData) {
			return syncSkipped, suffixed, nil
		}
		if !apply {
			return syncCollided, suffixed, nil
		}
		if err := writeAgentMemoryFile(suffixedPath, srcData); err != nil {
			return syncNone, "", err
		}
		return syncCollided, suffixed, nil
	}
	if !apply {
		return syncCopied, base, nil
	}
	if err := writeAgentMemoryFile(destPath, srcData); err != nil {
		return syncNone, "", err
	}
	return syncCopied, base, nil
}

// writeAgentMemoryFile persists data at path, creating parent directories.
func writeAgentMemoryFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("agent memory: mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("agent memory: write %s: %w", path, err)
	}
	return nil
}

// treeQualifiedCopyName builds the collision name: the `.wt-<worktree>`
// marker goes before the `.md` extension (feedback_x.md →
// feedback_x.wt-t223.md). Non-.md names get the marker appended.
func treeQualifiedCopyName(base, treeName string) string {
	ext := filepath.Ext(base)
	if ext == "" {
		return base + ".wt-" + treeName
	}
	return strings.TrimSuffix(base, ext) + ".wt-" + treeName + ext
}

// markdownLinkTarget captures a markdown link target ending in .md. Kept
// local so the drain core does not depend on the CLI package's unexported
// regex; the two describe the same index format.
var markdownLinkTarget = regexp.MustCompile(`\]\(([^)]+\.md)\)`)

// indexLinksTarget reports whether an index body links the named file.
func indexLinksTarget(indexBody, name string) bool {
	for _, m := range markdownLinkTarget.FindAllStringSubmatch(indexBody, -1) {
		if filepath.Base(filepath.ToSlash(m[1])) == name {
			return true
		}
	}
	return false
}

// indexLineFor derives the primary index line for destName, in priority
// order (plan §A.4 rule 2): the worktree index line for the source file
// (retargeted when the content landed under a tree-qualified name), then
// the file's frontmatter description, then its name.
func indexLineFor(srcAgentDir, srcFile, destName string) string {
	srcName := filepath.Base(srcFile)
	if data, err := os.ReadFile(filepath.Join(srcAgentDir, agentMemoryIndexName)); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			for _, m := range markdownLinkTarget.FindAllStringSubmatch(line, -1) {
				if filepath.Base(filepath.ToSlash(m[1])) == srcName {
					if destName == srcName {
						return strings.TrimRight(line, "\r")
					}
					return strings.Replace(strings.TrimRight(line, "\r"),
						"("+m[1]+")", "("+destName+")", 1)
				}
			}
		}
	}

	name := strings.TrimSuffix(destName, ".md")
	summary := name
	if fm, _, err := taxonomy.ParseFile(srcFile); err == nil {
		if fm.Name != "" {
			name = fm.Name
		}
		if fm.Description != "" {
			summary = fm.Description
		}
	}
	return fmt.Sprintf("- [%s](%s) — %s", name, destName, summary)
}

// ensureIndexLine appends exactly one index line for destName to the primary
// agent's MEMORY.md when (and only when) no line links it yet. A missing
// index is created with a header — the store is rebuildable from its index,
// so a drained topic must never land unindexed. apply=false computes
// whether a line would be added without writing.
func ensureIndexLine(primaryRoot, agent, line, destName string, apply bool) (bool, error) {
	agentDir := filepath.Join(primaryRoot, ".claude", "agent-memory", agent)
	indexPath := filepath.Join(agentDir, agentMemoryIndexName)

	content := "# Memory Index\n"
	if data, err := os.ReadFile(indexPath); err == nil {
		content = string(data)
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("agent memory: read %s: %w", indexPath, err)
	}
	if indexLinksTarget(content, destName) {
		return false, nil
	}
	if !apply {
		return true, nil
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += line + "\n"
	if err := os.WriteFile(indexPath, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("agent memory: write %s: %w", indexPath, err)
	}
	return true, nil
}

// MirrorAgentMemoryFile mirrors one just-written agent-memory file (the
// PostToolUse Write/Edit target) into the primary store. It is the
// write-time half of the drain: memory reaches primary at write time, before
// any disposal path can destroy the worktree.
//
// Returns mirrored=true when a copy (or collision refresh) was made. Every
// no-op case returns (false, nil): not agent-memory, the index itself, an
// unanchored path, or a session whose project root is the primary checkout
// itself (REQ-AM-006). An error means primary resolution or the copy failed
// — the hook caller logs it as a non-blocking notice and exits 0
// (REQ-AM-005).
func MirrorAgentMemoryFile(filePath string) (bool, error) {
	if !IsAgentMemoryMDPath(filePath) {
		return false, nil
	}
	if filepath.Base(filePath) == agentMemoryIndexName {
		// The index is never mirrored, only appended to in the primary
		// store, and only as a side effect of a topic landing.
		return false, nil
	}
	abs := filePath
	if !filepath.IsAbs(abs) {
		resolved, err := filepath.Abs(abs)
		if err != nil {
			return false, fmt.Errorf("agent memory: resolve %s: %w", abs, err)
		}
		abs = resolved
	}
	treeRoot, agent, rel, ok := SplitAgentMemoryPath(abs)
	if !ok {
		return false, nil
	}
	primary, inWorktree, err := agentMemoryPrimaryRootFn(treeRoot)
	if err != nil {
		return false, fmt.Errorf("agent memory mirror: %w", err)
	}
	if !inWorktree {
		return false, nil
	}
	if _, err := os.Stat(abs); err != nil {
		return false, fmt.Errorf("agent memory mirror: stat %s: %w", abs, err)
	}

	outcome, destName, err := syncAgentMemoryFile(primary, treeRoot, abs, agent, rel, true)
	if err != nil {
		return false, fmt.Errorf("agent memory mirror: %w", err)
	}
	if outcome == syncCopied || outcome == syncCollided {
		if !strings.Contains(rel, "/") {
			srcAgentDir := filepath.Join(treeRoot, ".claude", "agent-memory", agent)
			line := indexLineFor(srcAgentDir, abs, destName)
			if _, err := ensureIndexLine(primary, agent, line, destName, true); err != nil {
				// The copy landed; an index failure is a notice, not a
				// rollback trigger — doctor reports the orphan.
				return true, fmt.Errorf("agent memory mirror: %w", err)
			}
		}
		return true, nil
	}
	return false, nil
}
