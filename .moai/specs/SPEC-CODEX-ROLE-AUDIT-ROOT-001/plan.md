# SPEC-CODEX-ROLE-AUDIT-ROOT-001 — Implementation Plan

## §A Context

- Card t1324, Class C, Tier M. Worktree `.moai/worktrees/t1324`, branch `WT-roleaudit-root`, plan baseline HEAD `7ca01df35`.
- Cause analysis is complete (lead-verified, two lanes' live measurements): the `resolved == callerTop` equality in `codexAuditValidateRoot` is unsatisfiable on a shared MCP server. This plan designs the repair only.
- Anchors (read at plan time, HEAD `7ca01df35`):
  - `internal/cli/codex_audit_launch.go:293-338` — `codexAuditValidateRoot`: equality clause at lines 308-309; same-repo checks (git-common-dir comparison lines 311-323, registered worktree scan lines 324-337); `codexAuditResolve` symlink canonicalization (line 461).
  - `internal/cli/codex_audit_launch.go:167-225` — `prepareCodexAudit`: root validation is the FIRST check (line 173), so a root refusal fires before any process launch or file write.
  - `internal/cli/codex_audit_launch.go:793-836` — CLI `moai codex audit <role>`: passes `ProjectRoot: top, CallerDir: cwd, Root: top` — the CLI path already self-consistent.
  - `internal/cli/codex_audit_mcp.go:86-134` — `handleCodexRoleAudit`: computes `top` from `codexRoleAuditServerDir()` and passes `ProjectRoot: top, CallerDir: serverDir` — the defect's wiring point.
  - `internal/cli/codex_task.go` — `codex_task` tool: schema properties `background, prompt, resume_last, thread_id, work_key, write` (measured); no `project_root`.

## §B Known Issues

- The MCP handler derives BOTH `ProjectRoot` and `CallerDir` from the server process cwd; under the new contract `ProjectRoot` must derive from the **presented root**, and `CallerDir` becomes unnecessary to the root decision.
- `codex_role_audit`'s tool description text ("The worktree root must be the worktree this server started in") states the unsatisfiable contract and must be rewritten with the schema change.
- `codex_task` no-arg behavior change is breaking for existing primary-rooted callers: after M2 they must pass `project_root` explicitly. Accepted per card narrowing target (1) — the primary pin is the defect.
- **RESOLVED (2026-09-29, manager-spec, plan-audit iter1 D1) — codex_task no-arg caller sweep, measured in place**: sweep command `grep -rln "codex_task" internal/template/templates .claude/agents .claude/skills .claude/commands .claude/rules` → verbatim outcome (7 files): `internal/template/templates/.claude/agents/moai/super-advisor.md` · `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` · `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` · `internal/template/templates/.codex/agents/moai/super-advisor.toml` · `.claude/agents/moai/super-advisor.md` · `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` · `.claude/rules/moai/development/agent-authoring.md`. Classification: all 7 are documentation or tool-grant surfaces (`tools:` grant line, prose tool descriptions, catalogue tables) — zero literal invocation sites carry an argument list. The only shipped caller able to invoke `codex_task` is super-advisor (C1 local + C2 template mirror + C3 emitted `.codex` toml), whose invocations are model-shaped at runtime and — the schema having no `project_root` today (measured RED, AC-003) — are necessarily argument-less with respect to it. No caller relies on a non-primary tree through that absence: the tool acts on the server's primary checkout regardless of caller, and super-advisor's documented purpose (second-opinion delegation) names no tree target. Decision: the argument-less reliance EXISTS, so M2's refusal flip is not unconditional — the super-advisor caller update rides M2 (§F) and lands together with the flip.

## §C Pre-flight

- [ ] Re-confirm the RED observations at run start (acceptance.md cells carry exact commands): AC-001 requires an MCP stdio session with the server started in the primary checkout; AC-002/AC-003 re-run against a binary built from the run-phase tree (`go build -o /tmp/<tree>-moai ./cmd/moai`, invoke by path — tool-provenance §2.2).
- [ ] `go test -timeout 30m ./internal/cli/ -run '^(CodexAudit|CodexTask)$' -count=1` green before starting (characterization baseline).
- [ ] AC-001 re-observation MUST use a live stdio session (the session stays open across the `tools/call`), never a pipe-EOF one-shot: a pipe whose stdin closes after the writes deterministically surfaces the lower-layer refusal `the server did not start inside a git worktree` (`codex_audit_mcp.go:100`, which discards the underlying error) instead of the target equality-clause refusal — identify that artifact message on sight and switch session form, not diagnosis.
- [ ] Existing gate tests live in `internal/cli/mcp_project_root_worktree_gate_test.go` and `codex_audit_*_test.go` — extend, do not parallel-duplicate.

## §D Constraints

- NO silent primary fallback anywhere on the three surfaces (REQ-002, REQ-005).
- The same-repository verification (git-common-dir + registered `worktree list --porcelain`) is KEPT, semantics unchanged (REQ-003).
- Symlink canonicalization stays (the `project_root`-family convention canonicalizes accepted paths).
- CLI convention compliance (`internal/cli/CLAUDE.md`): no AskUserQuestion in CLI code; exit-code discipline (1 user error); stderr for refusals; `filepath.Abs` for user paths.
- Cross-platform: worktree registration comparison must not assume path separator; verify with a `GOOS=windows go build ./...` pass before commit.
- Lane-local test scope only; the `internal/cli` full suite is slot-lease territory (`moai slot acquire` per gitflow-lane-protocol §8) when it must run whole.

## §E Self-Verification

- E1: AC matrix re-run — every release-blocking cell flipped and every regression-guard cell still green; commands + verbatim outputs recorded in progress.md.
- E2: `go vet ./internal/cli/...` and `golangci-lint run internal/cli/...` clean.
- E3: `go test -timeout 30m ./internal/cli/ -run '^(CodexAudit|CodexTask|ProjectRoot)$' -count=1` green; coverage of the changed validator paths ≥ 85%.
- E4: grep the three surfaces for regression: no `callerTop` equality residue; tool descriptions no longer promise the server-started-in contract.
- E5: `GOOS=windows GOARCH=amd64 go build ./...` clean.

## §F Milestones

Order: decision-reversibility first — M1 carries the validation-contract and handler-wiring decisions most likely to move in review; M2 carries a breaking surface decision; M3 is additive user-facing surface; M4 is mechanical.

- **M1 — Root-verification repair (flips AC-001)**: rework `codexAuditValidateRoot` to drop the `resolved == callerTop` equality and gate the presented root on the same-repository registration checks; rewire `handleCodexRoleAudit` to pass the presented root as `ProjectRoot` (drop the server-cwd-derived `top`/`CallerDir` coupling); rewrite the tool description. Update/add table-driven tests: sibling registered worktree accepted; primary accepted when presented explicitly; out-of-repo refused; unregistered refused; empty root refused.
- **M2 — codex_task project_root (flips AC-003)**: add the `project_root` argument to the `codex_task` schema, gate it through the same validator, and refuse argument-less requests with a structured error. Tests for both arms. Caller update rides this milestone (§B D1 resolution): update the super-advisor definition prose (C1 `.claude/agents/moai/super-advisor.md` + C2 template mirror `internal/template/templates/.claude/agents/moai/super-advisor.md`) to instruct passing `project_root` = the session's own toplevel, and regenerate the emitted C3 `super-advisor.toml` with `make agents-emit` — the refusal flip lands together with this caller update, never before it.
- **M3 — CLI role-audit verb (flips AC-002)**: add `moai codex role-audit <role>` (registration under the `codex` cobra namespace, avoiding the multi-word Use-prefix collision trap), reusing the CLI-path request shape (`ProjectRoot/CallerDir/Root` from the process cwd) with the repaired validator; same refusal semantics on stderr with exit 1.
- **M4 — Re-measure and residue sweep (mechanical)**: re-run the full AC matrix per §E1, grep sweeps per §E4, windows build per §E5, and record evidence into progress.md.

## §G Anti-Patterns

- Do NOT keep `CallerDir` as a hidden second gate after removing the equality — one presented-root contract, one verification path.
- Do NOT widen acceptance to "any registered worktree of any repository" — the same-repo boundary is the security boundary.
- Do NOT let the CLI verb bypass validation by constructing its own request shape that skips `prepareCodexAudit`.
- Do NOT update `.claude/rules/moai/core/moai-mcp-tools.md` in run phase — sync-phase owns it.

## §H Cross-References

- acceptance.md — the two-cell AC matrix and RED/GREEN evidence contract.
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input — the sanctioned caller-presents-its-own-toplevel convention.
- SPEC-CODEX-AUDIT-READONLY-001 — the launcher this SPEC repairs the root gate of.
- Card t1324 — lead-verified cause analysis and live measurements.
