# M1 잔여 결함표 — 카드 6판정 × PTY 캡처 실측 (REQ-TRI-001)

- 측정 시점: run-phase M1 (수리 편집 0 상태)
- 측정 대상 트리: `WT-tux-render` @ `7ad734f38` (base `a404132e7`)
- 판정 방식: PTY 캡처 프레임 (tmux 80x30, `internal/cli/ptycaptest`, `MOAI_PTY_CAPTURE=1`)
- 프레임 증거: 본 디렉터리 `baseline-*.txt` 8매 (S1-S4 × en/ko)
- 측정 명령:
  `MOAI_PTY_CAPTURE=1 MOAI_PTY_CAPTURE_OUT=<spec>/evidence/baseline go test ./internal/cli/wizard/ -run 'TestPtyCapture_BaselineSurfaces|TestPtyCapture_NormalRun' -count=1 -v`
- 관측 출력 (발췈): `CONFIRM-INTERNAL-BLANK-ROWS confirm-fixture-en = 1`, `confirm-fixture-ko = 1`,
  `downgrade-confirm-en = 1`, `downgrade-confirm-ko = 1` — 전 confirm 표면 동일값.

## AC-TRI-003 전제값 조정 (observe-first 구조 흡수)

acceptance.md AC-TRI-003 이 인용한 "수리 전 = 2" 는 카드 판정문(2026-09-09 감사)의 값을 따랐으나,
본 M1 이 huh v2.0.3 + 현행 moaiWizardTheme 트리에서 실측한 값은 **1** 이다.
huh `field_confirm.go:261-263` 의 고정 `"\n\n"` 은 소스상 그대로 존재하지만(주석 참조),
이 두 개행이 만드는 가시 빈 행은 정확히 1개다(설명 줄 뒤 개행 1 + 빈 줄 1).
프레임 실측(`baseline-downgrade-confirm-en.txt` 2-5행: 제목/설명/빈 1행/버튼)이 권위다.
REQ-TRI-003 의 합격 기준(빈 행 ≤ 1)은 **이미 충족** — REQ-TRI-008 에 따라 수리 편집 없이 종결하고,
측정값 1 을 골드로 고정하는 회귀 가드를 둔다.

## 잔여 표

| # | 카드 판정 | M1 실측 (프레임 근거) | 상태 |
|---|---|---|---|
| 1 | F11-(1) 예/아니오 버튼 중앙정렬 | en/ko confirm 프레임 전부 버튼 라벨이 설명 첫 열과 동일 표시 열에서 시작 (`baseline-confirm-fixture-en.txt` 4↔6행, `baseline-downgrade-confirm-*.txt`) | **검증만 종결** (t586 수리 유지) — REQ-TRI-002 소스 스윕 가드로 재발 방지 (M2) |
| 2 | F11-(2) 확인 필드 내부 빈 행 | 실측 **1행** (전 confirm 표면) — ≤1 기준 충족, "수리 전 = 2" 전제는 실측 1로 조정 | **검증만 종결** (REQ-TRI-008) — 빈 행 ≤1 캡처 회귀 가드로 고정 (M2) |
| 3 | F11-(3) 선택 항목 설명 열 폭 | init/profile ko·en 프레임에서 옵션 4줄의 " - " 설명 시작 표시 열 동일 (`TestOptionDescriptionColumn_DisplayWidthAligned` 동반 PASS) | **검증만 종결** (t586 수리 유지) — 기존 View 검사 + 프레임로 비-회귀 (REQ-TRI-005) |
| 4 | F10-(1) init/프로필 프롬프트 영어 고정 | ko 프레임에 번역 표 ko 항목 문자열의 영어 원문 없음 — **예외 1건 발견**: acceptEdits stdout 고지문(`internal/cli/profile_setup.go:31`)은 여전히 영어 고정 (프레임 백로그, S5) | **수리 대상** — M3: 앵커 토큰 보존 현지화 (REQ-TRI-006/AC-TRI-007) |
| 5 | F10-(2) 버튼 Yes/No 고정 | ko 프레임 버튼 라벨 `예`/`아니오`, help 행도 `y 예 • n 아니오` 현지화 (`baseline-downgrade-confirm-ko.txt`) | **검증만 종결** |
| 6 | F10-(3) v1 위저드 처분 | huh v1 import 0건 (`go.mod`+import grep, plan §D 3행) — D1(v2 흡수 유지) 확인 | **검증만 종결** — huh v1 비-회귀 가드 (M3, REQ-TRI-007) |

## M2/M3 범위 확정 (plan.md M1 산출물 계약)

- **수리 편집은 행 4(acceptEdits 고지문) 1건뿐이다.** 나머지 5행은 수리 편집 0으로 종결(REQ-TRI-008).
- M2: REQ-TRI-002 정렬 소스 스윕 가드(뮤턴트 관측) + REQ-TRI-003 빈 행 ≤1 캡처 가드(실측 1 고정) + REQ-TRI-004/005 비-회귀 재검증.
- M3: 행 4 현지화 + AC-TRI-006 ko 프레임 잔존 0 검사 + REQ-TRI-007 huh v1 가드(뮤턴트 관측).
- M4: 종결 기록 + progress.md §E.2/§E.3.
