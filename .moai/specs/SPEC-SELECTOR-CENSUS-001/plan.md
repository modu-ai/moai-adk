---
id: SPEC-SELECTOR-CENSUS-001
title: "0-실행 테스트 판정 — 구현 계획"
version: "0.1.0"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "t341, evidence-gate, zero-execution, template-first"
tier: M
---

# 구현 계획 — SPEC-SELECTOR-CENSUS-001

카드 **t341** · 워크트리 `.claude/worktrees/t341` · 브랜치 `WT-selector-census` · 기준 트리 `a6bbbf82b`

## A. 맥락

`spec.md` §1-§2 가 실측한 상태 위에서 움직인다. 요약: 0개를 실행한 `go test` 가 증거 원장에 `IsTestPass=true` / `Outcome=success` 로 적히고, Stop 증거 게이트가 그것을 관측된 pass 로 읽는다.

## B. Tier

**Tier M.** 영향 파일 추정 6-8건 — `internal/hook/evidence_writer.go`, `internal/hook/evidence_writer_test.go`, `internal/hook/post_tool.go`(+ 그 테스트), `.claude/rules/moai/development/verification-completeness.md`, 그 템플릿 미러, `catalog.yaml`(미러 편집 시 해시 동반 변경). LOC 추정 300 미만~400. Tier M 대역(5-15 파일, 300-1000 LOC) 안이며 헌법급 변경이 아니다. 산출물 3종(spec/plan/acceptance), PASS 임계 0.80.

Tier 를 S 로 내리지 않는 이유: 규칙층 + 템플릿 미러가 붙어 파일 수가 5를 넘고, `make build` 재생성이 게이트로 끼어든다.

## C. 마일스톤 — 되돌리기 어려운 결정을 앞에

### M0 — 살아 있는 payload 관측 (전제 검증; 이 계획의 나머지가 여기에 걸려 있다)

`spec.md` §5 의 첫 전제를 관측으로 바꾼다. 실제 Claude Code PostToolUse payload 를 하나 잡아, `go test` 호출에 대해 (a) Bash stdout 이 실려 오는가, (b) `exit_code` 가 어느 위치에 실리는가(top-level / nested), (c) 텍스트가 감싼 JSON 인지 평문인지를 읽는다.

- 결과가 "stdout 이 실려 온다 + exit_code 도 실려 온다" 이면 계획 그대로 M1 로 간다.
- **stdout 이 실려 오지 않으면 §3.3 의 채널 선택과 §2.2 의 순서 제약이 무너진다** — 출력 토큰이 도착하지 않으므로 판정 키 자체를 다시 골라야 한다. 그 경우 M1 에 들어가지 않고 blocker 보고로 리드에 돌린다.

산출물: 캡처한 payload 를 `.moai/reports/t341/live-payload.json` 에 남기고, 판독을 진행 기록 §E.2 에 적는다.

**기계적 손잡이 — 이 마일스톤은 산문 게이트가 아니다.** M0 은 **AC-SEC-000** 으로 승격돼 있고 DoD 와 §E 의 E0 행이 그 산출물을 요구한다. 그래서 M0 을 건너뛰면 완료 판정이 붉어진다 — 종전에는 AC 7건이 전부 합성 fixture 위에서 돌아 건너뛰어도 전부 초록이었다. blocker 분기도 기준 안에서 판정된다: (a) 의 답이 "아니오" 이면 AC-SEC-000 (3) 이 `internal/hook/evidence_writer.go` 에 이 카드의 커밋이 **없을 것**을 요구하므로, "stdout 이 안 실리는데 M1 을 그냥 진행" 은 어느 분기로도 통과하지 못한다.

### M1 — 0-실행 거부권 (자료형·호출 계약이 바뀌는 지점)

`internal/hook/evidence_writer.go`:

