---
id: SPEC-HARNESS-CADENCE-BUILD-001
version: "0.1.2"
status: implemented
updated: 2026-07-13
---

# Progress — SPEC-HARNESS-CADENCE-BUILD-001

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-13 manager-spec: plan-phase artifact set authored (spec.md v0.1.0 — 34 REQs in 7 groups; plan.md — 5 milestones M1-M5, 3 open clarification markers; acceptance.md — 35 AC matrix rows + 6 scenarios + 6 edge cases). SPEC ID regex self-check PASS; 4 md byte-parity baselines measured PARITY; frontmatter 12-field schema validated. Ready for plan-audit. (Count correction at v0.1.1: this entry originally misstated 30 REQs / 34 ACs — plan-audit iter-1 D7.)
- 2026-07-13 manager-spec: plan-audit iter-1 revision (v0.1.1) — the 3 clarifications RESOLVED via orchestrator AskUserQuestion (registration-only thin-harness; gate-round question placement; minimal 3-field schedule) and recorded in plan.md §B Resolved clarifications; defect fixes applied: D2 AC-034 invocation-surface scoping, D3 AC-023 queue-path delta-frame (1 → 2), D4 new AC-HCB-002 (REQ-002 coverage), D5 REQ-044 formal modal, D6 AC-025 severity → MUST, D8 awk-bounded windowed greps + pinned heading anchors, D9 Retrofit-precedence pin (REQ-050 + AC-050 + plan M2). Current counts: 34 REQs / 36 AC matrix rows. SPEC ID regex self-check re-run PASS. No unresolved clarification markers remain. Ready for plan-audit iter-2.

## §E.2 Run-phase Evidence

- 2026-07-13 manager-develop (cycle_type=tdd): M1-M5 complete. Worktree branch `worktree-agent-a48c32b11d2683073`, base fast-forwarded to main HEAD `921cfbb24` pre-work (stale-base mitigation; two ff steps: 77809e76c→633e47c68→921cfbb24). Commits: M1 `b3be952c8` (manifest schedule schema + draft→in-progress), M2 `0cc25cec8` (doctrine surfaces), M3 `a941e861f` (CLI lifecycle awareness), M4 `b8a5c6ce7` (SKILL.md + mirrors + make build + catalog hash regen), M5 (this evidence commit). Evidence logs: `.moai/state/verify/cadence-build/` (m1-schedule-test.log, m3-schedule-test.log, m4-make-build.log, m4-template-tests.log, m5-full-test.log, m5-lint.log, m5-parity.log).

