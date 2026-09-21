# Implementation Plan — SPEC-MODEL-MATRIX-CONFIG-001

Tier L. Two milestones: M2 (retire the 36-cell axis and the config mirror) and M3 (effort actualization on both channels). Neither has started.

---

## §A Context

This SPEC carries the second seam of the `SPEC-MODEL-PROFILE-MATRIX-002` split: everything that makes the matrix the *only* source of truth, and everything that makes the matrix's effort axis actually reach a running agent.

Scope surfaces: `internal/config` (`model_routing.go`, `types.go`, their tests), both `workflow.yaml` copies, both `llm.yaml` copies, `internal/cli/init.go`, both `internal/cli/update.go` seams, `internal/web/agentfm.go`, `internal/settings/agentfm/agentfm.go`, and a new effort-application unit with its tests.

**Inherited state**: nothing here has landed. `model_routing_profiles` is still present in `workflow.yaml`; the effort application does not exist.

---

## §B Review first — the decisions most likely to change

### §B.1 The frontmatter rewrite revival and its four mitigations (M3)

The predecessor was retired for real side effects. The revival's safety rests entirely on the four mitigations holding **simultaneously**: `effort:`-only, deployed-tree-only, post-deploy, override-excluded. Dropping any one reproduces a variant of the original failure.

The `effort:`-only mitigation in particular must be implemented as **"pass leave-unchanged for model"**, never as a read-modify-write (`design.md` §C.2). A read-modify-write is observationally identical in the happy path and silently re-pins the model the moment the read is wrong — which is precisely the `[1m]`-hazard the predecessor was retired for.

### §B.2 Whether the `profiles:` warning branch should exist at all (M2)

REQ-MPME-007 splits on whether a user's customization is "representable as per-agent overrides". A group-keyed cell expands to 1-3 agent overrides mechanically, so representability is arguably always true under the old six-group shape. If it is always true, the warning branch is dead code shipped as a safety net — and dead safety nets rot. `research.md` §F carries this as an open clarification; resolve it before M2 lands, because the answer changes whether REQ-MPME-007 has one branch or two.

### §B.3 The `moai model profile --json` shape as a consumed contract (M3)

REQ-MPME-019 elevates this JSON from a reporting convenience to a contract a workflow script parses. `SPEC-MODEL-MATRIX-CORE-001` decides the shape (whether the `Group` key is dropped); this SPEC must not start M3 against an undecided shape, or the contract ships and then breaks.

Whether an agent-name filter flag is needed — so a script need not parse the full 11-entry array — is a shape decision too, and belongs in the same review.

### §B.4 The four seams and their deploy ordering (M3)

```
WRONG:  apply-effort → deploy-templates      (deploy overwrites the effort lines)
RIGHT:  deploy-templates → apply-effort
```

`internal/cli/update.go`'s `--profile` flag path already calls `ApplyProfile` after the deploy step, so the correct ordering exists at that seam. The other three must each be checked individually — the `init` seam has no prior deploy to race, and the web seam has no deploy at all. Getting this wrong makes a user's profile silently revert on every update, which presents as intermittent rather than broken.

---

## §C Known issues carried into execution

| # | Issue | Where it bites |
|---|---|---|
| K-1 | REQ-MPME-022's "no-op at medium" holds only after the predecessor's M1 step 7 re-sets the template baseline. Treating it as already true produces visible churn. | M3 — depends on `SPEC-MODEL-MATRIX-CORE-001` |
| K-2 | `TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` asserts precisely the behaviour M2 removes. It must be deleted, not amended. | M2 |
| K-3 | The comment block at `internal/web/agentfm.go` `applyPerfTierEdits` states frontmatter re-application is retired. M3 makes that comment false. | M3 |
| K-4 | The haiku-residual rule's surfaces 3 and 4 target artifacts M2 deletes. Their removal is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation — hand it off explicitly. | M2 → DOCS coupling |
| K-5 | The two `llm.yaml` copies have already drifted in the `glm.models` block (local has `opus`/`sonnet`/`haiku` alias keys the template lacks). A whole-block sync would import `haiku` into the template. M2 must edit the `profiles:` block only. | M2 |
| K-6 | `internal/web/agentfm.go` is edited by this SPEC and by `SPEC-MODEL-MATRIX-SURFACES-001` for unrelated reasons. | M3 — sequence, do not merge |

---

## §D Pre-flight

1. Confirm `SPEC-MODEL-MATRIX-CORE-001` has landed the matrix and, specifically, its M1 step 7 template-baseline re-set (K-1).
2. Confirm the `Group`-field JSON shape decision is recorded (§B.3).
3. Resolve the `profiles:` migration-representability clarification in `research.md` §F (§B.2).
4. `git fetch origin` and check divergence — a parallel session may be active on this shared checkout.
5. `go build ./... && go test ./internal/config/... ./internal/template/...` to establish a green baseline.

---

## §E Constraints carried into execution

