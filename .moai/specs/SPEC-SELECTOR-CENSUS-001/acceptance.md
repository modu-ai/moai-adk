---
id: SPEC-SELECTOR-CENSUS-001
title: "0-실행 테스트 판정 — 수락 기준"
version: "0.1.0"
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "t341, evidence-gate, zero-execution, two-cell, mutation-testing"
tier: M
---

# 수락 기준 — SPEC-SELECTOR-CENSUS-001

수락 기준 총 **8건**(AC-SEC-000 … AC-SEC-007). 폐기 기준 없음.

모든 RED-now 칸은 트리 **`a6bbbf82b`** 에 못 박혀 있다(`verification-completeness.md` §4 — 움직이는 ref 에 대고 재지 않는다). 각 칸은 명령·관측 출력·**왜 붉은지**를 함께 적는다.

**"red via missing test" 라는 표현은 이 문서에서 RED 사유로 쓰지 않는다** — 그 전제를 반증하는 것이 이 SPEC 의 존재 이유다(`spec.md` §2.4).

---

### AC-SEC-000 — M0 의 살아 있는 payload 관측이 산출물로 남는다 (REQ-SEC-002, REQ-SEC-004)

이 기준의 존재 이유: **AC-SEC-001 … AC-SEC-007 은 전부 합성 fixture 위에서 돈다.** M0 을 통째로 건너뛰어도 나머지 일곱은 초록이 된다 — `plan.md` M0 이 산문으로 세운 게이트에 기계적 손잡이를 붙이는 것이 이 기준이다.

**Given** 실제 Claude Code PostToolUse 훅이 `go test` Bash 호출에 대해 만든 payload 하나

**When** M0 이 그 payload 를 캡처한다

**Then** 다음 셋이 모두 성립한다.

1. `.moai/reports/t341/live-payload.json` 이 **존재하고** 유효 JSON 이다 — `test -f .moai/reports/t341/live-payload.json && jq -e . .moai/reports/t341/live-payload.json >/dev/null` 이 rc 0.
2. 진행 기록 §E.2 가 세 질문에 각각 답한다 — (a) Bash stdout 이 payload 에 실려 오는가 (예/아니오), (b) `exit_code` 가 실린 위치(top-level / nested / 부재), (c) 텍스트가 감싼 JSON 인가 평문인가. 세 답 각각이 그 payload 의 **어느 키**에서 읽혔는지를 함께 적는다.
3. **(a) 의 답이 "아니오" 이면 이 기준은 blocker 분기로 판정된다** — `plan.md` M0 의 blocker 사유가 진행 기록에 적히고, `internal/hook/evidence_writer.go` 에 이 카드의 변경이 **없다**(`git diff --name-only origin/develop...HEAD -- internal/hook/evidence_writer.go` 가 0줄). 즉 stdout 이 실려 오지 않는 세계에서도 이 기준은 이진 판정되며, "M1 을 그냥 진행" 은 어느 분기로도 통과하지 못한다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | 산출물이 없다. 붉은 **사유**: `ls .moai/reports/t341/live-payload.json` → `ls: .moai/reports/t341/live-payload.json: No such file or directory`, `ls-exit=1` (이 워크트리, 2026-08-29). 그 디렉터리에는 `discovery.md` 와 `plan-audit-iter1.md` 둘뿐이며 살아 있는 payload 는 이 카드에서 아직 한 번도 관측된 적이 없다 |
| **초록 경로** | M0 이 payload 를 캡처하면 (1) 이 성립하고, 판독을 §E.2 에 적으면 (2) 가 성립한다. (3) 은 (a) 가 "아니오" 인 경우에만 판정 대상이 된다 |

**뮤턴트 탐침과 그 한계**: 자기 손으로 쓴 합성 JSON 을 `live-payload.json` 이라는 이름으로 저장하는 구현은 (1) 을 만족하고, (2) 도 그 합성 파일에 대고 답하면 만족한다. **이 기준은 캡처된 payload 와 손으로 쓴 payload 를 기계적으로 가르지 못한다** — 그것이 이 기준의 알려진 한계이며 부채로 신고한다. 이 기준이 실제로 바꾸는 것은 다른 것이다: M0 을 건너뛰면 이제 **완료 판정이 붉어진다**(파일 부재 → rc≠0). 종전에는 건너뛰어도 일곱 기준이 전부 초록이었다. 조용한 전제가 시끄러운 전제로 바뀐다.

---