- 러너별 0-실행 토큰 taxonomy 를 `testCommandSignatures` 옆에 고정 목록으로 둔다(같은 파일, 같은 모양). 각 문자열은 M1 이 실제 러너 출력으로 확인한 뒤 넣는다 — 추정으로 넣지 않는다.
- 순수 함수 `detectZeroExecution(text) bool` 를 추가하고, `classifyTestCommand` 에서 **`deriveFromExitCode` 보다 앞에** 호출한다(`spec.md` §2.2). 판정되면 `isTest=true, isPass=false, isFail=false` 로 반환한다 — 0-실행은 실패가 아니라 **신호 없음**이다(부재 ≠ 실패, 기존 §R1 계약과 같은 방향).
- **진입 조건**: AC-SEC-000 이 초록(또는 blocker 분기로 판정)이기 전에는 M1 에 들어가지 않는다.
- 0-실행 표본 corpus 를 **하나의 변수**로 두고 AC-SEC-004 와 AC-SEC-006 이 그 같은 변수를 읽게 한다(사본 금지). 각 항목은 `detectZeroExecution` 을 실제로 발화시켜야 한다.
- 러너별 토큰은 **경계 있는 형태**로 고정한다 — 부분 문자열 `"0 passed"` 는 진짜 pass `10 passed` 를 삼킨다(AC-SEC-003 뮤턴트 탐침 2, 실측).
- 테스트: AC-SEC-001 · 002 · 003 · 004. 표 기반 서브테스트, `Test<Subject>_<Scenario>` 명명, fixture 는 `evidence_writer_test.go` 안에 둔다(기존 관례).

되돌리기 어려운 이유: 반환 계약과 호출 순서가 바뀌므로 뒤에 붙는 모든 것이 이 모양 위에 선다.

### M2 — 드러내기 (사용자가 보는 표면)

`internal/hook/post_tool.go` 의 Bash 분기에서, 0-실행이 판정된 경우 `HookOutput.SystemMessage` 에 안정된 센티널 토큰(예: `[moai:zero-swept]`)을 담은 한 줄 자문을 덧붙인다. navigator detect 자문(`post_tool.go:239`)과 같은 덧붙이기 방식을 그대로 쓴다 — 기존 `systemMessage` 문자열을 지우지 않는다.

원장 쪽은 0-실행을 **양성으로** 기록한다(§3.4). 차단하지 않고, 오류는 삼킨다.

테스트: AC-SEC-005.

### M3 — 지속 발화 장치

`testCommandSignatures` 를 순회하며 각 러너 축이 0-실행 표본을 갖는지 단언하는 테스트를 넣는다. 빠진 러너 이름을 실패 메시지에 출력한다. **세는 것으로 끝내지 않는다** — 같은 순회 안에서 각 표본을 `classifyTestCommand` 와 `detectZeroExecution` 에 실제로 먹이고 `isPass=false` · 거부권 발화를 단언한다(AC-SEC-006 조건 2). corpus 는 M1 이 만든 그 변수를 그대로 읽는다. 테스트: AC-SEC-006.

M2 뒤에 두는 이유: 표본 corpus 의 최종 모양이 M1·M2 에서 굳은 뒤라야 망라 단언이 무엇을 망라하는지가 고정된다.

### M4 — 규칙 조항 승격 + 템플릿 미러 (기계적)

- `.claude/rules/moai/development/verification-completeness.md` §1.1 의 *"Evidence footnote (not a rule)"* 문단을 `[ZONE:Evolvable] [HARD]` 조항 **한 개**로 승격한다. 그 파일의 다른 절을 다시 쓰지 않는다.
- `internal/template/templates/.claude/rules/moai/development/verification-completeness.md` 미러를 같은 변경에서 갱신한다(CLAUDE.local.md §2 Template-First). 미러는 **축자 복사가 아니다** — 템플릿 중립성(§25)에 따라 SPEC-ID·카드 id·내부 날짜를 담지 않는다.
- `make build` 실행. `catalog.yaml` 이 함께 바뀌면 같은 커밋에 담는다.
- **완료 조건은 `diff` rc 0** — 두 파일이 바이트 동일해야 한다(AC-SEC-007 (2)). 오늘 이미 rc 0 이며(실측), 승격 조항은 SPEC-ID·카드 id·날짜를 담지 않으므로 중립화 차이가 필요 없다. 중립화 차이가 실제로 필요해지면 진행하지 말고 blocker 로 보고한다.
- 테스트/검증: AC-SEC-007.

