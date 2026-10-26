---
id: SPEC-TUI-CHECKLIST-001
updated: 2026-10-27
---

# SPEC-TUI-CHECKLIST-001 — 실행 계획

마일스톤은 바뀌기 쉬운 결정부터 놓았다. 앞쪽(M0~M2)은 기준선·어휘·렌더러 모양처럼 리뷰에서 뒤집힐 수 있는 결정이고, 뒤쪽(M4~M6)은 연결과 정리처럼 기계적인 일이다. 시간 추정은 쓰지 않는다.

## §A 맥락

- 카드 t1969, Class C, P1. 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, plan 시점 HEAD `cb30e05b3e`(로컬 `main` 과 같음).
- 산출물: spec.md(REQ 20개), acceptance.md(AC 23개, 증거 원장 L1~L14), design.md, research.md, progress.md.
- 개발 방식: TDD(RED-GREEN-REFACTOR). 단, M0 은 특성화(기준선) 커밋이다.

### §A.1 Tier 판정 — L

| 근거 | 측정·추정 |
|---|---|
| 새 파일 | `internal/tui/checklist.go`, `checklist_test.go`, `internal/cli/checklist_session.go`, `checklist_session_test.go`, `checklist_plan.go`, `checklist_fallback_test.go`, 렌더러 골든 약 8개, CLI 골든 약 9개(대체 5 + 리치 4) |
| 고칠 파일 | 카탈로그 4개(en·ko·ja·zh), `internal/tui/testdata/MANIFEST.txt`, `internal/cli/init.go`, `init_warnings.go`, `reporter.go`, `severity_line.go`, `update.go`, `update_template_sync.go` |
| 파일 수 | 약 35개(15 초과) |
| 줄 수 | 렌더러 ~300, 세션·계획표 ~330, 연결 ~120, 카탈로그 4×~45, 테스트 ~700, 골든 ~300 → 1000 초과 |

줄 수는 추정이다(Gap G-7). 파일 수만으로도 Tier L 기준(15 초과)을 넘는다.

## §B init.go 변경 영역과 t1966 조율

`cb30e05b3e` 기준 행번호. t1966 hunk 목록은 research.md §7.

| 영역 | 행 | 하는 일 | t1966 hunk 와의 거리 |
|---|---|---|---|
| R1 | 383-389 (출력 관문) | `--verbose` 를 읽고 세션을 시작하는 한 줄과, 세션 종료 `defer` 한 줄을 printer 생성 줄 **앞**에 넣는다(가장 먼저 등록되어 가장 나중에 실행). 수집기에 세션을 넘긴다. | 332 뒤 삽입과 51줄, 561 뒤 삽입과 172줄 |
| R3 | 919-1075 (꼬리 단계) | 단계 앞마다 `cs.Item(…)` 한 줄을 **삽입만** 한다(기존 줄은 고치지 않음). 항목 키는 design.md §4 표. | 866 과 53줄 |
| R2 | 1077-1087 (완료 카드) | 세션이 켜져 있으면 카드 대신 요약 블록을 그리는 분기를 넣는다. 꺼져 있으면 지금 줄이 그대로 돈다. | 866 과 211줄 |

손대지 않는 줄: 728(보고기 연결 — t1966 의 730-749 삭제와 1줄 간격), 750(첫 Info — 삭제 구간 바로 다음), 780-790(자산 보장 — t1966 이 780 을 바꿈), 103-107 근처의 플래그 등록(t1966 이 바꿈). init 의 `--verbose` 등록은 새 파일 `checklist_session.go` 의 `init()` 함수에서 한다(패키지 변수 `initCmd` 는 모든 `init()` 보다 먼저 초기화된다).

조율 순서:

1. t1966 이 먼저 착지하면 이 카드는 run 시작 전에 `main` 을 흡수하고, research.md §7 표를 새 행번호로 다시 재고(영역 경계 갱신), AC-TCL-022 의 범위 숫자를 고친다(manager-spec 재위임 — 본문 수정이므로).
2. 이 카드가 먼저 착지하면 t1966 이 흡수한다. 겹치는 줄이 없으므로 텍스트 충돌은 없을 것으로 본다. 다만 t1966 이 `resuming` 계산을 위저드 앞으로 옮기므로(hunk `@@ -561,0 +586,20 @@`, `@@ -775,6 +800 @@`) 의미 충돌이 없는지 흡수 쪽이 확인한다.
3. 어느 쪽이든 흡수 뒤 `go build ./...` 와 AC-TCL-013·017 을 다시 돌린다.

## §C 사전 점검 (run 시작 시)

```bash
git branch --show-current
git rev-parse --short HEAD
git log --oneline -3 origin/main -- internal/cli/init.go
git diff --stat cb30e05b3e -- internal/cli/init.go
go build ./...
```

넷째 줄 출력이 비어 있지 않으면(이 카드 밖에서 init.go 가 바뀜) §B 1번 절차로 간다.

## §D 제약과 보존 목록

