---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "인수 기준 — codex verdict 합성의 관대 편향 제거"
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
tags: "codex, verdict, convergence, mutant, version-drift, acceptance"
---

# 인수 기준 — SPEC-CODEX-VERDICT-SYNTH-001

## §A 검증 규율 [HARD]

- 모든 AC 는 **Go 테스트 단언**으로 검증한다. grep 형태 AC 는 쓰지 않는다 — 이 작업의 대상은 문자열의 존재가 아니라 **채택된 값** 이기 때문이다.
- 실행: `go test ./internal/cli/... -timeout 600s`. `go test ./...` 금지.
- 각 AC 는 **자기가 죽이는 mutant** 를 함께 적는다. mutant 를 적지 못하는 AC 는 아무것도 관측하지 않는 AC 다.
- 지배 원칙(spec.md §0) **판정 불가를 통과로 읽지 않는다** 를 직접 고정하는 것은 AC-CVS-004 다.

## §B 대표 mutant 2종

### MUTANT-REP-1 — 감지 후 관대 채택

```go
// 통과해서는 안 된다
if stated != inferred {
    note = "signals disagreed"   // 감지함
}
return ReviewOutput{Verdict: "pass", SynthesisNote: note}  // 그러나 관대한 쪽 채택
```

"불일치가 로그에 남는다" / "경고가 나온다" 로만 표현된 AC 는 이 mutant 를 통과시킨다. 그래서 AC-CVS-002 는 **채택된 verdict 값** 을 건다.

### MUTANT-REP-2 — 인식기만 늘리고 fall-through 는 그대로

```go
// 통과해서는 안 된다
verdict := "pass"                       // ← 구조적 원인이 그대로 남아 있다
if m := codexStatedVerdict...  { ... }
if m := codexScoredVerdict...  { ... }  // 새 서식 인식기를 추가함
if codexFindingBullet...       { ... }
```

0.149.0 의 서식 하나를 알아보게 만들었을 뿐, **아는 서식이 하나도 안 맞을 때 여전히 `pass` 로 떨어진다**. 눈앞의 사례는 고쳐지고 다음 버전에서 같은 사고가 난다. AC-CVS-004 가 이것을 잡는다.

## §C 인수 기준

### AC-CVS-001 — 점수 표기 판정 인식 (REQ-CVS-001)

**Given** 리뷰 본문이 줄 머리에 `FAIL 0.75` 를 담고 있고 심각도 태그 불릿은 하나도 없을 때
**When** adversarial 모드로 `synthesizeReviewOutput` 을 호출하면
**Then** `Verdict` 는 `"fail"` 이다.

동일 형태로 `PASS 0.88` → `"pass"`, `INCONCLUSIVE 0.50` → `"inconclusive"` 를 함께 건다.

> **죽이는 mutant**: 점수 표기를 인식하지 않는 구현 (현재 트리 상태 — `codexStatedVerdict` 는 `Verdict:` 라벨을 요구하므로 `FAIL 0.75` 에 매치되지 않는다). 사전구현 트리에서 이 AC 는 **실패해야 한다**.

### AC-CVS-002 — 갈린 신호에서 보수 채택 (REQ-CVS-002) [MUTANT-REP-1 킬러]

**Given** 리뷰 본문이 `Verdict: pass` 를 명시하면서 동시에 `- [P1] SQL injection at db.go:12` 불릿을 담고 있을 때
**When** `synthesizeReviewOutput` 을 호출하면
**Then** 채택된 `Verdict` 는 **`"fail"`** 이고, `SynthesisNote` 는 비어 있지 않다.

**Given** 리뷰 본문이 `Verdict: fail` 을 명시하고 불릿은 없을 때
**When** 호출하면
**Then** 채택된 `Verdict` 는 **`"fail"`** 이다.

> **죽이는 mutant**: MUTANT-REP-1. 단언이 `SynthesisNote != ""` 만이면 통과하므로, **`Verdict` 값 단언이 필수**다.
> **죽이는 mutant (b)**: 순위 테이블이 `pass` 를 `inconclusive` 보다 보수적으로 매기는 구현. `Verdict: inconclusive` + 불릿 없음 → `"inconclusive"` (`"pass"` 아님) 케이스로 잡는다.