### AC-SEC-001 — go 0-매치 실행이 관측된 pass 로 기록되지 않는다 (REQ-SEC-001, REQ-SEC-003)

**Given** `classifyTestCommand("go test ./internal/config -run TestNope", <출력에 `ok  \tgithub.com/modu-ai/moai-adk/internal/config\t0.424s [no tests to run]` 를 담고 exit-code 신호는 없는 payload>)

**When** 분류기를 호출한다

**Then** `isTest=true`, `isPass=false`, `isFail=false` 이고, `buildBashRecord` 가 만든 레코드의 `IsTestPass` 가 `false`, `Outcome` 이 `success` 가 아니다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | 이 트리에서 그 표본은 `isPass=true` 로 분류된다. 붉은 **사유**: `deriveFromOutputText` 가 `"ok  \t"` 를 정밀 pass 마커로 보고(`evidence_writer.go:217`) 어떤 실행 수도 보기 전에 `return true, false, true` 한다(`:223`). 토큰이 코드에 아예 없다는 사실은 직접 잰다 — `grep -rnc 'no tests to run\|no test files\|no tests ran' internal/hook/ \| grep -v ':0'` → 출력 0줄, `grep-exit=1` (이 워크트리, 2026-08-29). 러너 쪽 재현: `go test ./internal/kanban -run TestT341NoSuchTestXYZ -count=1` → `ok  ... [no tests to run]`, `exit=0` |
| **초록 경로** | M1 이 0-실행 거부권을 넣으면 뒤집힌다. 통과 출력은 위 Then 세 값 + 레코드 두 필드 |

**뮤턴트 탐침**: 마커 목록에서 `"ok  \t"` 를 통째로 지우면 이 기준은 만족되지만 요구는 깨진다(진짜 pass 도 신호를 잃는다). 그래서 AC-SEC-003 과 짝으로만 채택한다.

---

### AC-SEC-002 — exit-code 0 을 담은 payload 에서도 0-실행이 pass 로 넘어가지 않는다 (REQ-SEC-002)

**Given** 같은 0-실행 출력 텍스트를 담되 **`{"exit_code": 0}` 를 함께 담은** payload

**When** `classifyTestCommand` 를 호출한다

**Then** `isPass=false`.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | `isPass=true`. 붉은 **사유**: `classifyTestCommand` 가 exit-code 신호를 **먼저** 보고 거기서 반환한다 — `evidence_writer.go:69` 가 `deriveFromExitCode` 이고 텍스트 휴리스틱은 `:74` 다. 0-매치 실행의 종료코드는 정확히 0 이므로(§AC-SEC-001 의 `exit=0` 실측) 텍스트는 읽히지도 않는다 |
| **초록 경로** | M1 이 거부권을 두 분기 **앞**에 놓으면 뒤집힌다. 통과 출력은 `isPass=false` |

**뮤턴트 탐침**: 거부권을 `deriveFromOutputText` **안**에만 넣은 구현은 AC-SEC-001 은 통과하지만 이 기준에서 붉다. 두 기준이 짝일 때만 실제 훅 경로가 덮인다.

---

### AC-SEC-003 — 진짜 pass 는 종전 그대로다 (REQ-SEC-005) · 반대 방향

비발화 방향은 발화 방향(AC-SEC-004)이 덮는 것과 **같은 러너 축 전부**를 덮는다 — go / cargo / pytest / jest·vitest. `npm`·`pnpm`·`yarn` 은 그 아래 러너의 출력을 그대로 흘리는 래퍼다 — **그 아래에 무엇이 있는지는 래퍼가 정하지 않는다.** 아래가 jest·vitest·pytest 일 때에 한해 (c)(d)(e) 표본에 흡수되고, node 내장 러너처럼 어느 표본과도 다른 출력 어휘를 쓰는 러너를 앞세울 수도 있다. 그 경우는 (f) 가 따로 고정한다. 한 축만 덮으면 대칭이 계측기 단위로만 성립하고 러너 단위로는 깨진다(`spec.md` §3.5).

**Given** 실행 수가 1 이상인 진짜 pass 출력 표본 다섯 — 축마다 하나씩(jest·vitest 는 두 개), **표본이 타는 코드 경로를 명시한다**:

| # | 러너 계열 | 진짜 pass 표본 | 오늘 이 표본을 pass 로 만드는 경로 |
|---|---|---|---|
| a | go | `ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.656s` | 정밀 마커 `"ok  \t"` (`evidence_writer.go:217`) |
| b | cargo | `test result: ok. 12 passed; 0 failed; 0 ignored` | 정밀 마커 `"test result: ok"` (`:219`). `" failed"` 를 담고도 정밀 마커가 앞서므로 pass 다 |
| c | pytest | `3 passed in 0.00s` | 카운트 분기 `" passed"` (`:232`) — 정밀 마커 **없음** |
| d | jest·vitest | `Tests:       5 passed, 5 total` | 카운트 분기 `" passed"` (`:232`) — 정밀 마커 **없음** |
| e | jest·vitest, 두 자리 수 | `Tests:       10 passed, 10 total` | 카운트 분기 `" passed"` (`:232`). 이 표본이 겨누는 것은 다른 실패다 — 아래 참조 |
| f | node 내장 러너 (`npm test` 로 도달) | `TAP version 13\n# Subtest: a\nok 1 - a\n1..1\n# tests 1\n# pass 1\n# fail 0` — **네 pass 마커를 하나도 담지 않는다** | **exit-code 경로 단독**: `deriveFromExitCode` (`evidence_writer.go:69` → `:163`, `exit_code == 0` → `return true, false, true`). `deriveFromOutputText` 는 정밀 마커도 `" passed"` 도 못 찾아 `ok=false` 로 빠진다(`:236`) — 오늘 이 표본을 pass 로 만드는 유일한 근거다 |

