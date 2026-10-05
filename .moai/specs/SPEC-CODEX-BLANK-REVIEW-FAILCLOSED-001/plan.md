---
id: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001
title: Implementation plan — blank codex review output must not synthesize a pass
version: 0.1.0
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, audit, fail-open, gate, issue-1632, internal/cli, t551"
---

# Implementation Plan — SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001

## §A Context

Single package (`internal/cli`), single defect family (exact-equality emptiness
checks on the codex review-text path). See `spec.md` §A for the measured
evidence and the five implicated sites.

---

## §B Decisions to settle FIRST — highest change-likelihood

These two are ordered ahead of the mechanical work because they are the ones a
reviewer should push back on. Neither is settled by this plan alone.

### B.1 — The pinned synthesizer test that contradicts this card

`internal/cli/codex_review_rpc_test.go:119` pins the case `"": "pass"` inside
`TestSynthesizeReviewOutput_FindingBulletsMapToFail`
(`internal/cli/codex_review_rpc_test.go:114-126`). That nail asserts, at the
synthesizer layer, exactly the behavior this card exists to stop laundering.

It is not a stale test. It is a deliberate record of the mode-keyed default that
`spec.md` REQ-CBR-007 preserves. Two routes are available, and run-phase must
choose one explicitly rather than discover the conflict mid-edit.

**Option (a) — repair at the guard and collection sites only. RECOMMENDED.**

Change `internal/cli/mcp_codex.go:817`, `:1134`, `:1137`, `:1253` to
whitespace-aware emptiness. Leave `synthesizeReviewOutput` and
`codexUnrecognizedVerdict` untouched. The pinned test keeps passing byte-for-byte
because the synthesizer's contract does not move.

- *For*: the smallest change that closes the loophole; it repairs the layer where
  the defect actually lives (a blank body must never REACH the synthesizer), and
  it leaves REQ-CBR-007's documented policy visibly intact. The synthesizer's
  `""` behavior becomes unreachable from production through this path, which is
  the correct relationship between a guard and a default.
- *Against*: the synthesizer remains lenient if some future caller reaches it
  directly, so the guard is the only thing standing between a blank body and a
  `pass`.

**Option (b) — also harden the synthesizer, as defense in depth.**

Additionally make `synthesizeReviewOutput` return `inconclusive` for a body with
no non-whitespace character, and UPDATE the pinned case at
`internal/cli/codex_review_rpc_test.go:119` from `"": "pass"` to
`"": "inconclusive"`, with a comment recording why the record changed and that
REQ-CBR-007 still governs non-blank unrecognized bodies.

- *For*: two independent layers must both fail before a blank body becomes a
  `pass`.
- *Against*: it edits a deliberately-placed record, and it moves a documented
  contract (`internal/cli/mcp_codex.go:1395-1415` explains the mode keying) for a
  case the guard already makes unreachable. It also widens the blast radius to
  every caller of the synthesizer, which this card has not surveyed.

**Recommendation: option (a).** The defect is reachability, not synthesis: the
synthesizer's answer for an unrecognized body is a documented policy, and the bug
is that an absent body is being routed to it at all. Fix the routing.

[HARD] Under EITHER option, the pinned test is never silently deleted. Option (a)
leaves it untouched; option (b) updates it in place with a rationale comment.
Deleting it removes the record of a decision this card explicitly preserves.

### B.2 — What "blank" means, mechanically

`strings.TrimSpace(s) == ""` is the proposed discriminator, and it must be the
SAME discriminator at all four sites.

A helper (for example `codexReviewTextIsBlank(string) bool`) is **RECOMMENDED,
not mandated**, so that a later edit cannot repair three sites and miss one —
which is precisely how the current defect is shaped. The recommendation is
deliberately weak: the helper is a three-line wrapper over a stdlib one-liner,
so it earns its keep only through that miss-one argument, and a reviewer may
reasonably prefer four inline calls.

[HARD] **No acceptance criterion mandates the helper.** Implementation shape is
a run-phase choice; the ACs govern behavior only. The residual-site sweep lives
in §E below as a self-verification checklist item, not as an AC — see §E note.

Open sub-question for run-phase: whether Unicode non-breaking space (U+00A0) is
in scope. `strings.TrimSpace` cuts on `unicode.IsSpace`, which DOES include
U+00A0. This is stated so the choice is visible, not left implicit.

---

## §C Constraints

1. [HARD] Blank-output and backend-unavailable remain distinguishable by their
   Summary text (REQ-CBR-005). The codebase already treats this
   distinguishability as load-bearing — see the comment at
   `internal/cli/mcp_codex.go:790-796`: *"an inconclusive that cannot be told
   apart from 'codex is missing' is a new silence, not a repair."*
2. [HARD] The unavailable path's fail-open is unchanged (REQ-CBR-006). codex is
   optional; collapsing the two states blocks the gate for every user without
   codex installed.
