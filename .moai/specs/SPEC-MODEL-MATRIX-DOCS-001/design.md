# Design — SPEC-MODEL-MATRIX-DOCS-001

Design decisions for the documentation surfaces and the guard realignment. Ordered by decision-reversibility: the canonical copy and the structural contradiction fix come first; the mechanical guard edit last.

Inherited from `SPEC-MODEL-PROFILE-MATRIX-002/design.md` §E and §F, plus the documentation-relevant rejected alternatives. The matrix data model (§A) stayed with `SPEC-MODEL-MATRIX-CORE-001`; the effort-channel mechanics (§C) went to `SPEC-MODEL-MATRIX-CONFIG-001`; the wizard design (§D) to `SPEC-MODEL-MATRIX-SURFACES-001`.

---

## §A Documentation strategy

### §A.1 The naming-inversion disclosure

This is the single most important documentation output and should not be buried. It needs one canonical statement, written once and referenced from the other surfaces rather than paraphrased four times per locale:

> `max` / `medium` / `low` name the **subscription tier** whose models the profile draws on — not a performance grade. Under the v1.1 leaderboard the Max profile is both cheaper per task and higher-scoring than the Medium profile.

Placement: `advanced/profile-matrix.md` is the natural home — it is the page that renders the matrix. README and `multi-llm/model-policy.md` link to it.

Paraphrasing this four times per locale is how it drifts: sixteen independently-worded copies of a claim about an ordering are sixteen chances to restate the ordering backwards.

### §A.2 The Max-Opus rationale

Separate from §A.1 and easy to conflate with it:

> The Max profile assigns Opus to `manager-develop` and `builder-harness` deliberately, because those agents' failures are expensive to recover from. This is a quality-first choice, not a benchmark optimum — the leaderboard would favour Fable for both.

Without this, a reader who has just absorbed §A.1 will read the Max column's Opus cells as an error. The two statements are adjacent for that reason, and neither substitutes for the other.

### §A.3 Contradiction resolution order

`multi-llm/model-policy.md` and `advanced/profile-matrix.md` currently contradict each other. Rather than editing both toward each other, designate one as authoritative:

- `advanced/profile-matrix.md` — **authoritative**: renders the 33-cell matrix, the naming disclosure, the Max-Opus rationale.
- `multi-llm/model-policy.md` — **narrative**: explains how to pick a profile for your plan, links to the matrix page, carries no per-agent table of its own.

Deleting the per-agent table from the narrative page is what makes the contradiction **structurally impossible to recur**, and it is what lets the zh copy's Haiku column be removed rather than corrected. Editing two contradicting tables toward each other leaves both able to drift again on the next change; deleting one removes the degree of freedom.

### §A.4 The rules-file self-contradiction

`.claude/rules/moai/development/model-policy.md` carries both the stale "all workers Sonnet 5 fixed" tier table and the modern resolver section. The stale table also names `model_routing_profiles` as "the 3-tier config SSOT" — a block `SPEC-MODEL-MATRIX-CONFIG-001` deletes. The table is therefore not merely outdated; it will reference a nonexistent artifact.

Delete the stale section; keep and extend the resolver section. The byte-parity twin must be edited in the same commit (C-2).

Updating rather than deleting was rejected: a section whose premise has been removed cannot be made correct by rewording, and keeping it invites a future reader to restore the block it names.

### §A.5 4-locale execution

Route through the `oss-docs` harness. Its canonical-locale chain and same-PR 4-locale obligation are the existing mechanism for exactly this failure mode; hand-editing 4 locales × 6 surfaces is where parity breaks.

### §A.6 Correcting already-published figures — the split's addition

The original design assumed S0 gated M6. It did not: M6 shipped and S0 remains undischarged, so the published figures trace to an unverified reading rather than to a record.

That changes what "S0 lands" means for this SPEC. Two dispositions exist when a confirmed figure differs from a published one:

| Disposition | Outcome |
|---|---|
| Record the delta in the predecessor's `progress.md` and leave the page | The wrong number stays live. **Rejected.** |
| Correct the published page, all four locales, in one change | The record and the page agree. **Required** (REQ-MPMD-011, D-6) |

The correction is not a new milestone; it is a step appended to M6 (`plan.md` §F step 10) and a criterion evaluated after the record exists (AC-MPMD-012). Its evaluation order is load-bearing: evaluated early, against an empty delta set, it reports PASS and closes nothing.

What no design can recover is the interval itself — live documentation carried unverified figures from the landing until the correction. That is recorded as accepted history rather than designed away.

---

## §B Guard realignment

The haiku-residual rule's surface list is edited in both directions in one change:

| Action | Surface | Reason | Gated on |
|---|---|---|---|
| add | web-console agentfm model option set | the value is removed; the guard prevents its return | `SPEC-MODEL-MATRIX-SURFACES-001` |
| add | v4manifest tier-suggestion table | same | `SPEC-MODEL-MATRIX-SURFACES-001` |
| remove | `model_routing_profiles` in workflow.yaml | block deleted | `SPEC-MODEL-MATRIX-CONFIG-001` |
| remove | `validRoutingModels` in model_routing.go | map deleted | `SPEC-MODEL-MATRIX-CONFIG-001` |

The two standing surfaces (agent frontmatter plus template mirror; `claude_models` in `llm.yaml`) and all four exemptions are unchanged.

A guard that scans for a deleted artifact is not merely dead — it is **actively misleading**, because a future reader will infer the artifact still exists. That is why the removals are coupled to the additions rather than deferred.

### §B.1 Why each direction is gated on a different predecessor

Adding a surface before it is clear turns the guard into an immediate failure against shipped code, which trains a reader to treat the guard's output as noise. Removing a surface before its artifact is deleted drops coverage while the artifact is still live. The two errors are opposite in kind and both are silent until someone reintroduces `haiku`, so both predecessors' states are verified before the rule is touched.

### §B.2 The property test's assertion count is itself an assertion

A property test asserting six of the seven invariants passes exactly as a seven-invariant test does. AC-MPMD-018 therefore requires the sub-assertion count to be visible in the run output rather than inferred from a green result — the same reason AC-MPMD-016 requires the seeded-failure direction rather than only the clean-tree direction.

---

## §C Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Rename profiles to match the inverted ordering | Breaks `llm.profile` values, the CLI flag, wizard values, config files in the field, and every doc surface simultaneously. Disclosure achieves the same reader outcome at a fraction of the blast radius. |
| Defer the naming-inversion disclosure to a follow-up | Shipping a matrix whose column names assert a false ordering, with no disclosure, is the failure this SPEC family exists to prevent. |
| Paraphrase the disclosure per page and per locale | Sixteen independently-worded copies of an ordering claim are sixteen chances to state it backwards (§A.1). |
| Edit the two contradicting pages toward each other | Leaves both able to drift apart again; deleting one table removes the degree of freedom (§A.3). |
| Update the stale rules-file tier section instead of deleting it | Its premise — `model_routing_profiles` as the config SSOT — is being deleted; rewording cannot make it correct (§A.4). |
| Correct the zh Haiku column rather than deleting the table it sits in | Corrects one locale's symptom and leaves the structural contradiction intact (§A.3). |
| Record an S0 delta upstream and leave the published page | Leaves the wrong number live; the correction is the whole point of discharging S0 late (§A.6). |
| Defer the retired guard-surface removals | A guard scanning a deleted artifact implies to a future reader that the artifact still exists (§B). |
| Regenerate the tokenomics infographics as part of M6 | Explicitly out of scope; M6 records the assessment result, and a regeneration inside a documentation milestone would expand it without a bound. |
