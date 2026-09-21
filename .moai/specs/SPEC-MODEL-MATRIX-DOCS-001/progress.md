# Progress — SPEC-MODEL-MATRIX-DOCS-001

Card: t1036 · Tier L · split from `SPEC-MODEL-PROFILE-MATRIX-002` on the M6 + M7 seam, 2026-09-20.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 21 / ceiling 25; AC 21 / ceiling 25 |
| Tier basis | REQ 21 and AC 21 both exceed the Tier M ceiling of 16; file scope far above 15 (13 docs-site pages × 4 locales, 4 README copies, 2 byte-parity rule twins, 2 enum files, the lint rule and its tests) |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 21 (REQ-MPMD-001 … REQ-MPMD-021) |
| Acceptance criteria | 21 (AC-MPMD-001 … AC-MPMD-021) |
| Predecessors | `SPEC-MODEL-MATRIX-CORE-001`, `SPEC-MODEL-MATRIX-CONFIG-001`, `SPEC-MODEL-MATRIX-SURFACES-001` |
| Successor | none — last of four |
| Status transition | inherited `in-progress` (NOT `(none) → draft`) — M6 landed under the original SPEC ID |

Open question carried to the Implementation Kickoff Approval gate: the infographic disposition — leave a visibly stale image beside a corrected table, or remove the image reference pending regeneration (`research.md` §F).

Unverified inputs inherited and assigned here: the ja/ko `multi-llm/model-policy.md` shape, the `advanced/*` line ranges, the tokenomics infographic contents, and cross-platform build status. All four survived M6's landing unmeasured. One more is added by the split: whether M6's landed pages actually resolve the three documentation contradictions, which no AC-level verification has established.

## §E.2 Run-phase Evidence

Run-phase is **PARTIAL**, inherited from `SPEC-MODEL-PROFILE-MATRIX-002`. One of this SPEC's two milestones has landed.

### Landed

| Unit | Scope | Evidence |
|---|---|---|
| M6 | 4-locale documentation + README ×4 — 13 docs-site pages per locale; profile-matrix, no-haiku-3tier, model-policy, agent-guide, faq, cli, init, update, introduction, config-sections, tokenomics-overview, what-is-moai-adk, init-wizard | squash `31da99a7b` (PR #1163) |

Verification observed at landing: `go test ./...` exit 0 / 0 FAIL · `golangci-lint run` 0 issues · `go vet ./...` clean · `hugo --minify --gc` exit 0 with zero warnings · `moai agent lint` 0 errors / 24 warnings (identical to the pre-change baseline), measured with a binary built from the change tree.

[HARD] **M6 shipped with its blocking precondition undischarged.** S0 — the leaderboard verification record, owned by `SPEC-MODEL-MATRIX-CORE-001` — was never performed. M6 proceeded on the user-supplied per-effort measurements instead. Every benchmark figure now live in the README set and in `advanced/no-haiku-3tier.md` therefore traces to an unverified reading rather than to a record.

The consequence is a correction obligation rather than a gate: REQ-MPMD-011's correction clause, `plan.md` §F step 10, and AC-MPMD-012. No criterion can retroactively close the interval during which live pages carried unverified figures; that window is recorded as accepted history.

### Not landed

| Unit | Status |
|---|---|
| M7 — Guard realignment + full verification | Not started. The haiku-residual rule still scans `model_routing_profiles` and `validRoutingModels`; neither of the two surfaces M7 adds is on its list; the consolidated matrix property test does not exist; no cross-platform build was run. |

M7 is additionally gated in both directions: its guard **additions** wait on `SPEC-MODEL-MATRIX-SURFACES-001` clearing both surfaces, and its guard **removals** wait on `SPEC-MODEL-MATRIX-CONFIG-001` deleting both artifacts.

### Gaps

- **21 acceptance criteria: 0 formally verified.** The verification above is toolchain-level (build / test / lint / docs-build / matrix cross-check), not AC-level.
- **`internal/` carries no `REQ-MPM2` / `AC-MPM2` markers** (grep count 0), so requirement-to-code traceability is unestablished for the landed work. No `REQ-MPMD` / `AC-MPMD` markers exist either.
- **No evidence that M6's edits achieved what their requirements asked.** The landing record names the pages it touched; a page being edited is not evidence the edit satisfied its requirement.
- **Four unverified inputs survived M6 unmeasured** (§E.1), and the infographic-disposition decision was never recorded.
- **The S0 correction (AC-MPMD-012) is currently vacuously satisfiable** — with no record, the delta set is empty and an early evaluation would report PASS while closing nothing. It must be evaluated after S0 lands.

### Residual risk

The benchmark driving the documented matrix measures **coding** agents. Documentation authoring, audit judgment, and SPEC authoring quality are not directly measured — those row placements rest on a similarity inference to multi-turn agentic work. Every row is reversible per-agent via `llm.agent_overrides`, and the pages should not imply otherwise.

The Max profile draws on Fable for six of its eleven cells, so an environment without Fable access degrades in a way the documentation must name rather than merely acknowledge (REQ-MPMD-014).

## §E.3 Run-phase Audit-Ready Signal

_<pending — run-phase is partial; M7 not started, the predecessor's S0 undischarged, the correction step outstanding, and 21 AC unverified. Not audit-ready.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — blocked on run-phase completion>_
