---
id: SPEC-V3R5-CONSTITUTION-DUAL-001
title: "Constitution Dual-Zone Formalization with Validate CLI"
version: 0.1.0
status: draft
created: 2026-05-20
updated: 2026-05-20
author: MoAI claude-bot (issue #1014)
priority: P1 High
phase: "v3.5.0 — Mega-Sprint W1"
module: "internal/constitution/, .claude/rules/moai/, CLAUDE.md"
dependencies:
    - SPEC-V3R2-CON-001
    - SPEC-V3R2-CON-002
related_gap:
    - F-006
related_problem: []
related_pattern:
    - S-4
related_principle:
    - P1
    - P2
    - P12
related_theme: "Layer 1: Constitution Hardening"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r5, constitution, dual-zone, validate-cli, drift-detection, ci-integration, sentinel-errors"
---

# SPEC-V3R5-CONSTITUTION-DUAL-001: Constitution Dual-Zone Formalization with Validate CLI

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-20 | MoAI claude-bot | Initial draft generated from issue #1014 triage. 19 REQs (Ubiquitous 5, Event-Driven 4, State-Driven 2, Optional 2, Unwanted 3, Integrity 3), 10 ACs (Given/When/Then), 6 explicit exclusions. Aligned with SPEC-V3R2-CON-001 v1.1.0 zone-codification template. |

---

## 1. Goal (목적)

Extend the FROZEN/EVOLVABLE zone primitive — partially codified in SPEC-V3R2-CON-001 — to **100% coverage** of the moai-adk rule tree, then **operationalize** it through a runtime-enforced `moai constitution validate` CLI verb backed by sentinel error governance and CI gating.

SPEC-V3R2-CON-001 introduced the zone registry and annotated 4 load-bearing files; it left a coverage gap (72 of ≥111 HARD rules = 65%). V3R5 closes the gap to 100% (≥111 entries) across 15 source files and adds:

1. A `zone_class` enum (`frozen-canonical | frozen-safety | evolvable-tuning | evolvable-experimental`) for sub-classification within Frozen and Evolvable zones.
2. An executable contract: `moai constitution validate [--strict] [--format json]` that diff-detects unregistered HARD rules, anchor drift, frozen entries missing canary_gate, and source-file deletions.
3. CI integration that blocks main-target PRs on validation drift (exit 1) or source-file loss (exit 2).
4. A `CONST-V3R5-NNN` namespace (parallel to V3R2 namespace, starting at 001) for new entries.

This SPEC **completes the zone codification**; the **5-layer safety amendment protocol** (autonomous tier-4 router, PreToolUse FrozenGuard hook) is W3 scope (SPEC-V3R3-HARNESS-AUTONOMY-001) and explicitly out of scope here.

## 2. Scope (범위)

### 2.1 In Scope (T2 Standard, W1)

**D1 — Inline annotation completeness**:
- Annotate every `[HARD]` rule across the canonical 4-file core with exactly one `[ZONE:Frozen]` or `[ZONE:Evolvable]` marker positioned immediately adjacent to the `[HARD]` token (same line preferred, next line acceptable).
- Canonical files (D1 scope):
  - `CLAUDE.md`
  - `.claude/rules/moai/core/moai-constitution.md`
  - `.claude/rules/moai/core/agent-common-protocol.md`
  - `.claude/rules/moai/core/askuser-protocol.md`
- Workflow/development files (D2-only — registry entries cover them; inline markers deferred to T3 follow-up unless trivial).

**D2 — zone-registry.md 100% coverage extension**:
- Add **39 new entries** in `CONST-V3R5-NNN` namespace (001..039), bringing total registry coverage to ≥111 entries.
- Extend the YAML entry schema with a new required field `zone_class:` of enum type:
  - `frozen-canonical` — load-bearing invariants (TRUST 5, SPEC+EARS, @MX, 16-language, Template-First, AskUserQuestion monopoly, Claude Code substrate)
  - `frozen-safety` — safety architecture, evaluator leniency prevention, GAN loop contract, pass-threshold floor
  - `evolvable-tuning` — orchestrator behavior knobs, adaptation weights, iteration limits
  - `evolvable-experimental` — feature-flagged or pilot-only rules
- Existing 72 entries gain `zone_class:` retroactively. Gap slots 047/048/050 + 073..149 remain reserved.
- Add `last_updated: YYYY-MM-DD` field per entry (enables REQ-CDL-019 STALE_ENTRY detection downstream).

**D3 — `moai constitution validate` CLI verb**:
- New subcommand under `internal/cli/constitution.go` (sibling to existing `moai constitution list` from V3R2-CON-001).
- Flags:
  - `--strict` — exit non-zero on any warning (default in CI mode per REQ-CDL-010)
  - `--format json|text` — output format selector (default `text`)
- Behavior:
  1. Load `.claude/rules/moai/core/zone-registry.md` and parse YAML entries.
  2. For each entry: verify `file:` path exists; verify `anchor:` resolves to an existing markdown heading or `[HARD]` line in the source.
  3. Scan all 15 source files for `[HARD]` occurrences; cross-check that each `[HARD]` has a registry entry.
  4. Verify every `zone: Frozen` entry has `canary_gate: true`.
  5. Verify every entry's `id:` matches the regex `^CONST-V3R[2-9]-\d{3}$` and is unique.
  6. Emit sentinel error codes (REQ-CDL-006..009, REQ-CDL-013, REQ-CDL-017..019).
- Exit codes:
  - `0` — clean
  - `1` — drift detected (HARD without entry, entry without HARD, anchor not found, frozen w/o canary, duplicate id, duplicate zone marker, stale entry)
  - `2` — source file referenced by registry no longer exists on disk

**CI integration**:
- New step in `.github/workflows/ci.yml` (or sibling workflow) runs `moai constitution validate --strict --format json` on every PR targeting `main`.
- Failure (exit 1 or 2) is a required status check per §18.7 branch protection extension (manual maintainer post-merge step).
- Environment variable `MOAI_CONSTITUTION_SKIP_VALIDATE=1` bypasses with a warning (REQ-CDL-011) — emergency hatch only.

### 2.2 Out of Scope (Explicit Exclusions)

- **EXCL-001**: W2 CORE-SLIM-001 (expert-backend/frontend/mobile retire, 12 W2-deferred lint findings) — separate W2 SPEC.
- **EXCL-002**: W3 HARNESS-AUTONOMY-001 (autonomous tier-4 router, PreToolUse FrozenGuard hook) — separate W3 SPEC.
- **EXCL-003**: Agent/skill frontmatter `zone:` field propagation — deferred to T3 follow-up SPEC.
- **EXCL-004**: PreToolUse FrozenGuard runtime hook scaffold — W3 scope.
- **EXCL-005**: Retroactive V3R3/V3R4 workflow-rule classification beyond the 15 currently-unmapped HARD sources — capped to existing surface.
- **EXCL-006**: Generation-time precommit hook (pre-push lefthook integration) — CI-time validation only for v3.5.0 cut; pre-commit deferred to follow-up if signal volume warrants.

## 3. Environment (환경)

### 3.1 Baseline state (verified 2026-05-20 on `main` HEAD `3bd2aa2`)

- `.claude/rules/moai/core/zone-registry.md` — 72 entries (`CONST-V3R2-001..046, 049, 051..072, 150..152`), gaps at 047/048/050 + 073..149.
- HARD rule counts per source file (`grep -c '\[HARD\]'`):
  - `CLAUDE.md`: 14
  - `.claude/rules/moai/core/moai-constitution.md`: 11
  - `.claude/rules/moai/core/agent-common-protocol.md`: 11
  - `.claude/rules/moai/core/askuser-protocol.md`: 3
  - `.claude/rules/moai/design/constitution.md`: 19
  - `.claude/rules/moai/workflow/worktree-integration.md`: 12
  - `.claude/rules/moai/workflow/ci-autofix-protocol.md`: 10
  - `.claude/rules/moai/workflow/ci-watch-protocol.md`: 8
  - `.claude/rules/moai/workflow/session-handoff.md`: 5
  - `.claude/rules/moai/workflow/context-window-management.md`: 5
  - `.claude/rules/moai/workflow/spec-workflow.md`: 3
  - `.claude/rules/moai/workflow/worktree-state-guard.md`: 1
  - `.claude/rules/moai/development/branch-origin-protocol.md`: 7
  - `.claude/rules/moai/development/skill-authoring.md`: 1
  - `.claude/rules/moai/development/agent-authoring.md`: 1
- **Total**: 111 HARD rules. **Gap**: 111 − 72 = **39 new entries required**.
- `internal/constitution/zone.go` — exists from V3R2-CON-001 (Zone enum + Rule struct with 6 fields). Will be extended (not replaced) with `ZoneClass` enum and `LastUpdated time.Time` field.
- `internal/cli/constitution.go` — has `list` subcommand. `validate` will be added as a sibling.

### 3.2 Adjacent SPEC context

- **SPEC-V3R2-CON-001** (status: implemented, v1.1.2): introduced zone registry primitive, 7 canonical FROZEN invariants, `moai constitution list` CLI. This SPEC extends both the data model and the CLI surface.
- **SPEC-V3R2-CON-002** (amendment protocol): conceptually adjacent; provides the 5-layer safety gate that PreToolUse FrozenGuard (W3) will enforce at runtime. V3R5 supplies the data plane; V3R2-CON-002 + W3 supply the control plane.
- **SPEC-V3R2-CON-003** (rule-tree consolidation): may shift file paths; if it merges before V3R5 ships, `file:` fields in registry entries are updated as a downstream maintenance task — does not block V3R5 if the consolidation is post-V3R5.

## 4. Stakeholders (이해관계자)

- **Primary**: moai-adk maintainers (Goos Kim, GoosLab) — author and ratify constitutional updates.
- **Secondary**: MoAI orchestrator (runtime) — consumes validated registry via `moai constitution list` and (future W3) PreToolUse FrozenGuard.
- **Tertiary**: SPEC authors — cross-reference `CONST-V3R5-NNN` IDs in future SPECs.
- **CI system**: GitHub Actions workflow — enforces `moai constitution validate --strict` on `main`-targeted PRs.

---

## 5. Requirements (EARS 형식)

### 5.1 Ubiquitous (REQ-CDL-001..005) — invariants

- **REQ-CDL-001**: Every `[HARD]` rule across the 15 constitutional source files **shall** carry exactly one `[ZONE:Frozen]` or `[ZONE:Evolvable]` marker positioned same-line or on the immediately following line.
- **REQ-CDL-002**: Every registry entry **shall** populate a `zone_class:` field with one of four enum values: `frozen-canonical`, `frozen-safety`, `evolvable-tuning`, `evolvable-experimental`.
- **REQ-CDL-003**: The `moai constitution validate` subcommand **shall** be invokable with `--format json` and emit a machine-parseable result envelope conforming to JSON Schema v1.0 (REQ-CDL-012).
- **REQ-CDL-004**: The zone registry **shall** achieve 100% HARD coverage with ≥111 entries (72 existing + 39 new V3R5 entries; gap slots retained as reserved IDs).
- **REQ-CDL-005**: New entries created under this SPEC **shall** use the `CONST-V3R5-NNN` namespace with `NNN` zero-padded three-digit numerals starting at 001 and incrementing contiguously (no gap slots in V3R5 range).

### 5.2 Event-Driven (REQ-CDL-006..009) — triggered

- **REQ-CDL-006**: **When** `moai constitution validate` detects drift between source `[HARD]` rules and registry entries, the CLI **shall** exit with code `1` and emit a `DRIFT_DETECTED` sentinel.
- **REQ-CDL-007**: **When** a PR targets `main`, the CI workflow **shall** execute `moai constitution validate --strict --format json` and block merge on non-zero exit unless `MOAI_CONSTITUTION_SKIP_VALIDATE=1` is set with explicit justification in the PR body.
- **REQ-CDL-008**: **When** a new `[HARD]` rule is added to a source file without a corresponding registry entry, the CLI **shall** emit `ZONE_UNREGISTERED` referencing the source file path and line number.
- **REQ-CDL-009**: **When** the zone-registry file is updated between invocations, the live registry parsing **shall** reflect the updates without any daemon restart (stateless parse per invocation).

### 5.3 State-Driven (REQ-CDL-010..011) — modal

- **REQ-CDL-010**: **While** running under CI (detected via `GITHUB_ACTIONS=true` or `CI=true`), the validate subcommand **shall** default to `--strict` mode even if the flag is not explicitly passed.
- **REQ-CDL-011**: **While** `MOAI_CONSTITUTION_SKIP_VALIDATE=1` is set, the validate subcommand **shall** emit a single-line warning to stderr (`[WARN] constitution validate bypassed via MOAI_CONSTITUTION_SKIP_VALIDATE`) and exit with code 0; no skipping the warning emission.

### 5.4 Optional (REQ-CDL-012..013) — conditional features

- **REQ-CDL-012**: **Where** `--format json` is requested, the output envelope **shall** conform to JSON Schema v1.0 with shape `{ "version": "1.0", "exit_code": 0|1|2, "errors": [{ "code": "<SENTINEL>", "file": "<path>", "line": <int>, "message": "<text>" }], "summary": { "total_hard": <int>, "registered": <int>, "coverage_pct": <float> } }` and **shall** be jq-compatible.
- **REQ-CDL-013**: **Where** a registry entry's `file:` path no longer exists on disk, the CLI **shall** emit `SOURCE_FILE_MISSING` and exit with code `2` (distinct from drift exit `1`).

### 5.5 Unwanted (REQ-CDL-014..016) — must-NOT

- **REQ-CDL-014**: The validate subcommand **shall not** modify any source file, registry file, or write any artifact outside of stdout/stderr. The CLI is strictly read-only.
- **REQ-CDL-015**: A registry entry with `zone: Frozen` **shall not** have `canary_gate: false` or a missing `canary_gate:` field. Violation emits `FROZEN_WITHOUT_CANARY` (exit 1).
- **REQ-CDL-016**: A registry entry **shall not** reference an `anchor:` value that does not resolve to either a markdown heading slug or a verbatim `[HARD]` substring in the referenced `file:`. Violation emits `ANCHOR_NOT_FOUND` (exit 1).

### 5.6 Integrity (REQ-CDL-017..019) — hybrid sentinel governance

- **REQ-CDL-017**: The registry **shall not** contain two entries with identical `id:` values. Violation emits `DUPLICATE_ID` (exit 1).
- **REQ-CDL-018**: A single `[HARD]` rule in a source file **shall not** carry two or more `[ZONE:*]` markers. Violation emits `DUPLICATE_ZONE_MARKER` (exit 1).
- **REQ-CDL-019**: A registry entry with `last_updated:` older than 365 days from validation execution date **shall** emit a non-blocking advisory `STALE_ENTRY` (exit 0 unless `--strict`). The `last_updated:` field becomes mandatory for all new V3R5 entries; existing V3R2 entries are grandfathered until first modification.

---

## 6. Acceptance Criteria (10 ACs, Given/When/Then)

- **AC-CDL-001 — D1 annotation completeness**
  - **Given** the 4 canonical core constitution files (CLAUDE.md + 3 core/*.md).
  - **When** the validator runs `grep -nE '\[HARD\]' <file>` and cross-checks each match for an adjacent `[ZONE:Frozen]` or `[ZONE:Evolvable]` token within the same line or the line immediately below.
  - **Then** every `[HARD]` match shall produce exactly one zone marker pairing. Auto-verifiable via `scripts/audit-zone-markers.sh` (delivered in run phase).

- **AC-CDL-002 — D2 100% coverage**
  - **Given** all 15 source files enumerated in §3.1 totaling 111 `[HARD]` rules.
  - **When** `moai constitution validate --format json` runs.
  - **Then** `summary.total_hard == 111`, `summary.registered == 111`, `summary.coverage_pct == 100.0`, and `errors` array contains zero `ZONE_UNREGISTERED` entries. Binds M=111 unification with AC-001.

- **AC-CDL-003 — D3 validate CLI happy path**
  - **Given** a clean repository state matching post-implementation main.
  - **When** `moai constitution validate` runs without flags.
  - **Then** exit code is `0`, stdout includes `coverage: 111/111 (100.0%)` line, stderr is empty.

- **AC-CDL-004 — D3 drift + sentinel coverage**
  - **Given** five synthetic mutations applied to fixture repo: (a) add a `[HARD]` rule with no registry entry; (b) rename a heading referenced by an anchor; (c) flip a Frozen entry to `canary_gate: false`; (d) delete a source file referenced by registry; (e) duplicate an entry `id:`.
  - **When** `moai constitution validate --strict --format json` runs on each fixture.
  - **Then** outputs contain the corresponding sentinel codes `ZONE_UNREGISTERED`, `ANCHOR_NOT_FOUND`, `FROZEN_WITHOUT_CANARY`, `SOURCE_FILE_MISSING`, `DUPLICATE_ID`. Exit codes are `1`, `1`, `1`, `2`, `1` respectively.

- **AC-CDL-005a — CI step exists and runs** (auto-verifiable)
  - **Given** the merged PR introducing this SPEC's run-phase artifacts.
  - **When** any subsequent PR targeting `main` triggers CI.
  - **Then** a `Constitution Validate` step appears in the workflow and runs to completion (pass or fail).

- **AC-CDL-005b — branch protection updated** (manual maintainer post-merge)
  - **Given** the maintainer runs the §18.7 `gh api -X PUT /branches/main/protection` invocation after merge.
  - **When** branch protection is queried with `gh api /branches/main/protection`.
  - **Then** the response includes `"Constitution Validate"` in `required_status_checks.contexts`.

- **AC-CDL-006 — zone_class enum validation**
  - **Given** the schema accepts only `frozen-canonical | frozen-safety | evolvable-tuning | evolvable-experimental`.
  - **When** a registry entry sets `zone_class: invalid-value`.
  - **Then** validate emits a YAML schema error and exits `1`.

- **AC-CDL-007 — CONST-V3R5-NNN ID format**
  - **Given** new entries in this SPEC use IDs `CONST-V3R5-001` through `CONST-V3R5-039`.
  - **When** validate parses each entry.
  - **Then** every `id:` matches `^CONST-V3R[2-9]-\d{3}$` and is unique across the registry.

- **AC-CDL-008 — live reload between writes**
  - **Given** a long-running test harness invokes validate twice with a registry edit between invocations.
  - **When** the second invocation runs.
  - **Then** the second result reflects the edited registry contents without any process restart, daemon, or warmup.

- **AC-CDL-009 — MOAI_CONSTITUTION_SKIP_VALIDATE bypass**
  - **Given** the registry contains a deliberate drift fault.
  - **When** `MOAI_CONSTITUTION_SKIP_VALIDATE=1 moai constitution validate` runs.
  - **Then** stderr contains the canonical warning string and exit code is `0`. Without the env var, the same invocation exits `1`.

- **AC-CDL-010 — validate CLI read-only assertion**
  - **Given** a checksum snapshot of all files under `.claude/rules/moai/` and `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/` taken before validation.
  - **When** `moai constitution validate --strict --format json` runs.
  - **Then** the post-validation checksum snapshot is byte-identical. The CLI process modifies no file on disk.

### 6.1 Edge cases (informative)

1. Multiple `[HARD]` rules on the same line — currently no such case exists; validator MAY treat as a single entry but SHOULD emit `MULTI_HARD_ON_LINE` advisory for tooling clarity.
2. `[HARD]` inside a fenced code block — must be skipped (false positive). Parser respects ` ``` ` fences.
3. Soft-wrapped marker (marker on next line) — REQ-CDL-001 explicitly permits adjacency on the line below to accommodate prose flow.
4. Anchor pointing to a heading slug that was kebab-renamed (e.g., `#1-hard-rules` → `#1-hard-rules-mandatory`) — emits `ANCHOR_NOT_FOUND`; remediation is registry update, not source rename revert.
5. Registry entry `file:` path uses backslashes on Windows fixtures — validator normalizes path separators before stat.
6. UTF-8 BOM in source files — parser strips BOM before regex match.

---

## 7. Constraints (제약 조건)

- **Backward compatibility**: V3R2 entry schema (6 fields) extended additively; new fields `zone_class:` and `last_updated:` are required on V3R5 entries, optional with grandfather clause on V3R2 entries.
- **CONST-V3R5 ID range**: 001..099 reserved for this SPEC; 100..149 reserved for in-namespace expansion; 150+ reserved for future cross-namespace mirrors (matches V3R2 convention).
- **Performance**: `moai constitution validate` must complete in <500ms on a warm cache repo (≤150 source files, ≤200 registry entries) and <2s cold.
- **Determinism**: Same input registry + same source tree must produce byte-identical JSON output across runs (sort errors by file:line ascending).
- **Cross-platform**: Validator must run on linux/amd64, darwin/arm64, darwin/amd64, windows/amd64 (the four release targets).
- **No external dependencies**: Stay within Go stdlib + existing moai-adk-go internal deps; no new go.mod additions for validation logic.

## 8. Open Questions (해결 필요 — run phase)

- **OQ1**: Should `STALE_ENTRY` advisory threshold (365 days) be configurable via `.moai/config/sections/system.yaml` or hardcoded? **Default**: hardcoded for v3.5.0 cut; promote to config in T3 follow-up if signal volume warrants.
- **OQ2**: Where does the `--strict` CI step output land? **Default**: GitHub Actions step summary via `$GITHUB_STEP_SUMMARY` with the JSON envelope rendered as a markdown table.
- **OQ3**: Should `zone_class: frozen-safety` entries require a secondary canary-gate signature (e.g., `safety_review_sha:`)? **Default**: deferred to W3 PreToolUse FrozenGuard SPEC.
- **OQ4**: Migration story for existing 72 V3R2 entries that lack `zone_class:` — auto-classify or require manual sweep? **Default**: lazy migration; entries gain `zone_class:` on next modification; bulk sweep delivered as a separate utility-scoped task in run phase if turnaround budget permits.

## 9. References

- **SPEC-V3R2-CON-001** (`.moai/specs/SPEC-V3R2-CON-001/spec.md`): zone codification primitive, 7 canonical FROZEN invariants, `moai constitution list` precedent.
- **SPEC-V3R2-CON-002**: amendment protocol (conceptual adjacency).
- **`.claude/rules/moai/core/zone-registry.md`**: existing 72-entry registry.
- **`.claude/rules/moai/design/constitution.md` §2**: pre-existing FROZEN/EVOLVABLE precedent inside the design subsystem (proves the pattern).
- **`.claude/rules/moai/development/spec-frontmatter-schema.md`**: 12-field canonical SPEC frontmatter SSOT.
- **`.moai/research/harness-autonomy-vision-2026-05-18.md` §3.1**: Two-Zone Architecture vision (motivates V3R5 closure).
- **`.moai/research/architecture-audit-2026-05-18.md` F-006**: zone-registry coverage-gap audit finding (65% → target 100%).
- **CLAUDE.local.md §18.7**: branch protection rule SSOT (CI required-status-checks list extended by AC-CDL-005b).

## 10. Glossary

- **Zone**: Top-level governance classification — `Frozen` (cannot be modified except by human + amendment protocol) or `Evolvable` (modifiable under safety gate).
- **zone_class**: Sub-classification under a Zone — distinguishes load-bearing canonical rules from safety mechanisms, tuning knobs, and experimental features.
- **Canary gate**: Pre-application shadow-evaluation requirement (Layer 2 of 5-layer safety architecture in `.claude/rules/moai/design/constitution.md` §5).
- **Sentinel error**: A stable, machine-readable error code (`ZONE_UNREGISTERED`, `ANCHOR_NOT_FOUND`, etc.) emitted by the validate CLI; consumers may script against the code without parsing prose.
- **HARD rule**: A clause marked with `[HARD]` indicating mandatory compliance (vs. SHOULD / advisory). 111 such rules existed in the moai-adk rule tree as of 2026-05-20.
- **CONST-V3R5-NNN namespace**: ID space allocated to V3R5 closure work. Parallel to (not replacing) V3R2 namespace.

---

**Status**: draft
**Tier**: T2 Standard
**Wave**: Mega-Sprint W1 (W0→W1→max(W2,W3)→W4→v3.5.0)
**Next phase**: `/moai plan SPEC-V3R5-CONSTITUTION-DUAL-001 --plan-audit` to produce `plan.md` (Phases A/B/C/D), then `acceptance.md`, then run-phase implementation.
