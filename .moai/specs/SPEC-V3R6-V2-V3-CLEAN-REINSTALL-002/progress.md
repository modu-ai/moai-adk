# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 — Progress

**Status**: completed (sync-phase closed; M4 bb886ecfa + M5 72f408178 NOT pushed — orchestrator owns landing of worktree branch crr-002-m4m5)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-16
- plan-auditor verdict: PASS
- plan-auditor score: 0.95 (Tier M threshold 0.80)
- skip-eligible: YES (score ≥ 0.90; governs ONLY Phase 1 re-execution — Implementation Kickoff Approval remains mandatory, obtained 2026-07-16)
- probes: P1 GEARS PASS / P2 root-cause PASS / P3 parent-contract non-weakening PASS / P4 Reproduction-First PASS
- blocking defects: none
- SHOULD-FIX: S1 (acceptance.md HARD-5 number collision, cosmetic) — deferred

## §E.2 Run-phase Evidence

### M1 — v3-version negative-override (REQ-CRR-001) — commit 73e7798ba (pushed)

**Scope**: 3 files touched (pathspec-only; unrelated `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` and `llm.yaml`/`README.ko.md` excluded).

| File | Change |
|------|--------|
| `internal/cli/v2_detection.go` | REQ-CRR-001 implementation: `probeVersionSignal` signature `(bool, string)` → `(bool, bool, string)`; `V2Fingerprint.V3VersionConfirmed` field added; `IsV2` aggregation changed from pure disjunction to `!V3VersionConfirmed && (S1 \|\| S2 \|\| S3)` |
| `internal/cli/v2_detection_test.go` | AC-CRR-002 reproduction test `TestDetectV2Fingerprint_V3Override_AC_CRR_002`; 2 aggregation cases updated to `wantIsV2: false` for v3+residue scenarios |
| `internal/cli/update_clean_install_test.go` | `makeScenarioB` fixture: `v3.0.0-rc2` → `v2.16.1` (partial-v2 project, not v3+residue — per AC-CRR-007 + Edge-2 design intent) |

**AC PASS/FAIL matrix (M1-relevant ACs)**:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CRR-001 | PASS | `go test -run TestDetectV2Fingerprint_V3Override ./internal/cli/` | `--- PASS: TestDetectV2Fingerprint_V3Override_AC_CRR_002` — V3VersionConfirmed=true, IsV2=false |
| AC-CRR-002 | PASS | `go test -run TestDetectV2Fingerprint_V3Override ./internal/cli/` | reproduction test: v3.0.0 + .agency/ + deprecated path → IsV2=false (loop-termination contract) |
| AC-CRR-007 | PASS | `go test -run TestRunCleanReinstall_ScenarioB ./internal/cli/` | `--- PASS: TestRunCleanReinstall_ScenarioB` — v2.* project + .agency/ → IsV2=true (clean-reinstall runs); v3+residue → IsV2=false (NOT activated) |

**Full-suite verification**:
- `go test ./internal/cli/... -count=1` → exit 0 (21.1s; all ScenarioA/B/C + v2_detection tests pass)
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/cli/...` → exit 0
- `golangci-lint run ./internal/cli/...` → exit 0 (no NEW issues; 6 pre-existing errcheck in `merge_test.go` untouched)
- Coverage: detectV2Fingerprint 95.0%, probeVersionSignal 80.0% (new v3.* branch covered by 5 test cases; pre-existing gaps in error/parse branches unchanged), probeDeprecatedPathSignal 100.0%

**Pending**: M2-M5 ACs not yet addressed (remaining milestones: M2 update.go integration, M3-M5 clean-reinstall path wiring).

> Note: §E.2 has no standalone M2/M3 entries — the M2 (`c4306e4e0`) and M3 (`c1508ce2c`) commits landed but their evidence rows were not appended here by the prior run-phase agents. The M4 entry below documents the completion of the M2 #1086 rejection (see M4 §finding).

### M4 — phantom-log gating + reproduction tests (REQ-CRR-006/008/009) — commit bb886ecfa (NOT pushed)

**Scope**: 3 files touched (specific-pathspec staging; no `git add -A`).

| File | Change |
|------|--------|
| `internal/cli/update_clean_install.go` | REQ-CRR-006: Step 4 REMOVE phase now re-scans post-REMOVE and derives `removedCount = len(pre-existing) − len(post-existing)` from the filesystem diff. `Removed N deprecated paths` emitted only when `removedCount > 0`; else `No deprecated paths found to remove`. |
| `internal/cli/update.go` | REQ-CRR-005 completion (see finding): added the non-project-cwd rejection specified in plan.md §F M2, placed **before** `acquireUpdateLock` and gated on `!binaryOnly`. |
| `internal/cli/update_clean_install_test.go` | `TestReproduction_FingerprintNonConvergence_Issue1084` (REQ-CRR-008), `TestReproduction_NonProjectDirectoryPollution_Issue1086` (REQ-CRR-009), `TestRunCleanReinstall_RemovalCountLogGating_AC_CRR_006` (REQ-CRR-006); gofmt reflow of a pre-existing M3 comment block. |

**FINDING — M2 #1086 was incompletely implemented (fixed in M4)**: M2 (`c4306e4e0`) added only the `&& isMoAIProject(cwd)` conjunct to the *clean-reinstall branch* of `runUpdate`. The early rejection that plan.md §F M2 explicitly specifies was never implemented, so a non-project cwd still fell through to the **v3 file-level sync**, which deployed the full embedded template tree into an arbitrary directory. `TestReproduction_NonProjectDirectoryPollution_Issue1086` exposed this on the current branch (FAIL: `.moai/`+`.claude/` created). M4 completes the M2 rejection (REQ-CRR-005 / AC-CRR-005). The gate sits **before `acquireUpdateLock`** because the lock itself `MkdirAll`s `.moai/` for its lockfile and its release removes only the lockfile — a post-lock gate still leaves an empty `.moai/` behind, violating AC-CRR-005(a).

**Reproduction-First (HARD-4 / CLAUDE.md §7 Rule 4)** — discharged against the true pre-fix baseline `0cdf18e07` (parent of M1; neither `isMoAIProject` nor `V3VersionConfirmed` present). Both reproduction tests were copied verbatim onto a `git worktree` at `0cdf18e07` and observed RED before any fix assertion existed:

```
--- FAIL: TestReproduction_FingerprintNonConvergence_Issue1084
    run 1: #1084 regression: IsV2 = true on a v3 project ... want false. signals: version=false agency=true deprecated=true
