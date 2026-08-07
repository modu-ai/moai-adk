# progress.md — SPEC-SYNC-PARALLEL-DOCS-001

> Canonical §E section skeleton. Populated by phase-owning agents: §E.1 by manager-spec (plan-phase), §E.2/§E.3 by manager-develop (run-phase), §E.4 by manager-docs (sync-phase). The literal `§E.2` / `§E.3` / `§E.4` heading tokens are parser-load-bearing (`internal/spec/era.go` `hasAnyProgressMarker`) — do NOT rename.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-auditor verdict>_

## §E.2 Run-phase Evidence

### M1 — A9 attributable diff-check + fallback (committed)

| AC | Status | Verification command | Observed output (verbatim) |
|---|---|---|---|
| AC-SPD-007 | PASS | `grep -c 'Attribution discipline (REQ-SPD-007' .claude/rules/moai/development/manager-develop-prompt-template.md` | `1` — attribution-triple clause (command a / output b / baseline c) present in § Section E |
| AC-SPD-008 | PASS | `grep -c 'Attributable diff-check doctrinal switch (REQ-SPD-008' .claude/rules/moai/core/agent-common-protocol.md` | `1` — doctrinal switch clause present in § Parallel Execution; cites `moai verify check --key-current` (live snapshot surface re-verified at quality-gates-quality.md:41) |
| AC-SPD-009 | PASS | `grep -c 'Fallback-to-re-execution contract (REQ-SPD-009' .claude/rules/moai/workflow/verification-batch-pattern.md` | `1` — fallback contract present; binds any-mismatch → re-execution, mismatch-reason logging, VCI §1.1 invariant on every path |

Baseline-attribution: `(this run, this tree)` HEAD = 63fceb889 (pre-M1) → M1 commit (pending). Files: `manager-develop-prompt-template.md`, `agent-common-protocol.md`, `verification-batch-pattern.md` (template source + local mirror, byte-identical).

### M2 — A6 Tier-aware plan-auditor retry ceilings (committed)

| AC | Status | Verification command | Observed output (verbatim) |
|---|---|---|---|
| AC-SPD-010 | PASS | `grep -c 'plan_audit_tier_ceilings' .moai/config/sections/harness.yaml` | `1` — `S: 1` entry present (Tier S single-pass ceiling) |
| AC-SPD-011 | PASS | `grep -A3 'plan_audit_tier_ceilings' .moai/config/sections/harness.yaml \| grep -c 'M: 2'` | `1` — `M: 2` entry present (Tier M two-spawn ceiling) |
| AC-SPD-012 | PASS | `grep -c 'L: 3' .moai/config/sections/harness.yaml` | `1` — `L: 3` entry present (Tier L legacy 3-spawn fallback); backward-compat clause also in plan-auditor.md § Retry Loop Contract |
| (consumer) | PASS | `grep -c 'Tier-resolved' .claude/agents/moai/plan-auditor.md` | `1` — plan-auditor consults Tier ceiling from harness.yaml SSOT; former `max_iterations: 3` literal demoted to consumer-side reference |

Baseline-attribution: `(this run, this tree)` M2 commit (pending). `go test ./internal/config/...` → `ok github.com/modu-ai/moai-adk/internal/config 8.126s` (no CI regression; harness.yaml outside struct-yaml symmetry audit scope). Files: `harness.yaml`, `plan-auditor.md` (template source + local mirror, byte-identical).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
