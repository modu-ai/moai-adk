# Sync 감사 보고서 — SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 (카드 t637)

- 감사자: sync-auditor (독립·회의적 평가)
- 대상 트리: `.claude/worktrees/t637`, 브랜치 `WT-acquire-branch-record`, HEAD `0499c861d`
- 구현 기준점: `4fcd2932e` (구현 직전 develop 흡수 병합)
- 하네스: Tier M → standard, 평가 프로필 `default`(flat 가중 백분율), 필수 통과 차원 = Functionality·Security
- 감사일: 2026-09-12

## 판정

**Overall Verdict: PASS** — 가중 조화평균 **90.9 / 100** (비가중 조화평균 89.6)

필수 통과 두 차원(Functionality, Security)이 모두 기준을 넘었고, 차단(blocking) 결함은 없다. 발견 사항 10건은 전부 비차단이며, 그중 문서 정확성 4건(F1-F4)은 병합 전에 고치기를 권한다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | AC-ILT-001..014 전부 관측 PASS. 슬롯 증거가 HEAD에 그대로 귀속됨(`git diff --stat f680dab46 HEAD -- internal/ cmd/` → 출력 없음). 픽스처 C1′/C2′/C3′ 출력이 기대값과 글자 단위로 일치. 가장자리(빈 `--branch`, 거절된 acquire, 구 레코드, 비-manual 모드) 모두 테스트로 고정 |
| Security (25%) | 92/100 | PASS | Critical/High 0건. 새 외부 입력 경로·셸 호출 없음. 경고는 `%q`로 분기명을 이스케이프. 저위험 1건(F9, 상태 텍스트의 card 원문 출력) |
| Craft (20%) | 85/100 | PASS | 설정 시접 함수 커버리지 100%(아래 증거). 뮤턴트 13+4행 전부 단언 수준에서 kill. `internal/cli` 커버리지 수치는 미측정(Gap). progress.md의 사실과 다른 서술 1건(F1)과 낡은 서술(F4) |
| Consistency (15%) | 87/100 | PASS | `omitempty` 선택 필드 패턴(PIDSource와 동일), 경고 접두사 공유, 템플릿 미러 바이트 동일·중립. 낡은 주석(F5), CHANGELOG 수치·표현(F2·F3), 코드블록 정렬(F6) |

가중 조화평균 = 1 / (0.40/95 + 0.25/92 + 0.20/85 + 0.15/87) = 90.9

## 요구사항별 확인

| 항목 | 결과 | 근거 |
|---|---|---|
| 운영자 결정 4항목(verdict.md §8: ① card 줄 ② branch_source + 경고 1줄, 거부 없음 ③ `--branch` 도움말 ④ 문서 `--card`) | 전부 전달 | 아래 각 행 |
| 범위 밖 미구현 | 준수 | `acquire --json` 객체 무변경, guard 무편집(`git diff --stat 4fcd2932e HEAD -- internal/hook/` → 출력 없음), 거부·게이트 추가 없음, 레코드 마이그레이션 없음 |
| t449 해석 순서 flag → config → caller | 보존 | `resolveIntegrationTarget`는 반환값 하나(source)만 추가. slot-016a 회귀 10개 PASS |
| `branch` 필드 의미(통합 TARGET) | 보존 | 필드명·값 무변경. source는 같은 trim 값에서 결정 |
| guard 비결정성(REQ-ILT-013) | 보존 | guard 파일 무편집, hook 통합락 테스트 14 PASS / 0 FAIL(이번 감사에서 재실행) |
| 구 레코드 호환(REQ-ILT-005) | 충족 | `TestIntegrationStatus_OldRecordKeepsTodaysBranchLine` — 텍스트 줄·JSON 키 부재·파일 바이트 불변을 모두 단언. 뮤턴트 f·l로 비공허성 확인 |
| 경고: stderr 전용·1회·manual+git-flow+빈 develop의 caller 폴백에서만·거절 시 없음 | 충족 | 조건 `source == caller && gitFlow.IsGitFlow()`가 레코드 기록 성공 뒤에만 평가됨(`integration.go:299-310`). 뮤턴트 a·b·e·g·i1·i2·k 전부 kill |
| 문서 8곳 | 충족 | 이번 감사 재측정 `8 / 0 / 1 / 0`, 증거 파일 `ac-011.txt`와 `diff` rc=0 |
| 템플릿 미러 바이트 동일·중립 | 충족 | `diff-rc=0`, 추가 줄 1, 중립성 적중 0, 양성 대조 1 |
| CHANGELOG 정확성 | 대체로 정확, 2곳 부정확 | F2, F3 |
| SPEC status·§E.4 수명주기 | 정상 | `draft→in-progress`(71480d586), `in-progress→completed`(38c058c09, status·updated만), `sync_commit_sha` 백필(0499c861d). `moai spec lint` → No findings |
| @MX 의무 | 의무 없음, 단 서술 오류 | 새 공개 함수 `LoadGitFlowIntegrationConfig`의 비테스트 호출자 2곳(fan_in < 3) → ANCHOR 불요, NOTE는 "고려" 수준. 그러나 §E.4의 "새 공개 함수 없음" 서술은 사실이 아님(F1) |

