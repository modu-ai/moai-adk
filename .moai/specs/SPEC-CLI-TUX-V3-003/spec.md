---
id: SPEC-CLI-TUX-V3-003
title: "moai update Decomposition + Change-Preview TUI (CLI TUX v3 — M3)"
version: "0.1.1"
status: draft
created: 2026-07-13
updated: 2026-07-14
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: "internal/cli/update.go → internal/cli/update/* + internal/merge"
lifecycle: spec-anchored
tags: "cli, tux, update, decomposition, preview-tui, bubbletea-v2, bubbles-v2, namespace-protection, three-way-merge, tier-l"
era: V3R6
tier: L
depends_on: [SPEC-CLI-TUX-V3-001]
related_specs: [SPEC-CLI-TUX-V3-002]
---

# SPEC-CLI-TUX-V3-003 — `moai update` 분해 + 변경 프리뷰 TUI (CLI TUX 현대화 M3: U-1~U-4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial draft — CLI TUX 현대화 계획 보고서(`.moai/reports/moai-cli-tux-modernization-plan-20260710.html`) §4 M3 + §5 U-1~U-4로부터 작성 |
| 0.1.1 | 2026-07-14 | manager-spec | plan-audit iter-1 delta-fix — D1: AC-TUX3-017 재앵커(indirect 잔존 허용 + first-party importer 실측 = internal/merge 3파일; 브리프의 "internal/cli 테스트 4파일"은 주석/skip 문자열 grep 노이즈로 정정), D2: §F 게이트 AC 번호 교정(M3c ↔ M3e), D3: progress.md §E.2 deferred-design-artifact 기록 의무 노트, D4: AC-TUX3-005 --stat → full-diff 증거 |

## §A Context

`internal/cli/update.go`는 **단일 파일 122KB / 3,276 LOC**(2026-07-13 실측)로, 7단계 업데이트 파이프라인 + namespace 보호 + 3-way merge 오케스트레이션을 전부 담고 있다. 계획 보고서 §5가 진단한 4개 결함(U-1~U-4):

1. **U-1 (P1)**: 파일 비대화 — 분석/백업/배포/병합/보고 관심사가 한 파일에 혼재. `internal/cli/update/` 서브패키지가 없다.
2. **U-2 (P1)**: 변경 요약이 텍스트 나열 + `merge.ConfirmMerge` 확인 프롬프트(update.go:677, 1038)뿐 — 분류 테이블/diff 프리뷰 UI 부재.
3. **U-3 (P1)**: Already up-to-date / Updated N files / Dry-run 결과가 각기 다른 출력 — 통일된 outcome 카드 부재. 백업 경로·복구 커맨드가 상시 표기되지 않는다.
4. **U-4 (P2)**: namespace 보호(`internal/cli/update_namespace_protect.go` + `isUserOwnedNamespace`, update.go:1387)는 동작하나 무언 — 사용자가 보존 사실을 확인할 수 없다.

**최악 리스크 (계획 §6)**: 분해 중 namespace 보호 회귀 = **사용자 자산 손실**. 기존 가드 테스트군 — `internal/template/split_namespace_test.go`(`TestSplitHarnessNamespaceNoLeak`), `internal/cli/update_namespace_hns_test.go`, `update_namespace_harness_v2_test.go`, `update_namespace_harness_v4_test.go`, `update_security_m2_test.go` — 전량 무수정 green이 본 SPEC의 HARD 제약이다.

**M1 잔여 부채 (의존성 주의)**: SPEC-CLI-TUX-V3-001은 PASS-WITH-DEBT(0.89)로 종결되었고 go.mod에는 `charm.land/lipgloss/v2` + `charm.land/fang/v2`만 랜딩했다. **bubbletea v1.3.10 / huh v1.0.0 / bubbles v1.0.0(indirect)은 여전히 v1이다** — bubbletea/huh/bubbles v1→v2 마이그레이션은 미완이며, 본 SPEC의 charm v2 사용(프리뷰 테이블 = bubbletea v2 + bubbles v2 table/viewport/list)은 이 부채를 해소하거나 명시적으로 우회 범위를 확정해야 한다(§B.7 + §E). `internal/merge/confirm.go`의 bubbletea v1 잔류는 M1이 @MX:DEBT(confirm.go:871, @MX:UPGRADE → 본 SPEC)로 기록했다 — 본 SPEC이 그 승격 트리거다.

본 SPEC은 4-마일스톤 중 **M3(update 재설계)** 만 다룬다. 모든 라인 번호 앵커는 run-phase 착수 시 재실측한다.

## §B Requirements (GEARS)

### B.1 변경 분류 데이터 모델 (설계 결정 — 프리뷰/폴백/보호의 공유 원천)

