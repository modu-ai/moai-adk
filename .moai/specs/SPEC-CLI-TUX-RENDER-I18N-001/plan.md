# SPEC-CLI-TUX-RENDER-I18N-001 — plan.md

> 카드 t756 · Tier M · Class C. plan 상태: audit-ready 지향(작성 시점 draft 0.1.0).

## §A Context

- 베이스: a404132e7(WT-tux-render 분기점, origin/develop tip + t700 병합). 로컬 develop(ba60eb6d5)이 앞서 있지만 흡수는 통합 창에서 하므로 본 계획은 베이스 그대로 판다.
- 선행 관계: SPEC-INIT-TUX-I18N-001(t586, completed, 본 베이스에 병합됨)이 카드 판정 6건 중 5건을 사실상 전달했다. 본 SPEC 은 그 위에 얹는 잔여 수리 + 가드 + 재검증이다(상세 대응표: spec.md §A.1).
- 검증 자산: `internal/cli/ptycaptest`(하니스), `internal/cli/wizard/ptycap_test.go`(tmux 캡처 케이스), `internal/cli/wizard/layout_alignment_test.go`(View 렌더 기반 정렬·간격 검사 — tmux 불필요, plan 단계에서 실행해 통과 확인).

## §B Known Issues — 카드 판정의 현재 트리 재고정표

카드의 행 참조는 2026-09-09 감사 기준으로 스테일하다. 아래로 재고정한다. SPEC 요구는 이 표의 "현재 기호"를 직접 인용하지 않고 관측 가능한 동작으로 서술했다(spec.md §C).

| 카드 참조 | 현재 트리 위치 / 기호 | 상태 판정 |
|---|---|---|
| `huh field_confirm.go:64,267-277` (버튼 중앙정렬) | huh v2.0.3 `field_confirm.go:53`(기본 `lipgloss.Center`), `:361`(`WithButtonAlignment`) | 라이브러리 기본값은 여전히 중앙 — 우리 쪽 2 생성 지점은 모두 좌측 지정 완료(`internal/cli/wizard/wizard.go:555`, `internal/cli/wizard/downgrade_confirm.go:31`) |
| `huh field_confirm.go:260-265` (확인 내부 빈줄) | huh v2.0.3 `field_confirm.go:261-263` — 제목 뒤 고정 `"\n\n"`(`!c.inline`일 때) | **잔여 결함** — 테마로 제거 불가. 기본 수리 경로는 로컬 확인 필드 래퍼(huh.Field 구현, 자체 View 조합); `Confirm.Inline(bool)`(`:136`)은 구조적 부적합(`:254` 제목↔설명 개행 제거, `:293` 버튼 앞 개행 부재)이라 M1 측정 참고로만 사용 |
| `huh_theme.go:95-100` (Yes/No 고정) | 소멸 — 테마는 `internal/cli/wizard/wizard.go` `newMoAIWizardTheme/moaiWizardStyles`, 버튼 라벨은 로케일 표(`ConfirmYes/ConfirmNo`)로 해석 | 수리됨(t586) |
| `wizard.go:549` (제목 MarginBottom) | `internal/cli/wizard/wizard.go:645`(`FieldSeparator` 개행 축소, t586 REQ-ITI-015) + `internal/cli/wizard/styles.go:87`(내부 `Styles.Title` — huh 테마와 별개 자산) | 필드 사이 간격 수리됨; 확인 필드 내부 간격은 위 행의 잔여 |
| `wizard.go:288-292` (선택 폭) | `internal/cli/wizard/wizard.go:317` `alignOptionLabels`(표시 폭 패딩) + `:299` `optionColumnWidthCap` | 수리됨(t586 AC-ITI-017) |
| `init.go:601-602` (영어 프로필 confirm) | 소멸 — init 은 프로필 확인 질문을 묻지 않음(t586 REQ-ITI-001) | 수리됨(제거) |
| `profile_setup.go:330` (구형 v1 위저드 1단계) | `internal/cli/profile_setup.go:330 부근` — v2 위저드 흡수 실행(`profileWizardRunner`), 로케일 해석 텍스트 | 수리됨(흡수) |
| (신규 발견) 영어 고정 고지문 | `internal/cli/profile_setup.go:31` `acceptEditsConfirmationLine` — stderr 고지, grep 앵커 토큰 계약(소유 SPEC: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 의 REQ-CCI-006/AC-CCI-006) | REQ-TRI-006 대상 — 현지화 시 앵커 토큰 보존 |

## §C Pre-flight (M1 착수 조건)

