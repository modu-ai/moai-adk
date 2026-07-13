# SPEC-CLI-TUX-V3-004 — Implementation Plan

## §A Context

CLI TUX 현대화 4-마일스톤 중 M4(전면 폴리시 + 시리즈 최종 게이트). Tier M — doctor/status/spec 렌더 계층 교체 + 배너 경량화 + help 재정리 + NO_COLOR/non-TTY/Windows 골든 매트릭스 갱신. 방법론: `quality.yaml development_mode: tdd`, 표면별 분화:

- **doctor/status/spec 렌더(M4b/M4c)**: characterization-first — 기존 골든(doctor_golden_test.go, status_golden_test.go)이 회귀 앵커; 렌더 계층 교체 후 골든 갱신은 파일별 사유 문서화.
- **배너 경량화(M4d)**: TDD — 신규 컴팩트 배너 계약 테스트 선행.
- **ratchet 0 게이트(M4e)**: 기계적 소탕 — grep 게이트가 완료 판정.

산출물 SSOT: spec.md(12 REQ) / acceptance.md(14 AC). 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M4 + §7 성공 지표.

§F 마일스톤 순서는 결정-가역성(decision-reversibility) 기준: 변경 가능성이 높은 결정(glamour 스타일·doctor 테이블 UX·배너 아이덴티티)을 앞에, 기계적 이행(fmt.Print 소탕·매트릭스 갱신)을 뒤에 배치한다.

## §B Known Issues (착수 전 인지 사항)

| # | 앵커 (run-phase 재실측) | 이슈 | 대응 |
|---|---|---|---|
| 1 | go.mod: glamour 미도입 (2026-07-13 실측) | glamour의 lipgloss 메이저 의존이 v1이면 v1/v2 혼재 그래프 재확대 | pre-flight #6에서 glamour 최신판의 의존 그래프 확정 후 도입; 문자열 토큰 경계(M1 설계) 승계 |
| 2 | ratchet 잔여 분포 — banner.go 12곳 포함 (2026-07-13 전체 40) | M2/M3 감소분 반영 후에도 doctor/status/spec/banner에 잔여 예상 | pre-flight #5 재실측으로 잔여 인벤토리 확정 → M4e 소탕 목록화 |
| 3 | uikit/banner.go + characterization_test.go + render_test.go | 배너 골든·characterization이 대형 로고 전제 | 경량화는 골든 대량 변경 — 파일별 사유 문서화(blind 재생성 금지) |
| 4 | doctor.go 30.5KB — 체크 실행과 출력이 결합 | 라이브 진행 삽입 시 판정 로직 접촉 위험 | 체크 실행 seam(리포터 인터페이스)과 렌더 분리 — 판정 로직 무변경(§D) |
| 5 | help 실표면(2026-07-14 실측): fang v2 렌더가 cobra Group(root.go:114-118) 타이틀을 대문자(`LAUNCH COMMANDS:` 등)로 표면화 + GroupID 미지정 커맨드는 별도 `COMMANDS` 섹션; help.go:101 renderRootHelpTUI는 fang help 인계에 가려진 dead surface | 종전 Title-Case 리터럴(`Launch Commands:`) 앵커는 실렌더(대문자 변환)와 불일치 — false anchor; 재정렬 지점이 렌더 경로 결정에 종속 | M4d 선행 verdict(keep-fang / renderRootHelpTUI 부활 / fang 커스터마이즈)로 렌더 경로 확정 후 재정렬 — 그룹/소속 변경 금지 + Group ID 무변경 단언 유지 (REQ-TUX4-007) |
| 6 | doctor_golden_test.go / status_golden_test.go 실존 | 갱신 대상 골든의 현행 green 여부 미확인 | pre-flight #4에서 현행 green 확인 후 착수 |
| 7 | REQ-TUX4-011 커버리지 ≥ 90% — 현행 baseline 미실측 | cli/template/hook가 이미 90% 미만일 가능성 | pre-flight #7 실측 → 미만 패키지는 비회귀 강등 + 갭 기록 (REQ-TUX4-011 내장 완화) |
| 8 | -002/-003 미완료 (draft) | depends_on strict fulfillment로 run 진입 차단 상태 | 본 SPEC run은 -002/-003 `completed` 이후에만 — Depends_on pre-flight가 기계 강제 |
| 9 | `moai spec` 서브커맨드 다수 (view/status/audit/lint/close/...) | glamour 적용 표면 확정 필요 | 범위는 view(+status 요약 표면)로 한정 — §C Out of Scope 명시 |
| 10 | fang(M1) help 렌더와 cobra group 정렬 상호작용 | fang이 help 정렬을 자체 처리하면 재정렬 지점이 다름 | pre-flight에서 fang help 렌더 경로 확인 후 정렬 지점 확정 |

## §C Pre-flight (착수 전 의무 검증)

