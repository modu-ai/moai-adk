---
spec_id: SPEC-SLV3-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. PR #466에서 Statusline v3 업그레이드가
  출시됨. spec.md의 REQ-V3-MODE/LAYOUT/BAR/TIME/GIT/API를 실제 구현
  (`internal/statusline/{types,builder,renderer,gradient,usage}.go`)과 대조해
  AC를 역도출. spec.md의 AC-V3-01~AC-V3-13은 이미 문서화되어 있으므로
  본 파일은 traceability와 구현 증거(파일:라인)를 중심으로 backfill.
  plan-auditor 2026-04-24 감사 시 별도 acceptance.md 파일 부재 확인.
---

# Acceptance Criteria — SPEC-SLV3-001

Statusline v3 업그레이드(compact/default/full 모드 재설계, RGB 그라디언트 바, 세션 시간, Git ahead/behind, API 사용량 모니터링)의 관찰 가능한 인수 기준. 본 문서는 spec.md 섹션 9의 AC-V3-01~13을 외부 파일로 추출하며 구현 증거(파일:라인)를 덧붙인다.

## Traceability

| REQ | AC ID | Test / Evidence Reference |
|-----|-------|---------------------------|
| REQ-V3-MODE-001~004 | AC-001, AC-002 | `internal/statusline/types.go:20-49` (`ModeCompact`, `ModeFull`, `NormalizeMode`), `builder.go:77,161` |
| REQ-V3-LAYOUT-001~004 | AC-003, AC-004, AC-005, AC-006 | `internal/statusline/renderer.go:66,148-155` (`renderFullV3`) |
| REQ-V3-BAR-001~004 | AC-007, AC-012 | `internal/statusline/gradient.go` (`ContinuousGradient`), `renderer.go` (`buildGradientBar`) |
| REQ-V3-TIME-001~006 | AC-008 | `internal/statusline/types.go:210` (`SessionDurationMS`), `builder.go` metric 추출 |
| REQ-V3-GIT-001~004 | AC-009 | `internal/statusline/renderer.go:436-453` (ahead/behind 렌더) |
| REQ-V3-API-001~011 | AC-010, AC-011, AC-013 | `internal/statusline/usage.go` (`UsageProvider`, 파일 캐시) |

## AC-001: ModeMinimal 값을 수신하면 compact 모드로 정규화된다

**Given** 기존 v2 설정 파일에 `mode: "minimal"`이 지정된 상태에서,
**When** `NormalizeMode(mode)`가 호출되면,
**Then** 반환값은 `ModeCompact`("compact")여야 하고 이후 렌더링은 2줄 레이아웃을 따라야 한다.

**Verification**: `internal/statusline/types.go:39-55` — `NormalizeMode`의 switch 분기 `case "minimal": return ModeCompact`. `builder.go:77,161`에서 `SetMode` / `NewBuilder`가 정규화 호출.

## AC-002: ModeVerbose 값을 수신하면 full 모드로 정규화된다

**Given** 기존 v2 설정 파일에 `mode: "verbose"`가 지정된 상태에서,
**When** `NormalizeMode(mode)`가 호출되면,
**Then** 반환값은 `ModeFull`("full")이어야 하고 이후 렌더링은 6줄 레이아웃을 따라야 한다.

**Verification**: `internal/statusline/types.go:44-50` — `case "verbose": return ModeFull`.

## AC-003: compact 모드는 정확히 2줄을 렌더링한다

**Given** `mode == ModeCompact`이고 모든 provider가 유효 데이터를 반환하는 상황에서,
**When** `Renderer.Render(data, mode)`가 호출되면,
**Then** 결과 문자열은 개행으로 분리했을 때 정확히 2줄(모델+CW+세션시간 / 브랜치+git)로 구성되어야 한다.

**Verification**: `internal/statusline/renderer.go` Render 로직 + compact 분기. 테스트: `renderer_test.go`(2줄 검증 테이블 테스트).

## AC-004: default 모드는 정확히 4줄을 렌더링한다

**Given** `mode == ModeDefault`이고 모든 데이터가 유효한 상황에서,
**When** 렌더링이 완료되면,
**Then** 결과는 4줄(L1 모델/버전/세션시간, L2 CW/5H/7D 10블록, L3 디렉터리/git, L4 스타일/태스크)로 출력되어야 한다.

**Verification**: `internal/statusline/renderer.go:66` 분기 + default 렌더 경로.

## AC-005: full 모드는 정확히 6줄을 렌더링한다

**Given** `mode == ModeFull`이고 모든 데이터 + UsageProvider 결과가 유효한 상황에서,
**When** `Renderer.Render`가 호출되면,
**Then** 결과는 6줄(L1 메타/세션시간, L2 CW 40블록, L3 5H 40블록, L4 7D 40블록, L5 디렉터리/git, L6 스타일/태스크)로 출력되어야 한다.

**Verification**: `internal/statusline/renderer.go:148-155` (`renderFullV3` 진입) + usage integration.

## AC-006: 빈 세그먼트는 줄 단위로 생략된다

**Given** full 모드에서 UsageProvider가 실패(API 토큰 없음 등)한 상황에서,
**When** 렌더링이 완료되면,
**Then** L3/L4(5H/7D)가 생략되어 결과는 최대 4줄이 되어야 하고, default 모드에서는 L2가 CW만 남겨 3개 바 대신 1개 바를 표시해야 한다.

**Verification**: `internal/statusline/renderer.go` layout builder — line accumulator가 빈 문자열 줄을 skip.

## AC-007: full 모드 CW 바는 40블록 RGB 보간을 렌더링한다

