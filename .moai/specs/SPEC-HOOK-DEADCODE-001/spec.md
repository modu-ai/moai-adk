---
id: SPEC-HOOK-DEADCODE-001
title: "internal/hook package dead-code cleanup (3 corroborated scopes)"
version: "0.1.0"
status: draft
created: 2026-07-03
updated: 2026-07-04
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
tags: "cleanup, dead-code, hook, refactor, go, internal-hook"
---

# SPEC-HOOK-DEADCODE-001 — internal/hook package dead-code cleanup (3 corroborated scopes)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-03 | 0.1.0 | Initial plan-phase draft. Tier M. Orchestrator-scoped evidence investigation (`go list -deps` + registry-registration grep + `deadcode` tool) corroborated 3 deletion scopes; manager-spec independently re-verified all claims and found a **material correction to Scope 3** (see §A.4) — the `handle-agent-hook.sh` wrapper is actively registered via 4 agents' frontmatter `hooks:` blocks, not unregistered/dead as the initial evidence framing suggested. | manager-spec |
| 2026-07-04 | 0.1.0 | Plan-auditor near-miss remediation (FAIL 0.77, Tier M threshold 0.80). D1: rewrote AC-HDC-011's over-broad `TeammateIdle, TaskCompleted` → zero grep (unsatisfiable + over-deletion-hazard) to the anchored `handle-agent-hook.sh.*TeammateIdle` form + a preservation constraint naming the 2 legitimate event-list lines (~74, ~350) that MUST be kept. D2: corrected the `resolver.go` `HookResponse` baseline from the never-measured "6 references" to the measured `grep -c` value **9** (§A.3 + AC-HDC-005). D3: completed the `dual_parse.go` whole-file-deletion rationale by acknowledging the 2 package-level error vars `deadcode`'s function-only analysis cannot see. D4/D5: noted `HookInput.Data`'s post-M3 orphaned-field retention decision + the `HookOutput.Data` same-name non-conflation. Body-only edits; scope unchanged. | manager-spec |

## §A. Context and Intent

### §A.1 Why this cleanup

`internal/hook` implements the Claude Code hook event handlers (`SessionStart`, `PreToolUse`, `PostToolUse`, `SubagentStop`, etc. — see `internal/hook/CLAUDE.md`). The package has accumulated dead code across two independent development threads that were never fully wired into the live registry (`internal/hook/registry.go`):

1. An **agent-specific handler architecture** (`internal/hook/agents/` + `internal/hook/lifecycle/`) that was built but never registered — the registry dispatches generically by `EventType`, not by agent identity.
2. A **dual-protocol parser** (`internal/hook/dual_parse.go`) with 5 functions that have zero production callers — only test files call them.
3. A **pointless data-injection line** in `internal/cli/hook.go`'s `runAgentHook` that writes to a struct field (`HookInput.Data`) that has zero readers anywhere in the repository.

`go vet`/`golangci-lint` do not catch this class of dead code (unexported symbols with test-only callers, or entire unregistered packages) — only `go list -deps` (package-graph absence) and `deadcode` (call-graph-from-main absence), cross-corroborated against registry-registration grep, can distinguish genuine dead code from `deadcode`'s ~120 false positives on registry-dispatched handlers (see §A.3 and Out of Scope).

### §A.2 Evidence methodology (why 3 tools, not 1)

`deadcode ./cmd/moai` alone reports **123 dead functions** across `internal/hook` — but the majority are registry-registered event handlers (`PreToolUse`, `SessionStart`, etc.) that `deadcode`'s static from-`main` call-graph analysis cannot see as reachable, because they are invoked through `registry.Dispatch(ctx, event, input)` (dynamic dispatch keyed by `EventType`), not a direct call `deadcode` can trace. Treating all 123 as deletable would be an **unobserved defect claim** in the `verification-claim-integrity.md` sense — a text/tool-pattern inference without independent corroboration.

The 3 scopes in this SPEC are each corroborated by **two independent absence signals**, not `deadcode` alone:

| Scope | Signal 1 | Signal 2 |
|-------|----------|----------|
| M1 (agents/+lifecycle/) | `go list -deps ./cmd/moai` — packages absent from binary's dependency graph | `grep` across `registry.go` + all production `.go` — zero import, zero registration |
| M2 (dual_parse.go) | `deadcode ./cmd/moai` — 5 functions reported `unreachable func` | Repo-wide caller grep — only test files (`dual_parse_test.go` + `response_test.go`) call them |
| M3 (`HookInput.Data`) | Repo-wide grep for `.Data\b` reads — zero matches outside the single write site | Confirmed: even the (dead, M1-scoped) `internal/hook/agents/*` handlers never read it |

### §A.3 Evidence re-verification — what manager-spec independently confirmed

