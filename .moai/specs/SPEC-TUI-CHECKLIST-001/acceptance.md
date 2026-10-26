# SPEC-TUI-CHECKLIST-001 — 인수 기준

문서 핀: 원장 L1~L14 의 RED-now 측정은 트리 `cb30e05b3e`(워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`)에서 했다. 자기 핀이 없는 기준은 이 핀을 따른다. L15~L19 는 트리 `d241917d17a3754541ef5ad52536d68023faf142` 에서 쟀고 그 핀을 따로 단다. 두 트리는 코드가 같다(`d241917d17` 은 SPEC 디렉터리만 더한 plan 커밋 — 1회차 감사 E-3).

이 파일은 검증 층이다. 요구사항(REQ)의 GEARS 문장은 spec.md §3 에 있고, 여기서는 각 기준을 Given-When-Then 과 이진 판정으로 적는다. 각 기준 제목의 `maps` 목록이 REQ 대응을 선언한다.

## §A 판정 규칙

- **명령 형식.** 모든 명령은 파이프·리다이렉트·`&&`·`;` 없는 한 번의 호출이다. 이 셸의 `grep` 은 ugrep 셸 함수라 일치 0건에서도 종료 코드 0 을 낸다(research.md §0). 그래서 모든 검색은 `/usr/bin/grep` 으로 적는다.
- **선택자 공허 방지.** `go test -run` 은 일치하는 테스트가 0개여도 `ok … [no tests to run]` 와 종료 코드 0 을 낸다(원장 L11 — AC-TCL-001 의 녹색 명령을 오늘 그대로 돌린 결과다). 그래서 테스트 기반 기준의 녹색 판정은 종료 코드 0 **그리고** 출력에 기준이 이름 붙인 `--- PASS: <테스트명> ` 줄(이름 뒤 공백 포함)이 하나도 빠짐없이 있는 것이다. 선택자는 `^…$` 로 묶고, `-v` 와 `-count=1` 을 항상 붙인다.
- **기준 종류.** release-blocking 기준은 RED-now 네 요소(명령, 그대로의 stdout, 종료 코드, 트리 SHA)를 갖춘다. 오늘 이미 녹색인 범위 경계 검사는 regression-guard 로 분류하고 통과로 세지 않는다.
- **이 카드의 커밋.** 범위 검사는 이 카드의 커밋만 센다 — 메시지에 카드 id `t1969` 가 있고 병합 커밋이 아닌 것(`git log --no-merges --grep=t1969`). `main` 을 흡수해도 다른 카드의 커밋이 범위 검사를 빨갛게 만들지 않는다(1회차 감사 D7).

## §B 증거 원장 (RED-now)

stdout 이 비어 있으면 `(빈 출력)` 으로 적고, stderr 는 따로 적는다.

| ID | 트리 | 명령 | stdout | stderr | 종료 코드 |
|---|---|---|---|---|---|
| L1 | `cb30e05b3e` | `ls internal/tui/checklist.go` | (빈 출력) | `ls: internal/tui/checklist.go: No such file or directory` | 1 |
| L2 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/en.yaml` | `0` | (없음) | 1 |
| L3 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ko.yaml` | `0` | (없음) | 1 |
| L4 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ja.yaml` | `0` | (없음) | 1 |
| L5 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` | `0` | (없음) | 1 |
| L6 | `cb30e05b3e` | `/usr/bin/grep -n "\"verbose\"" internal/cli/init.go` | (빈 출력) | (없음) | 1 |
| L7 | `cb30e05b3e` | `/usr/bin/grep -l "checklist" internal/cli/init.go internal/cli/update.go internal/cli/update_template_sync.go internal/cli/init_warnings.go internal/cli/severity_line.go internal/cli/reporter.go` | (빈 출력) | (없음) | 1 |
| L8 | `cb30e05b3e` | `ls internal/cli/testdata/checklist` | (빈 출력) | `ls: internal/cli/testdata/checklist: No such file or directory` | 1 |
| L9 | `cb30e05b3e` | `/usr/bin/grep -rn "func NewChecklist" internal/tui` | (빈 출력) | (없음) | 1 |
| L10 | `cb30e05b3e` | `ls internal/tui/checklist_test.go internal/cli/checklist_session.go internal/cli/checklist_session_test.go` | (빈 출력) | `ls: internal/cli/checklist_session_test.go: No such file or directory` / `ls: internal/cli/checklist_session.go: No such file or directory` / `ls: internal/tui/checklist_test.go: No such file or directory` | 1 |
| L11 | `cb30e05b3e` | `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/tui	0.301s [no tests to run]` | (없음) | 0 |
| L12 | `cb30e05b3e` | `/usr/bin/grep -c "ProgressLine(" internal/cli/update_template_sync.go` | `10` | (없음) | 0 |
| L13 | `cb30e05b3e` | `git diff --name-only cb30e05b3e -- internal/template/templates` | (빈 출력) | (없음) | 0 |
| L14 | `cb30e05b3e` | `/usr/bin/grep -c "emitSummary" internal/cli/init.go` | `1` | (없음) | 0 |
| L15 | `d241917d17` | `git diff -U0 cb30e05b3e -- internal/cli/init.go` | (빈 출력) | (없음) | 0 |
| L16 | `d241917d17` | `/usr/bin/grep -rln "TestChecklistInitHunkConfinement" internal/cli` | (빈 출력) | (없음) | 1 |
| L17 | `d241917d17` | `/usr/bin/grep -rln "TestChecklistPreservedLinesExactlyOnce" internal/cli` | (빈 출력) | (없음) | 1 |
| L18 | `d241917d17` | `/usr/bin/grep -rn "func checklistEnv" internal/cli` | (빈 출력) | (없음) | 1 |
| L19 | `d241917d17` | `ls internal/cli/testdata/checklist/fallback` | (빈 출력) | `ls: internal/cli/testdata/checklist/fallback: No such file or directory` | 1 |

- L11 은 RED 가 아니라 **대조군**이다. 아직 없는 테스트를 고르는 선택자가 `PASS` 와 종료 코드 0 을 낸다. §A 의 `--- PASS:` 요구가 여기서 나온다. 종료 코드만 보는 판정은 이 SPEC 의 모든 테스트 기반 기준을 오늘 녹색으로 만든다.
- L12·L14 는 기준선이다. L13 은 오늘 이미 비어 있어 AC-TCL-021 은 regression-guard 다.
- L15 는 AC-TCL-022 의 RED 전제("이 카드의 init.go hunk 가 0개")를 직접 잰다. 종료 코드 0 과 빈 출력은 hunk 가 하나도 없다는 뜻이다. L16 은 그 녹색 테스트가 아직 없음을 잰다.

## §C 인수 기준

각 기준의 "RED 이유" 는 오늘 왜 빨간지, "녹색 경로" 는 어느 마일스톤이 무엇으로 뒤집는지를 적는다. 뮤턴트는 기준을 통과하면서 요구를 어기는 구현을 시도해 본 결과다. "계속 발화" 는 이 검사가 멈췄을 때 누가 어떻게 알게 되는지다.

### AC-TCL-001 — 깨끗한 실행의 체크리스트 모양 (maps REQ-TCL-001, REQ-TCL-003)

- Given 고정 시계·`LightTheme`·고정 환경 요약과 design.md §4 의 init·update 계획표, When 모든 항목을 `Done` 으로 끝내고 요약을 그리면, Then 출력이 `internal/tui/testdata/checklist-init-clean.golden` 과 `checklist-update-clean.golden` 에 바이트 단위로 같고 머리줄 1개·환경 줄 1개·그룹 제목·항목당 한 줄 구조를 갖는다.
- 종류: release-blocking. RED 이유: L1·L9·L10 — 렌더러와 테스트가 없다. L11 이 같은 명령의 오늘 출력이며 `--- PASS:` 줄이 없다.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` → 종료 0, `--- PASS: TestChecklistGoldenClean ` 포함.
- 뮤턴트: 항목마다 두 줄(진행 줄 + 결과 줄)을 남기는 렌더러 → 골든 줄 수가 달라 실패. 골든을 렌더러 출력으로 덮어쓰는 뮤턴트 → AC-TCL-011 줄 예산과 AC-TCL-002 시퀀스 검사가 따로 잡는다.
- 계속 발화: 기본 `go test ./internal/tui` 실행에 포함. 골든을 새로 만들거나 지우면 `TestManifest` 가 실패한다 — 새 파일은 `internal/tui/golden/index_test.go:73`, 지운 파일은 `:80` (`manifest entry %s is missing on disk`).

### AC-TCL-002 — 실행 중인 줄만 제자리에서 다시 그림 (maps REQ-TCL-002)

- Given 한 항목에 `Update` 를 세 번 부른 뒤 `Done` 하고 다음 항목을 연 렌더러 출력, When 바이트를 훑으면, Then `\r\x1b[2K` 는 정확히 4번(갱신 3번 + 종료 1번) 나오고, 모든 `\n` 바로 다음 위치에는 `\r\x1b[2K` 가 오지 않는다(끝난 줄을 다시 그리지 않는다).
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistRedrawOnlyRunningLine$' -count=1 -v` → `--- PASS: TestChecklistRedrawOnlyRunningLine `.
- 뮤턴트: 화면 전체를 매번 다시 그리는 렌더러 → 시퀀스 수 초과로 실패. 다시 그리기를 전혀 안 하는 렌더러 → 수 부족으로 실패.
- 계속 발화: 패키지 기본 실행에 포함.

### AC-TCL-003 — 실패 시 상세 펼침 (maps REQ-TCL-004)

- Given 상세 줄 3개를 기록한 뒤 `Fail("boom")` 한 항목과, 상세 줄 2개를 기록하고 `Done` 한 항목, When 그리면, Then 출력이 `checklist-init-fail.golden` 과 같고, 실패 항목의 상세 3줄과 `boom` 은 들여쓴 채 모두 보이며 성공 항목의 상세 2줄은 하나도 보이지 않는다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistFailExpandsDetail$' -count=1 -v` → `--- PASS: TestChecklistFailExpandsDetail `.
- 뮤턴트: 실패 메시지만 펼치고 상세를 버리는 렌더러 → 상세 줄 부재로 실패. 모든 항목 상세를 펼치는 렌더러 → 성공 항목 상세 노출로 실패(양방향 검사).
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-004 — 경고만 펼치고 나머지는 접음 (maps REQ-TCL-005)

- Given 경고 하나를 기록한 항목과 경고 없는 항목, When 그리면, Then 경고 항목 줄은 `!` 로 끝나고 그 아래 경고 문장이 보이며, 경고 없는 항목의 상세는 보이지 않는다. 출력은 `checklist-init-warn.golden` 과 같다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistWarnExpandsOnlyWarned$' -count=1 -v` → `--- PASS: TestChecklistWarnExpandsOnlyWarned `.
- 뮤턴트: 경고 수만 요약에 세고 본문을 펼치지 않는 렌더러 → 경고 문장 부재로 실패.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-005 — 요약 블록 필드 (maps REQ-TCL-006)

- Given 바뀐 수 7, 보존 수 3, 경고 1, 고정 시계로 1.5초 경과, 백업 경로와 복구 명령을 담은 update 요약과, 파일 2·디렉터리 13·배포 모드·다음 단계·항목 밖 경고 문장 1개를 담은 init 요약, When 그리면, Then 각 숫자와 문자열이 요약에 한 번씩 나오고 경과 시간이 `1.5s` 로 보인다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistSummaryFields$' -count=1 -v` → `--- PASS: TestChecklistSummaryFields `.
- 뮤턴트: 백업 경로·복구 명령이나 항목 밖 경고 문장을 빠뜨린 요약 → 문자열 부재로 실패. 시계를 무시하고 `time.Now` 를 쓰는 렌더러 → `1.5s` 불일치로 실패.
- 계속 발화: 패키지 기본 실행.

### AC-TCL-006 — NO_COLOR 에서 SGR 없음 (maps REQ-TCL-007)

- Given `MonochromeTheme`(NO_COLOR 해석 결과)로 만든 렌더러, When 깨끗한 실행·실패·경고를 모두 그리면, Then 출력에 정규식 `\x1b\[[0-9;]*m` 일치가 0건이고 `checklist-init-nocolor.golden` 과 같다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistNoColorHasNoSGR$' -count=1 -v` → `--- PASS: TestChecklistNoColorHasNoSGR `.
- 뮤턴트: 테마를 무시하고 굵게 속성을 직접 넣는 렌더러 → SGR 일치로 실패. 빈 출력을 검사하는 테스트 뮤턴트 → 골든 비교가 막는다.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-007 — 모션 감소에서 점자 프레임 없음 (maps REQ-TCL-008)

- Given `MOAI_REDUCED_MOTION=1`, When 실행 중 항목에 `Update` 를 여러 번 부르면, Then 출력에 U+2800–U+28FF 문자가 하나도 없고 실행 중 줄은 `●` 를 쓴다. 대조로 같은 테스트가 환경 변수를 비우면 `⠋` 가 나온다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistReducedMotion$' -count=1 -v` → `--- PASS: TestChecklistReducedMotion `. 이 테스트는 `t.Setenv` 를 쓰므로 `t.Parallel` 을 쓰지 않는다.
- 뮤턴트: 항상 `●` 만 쓰는 렌더러 → 대조(환경 변수 비움)에서 `⠋` 부재로 실패.
- 계속 발화: 패키지 기본 실행.

### AC-TCL-008 — 네 언어 카탈로그 동치 (maps REQ-TCL-009)

- Given en·ko·ja·zh 카탈로그, When `checklist.` 로 시작하는 키 집합을 비교하면, Then 네 집합이 같고, 값이 모두 비어 있지 않고 키 자신과 다르며, 같은 키의 `%` 서식 지시자 개수가 네 언어에서 같고, 값에 U+1F300–1FAFF·U+2600–27BF·U+1F000–1F2FF 문자가 없으며, ja 값과 zh 값이 같은 키의 en 값과 다르다. 마지막 조건의 예외는 테스트 안의 허용 목록(초기값: 값이 `MCP` 하나뿐인 키)에 이름이 오른 키뿐이다. en 키 수는 design.md §6 표의 키 수와 같다.
- 종류: release-blocking. RED 이유: L2·L3·L4·L5 — 네 카탈로그 모두 `checklist.` 키가 0개다.
- 녹색 경로(M1): `go test ./internal/tui -run '^TestChecklistCatalogFourLocales$' -count=1 -v` → `--- PASS: TestChecklistCatalogFourLocales `. 보조 관측: `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` → en 과 같은 양수, 종료 0.
- 뮤턴트: ja 에 키를 빠뜨리면 `Translate` 가 en 으로 대체해 화면은 멀쩡해 보인다 → 집합 비교가 잡는다. 값에 키 이름을 그대로 넣는 뮤턴트 → 값 ≠ 키 조건이 잡는다. ja·zh 값을 en 문구 그대로 복사하는 뮤턴트 → ja·zh ≠ en 조건이 잡는다. 렌더러가 없는 키를 불러 원시 키가 화면에 남는 뮤턴트 → AC-TCL-023 과 AC-TCL-011 의 골든 원시 키 검사가 잡는다.
- 계속 발화: 패키지 기본 실행. 키를 하나 더하고 한 언어를 잊으면 바로 빨개진다.

### AC-TCL-009 — 한국어 폭 정렬과 미지원 언어 대체 (maps REQ-TCL-009)

- Given `Lang: "ko"` 렌더러와 `Lang: "xx"` 렌더러, When 같은 계획을 그리면, Then ko 출력이 `checklist-init-ko.golden` 과 같고(한글 2칸 폭 기준 정렬), xx 출력은 en 출력과 바이트 단위로 같다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistKoreanAndFallback$' -count=1 -v` → `--- PASS: TestChecklistKoreanAndFallback `.
- 뮤턴트: `len()` 으로 폭을 재는 렌더러 → ko 골든 정렬 불일치로 실패.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-010 — 그룹당 8개 이하 (maps REQ-TCL-010)

- Given `internal/cli` 의 계획표, When 그룹별 항목 수와 키를 검사하면, Then 모든 그룹이 1~8개 항목을 갖고, 모든 그룹·항목 키가 en 카탈로그에서 키 자신이 아닌 값으로 풀린다.
- 종류: release-blocking. RED 이유: L7·L10 — 계획표와 세션 코드가 없다.
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistPlanGroupCap$' -count=1 -v` → `--- PASS: TestChecklistPlanGroupCap `.
- 뮤턴트: 계획표 대신 호출 지점에 키를 흩어 두는 구현 → 이 검사는 녹색이어도 AC-TCL-011 의 실제 실행 줄 수가 잡는다(표만 보는 한계를 둘째 기준이 메운다).
- 계속 발화: `internal/cli` 기본 실행에 포함(CI 전체 실행).

### AC-TCL-011 — 한 화면 24줄 (maps REQ-TCL-010, REQ-TCL-009)

- Given 활성 판정 시접을 강제로 켠, spec.md §3.3 정의대로의 깨끗한 init 실행과 깨끗한 update 실행, When 두 작성기 출력을 합쳐 `\r\x1b[2K` 로 덮인 줄을 최종 화면으로 접고 ANSI 를 지운 뒤 줄을 세면, Then 둘 다 24줄 이하이고(update 정체성 띠처럼 넘김 규칙으로 보이는 줄도 센다), 접은 결과가 `internal/cli/testdata/checklist/rich/init-clean.golden`, `rich/update-clean.golden` 과 같으며, `rich/` 아래 모든 골든에 부분 문자열 `checklist.` 가 없다.
- 종류: release-blocking. RED 이유: L8·L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistScreenBudget$' -count=1 -v` → `--- PASS: TestChecklistScreenBudget/init `, `--- PASS: TestChecklistScreenBudget/update `, `--- PASS: TestChecklistScreenBudget/no_raw_keys `.
- 뮤턴트: 넘김 규칙을 잘못 걸어 항목 안 상세가 새는 구현 → 줄 수 초과로 실패. 정체성 띠를 세지 않는 측정 → 두 작성기 합산 조건이 막는다. 카탈로그에 없는 키를 부르는 구현 → 원시 키 검사로 실패.
- 계속 발화: `internal/cli` 기본 실행. 골든이 없으면 `t.Fatalf` 로 실패하고 건너뛰지 않는다(기존 `tuxReadGolden` 과 같은 방식, `internal/cli/tuxiu_characterization_test.go:115-122`).

### AC-TCL-012 — 활성 판정과 입력 수집 (maps REQ-TCL-011)

- Given design.md §3.1 의 다섯 조건과 그 입력을 모으는 한 함수 `checklistEnv`, When (1) 판정 단위로 조건마다 하나씩만 거짓으로 바꾼 경우와 모두 참인 경우를 판정하고, (2) 명령 수준에서 TTY 탐지 시접과 프로세스 스트림 시접(design.md §3.1)을 써 "stdout TTY·stderr 비TTY", `CI=1`, `TERM=dumb` 각각으로 init 과 update 를 돌리면, Then (1) 모두 참일 때만 활성이고, (2) 세 경우 모두 출력이 AC-TCL-013 의 깨끗한 실행 대체 골든과 같으며, 대조로 같은 시접에서 모든 조건이 참이면 체크리스트 머리줄이 나온다.
- 종류: release-blocking. RED 이유: L7·L10·L18 — 판정 함수, 입력 수집 함수, 세션 코드가 없다.
- 녹색 경로(M3 판정 단위, M4·M5 명령 수준): `go test ./internal/cli -run '^TestChecklistActivationPredicate$' -count=1 -v` → `--- PASS: TestChecklistActivationPredicate `, `--- PASS: TestChecklistActivationPredicate/all_true `, `--- PASS: TestChecklistActivationPredicate/stdout_not_tty `, `--- PASS: TestChecklistActivationPredicate/stderr_not_tty `, `--- PASS: TestChecklistActivationPredicate/writer_replaced `, `--- PASS: TestChecklistActivationPredicate/ci_set `, `--- PASS: TestChecklistActivationPredicate/term_dumb `, `--- PASS: TestChecklistActivationPredicate/verbose_on `. 그리고 `go test ./internal/cli -run '^TestChecklistGateCommandLevel$' -count=1 -v` → `--- PASS: TestChecklistGateCommandLevel/init_stderr_not_tty `, `--- PASS: TestChecklistGateCommandLevel/init_ci_set `, `--- PASS: TestChecklistGateCommandLevel/init_term_dumb `, `--- PASS: TestChecklistGateCommandLevel/update_stderr_not_tty `, `--- PASS: TestChecklistGateCommandLevel/update_ci_set `, `--- PASS: TestChecklistGateCommandLevel/update_term_dumb `, `--- PASS: TestChecklistGateCommandLevel/control_all_true `.
- 뮤턴트: 조건 하나(예: `TERM=dumb`)를 빼먹은 판정 → 해당 하위 테스트 실패. 늘 비활성인 판정 → `all_true`·`control_all_true` 실패(양방향). 호출 지점이 stderr 판정에 stdout 의 TTY 결과를 넘기는 구현 → `*_stderr_not_tty` 가 체크리스트를 그려 골든 불일치로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-013 — 대체 경로는 기준선과 같은 바이트, 기준선이 먼저 (maps REQ-TCL-012)

- Given M0 이 코드 변경 **전에**, 그 둘만 담은 커밋 하나로 넣은 `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 아래 골든 — init·update 각각 시나리오 세 개(깨끗한 실행 `clean`, 수집기 경고 1건 이상 `warn`, 실패 1건 `fail`)의 stdout·stderr 골든 12개, When 버퍼 작성기(비TTY)로 같은 고정 실행을 돌리면, Then 출력이 아래 정규화 뒤 골든과 바이트 단위로 같고, 정규화가 줄 수를 바꾸지 않으며, `warn` 골든에는 주입한 경고 문장이, `fail` 골든에는 주입한 실패 문장이 있다.
- 정규화 규칙(테스트가 이 순서로 적용하는 정규식 전부. 이 밖의 치환은 없다):
  - 프로젝트 임시 경로: 테스트가 만든 `t.TempDir()` 절대 경로 문자열 → `<ROOT>` (정규식이 아닌 고정 문자열 치환)
  - 백업 디렉터리 시각: `\b\d{8}_\d{6}\b` → `<TS>` (배치 형식 `20060102_150405`, `internal/cli/update/backup/retention_realdata_test.go:53`)
  - 버전: `\bv\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?\b` → `<VER>`
  - Go 런타임: `\bgo1\.\d+(\.\d+)?\b` → `<GO>` (update 정체성 띠의 `runtime.Version()`, `internal/cli/update_tux.go:80`)
  - 경과 시간: `\b\d+(\.\d+)?(ms|s)\b` → `<DUR>`
- 종류: release-blocking. RED 이유: L8·L19 — 기준선 골든 디렉터리가 없다. 이 기준은 기준선 커밋이 구현 커밋보다 앞서야 성립한다(design.md §8).
- 녹색 경로(M0 에서 녹색으로 들어오고, M4·M5 뒤에도 녹색이어야 함): `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` → `--- PASS: TestChecklistFallbackMatchesBaseline/init_clean `, `--- PASS: TestChecklistFallbackMatchesBaseline/init_warn `, `--- PASS: TestChecklistFallbackMatchesBaseline/init_fail `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_clean `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_warn `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_fail `.
- 순서 증거(M6 에서 각각 한 번의 호출로 돌리고 progress.md §E.2 에 그대로 기록. 셋 다 성립해야 녹색):
  - O1 `git log --format=%h cb30e05b3e..HEAD -- internal/cli/testdata/checklist/fallback internal/cli/checklist_fallback_test.go` → 정확히 한 줄, 그 SHA 가 M0 이다. 둘째 줄이 있으면(구현 뒤 골든 수정·재촬영·테스트 정규화 변경) 빨강.
  - O2 `git show --name-only --format= <M0>` → `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 아래 경로만. 다른 경로가 하나라도 있으면(코드 변경 혼입) 빨강.
  - O3 `git log --no-merges --grep=t1969 --format=%h cb30e05b3e..HEAD -- internal ':(exclude)internal/cli/testdata'` 가 내는 SHA 각각에 대해 `git merge-base --is-ancestor <M0> <SHA>` → 종료 0, 그리고 그 SHA 는 M0 과 다르다. 하나라도 종료 1 이면(M0 보다 먼저 들어간 코드 변경) 빨강.
- 뮤턴트: 기준선을 구현 뒤에 다시 찍어 수정 커밋으로 넣는 절차 → O1 이 두 줄. M0 에 `init.go` 를 섞는 절차 → O2. CLI 쪽 변경(`init_warnings.go` 등)을 M0 보다 먼저 커밋하는 절차 → O3. 경고 줄을 지우는 넓은 정규화 → 줄 수 불변 조건과 `warn` 골든의 주입 문장 조건이 잡는다.
- 계속 발화: 골든 비교는 `internal/cli` 기본 실행, 골든 부재는 `t.Fatalf`. 순서 증거 O1~O3 는 이 카드 범위의 일회성 검사이며 sync 감사가 다시 돈다. `main` 흡수로 기준선이 바뀌어야 하면 plan.md §B 의 재촬영 절차가 이 기준을 다시 쓴다.

### AC-TCL-014 — `--verbose` (maps REQ-TCL-013)

- Given init 과 update, 그리고 M0 이 함께 넣은 `fallback/update-verbose-stdout.golden`·`fallback/update-verbose-stderr.golden` — 그 고정 실행은 verbose 에서만 나오는 줄을 1줄 이상 낸다(예: 3-way 병합 실패 대체 경고, `internal/cli/update_noise.go:111-114`), When `--verbose` 를 주고 활성 판정 시접을 켠 상태로 돌리면, Then 체크리스트가 켜지지 않고, update 의 stdout·stderr 가 두 verbose 골든과 같으며, `update-verbose-stderr.golden` 에는 `update-clean-stderr.golden` 에 없는 줄이 1줄 이상 있고 두 verbose 골든 쌍은 비-verbose 쌍과 다르다. init 출력은 `fallback/init-clean-stderr.golden`·`init-clean-stdout.golden` 과 같다. init 의 `--help` 출력에 `--verbose` 가 있다.
- 종류: release-blocking. RED 이유: L6 — init 에 `--verbose` 가 없다. L8·L19 — 기준선이 없다.
- 녹색 경로(M3·M5): `go test ./internal/cli -run '^TestChecklistVerboseForcesFallback$' -count=1 -v` → `--- PASS: TestChecklistVerboseForcesFallback/init `, `--- PASS: TestChecklistVerboseForcesFallback/update `, `--- PASS: TestChecklistVerboseForcesFallback/update_verbose_differs `. 보조: `/usr/bin/grep -rn "Flags().Bool(\"verbose\"" internal/cli` 결과에 update 등록(`update.go:89`)과 init 등록 두 줄, 종료 0.
- 뮤턴트: init 에 플래그만 등록하고 판정에 쓰지 않는 구현 → 시접을 켠 상태에서 체크리스트가 켜져 실패. verbose stderr 경로를 깨뜨리는 구현 → stderr 골든 불일치. verbose 와 비-verbose 가 같은 출력을 내는 픽스처 → `update_verbose_differs` 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-015 — 항목 안은 접고, 밖은 넘김 (maps REQ-TCL-014)

- Given 활성 세션, When 항목이 열린 동안 `cmd.ErrOrStderr()` 에 `inside` 를 쓰고 항목이 닫힌 뒤 `outside` 를 쓰면, Then 실제 작성기에 `outside` 는 쓴 즉시 바이트 그대로 나오고 `inside` 는 성공 종료 뒤에도 나오지 않는다. 같은 항목을 `Fail` 로 닫은 경우에는 `inside` 가 펼쳐진다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistCapturePassThrough$' -count=1 -v` → `--- PASS: TestChecklistCapturePassThrough `.
- 뮤턴트: 모든 바이트를 버퍼에 모았다가 끝에 쏟는 구현 → "쓴 즉시" 검사(쓴 직후 실제 작성기 길이 확인)로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-016 — 경고는 항목 아래 한 번, 항목 밖 경고는 요약에 (maps REQ-TCL-015)

- Given 활성 init 실행 중 수집기 `Warn("w-in")`(항목 안), `Warn("w-out")`(항목 밖), `Collect("c-in")`(항목 안), `Collect("c-out")`(항목 밖), `emitSeverityLine(sevWarn, "s-in")`(항목 안), When 실행이 끝나면, Then 실제 작성기 출력에서 `w-in`·`c-in`·`s-in` 은 각각 정확히 1번(항목 아래 펼침), `w-out` 은 정확히 2번(기존과 같은 즉시 줄 + 요약 나열), `c-out` 은 정확히 1번(요약 나열)이고, `warning(s) during init:` 패널 머리 문자열은 0번이며, 요약의 경고 수는 4 — 수집기 수(결정 D-9, `internal/cli/init_warnings.go:59-63`)이며 `s-in` 은 세지 않는다. 대조로 같은 실행을 비활성으로 돌리면 패널 머리가 1번 나오고 기존 완료 카드의 경고 수도 4다.
- 종류: release-blocking. RED 이유: L7·L10. 기준선: L14 — 패널 연결이 지금 init.go 에 한 곳 있다.
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistWarningsExactlyOnce$' -count=1 -v` → `--- PASS: TestChecklistWarningsExactlyOnce `.
- 뮤턴트: `emitSeverityLine` 경로를 빠뜨린 구현 → `s-in` 이 접혀 0번 나와 실패. 패널을 그대로 둔 구현 → 패널 머리 1번으로 실패. 항목 밖 경고를 요약에 나열하지 않는 구현 → `c-out` 0번으로 실패. 심각도 줄을 경고 수에 더하는 구현 → 경고 수 5 로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-017 — 중도 종료와 성공 조기 반환 (maps REQ-TCL-016, REQ-TCL-023)

- Given 활성 세션에서 (a) 항목을 연 채 상세 `detail-x` 를 쓰고 오류를 반환하는 실행, (b) 위저드 취소 시접으로 첫 항목 전에 돌아오는 실행, (c) 중단 체크포인트가 있는 프로젝트에서 재개 경로(`init.go:858-864`)로 `nil` 을 반환하는 실행, When 끝나면, Then (a) 는 `✗` 줄과 `detail-x` 가 나오고, (b) 는 취소 문구(`Initialization cancelled.`, `init.go:575`)가 나오며, (c) 는 열린 항목이 `✓` 로 닫히고 `✗` 가 하나도 없으며 `Initialized MoAI project (resumed` 가 정확히 1번 나온다. 세 경우 모두 끝난 뒤 `cmd.OutOrStdout()`·`cmd.ErrOrStderr()` 가 원래 작성기를 돌려준다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M3·M4): `go test ./internal/cli -run '^TestChecklistAbortFlushes$' -count=1 -v` → `--- PASS: TestChecklistAbortFlushes/error `, `--- PASS: TestChecklistAbortFlushes/cancel `, `--- PASS: TestChecklistAbortFlushes/resumed_success `.
- 뮤턴트: 작성기를 되돌리지 않는 구현 → 원래 작성기 검사 실패. 취소 때 캡처를 버리는 구현 → 취소 문구 부재로 실패. 모든 조기 반환을 실패로 닫는 구현 → `resumed_success` 의 `✗` 로 실패. 재개 안내를 접는 구현 → 안내 0번으로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-018 — 출력 채널 (maps REQ-TCL-017)

