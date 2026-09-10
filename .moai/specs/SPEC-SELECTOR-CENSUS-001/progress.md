---
id: SPEC-SELECTOR-CENSUS-001
title: "0-실행 테스트 판정 — 진행 기록"
version: "0.1.0"
status: completed
created: 2026-08-29
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "t341, progress"
tier: M
---

# 진행 기록 — SPEC-SELECTOR-CENSUS-001

카드 **t341** · 브랜치 `WT-selector-census` · 기준 트리 `a6bbbf82b`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md` · `plan.md` · `acceptance.md` (Tier M 3종) + 이 파일
- 요구 7건(REQ-SEC-001..007) / 수락 8건(AC-SEC-000..007), Tier M 상한 16 이내
- 모든 RED-now 칸이 트리 `a6bbbf82b` 에 못 박혀 있고, 사유를 함께 적었다
- 미검증 전제 3건은 `spec.md` §5 에 명시했고, 그중 살아 있는 payload 관측은 `plan.md` M0 + **AC-SEC-000** 으로 승격했다
- 상태: `draft` — plan-audit 및 Implementation Kickoff Approval 대기

### iter-1 감사 대응 (`.moai/reports/t341/plan-audit-iter1.md`, PASS-WITH-DEBT 0.81)

blocking 5건을 닫았다. 전부 기준 문면 수정이며 `internal/` 코드는 건드리지 않았다.

| 결함 | 닫은 방법 | 이 라운드에서 실행한 명령 |
|---|---|---|
| D1 결속 검사 공허 | DoD 를 두 검사(삼중점 `origin/develop...HEAD` + `git status --porcelain`)로 교체, 실측 표 첨부. `plan.md` E5 → E5a/E5b | 결속 파일 1바이트 편집 + `git add` 후 두 형태 비교 → 종전 형태 `old-grep-exit=1`(초록), `git status --porcelain` → `M  …`(붉음). 이후 `git restore --staged` + `git checkout --` 로 원상 복구 확인 |
| D2 비발화 방향 1축 | AC-SEC-003 을 러너 축 전부(go/cargo/pytest/jest·vitest, 표본 5 × payload 2 = 10건)로 확장, 뮤턴트 탐침 2건 명시 | `pytest -q` → `3 passed in 0.00s` (정밀 마커 없음). `printf 'Tests:       10 passed, 10 total\n' \| grep -c '0 passed'` → `1` (부분 문자열 충돌) |
| D3 미러 판단 탈출구 | AC-SEC-007 (2) 를 `diff` rc 0 단언으로 교체, 탈출구 삭제 | `diff` 로컬 ↔ 미러 → 출력 0줄, `diff-exit=0` |
| D4 corpus 미결속 | AC-SEC-006 에 조건 (2) 신설 — 같은 변수 공유 + 각 표본이 `detectZeroExecution` 을 실제로 발화 | `evidence_writer.go:79` 판독(`return true, false, false` — 신호 없음도 `isPass=false`), 그래서 `isPass=false` 단언만으로는 불충분 |
| D5 M0 산문 게이트 | **AC-SEC-000** 신설 + DoD 줄 + `plan.md` E0 행 + M1 진입 조건 | `ls .moai/reports/t341/live-payload.json` → `No such file or directory`, `ls-exit=1` |
| D6 인용 좌표 | `spec.md` `:223`→분기/반환 분리, `:296-330`→`:309-330`(대입 `:328`) | `grep -n 'func buildBashRecord\|rec.IsTestPass'` → `309` / `328` |

**닫지 않은 것**: D7(요구 문면의 구현 표면 이름, optional — §3.3 이 이미 근거를 담고 있어 그대로 둔다), D8(AC-SEC-003 에 RED-now 칸 없음 — 여전히 신고된 부채다. 확장 후에도 열 표본 전부 오늘 초록이며, 관측하지 않은 RED 칸을 지어내지 않는다).

### iter-2 감사 이후 추가 편집 (D9, `.moai/reports/t341/plan-audit-iter2.md`, PASS-WITH-DEBT 0.894)

**이 편집은 마지막 감사 회차(iter-2, Tier M 상한) 이후에 착지했으며 재감사를 받지 않았다.** 성격은 **순수 추가(additive)** 다 — 기존 기준·DoD·RED-now 칸을 하나도 약화·개작·삭제하지 않았다.

- `acceptance.md` AC-SEC-003 에 **표본 (f)** 추가 — 네 pass 마커를 하나도 담지 않는 진짜 pass(node 내장 러너, `npm test` 로 도달) + `{"exit_code": 0}` payload 1건이 `isPass=true` 로 남을 것. 이것이 **exit-code 축(`deriveFromExitCode`, `:69`→`:163`)의 비발화 방향**을 처음으로 고정한다. payload 총계 10 → **11**((f) 는 짝을 만들지 않는다 — 사유는 기준 본문).
- 같은 절에 **뮤턴트 탐침 3** 추가 — 거부권을 "텍스트에 실행 수 근거 없음" 으로 구현해 exit-code pass 경로를 좁히는 형태. (f) 하나만 이 뮤턴트를 죽인다.
- `acceptance.md:86` 의 래퍼 문장 교정 — `npm`·`pnpm`·`yarn` 이 (c)(d)(e) 에 흡수된다는 **일반 주장은 거짓**이다. 아래 러너가 무엇인지는 래퍼가 정하지 않는다.
- `plan.md` §F 에 대응 안티패턴 1줄 추가 — M1 이 그 형태로 흘러가는 것을 계획층에서도 막는다.
- (f) 의 마커 부재는 이 트리(`a6bbbf82b`, 2026-08-29)에서 **실측**했다(`node --test` → `node-exit=0`, 마커 5종 `grep -c` 전부 `0`). 축자 문자열의 러너 판번 고정은 (b)(d) 와 같은 규율으로 M1 몫이다.
- iter-2 의 나머지 결함(D10·D11·D7·D8)은 **optional** 이라 손대지 않았다.

## §E.2 Run-phase Evidence

실행 트리: 워크트리 `.claude/worktrees/t341`, 브랜치 `WT-selector-census`, 진입 HEAD **`744cfab5e`** (develop `ee50984ab` 흡수 완료). 명령 원출력은 `.moai/state/verify/t341/` 에 남겼다.

### M0 판독 — AC-SEC-000 의 세 질문 (부분 충족)

관측 방법과 근거는 `.moai/reports/t341/m0-live-payload-observation.md`. 이 라운드는 그 보고를 **소비**했고 다시 관측하지 않았다.

| 질문 | 답 | 어느 키/신호에서 읽었는가 |
|---|---|---|
| (a) Bash stdout 이 payload 에 실려 오는가 | **예** | 직접 키 판독이 아니라 **두 분기를 의도적으로 불일치시켜** 갈랐다. 셸 종료코드 0(`\| cat`)인데 stdout 에 `--- FAIL` 이 있는 호출이 `.moai/evolution/telemetry/usage-2026-08-30.jsonl` 에 `is_test_fail:true` 로 적혔다. exit-code 분기는 `exit_code==0` 에서 pass 를 반환하므로(`evidence_writer.go:163`) 그 값을 낼 수 없다 → 텍스트 분기(정밀 fail 마커)가 돌았고, 그러려면 stdout 이 payload 에 있어야 한다 |
| (b) `exit_code` 위치 (top-level / nested / 부재) | **미확정** | 탐침이 가른 것은 "셸 rc 0 에 대해 exit-code **pass** 신호가 발화하지 않았다" 뿐이다. 필드 부재와 "있으나 다른 값" 을 이 방법은 구별하지 못한다. 추론으로 메우지 않는다 |
| (c) 감싼 JSON 인가 평문인가 | **미확정** | `decodeToolResponse`(`evidence_writer.go:111`)가 두 모양을 모두 처리하므로, 마커가 매칭됐다는 사실은 어느 쪽이 도착했는지에 대해 아무것도 증명하지 않는다 |

**AC-SEC-000 은 부분 충족이며 그렇게 신고한다.** 조건 (2) 의 세 답 중 (a) 만 관측으로 답했고 (b)(c) 는 미확정이다. 조건 (1) 의 산출물 `.moai/reports/t341/live-payload.json` 은 **존재하지 않는다** — 캡처 지점이 셸 래퍼와 훅 바이너리 둘뿐이고 둘 다 다른 세션과 공유하는 primary 체크아웃에 있어, payload 를 찍도록 고치는 것은 범위 밖이라 하지 않았다. 조건 (3) 의 blocker 분기는 발화하지 않는다 — (a) 의 답이 "예" 이기 때문이다.

M1 로 진행한 근거는 (a) 다: 판정 키(출력 토큰)가 실제로 도착한다는 것이 확인됐고, 그것이 `plan.md` M0 이 M1 진입에 걸어 둔 조건이다. (b)(c) 는 설계를 되돌릴 사유가 아니다 — 거부권이 **두 분기보다 앞**에 서므로 `exit_code` 의 위치·유무와 무관하게 발화한다.

**M0 이 남긴 파생 관측 하나**: 관측된 호출에서 exit-code 분기가 셸 rc 0 에 대해 발화하지 않았다. 가장 경제적인 해석은 살아 있는 Bash payload 가 `exit_code` 를 담지 않는다는 것이고, 그렇다면 텍스트 분기가 live-dominant 경로다. M1 이 각 러너에 대해 실제로 관측한 분기는 아래 표의 마지막 열에 적었다.

### 러너 출력 실측 (M1, 2026-08-30, 이 워크트리)

`plan.md` M1 의 "추정으로 넣지 않는다" 규율에 따라 다섯 러너를 모두 이 머신에서 돌렸다. 판번: go1.25 · pytest **8.4.2** · cargo **1.94.1** · jest **30.4.2** · vitest **3.2.7** · node **v22.14.0**.

| 러너 | 0-실행 실측 출력 | rc | 진짜 pass 실측 출력 | 오늘 pass 를 만드는 경로 |
|---|---|---|---|---|
| go | `ok  \t…/internal/kanban\t0.434s [no tests to run]` | 0 | `ok  \t…/internal/hook\t0.603s` | 정밀 마커 `"ok  \t"` |
| go | `?   \t…/cmd/moai\t[no test files]` | 0 | — | (오늘도 신호 없음) |
| pytest | `collected 0 items` + `no tests ran in 0.01s` | 5 | `3 passed in 0.00s` | 카운트 분기 `" passed"` |
| cargo | `test result: ok. 0 passed; 0 failed; 0 ignored; …` | 0 | `test result: ok. 2 passed; 0 failed; …` | 정밀 마커 `"test result: ok"` |
| jest | `No tests found, exiting with code 0` | 0 | `Tests:       2 passed, 2 total` / `Tests:       10 passed, 10 total` | 카운트 분기 `" passed"` |
| vitest | `No test files found, exiting with code 0` | 0 | `Tests  2 passed (2)` | (0-실행 출력에 마커 없음) |
| npm (래퍼) | `> jest --passWithNoTests` + `No tests found, exiting with code 0` | 0 | — | 아래 러너 출력을 축자로 흘린다 (실측) |
| node 내장 | — | — | `TAP version 13 … ok 1 - a … # pass 1` | **exit-code 경로 단독** — 네 pass 마커 전부 부재를 재확인 |

