# spec-compact — SPEC-V3R2-HRN-003

One-page reference card. For full text see `spec.md`, `research.md`, `plan.md`, `acceptance.md`, `tasks.md`.

## Identity

| Field | Value |
|-------|-------|
| ID | SPEC-V3R2-HRN-003 |
| Title | Hierarchical Acceptance Scoring (4-dimension × sub-criteria) |
| Phase | v3.0.0 — Phase 5 — Harness + Evaluator |
| Priority | P1 High |
| Status (spec.md) | draft (v0.2.0 audit pass complete) |
| Status (plan.md) | audit-ready |
| Breaking | false |
| Lifecycle | spec-anchored |

## Goal in 3 lines

Implement hierarchical 4-dimension × per-sub-criterion scoring in `internal/harness/scorer.go` + `rubric.go`, consuming SPC-001's `[]Acceptance` tree and producing structured JSON ScoreCards. Augment evaluator-active body with hierarchical output schema + rubric-citation enforcement; persist per-sub-criterion state to Sprint Contract per HRN-002 §11.4.1 substrate. Source pattern: pattern-library E-1 (priority 9, ADOPT) + E-3 (ADOPT); FROZEN constraints from design-constitution §12 Mechanism 1/3 + §5 floor.

## REQ Index (19 total)

| Modality | REQ IDs |
|----------|---------|
| Ubiquitous (5.1) | 001, 002, 003, 004, 005, 006 |
| Event-Driven (5.2) | 007, 008, 009, 010, 011 |
| State-Driven (5.3) | 012, 013, 014 |
| Optional (5.4) | 015, 016 |
| Unwanted (5.5) | 017, 018, 019 |

## AC Index (12 total; 4 self-demonstrate hierarchy with depth-3 grandchildren on AC-04)

| AC ID | Hierarchical? | Mapped REQs |
|-------|---------------|-------------|
| AC-HRN-003-01 | flat | 001 |
| AC-HRN-003-02 | flat (cross-refs REQ-017) | 002, 004, 017 |
| AC-HRN-003-03 | YES (3 children .a/.b/.c) | 007, 015 |
| AC-HRN-003-04 | YES (3 children + 2 grandchildren on .a) | 008 |
| AC-HRN-003-05 | YES (3 children .a/.b/.c) | 009 |
| AC-HRN-003-06 | flat (transitive — SPC-001 substrate) | 010 |
| AC-HRN-003-07 | YES (3 children .a/.b/.c) | 003, 005, 013, 016 |
| AC-HRN-003-08 | flat | 001, 012, 019 |
| AC-HRN-003-09 | flat | 006 |
| AC-HRN-003-10 | flat | 011 |
| AC-HRN-003-11 | flat | 018 |
| AC-HRN-003-12 | flat | 008, 014 |

## Drift Reconciliations (spec.md v0.2.0)

| Drift | Discovered | Resolution |
|-------|------------|------------|
| Profile format `.yaml` (spec) → `.md` (main) | Profiles ship at `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` since 2026-05-10 | REQ-005 v0.2.0 adopts `.md` parser; no `.yaml` migration |
| evaluator-active body REQ-006 "introduce" → augment | Body lines 47-54 already declare 4-dim table; lines 91-92 already cite §11.4.1 (HRN-002 M3) | REQ-006 v0.2.0 reframed as augment (hierarchical JSON section + rubric-citation requirement) |
| `internal/harness/gan_loop.go` does not exist | Confirmed `ls` no such file | REQ-011 wires via Go-side `WriteContract()` helper called by orchestrator-level SKILL.md (HRN-002 D1 inherited) |

## Files Touched in Run Phase