- **REQ-TUX3-001**: The update analysis result shall classify every affected file into exactly one of four classes — `add` / `update` / `preserve (user-owned)` / `conflict` — expressed as a single shared classification type consumed by BOTH the TUI preview table and the plain-text fallback (one source of truth; no per-surface re-derivation).
- **REQ-TUX3-002**: The `preserve (user-owned)` classification shall be derived from the same namespace-protection predicate that the deploy stage actually enforces (`isUserOwnedNamespace` family) — the preview shall not implement a parallel heuristic.

### B.2 U-1 파일 분해 (behavior-preserving DDD)

- **REQ-TUX3-003**: Before any extraction commit, characterization tests shall capture the current update pipeline's observable behavior (flag matrix × outcome classes × exit codes), and the same tests shall pass unmodified after decomposition.
- **REQ-TUX3-004**: The `internal/cli/update.go` single file shall be decomposed into `internal/cli/update/` subpackages `{plan, backup, deploy, merge, report}`, with the remaining root `update.go` reduced to cobra command wiring + orchestration glue.
- **REQ-TUX3-005**: The namespace-protection logic (`update_namespace_protect.go` + the `isUserOwnedNamespace` predicate family) shall move intact — zero semantic change — and ALL existing namespace guard tests (`TestSplitHarnessNamespaceNoLeak` class, `update_namespace_hns_test.go`, `update_namespace_harness_v2_test.go`, `update_namespace_harness_v4_test.go`, `update_security_m2_test.go`) shall remain green WITHOUT test-body modification (import-path adjustments only, individually documented).
- **REQ-TUX3-006**: The 3-way merge orchestration shall move intact; the `internal/merge` package test suite shall pass unmodified.
- **REQ-TUX3-007**: The `moai update` CLI surface — flag set (`--yes`, `--templates-only`, force/backup/dry-run class flags), flag semantics, and process exit codes — shall be preserved across the decomposition.

### B.3 U-2 변경 프리뷰 TUI

- **REQ-TUX3-008**: When update analysis completes on a TTY without `--yes`, the confirmation surface shall present a Bubble Tea v2 table classifying changes per the REQ-TUX3-001 model, with per-class counts and keyboard navigation.
- **REQ-TUX3-009**: Where a file row is selected in the preview table, a per-file diff preview (viewport) shall be reachable from the table before confirmation.
- **REQ-TUX3-010**: While `--yes` is set or the process runs non-TTY, the preview shall fall back to a plain-text classification summary (same shared model), and shall emit zero ANSI escape sequences under NO_COLOR or piped output.
- **REQ-TUX3-011**: The `internal/merge/confirm.go` checkbox UI shall be promoted from bubbletea v1 to a bubbles v2 list component, reused for update conflict resolution; upon promotion the M1 deferral annotations (`@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` at the `tea.NewProgram` call site) shall be removed, and the confirm test suite shall pass.

### B.4 U-3 outcome 카드 통일

- **REQ-TUX3-012**: When `moai update` terminates in any of the three primary outcomes — Already up-to-date / Updated N files / Dry-run — the result shall render through a single unified outcome-card renderer (one code path, outcome-parameterized).
- **REQ-TUX3-013**: The outcome card shall always display the backup location and the recovery command whenever a backup was created (including dry-run and failure paths where a backup exists).

### B.5 U-4 namespace 보호 가시화

- **REQ-TUX3-014**: The preview (both TUI table and text fallback) shall label user-owned namespace entries as `preserved (user-owned)`, making the silent protection visible.

### B.6 HARD 리스크 제약 (사용자 자산 보호)

- **REQ-TUX3-015**: The decomposition and preview work shall not weaken namespace-protection semantics — the update pipeline shall not delete or overwrite user-owned namespace files under any flag combination that previously preserved them (user-asset loss is the worst-case risk per plan §6).
- **REQ-TUX3-016**: A new integration test shall assert preview-vs-enforcement agreement: every file the preview classifies as `preserve (user-owned)` is actually left byte-identical by the deploy stage (classification ⇔ enforcement coherence).

### B.7 의존성 (M1 잔여 부채)

- **REQ-TUX3-017**: The go.mod dependency set shall include `charm.land/bubbletea/v2` and `charm.land/bubbles/v2` as direct dependencies (resolving the M1 residual debt if SPEC-CLI-TUX-V3-002 has not already landed them); the `github.com/charmbracelet/bubbletea` v1 direct dependency shall be removed once `internal/merge/confirm.go` is promoted (REQ-TUX3-011), unless a remaining v1 consumer is documented in progress.md §E.2.

### B.8 품질 게이트