**`acceptance.md` 가 미측정 상태로 적어 둔 문자열 두 개를 측정이 교정한다.** AC-SEC-004 의 Given 은 jest·vitest 의 0-실행 토큰을 `0 passed` 로 적었으나, **jest 30.4.2 도 vitest 3.2.7 도 0-실행 시 `0 passed` 를 내지 않는다** — 각각 `No tests found` / `No test files found` 다. 기준 본문이 "각 문자열은 M1 이 실제 러너 출력으로 확인해 고정한다" 고 지시하므로 이는 기준 위반이 아니라 그 지시의 이행이다. `0 passed` 형태 자체는 **cargo 에서 실측**돼 corpus 와 토큰 목록에 그대로 남아 있고, 판정은 러너를 가리지 않으므로 장차 `0 passed` 를 내는 jest·vitest 판번도 덮는다.

### 구현 — 무엇을 어디에 넣었는가

| 파일 | 변경 |
|---|---|
| `internal/hook/evidence_writer.go:34-196` | 0-실행 거부권 일체 — 센티널 상수, 러너 토큰 목록(`zeroExecutionLiterals`), 공유 corpus(`zeroExecutionSamples`), `detectZeroExecution` + `isZeroExecutionLine` / `hasExecutionEvidence` / `hasZeroCount`, 자문 조립기 2종. `testCommandSignatures` 바로 옆(`plan.md` M1) |
| `internal/hook/evidence_writer.go:232-237` | `classifyTestCommand` 안, `deriveFromExitCode`·`deriveFromOutputText` **양쪽보다 앞**에 거부권 호출. 판정되면 `(true, false, false)` — 0-실행은 실패가 아니라 신호 없음 |
| `internal/hook/evidence_writer.go:501-508` | `buildBashRecord` 가 0-실행을 `IsZeroExecution=true` 로 **양성 기록**. `Outcome` 은 기존 default 분기대로 `unknown` (≠ `success`) |
| `internal/telemetry/types.go:29-33` | `IsZeroExecution bool` (`omitempty`, 후방 호환). 기존 `IsTestPass`/`IsTestFail` 과 같은 패턴 |
| `internal/hook/post_tool.go:227-235` | Bash 분기에서 `maybeZeroExecutionAdvisory` 로 `HookOutput.SystemMessage` 에 자문 덧붙이기. `Decision` 미설정, 기존 메시지 미삭제 |
| `internal/hook/evidence_writer_zeroexec_test.go` | AC-SEC-001·002·003·004·005·006 테스트 + 실측 표본 상수 |
| `.claude/rules/moai/development/verification-completeness.md` §1.1 | 각주 → `[ZONE:Evolvable] [HARD]` 조항 승격 |
| `internal/template/templates/…/verification-completeness.md` | 미러 (바이트 동일) |

