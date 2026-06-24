---
id: SPEC-V3R6-BASH-RISK-GOVERNANCE-001
title: "Bash Risk-Amplifier Doctrine — subcommand soft-cap + warn-only hook signal"
version: "0.1.0"
status: completed
created: 2026-06-18
updated: 2026-06-20
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/development"
lifecycle: spec-anchored
tags: "bash,governance,risk,harness"
era: V3R6
---

# SPEC-V3R6-BASH-RISK-GOVERNANCE-001

## §A. Background / Problem

### A.1 Source Doctrine

Applied analysis of `github.com/wquguru/harness-books` (book1 ch04 + appendix A.3):

- **book1 ch04 thesis**: "工具是受管执行接口" (tools are managed-execution interfaces); "能力越通用，越要特殊看管" (the more general a capability, the denser its governance must be). Bash is named **风险放大器 (risk amplifier)**: a compound Bash command (`a && b | c; rm x`) multiplies blast radius. The book states Claude Code applies a subcommand-count cap on compound commands and partitions tools by concurrency safety (`isConcurrencySafe()`).
- **appendix A.3 litmus** (verbatim framing): *"if your Bash and ReadTool are governed the same way, your risk model is under-developed."*
- **§9.4**: the named principle this SPEC encodes.

### A.2 The Verified Gap

A grep for `risk.?amplifier|subcommand.?count|subcommand.?cap|compound.*command|isConcurrencySafe` across `.claude/rules/moai/` returns **zero hits** (verified 2026-06-18). moai-adk's PreToolUse hook (`handle-pre-tool.sh`) fires on `Write|Edit|Bash` **uniformly** (5s timeout) — there is:

- **No Bash-specific risk tier.** Bash (write/irreversible by default) is treated identically to Read/Glob/Grep (read-only).
- **No cap on compound-command fan-out.** A 7-segment pipeline (`a && b | c; d && e | f; rm x`) passes the same hook as `ls`.
- **Only one tool-category special-case exists**: the GLM-backend routing rule (`glm-web-tooling.md`) treats web-tool category specially. Bash governance is a real surface gap for a harness that runs `go test` / `gh pr` / `git push` / `moai hook` heavily.

This is the appendix A.3 litmus failing: Bash and ReadTool ARE governed the same way.

### A.3 Why Now (Sprint 15)

Sprint 15 is the harness-books application cohort. The harness executes heavy Bash fan-out (test suites, PR creation, hook forwarding). Introducing a warn-only, fail-open doctrine now — before the harness compounds the blast radius — is the cheapest insertion point.

## §B. Requirements (GEARS)

### REQ-BRG-001 — Bash risk tier classification (Ubiquitous)

The Bash tool SHALL be classified as risk-tier **"write/irreversible" by default**, DISTINCT from the Read/Glob/Grep risk-tier "read", per book1 ch04 (Bash as 风险放大器 / risk amplifier) and the appendix A.3 litmus.

### REQ-BRG-002 — Compound-command subcommand soft cap (State-driven)

**While** a Bash command is compound (contains one or more of: pipe `|`, logical-and `&&`, logical-or `||`, sequence `;`, backtick command substitution, `$(...)` POSIX substitution), the orchestrator/agent SHALL treat a subcommand count exceeding the **`BASH_SUBCOMMAND_SOFT_CAP`** threshold (default **5**) as a warn condition that MUST be surfaced either by splitting into a script file OR by explicit delegation.

The soft cap is a **named, tunable threshold constant** (`BASH_SUBCOMMAND_SOFT_CAP = 5`), not buried in prose. "Soft" means warn-only — it MUST NOT hard-fail or break existing workflows.

### REQ-BRG-003 — Destructive primitives require explicit confirmation (Event-driven)

**When** a Bash command contains a destructive primitive — `rm -rf`, `git push --force`, `git push --no-verify`, `git reset --hard`, SQL `DROP TABLE` / `TRUNCATE`, or `chmod -R 777` — the orchestrator/agent SHALL require explicit confirmation EVEN in `bypassPermissions` mode, cross-referencing the Implementation Kickoff Approval human-gate pattern in `CLAUDE.local.md` §19.1.

### REQ-BRG-004 — Warn-only, fail-open hook signal (Unwanted behavior — event-detected form)

**When** the PreToolUse hook detects a Bash command whose subcommand count exceeds `BASH_SUBCOMMAND_SOFT_CAP`, the hook SHALL emit a warn-only signal (stderr line AND structured JSON field) and the hook SHALL exit **0** (fail-open).

