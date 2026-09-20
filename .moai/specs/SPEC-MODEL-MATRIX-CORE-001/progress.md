# Progress — SPEC-MODEL-MATRIX-CORE-001

Card: t1036 · Tier L · split from `SPEC-MODEL-PROFILE-MATRIX-002` on the S0 + M1 seam, 2026-09-20.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 18 / ceiling 25; AC 13 / ceiling 25 |
| Tier basis | REQ 18 exceeds the Tier M ceiling of 16; file scope >15 (Go SSOT + tests, two config mirrors, ten agent frontmatter files × two mirrors, the report consumer) |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 18 (REQ-MPMC-001 … REQ-MPMC-018) |
| Acceptance criteria | 13 (AC-MPMC-001 … AC-MPMC-013) |
| Predecessor | none — root of the chain |
| Successors | `SPEC-MODEL-MATRIX-CONFIG-001`, `SPEC-MODEL-MATRIX-DOCS-001` |
| Status transition | inherited `in-progress` (NOT `(none) → draft`) — M1 landed under the original SPEC ID |

Open question carried to the Implementation Kickoff Approval gate: the `Group` field's disposition in `moai model profile --json` (`research.md` §F).

Unverified inputs inherited and assigned here: the DeepSWE `[low]` Fable row, the `[high]` Opus row, the canonical GLM 5.2 Pass@1 point value, and the leaderboard's token-column semantics — all S0's. Plus one added by the split: whether the landed matrix actually equals `spec.md` §A.3.

## §E.2 Run-phase Evidence

Run-phase is **PARTIAL**, inherited from `SPEC-MODEL-PROFILE-MATRIX-002`. One of this SPEC's two units has landed.

### Landed

| Unit | Scope | Evidence |
|---|---|---|
| M1 | 33-cell matrix redesign — Go SSOT `internal/template/profile_matrix.go`, `llm.yaml` + `workflow.yaml` (local + template mirrors), 10 agent frontmatter files ×2 mirrors, `model-policy.md` + `agent-authoring.md` ×2 mirrors, affected test fixtures | squash `31da99a7b` (PR #1163) |

Matrix as landed, per the inherited record: Opus 25 cells / Sonnet 8 cells; Fable 0 cells; `xhigh` 0 cells; `max` in exactly 2 cells (`manager-develop`, `super-advisor`, `high` column only). Cross-checked row-by-row against `template.DefaultProfileMatrix` and `moai model profile`.

[HARD] **That distribution does not match `spec.md` §A.3**, which carries six Fable cells, two `xhigh` cells, and no `max` effort. The inconsistency is inherited unreconciled and is recorded as a Gap below rather than resolved by the split.

Verification observed at landing: `go test ./...` exit 0 / 0 FAIL · `golangci-lint run` 0 issues · `go vet ./...` clean · `hugo --minify --gc` exit 0 with zero warnings · `moai agent lint` 0 errors / 24 warnings (identical to the pre-change baseline), measured with a binary built from the change tree.

### Not landed

| Unit | Status |
|---|---|
| **S0 — Leaderboard verification (BLOCKING)** | **NOT discharged.** The `research.md` §A probe is a lead, not evidence, and contradicts the user-supplied readings for two of three models. M1 and the downstream M6 proceeded on the user-supplied per-effort measurements instead; S0's own verification was never performed. Because the documentation S0 blocks has already shipped, discharging S0 is now remediation rather than prevention. |

### Unexecuted M1 remainder

M1's plan steps 2 and 3 were never executed despite the landing record. Measured in this tree (develop lineage):

- `internal/template/profile_matrix.go:209` — `var agentGroupMembership = map[string]string{`
- `internal/template/profile_matrix.go:464` — `func AgentGroup(agent string) (string, bool)`
- `internal/web/agentfm.go:491` — `template.AgentGroup(a.Name)` called as a live gate

Disposition — landing record wrong, or plan stale — is owned by card **t1037**. This SPEC records the remainder as its own; AC-MPMC-005 is recorded as unmet rather than relaxed.

### Gaps

- **13 acceptance criteria: 0 formally verified.** The verification above is toolchain-level (build / test / lint / docs-build / matrix cross-check), not AC-level.
- **`internal/` carries no `REQ-MPM2` / `AC-MPM2` markers** (grep count 0), so requirement-to-code traceability is unestablished for the landed work. No `REQ-MPMC` / `AC-MPMC` markers exist either — this SPEC's own vocabulary is newer than the code.
- **The `Group`-field clarification** in `research.md` §F remains open, and no Implementation Kickoff Approval is recorded for the original SPEC or this successor.
- **The landed-matrix / §A.3 divergence** above is unreconciled.

### Residual risk

The benchmark driving the matrix measures **coding** agents. Documentation authoring, audit judgment, and SPEC authoring quality are not directly measured — those row placements rest on a similarity inference to multi-turn agentic work. Every row is reversible per-agent via `llm.agent_overrides`.

A second figure was corrected during the original M6: the design report stated `xhigh` occupied 7 matrix cells before the change; counting the pre-change `defaultProfileMatrix` at `HEAD^` gave **6**. The docs carry 6; the original report itself is uncorrected.

## §E.3 Run-phase Audit-Ready Signal

_<pending — run-phase is partial; S0 outstanding, the M1 remainder unexecuted, and 13 AC unverified. Not audit-ready.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — blocked on run-phase completion>_
