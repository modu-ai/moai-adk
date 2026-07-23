# Implementation Plan — SPEC-MODEL-PROFILE-MATRIX-002

Tier L. Eight units of work: one blocking precondition (S0) and seven milestones (M1-M7).

---

## §A Context

`SPEC-MODEL-PROFILE-MATRIX-001` shipped a 3-column profile axis resolved through six agent groups. This SPEC replaces the group indirection with a direct 33-cell `profile → agent → {model, effort}` matrix, revives frontmatter `effort:` rewriting in a narrowed form so the effort axis actually reaches the `Agent`-tool spawn path, retires the unused 36-cell Tier×Phase axis and the `llm.yaml profiles:` mirror, and corrects the documentation across four locales.

Scope surfaces: `internal/template`, `internal/config`, `internal/cli`, `internal/web`, `internal/harness/v4manifest`, `internal/spec`, `docs-site`, `README*.md`, `.claude/rules/moai/development/model-policy.md` + template twin.

---

## §B Review first — the decisions most likely to change

These four decisions carry the highest change-likelihood and the highest cost of being wrong. They should absorb review attention before the mechanical milestones.

### §B.1 The 33 cell values themselves (M1)

The matrix is settled design input (`spec.md` §A.4) and must be transcribed verbatim, but it is also the artifact most likely to be revised after the S0 numbers land. Per `research.md` §B.1 the incoming values are **not** a reshape of the current 18 cells — they are new values. `max/develop` inverts from `fable/low` to `opus/xhigh`; `max/spec_auditors` drops from `fable/medium` to `fable/low`. A reviewer should confirm each of the 33 cells against `spec.md` §A.4 rather than diffing against today's matrix, because a diff-based review will mistake intentional inversions for errors.

### §B.2 Frontmatter rewrite revival and its four mitigations (M3)

The predecessor was retired for real side effects. The revival's safety rests entirely on the four mitigations holding **simultaneously**: `effort:`-only, deployed-tree-only, post-deploy, override-excluded. Dropping any one reproduces a variant of the original failure. The `effort:`-only mitigation in particular must be implemented as "pass leave-unchanged for model", never as a read-modify-write (`design.md` §C.2) — a read-modify-write is observationally identical in the happy path and silently re-pins the model when the read is wrong.

### §B.3 The `moai model profile --json` shape change (M1 + M3)

Removing groups strands the `Group` key. M3 then elevates this JSON to a consumed contract (the Workflow-path lookup route). Deciding the shape once, in M1, avoids two breaking changes. `research.md` §F carries this as an open clarification; it should be resolved before M1 lands, not after.

### §B.4 The naming-inversion disclosure and the Max-Opus rationale (M6)

These are the user-facing outputs the SPEC exists to produce. If M6 ships the matrix without them, a reader encounters column names asserting an ordering the data contradicts, and a Max column whose Opus cells look like mistakes. `design.md` §E.1 and §E.2 carry the canonical wording; it should be reviewed as copy, not as a checklist item.

---

## §C Known issues discovered during planning

| # | Issue | Where it bites |
|---|---|---|
| K-1 | The template agent frontmatter's current effort values differ from the incoming Medium column in four of ten files. REQ-MPM2-112's "no-op at medium" holds only after the template baseline is re-set. | M1/M3 ordering — `design.md` §C.5 |
| K-2 | `TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` asserts precisely the behavior M2 removes. It must be deleted, not amended. | M2 |
| K-3 | The comment block at `internal/web/agentfm.go` `applyPerfTierEdits` states frontmatter re-application is retired. M3 makes that comment false. | M3 |
| K-4 | The haiku-residual rule's surfaces 3 and 4 target artifacts M2 deletes. | M2 → M7 coupling |
| K-5 | The two `llm.yaml` copies have already drifted in the `glm.models` block (local has `opus`/`sonnet`/`haiku` alias keys the template lacks). A whole-block sync during M2 would import `haiku` into the template. M2 must edit the `profiles:` block only. | M2 |
| K-6 | `DefaultProfileMatrix()`'s Go type is unchanged while its inner-key semantics change from group key to agent name — a semantic change with no compiler signal. | M1 |
| K-7 | The wizard's option values (`high`/`medium`/`low`) do not match the profile vocabulary (`max`/`medium`/`low`); a user selecting "High" gets `profile: max`. | M4 |
| K-8 | `README.md`'s benchmark columns `$/solved` and `Tokens/solved` are derived metrics, not raw leaderboard columns. S0 must pin column semantics, not only values. | S0 → M6 |

