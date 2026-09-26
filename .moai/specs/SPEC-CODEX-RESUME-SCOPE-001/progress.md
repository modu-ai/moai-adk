---
id: SPEC-CODEX-RESUME-SCOPE-001
document: progress
card: t1201
---

# Progress — SPEC-CODEX-RESUME-SCOPE-001

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- 전제 측정: fixture transport 탐침(`.moai/reports/t1201/premise-probe.txt`, 소스 `premise-probe_test.go.txt`) — `resume_last` 가 카드 B의 `thr-6` 을 재개. MCP 서버 해석 경로(pid 82350: cwd `.claude/worktrees/t1169`, `CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go`) — 워크트리들이 레지스트리 하나를 공유.
- 운영자 결정 3건(`plan.md` §B): 혼합안, `thread_id` 는 이 프로젝트 레지스트리 기록분만, 거부는 구조화 JSON 만(`IsError` 없음). 모두 잠정 기본값이며 Implementation Kickoff 에서 운영자가 확정한다.
- fixture 응답 id 측정(`.moai/reports/t1201/fixture-probe.txt`): 기본 fixture 는 `turn/start` 를 `tid-fake` 로, 에코 변형은 요청 id 로 보낸다.
- plan-audit: iter-1 FAIL 0.75(`.moai/reports/t1201/plan-audit.md`) → v0.2.0 에서 D1~D8 수리, 선택 항목 D9~D13 반영. iter-2 재감사 대기.
- 규모: REQ 10개, AC 15개.
- LIVE 호출: plan 단계 0회.

## §E.2 Run-phase Evidence

- Kickoff: 운영자가 이 레인에서 2026-09-26 직접 승인. `plan.md` §B 세 결정(혼합안 / `thread_id` 는 이 프로젝트 레지스트리 기록분만 / 거부는 구조화 JSON, `IsError` 없음)을 그대로 확정.
- 기준: 워크트리 `t1201`, 브랜치 `WT-codex-resume-thread`, 착수 HEAD `b08e20910`, merge-base(origin/develop) `df526c9a9`.
- 증거 원문: `.moai/reports/t1201/run-evidence.txt`(gitignore 대상, 이 워크트리 기준 경로).
- 사전 점검(plan §C-2): 기준 트리에서 보존 대상 세 테스트 `go test ./internal/cli -run '…ResumeLast…|…SandboxPolicyResetOnReusedThread' -count=1 -v` → exit 0, 3 PASS.
- RED: 테스트만 쓴 상태 → 빌드 실패(exit 1). 데이터 필드만 선언한 상태 → 15개 중 14개 FAIL(exit 1), AC-CRS-001 이 `thread/resume threadId = "thr-6"` 로 이슈를 그대로 재현.
- GREEN: `go test ./internal/cli/... -run 'CodexTask|CodexJob' -count=1 -v` → exit 0, `--- PASS` 65줄·`--- FAIL` 0줄.
- `go vet ./internal/cli/` → exit 0. `gofmt -l <변경 파일>` → 출력 없음. `golangci-lint run ./internal/cli/...` → exit 0, `0 issues.`
- 패키지 전체 1회(슬롯 `internal-cli-suite` 임대 하에 직렬): `go test ./internal/cli/ -count=1 -timeout 25m` → exit 1, 실패 1건 `TestCodexSpawn_RealAssemblyThroughStubTmux`(`codex_launcher_test.go`, 이 변경 밖). 원인은 레인 환경 변수(`MOAI_FACTORY_WORKER=agent-35`, `MOAI_KANBAN_BACKEND=claude`)가 기대 명령에 섞여 든 것 — 같은 테스트를 스크럽 없이 돌리면 exit 1, `unset MOAI_KANBAN… MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test …` 로 돌리면 exit 0.
- 변경 함수 커버리지(`-run 'CodexTask|CodexJob' -coverprofile`): `hasThread`·`latestThreadForWorkKey`·`recordedThreads`·`codexRecordNewer`·`validCodexWorkKey`·`codexTaskRefusal` 100%, `handleCodexTask` 89.2%, `threadRecords` 78.9%(미커버는 읽기 실패 fail-open 분기).

