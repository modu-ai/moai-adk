# SPEC-CLI-TUX-V3-001 — Implementation Plan

## §A Context

CLI TUX 현대화 4-마일스톤 중 M1(기반 교체). Tier L — Charm v2 의존성 도입 + `internal/tui` 커널 포팅(1,744 LOC 실측) + 신규 printer 패키지 + fang root 교체가 걸린 constitutional-급 기반 작업. 방법론: `quality.yaml development_mode: tdd` — 단, 표면별로 분화한다:

- **M1a (tui v2 포팅)**: brownfield TDD의 pre-RED 분석 확장 = **characterization-first** (DDD ANALYZE-PRESERVE-IMPROVE 자세). 기존 tui 테스트 스위트 + `internal/tui/testdata/`(≈108 엔트리, run-phase 재실측 — research.md §C.4) + golden 디렉터리가 회귀 앵커.
- **M1b (printer 신규)**: 순수 TDD RED-GREEN-REFACTOR. 신규 패키지이므로 spec test가 계약을 먼저 정의.
- **M1c (fang + 회귀 매트릭스)**: 통합 지점 — 기존 root 테스트를 characterization으로 보존 후 fang 배선.

산출물 SSOT: spec.md(25 REQ) / acceptance.md(26 AC) / design.md(경계 설계) / research.md(실측 기록). 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html`.

## §B Known Issues (착수 전 인지 사항)

| # | 앵커 (run-phase 재실측) | 이슈 | 대응 |
|---|---|---|---|
| 1 | go.mod: huh v1.0.0 → lipgloss v1 의존 | v1/v2 메이저 혼재 불가피 | tui 문자열 토큰 경계 유지 (design.md §B); v1 제거는 stretch goal |
| 2 | init.go:223,227,237,327,334,488,500,509,514 | 경고 8곳 stdout 유출 (`fmt.Fprintf(cmd.OutOrStdout(), "Warning: ...")`) | M1b에서 Printer 경유 stderr 라우팅 |
| 3 | internal/core/project/reporter.go:43-45 | hex AdaptiveColor 리터럴 3개 (AC-CLI-TUI-013 계약 밖 잔존) + fmt.Printf stdout 직행 | Printer 어댑터로 흡수, hex 제거 |
| 4 | internal/merge/confirm.go (879 LOC, 모델 43-224) | 유일한 bubbletea v1 사용처 | v1 잔류 + @MX:DEBT (design.md §D 결정) — M3에서 bubbles v2 list 승격 |
| 5 | root.go:42 trivialCommands + Execute():61 | fang 래핑 시 lazy-init(REQ-PERF-003-A) 우회 위험 | fang.Execute 앞단에 isTrivialCommand 분기 보존 |
| 6 | cmd/moai/main.go ExitCoder | fang이 에러 체인을 가공하면 worktree verify 0/1/2/3 exit code 깨질 위험 | exit-code characterization test 선행 |
| 7 | internal/statusline/* (working tree 수정 중 — 병렬 세션) | 접촉 시 병렬 작업 충돌 | PRESERVE — 절대 무접촉 (§D) |
| 8 | tui golden/testdata ≈108 엔트리 (재실측 의무) | lipgloss v2 렌더러 변경으로 골든 대량 변동 가능 | 변동 골든은 개별 사유 문서화 (REQ-CTX-009); blind 재생성 금지 |
| 9 | bubbletea/bubbles/fang v2 API 상세 미실측 | v2 시그니처는 본 세션에서 미검증 (research.md §F) | run-phase 착수 시 Context7/godoc으로 API 확정 후 착수 |
| 10 | uikit/banner.go 12곳 fmt.Print* | ratchet 베이스라인의 최대 기여자이나 배너 경량화는 M4 소관 | M1에서 무접촉 — ratchet 수치만 관리 |

## §C Pre-flight (착수 전 의무 검증)

```bash
# 1. baseline 확인
git branch --show-current && git rev-parse HEAD

# 2. cross-platform build baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline (NEW vs pre-existing 구분)
golangci-lint run --timeout=5m 2>&1 | tail -5

# 4. tui characterization baseline — 포팅 전 전량 green 확인
go test ./internal/tui/... -count=1

# 5. ratchet baseline 재실측 (research.md §C의 46과 대조)
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l