| File | Purpose | M |
|------|---------|---|
| `internal/harness/scorer.go` (NEW ~250 LOC) | Dimension enum + ScoreCard tree + EvaluatorRunner.Score() + aggregation + must-pass firewall + WriteContract() | M2+M3 |
| `internal/harness/rubric.go` (NEW ~120 LOC) | Rubric struct + ParseRubricMarkdown for `.md` profiles + ValidateCitation | M2+M4 |
| `internal/harness/scorer_test.go` (NEW ~400 LOC) | All 12 ACs unit tests + integration | M5 |
| `internal/harness/rubric_test.go` (NEW ~150 LOC) | Markdown parser + validators | M5 |
| `internal/config/types.go` (+30 LOC) | EvaluatorConfig extensions: Profiles, Aggregation, MustPassDimensions | M4 |
| `internal/config/loader.go` (+40 LOC) | LoadHarnessConfig profile-loading step | M4 |
| `internal/config/errors.go` (+20 LOC) | 4 new sentinels: ErrUnknownDimension, ErrRubricCitationMissing, ErrFlatScoreCardProhibited, ErrMustPassBypassProhibited | M2 |
| `.claude/agents/moai/evaluator-active.md` (+35 LOC) | "## Hierarchical Score Output (Phase 5)" section | M4 |
| `.claude/skills/moai-workflow-gan-loop/SKILL.md` (+25 LOC) | "### Phase 3b: Hierarchical Scoring" subsection | M4 |
| Template mirrors of agent + SKILL files | Template-First per CLAUDE.local.md §2 | M4 |
| `.claude/rules/moai/core/zone-registry.md` (+25 LOC) | CONST-V3R2-154 (4-dim enum) + CONST-V3R2-155 (4 anchor levels) per OQ1 default | M5 |
| `internal/harness/testdata/profiles/malformed-{5dim,bypass}.md` (NEW) | Negative fixtures | M5 |

## Files NOT Modified

- `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` — already on main with correct rubric structure (read-side only).
- `.claude/rules/moai/design/constitution.md` — §11.4.1 already amended by HRN-002; HRN-003 reads §5/§11.4.1/§12 only.
- `internal/spec/{ears,parser,lint}.go` — SPC-001 surface; HRN-003 consumes `[]Acceptance` tree only.
- `internal/harness/{gan_loop.go}` — does not exist; HRN-003 inherits HRN-002 D1 (no Go-side runner).
- `internal/harness/evaluator_leak.go` — HRN-002 substrate; transitively satisfies REQ-013.

## Adjacent SPECs

| Type | SPEC | Relationship |
|------|------|--------------|
| Blocked by | SPEC-V3R2-CON-001 | FROZEN zone model + zone-registry infrastructure (M5 OQ1) |
| Blocked by | SPEC-V3R2-HRN-001 | HarnessConfig minimal substrate (HRN-003 M4 extends EvaluatorConfig) |
| Blocked by | SPEC-V3R2-HRN-002 | §11.4.1 fresh-judgment (REQ-013 cross-ref); evaluator-active body cross-reference |
| Blocked by | SPEC-V3R2-SPC-001 | Hierarchical AC parser; HRN-003 consumes `[]Acceptance` tree |
| Blocks | SPEC-V3R2-WF-003 | Multi-mode router thorough harness uses hierarchical scoring |
| Blocks | SPEC-V3R2-MIG-001 | Migrator output consumed by HRN-003 |
| Pattern source | pattern-library E-1 (priority 9 ADOPT) | Agent-as-a-Judge per-sub-criterion scoring |
| Pattern source | pattern-library E-3 (ADOPT) | Rubric-Anchored scoring |

## Open Questions (defaults proposed; auditor verify)

| OQ | Question | Default |
|----|----------|---------|
| OQ1 | Zone-registry mirror entries: M5 task or follow-up SPEC? | M5 task — register CONST-V3R2-154 + 155 |
| OQ2 | Aggregation default min vs mean? | CONFIRM `min` |
| OQ3 | Must-pass dimensions default set? | `[Functionality, Security]` exported; floor `[Security]` (REQ-018 prevents narrowing below) |
| OQ4 | JSON schema versioning? | YES — `schema_version: "v1"` |
| OQ5 | Rubric-anchor citation: strict reject vs warn? | STRICT reject + retry-on-reject max 2 |

## Risks (top 5)

| Risk | Severity | Mitigation milestone |
|------|----------|-----------------------|
| evaluator-active inconsistent JSON output | HIGH | M4 — schema declared in body; M3 — REQ-009 strict enforcement; retry-on-reject |
| Markdown rubric table parser brittle | MEDIUM | M2 — tolerant parser; M5 — malformed fixtures cover |
| Min-aggregation too strict for exploratory SPECs | MEDIUM | M3 — `mean` opt-in per REQ-015 + lenient profile |
| HRN-001 EvaluatorConfig field name conflict | MEDIUM | M4 — Decision D7 additive merge convention |
| OQ1 deferral leaves FROZEN constraints unregistered | MEDIUM | M5 — default register CONST-V3R2-154/155; fallback document deferral + open issue |

End of compact.
