---
id: SPEC-TUI-CHECKLIST-001
updated: 2026-10-27
---

# SPEC-TUI-CHECKLIST-001 — 실행 계획

마일스톤은 바뀌기 쉬운 결정부터 놓았다. 앞쪽(M0~M2)은 기준선·어휘·렌더러 모양처럼 리뷰에서 뒤집힐 수 있는 결정이고, 뒤쪽(M4~M6)은 연결과 정리처럼 기계적인 일이다. 시간 추정은 쓰지 않는다.

## §A 맥락

- 카드 t1969, Class C, P1. 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, plan 시점 HEAD `cb30e05b3e`(로컬 `main` 과 같음).
- 산출물: spec.md(REQ 23개), acceptance.md(AC 24개, 증거 원장 L1~L19), design.md, research.md, progress.md.
- 개발 방식: TDD(RED-GREEN-REFACTOR). 단, M0 은 특성화(기준선) 커밋이다.

### §A.0 회차 기록

| 회차 | 측정 HEAD | 결과 | 바뀐 것 |
|---|---|---|---|
| 초안 v0.1.0 | `cb30e05b3e` | plan 감사 1회차 FAIL 0.71(감사 HEAD `d241917d17`, 차단 6·권장 6·경미 8, MP-8 실패) | — |
| 델타 1 v0.1.1 | `d241917d17` | 재감사 대기 | 결함 D1~D20 수리. REQ 20→23(REQ-TCL-021·022·023), AC 23→24(AC-TCL-024), 원장 L15~L19, 대체 골든 하위 디렉터리 `fallback/`, 순서 증거 O1~O3, 경고 수 정의 D-9, 보존 줄 목록(design.md §3.3.1), 레인 결정 Q1~Q3 반영 |

### §A.1 Tier 판정 — L

| 근거 | 측정·추정 |
|---|---|
| 새 파일 | `internal/tui/checklist.go`, `checklist_test.go`, `internal/cli/checklist_session.go`, `checklist_session_test.go`, `checklist_plan.go`, `checklist_fallback_test.go`, 렌더러 골든 약 8개, CLI 골든 18개(대체 14 + 리치 4) |
| 고칠 파일 | 카탈로그 4개(en·ko·ja·zh), `internal/tui/testdata/MANIFEST.txt`, `internal/cli/init.go`, `init_warnings.go`, `reporter.go`, `severity_line.go`, `update.go`, `update_template_sync.go`, 보존 줄 알림을 위한 호출 지점 파일(`init_update_notice.go`, `worktree_advisory.go`, `update_noise.go`, `update/backup/*` 가운데 run 단계가 고른 것) |
| 파일 수 | 약 45개(15 초과) |
| 줄 수 | 렌더러 ~300, 세션·계획표 ~380, 연결 ~150, 카탈로그 4×~45, 테스트 ~900, 골든 ~400 → 1000 초과 |

줄 수는 추정이다(Gap G-7). 파일 수만으로도 Tier L 기준(15 초과)을 넘는다.

## §B init.go 변경 영역과 t1966 조율

영역 기준 SHA: `cb30e05b3e`(REQ-TCL-021·AC-TCL-022 가 이 값을 읽는다). 행번호는 이 SHA 기준. t1966 hunk 목록은 research.md §7.

| 영역 | 행 | 하는 일 | t1966 hunk 와의 거리 |
|---|---|---|---|
| R1 | 383-389 (출력 관문) | `--verbose` 를 읽고 세션을 시작하는 줄과, 세션 종료 `defer` 줄을 printer 생성 줄 **앞**에 넣는다(가장 먼저 등록되어 가장 나중에 실행). 수집기에 세션을 넘긴다(388 수정 가능). | 332 뒤 삽입과 51줄, 561 뒤 삽입과 172줄 |
| R3 | 919-1075 (꼬리 단계) | 단계 앞마다 `cs.Item(…)` 한 줄을 **삽입만** 한다(기존 줄은 고치지 않음 — AC-TCL-022 가 `b=0` 으로 검사). 항목 키는 design.md §4 표. | 866 과 53줄 |
| R2 | 1077-1087 (완료 카드) | 세션이 켜져 있으면 카드 대신 요약 블록을 그리는 분기를 넣는다. 꺼져 있으면 지금 줄이 그대로 돈다. | 866 과 211줄 |