- 고치지 않는 것: `internal/cli/wizard/**`, `internal/template/templates/**`, `.github/workflows/**`, `internal/tui/progress_line.go`, `internal/tui/status.go`, `internal/tui/glyphs.go`, `internal/cli/printer/**`, 기존 골든(`internal/cli/testdata/tuxiu/**`, 기존 `internal/tui/testdata/*.golden`).
- `internal/tui/testdata/MANIFEST.txt` 는 `UPDATE_GOLDEN=1` 로만 재생성한다.
- 테스트 시접을 바꾸는 테스트는 `t.Parallel` 없이 `t.Cleanup` 으로 되돌린다.
- 전체 `go test ./...` 금지. 바뀐 패키지 범위만 돌리고, 전체 `internal/cli` 실행은 `moai slot acquire` 아래에서 한 번 또는 CI 에 맡긴다.
- 커밋은 경로를 지정해 스테이징한다. 커밋 메시지마다 카드 id `t1969` 를 넣는다.

## §E 마일스톤

### M0 — 기준선 고정 (먼저, 단독 커밋) — Priority High

- `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback-{init-stderr,init-stdout,update-stdout,update-stderr,update-verbose-stdout}.golden` 을 **현재 코드**로 찍어 넣는다. 변동 값(백업 시각, 버전, Go 런타임 문자열) 정규화 규칙을 테스트에 둔다.
- 이 커밋은 코드 변경을 담지 않는다. 커밋 직후 AC-TCL-013 이 녹색이어야 한다.
- 이유: design.md §8.

### M1 — 어휘: 카탈로그 키 — Priority High

- design.md §6 의 키를 en·ko·ja·zh 에 넣는다. ja·zh 는 초안.
- `TestChecklistCatalogFourLocales`(AC-TCL-008) 먼저 RED, 그다음 키 추가로 GREEN.

### M2 — 렌더러 — Priority High

- `internal/tui/checklist.go`, 단위 골든(clean·fail·warn·nocolor·reduced·ko), `MANIFEST.txt` 재생성.
- AC-TCL-001~007, 009, 023.
- `@MX:ANCHOR` 후보: `NewChecklist`(init·update·보고기 세 곳에서 부름). 고루틴이 없으므로 `@MX:WARN` 은 필요 없다.

### M3 — CLI 세션과 계획표 — Priority Medium

- `internal/cli/checklist_session.go`: 활성 판정(시접), 캡처 작성기, 경고 수신, 종료·쏟아내기, init `--verbose` 등록.
- `internal/cli/checklist_plan.go`: design.md §4 표.
- AC-TCL-010, 012, 014(init 부분), 015, 017(error 부분).

### M4 — init 연결 — Priority Medium

- init.go R1·R2·R3(§B), `init_warnings.go`(세션에 경고 알림·활성 시 패널 생략), `reporter.go`(세션이 있으면 체크리스트 보고기), `severity_line.go`(경고 수신 인터페이스).
- 리치 골든 `rich-init-clean.golden`, `rich-init-warn.golden`.
- AC-TCL-011(init), 016, 017(cancel), 018(init), 019(init), 020(init), 022.

### M5 — update 연결 — Priority Medium

- `update.go` 의 `runUpdate` 앞부분(`:156-176` 근처)에서 세션 시작·종료, 언어 해석.
- `update_template_sync.go`: 단계 앞에 항목 한 줄씩 삽입. 세션이 켜져 있으면 `Section`·진행 막대(`:396`, `:1081`) 를 그리지 않고 결과를 요약으로 돌린다. `ProgressLine` 호출 10곳(L12)은 고치지 않는다.
- 리치 골든 `rich-update-clean.golden`, `rich-update-fail.golden`.
- AC-TCL-011(update), 014(update), 018(update), 019(update), 020(update).

### M6 — 마감 — Priority Low

- AC 전체 재실행과 progress.md §E.2·§E.3 기록, `TestChecklistInitHunkConfinement` 제거(AC-TCL-022 의 일회성 성격), lint·커버리지·교차 빌드.

## §F 자체 검증 산출물 (run 단계 보고)

각 항목은 명령·그대로의 출력·측정 HEAD 를 함께 적는다.

- E1 AC 표(AC-TCL-001~023, §A 의 `--- PASS:` 판정)
- E2 `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`
- E3 `go test -cover ./internal/tui`
- E4 `golangci-lint run --timeout=2m` 새 지적
- E5 AC-TCL-013 순서 증거(`git merge-base --is-ancestor`)
- E6 AC-TCL-022 의 `@@` 줄 원문

## §G 리더 확인 요청 (결정 지표)

- Q1 (결정 D-1): 비TTY·CI 출력을 지금 그대로 둘지, 평문 체크리스트로 바꿀지. 이 계획은 "지금 그대로" 로 썼다. 바꾸면 REQ-TCL-012 와 AC-TCL-013 이 다시 쓰이고, 기존 버퍼 기반 테스트 상당수가 영향을 받는다.
- Q2 (결정 D-6): 체크리스트가 켜졌을 때 init 의 경고 패널을 생략할지. 이 계획은 "생략하고 항목 아래 펼침으로 대신" 으로 썼다. 패널을 남기면 경고가 두 번 보이고 24줄 예산(AC-TCL-011)이 경고 수만큼 늘어난다.
- Q3: 사용자 자산 보장 단계(`init.go:780-790`)의 항목화를 t1966 착지 뒤 후속 카드로 넘기는 것에 동의하는지.

