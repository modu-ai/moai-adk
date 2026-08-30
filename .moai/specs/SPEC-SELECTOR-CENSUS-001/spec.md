---
id: SPEC-SELECTOR-CENSUS-001
title: "0-실행 테스트 판정 — 아무것도 쓸어담지 못한 실행을 관측된 pass 로 적지 않는다"
version: "0.1.0"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "t341, evidence-gate, test-selector, zero-execution, verification-completeness, template-first"
tier: M
---

# SPEC: 0-실행 테스트 판정

카드: **t341** · 워크트리 `.claude/worktrees/t341` · 브랜치 `WT-selector-census`
실측 트리: **`a6bbbf82b`** (fetch 시점 `origin/develop`, 2026-08-29)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-29 | manager-spec | 최초 작성. `.moai/reports/t341/discovery.md` 의 실측 위에 세웠고, 그 보고가 짚지 않은 이음매 하나를 이 판에서 추가로 읽어 담았다 — 0-실행 판정은 `deriveFromOutputText` 안이 아니라 **exit-code 분기보다 앞**에 서야 한다(§2.2). 요구 7 / 수락 7, Tier M |

---

## 1. 배경 — 침묵이 아니라 거짓 기록이다

카드가 세운 전제는 "선택자가 실제로 무엇을 골랐는지 세지 않는 검증에는 기계층 경고가 없다"였다. 이 트리에서 실측한 상태는 그보다 나쁘다. 기계층은 침묵하지 않는다 — **0개를 실행한 `go test` 를 관측된 테스트 pass 로 증거 원장에 적는다**. Stop 증거 게이트가 읽는 신호가 바로 그 원장이다.

측정 (이 워크트리, `a6bbbf82b`, 2026-08-29):

```
$ go test ./internal/kanban -run TestT341NoSuchTestXYZ -count=1 ; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.270s [no tests to run]
exit=0
```

`rc 0` 이 뜻하는 것은 "골라야 했던 것이 다 통과했다"가 아니라 "골라진 것이 다 통과했다"이며, 골라진 것이 공집합일 때도 같은 줄이 나온다.

정책층은 이 형태를 이미 알고 있다. `.claude/rules/moai/development/verification-completeness.md` §1.1 은 이를 *"Evidence footnote (not a rule)"* 로, §1.3 은 Selection 사례로 담는다. 그런데 조항이 아니라 각주이므로 기계층이 인용할 규칙이 없고, 저작 시점에 같은 형태가 계속 재발한다 — 이번 배치에서만 독립 실측 5건, 그리고 이 카드의 탐사 자체가 여섯 번째를 만들어 냈다(존재한 적 없는 파일에 대해 `PASS` 가 찍혔다, `discovery.md` §E3).

## 2. 실측된 기계층 상태

### 2.1 분류기가 0-실행을 정밀 pass 마커로 읽는다

`internal/hook/evidence_writer.go` `deriveFromOutputText` 는 `"ok  \t"` 를 **정밀 pass 마커**로 보고(`evidence_writer.go:217`), 어떤 실행 수도 보기 전에 `return true, false, true` 한다(분기 `:223`, 반환 `:224`). go 의 0-매치 줄은

```
ok  \tgithub.com/modu-ai/moai-adk/internal/config\t0.424s [no tests to run]
```

이므로 그 마커를 담는다. `buildBashRecord` 는 그 결과를 `IsTestPass=true`, `Outcome=success` 로 적는다(`:309-330`, 대입은 `:328`). 게이트는 관측된 pass 를 보고 조용해진다.

`[no tests to run]` / `[no test files]` / `no tests ran` 은 `internal/hook/` 어디에도 없다:

```
$ grep -rnc 'no tests to run\|no test files\|no tests ran' internal/hook/ | grep -v ':0'
$ echo "grep-exit=$?"
grep-exit=1
```

(출력 0줄 = 세 토큰 중 어느 것도 담은 파일이 없음.)

### 2.2 이 판이 추가로 읽은 이음매 — 판정 위치는 텍스트 휴리스틱 **앞**이다

`discovery.md` 는 이음매를 `deriveFromOutputText` 로 지목한다. 그 안에만 고치면 **살아 있는 훅 경로에서는 아무 일도 일어나지 않는다**. `classifyTestCommand` 는 구조화된 exit-code 신호를 **먼저** 보고 거기서 반환하기 때문이다:

