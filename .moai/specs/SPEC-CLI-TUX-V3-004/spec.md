---
id: SPEC-CLI-TUX-V3-004
title: "doctor/status/spec Surface Polish + Regression Matrix (CLI TUX v3 — M4)"
version: "0.1.1"
status: completed
created: 2026-07-13
updated: 2026-07-20
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli (doctor, status, spec) + internal/cli/uikit"
lifecycle: spec-anchored
tags: "cli, tux, doctor, status, spec-view, glamour, banner, help-groups, golden-tests, regression-matrix, tier-m"
era: V3R6
tier: M
depends_on: [SPEC-CLI-TUX-V3-001, SPEC-CLI-TUX-V3-002, SPEC-CLI-TUX-V3-003]
---

# SPEC-CLI-TUX-V3-004 — doctor/status/spec 표면 폴리시 + 회귀 매트릭스 (CLI TUX 현대화 M4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial draft — CLI TUX 현대화 계획 보고서(`.moai/reports/moai-cli-tux-modernization-plan-20260710.html`) §4 M4로부터 작성 |
| 0.1.1 | 2026-07-14 | manager-spec | plan-audit iter-1 delta-fix — D1(BLOCKING): help 표면 재실측(fang v2가 cobra Group 타이틀을 대문자로 실렌더 — 브리프의 "그룹 헤더 0" 주장 정정; renderRootHelpTUI는 fang에 가려진 dead surface 확인) → REQ-TUX4-007을 렌더 경로 verdict 선행 게이트로 재작성 + AC-TUX4-008 실측 리터럴 재앵커, D2: AC-TUX4-005 스타일 파일 경로 규약 확정, D5: AC-TUX4-004 grep을 internal/cli 전역으로 확대 |

## §A Context

CLI TUX 현대화의 마지막 마일스톤. M1(기반)·M2(init)·M3(update)이 핵심 플로우를 현대화한 뒤, M4는 잔여 표면을 폴리시하고 시리즈 전체의 성공 지표(계획 §7)를 최종 게이트로 닫는다:

1. **doctor**: `internal/cli/doctor.go`(30.5KB, 2026-07-13 실측)는 체크를 일괄 실행 후 결과를 나열한다 — 체크 항목 라이브 진행 + 섹션별 pass/fail 테이블(bubbles v2 table) 부재. 골든 테스트(`doctor_golden_test.go`)는 존재.
2. **status/spec**: `moai status`(status.go) / `moai spec view·status`(spec_view.go, spec_status.go)가 마크다운 콘텐츠(SPEC 요약·핸드오프)를 무서식 출력한다 — glamour 렌더 부재. **glamour는 go.mod에 미도입**(2026-07-13 실측).
3. **배너**: `internal/cli/uikit/banner.go`에 직접 `fmt.Print*` 12곳(실측) — 대형 ASCII 로고를 컴팩트 1-2행 아이덴티티 + pill 메타(버전/go/claude)로 경량화하고 Printer로 흡수한다.
4. **help 그룹**: `moai --help`의 실표면(2026-07-14 실측)은 fang v2 help 렌더다 — cobra Group(root.go:114-118 launch/project/tools) 타이틀이 fang 대문자 스타일(`LAUNCH COMMANDS:`/`PROJECT COMMANDS:`/`TOOLS:`)로 표면화되고, GroupID 미지정 커맨드(ast-grep/completion/goal/harness/help/migration)는 별도 `COMMANDS` 섹션에 나열된다. help.go:101 `renderRootHelpTUI`(별도 4-그룹 커스텀 표면)는 root.go SetHelpFunc로 설치되어 있으나 fang help 인계에 가려 실표면에 나타나지 않는다(dead surface). M4d는 help 렌더 경로를 먼저 결정한 뒤(REQ-TUX4-007), launch/project/tools 그룹 시맨틱 유지 + 그룹 내 사용 빈도순 재정렬을 채택 표면에 적용한다.
5. **회귀 매트릭스**: `NO_COLOR` × non-TTY × Windows Terminal 골든 테스트를 시리즈 전 표면에 대해 갱신 — M1~M4 누적 변경의 최종 회귀 검증.

