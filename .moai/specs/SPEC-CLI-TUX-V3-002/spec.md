---
id: SPEC-CLI-TUX-V3-002
title: "moai init Redesign — Deferred Self-Update + Unified Wizard + Live Progress (CLI TUX v3 — M2)"
version: "0.1.1"
status: in-progress
created: 2026-07-13
updated: 2026-07-14
author: manager-spec
priority: P0
phase: "v3.0.0 target"
module: "internal/cli (init.go, wizard) + internal/cli/printer"
lifecycle: spec-anchored
tags: "cli, tux, init, wizard, huh-v2, bubbletea-v2, bubbles-v2, spinner, warning-collector, tier-m"
era: V3R6
tier: M
depends_on: [SPEC-CLI-TUX-V3-001]
related_specs: [SPEC-CLI-TUX-V3-003]
---

# SPEC-CLI-TUX-V3-002 — `moai init` 재설계 (CLI TUX 현대화 M2: I-1~I-5)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial draft — CLI TUX 현대화 계획 보고서(`.moai/reports/moai-cli-tux-modernization-plan-20260710.html`) §4 M2 + §5 I-1~I-5로부터 작성 |
| 0.1.1 | 2026-07-14 | manager-spec | plan-audit iter-1 delta-fix — D1: `TestCharacterize_WizardTotalSteps` supersession 승인 행 plan.md §B #11 추가, D2: AC-TUX2-009를 go.mod direct 블록 검사에서 import-사용 증거로 재앵커 |

## §A Context

`moai init`의 사용자 경험에는 계획 보고서 §5가 진단한 5개 결함(I-1~I-5)이 있다:

1. **I-1 (P0)**: `runInit`(init.go:232, content-token 앵커 — 라인 번호는 run-phase 재실측)이 위저드 **이전에** 블로킹 네트워크 셀프업데이트(`runBinaryUpdateStep`)를 실행한다. 첫 인터랙션까지 네트워크 지연이 그대로 노출된다. `shouldSkipBinaryUpdate`(update.go:453)의 스킵 조건은 templates-only 플래그/`EnvSkipBinaryUpdate` env/dev 빌드뿐이다.
2. **I-2 (P0)**: huh v0.8.x YOffset 스크롤 버그 회피용 "질문 1개 = 폼 1개" 워크어라운드(`wizard.go:14-15` 주석)가 잔존하고, 스텝퍼 분모가 `wizardTotalSteps = 6`(wizard.go:99) 하드코딩이다 — 감사 실측으로 실제 표시 질문 수는 모드에 따라 7~9개까지 변동하며, `TotalVisibleQuestions` 동적 경로(wizard.go:76)와 하드코딩 경로(wizard.go:126)가 공존한다.
3. **I-3 (P0)**: PhaseExecutor 템플릿 배포가 정적 줄 출력만 제공한다 — 긴 단계가 무반응처럼 보인다. M1이 도입한 `internal/cli/printer`의 `Spinner`/`Progress` 핸들은 현재 stateless 단일 프레임(`spinnerFrame = "⠋"`, printer.go:241)이라 라이브 애니메이션이 없다.
4. **I-4 (P1)**: 경고가 단계별로 산발 출력된다(M1이 stderr 채널 규율은 잡았으나 수집·요약은 미해결 — init.go:509-510의 `result.Warnings` 나열이 유일한 집계점).
5. **I-5 (P2)**: 기존 프로젝트에 재실행하면 `--force` 필요 에러만 안내한다 — `moai update` 리다이렉트 힌트가 없다.

**M1 잔여 부채 (의존성 주의)**: SPEC-CLI-TUX-V3-001은 PASS-WITH-DEBT(0.89)로 종결되었고 go.mod에는 `charm.land/lipgloss/v2 v2.0.5` + `charm.land/fang/v2 v2.0.1`만 랜딩했다. **bubbletea v1.3.10 / huh v1.0.0 / bubbles v1.0.0(indirect)은 여전히 v1이다** — bubbletea/huh/bubbles v1→v2 마이그레이션은 미완이며, 본 SPEC의 charm v2 사용(라이브 스피너 = bubbles v2, 통합 폼 = huh v2)은 이 부채를 해소하거나 명시적으로 우회 범위를 확정해야 한다(§B.2/§B.3 + §E).

