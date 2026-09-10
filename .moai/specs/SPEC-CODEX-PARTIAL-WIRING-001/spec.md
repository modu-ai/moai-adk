---
id: SPEC-CODEX-PARTIAL-WIRING-001
title: "반쪽 배선(agent TOML만 존재) 상태의 doctor 탐지와 조치 안내"
version: "0.3.2"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, wiring, doctor, partial-wiring, half-wired, detection, advisory, t499"
tier: M
related_specs: [SPEC-CODEX-WIRING-001, SPEC-CODEX-SIDECAR-GUARD-001]
---

# SPEC-CODEX-PARTIAL-WIRING-001 — 반쪽 배선 상태의 doctor 탐지

> 카드: **t499** · 선행 SPEC: `SPEC-CODEX-WIRING-001`(REQ-CW-009 존재-게이트, REQ-CW-010 / AC-CW-012 doctor 행 소유)

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-09-07 | 최초 작성 (plan-phase, 카드 t499). 레인이 선행 수행한 재현 기록(`.moai/reports/t499/repro.md`)을 측정 전제로 삼고 그 위에 요구·수용 기준을 얹음 |
| 0.1.1 | 2026-09-07 | 리드 지적: §F의 AC 범위가 `~008`로 적혀 AC-CPW-009(뮤턴트 관문)가 판정 집합에서 빠져 있었다 — `~009`로 정정 |
| 0.2.0 | 2026-09-07 | plan-audit iter1(FAIL 0.76) 수리 라운드. D3 → REQ-CPW-009 신설(codex 부재 갈래 지시문 금지 — §D-1의 판단을 요구로 강제). D5 → REQ-CPW-010 신설(존재-게이트 무접촉)하고 AC-CPW-008을 그쪽으로 재매핑. D6 → §A.3 좌표 `:105` → `:92` 정정(원문 오기 사실을 지우지 않고 정정 절로 남김). D7(optional, 수용) → §D-1에 "REQ-CPW-006은 회귀 반경을 지키는 조항이지 claude-only 사용자를 지키는 조항이 아니다" 구분 추가. D1/D2/D4는 `acceptance.md` 소관 |
| 0.3.0 | 2026-09-07 | plan-audit iter2(FAIL 0.84, 뮤턴트 관문) 수리 라운드. D1 → AC-CPW-002 (d)를 토큰 부재에서 **성질**(부분문자열 `moai init` 금지)로 재작성 + 패러프레이즈 뮤턴트 M6 신설. D2 → **REQ-CPW-011 신설**(부재 갈래는 사용자 계층 훑기에 도달하지 않는다) + §A.4에 네 번째 보존 계약 추가 + AC-CPW-002 Given에 `stubCodexHome`(낡은 항목 ≥1) 고정 + 뮤턴트 M7 신설. D3 → `progress.md`의 lint 증거 문장 재작성(SPEC 표기는 바꾸지 않음). D4 → AC-CPW-007 좌표 `:394`/`:441` → `:404`/`:450`. D5 → RED-now 5셀을 명령/stdout/`exit:` 3필드로 분해. 더해 리드 승인 보강: RED-now 셀에 `grep -c` 관측을 짝지어 §D.0 4-조합 판독표 신설 |
| 0.3.1 | 2026-09-07 | 리드 질의(허용 목록 전환) 채택. AC-CPW-002 (d)를 **금지 목록 → 등가 단언(allowlist)**으로 전환 — 부재 갈래 Message를 시험 파일 리터럴과 등가로 못박아 어떤 패러프레이즈도 원리상 통과하지 못하게 함. (d′)로 리터럴 자체의 모양 검사를 남기고, 사람 판정 DoD 항목은 제거(검토 지점이 체크리스트에서 diff에 보이는 코드 한 줄로 이동). 등가 단언이 닫지 못하는 잔여분(리터럴 선택 자체)을 (d′) 주석에 명시. 뮤턴트 **M6′**(세 금지 토큰을 전부 피한 지시문 — 종전 처리로는 통과, 등가 단언으로는 RED) 신설 → 뮤턴트 8종 |
| 0.3.2 | 2026-09-07 | 리드의 어셈블리 판독 확인 + 한 곳 반박. 등가 단언 성립은 확인(`doctor_codex.go:193-201`). 다만 "깨끗한 home일 때만 성립"은 정정 — 올바른 구현에서는 훑기에 도달하지 않으므로 home 내용과 무관하게 성립하며, **낡은 home에서도 등가를 요구하는 쪽이 더 강하다**. AC-CPW-002 Given을 하위 케이스 (i) 깨끗한 home · (ii) 낡은 home으로 갈라 (a)·(d)를 둘 다에서 단언(진단은 픽스처 약화가 아니라 케이스 분리로 얻는다). 리드 질의 2 → AC-CPW-007은 흡수하지 않고 유지(codex 존재 갈래는 `joinCodexSummaries` 합성이라 흡수 불가). (d′)는 유지하되 "(d)와 서로 다른 대상을 잰다"로 재서술 |