---

## §D Pre-flight

Before starting any milestone:

1. `git fetch origin main` and check divergence — a parallel session may be active on this shared checkout (C-8).
2. Confirm the four open clarifications in `research.md` §F are resolved (they gate M1, M2, M4, and M6 respectively).
3. Confirm S0 is discharged before any M6 work begins.
4. `go build ./... && go test ./internal/template/... ./internal/config/...` to establish a green baseline.

---

## §E Constraints carried into execution

Reproduced from `spec.md` §D for execution reference — Template-First (`make build` after every template edit), template content neutrality (no SPEC IDs or REQ tokens under `internal/template/templates/**`), rule-file byte-parity twins edited in one commit, 4-locale parity mandatory, No-Haiku is a HARD non-skippable gate, `inherit` is never passed as an `Agent`-tool `model` argument, and every commit uses explicit pathspecs.

---

## §F Milestones

Milestones are listed in **execution order** because they carry hard dependencies. Their review priority is the reverse-reversibility ranking in §B.

### S0 — Leaderboard verification (BLOCKING)

**Priority: Critical. Blocks M6 entirely. Does not block M1-M5 or M7.**

Confirm the DeepSWE v1.1 figures against the live source and produce the verification record specified in `research.md` §A.5.

Steps:

1. Read the leaderboard for Fable 5, Opus 4.8, and GLM 5.2, capturing **the effort level of each row read**.
2. Resolve the GLM 5.2 Pass@1 conflict to a canonical value; state whether the source publishes a point estimate, an interval, or both.
3. Pin the token-column semantics (raw output tokens vs tokens-per-solved) so the README's derived columns can be reconstructed correctly.
4. Record the leaderboard version string and update date.
5. Produce a delta table against `spec.md` §A.2 for every metric that differs; write it to `progress.md`.

Do **not** proceed to M6 without this record. Do **not** copy figures from `research.md` §A.2 — that table is a summarised probe result explicitly marked insufficient.

Covers REQ-MPM2-001 … 004.

### M1 — 33-cell matrix redesign

**Priority: High. Blocks M2, M3, M7.**

1. Replace `defaultProfileMatrix` with the direct `profile → agent → {model, effort}` map, transcribing all 33 cells verbatim from `spec.md` §A.4.
2. Delete the six group constants and `agentGroupMembership`.
3. Delete `AgentGroup`; update `internal/cli/model.go` `resolveModelProfileReport` accordingly (per the §B.3 decision).
4. Add the explicit `Explore` row; keep the unmapped-agent `inherit` fallback.
5. Rename the resolver's second return value from `hasGroup` to a name describing what it now means (`injectable` / `mapped`) and update every call site.
6. Update `DefaultProfileMatrix()`'s doc comment to state the inner key is an agent name (K-6), and update the `@MX:ANCHOR` reason lines on the matrix and resolver.
7. Re-set the **template** agent frontmatter `effort:` values to the new Medium column (K-1), then `make build`.
8. Amend tests: rewrite the fidelity and low-column expectations; retarget the override-precedence test's second assertion; **split** the inherit test into an `Explore` case and an unmapped-agent case.
9. Add the display-order/matrix-key agreement assertion (`design.md` §A.3).

Covers REQ-MPM2-010 … 023.

### M2 — Retire the 36-cell axis and the config mirror

**Priority: High. Depends on M1. Blocks M7.**

