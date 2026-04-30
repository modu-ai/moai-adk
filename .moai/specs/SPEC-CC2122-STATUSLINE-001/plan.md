# SPEC-CC2122-STATUSLINE-001 구현 계획

## 개요

본 계획은 Claude Code v2.1.122 stdin JSON의 `effort.level` 과 `thinking.enabled` 필드를 `internal/statusline/` 패키지에 통합하는 작업의 마일스톤, 기술적 접근, 리스크, 의존성을 정의한다.

이 문서는 시간 추정(예: "2 days") 대신 **우선순위 라벨** 과 **단계 간 순서** 로 진행을 표현한다 (CLAUDE.local.md §14 준수).

## 기술적 접근 (High-Level)

기존 `RateLimitInfo` 패턴(types.go lines 72-83)을 그대로 차용하여 `*EffortInfo`, `*ThinkingInfo` nested pointer struct를 추가한다. 부재 시 `nil` → renderer가 silent omit 분기로 처리. 신규 enum 값에 대한 그레이스풀 폴백은 raw 값 패스스루 방식으로 구현한다 (whitelist 검증을 강제하지 않음).

`builder.go` 는 stdin → StatusData 변환 시 nil 검사 후 매핑하고, `renderer.go` 는 per-segment helper 함수 패턴을 따라 `renderEffortThinking()` (또는 동등한 이름) 헬퍼를 추가한다. 정확한 함수명/위치는 구현 단계에서 결정한다.

## Milestones

### M1 — Path Discovery (Priority: High)

**목표**: 변경 대상 파일 경로와 기존 테스트 패턴을 정확히 파악한다.

**작업 항목:**

- `internal/statusline/types.go` 의 `StdinData`, `StatusData`, `RateLimitInfo` 정의 위치 재확인
- `internal/statusline/builder.go` 의 stdin 파싱 로직 흐름 매핑 (어디서 RateLimits 가 매핑되는지)
- `internal/statusline/renderer.go` 의 per-segment helper 호출 순서 매핑 (default vs full 모드 분기)
- 기존 `*_test.go` 의 테이블 드리븐 테스트 컨벤션 (이름, struct, t.Run 패턴) 추출

**Done 기준:**

- 수정 대상 라인 범위가 PR 설명에 인용 가능한 수준으로 식별됨
- 기존 RateLimitInfo / GitWorktree 패턴이 신규 필드에 그대로 적용 가능함을 확인

### M2 — Type Extension (Priority: High)

**목표**: `types.go` 에 `EffortInfo`, `ThinkingInfo` 신규 struct 추가하고 `StdinData`, `StatusData` 에 필드를 노출한다.

**작업 항목:**

- `EffortInfo` struct: `Level string \`json:"level"\``
- `ThinkingInfo` struct: `Enabled bool \`json:"enabled"\``
- `StdinData` 에 `Effort *EffortInfo \`json:"effort"\`` 및 `Thinking *ThinkingInfo \`json:"thinking"\`` 추가
- `StatusData` 에 동일한 필드를 (또는 평탄화된 형태로) 추가하여 renderer 가 소비 가능하게 함
- godoc 주석은 영어 (코드 식별자 영어 정책 준수). REQ-006 의 한국어 @MX:NOTE 태그는 추후 부착.

**Done 기준:**

- `go build ./internal/statusline/...` 통과
- `go vet ./internal/statusline/...` 무경고

### M3 — Builder Integration (Priority: High)

**목표**: `builder.go` 가 stdin JSON 의 effort/thinking 을 StatusData 로 매핑하도록 한다.

**작업 항목:**

- 기존 RateLimits 매핑 분기 옆에 effort/thinking 매핑 추가
- nil-safe 매핑: `stdin.Effort == nil` 이면 `StatusData.Effort = nil` 그대로 전파
- 빈 문자열 effort.Level 처리 정책 결정: 빈 문자열 → 세그먼트 생략 (REQ-003과 일치)

**Done 기준:**

- builder_test.go 에 effort/thinking 케이스 최소 4개 (present, absent, null, empty-level) 추가하여 통과
- 기존 RateLimits / Workspace / Cost 매핑 테스트 회귀 없음

### M4 — Renderer Indicator (Priority: High)

**목표**: `renderer.go` 에 effort/thinking 세그먼트 렌더링 helper 추가. silent omit 동작 보장.

**작업 항목:**

- `renderEffortThinking(s *StatusData) string` (또는 동등한 이름) helper 추가
- 분기 로직:
  - `s.Effort == nil && s.Thinking == nil` → 빈 문자열 반환 (omit)
  - `s.Effort != nil && s.Effort.Level != ""` → `e:<level>` 추가
  - `s.Thinking != nil && s.Thinking.Enabled` → `·t` 추가
  - effort 부재 + thinking 활성 → `·t` 만 (또는 `t` prefix 결정은 구현 단계)
- raw fallback: invalid level 도 그대로 `e:<level>` 출력 (whitelist 미적용)
- default StatuslineMode 와 full StatuslineMode 양쪽에 통합 (위치/우선순위는 기존 세그먼트 컨벤션 따름)

**Done 기준:**