## §A. 측정 전제 (Verified baseline)

아래 사실은 전부 **레인이 이 워크트리에서, 이 트리(`ace1c5440`)로 빌드한 `./bin/moai`로 실측**했다.
원본 증거는 `.moai/reports/t499/repro.md` + `.moai/reports/t499/doctor-nocodex-path.txt`이며,
run-phase에서 다시 잴 필요는 없다(인용을 위한 재독은 허용).

### §A.1 plain init이 남기는 상태

`moai init <dir> --non-interactive`(즉 `--agent` 플래그 없음)는 `.codex/agents/moai/` 아래
agent 정의 TOML을 배포하고, 배선 파일 3종(`.codex/hooks.json`, `.codex/config.toml`,
`.moai/state/codex-wiring.json`)은 **하나도 만들지 않는다**. 즉 plain init의 산물이 곧 반쪽 배선이다.

### §A.2 그 상태에 대한 현행 doctor 판정 두 갈래

| PATH 조건 | 현행 `Codex Wiring` 행 | 문제 |
|---|---|---|
| `codex` 있음 | `codex installed, project not wired — run moai init --agent codex` | 지시문은 우연히 옳으나 **문구가 거짓** — 프로젝트는 Codex agent 정의를 갖고 있다 |
| `codex` 없음 | `not wired (claude-only project) — skipped` | **조용한 구멍** — agent 정의를 가진 프로젝트를 "claude-only"라고 단언한다 |

### §A.3 근본 원인 (읽은 것이지 추론이 아님)

`internal/cli/doctor_codex.go:92`

> 0.2.0 정정(감사 D6): 초판은 이 좌표를 `:105`로 적었다. 인용한 코드 조각은 옳았고 **좌표만** 틀렸다.
> 실측 — `grep -n 'wired := hooksErr' internal/cli/doctor_codex.go` → `92:`(트리 `ace1c5440`).
> 오기의 출처는 `.moai/reports/t499/repro.md:67`이며 그쪽은 리드가 정정했다.

```go
wired := hooksErr == nil || cfgErr == nil
```

판별식이 읽는 것은 배선 파일 2종뿐이다. `.codex/agents/`는 어느 분기에서도 조회되지 않으므로
반쪽 상태는 판별식 안에 자리가 없고, `!wired && !codexInstalled` 조기 반환이 §A.2의 조용한 OK를 만든다.

### §A.4 보존해야 하는 계약 4종

