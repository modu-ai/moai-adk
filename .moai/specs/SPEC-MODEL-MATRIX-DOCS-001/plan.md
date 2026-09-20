# Implementation Plan — SPEC-MODEL-MATRIX-DOCS-001

Tier L. Two milestones: M6 (4-locale documentation) and M7 (guard realignment and full verification). M6 has already landed under the original SPEC ID; M7 has not started.

---

## §A Context

This SPEC carries the fourth and last seam of the `SPEC-MODEL-PROFILE-MATRIX-002` split: every user-readable surface the matrix touches, and the guard realignment that follows the other three successors' deletions and clearances.

Scope surfaces: thirteen docs-site pages across four locales, four README copies, `.claude/rules/moai/development/model-policy.md` and its byte-parity template twin, `agent-authoring.md`, `dynamic-workflows.md`, `internal/spec/lint_haiku_residual.go`, and the consolidated matrix property test.

**Inherited state**: M6 landed in squash `31da99a7b` (PR #1163) — while its blocking precondition S0, owned by `SPEC-MODEL-MATRIX-CORE-001`, was never discharged. M7 never started.

---

## §B Review first — the decisions most likely to change

### §B.1 The naming-inversion disclosure and the Max-Opus rationale (M6)

These are the user-facing outputs the whole SPEC family exists to produce. If they ship without them, a reader encounters column names asserting an ordering the data contradicts, and a Max column whose Opus cells look like mistakes.

One canonical statement each, written once and referenced from the other surfaces rather than paraphrased four times per locale:

> `max` / `medium` / `low` name the **subscription tier** whose models the profile draws on — not a performance grade. Under the v1.1 leaderboard the Max profile is both cheaper per task and higher-scoring than the Medium profile.

> The Max profile assigns Opus to `manager-develop` and `builder-harness` deliberately, because those agents' failures are expensive to recover from. This is a quality-first choice, not a benchmark optimum — the leaderboard would favour Fable for both.

Placement: `advanced/profile-matrix.md` is the natural home (it is the page that renders the matrix). README and `multi-llm/model-policy.md` link to it. Review this as **copy**, not as a checklist item.

### §B.2 What to do about figures that are already live (M6)

This is the decision the split surfaced and the original SPEC never had to make. S0 was designed as a gate ahead of M6, and M6 shipped anyway; the figures now published trace to an unverified reading.

So when S0 finally lands a confirmed value that differs, there are two possible dispositions and only one is acceptable: record the delta upstream and leave the page (**rejected** — leaves the wrong number live), or correct the published page (**required**, REQ-MPMD-011 + D-6). Decide the mechanics of that correction — whether it rides a separate commit, and how the four locales stay in step — before S0 lands rather than under time pressure after.

### §B.3 Contradiction resolution by structure rather than by edit (M6)

`multi-llm/model-policy.md` and `advanced/profile-matrix.md` currently contradict each other. Rather than editing both toward each other, designate one as authoritative:

- `advanced/profile-matrix.md` — **authoritative**: renders the 33-cell matrix, the naming disclosure, the Max-Opus rationale.
- `multi-llm/model-policy.md` — **narrative**: explains how to pick a profile for your plan, links to the matrix page, carries no per-agent table of its own.

Deleting the per-agent table from the narrative page is what makes the contradiction structurally impossible to recur, and it is what lets the zh copy's Haiku column be removed rather than corrected. Editing the two pages toward each other would leave them able to drift apart again on the next change.

### §B.4 The guard's two-directional edit and its ordering (M7)

| Action | Surface | Reason | Gated on |
|---|---|---|---|
| add | web-console agentfm model option set | the value is removed; the guard prevents its return | `SPEC-MODEL-MATRIX-SURFACES-001` |
| add | v4manifest tier-suggestion table | same | `SPEC-MODEL-MATRIX-SURFACES-001` |
| remove | `model_routing_profiles` in workflow.yaml | block deleted | `SPEC-MODEL-MATRIX-CONFIG-001` |
| remove | `validRoutingModels` in model_routing.go | map deleted | `SPEC-MODEL-MATRIX-CONFIG-001` |

Both directions are gated on a different predecessor, and getting either order wrong is costly in a different way: adding early turns the guard into an immediate failure against shipped code; removing early drops coverage that is still needed. Verify both predecessors' state before touching the rule.

The two standing surfaces (agent frontmatter, `claude_models` in `llm.yaml`) and all four exemptions are unchanged.

---

## §C Known issues carried into execution

| # | Issue | Where it bites |
|---|---|---|
| K-1 | `README.md`'s benchmark columns `$/solved` and `Tokens/solved` are **derived** metrics, not raw leaderboard columns. Replacing one set of figures with another without pinning column semantics produces a table that is internally consistent but mislabelled. | M6 — depends on S0's semantics pin |
| K-2 | The currently-shipped README figures are the `[max]`-effort rows, while the SPEC framing depends on `[low]` / `[high]` rows. | M6 |
| K-3 | The haiku-residual rule's surfaces 3 and 4 target artifacts `SPEC-MODEL-MATRIX-CONFIG-001` deletes. | M7 — ordering (§B.4) |
| K-4 | The stale rules-file section names `model_routing_profiles` as the config SSOT — a block that will not exist. | M6 |
| K-5 | The ja and ko copies of `multi-llm/model-policy.md` were never inspected; only zh was. The claim that ja diverged into a third shape is a carried-forward unverified input. | M6 — re-measure |
| K-6 | The `advanced/profile-matrix.md` and `advanced/no-haiku-3tier.md` line ranges cited in the original brief were never opened during authoring. | M6 — re-measure |
| K-7 | Whether the tokenomics infographics actually embed benchmark numbers was never inspected. | M6 — assess, do not regenerate |

---

## §D Pre-flight

1. Confirm `SPEC-MODEL-MATRIX-CORE-001`'s S0 record exists before writing or correcting any benchmark figure (C-5).
2. Confirm `SPEC-MODEL-MATRIX-CONFIG-001` has deleted both artifacts before removing their guard surfaces (§B.4).
3. Confirm `SPEC-MODEL-MATRIX-SURFACES-001` has cleared both surfaces before adding them to the guard (§B.4).
4. Re-measure the four carried-forward unverified inputs (K-5, K-6, K-7, plus cross-platform build status) rather than inheriting them.
5. `git fetch origin` and check divergence — a parallel session may be active on this shared checkout.

---

## §E Constraints carried into execution

4-locale parity as a HARD obligation, rule-file byte-parity twins edited in one commit, template content neutrality, No-Haiku as a HARD non-skippable gate, no benchmark figure ahead of the S0 record, and explicit pathspecs on every commit.

---

## §F Milestones

### M6 — 4-locale documentation (LANDED, with an undischarged precondition)

**Priority: Medium. BLOCKED on the predecessor's S0. Depends on all three other successors for accuracy.**

Route through the `oss-docs` harness (D-5).

1. Replace the README benchmark table in all four locales with S0-confirmed figures, labelling the effort level of each row. — *landed, on unverified figures*
2. Rewrite `advanced/profile-matrix.md` in all four locales: 33-cell per-agent matrix, group table removed, plus the naming-inversion disclosure and the Max-Opus rationale (§B.1). — *landed*
3. Replace the `advanced/no-haiku-3tier.md` benchmark table in all four locales. — *landed, on unverified figures*
4. Rewrite `multi-llm/model-policy.md` in all four locales as the narrative page: remove the retired "all workers Sonnet 5" claim, remove the per-agent table entirely (which removes the zh Haiku column and normalizes the ja divergence in one move), link to the matrix page. — *landed*
5. Delete the stale tier-table section from `.claude/rules/moai/development/model-policy.md`; edit its byte-parity template twin in the same commit. — *landed*
6. Add `fable` to the model enum in `agent-authoring.md` and `dynamic-workflows.md`. — *landed*
7. Replace every `--model-policy` flag reference with `--profile`. — *landed*
8. Assess the tokenomics infographics for embedded stale figures and record the disposition decision (regeneration itself is out of scope). — *status unverified; K-7*
9. Verify 4-locale parity: same section count, same table shape, no locale missing an edit.
10. **Correction step, new to this SPEC**: when the predecessor's S0 record lands, compare every published figure against it and correct the published pages where they differ (REQ-MPMD-011, §B.2). All four locales in one change.

Covers REQ-MPMD-001 … 015.

### M7 — Guard realignment and full verification (NOT STARTED)

**Priority: Medium. Depends on `SPEC-MODEL-MATRIX-CONFIG-001` and `SPEC-MODEL-MATRIX-SURFACES-001` — in both directions (§B.4).**

1. Add the web-console agentfm model option set and the v4manifest tier-suggestion table to the haiku-residual rule's surfaces — after verifying both are clear.
2. Remove the `model_routing_profiles` and `validRoutingModels` surfaces from the same rule — after verifying both artifacts are gone (K-3).
3. Add the matrix property test asserting all seven invariants.
4. `go vet ./... && golangci-lint run && go test ./...`.
5. Cross-platform build for the release targets.

Covers REQ-MPMD-016 … 021.

---

## §G Anti-patterns to avoid

| # | Anti-pattern | Why it bites here |
|---|---|---|
| AP-1 | Recording an S0 delta upstream and leaving the published page | Leaves the wrong number live; the correction is the point (§B.2) |
| AP-2 | Copying benchmark numbers from the plan-phase probe table | That table is a summarised probe result explicitly marked insufficient for documentation |
| AP-3 | Replacing figures without pinning column semantics | `$/solved` and `Tokens/solved` are derived; a swap without semantics produces a consistent-but-mislabelled table (K-1) |
| AP-4 | Editing one locale and deferring the other three | 4-locale parity is a HARD obligation; deferral is how parity breaks |
| AP-5 | Editing the two contradicting pages toward each other | Leaves them able to drift apart again; delete one table instead (§B.3) |
| AP-6 | Updating the stale rules-file tier section instead of deleting it | Its premise — `model_routing_profiles` as the config SSOT — is gone (K-4) |
| AP-7 | Adding the guard surfaces before SURFACES clears them | Turns the guard into an immediate failure against shipped code (§B.4) |
| AP-8 | Removing the retired guard surfaces before CONFIG deletes the artifacts | Drops coverage that is still needed (§B.4) |
| AP-9 | Leaving the retired haiku-rule surfaces in place "harmlessly" | A guard scanning a deleted artifact implies to future readers that the artifact still exists |
| AP-10 | Inheriting the ja/ko shape, the `advanced/*` line ranges, or the infographic contents | All four were explicitly never measured; re-measure (K-5, K-6, K-7) |
| AP-11 | Regenerating the tokenomics infographics | Assessment is in scope; regeneration is explicitly deferred |
| AP-12 | `git add -A` on this shared checkout | A parallel session may have staged unrelated work; use explicit pathspecs |

---

## §H Cross-References

- `spec.md` — requirements, the disclosure statements, decisions, risks
- `acceptance.md` — AC matrix and verification commands
- `design.md` — documentation strategy, contradiction resolution order, guard realignment, rejected alternatives
- `research.md` — observed documentation contradictions, guard surfaces, open clarification, unverified items
- `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/` — predecessor; owns S0
- `.moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/` — predecessor; deletes the surfaces M7 removes
- `.moai/specs/SPEC-MODEL-MATRIX-SURFACES-001/` — predecessor; clears the surfaces M7 adds
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — retired predecessor stub