### AC-CVS-003 — 점수 표기 오탐 방지 (REQ-CVS-001)

**Given** 리뷰 본문이 산문 안에 `the suite reported PASS 12 times before the regression` 을 담고 있고 판정 라벨도 불릿도 없을 때
**When** adversarial 모드로 호출하면
**Then** `Verdict` 는 `"inconclusive"` 다 (`"pass"` 가 아니다).

> **죽이는 mutant**: 점수 표기 정규식을 `(?i)(pass|fail)\s+[\d.]+` 처럼 느슨하게 잡아 산문을 판정으로 읽는 구현.

### AC-CVS-004 — 서식 미상은 inconclusive (REQ-CVS-003) [MUTANT-REP-2 킬러 · §0 직접 고정]

**Given** 리뷰 본문이 **아는 서식 어느 것에도 맞지 않을** 때 — 판정 라벨(`Verdict:`)도, 점수 표기(`FAIL 0.75`)도, 심각도 태그 불릿(`- [P1]`)도 없는 산문. 예:

```
I walked the diff top to bottom and the caching layer looked reasonable to me.
Nothing further to add at this time.
```

**When** `codexMethodTurnStart`(adversarial) 모드로 호출하면
**Then** `Verdict` 는 **`"inconclusive"`** 다.

이 AC 는 codex CLI 의 특정 버전에 묶이지 않는다. 지금 아는 서식이 셋이든 나중에 다섯이든, **전부 안 맞으면 adversarial 경로의 값은 `inconclusive`** 라는 것이 걸린 내용이다.

> **죽이는 mutant**: MUTANT-REP-2 — 새 서식 인식기(`codexScoredVerdict`)를 추가했으나 `verdict := "pass"` fall-through 를 남긴 구현. AC-CVS-001 만으로는 이 mutant 가 **통과한다**(점수 표기는 인식하므로). fall-through 를 직접 겨냥하는 AC 가 따로 있어야 하는 이유다.
> **죽이는 mutant**: `verdict := "pass"` 기본값을 그대로 유지한 구현 (현재 트리 상태). 사전구현 트리에서 이 AC 는 **실패해야 한다**.

### AC-CVS-005 — native review 무불릿은 pass 로 보존 (REQ-CVS-003) [게이트 보호]

**Given** 리뷰 본문이 `The change introduces no blocking issues.` 이고 불릿도 판정 라벨도 없을 때
**When** `codexMethodReviewStart` 모드로 호출하면
**Then** `Verdict` 는 **`"pass"`** 다.

빈 문자열 입력에 대해서도 review 모드에서는 `"pass"` 를 유지한다 (기존 테이블 그대로).

> **죽이는 mutant (c)**: REQ-CVS-003 을 모드 구분 없이 적용해 review 모드까지 `inconclusive` 로 만든 구현. 이 mutant 는 `HandleCodexReviewGate` 의 clean-pass 를 깨뜨려 정상 변경을 차단하므로, 이 AC 가 없으면 회귀가 조용히 착지한다.

### AC-CVS-006 — 라이브 프로브 본문 고정 (REQ-CVS-001·002·003)

**Given** `.moai/reports/t229/live-probe-body.txt` 의 내용을 **원문 그대로** 픽스처로 읽었을 때
**When** adversarial 모드로 `synthesizeReviewOutput` 을 호출하면
**Then** `Verdict` 는 `"inconclusive"` 이며, **`"pass"` 가 아니다**.

> **죽이는 mutant**: 명시 verdict 파서를 제거하거나 좁혀 이 본문을 다시 `pass` 로 떨어뜨리는 회귀. 이 본문은 실제 사고의 원문이므로 회귀 고정 지점으로 쓴다.
> **참고**: 착수 트리 실측상 이 AC 는 **이미 통과한다**(t178·t186 착지분). 회귀 방지가 목적이며, 새 검출이 목적이 아니다 — spec.md §A.2 참조.

### AC-CVS-007 — 대소문자 무관 판정어 인식 (REQ-CVS-001)