손대지 않는 줄: 728(보고기 연결 — t1966 의 730-749 삭제와 1줄 간격), 750(첫 Info — 삭제 구간 바로 다음), 780-790(자산 보장 — t1966 이 780 을 바꿈), 858-866(재개 경로 — t1966 의 866 수정과 겹침. REQ-TCL-023 은 세션 종료 `defer` 가 명명된 반환값으로 처리하므로 이 줄을 고칠 필요가 없다), 103-107 근처의 플래그 등록(t1966 이 바꿈). init 의 `--verbose` 등록은 새 파일 `checklist_session.go` 의 `init()` 함수에서 한다(패키지 변수 `initCmd` 는 모든 `init()` 보다 먼저 초기화된다). 보존 줄(design.md §3.3.1) 가운데 init.go 안의 지점(862, 863, 885, 890, 944, 945, 1034, 1058)은 init.go 를 고치지 않고 호출되는 함수 쪽에서 세션에 알린다.

조율 순서:

1. **t1966 이 먼저 착지하면** 이 카드는 M3 시작 전(= `internal/cli` 의 testdata 밖 파일을 처음 고치기 전)에 `main` 을 흡수한다. 흡수 뒤:
   - (a) research.md §7 표를 새 행번호로 다시 재고, 영역 기준 SHA 를 흡수 커밋으로 바꾸고 R1·R2·R3 경계를 갱신한다(manager-spec 재위임 — 본문 수정이므로).
   - (b) `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` 를 흡수된 트리에서 돌린다. 녹색이면 기준선은 그대로 유효하다.
   - (c) 빨강이면(t1966 이 init·update 출력 문구를 바꿈) **재촬영 절차**를 탄다: manager-spec 델타 회차가 AC-TCL-013 을 다시 쓰고, 이 카드의 코드가 없는 트리(흡수 커밋 그 자체 — M3 전이므로 카드 코드가 없다)에서 골든을 다시 찍어 골든만 담은 커밋 하나로 넣는다. REQ-TCL-012 의 "기준선 트리"는 그 흡수 커밋이 되고, 순서 증거 O1 은 "M0 과 재촬영 커밋 두 줄"로, O3 은 재촬영 커밋을 기준으로 다시 쓴다.
   - M3 이후에 흡수해야 하는 상황이 오면 이 카드의 코드가 섞인 트리에서는 기준선을 찍을 수 없으므로, 진행을 멈추고 리더에게 막힘 보고를 낸다.
2. **이 카드가 먼저 착지하면** t1966 이 흡수한다. 겹치는 줄이 없으므로 텍스트 충돌은 없을 것으로 본다. 다만 t1966 이 `resuming` 계산을 위저드 앞으로 옮기므로(hunk `@@ -561,0 +586,20 @@`, `@@ -775,6 +800 @@`) 의미 충돌이 없는지 흡수 쪽이 확인한다.
3. 어느 쪽이든 흡수 뒤 `go build ./...` 와 AC-TCL-013·AC-TCL-017 을 다시 돌린다.

범위 검사(AC-TCL-019 뒤 절, AC-TCL-021, AC-TCL-022, AC-TCL-013 O3)는 이 카드의 커밋만 센다(`git log --no-merges --grep=t1969`). 그래서 흡수만으로 빨개지지 않는다.

## §C 사전 점검 (run 시작 시)

각 명령은 한 번의 호출이다. 출력을 progress.md §E.2 에 그대로 남긴다.

```bash
git branch --show-current
git rev-parse --short HEAD
git log --oneline -3 origin/main -- internal/cli/init.go
git diff --stat cb30e05b3e -- internal/cli/init.go
git rev-parse WT-moai-init-moai
git diff -U0 --output=<스크래치 경로>/t1966-init.diff cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go
/usr/bin/grep -n "^@@" <스크래치 경로>/t1966-init.diff
go build ./...
```