## D. 제약

- **셸 훅 쌍 규율**: 이 계획은 `.claude/hooks/moai/*.sh` 를 건드리지 않는다. 만약 M2 가 셸 경로로 선회하면 `.sh` 와 `.sh.tmpl` 을 **같은 변경에서** 함께 고쳐야 한다(CLAUDE.local.md §2.3) — 선회 자체가 blocker 보고 사유다.
- **`SPEC-TODO-SQLITE-001/acceptance.md` 는 건드리지 않는다**(t343 결속, `spec.md` §2.4). 편집이 필요하다고 판단되면 멈추고 리드에 보고한다.
- **검증 범위**: `go test ./internal/hook/... -count=1` + 건드린 패키지. 로컬 전체 스위트 금지(CLAUDE.local.md §4). 전 패키지 판정은 CI.
- **자기 적용**: 이 SPEC 의 검증 실행 자체가 0-실행이면 안 된다 — 테스트 실행 시 실행 수를 함께 읽는다.
- 커밋 메시지마다 카드 id `t341`.

## E. 자기 검증 (실행 단계가 채운다)

| # | 항목 | 명령 |
|---|---|---|
| E0 | M0 산출물 존재 + 판독 | `test -f .moai/reports/t341/live-payload.json && jq -e . .moai/reports/t341/live-payload.json >/dev/null` → rc 0, 그리고 §E.2 가 (a)(b)(c) 에 답함 (AC-SEC-000) |
| E1 | AC 7건 PASS/FAIL 표 | `go test ./internal/hook/... -run 'ZeroExec\|Classify' -count=1 -v` |
| E2 | 패키지 테스트 | `go test ./internal/hook/... -count=1` |
| E3 | 정적 검사 | `go vet ./internal/hook/...` |
| E4 | 템플릿 미러 동기 | `diff` 로컬 ↔ 미러 해당 절 + `make build` rc |
| E5a | 결속 파일 무변경 — 커밋된 편집 | `git diff --name-only origin/develop...HEAD -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` → 0줄 |
| E5b | 결속 파일 무변경 — 작업트리·인덱스 | `git status --porcelain -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` → 0줄 |

## F. 안티패턴 (하지 않을 것)

- 마커 목록에서 `"ok  \t"` 를 지워 문제를 "해결"하기 — 진짜 pass 신호가 함께 죽는다(AC-SEC-003 이 막는다).
- 0-실행을 **실패**로 기록하기 — 부재 ≠ 실패이며, 게이트를 붉게 만들어 다른 방향의 거짓 신호를 만든다.
- 거부권을 `deriveFromOutputText` 안에만 넣기 — 살아 있는 훅 경로에서 아무 일도 일어나지 않는다(AC-SEC-002 가 막는다).
- 어휘·시제 판별자로 선회하기(`spec.md` §3.2).
- 범위 밖 세 항목(grep 0-히트 기준, `t.Skip`, `sg test`)을 "겸사겸사" 넣기.
- 망라 테스트를 **세는 것으로만** 만들기 — 표본을 돌리지 않으면 `map[sig]""` 로도 초록이다(AC-SEC-006 뮤턴트 탐침).
- 결속 파일 무변경을 인수 없는 `git diff` 로 확인하기 — 스테이징·커밋된 편집을 놓친다(DoD 실측 표).
- M0 관측 없이 M1 로 진행하기 — AC-SEC-000 과 E0 가 이를 붉게 만든다.
- 0-실행 거부권을 "텍스트에 실행 수 근거가 없음" 으로 구현하기 — `deriveFromExitCode` 의 pass 경로를 함께 좁혀, 마커를 하나도 내지 않는 러너(`npm test`→node 내장)의 진짜 pass 를 죽인다(AC-SEC-003 표본 (f) 가 막는다).

## G. 교차 참조

- `.moai/reports/t341/discovery.md` — 실측 원본
- `.claude/rules/moai/development/verification-completeness.md` §1.1 / §1.3 / §2 / §4
- `CLAUDE.local.md` §2 · §2.3 · §4
