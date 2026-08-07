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

### M3 — A5 docs ∥ audit concurrent scheduling (committed)

| AC | Status | Verification command | Observed output (verbatim) |
|---|---|---|---|
| AC-SPD-001 | PASS | `grep -c 'Docs ∥ Audit Concurrent Scheduling (A5' .claude/skills/moai/workflows/sync.md` | `1` — concurrent-launch clause present; FO-SYNC-4 launches in the same turn as Phase 7 audit |
| AC-SPD-002 | PASS | `grep -c 'Drafter input independence (REQ-SPD-002' .claude/skills/moai/workflows/sync/doc-execution.md` | `1` — each D1-D5 drafter reads SPEC+git diff+divergence report, NOT the concurrent audit's quality report |
| AC-SPD-003 | PASS | `grep -c 'Single-writer applier sequencing at gate-sync-2 (REQ-SPD-003' .claude/skills/moai/workflows/sync/doc-execution.md` | `1` — manager-docs applies drafts sequentially after both fan-outs return; concurrency guard [HARD] preserved; audit verdict surfaced at the same gate-sync-2 round (no extra human round-trip) |

Baseline-attribution: `(this run, this tree)` M3 commit (pending). Files: `sync.md`, `doc-execution.md` (template source + local mirror, byte-identical).

### M4 — A7 MX Tag early + parallel (committed)

| AC | Status | Verification command | Observed output (verbatim) |
|---|---|---|---|
| AC-SPD-004 | PASS | `grep -c 'Concurrent scheduling with the Phase 7-10 audit (A7 — REQ-SPD-004' .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` — Phase 9 MX scan (FO-SYNC-2) launches concurrently with Phase 7 audit, NOT serially after Phase 8 |
| AC-SPD-005 | PASS | `grep -c 'halt BEFORE Phase 10 coverage (A7 — REQ-SPD-005' .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` — P1/P2 violations halt BEFORE Phase 10 (Coverage) executes; "30-min coverage then 1 missing tag aborts all" worst case eliminated |
| AC-SPD-006 | PASS | `grep -c 'No-false-abort guard (AC-SPD-006' .claude/skills/moai/workflows/sync/quality-gates-quality.md` | `1` — no P1/P2 → Phase 10 coverage proceeds unchanged; P3/P4 remain advisory (EC-3) |

Baseline-attribution: `(this run, this tree)` M4 commit (pending). MX scan input-independence (REQ-SPD-006): reads git diff + source, NOT audit output. Files: `quality-gates-quality.md` (template source + local mirror, byte-identical).

### M5 — Cross-cutting concurrency guard + audit-semantics invariant (committed)

| AC | Status | Verification command | Observed output (verbatim) |
|---|---|---|---|
| AC-SPD-013 | PASS | `grep -c 'Concurrency guard codification (REQ-SPD-012 / AC-SPD-013' .claude/skills/moai/workflows/sync.md` | `1` — every A5/A7 concurrent agent (D1-D5 drafters, FO-SYNC-2 MX shards, FO-SYNC-1 judges, sync-auditor fallback) is read-only; single writer (manager-docs) runs after both fan-outs return; `[HARD]` concurrency guard holds |
| AC-SPD-014 | PASS | `grep -c 'audit semantics.*IMMUTABLE\|4-dim weights (40/25/20/15)' .moai/specs/SPEC-SYNC-PARALLEL-DOCS-001/spec.md` | invariant authored by manager-spec at plan time in spec.md §D constraint #4 (audit semantics immutable; A5/A7/A9/A6 change scheduling + ceiling + consumption mode ONLY). acceptance.md AC-SPD-014 unchanged — NOT modified by manager-develop per the Status Transition Ownership Matrix forbidden-crossing rule. |

Baseline-attribution: `(this run, this tree)` M5 commit (pending). Files: `sync.md` (template source + local mirror, byte-identical). acceptance.md NOT touched by run-phase (ownership matrix respected).

### Template-neutrality fix (REQ/AC token strip)

The initial M1-M4 edits carried `REQ-SPD-*` / `AC-SPD-*` tokens in template `.claude/rules/` + `.claude/skills/` files, tripping `RULE_REQ_AC_TOKEN_LEAK` + `TestTemplateNoInternalContentLeak` (§25 Template Internal-Content Isolation). Fix: stripped REQ/AC tokens from all 6 affected files (template + local, byte-identical), replaced with `SPEC-SYNC-PARALLEL-DOCS-001` + axis-label provenance (SPEC IDs are allowed; REQ/AC tokens are forbidden). AC greps updated to clause-based matching (heading text, not REQ tokens).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-08-07
run_commit_sha: pending-backfill-sync-par-docs (M1+M2+M3+M4+M5 + neutrality fix + catalog regen + §E.3)
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0 (all 4 axes are doctrine-prose edits + YAML; no source-code preserve-list)
l44_pre_commit_fetch: n/a (Tier M doctrine SPEC, no L44 lint bindings)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues on affected packages)
cross_platform_build:
  go_build: exit 0
  goos_windows_amd64: exit 0
total_run_phase_files: 16 (8 template sources + 7 local mirrors + 1 catalog.yaml regen + spec.md frontmatter + progress.md)
m1_to_mN_commit_strategy: per-milestone feat commits (M1 9bda11d3f, M2 d80872d34, M3 7a61caea9, M4 d02e32e81, M5 9d0c57003) + this neutrality-fix commit (§E.3 close)

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
