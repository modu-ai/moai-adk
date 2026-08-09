# SPEC-HARNESS-LEARNING-EVO-002 — Acceptance Criteria (L2)

Every criterion below names the command that decides it, and that command actually decides it. Budget: 16 criteria (Tier M ceiling 16).

## §A. Verification conventions

- **Go-test criteria** are decided by the named test function's pass/fail, run with `-count=1`.
- **Grep criteria** are decided by the exact command's match count; a criterion asserting absence names the expected empty output.
- All analyzer criteria assert against synthetic fixtures under `internal/harness/delegationmap/testdata/`, never against live ledger content — the live ledger is gitignored runtime state, is gate-dependent, and is empty in CI. See `plan.md` §G AP-3.
- Fixtures are **generated**, not hand-written. AC-HLA-001 pins the generation path; a hand-authored JSONL fixture would silently drift from the row shape the producer actually emits.

## §B. Fixture integrity and reading

### AC-HLA-001 — fixtures are produced by the real finalize transform
- **Given** a fixture generator that constructs `routing.PendingRow` values and calls `PendingRow.Finalize(outcome)` to obtain each row
- **When** the generator writes every fixture under `testdata/` and the test re-runs it
- **Then** the committed fixtures are byte-identical to the regenerated output, and at least one fixture in the suite is produced through `Finalize()` rather than hand-authored JSON — so a dropped or mis-mapped field in the 15-field-plus-outcome transform fails this criterion instead of passing invisibly
- **Command**: `go test ./internal/harness/delegationmap/ -run TestFixtures_GeneratedViaFinalize -count=1`
- **Covers**: REQ-HLA-002 (row-shape fidelity)

### AC-HLA-002 — the analyzer reads through the routing reader, and refuses an oversized ledger
- **Given** the `delegationmap` package source, and separately a fixture ledger whose size exceeds the declared maximum
- **When** the package is searched for an independently-declared ledger path literal, and the analyzer is run against the oversized fixture
- **Then** the literal `routing-ledger.jsonl` appears nowhere in the package outside test fixtures, and the oversized run returns an empty finding list with a machine-readable size reason and no error
- **Command**: `grep -rn 'routing-ledger.jsonl' internal/harness/delegationmap/ --include='*.go' | grep -v _test` → expected: no output; and `go test ./internal/harness/delegationmap/ -run TestAnalyze_OversizedLedgerRefused -count=1`
- **Covers**: REQ-HLA-001

### AC-HLA-003 — per-subcommand aggregation is correct
- **Given** the fixture `aggregate_basic.jsonl` with 8 `plan` rows in which `manager-spec` appears 8 times and `Explore` 3 times
- **When** the aggregator runs
- **Then** the `plan` stat reports 8 qualifying rows, `manager-spec` count 8, and `Explore` count 3
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAggregate_Basic -count=1`
- **Covers**: REQ-HLA-002

### AC-HLA-004 — reroute and abort rows are excluded and reported separately
- **Given** the fixture `reroute_abort_only.jsonl` containing only `reroute` and `abort` rows
- **When** the analyzer runs
- **Then** qualifying rows is 0 for every subcommand, no finding is produced, and the reroute and abort counts appear in the result
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_RerouteAbortExcluded -count=1`
- **Covers**: REQ-HLA-003

## §C. Identity discrimination

### AC-HLA-005 — a non-catalog agent value never produces an undesignated-agent finding
- **Given** the fixture `non_catalog_agents.jsonl` with 10 qualifying `run` rows in which `general-purpose`, `workflow-subagent`, `hns-oss-docs-locale-translator-specialist`, and the spawn name `audit-hle` each appear in 9 of 10 rows (well clear of both thresholds)
- **When** the analyzer runs against a fixture delegation map that designates none of them
- **Then** no `undesignated_agent` finding names any of the four values, each is present in the result classified as non-catalog, and the run still emits findings for any retained-catalog agent that qualifies — proving the suppression is by classification and not by an empty result
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_NonCatalogNeverUndesignated -count=1`
- **Covers**: REQ-HLA-004

### AC-HLA-006 — unattributed delegations are counted and reported, never treated as an agent
- **Given** the fixture `unattributed_share.jsonl` with 10 qualifying `plan` rows in which 4 delegation entries carry the absent-identity marker defined by the producer SPEC
- **When** the analyzer runs
- **Then** the `plan` stat reports an unattributed share of 4 entries, the marker appears in no per-agent count and in no finding's agent field, and the unattributed share is present on every emitted proposal for `plan`
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_UnattributedShareReported -count=1`
- **Covers**: REQ-HLA-005

## §D. Thresholds and proposal kinds