```
evidence_writer.go:69   if pass, fail, ok := deriveFromExitCode(result); ok {   // ← 여기서 return
evidence_writer.go:74   if pass, fail, ok := deriveFromOutputText(result); ok { // ← 텍스트는 그 다음
```

훅 payload 가 `exit_code: 0` 을 담으면 — 0-매치 실행의 종료코드가 정확히 0 이다 — 텍스트는 읽히지도 않는다. 따라서 0-실행 판정은 두 분기보다 **앞**에 서는 거부권이어야 한다. (이 절은 `a6bbbf82b` 의 파일을 직접 읽어 세웠다.)

### 2.3 우연히 맞고 있는 인접 형태

`[no test files]` 와 pytest `no tests ran` 은 오늘 신호를 내지 않는다 — 다만 `ok  \t` 토큰을 담지 않아서일 뿐이고, 그 동작을 못 박는 테스트가 없다. 마커 목록 편집 한 번이면 회귀한다. 이 SPEC 은 그 우연을 **단언으로 바꾼다**.

### 2.4 착지한 산출물에 규범으로 박힌 같은 거짓 전제

`origin/develop:.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md:13`, AC-TOSQ-001 의 RED-now 칸은 이렇게 적혀 있다.

> Test name does not exist → suite failure ("no tests to run" surfaces red).

§1 의 측정이 이를 반증한다 — 붉지 않다. 같은 거짓 전제 위에 선 기준은 그 표에서 모두 **9건**이다: AC-TOSQ-001, 002, 003, 004, 005, 007, 008, 017, 018.

**이 숫자의 출처는 두 갈래이며, 갈래를 밝히지 않은 수치는 이 SPEC 이 겨누는 바로 그 결함이다**(`verification-completeness.md` §4, baseline 귀속 조항). 001 과 형제 6건(002-005, 007, 008)은 **이 레인이 이 워크트리에서 해당 셀을 직접 읽어** 확인했다. 017·018 두 건은 **카드 t343(lane-7) 레인의 재측정에서 온 것**이며, 이 SPEC 은 그 두 건을 스스로 재지 않았다. 017·018 을 이 레인의 관측으로 읽어서는 안 된다. AC-TOSQ-011 은 t343 이 **의도적으로 제외**했다 — 기제가 다르고(존재하지 않는 verb 의 비-0 종료), 재측정 대상도 아니었으므로 이 9건에 들어가지 않는다.

**[HARD] 교차 카드 결속.** t343(lane-7)이 **반대 축**에서 같은 셀을 인용한다("사실과 어긋난 RED 칸이 감사를 통과했다"). 그 셀을 먼저 편집하는 쪽이 다른 카드의 증거를 조용히 없앤다. `SPEC-TODO-SQLITE-001/acceptance.md:13` 에 손대는 변경은 **같은 변경 안에서 두 SPEC 을 함께 갱신**해야 한다. 이 SPEC 의 범위 안에서 그 셀은 **읽기 전용 증거**이며, 이 SPEC 은 그것을 고치지 않는다.

## 3. 범위

### 3.1 이음매 하나, 기제 하나

`internal/hook/evidence_writer.go` 의 기존 순수 함수 `classifyTestCommand` / `deriveFromOutputText` 안에서:

1. **0-실행 거부권.** 러너가 스스로 낸 0-실행 토큰이 출력에 있으면, 그 실행은 관측된 pass 로 기록되지 않는다. `testCommandSignatures` 에 이미 있는 러너 전부가 대상이다 — go(`[no tests to run]`, `[no test files]`), pytest(`no tests ran`, `collected 0 items`), jest·vitest(`0 passed`), cargo(`0 passed`).
2. **드러내기.** 0-실행은 **기록되지 않는 데서 그치지 않고 표면에 뜬다** — 침묵이 이 카드가 겨눈 병이다. 채널은 이미 있는 것 중 가장 싼 것을 고른다(§3.3).
3. **규칙 승격.** `verification-completeness.md` §1.1 의 각주를 조항 한 개로 승격한다. 기제가 인용할 규칙이 생기는 것이 전부이며, 그 파일을 다시 쓰지 않는다.

### 3.2 판정 키는 구조이지 어휘가 아니다

키는 **러너가 스스로 내는 출력 토큰 + 실행 수의 부재**다. 주장 문면의 시제·어법을 보는 판별자는 범위 밖이다 — lane-8 이 t342 에서 그 형태가 양방향으로 불건전함을 실측했고(오탐 7건), §2.4 의 셀 자체가 평범한 평서형 현재라 그런 판별자에는 보이지 않는다.

