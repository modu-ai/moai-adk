# SPEC-CC2122-STATUSLINE-001 인수 기준

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-30 | manager-spec | 초기 작성 — GWT-1 ~ GWT-10 시나리오 + Quality Gate + Definition of Done |
| 0.1.1 | 2026-04-30 | manager-spec | GWT-11 added per audit review-1 D1 fix — REQ-006 (한국어 @MX:NOTE) 검증 시나리오 보강 |

본 문서는 SPEC-CC2122-STATUSLINE-001 의 구현이 완료되었음을 검증하기 위한 Given-When-Then(GWT) 시나리오와 Definition of Done 을 정의한다. 각 시나리오는 `internal/statusline/` 패키지의 테이블 드리븐 테스트로 자동화된다.

## Given-When-Then 시나리오

### GWT-1: 기본 케이스 — effort + thinking 모두 활성

**Given:** stdin JSON 페이로드가 `"effort": {"level": "high"}` 와 `"thinking": {"enabled": true}` 를 포함한다.
**When:** statusline renderer 가 default StatuslineMode 로 출력을 생성한다.
**Then:** 렌더 출력 문자열에 `e:high·t` (또는 동등한 컴팩트 표기)가 포함되어야 하며, 다른 세그먼트(model, workspace 등)는 변경되지 않는다.

### GWT-2: thinking 비활성 — 점 인디케이터 부재

**Given:** stdin JSON 페이로드가 `"effort": {"level": "medium"}` 와 `"thinking": {"enabled": false}` 를 포함한다.
**When:** renderer 가 출력을 생성한다.
**Then:** 출력에 `e:medium` 이 포함되어야 하며, thinking 인디케이터(`·t`)는 포함되지 않아야 한다.

### GWT-3: 두 필드 모두 부재 — 하위 호환 silent omit

**Given:** stdin JSON 페이로드에 `effort` 와 `thinking` 키가 모두 부재한다 (Claude Code v2.1.121 이하의 stdin 형식).
**When:** renderer 가 출력을 생성한다.
**Then:** 출력 문자열 어디에도 `e:` prefix 또는 thinking 인디케이터가 등장하지 않아야 한다. 다른 모든 기존 세그먼트는 정상 출력되어야 한다.

### GWT-4: effort 만 존재, thinking 은 null

**Given:** stdin JSON 페이로드에 `"effort": {"level": "low"}` 가 있고 `"thinking": null` 인 경우.
**When:** renderer 가 출력을 생성한다.
**Then:** 출력에 `e:low` 만 포함되고 thinking 인디케이터는 부재한다. 어떤 panic 또는 nil pointer dereference 도 발생하지 않는다.

### GWT-5: thinking 만 활성, effort 는 null

**Given:** stdin JSON 페이로드에 `"effort": null` 이고 `"thinking": {"enabled": true}` 인 경우.
**When:** renderer 가 출력을 생성한다.
**Then:** 출력에 thinking 인디케이터(예: `·t` 또는 단독 형식)가 표시되고 `e:` prefix 는 부재한다. panic 발생 없음.

### GWT-6: 알려지지 않은 effort.level — raw fallback

**Given:** stdin JSON 페이로드에 `"effort": {"level": "bogus"}` (enum `low`/`medium`/`high`/`xhigh`/`max` 외의 값)가 포함된다.
**When:** renderer 가 출력을 생성한다.
**Then:** 출력에 `e:bogus` 가 그대로 포함되어야 한다 (whitelist 검증 없이 raw 값 패스스루). panic 또는 에러 발생 없음.

### GWT-7: 이전 Claude Code stdin — Go 언마샬 호환

**Given:** Claude Code v2.1.121 이하 형식의 stdin JSON (effort/thinking 키 자체가 부재).
**When:** `json.Unmarshal` 로 `StdinData` 구조체에 파싱한다.
**Then:** 파싱 결과 Go 구조체의 `Effort` 와 `Thinking` 필드는 모두 `nil` 이며, `json.Unmarshal` 은 에러를 반환하지 않는다.

### GWT-8: default StatuslineMode 회귀 검증

**Given:** stdin JSON 에 effort=high, thinking.enabled=true 가 포함되고 다른 모든 기존 필드(model, workspace, cost, context_window, rate_limits, output_style 등)도 정상 제공된다.
**When:** renderer 가 default StatuslineMode 로 출력을 생성한다.
**Then:** 기존 세그먼트(model 표시, git branch, context window 사용량 등)가 모두 정상 출력되며, 신규 effort/thinking 세그먼트가 추가된다. 3-line layout 의 line break 위치 및 segment 순서 컨벤션이 깨지지 않는다.

### GWT-9: full StatuslineMode 통합 검증

**Given:** stdin JSON 에 effort + thinking 이 활성된 상태이고 statusline 이 full StatuslineMode 로 동작한다.
**When:** renderer 가 출력을 생성한다.
**Then:** detailed layout 안에 effort/thinking 세그먼트가 노출되어야 하며, full 모드의 다른 추가 정보(rate limits 상세, cost breakdown 등)와 공존해야 한다.

### GWT-10: 테스트 스위트 + 커버리지 게이트

**Given:** 본 SPEC 의 모든 신규 코드 경로(types.go, builder.go, renderer.go 변경분)에 대해 테이블 드리븐 테스트가 작성됨.
**When:** `go test -cover ./internal/statusline/...` 명령을 실행한다.
**Then:**
- 모든 테스트가 통과한다 (`PASS`).
- 신규 추가된 코드 경로의 라인 커버리지가 100% 이다.
- 기존 builder_test.go / renderer_test.go 의 모든 케이스가 회귀 없이 통과한다.
- `go test -race ./internal/statusline/...` 도 통과한다 (race detector clean).