(b)(d) 의 축자 문자열은 M1 이 실제 러너 출력으로 확인해 고정한다(AC-SEC-004 와 같은 규율, `spec.md` §5). (a)(c) 는 이 트리에서 실측했다.

각 표본에 대해 exit-code 0 payload 와 exit-code 없는 payload 두 가지를 만든다. (a)~(e) 다섯 표본이 열 개다.

**표본 (f) 는 짝을 만들지 않는다 — `{"exit_code": 0}` 를 담은 payload 하나뿐이다.** exit-code 를 빼면 이 텍스트에는 pass 로 읽힐 신호가 하나도 남지 않아(`deriveFromOutputText` 가 `ok=false`) 오늘도 `isPass=false` 이며, 그것은 회귀가 아니라 정상이다. 따라서 payload 는 **열한 개**이고, 아래 **Then** 은 열한 번째에도 그대로 적용된다 — (f) 의 payload 도 `isPass=true`, `isFail=false`, `Outcome=success` 여야 한다. (f) 가 고정하는 것은 텍스트 축이 아니라 **exit-code 축의 비발화 방향**이다.

(f) 의 출력은 이 트리에서 실측했다(`a6bbbf82b`, 2026-08-29). `node --test /tmp/t341f/nt.test.js` → `node-exit=0`, 그리고 그 출력에 대해 네 pass 마커 + fail 마커를 전부 셌다 — `grep -c -P 'ok  \t'` → `0`, `grep -c -P 'ok \t'` → `0`, `grep -c 'test result: ok'` → `0`, `grep -c ' passed'` → `0`, `grep -c ' failed'` → `0`. 축자 문자열 자체는 M1 이 러너 판번을 명시해 고정한다((b)(d) 와 같은 규율, `spec.md` §5). `npm test` 가 `testCommandSignatures` 에 있다는 사실도 확인했다(`evidence_writer.go:29`).

**When** 열 개 payload 각각으로 분류기를 호출한다

**Then** 열 개 전부 `isPass=true`, `isFail=false` 이고 레코드가 `Outcome=success` 다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | **열 개 전부 오늘 초록이다 — 그리고 그것이 이 기준이 하는 일이다.** 이 기준은 회귀 고정자이지 결함 탐지자가 아니며, 홀로 채택되지 않는다: AC-SEC-001·002·004 와 **짝으로만** 의미를 가진다(`verification-completeness.md` §2). 시작 관측이 없는 약속이 되지 않도록, 이 기준이 실제로 무엇을 잡는지를 **오늘 잰 두 사실**로 못 박는다 — ① pytest 의 진짜 pass 는 `pytest -q` 실측 결과 `3 passed in 0.00s` 이며 정밀 마커도 0-실행 토큰도 담지 않는다(이 워크트리, 2026-08-29), 따라서 `:232` 의 `" passed"` 분기 하나에만 매달려 있다. ② `printf 'Tests:       10 passed, 10 total\n' \| grep -c '0 passed'` → `1` — 두 자리 수 pass 출력이 `"0 passed"` 를 **부분 문자열로 담는다**. 이 두 사실이 아래 두 뮤턴트의 존재 조건이다 |
| **초록 경로** | M1 이후에도 열 개 같은 값. 값이 바뀌면 M1 이 살아 있는 pass 신호를 좁힌 것이다 |

