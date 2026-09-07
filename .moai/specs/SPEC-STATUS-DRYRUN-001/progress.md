# SPEC-STATUS-DRYRUN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-STATUS-DRYRUN-001
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
status: draft
plan_status: audit-ready
plan_complete_at: 2026-09-07
plan_audit: "PASS 0.875 (iter 2/2; .moai/reports/t513/plan-audit.md)"
diagnosis: pre-confirmed (card t513; repro evidence at .moai/reports/t513/)
plan_basis: accepted fix direction R1-R4 from the delegation prompt
open_clarifications: none
```

## §F Phase 4 Mode Selection

Input parameters: tier M; scope = 2 source files + their test files (`internal/spec/status.go`, `internal/cli/spec_status.go`); domains = 1 (Go source, spec-status subsystem); file language mix = 100% Go; concurrency benefit = LOW (coding-heavy, single subsystem); agent-team prereqs = not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | Semantic change in shared parse/write paths — not a typo-level fix |
| serial | **selected** | Coding-heavy single-subsystem work; Anthropic coding-task parallelism caveat |
| fanout | not selected | <3 domains, <10 files; no independent research fan-out warranted |
| sweep | not selected | Not mechanical-uniform bulk transformation; 2-file scope |

Decision: serial

Justification: the fix touches one parse/write module plus its CLI consumer with tight inter-file coupling (shared `ParseStatus`); a single sequential `manager-develop` spawn with the full Section A-E delegation brief minimizes coordination cost and write-conflict risk inside the card worktree. `serial` is the default fallback for coding-heavy work per the decision tree.

Kickoff: Implementation Kickoff Approval granted by operator 2026-09-07 (AskUserQuestion); progression mode: autonomous (ac_converge armed at run-phase entry).

## §E.2 Run-phase Evidence

All evidence measured in this run, against this tree (`WT-spec-status-dryrun`), TDD RED→GREEN per milestone.

### E8 — verbatim RED failing-test output (pre-GREEN, test-first falsifiability)

M1 (read side, R1) — `go test -run 'TestParseStatus_FrontmatterAnchored|...' ./internal/spec/`:

```
--- FAIL: TestParseStatus_FrontmatterAnchored_HistoryTableHeader (0.00s)
    status_test.go:493: ParseStatus = "Notes", want "completed" (frontmatter must win over body table header)
--- FAIL: TestParseStatus_FrontmatterAnchored_BacktickedProse (0.00s)
    status_test.go:526: ParseStatus = "Notes", want "draft" (backticked prose must never be the status)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.439s
```

M2 (write side, R2) — `go test -run 'TestUpdateStatus_FrontmatterAnchored_BodyByteInvariance|TestUpdateStatus_FrontmatterInsert_IntoExistingBlock' ./internal/spec/`:

```
--- FAIL: TestUpdateStatus_FrontmatterAnchored_BodyByteInvariance (0.00s)
    status_test.go:636: frontmatter status not updated, got:
    status_test.go:647: body was modified by the status update:
    status_test.go:650: history table header cell was clobbered
--- FAIL: TestUpdateStatus_FrontmatterInsert_IntoExistingBlock (0.00s)
    status_test.go:671: status not inserted, got:
    status_test.go:693: status key not inserted inside the frontmatter block:
    status_test.go:699: body was modified by the status insertion:
FAIL
```

M3 (CLI, R3+R4) — `go test -run 'TestSyncGitSpecStatuses_...' ./internal/cli/`:

```
--- FAIL: TestSyncGitSpecStatuses_DryRunWritesNothing (0.40s)
    spec_status_test.go:77: dry-run modified spec.md:
    spec_status_test.go:88: dry-run summary must not claim a write count, got:   SPEC-DRYGIT-001: draft → implemented
--- FAIL: TestSyncGitSpecStatuses_InvalidStatusSkippedLoudly (0.42s)
    spec_status_test.go:174: expected stderr warning naming the SPEC and offending value, got stderr: "" stdout:   SPEC-INVALIDST-001: Notes → implemented
    spec_status_test.go:182: skipped SPEC must not be written
--- FAIL: TestSpecStatusHelpText_DryRunAccurate (0.00s)
    spec_status_test.go:202: help must document the --sync-git + --dry-run combination
FAIL
```

(Tests passing at RED that are regression guards, not RED-now criteria: REQ-002 fallback read tests, REQ-006 real-run write test — required-green by design.)

### AC-002 mutation proof (§D.2)

Deleting the frontmatter branch in `parseStatusFromContent` (temporary mutant, `if false { ... }`) turns both AC-002 tests RED on the same tree:

```
--- FAIL: TestParseStatus_FrontmatterAnchored_HistoryTableHeader (0.00s)
    status_test.go:493: ParseStatus = "Notes", want "completed" (frontmatter must win over body table header)
--- FAIL: TestParseStatus_FrontmatterAnchored_BacktickedProse (0.00s)
    status_test.go:526: ParseStatus = "Notes", want "draft" (backticked prose must never be the status)
