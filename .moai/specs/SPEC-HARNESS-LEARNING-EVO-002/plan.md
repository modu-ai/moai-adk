# SPEC-HARNESS-LEARNING-EVO-002 — Implementation Plan (L2)

> Derived from `spec.md`. Ordered by decision-reversibility: the contract and the two contested decisions first, the mechanical wiring last.

## §A. Context and measured baseline

### A.1 Baseline commands (re-runnable)

All figures were measured against the **primary checkout** `/Users/goos/MoAI/moai-adk-go` on 2026-08-09. They are not reproducible from this worktree: `.moai/state/`, `.moai/harness/`, and `.moai/evolution/` are gitignored runtime state and absent here. Small upward drift on re-run is expected.

| Claim | Command | Observed |
|---|---|---|
| The observe channel is empty | `wc -l < .moai/state/routing-ledger.jsonl` ; `jq -c '.delegations \| length' … \| sort \| uniq -c` | 4 rows; `4  0` |
| The proposal writer and layout work | `ls .moai/harness/proposals \| wc -l` ; `wc -l < .moai/harness/learning-history/tier-promotions.jsonl` | 73 dirs ; 175 rows |
| Agent identity is a mixed population | `grep '"subagent_stop"' .moai/harness/usage-log.jsonl \| jq -r '.agent_type // "(none)"' \| sort \| uniq -c \| sort -rn` | 2,783 rows; 1,941 attributed (69.7%); 842 `(none)`; `general-purpose` 204, `workflow-subagent` 186, ~140 one/two-occurrence spawn names |
| Nothing writes the map | `grep -rn 'delegation.yaml' internal/ --include='*.go' \| grep -v _test` | 0 matches |
| Two designations are conditional | `.moai/config/sections/delegation.yaml` `sync: agents: [manager-docs, sync-auditor]`, `plan: agents: [manager-spec, plan-auditor, Explore]` | both present |

### A.2 Existing plumbing this SPEC builds on

| Surface | Path | Role here |
|---|---|---|
| Reader | `internal/harness/routing/reader.go` | `NewReader(path).Read(Filter)` — the only ledger entry point; returns `([]Row, skipped int, error)` |
| Row shape + finalize | `internal/harness/routing/types.go` | `Row`, `PendingRow.Finalize(outcome)` — the fixture generator's source of truth |
| Proposal writer + layout | `internal/harness/proposalgen/{scaffolder,layout}.go` | `WriteProposals`, `ProposalDir`, `ListDraftIDs` — reused as-is |
| Event-type SSOT | `internal/harness/types.go` | `PatternBearingEventTypes()` — read for the isolation assertion, never extended |
| Delegation map | `.moai/config/sections/delegation.yaml` | `delegation.subcommands.<name>.agents` — read-only source of the designated set |
| Harness CLI parent | `internal/cli/harness_route.go` | `newHarnessRouteCmd`'s `AddCommand` block (the `ledger` verb is registered at line 144); the new `delegation` verb attaches here |
| Retained agent catalog | `CLAUDE.md` §4 | the 12-member membership test for REQ-HLA-004 |

## §B. Known issues the plan must not reintroduce