**판정 술어의 모양 — 계획서에 없던 결정이라 명시한다.** `detectZeroExecution` 은 **존재 게이트 + 줄 단위**다: 0-실행 토큰이 실제로 나타나야 하고(부재 기반이 아니다 — `plan.md` §F 가 금지한 형태이자 표본 (f) 가 죽이는 형태), 그러면서 실행 증거를 담은 **다른 줄**이 있으면 거부권이 취소된다. 두 번째 절이 없으면 새 결함이 생긴다: `go test ./...` 는 `[no test files]` 한 줄과 `ok` 수십 줄을 함께 찍으므로, 전체 통과한 스위트 실행이 통째로 "신호 없음" 이 되어 게이트를 **반대 방향으로** 무력화한다. 실패가 같은 출력에 섞인 경우도 같다 — 실패는 실패로 남아야 한다. 이 두 방향을 `TestZeroExecution_MixedOutputKeepsItsSignal` 이 고정한다.

### AC PASS/FAIL 표

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-SEC-000 | **PASS-WITH-DEBT (부분)** | `test -f .moai/reports/t341/live-payload.json` | rc **1** — 파일 없음. 조건 (1) 미충족, 조건 (2) 는 (a) 만 답함, 조건 (3) blocker 분기 미발화. 위 표 참조 |
| AC-SEC-001 | PASS | `go test ./internal/hook/ -run TestZeroExecution_GoZeroMatchIsNotAnObservedPass -count=1` | `ok  github.com/modu-ai/moai-adk/internal/hook 0.624s` |
| AC-SEC-002 | PASS | `go test ./internal/hook/ -run TestZeroExecution_ExitCodeZeroDoesNotCarryItPast -count=1` | 같은 실행에서 PASS |
| AC-SEC-003 | PASS | `go test ./internal/hook/ -run TestZeroExecution_GenuinePassIsUnchanged -count=1 -v` | 11 서브테스트 전부 PASS (`go`/`cargo`/`pytest`/`jest_single_digit`/`jest_double_digit` × 2 shape + `node_builtin/with_exit_zero`) |
| AC-SEC-004 | PASS | `go test ./internal/hook/ -run TestZeroExecution_CorpusSamplesAreVetoed -count=1` | PASS — corpus 10 표본 각각 `detectZeroExecution=true`, `isPass=false`(두 payload 모양 모두) |
| AC-SEC-005 | PASS | `go test ./internal/hook/ -run TestZeroExecution_SurfacesAsPostToolAdvisory -count=1` | PASS — `SystemMessage` 에 `[moai:zero-swept]`, `Decision` 빈 값, 레코드 `IsZeroExecution=true` · `IsTestPass=false` |
| AC-SEC-006 | PASS | `go test ./internal/hook/ -run TestZeroExecution_Corpus -count=1` | PASS — `testCommandSignatures` 9개 signature 전부 표본 보유, 역방향(고아 키) 도 0 |
| AC-SEC-007 | PASS | `diff <local> <mirror>` · `make build` | `diff-exit=0` (출력 0줄) · `make build` rc **0**. 조항이 `[ZONE:Evolvable] [HARD]` 로 승격됐고 각주 문구는 사라졌다 |