## 슬롯 증거에 대한 비판적 검토

- **뮤턴트 FAIL의 출처**: 13개 cli 행과 설정 3행·kanban 1행 모두 `_test.go:<line>:` 단언 메시지 뒤에 `--- FAIL`이 찍혀 있고, `build failed` / `undefined` / `syntax error` 흔적은 0건이다. 빌드 오류로 인한 가짜 kill은 없다.
- **실패 메시지와 선언된 변이의 정합**: d는 flag·config 모두 `"caller"`, e는 stdout에 경고가 섞여 JSON 디코드 실패, h는 blank_flag에서 `"flag"`, l은 구 레코드 JSON에 `"branch_source":""`가 나타난다 — 각 메시지가 선언된 한 줄 변이와 정확히 맞는다. 다만 적용된 변이 diff 자체는 증거 파일에 남아 있지 않다(F8).
- **원복 확인**: `restored-pass.txt` → PASS 20줄(최상위 14 + 하위), FAIL 0, `ok internal/cli`.
- **선택자 공허성**: slot-001..010, 016a 모두 `internal/cli` 줄이 `ok`이며 명명된 테스트 수만큼 `--- PASS`가 있다(003: 상위 1 + flag·blank_flag, 008: 2 + text·json).
- **픽스처 기대값 대조**:
  - C1′: `rc=0`, warn-count `1`, 경고가 `"WT-a"`를 명시, `  card:     tA`, `  branch:   WT-a (source: caller)`, worktree `/tmp/t637-fx-wt/cardA`, JSON `"card":"tA"`, `"branch_source":"caller"` — 일치
  - C2′: warn-count `1`(`WT-a`), `  card:     tB` + `  branch:   WT-a (source: caller)` — 일치
  - C3′: warn-count `0`, stderr 빈 파일, `  branch:   develop (source: config)`, worktree `/private/tmp/t637-fx-wt/develop`, JSON `"branch_source":"config"` — 일치
  - 실제 창 가드: before/after 해시 `063ae13b46b700f78ae34e9418bac488933033bf` 동일
- **귀속**: 슬롯은 HEAD `f680dab46`에서 돌았고, 이후 커밋 3개(c6539c2bf 증거, 38c058c09·0499c861d 문서)는 `internal/`·`cmd/`를 건드리지 않는다. 따라서 슬롯 증거는 현재 HEAD에 유효하다.

## 이번 감사에서 직접 실행한 명령 (발췌, 원문 출력)

```
$ go vet ./internal/cli/ ./internal/kanban/ ./internal/config/   → vet-rc=0
$ gofmt -l internal/cli internal/kanban internal/config          → (출력 없음) gofmt-rc=0
$ go test ./internal/config/... -run '^(TestLoadGitFlowDevelopBranch|TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch)$' -count=1 -v
  config-rc=0 · --- PASS 13 · --- FAIL 0
  --- PASS: TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch/{a..e} 5개 전부
$ go test ./internal/kanban/... -run 'IntegrationLock' -count=1 -v
  kanban-rc=0 · --- PASS 17 · --- FAIL 0
  --- PASS: TestIntegrationLock_BranchSourceRoundTripsAndOmitsWhenEmpty (0.00s)
$ go test ./internal/hook/... -run 'IntegrationLock' -count=1 -v  → hook-rc=0 · PASS 14 · FAIL 0
$ go test ./internal/template/... -run '^TestTemplateNoInternalContentLeak$' -count=1 -v
  tmpl-rc=0 · --- PASS: TestTemplateNoInternalContentLeak (0.57s)
$ go test ./internal/config/ -run '<위 두 테스트>' -coverprofile=… ; go tool cover -func
  loader_integration_branch.go:34: LoadGitFlowDevelopBranch      100.0%
  loader_integration_branch.go:57: IsGitFlow                     100.0%
  loader_integration_branch.go:64: LoadGitFlowIntegrationConfig  100.0%
$ git diff --stat 4fcd2932e HEAD -- internal/hook/               → (출력 없음)
$ git diff --stat f680dab46 HEAD -- internal/ cmd/               → (출력 없음)
$ diff -q <local kanban-dispatch.md> <template mirror>           → diff-rc=0
$ CMD-ILT-015 줄 2-4                                              → 1 / 0 / 1
$ CMD-ILT-014 재측정                                              → 8 / 0 / 1 / 0 (증거 ac-011.txt와 diff rc=0)
$ moai spec lint .moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001/spec.md
  ✓ No findings — all SPEC documents are valid
```

