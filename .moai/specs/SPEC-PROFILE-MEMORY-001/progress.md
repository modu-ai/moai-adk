# SPEC-PROFILE-MEMORY-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-02
tier: L
threshold: 0.85
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md]
req_count: 24
ac_count: 21
open_clarifications: 0
clarifications_resolved_at: 2026-08-02
audit_iterations:
  - iter: 1
    verdict: FAIL
    score: 0.68
    note: "Tier M(임계 0.80) 기준 감사. REQ 23 > Tier M 상한 16, D1 critical(해석기 미배선) 포함 12건"
  - iter: 2
    verdict: FAIL
    score: 0.82
    note: "Tier L(임계 0.85) 기준. 11/13 RESOLVED, 2 PARTIAL. 갭 0.03 — N1 critical(bare 런치 고지 공백) + N2 major(AC-PM-020 유도 방법 부재) 외 5건"
  - iter: 3
    verdict: PASS
    score: 0.88
    note: "N1-N7 델타 반영 후 재감사 완료. Tier L 임계 0.85 대비 +0.03. 점수 추이 0.68 → 0.82 → 0.88 로 단조 상승이므로 iter(N+1) < iter(N) STOP 신호는 발화하지 않았다. iter-3 은 최대 라운드였고 PASS 로 종료했으므로 PASS-with-debt / scope-reduction / 사용자 override 에스컬레이션은 불필요"