### 뮤턴트 탐침 — 일곱 개 전부 죽었다

기준이 공허하지 않다는 것은 **주장이 아니라 실행 결과**다. 각 뮤턴트를 실제로 심고 돌린 뒤 원복했다.

| 뮤턴트 | 죽인 기준 | 관측된 실패 |
|---|---|---|
| A. `hasPrecisePass` 에서 `"ok  \t"` 삭제 | AC-SEC-003 / `go` | `got (isTest,isPass,isFail) = (true,false,false), want (true,true,false)` |
| B. 거부권을 `deriveFromExitCode` **뒤**로 이동 | **AC-SEC-002 단독** (AC-SEC-001 은 통과) | `isPass = true with exit_code 0` |
| C. 카운트 분기 `" passed"` 삭제 | AC-SEC-003 / `pytest`·`jest_single_digit`·`jest_double_digit` | 세 서브테스트 `want (true,true,false)` |
| D. 0-카운트를 부분 문자열 `"0 passed"` 로 구현 | AC-SEC-003 / `jest_double_digit` | `detectZeroExecution = true on a genuine pass output` |
| E. 거부권을 "텍스트에 실행 증거 없음" 으로 구현 | AC-SEC-003 / **`node_builtin` 단독** | `detectZeroExecution = true on the node TAP pass output` |
| F. 자문을 `slog.Warn` 으로만 발신 | AC-SEC-005 | `SystemMessage = ""; want … "[moai:zero-swept]"` |
| G. corpus 를 signature 목록에서 파생시켜 빈 문자열로 채움 | AC-SEC-006 조건 (2) | `detectZeroExecution = false; the sample must fire the veto` |