1. Remove the `model_routing_profiles` block from `.moai/config/sections/workflow.yaml` and its template copy, including the explanatory comment lines that reference it.
2. Delete `RouteModelFor`, `ModelRoutingProfiles`, the `validRoutingModels` map, and the four `model_routing_profiles.*` validator branches.
3. Delete the orphaned tests in `internal/config/model_routing_test.go` (REQ-MPM2-032).
4. Remove the `profiles:` block from both `llm.yaml` copies — **the `profiles:` block only** (K-5). Remove the ~20 lines of group-explaining comment in the template copy at the same time.
5. Make the config loader tolerate a legacy `profiles:` or `model_routing_profiles` block as inert rather than failing (REQ-MPM2-037).
6. Implement `moai update` detection of a non-empty `profiles:` block, with migration into `agent_overrides` or a warning naming the affected cells (REQ-MPM2-035/036).
7. Delete `TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` (K-2).
8. Verify the matrix literal now exists in exactly two places.

Covers REQ-MPM2-030 … 037.

### M3 — Effort actualization (both channels)

**Priority: High. Depends on M1. Highest review priority after the matrix values.**

1. Implement the agent-effort application function per `design.md` §C.2: enumerate deployed agent files, skip override-carrying agents, skip unmapped agents, skip absent files, write `effort:` only.
2. Ensure the frontmatter patcher can express "leave `model:` unchanged" — extend it if the current signature cannot (§B.2 — read-modify-write is prohibited).
3. Wire the four seams: `internal/cli/init.go`, both `internal/cli/update.go` paths, `internal/web/agentfm.go`. Verify each is **after** any template deploy at that seam.
4. Adopt the continue-on-error posture: a single agent's rewrite failure logs and continues.
5. Rewrite the stale comment block at `applyPerfTierEdits` (K-3).
6. Expose the Workflow-path lookup route — document `moai model profile --json` as the contract, adding an agent-name filter flag if the full array is impractical for a script.
7. Add the two-channel explanation to documentation, including the explicit note that SPEC-001 DECISION-001's wording is superseded and that `Explore` has no file.
8. Tests: override-exclusion, template-immutability (assert no write under `internal/template/templates/`), model-line-preservation, absent-file skip, post-deploy ordering.

Covers REQ-MPM2-040 … 051.

### M4 — Init wizard question

**Priority: Medium. Independent of M1-M3; user-facing.**

1. Rewrite the `model_policy` question to subscription-tier framing; remove the `haiku` reference.
2. Change the option values to `max`/`medium`/`low` per `design.md` §D, keeping `NormalizeToTier` as a tolerant reader for a legacy `high`.
3. Write option descriptions that state the target subscription tier and do **not** assert a performance ordering (`design.md` §D.1).
4. Add `ko`, `ja`, and `zh` translations for the title, description, and every option label and description.
5. Test: the question renders localized text under each of the three non-English locales, and the persisted `llm.profile` is one of `max`/`medium`/`low` for every option.

Covers REQ-MPM2-060 … 065.

### M5 — Web console cleanup

**Priority: Medium. Independent of M1-M4.**

1. Remove `haiku` from the agentfm model selector option set.
2. Change the v4manifest lightblue tier suggestion to `sonnet / low`.
3. Re-word `agentfm.tier.desc` in all four locales to describe the effort-reapplication behavior M3 restores.
4. Remove the orphaned `mp.*` key family from all four locale files — re-grep for consumers first (`research.md` §G lists this as an unverified carried-forward input).
5. Verify no i18n key exists in a subset of locales after the edits.

Covers REQ-MPM2-070 … 074.

### M6 — 4-locale documentation

**Priority: Medium. BLOCKED on S0. Depends on M1, M2, M3 for accuracy.**

Route through the `oss-docs` harness (`design.md` §E.5).