- Given stdout·stderr 를 서로 다른 버퍼로 둔 활성 실행, When init 과 update 를 각각 돌리면, Then init 의 체크리스트 머리줄은 stderr 버퍼에만, update 의 머리줄은 stdout 버퍼에만 있다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistChannelPartition$' -count=1 -v` → `--- PASS: TestChecklistChannelPartition/init `, `--- PASS: TestChecklistChannelPartition/update `.
- 뮤턴트: 두 명령 모두 stderr 에 그리는 구현 → update 하위 테스트 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-019 — 질문 중 열린 줄 없음, 위저드 불변 (maps REQ-TCL-018)

- Given 병합 확인과 init 위저드 진입 지점의 시접, When 그 시접이 불릴 때 세션 상태를 기록하면, Then 열린 항목이 0개다. 그리고 이 카드의 커밋은 위저드 패키지를 바꾸지 않았다.
- 종류: release-blocking(앞 절), regression-guard(뒤 절). RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistNoOpenLineDuringPrompt$' -count=1 -v` → `--- PASS: TestChecklistNoOpenLineDuringPrompt/init `, `--- PASS: TestChecklistNoOpenLineDuringPrompt/update `. 보조: `git log --no-merges --grep=t1969 --name-only --format= cb30e05b3e..HEAD -- internal/cli/wizard` → (빈 출력), 종료 0. 이 카드의 커밋만 보므로 t1966 의 위저드 변경을 흡수해도 빨개지지 않는다.
- 뮤턴트: `analyze` 항목을 닫지 않고 확인 화면을 띄우는 구현 → 열린 항목 1개로 실패.
- 계속 발화: 앞 절은 기본 실행. 뒤 절은 이 카드 범위 확인용 일회성 검사이며 sync 감사가 다시 돈다(병합 뒤에는 멈춘다 — 의도된 한계).