FAIL
```

### E1 — AC PASS/FAIL matrix

| AC | Status | Verification command | Actual output (this run, this tree) |
|----|--------|----------------------|-------------------------------------|
| AC-001 | PASS | `go test -run TestParseStatus_FrontmatterAnchored_HistoryTableHeader ./internal/spec/` | `--- PASS: TestParseStatus_FrontmatterAnchored_HistoryTableHeader` |
| AC-002 | PASS | `go test -run TestParseStatus_FrontmatterAnchored ./internal/spec/` (both variants) + mutant RED above | `ok github.com/modu-ai/moai-adk/internal/spec` — both variants return frontmatter values, never `Notes` |
| AC-003 | PASS | repro step [2] (below) | Frontmatter column = `completed` / `completed` / `draft`; drift-cache has zero `"FrontmatterStatus": "Notes"`; both `Drifted: true` records are true positives (`Summary: 2/3 SPECs have status drift`) |
| AC-004 | PASS | `go test -run TestUpdateStatus_FrontmatterAnchored_BodyByteInvariance ./internal/spec/` | `--- PASS` — frontmatter line updated; history table + backticked prose byte-identical |
| AC-005 | PASS | `go test -run 'TestParseStatus_TableFormat|TestParseStatus_MarkdownListFormat|TestUpdateStatus_TableFormat|TestUpdateStatus_MarkdownListFormat|TestParseStatus_LegacyTableFallback_NoFrontmatter' ./internal/spec/` | `ok` — legacy no-frontmatter read AND write survive unchanged |
| AC-006 | PASS | `go test -run TestSyncGitSpecStatuses_DryRunWritesNothing ./internal/cli/` + repro steps [4]-[5] | `--- PASS`; repro: `Summary: dry-run, nothing written: 1 would update, …`; post-dry-run hashes byte-identical to pre (`739d712…` / `7927ee0…` / `02b0c23…`); `git status` empty |
| AC-007 | PASS | `go test -run TestSyncGitSpecStatuses_RealRunUpdatesFrontmatter ./internal/cli/` | `--- PASS` — frontmatter `status: implemented` written, `Summary: updated 1` |
| AC-008 | PASS | `go test -run TestSyncGitSpecStatuses_InvalidStatusSkippedLoudly ./internal/cli/` | `--- PASS` — stderr warning `skipping SPEC-INVALIDST-001: parsed status "Notes" is not a valid status`; file not written |
| AC-009 | PASS | `bash .moai/reports/t513/repro.sh /tmp/t513-repro-bin/moai-fixed` | `.moai/reports/t513/repro-output-fixed.txt` — full flip list flipped, persist list persisted (below) |
| AC-010 | PASS | `go test -run TestSpecStatusHelpText_DryRunAccurate ./internal/cli/` | `--- PASS` — `--dry-run` keeps "Preview change without writing"; Long documents `--sync-git --dry-run` |

### AC-009 — repro flip/persist diff vs `repro-output-buggy.txt`

FLIPPED (all required by plan.md §E.2):

- `--list` statuses: `Notes`/`Notes` → `completed` / `draft` (lines 12, 14)
- drift table Frontmatter column: `Notes` → real frontmatter values (lines 19, 21)
- drift-cache `FrontmatterStatus`: `Notes` → `completed` / `draft` (lines 32, 34, 44, 46)
- step [4] summary: `Summary: updated 2, skipped 1 …` → `Summary: dry-run, nothing written: 1 would update, 2 skipped (already done), 0 skipped (invalid status), 0 not found` — no write count
- step [5] hashes byte-identical to step [3] pre-values (`739d712…`, `7927ee0…`, `02b0c23…`)
- step [5] `git status`: two ` M` lines → empty
- step [6] corruption diff section: empty

PERSISTED (expected-unchanged, by design):

- `Summary: 2/3 SPECs have status drift` (true positives — DEMO/PROSE have no close commit in the fixture history)
- `"GitImpliedStatus": "implemented"` (git inference untouched)

Note: SPEC-DEMO-001 now skips as `already completed` in step [4] — correct-by-design once the parser reads its real frontmatter `completed` (AC-003); only PROSE-001 (`draft`) is a would-update.

### E2 — cross-platform build

```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E3 — coverage

```
$ go test -cover ./internal/spec/    → coverage: 90.5% of statements  (target ≥85% MET)
$ go test -cover ./internal/cli/     → coverage: 81.0% of statements  (target ≥90%)
$ go tool cover -func (clean serial profile, /tmp/t513-cli-cover2.out) — spec_status.go:
    newSpecStatusCmd      94.7%
    updateSpecStatus      93.8%
    listAllSpecs          76.9%
    syncGitSpecStatuses   66.1%
    getSPECIDsFromGitLog  85.7%
    stdinIsTerminal        0.0%
```

