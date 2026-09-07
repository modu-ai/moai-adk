# progress — SPEC-CODEX-EVENT-COVERAGE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-07
- tier: M (3 plan artifacts: spec.md / plan.md / acceptance.md + progress.md 추적 파일)
- plan-phase measurements: spec.md §C 표 M1-M5 (본 트리 실행 명령+관측 포함)
- open items: 없음 — [NEEDS CLARIFICATION] 마커 없음

## §E.2 Run-phase Evidence

Baseline: HEAD `e607dfa7f` (branch WT-codex-hook-events), this run, this tree. RED captured before GREEN (E8):

```text
$ go test ./internal/codexadapter/ -run 'TestEventTableRowCount|TestEventTableMapping|TestResolveInterruptNoCounterpart|TestResolveRecognizedButUnadapted' -v
    events_test.go:18: EventTable rows = 11, want 12
    events_test.go:128: Resolve(Interrupt) error = unknown codex hook event: "Interrupt", want an unadapted refusal
    events_test.go:46: expectation set size 12 != EventTable size 11
    events_test.go:107: Resolve(Interrupt) error = unknown codex hook event: "Interrupt", want an unadapted refusal
--- FAIL: TestEventTableMapping / TestResolveInterruptNoCounterpart / TestEventTableRowCount / TestResolveRecognizedButUnadapted
```

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CEV-001 | PASS | `awk '/^var EventTable/,/^}/' internal/codexadapter/events.go \| grep -c 'true},'` then same with `'false},'` | `6` and `6` (sum 12; false includes Interrupt) |
| AC-CEV-002 | PASS | `grep -rn 'EventInterrupt' internal/ \| grep -v _test \| grep -v codexadapter` | 0 matches (rc=1) |
| AC-CEV-003 | PASS | `go test ./internal/codexadapter/ -run 'TestResolve' -v` | `--- PASS: TestResolveInterruptNoCounterpart` (ErrUnadapted + no "dispatcher arg" assertion), `ok ... 0.409s` |
| AC-CEV-004 | PASS | `go test ./internal/codexadapter/... ./internal/codexwiring/...` | `ok ... codexadapter 0.571s` / `ok ... codexwiring 1.114s` — TestDispatcherArgsExist (empty-arg skip) + TestEventTableRowCount (12) GREEN |
| AC-CEV-005 | PASS | `go test ./internal/codexwiring/ -run 'TestRenderHooks_InterruptNeverInstalled' -v` | `--- PASS: TestRenderHooks_InterruptNeverInstalled (0.00s)` |
| AC-CEV-006 | PASS | `grep -c 'All eleven' internal/codexadapter/events.go` / `grep -c 'never an absence of' ...` | `0` and `0`; updated block names Interrupt + "no MoAI dispatcher counterpart" (events.go:50) |
| DoD#4 | PASS | `git diff --stat ace1c5440..HEAD -- internal/hook/` | empty output (0-row diff, REQ-CEV-005) |

M2 (campaign record: `.moai/reports/t496/codex-event-campaign.md`, evidence: `.moai/reports/t496/evidence/`):

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CEV-010 | PASS | per-event rows in the campaign record §2; captures at `evidence/captures/*.jsonl` | 6/6 rows carry fired/trigger-not-achieved verdict + exact command + observed output; 3 FIRED (SubagentStart, SubagentStop, Interrupt) with payloads; 3 trigger-not-achieved (PreCompact/PostCompact/PermissionRequest) with precondition evidence |
| AC-CEV-011 | PASS | campaign record §0-§1; `evidence/homecheck/zero-write-verdict.txt`; `codex --version` | `CODEX_HOME` SUPPORTED (run p0, transcript_path inside tmp home); key files' mtimes/hashes pre-campaign; 0 campaign-attributable real-home writes (parallel-lane `moai-codex-gate` sessions attributed by session_meta originator+cwd); pre-fix P0 run disclosed |
| AC-CEV-012 | PASS | campaign record §2 SubagentStop row; `evidence/runs/collab.jsonl` + `evidence/captures/SubagentStop.jsonl` | 0.147.0 not-fired observation REVERSED by 0.153.4 re-measurement: FIRED via collab_tool_call delegation, payload `last_assistant_message: "4"` |
| AC-CEV-013 | PASS | campaign record §4 disposition table | fires rows → adapter-extension decision (SubagentStart/Stop adapt-now; Interrupt follow-up-card with payload documented); not-fired rows → documented basis; 0 undispositioned rows |