### AC-TCL-020 — 숫자는 바뀌지 않음 (maps REQ-TCL-019)

- Given 같은 고정 프로젝트에서 한 번은 비활성(대체 경로), 한 번은 활성으로 돌린 init 과 update, 그리고 수집기 경고 2건과 심각도 경고 1건을 주입한 init, When 두 출력의 숫자를 뽑으면, Then init 의 디렉터리·파일 수와 경고 수(둘 다 수집기 수 2), update 의 갱신 파일 수·보존 수·경고 수가 두 실행에서 같고, 배포된 파일 목록(디스크)도 같다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistSummaryMatchesLegacyCounts$' -count=1 -v` → `--- PASS: TestChecklistSummaryMatchesLegacyCounts/init `, `--- PASS: TestChecklistSummaryMatchesLegacyCounts/update `.
- 뮤턴트: 요약이 계획표 항목 수를 "바뀐 수" 로 찍는 구현 → 숫자 불일치로 실패. 심각도 줄을 경고 수에 더하는 구현 → 활성 3 · 비활성 2 로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-021 — 배포 템플릿 불변 (maps REQ-TCL-020)

- Given 이 카드의 모든 커밋, When `git log --no-merges --grep=t1969 --name-only --format= cb30e05b3e..HEAD -- internal/template/templates` 를 돌리면, Then 출력이 비어 있고 종료 0 이다.
- 종류: regression-guard. 오늘 이미 녹색(L13 이 같은 범위의 트리 비교로 빈 출력)이므로 통과로 세지 않는다.
- 녹색 경로: 전 구간 유지. M6 와 sync 감사가 다시 돈다.
- 뮤턴트: 템플릿에 체크리스트 설정 키를 넣는 구현 → 출력에 경로가 나와 실패.
- 계속 발화: 일회성 범위 검사. 병합 뒤에는 멈춘다. 템플릿 중립성 자체는 기존 템플릿 테스트가 계속 지킨다.

