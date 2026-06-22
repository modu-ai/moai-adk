# Implementation Plan — SPEC-DIVECC-OBSERVABILITY-LOOP-001

> Tier M. Derived from `spec.md` §D requirements. Priority-ordered milestones; no time estimates (per `agent-common-protocol.md` § Time Estimation).

## §A. Context

Candidate N4 of Epic Dive-into-CC. Fills the single MISSING step (`cluster`) of the harness-learning loop `trace → eval → cluster → policy → repair`. Delivers a NEW read-only clustering engine (`internal/harness/cluster/`) + an observability surface (a report artifact + `moai harness clusters` CLI read surface). The premise is VERIFIED (`spec.md` §B) — run-phase re-confirms the field schema is current, then implements the deterministic clusterer. The scope is deliberately conservative: read-only aggregation + surfacing ONLY (user decisions Q1/Q2).

### §A.5 PRESERVE list (untouched — scope discipline)

The following are READ as inputs or PRESERVED entirely; the run-phase MUST NOT modify them:

- `internal/harness/proposalgen/` (`mapper.go` — the 1:1 promotion→candidate logic). PRESERVE.
- `internal/harness/applier.go` (`Apply` / `Rollback`). PRESERVE.
- `internal/harness/safety/` (L1-L5 decision logic). PRESERVE.
- `internal/harness/regression_gate.go` (`MetricTriple` is read as input; the gate logic is untouched). PRESERVE.
- the `autoApply` default (false, REQ-HL-005) — PRESERVE; no auto-apply path added.
- `internal/evolution/learning.go` (separate dormant surface — clean boundary). PRESERVE.
- the 56 manual `feedback_*.md` lessons — NOT migrated. PRESERVE.
- runtime files `.moai/harness/usage-log.jsonl`, `manifest.jsonl`, `tier-promotions.jsonl` — READ-ONLY inputs (B8).

EXTEND targets (the only new/changed surfaces): `internal/harness/cluster/` (NEW package), the harness CLI (NEW `clusters` subcommand). The `clusters` subcommand factory (`newHarnessClustersCmd`) is registered in the LIVE harness command tree — `newHarnessRouterCmd()` in `internal/cli/harness_route.go` (registered in `rootCmd` at `internal/cli/root.go:104`) — alongside `status` (`harness_route.go:99`). The shared `newHarnessStatusCmd` / `resolveProjectRoot` helpers it mirrors live in `internal/cli/harness.go`. Do NOT register `clusters` under `newHarnessCmd` (`harness.go:63`), which is a deprecation-marker tree NOT wired to `rootCmd`.

## §B. Known Issues / Constraints from research

- **SPEC-ID validity** (run-phase note, pre-empts a future false-flag): the multi-segment id `SPEC-DIVECC-OBSERVABILITY-LOOP-001` is VALID per the ENFORCED lint regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (`internal/spec/lint.go:578`), which admits one-or-more domain segments. The stale schema-prose regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` in `spec-frontmatter-schema.md` would APPEAR to reject the multi-segment id, but the enforced lint is authoritative. Do NOT re-flag this ID (same precedent as N1 — see `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/plan.md` §B).
- **B1 cross-platform** (`manager-develop-prompt-template.md` B1): pure-Go clustering, no syscall. Confirm `GOOS=windows GOARCH=amd64 go build ./...` stays green (run-phase AC-OBL-007).
- **B3/B11 subagent boundary** (C-HRA-008): `internal/harness/cluster/` MUST have NO `AskUserQuestion`/`mcp__askuser` calls. Add a `cluster/subagent_boundary_test.go` grep gate (AC-OBL-006). The CLI subcommand likewise — `internal/cli/CLAUDE.md` mandates this.
- **B4 canonical frontmatter**: 12 fields + `era: V3R6` + `tier: M` + `depends_on: [N1]` (already applied in spec.md).
- **B6 spec-lint heading**: Out of Scope is authored as `### Out of Scope — <topic>` H3 sub-sections (NOT a bare `## Out of Scope` h2) to avoid the `MissingExclusions` ERROR.
- **B7 path resolution (ALIGN with the existing convention)**: the CLI read surface resolves the input/report paths by REUSING the shared `resolveProjectRoot(cmd)` helper (`internal/cli/harness.go:95-111`) — the SAME helper `runHarnessStatus` uses (`harness.go:129`). That helper reads `--project-root` (flag + inherited flag) and falls back to `os.Getwd()` when the flag is empty (`harness.go:103-105`). REQ-OBL-011 requires REUSE of this helper and NO new divergent path-resolution path; it does NOT require a `$CLAUDE_PROJECT_DIR`-only path (verified: no harness CLI command reads `$CLAUDE_PROJECT_DIR` — `grep -rn 'CLAUDE_PROJECT_DIR\|EnvClaudeProjectDir' internal/cli/harness*.go` → 0; the `EnvClaudeProjectDir` constant at `internal/config/envkeys.go:80` is wired into hook/session/pre-push, NOT harness). Requiring a bespoke `$CLAUDE_PROJECT_DIR` path here would contradict the codebase, so the conservative fix is to mirror `runHarnessStatus` exactly.
- **B8/B10 working-tree hygiene + PRESERVE**: the clusterer writes ONLY its own report under `.moai/harness/learning-history/`; the runtime input files are read-only. Do not `git add` unrelated untracked files.
- **Determinism is the verification lever** (REQ-OBL-007): the signature key must be derived from a sorted dimension set + stable fields, and cluster output must be sorted by a stable key, so table-driven `t.TempDir` fixtures produce byte-identical output. Any map-iteration-order leak into output would break this — sort before emit.
- **Schema currency**: `LogSchemaVersion = "v2.1"` and the `Outcome*` omitempty fields drop genuine zeros — a `kept` Δ=0 outcome is identifiable by `outcome_verdict == "kept"` with triple fields absent (`types.go:147-149`). The clusterer keys on `Verdict` + `Regressed`, both present on the events it clusters, so the omitempty drop does not affect signature derivation.

