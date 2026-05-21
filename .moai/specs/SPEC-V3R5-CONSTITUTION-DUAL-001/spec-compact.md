---
id: SPEC-V3R5-CONSTITUTION-DUAL-001
title: "Constitution Dual-Zone Formalization with Validate CLI — Compact Extract"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
module: ".claude/rules/moai + internal/constitution + internal/cli"
lifecycle: spec-anchored
tags: "constitution, dual-zone, frozen, evolvable, zone-registry, mega-sprint, w1, compact"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial compact extract |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision — addressed plan-auditor BLOCKING defects (zone-registry 75→72, HARD rules 102→111 empirical, AC methodology unified) + SHOULD defects (5 new ACs for traceability, V3R5-001 namespace, plan.md/acceptance.md Out of Scope sections, 3 sentinel REQs added, AC-CDL-005 split). |

> Auto-extracted from spec.md. Contains REQ-CDL-* + AC-CDL-* + files-to-modify + exclusions only. For full context, narrative, background, and rationale, see `spec.md`.

---

## Empirical Baseline (binding for all REQ/AC computations)

At main HEAD `3bd2aa291`:

- **15 canonical source files** (CLAUDE.md + 14 files under `.claude/rules/moai/{core,workflow,development,design}/`)
- **111 [HARD] occurrences** (sum of `grep -hcE '\[HARD\]'` across the 15 files)
- **72 zone-registry entries** (CONST-V3R2-001..046 + 049 + 051..072 + 150..152; 3 internal gaps at 047/048/050)
- **Coverage: 72/111 = 65%** → Phase B target: 39 new entries (CONST-V3R5-001..039) → final N≥111 (100% coverage)

---

## REQ-CDL-* (EARS Requirements)

### Ubiquitous

- **REQ-CDL-001**: The system shall ensure that every `[HARD]` rule in the 15 canonical constitution source files (per spec.md §2.2) carries exactly one `[ZONE:Frozen]` or `[ZONE:Evolvable]` marker.

- **REQ-CDL-002**: The zone classification field `zone_class` in `zone-registry.md` entries shall be one of: `frozen-canonical` | `frozen-safety` | `evolvable-tuning` | `evolvable-experimental`.

- **REQ-CDL-003**: The `moai constitution validate` CLI verb shall be invokable with the `--format json` flag for CI integration.

- **REQ-CDL-004**: The `zone-registry.md` document shall achieve 100% [HARD] coverage (entries ≥ HARD rules count across the 15 source files in §2.2; baseline M=111).

- **REQ-CDL-005**: All zone-registry entries newly introduced by this SPEC shall use the ID format `CONST-V3R5-NNN` beginning at `CONST-V3R5-001` (parallel namespace to existing CONST-V3R2-* series).

### Event-Driven

- **REQ-CDL-006**: When `moai constitution validate` detects clause drift, the CLI shall exit with code 1 (or 2 if source file missing) AND emit JSON report listing affected entries with `status: DRIFT` or `status: SOURCE_FILE_MISSING`.

- **REQ-CDL-007**: When CI runs the validate step on a PR, the workflow MUST fail if drift is detected on main-target PRs.

- **REQ-CDL-008**: When a [HARD] rule is added without a corresponding zone-registry entry, `moai constitution validate --strict` MUST fail with error key `ZONE_UNREGISTERED`.

- **REQ-CDL-009**: When `zone-registry.md` is updated, `moai constitution list --zone evolvable` MUST reflect the change without restart.

### State-Driven

- **REQ-CDL-010**: While CI mode (env `CI=true`), `moai constitution validate` MUST default to `--strict` mode.

- **REQ-CDL-011**: While `MOAI_CONSTITUTION_SKIP_VALIDATE=1` env is set, validate MUST emit warning and exit 0.

### Optional

- **REQ-CDL-012**: If `--format json` is provided, output MUST be JSON Schema v1.0 compatible with `jq` parsing (fields: `status`, `drift_count`, `missing_count`, `unregistered_count`, `entries`).

- **REQ-CDL-013**: If a constitution `.md` file is deleted, validate MUST list affected entries with `status: SOURCE_FILE_MISSING` (exit 2).

### Unwanted

- **REQ-CDL-014**: The `validate` CLI MUST NOT modify any source files (read-only).

- **REQ-CDL-015**: Entries MUST NOT use zone `Frozen` while having `canary_gate: false` (FROZEN_WITHOUT_CANARY error).

- **REQ-CDL-016**: New zone-registry entries MUST NOT reference anchors that do not exist in target file (ANCHOR_NOT_FOUND error).

### Integrity (NEW in iteration 2)

- **REQ-CDL-017**: A zone-registry entry's `id` field MUST be globally unique. Duplicate IDs trigger `DUPLICATE_ID` error with exit code 1 regardless of `--strict`.

- **REQ-CDL-018**: A single `[HARD]` rule MUST NOT be annotated with more than one `[ZONE:*]` marker. Multi-marker triggers `DUPLICATE_ZONE_MARKER` warning (does not affect exit unless `--strict --fail-on-warning`).

