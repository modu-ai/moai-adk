# SPEC-TUI-CHECKLIST-001 — 인수 기준

문서 핀: 모든 RED-now 측정은 트리 `cb30e05b3e`(워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`)에서 했다. 자기 핀이 없는 기준은 이 핀을 따른다.

이 파일은 검증 층이다. 요구사항(REQ)의 GEARS 문장은 spec.md §3 에 있고, 여기서는 각 기준을 Given-When-Then 과 이진 판정으로 적는다. 각 기준 제목의 `maps` 목록이 REQ 대응을 선언한다.

## §A 판정 규칙

- **명령 형식.** 모든 명령은 파이프·리다이렉트·`&&`·`;` 없는 한 번의 호출이다. 이 셸의 `grep` 은 ugrep 셸 함수라 일치 0건에서도 종료 코드 0 을 낸다(research.md §0). 그래서 모든 검색은 `/usr/bin/grep` 으로 적는다.
- **선택자 공허 방지.** `go test -run` 은 일치하는 테스트가 0개여도 `ok … [no tests to run]` 와 종료 코드 0 을 낸다(원장 L11 — 이 SPEC 의 AC-TCL-001 녹색 명령을 오늘 그대로 돌린 결과다). 그래서 테스트 기반 기준의 녹색 판정은 종료 코드 0 **그리고** 출력에 기준이 이름 붙인 `--- PASS: <테스트명> ` 줄(이름 뒤 공백 포함)이 하나도 빠짐없이 있는 것이다. 선택자는 `^…$` 로 묶고, `-v` 와 `-count=1` 을 항상 붙인다.
- **기준 종류.** release-blocking 기준은 RED-now 네 요소(명령, 그대로의 stdout, 종료 코드, 트리 SHA)를 갖춘다. 오늘 이미 녹색인 범위 경계 검사는 regression-guard 로 분류하고 통과로 세지 않는다.

## §B 증거 원장 (RED-now, 트리 `cb30e05b3e`)

stdout 이 비어 있으면 `(빈 출력)` 으로 적고, stderr 는 따로 적는다.

| ID | 명령 | stdout | stderr | 종료 코드 |
|---|---|---|---|---|
| L1 | `ls internal/tui/checklist.go` | (빈 출력) | `ls: internal/tui/checklist.go: No such file or directory` | 1 |
| L2 | `/usr/bin/grep -c "^checklist\." internal/tui/messages/en.yaml` | `0` | (없음) | 1 |
| L3 | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ko.yaml` | `0` | (없음) | 1 |
| L4 | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ja.yaml` | `0` | (없음) | 1 |
| L5 | `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` | `0` | (없음) | 1 |
| L6 | `/usr/bin/grep -n "\"verbose\"" internal/cli/init.go` | (빈 출력) | (없음) | 1 |
| L7 | `/usr/bin/grep -l "checklist" internal/cli/init.go internal/cli/update.go internal/cli/update_template_sync.go internal/cli/init_warnings.go internal/cli/severity_line.go internal/cli/reporter.go` | (빈 출력) | (없음) | 1 |
| L8 | `ls internal/cli/testdata/checklist` | (빈 출력) | `ls: internal/cli/testdata/checklist: No such file or directory` | 1 |
| L9 | `/usr/bin/grep -rn "func NewChecklist" internal/tui` | (빈 출력) | (없음) | 1 |
| L10 | `ls internal/tui/checklist_test.go internal/cli/checklist_session.go internal/cli/checklist_session_test.go` | (빈 출력) | `ls: internal/cli/checklist_session_test.go: No such file or directory` / `ls: internal/cli/checklist_session.go: No such file or directory` / `ls: internal/tui/checklist_test.go: No such file or directory` | 1 |
| L11 | `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/tui	0.301s [no tests to run]` | (없음) | 0 |
| L12 | `/usr/bin/grep -c "ProgressLine(" internal/cli/update_template_sync.go` | `10` | (없음) | 0 |
| L13 | `git diff --name-only cb30e05b3e -- internal/template/templates` | (빈 출력) | (없음) | 0 |
| L14 | `/usr/bin/grep -c "emitSummary" internal/cli/init.go` | `1` | (없음) | 0 |

- L11 은 RED 가 아니라 **대조군**이다. 아직 없는 테스트를 고르는 선택자가 `PASS` 와 종료 코드 0 을 낸다. §A 의 `--- PASS:` 요구가 여기서 나온다. 종료 코드만 보는 판정은 이 SPEC 의 모든 테스트 기반 기준을 오늘 녹색으로 만든다.
- L12·L14 는 기준선이다. L13 은 오늘 이미 비어 있어 AC-TCL-021 은 regression-guard 다.

## §C 인수 기준

각 기준의 "RED 이유" 는 오늘 왜 빨간지, "녹색 경로" 는 어느 마일스톤이 무엇으로 뒤집는지를 적는다. 뮤턴트는 기준을 통과하면서 요구를 어기는 구현을 시도해 본 결과다. "계속 발화" 는 이 검사가 멈췄을 때 누가 어떻게 알게 되는지다.

### AC-TCL-001 — 깨끗한 실행의 체크리스트 모양 (maps REQ-TCL-001, REQ-TCL-003)

- Given 고정 시계·`LightTheme`·고정 환경 요약과 design.md §4 의 init·update 계획표, When 모든 항목을 `Done` 으로 끝내고 요약을 그리면, Then 출력이 `internal/tui/testdata/checklist-init-clean.golden` 과 `checklist-update-clean.golden` 에 바이트 단위로 같고 머리줄 1개·환경 줄 1개·그룹 제목·항목당 한 줄 구조를 갖는다.
- 종류: release-blocking. RED 이유: L1·L9·L10 — 렌더러와 테스트가 없다. L11 이 같은 명령의 오늘 출력이며 `--- PASS:` 줄이 없다.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` → 종료 0, `--- PASS: TestChecklistGoldenClean ` 포함.
- 뮤턴트: 항목마다 두 줄(진행 줄 + 결과 줄)을 남기는 렌더러 → 골든 줄 수가 달라 실패. 골든을 렌더러 출력으로 덮어쓰는 뮤턴트 → AC-TCL-011 줄 예산과 AC-TCL-002 시퀀스 검사가 따로 잡는다.
- 계속 발화: 기본 `go test ./internal/tui` 실행에 포함. 골든을 지우거나 새로 만들면 `TestManifest`(`internal/tui/golden/index_test.go:44-73`)가 `MANIFEST.txt` 불일치로 실패한다.

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

- Given 바뀐 수 7, 보존 수 3, 경고 1, 고정 시계로 1.5초 경과, 백업 경로와 복구 명령을 담은 update 요약과, 파일 2·디렉터리 13·배포 모드·다음 단계를 담은 init 요약, When 그리면, Then 각 숫자와 문자열이 요약에 한 번씩 나오고 경과 시간이 `1.5s` 로 보인다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistSummaryFields$' -count=1 -v` → `--- PASS: TestChecklistSummaryFields `.
- 뮤턴트: 백업 경로·복구 명령을 빠뜨린 요약 → 문자열 부재로 실패. 시계를 무시하고 `time.Now` 를 쓰는 렌더러 → `1.5s` 불일치로 실패.
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

- Given en·ko·ja·zh 카탈로그, When `checklist.` 로 시작하는 키 집합을 비교하면, Then 네 집합이 같고, 값이 모두 비어 있지 않고 키 자신과 다르며, 같은 키의 `%` 서식 지시자 개수가 네 언어에서 같고, 값에 U+1F300–1FAFF·U+2600–27BF·U+1F000–1F2FF 문자가 없다. en 키 수는 design.md §6 표의 키 수와 같다.
- 종류: release-blocking. RED 이유: L2·L3·L4·L5 — 네 카탈로그 모두 `checklist.` 키가 0개다.
- 녹색 경로(M1): `go test ./internal/tui -run '^TestChecklistCatalogFourLocales$' -count=1 -v` → `--- PASS: TestChecklistCatalogFourLocales `. 보조 관측: `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` → en 과 같은 양수, 종료 0.
- 뮤턴트: ja 에 키를 빠뜨리면 `Translate` 가 en 으로 대체해 화면은 멀쩡해 보인다 → 집합 비교가 잡는다. 값에 키 이름을 그대로 넣는 뮤턴트 → 값 ≠ 키 조건이 잡는다.
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

### AC-TCL-011 — 한 화면 24줄 (maps REQ-TCL-010)

- Given 활성 판정 시접을 강제로 켠 깨끗한 init 실행과 깨끗한 update 실행, When 출력에서 `\r\x1b[2K` 로 덮인 줄을 최종 화면으로 접고 ANSI 를 지운 뒤 줄을 세면, Then 둘 다 24줄 이하다. 접은 결과는 `internal/cli/testdata/checklist/rich-init-clean.golden`, `rich-update-clean.golden` 과 같다.
- 종류: release-blocking. RED 이유: L8·L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistScreenBudget$' -count=1 -v` → `--- PASS: TestChecklistScreenBudget/init `, `--- PASS: TestChecklistScreenBudget/update `.
- 뮤턴트: 넘김 규칙을 잘못 걸어 항목 안 상세가 새는 구현 → 줄 수 초과로 실패.
- 계속 발화: `internal/cli` 기본 실행. 골든이 없으면 `t.Fatalf` 로 실패하고 건너뛰지 않는다(기존 `tuxReadGolden` 과 같은 방식, `internal/cli/tuxiu_characterization_test.go:115-122`).

### AC-TCL-012 — 활성 판정 (maps REQ-TCL-011)

- Given design.md §3.1 의 다섯 조건, When 조건마다 하나씩만 거짓으로 바꾼 다섯 경우와 모두 참인 한 경우를 판정하면, Then 모두 참일 때만 활성이고 다섯 경우는 모두 비활성이다.
- 종류: release-blocking. RED 이유: L7·L10.
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistActivationPredicate$' -count=1 -v` → `--- PASS: TestChecklistActivationPredicate ` 와 하위 테스트 6개의 `--- PASS:` 줄.
- 뮤턴트: 조건 하나(예: `TERM=dumb`)를 빼먹은 판정 → 해당 하위 테스트 실패. 늘 비활성인 판정 → "모두 참" 하위 테스트 실패(양방향).
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-013 — 대체 경로는 오늘과 같은 바이트 (maps REQ-TCL-012)

- Given M0 에서 코드 변경 **전에** 별도 커밋으로 넣은 `internal/cli/testdata/checklist/fallback-init-stderr.golden`, `fallback-init-stdout.golden`, `fallback-update-stdout.golden`, `fallback-update-stderr.golden`, When 버퍼 작성기(비TTY)로 같은 고정 실행을 돌리면, Then 출력이 변동 값 정규화(백업 시각, 버전 문자열, Go 런타임 문자열) 뒤 네 골든과 바이트 단위로 같고, 정규화 뒤 골든의 줄 수가 0 이 아니다.
- 종류: release-blocking. RED 이유: L8 — 기준선 골든이 없다. 이 기준은 기준선 커밋이 구현 커밋보다 앞서야 성립한다(design.md §8).
- 녹색 경로(M0 에서 녹색으로 들어오고, M4·M5 뒤에도 녹색이어야 함): `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` → `--- PASS: TestChecklistFallbackMatchesBaseline/init `, `--- PASS: TestChecklistFallbackMatchesBaseline/update `. 순서 증거: 기준선 골든을 처음 넣은 커밋(`git log --format=%h --diff-filter=A cb30e05b3e..HEAD -- internal/cli/testdata/checklist/fallback-init-stderr.golden`)이 렌더러를 처음 넣은 커밋(`git log --format=%h --diff-filter=A cb30e05b3e..HEAD -- internal/tui/checklist.go`)의 진짜 조상이다 — `git merge-base --is-ancestor <기준선 SHA> <렌더러 SHA>` 종료 0, 두 SHA 는 서로 다르다.
- 뮤턴트: 기준선을 구현 뒤에 다시 찍는 절차 → 조상 검사가 실패. 모든 줄을 지우는 정규화 → 줄 수 0 이 아님 조건이 잡는다.
- 계속 발화: `internal/cli` 기본 실행. 골든 부재는 `t.Fatalf`.

### AC-TCL-014 — `--verbose` (maps REQ-TCL-013)

- Given init 과 update, When `--verbose` 를 주고 활성 판정 시접을 켠 상태로 돌리면, Then 체크리스트가 켜지지 않고, update 출력은 M0 의 `fallback-update-verbose-stdout.golden` 과 같으며, init 출력은 `fallback-init-stderr.golden` 과 같다. init 의 `--help` 출력에 `--verbose` 가 있다.
- 종류: release-blocking. RED 이유: L6 — init 에 `--verbose` 가 없다. L8 — 기준선이 없다.
- 녹색 경로(M3·M5): `go test ./internal/cli -run '^TestChecklistVerboseForcesFallback$' -count=1 -v` → `--- PASS: TestChecklistVerboseForcesFallback/init `, `--- PASS: TestChecklistVerboseForcesFallback/update `. 보조: `/usr/bin/grep -rn "Flags().Bool(\"verbose\"" internal/cli` 결과에 update 등록(`update.go:89`)과 init 등록 두 줄, 종료 0.
- 뮤턴트: init 에 플래그만 등록하고 판정에 쓰지 않는 구현 → 시접을 켠 상태에서 체크리스트가 켜져 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-015 — 항목 안은 접고, 밖은 넘김 (maps REQ-TCL-014)

- Given 활성 세션, When 항목이 열린 동안 `cmd.ErrOrStderr()` 에 `inside` 를 쓰고 항목이 닫힌 뒤 `outside` 를 쓰면, Then 실제 작성기에 `outside` 는 쓴 즉시 바이트 그대로 나오고 `inside` 는 성공 종료 뒤에도 나오지 않는다. 같은 항목을 `Fail` 로 닫은 경우에는 `inside` 가 펼쳐진다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistCapturePassThrough$' -count=1 -v` → `--- PASS: TestChecklistCapturePassThrough `.
- 뮤턴트: 모든 바이트를 버퍼에 모았다가 끝에 쏟는 구현 → "쓴 즉시" 검사(쓴 직후 실제 작성기 길이 확인)로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-016 — 경고는 정확히 한 번 (maps REQ-TCL-015)

- Given 활성 init 실행 중 수집기 `Warn` 2번(항목 안 1, 항목 밖 1), `Collect` 1번(항목 안), `emitSeverityLine(sevWarn)` 1번(항목 안), When 실행이 끝나면, Then 네 경고 문장이 실제 작성기 출력에 각각 정확히 한 번 나오고, `warning(s) during init:` 패널 머리 문자열은 0번 나오며, 요약의 경고 수는 4다. 대조로 같은 실행을 비활성으로 돌리면 패널 머리가 1번 나온다.
- 종류: release-blocking. RED 이유: L7·L10. 기준선: L14 — 패널 연결이 지금 init.go 에 한 곳 있다.
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistWarningsExactlyOnce$' -count=1 -v` → `--- PASS: TestChecklistWarningsExactlyOnce `.
- 뮤턴트: `emitSeverityLine` 경로를 빠뜨린 구현 → 그 경고가 접혀 0번 나와 실패. 패널을 그대로 둔 구현 → 2번 나와 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-017 — 중도 종료 시 쏟아내기 (maps REQ-TCL-016)

- Given 활성 세션에서 항목을 연 채 상세 `detail-x` 를 쓰고 오류를 반환하는 실행, 그리고 위저드 취소 시접으로 첫 항목 전에 돌아오는 실행, When 끝나면, Then 첫 실행은 `✗` 줄과 `detail-x` 가 나오고, 둘째 실행은 취소 문구(`Initialization cancelled.`, `init.go:575`)가 나오며, 두 경우 모두 끝난 뒤 `cmd.OutOrStdout()`·`cmd.ErrOrStderr()` 가 원래 작성기를 돌려준다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M3·M4): `go test ./internal/cli -run '^TestChecklistAbortFlushes$' -count=1 -v` → `--- PASS: TestChecklistAbortFlushes/error `, `--- PASS: TestChecklistAbortFlushes/cancel `.
- 뮤턴트: 작성기를 되돌리지 않는 구현 → 원래 작성기 검사 실패. 취소 때 캡처를 버리는 구현 → 취소 문구 부재로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-018 — 출력 채널 (maps REQ-TCL-017)

