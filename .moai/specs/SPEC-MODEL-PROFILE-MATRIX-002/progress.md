# Progress — SPEC-MODEL-PROFILE-MATRIX-002

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set authored 2026-07-23: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

- Tier L, `status: draft`.
- 72 requirements (REQ-MPM2-001 … 113, non-contiguous by milestone block).
- 64 acceptance criteria (AC-MPM2-001 … 103, non-contiguous by milestone block).
- 8 units of work: S0 (blocking precondition) + M1 … M7.
- 4 open `[NEEDS CLARIFICATION]` items recorded in `research.md` §F — all must be resolved before Implementation Kickoff Approval.
- S0 is **not discharged**. The plan-phase probe result in `research.md` §A is a lead, not evidence; it contradicts the user-supplied readings for two of three models and is explicitly marked insufficient for documentation use.
- 8 unverified inputs carried forward and enumerated in `research.md` §G; each is assigned to a milestone for re-measurement.

## §E.2 Run-phase Evidence

Run-phase is **PARTIAL**. Two of eight units of work have landed; the blocking precondition S0 and four milestones remain. `status: in-progress` (not `implemented`) reflects that state.

### Landed

| Unit | Scope | Evidence |
|------|-------|----------|
| M1 | 33-cell matrix redesign — Go SSOT `internal/template/profile_matrix.go`, `llm.yaml` + `workflow.yaml` (local + template mirrors), 10 agent frontmatter files ×2 mirrors, `model-policy.md` + `agent-authoring.md` ×2 mirrors, affected test fixtures | squash `31da99a7b` (PR #1163) |
| M6 | 4-locale documentation + README ×4 — 13 docs-site pages per locale; profile-matrix, no-haiku-3tier, model-policy, agent-guide, faq, cli, init, update, introduction, config-sections, tokenomics-overview, what-is-moai-adk, init-wizard | same squash `31da99a7b` |

Also landed in the same squash, outside the milestone list: the `moai init` wizard's on-screen model-policy labels (`internal/cli/profile_setup_translations.go` ×4 locales), which still read `"Max - Fable 5 (low) + Opus 4.8 (high)"` while the option value was already `high`. Not a milestone deliverable — a defect the M1 change surfaced.

Matrix as landed: Opus 25 cells / Sonnet 8 cells; Fable 0 cells; `xhigh` 0 cells; `max` in exactly 2 cells (`manager-develop`, `super-advisor`, `high` column only). Cross-checked row-by-row against `template.DefaultProfileMatrix` and `moai model profile`.

Verification observed at landing: `go test ./...` exit 0 / 0 FAIL · `golangci-lint run` 0 issues · `go vet ./...` clean · `hugo --minify --gc` exit 0 with zero warnings · `moai agent lint` 0 errors / 24 warnings (identical to the pre-change baseline), measured with a binary built from the change tree.

### Not landed

| Unit | Status |
|------|--------|
| **S0 — Leaderboard verification (BLOCKING)** | **NOT discharged.** §E.1 records that the `research.md` §A probe is a lead, not evidence, and contradicts the user-supplied readings for two of three models. M1/M6 proceeded on the user-supplied per-effort measurements instead; S0's own verification was never performed. |
| M2 — Retire the 36-cell axis + config mirror | Not started. `model_routing_profiles` still present in `workflow.yaml`. |
| M3 — Effort actualization (both channels) | Not started. |
| M4 — Init wizard question | Not started. The wizard **label** fix above is not the M4 deliverable (M4 adds a question; this corrected existing option text). |
| M5 — Web console cleanup | Not started under this SPEC. |
| M7 — Guard realignment + full verification | Not started. |

### Gaps

- **64 acceptance criteria: 0 formally verified.** `acceptance.md` carries 64 unique `AC-MPM2-*` ids; none were evaluated against a PASS/FAIL matrix during M1/M6. The verification listed above is toolchain-level (build / test / lint / docs-build / matrix cross-check), not AC-level.
- **`internal/` carries no `REQ-MPM2` / `AC-MPM2` markers** (grep count 0), so requirement-to-code traceability is unestablished.
- **4 `[NEEDS CLARIFICATION]` items** in `research.md` §F remain open. §E.1 states all four must be resolved before Implementation Kickoff Approval; no Kickoff Approval is recorded for this SPEC.
- **8 unverified inputs** enumerated in `research.md` §G were not re-measured.

### Residual risk

The benchmark driving the matrix measures **coding** agents. Documentation authoring, audit judgment, and SPEC authoring quality are not directly measured — those row placements rest on a similarity inference to multi-turn agentic work, recorded as R1 (High) in the design report. Every row is reversible per-agent via `llm.agent_overrides`.

A second figure was corrected during M6: the design report states `xhigh` occupied 7 matrix cells before the change; counting the pre-change `defaultProfileMatrix` at `HEAD^` gives **6**. The docs carry 6; the report itself is uncorrected.

## §E.3 Run-phase Audit-Ready Signal

_<pending — run-phase is partial; S0 + M2/M3/M4/M5/M7 outstanding and 64 AC unverified. Not audit-ready.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — blocked on run-phase completion>_
