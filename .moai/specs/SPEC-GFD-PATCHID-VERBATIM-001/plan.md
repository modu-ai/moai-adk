# SPEC-GFD-PATCHID-VERBATIM-001 — plan

> card t1561 (P1, data-destruction class) · branch `WT-landing-patchid` · HEAD 기저 `cad44a751`

## §A Context

- **A.1 위치**: 카드 워크트리 `.moai/worktrees/t1561`, 브랜치 `WT-landing-patchid`, HEAD `cad44a75163b6f0f056551354ef8008fe799090f` = `origin/main` 분기 시점 일치.
- **A.2 대상 파일 (4개 — 수리 2 + 테스트 2)**:
  - `internal/cli/worktree/landing_predicate.go` — 수리 대상. `landingPatchIDs()`(:92-104)의 고정 `--stable` 인자, 지원 감지 시임 신설, 미지원 시 오류 경로, 커밋별 공백 충실 동치 술어의 신설 공개(세션 ②팔이 쓴다).
  - `internal/cli/worktree/landing_predicate_test.go` — 신규 테스트 추가만. RED 테스트(`:554`)의 단언은 고정.
  - `internal/cli/session_worktree.go` — 수리 대상(제2 결함 지점). `gitBranchLandedReal` ②팔(`git cherry`, :890)을 공백 충실 동치로 교체 — ①·③팔과 조기 반환 형상 유지, 미지원 git 은 오류(호출자 preserve, REQ-WSS-302 계약 그대로).
  - `internal/cli/session_worktree_landing_test.go` — 세션 종료 경로 신설 관문 1건 추가(기존 `TestCleanupSessionWorktree_*` 단언 무변경).
- **A.3 RED 상태**: `TestLandingPredicateWhitespaceDivergenceKeepsTree` 가 이 트리에서 실패 중(`.moai/reports/t1561/red-reproduction.md`). `quality.yaml` `development_mode: tdd` — run 단계는 RED 관문 확인 → GREEN → 보강 순서.
- **A.4 판정 명령(수정 패키지 스코프)**: `go test ./internal/cli ./internal/cli/worktree/ -count=1 -timeout 30m`. 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다 — CI 몫이다. `internal/cli` 뿌리 패키지 스위트는 분 단위의 무거운 실행이므로 M3 배치 전 `moai slot acquire --resource <이름> --max-duration <상한>` 으로 자원 임대를 먼저 잡는다(레인-로컬 규율).
- **A.5 PRESERVE**: 기존 테스트 5함수(`TestLandingPredicateSquashSafe` F1~F9, `LaterChangeObservation`, `CommitCap`, `GHReadsTheCardBranch`, `GHFailureIsFailClosed`)의 단언 + RED 관문 1건(본 카드 신설, 단언 고정) + 세션 쪽 기존 `TestCleanupSessionWorktree_*` 5함수의 단언 · 호출자 계약(`landedBeyondAncestry`/`LandedByPatchID` 서명과 의미) · 세션 술어의 ①팔(조상)·③팔(공유 누적 patch-id)과 ②팔의 커밋별 동치 확인력(개별 커밋 착지 케이스) · `landingPatchIDCommitCap`/`landingGHTimeout` 값 · 파일 헤더의 fail-closed 계약 문장.

## §B Known Issues (관련 축만)

- **B5 CI 3-tier**: spec-lint·golangci-lint·Test 는 각자 따로 죽는다. 수리 전 기준선(vet·lint 두 패키지 클린 — acceptance.md §B 셀, RED 테스트 1건 적색, 세션 신설 관문 부재)과 NEW 결함을 구분해 보고한다.
- **B8 워킹 트리 위생**: 커밋은 명시 pathspec 스테이징(`git add internal/cli/worktree/landing_predicate.go internal/cli/worktree/landing_predicate_test.go internal/cli/session_worktree.go internal/cli/session_worktree_landing_test.go`). `.moai/state`·런타임 파일 편집 금지. 현재 트리의 ` M internal/cli/worktree/landing_predicate_test.go`(RED 테스트)는 카드 자산 — 유지한다.
- **B10 범위 규율**: §A.5 PRESERVE 목록 밖으로 손대지 않는다. 인접 파일(sweep/done 본문)은 호출자로서 읽기만 한다.
- **B11 블로커 프로토콜**: 사용자 결정이 필요해지면 AskUserQuestion 대신 구조화 블로커 보고.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse HEAD        # WT-landing-patchid @ cad44a751…
go test ./internal/cli/worktree/ -run '^TestLandingPredicateWhitespaceDivergenceKeepsTree$' -count=1 -timeout 30m -v
                                                        # RED 확인 — 실패가 정상 (want preserve 관측)
