# Progress — SPEC-HARNESS-LOOP-CLOSURE-001

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: S
- scope (file count): 4 — `types.go` (EXTEND), `lineage.go` (NEW), `applier.go` (EXTEND Apply only), `lineage_test.go` (NEW)
- domain count: 1 (Go source, `internal/harness`)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy; the 4 files are tightly coupled around `applier.go` `Apply()`)
- Agent Teams prereqs: not met (harness level not `thorough` / team disabled)

**Mode evaluation**
- Mode 1 trivial — not selected (multi-file additive feature + tests)
- Mode 2 background — not selected (Write operations require foreground)
- Mode 3 agent-team — not selected (prereqs unmet + single-domain)
- Mode 4 parallel — not selected (coding-heavy, single coupled file set; Anthropic coding-task parallelism caveat)
- Mode 5 sub-agent — **selected** (coding-heavy single-domain default)
- Mode 6 workflow — not selected (< 30 files, not mechanical-uniform)

**Decision: sub-agent**

**Justification**: Coding-heavy, single-domain (`internal/harness`), 4 tightly-coupled files centered on the `applier.go` `Apply()` extension. Per Anthropic's coding-task parallelism caveat, sequential sub-agent (Mode 5) is the correct default. GATE-2 approved; plan-auditor iter-1 PASS-WITH-DEBT 0.83 (Tier S threshold 0.80) with all SHOULD-FIX (D1/D2/D3) + D6 defects patched orchestrator-direct prior to run-phase entry. Phase 0.5 gate already executed this session (not re-run; score 0.83 < 0.90 skip threshold is moot because the gate ran once and defects are patched).

## §Run-phase Evidence

cycle_type: tdd (RED-GREEN-REFACTOR). Tier S. Single-pass M1-M4 (no Round split).

### Design decision (manifest path injection)

Per plan.md §D.3, the manifest path is threaded through `Apply()` as an **Applier field**
(`manifestPath`, precedent: the existing `allowWrites` field) rather than a new parameter. This
keeps the 4-arg `Apply(proposal, evaluator, snapshotBase, sessions)` signature unchanged, so all
7 existing `applier_test.go` call-sites compile without edit. Lineage tests inject a `t.TempDir()`
manifest path via the new `newApplierWithManifest(manifestPath)` constructor. When `manifestPath`
is empty (the `NewApplier()` default), `Apply()` skips the lineage write — preserving pre-lineage
behavior. Ground-truth confirmed at run-start: ZERO production `.Apply(` call-sites exist (only
tests call Apply), so the field-vs-param choice is non-breaking either way; the field route was
chosen for minimal-diff scope discipline.

### E1 — 8-AC PASS/FAIL matrix

| AC | REQ | Status | Verification Command | Actual Output |
|----|-----|--------|----------------------|---------------|
| AC-HLC-001 | REQ-HLC-001/003 | PASS | `go test -run 'TestApply.*Lineage.*Approved' -v` + anti-vacuous guard | `--- PASS: TestApply_LineageApproved`; guard: >=1 PASS, no "no tests to run" (PASS-count 1); LoadManifest len==1, applied_surface=="description" |
| AC-HLC-002 | REQ-HLC-004/008 | PASS | `go test -run 'TestLineage.*Rejected\|TestApply.*Frozen.*Lineage'` | `ok` — one rejected entry, non-empty reason, no SKILL.md write, no snapshot dir; rejection error still returned |
| AC-HLC-003 | REQ-HLC-005/006 | PASS | `go test -run 'TestApply.*Pending\|TestLineage.*Pending.*NoEntry'` | `ok` — `*ApplyPendingError` returned, OversightPayload non-nil, LoadManifest len==0 |
| AC-HLC-004 | REQ-HLC-007 | PASS | `go test -run 'TestApply.*PreservesFrontmatterBody\|TestEnrichDescription' -v` + anti-vacuous guard | guard: >=1 PASS, no "no tests to run" (PASS-count 7); body byte-identical, non-description frontmatter preserved, description enriched |
| AC-HLC-005 | REQ-HLC-001/002 | PASS | `go test -run 'TestLineage.*AppendOnly\|TestLoadManifest'` | `--- PASS` x 3 (MissingFileReturnsEmpty, SkipsBlankLines, AppendOnly); two writes -> len==2 in order; missing file -> (empty, nil) |
| AC-HLC-006 | REQ-HLC-009 | PASS | `go test -run 'TestApply.*SnapshotCreatedFirstApply\|TestSnapshot.*Created'` | `--- PASS: TestApply_SnapshotCreatedFirstApply`; base dir created from absent, per-apply dir + manifest.json present, backup byte-for-byte == pre-apply |
| AC-HLC-007 | REQ-HLC-010 | PASS | `go test -run 'TestRestoreSnapshot.*ByteIdentical\|TestApply.*Rollback'` | `--- PASS` x 2 (RestoreSnapshot_RestoresByteIdentical, Apply_Rollback); post-restore bytes == pre-apply bytes |
| AC-HLC-008 | REQ-HLC-011/012 | PASS | full suite + boundary test + protected-artifact + template grep | full harness suite `ok` (all 8 packages); TestSubagentBoundary `ok`; protected artifacts unchanged; TEMPLATE_CLEAN |