## §C. Pre-flight (run-phase, before writing code)

```bash
# 1. branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. cross-platform build baseline (must stay green after the new package lands)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. confirm the apply_outcome field schema is current (the clusterer's substrate)
grep -n 'EventTypeApplyOutcome\|OutcomeVerdict\|OutcomeRegressed\|OutcomeProposalID' internal/harness/types.go

# 5. confirm PRESERVE targets exist + are untouched
ls internal/harness/proposalgen/ internal/harness/applier.go internal/harness/regression_gate.go internal/evolution/learning.go

# 6. confirm the read-only CLI aggregator pattern to mirror
grep -n 'newHarnessStatusCmd\|resolveProjectRoot\|AggregatePatterns' internal/cli/harness.go

# 7. confirm the LIVE registration site (clusters registers HERE, alongside status)
grep -n 'newHarnessStatusCmd\|AddCommand' internal/cli/harness_route.go   # newHarnessRouterCmd is the live tree
grep -n 'newHarnessRouterCmd' internal/cli/root.go                        # registered in rootCmd
grep -n 'AddCommand(newHarnessCmd' internal/cli/*.go || echo "newHarnessCmd NOT in rootCmd (deprecation marker — do NOT register clusters there)"
```

## §D. Constraints

- **Read-only boundary (REQ-OBL-013/014)**: the clusterer reads JSONL inputs and writes ONLY its own report file. ZERO modifications to `applier.go`, `proposalgen/`, `regression_gate.go`, the entire `internal/harness/safety/` directory (L1-L5), the `autoApply` default (`harness.go:484`), or `evolution/learning.go`. Verified by `git diff --stat` over the FULL REQ-OBL-014 surface (including `internal/harness/safety/`) showing changes confined to `internal/harness/cluster/` + the harness CLI `clusters` registration (`internal/cli/harness_route.go` + the shared factory in `internal/cli/harness.go`) (+ tests), plus a positive autoApply-default-unchanged check (AC-OBL-005).
- **Subagent boundary (REQ-OBL-015)**: no `AskUserQuestion`/`mcp__askuser` in `internal/harness/cluster/` or the CLI subcommand. CI guard test mandatory.
- **Determinism (REQ-OBL-007)**: no ML, no randomness, stable sort before emit.
- **No auto-apply (Q2)**: the engine never triggers an apply, never writes a proposal, never changes the `autoApply` default. Observation/surfacing ONLY.
- **TRUST 5 + coverage ≥ 85%** for the new `cluster` package (run-phase AC-OBL-008).
- **Conventional Commits**, `--no-verify` prohibited, `🗿 MoAI` trailer.

## §E. Self-Verification (run-phase deliverables map)