- **B1 — namespace capture.** Adding a `delegation_map` member to `harness.PatternBearingEventTypes()` would silently widen the *observer's* event taxonomy and the proposalgen format regex derived from it. Forbidden by REQ-HLA-011.
- **B2 — identity naivety.** Treating every observed `agent_type` value as an agent name produces `undesignated_agent` findings for `general-purpose`, `workflow-subagent`, harness specialists, and ~140 spawn names — a proposal stream that is 100% noise on the measured population (§A.1).
- **B3 — absence-reasoning without exclusions.** `designated_never_spawned` reasons from absence. `sync-auditor` runs only at harness `thorough`; `plan-auditor` is skipped by the Plan Audit Gate skip-eligibility path. Without REQ-HLA-008 the analyzer reports prescribed behavior as a defect.
- **B4 — fixture drift.** Hand-written JSONL fixtures encode the author's belief about the row shape, not the producer's actual output. `PendingRow.Finalize` maps 15 fields plus the outcome and normalizes two nil slices; a dropped or mis-mapped field would be invisible to every hand-written fixture. AC-HLA-001 pins generation through `Finalize()`.
- **B5 — silent map mutation.** Any code path that opens `delegation.yaml` for writing, including a "harmless" comment or formatting rewrite.
- **B6 — assuming one session yields one row carrying both its delegations and its outcome.** It does not, reliably. Live dogfood of 001 established that a single `claude -p` invocation fires UserPromptSubmit and Stop **twice each** (trace in `SPEC-HARNESS-LEARNING-EVO-001/progress.md` §E.2), and one observed run produced **two** rows for one session, both `delegations: []` with `outcome: "success"`. The hypothesis — unreproduced, so stated as one — is that a terminal test signal arriving before the first subagent stops closes the row early at the mid-session Stop, leaving the later delegation to land on a fresh row.

  This is `SPEC-HARNESS-LEARNING-EVO-001` §F **R8**, open by decision and not fixed in 001. It bears directly on this SPEC because the row carrying the outcome is the row this analyzer counts as qualifying: delegations can be split across rows, or absent from the qualifying row entirely. An aggregation that assumes the one-session-one-complete-row shape will **under-count support ratios** and **over-count empty-delegation rows**, and both errors push findings in the same direction — toward suppressing real patterns rather than inventing false ones, which makes the failure quiet.

  The `MinQualifyingRows = 5` and `MinSupportRatio = 0.60` constants (§F M1) were chosen against the assumed shape; they are not invalidated by B6, but their effective strictness under row-splitting is unmeasured. Decide explicitly during M2 whether aggregation groups by `session_id` across rows or treats each row independently — and record which, because the choice changes what a support ratio means.

## §C. Pre-flight