- 기존 statusline 출력 길이 회귀 없음 (default 모드 3-line layout 유지)
- 신규 helper 가 nil 입력에 대해 panic 없이 빈 문자열 반환

### M5 — Test Coverage (Priority: High)

**목표**: 테이블 드리븐 테스트로 acceptance.md 의 GWT-1 ~ GWT-10 시나리오를 커버한다.

**작업 항목:**

- `renderer_test.go` 또는 신규 `effort_thinking_test.go` 에 6+ 케이스 테이블:
  1. effort=high + thinking=true → `e:high·t` 포함
  2. effort=medium + thinking=false → `e:medium` 포함, 점 없음
  3. effort=nil + thinking=nil → 출력에 `e:` prefix 없음
  4. effort 있음, thinking 만 nil → effort 만 표시, panic 없음
  5. effort=nil, thinking=true → thinking 인디케이터만, panic 없음
  6. effort.Level="bogus" + thinking=false → `e:bogus` 그대로
- builder_test.go 에 stdin JSON 언마샬 케이스 추가:
  7. v2.1.121 이하 stdin (effort/thinking 키 부재) → Go struct 의 두 필드 모두 nil, JSON 에러 없음
- statusline mode 회귀 테스트:
  8. default mode 에서 effort/thinking 추가 후 기존 세그먼트 모두 보존
  9. full mode 에서 effort/thinking 가시성 확인
- coverage: `go test -cover ./internal/statusline/...` 로 신규 코드 경로 100% 확인

**Done 기준:**

- `go test ./internal/statusline/...` 전체 통과
- 신규 추가된 코드 경로에 대한 라인 커버리지 100%
- 기존 builder_test.go / renderer_test.go 회귀 없음

### M6 — MX Tag Annotation + Final Validation (Priority: Medium)

**목표**: 신규 exported 타입에 `@MX:NOTE` (REQ-006 한국어 정책) 부착, 최종 품질 게이트 통과, 커밋 준비.

**작업 항목:**

- `EffortInfo` 와 `ThinkingInfo` struct 위에 `@MX:NOTE` 태그 부착 (한국어, code_comments=ko 준수)
- `go test ./...` (전체) 통과 확인 — cascading failure 검출 (CLAUDE.local.md §6 Go Test Execution Rules)
- `go vet ./...` 통과
- `golangci-lint run` 무경고 (가능한 경우)
- 커밋 메시지 초안: `feat(statusline): integrate Claude Code v2.1.122 effort + thinking fields (SPEC-CC2122-STATUSLINE-001)`

**Done 기준:**

- 모든 게이트 (vet/test/lint) 통과
- @MX:NOTE 태그가 두 신규 타입에 부착됨
- PR 설명에 SPEC-CC2122-STATUSLINE-001 명시

## Dependencies

- **상위 의존**: SPEC-CC2122-HOOK-001 (이미 main 적용 완료) — 동일 v2.1.122 릴리스 통합 시리즈의 일부. 직접적 코드 의존성 없음.
- **하위 의존 후보**: 향후 statusline 시각화 옵션(컬러, 단축형) SPEC 의 입력이 될 수 있음.
- **외부 의존**: Claude Code v2.1.122 이상에서 stdin JSON 에 effort/thinking 필드가 실제로 전송되는지 — 릴리스 노트로 확인됨, 통합 테스트는 구현 단계에서 수행.

## Risks and Mitigations

| 리스크 | 영향도 | 완화 전략 |
|--------|--------|-----------|
| Claude Code 가 미래 버전에서 enum 값을 추가 (예: `"ultra"`) | Medium | REQ-004 의 raw fallback 패스스루로 대응. whitelist 검증을 일부러 추가하지 않음. |
| effort/thinking 세그먼트 추가로 default 모드 line width 초과 | Low | M4 에서 default 모드 길이 회귀 테스트 추가. 인디케이터를 컴팩트(`e:high·t`)하게 유지. |
| nil pointer dereference (renderer 가 builder 에서 nil 매핑 누락) | High | builder/renderer 양쪽에 nil 검사 명시. 테스트 케이스 GWT-4, GWT-5 로 검증. |
| 기존 statusline 사용자(이전 Claude Code 버전)의 출력이 갑자기 깨짐 | High | REQ-003 silent omit 강제. M3/M4 nil 매핑 + M5 의 케이스 7 (v2.1.121 이하 stdin) 통과 필수. |
| @MX:NOTE 한국어 정책 누락 (영어로 작성) | Low | M6 에서 code_comments 설정 재확인 후 부착. |

## Out of Scope (Plan-Level)

- 신규 statusline 모드 추가 (예: `effort-only` mode) — 별도 SPEC 분리.
- effort/thinking 변경 시 hook 트리거 — `internal/hook/` 에 영향 없음.
- 컬러/스타일 customization — 별도 SPEC.
- Windows-specific terminal rendering 검증 — 기존 statusline 의 cross-platform 동작에 의존.

## Approval / Next Step

본 plan.md 와 spec.md, acceptance.md 가 plan-auditor 의 승인을 받으면, manager-cycle 또는 manager-tdd 가 `/moai run SPEC-CC2122-STATUSLINE-001` 으로 위임받아 M1 → M6 순으로 구현한다.
