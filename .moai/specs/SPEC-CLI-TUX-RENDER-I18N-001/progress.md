---
spec: SPEC-CLI-TUX-RENDER-I18N-001
card: t756
phase: plan
plan_status: audit-ready (draft 0.1.0)
updated: 2026-09-14
---

# progress — SPEC-CLI-TUX-RENDER-I18N-001

## §A Plan-phase 산출물

- `spec.md` 0.1.0 (draft) — GEARS REQ-TRI-001~008, Out of Scope 3건
- `plan.md` — 재고정표(§B), Pre-flight(§C), M1-M4
- `acceptance.md` — AC-TRI-001~009 (전 AC 캡처 프레임 합격 기준)

## §B Plan-phase 결정 기록

- D1 (카드 위임 결정 "v1 흡수 vs 최소화"): **흡수 유지 권고** — t586 이 흡수를 완료했고 v1 코드·의존이 트리에 없음(§D 항목 3·4). Kickoff 게이트에서 최종 확인 대상.
- D2: 카드 6판정 중 5건이 선행 SPEC-INIT-TUX-I18N-001(completed, 베이스 병합됨)로 전달됨 → 본 SPEC 을 잔여 수리 + 가드 + 재검증으로 범위 확정. REQ-TRI-008(검증만 종결)로 과잉 수리 방지.
- D3: 확인 필드 내부 고정 빈 줄의 기본 수리 경로는 로컬 확인 필드 래퍼(huh.Field 구현, 자체 View 조합)로 판독. 라이브러리 인라인 모드(huh v2.0.3 `field_confirm.go:136`)는 구조적 부적합(`:254` 제목↔설명 개행 제거, `:293` 버튼 앞 개행 부재)이라 M1 측정 참고로만 관찰 — plan-audit 1차 D1 반영.

## §C Plan-phase 조사 기록 (file:line)

- 버튼 정렬: huh v2.0.3 `field_confirm.go:53`(기본 Center), `:361`(WithButtonAlignment). 우리 생성 지점 2곳 모두 좌측 지정: `internal/cli/wizard/wizard.go:555`, `internal/cli/wizard/downgrade_confirm.go:31`.
- 확인 내부 빈 줄: huh v2.0.3 `field_confirm.go:261-263` — `!c.inline`일 때 제목 뒤 `"\n\n"` 고정 출력(테마로 제거 불가 → REQ-TRI-003).
- 필드 사이 간격: `internal/cli/wizard/wizard.go:645` — `FieldSeparator` 개행 축소(t586 REQ-ITI-015).
- 선택 폭: `internal/cli/wizard/wizard.go:299`(`optionColumnWidthCap`), `:317`(`alignOptionLabels`, 표시 폭 패딩).
- init 프로필 confirm 소멸: `internal/cli/init.go:617` 주석(REQ-ITI-001 — 프로필 질문 미포함).
- v1 흡수: `internal/cli/profile_setup.go:330` 부근(v2 위저드 실행, 로케일 해석).
- huh v1 부재: `go.mod`에 v1 무, `internal/`·`cmd/` 에서 v1 import grep 0건.
- 신규 발견(영어 고정 고지문): `internal/cli/profile_setup.go:31` `acceptEditsConfirmationLine` — grep 앵커 토큰 계약(REQ-CCI-006) → REQ-TRI-006/AC-TRI-007.
- 캡처 자산: `internal/cli/wizard/ptycap_test.go`(tmux 케이스), `layout_alignment_test.go`(View 렌더 검사), `internal/cli/ptycaptest/`(하니스).

## §D Plan-phase Evidence (측정 기록)

| # | Claim | Command | Observed output |
|---|---|---|---|
| 1 | 렌더 정렬·간격 기존 검사가 이 트리에서 통과 | `go test ./internal/cli/wizard/ -run 'TestLayout_NoBlankBetweenFields' -count=1` (plan-audit 1차 -v 관측: 해당 셀렉터 목록과 일치하는 검사는 이 1건) | `ok github.com/modu-ai/moai-adk/internal/cli/wizard 0.607s` |
| 2 | 확인 생성 지점 2곳 + 모두 좌측 정렬 | `grep -rn "huh.NewConfirm" internal/ ; grep -rn "WithButtonAlignment" internal/ \| grep -v _test` | 2 지점(`wizard.go:544`, `downgrade_confirm.go:26`) / 2 정렬 지정(`wizard.go:555`, `downgrade_confirm.go:31`) |
| 3 | huh v1 잔존 없음 | `grep -rn "charmbracelet/huh\"" internal/ cmd/ ; grep -n "charmbracelet/huh" go.mod` | 양쪽 모두 0건(빈 출력) |
| 4 | t586 완료·본 베이스 포함 | `.moai/specs/SPEC-INIT-TUX-I18N-001/spec.md` frontmatter 판독 | `status: completed`, `updated: 2026-09-13` |
| 5 | SPEC ID 규격·고유성 | 정규식 셀프체크 + specs 디렉터리 조회 | `PASS` / 기존 ID 충돌 없음 |
| 6 | 라이브러리 기본 정렬·고정 빈 줄 존재 | huh v2.0.3 `field_confirm.go` 판독 | `:53 buttonAlignment: lipgloss.Center`, `:261-263` `"\n\n"` 출력부, `:136 Inline`, `:361 WithButtonAlignment` |

## §E 감사 신호

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-CLI-TUX-RENDER-I18N-001
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 8
ac_count: 9
gears_compliance: all REQ use GEARS patterns (Event/State/Ubiquitous), no IF/THEN
out_of_scope: 3 H3 sections with bullets (question-structure overhaul, huh upstream patch, new locale content)
capture_gate: every render/i18n AC judged on PTY frames (card HARD clause honored)
stale_ref_resolution: all card refs re-anchored; table in plan.md §B
blockers: none
needs_clarification_markers: none
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