3. [HARD] `codexUnrecognizedVerdict`'s mode-keyed policy for a NON-blank
   unrecognized body is unchanged (REQ-CBR-007).
4. Scope stays inside `internal/cli`. No convergence-layer edits.
5. Every AC must be mechanically checkable by a Go test in package
   `internal/cli`.

---

## §D Milestones

Ordered by decision-reversibility: the contract-shaped work first, the mechanical
sweep last.

### M1 — Settle B.1 and land the discriminator seam (Priority: High)

- Record the B.1 choice in `progress.md` §E.2 with its rationale.
- Introduce the single blankness discriminator (B.2) with its own unit test,
  including the U+00A0 decision.
- Characterization first: pin the CURRENT behavior of all four sites before
  changing any of them, so the change is visible as a diff in test expectations
  rather than as an unexplained new test.

### M2 — Repair collection + selection together (Priority: High)

- `internal/cli/mcp_codex.go:1134`, `:1137`, `:1253`.
- The selection site `:1253` is pinned by AC-CBR-003: a blank structured
  `exitedReviewMode.review` alongside a REAL `agentMessage.text` must select the
  agent text and produce `pass`. That is the case a guard-only repair would
  break.
- The collection sites `:1134` / `:1137` are pinned by AC-CBR-009 (the overwrite
  case) — NOT by AC-CBR-003, which passes even when they are left unrepaired.
  Their only behavioral effect is that a later blank item must stop clobbering an
  earlier real one in the loop at `:1125-1139`.

### M3 — Repair the guard and its summary (Priority: High)

- `internal/cli/mcp_codex.go:817` — whitespace-aware, with a Summary distinct
  from the unavailable wording (REQ-CBR-005).

### M4 — Three-state control matrix + gate visibility (Priority: High)

- The full A/B/C matrix from `acceptance.md` §A, driven through the real
  `runCodexReviewRPC` path with a stubbed conn, in the shape the two probes in
  `.moai/reports/t551/` already demonstrate.
- REQ-CBR-009: `GateUnmet` present under a `required` gate.

### M5 — Close the spec.md §E gaps (Priority: Medium)

- Measure the convergence-layer behavior that §E records as unmeasured, or
  restate it as a residual gap. Do not inherit the plan-phase reading as a
  result.

---

## §E Self-verification

- [ ] All four sites (`:817`, `:1134`, `:1137`, `:1253`) use the SAME
      discriminator, and no residual exact-equality emptiness check remains on
      the review-text path — that path being exactly those four sites, enumerated
      here rather than left to a grep to define.

      This is a checklist item, NOT an acceptance criterion. It was demoted from
      the AC set deliberately: as an AC it asserted implementation shape rather
      than behavior (violating the helper-is-recommendation decision in §B.2),
      and its absence half was vacuous-capable — a zero-match grep with no
      positive control passes whether or not the sweep found anything. If
      run-phase wants it mechanized, the grep MUST first be shown to MATCH on the
      pre-repair tree; an unmatched grep is not evidence.
- [ ] The B.1 decision is recorded with a rationale, and the pinned test at
      `internal/cli/codex_review_rpc_test.go:119` is either untouched or updated
      in place — never deleted.
- [ ] The three-state control (blank / real-clean / unavailable) is observed in
      one test, and the three produce three distinguishable signals.
- [ ] `go test ./internal/cli/...` green; `go vet ./internal/cli/...` clean.

---

## §F Risks and anti-patterns

- **AP-1 — repairing the guard alone.** Converts the blank-structured-review +
  real-agent-message case from a correct `pass` into a wrong `inconclusive`.
  M2 exists to prevent this and must land with M3.
- **AP-2 — asserting only the blank case.** An AC that observes only state (A) is
  the defect re-enacted in the test: it cannot tell "blocked everything" from
  "blocked the right thing". The control states (B) and (C) are mandatory.
- **AP-3 — collapsing blank into unavailable.** Reusing the
  `codex unavailable: ...` Summary for the blank case satisfies the verdict but
  violates constraint C.1 and produces a new silence.
- **AP-4 — vacuous green.** A test whose stub never delivers an `item/completed`
  notification at all will read as "blank" for the wrong reason. Each fixture
  must be shown to reach the site it claims to exercise — the reachability probe
  in `.moai/reports/t551/probe-reachability-20260908.txt` is the pattern.
- **AP-5 — widening to sibling axes.** Axes 3/4/5 of #1632 are other cards; axes
  1/2 are already closed (`spec.md` §D). Touching them here makes the diff
  unreviewable.

---

## §G Cross-references

- `spec.md` §A (measured evidence), §C (verdict-value argument), §D (scope)
- `acceptance.md` §A (three-state control matrix)
- `.moai/reports/t551/probe-reachability-20260908.txt`
- `.moai/reports/t551/probe-synthesizer-20260908.txt`
- GitHub issue #1632