### AC PASS/FAIL matrix (36 rows; verification commands run verbatim, this tree)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-HCB-010 | PASS | `go test ./internal/harness/v4manifest/ -run 'Schedule' -count=1` | `ok ... 0.427s` — loop `30m` + cron `0 3 * * *` fixtures pass Validate (m1-schedule-test.log) |
| AC-HCB-011 | PASS | same run + full package `go test ./internal/harness/v4manifest/ -count=1` | `ok` — TestValidate_ScheduleAbsentRegression PASS; zero pre-existing tests modified |
| AC-HCB-012 | PASS | same run | TestValidate_ScheduleModeViolation PASS — `"run"`/`""`/`"write"`/case-variant rejected, error names `discovery-only` |
| AC-HCB-013 | PASS | same run | TestValidate_ScheduleMechanismAndInterval PASS — `"daemon"`→mechanism error; `""` interval→interval error |
| AC-HCB-014 | PASS | same run | TestValidate_ScheduleDecoderNoSilentModeDefault PASS — JSON omitting mode decodes empty + Validate rejects |
| AC-HCB-001 | PASS | `awk '/^#+ /{f=($0~/Recurrence/)} f' harness-builder.md` window | window non-empty (4 lines); names `PLAN→GENERATE gate round`; `grep -c discovery-only`=1; mirror byte-identical |
| AC-HCB-002 | PASS | same window greps | interval=1, loop=1, cron=1, `session-scoped`=1, `persistent` (ci)=1; mirror byte-identical |
| AC-HCB-003 | PASS | `grep -A3 'schedule' harness-builder.md` | Artifact 5 bullet shows interval/mechanism/mode (3/3/3 matches) + `omission-on-decline`=1; both trees |
| AC-HCB-020 | PASS | `grep -c "Scheduled Harness Discovery" cadence-bridge.md` | 1; `### Recipe 4` heading count 1; mirror byte-identical |
| AC-HCB-021 | PASS | awk `Recipe 4` window greps | window 16 lines; read-only=2, `queue persistence`=1, `no commits`=1, `no pushes`=1, run-phase=1, `<harness-name>` placeholder=1; `/harness:<name>` in window=0 |
| AC-HCB-022 | PASS | `grep -c "scheduled runs never commit, never push, never enter run-phase" cadence-bridge.md` | 1 (live) and 1 (mirror) — invariant unmodified, not restated |
| AC-HCB-023 | PASS | `grep -c ".moai/reports/cadence/<date>.md" cadence-bridge.md` | 2 (baseline 1 + 1 inside the Recipe 4 payload); window names `Discovery-to-Queue Contract`=1; mirror identical |
| AC-HCB-024 | PASS | awk eligibility window pipe-line count | 7 pipe-lines (baseline 6, +1); first-column `harness discovery` (ci) exactly 1 |
| AC-HCB-025 | PASS | Recipe 4 window grep | `degrades to the native /loop`=1 in window + updated catalog fallback bullet; both trees |
| AC-HCB-030 | PASS | file greps + awk ACTIVATE window `grep -n` | CronCreate=2 file-wide; session-scoped=2; ACTIVATE heading unique (1); window first-match `smoke gate` line 8 < `CronCreate` line 12 |
| AC-HCB-031a | PASS | `go test ./internal/cli/harness/ -run 'List' -count=1` | TestListHarnesses_ScheduleSurfaced PASS — Entry.Schedule populated; `--json` contains `"schedule"`; text row shows `schedule: nightly via cron` |
| AC-HCB-031b | PASS | same run | TestListHarnesses_ScheduleAbsentOmitted PASS — `--json` omits key; baseline text row byte-identical |
| AC-HCB-032 | PASS | `go test ./internal/cli/harness/ -run 'Remove' -count=1` | TestRemoveHarness_ScheduleUnregisterNotice PASS — cron→`CronDelete`, loop→cancellation, schedule-less→no notice; atomicity green |
| AC-HCB-033 | PASS | `go test ./internal/cli/harness/ -run 'Doctor' -count=1` | TestDoctor_ScheduleInvalidModeError PASS — `"mode":"write"` → ERROR DoctorFinding via doctor scan path + cobra doctor non-zero exit |
| AC-HCB-034 | PASS | boundary grep (delta-framed) | 5 matches = the 5 pre-existing documentation/help-text literals (propose.go ×3, execute.go ×1, install.go ×1) — 0 NEW; `grep -rn 'mcp__' internal/cli/harness/` excl. tests = 0 |
| AC-HCB-035 | PASS | `./bin/moai harness list` + `./bin/moai harness doctor` on this repo | list shows 3 harnesses, `grep -c "schedule:"`=0; doctor exit 0 |
| AC-HCB-040 | PASS | file greps on harness-builder.md | resolve-library-id=1, WebFetch≥1, `PLAN-aggregate input`=1; both trees |
| AC-HCB-042 | PASS | file greps | `MCP Fallback`=1, `never block`=1 ("never blocks"); both trees |
| AC-HCB-043 | PASS | `grep -c "glm-web-tooling"` | 1; both trees |
| AC-HCB-044 | PASS (SHOULD) | grep near research sub-step | `skip is recorded with rationale` present in the research sub-step sentence; both trees |
| AC-HCB-050 | PASS | `grep -c "Schedule Retrofit" harness-build-entry.md` + awk window | 2 matches; window carries detection rule (`commands/harness/<name>.md` + `scheduling intent`) + precedence `evaluated BEFORE`=1 |
| AC-HCB-051 | PASS | awk window `grep -n` first-matches | recurrence(ci) line 5 < `moai harness edit` line 7 < `CronCreate` line 9 — strictly increasing; CronCreate only in the Registration step (1 window match) |
| AC-HCB-052 | PASS | `grep -c "without fabricating a manifest"` | 1; `unavailable for manifest-less`=1 (user-informed limitation); both trees |
| AC-HCB-053 | PASS | `ls internal/template/templates/.claude/{commands,agents}/harness/` + leak test | both dirs absent (`No such file or directory`); `go test ./internal/template/ -run 'SplitHarnessNamespace'` ok |
| AC-HCB-060 | PASS | `diff` live vs mirror ×4 | exit 0 ×4 (m5-parity.log) |
| AC-HCB-061 | PASS | neutrality grep on the 4 files + `internal/template/templates/` | 0 matches (no SPEC IDs / REQ tokens) |
| AC-HCB-062 | PASS | `make build` + `go test ./internal/template/... -count=1` | both exit 0 (m4-make-build.log, m4-template-tests.log) |
| AC-HCB-070 | PASS | `go test ./... -count=1` | exit 0, 99 packages ok, 0 FAIL (m5-full-test.log) |
| AC-HCB-071 | PASS | `golangci-lint run --timeout=3m internal/harness/v4manifest/... internal/cli/harness/...` | `0 issues.` (m5-lint.log) |
| AC-HCB-072 | PASS | `./bin/moai harness doctor` | exit 0; `Scanned 3 harness(es): 0 ERROR, 2 INFO` — github/release remain INFO (Runner asymmetry preserved) |
| AC-HCB-073 | PASS-WITH-DEBT | `./bin/moai spec lint` | 1 pre-existing transient WARNING `StatusGitConsistency` (frontmatter mid-lifecycle status vs git-implied `implemented` from the plan-phase `feat(...)` commit subjects). Baseline-attributed: the SAME warning fires on the untouched main checkout with `status: draft` — it predates run-phase and self-resolves at sync close (terminal statuses exempted by `terminalStatusEnum`). 0 findings introduced by run-phase changes |