### E2 — Cross-platform build

```
$ go build ./...                          -> exit 0
$ GOOS=windows GOARCH=amd64 go build ./... -> exit 0
```

### E3 — Coverage on lineage path (>=85% target)

```
$ go test -cover ./internal/harness/   -> coverage: 87.7% of statements (package total)
```

Function-level on the new lineage path:
- `lineage.go` `WriteLineageEntry` 83.3%, `LoadManifest` 85.0%
- `applier.go` `writeLineage` 100.0%, `rejectReason` 100.0%, `newApplierWithManifest` 100.0%

Residual uncovered lines are defensive error branches (json.Marshal failure on a plain
string/time struct = unreachable; os.OpenFile/f.Write failure = non-deterministic without
fault injection). The MkdirAll-failure branch IS covered (TestWriteLineageEntry_ParentDirCreateFails),
and the rejectReason synthesized-fallback branch IS covered (TestApply_RejectedSynthesizedReason).

### E4 — C-HRA-008 subagent boundary

```
$ go test ./internal/harness/ -run TestSubagentBoundary -> ok (canonical call-site guard green)
$ grep -rn 'AskUserQuestion(' internal/harness/ | grep -v '_test.go'  -> 0 call-site matches
```

Note: the acceptance.md AC-HLC-008 raw grep (`AskUserQuestion\|mcp__askuser`, comment-line
filter only) surfaces ONE pre-existing string-literal occurrence at
`proposalgen/scaffolder.go:111` ("the orchestrator AskUserQuestion gate" documentation prose).
This file has 0 diff lines in this SPEC (`git diff --stat` confirms untouched) and is the exact
documentation-prose false-positive the canonical `TestSubagentBoundary` guard was refined to
ignore (it targets the call-site pattern `AskUserQuestion(`, of which there are zero in non-test
harness source). The boundary is genuinely clean for this SPEC's changes.

### E5 — Lint

```
$ golangci-lint run ./internal/harness/... --timeout=3m -> 0 issues
```

### E6 — Full suite + cascade

```
$ go test ./internal/harness/... -count=1                          -> ok (all 8 packages)
$ go test ./internal/harness/... ./internal/cli/harness/... -count=1 -> ok (downstream importers, exit 0)
```

### E7 — Files changed (scope)

- `internal/harness/types.go` (EXTEND — added `LineageEntry` struct, additive omitempty)
- `internal/harness/lineage.go` (NEW — `WriteLineageEntry` + `LoadManifest`)
- `internal/harness/applier.go` (EXTEND `Apply` — accept/reject lineage wiring + `writeLineage`/`rejectReason`/`newApplierWithManifest` helpers; NO snapshot/restore/safety logic change)
- `internal/harness/lineage_test.go` (NEW — 13 tests covering 8 ACs + branch coverage)
- `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/spec.md` (frontmatter `status: draft -> in-progress`)
- `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/progress.md` (this evidence; created in run-phase — was untracked at plan-phase HEAD)

Protected runtime artifacts (`usage-log.jsonl`, `tier-promotions.jsonl`, `.moai/state/*`,
`.moai/cache/*`) UNCHANGED. `createSnapshot`/`RestoreSnapshot`/safety pipeline/L1 guard UNCHANGED.
`auto_apply: false` (harness.yaml) UNCHANGED. FROZEN constants UNCHANGED. No quality verdict added.

### Run-phase audit-ready signal

```yaml
run_complete_at: 2026-06-14
run_commit_sha: 3829f3a26
run_status: implemented
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: unchanged
cross_platform_build:
  host: pass
  windows_amd64: pass
new_warnings_or_lints_introduced: 0
coverage_package_total: "87.7%"
c_hra_008_boundary: clean
template_neutrality: clean
protected_artifacts_unchanged: true
total_run_phase_files: 6
m1_to_mN_commit_strategy: single-commit (Tier S, all M1-M4 in one feat commit)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-14
sync_commit_sha: <backfill at git commit — canonical sync-phase sha>
sync_status: implemented
docs_updated: [CHANGELOG.md]
specs_artifacts_modified:
  - "spec.md": "status in-progress → implemented"
  - "progress.md": "backfilled run_commit_sha + sync-phase evidence"
```
