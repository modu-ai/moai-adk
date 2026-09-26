package cli

// mcp_worktree_root.go — SPEC-MCP-WORKTREE-UNTRACKED-001.
//
// A repository that keeps .moai/ untracked gives every linked worktree a tree
// with no .moai of its own. validateProjectRoot's existing test (".moai is a
// directory") rejects such a tree, and omitting project_root silently acts on
// the primary checkout. This file holds the second, git-backed acceptance
// branch, the git-free "config-orphaned" predicate, and the audit-gate routing
// that keeps the primary's explicit workflow.audit.gates binding on such a tree.
//
// Every git inspection here runs scrubbed (no inherited GIT_DIR / GIT_WORK_TREE
// / GIT_COMMON_DIR / GIT_INDEX_FILE / GIT_CEILING_DIRECTORIES, LC_ALL=C), is
// decided by exit status and output shape — never message text — and is never
// retried through an alternative invocation: an inspection that cannot
// complete is a rejection (validator) or a fail-closed gate (gate read).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// scrubbedGitVars are removed from every git inspection's environment, so a
// variable inherited from the server process cannot redirect git at another
// repository (REQ-MWU-006).
var scrubbedGitVars = []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_CEILING_DIRECTORIES"}

// scrubbedGitEnv returns the process environment minus scrubbedGitVars and any
// locale override, with LC_ALL=C set.
func scrubbedGitEnv() []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		drop := name == "LC_ALL"
		for _, v := range scrubbedGitVars {
			if name == v {
				drop = true
			}
		}
		if !drop {
			env = append(env, kv)
		}
	}
	return append(env, "LC_ALL=C")
}

// runScrubbedGit runs `git -C dir args...` in the scrubbed environment and
// returns stdout. Any failure — git missing, non-zero exit — is an error; the
// caller never inspects git's message text.
func runScrubbedGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = scrubbedGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// singleGitPath parses output that must be exactly one non-empty absolute path
// line, and returns it canonicalized. Any other shape is an error.
func singleGitPath(out string) (string, error) {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 || strings.TrimSpace(lines[0]) == "" || !filepath.IsAbs(lines[0]) {
		return "", errors.New("unexpected git output shape")
	}
	return filepath.EvalSymlinks(lines[0])
}

// worktreeListEntry is one `git worktree list --porcelain` record.
type worktreeListEntry struct {
	path     string // canonical; "" when the listed path cannot be canonicalized
	prunable bool
}

// parseWorktreePorcelain parses `git worktree list --porcelain`. Every record
// must open with a `worktree <path>` line; any other shape is an error. A
// listed path that cannot be canonicalized (a deleted directory) is kept with
// an empty canonical path so it simply never matches (REQ-MWU-007).
func parseWorktreePorcelain(out string) ([]worktreeListEntry, error) {
	var entries []worktreeListEntry
	inRecord := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			inRecord = false
			continue
		}
		if !inRecord {
			path, ok := strings.CutPrefix(line, "worktree ")
			if !ok || !filepath.IsAbs(path) {
				return nil, errors.New("unexpected git worktree list shape")
			}
			canon, err := filepath.EvalSymlinks(path)
			if err != nil {
				canon = ""
			}
			entries = append(entries, worktreeListEntry{path: canon})
			inRecord = true
			continue
		}
		if line == "prunable" || strings.HasPrefix(line, "prunable ") {
			entries[len(entries)-1].prunable = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("git worktree list returned no entries")
	}
	return entries, nil
}

// errAmbiguousLayout marks a repository whose primary checkout cannot be
// identified unambiguously (--separate-git-dir, a submodule-internal git dir,
// a bare repository).
var errAmbiguousLayout = errors.New("ambiguous repository layout (separate git dir, submodule, or bare repository)")

// identifyPrimaryCheckout returns the primary checkout of the repository that
// contains dir, plus the porcelain worktree listing (REQ-MWU-004): the parent
// of the git common dir, accepted only when the common dir is named `.git`, the
// parent's own `.git` resolves to that same common dir, and the parent is the
// first listed worktree.
func identifyPrimaryCheckout(dir string) (string, []worktreeListEntry, error) {
	out, err := runScrubbedGit(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", nil, fmt.Errorf("git could not inspect the repository: %w", err)
	}
	common, err := singleGitPath(out)
	if err != nil {
		return "", nil, fmt.Errorf("git could not inspect the repository: %w", err)
	}
	if filepath.Base(common) != ".git" {
		return "", nil, errAmbiguousLayout
	}
	primary := filepath.Dir(common)
	primaryGit, err := filepath.EvalSymlinks(filepath.Join(primary, ".git"))
	if err != nil || primaryGit != common {
		return "", nil, errAmbiguousLayout
	}
	listOut, err := runScrubbedGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", nil, fmt.Errorf("git could not list worktrees: %w", err)
	}
	entries, err := parseWorktreePorcelain(listOut)
	if err != nil {
		return "", nil, err
	}
	if entries[0].path != primary {
		return "", nil, errAmbiguousLayout
	}
	return primary, entries, nil
}

