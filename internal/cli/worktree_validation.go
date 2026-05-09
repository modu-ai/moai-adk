// worktree_validation.go: isolation:worktree Agent() return value validation helper.
// REQ-RA-005 (validation helper), REQ-RA-010 (WORKTREE_PATH_INVALID sentinel).
//
// file exists as standalone helper for future SubagentStart hook re-delegation and external integration
// exists as a standalone helper. currently has no direct callsites
// (plan-stage expectation: 3-5 callsites → actual grep result: 0,
// only behavior validated in tests. actual wire-up to be performed in follow-up SPEC.
package cli

import (
"errors"
"fmt"
)

// worktreeReturn models worktree-related fields from Agent() call return value.
// isolation field: "worktree", "", "none", etc.
//
// same type is also defined in type test file (launcher_worktree_validation_test.go).
// after M4 completion, replace with type reference from test file.
//
// Note: since test file is in same package (cli), having same Name type in test file
// causes compile error. removed duplicate worktreeReturn definition from test file.
type worktreeReturn struct {
WorktreePath string
WorktreeBranch string
IsolationMode string
}

// WorktreePathInvalidError: typed error raised when isolation:worktree call returns broken worktreePath
// typed error that occurs when.
// Can be identified with errors.Is(err, ErrWorktreePathInvalid) or errors.As.
//
// AC-RA-10: error message includes agentName + reason context.
type WorktreePathInvalidError struct {
AgentName string
Reason string
}

// Error returns error message containing "WORKTREE_PATH_INVALID" sentinel string.
func (e *WorktreePathInvalidError) Error() string {
return fmt.Sprintf("WORKTREE_PATH_INVALID: agent=%q reason=%q", e.AgentName, e.Reason)
}

// Is returns true if target is ErrWorktreePathInvalid.
// supports errors.Is(err, ErrWorktreePathInvalid) pattern.
func (e *WorktreePathInvalidError) Is(target error) bool {
return errors.Is(target, ErrWorktreePathInvalid)
}

// ErrWorktreePathInvalid is a sentinel error used when isolation:worktree call returns broken worktreePath
// sentinel error used when.
// can be identified with errors.Is(err, ErrWorktreePathInvalid) (REQ-RA-005, REQ-RA-010).
var ErrWorktreePathInvalid = errors.New("WORKTREE_PATH_INVALID")

// @MX:NOTE: [AUTO] validateWorktreeReturn is a defense-in-depth helper that blocks Layer 2-4 of the SPEC-V3R3-RETIRED-AGENT-001 5-layer defect chain
// (worktree allocation broken state propagation + path interpolation literal "{}")
// helper for defense-in-depth. Even if Layer 1 (SubagentStart guard in internal/hook/subagent_start.go
// is blocked, this still operates as in-depth defense.
// current callsite: none (standalone helper); wire-up planned in follow-up SPEC.
//
// validateWorktreeReturn validate worktreePath in return value from isolation:"worktree" Agent() call
// validate if it is valid.
//
// Behavior:
// - if isolationMode is not "worktree", skip validation and return nil.
// - if result is nil or WorktreePath is empty string, return WORKTREE_PATH_INVALID error.
// - WorktreeBranch is optional — missing or empty string is allowed.
//
// REQ-RA-005: add validation helper.
// REQ-RA-010: prohibit propagation of broken worktreePath without validation.
//
// current callsite: none (standalone helper; wire-up planned in follow-up SPEC.
func validateWorktreeReturn(result *worktreeReturn, isolationMode string, agentName string) error {
// skip validation if isolation is not "worktree" (REQ-RA-005 edge case)
if isolationMode != "worktree" {
return nil
}

// check nil return value
if result == nil {
return &WorktreePathInvalidError{
AgentName: agentName,
Reason: "nil worktreeReturn",
}
}

// check empty worktreePath
if result.WorktreePath == "" {
return &WorktreePathInvalidError{
AgentName: agentName,
Reason: "empty WorktreePath",
}
}

// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-02 fix:
// literal "{}" or "[object Object]" which is the product of Layer 4 of the 5-layer defect chain
// Reject pattern. Satisfies AC-RA-18 critical assertion (validation layer raises
// WORKTREE_PATH_INVALID before any path interpolation).
switch result.WorktreePath {
case "{}", "[object Object]", "null", "undefined":
return &WorktreePathInvalidError{
AgentName: agentName,
Reason: "literal stringified-object pattern in WorktreePath: " + result.WorktreePath,
}
}

return nil
}