- **REQ-TUX3-018**: Each new `internal/cli/update/*` subpackage shall achieve >= 85% statement coverage, and the `internal/cli` / `internal/template` / `internal/hook` package coverage shall not regress below the pre-flight baseline.
- **REQ-TUX3-019**: The full repository test suite (`go test ./... -count=1`), `golangci-lint run` (no NEW findings), and cross-platform builds (`GOOS=windows`, `GOOS=linux`) shall pass.
- **REQ-TUX3-020**: The count of direct `fmt.Printf`/`fmt.Println`/`fmt.Print(` calls in non-test `internal/cli` sources shall be strictly lower than the re-measured pre-flight baseline as update-surface call sites migrate to the Printer (ratchet 승계 — REQ-CTX-017).

## §C Out of Scope (Non-Goals)

본 SPEC은 CLI TUX 현대화 계획의 M3만 다룬다.

### Out of Scope — init 표면 (M2 / SPEC-CLI-TUX-V3-002)
- 셀프업데이트 순서 교정(I-1)은 M2 소관 — `runBinaryUpdateStep`/`shouldSkipBinaryUpdate`의 **호출 위치**는 본 SPEC에서 변경하지 않는다(분해 시 파일 이동은 허용, 시맨틱·호출 순서 무변경).
- 위저드·라이브 init 진행률·경고 수집기는 M2 소관.

### Out of Scope — doctor/status/spec 폴리시 (M4 / SPEC-CLI-TUX-V3-004)
- doctor 대시보드, glamour 렌더, 배너 경량화, help 재정리, 전면 골든 매트릭스 갱신은 M4 소관.

### Out of Scope — huh v2 승격
- update 플로우의 huh 사용(`huh.NewConfirm`, update.go:165)은 본 SPEC에서 huh v1 잔류 허용 — huh 메이저 결정은 M2 스파이크 소관이며, 본 SPEC의 프리뷰/컨펌 승격 대상은 bubbletea/bubbles 계열이다.

### Out of Scope — 템플릿 콘텐츠·namespace 정책 변경
- namespace 보호 **정책**(어떤 경로가 user-owned인가)의 변경은 금지 — 본 SPEC은 로직의 무손실 이동 + 가시화만 수행한다. 정책 변경은 별도 SPEC.

### Out of Scope — statusline / internal/tui 커널
- 무접촉. tui 문자열 토큰 소비만.

## §D Non-Functional Constraints

- **Behavior preservation HARD**: 분해는 DDD(characterization-first). 관측 가능 행동(출력 분류·백업 생성·병합 결과·exit code) 변경 금지 — 변경은 표시 계층(프리뷰/카드)에만 허용.
- **가드 테스트 무수정 원칙**: namespace 가드 테스트의 단언(assertion) 수정 금지. import 경로 조정만 허용하며 파일별 사유를 progress.md §E.2에 기재.
- **Windows/CI 안전**: 프리뷰 TUI는 non-TTY/`--yes` 폴백 페어 의무. `GOOS=windows` 빌드 필수 (bubbletea v2 Windows 콘솔 경로 포함).
- **커버리지**: 신규 update 서브패키지 각 ≥ 85% (REQ-TUX3-018); cli/template/hook 비회귀.
- **성능**: 프리뷰 테이블은 파일 수백 건 규모에서 렌더 지연 없이 동작 (분석 결과는 이미 메모리 내 — 추가 I/O 금지).

## §E Dependencies & Interop

- **depends_on**: SPEC-CLI-TUX-V3-001 (completed) — printer/fang/tui v2 + confirm.go @MX:DEBT 기록이 전제. 본 SPEC이 그 @MX:UPGRADE 트리거를 이행한다.
- **M1 잔여 부채**: bubbletea/huh/bubbles v1→v2 마이그레이션 미완 (go.mod 실측: bubbletea v1.3.10 direct, bubbles v1.0.0 indirect, huh v1.0.0 direct). 본 SPEC의 charm v2 사용은 이 부채를 해소하거나(REQ-TUX3-017: bubbletea/bubbles v2 도입 + confirm.go 승격 후 bubbletea v1 제거 시도) 명시적으로 잔존 소비자를 문서화해야 한다.
- **related_specs**: SPEC-CLI-TUX-V3-002 — bubbletea/bubbles v2 도입은 -002/-003 중 선행 착수 SPEC이 랜딩(중복 도입 방지). -002와 실행 순서 제약은 없다(둘 다 -001에만 의존).
- **병렬 작업 격리**: 본 SPEC의 커밋 범위는 `internal/cli/update.go` + 신규 `internal/cli/update/**` + `internal/cli/update_namespace_protect.go`(이동) + `internal/merge/**` + go.mod/go.sum + 관련 테스트로 한정.