- 넷째 줄 출력이 비어 있지 않으면(이 카드 밖에서 init.go 가 바뀜) §B 1번 절차로 간다.
- 다섯째~일곱째 줄은 t1966 진행을 다시 잰다. ref SHA 가 research.md §7 의 기록(델타 1회차 측정 `1eebf7ba6d`)과 다르거나 hunk 머리가 바뀌었으면 research.md §7 을 갱신하고, R1·R2·R3 와 최소 간격(현재 53줄)을 다시 계산한다. 간격이 0 이하가 되면(겹침) 진행을 멈추고 리더에게 막힘 보고를 낸다.
- **t1966 레인 알림.** run 시작 때 한 번, t1966 레인에 사실 통지를 보낸다: "t1969 가 init.go 의 R1 383-389·R3 919-1075(삽입만)·R2 1077-1087 을 고친다(기준 `cb30e05b3e`). 728·750·780-790·858-866 은 건드리지 않는다." 통지는 사실 전달이며 요청이 아니다(cross-session-messaging 규칙). 전달 결과(큐 적재·보류·분실)를 progress.md 에 남기고, 완료 판단은 통지 응답이 아니라 위 재측정으로 한다.
- 공유 ref 는 커밋된 것만 보인다. t1966 워크트리의 미커밋 변경은 이 점검으로 보이지 않는다(Gap G-1).

## §D 제약과 보존 목록

- 고치지 않는 것: `internal/cli/wizard/**`, `internal/template/templates/**`, `.github/workflows/**`, `internal/tui/progress_line.go`, `internal/tui/status.go`, `internal/tui/glyphs.go`, `internal/cli/printer/**`, 기존 골든(`internal/cli/testdata/tuxiu/**`, 기존 `internal/tui/testdata/*.golden`), M0 이후의 `internal/cli/testdata/checklist/fallback/**` 와 `internal/cli/checklist_fallback_test.go`.
- `internal/tui/testdata/MANIFEST.txt` 는 `UPDATE_GOLDEN=1` 로만 재생성한다.
- 테스트 시접을 바꾸는 테스트는 `t.Parallel` 없이 `t.Cleanup` 으로 되돌린다.
- 전체 `go test ./...` 금지. 바뀐 패키지 범위만 돌리고, 전체 `internal/cli` 실행은 `moai slot acquire` 아래에서 한 번 또는 CI 에 맡긴다.
- 커밋은 경로를 지정해 스테이징한다. 커밋 메시지마다 카드 id `t1969` 를 넣는다 — 범위 검사가 이 id 로 이 카드의 커밋을 고른다.

## §E 마일스톤

### M0 — 기준선 고정 (먼저, 단독 커밋) — Priority High