| 계약 | 좌표 | 내용 |
|---|---|---|
| 존재-게이트 | `internal/codexwiring/wire.go:51-56` | `RefreshWiring`는 `wiringFilesExist`가 거짓이면 `Result{}, nil`을 돌려주고 **아무것도 만들지 않는다** |
| 읽기 전용·비차단 | `internal/cli/doctor_codex.go` 파일 머리 주석 | 이 검사는 보고만 하며 `.codex/` 파일을 만들지도 고치지도 지우지도 않고, `~/.codex/config.toml`을 건드리지 않으며, 게이트로 쓰이지 않는다 |
| un-nagging 불변 | `internal/cli/doctor_codex.go`의 `!wired && !codexInstalled` 조기 반환 + `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent` | 진짜 claude-only 프로젝트(`.codex/` 자체가 없음)를 codex 없는 머신에서 볼 때 검사는 침묵한다 |
| **사용자 계층 훑기의 도달 범위**(0.3.0 추가) | `internal/cli/doctor_codex.go:189`(`codexStaleSkillFinding` 호출, 함수 선언 `:376`) + 그 위 `:186-188` 주석 | 이 훑기는 "Codex가 관여할 때만 도달한다"는 전제 위에 있고, 오늘 그 전제는 조기 반환이 지킨다. **half-wired × codex 부재가 그 전제를 깨는 첫 사례**이므로, 이 갈래가 훑기에 도달하는지 여부는 이 SPEC이 명시해야 하는 계약이다(REQ-CPW-011) |

### §A.5 골든 픽스처 영향 (읽어서 확인)

`internal/cli/doctor_golden_test.go`의 하네스는 빈 `t.TempDir()`로 `t.Chdir`하고
`codexWiringLookPath`를 `codex` 부재로 고정한다. 즉 골든이 잡는 프로젝트에는 `.codex/`가 아예 없으므로,
`doctor-{light,dark,nocolor}.golden`의 `not wired (claude-only project) — skipped` 행은
이 SPEC의 변경 뒤에도 그대로여야 한다(REQ-CPW-006).

## §B. 문제 서술 (Why)

plain init을 쓴 모든 프로젝트가 반쪽 배선 상태로 남는데, 그 상태를 진단하는 유일한 표면인 `moai doctor`가
두 갈래 모두에서 **사실이 아닌 문장**을 낸다. codex가 깔린 머신에서는 배선이 하나도 없다는 취지의 문구가
agent 정의의 존재를 지우고, codex가 없는 머신에서는 아예 "claude-only"라고 잘못 단언한다.
사용자는 자기 프로젝트가 어느 상태인지 doctor에게서 알 수 없다.

## §C. 요구 (GEARS)

> 주체는 `moai doctor`의 `Codex Wiring` 검사(`checkCodexWiring`)다.