### Invariant rows

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Catalog invariant sentence unmodified, count 1 | PASS | AC-HCB-022 row above |
| Additive-only schema — zero behavior change for schedule-less manifests | PASS | AC-HCB-011/031b/035 rows; omitempty round-trip test |
| CLI never prompts (C-3) | PASS | AC-HCB-034 row; boundary grep delta 0 |
| No scheduler mechanics in Go (C-4) | PASS | no daemon/timer/cron engine added; CLI emits declared-schedule text only |
| Template neutrality (C-5 / REQ-HCB-061) | PASS | AC-HCB-061 row; template CI guards green |
| Cross-platform build | PASS | `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (post-M5 tree) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: 05dda2222  # M1 b3be952c8 / M2 0cc25cec8 / M3 a941e861f / M4 b8a5c6ce7; M5 = the commit carrying this section (self-referential — backfill in the next commit or at sync)
run_status: PASS-WITH-DEBT
ac_pass_count: 35
ac_pass_with_debt_count: 1  # AC-HCB-073 — pre-existing transient StatusGitConsistency WARNING, self-resolving at sync close
ac_fail_count: 0
preserve_list_post_run_count: 0  # no PRESERVE-listed file modified; scope held to plan §F file sets
l44_pre_commit_fetch: "git fetch origin main + rev-list --left-right = 1 0 pre-work (origin ahead; worktree ff'd to 921cfbb24 before M1)"
l44_post_push_fetch: "n/a — no push (worktree branch; orchestrator integrates per stale-worktree ff/cherry-pick discipline)"
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues; gofmt clean on all new/edited lines (5 pre-existing gofmt-unclean files in v4manifest left untouched per scope discipline)
cross_platform_build:
  darwin: "go build ./... exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
coverage:
  internal_harness_v4manifest: "100.0% (go test -cover, this tree)"
  internal_cli_harness: "81.6% (baseline main checkout 74.8% — +6.8pp improvement; 85% DoD gate was already unmet at baseline; residual gaps are pre-existing: NewHarnessV4EditCmd 0%, RemoveHarness error paths)"
total_run_phase_files: 13  # 3 Go source + 2 Go test + 3 doctrine md + SKILL.md + 4 mirrors + catalog.yaml (+ 4 SPEC frontmatter files)
m1_to_mN_commit_strategy: "per-milestone pathspec-scoped commits on the agent worktree branch (M1/M2/M3/M4/M5); no push per delegation instruction — Tier M main-direct integration performed by the orchestrator"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
