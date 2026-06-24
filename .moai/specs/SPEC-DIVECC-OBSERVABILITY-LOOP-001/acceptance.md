# Acceptance Criteria — SPEC-DIVECC-OBSERVABILITY-LOOP-001

> GEARS-format acceptance criteria with Given-When-Then scenarios. Every AC is observable (test output, grep output, file existence, `git diff`). Per `verification-claim-integrity.md`, each AC names the evidence that proves it. Companion to `spec.md` (the canonical lint target) + `plan.md`. Traceability to REQ-OBL-XXX in each AC heading.

## §A. Definition of Done

The SPEC is done when: a deterministic read-only clustering engine exists in `internal/harness/cluster/` that ingests `apply_outcome` events and groups them by a deterministic signature key into `FailureCluster`s; a report artifact is emitted under `.moai/harness/learning-history/`; a `moai harness clusters` CLI read surface prints the clusters (with `--json`); the engine writes ONLY its own report (zero diff to applier/proposalgen/safety/autoApply/evolution); the subagent-boundary grep is clean; cross-platform build is green; and the new package coverage is ≥ 85%.

## §B. AC Matrix

### AC-OBL-001 — Ingestion + deterministic grouping (covers REQ-OBL-001, REQ-OBL-005, REQ-OBL-006)

- **Given** a `usage-log.jsonl` fixture containing several `apply_outcome` events (some `rolled-back` with overlapping `outcome_regressed` dimension sets),
- **When** the clusterer ingests the log and groups events,
- **Then** events sharing an identical signature key (sorted `outcome_regressed` dimension set + `outcome_verdict` + `outcome_decision` — all fields present on the `apply_outcome` event per `types.go:151-176`; NOT `pattern_key`, which is absent from the event and one-way hashed into `outcome_proposal_id` per `mapper.go:97,102-107`) **shall** be grouped into one `FailureCluster` carrying the signature, member count, member event refs, representative regressed dimensions, and first/last-seen timestamps.
- **Evidence**: `go test -run TestClusterEvents ./internal/harness/cluster/` PASS; the test asserts that two `rolled-back` events with `outcome_regressed: ["coverage"]` and the same `outcome_verdict`/`outcome_decision` land in one cluster with `count == 2`, and an event with `outcome_regressed: ["lint"]` lands in a separate cluster. The test also asserts the signature key derivation reads NO `pattern_key` field (the fixture events carry none).

### AC-OBL-002 — Determinism (byte-identical repeated runs) (covers REQ-OBL-007)

- **Given** a fixed `usage-log.jsonl` fixture,
- **When** the clusterer runs twice over the same input,
- **Then** the two cluster outputs **shall** be byte-identical (stable sort by signature key; no ML, no randomness, no `time.Now()` leak into output).
- **Evidence**: `go test -run TestClusterDeterministic ./internal/harness/cluster/` PASS; the test runs `ClusterEvents` twice and asserts `reflect.DeepEqual` (or byte-equal serialized output). A grep for `math/rand`/`time.Now()` in the signature/emit path returns no use in output ordering.

### AC-OBL-003 — `kept` outcomes excluded (covers REQ-OBL-008)

- **Given** a `usage-log.jsonl` fixture mixing `outcome_verdict: "kept"` and `outcome_verdict: "rolled-back"` events,
- **When** the clusterer clusters the log,
- **Then** no `kept` event **shall** appear in any failure cluster — only failure/`rolled-back` signatures cluster.
- **Evidence**: `go test -run TestClusterExcludesKept ./internal/harness/cluster/` PASS; the test asserts the total member count across all clusters equals the count of `rolled-back` events (kept events contribute zero).

### AC-OBL-004 — CLI read surface (covers REQ-OBL-010, REQ-OBL-011, REQ-OBL-012)