1. Replace the README benchmark table in all four locales with S0-confirmed figures, labelling the effort level of each row.
2. Rewrite `advanced/profile-matrix.md` in all four locales: 33-cell per-agent matrix, group table removed, plus the naming-inversion disclosure (`design.md` §E.1) and the Max-Opus rationale (§E.2).
3. Replace the `advanced/no-haiku-3tier.md` benchmark table in all four locales.
4. Rewrite `multi-llm/model-policy.md` in all four locales as the narrative page: remove the retired "all workers Sonnet 5" claim, remove the per-agent table entirely (which removes the zh Haiku column and normalizes the ja divergence in one move), link to the matrix page.
5. Delete the stale tier-table section from `.claude/rules/moai/development/model-policy.md`; edit its byte-parity template twin in the same commit.
6. Add `fable` to the model enum in `agent-authoring.md` and `dynamic-workflows.md`.
7. Replace every `--model-policy` flag reference with `--profile`.
8. Assess `assets/images/readme/tokenomics-harness-{en,ko,ja,zh}.png` for embedded stale figures and record the disposition decision (regeneration itself is out of scope per `spec.md` §C).
9. Verify 4-locale parity: same section count, same table shape, no locale missing an edit.

Covers REQ-MPM2-080 … 092.

### M7 — Guard realignment and full verification

**Priority: Medium. Depends on M1, M2, M5.**

1. Add the web-console agentfm model option set and the v4manifest tier-suggestion table to the haiku-residual rule's surfaces.
2. Remove the `model_routing_profiles` and `validRoutingModels` surfaces from the same rule (K-4).
3. Add the matrix property test asserting all seven invariants of `design.md` §A.3.
4. `go vet ./... && golangci-lint run && go test ./...`.
5. Cross-platform build for the release targets.

Covers REQ-MPM2-100 … 105.

### Cross-cutting

REQ-MPM2-110 … 113 attach to M6 (the Fable-unavailability fallback and the GLM/CG pairing note are documentation outputs), M2 (the no-silent-drop prohibition), and M3 (the dev-repo self-application no-op, which becomes true only after M1 step 7).

---

## §G Anti-patterns to avoid

| # | Anti-pattern | Why it bites here |
|---|---|---|
| AP-1 | Diffing the new matrix against the current 18 cells to "verify" the transcription | The new values are not a reshape; a diff review mistakes intentional inversions for errors (§B.1) |
| AP-2 | Read-modify-write on agent frontmatter to preserve `model:` | Observationally identical in the happy path; silently re-pins the model when the read is wrong |
| AP-3 | Whole-block sync of the two `llm.yaml` copies | Imports the local-only `glm.models` alias keys — including `haiku` — into the template (K-5) |
| AP-4 | Iterating matrix keys when applying effort | Breaks on `Explore`, which is mapped but has no file |
| AP-5 | Applying effort before template deploy | The deploy overwrites the effort lines; the user's profile silently reverts each update |
| AP-6 | Deleting `agentfm.tier.desc` instead of re-wording it | M3 makes the described behavior true again; deletion loses a now-accurate string |
| AP-7 | Copying benchmark numbers from `research.md` §A.2 | That table is a summarised probe result explicitly marked insufficient for documentation |
| AP-8 | Editing one locale and deferring the other three | 4-locale parity is a HARD obligation; deferral is how parity breaks |
| AP-9 | Leaving the retired haiku-rule surfaces in place "harmlessly" | A guard scanning a deleted artifact implies to future readers that the artifact still exists |
| AP-10 | Treating REQ-MPM2-112 ("no-op at medium") as already true | It becomes true only after the template baseline is re-set (K-1) |
| AP-11 | `git add -A` on this shared checkout | A parallel session may have staged unrelated work; use explicit pathspecs |

---

## §H Cross-References

- `spec.md` — requirements, the 33-cell matrix, decisions, risks
- `acceptance.md` — AC matrix and verification commands
- `design.md` — data-model rationale, two-channel mechanics, rejected alternatives
- `research.md` — observed current-state facts, S0 probe status, open clarifications, unverified items
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-001/` — predecessor
