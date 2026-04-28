---
id: SPEC-WF-AUDIT-GATE-001
version: "1.0.0"
status: draft
created_at: 2026-04-25
updated_at: 2026-04-25
author: GOOS
priority: High
labels: [workflow, plan-audit, gate, governance, dogfood]
issue_number: null
depends_on: []
related_specs: []
---

# SPEC-WF-AUDIT-GATE-001 Task Decomposition

> 구현 작업 분해 + REQ/AC 매핑 + TRUST 5 게이트
> 최종 갱신: 2026-04-25
> Phase 매핑: plan.md §3 Phase A-F
> Task 개수: **18**

---

## 범례

- **Owner-role**: `implementer` (skill/Go 구현), `tester` (TDD 작성), `reviewer` (검수), `researcher` (조사·annotation). 팀 모드 시 `workflow.yaml` role profile로 스폰.
- **Isolation**: 팀 모드 병렬 실행 시 `worktree` 필요 여부 (구현/테스트 파일 쓰기 → worktree, 읽기 전용/skill 마크다운 수정 → 불필요).
- **Blocks**: 이 task 완료가 선행 조건인 후속 task ID.
- **TRUST 5 게이트**: 각 task DoD에 5 pillar 적용 (Tested/Readable/Unified/Secured/Trackable).

---

## Parallel Group 요약

| Group | Tasks | 사전 조건 |
|-------|-------|-----------|
| G-A (직렬) | T-01, T-02, T-03 | 없음 (Phase A 기반) |
| G-B (병렬) | T-04, T-05 | T-03 완료 |
| G-C (직렬) | T-06 | G-B 완료 |
| G-D (직렬) | T-07, T-08 | T-06 완료 |
| G-E (병렬) | T-09, T-10, T-11, T-12 | T-08 완료 (RED 단계) |
| G-E2 (직렬) | T-13, T-14, T-15 | G-E 완료 (GREEN/REFACTOR) |
| G-F (직렬) | T-16, T-17, T-18 | T-15 완료 |

---

## Phase A: 기반 — 디렉터리 + skill 단락 skeleton + 트윈 동기

### `T-01` 보고서 디렉터리 생성 + .gitignore 보강
- **REQ 매핑**: `REQ-WAG-004`
- **AC 매핑**: `AC-WAG-10`
- **Owner-role**: implementer
- **Isolation**: 불필요 (마크다운 + .gitkeep)
- **File ownership**:
  - `.moai/reports/plan-audit/.gitkeep`
  - `internal/template/templates/.moai/reports/plan-audit/.gitkeep`
  - `.gitignore` (보강)
  - `internal/template/templates/.gitignore` (보강)
- **의존성**: 없음 (루트 task)
- **Blocks**: T-02
- **설명**: `.moai/reports/plan-audit/` 디렉터리를 `.gitkeep`로 추적 가능하게 생성. `.gitignore`에 `.moai/reports/plan-audit/*.md` 추가 (보고서는 로컬 산출물). 트윈 2개 동시 변경.
- **DoD**:
  - [ ] **Tested**: `os.Stat(".moai/reports/plan-audit")` IsDir true (tasks.md §5 검증 스크립트)
  - [ ] **Readable**: `.gitkeep` 헤더 주석 포함 ("# moai plan-audit reports — see SPEC-WF-AUDIT-GATE-001")
  - [ ] **Unified**: 두 .gitignore 변경이 `diff` 결과 동등 (트윈 일치)
  - [ ] **Secured**: 디렉터리 권한 0755, `.gitignore` 패턴이 `.moai/reports/` 외부 영향 없음
  - [ ] **Trackable**: 커밋 메시지 `chore(plan-audit): create reports directory (SPEC-WF-AUDIT-GATE-001 T-01)`

### `T-02` `run.md` Phase 0.5 placeholder 단락 추가 (solo)
- **REQ 매핑**: `REQ-WAG-001`
- **AC 매핑**: `AC-WAG-01` (사전조건)
- **Owner-role**: implementer
- **Isolation**: 불필요 (단일 마크다운 파일 2 트윈)
- **File ownership**:
  - `.claude/skills/moai/workflows/run.md`
  - `internal/template/templates/.claude/skills/moai/workflows/run.md`