## §H 위험

| ID | 위험 | 영향 | 대응 |
|---|---|---|---|
| R-1 | t1966 과 init.go 동시 변경 | 흡수 때 충돌·의미 충돌 | §B 영역 분리(최소 53줄 간격), 흡수 뒤 AC-TCL-013·017 재실행, AC-TCL-022 |
| R-2 | 캡처가 대화형 화면을 삼킴 | 사용자가 질문을 못 봄 | 위저드·확인은 `cmd` 작성기를 쓰지 않음(research.md §4). AC-TCL-019 가 열린 줄 0 을 확인. 실제 TTY 육안 확인은 Gap G-2 |
| R-3 | 접힌 상세에 경고가 섞여 사라짐 | 정보 손실 | 경고는 신호(수집기·`emitSeverityLine`)로 따로 알림. 문자열로 "경고처럼 보이는 줄" 을 고르지 않는다. 신호 없이 그냥 쓰인 경고 문장은 항목이 성공하면 접힌다 — run 단계가 init·update 경로의 직접 쓰기 경고(`update_template_sync.go:948`, `uikit.SymWarning()`)를 목록으로 만들어 신호 경로로 옮긴다 |
| R-4 | 24줄 예산 초과 | 한 화면에 안 들어감 | 항목 묶음(design.md §4), AC-TCL-011 |
| R-5 | ja·zh 초안 품질 | 어색한 문구 | 원어민 검수는 범위 밖(Gap G-5), 키 동치는 AC-TCL-008 |
| R-6 | 활성 판정 오탐 — 터미널인데 출력이 리디렉트된 경우 | 체크리스트가 파일에 기록됨 | 판정이 stdout·stderr 둘 다 터미널일 것을 요구 |
| R-7 | `ProgressLine` 의 비TTY 줄이 상세로 접힐 때 형식 변화 | 실패 펼침 모양이 예상과 다름 | AC-TCL-003 골든과 M5 실패 골든 |
| R-8 | 렌더러가 하나라는 제약(카드)과 보고기 어댑터 | 사실상 두 번째 렌더러처럼 보임 | 어댑터는 이벤트만 넘기고 그리지 않는다. 그리는 코드는 `checklist.go` 하나 |
| R-9 | 세션 종료 `defer` 순서 착오 | 패널·종료 블록이 캡처에 갇힘 | design.md §3.4, AC-TCL-017 |
| R-10 | `TERM=dumb`·Windows 콘솔의 줄 지우기 지원 | 잔재 문자 | `TERM=dumb` 는 대체 경로. Windows 실제 콘솔 확인은 Gap G-3 |

## §I 가정과 Gap

가정:

- A-1 huh 폼과 Bubble Tea 미리보기는 `cmd` 작성기가 아닌 자기 출력(터미널)에 쓴다(research.md §4 의 코드 읽기에 근거, 실행 관측 아님).
- A-2 init 위저드 답의 언어를 세션이 받을 수 있다. 받지 못하면 프로필 로케일, 그다음 `en`.
- A-3 `TestManifest` 는 골든이 지워진 경우에도 실패한다. 이번에 읽은 범위(`internal/tui/golden/index_test.go:70-73`)는 "새 파일" 방향만 보였다. 반대 방향은 run 단계가 확인한다.

Gap:

- G-1 트리 격리 가드가 `git -C <t1966 워크트리>` 를 거부했다. 대신 공유 ref `WT-moai-init-moai` 를 이 트리에서 읽었다. 그래서 t1966 워크트리의 커밋되지 않은 변경은 보지 못했다.
- G-2 실제 TTY 에서 위저드·확인 화면과 체크리스트가 섞이지 않는지 육안 확인을 하지 않았다(plan 단계라 코드가 없음).
- G-3 Windows 콘솔에서의 `\r\x1b[2K` 동작을 확인하지 않았다.
- G-4 이 계획의 `moai spec lint` 판정은 `go run ./cmd/moai` 로 이 트리에서 빌드한 바이너리가 냈다(progress.md §E.1).
- G-5 ja·zh 문구 원어민 검수 없음.
- G-6 긴 라벨 자르기 동작은 정하지 않았다.
- G-7 §A.1 의 줄 수는 추정이다.

## §J 참고

- SPEC-V3R6-UPDATE-PROGRESS-001 — `ProgressLine`
- SPEC-CLI-TUX-INIT-UPDATE-001 — 글리프 SSOT, update 카드, 정체성 띠
- SPEC-V3R6-UPDATE-NOISE-001 — update `--verbose` 의 뜻
- SPEC-INIT-QUIET-WIZARD-001 — init 위저드 질문 범위