internal/cli package-level 81.0% is dominated by pre-existing untested subcommand surfaces outside this SPEC's scope (`init`, `update`, `cc`, `glm`); the SPEC's changed file is well covered. syncGitSpecStatuses' residual uncovered blocks are error seams without injection points (findProjectRoot / git-log failure), the untestable interactive `fmt.Scanln` prompt, and the SPEC-file-missing + no-SPEC-IDs early returns — the latter two were covered by M4 sweep tests (`TestSyncGitSpecStatuses_SpecNotInSpecsDirCountedNotFound`, `TestSyncGitSpecStatuses_NoSpecIDsInGitLog`, both PASS). Gap: package-level 81.0% remains below the 90% convention — pre-existing, not introduced by this SPEC (no cli package file outside spec_status.go was modified).

### E4 — subagent-boundary grep

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/spec/ internal/cli/spec_status.go internal/cli/spec_status_test.go internal/cli/exitcode_streams_test.go | grep -v '_test.go' | grep -v '// '
→ 0 matches
```

### E5 — lint

```
$ golangci-lint run --timeout=2m                      → 0 issues (pre-change baseline)
$ golangci-lint run ./internal/spec/... ./internal/cli/... --timeout=2m → 0 issues (post-change)
NEW findings: 0
```

### E6 — commits (unpushed; lane never pushes — lead batch-pushes develop)

- `b9715249d` fix(SPEC-STATUS-DRYRUN-001): M1 — frontmatter-anchored status read (R1)  [carries `draft → in-progress` frontmatter transition]
- `cb3725554` fix(SPEC-STATUS-DRYRUN-001): M2 — frontmatter-anchored status write (R2)
- `2f30b220f` fix(SPEC-STATUS-DRYRUN-001): M3 — honor --dry-run on spec status --sync-git (R3+R4)
- M4 commit: M4 sweep tests (2) + progress.md §E.2/§E.3 + `repro-output-fixed.txt` evidence

### Scope deviations (justified)

- `internal/cli/exitcode_streams_test.go` — mechanical 2-arg → 3-arg call update for the new `syncGitSpecStatuses(cmd, autoConfirm, dryRun)` signature (`dryRun=false`; the tested non-TTY abort path is behavior-neutral). Required by the compiler; within the SPEC's semantic envelope.
- `TestCatalogHashParity` in `internal/spec` FAILS on this tree — PRE-EXISTING (catalog.yaml hash drift vs `internal/template/templates/.claude/agents/{manager-develop,manager-lead,manager-design,e2e-tester}.md`, none of which this SPEC touches; `git diff` at measurement shows only scope files modified). Out of scope: catalog regeneration is a separate cascade not attributable to this SPEC's envelope.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: "2f30b220f"
run_commit_sha_note: "last code-bearing commit (M3); §E.2/§E.3 evidence + repro-output-fixed.txt ride the M4 commit"
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
preserve_list_note: "ValidStatuses enum, legacy fallback paths (REQ-002/REQ-004), package-level regex promotion (REQ-PERF-004-A), internal/kanban/status_read.go, drift engine, git inference — all untouched"
l44_pre_commit_fetch: "n/a — lane worktree, no push; develop integration via lead batch"
l44_post_push_fetch: "n/a — no push performed by this lane"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
  linux_amd64: "not separately measured (pure Go string/regex work; windows cross-build proves the cross-platform surface)"
total_run_phase_files: 8
total_run_phase_files_note: "status.go, status_test.go, spec_status.go, spec_status_test.go, exitcode_streams_test.go (justified cascade), spec.md (status transition only), progress.md (this), repro-output-fixed.txt (evidence)"
m1_to_mN_commit_strategy: "per-milestone commits M1-M4 (RED→GREEN per milestone)"
known_red_preexisting: "TestCatalogHashParity (internal/spec) — pre-existing catalog hash drift, out of scope (see §E.2)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: "pending-backfill-sync"
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-STATUS-DRYRUN-001' CHANGELOG.md → 0 (pre-emission, no duplicate entry)"
b12_self_test_b: "AC count: acceptance.md SSOT yields 10 distinct AC identifiers (AC-001..AC-010); CHANGELOG entry cites 10 (AC-001..010) — match"
b12_self_test_c: "every file path claimed in the CHANGELOG entry verified to exist: internal/spec/status.go, internal/cli/spec_status.go, internal/cli/exitcode_streams_test.go, .moai/specs/SPEC-STATUS-DRYRUN-001/progress.md, .moai/reports/t513/repro-output-fixed.txt"
changelog_entry_position: "CHANGELOG.md [Unreleased] → ### Fixed, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (this sync commit)"
  updated_field: "2026-09-07 (already current — no change needed)"
mx_compliance_check:
  status: pass
  note: "MX tags validated as a sync sub-step — @MX:ANCHOR on frontmatterBlock (carries @MX:REASON), @MX:NOTE on package-level regexes; no dangling @MX:TODO introduced by this branch"
canary_compliance_check:
  status: pass
  note: "no policy forward-looking clauses defined by this SPEC; no canary self-test applies"
sync_phase_scope: "markdown-only — CHANGELOG entry, spec.md frontmatter status transition, this §E.4, orchestrator-rerun evidence file; zero code changes"
```