- **의존성**: T-01
- **Blocks**: T-04 (Phase B 본문 작성)
- **설명**: Phase 0과 Phase 1 사이에 `## Phase 0.5: Plan Audit Gate (TBD — see SPEC-WF-AUDIT-GATE-001)` 헤더만 추가. 본문은 Phase B에서 채움.
- **DoD**:
  - [ ] **Tested**: skill audit 테스트(`internal/template/skills_audit_test.go` 또는 신설)에서 `Phase 0.5` 헤더 존재 검증
  - [ ] **Readable**: 헤더 1줄 + TBD 표기 + SPEC-ID 백링크
  - [ ] **Unified**: 두 트윈 byte-level 일치
  - [ ] **Secured**: N/A (마크다운)
  - [ ] **Trackable**: 커밋 메시지 SPEC-ID 레퍼런스

### `T-03` `team/run.md` placeholder + `plan.md` audit-ready 출력 + `spec-workflow.md` 보강
- **REQ 매핑**: `REQ-WAG-005`
- **AC 매핑**: `AC-WAG-05` (사전조건)
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.claude/skills/moai/team/run.md` + 트윈
  - `.claude/skills/moai/workflows/plan.md` + 트윈
  - `.claude/rules/moai/workflow/spec-workflow.md` + 트윈
- **의존성**: T-02
- **Blocks**: T-04, T-05
- **설명**: (a) team/run.md에 동일 placeholder 추가, (b) plan.md 종료 단락에 "Output progress.md `plan_complete_at: <ISO-8601>` to signal audit-ready" instruction 추가, (c) spec-workflow.md "Phase Transitions"의 "Plan to Run" 항목 보강.
- **DoD**:
  - [ ] **Tested**: 3 파일 모두 헤더/문구 존재 grep 검증
  - [ ] **Readable**: 변경된 단락이 기존 헤더 패턴과 일관(H2/H3 레벨 일치)
  - [ ] **Unified**: 6개 파일(원본 + 트윈 3쌍) byte-level 일치
  - [ ] **Secured**: N/A
  - [ ] **Trackable**: 단일 커밋, SPEC-ID 명시

---

## Phase B: solo `run.md` 게이트 본문 작성

### `T-04` Phase 0.5 본문 5-step instruction 작성
- **REQ 매핑**: `REQ-WAG-001`, `REQ-WAG-002`, `REQ-WAG-003`, `REQ-WAG-004`
- **AC 매핑**: `AC-WAG-01`, `AC-WAG-02`, `AC-WAG-03`, `AC-WAG-04`
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.claude/skills/moai/workflows/run.md` + 트윈
- **의존성**: T-03
- **Blocks**: T-06
- **설명**: Phase 0.5 단락에 5 sub-step 본문 작성. (1) plan 산출물 hash, (2) 24h cache, (3) plan-auditor 호출 syntax, (4) verdict 4-way 분기, (5) progress.md persist + report append. 사용자 결정은 AskUserQuestion으로 위임함을 명시.
- **DoD**:
  - [ ] **Tested**: 통합 테스트 `TestRunInvokesPlanAuditorBeforeImplementation` skeleton 작성 (Phase E에서 채움)
  - [ ] **Readable**: 5 sub-step 모두 H3 헤더 + 1줄 instruction
  - [ ] **Unified**: 두 트윈 일치
  - [ ] **Secured**: 파일 경로 패턴 명시 시 `.moai/reports/plan-audit/<SPEC>-<DATE>.md` 정확히 명기 (path traversal 방지 instruction)
  - [ ] **Trackable**: 단일 커밋

