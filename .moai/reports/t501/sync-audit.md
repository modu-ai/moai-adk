# SPEC-CODEX-TEST-GAPS-001 — sync-audit (card t501, lens --deep)

- 감사자: sync-auditor (독립 재판정 — 실행 단계 보고를 신뢰하지 않고 전수 재측정)
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501` — 브랜치 `WT-codex-uncovered`, HEAD `82674b94c` (측정 전후 불변 확인)
- Diff base: `24df2ae45` (plan-phase 수정 라운드) → HEAD `82674b94c`
- 감사일: 2026-09-07

## Evaluation Report

SPEC: SPEC-CODEX-TEST-GAPS-001
Overall Verdict: **PASS** (harmonic mean 97.9/100 — must-pass 차원 Functionality·Security 모두 독립 통과)

### Dimension Scores

| Dimension | Score | Verdict | Evidence (이번 실행, 이 트리에서 직접 재측정한 값) |
|-----------|-------|---------|----------|
| Functionality (40%) | 100/100 | PASS | 8개 신규 테스트 전부 재실행하여 `--- PASS` 직접 관측: `--- PASS: TestTerminateCodexProcessHelper (0.00s)` · `--- PASS: TestTerminateCodexProcess (0.00s)` · `ok github.com/modu-ai/moai-adk/internal/cli 0.715s`, `--- PASS: TestCodexIDMatches (0.00s) ok ... 2.508s`, `--- PASS: TestRealCodexConnPid (0.00s) ok ... 0.723s`, `--- PASS: TestCodexCountExecutingImports / TestCodexGatePrintf / TestDefaultCodexInitGenerator / TestAwaitCodexResponse / TestCodexSessionError` 각 1행 + `ok ... 0.744s` — `[no tests to run]` 토큰 0회(공허 스윕 아님). AC-CTG-010: `awk -F'\t' '$NF=="0.0%"'` 기준 6파일 추출(127행)의 진짜 0.0% 행은 정확히 1개 = `codex_review_gate.go:183 runCodexReviewGate`(118행 baseline 분모 밖) → 118행 중 0개. baseline 대비 4개 0.0% 함수 전부 소멸: `terminateCodexProcess 75.0%`(FindProcess 오류 분기 = §D 기록 스킵), `mcp_codex.go:494 pid 100.0%`, `mcp_codex.go:602 Error 100.0%` · `:603 Unwrap 100.0%`, `codexIDMatches 100.0%`(60% 미만 해소). GOOS=windows 테스트 컴파일 `windows-test-compile-ok rc=0` 재현 |
| Security (25%) | 100/100 | PASS | 유니언 게이트(AC-CTG-009): `git diff --name-only 24df2ae45..HEAD`의 .go 피연산자 = 테스트 파일 4개뿐, 비테스트 .go 필터 = 0, `git status --short` = 0행 → 프로덕션 보안면 변경 0. 자식 프로세스 위생: `t.Cleanup`이 자식 관측 첫 단언 앞에 등록됨(`codex_job_control_test.go` 성공 분기), 자식은 재실행 헬퍼 패턴으로 단명·정리 보장. 픽스처에 실제 시크릿 없음(`"disk full (t501 fixture)"` 등 합성값). 서브에이전트 경계 그렙: 4개 터치 파일에서 `AskUserQuestion\|mcp__askuser` 0매치(rc=1) |
| Craft (20%) | 92/100 | PASS | `golangci-lint run internal/cli/...` → `0 issues.`, `go vet ./internal/cli/` → rc=0 무출력, `gofmt -l` → 빈 출력 — 전부 본 실행 재측정. 변이체 내성: 감사자 독자 변이체(`realCodexConn.pid` 양성 분기 `return c.cmd.Process.Pid` → `return 0`)에서 `mcp_codex_test.go:659: pid with a live Process = 0, want 37089` · `--- FAIL: TestRealCodexConnPid` RED 직접 관측 후 완전 복원(`git status --short` 0행, HEAD `82674b94c` 불변, GREEN 재관측) — 테스트 공허 아님 입증. 블록 수준 검증: `mcp_codex.go:494`의 3블록 전부 count=1. 감점 요인: 패키지 전체 커버리지 80.8%는 기관 목표 85% 미만(진입 시 80.7%의 계승 baseline — 테스트 전용 단일 패키지 카드의 범위 밖, CHANGELOG/progress에 정직 공시됨) |
| Consistency (15%) | 100/100 | PASS | 4개 모두 기존 형제 테스트 파일 확장(새 파일 0) — plan.md §A가 지정한 배치. 레포 관용구 준수: 테이블 드리븐, `fakeCodexConn`/시임 재사용, `t.TempDir()`, 영어 주석, 섹션 구분 주석. CHANGELOG 항목 1건(`grep -c` = 1)이 형제 SPEC 항목과 동일 서식. diff의 마지막 커밋이 sync close(`82674b94c chore(...): sync-phase artifacts + 3-phase close`) — 닫기 뒤 코드 착지 0(t488 불변 충족) |

### Deep-lens 공격 항목 판정 (디스패치 6개 요건)

1. **테스트 재실행 + 독자 변이체** — 8/8 테스트 `-v` 셀렉터 재실행 전부 GREEN(요건 ≥3 초과). 독자 변이체 1건(양성 분기 영화) RED → 완전 복원 → 재관측 GREEN. 실행 단계가 기록한 8-변이체 표의 규율이 실재함을 독립 재현으로 확인.
2. **시임 교체 탈출구 공격** — 기각(테스트 견고). `codexTerminateProcess`는 패키지 초기화 때 `terminateCodexProcess` 함수값으로 1회 초기화되는 var이고(`codex_job_control.go:84`), M1 테스트는 함수 식별자를 직접 호출한다. var 재바인딩(기존 job-control 테스트의 관용구)은 함수값을 바꾸지 못하며, `terminateCodexProcess(pid)`가 다른 대상으로 해석되는 경로는 함수 본문 편집뿐 — 그것은 프로덕션 diff라 유니언 게이트가 커밋 시점에 잡는다. 직접 호출 구축은 구조적으로 변이 내성적. 프로덕션 호출부는 시임 경유 1곳(`codex_job_control.go:258`)으로 확인.
3. **pid 66.7% 공격** — **디스패치 전제 정정.** 66.7%는 pid 테스트의 대상이 아니다. 테스트 대상 `(realCodexConn).pid`(mcp_codex.go:494, 3분기)는 블록 수준 100.0%(3블록 전부 count=1) — 미커버 분기 없음. 66.7%는 **다른 리시버** `(codexSessionHandle).pid`(mcp_codex.go:701)로, baseline에서도 66.7%(변화 0), 블록 `702.31,704.3`(nil 가드 `return 0` 분기) count=0. 해당 분기는 닐 리시버 호출(`var h *codexSessionHandle; h.pid()`)로 밀폐 환경 커버 가능하나, 이 SPEC이 이름 붙인 REQ가 없고(§F 범위 밖) AC-CTG-010 분모(0.0% = 0)도 충족하므로 결함 아님 — 후속 후보 권고(F1).
4. **CHANGELOG 정직성** — 검증 통과. "tests-only": 유니언 게이트 재측정 EMPTY(피연산자 4개 테스트 파일로 비어있지 않음을 먼저 증명). "80.7→80.8": baseline 80.7%는 plan-phase 측정치로 spec.md §B.2에 자체 verbatim 명령+출력과 함께 귀속돼 있고, 신규 80.8%는 `.moai/state/verify/t501/cover-after.log` verbatim(`ok ... 337.374s coverage: 80.8% of statements` / `rc=0`). "zero 0.0% among the 118": awk 재측정으로 정확히 참(127행 중 유일한 0.0% = 분모 밖 runCodexReviewGate). "8 new tests across 4 test files": diff에 정확히 8개 테스트 함수 존재 확인. 측정이 뒷받침하지 않는 수치 없음.
5. **유니언 게이트 + 마지막 쓰기** — 둘 다 EMPTY 재측정. 커밋 순서: `cd855f296`(iter2 보고) → `e652acf40`(§F) → `ffa06117f`(M1) → `1269e6d8d`(M2-M7) → `c79428b9c`(run 증거) → `82674b94c`(sync close) — 닫기가 마지막 커밋.
6. **SPEC 닫힘 정합성** — `git diff 24df2ae45..HEAD`의 spec/plan/acceptance diff는 `status: draft → completed` 한 줄뿐(plan/acceptance diff 0). §E.4 `sync_commit_sha: "pending-backfill-sync"`는 D3 승인 플레이스홀더(후속 커밋 백필 예정 부채명시). §A History v0.1.0/v0.2.0 양측 intact.

### Findings (구조적 결함 목록)

- **F1 [Advisory] [optional]** `internal/cli/mcp_codex.go:701-706` — `(codexSessionHandle).pid` 66.7%: nil-가드 `return 0` 분기(블록 702.31,704.3) 미커버. 닐 리시버 호출로 밀폐 커버 가능하나 이 SPEC의 REQ가 이름 붙이지 않았고(§F 범위 밖), 0.0%가 아니며 baseline 대비 변화 0. **신뢰도: high**(블록 데이터 직접 측정). Required fix: 이 카드의 수정사항 아님 — 2줄 테스트 추가는 후속 카드 후보로 기록만.
- **F2 [Advisory] [optional]** `internal/cli/codex_job_control_test.go:726-731` — AC-CTG-001의 "t.Cleanup before the first assertion" 문언과 구현의 정렬: 등록은 거절 분기 단언들(자식이 아직 존재하지 않아 유출 불가능) 뒤, 자식이 존재한 뒤의 모든 단언 앞. 누출-안전 실질은 충족(plan.md §F M1이 성공-분기 국지 정렬로 정의). **신뢰도: high**(테스트 본문 직독). Required fix: 없음 — 문언을 성공-분기 기준으로 읽는 것이 plan의 정의이며, 후속 SPEC에서 문구 정밀화 여지.
- **F3 [Advisory] [optional]** 패키지 전체 — `internal/cli` 80.8% < 기관 목표 85%, 전체 패키지 진짜 0.0% 함수 76개(그중 codex 파일 1개 = runCodexReviewGate). 전부 이 카드 이전의 계승 baseline이고 테스트 전용 단일 패키지 카드 범위 밖; CHANGELOG가 정직하게 공시. **신뢰도: high**(awk 전수 측정). Required fix: 본 카드 소관 아님 — 후속 카드로만. (디스패치 지시에 따라 runCodexReviewGate를 SPEC 결함으로 재제기하지 않음.)

**Blocking 0건 / Should-fix 0건 / Advisory 3건.** must-pass 방화벽(Functionality + Security) 양쪽 독립 통과 — 전체 PASS.

### Recommendations

- F1의 `(codexSessionHandle).pid` nil-분기 2줄 테스트는 다음 테스트-보강 성격 카드에서 무비용으로 흡수 가능(REQ 명시 없이는 이 카드 범위 밖).
- §E.4의 `pending-backfill-sync` 플레이스홀더는 후속 커밋에서 실제 SHA로 백필할 부채 — 리드가 병합 커밋 확정 후 백필 커밋을 잊지 않을 것(스키마 D3 절차).

### Gaps (명시적으로 관측하지 않은 것)

- 실행 단계가 기록한 8개 변이체의 당시 RED 출력은 과거 일회성 사건이라 재관측 불가 — 진행 기록 표의 verbatim 인용과 독자 변이체 1건의 성공적 재현으로 간접 검증(전수 재현은 시도하지 않음).
- `go test ./internal/cli/` 전체 스위트를 본 감사에서 재실행하지 않음(레인 부하 규율 + 337s 소요) — 패키지 `ok`/커버리지 수치는 실행 단계 로그(`cover-after.log`)와 감사자 셀렉터 실행으로 교차 확인. `go test ./...`·CI 판정은 원격 CI 몫.

### Residual-risk

- 본 감사의 셀렉터 GREEN은 darwin 로컬 환경 기준 — windows는 테스트-바이너리 컴파일까지만 검증됨(런타임 거동은 CI 매트릭스 몫).
- 시임 var(`codexTerminateProcess`)를 교체한 채로 M1 테스트와 기존 job-control 테스트가 같은 패키지에서 동시에 도는 상호작용은 존재하지 않음이 코드 구조상 확인됐으나, 실행 시점 직렬성은 go test의 순차 실행에 의존한다(병렬 선언 없음 확인).
