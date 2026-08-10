// Package cli — `moai hook codex-review-gate` Stop-hook logic
// (SPEC-MOAI-MCP-SERVER-001 M2, REQ-MCP-008 / AC-MCP-009 / AC-MCP-010).
//
// codex_review_gate.go holds the PURE gate logic. The CLI subcommand wiring
// (stdin read → project-root resolve → config-gate read → this handler →
// stdout) lives in hook.go runCodexReviewGate. Keeping the logic pure + in the
// cli package lets it reuse the codex backend (mcp_codex.go, same package) and
// the readCodexReviewGateEnabled config reader without an import cycle
// (internal/cli → internal/hook already exists for the HookInput/HookOutput
// types). It returns the standard ALLOW/BLOCK HookOutput and NEVER invokes
// AskUserQuestion (subagent boundary — the orchestrator translates a BLOCK).
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"

	"github.com/spf13/cobra"
)

// reviewGateRuntimePrefixes are the working-tree paths the change detector
// ignores: they are hook/session-written state that drifts on every turn, so
// counting them as "reviewable" would make the self-gate fire on every Stop
// and defeat AC-MCP-009 (no false block on a non-editing turn).
var reviewGateRuntimePrefixes = []string{
	".moai/state/",
	".moai/cache/",
	".moai/reports/",
	".moai/logs/",
	".moai/harness/",
	".claude/agent-memory/",
}

// reviewGateChangeDetector is the injectable "is there reviewable uncommitted
// work?" seam. The production default runs `git status --porcelain` and filters
// runtime-managed paths; tests swap it to drive the self-gate deterministically
// (true = reviewable change present, false = clean / no-edit turn).
var reviewGateChangeDetector = hasReviewableChanges

// HandleCodexReviewGate is the Stop-hook gate logic. ALLOW/BLOCK contract:
//   - empty HookOutput ({})                            = ALLOW (Claude may stop)
//   - {Decision: "block", Reason: "..."}               = BLOCK (keep working)
//
// Decision order (AC-MCP-009 self-gate + AC-MCP-010 opt-in + REQ-MCP-012
// fail-open):
//  1. gate disabled (config off)            → ALLOW (opt-in default-off, C6)
//  2. stop_hook_active (loop prevention)    → ALLOW (mandatory CC protocol)
//  3. no reviewable uncommitted change      → ALLOW (self-gate; no false block)
//  4. codex missing                         → ALLOW (fail-open; can't trap the session)
//  5. codex review pass / inconclusive      → ALLOW
//  6. codex review FAIL                     → BLOCK (the gate's only block path)
//
// `enabled` is read by the caller (runCodexReviewGate via
// readCodexReviewGateEnabled) and passed in so this function stays free of
// config I/O (testable as pure logic). It is re-checked here as defense-in-depth.
func HandleCodexReviewGate(input *hook.HookInput, enabled bool, projectDir string) (*hook.HookOutput, error) {
	allow := &hook.HookOutput{}
	if !enabled {
		return allow, nil // (1) opt-in default-off
	}
	if input != nil && input.StopHookActive {
		return allow, nil // (2) loop prevention — never re-block an already-continuing turn
	}
	if !reviewGateChangeDetector(projectDir) {
		return allow, nil // (3) self-gate — nothing reviewable ⇒ no false block
	}

	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return allow, nil // (4) fail-open: a missing reviewer must not trap the session
	}

	// The 900s override (config.DefaultCodexReviewGateTimeout) is pinned in the
	// hook manifest (M4) for this hook only; here we enforce it on the codex
	// call itself so a hung review cannot stall the Stop beyond the manifest
	// budget. The moai-default 5s hook timeout does NOT apply (AC-MCP-010).
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultCodexReviewGateTimeout)
	defer cancel()
	out, rpcErr := runCodexReviewRPC(ctx, binaryPath, codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted,
		// cwd lets codex review the uncommitted changes in THIS project's tree;
		// without it thread/start reviews the app-server's own cwd, not projectDir.
		"cwd": projectDir,
	})
	if rpcErr != nil {
		// (5) fail-open: an inconclusive or erroring reviewer ⇒ ALLOW. The error
		// rides back with the ALLOW so runCodexReviewGate can log WHY on stderr;
		// it does not change the decision. Swallowing it here made a gate that
		// was turned on but structurally unable to reach a verdict look exactly
		// like a gate that had reviewed the change and found nothing wrong.
		return allow, rpcErr
	}
	if isBlockVerdict(out.Verdict) {
		return &hook.HookOutput{
			Decision: hook.DecisionBlock, // (6) the gate's only BLOCK path
			Reason:   "codex review gate: " + out.Summary,
		}, nil
	}
	return allow, nil // pass / inconclusive ⇒ ALLOW
}