`no tests to run` 문구는 config·hook·template 출력에서 명명 테스트가 없는 하위 패키지(`toolpolicy`, `hook/trace` 등) 줄에서만 나왔고, 명명 테스트가 있는 패키지는 모두 `ok`와 해당 `--- PASS` 줄을 가진다.

## 발견 사항 (structured defect-list)

| id | 심각도 | 차단 | 확신 | 위치 | 결함 | 필요한 수정 |
|---|---|---|---|---|---|---|
| F1 | Medium | 비차단 (병합 전 수정 권장) | 높음 | `.moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001/progress.md` §E.4 "MX tags" 문단 | "no new exported function … was introduced by this SPEC's run-phase"는 사실이 아니다. 새 공개 기호: `config.LoadGitFlowIntegrationConfig`, `config.GitFlowIntegrationConfig`, `(GitFlowIntegrationConfig).IsGitFlow`, `kanban.BranchSourceFlag/Config/Caller`. 결론(태그 불요)은 맞지만 전제가 틀린 sync 신호다 | 문단을 "새 공개 기호 N개 도입; `LoadGitFlowIntegrationConfig` 비테스트 호출자 2곳으로 fan_in < 3 → @MX:ANCHOR 불요; godoc이 계약을 설명하므로 @MX:NOTE 미추가"처럼 측정 사실로 다시 쓴다 |
| F2 | Low | 비차단 (수정 권장) | 높음 | `CHANGELOG.md` 해당 항목 "The five documented `acquire` invocations" | 호출 지점은 8곳(AC-ILT-011 측정값), 파일이 5개다. 수치가 동작·증거와 어긋난다 | "The eight documented `acquire` invocations across five files"로 고친다 |
| F3 | Low | 비차단 (수정 권장) | 높음 | `CHANGELOG.md` 해당 항목 "no `--branch` was given" | 실제 조건은 "non-blank `--branch`가 없을 때"다. `--branch "   "`도 경고를 낸다(REQ-ILT-006 문구도 non-blank) | "no non-blank `--branch` was given"으로 고친다 |
| F4 | Low | 비차단 (수정 권장) | 높음 | `progress.md` 4행 헤더, §E.1 "rows a-f", §E.2 도입부·§E.2.5 "NOT observed" | 헤더가 여전히 "Status: in-progress (run-phase, pre-slot done)"이고, §E.2.5는 §E.3에서 해소됐다는 표시가 없다. §E.1의 "rows a-f"는 v0.1 시점 값이다 | 헤더를 completed로 갱신하고 §E.2.5에 "§E.3에서 해소" 한 줄을 붙인다 (이력 문단은 그대로 둔다) |
| F5 | Low | 비차단 | 높음 | `internal/cli/integration.go:105-108` (resolveIntegrationTarget godoc) | 설정 브랜치 출처로 `LoadGitFlowDevelopBranch`를 지목하지만 acquire는 이제 `LoadGitFlowIntegrationConfig`를 호출한다. `LoadGitFlowDevelopBranch`는 비테스트 호출자가 0곳이 됐다(삭제는 범위 밖이므로 유지가 맞다) | 주석의 참조를 `LoadGitFlowIntegrationConfig(...).DevelopBranch`로 바꾼다. 함수 존치 여부는 별도 판단 |
| F6 | Low | optional | 높음 | `.claude/rules/local/gitflow-lane-protocol.md` 51행 코드블록 | `--card <card-id>` 삽입으로 주석 열 정렬이 다음 두 줄과 어긋났다 | 세 줄의 `#` 열을 맞춘다 (미관) |
| F7 | Info | optional | 중간 | `gitflow-lane-protocol.md:47` | `--card`와 함께 `--name <lane>`도 새로 들어갔다(원래 맨 `acquire`). 호출 안의 변경이라 §E 위반은 아니나 REQ-ILT-010이 요구한 것 이상이다 | 조치 불요. 의도라면 그대로 둔다 |
| F8 | Low | optional | 높음 | `.moai/reports/t637/ac-evidence/mut-*.txt` | 각 파일에 테스트 출력만 있고 적용한 변이 diff는 없다. 변이의 정체는 실패 메시지로 추론 가능하며 13행 모두 정합하지만, 직접 증거는 아니다 | 이후 슬롯에서는 행마다 `diff <backup> <mutated>` 출력을 같은 파일 머리에 남긴다 |
| F9 | Low | optional | 낮음 | `internal/cli/integration.go:231-232` | `status` 텍스트가 `lock.Card`를 원문 그대로 찍는다. 개행·ANSI 제어문자를 담은 card 값이면 가짜 상태 줄을 만들 수 있다(경고는 `%q`로 안전). 값은 로컬 운영자 입력이고 `branch`·`worktree`도 이미 원문 출력이라 새로운 신뢰 경계는 아니다 | 선택: `%q` 또는 제어문자 제거. 추측성 보강이므로 자동 수정 대상 아님 |
| F10 | Info | optional | 중간 | `internal/cli/integration.go:309-310` | 호출자 분기를 얻지 못하면(`currentBranch()`가 `""`) 경고가 `branch ""`로 나간다. git 저장소 밖 실행이라는 드문 경로 | 조치 불요 (관찰 기록) |