- Given stdout·stderr 를 서로 다른 버퍼로 둔 활성 실행, When init 과 update 를 각각 돌리면, Then init 의 체크리스트 머리줄은 stderr 버퍼에만, update 의 머리줄은 stdout 버퍼에만 있다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistChannelPartition$' -count=1 -v` → `--- PASS: TestChecklistChannelPartition/init `, `--- PASS: TestChecklistChannelPartition/update `.
- 뮤턴트: 두 명령 모두 stderr 에 그리는 구현 → update 하위 테스트 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-019 — 질문 중 열린 줄 없음, 위저드 불변 (maps REQ-TCL-018)

- Given 병합 확인과 init 위저드 진입 지점의 시접, When 그 시접이 불릴 때 세션 상태를 기록하면, Then 열린 항목이 0개다. 그리고 위저드 패키지는 바뀌지 않았다.
- 종류: release-blocking(앞 절), regression-guard(뒤 절). RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistNoOpenLineDuringPrompt$' -count=1 -v` → `--- PASS: TestChecklistNoOpenLineDuringPrompt/init `, `--- PASS: TestChecklistNoOpenLineDuringPrompt/update `. 보조: `git diff --name-only cb30e05b3e -- internal/cli/wizard` → (빈 출력), 종료 0.
- 뮤턴트: `analyze` 항목을 닫지 않고 확인 화면을 띄우는 구현 → 열린 항목 1개로 실패.
- 계속 발화: 앞 절은 기본 실행. 뒤 절은 이 카드 범위 확인용 일회성 검사이며 sync 감사가 다시 돈다(병합 뒤에는 멈춘다 — 의도된 한계).