- **REQ-CDL-019**: An entry older than 90 days SHOULD be reported as `STALE_ENTRY` warning (observation-only).

---

## AC-CDL-* (Acceptance Criteria — Given/When/Then summary)

- **AC-CDL-001 (D1 annotation completeness)**: Given 15 constitution files with X=111 [HARD] rules. When `grep -hcE '\[ZONE:(Frozen|Evolvable)\]'` summed returns Y. Then Y >= X.

- **AC-CDL-002 (D2 100% coverage)**: Given `moai constitution list --format json | jq '.entries | length'` returns N. When total HARD rules count across the 15 source files (§2.2) is M=111. Then N >= M (target N≥111).

- **AC-CDL-003 (D3 validate happy path)**: Given registry in sync with sources. When `moai constitution validate --strict --format json` runs. Then exit 0, status=ok, drift_count=0, completes < 5s.

- **AC-CDL-004 (D3 drift detection + multi-key fixtures)**: Given a [HARD] rule text modified to no longer match registered clause (or other error key trigger). When validate runs. Then exit 1 (or 2 for SOURCE_FILE_MISSING), entry has appropriate status (DRIFT/SOURCE_FILE_MISSING/ZONE_UNREGISTERED/FROZEN_WITHOUT_CANARY/ANCHOR_NOT_FOUND/DUPLICATE_ID/DUPLICATE_ZONE_MARKER/STALE_ENTRY).

- **AC-CDL-005a (CI integration — automatable)**: Given PR targeting main + ci.yml updated. When workflow runs. Then validate step fails PR on drift (CI auto-verifiable in single run).

- **AC-CDL-005b (Branch protection — manual maintainer verification)**: Given Phase D Task D-2 applied by admin post-merge. When `gh api .../branches/main/protection` invoked. Then required_status_checks.contexts has 5 entries including "Constitution Validate". **Manual verification** required (admin permission, not CI-automatable).

- **AC-CDL-006 (zone_class enum compliance, NEW)**: Given all entries declare `zone_class`. When values checked against allow-list. Then every value ∈ {frozen-canonical, frozen-safety, evolvable-tuning, evolvable-experimental}; INVALID_ZONE_CLASS rejection otherwise.

- **AC-CDL-007 (ID format compliance, NEW)**: Given new V3R5 entries. When `grep -oE '^- id: CONST-V3R5-[0-9]{3}$'` per entry. Then every new ID matches regex, starts at CONST-V3R5-001, contiguous within V3R5 namespace.

- **AC-CDL-008 (live reload, NEW)**: Given list invoked once. When registry.md edited, list re-invoked in same shell. Then second invocation reflects update without restart.

- **AC-CDL-009 (SKIP_VALIDATE override, NEW)**: Given controlled drift. When `MOAI_CONSTITUTION_SKIP_VALIDATE=1` set. Then exit 0 + stderr warning + no drift in stdout. CI workflow clears env to prevent bypass.

- **AC-CDL-010 (read-only assertion, NEW)**: Given read-only filesystem (`chmod -R -w .`). When validate runs. Then complete validation report without EACCES/EROFS errors, no write attempts.

---

## REQ ↔ AC Traceability (100% coverage)

| REQ ID | Primary AC | Verification Method |
|--------|------------|---------------------|
| REQ-CDL-001 | AC-CDL-001 | Grep-based zone marker count vs HARD count |
| REQ-CDL-002 | AC-CDL-006 | YAML schema validation against 4-enum allow-list |
| REQ-CDL-003 | AC-CDL-003 | CLI invocation with `--format json` + jq parse |
| REQ-CDL-004 | AC-CDL-002 | Registry entry count N ≥ HARD count M=111 |
| REQ-CDL-005 | AC-CDL-007 | grep regex `^CONST-V3R5-[0-9]{3}$` per new entry |
| REQ-CDL-006 | AC-CDL-004 | Controlled drift fixture + exit-code assertion |
| REQ-CDL-007 | AC-CDL-005a | CI workflow step exit code on drift PR |
| REQ-CDL-008 | AC-CDL-004 | Unregistered HARD rule fixture → error key |
| REQ-CDL-009 | AC-CDL-008 | Edit-then-invoke test without restart |
| REQ-CDL-010 | AC-CDL-005a | CI env asserts strict mode |
| REQ-CDL-011 | AC-CDL-009 | Env override fixture → exit 0 + stderr warn |
| REQ-CDL-012 | AC-CDL-003 | jq-parseable JSON with all 5 top-level fields |
| REQ-CDL-013 | AC-CDL-004 | File-deletion fixture → exit 2 + status |
| REQ-CDL-014 | AC-CDL-010 | Read-only filesystem assertion test |
| REQ-CDL-015 | AC-CDL-004 | Frozen + canary_gate=false fixture → reject |
| REQ-CDL-016 | AC-CDL-004 | Missing-anchor fixture → reject |
| REQ-CDL-017 | AC-CDL-004 | Duplicate id fixture → exit 1 always |
| REQ-CDL-018 | AC-CDL-001 | Multi-marker line emits warning |
| REQ-CDL-019 | AC-CDL-004 | Entry timestamp > 90 days fixture → warning |