// validateLinkedWorktreeRoot is the second acceptance branch of
// validateProjectRoot, reached only when canonical has no .moai directory
// (REQ-MWU-002..007). It accepts canonical when it is the top level of a linked
// worktree listed without a prunable mark, in a repository whose primary
// checkout has a .moai directory; every other shape is rejected with the
// failed condition named. raw is the caller's spelling, used in messages.
func validateLinkedWorktreeRoot(raw, canonical string) (string, error) {
	reject := func(reason string) (string, error) {
		return "", fmt.Errorf("project_root %q has no .moai directory and is not an accepted linked worktree: %s", raw, reason)
	}

	out, err := runScrubbedGit(canonical, "rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil {
		return reject("git could not inspect it as a worktree (not a git repository, an unregistered or unreadable worktree, or git unavailable)")
	}
	top, err := singleGitPath(out)
	if err != nil {
		return reject("git could not inspect it as a worktree (unexpected output)")
	}
	if top != canonical {
		return reject("it is not the top level of a worktree (a subdirectory of " + top + ")")
	}

	primary, entries, err := identifyPrimaryCheckout(canonical)
	if err != nil {
		return reject(err.Error())
	}
	if primary == canonical {
		return reject("it is the repository's primary checkout, which has no .moai directory")
	}
	if info, statErr := os.Stat(filepath.Join(primary, ".moai")); statErr != nil || !info.IsDir() {
		return reject("its primary checkout " + primary + " has no .moai directory")
	}
	for _, e := range entries[1:] {
		if e.path != canonical {
			continue // other entries — stale or not — are not a reason to reject
		}
		if e.prunable {
			return reject("it is listed as a prunable worktree")
		}
		return canonical, nil
	}
	return reject("it is not a registered worktree of " + primary)
}

// isConfigOrphanedRoot reports whether root has no workflow config of its own
// AND carries positive, git-free evidence of being a linked worktree top level
// (spec §4.4): `<root>/.git` is a regular file whose `gitdir:` path — resolved
// against root when relative, never against the process working directory —
// names an existing directory whose parent is named `worktrees` and which holds
// a `commondir` file. Two file reads, no subprocess. There is deliberately no
// check that the admin directory points back at root: such a check could only
// move a root toward today's fail-open gate.
func isConfigOrphanedRoot(root string) bool {
	if strings.TrimSpace(root) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")); err == nil {
		return false
	}
	dotGit := filepath.Join(root, ".git")
	info, err := os.Lstat(dotGit)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return false
	}
	gitdir := ""
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "gitdir:"); ok {
			gitdir = strings.TrimSpace(v)
			break
		}
	}
	if gitdir == "" {
		return false
	}
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(root, gitdir)
	}
	adminInfo, err := os.Stat(gitdir)
	if err != nil || !adminInfo.IsDir() {
		return false
	}
	if filepath.Base(filepath.Dir(filepath.Clean(gitdir))) != "worktrees" {
		return false
	}
	cdInfo, err := os.Stat(filepath.Join(gitdir, "commondir"))
	return err == nil && cdInfo.Mode().IsRegular()
}

// gateAssumedRequiredNote is the gate_unmet / residual-note wording for a
// config-orphaned root whose primary checkout could not be identified
// (REQ-MWU-012), distinguishable from a primary that declares `required`.
const gateAssumedRequiredNote = "workflow.audit.gates.codex assumed `required` because the primary checkout of this worktree could not be identified"

// resolveAuditGates returns the workflow.audit.gates block that governs an
// audit of root (REQ-MWU-011/012). A root that is not config-orphaned reads its
// own workflow.yaml exactly as before and runs no git. A config-orphaned root
// reads the gate of its primary checkout; when that primary cannot be
// identified the codex gate is treated as `required` and assumedNote says why.
func resolveAuditGates(root string) (gates config.AuditGates, assumedNote string) {
	if !isConfigOrphanedRoot(root) {
		return workflowAuditPins(root).Gates, ""
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return config.AuditGates{Codex: config.AuditGateRequired}, gateAssumedRequiredNote
	}
	return workflowAuditPins(primary).Gates, ""
}

// receiptCodexGateRequired is the receipt-id exposure read of recordAuditReceipt,
// routed the same way. auditreceipt.CodexGateRequired itself is unchanged (the
// hook-side receipt guard keeps reading the session tree's own config).
func receiptCodexGateRequired(root string) bool {
	if !isConfigOrphanedRoot(root) {
		return auditreceipt.CodexGateRequired(root)
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return true
	}
	return auditreceipt.CodexGateRequired(primary)
}

// withRootBlock returns data re-shaped as a JSON object carrying an added
// "_root" key, with every existing field kept verbatim (numbers are decoded as
// json.Number so their text does not change). Data that does not marshal to a
// JSON object is returned unchanged.
func withRootBlock(data any, rootBlock map[string]any) any {
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil || m == nil {
		return data
	}
	m["_root"] = rootBlock
	return m
}

// worktreeWarning is the _root.worktree_warning text of REQ-MWU-013.
const worktreeWarning = "project_root is a linked worktree whose repository does not track .moai; " +
	"this answer was read from the worktree tree and may be empty because .moai is not tracked there — " +
	"an empty result does not mean the project has no SPECs"