# 6. charm.land v2 최신 버전 재확인
go list -m -versions charm.land/lipgloss/v2 | tr ' ' '\n' | tail -1
go list -m -versions charm.land/bubbletea/v2 | tr ' ' '\n' | tail -1
go list -m -versions charm.land/bubbles/v2 | tr ' ' '\n' | tail -1
go list -m -versions charm.land/fang/v2 | tr ' ' '\n' | tail -1

# 7. v2 API 확정 (Known Issue #9) — Context7 resolve-library-id + query-docs 또는 go doc
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (절대 무접촉)**: `internal/statusline/**` (병렬 세션 작업 중), `CLAUDE.local.md`, `.moai/state/**`, `.moai/reports/**`, 무관 SPEC 디렉터리. working tree의 기존 수정분(`internal/statusline/cache_hit_test.go`, `internal/statusline/renderer.go`)을 커밋에 포함 금지 — `git add`는 specific path만.
- **공개 계약 보존**: `internal/tui` Theme 구조체의 string 필드명·`Catppuccin*` 상수·exported 함수 시그니처 변경 금지(REQ-CTX-004). 소비자(statusline/wizard/uikit) 소스 수정 없이 컴파일되어야 함(REQ-CTX-010).
- **위저드 huh v1 잔류**: `internal/cli/wizard`의 huh 의존 코드는 로직 무변경 (M2 소관).
- **Template-First 비적용**: 본 SPEC은 내부 Go 코드만 — `internal/template/templates/**` 무접촉.
- **금지 명령**: `--no-verify`, `--amend`, force-push. Conventional Commits (`feat(SPEC-CLI-TUX-V3-001): M1a ...`) + `🗿 MoAI` trailer 의무.
- **골든 blind 재생성 금지**: 골든 파일 변경은 파일별 사유를 progress.md §E.2에 기록.

## §E Self-Verification (완료 보고 의무 항목)

- E1: acceptance.md 26 AC 전 행 PASS/FAIL 매트릭스 + verbatim 출력 (vci 5-section 형식).
- E2: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` verbatim.
- E3: `go test -cover ./internal/cli/printer/` ≥ 90.0% + `go test ./internal/tui/... ./internal/cli/... ./internal/merge/... -count=1` green.
- E4: subagent boundary grep — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/printer/ | grep -v _test` 0건.
- E5: `golangci-lint run --timeout=5m` NEW 항목 0 (baseline 대조).
- E6: ratchet 재실측 — 46 이하 + 실제 감소치 보고.
- E7: 신규 commit SHA 목록 + push 상태.

## §F Milestones (priority order)

### M1a — 의존성 도입 + internal/tui v2 포팅 (characterization-first)