**탐침 D 가 처음에는 살아남았고, 그것이 표본 하나를 고쳤다.** 최초 `sampleJestPass10` 은 jest 요약 **블록 전체**(`Test Suites: 1 passed…` 줄 포함)였고, 줄 단위 판정에서 이웃 줄의 실행 증거가 거부권을 취소해 버려 뮤턴트가 통과했다. `acceptance.md` 표본 (e) 가 명시한 대로 **요약 줄 하나**로 좁힌 뒤 뮤턴트가 죽었다. 기준이 지목한 충돌은 실재했고, 그것을 드러낸 것은 탐침을 실제로 돌린 일이다.

### 계획서 대비 이탈 3건 (전부 신고)

1. **테스트 파일 위치.** `plan.md` M1 은 fixture 를 `evidence_writer_test.go` 안에 두라고 했으나, 워크트리 격리 가드가 heredoc/compound Bash 호출을 거부해 그 파일에 이어붙일 수 없었다. 같은 패키지의 형제 파일 `internal/hook/evidence_writer_zeroexec_test.go` 에 두었다 — 패키지·헬퍼·관례는 동일하고, 어느 기준도 파일명을 요구하지 않는다.
2. **`internal/telemetry/types.go` 변경.** `plan.md` §B 의 영향 파일 추정(6-8건)에 없던 파일이다. AC-SEC-005 의 "원장에 0-실행으로 **양성 기록**" 을 만족하려면 필드가 필요했고, 기존 `IsTestPass`/`IsTestFail` 과 같은 `omitempty` 추가라 후방 호환이다. §D 제약 어디에도 저촉되지 않는다.
3. **판정 술어에 "실행 증거 있으면 취소" 절 추가.** `plan.md` M1 은 토큰 존재만으로 판정하는 모양을 그렸다. 그대로 구현하면 `go test ./...` 의 정상 통과가 통째로 무효화된다(위 "판정 술어의 모양" 참조). 계획서의 안티패턴(부재 기반 판정)은 그대로 지켰다 — 존재 게이트가 먼저다.

