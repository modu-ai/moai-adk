# SPEC-CLI-TUX-V3-004 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. **Given** TTY 세션에서 사용자가 `moai doctor`를 실행할 때, **When** 체크가 순차 실행되면, **Then** 각 체크의 라이브 진행 피드백이 stderr에 나타나고, 완료 시 섹션별 pass/fail 테이블이 렌더된다.
2. **Given** CI 환경(non-TTY, `NO_COLOR=1`)에서, **When** `moai doctor` / `moai status` / `moai spec view`를 실행하면, **Then** 세 표면 모두 ANSI 이스케이프가 한 바이트도 없고 plain 정렬 텍스트(또는 마크다운 passthrough)로 출력된다.
3. **Given** M2/M3까지 감소해 온 fmt.Print ratchet에서, **When** 본 SPEC의 소탕이 완료되면, **Then** `internal/cli` 비테스트 소스의 직접 fmt.Printf/Println/Print( 호출이 0건이다(시리즈 최종 게이트).
4. **Given** 배너 경량화 후, **When** `moai`(무인자)를 실행하면, **Then** 컴팩트 1-2행 아이덴티티 + 버전/go/claude pill 메타가 렌더되고 대형 ASCII 로고는 나타나지 않는다.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-TUX4-001 | REQ-TUX4-001 | `go test ./internal/cli/ -run 'DoctorLiveProgress\|DoctorStep' -count=1 -v` (신규 테스트 실존 grep 병행) | PASS — TTY 모드에서 체크별 진행 피드백이 Printer 핸들 경유로 방출됨 단언 |
| AC-TUX4-002 | REQ-TUX4-002 | `go test ./internal/cli/ -run 'DoctorTable\|DoctorSectionResult' -count=1 -v` + `grep -rn 'charm.land/bubbles/v2' internal/cli/doctor*.go \| grep -v '_test.go'` | PASS — 섹션별 pass/fail 테이블 + 카운트 렌더 단언; grep ≥ 1 (bubbles v2 table 실사용 도달 증명) |
| AC-TUX4-003 | REQ-TUX4-003 | `go test ./internal/cli/ -run 'DoctorGolden' -count=1 -v` + `NO_COLOR=1 go run ./cmd/moai doctor 2>/dev/null \| grep -c $'\x1b'` | 골든 PASS + grep 0 (grep exit 1 = no matches) |
| AC-TUX4-004 | REQ-TUX4-004 | `go list -m github.com/charmbracelet/glamour` + `grep -rln 'charmbracelet/glamour' internal/cli --include='*.go' \| grep -v '_test.go'` + status.go/spec_view.go 각각의 glamour-경유 렌더 호출 지점 grep(직접 import 또는 공유 헬퍼 심볼 — 심볼명은 run-phase 확정 후 progress.md §E.2 기재) | glamour resolve + 비테스트 internal/cli import 파일 ≥ 1 + status/spec_view 양쪽 렌더 경로가 glamour-경유 함수를 호출 (도달 가능성 — 토큰 존재가 아닌 렌더 경로 배선; 공유 헬퍼 파일 배치로 인한 false-fail 방지를 위해 status.go/spec_view.go 파일 한정 grep 금지) |
| AC-TUX4-005 | REQ-TUX4-004 | `grep -rn '#[0-9A-Fa-f]\{6\}' internal/cli/status.go internal/cli/spec_view.go internal/cli/glamour_style.go \| grep -v '_test\.go:' \| grep -vE ':[0-9]+:[[:space:]]*//' \| wc -l` — 경로 규약: glamour 스타일 구성은 `internal/cli/glamour_style.go`(신규)로 저작; 실제 파일 경로는 M4a에서 최종 확정(run-phase finalize 조항) — 규약과 다른 경로 채택 시 progress.md §E.2에 실경로 + 치환된 검증 명령 기재 | 0 — glamour 스타일이 tui 토큰 참조로만 구성 (신규 hex 리터럴 부재; AC-CLI-TUI-013 승계) |
| AC-TUX4-006 | REQ-TUX4-005 | `go test ./internal/cli/ -run 'StatusGolden\|SpecViewPlain' -count=1 -v` + `NO_COLOR=1 go run ./cmd/moai status 2>/dev/null \| grep -c $'\x1b'` | 골든 PASS + grep 0 — non-TTY/NO_COLOR에서 plain passthrough |
| AC-TUX4-007 | REQ-TUX4-006 | `go test ./internal/cli/uikit/ -run 'CompactBanner\|BannerPill' -count=1 -v` + `grep -c 'fmt\.Print' internal/cli/uikit/banner.go` | PASS — 1-2행 아이덴티티 + pill(version/go/claude) 단언; grep 0 (12곳 전량 Printer 흡수) |
| AC-TUX4-008 | REQ-TUX4-007 | (1) `grep -n 'help render path verdict' .moai/specs/SPEC-CLI-TUX-V3-004/progress.md` — verdict 기록이 재정렬·골든 커밋보다 선행(커밋 타임라인 대조); (2) 채택 표면 헤더 검사 — verdict=keep-fang 시 `go run ./cmd/moai --help` 출력에 `LAUNCH COMMANDS:` / `PROJECT COMMANDS:` / `TOOLS:` 각각 grep ≥ 1 (2026-07-14 실측 리터럴 — fang이 cobra Group 타이틀을 대문자 변환); verdict=부활/커스터마이즈 시 §E.2에 기록된 채택 표면의 실제 헤더 리터럴로 동일 검사; (3) `go test ./internal/cli/ -run 'HelpGolden\|HelpGroupOrder' -count=1 -v` | verdict 기록 ≥ 1 + 채택 표면에 3그룹 헤더 전부 존재(launch/project/tools 그룹 시맨틱 보존) + 재정렬 골든 PASS (미채택 경로 검사는 N/A 명기 — -002 M2a 스파이크 패턴 미러) |
| AC-TUX4-009 | REQ-TUX4-008 | `go test ./internal/cli/ ./internal/cli/uikit/ -run 'Golden' -count=1` + `GOOS=windows GOARCH=amd64 go build ./...` + progress.md §E.2의 골든 변경 사유 목록 검사 | 골든 전량 PASS + windows 빌드 exit 0 + 변경 골든 파일별 사유 기재 존재 |
| AC-TUX4-010 | REQ-TUX4-009 | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | **0** — 시리즈 최종 grep 게이트 (계획 §7 성공 지표) |
| AC-TUX4-011 | REQ-TUX4-010 | `go test ./internal/cli/ -run 'StdoutClean\|NoWarnOnStdout' -count=1 -v` | PASS — doctor/status/spec/banner 경로의 stdout 캡처에 "Warning"/상태 문자열 부재 (골든) |
| AC-TUX4-012 | REQ-TUX4-011 | `go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/...` | 각 패키지 ≥ 90.0% — 또는 pre-flight #7 baseline이 90% 미만인 패키지에 한해 비회귀 + progress.md §E.2 갭 기록 |
| AC-TUX4-013 | REQ-TUX4-012 | `go test ./... -count=1 && golangci-lint run --timeout=5m && GOOS=windows GOARCH=amd64 go build ./... && GOOS=linux GOARCH=amd64 go build ./...` | 전체 green + lint NEW 0 + 3-타깃 빌드 exit 0 |
| AC-TUX4-014 | REQ-TUX4-001, REQ-TUX4-002 | 코드 리뷰 + `git diff <preflight-HEAD>..HEAD --stat -- internal/cli/doctor.go` 검토: 판정 로직 함수의 시그니처·판정 기준 무변경 | doctor 체크 판정 로직 diff가 진행 리포터 seam 삽입에 한정 (렌더 계층 분리 증명) |

## §C Edge Cases

- **doctor 체크 실패 혼재**: 일부 섹션 fail 시에도 테이블은 전 섹션 렌더 + exit code는 기존 doctor 시맨틱 보존(판정 로직 무변경).
- **glamour 폭 계산**: 터미널 폭 미검출(non-TTY) 시 고정 폭 폴백 — 줄바꿈 깨짐으로 골든 불안정해지지 않게 폭 고정.
- **CJK 렌더**: status/spec 마크다운의 한국어 콘텐츠 — glamour 렌더 폭 계산이 동아시아 폭에서 안정적인지 골든으로 검증.
- **빈 SPEC 디렉터리에서 spec view**: 조회 대상 부재 시 기존 에러 시맨틱 보존 — glamour 경유로 에러 경로가 바뀌지 않음.
- **NO_COLOR 값 시맨틱**: 비어있지 않은 임의 문자열 = set — 시리즈 공통(detect.go isEnvSet 승계).
- **배너 pill의 claude 버전 미검출**: claude CLI 부재 환경에서 pill이 자리표시자로 우아하게 강등(에러/공백 pill 금지).
- **help 골든과 명령 증감**: 이후 명령 추가 시 골든이 깨지기 쉬움 — 골든은 그룹 헤더·정렬 규칙 중심으로 단언(전체 문자열 스냅숏 최소화).

## §D.5 Quality Gate / Definition of Done

- 14 AC 전 행 PASS — verbatim 명령 출력을 progress.md §E.2에 인용 (vci §3 5-section 형식).
- M4a~M4e 마일스톤별 커밋 분리 + 골든 변경 파일별 사유 목록.
- **시리즈 최종 게이트**: ratchet 0 (AC-TUX4-010) + stdout 무경고 (AC-TUX4-011) + 커버리지 ≥ 90%/비회귀 (AC-TUX4-012) — 계획 §7 성공 지표 3종 최종 확인.
- doctor/status/spec 판정·데이터 로직 무변경 증명 (AC-TUX4-014).
- frontmatter `status` 전이는 소유 매트릭스 준수 (draft→in-progress는 manager-develop).