1. `go test ./internal/cli/wizard/ -run 'TestLayout_NoBlankBetweenFields' -count=1` — plan 단계 기준 통과 확인(이미 수행, progress.md §D).
2. tmux 존재 확인 + `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run '^TestPtyCapture_' -count=1` — 기존 캡처 경로가 이 트리에서 작동함을 확인.
3. 대화형 표면 목록화: 위저드 그룹·다운그레이드 확인창·프로필 확인 외에 `moai init/update/profile` 경로의 확인·선택 표면을 전수 나열(REQ-TRI-001 잔여 표의 행 집합이 된다).

## §D Constraints

- plan 전제: 구현 편집 금지(본 plan 단계), `.moai/reports/` 쓰기 금지 — 증거는 `.moai/specs/SPEC-CLI-TUX-RENDER-I18N-001/` 아래에만.
- 확인 필드 인라인 모드 전환은 렌더 3면(제목/설명/버튼)의 표시 열을 모두 바꿀 수 있으므로, M2 수리 시 기존 골든을 함께 갱신하고 REQ-TRI-004/005 비-회귀 검사를 같은 커밋에서 돌린다.
- 카드 [HARD] 준수: 모든 합격 판정은 캡처 프레임 기준. 마크업 단위 검사는 도달성 전제로만.

## §E Self-Verification

run-phase 종료 시 `progress.md` §E.2/§E.3 에 증거와 감사 준비 신호를 남긴다. 본 plan 단계의 검증 기록은 `progress.md` §D 에 있다.

## §F Milestones (우선순위 순 — 변경 가능성 높은 결정이 앞)

### M1 — 기준 캡처와 잔여 결함표 (Priority High)

- 표면 전수 캡처(en/ko × 80열), 카드 6판정 대응 잔여 표 산출, 증거를 SPEC 디렉터리에 내보내기. REQ-TRI-001, REQ-TRI-008.
- 산출물: 잔여 표(행마다 상태 = 수리 대상 / 검증만 종결).
- **이 표가 M2/M3 범위를 확정한다.** 캡처가 적합을 보이는 항목은 수리하지 않는다.

### M2 — 렌더 잔여 수리 + 재발 방지 가드 (Priority High)

- 확인 필드 내부 간격 축소(REQ-TRI-003) — **기본 경로: 로컬 확인 필드 래퍼**(huh.Field 를 구현한 우리 쪽 타입, 자체 View 조합으로 제목·설명과 버튼 줄 사이 간격 제어). 라이브러리 `Inline(true)` 는 구조적으로 부적합(huh `field_confirm.go:254` 제목↔설명 개행 제거, `:293` 버튼 줄 앞 개행 부재 — 제목·설명·버튼이 한 줄로 합쳐짐)이므로 M1 측정 참고로만 관찰하고 M2 구현 경로에서 제외한다. 캡처 프레임으로 빈 행 ≤1 확인.
- 확인 생성 지점 정렬 소스 스윕 가드(REQ-TRI-002) — 뮤턴트(정렬 지정 제거)에서 실패함을 관측.
- REQ-TRI-004/005 비-회귀 재검증 + 골든 갱신.

### M3 — i18n 잔여 훑기 (Priority Medium)

- M1 잔여 표의 미번역 표면을 번역 표로 해석(REQ-TRI-006). ko 캡처 프레임에 영어 원문 잔존 0 확인.
- 앵커 토큰 계약 문자열(acceptEdits 고지 등)은 토큰 보존 현지화 + 기존 앵커 검사 갱신.
- huh v1 비-회귀 가드(REQ-TRI-007) — 뮤턴트(import 추가)에서 실패함을 관측.

### M4 — 종결 기록 (Priority Low)

- 검증만 종결 항목의 증거 정리(REQ-TRI-008), golden 최종 갱신, progress.md §E.2/§E.3 작성.

## §G Anti-Patterns

- 마크업 단언만으로 렌더 AC 를 PASS 처리(공허한 초록 — 카드 [HARD] 위반).
- 캡처가 적합을 보여주는 코드를 "개선" 명목으로 고치는 것(REQ-TRI-008 위반).
- huh upstream 패치나 fork로 잔여 간격을 제거하는 것(spec.md §D 제약 위반).
- tmux 게이트 없는 캡처 검사가 조용히 통과하는 것(건너뜀이 관측 가능해야 한다).

## §H Cross-References

- 선행: SPEC-INIT-TUX-I18N-001(completed — 본 SPEC 의 기준선), SPEC-CLI-WIZARD-RESTRUCTURE-001(v2 위저드 구조).
- 감사 근거: `.moai/reports/init-tui-audit-20260909.html` §1(primary 체크아웃, 읽기전용).
- huh v2.0.3 소스: 모듈 캐시 `charm.land/huh/v2@v2.0.3/field_confirm.go`(:53 기본 정렬, :136 Inline, :261-263 고정 빈 줄, :361 WithButtonAlignment).
- 결정 항목 D1(v1 처분 = v2 흡수 유지 권고): spec.md §A.2 — Kickoff 게이트에서 최종 확인.
