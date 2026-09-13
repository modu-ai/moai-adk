# t694 Verdict — moai update 진행 TUI 깨짐 (운영자 실측 2026-09-13)

Date: 2026-09-13 · Lane: lane-1 · Branch: `WT-update-tui-layout` (develop `4f9025151` 기점) · Class B (run→sync)

## Claim

운영자가 관측한 3 증상 — ① 분류표 14행에서 끊김 ② conflict 2건 미표시 ③ 진행바 줄 누적·출력 일괄 노출 — 의 원인을 코드로 확정하고, tmux PTY 캡처로 재현한 뒤, 확정된 기제에 대한 최소 수리를 적용해 동일 PTY에서 전후 차이를 관측한다.

## 재현 방법 ([HARD] 준수 — 이 저장소 체크아웃에서 update 미실행)

- 기점 프로젝트: `/tmp/t694-proj` 를 **v3.1.2 바이너리**(`/tmp/moai-v312`, `git archive v3.1.2` 로 /tmp 빌드)로 init → rc.10↔develop 은 템플릿 드리프트 0이라 v3.1.2 기점 사용(240파일 드리프트, 62 change rows 발생)
- 대상 바이너리: `/tmp/moai-t694` (본 브랜치 develop `4f9025151` 빌드)
- 실행: `tmux new-session -d -x 120 -y 30` PTY 안에서 `update --templates-only`, `tmux capture-pane -p` 로 화면 캡처, `send-keys y` 로 확인

## Root cause (확정)

**①② 분류표 끊김·conflict 미표시** — 두 코드 결함의 합성:

1. `update_template_sync.go` `confirmViaPreview` 가 `update.PreviewOptions{Interactive: true}` 만 넘겨 **Width/Height 가 0** → `newPreviewModel` 의 폴백 `80×24` 하드코딩으로 렌더링 (`preview_tui.go` `if height <= 0 { height = 24 }`).
2. `previewModel.Update` 는 **`tea.WindowSizeMsg` 를 처리하지 않음** → 런타임이 프로그램 시작 시 전달하는 실제 터미널 크기가 무시되고, 테이블은 `WithHeight(24-8=16)` (= 데이터행 14) 로 고정. 62행 중 44행이 fold 아래.
3. 표시 순서가 파일 순 그대로라 **conflict 2행이 마지막 부근에 배치** → 첫 화면(14행)에서 보이지 않음.

**③ 진행바 누적·일괄 노출**:

4. `update_template_sync.go:573` 이 각 단계 완료 후 `Fprintln(renderDeployProgress(i+1, ...))` 로 bar 를 **새 줄에 추가**하는데, `renderDeployProgress` 는 마지막 단계가 아니면 **● running 글리프**를 찍음 → 완료된 단계들의 bar 가 "진행 중인데 안 끝나는" 줄로 여러 줄 누적. (bar 라인 자체의 스택은 legacy "N/M steps complete" 텍스트가 그러했듯 의도된 이력 표기 — 결함은 완료 스냅샷이 running 상태를 주장한 것.)
5. "일괄 노출": 재현 실측상 62파일/584파일 배포가 확인 후 **1.2초 이내**에 완료 → inline TUI 가 preview 프레임을 지운 직후 요약 ~30줄이 단발로 출력. 점진 피드백 표면이 확인~결과 배너 사이에 없음.

## 수리 (확정 기제에 대한 최소 변경)

| 파일 | 변경 |
|---|---|
| `internal/cli/update/preview_tui.go` | ① `Update` 에 `tea.WindowSizeMsg` 처리 — `table.SetWidth/SetHeight`, `viewport.SetWidth/SetHeight` 로 실시간 리사이즈 (inset 상수 `previewTableInset=8`·`previewViewportInset=4` 로 생성자와 단일화) ② 행 표시 순서를 **conflict 우선** stable sort (`displayRank`) — 요약 카드는 classOrder 그대로 |
| `internal/cli/update_tux.go` | ③ `renderDeployProgress` 의 lead 글리프를 완료(✓)로 — 완료 스냅샷이 running 을 주장하지 않게. bar 라인의 이력 표기 설계(REQ-TUXIU-014)는 유지 |
| `internal/cli/update/preview_layout_test.go` (신규) | 리사이즈 델타·conflict 최상단 회귀 2건 |
| goldens 2건 재생성 + `preview_test.go` selectRow 테스트를 경로 기반 선택으로 | 표시 순서 변경의 의도된 파급 |

## Evidence

| # | Command / 관측 | Output |
|---|---|---|
| E1 | PTY 재현 (수리 전 빌드), preview 화면 | `evidence-prefix-preview.txt` — 62행 분류 중 **14행에서 끊김**, conflict 2건 미표시 (운영자 증상 그대로) |
| E2 | 동일 런, 'y' 후 배치 화면 | `evidence-prefix-deploy.txt` — `● 4/5 steps` 잔재가 `✓ 5/5` 위에 누적 |
| E3 | 시간축 프레임 (0.4s×6) | 프레임 1(확인 후 ~0.4s)에 이미 `Updated 584 files` — 배포 <1.2s 단발 노출 확인 |
| E4 | 수리 후 동일 PTY, preview 화면 | `evidence-postfix-preview.txt` — **conflict 2행 최상단 표시**, 가시 행 14→20 (120×30) |
| E5 | 수리 후 동일 PTY, 배치 화면 | `evidence-postfix-deploy.txt` — 중간 bar 도 `✓` lead |
| E6 | `go test ./internal/cli/update/ -count=1` | ok (golden 2건 재생성 포함 전체) |
| E7 | `go test ./internal/cli/ -run 'TestRenderDeployProgress|TestTuxiu' -count=1` | ok — `update_tux_test.go` 의 ● 기대를 ✓ 로 갱신 (AC-TUXIU-005 본질인 bar 셀·N/M 표기 단언은 유지) |
| E8 | gofmt -l(빈 출력) · go vet ./internal/cli/update/ | 클린 |

## Gaps

- 운영자의 원래 환경(터미널 크기·프로젝트) 미실측 — 본 재현은 120×30 PTY·v3.1.2 기점 62행 시나리오다. 증상 형태는 E1·E2로 동일 계열로 판정되나 수치(14행 등)는 터미널 크기에 의존한다.
- "일괄 노출"의 근본(배포 속도 자체)은 코드 결함이 아니라 소규모 프로젝트의 물리 속도에 가까워 수리하지 않았다 — 큰 프로젝트에서 단계별 ✓ 라인이 점진 출력되는지는 대형 fixture 미구축으로 미관측.
- `internal/cli` 패키지 전량 테스트는 로컬 미실행(CI 몫) — E7 선택자 범위만 측정.
- tmux(궁극적으로 tmux 서버 종료)가 화면을 지우는 순간의 세부 프레임은 캡처 한계로 미관측.

## Residual risk

- 62행 > 20행(120×30)처럼 행수가 터미널보다 크면 여전히 스크롤이 필요하다 — 힌트 바에 스크롤 키 안내(up/down)가 없어 "끊김"으로 읽힐 여지는 남는다 (후속 UX 카드 후보).
- bar 라인의 이력 스택 자체는 설계 유지 — 완전 단일 라인 재작성이 필요하면 REQ-TUXIU-014 재판정이 선행돼야 한다.

🗿 MoAI
