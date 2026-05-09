# Plan — SPEC-V3R2-SPC-001 EARS + Hierarchical Acceptance Criteria

> Phase 1B implementation plan. Based on research.md gap analysis: ~80% of behaviour already implemented in `internal/spec/`. Plan focuses on documentation, audit, performance benchmark, and CON-002 amendment paperwork.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-spec (Wave 9 plan author) | Initial plan covering M1–M5 milestones for SPC-001. |

---

## 1. Goal Recap

Establish hierarchical (parent → child) acceptance-criteria schema for moai SPECs while preserving 100% back-compat with the 185 existing flat SPECs. Land the FROZEN-zone amendment per CON-002 protocol.

## 2. Approach

Audit-and-fill rather than greenfield. Research confirms the Go data model, parser, errors, lint integration, and `moai spec view` are already on `main`. Plan tasks therefore consist of:

- Verifying gaps surfaced in research §5 (`--shape-trace` field emission, perf benchmark).
- Authoring the hierarchical-AC documentation block in `.claude/rules/moai/workflow/spec-workflow.md`.
- Filing CON-002 amendment evidence (Canary, HumanOversight).
- Cross-linking SPC-003 / HRN-002 / HRN-003 / MIG-001.

## 3. Milestones

### M1 — Schema design + research delivery — Priority Critical

Owner: `manager-spec`
Deliverables:
- This plan.md, the research.md, the acceptance.md, and tasks.md as a 6-file plan bundle.
- Self-demonstrating hierarchical AC inside acceptance.md.
- File:line anchor coverage (≥25, achieved 48 in research.md §10).

mx_plan tags:
- `@MX:NOTE` on research.md §6 schema canonical form.
- `@MX:ANCHOR` on plan.md §3 milestone table (fan_in expected ≥3 from MIG-001, SPC-003, HRN-002 plans).

Exit criteria:
- plan-auditor independent verification PASS.
- ≥25 file:line anchors in research.md.
- All 18 REQs listed in spec.md §5 mapped to ≥1 AC in acceptance.md.

### M2 — Performance benchmark + parser hardening — Priority High

Owner: `expert-backend` (delegated by manager-tdd in run phase)
File:line anchors:
- `internal/spec/parser_test.go` (new benchmark function `BenchmarkParse365Leaves`).
- `internal/spec/parser.go:200-227` (auto-wrap path under bench harness).
- `internal/spec/testdata/hierarchical-ac/spec.md` (extend fixture or add `internal/spec/testdata/perf-365-leaves/spec.md`).

Tasks:
- Generate fixture with 55 top-level parents × ~6.6 average children = 365 leaves matching DevAI shape (REQ-SPC-001-001, AC-SPC-001-14).
- Add `go test -bench .` assertion: `<500ms` per AC-SPC-001-14.
- Cover `acceptance_format: flat` malformed-frontmatter cases (research.md R6).

mx_plan tags:
- `@MX:WARN reason="365-leaf parse perf budget"` on the benchmark function.
- `@MX:NOTE` on the perf fixture explaining DevAI 55→365 ratio.

Exit criteria:
- Benchmark passes <500ms locally on Apple M-series and Linux GitHub runner.
- All existing `internal/spec/...` tests remain green.

### M3 — Documentation + spec-workflow.md amendment — Priority High

Owner: `manager-docs` (run phase — sync sub-task) with manager-spec author oversight
File:line anchors:
- `.claude/rules/moai/workflow/spec-workflow.md:88-104` — insert hierarchical-AC schema block after Plan Phase output bullet list.
- `.claude/rules/moai/core/zone-registry.md:16-19` — update CONST-V3R2-001 entry with cross-link to SPC-001.
- `.claude/skills/moai-workflow-spec/SKILL.md` — add §"Hierarchical Acceptance" subsection under "EARS Format Deep Dive".
- `internal/cli/spec_view.go:41-158` — verify `--shape-trace` emits `depth` + `parent_id` fields (REQ-SPC-001-031, AC-SPC-001-08); add CLI test if missing.