```

## §E.2 Run-phase Evidence

### 마일스톤과 커밋

M1~M5 는 각 마일스톤의 RED 테스트를 먼저 작성한 뒤 구현했다. M6 은 검증 전용 마일스톤이라
소스 변경이 없고, 이 문서가 그 산출물이다.

| 마일스톤 | 커밋 | 변경 파일 |
|----------|------|-----------|
| M1 — 원장 스키마 + 해석기/기록기 | `237fe01a4` | `internal/profile/profile.go`, `profile_test.go`, `project_scope_test.go` (+ plan-phase 산출물 6종, `draft → in-progress`) |
| M2 — 호출자 배선 | `6d5b3696d` | `internal/cli/web.go`, `internal/web/app.go` |
| M2 후속 — gofmt | `dab895cfd` | `internal/web/app.go` |
| M3 — 해석·기록 순서 재배열 + 시임 | `90dca4f4c` | `internal/cli/launcher.go`, `launcher_project_scope_test.go` |
| M4 — 새 프로필 고지 | `525b3c723` | `internal/cli/launcher.go`, `launcher_fresh_notice_test.go` |
| M5 — 테스트 격리 정비 | `fa0d5898a` | `internal/profile/main_test.go`, `profile_test.go`, `zz_sandbox_guard_test.go`, `internal/web/main_test.go` |
| M5 후속 — 주석 플랫폼 중립화 | `d6d2eac40` | `internal/profile/profile.go` |
| 검증 중 발견한 결함 수정 | `b4ae291e6` | `internal/cli/launcher.go` |
| M6 — AC 검증 | (이 커밋) | `progress.md` |

`internal/template/templates/` 는 8개 커밋 어디에서도 건드리지 않았다
(`git diff --name-only 237fe01a4^..b4ae291e6 | grep -c 'internal/template/templates/'` → `0`).

### AC 판정 행렬

`Actual Output` 열은 `acceptance.md` 의 판정 명령을 그대로 실행한 결과이지 요약이 아니다.
아래 값은 최종 트리(`b4ae291e6`)에서 재실행해 관측한 것이다.

| AC | 판정 | 검증 명령 | 실제 출력 |
|----|------|-----------|-----------|
| AC-PM-001 | PASS | `go test ./internal/profile/ -v -count=1` | `--- PASS: TestRecordForProject_PreservesLegacyKeys (0.01s)` |
| AC-PM-002 | PASS | 같은 실행 | `--- PASS: TestResolveForProject_ProjectScopeWinsOverGlobal (0.02s)` |
| AC-PM-003 | PASS | 같은 실행 | `--- PASS: TestResolveForProject_LegacyLedgerUnchanged (0.01s)` |
| AC-PM-004 | PASS | 같은 실행 | `--- PASS: TestResolveForProject_OptOutDisablesBothLookups (0.01s)` |
| AC-PM-005 | PASS | 같은 실행 | `--- PASS: TestForProject_EmptyRootFallsBackToGlobal (0.03s)` |
| AC-PM-006 | PASS | 같은 실행 | `--- PASS: TestResolveForProject_StaleProjectEntrySkipped (0.01s)` |
| AC-PM-007 | PASS | 같은 실행 | `--- PASS: TestProjectKey_NormalizationSymmetric (0.01s)` |
| AC-PM-008 | PASS | 같은 실행 | `--- PASS: TestRecordForProject_RejectsMissingDirectory (0.01s)` |
| AC-PM-009 | PASS | `go test ./internal/cli/ -v -count=1 -run '^TestUnifiedLaunch'` | `--- PASS: TestUnifiedLaunch_FirstTimeNewProfileIsRecorded (0.02s)` |
| AC-PM-010 | PASS | 같은 실행 | `--- PASS: TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch (0.02s)` |
| AC-PM-011 | PASS | `go test ./internal/profile/ ./internal/cli/ -run '^(TestHasClaudeConfig_DecidesOnClaudeJSONAlone\|TestFreshProfileNotice_WriterContent)$' -v -count=1` + grep -c | `2` (기대 2) |
| AC-PM-012 | PASS | `acceptance.md` §C 의 3개 부재-grep 그대로 | `PASS-a`, `PASS-b`, `PASS-c` |
| AC-PM-013 | PASS | `grep -c 'recordLastProfile func(name string) error' internal/web/app.go` + `go test ./internal/web/` | `1`, `ok ... internal/web 6.230s` |
| AC-PM-014 | PASS | (1) 두 패키지 가드 grep -c → `2`; (2) 홈 원장 해시 왕복 → `PASS-untouched`; (3) 반증 왕복 **실행(amendment)** — 양 패키지 `sandboxProfileBaseDir()` 무력화 시 `TestProfileBaseDirIsSandboxed` FAIL → 원복 후 PASS | (1) `2` (기대 2); (2) before=after=`0e59ed31…5be7538e`; (3) RED `main_test.go:72`(profile)/`:68`(web) `BaseDirOverride is empty` → GREEN PASS (amendment 커밋에서 관측) |
| AC-PM-015 | PASS | 상수 선언 grep + 리터럴 산재 grep | `34: projectsKey = "projects"`, `39: claudeConfigStateFile = ".claude.json"`, 산재 `PASS` |
| AC-PM-016 | PASS | `go build ./...` / `go vet ./...` / `GOOS=windows GOARCH=amd64 go build ./...` / `golangci-lint run ./internal/profile/... ./internal/cli/... ./internal/web/...` | 모두 exit 0, 린트 `0 issues.` |
| AC-PM-017 | PASS | `go test ./internal/cli/ -v -count=1 -run '^TestUnifiedLaunch'` | `--- PASS: TestUnifiedLaunch_UsesProjectScopedResolution (0.02s)` |
| AC-PM-018 | PASS | 같은 실행 | `--- PASS: TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce (0.09s)` (A/B/C-1/C-2 4케이스 포함) |
| AC-PM-019 | PASS | 테스트 + 배선 grep 2건 | `--- PASS: TestGetCurrentNameForProject_ProjectScoped (0.02s)`; `GetCurrentNameForProject(projectRoot)` → `1`, 전역 전용 잔존 → `0` (`PASS-wired`) |
| AC-PM-020 | PASS | (1) 원자 쓰기 프리미티브 grep; (2) `TestRecordForProject_NoPartialStateOnFailure` | (1) 4행 — `436: os.CreateTemp`, `455: os.Rename` (+369/432 주석) → `PASS-primitives`; (2) `--- PASS: TestRecordForProject_NoPartialStateOnFailure (0.02s)` |
| AC-PM-021 | PASS | (1) 후행 정렬 확인; (2) 패키지 전체 실행; (3) 반증 왕복 **실행(amendment)** — `TestGetBaseDir_Default` 의 `t.Cleanup` 복원 제거 시 전체 패키지 실행에서 `TestSandboxSurvivesPackageRun` FAIL → 원복 후 PASS | (1) `/bin/ls internal/profile/*_test.go \| sort \| tail -1` → `internal/profile/zz_sandbox_guard_test.go`; (2) `--- PASS: TestSandboxSurvivesPackageRun (0.00s)`; (3) RED `zz_sandbox_guard_test.go:38` `BaseDirOverride was cleared` → GREEN PASS (amendment 커밋에서 관측) |

집계: 21건 평가, FAIL 0건, PASS 21건 (AC-PM-014(3)/021(3) 반증 왕복이 amendment run 에서 실행돼 기존 PASS-WITH-DEBT 2건이 PASS 로 승격).

### 검증 중 발견해 고친 결함 1건

첫 린트 실행에서 신규 지적 1건이 나왔다.

```
internal/cli/launcher.go:147:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
```

원인은 M4 에서 마지막 사용 프로필 고지를 새 `launcherStderr` 시임 경유로 돌린 것이다.
errcheck 가 `os.Stderr` 에 적용하던 면제가 시임을 거치면서 사라졌다. 같은 파일이 이미 쓰던
`_, _ = fmt.Fprintf(...)` 형태로 맞추고 `b4ae291e6` 로 커밋했으며, 이후 게이트는 `0 issues.` 를
출력한다. 사전 존재 지적이 아니라 이 SPEC 이 만든 신규 지적이었으므로 이월하지 않고 그 자리에서 닫았다.

### plan 이 run-phase 확인 대상으로 표시한 2건

**AC-PM-020 의 rename 유도 레시피.** `plan.md` 는 원장 경로를 비어 있지 않은 디렉터리로 만들어
`os.Rename` 을 실패시키는 방법이 POSIX `rename(2)` 의미론에 근거한 **추론**이며 계획 단계에서
실행해 확인하지 않았다고 명시했다. run-phase 에서 `TestRecordForProject_NoPartialStateOnFailure`
가 이 레시피로 통과하므로 macOS 에서는 재현된다. 시임 도입 대안으로 전환할 필요는 없었다.
Windows 는 미검증 — 아래 잔여 위험 참조.

**AC-PM-021 의 후행 가드 — amendment 로 반증 왕복 완료.** `TestSandboxSurvivesPackageRun`
은 패키지 전체 실행에서 통과한다. 원본 close 에서는 AC-PM-021(3) 이 요구하는 반증 왕복
(`TestGetBaseDir_Default` 의 `t.Cleanup` 복원을 임시 제거 → FAIL 관측 → 원복 → PASS 관측)을
실행하지 않아 PASS-WITH-DEBT 로 남겨뒀다. amendment run(flip 커밋 `e0c61e318` 이후)에서 이
반증 왕복을 실행했다: `t.Cleanup` 복원 라인을 임시 주석 처리하고 `go test ./internal/profile/`
전체 실행 돌리면 `TestSandboxSurvivesPackageRun` 이 `zz_sandbox_guard_test.go:38` 에서
`BaseDirOverride was cleared by an earlier test and never restored` 로 FAIL 한다.
라인을 원복하면 다시 PASS. AC-PM-014(3) 도 같은 amendment 에서 실행했다 — 양 패키지
`sandboxProfileBaseDir()` 호출(및 `restore()`)을 임시 주석 처리하면
`TestProfileBaseDirIsSandboxed` 가 `main_test.go:72`(profile)/`:68`(web) 에서
`BaseDirOverride is empty` 로 FAIL 하고, 원복 시 PASS 한다. 두 가드 모두 장식이 아니라
load-bearing 임이 확인됐으므로 두 AC 는 PASS 로 승격했다. 양 왕복 모두 홈 원장 해시가
불변(`c9de8105…0fc42` == 사전 해시)했음을 관측해 샌드박스가 유지됐음도 확인했다.

### 증거 수집 중 관측한 판정 명령 취약점 2건

**AC-PM-021(1) 은 셸 alias 에 걸린다.** 이 머신에서 `ls` 는 `ls -la` 의 alias 라
`ls internal/profile/*_test.go | sort | tail -1` 이 파일명이 아니라 long-format 행
(`-rw-r--r--@ 1 goos staff 13701 Jul 7 01:30 internal/profile/sync_test.go`)을 내보내고,
정렬 키가 권한 문자열이 되면서 엉뚱한 파일을 지목한다. `/bin/ls` 로 우회하면 기대값
`zz_sandbox_guard_test.go` 가 그대로 나온다. 구현 결함이 아니라 판정 명령의 취약점이며,
`acceptance.md` 는 이 에이전트의 소유 범위가 아니므로 여기에만 기록한다.

**`internal/cli` 전체 실행에 외부 실패가 간헐적으로 섞인다.** 오케스트레이터가 넘긴 인계에는
`TestHookCommandFlushesLastHandlerEntry` 가 FAIL(8.64s) 로 기록돼 있었다. 이 테스트는
`internal/cli/hook_flush_test.go` 에 있고 커밋 `f8aaf7ea4` 로 들어온 **병렬 세션의**
SPEC-HOOK-TRACE-FLUSH-001 산출물이다. 다만 최종 트리에서 재실행한 3-패키지 전체 실행
(`go test ./internal/profile/ ./internal/cli/ ./internal/web/ -count=1`)에서는 `internal/cli`
가 `ok ... 271.576s` 로 통과했고 FAIL 행이 한 건도 없었다. 즉 결정적 실패가 아니라 간헐적
실패이며, 어느 쪽이든 이 SPEC 의 범위 밖이다. 해당 파일과 `internal/hook/trace/` 는 건드리지 않았다.

### 커버리지 (AC 아님 — 기준선 대비 관측)

`acceptance.md` 는 커버리지 AC 를 두지 않았으므로 판정 대상이 아니다. 다만 TRUST 5 의 85%
목표에 두 패키지가 미달하므로 기준선 귀속과 함께 기록한다. 기준선은 `237fe01a4^` 를 detached
worktree 로 체크아웃해 같은 명령으로 측정했다(공유 트리 미변경).

| 패키지 | 기준선 (`237fe01a4^`) | 최종 (`b4ae291e6`) | 증감 |
|--------|----------------------|--------------------|------|
| `internal/profile` | 82.9% 미만 — 81.2% | 82.9% | +1.7pp |
| `internal/web` | 59.6% | 59.6% | 변화 없음 |

85% 미달은 이 SPEC 이 만든 것이 아니라 사전 존재 부채이며, `internal/profile` 은 오히려
개선됐다. `internal/cli` 는 패키지 실행이 271초라 커버리지를 측정하지 않았다(미검증).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02
run_commit_sha: b4ae291e6
run_status: audit-ready
ac_total_count: 21
ac_pass_count: 21
ac_pass_with_debt_count: 0
ac_fail_count: 0
ac_pass_with_debt_ids: []
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
lint_defects_found_and_fixed_in_run: 1
cross_platform_build:
  status: pass
  host: "go build ./... exit 0; go vet ./... exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 17
template_files_touched: 0
m1_to_mN_commit_strategy: >-
  마일스톤당 1커밋 원칙에 후속 커밋 3건을 더한 8커밋 —
  M1 237fe01a4, M2 6d5b3696d (+ gofmt 후속 dab895cfd),
  M3 90dca4f4c, M4 525b3c723, M5 fa0d5898a (+ 주석 중립화 후속 d6d2eac40),
  검증 중 발견한 errcheck 결함 수정 b4ae291e6.
  M6 은 검증 전용이라 소스 변경이 없고 이 progress.md 커밋이 그 산출물이다.
unverified:
  - "AC-PM-020 rename 레시피의 Windows 재현 여부 미검증 (macOS 에서만 확인)"
  - "internal/cli 커버리지 미측정 (패키지 실행 271초)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: 53756d4f1
sync_status: audit-ready
b12_self_test_a: "grep -c 'SPEC-PROFILE-MEMORY-001' CHANGELOG.md → 0 (사전 중복 없음, 발행 진행)"
b12_self_test_b: "acceptance.md 의 고유 AC 식별자 21건 = CHANGELOG 항목이 명시한 21건 (19 PASS / 2 PASS-WITH-DEBT / 0 FAIL)"
b12_self_test_c: >-
  CHANGELOG 가 지목한 모든 경로를 ls 로 확인 —
  internal/profile/, internal/cli/launcher.go, internal/cli/web.go, internal/web/app.go,
  docs-site/content/{ko,en,ja,zh}/cli-reference/{profile,web}.md (8건),
  docs-site/static/images/profile/*.png (3건) 모두 존재.
changelog_entry_position: "[Unreleased] → Added, 최상단 (SPEC-REF-SEO-ABSORB-001 항목 바로 위)"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (단일 sync 커밋에 병합)"
  plan_md: "N/A — frontmatter 블록 없음"
  acceptance_md: "N/A — frontmatter 블록 없음"
  design_md: "N/A — frontmatter 블록 없음"
  research_md: "N/A — frontmatter 블록 없음"
  progress_md: "N/A — frontmatter 블록 없음"
  note: >-
    이 SPEC 세트에서 frontmatter 를 가진 산출물은 spec.md 하나뿐이다(head -1 로 6종 전수 확인).
    updated 는 이미 2026-08-02 이고 sync 커밋도 같은 날짜라 값 변화가 없다.
    phase / version 등 다른 필드는 손대지 않았다.
canary_compliance_check:
  applicable: false
  reason: "이 SPEC 은 자기 자신이 sync 에서 검증할 전방위 정책을 정의하지 않는다"
readme_decision: >-
  README 4종 미수정. Claude 프로필 서브시스템은 어느 README 에도 서술된 적이 없고
  (grep 히트는 모두 무관한 llm.profile 모델 프로파일), 동작이 아직 릴리스에 포함되지
  않았다. 없던 절을 새로 만드는 것은 근거 없는 편집이라 건너뛰었다.
carried_debt:
  - "AC-PM-020 rename 레시피 Windows 미검증"
  - "internal/cli 커버리지 미측정"
push_state: "미푸시 — 사용자가 push 결정을 보류했다"
```

---

## §F Phase 4 Mode Selection

amendment run-phase 진입 (flip 커밋 `e0c61e318`, completed → in-progress). 본 amendment는 부채 2건(AC-PM-014(3)/021(3))의 sandbox-guard 반증 왕복(falsification round-trip)을 실행하는 좁은 검증 작업이다.

입력 파라미터:
- tier: L (원본 분류 유지; 단 amendment scope는 좁음)
- scope: 2개 가드 파일 임시 무력화/원복 + 타겟 테스트 실행 — production-code 최종 변경 0
- domain count: 1 (`internal/profile` 샌드박스 가드 중심; `internal/web`은 AC-PM-014(1) grep만)
- file language mix: Go 테스트 파일
- concurrency benefit: LOW (단일 milestone 순차 검증)

모드 평가:
- Mode 1 trivial — 제외: 임시 수정/원복/검증 루프가 존재
- Mode 2 background — 제외: 결과가 다음 sync에 필요
- Mode 3 agent-team — RETIRED
- Mode 4 parallel — 제외: 단일 도메인 순차 검증, 병렬 이익 없음
- Mode 6 workflow — 제외: 좁은 검증, ≥30파일 기계 변환 아님
- Mode 5 sub-agent — 선택: 단일 milestone, 순차

Decision: sub-agent

정당화: 반증 왕복은 순차적(가드 무력화 → FAIL 관측 → 원복 → PASS 관측)이고 단일 패키지 범위이며 코드 최종 변경이 없는 검증 작업이다. 병렬화 이익이 없으므로 Anthropic coding-task parallelism caveat와 무관하게 Mode 5가 정합한다. manager-develop(cycle_type=tdd) 1회 위임.
