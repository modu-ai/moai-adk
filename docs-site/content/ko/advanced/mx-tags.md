---
title: "@MX TAG 시스템"
weight: 61
draft: false
---

@MX TAG는 코드에 직접 다는 주석으로, AI 에이전트가 개발 세션과 세션 사이에 **컨텍스트·불변량·위험 구역**을 넘겨주는 표준 수단입니다. 프롬프트는 무시될 수 있어도 코드에 새겨둔 주석은 코드와 함께 살아남습니다. 덕분에 다음 에이전트는 코드를 처음 읽는 순간 의도와 제약을 곧바로 파악합니다.

> @MX TAG를 실제로 다루는 일(스캔·추가·질의)은 `/moai mx` 명령어가 맡습니다. 이 페이지는 태그 시스템 자체의 프로토콜과 라이프사이클을 다룹니다.

## 태그 문법

```go
// @MX:TAG_TYPE: [설명]
// @MX:SUB_KEY: [하위 값]
```

태그는 소스에 인라인으로 붙는 주석이지 별도의 JSON 원장이 아닙니다. `grep`이나 `moai mx query`로 모아 봅니다.

## 태그 타입

| 태그 | 용도 | 필수 서브라인 |
|------|------|----------------|
| `@MX:NOTE` | 컨텍스트와 의도 전달 | — |
| `@MX:WARN` | 위험 구역 표시 | `@MX:REASON` |
| `@MX:ANCHOR` | 불변 계약 (높은 fan_in) | `@MX:REASON` |
| `@MX:TODO` | 미완성 작업 | — |
| `@MX:DEBT` | 의도적 단순화 (작동하는 코드) | `@MX:CEILING` + `@MX:UPGRADE` |

## 서브라인

`@MX:SPEC` · `@MX:LEGACY` · `@MX:REASON` · `@MX:TEST` · `@MX:PRIORITY` · `@MX:CEILING` · `@MX:UPGRADE`

- `@MX:REASON`은 WARN·ANCHOR에 **필수**입니다.
- `[AUTO]` 접두어는 에이전트가 생성한 태그에 **필수**입니다.

## 추가 시점

**@MX:NOTE** — 매직 상수, 100줄 초과 exported 함수에 godoc 부재, 설명 없는 비즈니스 규칙.

**@MX:WARN** — `context.Context` 없는 goroutine/channel, 순환 복잡도 15 이상, 전역 상태 변경, if-분기 8개 이상.

**@MX:ANCHOR** — fan_in 3 이상, 공개 API 경계, 외부 시스템 통합 지점.

**@MX:TODO** — 테스트 파일 없는 공개 함수, 미구현 SPEC 요구사항, 처리 없이 반환되는 에러.

**@MX:DEBT** — 의도적 단순화를 채택했고, 명시된 한계(`@MX:CEILING`) 내에서 정확하며, 재방문 트리거(`@MX:UPGRADE`)가 있을 때.

## DEBT — 작동하는 단순화의 명시적 한계

`@MX:DEBT`는 미완성 작업 표시가 아닙니다. 코드는 **이미 완성돼 정확히 동작**하며, 다만 명시한 한계 안에서 일부러 단순하게 짰다는 사실을 기록합니다. 하위 라인 두 개가 따라붙습니다.

```go
// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
```

`@MX:UPGRADE`가 없는 DEBT는 끝날 조건이 없어 **조용히 부패(rot)** 합니다. `moai mx query --kind DEBT --json`은 이런 항목을 `"rotRisk": "no-trigger"`로 표시합니다. 부패의 신호는 `@MX:UPGRADE`가 없다는 점이며, `@MX:CEILING`이 없는 것은 품질 메모일 뿐 부패 판정 기준이 아닙니다.

> `@MX:TODO`는 GREEN 단계에서 마무리할 미완성 작업(코드가 아직 완성되지 않음)을, `@MX:DEBT`는 완성돼 정확히 동작하지만 한계를 명시해 둔 단순화(코드는 완성됨)를 가리킵니다. DEBT는 여러 GREEN 단계를 넘어 그대로 남아 있어도 정상이며, TODO의 "3회 미해결 시 WARN 승격" 규칙도 적용되지 않습니다.

## 업데이트·제거 시점

- **ANCHOR** — fan_in이 바뀌거나 SPEC이 갱신되면 함께 갱신. 자동 삭제는 금지하고, 리포트를 거쳐 NOTE로 강등.
- **NOTE** — 함수 시그니처가 바뀌면 다시 검토.
- **WARN** — 위험한 구조를 개선했으면 제거.
- **TODO** — 해결되면(테스트 통과 또는 구현 완료) 제거. 세 번 반복해도 남아 있으면 WARN으로 승격.
- **DEBT** — 한계나 트리거가 바뀌면 갱신. `@MX:UPGRADE` 트리거가 발화해 단순화를 교체할 때 제거하며, 다른 작업이 끝났는지와는 무관합니다. 자동 승격은 없습니다.

## 라이프사이클 요약

```text
TODO     RED/ANALYZE 생성 → GREEN/IMPROVE 해결(제거) → 3회 미해결 시 WARN 승격
ANCHOR   fan_in ≥ 3 생성 → 호출 수·SPEC 변화 시 갱신 → fan_in < 3 시 NOTE 강등(리포트) → 자동 삭제 없음
WARN     위험 감지 시 생성 → 구조적이면 지속 → 해결 시 제거
NOTE     컨텍스트 필요 시 생성 → 시그니처 변경 후 갱신 → 코드 삭제 시 폐기
DEBT     의도적 단순화 시 생성 → UPGRADE 트리거 발화 시 해결(단순화 교체) → 자동 승격 없음
```

## 언어별 주석 문법

| 언어 | 접두어 | 예시 |
|------|--------|------|
| Go · Java · TS · Rust · C/C++ · Swift · Kotlin · Dart · Zig · Scala | `//` | `// @MX:NOTE:` |
| Python · Ruby · Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

## 설정 (`.moai/config/sections/mx.yaml`)

- **thresholds** — `fan_in_anchor`, `complexity_warn`, `branch_warn`
- **limits** — `anchor_per_file`(기본 3), `warn_per_file`(기본 5). 한도를 넘으면 ANCHOR는 fan_in이 낮은 것부터 강등하고, WARN은 P1–P5 우선순위만 남깁니다.
- **exclude** — `**/*_generated.go`, `**/vendor/**`, `**/mock_*.go` 등 태깅에서 빼는 패턴
- **require_reason_for** — REASON을 반드시 달아야 하는 태그 타입

## 태그 언어

태그 설명과 `@MX:REASON`은 `.moai/config/sections/language.yaml`의 `code_comments` 설정을 따릅니다(기본 `en`). 한국어 프로젝트라면 `code_comments: ko`로 설정해 태그를 한국어로 작성할 수 있습니다.

## 다음 단계

- [Hooks 가이드](/ko/advanced/hooks-guide) — 훅과 함께 코드 컨텍스트를 다루는 기반
- [SPEC 기반 개발](/ko/core-concepts/spec-based-dev) — SPEC 라이프사이클과 @MX TAG 연동
- [TRUST 5 품질 프레임워크](/ko/core-concepts/trust-5) — Readable 원칙과 @MX:NOTE