The hook SHALL NOT block, exit non-zero, or otherwise prevent the Bash call from proceeding. A blocking hook here would itself instantiate the death-spiral hazard book1 ch06 warns about. This fail-open constraint is non-negotiable.

### REQ-BRG-005 — No regression to existing CLAUDE.md safeguards (Unwanted behavior)

The Bash Risk-Amplifier Doctrine SHALL NOT contradict, weaken, or supersede `CLAUDE.md §7 Safe Development Protocol` or `CLAUDE.md §14 Parallel Execution Safeguards`. The doctrine is **strictly additive** — it layers a Bash-specific risk tier on top of the existing uniform PreToolUse hook behavior.

### REQ-BRG-006 — Grep reproducibility / vocabulary introduction (Ubiquitous)

The doctrine SHALL introduce the `risk-amplifier` and `subcommand` (and `BASH_SUBCOMMAND_SOFT_CAP`) vocabulary into active doctrine such that the grep reproducibility gap (`grep -rniE 'risk.?amplifier|subcommand.?count|subcommand.?cap|compound.*command|isConcurrencySafe' .claude/rules/moai/` returns non-zero) is closed.

## §C. Scope

### C.1 In Scope

- `.claude/rules/moai/development/coding-standards.md` — new §Bash Risk-Amplifier Doctrine section.
- `.claude/hooks/moai/handle-pre-tool.sh` — warn-only subcommand-count signal before the `exec` forward, fail-open.

### C.2 Out of Scope

- `glm-web-tooling.md` — separate concern (GLM-backend web-tool MCP routing).
- Hook timeout changes (5s preserved verbatim).
- Hook matcher scope changes (`Write|Edit|Bash` preserved).
- Any Go-binary (`moai hook pre-tool`) logic changes — the warn signal is emitted at the shell-wrapper layer before forwarding.
- Hard-fail enforcement of the soft cap (explicitly forbidden by REQ-BRG-004).

## §D. Constraints

- **WARN-only, fail-open** — non-negotiable. The hook MUST exit 0 even when the warning fires. (REQ-BRG-004)
- **Soft cap = warn threshold, not hard limit.** Existing workflows MUST NOT break. (REQ-BRG-002)
- **Preserve hook contract** — 5s timeout unchanged, `Write|Edit|Bash` matcher scope unchanged.
- **Additive only** — no contradiction with CLAUDE.md §7 / §14. (REQ-BRG-005)
- **Named constant** — the soft-cap value MUST be a named, tunable threshold, not prose. (REQ-BRG-002)
- **Book citations required** — book1 ch04 + appendix A.3 litmus MUST be cited. (AC-BRG-006)

## §E. Cross-References

- `CLAUDE.local.md` §19.1 — Implementation Kickoff Approval human-gate pattern (cross-ref for REQ-BRG-003 destructive-primitive confirmation).
- `CLAUDE.md` §7 Safe Development Protocol — compatibility target (REQ-BRG-005).
- `CLAUDE.md` §14 Parallel Execution Safeguards — compatibility target (REQ-BRG-005).
- `.claude/rules/moai/core/glm-web-tooling.md` — the only other tool-category special-case (explicitly out of scope per §C.2).
- `github.com/wquguru/harness-books` book1 ch04 + appendix A.3 + §9.4 — source doctrine.

## §F. Exclusions (What NOT to Build)

- **NOT building**: a hard-fail hook that blocks compound Bash commands. The warn-only constraint (REQ-BRG-004) forbids this.
- **NOT building**: a Go-binary-level (`internal/cli/hook/`) subcommand counter. The warn signal lives at the shell-wrapper layer.
- **NOT building**: changes to the existing 5s hook timeout or the `Write|Edit|Bash` matcher scope.
- **NOT building**: an `isConcurrencySafe()` tool-partition taxonomy (book1 ch04 names it; this SPEC encodes only the Bash-specific risk tier, not a full concurrency-safety partition).
- **NOT building**: changes to `glm-web-tooling.md` or any GLM-routing surface.
- **NOT building**: automated enforcement of destructive-primitive confirmation via the hook — REQ-BRG-003 is an orchestrator/agent obligation (doctrine-level), not a hook block. The hook is warn-only on subcommand count (REQ-BRG-004); destructive-primitive confirmation is the orchestrator's human-gate responsibility per §19.1.

## §G. History

- 2026-06-18 — plan-phase artifact set authored (spec.md + plan.md + acceptance.md + progress.md §E skeleton). Tier S, Sprint 15 P1b harness-books application cohort. Era V3R6 (explicit frontmatter `era: V3R6` to avoid H-2 transient misclassification per lifecycle-sync-gate.md).