**뮤턴트 탐침 1 (D2 가 실제로 쓰인 뮤턴트)**: 0-실행 거부권을 옳게 넣으면서 "안전을 위해" `:231-233` 의 `" passed"` 절을 함께 지운다. 확장 전 기준으로는 일곱 개가 전부 초록이었다 — 표본 (a) 가 정밀 마커로 살아남기 때문이다. 확장 후에는 (c)(d)(e) 가 붉어진다.

**뮤턴트 탐침 2**: jest·vitest 의 0-실행 거부권을 `"0 passed"` **부분 문자열**로 구현한다. 표본 (e) `10 passed` 가 그 문자열을 담으므로 진짜 pass 가 0-실행으로 판정된다 — 위 ② 로 실측된 충돌이며, 표본 (e) 가 없으면 잡히지 않는다. M1 은 러너별 토큰을 경계 있는 형태로(예: 줄 시작·단어 경계·`Tests:` 접두 포함) 고정해야 한다.

**뮤턴트 탐침 3 (exit-code 축)**: 0-실행 거부권을 옳게 넣으면서, 그것을 "텍스트에 인식 가능한 실행 수 근거가 없으면 0-실행으로 본다" 는 형태로 구현해 `deriveFromExitCode` 의 pass 경로를 좁힌다. `plan.md` 가 M1 에 지시하는 삽입 위치가 정확히 그 호출 지점(`:69`) 바로 앞이라 흘러가기 쉬운 형태다. 표본 (a)~(e) 는 전부 텍스트 마커를 담고 있어 이 뮤턴트에서도 살아남는다 — **(f) 하나만 붉어진다.** (f) 가 없으면 이 뮤턴트는 기준 전부를 통과하면서 `npm test`→node 내장 러너 계열의 진짜 pass 를 조용히 `isPass=false` 로 만든다(REQ-SEC-005 위반).

---

### AC-SEC-004 — 인접 러너 형태가 우연이 아니라 단언으로 고정된다 (REQ-SEC-001, REQ-SEC-003)

**Given** 러너별 0-실행 표본 — go `[no test files]`, pytest `no tests ran` 및 `collected 0 items`, jest·vitest `0 passed`, cargo `0 passed`(각 문자열은 M1 이 실제 러너 출력으로 확인해 고정한다, `spec.md` §5)

**When** 각 표본으로 분류기를 호출한다

**Then** 모두 `isPass=false` 다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | jest·vitest `0 passed` 와 cargo `0 passed` 는 `isPass=true` 로 분류된다 — 붉은 **사유**: 카운트 휴리스틱이 `" passed"` 를 담기만 하면 통과로 읽는다(`evidence_writer.go:232`, (b) 분기; 짝은 `:229` 의 `" failed"`). go `[no test files]` 와 pytest `no tests ran` 은 오늘 신호를 내지 않지만 **어떤 테스트도 그것을 고정하지 않는다** — 이 기준이 없는 상태 자체가 붉다(`grep -rn 'no test files' internal/hook/` → 0건, AC-SEC-001 의 grep 과 같은 실행) |
| **초록 경로** | M1 이 러너별 표본 corpus 와 거부권을 넣으면 다섯 표본 모두 `isPass=false` |

---

### AC-SEC-005 — 0-실행이 표면에 뜬다 (REQ-SEC-004)

**Given** 0-실행 표본을 담은 PostToolUse `Bash` 입력

**When** PostToolUse 핸들러를 호출한다

**Then** 반환된 `HookOutput.SystemMessage` 가 안정된 센티널 토큰을 담고, 원장 레코드가 그 실행을 0-실행으로 양성 기록하며, `Decision` 은 비어 있다(차단하지 않는다). 핸들러 내부에서 오류가 나도 반환은 정상이다(fail-open).

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | `SystemMessage` 가 비어 있다. 붉은 **사유**: `post_tool.go` 의 Bash 분기는 `logEvidence(input)` 만 호출하고 그 함수는 "never alters HookOutput" 계약을 명시한다(`post_tool.go:219-224`). 0-실행을 표시하는 코드 경로가 존재하지 않는다 |
| **초록 경로** | M2 가 navigator detect 와 같은 방식으로(`post_tool.go:239`) 같은 필드에 자문을 덧붙이면 뒤집힌다 |