git status --short                                      # M internal/cli/worktree/landing_predicate_test.go(RED 자산)와 untracked SPEC 산물뿐이어야 한다
```

## §D Constraints

- 파일 4개(§A.2)만. 템플릿·Makefile·catalog·config 변경 금지.
- RED 테스트 단언 무변경(주석 조정만 허용).
- 코드 주석·godoc 영어, 파일의 기존 밀도·어법 유지.
- Conventional Commits, 커밋 메시지에 `(card t1561)` 포함, 🗿 MoAI 트레일러.
- 미지원 git 에서 `--stable` 하락 경로 금지(REQ-GPV-003).
- 전체 스위트 로컬 실행 금지 — 패키지 스코프만(A.4).

## §E Self-Verification

`acceptance.md` AC-GPV-001..007 의 PASS/FAIL 행렬을 5-섹션 형식(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)으로 `progress.md` §E.2에 기재한다. 각 항은 (a) 명령 (b) 원문 출력 (c) 이번 런·이 트리+HEAD 를 함께 적는다. E8(RED 원문)은 Pre-flight 단계 관측분을 인용한다.

## §F Milestones (의사결정 역전 가능성 순 — 바뀔 결정이 앞에 온다)

### M1 — 계층 2 공백 충실 전환 (Priority High — 이 SPEC 의 유일한 설계 밀도 구간)

1. 지원 감지 시임: 지원 여부를 담는 패키지 변수 하나를 신설한다(기존 `landingGH`/`landingPatchIDCommitCap` 시임 어법과 같은 모양 — nil/미초기화면 실측, 테스트가 대입해 양쪽을 재현). 실측은 `git patch-id --verbatim` 을 실제로 실행해 관측하는 방식으로 한다 — `git --version` 문자열 파싱은 하지 않는다(빌드 변이·로케일에 부서진다).
2. `landingPatchIDs()` 의 고정 `--stable` 인자를 감지 결과가 고른 모드로 바꾼다. 카드 누적 diff 와 ref log 스트림이 같은 모드를 타는지가 핵심이며(REQ-GPV-002 양면 적용), 비교 형상·상한 로직은 그대로다.
3. 미지원 확정 시 오류 반환 — `LandedByPatchID` 의 기존 계약(오류 = cannot answer → 호출자 preserve)이 그대로 이어준다(REQ-GPV-003). `landedBeyondAncestry` 는 계층 3 을 계속 시도한다.
4. 커밋별 공백 충실 동치 술어를 worktree 패키지에 공개한다(`git cherry` 의 동치 의미를 감지 모드의 patch-id 로 재현하는 함수 — 예: `LandedByCommitPatchIDs(dir, tip, ref)`). 세션 종료 ②팔이 이것으로 `git cherry` 를 대체하고, 미지원 git 에서는 오류로 답해 `gitBranchLandedReal` 의 기존 이상 처리(오류 → 호출자 preserve, REQ-WSS-302)가 이어받는다.
- **판정**: RED 테스트 GREEN + 기존 5함수와 세션 기존 5함수 전체 GREEN(REQ-GPV-004 — §A.5 의 "기존 5함수 + RED 관문(신설)" 분리 표기 따름).

### M2 — 회귀 보강 테스트 (Priority Medium — M1 위의 기계적 보강)

1. `TestLandingPredicateVerbatimUnsupportedIsFailClosed` 신설: 시임을 미지원으로 → 하위 케이스 3개를 이 이름으로 고정한다 — `landedbypatchid_errors`(`LandedByPatchID` 오류), `sweep_preserves_without_pr`(PR 없으면 sweep preserve), `layer3_still_decides`(PR 있으면 계층 3 이 DISPOSE). 기존 픽스처 어법(`newGFDFixture` + `installGH`)을 그대로 따른다.
2. 다중 커밋 누적 공백 갈림 케이스 신설 — 이름 고정 `TestLandingPredicateWhitespaceDivergenceMultiCommitKeepsTree`: 두 커밋 카드 + squash amend 공백 변형 → preserve(F2 의 공백 형제 — RED 기록 잔여위험 "단일 시나리오 RED" 소화).
3. 세션 종료 경로 관문 신설 — 이름 고정 `TestSessionExitWhitespaceDivergencePreserves`(`internal/cli/session_worktree_landing_test.go`, 기존 `TestCleanupSessionWorktree_*` 픽스처 어법): 단일 커밋 카드 + 공백 변형 squash 착지 상태 → 세션 출구가 preserve 를 답한다(D1 제2 지점의 RED/GREEN 관문).
- **판정**: 신설 테스트 GREEN, 기존 테스트 무변경 유지.

### M3 — 검증 배치 (Priority Low — 기계 검증, 순서 고정)

1. 한 턴 병렬 일괄: `go test ./internal/cli ./internal/cli/worktree/ -count=1 -timeout 30m`(§A.4 의 두 패키지) · `go vet ./internal/cli ./internal/cli/worktree/` · `golangci-lint run ./internal/cli ./internal/cli/worktree/`(AC-GPV-003 When 와 같은 두 경로, 신규 0건) · AC-GPV-004 grep · `git diff --stat`(§A.2 의 4파일 확인).
2. `progress.md` §E.2·§E.3 기재(증거 원문 포함).

## §G Anti-Patterns

- RED 테스트를 약화해 GREEN 을 만지 않는다 — 관문의 단언은 고정이다.
- 미지원 git 에서 `--stable` 조용히 하락을 붙이지 않는다(REQ-GPV-003 위반).
- `git --version` 출력 파싱으로 기능을 감지하지 않는다 — 실제 기능 실행 관측으로 판정한다.
- 계층 1·3, 세션 술어의 ①·③팔과 호출 계약, 상한·타임아웃 값을 손대지 않는다 — 세션 쪽 예외는 ②팔(`git cherry` 동치)의 공백 충실화뿐이다.
- 전체 스위트를 로컬에서 돌리지 않는다.

## §H Cross-References

- `SPEC-GITHUB-FLOW-DEFAULT-001` REQ-GFD-002 — 착지 판정 세 계층의 창시 요구(전임, in-progress).
- `internal/cli/worktree/landing_predicate.go` 파일 헤더 — "답할 수 없는 계층은 결코 착지가 아니다" 계약.
- `.moai/reports/t1561/red-reproduction.md` — RED 재현 기록(해시 수준 프로브 + 술어 수준 실패 출력).
- `.moai/reports/t1561/plan-audit-iter1.md` — plan-audit 1회차(0.83 FAIL) 판정서 — D1-D4 수리의 근거와 증거 표(cherry 프로브·verbatim 독립 프로브 포함).
- `decision-index.md` Q1·Q2 — 이 plan 이 적용한 DEFAULT-APPLIED 결정의 원장.

🗿 MoAI