| AC | 판정 | 테스트 | 실제 출력 |
|---|---|---|---|
| AC-CRS-001 | PASS | `TestCodexTaskResumeScope_WorkKeyResumesOwnCardThread` | `--- PASS` (RED 때 `threadId = "thr-6"`) |
| AC-CRS-002 | PASS | `TestCodexTaskResumeScope_SoleThreadReportsBasis` | `--- PASS` |
| AC-CRS-003 | PASS | `TestCodexTaskResumeScope_ThreadIDResumesExactThread` | `--- PASS` |
| AC-CRS-004 | PASS | `TestCodexTaskResumeScope_UnrecordedThreadIDRefused` | `--- PASS` |
| AC-CRS-005 | PASS | `TestCodexTask_ResumeLastReusesRecordedThread`, `TestCodexTask_ResumeLastWithNoRecordedThread`, `TestCodexTask_SandboxPolicyResetOnReusedThread` | 3 × `--- PASS`; `codex_task_test.go` 의 기준 대비 diff 빈 출력 |
| AC-CRS-006 | PASS | `TestCodexTaskResumeScope_AmbiguousResumeLastRefused` | `--- PASS` |
| AC-CRS-007 | PASS | `TestCodexTaskResumeScope_SharedThreadRecordsCountOnce` | `--- PASS` |
| AC-CRS-008 | PASS | `TestCodexTaskResumeScope_UnmatchedWorkKeyOpensNewThread` | `--- PASS` |
| AC-CRS-009 | PASS | `TestCodexTaskResumeScope_BackgroundRecordsWorkKey` | `--- PASS` |
| AC-CRS-010 | PASS | `TestCodexTaskResumeScope_InvalidWorkKeyRefused` (blank / 129B / control) | `--- PASS` |
| AC-CRS-011 | PASS | `TestCodexTaskResumeScope_CandidatesBoundedAndOrdered` | `--- PASS` |
| AC-CRS-012 | PASS | `TestCodexTaskResumeScope_ThreadIDPrecedenceAndUnusedSelectors` | `--- PASS` |
| AC-CRS-013 | PASS | `TestCodexTaskResumeScope_SchemaDeclaresScoping` | `--- PASS` |
| AC-CRS-014 | PASS | `TestCodexTaskResumeScope_RepresentativeRecordAcrossRecords` | `--- PASS` |
| AC-CRS-015 | PASS | `TestCodexTaskResumeScope_ReportsSentIDOnMismatchAndAckError` | `--- PASS` |
| (AC 없음, plan §D 회귀) | PASS | `TestCodexTaskResumeScope_PreSendFailureCarriesNoResumeFields` (세션 시작 실패 / `initialize` 거부) | `--- PASS` — 필드가 생기기 전에도 통과하는 회귀 방지 테스트이지 RED 로 이끈 테스트가 아니다 |

설계 편차 1건: `plan.md` §A 의 대상 파일 목록 밖인 `internal/cli/mcp_codex.go` 를 최소 수정했다 — `codexSessionError` 에 `threadRequestSent` 표지를 더하고, 스레드 요청을 쓴 뒤의 실패(ack 거부·스레드 id 없음)만 그 표지를 켠다. `openCodexSessionOn` 이 초기화와 스레드 요청을 한 번에 수행하므로, 이 표지 없이는 plan §D 의 "요청을 쓴 직후" 경계를 오류 문자열 매칭 말고는 가를 수 없었다. 스레드 요청 경로의 오류 문구·반환 모양은 바뀌지 않는다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 4122f7a4c   # M1 구현 커밋; 이 progress 갱신은 후속 커밋
run_status: complete-with-gap
ac_pass_count: 15
ac_fail_count: 0
added_regression_tests: 1   # PreSendFailureCarriesNoResumeFields (AC 없음)
preserve_list_post_run_count: 3   # AC-CRS-005 세 테스트, 파일 무변경
l44_pre_commit_fetch: not-run   # push 금지 레인; 리드 일괄 push
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build: not-measured   # 로컬 darwin 만; 매트릭스는 develop push CI
full_package_run: "exit 1 — 1 failure outside diff (TestCodexSpawn_RealAssemblyThroughStubTmux, lane env leak; passes scrubbed)"
total_run_phase_files: 6   # codex_task.go, codex_jobs.go, mcp_codex.go, mcp_server.go, codex_task_resume_scope_test.go, spec.md
m1_to_mN_commit_strategy: "M1 구현+테스트+status 전이 1커밋, progress 증거 1커밋"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash
sync_status: complete
b12_self_test_a: "grep -c SPEC-CODEX-RESUME-SCOPE-001 CHANGELOG.md -> 0 before emission"
b12_self_test_b: "distinct AC ids in acceptance.md = 15; CHANGELOG entry cites AC-CRS-001..015 (15)"
b12_self_test_c: "ls of every path cited in the entry -> all present"
changelog_entry_position: "[Unreleased] / ### Changed, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "no status field"
  acceptance_md: "no status field"
docs_sync: "none — grep 'resume_last' over docs-site/content/ and .claude/rules -> 0 files"
cross_spec_note: "SPEC-CODEX-PHASE2-001 HISTORY +1 line (REQ-CX2-008 narrowed)"
known_unrelated_residual: "TestCodexSpawn_RealAssemblyThroughStubTmux fails only under lane env (MOAI_KANBAN_*/MOAI_FACTORY_*); passes scrubbed"
push: not-performed   # lead batch push
```