### AC-TCL-022 — init.go 변경 영역 제한 (maps REQ-TCL-021)

- Given 이 카드의 커밋들과 plan.md §B 의 영역 기준 SHA(처음 값 `cb30e05b3e`), When `git log --no-merges --grep=t1969 -p -U0 --format= <영역 기준>..HEAD -- internal/cli/init.go` 의 hunk 머리를 읽으면, Then hunk 가 1개 이상이고, 모든 hunk 의 이전 쪽 범위 `-a,b` 가 한 영역 안에 있다 — `b>0` 이면 시작 줄 `a` 와 끝 줄 `a+b-1` 이 모두, `b=0`(삽입)이면 `a` 가 R1 `[383,389]`, R3 `[919,1075]`, R2 `[1077,1087]` 가운데 하나에 들며, R3 안의 hunk 는 모두 `b=0` 이다.
- 종류: release-blocking. RED 이유: L15 — 이 카드의 init.go hunk 가 0개(빈 출력, 종료 0)라 "1개 이상"이 거짓이다. L16 — 녹색 테스트가 아직 없다. 빨간 이유는 "연결이 아직 없다"이며, 범위 밖 변경 때문이 아니다.
- 녹색 경로(M4): M4 가 넣는 `TestChecklistInitHunkConfinement` 가 위 명령의 출력을 해석해 판정한다 → `go test ./internal/cli -run '^TestChecklistInitHunkConfinement$' -count=1 -v` → `--- PASS: TestChecklistInitHunkConfinement `. git 이 없거나 이력이 얕으면 건너뛰지 않고 분명한 메시지로 실패한다.
- 뮤턴트: 보고기 줄(`init.go:728`)을 고친 구현 → 시작 줄 728 이 범위 밖이라 실패. 1070 에서 시작해 1100 까지 지우는 hunk → 끝 줄 1100 이 범위 밖이라 실패. R3 안의 기존 줄을 고치는 hunk(`b>0`) → R3 삽입 전용 조건으로 실패.
- 계속 발화: 이 카드 범위의 일회성 조율 검사라 테스트는 M6 에서 지운다. 지우기 전 마지막 실행(명령·그대로의 출력·측정 SHA)을 progress.md §E.2 에 남기고, §F 완료 정의는 그 기록으로 판정한다. sync 감사는 그 기록을 읽는다.