- **Given** the `moai harness clusters` subcommand is registered in the LIVE harness command tree (`newHarnessRouterCmd`, `harness_route.go`, registered in `rootCmd` at `root.go:104` — NOT the deprecation-marker `newHarnessCmd`),
- **When** a user runs `moai harness clusters --help`, `moai harness clusters` (text), and `moai harness clusters --json` (machine-readable),
- **Then** `moai harness clusters --help` **shall** exit 0 (proving the command is wired into the live `rootCmd` tree); the command **shall** resolve the project root by REUSING the shared `resolveProjectRoot(cmd)` helper — the SAME helper `runHarnessStatus` uses — introducing NO new divergent path-resolution; it **shall** compute and print the clusters; and it **shall** report an empty result with exit 0 when there are zero clusters.
- **Evidence**:
  - `moai harness clusters --help` exits 0 (the command is registered in the live `rootCmd` tree via `newHarnessRouterCmd`; verified the registration site is `harness_route.go`, NOT the unwired `newHarnessCmd` in `harness.go`).
  - `go test -run TestHarnessClustersCmd ./internal/cli/...` PASS — asserts text output lists clusters and `--json` emits valid JSON on stdout.
  - `go test -run TestHarnessClustersEmpty ./internal/cli/...` PASS — empty/absent log → empty result, exit 0 (not error).
  - **Positive root-resolution assertion (NOT a vacuous "no os.Getwd" grep)**: `grep -n 'resolveProjectRoot' internal/cli/harness.go` shows the `clusters` handler calls the SAME `resolveProjectRoot(cmd)` as `runHarnessStatus` (`harness.go:129`); the `clusters` handler defines NO bespoke root-resolution function of its own (i.e. the clusters path adds zero new `func` that resolves a project root — it reuses the shared helper verbatim). Note: `resolveProjectRoot` itself legitimately falls back to `os.Getwd()` (`harness.go:103-105`), so a "grep shows no `os.Getwd`" check would pass vacuously and is deliberately NOT used as the evidence here; the evidence is the positive reuse of the shared helper.

### AC-OBL-005 — Read-only boundary (zero diff to the apply/proposal path) (covers REQ-OBL-013, REQ-OBL-014)

- **Given** the clusterer's deliverable is read-only aggregation + surfacing,
- **When** the run-phase changes are diffed,
- **Then** `git diff --stat` **shall** be confined to `internal/harness/cluster/` + the `clusters` registration in the harness CLI (`internal/cli/harness.go` and/or `internal/cli/harness_route.go`) (+ their tests), and **shall** show ZERO changes to `internal/harness/applier.go`, `internal/harness/proposalgen/`, `internal/harness/regression_gate.go`, the entire `internal/harness/safety/` directory (the L1-L5 decision logic), the `autoApply` default, and `internal/evolution/learning.go`.
- **Evidence**:
  ```bash
  # Full REQ-OBL-014 surface — every PRESERVE target, including the safety/ directory:
  git diff --stat internal/harness/applier.go internal/harness/proposalgen/ \
                  internal/harness/regression_gate.go internal/harness/safety/ \
                  internal/evolution/learning.go
  # Expected: empty (zero changed files)

  # Positive autoApply-default-unchanged assertion: the default is `AutoApply: false`
  # at internal/cli/harness.go:484 (defaultLearningConfig). Confirm it is untouched
  # AND that the cluster package never writes/flips it:
  git diff internal/cli/harness.go | grep -E '^\+.*AutoApply'   # Expected: no added line touching AutoApply
  grep -rn 'AutoApply' internal/harness/cluster/                # Expected: no match (cluster never references the autoApply default)
  ```

### AC-OBL-006 — Subagent boundary (covers REQ-OBL-015)

- **Given** `internal/harness/cluster/` runs in subagent context,
- **When** the boundary grep gate runs,
- **Then** the package **shall** contain no `AskUserQuestion` / `mcp__askuser` call.
- **Evidence**:
  ```bash
  grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/cluster/ \
    | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[ \t]*//'
  # Expected: no output
  ```
  Plus a CI guard test `internal/harness/cluster/subagent_boundary_test.go` (mirroring `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion`) PASS.

### AC-OBL-007 — Cross-platform build (covers plan §B B1)

- **Given** the clusterer is pure Go (no syscall),
- **When** the cross-platform build runs,
- **Then** both host and Windows cross-builds **shall** exit 0.
- **Evidence**:
  ```bash
  go build ./...                          # exit 0
  GOOS=windows GOARCH=amd64 go build ./... # exit 0
  ```

