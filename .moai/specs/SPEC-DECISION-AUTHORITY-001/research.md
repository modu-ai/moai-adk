---
id: SPEC-DECISION-AUTHORITY-001
title: "Research — surfaces, precedent, and authority-register measurement"
version: "0.1.0"
created: 2026-09-13
---

# research.md — SPEC-DECISION-AUTHORITY-001

All measurements on tree `62fbd6baf` (worktree `.claude/worktrees/t692`, branch
`WT-judgment-authority`), 2026-09-13.

## 1. Surfaces read (all read in full or targeted this run)

| Surface | Read | Relevance |
|---|---|---|
| `.moai/reports/t692/issue-1683-proposal.md` | Full (276 lines) | The proposal: authority chain, four labels, verdict actions, zero-flag rule, compatibility note |
| `.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/spec.md` | Full (477 lines) | The split (§A.1), the three adopted conditions, REQ-JFM structure, §F exclusions that bind this SPEC |
| `.claude/agents/moai/manager-spec.md` | Full (242 lines) | The flow owner: Steps 1-6, artifact ownership, forbidden modifications |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | Lines 1-299 | Phase 9/10/11; the Implementation Kickoff Approval [HARD] clause at `:217-223`; plan HTML emission (`:213`); the FAIL retry loop |
| `.claude/skills/moai/workflows/plan/clarity-interview.md` | Full (239 lines) | Phase 4 interview loop — where unresolved decisions surface during clarification; Phase DP1 Plan Review gate |
| `.moai/config/sections/interview.yaml` | Full (39 lines) | Carries `recommendation_mode: pull` (dogfood) + plan/project round config |
| `.moai/project/{product,structure,tech}.md` | Heads (60/40/40 lines) | D4 authority-register verification |
| `.claude/rules/moai/development/spec-frontmatter-schema.md` | Full | Artifact-statelessness rule (decision-index.md MUST NOT carry `status:`) |
| `.claude/rules/moai/development/verification-completeness.md` | Full | Two-cell AC discipline (§2, §2.1), empty-sweep visibility (§1.1) |

## 2. Precedent: the config axis shape (CONST-2)

The `recommendation_mode` axis landed in SPEC-JUDGMENT-FIRST-MODE-001 with this exact Go
wiring, measured this tree:

- `internal/config/types.go:1360` — `RecommendationMode string` with `yaml:"recommendation_mode"`.
- `internal/config/defaults.go:1163` — `RecommendationMode: "push"` (distributed default).
- `ResolvedRecommendationMode()` (`types.go:1369`) — absent/empty/unrecognized resolves to the
  default; the raw value is preserved on the struct so an unrecognized setting is observable.
- `internal/config/interview_recommendation_mode_test.go` — round-trip tests: default-is-push,
  absent-key, pull round-trip, explicit push.

The new axis is the same shape: `DecisionGate string` + `yaml:"decision_gate"`, default
`"off"`, `ResolvedDecisionGate()`, same test file family. **Orthogonality note:** the two
resolvers must not read each other — each resolves from its own field only (REQ-DA-018).

## 3. Positive-control incident (documented for the AC evidence ledger)

The first positive-control probe — `grep -c "recommendation_mode" internal/config/defaults.go`
— returned `0` (exit 1) although the key demonstrably landed in SPEC-JFM. Root cause: the YAML
key lives in `types.go` as a struct tag; `defaults.go` carries only the CamelCase Go field
name. A zero-hit probe against the wrong file is not evidence of absence. The corrected
positive control (`types.go`) returns `1` (exit 0). Recorded because the same mistake in the
release-blocking AC-DA-001 would have manufactured a false RED.

## 4. Mirror-state measurement (CONST-8 basis)

| File | Local vs template mirror |
|---|---|
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | byte-identical (`diff -q` clean) |
| `.claude/rules/moai/core/askuser-protocol.md` | byte-identical |
| `.claude/agents/moai/manager-spec.md` | **differs** (intentional C1↔C2 branching) |
| `.claude/skills/moai/workflows/plan/clarity-interview.md` | **differs** |
| `.moai/config/sections/interview.yaml` | **differs** (local dogfood `pull` vs template default `push`) |

Consequence: the mirror AC (AC-DA-018) asserts token presence in BOTH trees per edited file,
never `diff -q` identity.

## 5. The kickoff gate — exact anchor

`.claude/skills/moai/workflows/plan/spec-assembly.md:217`:

> `[HARD] The Implementation Kickoff Approval AskUserQuestion gate stays MANDATORY and
> score-independent. The plan HTML report ENRICHES the review surface (inline prose → rich
> HTML); it does NOT replace the gate, does NOT auto-bypass it, and does NOT relax its three
> canonical options ... or the (권장) first-option label (withheld under
> `recommendation_mode: pull`; the gate itself is unchanged).`

The precedent for "enrichment" is already in this clause: the plan HTML report (Step 2.3.3a)
enriches the review surface without touching the gate. The decision-index presentation joins
that enrichment family — additive, fail-open, never a gate modification. Line number is a
locating aid at `62fbd6baf`; locate by the quoted clause text, not arithmetic.

## 6. Where unresolved decisions surface today (the gap the index fills)

- **Phase 4 clarity interview** (`clarity-interview.md`): rounds on Scope / Constraints /
  Success / Edge cases / Priority. User answers are recorded to `interview.md`. A question the
  user does NOT answer (or answers "whatever you think") has no sink — it is resolved
  implicitly by the author downstream. This is failure #1's entry point.
- **Phase 8 SPEC planning** (`clarity-interview.md`): manager-spec drafts GEARS requirements;
  nothing in the step list asks "whose decision was this?".
- **Phase DP1 Plan Review** (`clarity-interview.md:188-215`): the user approves the SPEC as a
  whole; unresolved decisions are indistinguishable from resolved ones inside the artifact.
- **Implementation Kickoff Approval** (`spec-assembly.md:217`): approves entry into run — the
  only human gate at the decision point, and today it sees no decision inventory.

## 7. `.moai/reports/**` gitignore measurement (corrects the D4 premise)

- `git check-ignore -v .moai/reports/t692/issue-1683-proposal.md` → exit **1** (not ignored).
- Local `.gitignore:355` ignores `.moai/reports/*.md` (root level only); `:273` ignores
  `.moai/reports/plan-audit/*.md`; `:107-109` comment: card evidence under `.moai/reports/**`
  is "durable audit evidence, not operational".
- The proposal file is untracked (`??` in git status).

Conclusion (carried into REQ-DA-010): the citable-authority criterion is **committed**, not
"not-gitignored". `product.md`, completed SPEC HISTORY rows, config sections, and the
constitution are all committed and re-derivable with `git show`; the proposal currently is not.

## 8. Risks identified

- **R1 — Doctrine-token decay.** All coordinates in this SPEC are locating aids at
  `62fbd6baf`. M1-M4 inserts renumber lines; every AC anchor is stated as content to grep, not
  a line number.
- **R2 — Off-mode regression surface is wide.** Three doctrine files + one config file +
  agent body change; the regression property (AC-DA-004) must sweep all six trees (3 surfaces
  × local + template).
- **R3 — Label drift.** "DECIDED" is a common English word; the vocabulary AC must anchor on
  the distinctive token `POLICY-COVERED` (zero hits at base across all surfaces) rather than
  `DECIDED`.
- **R4 — product.md ownership crossing.** manager-spec writing `.moai/project/product.md`
  would cross the canonical ownership boundary (manager-docs owns project-doc scaffolding).
  REQ-DA-016 routes the act to the operator's decision instead of encoding an owner.