### AC-TCL-020 — 숫자는 바뀌지 않음 (maps REQ-TCL-019)

- Given 같은 고정 프로젝트에서 한 번은 비활성(대체 경로), 한 번은 활성으로 돌린 init 과 update, When 두 출력의 숫자를 뽑으면, Then init 의 디렉터리·파일·경고 수, update 의 갱신 파일 수·보존 수·경고 수가 두 실행에서 같고, 배포된 파일 목록(디스크)도 같다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M4·M5): `go test ./internal/cli -run '^TestChecklistSummaryMatchesLegacyCounts$' -count=1 -v` → `--- PASS: TestChecklistSummaryMatchesLegacyCounts/init `, `--- PASS: TestChecklistSummaryMatchesLegacyCounts/update `.
- 뮤턴트: 요약이 계획표 항목 수를 "바뀐 수" 로 찍는 구현 → 숫자 불일치로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-021 — 배포 템플릿 불변 (maps REQ-TCL-020)

- Given 이 카드의 모든 커밋, When `git diff --name-only cb30e05b3e -- internal/template/templates` 를 돌리면, Then 출력이 비어 있고 종료 0 이다.
- 종류: regression-guard. 오늘 이미 녹색(L13)이므로 통과로 세지 않는다.
- 녹색 경로: 전 구간 유지. M6 와 sync 감사가 다시 돈다.
- 뮤턴트: 템플릿에 체크리스트 설정 키를 넣는 구현 → 출력에 경로가 나와 실패.
- 계속 발화: 일회성 범위 검사. 병합 뒤에는 멈춘다. 템플릿 중립성 자체는 기존 템플릿 테스트가 계속 지킨다.