### 결속 파일 무변경 (DoD)

| 검사 | 명령 | 출력 |
|---|---|---|
| 커밋된 편집 | `git diff --name-only origin/develop...HEAD -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` | 0줄 |
| 작업트리·인덱스 | `git status --porcelain -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` | 0줄 |

### 자기 적용 — 이 SPEC 의 검증 실행이 0-실행이 아님

`go test ./internal/hook/ -run TestZeroExecution -count=1 -v` 가 **8개 최상위 테스트 + 11개 서브테스트**를 `--- PASS` 로 열거했다. 실행 수가 0이 아님을 목록으로 확인했고, 출력에 `[no tests to run]` 은 없다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: c6371085c
run_status: complete-with-debt
ac_pass_count: 7
ac_fail_count: 0
ac_partial_count: 1   # AC-SEC-000 — 산출물 미생성, (b)(c) 미확정
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run   # 레인 워크트리, 진입 시점에 develop 흡수 완료
l44_post_push_fetch: not-run    # 이 카드는 push 하지 않는다 (리드가 통합)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass   # go vet ./internal/hook/... ./internal/telemetry/... rc 0
  windows_amd64: pass  # GOOS=windows GOARCH=amd64 go build ./... rc 0
coverage:
  internal_hook: "85.1%"
  internal_telemetry: "82.7%"   # 이 변경은 구조체 필드만 추가해 statement 0건 — 사전 존재 수치
total_run_phase_files: 7
m1_to_mN_commit_strategy: single-commit   # M1~M4 를 한 커밋에 담는다 (한 이음매, 한 기제)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: "c56bba789"   # a commit cannot cite its own hash; backfilled in the immediately following commit
sync_status: completed
changelog_entry: added
changelog_entry_position: "CHANGELOG.md [Unreleased] -> ### Added, first entry (line 12), inserted above SPEC-BINLAG-INVOCATION-001"
b12_self_test_a: "grep -c 'SELECTOR-CENSUS' CHANGELOG.md -> 0, exit 1 (pre-emission; no duplicate entry from a parallel session). Post-emission the same command returns 1"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 8 (AC-SEC-000..007). Non-zero and plausible; matches the 8-row PASS/FAIL matrix in §E.2 and the count the CHANGELOG entry states (7 PASS + 1 PASS-WITH-DEBT)"
b12_self_test_c: "ls on every path claimed in the entry -> all present: .moai/specs/SPEC-SELECTOR-CENSUS-001/spec.md, internal/hook/evidence_writer.go, internal/hook/evidence_writer_zeroexec_test.go, internal/hook/post_tool.go, internal/telemetry/types.go, .claude/rules/moai/development/verification-completeness.md, internal/template/templates/.claude/rules/moai/development/verification-completeness.md, .moai/reports/t341/{discovery,m0-live-payload-observation,plan-audit-iter1,plan-audit-iter2}.md"
frontmatter_status_transitions:
  spec_md: "draft -> completed (merged 3-phase close on this sync commit); updated: 2026-08-29 -> 2026-08-31. NOTE: the source state was `draft`, not `in-progress` — the run phase never performed the `draft -> in-progress` step. Recorded as it happened rather than as the canonical path"
  progress_md: "in-progress -> completed; updated: already 2026-08-31, re-affirmed"
  plan_md: "NOT transitioned — carries no `status:` field (grep '^status:' -> no match). Removed at 97dc597bf under ArtifactStatusFieldForbidden (SPEC-ARTIFACT-STATELESS-001, card t357). Re-adding one to satisfy an 'all four artifacts' reading would create the exact drift that rule forbids"
  acceptance_md: "NOT transitioned — same as plan_md; no `status:` field, none added"