### `T-05` `team/run.md` 게이트 본문 작성
- **REQ 매핑**: `REQ-WAG-005`
- **AC 매핑**: `AC-WAG-05`
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.claude/skills/moai/team/run.md` + 트윈
- **의존성**: T-03
- **Blocks**: T-06
- **설명**: T-04와 동일한 5 sub-step 본문 + team-specific 차이 명시 (main session 단독 호출, TeamCreate 이전, teammate spawn 시 cache hit 확인). 기존 "Phase 1 — Task Decomposition" 직전 위치.
- **DoD**:
  - [ ] **Tested**: 통합 테스트 `TestTeamRunBlockedBeforeTeammateSpawn` skeleton (Phase E에서 채움)
  - [ ] **Readable**: 단락 구조 T-04와 1:1 대응 (인지 부하 최소화)
  - [ ] **Unified**: 두 트윈 일치
  - [ ] **Secured**: teammate spawn 분기에 verdict 검사 명시
  - [ ] **Trackable**: 단일 커밋

---

## Phase C: 통합 검증 (skill 마크다운 자체 일관성)

### `T-06` skill audit 테스트 신설/보강 (SPEC-THIN-CMDS-001 패턴 차용)
- **REQ 매핑**: 전체 (정합성)
- **AC 매핑**: 사전조건
- **Owner-role**: tester
- **Isolation**: 불필요 (테스트 파일 추가)
- **File ownership**:
  - `internal/template/skills_audit_test.go` (신설 또는 기존 audit 테스트 보강)
- **의존성**: T-04, T-05
- **Blocks**: T-07
- **설명**: `internal/template/commands_audit_test.go`와 유사한 패턴으로 `run.md`/`team/run.md`/`plan.md`/`spec-workflow.md` 4 파일에서 다음 패턴 grep 검증: (a) `Phase 0.5: Plan Audit Gate` 헤더, (b) `plan-auditor` 단어 등장, (c) `--skip-audit` 단어 등장, (d) `INCONCLUSIVE` 단어 등장, (e) `.moai/reports/plan-audit/` 경로 등장.
- **DoD**:
  - [ ] **Tested**: `go test ./internal/template/...` PASS
  - [ ] **Readable**: 테스트 함수명이 검증 의도 직설 (`TestRunMdContainsPhase05Header` 등)
  - [ ] **Unified**: 기존 audit 테스트와 스타일 일치 (table-driven)
  - [ ] **Secured**: 테스트는 `t.TempDir()` 무관 (정적 파일 검증)
  - [ ] **Trackable**: 커밋 메시지 SPEC-ID 명시

---

## Phase D: `--skip-audit` + INCONCLUSIVE 명세

### `T-07` `--skip-audit` 우회 절 작성 (run.md + team/run.md)
- **REQ 매핑**: `REQ-WAG-006`
- **AC 매핑**: `AC-WAG-06`
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.claude/skills/moai/workflows/run.md` + 트윈
  - `.claude/skills/moai/team/run.md` + 트윈
