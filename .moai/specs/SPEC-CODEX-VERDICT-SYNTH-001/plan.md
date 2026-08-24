---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "구현 계획 — 모르는 서식을 통과로 읽지 않는다"
version: "0.4.0"
status: draft
created: 2026-08-24
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: S
tags: "codex, verdict, version-drift, tdd"
---

# 구현 계획 — SPEC-CODEX-VERDICT-SYNTH-001

## §A 맥락

지배 원칙은 spec.md §0 — **판정 불가를 통과로 읽지 않는다**. 일차 근거는 `.moai/reports/t229/premise-revision.md` 의 7행 실측표이며, `cause.md` 와 어긋나면 그쪽이 이긴다.

Tier S. 결함 3건(G1·G3·G4) + 회귀 방어. 착수 순서는 확정돼 있다: **이 SPEC → t234 (= GitHub #1632)**. `Findings: []Finding{}` 는 손대지 않는다.

**검증은 `go test` 로만 한다.** MCP 라이브 프로브를 근거로 쓰지 않는다 — spec.md §A.2 의 바이너리 랙(설치본이 `origin/main` 보다 259 커밋 뒤처짐)이 그 경로를 신뢰할 수 없게 만든다.

## §B 착수 전 확인 (pre-flight)

```bash
# 1) 관대 fall-through 가 아직 있는가 (있어야 정상 — 이 SPEC 의 대상)
grep -n 'verdict := "pass"' internal/cli/mcp_codex.go

# 2) 기존 테스트가 통과하는가 (base 초록 확인)
go test ./internal/cli/ -run TestSynthesizeReviewOutput -v -timeout 600s
```

## §C 설계 결정

### C.1 [결정] fall-through 를 고치는 것이 본체다

이 작업의 무게중심은 인식기 추가가 아니라 **아는 서식이 하나도 안 맞았을 때의 값**이다. 인식기를 몇 개 두든, 그 값이 adversarial 에서 `pass` 로 남으면 다음 버전에서 같은 사고가 난다(spec.md §A.4).

- 아는 서식은 목록으로 둔다 (`codexStatedVerdict`, 신설 `codexScoredVerdict`, `codexFindingBullet`)
- 목록이 전부 비었을 때의 값은 **모드 기본값**
- 새 서식을 나중에 추가할 때 이 fall-through 는 건드리지 않는다

### C.2 [결정] 모드 seam — 기존 `method` 를 그대로 쓴다

`synthesizeReviewOutput` 의 유일한 프로덕션 호출자는 `runTurn`(`internal/cli/mcp_codex.go:680`)이고 그 시그니처가 이미 `method string` 을 받는다. 새 파라미터를 뚫을 필요가 없다.

```go
func synthesizeReviewOutput(reviewText, method string) ReviewOutput
```

| method | 모드 | 아는 서식이 하나도 안 맞을 때 |
|---|---|---|
| `codexMethodReviewStart` | native review | `pass` (보존 — 게이트 clean-pass) |
| `codexMethodTurnStart` | adversarial | `inconclusive` |
| 그 외 / 빈 문자열 | 미상 | `inconclusive` (보수) |

**기각한 대안**: 별도 `mode` enum 을 호출 체인에 뚫는 안. `method` 가 이미 모드를 1:1 로 결정하므로 두 번째 진실 원천이 생기고, 둘이 어긋나면 아무도 모른다.

> t234 착수자 주의: 이 시그니처 변경을 되돌리면 모드 구분이 사라져 spec.md §0 이 깨진다.

### C.3 [결정] 점수 표기 인식기는 좁게

`codexStatedVerdict` 를 넓히지 않고 `codexScoredVerdict` 를 신설한다. 기존 정규식은 "verdict 라벨이 줄 머리에 온다" 는 좁은 계약을 의도적으로 지키고 있고, 여기에 점수 형태를 욱여넣으면 그 좁음이 깨진다.

`FAIL` / `PASS` 는 산문에도 흔하므로 **줄 머리 + 대문자 판정어 + 공백 + 0~1 소수** 로 좁게 잡는다. `"the suite reported PASS 12 times"` 가 매치되면 안 된다(AC-CVS-002 후반).

### C.4 [결정] 불일치 기록 필드

`ReviewOutput` 과 `PerBackendVerdict` 에 `SynthesisNote string \`json:"synthesis_note,omitempty"\`` 를 추가한다. `omitempty` 라 기존 소비자의 JSON 은 바뀌지 않는다. `converge()` 는 비어 있지 않은 `SynthesisNote` 가 하나라도 있으면 `disagreement_flag` 를 세우고 `residual_risk_note` 에 담되, **`overall_verdict` 를 새로 막지는 않는다**(`mcp_convergence.go:134` 의 기존 불변식 C3).

~~미확인 사항~~ **종결(실측)**: `review-output.schema.json` 은 **이 저장소에 파일로 존재하지 않는다** — `find . -name 'review-output.schema.json' -not -path './.git/*'` 무출력, `git ls-files | grep -c` → `0`. 이름이 등장하는 곳은 SPEC 문서와 Go 주석 등 산문 참조뿐이다. 따라서 `SynthesisNote` 추가는 파일로 고정된 스키마를 깨지 않는다. 착수 시 이 항목을 다시 열 필요가 없다.

### C.5 [비결정] 보수 채택의 구현 형태 — 요구사항 아니되, 성질은 AC 가 강제한다

순위 테이블(`verdictRank`/`moreConservative`)을 쓸지 대입 순서를 유지할지는 **구현자 자유**다. 요구사항으로 두지 않는다 — 현재 트리에 반례가 없으므로 결함이 아닌 것을 고치는 데 감사 예산을 쓰지 않는다(spec.md §A.5 전반부).

**단, M2 가 세 번째 신호를 들이는 순간 성질 자체는 구속된다.** `codexScoredVerdict` 는 pass / fail / inconclusive 어느 값이든 대입할 수 있으므로, 아래 C.1 의 신호 순서(`stated` → `scored` → `bullet`) 위에서 **대입을 한 줄 더 얹는 방식**으로 쓰면 나중 대입이 앞선 보수적 값을 덮는다:

```
body:  "Verdict: fail\n\nPASS 0.95 / 1.00"
       stated → fail  →  scored 덮어씀 → pass  →  불릿 없음  →  최종: pass
```

차단 판정이 통과로 세탁되며, 이는 spec.md §0 위반을 이 SPEC 이 새로 만들어 내는 형태다.

**지켜야 할 규칙은 spec.md §A.5 의 P-CONS 다 — 채택값은 신호 집합의 최댓값(`fail` > `inconclusive` > `pass`).** 위 반례는 그 규칙을 어긴 한 가지 사례일 뿐, 규칙 자체가 아니다. 규칙을 **"나중 신호가 앞선 신호를 덮지 않는다"** 같은 순서 서술로 옮겨 적지 말 것 — 신호가 셋인 동안만 맞고 넷째가 오면 다시 뚫린다. 집합의 최댓값에는 순서가 없다.

**AC-CVS-006 이 이 규칙을 채택값으로 고정한다** — 어떤 내부 형태를 택하든 그 AC 를 통과해야 한다. 순위 테이블은 만족시키는 가장 단순한 방법이지 유일한 방법은 아니며, 반대로 `stated × scored` 쌍만 특수 분기로 처리하는 방법은 오늘은 통과해도 넷째 신호에서 무너진다(acceptance.md §B-2 mutant g).

## §D 마일스톤

되돌리기 어려운 결정 먼저.

### M1 — 모드 seam + fall-through 교정 (본체)

- `synthesizeReviewOutput(reviewText, method string)` 시그니처 변경, `runTurn` 호출부 갱신
- adversarial fall-through 를 `inconclusive` 로, native 는 `pass` 유지 (REQ-CVS-001, REQ-CVS-004)
- RED 먼저: acceptance.md §B corpus 를 순회하는 속성형 테스트. 사전구현 트리에서 **C1~C8 전부**가 실패해야 한다(감사 iter1 실측: 8건 모두 `pass` 로 합성)

### M2 — 점수 표기 인식기 (REQ-CVS-002) [착수 게이트: AC-CVS-006]

- **착수 전제**: acceptance.md 에 AC-CVS-006(다중 신호 보수 해소)이 테스트로 존재해야 한다. 세 번째 신호를 들이는 순간 그 AC 가 유일한 가드가 되므로, 없는 상태로 M2 를 쓰면 spec.md §0 위반이 초록 신호를 달고 착지한다
- `codexScoredVerdict` 신설, 신호 목록에 편입
- **대입 한 줄 추가로 쓰지 말 것** — C.5 의 세탁 반례가 정확히 그 형태다
- RED: C8(`FAIL 0.75 / 1.00`) → `fail`, 그리고 산문 오탐 케이스
- GREEN 직후 AC-CVS-006 을 별도로 확인 (가드 AC 이므로 RED 기대 없음)

### M3 — 불일치 기록 (REQ-CVS-003)

- `SynthesisNote` 추가 + `converge()` 반영
- RED: 채택값 `fail` **과** 기록 필드 비어 있지 않음을 동시에 단언

### M4 — 회귀 고정 (REQ-CVS-004)

- 기존 테스트를 native 모드 명시 호출로 확장 (삭제 금지)
- 라이브 프로브 본문 픽스처 고정, `codex_task` 출력 불변 확인

## §E 자가 검증

```bash
go test ./internal/cli/... -timeout 600s
go vet ./internal/cli/...
GOOS=windows go vet ./internal/cli/...
```

## §F 위험과 안티패턴

| 위험 | 완화 |
|---|---|
| native clean-pass 를 `inconclusive` 로 바꿔 **관측된 판정을 미관측으로 오보** (차단은 되지 않음 — 게이트는 `fail` 접두사만 막는다) | AC-CVS-003. M1 RED 에 포함 |
| M2 가 세 번째 신호를 대입 한 줄로 얹어 보수적 값을 덮음 | AC-CVS-006 (M2 착수 게이트) |
| t234 (= #1632) 와 같은 함수에서 충돌 | 순서 확정: 이 SPEC 선착지. spec.md §D |
| `codex_task` 가 같은 `runTurn` 을 쓰므로 fall-through 가 바뀜 | `codex_task` 는 `Verdict` 를 쓰지 않고 `Summary` 만 쓴다. AC-CVS-005 가 고정 |
| 점수 표기 정규식이 산문을 오탐 | C.3 의 좁은 계약 + AC-CVS-002 후반 |
| 바이너리 랙으로 "안 고쳐졌다" 오판 | 검증은 `go test`. 라이브 프로브 금지 (spec.md §A.2) |

**안티패턴 1 — 인식기만 늘리고 fall-through 를 남겨둠**: 점수 표기 정규식을 하나 추가했는데 `verdict := "pass"` 는 그대로인 구현. 눈앞의 사례(G1)는 닫히고 **구조적 원인(G3)은 그대로** 남아 다음 버전에서 재발한다. AC-CVS-001 이 잡는다.

**안티패턴 2 — 서식 열거형 테스트**: 구현이 읽도록 만든 서식만 열거한 테스트. 구현을 검증하지 않고 되풀이한다. acceptance.md §B 참조 — 단언은 corpus 를 순회해야 하고, 구성원 추가가 단언문 수정을 요구하면 안 된다.

**안티패턴 3 — 세 번째 신호를 대입 한 줄로 얹기**: `codexScoredVerdict` 를 기존 대입 열에 추가하는, M2 를 쓰는 **가장 자연스러운 방식**. 나중 대입이 앞선 보수적 값을 덮어 `Verdict: fail` + `PASS 0.95` 가 `pass` 로 세탁된다. AC-CVS-006 의 K1 이 잡는다.

**안티패턴 3-b — 쌍 특수화로 막기**: 위 반례를 보고 `stated × scored` 조합만 따로 분기해 처리하는 수리. 오늘의 세 신호에서는 맞게 동작하므로 통과처럼 보이지만, **넷째 신호가 오는 날 같은 자리가 다시 뚫린다** — 구멍이 사라진 게 아니라 한 겹 옆으로 옮겨 간 것이다. 고쳐야 할 것은 쌍이 아니라 **집합의 최댓값을 취한다는 규칙**(P-CONS)이며, AC-CVS-006 의 K5·K6 이 이 수리를 가른다.

**안티패턴 4 — 감지만 하고 관대한 쪽 채택**: 불일치를 기록만 하고 `pass` 를 반환하는 구현. AC-CVS-004 가 **채택값** 을 걸어 잡는다.