### AC-TCL-023 — 글리프 허용 목록과 원시 키 부재 (maps REQ-TCL-001, REQ-TCL-003, REQ-TCL-009)

- Given AC-TCL-001, AC-TCL-002, AC-TCL-003, AC-TCL-004, AC-TCL-005, AC-TCL-006, AC-TCL-007, AC-TCL-009 가 만드는 모든 렌더러 출력과 `internal/tui/testdata/checklist-*.golden`, When ANSI 를 지우고 문자를 훑으면, Then 상태 자리 문자는 `✓ ✗ ! · ● ○ ◆ ◇` 또는 U+2800–28FF 점자뿐이고, 출력 어디에도 U+27A4(`➤`)와 U+2013(`–`)이 없으며, 모든 렌더러 골든에 부분 문자열 `checklist.` 가 없다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGlyphWhitelist$' -count=1 -v` → `--- PASS: TestChecklistGlyphWhitelist `.
- 뮤턴트: 카드 문구 그대로 `➤` 를 쓴 렌더러 → 실패. 카탈로그에 없는 키를 불러 `Translate` 가 키 이름을 돌려주고 그 출력으로 골든을 찍은 경우 → 원시 키 검사로 실패.
- 계속 발화: `internal/tui` 기본 실행.

### AC-TCL-024 — 보존 줄은 성공 항목에서도 한 번 (maps REQ-TCL-022)

