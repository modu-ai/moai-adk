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

### M1 remainder — disposition ruled (card t1037)

Card **t1037** answered the question this section previously carried open. **The plan is stale; the
landing record is not wrong.** M1's plan steps 2 and 3 were not executed, and that non-execution was
a deliberate decision taken at landing time rather than an omission. Full judgment and evidence:
`.moai/reports/t1037/verdict.md`.

All coordinates below were re-measured in this tree at HEAD `8f87f2359` after absorbing develop —
not carried over from the pre-merge measurement.

#### What the squash did and did not touch

Symbol counts across `31da99a7b^` (pre-landing) / `31da99a7b` (the landing squash) / HEAD, all from
`internal/template/profile_matrix.go`:

| Symbol | pre | post | HEAD | Reading |
|---|---|---|---|---|
| `agentGroupMembership` | 3 | 3 | 3 | untouched by the squash |
| `func AgentGroup` | 1 | 1 | 1 | untouched by the squash |
| `{Model:` | 22 | 36 | 42 | **positive control** — the instrument is live |
| `GroupExplore` | 6 | 3 | 3 | **positive control** |

The two controls matter: they are the same file and the same tool, and they move. So the invariant
3/3/3 and 1/1/1 are an absence of change, not a failed measurement.

#### The decision is recorded in the code itself

The string `display classification` occurs 0 times before the squash and 1 time after — the squash
**authored** the retention rationale. At HEAD it reads, `internal/template/profile_matrix.go:481-483`:

```go
// Lookup is by agent NAME, not by group: per-agent cells split two of the former
// groups, so the group layer no longer carries routing information and survives
// only as a display classification (see AgentGroup).
```

The plan's reasoning was "the layer stops carrying routing, therefore delete it". The executor
reached "the layer stops carrying routing, therefore keep it as a display classification" and wrote
that down. `plan.md` was never updated to match.

#### Per-step state

| Step | State | Evidence |
|---|---|---|
| 1 — agent-keyed matrix | landed | `profile_matrix.go:259` doc comment, `:320` declaration |
| 2 — delete group constants + `agentGroupMembership` | **retired** | deliberate deviation, above; symbol still at `:209` |
| 3 — delete `AgentGroup` | **retired** | same; symbol still at `:464`, live consumers below |
| 4 — explicit `Explore` row | landed | `:334`, `:349`, `:364` — all three profiles |
| 5 — rename `hasGroup`, update every call site | **PARTIAL — disposition NOT ruled** | declaration renamed to `mapped` at `:487`; `hasGroup` survives at 8 occurrences (`internal/cli/model.go:99`, `:110`; `internal/template/profile_matrix_test.go:278`, `:285`, `:286`, `:297`, `:302`, `:303`) |
| 6 — doc comment states agent-name key | landed | `:259`, `:318` |

[HARD] Step 5 is **not** covered by the ruling above. Steps 2 and 3 have an authored decision on
record; step 5 has none, and its rename rationale (the value names membership, not routing) applies
to the call sites exactly as it applies to the declaration. It may therefore be genuine unfinished
execution rather than a stale plan item. t1037 did not decide it and this record does not either.

#### The retained layer carries two loads, and the code comment names only one

`template.AgentGroup` has exactly two consumers (`grep -rn --include='*.go' 'AgentGroup' .`):

| Consumer | What it uses | Matches "display classification"? |
|---|---|---|
| `internal/cli/model.go:101` | the group **string** → report `Group` column (`:38`), table output (`:149`, `:154`) | yes |
| `internal/web/agentfm.go:491` | **the bool only** (`if _, ok := ...; !ok`) → rejects override submissions from non-matrix agents | **no — this is a gate, not display** |

So the retention rationale's own claim is inaccurate. A later reader who removes the layer on the
strength of that sentence would move the display column and silently break the gate. Raised as a
separate card candidate; deliberately NOT folded into t1037 or this SPEC.

#### What this does to AC-MPMC-005

AC-MPMC-005 asserts that no group constant, no `agentGroupMembership`, and no `AgentGroup` exists
after M1. Under this ruling the criterion asserts an absence that a deliberate landing-time decision
chose never to create, so it is carried as an **explicitly-owned debt** per `acceptance.md` §H
closure gate 5 — not relaxed, and not rewritten. Whether REQ-MPMC-006 / 007 / 008 should now be
retired with it is a product decision that remains **open and unowned**: t1037's scope was the
disposition of the plan steps, not the fate of the requirements.

### Gaps

- **13 acceptance criteria: 0 formally verified.** The verification above is toolchain-level (build / test / lint / docs-build / matrix cross-check), not AC-level.
- **`internal/` carries no `REQ-MPM2` / `AC-MPM2` markers** (grep count 0), so requirement-to-code traceability is unestablished for the landed work. No `REQ-MPMC` / `AC-MPMC` markers exist either — this SPEC's own vocabulary is newer than the code.
- **The `Group`-field clarification** in `research.md` §F remains open, and no Implementation Kickoff Approval is recorded for the original SPEC or this successor.
- **The landed-matrix / §A.3 divergence** above is unreconciled.

### Residual risk

The benchmark driving the matrix measures **coding** agents. Documentation authoring, audit judgment, and SPEC authoring quality are not directly measured — those row placements rest on a similarity inference to multi-turn agentic work. Every row is reversible per-agent via `llm.agent_overrides`.

A second figure was corrected during the original M6: the design report stated `xhigh` occupied 7 matrix cells before the change; counting the pre-change `defaultProfileMatrix` at `HEAD^` gave **6**. The docs carry 6; the original report itself is uncorrected.

## §E.3 Run-phase Audit-Ready Signal

_<pending — run-phase is partial; S0 outstanding and 13 AC unverified. The M1 remainder is no longer an open question (t1037 ruled the plan stale, §E.2), but AC-MPMC-005 is carried as owned debt and step 5's disposition is unruled. Not audit-ready.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — blocked on run-phase completion>_