--- FAIL: TestReproduction_NonProjectDirectoryPollution_Issue1086
    AC-CRR-005(d): runUpdate returned nil in a non-project cwd; want a non-nil error
    AC-CRR-005(a): .moai/ was created in a non-project cwd
```

The AC-CRR-006 test was likewise confirmed RED at pre-fix (`AC-CRR-006(b): phantom log emitted with zero actual removals; found "Removed"`).

**AC PASS/FAIL matrix (M4-relevant ACs)**:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CRR-002 | PASS | `go test -run TestReproduction_FingerprintNonConvergence_Issue1084 ./internal/cli/` | `--- PASS` — v3 project + .agency/ + deprecated residue → IsV2=false, stable across 2 reads; language.yaml ko preserved |
| AC-CRR-005 | PASS | `go test -run 'TestReproduction_NonProjectDirectoryPollution_Issue1086\|TestIsMoAIProject_AC_CRR_005' ./internal/cli/` | `--- PASS` (both) — non-project cwd: no `.moai/`/`.claude/` created, structured `not a moai project` error naming the marker + `moai init`, exits non-zero, no cwd leak |
| AC-CRR-006 | PASS | `go test -run TestRunCleanReinstall_RemovalCountLogGating_AC_CRR_006 ./internal/cli/` | `--- PASS` (2 subtests) — 0 removals → no `Removed` line + `No deprecated paths found to remove`; 2 removals → `Removed 2 deprecated paths`; filesystem-diff derived |
| AC-CRR-008 | PASS | `git diff c1508ce2c HEAD -- internal/cli/update_preserve_inventory.go \| wc -l` | `0` (PRESERVE inventory untouched) |

### M5 — three-run idempotency + cross-platform parity (REQ-CRR-011, AC-CRR-009/010) — commit 72f408178 (NOT pushed)

**Scope**: 1 file touched (`internal/cli/update_clean_install_test.go`, test-only).

| File | Change |
|------|--------|
| `internal/cli/update_clean_install_test.go` | `TestRunUpdate_ThreeRunIdempotency_V3Project` (REQ-CRR-011/AC-CRR-009), `TestFingerprintPredicate_CrossPlatformParity` (AC-CRR-010); `runtime` import added. |

**AC PASS/FAIL matrix (M5-relevant ACs)**:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CRR-009 | PASS | `go test -run TestRunUpdate_ThreeRunIdempotency_V3Project ./internal/cli/` | `--- PASS` — 3 runs on v3 project: `conversation_language: ko` survives every run; file byte-stable from run 1 onward; no v2-to-v3 backup dir on runs 2/3; no REMOVE-phase log on runs 2/3 |
| AC-CRR-010 (S2) | PASS | `go test -run TestFingerprintPredicate_CrossPlatformParity ./internal/cli/` | `--- PASS` on darwin — `filepath.Join` marker resolution; isMoAIProject accepts v3 / rejects non-project; verdict stable-false on v3+residue. Linux/Windows CI-matrix exercised (build passes all 3). |

**Note on AC-CRR-009(a) semantics**: run 1 legitimately normalizes the minimal fixture via file-level sync (AC-CRR-009(d)) — the SHA-256 hash-diff classifies `language.yaml` as user-modified and preserves `conversation_language: ko` while merging in the remaining template keys (parent REQ-VVCR-007). The test pins **convergence** (byte-stable runs 1→2→3, ko never lost), not byte-equality to the original minimal input.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-07-16
- run_commit_sha: 72f408178 (M5, final); bb886ecfa (M4). NOT pushed — orchestrator owns landing decision for temporary worktree branch `crr-002-m4m5`.
- run_status: run-phase M4+M5 COMPLETE (M1-M3 pre-existing in branch)
- ac_pass_count: 6 M4/M5-scoped ACs verified PASS (AC-CRR-002, -005, -006, -008, -009, -010)
- ac_fail_count: 0
- preserve_list_post_run_count: `git diff c1508ce2c HEAD -- internal/cli/update_preserve_inventory.go` → 0 lines (HARD-2 non-weakening confirmed); `internal/defs/` diff → 0 lines (HARD-3 43-entry DeprecatedPaths frozen); `internal/template/templates/` diff → 0 lines (template isolation preserved)
- l44_pre_commit_fetch: N/A — isolated worktree branch, no push performed; orchestrator owns landing
- l44_post_push_fetch: N/A — no push
- new_warnings_or_lints_introduced: 0 (`golangci-lint run ./internal/cli/... --timeout=2m` → `0 issues.`; baseline also 0)
- cross_platform_build:
    - darwin_arm64: `go build ./...` → exit 0
    - windows_amd64: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
    - linux_amd64: `GOOS=linux GOARCH=amd64 go build ./...` → exit 0
- total_run_phase_files: 3 (update.go, update_clean_install.go, update_clean_install_test.go) across M4+M5
- m1_to_mN_commit_strategy: per-milestone commits — bb886ecfa (M4), 72f408178 (M5); Conventional Commits with 🗿 MoAI trailer
- coverage: isMoAIProject 100.0%, detectV2Fingerprint 95.0%, update_clean_install.go REMOVE-gate both branches covered (`262.22,264.3 → 1 1`, `264.8,266.3 → 1 1`); package-level internal/cli 74.2% (pre-existing baseline — update_clean_install.go 73.6% dominated by pre-existing untested Step 5/6/7 deployer error paths, NOT M4/M5 additions)
- subagent_boundary (E4): my 3 touched files have 0 `AskUserQuestion` references; the 32 repo-wide matches are pre-existing help-text/comment string literals in unrelated CLI files (harness.go, pr_watch_cmd.go), not subagent prompt invocations
- full_suite: `go test ./internal/cli/... -count=1` → all packages `ok` (0 FAIL)

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-07-16
- sync_commit_sha: pending-backfill-sync (backfilled in a follow-up commit on this branch — self-referential-SHA workaround, D3 exemption)
- sync_status: complete (single sync commit carries the merged `implemented → completed` transition; no separate Mx-phase chore commit)
- b12_self_test_a: `grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002' CHANGELOG.md` → 0 (pre-emission; confirmed no prior entry before this sync commit)
- b12_self_test_b: acceptance.md SSOT AC row count (`grep -cE '^\| AC-CRR-[0-9]+ \|'` matched against `§D.1 Severity Classification` table) = 10 (9 S1 + 1 S2); CHANGELOG entry references all 10 AC-CRR-NNN tokens across the 4 numbered sub-changes + closing PRESERVE-non-weakening clause
- b12_self_test_c: file paths cited in CHANGELOG entry (`internal/cli/v2_detection.go`, `internal/cli/update.go`, `internal/cli/update_clean_install.go`, `internal/cli/update_preserve_inventory.go`) verified present via `ls` before commit
- changelog_entry_position: `## [Unreleased]` → `### Fixed` (new subsection; first entry in Unreleased)
- frontmatter_status_transitions.spec_md: draft → completed (updated: 2026-07-16, unchanged date — created==updated==today)
- frontmatter_status_transitions.plan_md: not modified (out of sync-phase scope; body/frontmatter untouched)
- frontmatter_status_transitions.acceptance_md: not modified (out of sync-phase scope; body/frontmatter untouched)
- canary_compliance_check.push: NOT performed — orchestrator owns landing of worktree branch `crr-002-m4m5`
- canary_compliance_check.git_add_scope: specific-path only (CHANGELOG.md, spec.md, progress.md) — no `git add -A`

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M
- scope: ~5-10 files (internal/cli/update.go, v2_detection.go, update_clean_install.go, update_preserve_inventory.go, update_cleanup.go + tests)
- domain count: 1 (internal/cli clean-reinstall code path)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, sequential dependency between milestones)

**Mode evaluation**:
- Mode 1 (trivial): not selected — semantic regression repair, not a typo
- Mode 2 (background): not selected — write-capable implementation work
- Mode 3 (agent-team): RETIRED — never selected
- Mode 4 (parallel): not selected — single domain, coding-heavy (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): **selected** — single sequential manager-develop per milestone
- Mode 6 (workflow): not selected — scope < ~30 files, not mechanical-uniform

**Decision**: sub-agent (Mode 5)

**Justification**: Coding-heavy single-domain regression repair. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. Milestones M1→M5 have sequential dependencies (M1 fingerprint fix is the highest-irreversibility data-model change that M2-M5 build on). Tier M Section A-E delegation template applies.

**Implementation Kickoff Approval**: obtained 2026-07-16 (user selected "run-phase 진입 (권장)"). cycle_type=tdd (existing v2_detection_test.go / update_clean_install_test.go baseline).