Tasks:
- Author the canonical schema example block (research.md §6) into spec-workflow.md.
- Document Given inheritance semantics + leaf-only REQ tail rule.
- Add `--shape-trace` audit + test coverage.
- Update SKILL.md skill body with concise schema reference (Quick Reference + Implementation Guide).
- Cross-link SPC-002 (@MX TAG), SPC-003 (linter), HRN-002 (Sprint Contract), HRN-003 (per-leaf scoring), MIG-001 (migrator).

mx_plan tags:
- `@MX:NOTE` on spec-workflow.md amendment block: "Hierarchical AC introduced by SPC-001; see acceptance.md for self-demonstration."

Exit criteria:
- Doc lint (`scripts/docs-i18n-check.sh` if applicable) passes.
- `moai spec view <SPEC-XXX> --shape-trace` output includes `depth` + `parent` for each node.
- SKILL.md update preserves existing 6-pillar structure.

### M4 — Existing SPEC migration handoff to MIG-001 — Priority Medium

Owner: `manager-spec` (handoff coordination only) — implementation owned by SPEC-V3R2-MIG-001 run phase
File:line anchors:
- `.moai/specs/SPEC-V3R2-MIG-001/spec.md` — confirm references to BC-V3R2-011 + SPC-001 wrap behaviour are current.
- `.moai/specs/SPEC-V3R2-SPC-001/spec.md:151` — Migration paragraph cross-links MIG-001.
- `internal/spec/parser.go:200-227` — read-path compatibility already covers all 185 SPECs.

Tasks:
- Author handoff note in MIG-001 (or its plan.md when filed) referencing SPC-001 wrap-synthesis behaviour.
- Verify MIG-001 SPEC's "AC wrap" paragraph aligns with `parser.go:200-227` actual behaviour.
- No source changes to existing 185 SPECs in this milestone — runtime auto-wrap is sufficient (research §7 D4).

mx_plan tags:
- `@MX:TODO resolved="MIG-001 run phase"` on the SPC-001 §2.1 migration bullet.

Exit criteria:
- MIG-001 SPEC confirms it owns the optional cosmetic rewrite.
- No regression in lint test suite.

### M5 — REFACTOR + MX tags + completion gate (CON-002 paperwork) — Priority Critical

Owner: `manager-quality` + `manager-spec`
File:line anchors:
- `.moai/specs/SPEC-V3R2-SPC-001/spec.md:225-231` — §11.3 Amendment Safety Gate evidence.
- `internal/spec/...` — all existing tests must remain green.
- `.moai/specs/SPEC-V3R2-SPC-001/progress.md` — final status `audit-ready` after plan, `complete` after run-phase Canary.

Tasks:
- **FrozenGuard evidence**: confirm SPC-001 amendment is strictly additive (children optional, flat parseable). Document in progress.md.
- **Canary evidence**: re-parse the last 10 landed v2 SPECs (e.g., SPEC-V3R2-WF-005, SPEC-V3R3-CLI-TUI-001 M1/M2/M3, SPEC-V3R3-CIAUT-001, SPEC-V3R2-WF-002/003/004, SPEC-V3R2-CON-002, SPEC-V3R2-MIG-001) with the new parser. All MUST parse without warnings and yield identical REQ-coverage. Capture script output.
- **ContradictionDetector evidence**: confirm no existing rule prohibits nesting. Already noted in spec.md §11.3; cite zone-registry CONST-V3R2-001 as evidence.
- **RateLimiter evidence**: SPC-001 is 1 of ≤3 FROZEN amendments per v3.x cycle (with HRN-002, optionally CON-002 itself). Document.
- **HumanOversight evidence**: maintainer approval at plan-auditor iteration 2; record approval timestamp + reviewer in landing PR description.
- Add `@MX:ANCHOR fan_in=4` to acceptance.md for the canonical hierarchical example (cross-referenced by SPC-002, SPC-003, HRN-002, HRN-003).
- Add `@MX:NOTE` on `internal/spec/ears.go:11-18` Acceptance struct: "Hierarchical AC schema per SPC-001 §11.2".