- Given 활성 세션과 design.md §3.3.1 보존 줄 목록에서 고른 네 지점에 주입한 실패·안내 — (a) update 배포 모드 재주장 실패(`internal/cli/update_template_sync.go:948`), (b) init 설정 스냅숏 쓰기 실패(`init.go:885` → `internal/cli/update/backup/file_snapshot.go:95` → `settings_snapshot.go:115`), (c) init 지연 갱신 안내(`init.go:1058` → `internal/cli/init_update_notice.go:107`, 시접 `deferredUpdateCheck`), (d) init 워크트리 안내(`init.go:1034` → `internal/cli/worktree_advisory.go:57`), When 각 항목이 성공으로 끝나면, Then 각 문장이 실제 작성기 출력에 정확히 한 번 나오고, (a)·(b) 를 담은 항목 줄은 `!` 로, (c)·(d) 를 담은 항목 줄은 `✓` 로 끝나며, 요약의 경고 수는 주입 전과 같다(결정 D-9).
- 종류: release-blocking. RED 이유: L17 — 이 테스트가 없다. L7·L10 — 세션 코드가 없다. 오늘(비활성 경로)은 네 문장이 모두 보이지만, 체크리스트가 켜진 경로 자체가 없어 이 기준의 활성 조건을 만들 수 없다.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistPreservedLinesExactlyOnce$' -count=1 -v` → `--- PASS: TestChecklistPreservedLinesExactlyOnce/update_deploy_mode_reassert `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/init_settings_snapshot `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/init_update_notice `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/init_worktree_advisory `.
- 뮤턴트: 보존 줄 목록을 무시하고 성공 항목의 상세를 모두 접는 구현 → 네 문장 0번으로 실패. 성공 항목의 상세를 모두 펼치는 구현 → AC-TCL-015(`inside` 미노출)와 AC-TCL-011(줄 예산)이 잡는다. 보존 줄을 즉시 넘기고 종료 때 또 펼치는 구현 → 2번으로 실패. 안내를 경고로 세는 구현 → 경고 수와 `✓` 조건으로 실패.
- 계속 발화: `internal/cli` 기본 실행. 보존 줄 목록에 지점을 더하면 이 테스트에 하위 테스트를 더한다(design.md §3.3.1).