### 3.3 드러내기 채널 — `HookOutput.SystemMessage`

고른 것: **`internal/hook/post_tool.go` 의 PostToolUse 경로에서 `HookOutput.SystemMessage`**.

근거는 값이 아니라 이미 있음이다. (a) 증거 기록이 일어나는 바로 그 핸들러가 이미 `SystemMessage` 를 담아 반환한다(`post_tool.go:280`). (b) 같은 함수 안에 자문 전용 선례가 있다 — navigator detect 가 `emitNavigatorDetectAdvisory` 로 같은 필드에 덧붙인다(`post_tool.go:239`). (c) 필드 자체의 계약이 "Warning message shown to user" 다(`types.go:366`). 밀려난 대안: `pre_tool.go:643-653` 의 자문은 **실행 전** 표면이라 출력이 아직 없고, `.claude/hooks/moai/handle-pre-tool.sh:24-53` 의 셸 WARN 은 stderr 로만 가며 `.sh`/`.sh.tmpl` 쌍 유지 부담을 새로 만든다.

**자문 전용이다.** deny 가 아니고, 게이트가 아니며, fail-open 이다(`logEvidence` 의 기존 계약과 같다).

### 3.4 지속 발화 (verification-completeness §1.3)

이 판정이 **멈춘 것을 독자가 어떻게 아는가**에 이 SPEC 이 내는 답: **표본 corpus 가 `testCommandSignatures` 를 망라함을 테스트가 단언한다.** 러너가 목록에 추가됐는데 0-실행 표본이 없으면 그 테스트가 붉어진다. 표본이 조용히 좁아지는 길이 막힌다.

부수로, 0-실행은 원장에 **양성으로** 기록된다(§4 REQ-SEC-004) — 기록이 사라지는 것과 판정이 멈춘 것이 구별되지 않는 상태를 만들지 않기 위해서다.

### 3.5 자기 적용 (t241)

이 가드는 자기가 강제하는 규칙의 적용 대상이다. 수락 기준은 **양방향을 모두 관측**한다 — 심어 둔 표본이 발화하고, 진짜 pass 는 발화하지 **않는다**. 한 방향만 확인된 가드는 꺼진 가드와 구별되지 않는다. §2 의 뮤턴트 탐침을 각 기준에 붙인다.

**그 대칭은 계측기 단위가 아니라 러너 계열 단위로 성립해야 한다.** 비발화 방향을 go 의 정밀 마커 하나로만 고정하면, 0-실행 거부권을 옳게 넣으면서 `deriveFromOutputText` 의 카운트 분기(`evidence_writer.go:232`, `" passed"`)를 함께 지우는 구현이 모든 기준을 만족한다 — 그리고 pytest 의 진짜 pass `3 passed in 0.00s`(이 트리에서 실측, `pytest -q` 2026-08-29)가 조용히 `isPass=false` 가 된다. 그래서 AC-SEC-003 은 발화 방향과 **같은 러너 축 전부**(go / cargo / pytest / jest·vitest)에 대해 진짜 pass 를 고정한다.

### 3.6 제외

이 SPEC 이 **짓지 않는 것**. 셋 다 같은 추론 오류를 공유하지만 실행 이음매를 공유하지 않는다 — 앞의 둘은 저작층이고, 셋째는 명령 signature 가 추가되면 같은 계수 taxonomy 로 공짜로 딸려 온다. 기계적으로 덮으려면 두 번째, 다른 계측기가 필요하다. 별도 카드 소관이다.

### Out of Scope — 0-히트 통과 조건을 가진 grep 기반 수락 기준

- grep 이 0건을 맞히면 통과하는 수락 기준의 판별·차단.
- 그런 기준을 SPEC 저작 시점에 막는 lint 규칙.

### Out of Scope — 단언 앞의 `t.Skip`

- 단언에 닿기 전에 `t.Skip` 으로 빠져나가는 테스트의 판별.
- skip 사유의 분류나 skip 예산.

### Out of Scope — 빈 룰셋에 대한 `sg test`

- `sg test` 의 `0 passed` 판정. `testCommandSignatures` 에 `sg test` 를 추가하는 것 자체도 이 SPEC 밖이다.

### Out of Scope — 착지한 SPEC 문면 수리