Template-First (`make build` after every template edit), template content neutrality, No-Haiku as a HARD non-skippable gate, legacy-block load tolerance as a compatibility requirement rather than a nicety, and explicit pathspecs on every commit.

---

## §F Milestones

### M2 — Retire the 36-cell axis and the config mirror

**Priority: High. Depends on `SPEC-MODEL-MATRIX-CORE-001`. Hands off two guard-surface removals to `SPEC-MODEL-MATRIX-DOCS-001`.**

1. Remove the `model_routing_profiles` block from `.moai/config/sections/workflow.yaml` and its template copy, including the explanatory comment lines that reference it.
2. Delete `RouteModelFor`, `ModelRoutingProfiles`, the `validRoutingModels` map, and the four `model_routing_profiles.*` validator branches.
3. Delete the orphaned tests in `internal/config/model_routing_test.go` (REQ-MPME-003).
4. Remove the `profiles:` block from both `llm.yaml` copies — **the `profiles:` block only** (K-5). Remove the ~20 lines of group-explaining comment in the template copy at the same time.
5. Make the config loader tolerate a legacy `profiles:` or `model_routing_profiles` block as inert rather than failing (REQ-MPME-008).
6. Implement `moai update` detection of a non-empty `profiles:` block, with migration into `agent_overrides` or a warning naming the affected cells (REQ-MPME-006/007/009).
7. Delete `TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` (K-2).
8. Verify the matrix literal now exists in exactly two places.
9. Record the handoff to `SPEC-MODEL-MATRIX-DOCS-001`: the haiku-residual rule now scans two deleted artifacts (K-4).

Covers REQ-MPME-001 … 009.

### M3 — Effort actualization (both channels)

**Priority: High. Depends on `SPEC-MODEL-MATRIX-CORE-001`. Highest review priority in this SPEC.**

1. Implement the agent-effort application function per `design.md` §C.2: enumerate deployed agent files, skip override-carrying agents, skip unmapped agents, skip absent files, write `effort:` only.
2. Ensure the frontmatter patcher can express "leave `model:` unchanged" — extend it if the current signature cannot (§B.1 — read-modify-write is prohibited).
3. Wire the four seams: `internal/cli/init.go`, both `internal/cli/update.go` paths, `internal/web/agentfm.go`. Verify each is **after** any template deploy at that seam.
4. Adopt the continue-on-error posture: a single agent's rewrite failure logs and continues. A half-applied profile is strictly better than a failed `moai update`.
5. Rewrite the stale comment block at `applyPerfTierEdits` (K-3).
6. Expose the Workflow-path lookup route — document `moai model profile --json` as the contract, adding an agent-name filter flag if the full array is impractical for a script.
7. Produce the two-channel explanation content, including the explicit note that `SPEC-MODEL-PROFILE-MATRIX-001` DECISION-001's wording is superseded and that `Explore` has no file. Hand the placement and locale work to `SPEC-MODEL-MATRIX-DOCS-001`.
8. Tests: override-exclusion, template-immutability (assert no write under `internal/template/templates/`), model-line-preservation, absent-file skip, post-deploy ordering, dev-repo no-op at `medium`.

Covers REQ-MPME-010 … 022.

---

## §G Anti-patterns to avoid

| # | Anti-pattern | Why it bites here |
|---|---|---|
| AP-1 | Read-modify-write on agent frontmatter to preserve `model:` | Observationally identical in the happy path; silently re-pins the model when the read is wrong (§B.1) |
| AP-2 | Whole-block sync of the two `llm.yaml` copies | Imports the local-only `glm.models` alias keys — including `haiku` — into the template (K-5) |
| AP-3 | Iterating matrix keys when applying effort | Breaks on `Explore`, which is mapped but has no file |
| AP-4 | Applying effort before template deploy | The deploy overwrites the effort lines; the user's profile silently reverts each update (§B.4) |
| AP-5 | Treating REQ-MPME-022 ("no-op at medium") as already true | It becomes true only after the predecessor's M1 step 7 (K-1) |
| AP-6 | Leaving the retired haiku-rule surfaces for someone to notice | A guard scanning a deleted artifact implies to future readers that the artifact still exists; hand it off explicitly (K-4) |
| AP-7 | Making the config loader reject a legacy block | Breaks existing projects on upgrade; tolerance is a requirement, not a nicety |
| AP-8 | Merging this SPEC's `agentfm.go` edits with SURFACES' | Two independent concerns in one file; sequence them (K-6) |
| AP-9 | `git add -A` on this shared checkout | A parallel session may have staged unrelated work; use explicit pathspecs |

---

## §H Cross-References

- `spec.md` — requirements, the two-channel schema evidence, decisions, risks
- `acceptance.md` — AC matrix and verification commands
- `design.md` — the two channels, the ordering hazard, the template-baseline prerequisite, rejected alternatives
- `research.md` — observed current-state facts, open clarifications, unverified items
- `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/` — predecessor
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — retired predecessor stub