본 SPEC은 4-마일스톤 중 **M2(init 재설계)** 만 다룬다. 모든 라인 번호 앵커는 run-phase 착수 시 재실측한다.

## §B Requirements (GEARS)

### B.1 셀프업데이트 순서 교정 (I-1, P0)

- **REQ-TUX2-001**: When `moai init` runs in interactive (wizard) mode, the process shall reach the first wizard interaction without any prior network call — the binary self-update check (`runBinaryUpdateStep`) shall not execute before wizard completion.
- **REQ-TUX2-002**: Where the deferred self-update check detects a newer binary after wizard completion, the result shall surface as a non-blocking stderr notice (including the `moai update` follow-up hint); the check shall not re-exec the process after wizard answers have been collected.
- **REQ-TUX2-003**: While any existing skip condition holds (`templates-only`-class flag, `EnvSkipBinaryUpdate` env guard, dev-build version detection per `shouldSkipBinaryUpdate`), the deferred self-update check shall remain skipped — skip semantics are behavior-preserving.
- **REQ-TUX2-004**: While `moai init` runs non-interactively (all wizard inputs supplied by flags, or non-TTY), the deferred check shall preserve the same "no network before first output" property and shall not block phase execution on network I/O.

### B.2 huh v2 스파이크 + 통합 위저드 (I-2, P0)

- **REQ-TUX2-005**: The huh v2 compatibility spike shall execute as the first milestone, and its verdict (YOffset scroll defect resolved / persists under the huh v2 + bubbletea v2 pair) shall be recorded in progress.md §E.2 with reproduction evidence before any wizard-restructuring commit lands.
- **REQ-TUX2-006**: Where the spike confirms the scroll defect is resolved, the init wizard shall present a single multi-group form (2-3 field groups + stepper) replacing the one-question-one-form workaround.
- **REQ-TUX2-007**: Where the spike fails (plan B), the wizard shall retain the current per-question form structure with v2-styled theming only, and the plan-B decision shall be recorded in progress.md §E.2 — wizard behavior (question set, defaults, validation) stays byte-compatible.
- **REQ-TUX2-008**: The wizard stepper denominator shall be derived from the actual visible question count (single dynamic source); the hardcoded `wizardTotalSteps = 6` constant shall be removed from non-test wizard sources.
- **REQ-TUX2-009**: The wizard shall preserve locale rendering (`RunWithLocale`) and standard/advanced mode question-visibility semantics — existing wizard test suite passes, or each legitimately changed golden is individually documented in progress.md §E.2.

### B.3 라이브 진행 피드백 (I-3, P0)

- **REQ-TUX2-010**: While PhaseExecutor deploys templates on a TTY, the progress surface shall render an animated spinner with a live file-count progress readout (e.g. `42/96 files`), driven through the `internal/cli/printer` Spinner/Progress handles.
- **REQ-TUX2-011**: While the process runs non-TTY, or with `NO_COLOR` set, or with `MOAI_REDUCED_MOTION` set, the progress surface shall degrade to the existing plain line output — zero ANSI escape sequences and zero animation frames on either channel.
- **REQ-TUX2-012**: The go.mod dependency set shall include `charm.land/bubbletea/v2` and `charm.land/bubbles/v2` as direct dependencies backing the animated spinner (M1 residual-debt resolution); the huh major version follows the spike verdict (v2 on success; v1.0.0 retained under plan B with the retention decision documented).

### B.4 경고 수집기 (I-4, P1)

- **REQ-TUX2-013**: When `moai init` terminates (success or failure), all warnings accumulated during the run shall be re-emitted exactly once as a single consolidated stderr summary panel carrying the warning count and each message.
- **REQ-TUX2-014**: The init command shall not write any warning text to stdout — stdout carries data only (channel discipline inherited from REQ-CTX-012/016).

### B.5 기존 프로젝트 리다이렉트 힌트 (I-5, P2)

- **REQ-TUX2-015**: When `moai init` targets a directory already containing an initialized MoAI project and `--force` is absent, the error surface shall include a redirect hint naming `moai update` as the likely intended command.

### B.6 완료 카드

- **REQ-TUX2-016**: When `moai init` succeeds, the completion card shall display the next-action sequence (`cd <project>` → `moai cc` → `/moai plan`) and, where warnings were collected, a one-line pointer to the stderr warning summary.