### AC-OBL-008 — Coverage ≥ 85% on the new package (covers TRUST 5)

- **Given** the new `internal/harness/cluster/` package,
- **When** coverage is measured,
- **Then** statement coverage **shall** be ≥ 85%.
- **Evidence**: `go test -cover ./internal/harness/cluster/...` reports `coverage: ≥85.0% of statements`.

### AC-OBL-009 — Report artifact emitted deterministically (covers REQ-OBL-009)

- **Given** a non-empty cluster set,
- **When** the clusterer emits its report,
- **Then** a deterministic report artifact **shall** be written under `.moai/harness/learning-history/`, byte-identical across repeated runs over the same input, enumerating each cluster (signature, count, representative dimensions, first/last seen).
- **Evidence**: `go test -run TestClusterReport ./internal/harness/cluster/` PASS — writes the report to a `t.TempDir()` learning-history path, runs twice, and asserts byte-equal report content. The report path is under `learning-history/` and is the ONLY file the clusterer writes (per AC-OBL-005).

### AC-OBL-010 — Ingestion edge cases (fail-open / optional manifest / empty log) (covers REQ-OBL-002, REQ-OBL-003, REQ-OBL-004)

- **Given** a table-driven set of `usage-log.jsonl` fixtures exercising the three ingestion edge cases,
- **When** the clusterer's `LoadEvents` (ingestion entry point) processes each fixture,
- **Then**:
  - a fixture with one malformed JSONL line interleaved among valid `apply_outcome` lines **shall** skip the malformed line and ingest the remaining valid lines without aborting (fail-open — REQ-OBL-003, EC-2);
  - a fixture whose `usage-log.jsonl` is present but whose optional `manifest.jsonl` is absent **shall** ingest from `usage-log.jsonl` alone without error (REQ-OBL-002, EC-3);
  - a fixture whose `usage-log.jsonl` is absent or empty **shall** produce zero clusters and a successful (non-error) result — NOT an error (REQ-OBL-004, EC-1).
- **Evidence**: `go test -run TestLoadEvents ./internal/harness/cluster/` PASS — a table-driven test (`t.TempDir()` fixtures, one sub-case per edge above) asserting: (a) the malformed-line case returns the count of valid events with `err == nil`; (b) the manifest-absent case returns the `usage-log.jsonl` events with `err == nil`; (c) the absent/empty-log case returns zero events + zero clusters + `err == nil` (mirroring the AC-OBL-001..009 Given-When-Then + reproducible `go test -run` style).

## §C. Edge cases

- **EC-1 — empty / absent log**: `usage-log.jsonl` absent or empty → zero clusters, exit 0, NOT an error (REQ-OBL-004 / REQ-OBL-012). Bound by AC-OBL-004 (empty case).
- **EC-2 — malformed JSONL line**: a corrupt or non-`apply_outcome` line is skipped; the run continues over the remaining valid lines (REQ-OBL-003). The clusterer never aborts on one bad line.
- **EC-3 — manifest.jsonl absent**: the optional `manifest.jsonl` supplementary input is absent → the clusterer proceeds on `usage-log.jsonl` alone without error (REQ-OBL-002).
- **EC-4 — all-`kept` log**: a log containing only `kept` outcomes → zero failure clusters (REQ-OBL-008). The total member count is zero; the report is the empty-cluster report.
- **EC-5 — single-member cluster**: a unique failure signature with one member event still forms a valid `FailureCluster` with `count == 1` (no minimum-size threshold; clustering is grouping, not filtering-by-frequency).

## §D. Quality gate

- No unobserved claim: every AC names a reproducible command + expected output (AC-OBL-001..010).
- Read-only boundary: zero diff to applier/proposalgen/safety/autoApply/evolution (AC-OBL-005) — the conservative-scope invariant.
- Subagent boundary: `internal/harness/cluster/` has no `AskUserQuestion` (AC-OBL-006).
- Determinism: byte-identical repeated-run output (AC-OBL-002, AC-OBL-009) — the verification lever for an ML-free engine.
- Coverage: ≥ 85% on the new package (AC-OBL-008).
- Cross-platform: host + Windows build green (AC-OBL-007).