- `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 골든 14개(design.md §7: init·update 각각 clean·warn·fail 의 stdout·stderr 12개 + update verbose stdout·stderr 2개)를 **현재 코드**로 찍어 넣는다.
- 정규화는 acceptance.md AC-TCL-013 에 열거한 정규식만 쓴다. 정규화가 줄 수를 바꾸지 않음을 테스트가 단언한다. `warn`·`fail` 골든에 주입 문장이 있음을 단언한다(양성 대조). verbose 고정 실행이 verbose 전용 줄을 1줄 이상 내도록 픽스처를 만든다(AC-TCL-014).
- 이 커밋은 위 두 경로 밖의 파일을 담지 않는다(AC-TCL-013 O2). 커밋 직후 AC-TCL-013 이 녹색이어야 한다.
- 이유: design.md §8.

### M1 — 어휘: 카탈로그 키 — Priority High

- design.md §6 의 키를 en·ko·ja·zh 에 넣는다. ja·zh 는 초안이며 en 과 다른 문구여야 한다(허용 목록 예외만).
- `TestChecklistCatalogFourLocales`(AC-TCL-008) 먼저 RED, 그다음 키 추가로 GREEN.

### M2 — 렌더러 — Priority High

- `internal/tui/checklist.go`, 단위 골든(clean·fail·warn·nocolor·reduced·ko), `MANIFEST.txt` 재생성.
- AC-TCL-001, AC-TCL-002, AC-TCL-003, AC-TCL-004, AC-TCL-005, AC-TCL-006, AC-TCL-007, AC-TCL-009, AC-TCL-023.
- `@MX:ANCHOR` 후보: `NewChecklist`(init·update·보고기 세 곳에서 부름). 고루틴이 없으므로 `@MX:WARN` 은 필요 없다.

### M3 — CLI 세션과 계획표 — Priority Medium

- 시작 전 §B 1번 조건(t1966 착지 여부)을 확인한다.
- `internal/cli/checklist_session.go`: 입력 수집 함수 `checklistEnv`, 활성 판정과 시접 세 개(design.md §3.1), 캡처 작성기, 경고 수신, 보존 줄 수신, 종료(오류·취소·`nil` 조기 반환), init `--verbose` 등록.
- `internal/cli/checklist_plan.go`: design.md §4 표.
- AC-TCL-010, AC-TCL-012(판정 단위), AC-TCL-014(init 부분), AC-TCL-015, AC-TCL-017(error 부분).

### M4 — init 연결 — Priority Medium

- init.go R1·R2·R3(§B), `init_warnings.go`(세션에 경고 알림·활성 시 패널 생략·항목 밖 경고를 요약에 넘김), `reporter.go`(세션이 있으면 체크리스트 보고기), `severity_line.go`(경고 수신 인터페이스), init 쪽 보존 줄 지점 P1~P5(design.md §3.3.1).
- 리치 골든 `rich/init-clean.golden`, `rich/init-warn.golden`.
- AC-TCL-011(init), AC-TCL-012(init 명령 수준), AC-TCL-016, AC-TCL-017(cancel, resumed_success), AC-TCL-018(init), AC-TCL-019(init), AC-TCL-020(init), AC-TCL-022, AC-TCL-024(init 하위 테스트).

### M5 — update 연결 — Priority Medium

- `update.go` 의 `runUpdate` 앞부분(`:156-176` 근처)에서 세션 시작·종료, 언어 해석.
- `update_template_sync.go`: 단계 앞에 항목 한 줄씩 삽입. 세션이 켜져 있으면 `Section`·진행 막대(`:396`, `:1081`) 를 그리지 않고 결과를 요약으로 돌린다. `ProgressLine` 호출 10곳(L12)은 고치지 않는다. update 쪽 보존 줄 지점 P6~P11.
- 리치 골든 `rich/update-clean.golden`, `rich/update-fail.golden`.
- AC-TCL-011(update), AC-TCL-012(update 명령 수준), AC-TCL-014(update), AC-TCL-018(update), AC-TCL-019(update), AC-TCL-020(update), AC-TCL-024(update 하위 테스트).

### M6 — 마감 — Priority Low

- AC 전체 재실행과 progress.md §E.2·§E.3 기록.
- AC-TCL-013 순서 증거 O1·O2·O3 실행과 기록.
- `TestChecklistInitHunkConfinement` 의 마지막 실행(명령·그대로의 출력·종료 코드·측정 SHA)을 기록한 **뒤** 제거(AC-TCL-022 의 일회성 성격).
- lint·커버리지·교차 빌드.

## §F 자체 검증 산출물 (run 단계 보고)

각 항목은 명령·그대로의 출력·측정 HEAD 를 함께 적는다.

- E1 AC 표(AC-TCL-001 부터 AC-TCL-024 까지, §A 의 `--- PASS:` 판정. AC-TCL-022 는 제거 전 기록)
- E2 `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`
- E3 `go test -cover ./internal/tui`
- E4 `golangci-lint run --timeout=2m` 새 지적
- E5 AC-TCL-013 순서 증거 O1·O2·O3 원문
- E6 AC-TCL-022 의 `@@` 줄 원문
- E7 §C 사전 점검 출력과 t1966 레인 알림의 전달 결과

## §G 레인 결정 (Kickoff 기록에서 리더에게 알림)

세 결정은 레인 판단(사다리 마지막 단계)으로 정했다. 근거 기록: `.moai/reports/t1969/progress.md` 결정 기록. 모두 되돌릴 수 있다.

- **Q1 (결정 D-1) 비TTY·CI 출력 = 기준선과 같은 바이트.** 독해: 카드 문구 "TTY 아님/CI/--verbose 면 평문 줄 출력으로 폴백"을 "오늘의 평문 줄 출력으로 폴백"으로 읽었다. 오늘 비TTY 출력은 이스케이프 없는 평문 줄이라 문구를 문자 그대로 충족하고, 완료 조건의 "비TTY 골든"은 M0 기준선이 채운다. 공개한 가정이며 부채로 들고 간다. **리더가 다른 독해를 택하면**(카드가 `--verbose` 만 "현재 상세 그대로"라 따로 적었으니 비TTY·CI 는 평문 체크리스트로 바꾸라는 뜻): manager-spec 델타 회차가 REQ-TCL-012·AC-TCL-013·AC-TCL-014 와 M0 을 다시 쓰고, 렌더러에 평문 모드를 더하며(design.md §2 의 YAGNI 결정 철회), 기존 버퍼 기반 테스트 가운데 init·update 출력을 단언하는 것들을 새 모양으로 옮긴다.
- **Q2 (결정 D-6) 체크리스트가 켜지면 init 경고 패널 생략.** 항목 안 경고는 항목 아래 한 번 펼치고, 항목 밖 경고는 요약이 다시 나열한다(REQ-TCL-015, design.md §3.3). 경고 수는 수집기 수 하나로 정의했다(결정 D-9). **리더가 패널 유지를 택하면**: REQ-TCL-015 와 AC-TCL-016 의 패널 0번 조건을 1번으로 바꾸고, 24줄 예산(AC-TCL-011)은 깨끗한 실행 기준이라 그대로 두되 경고가 있는 실행은 예산 밖임을 명시한다.
- **Q3 사용자 자산 보장 단계(`init.go:780-790`) 항목화는 후속 카드.** t1966 이 `:775-780` 을 바꾸므로 그 착지 뒤로 미룬다. 이 단계는 `executor.Execute`(`init.go:829`)보다 먼저 돌아 열린 항목이 없으므로 출력이 넘김 규칙으로 그대로 보인다. **리더가 이번 카드에 넣기를 택하면**: t1966 착지 뒤 §B 1번 절차로 흡수하고, 계획표에 `assets` 항목을 더하는 델타 회차를 연다.
- **글리프 (결정 D-2).** 카드의 `➤`·`–` 대신 `◆`·`○`. 이모지 검사와 허용 목록 충돌 때문이다(design.md §1.1).

## §H 위험

| ID | 위험 | 영향 | 대응 |
|---|---|---|---|
| R-1 | t1966 과 init.go 동시 변경 | 흡수 때 충돌·의미 충돌 | §B 영역 분리(최소 53줄 간격), §C 재측정과 t1966 레인 알림, 흡수 뒤 AC-TCL-013·AC-TCL-017 재실행, AC-TCL-022 |
| R-2 | 캡처가 대화형 화면을 삼킴 | 사용자가 질문을 못 봄 | 위저드·확인은 `cmd` 작성기를 쓰지 않음(research.md §4). AC-TCL-019 가 열린 줄 0 을 확인. 실제 TTY 육안 확인은 Gap G-2 |
| R-3 | 접힌 상세에 경고·오류·안내가 섞여 사라짐 | 정보 손실 | 경고는 신호(수집기·`emitSeverityLine`)로, 직접 쓰기 줄은 보존 줄 목록(design.md §3.3.1, REQ-TCL-022)으로 따로 알린다. 문자열로 "경고처럼 보이는 줄" 을 고르지 않는다. AC-TCL-024 가 네 지점을 주입해 확인한다. 목록 밖에 새 직접 쓰기 경고가 생기면 접힌다 — 남은 위험 |
| R-4 | 24줄 예산 초과 | 한 화면에 안 들어감 | 항목 묶음(design.md §4), AC-TCL-011. update 깨끗한 실행은 예산에 딱 닿는다(design.md §4 줄 예산). M5 실측이 넘으면 기준을 느슨하게 하지 않고 manager-spec 에 되돌린다 |
| R-5 | ja·zh 초안 품질 | 어색한 문구 | 원어민 검수는 범위 밖(Gap G-5), 키 동치·en 과 다름은 AC-TCL-008 |
| R-6 | 활성 판정 오탐 — 터미널인데 출력이 리디렉트된 경우 | 체크리스트가 파일에 기록됨 | 입력 수집을 한 함수로 모으고 fd 를 따로 판정(design.md §3.1). AC-TCL-012 명령 수준 하위 테스트 `*_stderr_not_tty` 가 호출 지점 오류까지 잡는다 |
| R-7 | `ProgressLine` 의 비TTY 줄이 상세로 접힐 때 형식 변화 | 실패 펼침 모양이 예상과 다름 | AC-TCL-003 골든과 M5 실패 골든 |
| R-8 | 렌더러가 하나라는 제약(카드)과 보고기 어댑터 | 사실상 두 번째 렌더러처럼 보임 | 어댑터는 이벤트만 넘기고 그리지 않는다. 그리는 코드는 `checklist.go` 하나 |
| R-9 | 세션 종료 `defer` 순서 착오 | 패널·종료 블록이 캡처에 갇힘 | design.md §3.4, AC-TCL-017 |
| R-10 | `TERM=dumb`·Windows 콘솔의 줄 지우기 지원 | 잔재 문자 | `TERM=dumb` 는 대체 경로. Windows 실제 콘솔 확인은 Gap G-3 |
| R-11 | AC-TCL-022 테스트가 git 이력을 읽음 | 얕은 클론 CI 에서 M4~M6 사이 빨강 | 테스트는 이력 부족을 분명한 메시지로 실패시킨다. M6 에서 지운다. M4~M6 사이 원격 CI 후보 푸시가 있으면 그 실패를 이 사유로 분류한다 |

## §I 가정과 Gap

가정:

- A-1 huh 폼과 Bubble Tea 미리보기는 `cmd` 작성기가 아닌 자기 출력(터미널)에 쓴다(research.md §4 의 코드 읽기에 근거, 실행 관측 아님).
- A-2 init 위저드 답의 언어를 세션이 받을 수 있다. 받지 못하면 프로필 로케일, 그다음 `en`.

확인된 사실(이전 가정 A-3):

- `TestManifest` 는 골든이 새로 생긴 경우(`internal/tui/golden/index_test.go:73` `golden file %s is new`)와 지워진 경우(`:80` `manifest entry %s is missing on disk`)를 모두 실패시킨다. 델타 1회차에서 HEAD `d241917d17` 의 `:68-83` 을 읽어 확인했다.

Gap:

- G-1 트리 격리 가드가 `git -C <t1966 워크트리>` 를 거부했다. 대신 공유 ref `WT-moai-init-moai` 를 이 트리에서 읽었다(델타 1회차 재측정 `1eebf7ba6d`). t1966 워크트리의 커밋되지 않은 변경은 보지 못했다. §C 가 run 시작 때 다시 잰다.
- G-2 실제 TTY 에서 위저드·확인 화면과 체크리스트가 섞이지 않는지 육안 확인을 하지 않았다(plan 단계라 코드가 없음).
- G-3 Windows 콘솔에서의 `\r\x1b[2K` 동작을 확인하지 않았다.
- G-4 이 계획의 `moai spec lint` 판정은 `go run ./cmd/moai` 로 이 트리에서 빌드한 바이너리가 냈다(progress.md §E.1).
- G-5 ja·zh 문구 원어민 검수 없음.
- G-6 긴 라벨 자르기 동작은 정하지 않았다.
- G-7 §A.1 의 줄 수는 추정이다.
- G-8 update 깨끗한 실행의 24줄 적합성은 실측하지 않았다(design.md §4 의 줄 예산은 계산값).

## §J 참고

- SPEC-V3R6-UPDATE-PROGRESS-001 — `ProgressLine`
- SPEC-CLI-TUX-INIT-UPDATE-001 — 글리프 SSOT, update 카드, 정체성 띠
- SPEC-V3R6-UPDATE-NOISE-001 — update `--verbose` 의 뜻
- SPEC-INIT-QUIET-WIZARD-001 — init 위저드 질문 범위