### AC-HLA-007 — both thresholds suppress a proposal
- **Given** the fixture `below_min_rows.jsonl` with 4 qualifying `plan` rows against `MinQualifyingRows = 5`, and the fixture `stray_agent.jsonl` with 10 qualifying `run` rows in which an undesignated retained-catalog agent appears once (support 0.10, below `MinSupportRatio = 0.60`)
- **When** the analyzer runs against each
- **Then** the first yields an empty finding list with a result reason naming the row threshold, and the second yields no finding naming the stray agent
- **Command**: `go test ./internal/harness/delegationmap/ -run 'TestAnalyze_BelowMinRows|TestAnalyze_StrayRowSuppressed' -count=1`
- **Covers**: REQ-HLA-006

### AC-HLA-008 — both proposal kinds are produced, and only those two
- **Given** the fixture `two_kinds.jsonl` plus a fixture delegation map, arranged so one undesignated retained-catalog agent clears both thresholds for `run` and one designated agent that is not in the conditional-exclusion set is observed zero times for `sync`
- **When** the analyzer runs
- **Then** exactly two findings are produced, one of kind `undesignated_agent` and one of kind `designated_never_spawned`, and the `Kind` enum admits no third value
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_TwoKindsOnly -count=1`
- **Covers**: REQ-HLA-007

### AC-HLA-009 — conditionally-invoked designations are excluded from designated-never-spawned
- **Given** the fixture `conditional_designations.jsonl` with 12 qualifying `sync` rows in which `sync-auditor` never appears, and 12 qualifying `plan` rows in which `plan-auditor` never appears, against a fixture map designating both
- **When** the analyzer runs
- **Then** no `designated_never_spawned` finding names `sync-auditor` or `plan-auditor`, the exclusion set is a declared value carrying a rule citation per entry, and a third designated agent absent from the exclusion set and observed zero times in the same fixture still produces its finding
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_ConditionalDesignationsExcluded -count=1`
- **Covers**: REQ-HLA-008

## §E. Emission and isolation

### AC-HLA-010 — the delegation map is read, and never written
- **Given** a fixture `delegation.yaml` with known designated agents, made read-only for the duration of the test
- **When** the analyzer runs against it
- **Then** the designated set is resolved correctly, the analyzer succeeds, and the file's modification time and bytes are unchanged
- **Command**: `go test ./internal/harness/delegationmap/ -run TestMapReader_ReadOnly -count=1`
- **Covers**: REQ-HLA-009

### AC-HLA-011 — the pattern-key namespace is isolated, and the event-type SSOT is untouched
- **Given** the analyzer's emitted candidates for `two_kinds.jsonl`, and `internal/harness/types.go`
- **When** each `pattern_key` is tested against the existing proposalgen actionable-pattern regex, and the SSOT file is searched for the new namespace token
- **Then** every key starts with `delegation_map:` and none matches the existing regex; `delegation_map` does not appear in `internal/harness/types.go`; and `PatternBearingEventTypes()` returns the same members as before the change
- **Command**: `go test ./internal/harness/delegationmap/ -run TestPatternKey_NamespaceIsolated -count=1`; `grep -c 'delegation_map' internal/harness/types.go` → expected `0`; and `go test ./internal/harness/ -run TestPatternBearingEventTypes -count=1`
- **Covers**: REQ-HLA-010, REQ-HLA-011

### AC-HLA-012 — each proposal carries its evidence
- **Given** a written draft directory from the `two_kinds.jsonl` run
- **When** its `proposal.json` and `spec.md` are read
- **Then** both carry the observation count, the support ratio, the qualifying-row count for the subcommand, the unattributed share, and the proposal kind
- **Command**: `go test ./internal/harness/delegationmap/ -run TestProposal_CarriesEvidence -count=1`
- **Covers**: REQ-HLA-012