### AC-TCL-022 — init.go 변경 영역 제한 (C-2, t1966 조율)

- Given 이 카드의 커밋들, When `git diff -U0 cb30e05b3e -- internal/cli/init.go` 를 돌리면, Then hunk 가 1개 이상이고, 모든 hunk 머리의 이전 쪽 시작 줄이 R1 `[383,389]`, R3 `[919,1075]`, R2 `[1077,1087]` 가운데 하나에 든다.
- 종류: release-blocking. RED 이유: 오늘 hunk 가 0개라 "1개 이상" 이 거짓이다(L7 이 init.go 에 체크리스트 연결이 없음을 보인다). 빨간 이유는 "연결이 아직 없다" 이며, 범위 밖 변경 때문이 아니다.
- 녹색 경로(M4): M4 가 넣는 `TestChecklistInitHunkConfinement` 가 `git diff` 출력을 해석해 판정한다 → `go test ./internal/cli -run '^TestChecklistInitHunkConfinement$' -count=1 -v` → `--- PASS: TestChecklistInitHunkConfinement `. git 이 없는 환경에서는 건너뛰지 않고 분명한 메시지로 실패한다.
- 뮤턴트: 보고기 줄(`init.go:728`)을 고친 구현 → 그 hunk 의 시작 줄 728 이 범위 밖이라 실패.
- 계속 발화: 기준 SHA 를 고정했으므로 이 카드가 병합된 뒤에는 의미가 없다. 그래서 테스트는 M6 에서 지운다(일회성 조율 검사). sync 감사는 지우기 전 마지막 출력을 증거로 남긴다.

