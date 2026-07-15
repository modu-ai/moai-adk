---
title: "@MX TAG 시스템"
weight: 61
draft: false
---

@MX TAG는 코드 레벨 주석으로, AI 에이전트가 개발 세션 사이에 **컨텍스트·불변량·위험 구역**을 전달하는 표준 수단입니다. 프롬프트는 무시될 수 있지만 코드에 새겨진 주석은 코드와 함께 살아남아, 다음 에이전트가 코드를 처음 읽는 순간 의도와 제약을 즉시 파악할 수 있습니다.

> @MX TAG의 운영(스캔·추가·질의)은 `/moai mx` 명령어로 수행합니다. 이 페이지는 태그 시스템 자체의 프로토콜과 라이프사이클을 다룹니다.

## 태그 문법

```go
// @MX:TAG_TYPE: [설명]
// @MX:SUB_KEY: [하위 값]
```

태그는 인라인 소스 주석이지 별도의 JSON 원장이 아닙니다. `grep` 또는 `moai mx query`로 수집됩니다.

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

`@MX:DEBT`는 미완성 작업 표시가 아닙니다. 코드는 **이미 완성되어 정확히 동작**하지만, 명시된 한계 내에서의 의도적 단순화임을 기록합니다. 두 하위 라인이 따릅니다.

```go
// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
```

`@MX:UPGRADE`이 없는 DEBT는 종료 조건이 없어 **조용히 부패(rot)** 합니다. `moai mx query --kind DEBT --json`은 이를 `"rotRisk": "no-trigger"`로 표시합니다. 부패 신호는 `@MX:UPGRADE` 부재이며, `@MX:CEILING` 부재는 품질 메모일 뿐 부패의 기준이 아닙니다.

> `@MX:TODO`는 GREEN 단계에서 해결되는 미완성 작업(코드가 아직 완성되지 않음)을, `@MX:DEBT`는 완성되어 정확히 동작하지만 명시적 한계를 가진 단순화(코드는 완성됨)를 표시합니다. DEBT는 여러 GREEN 단계에 걸쳐 정상적으로 유지될 수 있으며 TODO의 "3회 미해결 시 WARN 승격" 규칙이 적용되지 않습니다.

## 업데이트·제거 시점

- **ANCHOR** — fan_in 변화 또는 SPEC 업데이트 시 갱신. 자동 삭제 금지, 리포트로 NOTE 강등.
- **NOTE** — 함수 시그니처 변경 시 재검토.
- **WARN** — 위험 구조 개선 시 제거.
- **TODO** — 해결 시(테스트 통과 또는 구현 완료) 제거. 3회 반복 미해결 시 WARN으로 승격.
- **DEBT** — 한계 또는 트리거 변화 시 갱신. `@MX:UPGRADE` 트리거 발화로 단순화가 교체될 때 제거하며, 다른 작업 완료와 무관합니다. 자동 승격 없음.

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
- **limits** — `anchor_per_file` (기본 3), `warn_per_file` (기본 5). 초과 시 ANCHOR는 최저 fan_in부터 강등, WARN은 P1–P5 우선만 유지.
- **exclude** — `**/*_generated.go`, `**/vendor/**`, `**/mock_*.go` 등 태깅 제외 패턴
- **require_reason_for** — REASON이 필수인 태그 타입

## 태그 언어

태그 설명과 `@MX:REASON`은 `.moai/config/sections/language.yaml`의 `code_comments` 설정을 따릅니다 (기본 `en`). 한국어 프로젝트라면 `code_comments: ko`로 설정하면 태그가 한국어로 작성됩니다.

## 다음 단계

- [Hooks 가이드](/ko/advanced/hooks-guide) — 훅과 함께 코드 컨텍스트를 다루는 기반
- [SPEC 기반 개발](/ko/core-concepts/spec-based-dev) — SPEC 라이프사이클과 @MX TAG 연동
- [TRUST 5 품질 프레임워크](/ko/core-concepts/trust-5) — Readable 원칙과 @MX:NOTE
