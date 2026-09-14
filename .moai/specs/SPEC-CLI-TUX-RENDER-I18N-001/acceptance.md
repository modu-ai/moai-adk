# SPEC-CLI-TUX-RENDER-I18N-001 — acceptance.md

> 카드 t756 · Tier M. 모든 렌더/i18n AC 의 합격 증거는 PTY 캡처 프레임이다(카드 [HARD]). 마크업·내부 구조 단언은 프레임 도달성의 전제로만 쓴다.

## §A 품질 게이트

- TRUST 5 준수. 캡처 검사는 `internal/cli/ptycaptest` 게이트(tmux + `MOAI_PTY_CAPTURE=1`) 아래에서 실행하며, 게이트 부재 시 건너뜀이 관측 가능해야 한다(선행 SPEC AC-ITI-019 계약 승계).
- 프레임 계약: 80열, ANSI 제거 후 판정, 표시 열 계산은 동아시아 전각 2칸(`go-runewidth` 관례 승계).
- 각 수리 AC 는 (1) 수리 전 프레임에서 결함이 관측됨(양성 전제), (2) 수리 후 프레임에서 기준 충족, (3) 골든 고정 — 세 단계를 남긴다.

## §C 검증 대상 표면

| id | 표면 | 픽스처/케이스 |
|---|---|---|
| S1 | init 위저드 첫 페이지(en/ko) | 기존 `init-first-page` 캡처 케이스 승계 |
| S2 | init/update 확인형 질문(그룹 내 confirm) | `Question{Type: confirm}` 픽스처를 `buildUnifiedForm` 으로 렌더 |
| S3 | 다운그레이드 확인창(en/ko) | 확인창 생성 헬퍼 직접 렌더 |
| S4 | 프로필 위저드 확인·선택 그룹(ko) | 실제 프로필 질문 세트 |
| S5 | acceptEdits 정규화 고지문(stderr) | 고지 출력 함수 직접 호출 |
| S6 | M1 조사에서 밝혀지는 추가 대화형 표면 | M1 잔여 표에 행 추가 후 캡처 케이스 등록 |

## §D AC 매트릭스

| AC | 표면 | 측정 | 합격 기준 |
|---|---|---|---|
| AC-TRI-001 | S1-S6 | PTY 캡처(en·ko × 80열) | 표면마다 기준 프레임 존재 + 프레임마다 기준 문자열 도달성 전제 통과 + 잔여 결함표 산출 |
| AC-TRI-002 | 전 confirm 생성 지점 | 소스 스윕 가드 검사 | 모든 확인 필드 생성 지점이 좌측 정렬 명시; 정렬 누락 뮤턴트에서 FAIL |
| AC-TRI-003 | S2·S3·S4 | 캡처 프레임 | 확인 필드 제목 줄과 버튼 줄 사이 빈 행 ≤ 1 (수리 전 = 2) |
| AC-TRI-004 | S1·S4 | 캡처 프레임 | 연속 필드 사이 빈 줄 0, 선택 필드 아래 빈 카드 줄 0 (선행 기준 유지) |
| AC-TRI-005 | S1·S4 | 캡처 프레임 | 같은 선택 목록의 옵션 줄 4개 설명 시작 표시 열 동일 |
| AC-TRI-006 | S1-S4·S6 | ko 캡처 프레임 | 번역 표에 ko 항목이 있는 문자열의 영어 원문 잔존 0 |
| AC-TRI-007 | S5 | 고지 출력 + 앵커 검사 | 현지화된 고지문이 앵커 토큰(`acceptEdits`, `settings.local.json`) 보존 |
| AC-TRI-008 | 패키지 의존 | 의존 스윕 가드 검사 | huh v1 import 존재 뮤턴트에서 FAIL, 정상 트리 PASS |
| AC-TRI-009 | 잔여 표 전체 | M1 기준 프레임 | 적합 확인된 카드 판정 항목이 수리 편집 0으로 "검증만 종결" 기록 + 프레임 증거 |

## §E Given-When-Then