| ID | 유형 | 요구 |
|---|---|---|
| REQ-CPW-001 | Ubiquitous | `Codex Wiring` 검사는 프로젝트의 Codex 배선 상태를 **세 가지**로 구분해야 한다(SHALL): `unwired`(agent 정의 없음 + 배선 파일 없음), `half-wired`(agent 정의 있음 + 배선 파일 둘 다 없음), `wired`(배선 파일이 하나 이상 있음) |
| REQ-CPW-002 | Event-driven | `half-wired`이고 `codex`가 PATH에서 해석될 때, 검사는 그 상태를 이름 붙여 보고하고(SHALL) 조치 지시문 `run moai init --agent codex`를 **Message에** 실어야 한다 |
| REQ-CPW-003 | Event-driven | `half-wired`이고 `codex`가 PATH에 없을 때, 검사는 그 상태를 이름 붙여 보고해야 한다(SHALL) — 이때 검사 상태는 `CheckOK`(정보성)로 유지한다 |
| REQ-CPW-004 | Unwanted | 어떤 분기에서도 검사는 `half-wired` 프로젝트를 `claude-only`라고 서술해서는 안 된다(SHALL NOT) |
| REQ-CPW-005 | Ubiquitous | 반쪽 상태의 판별은 agent 정의 디렉터리에 정의 파일이 **하나 이상 존재하는지**로만 이뤄져야 한다(SHALL). 고정된 개수(현행 11)를 판별식에 넣어서는 안 된다 |
| REQ-CPW-006 | State-driven | 프로젝트가 `unwired`인 동안, 두 PATH 갈래의 기존 Message는 **글자 그대로 보존**되어야 한다(SHALL) — codex 부재 시 `not wired (claude-only project) — skipped`, codex 존재 시 `codex installed, project not wired — run moai init --agent codex` |
| REQ-CPW-007 | Unwanted | 검사는 `.codex/` 파일이나 `~/.codex/config.toml`을 생성·수정·삭제해서는 안 되며(SHALL NOT), 어떤 발견도 `moai doctor`의 종료 코드를 바꾸어서는 안 된다 |
| REQ-CPW-008 | Ubiquitous | 새로 추가되는 Message 문구는 기존 폭 상한 `codexMessageWidthCeiling`(113 runes)을 지켜야 한다(SHALL) |
| REQ-CPW-009 | Unwanted | `half-wired`이고 `codex`가 PATH에 없을 때, 그 Message는 조치 지시문(`initCodexAdvice` = `run moai init --agent codex`)을 실어서는 안 된다(SHALL NOT). 실행할 수 없는 지시는 조치 안내가 아니라 잔소리이며, §D-1의 판단은 이 요구로 강제되어야만 구현을 구속한다 |
| REQ-CPW-010 | Unwanted | 이 SPEC의 구현은 `internal/codexwiring`의 `RefreshWiring` / `wireProject` / `wiringFilesExist` 동작을 변경해서는 안 된다(SHALL NOT) — 존재-게이트(§A.4)는 이 카드에서 관측 대상이지 수정 대상이 아니다 |
| REQ-CPW-011 | Event-driven | `half-wired`이고 `codex`가 PATH에 없을 때, 사용자 계층 config(`~/.codex/config.toml`)가 낡은 `[[skills.config]]` 항목을 선언하고 있더라도 검사 상태는 `CheckOK`로 유지되어야 하며(SHALL), 그 항목에서 비롯된 소견이 Message나 Detail에 등장해서는 안 된다 — 즉 이 갈래는 사용자 계층 훑기(§A.4 네 번째 계약)에 도달하지 않는다 |

## §D. 설계 결정 (근거를 남긴다)

### D-1. codex 부재 갈래를 Warn으로 올리지 않는 이유

§A.1대로 **plain init이 모든 프로젝트에 agent TOML을 깐다**. 그러므로 반쪽 상태를 무조건 `CheckWarn`으로
올리면, Codex를 쓸 생각이 전혀 없는 claude-only 사용자의 모든 프로젝트에 경고 행이 상시로 뜬다.
이는 §A.4의 un-nagging 계약이 막으려던 바로 그 결과다. 따라서 갈래를 나눈다: codex가 실제로 깔려 있어
지시문이 **당장 실행 가능할 때만** Warn(REQ-CPW-002), 깔려 있지 않으면 상태는 OK로 두되 문장만 사실에 맞춘다
(REQ-CPW-003). 거짓 단언은 없애고 잔소리는 늘리지 않는다.

**0.2.0 보강(감사 D7 — 무엇이 무엇을 지키는지 구분).** REQ-CPW-006(기존 문구 보존)을 "claude-only 사용자를
지켜 주는 조항"으로 읽으면 안 된다. §A.1대로 오늘 `moai init`을 돌린 사용자는 half-wired가 되어 **새 문구**
쪽에 떨어지므로, REQ-CPW-006이 실제로 덮는 집단은 `.codex/`가 아예 없는 프로젝트 — 옛 프로젝트, 손으로 만든
프로젝트, 그리고 골든 픽스처 — 로 사실상 한정된다. 즉 REQ-CPW-006은 **회귀 반경**을 지키는 조항이고(그 값어치는
그대로다), claude-only 사용자를 잔소리에서 지키는 것은 REQ-CPW-003의 `CheckOK`와 REQ-CPW-009의 지시문 금지다.