1. Confirm the worktree is `feat/harness-learning-evo` and `git status` is clean of foreign edits.
2. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` green before the first edit (baseline attribution).
3. `golangci-lint run --timeout=2m 2>&1 | tail -5` — record the pre-existing baseline so NEW findings are separable.
4. `grep -rn 'delegation.yaml' internal/ --include='*.go' | grep -v _test` — record the 0-match starting state so the post-change single-match (the read-only loader) is attributable.

## §D. Constraints

- Read-only against `delegation.yaml`; no write path, ever (REQ-HLA-009, B5).
- No `AskUserQuestion` in the new package or the new CLI file; a boundary grep test mirrors `internal/harness/routing/subagent_boundary_test.go`.
- No template edits, no `.claude/skills/` edits, no frozen-allowlist edit.
- Do not extend `harness.PatternBearingEventTypes()` and do not route through `proposalgen.MapPromotions`.
- Do not touch runtime-managed files (`.moai/harness/*`, `.moai/state/*`).

## §E. Decisions carried into audit

### D1 — reader contract: accept the materializing read, bound it by size

The original SPEC carried two requirements that contradicted each other: "read through the existing routing reader" and "stream the ledger line by line and hold no whole-file buffer". `internal/harness/routing/reader.go:31` is `func (r *Reader) Read(f Filter) ([]Row, int, error)` — it materializes a slice, and no streaming variant exists.

**Chosen: relax to the reader's real contract, with a declared size bound.** Adding a streaming variant to `routing.Reader` would place a change to the producer's package inside this SPEC's milestone set, which makes this SPEC depend on its sibling and destroys the fixture-based independence that lets the two proceed in either order. Instead REQ-HLA-001 keeps the existing `Read` call, forbids retaining per-row state beyond the aggregate, and adds a declared maximum ledger size above which the analyzer declines the read and returns a machine-readable reason. The exposure is capped rather than removed (residual risk R4). Revisit if a real ledger ever approaches the bound; a streaming variant is then a change to the producer's package, correctly scoped as its own SPEC.

### D2 — catalog membership source

REQ-HLA-004's discrimination needs the retained-agent catalog. Two sources exist: the prose list in `CLAUDE.md` §4, and the union of agent names in `delegation.yaml`'s own `subcommands.*.agents`.

**Chosen: a declared constant in the package, seeded from `CLAUDE.md` §4, with the delegation map used only to resolve *designations* (not membership).** Deriving membership from the map would make the discrimination circular — an agent absent from every designation could never be found undesignated, which is precisely the finding the analyzer exists to produce. The cost is a snapshot that can go stale against a catalog change (residual risk R2); the constant carries a comment naming its source so a catalog edit has somewhere to land.

### D3 — F3 bounding

The original falsifier ("every proposal surfaced at the Tier-4 gate is rejected") was universally quantified over an unbounded stream with no N and no termination condition, so it could never be discharged. `spec.md` §E bounds it: a review window that opens at 50 qualifying rows and closes after the first 10 proposals reach the gate, with a ≥8-of-10 rejection threshold.

## §F. Milestones

Ordered by decision-reversibility. M1 carries the contract every later milestone binds to; M5 is mechanical.

### M1 — Types, thresholds, and generated fixtures (highest change-likelihood)

Authored before any analyzer logic, so the contract is reviewable on its own and L2 is testable without waiting for real data.

- New package `internal/harness/delegationmap/`.
- `types.go`: `SubcommandStat` (qualifying rows, per-agent counts, unattributed count, reroute/abort counts), `Finding` (kind, subcommand, agent, observation count, support ratio, qualifying rows, unattributed share), `Result` (findings, reason, malformed lines, evaluated subcommands).
- `Kind` enum: `undesignated_agent`, `designated_never_spawned` — and no third value (REQ-HLA-007).
- Declared constants with stated rationale: `MinQualifyingRows = 5`, `MinSupportRatio = 0.60` (mirroring the `proposalgen` precedent of a single declared constant rather than config-driven tuning), `MaxLedgerBytes` (§E D1), the retained-catalog membership set (§E D2), and the conditional-invocation exclusion set with a rule citation per entry (REQ-HLA-008).
- `testdata/` fixtures **generated** by a Go generator that builds `routing.PendingRow` values and calls `Finalize()` (B4, AC-HLA-001): below-threshold, at-threshold, undesignated-agent, designated-never-spawned, conditional-designations, non-catalog-agents, unattributed-share, reroute/abort-only, malformed-line, oversized, and empty.

**Why first:** every later milestone binds to these types and constants, and the two contested decisions (§E D1, D2) are both encoded here.

### M2 — Aggregation and identity discrimination

- `aggregate.go`: consume rows via `routing.NewReader(...).Read(routing.Filter{})`, fold into `SubcommandStat`, retaining no per-row state (REQ-HLA-001/002/003).
- Apply the size guard before the read; on exceed, return the empty result with the size reason.
- Classify each observed `agent` value into catalog / non-catalog / unattributed (REQ-HLA-004/005). The classification is total — every value lands in exactly one bucket, and none is dropped.

### M3 — Map reader and finding production

- `mapreader.go`: read-only YAML load of `.moai/config/sections/delegation.yaml` into `map[string][]string` (subcommand → designated agents). Read path only (REQ-HLA-009).
- `analyze.go`: apply thresholds, apply the conditional-invocation exclusion, produce `Finding`s (REQ-HLA-006/007/008).

### M4 — Proposal emission

- `proposal.go`: build `proposalgen.ProposalCandidate` values with `PatternKey = "delegation_map:<subcommand>:<sha256(kind|subcommand|agent)[:8]>"` and `DraftID = "PROPOSAL-" + sha256(patternKey)[:8]`, then emit via `proposalgen.WriteProposals(proposalgen.ProposalDir(root), candidates)` (REQ-HLA-010/011/012/013).
- Body rendering carries the observation count, support ratio, qualifying-row count, unattributed share, and kind, plus the explicit statement that application requires the Tier-4 approval gate.

**Namespace note:** `delegation_map` is not a member of `PatternBearingEventTypes()`, so `proposalgen.MapPromotions`'s format regex rejects it by construction — which is the desired isolation, and the reason the analyzer bypasses the mapper rather than extending it (B1).

### M5 — CLI surface

- `moai harness delegation analyze [--json] [--dry-run] [--limit N] [--ledger PATH] [--map PATH]` in a new `internal/cli/harness_delegation.go`, registered in the existing `AddCommand` block of `internal/cli/harness_route.go` alongside `newHarnessLedgerCmd()` (REQ-HLA-015).
- Boundary grep test mirroring the routing package's `subagent_boundary_test.go`.

## §G. Anti-patterns

- **AP-1 — event-type widening.** Adding `delegation_map` to `PatternBearingEventTypes()` to make the existing mapper accept the new keys (B1).
- **AP-2 — naive identity.** Emitting an `undesignated_agent` finding for any observed value that is not in the map, without the catalog-membership test (B2).
- **AP-3 — live-data AC.** Writing an acceptance criterion that reads `.moai/state/routing-ledger.jsonl` and asserts content. The file is gitignored runtime state, gate-dependent, and empty in CI.
- **AP-4 — hand-written fixtures.** Authoring `testdata/*.jsonl` by hand instead of generating through `PendingRow.Finalize()` (B4).
- **AP-5 — silent map write.** Any code path that opens `delegation.yaml` for writing, including a "harmless" comment or formatting rewrite (B5).
- **AP-6 — exclusion creep.** Adding an entry to the conditional-invocation exclusion set without a rule citation, turning a justified carve-out into an allowlist that suppresses real findings (residual risk R3).
- **AP-7 — circular membership.** Deriving the retained-agent catalog from the delegation map, which makes an undesignated agent undiscoverable by construction (§E D2).

## §H. Deferred / follow-up (not in this SPEC)

- Ledger schema v2 with an injected-skills field, which would unlock skill-level proposals.
- Threshold tuning driven by real data once the `spec.md` §E review window has run.
- A streaming variant of `routing.Reader`, correctly scoped as a change to the producer's package (§E D1).
- Wiring analyzer findings into `moai harness propose` as a second source, rather than a separate verb.

## §I. File inventory (run phase)

| Path | Change |
|---|---|
| `internal/harness/delegationmap/types.go` | new — stats, findings, kinds, declared constants |
| `internal/harness/delegationmap/aggregate.go` | new |
| `internal/harness/delegationmap/mapreader.go` | new — the only `delegation.yaml` reference in the tree |
| `internal/harness/delegationmap/analyze.go` | new |
| `internal/harness/delegationmap/proposal.go` | new |
| `internal/harness/delegationmap/*_test.go` | new |
| `internal/harness/delegationmap/subagent_boundary_test.go` | new — mirrors the routing package's boundary grep |
| `internal/harness/delegationmap/fixturegen_test.go` | new — generates `testdata/` through `PendingRow.Finalize()` |
| `internal/harness/delegationmap/testdata/*.jsonl` | new fixtures (generated) |
| `internal/cli/harness_delegation.go` | new — CLI verb |
| `internal/cli/harness_route.go` | edit — one `AddCommand` line |

Roughly 11 files, of which 1 is an edit. No template files, no `.claude/` files.

## §J. Tier classification

**Tier M.** Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, the scope-based table places Tier M at 300-1000 LOC and 5-15 files affected; this SPEC's inventory is ~11 files, single-domain (Go, one new package plus one registration line), and well under 1000 LOC. The requirement and acceptance-criterion counts (16 and 16) sit exactly at the Tier M ceiling, which is the binding constraint that motivated the split from the original 33/36 SPEC. Tier M carries the 3-artifact set (spec.md + plan.md + acceptance.md) plus `progress.md`, and a plan-auditor PASS threshold of 0.80.

## §K. Cross-references

- `SPEC-HARNESS-LEARNING-EVO-001` — the L1 producer of the rows this analyzer reads (sibling, not a dependency; see `spec.md` §F)
- `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §2.5, §5 row P3
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (the budget that forced the split), § Plan Audit Gate skip policy (the `plan-auditor` conditional-invocation citation), § Mode Dispatch (the `sync-auditor` harness-level citation)
- `.claude/rules/moai/workflow/skill-routing.md` — the consumer of `delegation.yaml`
- `.claude/rules/moai/core/verification-claim-integrity.md` — the `spec.md` §E falsification discipline
- `CLAUDE.md` §4 — the retained agent catalog seeding the membership constant (§E D2)
