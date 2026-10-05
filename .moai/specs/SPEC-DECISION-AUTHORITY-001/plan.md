---
id: SPEC-DECISION-AUTHORITY-001
title: "Implementation plan — decision-gate axis, decision-index artifact, kickoff enrichment"
version: "0.1.0"
created: 2026-09-13
---

# plan.md — SPEC-DECISION-AUTHORITY-001

## §A Context

- **Epic**: issue #1683 Human Decision Authority — a two-card series. Item #2
  (SPEC-JUDGMENT-FIRST-MODE-001, `status: completed`) landed first; this SPEC is item #1.
- **Worktree**: `.claude/worktrees/t692`, branch `WT-judgment-authority`, base `62fbd6baf`
  (local develop head at authoring).
- **Tier**: L (5 artifacts + progress.md). plan-auditor PASS threshold 0.85.
- **Plan-phase artifacts**: `.moai/specs/SPEC-DECISION-AUTHORITY-001/{spec,plan,acceptance,design,research,progress}.md`.
- **Cycle**: TDD (`quality.yaml` `constitution.development_mode` default) — the config-axis
  milestone M1 is RED-GREEN testable; the doctrine-text milestones M2-M3 are verified by the
  static sweeps of acceptance.md.
- **PRESERVE targets** (unchanged by every milestone):
  - `.claude/agents/moai/plan-auditor.md` (REQ-DA-014 — byte-unchanged)
  - `.claude/rules/moai/core/askuser-protocol.md` (the pull convention; read-only dependency)
  - `.claude/skills/moai/workflows/plan/spec-assembly.md:217` gate clause **structure** — the
    `[HARD]` mandatory/score-independent sentence and its three canonical options (the
    decision-index presentation is inserted beside them, never into them)
  - `internal/config/types.go` `RecommendationMode` field and `ResolvedRecommendationMode()`
  - All other SPEC directories under `.moai/specs/`

## §B Known Issues (filtered to relevant categories)

- **B4 Frontmatter schema** — `decision-index.md` is stateless: NO `status:` field
  (spec-frontmatter-schema.md § Artifact Statelessness). plan/acceptance/design/research of
  this SPEC likewise carry none.
- **B6 spec-lint heading convention** — this spec.md's exclusions use `### Out of Scope —`
  H3 sub-headings with `-` bullets (OutOfScopeRule).
- **B3/B11 Subagent boundary** — no hook, no Go pipeline change, no AskUserQuestion anywhere
  in the deliverables. The gate presentation is doctrine text executed by the orchestrator.
- **B8 Working-tree hygiene** — do not touch `.moai/state/`, `.moai/cache/`, runtime logs.
- **CONST-3 (moai update wipe)** — every `.claude/` / `.moai/config/` file edited here needs
  its template mirror in the same milestone or the next `moai update` destroys the change.
- **CONST-8 (C1↔C2 branching)** — `manager-spec.md`, `clarity-interview.md`,
  `interview.yaml` mirrors already diverge from local copies. Mirror verification is
  token-presence in both trees, never `diff -q`.
- **R1 Coordinate decay** — every line number in this plan is a locating aid at `62fbd6baf`.
  Locate by the quoted clause text.

## §C Pre-flight

```bash
git -C .claude/worktrees/t692 branch --show-current     # WT-judgment-authority
git -C .claude/worktrees/t692 rev-parse --short HEAD     # re-read before first commit
go build ./...                                           # baseline green
go test ./internal/config/...                            # baseline green (scoped, not full suite)
golangci-lint run --timeout=2m ./internal/config/...     # baseline clean
```

RED probes P1-P18 (acceptance.md §E evidence ledger) were executed this tree at `62fbd6baf`
before artifact authoring; re-pin any re-measured probe at the milestone that flips it.

## §D Constraints

Verbatim from spec.md §D (CONST-1..CONST-8). Non-negotiable highlights: template default
`off`; plan-auditor untouched; no deny paths; no time estimates; mirror-the-change-not-the-
bytes.