// isBlockVerdict reports whether a codex verdict should BLOCK the session. A
// BLOCK fires ONLY on a clear failure; the inconclusive fail-open value and any
// non-fail verdict ALLOW. Matched case-insensitively on a "fail" prefix so
// "fail" / "FAIL" / "failed" all block while "inconclusive" / "pass" do not.
func isBlockVerdict(verdict string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(verdict)), "fail")
}

// hasReviewableChanges reports whether the working tree at projectDir carries
// any uncommitted change to a NON-runtime-managed path. It shells out to
// `git status --porcelain` (the SAME signal internal/hook/security turn-review
// uses). Fail-open: a non-git dir, missing git, or git error ⇒ false (nothing
// reviewable ⇒ ALLOW), so the gate never false-blocks outside a git repo.
func hasReviewableChanges(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	out, err := exec.Command("git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return reviewableFromPorcelain(string(out))
}

// reviewableFromPorcelain is the pure parser over a `git status --porcelain`
// payload: it returns true iff at least one changed path is NOT under a
// runtime-managed prefix. Split out so the filtering rule is unit-testable
// without a git repo. Porcelain line shape is fixed-width: columns 0-1 are the
// XY status (either may be a space), column 2 is a space, and the path begins
// at column 3 — so the path MUST be sliced from the raw line, NOT from a
// TrimSpace'd copy (TrimSpace would strip a leading-space status and shift the
// path off by one, dropping its leading "." and defeating the prefix filter).
// Renames use "XY <old> -> <new>"; the prefix check against <old> is sufficient.
func reviewableFromPorcelain(porcelain string) bool {
	for _, raw := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) <= 3 {
			continue // malformed — no path column
		}
		path := strings.TrimSpace(raw[3:]) // path column; trim trailing space / CR only
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[:idx] // rename source for the prefix check
		}
		if path == "" || isRuntimeManagedPath(path) {
			continue
		}
		return true
	}
	return false
}

// isRuntimeManagedPath reports whether path falls under a hook/session-written
// prefix the gate must ignore (otherwise every turn's .moai/state drift would
// trip the self-gate).
func isRuntimeManagedPath(path string) bool {
	for _, p := range reviewGateRuntimePrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// runCodexReviewGate is the cobra RunE for `moai hook codex-review-gate`
// (SPEC-MOAI-MCP-SERVER-001 M2 REQ-MCP-008). It reads the Stop-hook stdin JSON,
// resolves the project root, reads the workflow.codex.review_gate.enabled
// opt-in flag, and forwards to HandleCodexReviewGate. The output is emitted as
// JSON on stdout (the Stop-hook contract). Fail-open: any error logs to stderr
// and exits 0 so the Stop pipeline is never broken (REQ-MCP-012).
func runCodexReviewGate(cmd *cobra.Command, _ []string) error {
	input, err := readHookInput(cmd.InOrStdin())
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)
		// Fail-open: emit an empty ALLOW.
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	projectDir := resolveProjectDirFromInput(input)
	enabled := readCodexReviewGateEnabled(projectDir)
	out, gateErr := HandleCodexReviewGate(input, enabled, projectDir)
	if gateErr != nil {
		// Fail-open: a handler error MUST NOT trap the Stop pipeline.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-gate: error:", gateErr)
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	if out == nil {
		out = &hook.HookOutput{}
	}
	return emitHookOutput(cmd.OutOrStdout(), out)
}

// readHookInput reads and parses the hook stdin JSON into a HookInput. It is
// broken out so the fail-open path can distinguish "could not read stdin" from
// "handler errored".
func readHookInput(r io.Reader) (*hook.HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	var input hook.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parse stdin: %w", err)
	}
	return &input, nil
}

// emitHookOutput serializes out as JSON to w. A nil-safe helper used by the
// fail-open path and the success path alike.
func emitHookOutput(w io.Writer, out *hook.HookOutput) error {
	enc := json.NewEncoder(w)
	return enc.Encode(out)
}

// resolveProjectDirFromInput picks the best available project dir from the hook
// input: ProjectDir first, then CWD, then the CLAUDE_PROJECT_DIR environment
// variable, then empty (the gate self-gates on empty).
//
// The CWD arm is load-bearing: Claude Code sends `cwd` in the Stop payload and
// treats `project_dir` as a legacy field (internal/hook/types.go), so a
// ProjectDir-only resolution returned "" on every real invocation and both gate
// readers fail-CLOSED to false no matter how the project was configured. The
// env arm mirrors how the shell wrappers resolve the project root, so the two
// layers agree on which workflow.yaml they are reading.
func resolveProjectDirFromInput(input *hook.HookInput) string {
	if input != nil {
		if input.ProjectDir != "" {
			return input.ProjectDir
		}
		if input.CWD != "" {
			return input.CWD
		}
	}
	return os.Getenv("CLAUDE_PROJECT_DIR")
}