All claims below were re-run by this agent (not carried over from the handoff) against the current working tree (`HEAD` = `d26feac32`, working tree has unrelated uncommitted edits in `internal/hook/{pre_tool.go,registry.go,types.go,coverage_boost_test.go}` from a separate in-progress task — untouched by this investigation):

- `go list -deps ./cmd/moai | grep -E 'internal/hook/(agents|lifecycle)'` → **empty** (exit 1). Confirmed absent from dependency graph.
- `grep -rl 'internal/hook/agents|internal/hook/lifecycle' --include='*.go' .` (excluding `_test.go`) → **no production importer**.
- `grep -n 'hook/agents|hook/lifecycle' internal/hook/registry.go` → **no match**.
- File-level breakdown (corrects the handoff's "18 files, 1002 LOC" framing, which conflated file-count and production-LOC): **18 files total** = 13 production `.go` (531 LOC in `agents/` + 471 LOC in `lifecycle/` = **1002 LOC production**) + 5 test `.go` (**1869 LOC test**). **2871 LOC combined.**
- `deadcode ./cmd/moai` reports exactly the 5 named functions in `dual_parse.go` as `unreachable func`: `ParseHookOutput`, `synthesizeFromExitCode`, `ValidateHookResponse`, `ToHookOutput`, `ToHookResponse`.
- **Whole-file-deletion completeness — `dual_parse.go`'s 2 package-level error vars.** Beyond the 5 functions, `dual_parse.go` also declares 2 package-level error vars: `ErrHookProtocolLegacyRejected` (line 13) and `ErrHookInvalidPermissionDecision` (line 16). `deadcode` did NOT report these — it analyzes the function call graph only and structurally cannot see package-level `var` declarations, which is why they were absent from the original tool output. Independent repo-wide grep confirms both are **dead-with-the-functions**: `grep -rn 'ErrHookInvalidPermissionDecision' --include='*.go' .` (excluding `dual_parse.go` + `dual_parse_test.go`) → zero matches (it is referenced only inside the deleted `ValidateHookResponse`, `dual_parse.go:89`, plus the deleted `dual_parse_test.go` tests), and `grep -rn 'ErrHookProtocolLegacyRejected' --include='*.go' .` (same exclusions) → zero matches (fully orphaned — referenced by nothing at all, not even the 5 functions). Neither var is used anywhere outside the whole-file-deletion set, so the file carries no live symbol; this completes the M2 whole-file-deletion rationale.
- `HookResponse` type (the thing these 5 functions operate on) is defined in `internal/hook/response.go:11` and is a live, actively-used field type in `internal/permission/resolver.go` (`ctx.HookResponse`; measured pre-M2 baseline `grep -c 'HookResponse' internal/permission/resolver.go` = **9** matched lines — the earlier "6 references" figure was never actually measured and is corrected here to the observed `grep -c` value) — **confirmed MUST-PRESERVE**, distinct from the 5 dead functions.
- **New finding beyond the handoff**: `internal/hook/response_test.go` (a *different* file from `dual_parse_test.go`) contains **3 test functions** — `TestPermissionDecisionValues`, `TestHookResponseContinue`, `TestHookResponseAdditionalContextTruncation` — that call `ValidateHookResponse`/`ToHookOutput` directly (8 call-site references). These exercise real behavior (a 64 KiB `AdditionalContext` truncation, `PermissionDecision` enum validation, `Continue *bool` propagation) that exists **only** inside the dead functions — confirmed via a repo-wide grep that no other file implements this truncation/validation. Deleting `dual_parse.go` without also addressing these 3 test functions in `response_test.go` will break `go build`/`go test` with undefined-symbol errors. M2 scope is expanded accordingly (§ M2 below).
- `internal/hook/response.go:9` carries a doc-comment prose reference to `ParseHookOutput` (not a call) that becomes a dangling reference after M2 and must be updated.
- `HookInput.Data json.RawMessage \`json:"-"\`` (defined `internal/hook/types.go:316`, part of the `HookInput` struct) has exactly one write site repo-wide (`internal/cli/hook.go:340`, inside `runAgentHook`) and **zero read sites** repo-wide (including inside the M1-scoped, already-dead `internal/hook/agents/*` handlers that were presumably the field's intended consumer). After M3 removes that sole write site, the `HookInput.Data` field *declaration* itself becomes fully orphaned (zero readers AND zero writers). This SPEC intentionally **RETAINS the field declaration** — REQ-HDC-007 scopes M3 to removing the dead *writer*, not the field — and records removing the now-orphaned `HookInput.Data` declaration from the `HookInput` struct as an explicit micro-follow-up, deliberately kept out of M3 so the milestone diff stays limited to the injection removal + the 2 doc corrections. Note this `HookInput.Data` field is DISTINCT from the same-named but live `HookOutput.Data` field (`internal/hook/types.go:380`, written via the `Data:` struct literal inside `NewAllowOutputWithData`) — the two MUST NOT be conflated (the `NewAllowOutputWithData` writer is unaffected by this SPEC).

### §A.4 Material correction to the Scope 3 evidence handoff

The initial evidence framing described the `handle-agent-hook.sh` wrapper as "deployed... but registered 0 times in settings.json.tmpl", implying it is orphaned/unregistered scaffolding, and offered "delete the unregistered wrapper + retire agent-hooks.md" as a design option.

**Independent re-verification found this framing incomplete.** `grep -l '^hooks:' .claude/agents/moai/*.md` shows **4 live retained agents** (`manager-develop.md`, `manager-docs.md`, `manager-spec.md`, `sync-auditor.md`) declare a YAML frontmatter `hooks:` block — a Claude Code registration surface **distinct from** `settings.json.tmpl` — that invokes `handle-agent-hook.sh` with real action strings:

| Agent | PreToolUse | PostToolUse | SubagentStop |
|-------|-----------|--------------|--------------|
| manager-develop | `develop-pre-implementation` | `develop-post-implementation` | `develop-completion` |
| manager-docs | — | `docs-verification` | `docs-completion` |
| manager-spec | — | — | `spec-completion` |
| sync-auditor | — | — | `evaluator-completion` (undocumented in `agent-hooks.md`'s Actions table — a doc-staleness finding, see §A.5) |

`internal/cli/hook.go`'s `runAgentHook` infers `EventType` from the action-string suffix (`endsWithAny`) and dispatches via `registry.Dispatch(ctx, event, input)` — the **same registered generic event handlers** used by global hooks, not agent-specific ones (agent-specificity was the never-completed `internal/hook/agents/factory.go` architecture that M1 deletes). The wrapper script, the CLI subcommand, and the 4 agent-frontmatter registrations are **all live and load-bearing** — deleting any of them would break 4 production agents' hook wiring. **Design Option (a) as originally framed is invalidated.** See § M3 below for the corrected scope.

### §A.5 Secondary doc-staleness findings (in scope for M1/M3, not new deletion targets)

- `agent-hooks.md`'s "Handler Architecture" section (both `.claude/rules/moai/core/agent-hooks.md` and its template mirror) states dispatch goes through `internal/hook/agents/factory.go` — this becomes factually false once M1 deletes that package. Must be corrected, not merely deleted-with-the-package (the doc is a live, loaded rule file — see `agent-hooks.md` frontmatter `paths:` scope).
- `agent-hooks.md`'s "Agent Hook Actions" table omits the `sync-auditor` → `evaluator-completion` row (present in the live agent frontmatter, absent from the doc) — a pre-existing staleness unrelated to this cleanup's deletions, but adjacent enough to fix in the same M3 commit (single-file edit, near-zero incremental risk).
- `hooks-system.md:322` (both local + template mirror) states `handle-agent-hook.sh` handles "TeammateIdle, TaskCompleted events (team mode)" — **independently verified false**: `settings.json.tmpl` registers `handle-teammate-idle.sh` and `handle-task-completed.sh` for those two events, not `handle-agent-hook.sh`.

## §B. Requirements (GEARS format)

- **REQ-HDC-001** (Ubiquitous): The `internal/hook` module SHALL contain zero packages absent from the `moai` binary's dependency graph, as measured by `go list -deps ./cmd/moai`.
- **REQ-HDC-002** (Event-driven): When `go list -deps ./cmd/moai` is run after M1 lands, the system SHALL NOT list `internal/hook/agents` or `internal/hook/lifecycle` in its output.
- **REQ-HDC-003** (Ubiquitous): The system SHALL preserve the `HookResponse` type (`internal/hook/response.go:11`) and its live consumer `internal/permission/resolver.go` structurally unmodified through M2.
- **REQ-HDC-004** (Event-driven): When the 5 unreachable functions in `internal/hook/dual_parse.go` are removed, the system SHALL also remove or rewrite every test function that exclusively exercises those functions (in `dual_parse_test.go` AND `internal/hook/response_test.go`), such that `go build ./... && go test ./...` succeeds with zero undefined-symbol errors.
- **REQ-HDC-005** (Ubiquitous): The `moai hook agent <action>` CLI subcommand (`internal/cli/hook.go` `runAgentHook`) SHALL remain fully functional and behaviorally unmodified (cobra registration, `EventType` inference, `registry.Dispatch` call) across all 3 milestones.
- **REQ-HDC-006** (Unwanted behavior): When M3 is implemented, the system SHALL NOT delete, rename, or functionally alter `.claude/hooks/moai/handle-agent-hook.sh(.tmpl)` or any of the 4 agent-frontmatter `hooks:` blocks (`manager-develop.md`, `manager-docs.md`, `manager-spec.md`, `sync-auditor.md`) that invoke it.
- **REQ-HDC-007** (Ubiquitous): The `HookInput.Data` field (`internal/hook/types.go`) SHALL have zero writers repo-wide once M3 is implemented (the sole existing writer, `runAgentHook`, is removed because the field has zero readers repo-wide).
- **REQ-HDC-008** (Ubiquitous): Documentation describing the agent-hook dispatch mechanism (`agent-hooks.md`, `hooks-system.md`, both live copy and template mirror) SHALL accurately reflect the actual generic `EventType`-based dispatch behavior after M1/M3 — no reference to the deleted `internal/hook/agents/factory.go` per-agent architecture, and no incorrect claim that `handle-agent-hook.sh` serves `TeammateIdle`/`TaskCompleted`.
- **REQ-HDC-009** (State-driven): While a milestone (M1/M2/M3) is being implemented, the run-phase agent SHALL re-run `go list -deps ./cmd/moai`, `go build ./...`, and `go test ./...` after each scope's deletion, before proceeding to the next milestone.
- **REQ-HDC-010** (Unwanted behavior): The system SHALL NOT delete any of the ~120 other `deadcode`-reported findings in `internal/hook/*` that lack independent corroboration via BOTH `go list -deps` absence AND registry/frontmatter-registration absence (the false-positive class: registry-dispatched handlers invisible to `deadcode`'s static from-`main` analysis).

## §C. Non-Functional Constraints

- Zero behavior change for any live-registered hook path (`PreToolUse`, `PostToolUse`, `SubagentStop`, `SessionStart`, etc.) — this is a pure dead-code removal, not a refactor of live logic.
- Each milestone MUST be independently landable and independently `go build`/`go test` green — no milestone may leave the tree in a broken intermediate state.
- Template-First discipline (`CLAUDE.local.md §2`) applies to every doc edit — `agent-hooks.md` and `hooks-system.md` each have a live copy (`.claude/rules/moai/core/`) and a template mirror (`internal/template/templates/.claude/rules/moai/core/`); both MUST be edited byte-identically in the same commit (mirror-parity, verified by `internal/template/rule_template_mirror_test.go`).
- No SPEC ID, REQ token, or other internal-development artifact may be introduced into template-mirrored doc content beyond what already exists there (Template Internal-Content Isolation, `CLAUDE.local.md §25`).

## Exclusions

### Out of Scope — Blanket deadcode-reported deletions

- The ~120 other functions `deadcode ./cmd/moai` reports as unreachable across `internal/hook/*` are explicitly OUT OF SCOPE for this SPEC. Most are registry-registered event handlers (`SessionStart`, `PostCompact`, etc.) that `deadcode`'s static from-`main` analysis cannot see as reachable through `registry.Dispatch`'s dynamic `EventType` keying — deleting them would break live functionality. Per `verification-claim-integrity.md` §1.1 surface 3, a `deadcode`-only signal is a hypothesis, not a verified defect; this SPEC only acts on the 3 scopes independently corroborated by a second absence signal (§A.2, §A.3).

### Out of Scope — internal/hook/agents/factory.go re-wiring (feature completion)

- Completing the never-finished "agent-specific handler" architecture (wiring `internal/hook/agents/factory.go` into `registry.go` so `moai hook agent <action>` actually dispatches per-agent instead of generically by `EventType`) is explicitly OUT OF SCOPE. That is a feature-completion project (new production logic, new tests, a registry-semantics design decision), not a dead-code cleanup. This SPEC deletes the never-wired scaffolding (M1); a future SPEC may propose completing the wiring instead, if there is a concrete need for per-agent dispatch behavior beyond today's generic `EventType` routing.

### Out of Scope — moai hook agent subcommand / wrapper removal

- Deleting or altering `moai hook agent <action>` (`internal/cli/hook.go`), `.claude/hooks/moai/handle-agent-hook.sh(.tmpl)`, or any of the 4 agent-frontmatter `hooks:` registrations is explicitly OUT OF SCOPE. §A.4 establishes these are live, actively-dispatched infrastructure for 4 production agents — this corrects the initial evidence handoff's "unregistered wrapper" framing. Only the dead `HookInput.Data` injection (§M3) and the doc inaccuracies it exposed (§A.5) are in scope.

### Out of Scope — Claude Code runtime verification of agent-frontmatter `hooks:`

- This SPEC treats the 4 agents' `hooks:` frontmatter blocks as load-bearing based on consistent codebase usage across 4 agent files and the presence of matching action-string dispatch logic in `runAgentHook`. It does NOT independently verify against Claude Code's own runtime/source that agent-frontmatter `hooks:` blocks are honored as documented — that would require an external Claude Code platform investigation, out of scope for an `internal/hook` package cleanup.