- **AC-TRI-001** **Given** tmux 가 있는 환경과 §C 표면 목록, **When** `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run '^TestPtyCapture_' -count=1` (신규 표면 케이스 포함)으로 캡처하면, **Then** 표면×로케일 조각마다 내보낸 프레임 파일이 존재하고 각 프레임이 기준 문자열(예: S1의 `Select conversation language` 와 옵션 줄)을 포함함이 먼저 단정되며, 프레임들이 `.moai/specs/SPEC-CLI-TUX-RENDER-I18N-001/evidence/` 아래 남고 잔여 결함표(카드 6판정 × 상태)가 작성돼 있다.
- **AC-TRI-002** **Given** wizard 패키지 소스, **When** 확인 필드 생성 지점 소스 스윕 가드 검사(예: `go test ./internal/cli/wizard/ -run 'TestConfirmAlignmentSweep' -count=1`)를 돌리면, **Then** 모든 지점이 좌측 정렬 명시를 가져 PASS 하고, 임의 지점의 정렬 지정을 제거한 뮤턴트에서 FAIL 이 관측돼 있다.
- **AC-TRI-003** **Given** S2·S3·S4 확인 표면의 수리 전 기준 프레임, **When** 프레임에서 확인 필드 제목 줄 다음 버튼 줄 사이의 빈 행 수를 세면, **Then** 수리 전 프레임에서 2가 관측되고(결함 재현), 수리 후 프레임에서 1 이하이며 골든이 이를 고정한다.
- **AC-TRI-004** **Given** S1·S4 프레임, **When** 연속 필드 사이 빈 줄과 선택 필드 아래 빈 카드 줄(`┃` 뒤 공백뿐인 행)을 세면, **Then** 각각 0 이다(선행 SPEC 기준 감쇠 없음).
- **AC-TRI-005** **Given** S1·S4의 대화 언어 선택 옵션 줄 4개, **When** 각 줄의 설명 시작 위치를 표시 폭으로 계산하면, **Then** 네 값이 동일하다.
- **AC-TRI-006** **Given** ko 로케일로 강제한 S1-S4·S6 프레임과 번역 표의 ko 항목 목록, **When** 프레임에서 각 ko 항목에 대응하는 영어 원문을 찾으면, **Then** 일치가 0 이다. (M1 잔여 표의 i18n 행 전체가 이 검사로 닫힌다.)
- **AC-TRI-007** **Given** acceptEdits 정규화 고지의 현지화 출력, **When** 기존 앵커 검사(`go test ./internal/cli/ -run 'TestEmitAcceptEditsConfirmationAnchor' -count=1` — 위치는 run 단계에서 확정)를 돌리면, **Then** 앵커 토큰 2종이 현지화 문자열 안에 보존돼 검사가 PASS 하고, ko 출력에 고지문의 번역문이 나온다.
- **AC-TRI-008** **Given** init/update/profile 표면 패키지들의 의존 그래프, **When** huh v1 비-회귀 가드 검사를 돌리면, **Then** 정상 트리에서 PASS 하고 huh v1 import 를 추가한 뮤턴트에서 FAIL 이 관측돼 있다.
- **AC-TRI-009** **Given** M1 잔여 결함표의 "검증만 종결" 행, **When** 해당 행의 근거를 확인하면, **Then** 각 행이 기준 프레임 경로와 적합 판독 근거를 가지며 해당 표면에 대한 수리 diff 가 0 임이 커밋 범위에서 확인된다.

## §F 추적성

| REQ | AC |
|---|---|
| REQ-TRI-001 | AC-TRI-001 |
| REQ-TRI-002 | AC-TRI-002 |
| REQ-TRI-003 | AC-TRI-003 |
| REQ-TRI-004 | AC-TRI-004 |
| REQ-TRI-005 | AC-TRI-005 |
| REQ-TRI-006 | AC-TRI-006, AC-TRI-007 |
| REQ-TRI-007 | AC-TRI-008 |
| REQ-TRI-008 | AC-TRI-009 |

## §G Definition of Done

- 잔여 결함표의 모든 행이 "수리 후 재검증 PASS" 또는 "검증만 종결"로 닫힘.
- 모든 렌더/i18n AC 의 증거 프레임이 SPEC 디렉터리 아래에 존재하고 골든이 갱신됨.
- 뮤턴트 관측 기록(AC-TRI-002, AC-TRI-008)이 progress.md §E.2 에 남음.
- 변경 패키지(`go test ./internal/cli/... ./internal/cli/wizard/`) 통과 — 전체 스위트는 CI 몫.