Coverage: 19 REQs ↔ 10 ACs. Every REQ has ≥1 AC.

---

## Files to Modify

### [MODIFY] Documentation (.md files) — 15 canonical source files (per spec.md §2.2)

- `CLAUDE.md` — §1/§7/§8/§14/§19 ZONE markers inline
- `.claude/rules/moai/core/agent-common-protocol.md`
- `.claude/rules/moai/core/askuser-protocol.md`
- `.claude/rules/moai/core/moai-constitution.md`
- `.claude/rules/moai/design/constitution.md` — align with existing §2
- `.claude/rules/moai/workflow/ci-autofix-protocol.md`
- `.claude/rules/moai/workflow/ci-watch-protocol.md`
- `.claude/rules/moai/workflow/context-window-management.md`
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/spec-workflow.md`
- `.claude/rules/moai/workflow/worktree-integration.md`
- `.claude/rules/moai/workflow/worktree-state-guard.md`
- `.claude/rules/moai/development/agent-authoring.md`
- `.claude/rules/moai/development/branch-origin-protocol.md`
- `.claude/rules/moai/development/skill-authoring.md`

### [MODIFY] Registry

- `.claude/rules/moai/core/zone-registry.md` — CONST-V3R5-001..039 entries + zone_class field (retroactive on all 111 entries)

### [NEW] Go source

- `internal/constitution/registry.go` — YAML parser (~50 LOC)
- `internal/constitution/validator.go` — drift detection (~150 LOC)
- `internal/constitution/validator_test.go` — unit + fixture tests (~300 LOC)
- `internal/cli/constitution_validate.go` — Cobra subcommand (~100 LOC)

### [MODIFY] Go source

- `internal/cli/constitution.go` — register validate subcommand (~10 LOC) — sibling to existing guard/list/amend

### [MODIFY] CI

- `.github/workflows/ci.yml` — validate step (~20 LOC YAML)
- `Makefile` — `preflight` target add validate (~1 LOC)

---

## Scope Boundaries

### Out of Scope (Exclusions — What NOT to Build)

- **EXCL-001**: NO PreToolUse Frozen Guard hook implementation (W3 scope)
- **EXCL-002**: NO agent/skill frontmatter `zone:` field (T3 Full / follow-up SPEC)
- **EXCL-003**: NO expert-backend / expert-frontend / expert-mobile retirement (W2 scope)
- **EXCL-004**: NO retroactive classification of workflow rules beyond the 15 source files enumerated in spec.md §2.2
- **EXCL-005**: NO structural change to `.claude/rules/moai/design/constitution.md` (inline markers only)
- **EXCL-006**: NO precommit/generation-time check (CI-time only)

---

## Key Constraints

- Backward compatibility: existing `moai constitution list`/`guard`/`amend` behavior unchanged (C-CDL-001, C-CDL-006)
- 16-language neutrality: sentinel error keys (`DRIFT`, `SOURCE_FILE_MISSING`, `ZONE_UNREGISTERED`, `FROZEN_WITHOUT_CANARY`, `ANCHOR_NOT_FOUND`, `DUPLICATE_ID`, `DUPLICATE_ZONE_MARKER`, `STALE_ENTRY`, `INVALID_ZONE_CLASS`) language-agnostic
- Performance: validate < 5 seconds (current corpus ≤ 150 entries, ≤ 15 sources)
- No new external dependencies (Go stdlib + existing yaml.v3 only)
- TRUST 5: coverage ≥ 85%, linter zero warnings, read-only validator
- CLI verb orthogonality: `validate` (new) is **directly orthogonal** to existing `guard` — `guard` checks pre-amendment, `validate` checks post-acceptance drift. Both coexist (C-CDL-006).

---

## Sentinel Error Keys (all governed by REQs in iteration 2)

| Key | Meaning | Exit Code | Governing REQ |
|-----|---------|-----------|---------------|
| `DRIFT` | Registered clause no longer matches source | 1 | REQ-CDL-006 |
| `SOURCE_FILE_MISSING` | Referenced source file does not exist | 2 | REQ-CDL-006, REQ-CDL-013 |
| `ZONE_UNREGISTERED` | [HARD] rule in source has no registry entry | 1 (with --strict) | REQ-CDL-008 |
| `FROZEN_WITHOUT_CANARY` | Entry has zone=Frozen + canary_gate=false | 1 (with --strict) | REQ-CDL-015 |
| `ANCHOR_NOT_FOUND` | Entry's anchor does not exist in source | 1 (with --strict) | REQ-CDL-016 |
| `DUPLICATE_ID` | Two entries share same id | 1 (always) | REQ-CDL-017 (NEW) |
| `DUPLICATE_ZONE_MARKER` | Single [HARD] rule has multiple zone markers | warning only | REQ-CDL-018 (NEW) |
| `STALE_ENTRY` | Entry not updated in 90+ days | warning only | REQ-CDL-019 (NEW) |
| `INVALID_ZONE_CLASS` | zone_class not in 4-enum allow-list | 1 (with --strict) | REQ-CDL-002 + AC-CDL-006 |