### AC-HLA-013 — the analyzer is deterministic and idempotent
- **Given** the `two_kinds.jsonl` fixture
- **When** the analyzer runs twice into the same output directory
- **Then** the candidate slices are equal element-for-element in the same order, and the second run leaves every written file byte-identical
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_DeterministicIdempotent -count=1`
- **Covers**: REQ-HLA-013

### AC-HLA-014 — absent, empty, and malformed inputs degrade gracefully
- **Given** three inputs — a non-existent path, an empty file, and `all_malformed.jsonl`
- **When** the analyzer runs on each
- **Then** each returns an empty finding list, a non-empty machine-readable reason, a malformed-line count matching the fixture, and no error
- **Command**: `go test ./internal/harness/delegationmap/ -run TestAnalyze_GracefulNoOp -count=1`
- **Covers**: REQ-HLA-014

### AC-HLA-015 — the CLI verb exposes dry-run and JSON
- **Given** a temp root with a fixture ledger and map
- **When** `moai harness delegation analyze --json --dry-run` runs
- **Then** it exits 0, emits a parseable JSON result on stdout, and creates no directory under the proposals path
- **Command**: `go test ./internal/cli/ -run TestHarnessDelegationAnalyzeCmd -count=1`
- **Covers**: REQ-HLA-015

### AC-HLA-016 — the boundary holds: no map write, no allowlist edit, no prompt, no skill proposal
- **Given** the whole non-test Go tree, the branch diff against its merge base, and every fixture in `testdata/`
- **When** each is inspected
- **Then** the only occurrences of `delegation.yaml` are in the read-only loader with no `os.WriteFile` / `os.Create` targeting it; `.claude/lsel/frozen-allowlist.json` is absent from the changed-file list and still contains `^\.moai/config/sections/`; `AskUserQuestion` appears in none of the new files; and no finding carries a skill field, the `Finding` type declaring none
- **Command**: `grep -rn 'delegation.yaml' internal/ --include='*.go' | grep -v _test` → expected: only `internal/harness/delegationmap/mapreader.go`; `git diff --name-only origin/main...HEAD | grep -c 'frozen-allowlist.json'` → expected `0`; `grep -c 'moai/config/sections' .claude/lsel/frozen-allowlist.json` → expected `≥ 1`; `grep -rn 'AskUserQuestion' internal/harness/delegationmap/ internal/cli/harness_delegation.go` → expected: no output; and `go test ./internal/harness/delegationmap/ -run TestAnalyze_NoSkillProposals -count=1`
- **Covers**: REQ-HLA-016

## §F. Quality gate

- `go build ./...` exits 0, and `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `go test ./... -count=1` exits 0 (full suite, not the affected packages alone).
- `golangci-lint run` reports no new finding against the pre-change baseline.
- `moai spec lint .moai/specs/SPEC-HARNESS-LEARNING-EVO-002/spec.md --strict` exits 0.
- `git diff --name-only origin/main...HEAD | grep -cE '^(internal/template/templates/|\.claude/skills/)'` returns `0`.

## §G. Definition of Done

- All 16 criteria pass, each cited with its command and observed output.
- The traceability matrix in §H shows every requirement covered.
- The falsification review named in `spec.md` §E is scheduled, not performed — its window opens at 50 qualifying rows, which do not exist at merge time.

## §H. Residual risk

- **R1 — thresholds are guesses.** `MinQualifyingRows = 5` and `MinSupportRatio = 0.60` are chosen by analogy to the existing `proposalgen` constant, not from data. They are the most likely thing to be wrong, and they can only be validated after the producer SPEC has generated real rows.
- **R2 — the catalog membership test is a snapshot.** REQ-HLA-004's discrimination depends on the retained-agent catalog (`CLAUDE.md` §4). A catalog change that this package does not track would silently reclassify observations. The membership source and its staleness exposure are recorded in `plan.md` §E D2.
- **R3 — the exclusion set can become a silent allowlist.** REQ-HLA-008's exclusion suppresses real findings by design. Each entry carries a rule citation for exactly this reason, but nothing mechanically re-checks that the cited rule still makes the invocation conditional.
- **R4 — the memory bound is a size guard, not streaming.** `routing.Reader.Read` materializes the whole filtered row set; adding a streaming variant would have made this SPEC depend on a change in its sibling, which the fixture-based strategy exists to avoid. REQ-HLA-001's declared maximum caps the exposure rather than removing it (`plan.md` §E D1).
- **R5 — spawn-name collisions inflate catalog counts.** A spawn *named* exactly like a retained agent is indistinguishable from a type spawn, so per-agent counts may include named spawns of a different underlying type. The direction of the error is toward over-counting a designated agent, which suppresses `designated_never_spawned` rather than fabricating `undesignated_agent` — the safer direction, but not a neutral one.
- **R6 — F3 is bounded but slow.** The falsifier needs 10 proposals through a human gate; until the window closes, the central assumption stays unfalsified rather than confirmed.

## §I. Traceability matrix

| REQ | Criteria |
|---|---|
| REQ-HLA-001 | AC-HLA-002 |
| REQ-HLA-002 | AC-HLA-001, AC-HLA-003 |
| REQ-HLA-003 | AC-HLA-004 |
| REQ-HLA-004 | AC-HLA-005 |
| REQ-HLA-005 | AC-HLA-006 |
| REQ-HLA-006 | AC-HLA-007 |
| REQ-HLA-007 | AC-HLA-008 |
| REQ-HLA-008 | AC-HLA-009 |
| REQ-HLA-009 | AC-HLA-010, AC-HLA-016 |
| REQ-HLA-010 | AC-HLA-011 |
| REQ-HLA-011 | AC-HLA-011 |
| REQ-HLA-012 | AC-HLA-012 |
| REQ-HLA-013 | AC-HLA-013 |
| REQ-HLA-014 | AC-HLA-014 |
| REQ-HLA-015 | AC-HLA-015 |
| REQ-HLA-016 | AC-HLA-016 |
