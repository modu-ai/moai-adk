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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// The git inspection helpers, the config-orphaned predicate, and primary
// identification live in internal/auditreceipt so the hook package reaches the
// same answer (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-001). These names keep the
// call sites of this package unchanged.
var runScrubbedGit = auditreceipt.RunScrubbedGit

func singleGitPath(out string) (string, error) { return auditreceipt.SingleGitPath(out) }

func identifyPrimaryCheckout(dir string) (string, []auditreceipt.WorktreeEntry, error) {
	return auditreceipt.IdentifyPrimaryCheckout(dir)
}

func isConfigOrphanedRoot(root string) bool { return auditreceipt.IsConfigOrphanedRoot(root) }

// gateAssumedRequiredNote is the gate_unmet / residual-note wording for a
// config-orphaned root whose primary checkout could not be identified
// (REQ-MWU-012), distinguishable from a primary that declares `required`.
const gateAssumedRequiredNote = auditreceipt.GateAssumedRequiredNote

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
		if e.Path != canonical {
			continue // other entries — stale or not — are not a reason to reject
		}
		if e.Prunable {
			return reject("it is listed as a prunable worktree")
		}
		return canonical, nil
	}
	return reject("it is not a registered worktree of " + primary)
}

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