mx_plan tags:
- `@MX:ANCHOR fan_in=4` on acceptance.md self-demonstrating example.
- `@MX:NOTE` on `internal/spec/ears.go:11` `Acceptance` struct definition.
- `@MX:WARN reason="FROZEN-zone amendment"` on the spec-workflow.md insertion point.

Exit criteria:
- progress.md `plan_status: audit-ready` (after this PR merges).
- Canary script output committed under `.moai/specs/SPEC-V3R2-SPC-001/canary-v2-reparse.txt` during run phase.
- HumanOversight approval recorded in landing PR.

---

## 4. Technical Approach

### 4.1 Stack & touch points

- Go subsystem: `internal/spec/` (no behavioural change; only test additions).
- CLI surface: `internal/cli/spec_view.go` (`--shape-trace` field audit only).
- Rule files: `.claude/rules/moai/workflow/spec-workflow.md`, `.claude/rules/moai/core/zone-registry.md`.
- Skill: `.claude/skills/moai-workflow-spec/SKILL.md`.
- SPEC corpus: 185 SPECs untouched (read-path auto-wrap covers compatibility).

### 4.2 Risk register (carried from research §9)

| # | Risk | Severity | Mitigation milestone |
|---|------|----------|-----------------------|
| R1 | Tab-vs-space indentation | MEDIUM | M3 documents 2-space minimum |
| R2 | 365-leaf perf unmeasured | LOW | M2 benchmark |
| R3 | `--shape-trace` field emission | MEDIUM | M3 audit + CLI test |
| R4 | CON-002 paperwork incomplete | HIGH | M5 |
| R5 | Tree glyph rendering | LOW | M3 ASCII fallback test |
| R6 | YAML frontmatter edge cases | LOW | M2 fixture coverage |

### 4.3 Dependencies

Blocked by:
- SPEC-V3R2-CON-001 — confirmed FROZEN status of EARS modality (zone-registry CONST-V3R2-001 already in main).

Blocks:
- SPEC-V3R2-SPC-003 — SPEC linter consumes hierarchical schema.
- SPEC-V3R2-HRN-002 — Sprint Contract per-leaf state.
- SPEC-V3R2-HRN-003 — evaluator-active per-leaf scoring.
- SPEC-V3R2-MIG-001 — optional cosmetic AC rewrite.

### 4.4 Out-of-scope confirmations

Per spec.md §2.2 and research §8:
- REQ ID hierarchy (REQs stay flat).
- EARS modality additions (CON-001 FROZEN).
- Sprint Contract state shape (HRN-002).
- Per-leaf scoring (HRN-003).
- @MX TAG schema (SPC-002).
- Linter implementation (SPC-003).
- Migrator implementation (MIG-001).

---

## 5. Quality Gates

### 5.1 Plan-phase gates (this PR)

- [ ] `plan-auditor` independent verification — PASS at iteration ≤2.
- [ ] research.md ≥25 file:line anchors — achieved 48.
- [ ] Every REQ in spec.md §5 mapped to ≥1 AC in acceptance.md (18 REQs verified).
- [ ] Every AC in acceptance.md cites ≥1 REQ.
- [ ] tasks.md uses `T-SPC001-NN` naming with owner role per task.
- [ ] progress.md frontmatter `plan_status: audit-ready`.
- [ ] No `spec.md` modifications in this PR (read-only) — spec.md remains v0.1.0.

### 5.2 Run-phase gates (future PR, scoped by tasks.md)

- [ ] `go test ./internal/spec/...` and `go test ./internal/cli/...` green.
- [ ] `go test -bench BenchmarkParse365Leaves ./internal/spec/...` <500ms.
- [ ] `moai spec view SPEC-XXX --shape-trace` test fixture passes.
- [ ] Canary re-parse evidence committed.
- [ ] CON-002 HumanOversight approval recorded in PR description.
- [ ] TRUST 5: Tested / Readable / Unified / Secured / Trackable all green.

### 5.3 Sync-phase gates