- **의존성**: T-06
- **Blocks**: T-08
- **설명**: 양 파일에 `### When --skip-audit Flag Is Provided` 절 추가. 우회 동작, 보고서 4 필드 (`verdict: BYPASSED`, `bypass_at`, `bypass_user`, `bypass_reason`), 비대화형 환경 처리(`bypass_reason: "non-interactive"`) 명시. 환경변수 `MOAI_SKIP_PLAN_AUDIT=1` 동등 처리 명시.
- **DoD**:
  - [ ] **Tested**: skill audit 테스트에 `--skip-audit` 절 존재 검증 추가
  - [ ] **Readable**: 4 필드 모두 코드 블록(```yaml)으로 예시 제공
  - [ ] **Unified**: 두 파일 본문 1:1 대응 (team-specific 부분만 차이)
  - [ ] **Secured**: rationale 입력은 마크다운 escape 후 기록 instruction 명시
  - [ ] **Trackable**: 커밋 메시지 SPEC-ID + REQ-WAG-006 명시

### `T-08` INCONCLUSIVE fall-back 절 작성 (run.md + team/run.md)
- **REQ 매핑**: `REQ-WAG-007`
- **AC 매핑**: `AC-WAG-07`
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.claude/skills/moai/workflows/run.md` + 트윈
  - `.claude/skills/moai/team/run.md` + 트윈
- **의존성**: T-07
- **Blocks**: T-09 ~ T-12 (Phase E RED 시작)
- **설명**: 양 파일에 `### When Plan-Auditor Fails or Times Out` 절 추가. timeout(60s 기본), malformed output, panic 3 케이스 모두 INCONCLUSIVE 분류. AskUserQuestion 3-way (retry / proceed-with-acknowledgement / abort). proceed 시 progress.md `inconclusive_acknowledged_by: <user>` 강제 기록. retry 횟수 제한(OPEN QUESTION Q3) 결정 후 명시.
- **DoD**:
  - [ ] **Tested**: skill audit 테스트에 `INCONCLUSIVE` + `proceed-with-acknowledgement` 단어 존재 검증
  - [ ] **Readable**: 3 케이스(timeout/malformed/panic) 명시 표
  - [ ] **Unified**: 두 파일 1:1 대응
  - [ ] **Secured**: PASS 자동 처리 금지 명시 (음성 instruction)
  - [ ] **Trackable**: OPEN QUESTION Q3 해소 결과를 단락 끝 주석으로 기록

---

## Phase E: 통합 테스트 작성 (TDD RED → GREEN → REFACTOR)

### `T-09` RED — `run_audit_gate_integration_test.go` 작성 (5 AC)
- **REQ 매핑**: `REQ-WAG-001`, `REQ-WAG-002`, `REQ-WAG-003`, `REQ-WAG-006`, `REQ-WAG-007`
- **AC 매핑**: `AC-WAG-01`, `AC-WAG-02`, `AC-WAG-03`, `AC-WAG-06`, `AC-WAG-07`
- **Owner-role**: tester
- **Isolation**: worktree (테스트 파일 + dummy fixture 작성)
- **File ownership**:
  - `internal/cli/run_audit_gate_integration_test.go`
  - `internal/cli/testdata/audit-gate/SPEC-DUMMY-{PASS,FAIL,BYP,INC}-001/` (4 fixture)
- **의존성**: T-08
- **Blocks**: T-13
- **설명**: 5 테스트 함수 작성 — `TestRunInvokesPlanAuditorBeforeImplementation`, `TestRunBlockedOnAuditFail`, `TestRunProceedsOnAuditPassAndPersistsVerdict`, `TestSkipAuditFlagRecordsBypassWithUserRationale`, `TestEnvVarSkipAuditEquivalentToFlag`, `TestPlanAuditorFailureClassifiesAsInconclusive`. PlanAuditor 인터페이스를 정의하고 mock 구현 사용. 빌드 태그 `//go:build integration`. RED 단계 — 모두 fail 확인.
- **DoD**:
  - [ ] **Tested**: `go test -tags=integration -run TestRun -count=1 ./internal/cli/...` 6 테스트 모두 FAIL (RED 확인)
  - [ ] **Readable**: 각 테스트가 GIVEN/WHEN/THEN 주석 포함
  - [ ] **Unified**: table-driven 패턴 + `t.Run(name, ...)` 서브테스트
  - [ ] **Secured**: fixture는 `t.TempDir()` 또는 testdata/ 내부, 실제 SPEC 영향 없음
  - [ ] **Trackable**: 각 테스트 함수 godoc에 AC-ID 백링크

### `T-10` RED — `run_audit_gate_grace_test.go` (시간 의존)
- **REQ 매핑**: `REQ-WAG-002` 변형
- **AC 매핑**: `AC-WAG-08`
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**:
  - `internal/cli/run_audit_gate_grace_test.go`
- **의존성**: T-08
- **Blocks**: T-13
- **설명**: `TestGraceWindowWarnOnlyMode`, `TestGraceWindowExpiryRevertsToBlockingMode` 2 테스트. `Clock` 인터페이스 + `FakeClock` 사용. T-09와 병렬 실행 가능 (다른 파일).
- **DoD**:
  - [ ] **Tested**: 2 테스트 RED 확인
  - [ ] **Readable**: T0/T+3/T+8 시점 주입을 헬퍼로 추출
  - [ ] **Unified**: t.Run 서브테스트 패턴
  - [ ] **Secured**: 시간 주입은 `MOAI_AUDIT_GATE_T0` 환경변수 또는 Clock 인터페이스 (전역 `time.Now` 변경 금지)
  - [ ] **Trackable**: AC-WAG-08 백링크

### `T-11` RED — `run_audit_gate_cache_test.go` + `run_audit_gate_filesystem_test.go`
- **REQ 매핑**: `REQ-WAG-003` 변형, `REQ-WAG-004` 변형
- **AC 매핑**: `AC-WAG-09`, `AC-WAG-10`
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**:
  - `internal/cli/run_audit_gate_cache_test.go`
  - `internal/cli/run_audit_gate_filesystem_test.go`
- **의존성**: T-08
- **Blocks**: T-13
- **설명**: 캐시 hit/invalidation 테스트 2개 + 디렉터리 자동 생성/readonly fall-back 테스트 2개.
- **DoD**:
  - [ ] **Tested**: 4 테스트 RED 확인
  - [ ] **Readable**: 캐시 키(plan artifact hash) 계산 헬퍼 추출
  - [ ] **Unified**: 동일 패턴
  - [ ] **Secured**: readonly fall-back 테스트는 `os.Chmod(0444)` 후 cleanup `t.Cleanup(func(){ os.Chmod(0755) })` 보장
  - [ ] **Trackable**: AC-WAG-09, AC-WAG-10 백링크

### `T-12` RED — `team_run_audit_gate_test.go` + `dogfood_self_audit_test.go`
- **REQ 매핑**: `REQ-WAG-005`, dogfood
- **AC 매핑**: `AC-WAG-05`, `AC-WAG-11`
- **Owner-role**: tester
- **Isolation**: worktree
- **File ownership**:
  - `internal/cli/team_run_audit_gate_test.go`
  - `internal/cli/dogfood_self_audit_test.go`
- **의존성**: T-08
- **Blocks**: T-13
- **설명**: team 모드 게이트 차단 테스트 1개 + dogfood self-audit 테스트 1개. dogfood 테스트는 본 SPEC 디렉터리(`.moai/specs/SPEC-WF-AUDIT-GATE-001/`)를 plan-auditor mock 또는 실 호출에 입력.
- **DoD**:
  - [ ] **Tested**: 2 테스트 RED 확인
  - [ ] **Readable**: dogfood 테스트가 본 SPEC 경로를 hardcode 하지 않고 `runtime.Caller` 또는 환경변수 사용
  - [ ] **Unified**: 동일 패턴
  - [ ] **Secured**: team 테스트는 실제 tmux/TeamCreate 호출 없이 mock TeamCreator 사용
  - [ ] **Trackable**: AC-WAG-05, AC-WAG-11 백링크

### `T-13` GREEN — `internal/runtime/audit_gate.go` 구현 (orchestration)
- **REQ 매핑**: `REQ-WAG-001` ~ `REQ-WAG-007` 전체
- **AC 매핑**: `AC-WAG-01` ~ `AC-WAG-07` 전체
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**:
  - `internal/runtime/audit_gate.go`
  - `internal/runtime/audit_gate_test.go` (단위 테스트)
- **의존성**: T-09, T-10, T-11, T-12 (모든 RED 완료)
- **Blocks**: T-14
- **설명**: gate orchestration 로직. `PlanAuditor` 인터페이스 정의, verdict 4-way 분기, grace window 결정, AskUserQuestion 분기 위임 (orchestrator 호출 경로). 단위 테스트는 분기 로직 검증, 통합 테스트는 T-09~T-12에서 검증.
- **DoD**:
  - [ ] **Tested**: 단위 테스트 + T-09~T-12 통합 테스트 GREEN 전환
  - [ ] **Readable**: PlanAuditor 인터페이스 godoc 완비, verdict enum 4 값 명시
  - [ ] **Unified**: `gofmt -l` empty, `golangci-lint run` zero error
  - [ ] **Secured**: AskUserQuestion 호출은 main session에서만 (subagent에서 호출 금지 검증)
  - [ ] **Trackable**: 함수 godoc에 REQ-ID/AC-ID 매핑

### `T-14` GREEN — `internal/runtime/audit_cache.go` + `audit_report.go` 구현
- **REQ 매핑**: `REQ-WAG-003`, `REQ-WAG-004`
- **AC 매핑**: `AC-WAG-04`, `AC-WAG-09`, `AC-WAG-10`
- **Owner-role**: implementer
- **Isolation**: worktree
- **File ownership**:
  - `internal/runtime/audit_cache.go`
  - `internal/runtime/audit_report.go`
  - 단위 테스트 2개
- **의존성**: T-13
- **Blocks**: T-15
- **설명**: (a) plan artifact hash 계산(OPEN QUESTION Q1 결정 — 정규화 방식), 24h TTL cache, hash 기반 invalidation. (b) daily report 파일 append, 권한 거부 시 INCONCLUSIVE 위임. `filepath.Clean` + projectDir scope 검증.
- **DoD**:
  - [ ] **Tested**: 캐시/보고서 단위 테스트 + T-11 통합 테스트 GREEN 전환
  - [ ] **Readable**: hash 알고리즘이 `audit_cache.go` 헤더 주석에 명시
  - [ ] **Unified**: 모든 시간 비교는 UTC ISO-8601, `Clock` 인터페이스 경유
  - [ ] **Secured**: 보고서 경로 path traversal 방어 (`strings.HasPrefix(filepath.Clean(p), projectDir)` 검증)
  - [ ] **Trackable**: 커밋 분리 가능 (cache 1 + report 1)

### `T-15` REFACTOR — 중복 제거 + 통합 테스트 전체 GREEN 확인
- **REQ 매핑**: 전체
- **AC 매핑**: 전체
- **Owner-role**: implementer + reviewer
- **Isolation**: worktree
- **File ownership**:
  - `internal/runtime/audit_*.go` 전체
  - `internal/cli/run_audit_gate_*_test.go` 전체
- **의존성**: T-14
- **Blocks**: T-16
- **설명**: 중복 헬퍼 추출, 함수 네이밍 일관성, 단위/통합 테스트 모두 GREEN 확인. `go test -race ./...` 통과 확인.
- **DoD**:
  - [ ] **Tested**: `go test -race -tags=integration ./internal/...` 전체 GREEN
  - [ ] **Readable**: exported 심볼 100% godoc, 함수 평균 LOC ≤ 30
  - [ ] **Unified**: `gofmt -l ./...` empty, `golangci-lint run ./...` zero error
  - [ ] **Secured**: `go vet ./...` clean, 입력 검증 모든 진입점에 적용
  - [ ] **Trackable**: REFACTOR 단일 커밋 (TDD 사이클 보존)

---

## Phase F: 템플릿 동기 + dogfood + grace window 개시

### `T-16` Template-First 최종 동기 + `make build` 검증
- **REQ 매핑**: 정합성
- **AC 매핑**: 사전조건
- **Owner-role**: implementer
- **Isolation**: 불필요 (build 작업)
- **File ownership**:
  - 모든 트윈 파일 (Phase A-D에서 변경한 파일)
  - `internal/template/embedded.go` (자동 생성)
- **의존성**: T-15
- **Blocks**: T-17
- **설명**: 모든 영향 파일이 `.claude/`와 `internal/template/templates/.claude/` 사이 byte-level 일치 확인. `make build && make install` 후 embedded 재생성 검증.
- **DoD**:
  - [ ] **Tested**: `go test ./internal/template/...` 전체 PASS
  - [ ] **Readable**: 변경 파일 목록을 `git diff --name-only` 결과로 PR 본문에 첨부
  - [ ] **Unified**: 모든 트윈 쌍 `diff` 결과 empty
  - [ ] **Secured**: embedded 재생성 시 path traversal 없음 (`go:embed` 신뢰 경로)
  - [ ] **Trackable**: 커밋 메시지 `chore(templates): sync plan-audit gate (SPEC-WF-AUDIT-GATE-001 T-16)`

### `T-17` dogfood self-audit 실행 + 보고서 생성
- **REQ 매핑**: dogfood
- **AC 매핑**: `AC-WAG-11`
- **Owner-role**: implementer + reviewer
- **Isolation**: 불필요
- **File ownership**:
  - `.moai/reports/plan-audit/SPEC-WF-AUDIT-GATE-001-2026-04-25.md` (자동 생성)
- **의존성**: T-16
- **Blocks**: T-18
- **설명**: 본 SPEC 디렉터리를 plan-auditor에 입력하여 self-audit 실행. verdict=PASS 확인, 보고서 자동 생성. dogfood 통합 테스트 `TestSelfAuditPassesOnOwnSpec` 통과.
- **DoD**:
  - [ ] **Tested**: `go test -tags=integration -run TestSelfAuditPassesOnOwnSpec ./internal/cli/...` PASS
  - [ ] **Readable**: 보고서가 verdict 명확히 PASS 표기 + 4 must-pass 모두 통과
  - [ ] **Unified**: 보고서 형식이 다른 SPEC 보고서와 일치
  - [ ] **Secured**: 보고서 파일 경로가 `.moai/reports/plan-audit/` 내부
  - [ ] **Trackable**: 보고서가 git status에 표시되며 추적 가능 (`.gitkeep` 외 첫 번째 보고서로 PR에 첨부)

### `T-18` Grace window 개시 + status 전환 + CHANGELOG 업데이트
- **REQ 매핑**: §6 마이그레이션
- **AC 매핑**: `AC-WAG-08` 의 시간 기준점
- **Owner-role**: implementer
- **Isolation**: 불필요
- **File ownership**:
  - `.moai/state/audit-gate-merge-at.txt` (T0 기준점)
  - `CHANGELOG.md`
  - `.moai/specs/SPEC-WF-AUDIT-GATE-001/spec.md` (status 전환)
- **의존성**: T-17
- **Blocks**: 없음 (최종 task)
- **설명**: (a) merge timestamp `T0`를 `.moai/state/audit-gate-merge-at.txt`에 ISO-8601로 기록. (b) CHANGELOG.md에 `[Unreleased]` → `## SPEC-WF-AUDIT-GATE-001 — Plan Audit Gate (grace window 7d)` 항목 추가. (c) spec.md frontmatter `status: draft → implemented` 전환.
- **DoD**:
  - [ ] **Tested**: T0 파일 존재 + grace window 테스트가 T0 읽어 동작 확인
  - [ ] **Readable**: CHANGELOG 항목이 사용자 영향(7일 grace, --skip-audit, INCONCLUSIVE)을 평이한 한국어로 설명
  - [ ] **Unified**: status 전환은 frontmatter 한 필드만 수정 (FROZEN 본문 변경 없음)
  - [ ] **Secured**: T0 파일은 평문 ISO-8601만 포함 (인젝션 가능 콘텐츠 없음)
  - [ ] **Trackable**: 단일 commit + tag (예: `audit-gate-merge`) 제안 (선택)

---

## REQ → Task 추적성 매트릭스

| REQ | 1차 Task | 보강 Task | 검증 Task |
|-----|---------|---------|---------|
| `REQ-WAG-001` | T-04 | T-02 | T-09, T-13 |
| `REQ-WAG-002` | T-04 | T-08 | T-09, T-13 |
| `REQ-WAG-003` | T-04 | T-14 | T-09, T-13, T-14 |
| `REQ-WAG-004` | T-01, T-04 | T-14 | T-11, T-14 |
| `REQ-WAG-005` | T-05 | T-03 | T-12, T-13 |
| `REQ-WAG-006` | T-07 | T-04 | T-09, T-13 |
| `REQ-WAG-007` | T-08 | — | T-09, T-13 |
| dogfood | T-17 | — | T-12, T-17 |
| Template-First | T-01, T-02, T-03, T-04, T-05, T-07, T-08 | — | T-16 |

전체 REQ 7건 + dogfood + Template-First → Task 18개 100% 매핑.

---

## AC → Task 추적성 매트릭스

| AC | RED Task | GREEN Task | 최종 검증 Task |
|----|---------|-----------|--------------|
| `AC-WAG-01` | T-09 | T-13 | T-15 |
| `AC-WAG-02` | T-09 | T-13 | T-15 |
| `AC-WAG-03` | T-09 | T-13, T-14 | T-15 |
| `AC-WAG-04` | T-11 | T-14 | T-15 |
| `AC-WAG-05` | T-12 | T-13 | T-15 |
| `AC-WAG-06` | T-09 | T-13 | T-15 |
| `AC-WAG-07` | T-09 | T-13 | T-15 |
| `AC-WAG-08` | T-10 | T-13 | T-15 |
| `AC-WAG-09` | T-11 | T-14 | T-15 |
| `AC-WAG-10` | T-11 | T-14 | T-15 |
| `AC-WAG-11` | T-12 | T-17 | T-17 |

---

## 작업 순서 (Critical Path)

```
T-01 → T-02 → T-03
                ↓
        ┌──── T-04 ────┐
        │              │
        T-05      [둘 다 완료 후]
        │              │
        └──── T-06 ────┘
                ↓
              T-07
                ↓
              T-08
                ↓
        ┌──┬──┬──┐
        T-09 T-10 T-11 T-12  (병렬 — RED 단계)
        └──┴──┴──┘
                ↓
              T-13
                ↓
              T-14
                ↓
              T-15
                ↓
              T-16
                ↓
              T-17
                ↓
              T-18
```

병렬 처리 가능 구간: G-B (T-04, T-05), G-E (T-09 ~ T-12).

---

## 작업량 추정 (priority-based, no time)

- Phase A (T-01 ~ T-03): Priority Critical — 후속 모든 작업의 사전조건
- Phase B (T-04, T-05): Priority Critical — 게이트 본문, 사용자 visible
- Phase C (T-06): Priority High — skill 정합성 검증
- Phase D (T-07, T-08): Priority High — 우회/실패 경로 명세
- Phase E RED (T-09 ~ T-12): Priority Critical — TDD 사이클 시작
- Phase E GREEN (T-13 ~ T-15): Priority Critical — Go 구현 핵심
- Phase F (T-16 ~ T-18): Priority Critical — Template-First + dogfood + 출시

총 **18 task**, 모두 SPEC-WF-AUDIT-GATE-001 추적성 보장.