### B.7 품질 게이트

- **REQ-TUX2-017**: The full repository test suite (`go test ./... -count=1`) and `golangci-lint run` shall pass with no NEW findings relative to the pre-flight baseline, and cross-platform builds (`GOOS=windows`, `GOOS=linux`) shall succeed.
- **REQ-TUX2-018**: The count of direct `fmt.Printf`/`fmt.Println`/`fmt.Print(` calls in non-test `internal/cli` sources shall be strictly lower than the re-measured pre-flight baseline (2026-07-13 실측 40) as init/wizard call sites migrate to the Printer (ratchet 승계 — REQ-CTX-017).

## §C Out of Scope (Non-Goals)

본 SPEC은 CLI TUX 현대화 계획의 M2만 다룬다.

### Out of Scope — update 표면 (M3 / SPEC-CLI-TUX-V3-003)
- `update.go` 분해, 변경 프리뷰 TUI, outcome 카드, namespace 보호 가시화는 전부 M3 소관.
- `internal/merge/confirm.go`의 bubbles v2 list 승격(@MX:DEBT 해소)도 M3 소관 — 본 SPEC은 confirm.go 무접촉.

### Out of Scope — doctor/status/spec 폴리시 (M4 / SPEC-CLI-TUX-V3-004)
- doctor 라이브 대시보드, glamour 렌더, 배너 경량화, help 그룹 재정리, NO_COLOR/non-TTY/Windows 골든 매트릭스 전면 갱신은 M4 소관.
- 배너(uikit/banner.go 12곳 fmt.Print*)는 본 SPEC 무접촉 — ratchet 감소는 init/wizard 표면에서만 발생.

### Out of Scope — internal/tui 커널 재포팅
- lipgloss v2 포팅은 M1에서 완료. 본 SPEC은 tui 공개 문자열 토큰 계약(Theme/Catppuccin*)을 소비만 하고 수정하지 않는다.

### Out of Scope — lipgloss v1 완전 제거
- huh 스파이크가 plan B로 귀결되면 huh v1.0.0 + lipgloss v1 간접 그래프가 잔존한다. v1 완전 제거는 본 SPEC의 요구가 아니다(스파이크 성공 시의 부수 효과로만 발생 가능).

### Out of Scope — statusline 시각 언어
- `internal/statusline`은 본 SPEC 무접촉.

## §D Non-Functional Constraints

- **Behavior preservation 기본**: 위저드 질문 집합·기본값·검증 로직·결과 스키마(`WizardResult`)는 구조 개편과 무관하게 보존. 변경되는 것은 표시 계층뿐.
- **채널 규율**: stdout=데이터, stderr=상태/경고 (M1 계약 승계). 골든 테스트로 기계 검증.
- **Windows/CI 안전**: 모든 TTY 신기능은 non-TTY plain 폴백을 페어로 가진다. `GOOS=windows` 빌드 필수.
- **네트워크 0 원칙**: 첫 위저드 인터랙션 전 네트워크 호출 0회(계획 §7 성공 지표). 셀프업데이트는 지연·비차단.
- **스파이크 게이트**: I-2 통합 폼 작업은 스파이크 verdict 기록 전 착수 금지 — plan B 경로가 문서화된 탈출구다.

## §E Dependencies & Interop

- **depends_on**: SPEC-CLI-TUX-V3-001 (completed) — printer 패키지·fang root·tui v2 포팅이 전제.
- **M1 잔여 부채**: bubbletea/huh/bubbles v1→v2 마이그레이션 미완 (go.mod 실측: bubbletea v1.3.10, huh v1.0.0, bubbles v1.0.0 indirect). 본 SPEC의 charm v2 사용은 이 부채를 해소하거나(REQ-TUX2-012) 명시적으로 우회 범위를 확정해야 한다(huh는 스파이크 verdict에 종속).
- **related_specs**: SPEC-CLI-TUX-V3-003 — bubbletea/bubbles v2 도입을 -002/-003 중 먼저 착수하는 쪽이 랜딩한다(둘 다 필요). 본 SPEC이 선행하면 -003은 의존성 도입 단계를 스킵한다.
- **병렬 작업 격리**: 본 SPEC의 커밋 범위는 `internal/cli/{init.go,wizard/,printer/}` + go.mod/go.sum + 관련 테스트로 한정.