## §D 경계 사례

- 경고와 실패가 한 항목에 같이 오면 `✗` 하나로 끝내고, 경고와 상세를 모두 펼친다.
- 항목 없이 끝나는 실행(위저드 취소)은 머리줄도 그리지 않는다. 머리줄은 첫 항목이 열릴 때 그린다.
- 터미널 폭보다 긴 항목 라벨은 자르지 않는다. 자르기는 범위 밖이며 Gap G-6 로 남긴다(plan.md §I).
- `NO_COLOR` 와 모션 감소가 함께 켜져도 AC-TCL-006·AC-TCL-007 이 각각 성립해야 한다.

## §E 품질 관문

- `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...` 종료 0.
- `go vet ./internal/tui/... ./internal/cli/...` 종료 0.
- 바뀐 패키지 범위 테스트는 §C 의 기준별 명령으로 돈다. 전체 `internal/cli` 실행은 기계 부하 규칙에 따라 자원 임대(`moai slot acquire`) 아래에서 한 번, 또는 CI 에 맡긴다.
- `internal/tui` 커버리지 85% 이상: `go test -cover ./internal/tui` 출력을 그대로 기록.
- `golangci-lint run --timeout=2m` 에서 새 지적 0.

## §F 완료 정의

- AC-TCL-001 부터 AC-TCL-020 까지, AC-TCL-023, AC-TCL-024 가 §A 의 녹색 판정을 만족한다.
- AC-TCL-022 는 M6 에서 테스트를 지우기 **전** 마지막 실행 기록(명령, 그대로의 출력, 종료 코드, 측정 SHA)이 §A 의 녹색 판정을 만족한 것으로 판정한다. 지운 뒤의 `[no tests to run]` 은 판정에 쓰지 않는다.
- AC-TCL-021 은 regression-guard 로 유지된다.
- AC-TCL-013 의 순서 증거 O1·O2·O3 가 모두 성립한다.
- `go test ./internal/tui/golden -run '^(TestManifest|TestGoldenCount)$' -count=1 -v` → `--- PASS: TestManifest `, `--- PASS: TestGoldenCount `.
- progress.md §E.2 에 각 기준의 명령·그대로의 출력·종료 코드·측정 HEAD 가 기록돼 있다.
