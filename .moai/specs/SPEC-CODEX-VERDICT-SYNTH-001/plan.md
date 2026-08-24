---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "구현 계획 — codex verdict 합성의 관대 편향 제거"
version: "0.2.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, audit-multi, verdict, convergence, version-drift"
---

# 구현 계획 — SPEC-CODEX-VERDICT-SYNTH-001

## §A 맥락

지배 원칙은 spec.md §0 — **판정 불가를 통과로 읽지 않는다**. 이 계획은 spec.md §A.4 의 G1~G4 를 없애며, 그 형태는 "정규식을 0.149.0 에 다시 맞추기" 가 아니라 **서식 드리프트를 견디는 판정** 이다(spec.md §A.3).

착수 전 반드시 spec.md **§A.2 (cause.md 스테일 구간)** 을 읽을 것 — R1·R2 의 큰 몫은 t178·t186 에서 이미 착지했고, 여기서 다루는 것은 잔여분이다.

착수 순서는 확정돼 있다: **이 SPEC 이 t234 (= GitHub #1632) 보다 먼저 착지한다**(spec.md §C). `Findings: []Finding{}` 는 손대지 않는다.

## §B 착수 전 확인 (pre-flight)

착수하는 레인은 아래를 **실행해서** 확인한다. 인용은 실행 출력으로 한다.

```bash
# 1) 명시 verdict 파서가 트리에 있는가 (있어야 정상 — 없으면 base 가 t178 이전)
grep -n 'codexStatedVerdict' internal/cli/mcp_codex.go

# 2) 관대 기본값이 아직 있는가 (있어야 정상 — 이 SPEC 의 대상)
grep -n 'verdict := "pass"' internal/cli/mcp_codex.go

# 3) 기존 테스트가 통과하는가 (base 초록 확인)
go test ./internal/cli/ -run TestSynthesizeReviewOutput -v -timeout 600s
```

## §C 설계 결정

### C.1 [결정] 모드 seam — 기존 `method` 를 그대로 쓴다

`synthesizeReviewOutput` 의 유일한 프로덕션 호출자는 `runTurn`(`internal/cli/mcp_codex.go:680`)이고, 그 시그니처가 이미 `method string` 을 받는다. 따라서 **새 파라미터를 파이프라인에 뚫을 필요가 없다**.

```go
// 변경 전
func synthesizeReviewOutput(reviewText string) ReviewOutput

// 변경 후 (제안)
func synthesizeReviewOutput(reviewText, method string) ReviewOutput
```

`runTurn` 의 마지막 줄 `return synthesizeReviewOutput(reviewText), nil` 을 `return synthesizeReviewOutput(reviewText, method), nil` 로 바꾸는 것이 전부다.

모드 매핑:

| method | 모드 | 아는 서식이 하나도 안 맞을 때 |
|---|---|---|
| `codexMethodReviewStart` | native review | `pass` (보존 — 게이트 clean-pass 가 걸림) |
| `codexMethodTurnStart` | adversarial | `inconclusive` |
| 그 외 / 빈 문자열 | 미상 | `inconclusive` (보수) |

**대안으로 기각한 것**: 별도 `mode` enum 을 만들어 호출 체인에 뚫는 안. `method` 가 이미 모드를 1:1 로 결정하므로 두 번째 진실 원천을 만드는 셈이고, 둘이 어긋나면 아무도 모른다.

> t234 착수자 주의: 이 시그니처 변경을 되돌리면 모드 구분이 사라져 spec.md §0 의 원칙이 깨진다.

### C.2 [결정] 보수 채택은 명시적 순위 함수로

대입 순서가 아니라 순위 테이블로 바꾼다. 대입 순서 구현은 신호가 3개가 되는 순간 조용히 틀린다.

```go
func verdictRank(v string) int  // fail=2, inconclusive=1, pass=0, 그 외=1
func moreConservative(a, b string) string
```

합성은 `모드기본값 → 명시 verdict → 불릿 추론` 세 신호의 **최댓값**을 채택한다.

### C.3 [결정] 서식 인식기는 목록으로, 기본값은 목록 밖에

점수 표기 인식은 `codexStatedVerdict` 를 넓히지 않고 `codexScoredVerdict` 를 새로 두어 처리한다. 기존 정규식은 "verdict 라벨이 줄 머리에 온다" 는 좁은 계약을 의도적으로 지키고 있고(주석에 기록됨), 여기에 점수 형태를 욱여넣으면 그 좁음이 깨진다.

**드리프트 내성의 핵심은 인식기를 늘리는 것이 아니라 fall-through 를 고치는 것이다.** 인식기를 몇 개 두든, 하나도 안 맞았을 때 adversarial 경로가 `pass` 로 떨어지면 다음 버전에서 같은 사고가 난다. 따라서:

- 아는 서식은 목록으로 둔다 (`codexStatedVerdict`, `codexScoredVerdict`, `codexFindingBullet`)
- 목록이 전부 비었을 때의 값은 **모드 기본값** 이며, adversarial 에서는 `inconclusive` 다
- 새 서식을 나중에 추가할 때 이 fall-through 는 건드리지 않는다

주의: `FAIL` / `PASS` 같은 단어는 산문에도 흔하다. 점수 표기 정규식은 **줄 머리 + 대문자 판정어 + 공백 + 0~1 소수** 로 좁게 잡는다. `"the suite reported PASS 12 times"` 가 매치되면 안 된다.

### C.4 [결정] 합성 근거는 `ReviewOutput` 에 필드 추가

두 신호가 갈렸을 때 그 사실을 실어 나를 자리가 지금 없다. `ReviewOutput` 과 `PerBackendVerdict` 에 `SynthesisNote string \`json:"synthesis_note,omitempty"\`` 를 추가한다. `omitempty` 라서 기존 소비자의 JSON 은 바뀌지 않는다.

`converge()` 는 required 든 advisory 든 **비어 있지 않은 `SynthesisNote` 가 하나라도 있으면** `disagreement_flag` 를 세우고 `residual_risk_note` 에 그 내용을 넣는다. 단, **`overall_verdict` 를 새로 막지는 않는다** — `disagreement_flag` 는 새 차단 범주가 아니라는 기존 불변식(C3, `mcp_convergence.go:134`)을 지킨다.

미확인 사항: `review-output.schema.json` 이 외부 계약으로 고정돼 있다면 필드 추가가 스키마 변경에 해당할 수 있다. 착수 시 그 파일의 소유 범위를 먼저 확인할 것.

## §D 마일스톤

순서는 **되돌리기 어려운 결정 먼저**다. 시그니처·스키마 변경이 위, 기계적 변경이 아래.

### M1 — 합성 시그니처 + 순위 테이블 + 드리프트 fall-through (되돌리기 가장 어려움)

- `synthesizeReviewOutput(reviewText, method string)` 로 시그니처 변경, `runTurn` 호출부 갱신
- `verdictRank` / `moreConservative` 도입, 대입 순서 구현 제거
- 모드별 fall-through 도입 (REQ-CVS-002, REQ-CVS-003)
- RED: 서식 미상 adversarial 본문이 `inconclusive`, 무불릿 review 본문이 `pass` 임을 거는 테스트 먼저

### M2 — 점수 표기 인식기 (REQ-CVS-001)

- `codexScoredVerdict` 추가, 세 신호 합성에 편입
- RED: `FAIL 0.75` / `PASS 0.88` 본문, 그리고 산문 오탐 방지 케이스

### M3 — 합성 근거 표면화 (REQ-CVS-004)

- `ReviewOutput.SynthesisNote` / `PerBackendVerdict.SynthesisNote` 추가
- `converge()` 에서 `disagreement_flag` + `residual_risk_note` 반영
- RED: 신호가 갈린 codex 본문 → `disagreement_flag == true`, 그러나 `overall_verdict` 는 기존 정책대로

### M4 — 회귀 고정 (REQ-CVS-005)

- `TestSynthesizeReviewOutput_FindingBulletsMapToFail` 를 review-mode 명시 호출로 확장 (삭제 금지)
- `codex_task` 의 반환 `Output` 텍스트 불변 확인

## §E 자가 검증

```bash
go test ./internal/cli/... -timeout 600s
go vet ./internal/cli/...
GOOS=windows go vet ./internal/cli/...
```

`go test ./...` 은 실행하지 않는다 (저장소 규율 — 전 패키지 판정은 CI 몫).

## §F 위험과 안티패턴

| 위험 | 완화 |
|---|---|
| review-mode clean-pass 를 깨뜨림 → Stop 훅 게이트가 정상 변경을 막음 | AC-CVS-005 가 review-mode 무불릿 = `pass` 를 못 박는다. M1 RED 에 포함 |
| t234 (= #1632) 와 같은 함수에서 충돌 | 순서 확정: 이 SPEC 선착지. spec.md §C 에 산문으로 기록 |
| `codex_task` 가 같은 `runTurn` 을 쓰므로 fall-through 가 바뀜 | `codex_task` 는 `Verdict` 를 쓰지 않고 `Summary` 만 쓴다(`codex_task.go:118` 주석). AC-CVS-008 이 고정 |
| 점수 표기 정규식이 산문을 오탐 | C.3 의 좁은 계약 + AC-CVS-003 의 오탐 케이스 |
| 스테일 바이너리/서버 프로세스로 "안 고쳐졌다" 오판 | 검증은 `go test` 로 한다. MCP 라이브 프로브를 근거로 쓰지 않는다 (spec.md §A.2) |

**안티패턴 1 — 감지만 하고 관대한 쪽 채택**: 불일치를 로그에만 남기고 `pass` 를 그대로 반환하는 구현. acceptance.md 의 대표 mutant 이며, AC-CVS-002 가 **채택값** 을 걸어 잡는다.

**안티패턴 2 — 인식기만 늘리고 fall-through 를 남겨둠**: 0.149.0 서식용 정규식을 하나 더 추가했는데 `verdict := "pass"` 는 그대로인 구현. 눈앞의 사례는 고쳐지고 **구조적 원인은 그대로 남는다** — 다음 버전에서 같은 사고가 난다. AC-CVS-004 가 이것을 잡는다.