**시리즈 최종 게이트 (계획 §7 성공 지표)**: 본 SPEC이 시리즈의 마지막이므로 다음 지표를 최종 검증한다 — (a) `internal/cli` 비테스트 소스의 직접 `fmt.Printf`/`fmt.Println`/`fmt.Print(` 호출 **0건**(전량 Printer 경유; 2026-07-13 실측 40에서 M2/M3 감소분을 이어받아 본 SPEC이 0으로 마감), (b) 경고/상태 문자열의 stdout 유출 0건(골든), (c) cli/template/hook 커버리지 ≥ 90% 유지.

본 SPEC은 M1~M3 완료를 전제한다(depends_on 3건 — Depends_on pre-flight가 `status: completed`를 강제). 모든 라인 번호·수치 앵커는 run-phase 착수 시 재실측한다.

## §B Requirements (GEARS)

### B.1 doctor 라이브 진행 + 결과 테이블

- **REQ-TUX4-001**: While `moai doctor` checks execute on a TTY, the progress surface shall render live per-check feedback (spinner/step lines through the `internal/cli/printer` handles) instead of silent batch execution.
- **REQ-TUX4-002**: When `moai doctor` completes, the results shall render as a per-section pass/fail table (bubbles v2 table on TTY; aligned plain-text table on non-TTY) with per-section and overall counts.
- **REQ-TUX4-003**: While the process runs non-TTY or with `NO_COLOR` set, the doctor output shall contain zero ANSI escape sequences, and the refreshed doctor golden tests shall assert this.

### B.2 status/spec glamour 렌더

- **REQ-TUX4-004**: The go.mod dependency set shall include `github.com/charmbracelet/glamour` as a direct dependency, and `moai status` and `moai spec view` shall render their markdown payloads through glamour with a style derived from `internal/tui` tokens (no new raw hex literals outside internal/tui — AC-CLI-TUI-013 계약 승계).
- **REQ-TUX4-005**: While the process runs non-TTY or with `NO_COLOR` set, the glamour-rendered surfaces shall fall back to plain markdown passthrough with zero ANSI escape sequences.

### B.3 배너 경량화 + help 그룹 재정리

- **REQ-TUX4-006**: The banner shall render as a compact 1-2 line identity with pill metadata (version / go / claude) replacing the large ASCII logo, and the `internal/cli/uikit/banner.go` direct `fmt.Print*` call sites (실측 12곳) shall be absorbed into the Printer abstraction.
- **REQ-TUX4-007**: When M4d begins, the help render path shall first be decided as an explicit spike-style verdict recorded in progress.md §E.2 before any reorder/golden commit — options: (a) keep fang v2 group rendering (현행 실표면, 2026-07-14 실측 — cobra Group 타이틀이 대문자로 표면화), (b) revive the shadowed `renderRootHelpTUI` custom surface (help.go:101 dead surface — 부활 시 4-그룹→launch/project/tools 3-그룹 정합 결정 포함), (c) customize fang's help rendering where supported (지원 여부는 스파이크로 확정, 미지원 시 plan-B 기록 — -002 M2a 패턴 미러). While the adopted surface renders root help, the launch/project/tools grouping semantics shall be retained with commands reordered by usage frequency within each group, and the help golden shall be refreshed against the adopted surface's actual header literals.

### B.4 회귀 매트릭스 (시리즈 최종)

- **REQ-TUX4-008**: The golden-test regression matrix shall cover the polished surfaces (doctor / status / spec view / banner / help) under each of: `NO_COLOR=1`, non-TTY (piped) output, and the Windows build target (`GOOS=windows GOARCH=amd64 go build ./...`), with goldens refreshed and each legitimately changed golden individually documented in progress.md §E.2.
- **REQ-TUX4-009**: The count of direct `fmt.Printf`/`fmt.Println`/`fmt.Print(` calls in non-test `internal/cli` sources shall reach ZERO — all output routes through the Printer (series-final grep gate per plan §7; ratchet 46→40→M2/M3 감소분→0).
- **REQ-TUX4-010**: The polished surfaces shall not leak any warning/status strings to stdout — stdout carries data only (golden-verified across doctor/status/spec/banner paths).

### B.5 품질 게이트