## §E Self-Verification

Each §E item reported per verification-claim-integrity §3 (Claim / Evidence /
Baseline-attribution / Gaps / Residual-risk), with the attribution triple (command, verbatim
output, tree SHA):

- **E1** — AC binary PASS/FAIL matrix against acceptance.md §D.
- **E2** — `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (config
  package touched → cross-platform build applies).
- **E3** — `go test -cover ./internal/config/...` ≥ 85% (the new resolver + field).
- **E5** — `golangci-lint run --timeout=2m` — NEW issues vs pre-existing baseline, named
  separately.
- **E6** — branch HEAD + commit list; commits carry `feat(SPEC-DECISION-AUTHORITY-001): M<N>`
  subjects and the card id (t692) in the body; `🗿 MoAI` trailer.
- **E8** — RED failure output for M1 (the `TestDecisionGate` tests written first and observed
  failing before the resolver lands).

## §F Milestones

Ordered by decision-reversibility — the data-model change first, user-facing flow next,
mechanical mirror/build steps last.

### M1 — Config axis (Priority High; data-model decision)

Files: `internal/config/types.go`, `internal/config/defaults.go`,
`internal/config/loader_interview.go` (no change expected — loader is generic),
`internal/config/interview_decision_gate_test.go` (new),
`internal/template/templates/.moai/config/sections/interview.yaml`,
`.moai/config/sections/interview.yaml` (local dogfood: `decision_gate: on`).

1. RED: write `TestDecisionGateDefaultIsOff`, `TestDecisionGateAbsentKeyIsOff`,
   `TestDecisionGateOnRoundTrip`, `TestDecisionGateUnknownValueResolvesOff` — observe them
   fail (E8).
2. Add `DecisionGate string` + `yaml:"decision_gate"` to `InterviewConfig`; default `"off"`;
   `ResolvedDecisionGate()` mirroring `ResolvedRecommendationMode()` semantics (absent/empty/
   unknown → `off`; only `on` activates). The resolver reads its own field only (REQ-DA-018).
3. GREEN: tests pass; `go test ./internal/config/...` green.
4. Template `interview.yaml`: add `decision_gate: off` with a one-line comment. Local
   `interview.yaml`: `decision_gate: on` (dogfood, per OD-1).

Flips: AC-DA-001, AC-DA-002, AC-DA-003, AC-DA-017.

### M2 — Index authoring flow (Priority High; user-facing flow decision)

Files: `.claude/agents/moai/manager-spec.md` + mirror,
`.claude/skills/moai/workflows/plan/clarity-interview.md` + mirror.

1. manager-spec Step 4 (Create SPEC Documents): add the decision-index authoring obligation —
   **Where** `interview.decision_gate` is `on`, author `.moai/specs/SPEC-{ID}/decision-index.md`
   (stateless, fixed row shape per spec.md §B) alongside the Tier artifact set; **Where** off,
   no index is created (REQ-DA-005/006/007).
2. The four-label routing rules with the authority register (REQ-DA-008/009/010) and the
   escalation rule (REQ-DA-011) — quoted from spec.md §C.3, adapted to manager-spec's voice.
3. clarity-interview Phase 4/Phase 8: decisions the user does not settle, or whose authority
   no document carries, are recorded as index-row candidates instead of resolved by default
   (REQ-DA-012); each row carries Detect → Explain → Ask, no preferred answer (REQ-DA-017).
4. All mode-conditioned mentions carry the `Where ... on` / `off` conditioning inside their
   enclosing markdown block (AC-DA-004 sweep shape, per SPEC-JFM 0.2.4's block-window rule).

Flips: AC-DA-004, AC-DA-005, AC-DA-006, AC-DA-007, AC-DA-008, AC-DA-009, AC-DA-016.

### M3 — Kickoff gate enrichment (Priority High; user-facing gate decision)

Files: `.claude/skills/moai/workflows/plan/spec-assembly.md` + mirror.

1. After the plan HTML emission step (Step 2.3.3a family), add the decision-index presentation
   step: **Where** the gate is `on` and `decision-index.md` exists, the orchestrator surfaces
   the index rows as additive prose context in the SAME turn the Implementation Kickoff
   Approval `AskUserQuestion` fires — same fail-open shape as the HTML report (index absent or
   unreadable → skip silently, gate unchanged) (REQ-DA-013).
2. Verdict recording: operator verdicts per row (`DECIDE` / `NEED_ANALYSIS` / `NEED_EVIDENCE`
   / `DEFER`) are written back into the row after the gate (REQ-DA-015); product-level
   verdicts surface the `product.md` reconciliation decision with the manager-docs ownership
   boundary named (REQ-DA-016, §E.3 OD-2).
3. The zero-rows clause: zero judgment points is not approval (REQ-DA-019).
4. [PRESERVE] The `[HARD]` mandatory-gate sentence, its three canonical options, and the
   `(권장)`-label clause stay verbatim; the insertion is beside, not inside (AC-DA-011).

Flips: AC-DA-010, AC-DA-013, AC-DA-014, AC-DA-015.

### M4 — Template mirrors + embed regeneration (Priority Medium; mechanical)

1. Apply the M2/M3 changes to the three template mirrors
   (`internal/template/templates/.claude/agents/moai/manager-spec.md`,
   `.../plan/spec-assembly.md`, `.../plan/clarity-interview.md`).
2. Template neutrality pass on every template-side change (REQ-DA-021): no SPEC IDs, no REQ
   tokens, no internal dates, no SHAs, no macOS-bias, no `CLAUDE.local.md` references.
3. `make build` to regenerate the embedded filesystem; `make agents-emit` not required (no
   `.codex/` agent toml consumed here — but verify `make build`'s `agents-emit-check` passes
   since `manager-spec.md` is an emitted source).
4. Mirror verification: token-presence in both trees per edited file (AC-DA-018).

Flips: AC-DA-018, AC-DA-021 via the neutrality guard.

### M5 — Behavioral verification + close prep (Priority Medium)

1. Off-mode behavioral repro: with `decision_gate: off` (or absent), execute a plan-phase
   pass on a throwaway SPEC under `/tmp` — assert no `decision-index.md` is created and the
   kickoff flow composes unchanged (AC-DA-019).
2. On-mode behavioral repro: with `decision_gate: on`, authoring surfaces produce the index
   with the four labels; an unverifiable anchor routes to `FOUNDER`.
3. Run the full acceptance.md sweep set; populate §E items E1-E6; hand off to plan-auditor.

Flips: AC-DA-019; final state of all sweep ACs.

## §G Anti-Patterns

- **Do not** put a `status:` field in decision-index.md (statelessness).
- **Do not** let the resolver read `recommendation_mode` or vice versa (orthogonality).
- **Do not** edit the `[HARD]` gate sentence, the three canonical options, or the label
  clause — enrich beside, never inside.
- **Do not** cite untracked files (`.moai/reports/**` working copies) as authority anchors.
- **Do not** write the product.md reconciliation in manager-spec (ownership boundary — OD-2).
- **Do not** fix the M2 mirror by copying the local file over it (`cp` is prohibited for
  mirrors per CLAUDE.local.md); apply the same change to both trees.
- **Do not** run `go test ./...` locally (lane load discipline — scoped packages only).

## §H Cross-References

- spec.md §C (requirements), §D (constraints), §E.3 (operator decisions OD-1/2/3), §F
  (exclusions)
- acceptance.md §E evidence ledger (probes P1-P19) and §D AC matrix
- design.md §2 (resolver design), §3 (index format), §4 (gate integration), §5 (verification
  design)
- SPEC-JUDGMENT-FIRST-MODE-001 spec.md §A.3 (auditor measurement), §F (carried exclusions)
- verification-completeness.md §2.1 (two-cell discipline), §1.1 (empty-sweep visibility)
