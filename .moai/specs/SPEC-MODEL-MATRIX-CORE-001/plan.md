# Implementation Plan — SPEC-MODEL-MATRIX-CORE-001

Tier L. Two units of work: one blocking precondition (S0) and one milestone (M1). M1 has already landed under the original SPEC ID; S0 has not been discharged.

---

## §A Context

This SPEC carries the first seam of the `SPEC-MODEL-PROFILE-MATRIX-002` split: the 33-cell agent-direct matrix that replaces the six-group indirection, and the leaderboard verification record every downstream benchmark figure must trace to.

Scope surfaces: `internal/template/profile_matrix.go` and its tests, `internal/cli/model.go` (`resolveModelProfileReport`), the `llm.yaml` and `workflow.yaml` mirror pair, ten agent frontmatter files across both mirrors, and `progress.md` as the S0 record's home.

**Inherited state**: M1 landed in squash `31da99a7b` (PR #1163); S0 did not. Two M1 plan steps below (steps 2 and 3) were never executed despite the landing record — card t1037 owns that disposition, and the steps stay listed here so the remainder is not lost.

---

## §B Review first — the decisions most likely to change

### §B.1 The 33 cell values themselves (M1)

The matrix is settled design input (`spec.md` §A.3) and must be transcribed verbatim, but it is the artifact most likely to be revised after the S0 numbers land. Per `research.md` §B.1 the incoming values are **not** a reshape of the previous 18 cells — they are new values. `max/develop` inverts from `fable/low` to `opus/xhigh`; `max/spec_auditors` drops from `fable/medium` to `fable/low`. Confirm each of the 33 cells against `spec.md` §A.3 rather than diffing against the old matrix: a diff-based review mistakes intentional inversions for errors.

### §B.2 The `moai model profile --json` shape change (M1)

Removing groups strands the `Group` key of `modelProfileEntry`. `SPEC-MODEL-MATRIX-CONFIG-001` then elevates this JSON to a consumed contract (the Workflow-path lookup route). Deciding the shape once, here, avoids two breaking changes. `research.md` §F carries this as an open clarification; it must be resolved before M1 is considered complete, not after — and CONFIG's M3 cannot start against an undecided shape.

### §B.3 Whether the landed matrix actually matches §A.3

`progress.md` §E.2 records the matrix as landed with a distribution — Opus 25 cells / Sonnet 8 / Fable 0 / `xhigh` 0 / `max` in 2 cells — that does **not** match `spec.md` §A.3, which carries six Fable cells and two `xhigh` cells and no `max` effort at all. This divergence is inherited verbatim and unreconciled. Reviewing it is the highest-value read in this SPEC: either the landed matrix departed from the authoritative table, or the landing record describes a different matrix. Resolve before treating M1 as discharged.

### §B.4 The S0 record's status as remediation

S0 blocks documentation that has already shipped. The record is therefore not a gate ahead of the work but a correction behind it, and a confirmed delta must reach the published pages via `SPEC-MODEL-MATRIX-DOCS-001` rather than stopping at `progress.md`.

---

## §C Known issues carried into execution

| # | Issue | Where it bites |
|---|---|---|
| K-1 | The template agent frontmatter's effort values differ from the incoming Medium column in several of ten files. The downstream "no-op at medium" claim holds only after the template baseline is re-set. | M1 step 7 → CONFIG's M3 |
| K-2 | `DefaultProfileMatrix()`'s Go type is unchanged while its inner-key semantics change from group key to agent name — a semantic change with no compiler signal. | M1 |
| K-3 | `README.md`'s benchmark columns `$/solved` and `Tokens/solved` are derived metrics, not raw leaderboard columns. S0 must pin column semantics, not only values. | S0 → DOCS |
| K-4 | `profileMatrixAgentOrder` carries 11 entries while `agentGroupMembership` carries 10 — the asymmetry that made `Explore` an accidental `inherit`. | M1 |
| K-5 | M1 plan steps 2 and 3 are recorded as landed but are observably unexecuted in this tree (`profile_matrix.go:209`, `profile_matrix.go:464`, consumer at `internal/web/agentfm.go:491`). | M1 — disposition owned by card t1037 |

---

## §D Pre-flight

1. `git fetch origin` and check divergence — a parallel session may be active on this shared checkout.
2. Resolve the `Group`-field clarification in `research.md` §F; it gates M1's report-shape step and CONFIG's M3.
3. Reconcile the §B.3 divergence between the landed matrix distribution and `spec.md` §A.3.
4. Establish a green baseline for the affected packages: `go build ./... && go test ./internal/template/...`.

---

## §E Constraints carried into execution

Template-First (`make build` after every template edit), template content neutrality, No-Haiku as a HARD non-skippable gate, `inherit` never passed as an `Agent`-tool `model` argument, and explicit pathspecs on every commit.

---

## §F Milestones

### S0 — Leaderboard verification (BLOCKING, NOT DISCHARGED)

**Priority: Critical. Blocks `SPEC-MODEL-MATRIX-DOCS-001` entirely. Does not block M1.**

Confirm the DeepSWE v1.1 figures against the live source and produce the verification record specified in `research.md` §A.5.

1. Read the leaderboard for Fable 5, Opus 4.8, and GLM 5.2, capturing **the effort level of each row read**.
2. Resolve the GLM 5.2 Pass@1 conflict to a canonical value; state whether the source publishes a point estimate, an interval, or both.
3. Pin the token-column semantics (raw output tokens vs tokens-per-solved) so derived README columns can be reconstructed correctly.
4. Record the leaderboard version string and update date.
5. Produce a delta table against `spec.md` §A.1 for every metric that differs; write it to `progress.md`.
6. **Remediation step, new to this SPEC**: where a confirmed figure differs from what already shipped in the landed M6 documentation, raise the correction to `SPEC-MODEL-MATRIX-DOCS-001` rather than closing S0 against `progress.md` alone.

Do **not** copy figures from `research.md` §A.2 — that table is a summarised probe result explicitly marked insufficient.

Covers REQ-MPMC-001 … 004.

### M1 — 33-cell matrix redesign (LANDED, with an unexecuted remainder)

**Priority: High. Blocks `SPEC-MODEL-MATRIX-CONFIG-001` and `SPEC-MODEL-MATRIX-DOCS-001`.**

1. Replace `defaultProfileMatrix` with the direct `profile → agent → {model, effort}` map, transcribing all 33 cells verbatim from `spec.md` §A.3. — *landed*
2. Delete the six group constants and `agentGroupMembership`. — **NOT executed** (t1037)
3. Delete `AgentGroup`; update `internal/cli/model.go` `resolveModelProfileReport` accordingly (per the §B.2 decision). — **NOT executed** (t1037); the consumer at `internal/web/agentfm.go:491` is live
4. Add the explicit `Explore` row; keep the unmapped-agent `inherit` fallback.
5. Rename the resolver's second return value from `hasGroup` to a name describing what it now means (`injectable` / `mapped`) and update every call site.
6. Update `DefaultProfileMatrix()`'s doc comment to state the inner key is an agent name (K-2), and update the `@MX:ANCHOR` reason lines on the matrix and resolver.
7. Re-set the **template** agent frontmatter `effort:` values to the new Medium column (K-1), then `make build`.
8. Amend tests: rewrite the fidelity and low-column expectations; retarget the override-precedence test's second assertion; **split** the inherit test into an `Explore` case and an unmapped-agent case.
9. Add the display-order/matrix-key agreement assertion (`design.md` §A.3).

Covers REQ-MPMC-005 … 018.

---

## §G Anti-patterns to avoid

| # | Anti-pattern | Why it bites here |
|---|---|---|
| AP-1 | Diffing the new matrix against the old 18 cells to "verify" the transcription | The new values are not a reshape; a diff review mistakes intentional inversions for errors (§B.1) |
| AP-2 | Copying benchmark numbers from `research.md` §A.2 | That table is a summarised probe result explicitly marked insufficient for documentation |
| AP-3 | Treating M1 as discharged because the landing record says so | Steps 2 and 3 are observably unexecuted and the landed distribution does not match §A.3 (§B.3, K-5) |
| AP-4 | Removing `agentGroupMembership` / `AgentGroup` on this SPEC's own authority | Card t1037 owns the disposition; a live consumer exists |
| AP-5 | Closing S0 against `progress.md` while the shipped pages still carry unverified figures | S0 is remediation here, not prevention (§B.4) |
| AP-6 | `git add -A` on this shared checkout | A parallel session may have staged unrelated work; use explicit pathspecs |

---

## §H Cross-References

- `spec.md` — requirements, the 33-cell matrix, decisions, risks
- `acceptance.md` — AC matrix and verification commands
- `design.md` — data-model rationale, the unmapped-agent path, rejected alternatives
- `research.md` — observed current-state facts, S0 probe status, open clarifications, unverified items
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — retired predecessor stub