1. Pre-flight #4의 tui 스위트 green 기록(characterization 앵커 확정).
2. `go get charm.land/{lipgloss,bubbletea,bubbles,fang}/v2@<pinned>` — REQ-CTX-001 버전 핀.
3. v2 API 확정 스파이크(Known Issue #9): lipgloss v2의 Style/Color/배경 감지 시그니처를 문서로 확정, 포팅 규칙표 작성.
4. tui 15개 파일 포팅 — 파일 단위 커밋, 각 커밋 후 `go test ./internal/tui/... -count=1`. 공개 문자열 토큰 계약 불변.
5. 소비자 무변경 검증: `go build ./internal/statusline/ ./internal/cli/wizard/ ./internal/cli/uikit/` + 각 테스트 green.
6. **게이트**: LSP zero errors + lint NEW 0 + AC-CTX-001~010 통과.

### M1b — internal/cli/printer 신규 (TDD RED-GREEN-REFACTOR) + reporter/init 배선

1. RED: Printer 인터페이스 계약 테스트(메서드 집합·채널 규율·3모드·ANSI 부재) 선작성 — 실패 확인.
2. GREEN: 최소 구현 (tty-rich는 tui 토큰 소비, plain은 무장식, json은 구조화 라인).
3. REFACTOR: spinner/progress 핸들 정리, `MOAI_REDUCED_MOTION` 존중 확인.
4. `ConsoleReporter` 어댑터/재배선 (hex 제거 — REQ-CTX-008/015) + init.go 대표 경로(성공/경고) 전환 (REQ-CTX-016).
5. ratchet 재실측 — baseline 46 대비 감소 확인 (REQ-CTX-017).
6. **게이트**: printer 커버리지 ≥ 90% + AC-CTX-011~017 통과.

### M1c — Fang 통합 + confirm.go 결정 기록 + 회귀 매트릭스

1. exit-code + trivial fast-path characterization test 선행 (Known Issue #5/#6; exit-code 테스트명은 `TestFangExitCoderCharacterization`으로 고정 — AC-CTX-020 실존 grep과 일치 의무).
2. `fang.Execute(ctx, rootCmd)` 배선 — initConsole/lazy-init 분기 보존, 테마는 tui 토큰 주입 (REQ-CTX-018/019/022).
3. `internal/merge/confirm.go`에 @MX:DEBT(+CEILING/UPGRADE) 주석 — design.md §D 결정 반영 (REQ-CTX-023).
4. 회귀 매트릭스: non-TTY(`| cat` 파이프) × `NO_COLOR=1` × `GOOS=windows` 빌드 × exit code — AC-CTX-018~023.
5. **게이트**: 전체 스위트 green + lint NEW 0 + §E 자가 검증 전 항목.

### 검증 명령 배치 (M1c 종료 시, 단일 턴 병렬 실행)

```bash
go test ./... -count=1 > /tmp/moai-verify/1-full.log 2>&1; echo "exit=$?"; tail -20 /tmp/moai-verify/1-full.log
go test -coverprofile=/tmp/moai-verify/printer.out ./internal/cli/printer/ > /tmp/moai-verify/2-cover.log 2>&1; echo "exit=$?"; tail -5 /tmp/moai-verify/2-cover.log
GOOS=windows GOARCH=amd64 go build ./... > /tmp/moai-verify/3-win.log 2>&1; echo "exit=$?"
golangci-lint run --timeout=5m > /tmp/moai-verify/4-lint.log 2>&1; echo "exit=$?"; tail -10 /tmp/moai-verify/4-lint.log
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l
NO_COLOR=1 go run ./cmd/moai --help 2>&1 | grep -c $'\x1b'; echo "expect 0"
grep -rn '#[0-9A-Fa-f]\{6\}' internal/core/project internal/cli --include='*.go' | grep -v '_test\.go:' | grep -vE ':[0-9]+:[[:space:]]*//' | wc -l; echo "expect 0 (comment-line 제외 — AC-CTX-008과 동일 형식)"
```

## §G Anti-Patterns and Risks

- **Anti-pattern: 골든 일괄 재생성으로 포팅 "완료" 선언** — lipgloss v2 출력 차이를 골든 덮어쓰기로 삼키면 behavior-preserving 검증이 무력화. 파일별 사유 문서화 의무(§D).
- **Anti-pattern: tui 공개 API를 lipgloss v2 타입으로 확장** — `lipgloss.Style`(v2) 반환 함수를 export하는 순간 v1 소비자(huh 테마 경로)가 깨진다. 경계는 문자열 토큰뿐(design.md §B).
- **Anti-pattern: fang 배선과 출력 채널 전환을 한 커밋에 혼합** — 회귀 원인 분리 불가. M1a/M1b/M1c 커밋 분리.
- **Risk: bubbletea v2가 lipgloss v2를 강제 → 간접 그래프 팽창** — confirm.go v1 잔류 결정으로 M1에서는 bubbletea v2 실사용이 없을 수 있음. bubbles v2 스피너를 printer가 소비하는 순간 v2 그래프가 활성화 — go.mod tidy 결과를 커밋별 확인.
- **Risk: fang의 에러 출력 형식 변경으로 기존 에러-메시지 의존 테스트 파손** — characterization 선행(M1c-1)으로 파손 지점을 먼저 드러냄.
- **Risk: charm.land vanity 도메인 프록시 장애 시 CI 실패** — GOPROXY 캐시로 완화; 재현 시 GONOSUMDB 아닌 표준 프록시 경로 유지.
- **Risk: 병렬 세션(statusline) 충돌** — §D PRESERVE 및 pre-spawn sync check 준수.

## §H Cross-References

- 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §1-§8 (M1 = §4 첫 항목).
- 후속: SPEC-CLI-TUX-V3-002 (M2 init), -003 (M3 update + confirm.go 승격), -004 (M4 폴리시).
- 계약 승계: AC-CLI-TUI-013 (hex 금지, SPEC-CLI-UIKIT-KERNEL-001), REQ-PERF-003-A (lazy-init).
- 규약: `internal/cli/CLAUDE.md` (output streams / exit codes / env keys), CLAUDE.local.md §6 (테스트 격리) · §14 (하드코딩 금지).
- 방법론: moai-workflow-tdd (M1b) + moai-workflow-ddd characterization (M1a).