```bash
# 1. baseline 확인
git branch --show-current && git rev-parse HEAD

# 2. depends_on 충족 확인 (strict: completed)
grep -m1 'status:' .moai/specs/SPEC-CLI-TUX-V3-002/spec.md
grep -m1 'status:' .moai/specs/SPEC-CLI-TUX-V3-003/spec.md

# 3. cross-platform build + lint baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=5m 2>&1 | tail -5

# 4. 골든 현행 green 앵커
go test ./internal/cli/ -run 'Golden' -count=1
go test ./internal/cli/uikit/ -count=1

# 5. ratchet 잔여 인벤토리 재실측 (M2/M3 감소분 반영 후)
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go'

# 6. glamour 최신판 + 의존 그래프 확정
go list -m -versions github.com/charmbracelet/glamour | tr ' ' '\n' | tail -1
# 도입 전 glamour의 lipgloss 메이저 의존을 `go mod graph`로 확인

# 7. coverage baseline (REQ-TUX4-011 강등 판단 기준점)
go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/... 2>&1 | grep coverage
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (절대 무접촉)**: doctor 체크 판정 로직(항목·판정 기준), status/spec 데이터 조회 로직, `internal/statusline/**`, `internal/cli/{init.go,update.go,update/**,wizard/**}`(M2/M3 산출물 — 골든 매트릭스 검증 외 수정 금지), 무관 SPEC 디렉터리.
- **렌더 계층만 교체**: 표시 데이터(체크 결과 집합·상태 필드)의 의미 변경 금지 — 골든 diff는 서식 변화만.
- **그룹 안정성**: cobra Group ID(launch/project/tools) 및 명령 소속 무변경 — 정렬만.
- **hex 금지 승계**: 신규 렌더(테이블/glamour 스타일/pill)는 tui 토큰만 소비 — AC-CLI-TUI-013.
- **Template-First 비적용**: 내부 Go 코드만 — `internal/template/templates/**` 무접촉.
- **금지 명령**: `--no-verify`, `--amend`, force-push. Conventional Commits (`feat(SPEC-CLI-TUX-V3-004): M4a ...`) + `🗿 MoAI` trailer 의무.

## §E Self-Verification (완료 보고 의무 항목)

- E1: acceptance.md 14 AC 전 행 PASS/FAIL 매트릭스 + verbatim 출력 (vci 5-section 형식).
- E2: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` verbatim.
- E3: `go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/...` — ≥ 90% 또는 비회귀 강등 사유.
- E4: ratchet 최종 grep — **0건** verbatim (시리즈 최종 게이트).
- E5: `golangci-lint run --timeout=5m` NEW 항목 0.
- E6: 골든 변경 파일별 사유 목록 (progress.md §E.2 인용).
- E7: 신규 commit SHA 목록 + push 상태.

## §F Milestones (priority order — 결정-가역성 우선)

### M4a — glamour 도입 + 스타일 결정 (Priority: High, 설계 결정)

1. Pre-flight #6로 glamour 버전·의존 그래프 확정 — lipgloss 메이저 충돌 시 대응 결정(스타일 JSON 직접 구성 등) 기록.
2. tui 토큰 → glamour 스타일 매핑 결정 (hex 직접 기입 금지 — 토큰 참조 구성).
3. `go get github.com/charmbracelet/glamour@<pinned>` 도입 커밋.
4. **게이트**: `go build ./...` green + 스타일 결정 progress.md §E.2 기록.

### M4b — status/spec glamour 렌더 (Priority: High, UX)

1. characterization: status/spec view 현행 출력 골든 앵커.
2. RED→GREEN: glamour 렌더 경유 + non-TTY/NO_COLOR plain passthrough 폴백. (REQ-TUX4-004/005)
3. **게이트**: AC-TUX4-004~006 통과.

### M4c — doctor 라이브 진행 + 결과 테이블 (Priority: High, UX)

1. characterization: doctor 골든 현행 green 앵커.
2. 체크 실행에 진행 리포터 seam 삽입(판정 로직 무변경) → Printer step/spinner 라이브 피드백. (REQ-TUX4-001)
3. 결과를 섹션별 pass/fail 테이블로 렌더(bubbles v2 table TTY / plain 정렬 폴백). (REQ-TUX4-002)
4. 골든 갱신 — NO_COLOR/non-TTY ANSI 부재 단언 포함. (REQ-TUX4-003)
5. **게이트**: AC-TUX4-001~003 통과.

### M4d — 배너 경량화 + help 그룹 재정리 (Priority: Medium, 브랜드/정보 구조)

1. RED: 컴팩트 배너 계약 테스트(1-2행 + pill 메타 버전/go/claude, 글리프 화이트리스트, hex 부재).
2. GREEN: banner.go 재구현 + 12곳 fmt.Print* Printer 흡수. (REQ-TUX4-006)
3. **help 렌더 경로 verdict(-002 M2a 스파이크 패턴 미러)**: keep-fang(현행 — cobra Group 대문자 표면화, 2026-07-14 실측) / renderRootHelpTUI 부활(help.go:101 dead surface; 4-그룹→3-그룹 정합 포함) / fang 렌더 커스터마이즈(지원 여부 확인, 미지원 시 plan-B) 중 채택 — verdict + 근거 + 채택 표면의 실제 헤더 리터럴을 progress.md §E.2에 기록, 재정렬·골든 커밋보다 선행. (REQ-TUX4-007)
4. verdict 채택 표면에서 help 그룹 내 사용 빈도순 재정렬 + help 골든 갱신 (그룹 ID/소속 무변경 단언). (REQ-TUX4-007)
5. **게이트**: AC-TUX4-007~008 통과 + uikit 스위트 green.

### M4e — ratchet 0 소탕 + 회귀 매트릭스 (Priority: Medium, 시리즈 최종 게이트 — 기계적)

1. Pre-flight #5 인벤토리의 잔여 fmt.Print* 전량 Printer 전환 — 파일 단위 커밋.
2. 최종 grep 게이트: `internal/cli` 비테스트 0건. (REQ-TUX4-009)
3. 회귀 매트릭스: 폴리시 표면(doctor/status/spec/banner/help) × {NO_COLOR=1, 파이프 non-TTY, GOOS=windows 빌드} 골든 갱신 + stdout 무경고 골든. (REQ-TUX4-008/010)
4. 커버리지 측정 — ≥ 90% 또는 비회귀 강등 + 갭 기록. (REQ-TUX4-011)
5. **게이트**: 전체 스위트 green + lint NEW 0 + §E 자가 검증 전 항목. (REQ-TUX4-012)

### 검증 명령 배치 (M4e 종료 시, 단일 턴 병렬 실행)

```bash
go test ./... -count=1 > /tmp/moai-verify/1-full.log 2>&1; echo "exit=$?"; tail -20 /tmp/moai-verify/1-full.log
go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/... > /tmp/moai-verify/2-cover.log 2>&1; echo "exit=$?"; grep coverage /tmp/moai-verify/2-cover.log
GOOS=windows GOARCH=amd64 go build ./... > /tmp/moai-verify/3-win.log 2>&1; echo "exit=$?"
golangci-lint run --timeout=5m > /tmp/moai-verify/4-lint.log 2>&1; echo "exit=$?"; tail -10 /tmp/moai-verify/4-lint.log
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l; echo "expect 0 (시리즈 최종 게이트)"
NO_COLOR=1 go run ./cmd/moai doctor 2>/dev/null | grep -c $'\x1b'; echo "expect 0"
NO_COLOR=1 go run ./cmd/moai --help 2>&1 | grep -c $'\x1b'; echo "expect 0"
```

## §G Anti-Patterns and Risks

- **Anti-pattern: 골든 일괄 재생성으로 폴리시 "완료" 선언** — 서식 변경과 데이터 변경을 구분 못 하게 됨. 파일별 사유 문서화 의무(§D).
- **Anti-pattern: doctor 판정 로직에 렌더 관심사 침투** — 진행 리포터 seam으로 분리; 판정 결과 스키마 무변경.
- **Anti-pattern: glamour 스타일에 hex 직접 기입** — tui 토큰 참조로만 구성(AC-CLI-TUI-013 승계).
- **Anti-pattern: ratchet 0을 위해 fmt.Fprintf(os.Stderr, ...)로 우회** — 게이트의 정신은 "전량 Printer 경유"다. 우회 패턴은 grep 확장(AC-TUX4-011의 stdout 유출 골든)으로 노출된다.
- **Risk: glamour의 lipgloss v1 의존으로 그래프 재확대** — pre-flight #6 확정 전 도입 금지; 충돌 시 스타일 직접 구성 대안 기록.
- **Risk: help 재정렬이 스크립트 파싱 사용자를 깨뜨림** — help 텍스트는 계약이 아니나, 그룹 구조 유지로 완충; CHANGELOG에 명시.
- **Risk: -002/-003 지연으로 본 SPEC 장기 대기** — depends_on strict 게이트가 의도된 동작 — 우회(--ignore-deps) 시 ratchet 0 게이트가 성립 불가함을 blocker로 보고.

## §H Cross-References

- 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M4 + §7 성공 지표(본 SPEC이 최종 게이트 소유).
- 선행: SPEC-CLI-TUX-V3-001 (completed) / -002 (M2) / -003 (M3) — 전부 `completed` 필수.
- 계약 승계: REQ-CTX-011~014 (Printer), REQ-CTX-017→TUX2-018→TUX3-020→**TUX4-009 (ratchet 종착 0)**, AC-CLI-TUI-013 (hex 금지).
- 규약: `internal/cli/CLAUDE.md` (output streams), CLAUDE.local.md §6 (테스트 격리) · §14 (하드코딩 금지).
- 방법론: characterization (M4b/M4c 골든 앵커) + moai-workflow-tdd (M4d).