- [ ] `.claude/rules/moai/workflow/spec-workflow.md` amendment landed.
- [ ] `.claude/skills/moai-workflow-spec/SKILL.md` body updated.
- [ ] zone-registry CONST-V3R2-001 cross-linked to SPC-001.
- [ ] CHANGELOG entry under `### Changed` referencing BC-V3R2-011.
- [ ] docs-site 4-language sync (per CLAUDE.local.md §17) deferred to MIG-001 cycle if rule-only doc change.

---

## 6. Rollout Strategy

This is an additive amendment — no breaking change at runtime because the parser's auto-wrap behaviour is already shipped. The "breaking: true" flag in spec.md frontmatter reflects the **schema-level** semantic change (flat → hierarchical is a contract evolution), not a runtime regression. Rollout proceeds as:

1. Plan PR (this) — research, plan, AC, tasks artifacts only.
2. plan-auditor iteration → approval.
3. Plan PR merge.
4. Run PR — perf benchmark + `--shape-trace` audit + spec-workflow.md amendment + Canary evidence.
5. Run PR merge.
6. Sync PR — SKILL.md + CHANGELOG + cross-links.
7. SPC-003 plan kicks off, consuming hierarchical schema.

No rollback procedure required — additive change. If perf benchmark fails, rollback strategy: revert M2 fixture only (parser code is already on main and unaffected).

---

## 7. Resource Estimates (priority-based; no time predictions per agent-common-protocol)

| Milestone | Priority | Files touched (estimate) | LOC delta (estimate) |
|-----------|----------|--------------------------|----------------------|
| M1 | Critical | 6 (this plan bundle) | +1500 |
| M2 | High | 3 (`parser_test.go`, perf fixture, optional small parser fix) | +200 |
| M3 | High | 4 (`spec-workflow.md`, `zone-registry.md`, `SKILL.md`, `spec_view.go` test) | +250 |
| M4 | Medium | 1 (MIG-001 plan note) | +30 |
| M5 | Critical | 3 (`progress.md`, Canary script output, PR description) | +400 |

Total estimated delta across all milestones: ~2400 LOC of which ~1500 is plan documentation and ~900 is code/test additions.

---

## 8. Communication & Handoff

### 8.1 To plan-auditor

- Read research.md §5 gap analysis first — answers "is this work redundant with what's already on main?"
- Read research.md §10 file:line anchors for grounded claims.
- Verify acceptance.md self-demonstrates hierarchical schema (parent → child tree on at least 3 ACs).
- Confirm tasks.md owner roles align with manager-spec / expert-backend / manager-quality split.

### 8.2 To run-phase agent (manager-tdd or expert-backend)

- Start from M2 (perf benchmark) — lowest risk, validates parser at scale.
- Then M3 (doc amendment) — coordinate with manager-docs for SKILL.md + spec-workflow.md.
- Then M5 (CON-002 paperwork) — Canary script must run before HumanOversight approval.
- Then M4 (MIG-001 handoff note) — final cross-link.

### 8.3 To dependents (SPC-003, HRN-002, HRN-003, MIG-001)

- SPC-003 plan-phase agent: SPEC linter MUST consume `internal/spec/lint.go` `collectAllREQIDs` (line 394-403) for tree-aware REQ↔AC coverage.
- HRN-002: Sprint Contract durable state references leaf IDs `AC-XXX-NN.a.i` produced by parser tree.
- HRN-003: per-leaf evaluator scoring iterates `Acceptance.Children` recursively (model already in `internal/spec/ears.go`).
- MIG-001 run phase: optional cosmetic AC rewrite consults parser auto-wrap behaviour as the runtime equivalent.

---

## 9. Done Definition

This plan is `audit-ready` when:
- All 6 plan files exist (`spec.md`, `research.md`, `plan.md`, `acceptance.md`, `tasks.md`, `progress.md` + `spec-compact.md`).
- plan-auditor PASS at iteration ≤2.
- PR `plan(spec): SPEC-V3R2-SPC-001 — EARS + hierarchical acceptance criteria` opened against main.
- progress.md frontmatter `plan_status: audit-ready`.

End of plan.