차단 결함 0건. 심각도 분포: Medium 1 · Low 7 · Info 2.

## 권고

- 병합 전: F1-F4를 한 번의 문서 커밋으로 고친다. 모두 문서 정확성 문제이며 코드·테스트 재측정이 필요 없다.
- F5는 다음에 `integration.go`를 만질 때 함께 고친다(주석 한 줄).
- F6-F10은 재량 사항이다. 자동으로 수정 경로에 넣지 않는다.

## Gaps (관측하지 못한 것)

- `internal/cli` 테스트·빌드·린트는 이번 감사에서 다시 돌리지 않았다(컴파일 슬롯 반납, HARD 제약). 해당 판정은 슬롯 증거(HEAD `f680dab46`, 이후 `internal/`·`cmd/` 무변경 확인)에 귀속된다.
- `internal/cli` 변경분의 커버리지 수치는 측정하지 못했다. 새 분기마다 테스트가 있고 뮤턴트로 비공허성이 확인됐지만, 백분율은 미관측이다.
- 교차 모델 감사(`audit_multi`/codex/GLM)는 돌리지 않았다. 백엔드가 받는 대상이 `uncommittedChanges`(없음) 또는 `baseBranch`(원격 기본 head/`main` 기준 — develop 흡수분까지 섞인 다른 diff)뿐이라, 이 카드의 변경(`4fcd2932e..HEAD`)을 겨눌 수 없다.
- 전체 스위트와 크로스 플랫폼 빌드는 CI 몫이다. develop push 이후의 CI 판정은 관측 전이다.

## Residual-risk (관측했음에도 남는 위험)

- `CLAUDE.local.md` 세 줄 편집(이 트리 370/389/403)은 primary 체크아웃의 미커밋 사본(338/357/371)과 줄 위치가 다르다. 이 카드의 develop 병합에서는 충돌하지 않지만, primary 작업이 main에 들어가거나 release PR이 develop 사본을 옮길 때 충돌할 수 있다.
- 경고 억제는 git-strategy 파일 해석에 기댄다. `moai update`가 manual 블록을 템플릿 기본값(github-flow)으로 되돌리면 경고 조건(`IsGitFlow`)이 거짓이 되어 경고 없이 caller 폴백한다 — 다만 이제 `status`에 `(source: caller)`가 찍히므로 폴백 자체는 보인다.
- 슬롯 기록이 남긴 관찰(슬롯 도중 실제 락 파일이 `lane-8` 보유로 생겨남, 리드 공지와 불일치)은 이 SPEC의 결함이 아니지만 리드가 확인할 사안이다. 픽스처가 실제 창을 건드리지 않았다는 사실은 해시 동일로 확인됐다.