### AC-TCL-023 — 글리프 허용 목록 (maps REQ-TCL-001, REQ-TCL-003)

- Given AC-TCL-001~009 가 만드는 모든 렌더러 출력, When ANSI 를 지우고 문자를 훑으면, Then 상태 자리 문자는 `✓ ✗ ! · ● ○ ◆ ◇` 또는 U+2800–28FF 점자뿐이고, 출력 어디에도 U+27A4(`➤`)와 U+2013(`–`)이 없다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGlyphWhitelist$' -count=1 -v` → `--- PASS: TestChecklistGlyphWhitelist `.
- 뮤턴트: 카드 문구 그대로 `➤` 를 쓴 렌더러 → 실패.
- 계속 발화: `internal/tui` 기본 실행.

## §D 경계 사례

- 경고와 실패가 한 항목에 같이 오면 `✗` 하나로 끝내고, 경고와 상세를 모두 펼친다.
- 항목 없이 끝나는 실행(위저드 취소)은 머리줄도 그리지 않는다. 머리줄은 첫 항목이 열릴 때 그린다.
- 터미널 폭보다 긴 항목 라벨은 자르지 않는다. 자르기는 범위 밖이며 Gap G-6 로 남긴다(plan.md §I).
- `NO_COLOR` 와 모션 감소가 함께 켜져도 AC-TCL-006·007 이 각각 성립해야 한다.

## §E 품질 관문

- `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...` 종료 0.
- `go vet ./internal/tui/... ./internal/cli/...` 종료 0.
- 바뀐 패키지 범위 테스트는 §C 의 기준별 명령으로 돈다. 전체 `internal/cli` 실행은 기계 부하 규칙에 따라 자원 임대(`moai slot acquire`) 아래에서 한 번, 또는 CI 에 맡긴다.
- `internal/tui` 커버리지 85% 이상: `go test -cover ./internal/tui` 출력을 그대로 기록.
- `golangci-lint run --timeout=2m` 에서 새 지적 0.

## §F 완료 정의

- AC-TCL-001~020, 022, 023 이 §A 의 녹색 판정을 만족한다. AC-TCL-021 은 regression-guard 로 유지된다.
- `go test ./internal/tui/golden -run '^(TestManifest|TestGoldenCount)$' -count=1 -v` → `--- PASS: TestManifest `, `--- PASS: TestGoldenCount `.
- progress.md §E.2 에 각 기준의 명령·그대로의 출력·종료 코드·측정 HEAD 가 기록돼 있다.
- 기준선 커밋이 렌더러 커밋의 조상이다(AC-TCL-013 순서 증거).