| REQ | Verified by |
|-----|-------------|
| REQ-OBL-001/002/003/004 | ingestion tests (AC-OBL-010, table-driven `TestLoadEvents`): `apply_outcome` parsed; manifest optional; malformed line skipped (fail-open); empty/absent log → 0 clusters + success |
| REQ-OBL-005/006/007/008 | clustering tests: deterministic signature key from `outcome_regressed`+`outcome_verdict`+`outcome_decision` (NOT `pattern_key`); grouping into `FailureCluster`; byte-identical repeated-run output; `kept` excluded |
| REQ-OBL-009 | report-artifact test: deterministic report emitted under `.moai/harness/learning-history/` |
| REQ-OBL-010/011/012 | CLI tests: `moai harness clusters --help` exits 0 (live-tree wiring via `newHarnessRouterCmd`); prints clusters; `--json` machine-readable; reuses `resolveProjectRoot` (no new divergent path); empty → exit 0 |
| REQ-OBL-013/014 | `git diff --stat` (FULL surface incl. `internal/harness/safety/`) confined to `cluster/` + harness CLI `clusters` registration (+ tests); applier/proposalgen/regression_gate/safety/autoApply-default/evolution unchanged + positive autoApply-default-unchanged check |
| REQ-OBL-015 | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/cluster/ \| grep -v _test.go \| grep -v "//"` → 0 |

## §F. Milestones (priority-ordered)

- **M1 — Confirm schema + scaffold the package**: re-run §C pre-flight; create `internal/harness/cluster/` package skeleton with the `FailureCluster` type and an ingestion entry point (`LoadEvents` / `ClusterEvents`). RED tests for ingestion (REQ-OBL-001/002/003/004).
- **M2 — Deterministic signature extraction**: implement the signature-key derivation (sorted `outcome_regressed` dimension set + `outcome_verdict` + `outcome_decision` — NOT `pattern_key`, per REQ-OBL-005) and the grouping into `FailureCluster`. RED→GREEN tests for determinism + `kept` exclusion (REQ-OBL-005/006/007/008).
- **M3 — Report artifact**: emit the deterministic report under `.moai/harness/learning-history/`; test that the report is byte-identical across runs (REQ-OBL-009).
- **M4 — CLI read surface**: register the `clusters` subcommand factory (`newHarnessClustersCmd`) in the LIVE harness tree — `newHarnessRouterCmd()` in `internal/cli/harness_route.go` (registered in `rootCmd` at `root.go:104`), alongside `status` (`harness_route.go:99`) — mirroring `newHarnessStatusCmd`. Reuse the shared `resolveProjectRoot(cmd)` helper (no new divergent path); add `--json` flag; empty-set exit 0 (REQ-OBL-010/011/012). Do NOT register under the unwired `newHarnessCmd` (`harness.go:63`). CLI tests including `moai harness clusters --help` exits 0 (proves live-tree wiring, AC-OBL-004).
- **M5 — Boundary + coverage hardening**: add the `cluster/subagent_boundary_test.go` grep gate (REQ-OBL-015); verify read-only boundary via `git diff --stat` (REQ-OBL-013/014); raise package coverage ≥ 85%; cross-platform build (AC-OBL-007).
- **M6 — Self-verify + close**: run the §E verification matrix; confirm the PRESERVE list is untouched; commit + push per Hybrid Trunk Tier M.

## §G. Anti-Patterns to avoid

- AP — letting the clusterer write back into the proposal/apply path (violates the read-only boundary; the WHOLE point of N4's conservative scope).
- AP — adding root-cause proposal generation or N:M mapping "while we're here" (this was the REJECTED Tier-L alternative — `spec.md` §C).
- AP — introducing map-iteration-order or timestamp-now into cluster output (breaks determinism REQ-OBL-007; sort before emit).
- AP — touching `evolution/learning.go` or the 56 `feedback_*.md` lessons (out of scope; clean boundary).
- AP — adding an auto-apply path or flipping the `autoApply` default (violates Q2 / REQ-OBL-014).
- AP — introducing a NEW, bespoke path-resolution path in the `clusters` handler instead of reusing the shared `resolveProjectRoot(cmd)` helper (B7/REQ-OBL-011 — mirror `runHarnessStatus` exactly; do NOT invent a `$CLAUDE_PROJECT_DIR`-only path the rest of the harness CLI does not use).
- AP — registering `clusters` under the deprecation-marker `newHarnessCmd` (`harness.go:63`) instead of the live `newHarnessRouterCmd` (`harness_route.go`); the former is NOT wired to `rootCmd`, so `moai harness clusters --help` would not exist (D5 / AC-OBL-004).

## §H. Cross-References

- `spec.md` §B — the verified field-schema evidence the clusterer ingests.
- `acceptance.md` — Given-When-Then ACs.
- `internal/cli/harness.go:117-183` — the `harness status` read-only aggregator the CLI surface mirrors.
- `internal/cli/CLAUDE.md` — subagent boundary + stream discipline + `$CLAUDE_PROJECT_DIR` resolution.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — B-series known issues (B1/B3/B7/B8/B10/B11).