**Given** 리뷰 본문 첫 줄이 `VERDICT: FAIL — merge blocked.` 일 때
**When** 호출하면
**Then** `Verdict` 는 `"fail"` 이다. `verdict: fail` / `**Verdict:** Fail` 형태도 동일하게 `"fail"` 이다.

> **죽이는 mutant (a)**: 판정어 파서를 소문자만 매치하도록 좁힌 구현.

### AC-CVS-008 — 불일치가 수렴 결과로 드러난다 (REQ-CVS-004)

**Given** codex 백엔드가 `SynthesisNote` 가 비어 있지 않은 `PerBackendVerdict` 를 돌려주고 다른 백엔드는 모두 `pass` 일 때
**When** `converge` 를 호출하면
**Then** `DisagreementFlag` 는 `true` 이고, `ResidualRiskNote` 는 그 note 내용을 담는다.

**그리고** 같은 입력에서 `OverallVerdict` 는 기존 정책이 산출하던 값 그대로다 — `disagreement_flag` 는 새 차단 범주가 아니다.

> **죽이는 mutant**: `SynthesisNote` 를 `ReviewOutput` 에만 두고 `PerBackendVerdict` 로 전달하지 않아 `converge` 가 영원히 못 보는 구현 (필드는 추가됐는데 배선이 없는 형태).
> **죽이는 mutant**: 불일치를 새 차단으로 승격시켜 `OverallVerdict` 를 `fail` 로 바꾸는 구현 — `OverallVerdict` 불변 단언이 잡는다.

### AC-CVS-009 — codex_task 출력 불변 (REQ-CVS-005)

**Given** `codex_task` 경로가 판정어도 불릿도 없는 본문을 받았을 때
**When** 태스크가 완료되면
**Then** 반환된 `CodexTaskResult.Output` 은 본문 텍스트 그대로이며, 이 SPEC 이전과 동일하다.

> **죽이는 mutant**: 모드 fall-through 변경이 `codex_task` 의 출력 경로까지 흘러들어 텍스트가 바뀌는 구현.

### AC-CVS-010 — 기존 테스트 보존 (REQ-CVS-005)

**Given** `TestSynthesizeReviewOutput_FindingBulletsMapToFail`(`internal/cli/codex_review_rpc_test.go:114`) 의 4개 입력
**When** review 모드를 명시해 호출하도록 확장된 뒤에도
**Then** 4건 모두 기존 기대값(`fail`/`fail`/`pass`/`pass`)을 유지한다.

> **죽이는 mutant**: 기존 테스트를 삭제하거나 기대값을 새 동작에 맞춰 고쳐 회귀를 은폐하는 구현.

## §D 품질 게이트

| 항목 | 기준 |
|---|---|
| 테스트 | `go test ./internal/cli/... -timeout 600s` 전량 통과 |
| 정적 분석 | `go vet ./internal/cli/...` 무경고 |
| 크로스 플랫폼 | `GOOS=windows go vet ./internal/cli/...` 통과 |
| 커버리지 | 변경된 함수(`synthesizeReviewOutput`, `converge`)에 대해 위 AC 가 모두 실행 경로를 덮음 |
| fail-open | 어떤 AC 도 hard error 반환을 기대하지 않음 |

## §E Definition of Done

- [ ] AC-CVS-001 ~ AC-CVS-010 전부 통과
- [ ] AC-CVS-001·AC-CVS-004 가 **사전구현 트리에서 실패했음** 을 RED 단계 출력으로 기록 (새 검출력이 실재함의 증거)
- [ ] AC-CVS-004 가 AC-CVS-001 과 **독립적으로** 실패했음을 확인 — 둘이 같은 mutant 만 잡으면 MUTANT-REP-2 가 빠져나간다
- [ ] AC-CVS-006 은 사전구현 트리에서도 통과함을 기록 (회귀 고정 목적임을 명시)
- [ ] `TestSynthesizeReviewOutput_FindingBulletsMapToFail` 삭제되지 않음
- [ ] §D 품질 게이트 전량 통과 출력이 progress.md §E.2 에 인용됨
- [ ] 착지 후 리드에게 t234 (= GitHub #1632) 착수 가능 신호 전달 (spec.md §C — 순서 확정)