M3 (conditional adapt):

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CEV-020 | PASS (branch b — 2 adapt-now rows) | M3 RED verbatim: `adapted rows = 6, want 8` / `SubagentStart: adapted = false, want true` (captured pre-GREEN); GREEN: `go test ./internal/codexadapter/ ./internal/codexwiring/` | `ok ... codexadapter 1.116s` / `ok ... codexwiring 0.598s`; census `true}=8` `false}=4`; `go test ./internal/cli/ -run 'Codex|Hooks'` → `ok ... 45.954s` |

Quality gates (D.2): `go test -cover ./internal/codexadapter/ ./internal/codexwiring/` → 87.3% / 88.2%; `golangci-lint run ./internal/codexadapter/...` → `0 issues.`; `go vet` both packages rc=0; `gofmt -l` empty. AC-CEV-002 re-checked post-M3: 0 matches. DoD#4 re-checked post-M3: empty diff.

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete (M1 + M2 + M3-adapt; REQ-CEV-012 split not needed — M2 closed in the same run)
- run_complete_at: 2026-09-07
- run_commit_sha: f865e57a0 (M2 record; M1 code b4653524b, M1 evidence 88c7d0c84, M3 adapt 865e12c80)
- ac_pass_count: 10 (AC-CEV-001..006, 010..013, 020)
- ac_fail_count: 0
- preserve_list_post_run_count: internal/hook untouched (merge-base diff 0, measured twice)
- l44_pre_commit_fetch: n/a (card worktree branch, no push per lane protocol)
- l44_post_push_fetch: n/a (lead batch-pushes develop)
- new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues on touched package; vet clean)
- cross_platform_build: not run this lane (no build-system files touched; production diff is 1 Go file of table data + comments) — CI verdict on origin/develop pending lead push
- total_run_phase_files: 7 (4 code/test + 1 SPEC frontmatter + 1 progress + campaign record/evidence set)
- m1_to_mN_commit_strategy: M1 code+tests, M1 evidence, M3 adapt, M2 record — 4 commits, evidence committed before each next change wave

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: PENDING_THIS_COMMIT
sync_status: complete
changelog_entry: CHANGELOG.md [Unreleased] §Added — SPEC-CODEX-EVENT-COVERAGE-001 bullet (b12 pre-emission grep 0, AC inventory 11 distinct IDs all PASS, file paths ls-verified)
docs_surfaces_assessed: docs-site/content/{en,ko,ja,zh}/advanced/codex-dual-harness.md updated (12-event table, 0.153.4 basis, SubagentStart/Stop adapted + RenderHooks install, honest trigger-not-achieved verdicts); README{,.ko,.ja,.zh}.md — no codex hook-event enumeration (grep rc=1), no update needed; hugo build verified warning-free post-edit
frontmatter_status_transitions:
  spec.md: in-progress -> completed (single sync commit, 3-phase close)
  plan.md: no status field (frontmatter carries none); updated already 2026-09-07
  acceptance.md: no status field; updated already 2026-09-07
  progress.md: no frontmatter (E.1-E.4 signal file)
mx_tag_validation: sync sub-step — no new exported surface this phase (docs+frontmatter only); run-phase M1/M3 annotations stand as landed
b12_self_test_a: pre-emission grep SPEC-CODEX-EVENT-COVERAGE-001 in CHANGELOG.md = 0
b12_self_test_b: distinct AC ids in acceptance.md = 11 (AC-CEV-001..006, 010..013, 020), all PASS in §E.2
b12_self_test_c: CHANGELOG-cited paths (docs-site 4 locales, campaign record) ls-verified before commit
```

Note: `sync_commit_sha` is stamped on the sync commit itself (chicken-and-egg); the commit message and this file's landing prove the binding. Observation carried to the lead: §E.3 `ac_pass_count: 10` enumerates 11 PASS ACs (001..006, 010..013, 020) — §E.2/§E.3 are manager-develop-owned and left unmodified.