- **REQ-TUX4-011**: The statement coverage of `internal/cli`, `internal/template`, and `internal/hook` shall be >= 90% per the plan §7 success metric; where the re-measured pre-flight baseline is already below 90% for a package, the gate degrades to strict non-regression for that package with the gap recorded in progress.md §E.2.
- **REQ-TUX4-012**: The full repository test suite (`go test ./... -count=1`), `golangci-lint run` (no NEW findings), and cross-platform builds (`GOOS=windows`, `GOOS=linux`) shall pass.

## §C Out of Scope (Non-Goals)

본 SPEC은 CLI TUX 현대화 계획의 M4만 다룬다.

### Out of Scope — init/update 플로우 재개편 (M2/M3 소관 완료분)
- init 위저드·셀프업데이트 순서·update 분해/프리뷰의 기능 변경은 다루지 않는다 — 본 SPEC은 해당 표면의 골든 매트릭스 검증만 수행.
- M2/M3 산출물에서 발견되는 기능 결함은 본 SPEC 범위가 아니라 해당 SPEC의 부채/후속으로 처리.

### Out of Scope — statusline 시각 언어
- `internal/statusline`은 시리즈 전체에서 범위 외로 고정(계획 §6) — 토큰 팔레트 공유 문서화만 M1에서 완료. 본 SPEC 무접촉.

### Out of Scope — glamour의 전면 적용
- glamour 렌더는 `moai status` + `moai spec view` 2개 표면에 한정. 다른 마크다운 출력 표면(handoff 열람 등)으로의 확대는 후속 SPEC.

### Out of Scope — 신규 doctor 체크 항목
- doctor의 체크 **내용**(항목 추가/제거/판정 로직)은 무변경 — 본 SPEC은 진행 표시와 결과 렌더 계층만 교체한다.

### Out of Scope — internal/cli 외부 패키지의 fmt.Print 소탕
- REQ-TUX4-009의 0-게이트 범위는 `internal/cli` 비테스트 소스로 한정(계획 §7 문언 그대로). `internal/core` 등 타 패키지의 잔여 호출은 범위 외.

## §D Non-Functional Constraints

- **표시 계층만 변경**: doctor 판정·status 데이터·spec 조회 로직 무변경 — 렌더 계층 교체는 behavior-preserving(데이터 표면 기준) + 골든 갱신 사유 개별 문서화.
- **채널 규율**: stdout=데이터, stderr=상태/경고 — 시리즈 공통 HARD, 골든으로 기계 검증.
- **Windows/CI 안전**: 모든 TTY 신기능은 non-TTY plain 폴백 페어. `GOOS=windows` 빌드 필수.
- **배너 경량화의 브랜드 보존**: 컴팩트 아이덴티티도 MoAI 글리프 어휘(`✓ ✗ ! ● ○ ◆` 화이트리스트)와 tui 토큰만 사용 — 신규 hex 금지.
- **help 재정렬의 안정성**: 그룹 ID(launch/project/tools) 및 명령 소속 무변경 — 정렬 순서만.

## §E Dependencies & Interop

- **depends_on**: SPEC-CLI-TUX-V3-001 (completed) + SPEC-CLI-TUX-V3-002 (draft) + SPEC-CLI-TUX-V3-003 (draft) — Depends_on pre-flight의 strict fulfillment(`status: completed`)에 따라 -002/-003 완료 전 run 진입이 차단된다. REQ-TUX4-009(ratchet 0)와 REQ-TUX4-008(전 표면 매트릭스)이 M2/M3 산출물 위에서만 성립하기 때문.
- **glamour 신규 의존**: `github.com/charmbracelet/glamour` 도입 — lipgloss v2 기반 스택과의 호환 버전을 run-phase pre-flight에서 확정(glamour의 lipgloss 메이저 의존 확인 의무).
- **bubbletea/bubbles v2**: -002/-003이 랜딩한 v2 스택을 소비만 한다(도입 책임 없음).
- **병렬 작업 격리**: 본 SPEC의 커밋 범위는 `internal/cli/{doctor*,status*,spec_view*,spec_status*,root.go}` + `internal/cli/uikit/**` + go.mod/go.sum + 관련 골든/테스트로 한정.
