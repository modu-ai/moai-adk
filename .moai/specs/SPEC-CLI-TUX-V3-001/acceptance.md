# SPEC-CLI-TUX-V3-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. **Given** CI 환경(non-TTY, `NO_COLOR=1`)에서 사용자가 `moai --help`를 실행할 때, **When** fang이 help를 렌더하면, **Then** ANSI 이스케이프 시퀀스가 한 바이트도 출력되지 않고 exit code 0으로 종료된다.
2. **Given** `moai init` 진행 중 프로필 동기화가 실패하는 상황에서, **When** 마이그레이션된 경고 경로가 경고를 방출하면, **Then** "Warning:" 문자열은 stderr에만 나타나고 stdout에는 데이터 외 어떤 상태 문자열도 섞이지 않는다.
3. **Given** lipgloss v1을 소비하는 statusline·wizard(huh v1 테마)·uikit 패키지가 있을 때, **When** internal/tui가 lipgloss v2로 포팅된 후 전체 빌드를 수행하면, **Then** 세 패키지는 소스 무수정으로 컴파일되고 기존 테스트가 전부 green이다 (문자열 토큰 경계 보존의 증거).

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CTX-001 | REQ-CTX-001 | `go list -m charm.land/lipgloss/v2 charm.land/bubbletea/v2 charm.land/bubbles/v2 charm.land/fang/v2` | 4개 모듈 전부 resolve; lipgloss ≥ v2.0.5, bubbletea ≥ v2.0.8, bubbles ≥ v2.1.1, fang ≥ v2.0.1 |
| AC-CTX-002 | REQ-CTX-002 | `go list -m github.com/charmbracelet/huh github.com/charmbracelet/lipgloss github.com/charmbracelet/bubbletea && go build ./...` | huh v1.0.0 + lipgloss v1.x + bubbletea v1.x 공존 상태로 build exit 0 |
| AC-CTX-003 | REQ-CTX-003 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./... && GOOS=linux GOARCH=amd64 go build ./...` | 3개 타깃 전부 exit 0 |
| AC-CTX-004 | REQ-CTX-004 | `grep -rn '"github.com/charmbracelet/lipgloss"' internal/tui --include='*.go' \| grep -v _test.go` | 0 matches — tui 비테스트 소스는 charm.land/lipgloss/v2만 import |
| AC-CTX-005 | REQ-CTX-005 | `go test ./internal/tui/ -run 'NoColor\|Monochrome\|Detect\|Resolve' -count=1 -v` | PASS — NO_COLOR 우선순위 체인 + detect_test.go Resolve 체인(TestResolve/TestThemeResolve) 기존 테스트 green |
| AC-CTX-006 | REQ-CTX-006 | `go test ./internal/tui/ -run 'Theme\|MoaiTheme\|Profile' -count=1 -v` | PASS — light/dark/auto/invalid 분기 기존 테스트 green |
| AC-CTX-007 | REQ-CTX-007 | `go test ./internal/tui/ -run 'ReducedMotion\|Spinner\|Progress' -count=1 -v` | PASS — 정적 변형(● / filled bar) 시맨틱 유지 |
| AC-CTX-008 | REQ-CTX-008 | `grep -rn '#[0-9A-Fa-f]\{6\}' internal/core/project internal/cli --include='*.go' \| grep -v '_test\.go:' \| grep -vE ':[0-9]+:[[:space:]]*//' \| wc -l` | 0 — reporter.go 코드 hex 3건(43-45) 제거 후 code-level hex-free. baseline 3 (comment-line 제외 필터 적용 후 실측 2026-07-10; 전체 hex 라인은 5 = reporter 3 코드 + wizard/styles.go:17,20 주석 2 — 주석 2건은 필터로 배제하고 wizard는 무수정 유지, REQ-CTX-010) |
| AC-CTX-009 | REQ-CTX-009 | `go test ./internal/tui/... -count=1` + `git diff <preflight-HEAD>..HEAD --stat -- internal/tui/testdata internal/tui/golden` — run 착수 시 pre-flight #1의 `git rev-parse HEAD` sha를 progress.md §E.2에 `preflight_head:`로 기록하고 그 sha를 base로 사용 | 스위트 PASS; diff는 본 SPEC run-phase 커밋 범위만 포함(무기준 `git diff`의 커밋-후 공허 PASS 차단) — 골든 변경이 있다면 변경 파일 각각이 progress.md §E.2에 사유와 함께 기재됨 |
| AC-CTX-010 | REQ-CTX-010 | `git diff <preflight-HEAD>..HEAD --stat -- internal/statusline internal/cli/wizard internal/cli/uikit` (AC-CTX-009와 동일한 progress.md §E.2 `preflight_head:` sha 사용) + `go test ./internal/statusline/... ./internal/cli/wizard/... ./internal/cli/uikit/... -count=1` | 보호 3경로 diff 0 — 본 SPEC 커밋 범위 기준이라 병렬 세션의 working-tree dirty 파일(internal/statusline/*)로 인한 false-FAIL 없음 + 테스트 전부 PASS |
| AC-CTX-011 | REQ-CTX-011 | `go test ./internal/cli/printer/ -run 'Interface\|Contract' -count=1 -v` | PASS — Info/Warn/Error/Success/Data/Step + spinner/progress 핸들 계약 테스트 green |
| AC-CTX-012 | REQ-CTX-012 | `go test ./internal/cli/printer/ -run 'ChannelDiscipline' -count=1 -v` | PASS — Data→stdout writer only, Info/Warn/Error/Success/Step→stderr writer only 단언 |
| AC-CTX-013 | REQ-CTX-013 | `go test ./internal/cli/printer/ -run 'Mode' -count=1 -v` | PASS — tty-rich/plain/json 3모드 각각의 렌더 계약 테스트 green |
| AC-CTX-014 | REQ-CTX-014 | `go test ./internal/cli/printer/ -run 'NoANSI\|PlainNoEscape' -count=1 -v` | PASS — plain/json 모드 및 NO_COLOR 시 출력에 `\x1b[` 부재 단언 |
| AC-CTX-015 | REQ-CTX-015 | `grep -c 'fmt\.Printf' internal/core/project/reporter.go; go test ./internal/core/project/ -run 'Reporter' -count=1 -v` | grep 0 + Reporter 테스트 PASS (Printer 경유 확인) |
| AC-CTX-016 | REQ-CTX-016 | `go test ./internal/cli/ -run 'InitWarn.*Stderr\|WarningChannel' -count=1 -v` | PASS — 마이그레이션된 init 경고 경로의 stdout 캡처에 "Warning:" 부재, stderr 캡처에 존재 |
| AC-CTX-017 | REQ-CTX-017 | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | ≤ 46 (baseline, research.md §C) 이면서 46 미만으로 감소; 실측치 progress.md §E.2 기록 |
| AC-CTX-018 | REQ-CTX-018 | `grep -rn 'fang\.Execute' internal/cli cmd/moai --include='*.go' \| grep -v _test.go` | ≥ 1 match — root 실행 경로가 fang.Execute 경유 |
| AC-CTX-019 | REQ-CTX-019 | `go test ./internal/cli/ -run 'Trivial\|LazyInit\|LightDeps' -count=1 -v` | PASS — version/help/completion이 InitDependencies 전체 그래프를 트리거하지 않음 단언 |
| AC-CTX-020 | REQ-CTX-020 | `grep -rn 'func TestFangExitCoderCharacterization' internal/cli cmd/moai --include='*_test.go'` + `go test ./internal/cli/... ./cmd/... -run 'TestFangExitCoderCharacterization' -count=1 -v` | grep ≥ 1 match(신규 characterization 테스트 실존 증명 — 기존 무관 ExitCode/ExitCoder 테스트 매칭에 의한 false-PASS 차단) + PASS — ExitCoder 체인(worktree verify 0/1/2/3 포함) exit code 보존 |
| AC-CTX-021 | REQ-CTX-021 | `NO_COLOR=1 go run ./cmd/moai --help 2>&1 \| grep -c $'\x1b'` 및 `NO_COLOR=1 go run ./cmd/moai init --help 2>&1 \| grep -c $'\x1b'` | 둘 다 0 (grep exit 1 = no matches) — 파이프(non-TTY) + NO_COLOR에서 ANSI 무출력 |
| AC-CTX-022 | REQ-CTX-022 | `grep -rn '#[0-9A-Fa-f]\{6\}' <fang 테마 배선 파일> \| grep -v _test.go` + 코드 리뷰: fang 테마 생성부가 `internal/tui` 토큰 참조 | hex 0건 + tui 토큰 소비 확인 (배선 파일 경로는 run-phase에 확정, progress.md §E.2 기재) |
| AC-CTX-023 | REQ-CTX-023 | `grep -n '@MX:DEBT' internal/merge/confirm.go && grep -n '@MX:CEILING' internal/merge/confirm.go && grep -n '@MX:UPGRADE' internal/merge/confirm.go && go test ./internal/merge/ -count=1` | REQ-CTX-023이 의무화한 3태그(@MX:DEBT + @MX:CEILING + @MX:UPGRADE, SPEC-CLI-TUX-V3-003 트리거) 전부 존재 + merge 스위트 PASS + design.md §D에 결정 기록 |
| AC-CTX-024 | REQ-CTX-024 | `go test -cover ./internal/cli/printer/` | `coverage: ≥ 90.0% of statements` |
| AC-CTX-025 | REQ-CTX-025 | `go test ./... -count=1 && golangci-lint run --timeout=5m` | 전체 스위트 green + lint NEW 항목 0 (pre-flight baseline 대조) |
| AC-CTX-026 | REQ-CTX-009, REQ-CTX-017 | progress.md §E.2 검사: 골든 변경 사유 목록 + ratchet 실측치 기재 여부 | 두 기록 모두 존재 (문서화 게이트) |

## §C Edge Cases

- **TTY 감지 경계**: CI에서 stdout은 파이프이나 stderr가 TTY인 혼합 케이스 — Printer 모드 판단은 채널별이 아닌 구성 시점 1회로 고정하되, 두 채널 모두 escape-free여야 함(plain 모드).
- **NO_COLOR 값 시맨틱**: NO_COLOR 표준은 "비어있지 않은 임의 문자열 = set" — `NO_COLOR=0`도 set으로 취급(기존 detect.go isEnvSet 시맨틱 보존).
- **Windows 콘솔**: `initConsole()`(VT enable, console_windows.go)이 fang.Execute보다 먼저 실행되어야 함 — 순서 역전 시 Windows에서 escape 깨짐.
- **`moai` 인자 없음(root Run)**: 배너 + help 출력 경로가 fang 래핑 후에도 동작; 배너는 M4까지 현행 유지(ratchet 수치에 포함된 12곳 무접촉).
- **golden 개행/폭 차이**: lipgloss v2의 width 계산 변경이 동아시아 폭(runewidth) 렌더에 영향 줄 수 있음 — CJK 문자열 골든을 명시적으로 재확인.
- **go.mod tidy 부수 효과**: v2 모듈 추가로 indirect 그래프가 변할 때 기존 pinned indirect(catppuccin/go 등) 버전 튀지 않는지 diff 확인.
- **fang의 --version 자동화 vs 기존 version.GetVersion()**: 이중 version 표면이 생기지 않도록 단일 소스 유지 — 기존 `moai version`/`--version` 출력 형식 회귀 테스트.

## §D.5 Quality Gate / Definition of Done

- 26 AC 전 행 PASS — verbatim 명령 출력을 progress.md §E.2에 인용 (vci §3 5-section 형식).
- M1a/M1b/M1c 마일스톤별 커밋 분리 + 각 게이트(LSP zero errors, lint NEW 0) 통과 기록.
- 신규 printer 패키지 커버리지 ≥ 90%, tui 스위트 무손실, statusline/wizard/uikit 무수정 green.
- ratchet 실측치가 baseline 46 미만으로 감소했고 명령·수치가 progress.md에 기록됨.
- `.moai/specs/SPEC-CLI-TUX-V3-001/` 산출물의 frontmatter `status` 전이는 소유 매트릭스 준수 (draft→in-progress는 manager-develop).