### D-2. codex 존재 갈래의 기존 문구를 "고치는가"

고치지 않는다 — **전제를 좁힌다**. `codex installed, project not wired — run moai init --agent codex`는
agent 정의까지 정말로 없는 프로젝트에서는 참이다. 이 SPEC은 그 문장이 나오는 조건을 `unwired`로 한정하고
(REQ-CPW-006), `half-wired`에는 별도 문장을 준다(REQ-CPW-002). 결과적으로 카드가 지적한 "문구가 거짓"인
사례는 사라지되, 기존 문장을 재작성하지 않으므로 그 문장을 고정한 기존 시험
(`TestCheckCodexWiring_UnwiredWithCodexInstalledWarns` — 빈 `t.TempDir()`을 쓰므로 여전히 `unwired`)은
수정 없이 계속 통과한다. 문구 재작성보다 조건 한정이 회귀 반경이 작다.

### D-3. 개수를 판별식에 넣지 않는 이유

11이라는 수는 템플릿 사실이고 템플릿 편집 한 번이면 바뀐다(레인의 잔여 위험 기록). 개수에 건 판별식은
템플릿이 12개가 되는 날 조용히 틀린다. 그래서 "비어 있지 않은가"만 묻는다(REQ-CPW-005).

## §E. 범위 밖 (out of scope)

### Out of Scope — 배선 파일 생성

- `moai doctor`가 `.codex/hooks.json` 또는 `.codex/config.toml`을 만드는 것
- `moai update` 또는 `codexwiring.RefreshWiring`가 반쪽 프로젝트를 자동 배선하는 것 — 존재-게이트(§A.4)는 **그대로 보존**되며 이 SPEC은 그 계약을 건드리지 않는다
- 반쪽 상태의 자동 수리(auto-repair)나 사용자 확인 없는 마이그레이션

### Out of Scope — doctor의 성격 변경

- 이 발견으로 `moai doctor`의 종료 코드를 바꾸거나 게이트로 승격시키는 것
- `~/.codex/config.toml`(user-layer) 쓰기
- `Codex Wiring` 외 다른 doctor 검사 행의 문구·상태 변경

### Out of Scope — init 쪽 동작

- plain init이 agent TOML을 배포하는 현행 동작 자체의 변경(배포를 막거나 `--agent` 뒤로 옮기는 것)
- 대화형 위저드 경로(`SPEC-INIT-HARNESS-PROMPT-001`)의 `agent_wiring` 응답 처리 — 레인이 측정하지 않은 갭이며, 이 SPEC은 위저드 동작을 규정하지 않는다
- Codex agent TOML의 내용·개수·발행 방식(`internal/template/agentemit` 소관)

## §F. 수용 기준

`acceptance.md`의 AC-CPW-001 ~ AC-CPW-009가 이 SPEC의 판정 기준이다. AC-CPW-009(뮤턴트 8종)는
다른 여덟 항목이 실제로 판별식을 잡고 있는지를 검사하는 반증 관문이므로 이 집합에서 빠질 수 없다.

0.2.0에서 아홉 항목은 두 부류로 나뉘어 표기된다 — **릴리스 게이트 6건**(AC-CPW-001 / -002 / -005 / -006 /
-007 / -009, 각각 트리 `ace1c5440`에서 실측한 RED-now 셀을 갖는다)과 **회귀 가드 3건**(AC-CPW-003 / -004 /
-008, 보존 기준이라 원리상 RED-now를 가질 수 없으므로 게이트로 세지 않는다).

## §G. 참조

- `.moai/reports/t499/repro.md` — 측정 전제의 원본
- `SPEC-CODEX-WIRING-001` — REQ-CW-009(존재-게이트) / REQ-CW-010 · AC-CW-012(doctor 행) 소유
- `internal/cli/doctor_codex.go` · `internal/cli/doctor_codex_test.go` · `internal/codexwiring/wire.go`