- `SPEC-TODO-SQLITE-001/acceptance.md:13` 및 §2.4 가 세는 나머지 8건(합 9건)의 문면 수리. §2.4 의 결속 때문에 t343 과 같은 변경에서만 다뤄야 하며, 이 SPEC 에서 그 셀은 읽기 전용 증거다.

### Out of Scope — 어휘·시제 판별자

- 주장 문면의 어법으로 결함을 가르는 판별자 일체(§3.2).

## 4. 요구 (GEARS)

**REQ-SEC-001 (Ubiquitous).** 증거 분류기는 실행된 테스트가 0건인 테스트 실행을 관측된 pass 로 기록하지 **않는다**. 그런 실행의 원장 결과는 `success` 가 아니다.

**REQ-SEC-002 (When, 순서 제약).** 출력이 0-실행 토큰을 담을 때, 분류기는 구조화된 exit-code 신호와 출력 텍스트 휴리스틱 **어느 쪽보다도 먼저** 그 사실을 판정한다. (§2.2 — exit-code 분기가 앞서면 텍스트는 읽히지 않는다.)

**REQ-SEC-003 (Ubiquitous, 판정 키).** 판정 키는 러너가 스스로 낸 출력 토큰과 실행 수의 부재이며, 주장·기준·커밋 문면의 어휘를 입력으로 삼지 않는다.

**REQ-SEC-004 (When, 드러내기).** 0-실행이 판정되면, PostToolUse 훅은 안정된 센티널 토큰을 담은 자문을 `HookOutput.SystemMessage` 로 내고, 원장에는 그 실행을 0-실행으로 **양성 기록**한다. 자문은 차단하지 않으며 실패 시 조용히 통과한다(fail-open).

**REQ-SEC-005 (Ubiquitous, 반대 방향).** 실행 수가 1 이상인 진짜 pass 는 종전과 똑같이 `IsTestPass=true` / `Outcome=success` 로 기록된다. 이 SPEC 은 살아 있는 pass 신호를 좁히지 않는다.

**REQ-SEC-006 (Where, 지속 발화).** `testCommandSignatures` 에 러너 signature 가 있는 경우, 표본 corpus 는 그 러너의 0-실행 표본을 담는다. 담지 않으면 테스트가 실패한다.

**REQ-SEC-007 (Ubiquitous, 규칙층).** `verification-completeness.md` §1.1 의 0-스윕 각주는 조항으로 승격되고, `internal/template/templates/.claude/rules/moai/development/` 의 미러가 같은 변경에서 갱신되며 `make build` 가 실행된다.

## 5. 전제 (검증되지 않음 — 실행 단계 M0 의 입력)

- **살아 있는 payload 가 Bash stdout 을 담는가는 관측되지 않았다.** `evidence_writer.go` 는 감싼 `tool_response` 객체를 전제로 쓰였고 그 테스트는 합성 fixture 만 쓴다. 실제 Claude Code PostToolUse payload 를 잡아 읽은 적이 없다. **M0 이 그것을 관측하기 전까지 §3.3 의 채널 선택과 §2.2 의 순서 제약은 가설이다.** 관측 결과가 다르면 M1 이전에 설계를 되돌린다. 이 전제는 **AC-SEC-000 으로 승격돼 완료 판정에 걸린다** — 산문 게이트가 아니라 산출물(`.moai/reports/t341/live-payload.json`)과 §E.2 판독을 요구하는 기준이며, blocker 분기도 그 기준 안에서 판정된다.
- jest / vitest / cargo 의 0-실행 출력 토큰은 이 트리에서 실행되지 않았다(go 와 pytest 형태만). M1 이 각 러너의 실제 출력 문자열을 확인한 뒤 표본을 고정한다.
- 같은 "missing test → red" 전제를 담은 다른 착지 SPEC 이 몇 건인지 세지 않았다. `SPEC-TODO-SQLITE-001` 하나만 읽었다.

## 6. 교차 참조

- `.claude/rules/moai/development/verification-completeness.md` §1.1 / §1.3 / §2 — 승격 대상 각주, Selection 사례, 두 칸 규율
- `.moai/reports/t341/discovery.md` — 이 SPEC 의 실측 원본
- `SPEC-TODO-SQLITE-001/acceptance.md:13` — §2.4 의 반증 대상 (읽기 전용, t343 과 결속)
- `CLAUDE.local.md` §2 / §2.3 — Template-First 및 `.sh`/`.sh.tmpl` 쌍 규율