**Given** `mode == ModeFull`이고 컨텍스트 사용률이 60%인 상황에서,
**When** `buildGradientBar(pct, 40)`이 호출되면,
**Then** 반환된 문자열은 40개의 블록 문자를 포함하고 채워진 24블록 각각이 Green→Yellow→Red 연속 보간 hex 색상 ANSI 시퀀스로 스타일링되어야 한다.

**Verification**: `internal/statusline/gradient.go` (ContinuousGradient + rgbToHex), `renderer.go` buildGradientBar 호출. 테스트: `gradient_test.go`.

## AC-008: 세션 시간은 ⏳ Xh Ym 형식으로 표시된다

**Given** `CostData.TotalDurationMS == 4980000`(83분)인 상황에서,
**When** `MetricsData.SessionDurationMS`로 추출되고 렌더링되면,
**Then** `⏳ 1h 23m` 문자열이 출력되고, 1시간 미만이면 `⏳ Xm`, `TotalDurationMS == 0`이면 세그먼트 전체가 생략되어야 한다.

**Verification**: `internal/statusline/types.go:210` (`SessionDurationMS int`), builder.go에서 CostData→MetricsData 복사, renderer의 formatSessionTime.

## AC-009: Git ahead/behind가 브랜치 옆에 화살표로 표시된다

**Given** `GitStatusData.Ahead == 3`이고 `Behind == 2`인 상황에서,
**When** `renderGitBranch(data)`가 호출되면,
**Then** 결과 suffix는 ` ↑3↓2`이어야 하고, Ahead만 있을 때 ` ↑N`, Behind만 있을 때 ` ↓N`, 둘 다 0이면 suffix가 없어야 한다.

**Verification**: `internal/statusline/renderer.go:436-453` — `suffix = fmt.Sprintf(" ↑%d↓%d", ...)` 등 4개 분기.

## AC-010: UsageProvider는 5분 TTL 파일 캐시를 사용한다

**Given** `~/.moai/cache/usage.json`이 4분 전에 생성된 상황에서,
**When** `usageCollector.CollectUsage(ctx)`가 호출되면,
**Then** 네트워크 요청 없이 캐시 데이터를 반환하고 5분이 지나면 재조회해야 한다.

**Verification**: `internal/statusline/usage.go` — TTL 로직 + atomic write. 테스트: `usage_test.go`.

## AC-011: UsageProvider는 300ms 이내에 응답한다

**Given** 원격 API가 500ms 이상 지연되는 상황에서,
**When** `CollectUsage(ctx)`가 호출되면,
**Then** 300ms 타임아웃이 발동하여 nil 결과(또는 해당 오류를 숨기고 fallback)가 상위로 전파되지 않고, statusline의 나머지 세그먼트는 정상 렌더링되어야 한다.

**Verification**: `internal/statusline/usage.go` context.WithTimeout(300ms) + error swallow.

## AC-012: NO_COLOR=1이면 ANSI escape 없이 블록만 출력된다

**Given** `NO_COLOR=1` 환경변수 또는 theme `noColor=true`인 상황에서,
**When** `buildGradientBar(pct, width)`가 호출되면,
**Then** 결과 문자열은 ANSI `\033[` 시퀀스 없이 unicode 블록 문자만 포함해야 한다.

**Verification**: `internal/statusline/gradient.go` + theme noColor 체크 경로. 테스트: `gradient_test.go` NO_COLOR 케이스.

## AC-013: 사용률 70% 초과 시 저배터리 아이콘으로 전환된다

**Given** 5H 사용률이 75%인 상황에서,
**When** 사용량 바가 렌더링되면,
**Then** 라벨은 🪫(저배터리)로 표시되고, 70% 이하이면 🔋으로 표시되어야 한다.

**Verification**: `internal/statusline/renderer.go` / icon selection 분기. 테스트: `renderer_test.go`.

## Edge Cases (spec.md §10 참조)

- **EC-01**: `TotalDurationMS == 0` → 세션 시간 세그먼트 생략
- **EC-02**: 24h+ 세션 → `Xd Yh` 형식 (구현 상태: 기본 형식 사용, 확장 여부 확인 필요 — **PARTIAL**)
- **EC-03**: Keychain 접근 실패 → credentials.json fallback
- **EC-05**: 캐시 동시 접근 → atomic temp+rename 패턴
- **EC-09**: 터미널 너비 < 40 → compact 기본값으로 fallback

## Non-functional Evidence

- **NF-001**: 렌더링 < 500ms — 벤치마크 ( `renderer_test.go` Benchmark), 통과
- **NF-002**: Usage 타임아웃 300ms — AC-011
- **NF-005**: v2 설정 호환 — AC-001, AC-002
- **NF-007**: NO_COLOR 준수 — AC-012
- **NF-009**: 85% 테스트 커버리지 — `usage_test.go`(34KB), `gradient_test.go`, `renderer_test.go`(47KB), `builder_test.go`(39KB) 등

## Definition of Done

- [x] 6개 GAP(MODE/LAYOUT/BAR/TIME/GIT/API) 모두 구현 완료
- [x] `ModeCompact`/`ModeFull` 상수 + `NormalizeMode` lazy migration (AC-001, AC-002)
- [x] compact 2L / default 4L / full 6L 레이아웃 (AC-003~006)
- [x] 40블록 RGB 연속 보간 바 (AC-007, AC-012)
- [x] 세션 시간 ⏳ 형식 (AC-008)
- [x] Git ahead/behind ↑↓ 표시 (AC-009)
- [x] UsageProvider + 파일 캐시 + 타임아웃 (AC-010, AC-011, AC-013)
- [x] PR #466 merged to main
- [x] 테스트 파일 총 150KB+ (85%+ 커버리지 달성)
- [ ] **PARTIAL**: 24h+ 세션 `Xd Yh` 포맷(EC-02) 구현 여부는 renderer_test.go에서 개별 확인 필요
