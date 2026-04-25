---
spec_id: SPEC-STATUSLINE-002
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. 커밋 beba31b7c 및 후속 커밋에서
  Claude Code v2.1.80 rate_limits statusline 지원이 출시됨.
  spec.md §2-3의 REQ-SL-010/011과 실제 구현(`internal/statusline/types.go`,
  `builder.go:219-220`, `renderer.go:174-188,272,284`)을 대조하여 AC 역도출.
  plan-auditor 2026-04-24 감사 시 acceptance.md 부재 확인.
---

# Acceptance Criteria — SPEC-STATUSLINE-002

Claude Code v2.1.80 statusline JSON의 `rate_limits` 필드를 파싱하여 5시간/7일 rate limit 사용량을 표시하는 기능의 관찰 가능한 인수 기준.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| REQ-SL-010 (rate_limits 필드 파싱) | AC-001, AC-002 | `internal/statusline/types.go:67,71-81,184` |
| REQ-SL-011 (렌더러 우선 사용) | AC-003, AC-004, AC-005 | `internal/statusline/renderer.go:171-188,272,284`, `builder.go:219-220,252-255` |

## AC-001: StdinData와 StatusData가 rate_limits 필드를 정의한다

**Given** statusline 타입 정의가 컴파일된 상태에서,
**When** `StdinData`와 `StatusData` 구조체를 참조하면,
**Then** 두 구조체 모두 `RateLimits *RateLimitInfo` 필드를 가지며 StdinData의 JSON 태그는 `json:"rate_limits"`여야 한다.

**Verification**: `internal/statusline/types.go:67` (`RateLimits *RateLimitInfo \`json:"rate_limits"\``), `:184` (`RateLimits *RateLimitInfo // Rate limit info from Claude Code (nil when unavailable)`).

## AC-002: RateLimitInfo / RateLimitWindow 타입이 정의된다

**Given** rate limit 파싱을 위한 타입 모듈에서,
**When** 타입 정의를 조회하면,
**Then** `RateLimitInfo` 구조체는 `FiveHour *RateLimitWindow`와 `SevenDay *RateLimitWindow` 필드를 가져야 하고, `RateLimitWindow`는 사용률/리셋 시간 필드(`UsedPercentage`, `ResetsAt`)를 가져야 한다.

**Verification**: `internal/statusline/types.go:71-81` — `RateLimitInfo` 및 `RateLimitWindow` 구조체 정의. `renderer.go:175,176,187,188`에서 `UsedPercentage`, `ResetsAt` 사용 확인.

## AC-003: builder가 stdin의 RateLimits를 StatusData로 전달한다

**Given** Claude Code가 statusline JSON에 `rate_limits` 객체를 제공하는 상황에서,
**When** `collectAll(input)`이 실행되면,
**Then** `input.RateLimits`가 nil이 아니면 `data.RateLimits = input.RateLimits`로 복사되어야 하고, 이후 usage provider 호출이 스킵되어야 한다.

**Verification**: `internal/statusline/builder.go:219-220` (`if input != nil && input.RateLimits != nil { data.RateLimits = input.RateLimits }`), `:252-255` (`if b.usageProvider != nil && (input == nil || input.RateLimits == nil)` — 조건부 MoAI API 호출 스킵).

## AC-004: renderer는 RateLimits를 MoAI API Usage보다 우선 사용한다

**Given** `data.RateLimits != nil` 및 `data.RateLimits.FiveHour != nil`인 상황에서,
**When** `renderFullV3(data)`가 실행되면,
**Then** `pct5H = int(data.RateLimits.FiveHour.UsedPercentage)`가 우선 설정되고, RateLimits가 nil일 때만 `data.Usage`로 폴백해야 한다.

**Verification**: `internal/statusline/renderer.go:171` 주석 `// Prefer RateLimits (from Claude Code v2.1.80+ statusline JSON) over Usage (MoAI API call)`, `:174-188` 분기 로직.

## AC-005: 5H 사용률은 상대 시간, 7D 사용률은 절대 시간으로 리셋이 표시된다

**Given** `RateLimits.FiveHour.ResetsAt` 및 `SevenDay.ResetsAt` 둘 다 유효한 타임스탬프인 상황에서,
**When** 렌더링이 완료되면,
**Then** 5H 리셋은 `formatResetTimeRelative(...)` 결과로 (예: `in 2h 34m`) 표시되어야 하고, 7D 리셋은 `formatResetTimeAbsolute(...)` 결과로 (예: `Mon 10:00`) 표시되어야 한다.

**Verification**: `internal/statusline/renderer.go:176` (`reset5H = formatResetTimeRelative(data.RateLimits.FiveHour.ResetsAt)`), `:188` (`reset7D = formatResetTimeAbsolute(data.RateLimits.SevenDay.ResetsAt)`).

## Edge Cases

- **EC-01**: Claude Code < v2.1.80 → `RateLimits == nil` → renderer는 기존 MoAI API Usage 경로로 폴백 (builder.go:252-255 조건부 provider 호출).
- **EC-02**: `RateLimits != nil`이지만 `FiveHour == nil` → 해당 바 세그먼트는 `pct5H = 0`으로 처리되거나 생략(renderer.go:174의 nil 체크).
- **EC-03**: 사용량 수집 중복 방지 — Claude Code가 직접 제공하면 MoAI의 300ms 타임아웃 네트워크 호출 자체를 수행하지 않음(builder.go:255 주석 및 조건).

## Non-functional Evidence

- **기존 UsageProvider 경로 보존**: `builder.go:252-255`의 조건부 호출로 하위 호환성 보장.
- **Zero-downtime 전환**: Claude Code 2.1.80 미만 환경에서는 기존 API 경로가 그대로 동작.

## Definition of Done

- [x] `StdinData.RateLimits` + `StatusData.RateLimits` 필드 추가 (AC-001)
- [x] `RateLimitInfo` / `RateLimitWindow` 타입 정의 (AC-002)
- [x] builder에서 stdin → StatusData 전달 + provider 중복 호출 방지 (AC-003)
- [x] renderer에서 RateLimits 우선 사용 (AC-004)
- [x] 5H 상대 시간 / 7D 절대 시간 포맷팅 (AC-005)
- [x] 커밋 beba31b7c 및 관련 후속 커밋 merged to main
- [x] 상태: spec.md frontmatter `status: completed` 표시됨