**뮤턴트 탐침**: 자문을 stderr 로만 보내는 구현은 "경고를 냈다"를 만족하지만 이 기준에서 붉다 — 반환 payload 를 보기 때문이다.

---

### AC-SEC-006 — 표본 corpus 가 러너 taxonomy 를 망라한다 (REQ-SEC-006) · 지속 발화

**Given** `testCommandSignatures` 의 항목 집합과 0-실행 표본 corpus. **corpus 는 AC-SEC-004 가 분류기에 먹이는 바로 그 collection 이다** — 두 기준이 같은 하나의 변수를 읽으며, 망라 테스트용 사본을 따로 만들지 않는다.

**When** 망라 테스트를 실행한다

**Then** 다음 둘이 **함께** 성립한다.

1. signature 집합의 모든 러너 축이 corpus 에 최소 1개 표본을 갖는다. 없으면 테스트가 **실패하고 빠진 러너 이름을 출력한다**.
2. corpus 의 **각 항목이 실제로 분류기에 먹여지고**, `isTest=true` · `isPass=false` 이면서 **`detectZeroExecution(항목) == true`** 다. 세 번째 절이 핵심이다 — `isPass=false` 만 요구하면 부족하다: `classifyTestCommand` 는 인식 가능한 신호가 하나도 없을 때도 `return true, false, false` 하므로(`evidence_writer.go:79`, "부재 ≠ 실패"), **아무 의미 없는 문자열도 `isPass=false` 를 낸다.** 표본은 거부권을 **실제로 발화시켜야** 하며, 그저 pass 가 아니기만 해서는 안 된다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | corpus 도 망라 테스트도 존재하지 않는다. 붉은 **사유**: `evidence_writer_test.go` 는 `testCommandSignatures` 를 순회하는 단언을 담지 않으며(`grep -n 'testCommandSignatures' internal/hook/evidence_writer_test.go` → 0건), 0-실행 fixture 자체가 없다(AC-SEC-001 의 grep) |
| **초록 경로** | M3 이 망라 테스트를 넣으면 초록. **이후 러너가 추가되면 표본 없이는 다시 붉어진다** — 이것이 이 SPEC 의 §1.3 답이다: 판정이 조용히 좁아지는 길이 CI 에서 막힌다 |

**독자가 멈춤을 아는 법**: corpus 축소·마커 목록 편집·러너 추가 셋 다 이 테스트를 붉게 만든다. 요청하지 않아도 도착하는 신호다.

**뮤턴트 탐침**: corpus 를 signature 목록에서 파생시켜 `map[sig]""` 로 채운다. 조건 (1) 은 정의상 항상 초록이고, 빈 문자열은 분류기에서 `isPass=false` 를 내므로(신호 없음 경로, `evidence_writer.go:79`) **`isPass=false` 만 요구하는 판본에서도 초록이다** — 아무것도 단언하지 않는 망라 테스트가 통과한다. 조건 (2) 의 `detectZeroExecution` 절이 이 뮤턴트를 붉게 만드는 유일한 절이다: 빈 문자열에는 0-실행 토큰이 없으므로 거부권이 발화하지 않는다.

---

### AC-SEC-007 — 규칙 조항과 템플릿 미러가 같은 변경 안에 있다 (REQ-SEC-007)

**Given** `verification-completeness.md` 의 §1.1 0-스윕 각주

**When** 조항 승격 후 로컬 판과 `internal/template/templates/.claude/rules/moai/development/verification-completeness.md` 를 비교하고 `make build` 를 실행한다

**Then** (1) 로컬 파일이 각주가 아닌 `[HARD]` 조항을 담고, (2) **두 파일이 바이트 동일하다** — `diff .claude/rules/moai/development/verification-completeness.md internal/template/templates/.claude/rules/moai/development/verification-completeness.md` 가 rc 0, 출력 0줄. 판단 탈출구 없음: 사유를 적으면 통과하는 형태가 아니다. (3) `make build` 가 rc 0 이다.

조건 (2) 가 rc 0 을 그대로 단언할 수 있는 근거는 실측이다 — 이 트리에서 두 파일은 **오늘 이미 바이트 동일**하다 — 위 `diff` 명령을 그대로 실행해 출력 0줄 · `diff-exit=0` 을 관측했다(이 워크트리, `a6bbbf82b`, 2026-08-29). 그리고, 승격될 조항은 SPEC-ID·카드 id·내부 날짜를 담지 않는 일반 산문이라 템플릿 중립성(`CLAUDE.local.md` §25)이 요구하는 중립화 차이가 생길 이유가 없다. 만약 M4 가 중립화 차이를 **실제로** 필요로 하게 되면, 그것은 이 기준을 우회할 사유가 아니라 **기준을 고치고 재감사할 사유**이며 blocker 로 보고한다.