### GWT-11: REQ-006 검증 — 한국어 @MX:NOTE 태그 부착

**Given:** `.moai/config/sections/language.yaml` 의 `code_comments` 설정값이 `"ko"` 이고, 구현 단계의 M6(MX 태그 부착 마일스톤)에서 manager-cycle (또는 manager-tdd) 가 `internal/statusline/types.go` 의 신규 exported 타입 `EffortInfo` 와 `ThinkingInfo` 위에 `@MX:NOTE` 태그를 추가했다.
**When:** 변경된 Go 파일들에 대해 다음 정규식 grep 을 수행한다 (예: `grep -nP '@MX:NOTE: [\p{Hangul}]' internal/statusline/types.go`), 그리고 자동화된 테스트 `TestStatuslineEffortThinking_KoreanMXTags` (또는 동등한 이름의 테스트)가 동일한 정규식 매칭을 수행한다.
**Then:**
- grep 명령이 `EffortInfo` 와 `ThinkingInfo` 두 타입 각각에 대해 최소 1회 이상의 매칭을 반환한다 (즉, 두 타입 모두 한국어로 시작하는 `@MX:NOTE` 설명 텍스트를 보유함).
- 자동화된 테스트 `TestStatuslineEffortThinking_KoreanMXTags` 가 PASS 한다.
- 코드 식별자(타입명 `EffortInfo`/`ThinkingInfo`, 필드명 `Level`/`Enabled`, JSON 태그)는 영어로 유지된다 — 한국어는 `@MX:NOTE` 의 설명 텍스트(콜론 뒤 본문)에만 적용된다.
- 만약 `code_comments` 설정값이 `"ko"` 가 아닌 다른 값(예: `"en"`)일 경우 본 시나리오는 적용되지 않으며, 테스트는 설정값에 따라 적절히 skip 되거나 분기되어야 한다.

## Edge Cases (테스트 추가 권장)

다음은 GWT-1~11 외에 구현 단계에서 명시적으로 다뤄야 할 추가 엣지 케이스:

- **빈 문자열 effort.Level:** `"effort": {"level": ""}` → REQ-003 정책에 따라 세그먼트 생략 (level=빈 문자열을 부재와 동일 취급).
- **JSON 형식 오류:** stdin JSON 자체가 깨진 경우 — 기존 statusline 의 에러 처리 정책을 따르며, 신규 필드가 영향을 주지 않아야 한다.
- **모든 enum 값 정상 출력:** `low`, `medium`, `high`, `xhigh`, `max` 다섯 값을 모두 테스트 케이스로 포함하여 raw fallback 분기에 의존하지 않게 한다.

## Quality Gate Criteria

본 SPEC 의 인수 조건은 다음을 모두 충족해야 한다:

- [ ] **EARS 준수**: spec.md 의 6개 REQ 가 모두 EARS 패턴(WHEN/IF/WHILE/WHERE) 으로 표현됨
- [ ] **GWT-1 ~ GWT-11**: 11개 시나리오 모두 자동화된 테스트로 통과
- [ ] **하위 호환**: GWT-3, GWT-7 통과로 v2.1.121 이하 사용자에 대한 무회귀 보장
- [ ] **그레이스풀 폴백**: GWT-6 통과로 미래 enum 추가에 대한 안전성 확보
- [ ] **회귀 없음**: `go test ./...` 전체 통과 (CLAUDE.local.md §6 Go Test Execution Rules)
- [ ] **race-free**: `go test -race ./internal/statusline/...` 통과
- [ ] **vet/lint**: `go vet ./...` 무경고, `golangci-lint run` 무경고
- [ ] **커버리지**: 신규 코드 경로 100% 라인 커버리지
- [ ] **REQ-006 준수**: 신규 exported 타입에 한국어 `@MX:NOTE` 태그 부착 (code_comments=ko 정책)

## Definition of Done

다음 조건이 모두 만족되면 SPEC-CC2122-STATUSLINE-001 은 완료(`status: completed`)로 간주한다:

1. `internal/statusline/types.go` 에 `EffortInfo`, `ThinkingInfo` struct 가 추가되고 `StdinData`/`StatusData` 가 확장됨.
2. `internal/statusline/builder.go` 가 effort/thinking 매핑을 nil-safe 하게 수행함.
3. `internal/statusline/renderer.go` 에 effort/thinking 세그먼트 helper 가 추가되고 silent omit / raw fallback 분기가 구현됨.
4. 위 GWT-1 ~ GWT-11 시나리오가 모두 자동화된 테스트로 작성·통과됨.
5. `go test ./...` 전체 통과, `go vet ./...` 무경고.
6. 신규 exported 타입에 한국어 `@MX:NOTE` 태그가 부착됨 (REQ-006).
7. PR 이 main 에 머지되어 `merged_pr` / `merged_commit` 정보가 spec.md 의 frontmatter 에 기록됨.
8. plan-auditor 의 사후 감사(post-merge audit)에서 통과 판정.

## 검증 방법 요약

| 시나리오 | 검증 도구 | 자동화 여부 |
|---------|-----------|-----------|
| GWT-1 ~ GWT-6 | `go test ./internal/statusline/...` (renderer_test.go) | 자동화 |
| GWT-7 | `go test ./internal/statusline/...` (builder_test.go 또는 types_test.go) | 자동화 |
| GWT-8, GWT-9 | `go test ./internal/statusline/...` (renderer_test.go, mode 분기) | 자동화 |
| GWT-10 | `go test -cover -race ./internal/statusline/...` | 자동화 |
| Edge cases | 위와 동일 테스트 파일에 추가 | 자동화 |
| Definition of Done #7, #8 | manager-git PR 머지 후 plan-auditor 호출 | 수동 트리거 (자동 검증) |