docs_surfaces_touched:
  changelog: "CHANGELOG.md — one entry added under [Unreleased] -> Added"
  readme: "none. grep -rlniE 'zero-swept|no tests to run|evidence_writer' README{,.ko,.ja,.zh}.md -> 0 files"
  docs_site: "none. Same grep across docs-site/content/** and .moai/docs/** -> 0 files. Control grep for 'moai' across README.md + docs-site/content matched 637 files, so the scanned surface is live and the 0 is a measured zero rather than a mis-scoped scan"
  rationale: "the deliverable changes internal hook classification behaviour and adds no operator-facing CLI verb, settings key, hook event, or wrapper script, so no user-facing surface is owed"
mx_tag_validation: "not performed in this sync commit. The sync commit touches only CHANGELOG.md, this file's frontmatter + §E.4, and spec.md frontmatter — no .go file is modified here, and the run-phase Go surface (evidence_writer.go, post_tool.go, types.go, evidence_writer_zeroexec_test.go) was not re-scanned for @MX annotations by this session. Reported as an unperformed step, not as a clean result"
canary_compliance_check:
  applicable: true
  reason: "this SPEC defines a forward-looking rule (a zero-execution invocation is not an observed pass) and the run phase applied it to its own verification run"
  result: "`go test ./internal/hook/ -run TestZeroExecution -count=1 -v` enumerated 8 top-level tests and 11 subtests as `--- PASS`, with no `[no tests to run]` line — this SPEC's own verification is not a zero-execution run under its own predicate. Consumed from §E.2; NOT re-executed by this sync session"
verification_re_execution: "none. This sync session ran no test, no lint, and no build. Every numeric claim in the CHANGELOG entry (85.1% / 82.7% coverage, the seven mutant verdicts, the five-runner token measurements, the 8+11 self-application count, `make build` rc 0, the darwin/windows exits) is CONSUMED from §E.2/§E.3 as run-phase evidence at commit c6371085c, not re-measured here. `git merge-base --is-ancestor c6371085c HEAD` -> exit 0 confirms that evidence's commit is an ancestor of this tree"
not_observed:
  ci: "no CI run on this branch has been read by this session. WT-selector-census is unpushed; whether origin/develop's own CI is green at b9149857c is the lead's read, not this session's"
  full_suite: "`go test ./...` NOT run — prohibited locally in this repository. The full-suite verdict is CI's"
  audit: "no sync-auditor pass has been run against this sync commit"
  spec_audit: "mcp__moai__spec_audit / `moai spec lint` NOT run by this session against this worktree"
  ac_sec_000_debt: "carried forward unchanged and unrepaired — .moai/reports/t341/live-payload.json still does not exist, and questions (b) exit_code position/presence and (c) wrapped-JSON vs plain payload shape remain undetermined. This sync phase did not attempt to close them"
push_state: "not pushed, not merged, no PR opened by this session. The lead owns the integration window"
scope_discipline: "this sync commit modifies ONLY CHANGELOG.md, this file (frontmatter + §E.4), and spec.md frontmatter (`status:` + `updated:`). No body content of spec.md / plan.md / acceptance.md was touched, and no `status:` field was added to plan.md or acceptance.md"
```