| 칸 | 내용 |
|---|---|
| **RED-now (`a6bbbf82b`)** | 로컬 파일의 해당 문단이 문자 그대로 *"Evidence footnote (not a rule)"* 로 시작한다(`verification-completeness.md` §1.1 말미) — 조항이 아니므로 기계층이 인용할 수 없다. 붉은 **사유**: 승격이 아직 없다 |
| **초록 경로** | M4 가 조항 승격 + 미러 + `make build` 를 한 변경에 담으면 세 조건 모두 성립 |

**뮤턴트 탐침**: 로컬만 고치고 미러를 빼면 조건 (1)은 만족하나 (2)에서 붉다. `moai update` 가 로컬 판을 템플릿판으로 되돌리는 경로(CLAUDE.local.md §2.3)를 이 기준이 막는다.

---

## 완료 정의 (Definition of Done)

- 수락 기준 8건 전부 초록이며, 각 기준의 RED-now 가 **구현 전 트리에서 실제로 관측된 뒤** 뒤집혔다.
- **M0 산출물이 존재한다** — `test -f .moai/reports/t341/live-payload.json` 이 rc 0 이고, 그 판독이 진행 기록 §E.2 에 AC-SEC-000 의 세 질문 (a)(b)(c) 에 답하는 형태로 적혀 있다. (a) 가 "아니오" 인 경우에는 blocker 보고가 §E.2 에 적히고 `internal/hook/evidence_writer.go` 에 이 카드의 커밋이 없다 — 어느 분기든 **판정된다**. M0 을 건너뛴 상태는 이 줄에서 붉다.
- 양방향이 모두 관측됐다 — AC-SEC-001·002·004(발화)와 AC-SEC-003(비발화, 러너 축 전부).
- `go test ./internal/hook/... -count=1` 이 초록이고, 그 실행이 **0-실행이 아님**을 실행 수로 확인한다(이 SPEC 의 자기 적용).
- **`.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` 가 이 변경에서 한 바이트도 바뀌지 않았다** — t343 결속(`spec.md` §2.4). 다음 **두 검사가 모두** 빈 출력이어야 하며, 하나라도 출력이 있으면 붉다:
  1. `git diff --name-only origin/develop...HEAD -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` — 브랜치에 **커밋된** 편집을 잡는다(병합 기준점 대비 삼중점).
  2. `git status --porcelain -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` — **작업트리와 인덱스** 편집을 함께 잡는다(스테이징으로 숨지 않는다).

  **왜 종전 형태를 버렸는가 — 실측.** 종전 DoD 는 인수 없는 `git diff --name-only \| grep SPEC-TODO-SQLITE-001` 이었다. 그 파일에 한 바이트를 넣고 `git add` 한 뒤 이 워크트리(`a6bbbf82b`, 2026-08-29)에서 잰 결과:

  | 형태 | 편집+스테이징 상태에서의 출력 | 판정 |
  |---|---|---|
  | 종전 `git diff --name-only \| grep …` | 출력 0줄, `old-grep-exit=1` | **초록 — 놓친다** |
  | 새 검사 2 `git status --porcelain -- <path>` | `M  .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` | **붉다 — 잡는다** |

  (측정 후 `git restore --staged` + `git checkout --` 로 되돌렸고 `git status --porcelain -- <path>` 가 다시 빈 출력임을 확인했다.) 종전 형태는 M4 가 편집하고 **커밋하면** 영구히 초록이 되는 구조적 상시-초록 검사이며, 이 SPEC 이 스스로 범위 밖으로 선언한 "0-히트 통과 조건" 바로 그 형태였다(`spec.md` §3.6). 검사 1 의 비공허성도 같은 트리에서 확인했다 — `git diff --name-only HEAD~3...HEAD -- .claude/rules/moai/development/spec-frontmatter-schema.md` 는 그 경로를 출력한다(범위 안에 그 파일을 건드린 커밋이 있으므로). 지금 이 브랜치는 커밋이 0건이라(`git merge-base origin/develop HEAD` = `a6bbbf82b` = HEAD) 검사 1 은 오늘 빈 출력이며, M4 가 결속 파일을 커밋하는 순간 붉어진다.
- 커밋 메시지마다 카드 id `t341` 이 들어 있다.
