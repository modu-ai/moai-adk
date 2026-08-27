# SPEC-CODEX-INIT-001 — AC 뮤턴트 분석 (적대적 감사)

대상: `.moai/specs/SPEC-CODEX-INIT-001/acceptance.md` AC-CI-001 ~ AC-CI-011 (11건 전수, 표본 없음).
방법: 각 AC 에 대해 **그 AC 의 모든 단언을 문자 그대로 만족하면서 그 AC 가 인용한 REQ 를 위반하는 구현**(뮤턴트)을 실제로 작성 시도한다. 작성에 성공하면 MUTANT-WRITABLE.
REQ 원문은 `spec.md` §E 에서 그대로 인용한다 (영문 원문 유지).

**전 AC 를 관통하는 한 뮤턴트를 먼저 기록한다** — 개별 절에서 반복하지 않기 위해서다.

> **M-SPAWN (기동 seam 이 두 개다).** 이 SPEC 의 모든 음성 판정은 `exec 0회` 로 기동 부재를 증명한다. 그런데 런처 SPEC 의 AC-CL-003 이 고정한 대로 기동 seam 은 **둘**이다 — `exec` 과 `spawnLaunch`(`--spawn`). 게이트를 `--spawn` 경로에서만 건너뛰는 구현(예: `if !opts.Spawn { gate() }`)은 `exec` 호출이 영원히 0 이므로 AC-CI-001·003·004·010·011 의 `exec 0회` 단언을 **전부** 만족하면서, `moai codex cli --spawn` 으로 미배선 세션을 새 창에 띄운다. REQ-CI-002 의 *"The offer shall bind both launch verbs identically"* 와 REQ-CI-010 의 *"the system shall not launch"* 를 정면으로 위반한다.
> **조이는 방법**: 모든 음성 단언을 `exec 호출 0회 AND spawnLaunch 호출 0회` 로 바꾸고, 최소한 AC-CI-001 의 12칸과 AC-CI-010 의 6칸을 `--spawn` 유무 2종으로 다시 곱한다.

---

## AC-CI-001
- **상태**: MUTANT-WRITABLE
- **mutant**: 게이트가 런처 분류기를 **부르지 않고**, `codex_init.go` 안에서 `os.Stat(filepath.Join(root, ".codex", "hooks.json"))` · `os.Stat(... "config.toml")` + `codexadapter.ValidateConfig` 로 상태를 **자체 재구현** 한다. 판정 결과는 S1~S6 정의와 동일하게 맞춰 놓는다. 두 동사는 같은 함수를 통과하므로 12칸이 전부 기대대로 나온다.
  (부수 뮤턴트 2종: ① 제안을 띄운 뒤 응답과 무관하게 exec 한다 — AC-CI-001 은 *"제안 전 exec 0회"* 만 세므로 통과한다. ② M-SPAWN.)
- **위반되는 REQ 조항**: REQ-CI-001 — *"It shall obtain that judgement by calling the readout's wiring classifier and acting on the state that call returns — calling it and then deciding by other means is the same defect as defining a second test."*
- **왜 AC 가 못 잡는가**: 12칸 전부에서 **디스크 배치와 참 상태가 일치** 하므로, 자체 재분류의 출력은 분류기 출력과 구별되지 않는다. AC-CI-001 에는 분류기 호출 횟수 단언도, 모순 주입도, 파일 검사 grep 단언도 없다 — 그 셋은 전부 AC-CI-002 에만 있고, AC-CI-002 는 뒤에서 보듯 동사 축을 돌지 않는다. 즉 12칸은 "결과가 맞다" 만 증명하고 "출처가 하나다" 는 증명하지 않는다.
- **조이는 방법**: 12칸 각각에서 (a) 분류기 호출 횟수 ≥1, (b) 12칸 중 최소 2칸은 분류기가 디스크와 **모순되는** 상태를 반환하도록 스텁하고 결과가 반환값을 따를 것, (c) `exec 0회` 를 "제안 전" 이 아니라 **실행 종료 시점** 기준으로 셀 것 — 이 세 줄을 AC-CI-001 자체에 넣는다.

## AC-CI-002
- **상태**: MUTANT-WRITABLE
- **mutant**: `cli` 동사는 주입된 반환 상태를 그대로 소비하고, `app` 동사는 분류기를 부른 뒤 **반환값을 버리고** 디스크에서 자체 재분류한다 (`hooksExists && tomlExists` 두 줄). AC-CI-002 의 8칸을 `cli` 로 돌면 여덟 칸 모두 통과하고, AC-CI-001 의 12칸은 디스크와 참 상태가 일치하므로 `app` 도 통과한다. 파일 검사 grep 은 상수 결합(`".codex"` + `string(filepath.Separator)` + `"hooks"+".json"`)이나 별도 헬퍼 패키지 호출로 회피한다.
- **위반되는 REQ 조항**: REQ-CI-002 — *"The offer shall bind both launch verbs identically; a verb that launches into an incomplete project while its sibling gates is a defect, not a variation."* (그리고 REQ-CI-001 의 *"acting on the state that call returns"*)
- **왜 AC 가 못 잡는가**: AC-CI-002 는 **주입 상태 × 디스크 배치** 2축만 곱하고 **동사 축을 곱하지 않는다** ("8칸" 표에 동사 열이 없다). 표가 어느 동사로 도는지 명시되지 않아, 한 동사만 도는 시험이 규격을 충족한다. 부수적으로 *"파일 검사 grep 0건"* 은 리터럴 문자열 검색이라 계산된 경로·간접 호출을 통과시키며, *"분류기 호출 횟수 ≥1회"* 는 반환값 소비를 증명하지 못한다 (AC 자신이 서두에서 인정한 결함이 grep 단언에는 그대로 남아 있다).
- **조이는 방법**: 표를 **주입 상태 4 × 디스크 배치 2 × 동사 2 = 16칸** 으로 확장하고, "파일 검사 0건" 을 grep 이 아니라 **파일시스템 seam 계수** 로 바꾼다 — 게이트 구간에서 `.codex/` 하위를 `Stat`/`Open` 한 횟수가 0 임을 스텁 FS 로 관측한다 (문자열이 아니라 행위를 센다).

## AC-CI-003
- **상태**: MUTANT-WRITABLE
- **mutant**: 거절 시 프로젝트 트리에는 아무것도 쓰지 않되, `CODEX_HOME/.moai-declined` 에 거절 기록을 append 한다 (또는 `$HOME/.cache/moai/codex-decline.log`). 10칸 전부에서 생성기 0회·exec 0회·프로젝트 스냅샷 무변경·취소 종료 코드·상태 문구가 성립한다.
  (부수 뮤턴트: 기존 `CLAUDE.md` 를 읽어 재작성한 뒤 `os.Chtimes` 로 mtime 을 복원한다 — 목록·mtime 스냅샷은 내용 변화를 보지 못한다.)
- **위반되는 REQ 조항**: REQ-CI-003 — *"Where the operator declines, the system shall write nothing and shall not launch"*
- **왜 AC 가 못 잡는가**: 스냅샷의 범위가 *"프로젝트 트리 전체"* 로 한정돼 있어 트리 밖 쓰기를 원리적으로 관측하지 않는다. AC-CI-008(트리 밖 무쓰기)은 **I1~I5 · L1 · L2 계약 fixture 에만** 적용되며 거절 경로를 돌지 않는다. 또한 스냅샷 축이 *목록 + mtime* 이라 동일 크기 in-place 재작성 + mtime 복원을 통과시킨다.
- **조이는 방법**: 거절 10칸의 스냅샷을 (a) **내용 해시** 기반으로 바꾸고, (b) 범위를 프로젝트 트리 + 임시 HOME + 임시 `CODEX_HOME` + **프로젝트의 부모 디렉터리** 로 넓힌다 (AC-CI-008 과 같은 스냅샷 함수를 공유하게 한다).

## AC-CI-004
- **상태**: MUTANT-WRITABLE
- **mutant**: 수락 시 생성기를 정확히 1회 (`--agent codex` 인자 포함) 호출한 뒤, **지시 계약 단계를 아예 실행하지 않고** 곧바로 기동한다. `codex_contract.go` 는 파일로는 존재하고 단위 시험도 붙어 있으나 게이트에서 호출되지 않는다 (배선 없는 죽은 코드).
  (부수 뮤턴트 2종: ① 수락 시 생성기만 부르고 **기동하지 않은 채** 성공 종료한다 — AC-CI-004 에 exec 단언이 없다. ② 비대화형 판정을 "표준입력이 파이프인가" 로만 하여, TTY 이면서 응답 불가인 환경에서 프롬프트를 발행한다 — 시험 환경은 파이프이므로 10칸이 전부 프롬프트 0회로 통과한다.)
- **위반되는 REQ 조항**: REQ-CI-005 — *"Initialization shall ensure the project carries an `AGENTS.md` and a `CLAUDE.md` linked by a single import directive, so both harnesses resolve one instruction source."*
- **왜 AC 가 못 잡는가**: AC-CI-004 의 단언은 **생성기 호출 횟수·인자**, **배선 파일 쓰기 grep 0건**, 그리고 비대화형 블록의 **프롬프트 발행 0회** 뿐이다. 계약 단계의 실행 여부에 대한 단언이 한 줄도 없다. 계약 자체는 AC-CI-005~007 이 판정하지만, 그 AC 들은 "초기화" 를 계약 함수에 직접 걸어도 성립하므로 **게이트에서의 도달성** 은 어느 AC 도 증명하지 않는다 — SPEC 전체를 통틀어 "수락 → 계약이 실제로 쓰인다" 를 세는 곳이 없다.
- **조이는 방법**: 수락 10칸에 세 줄을 추가한다 — (a) 계약 확보 함수 호출 횟수가 정확히 1이고 그 호출이 생성기 호출 **이후** 다, (b) 수락 후 실제 디스크에 `AGENTS.md` · `CLAUDE.md` 가 존재하고 `@AGENTS.md` 줄이 1건이다, (c) 성공 수락 경로에서 exec 이 정확히 1회다.

## AC-CI-005
- **상태**: MUTANT-WRITABLE
- **mutant**: `@AGENTS.md` 줄을 **실행되지 않는 위치** 에 넣는다. 구체적으로 `CLAUDE.md` 끝에 `<!-- generated -->` 주석과 세 백틱 코드펜스를 붙이고 그 **펜스 안에** `@AGENTS.md` 한 줄을 넣는다. 줄 자체는 파일에 정확히 1건 존재하고, 기존 사용자 내용은 부분 문자열로 온전하며, `AGENTS.md` 는 바이트 무변경이다. 실제 하네스는 코드펜스 안의 `@AGENTS.md` 를 import 로 해석하지 않는다.
  (부수 뮤턴트: I1·I3 에서 `AGENTS.md` 를 **0바이트** 로 만든다 — "carries an AGENTS.md" 를 문자 그대로 만족하면서 지시 원본이 비어 있다. AC-CI-005 는 새로 만든 파일의 내용에 대해 아무 단언도 하지 않는다.)
- **위반되는 REQ 조항**: REQ-CI-005 — *"...linked by a single import directive, so both harnesses resolve one instruction source."*
- **왜 AC 가 못 잡는가**: AC 는 *"`@AGENTS.md` 줄 수는 정확히 1"* 이라는 **줄 세기** 로만 판정한다. 서두에서 "판정 대상 형태를 먼저 고정한다" 고 선언했지만 고정된 것은 *문자열의 모양* 이지 *그 줄이 놓이는 위치의 실행 가능성* 이 아니다. AC-CI-007 의 도달성 시험도 시험 쪽 follower 가 `^@(\S+)` 를 순진하게 정규식 매칭하면 펜스 안의 줄을 그대로 따라가므로 함께 통과한다.
- **조이는 방법**: (a) 그 줄이 **줄머리에서 시작하고 코드펜스·인용블록·HTML 주석 안이 아님** 을 단언한다 (fixture 에 펜스 안 `@AGENTS.md` 를 심은 음성 케이스 I6 을 추가해, 그것을 "이미 있음" 으로 오인해 줄을 안 넣는 구현이 떨어지게 한다). (b) 새로 생성된 `AGENTS.md` 가 **비어 있지 않다** 를 단언한다.

## AC-CI-006
- **상태**: MUTANT-WRITABLE
- **mutant**: 링크 존재 판정을 `strings.HasSuffix(content, "@AGENTS.md\n")` 로 구현한다. 즉 **파일 끝에 있을 때만** "이미 있음" 으로 본다. I5 fixture 의 `CLAUDE.md` 가 그 줄을 마지막 줄에 담고 있으면 7종 × 2회가 전부 바이트 동일로 통과한다. 그러나 이 저장소의 실제 `CLAUDE.md` 처럼 import 가 **문서 앞부분(§0)** 에 있는 파일에서는 매 실행마다 줄이 하나씩 추가돼 비멱등이 된다.
  (부수 뮤턴트: 판정 문자열을 `"@AGENTS.md\n"` 로 고정 — CRLF 로 저장된 `CLAUDE.md` 에서는 매 실행 append. Windows 사용자에게만 재현되고 AC-CI-009 의 `GOOS=windows go vet` 은 컴파일만 증명하므로 잡히지 않는다.)
- **위반되는 REQ 조항**: REQ-CI-007 — *"Initialization shall be idempotent: running it twice shall leave every instruction file byte-identical to its state after the first run."*
- **왜 AC 가 못 잡는가**: 멱등은 **fixture 의 형태** 에 의존하는데, I4·I5 의 정의가 *"둘 다 있는데 줄이 없음"* · *"둘 다 있고 줄도 있음"* 뿐이라 **줄의 위치와 개행 코드가 미규정** 이다. 시험 작성자가 자연스럽게 파일 끝에 두면 접미사 매칭 뮤턴트가 통과한다. 2회 실행 비교라는 축은 옳지만, 비교 대상이 뮤턴트가 이미 안정된 형태(끝줄)뿐이다.
- **조이는 방법**: I5 를 세 변종으로 쪼갠다 — 줄이 **파일 맨 앞**, **중간**, **끝** 인 세 fixture, 그리고 CRLF 저장 변종 1종. 그리고 2회가 아니라 **3회** 실행해 1↔2, 2↔3 두 비교를 모두 건다 (2회차에서만 안정되는 구현을 배제).

## AC-CI-007
- **상태**: MUTANT-WRITABLE
- **mutant**: `AGENTS.md` 에 `@CLAUDE.local.md` 를 1건 넣고(설계대로), **동시에** `AGENTS.md` 가 import 하는 하위 파일(예: 생성기가 까는 `.codex/AGENTS.md` 또는 새로 만든 `AGENTS.internal.md`)에도 같은 import 를 1건 넣는다. 두 진입 파일에서 sentinel 이 도달하고, `CLAUDE.local.md` 는 바이트 무변경이며, **`AGENTS.md` 와 `CLAUDE.md` 두 파일만 세면 참조는 정확히 1건** 이다. 실제 로딩에서는 같은 내용이 두 번 들어온다.
- **위반되는 REQ 조항**: REQ-CI-008 — *"...shall make its content reachable from both harnesses through the import chain, referenced from exactly one place so it is not loaded twice"*
- **왜 AC 가 못 잡는가**: 참조 개수를 세는 **범위가 `AGENTS.md` 와 `CLAUDE.md` 두 파일로 한정** 되어 있다 (*"`AGENTS.md` 와 `CLAUDE.md` 를 통틀어 정확히 1건"*). REQ 가 금지하는 것은 "두 번 로드" 인데, 로드는 전이 import 폐포 전체에서 일어나므로 측정 범위가 요구 범위보다 좁다. AC 가 이미 도달성 시험을 위해 폐포를 순회하면서도, 세는 것은 두 파일뿐이라는 비대칭이 그대로 구멍이다.
- **조이는 방법**: 참조 개수를 **도달성 순회로 수집한 전이 폐포 전체** 에서 세고 정확히 1건임을 단언한다 (도달성과 계수에 같은 순회 결과를 쓴다). L2 의 "0건" 단언도 같은 폐포 범위로 바꾼다.

## AC-CI-008
- **상태**: MUTANT-WRITABLE
- **mutant**: 초기화가 **프로젝트의 부모 디렉터리** 에 캐시를 쓴다 — 예컨대 `filepath.Join(projectRoot, "..", ".moai-codex-init.state")`. 임시 HOME 과 임시 `CODEX_HOME` 두 스냅샷은 완전히 동일하게 유지된다.
  (부수 뮤턴트: `$CODEX_HOME/config.toml` 을 **같은 크기로 in-place 덮어쓰고** `os.Chtimes` 로 mtime 을 복원한다 — 목록·mtime 스냅샷은 동일하다.)
- **위반되는 REQ 조항**: `spec.md` §F 비기능 — *"초기화는 프로젝트 트리 밖(사용자 홈, `CODEX_HOME`)에 쓰지 않는다"* 및 AC 자신의 Then 문 *"초기화는 프로젝트 트리 안에만 쓴다"*
- **왜 AC 가 못 잡는가**: AC 의 Then 은 "트리 안에만 쓴다" 라는 **전칭 주장** 인데, Given 이 관측하는 것은 **두 디렉터리** 뿐이다. 주장 범위와 측정 범위가 어긋난 전형적 형태다. 게다가 스냅샷 축이 *파일 목록 + mtime* 이라 내용 변조를 원리적으로 보지 못한다.
- **조이는 방법**: 프로젝트·HOME·`CODEX_HOME` 을 **하나의 샌드박스 루트 아래** 배치하고, 그 루트 전체를 **경로 + 내용 해시** 로 스냅샷한 뒤 "프로젝트 서브트리 밖 항목의 집합이 동일" 을 단언한다. 그리고 이 스냅샷을 계약 fixture 뿐 아니라 **거절·비대화형·실패 경로에도** 적용한다.

## AC-CI-009
- **상태**: MUTANT-WRITABLE
- **mutant**: 이 SPEC 이 추가하는 시험 함수 이름에서 `Codex` 를 뺀다 — `TestInitGateStateVerbMatrix`, `TestContractIdempotent`, `TestPathGuardMatrix` … 게이트 명령 `go test ./internal/cli/... -run 'Codex' -timeout 600s` 는 **한 케이스도 실행하지 않고** rc 0 을 낸다. 나머지 게이트(build·vet·lint·중립성)는 시험 내용과 무관하게 통과한다.
  (부수 뮤턴트: 경로 봉쇄를 `//go:build !windows` 파일에만 구현하고 windows 쪽에는 무조건 `return nil` 인 짝을 둔다. `GOOS=windows go vet` 은 **컴파일만** 증명하므로 초록이고, AC-CI-011 의 12칸은 darwin/linux 에서만 도므로 통과한다.)
- **위반되는 REQ 조항**: AC-CI-009 는 "전 REQ" 를 인용한다. 무력화되는 대표 조항은 REQ-CI-011 — *"Initialization shall treat an instruction path that is not a regular file inside the project as out of bounds"* — 로, windows 에서는 그 봉쇄가 실행되지 않는다.
- **왜 AC 가 못 잡는가**: `-run 'Codex'` 는 **이름 정규식** 이라 "시험이 존재하고 실행됐다" 를 증명하지 않는다. rc 0 은 "선택된 시험이 0개" 와 "선택된 시험이 전부 통과" 를 구별하지 않는다 (부재로 인한 0 을 성공으로 읽는 형태). `GOOS=windows go vet` 은 크로스 플랫폼 **동작** 이 아니라 컴파일 가능성만 관측한다.
- **조이는 방법**: (a) `-run 'Codex'` 실행에 `-v` 를 붙여 **실행된 케이스 수 ≥ N**(이 SPEC 이 정의하는 12+8+10+10+6+12 칸의 하한)을 단언하거나, 이름 정규식을 버리고 신규 시험 파일이 속한 패키지를 통째로 돌린다. (b) 경로 봉쇄 행렬(AC-CI-011)을 **windows 러너에서도** 돌리도록 게이트에 한 줄 추가한다.

## AC-CI-010
- **상태**: MUTANT-WRITABLE
- **mutant**: E2(계약 쓰기 실패)에서 계약 함수가 `os.Create(CLAUDE.md)` 로 **먼저 잘라내고** 쓰기 단계에서 실패한다. 결과: exec 0회, 종료 코드 비성공, "계약 쓰기 실패 + 조치" 문구, 그리고 E1·E3 에서의 계약 호출 0회 — 여섯 칸 단언이 **전부** 성립한다. 그런데 사용자가 작성한 `CLAUDE.md` 는 0바이트로 남는다.
  (부수 뮤턴트: M-SPAWN — 실패 경로가 `--spawn` 일 때 `spawnLaunch` 로 그대로 기동한다.)
- **위반되는 REQ 조항**: REQ-CI-006 — *"Where either file already exists, its content shall be preserved byte-for-byte and only the missing link added; initialization shall never rewrite instruction content a person authored."*
- **왜 AC 가 못 잡는가**: AC-CI-010 의 단언은 전부 **행위 계수(exec·계약 호출)와 종료 코드·문구** 이고, 실패 후 **디스크에 남은 상태** 에 대한 단언은 E3 의 *"남은 상태를 그대로 보고"* 한 줄뿐인데 그마저 배선 파일(`config.toml`)에 관한 것이지 지시 파일이 아니다. 내용 보존은 AC-CI-005 가 판정하지만 그것은 **성공 경로 전용** fixture 다 — 실패 경로에서의 보존을 세는 AC 가 없다.
- **조이는 방법**: E1~E3 여섯 칸 각각에 *"실패 후 `AGENTS.md` · `CLAUDE.md` · `CLAUDE.local.md` 가 실행 전과 바이트 동일"* 을 추가한다 (원자적 쓰기 — 임시 파일 + rename — 를 강제하는 효과). 더불어 `exec 0회` 를 `exec 0회 AND spawnLaunch 0회` 로 바꾸고 여섯 칸을 `--spawn` 유무로 곱한다.

## AC-CI-011
- **상태**: MUTANT-WRITABLE
- **mutant**: 봉쇄를 **거부 목록** 으로 구현한다 — `os.Lstat` 으로 얻은 모드에 대해 `m.IsDir() || m&os.ModeSymlink != 0 || m&os.ModeNamedPipe != 0` 일 때만 거부하고 그 외는 통과시킨다. 12칸(외부 심링크·내부 심링크·디렉터리·FIFO × 3파일)이 전부 거부되어 통과한다. 그러나 **유닉스 소켓**(`os.ModeSocket`), **디바이스 노드**(`os.ModeDevice` / `os.ModeCharDevice`), windows 의 **reparse point** 는 통과해 그대로 읽고 쓴다.
  **두 번째 뮤턴트(더 심각)**: `Lstat` 을 **마지막 경로 요소에만** 건다. 저장소가 `docs/` 를 프로젝트 밖으로 향하는 심링크로 만들고 지시 파일을 `docs/AGENTS.md` 로 참조하게 하면, 마지막 요소는 일반 파일이므로 통과하고 쓰기는 프로젝트 밖에 떨어진다. 12 fixture 는 **전부 leaf 이름만** 조작하므로 이 뮤턴트를 건드리지 않는다.
- **위반되는 REQ 조항**: REQ-CI-011 — *"Initialization shall treat an instruction path that is not a regular file inside the project as out of bounds: it shall neither read, write, nor import through such a path, and shall report the condition instead."* (그리고 §F — *"경로 봉쇄는 스냅샷이 아니라 **구조** 로 건다"*)
- **왜 AC 가 못 잡는가**: AC 자신이 *"REQ-CI-011 은 '일반 파일이 아닌 경로' 전부를 거부한다"* 고 적어 놓고, 행렬은 **네 종류만** 열거한다 — 소켓·디바이스가 빠진 열거형 부분 표본이다. 그리고 12칸의 배치가 모두 *"그 이름의 …"*, 즉 **leaf 조작** 이라 부모 디렉터리 성분을 통한 탈출(계산된/간접 경로)은 시험 공간에 존재하지 않는다. 근거 문단이 *"해석 경로 봉쇄"* 를 말하지만 그것을 관측하는 칸이 없다.
- **조이는 방법**: (a) 경로 종류에 **유닉스 소켓** 과 **디바이스 노드**(가능한 플랫폼) 행을 추가해 15~18칸으로 넓히거나, 종류 열거 대신 *"`Lstat` 모드가 `IsRegular()` 가 아닌 모든 fixture"* 를 생성하는 테이블로 바꾼다. (b) **부모 성분이 프로젝트 밖 심링크인** fixture(`docs -> /tmp/outside`, 지시 파일은 `docs/AGENTS.md`)를 3파일 각각에 추가하고, 그 칸에서도 `SENTINEL-OUTSIDE-3k3` 무변경·미도달을 단언한다.

---

## 요약

| AC id | 상태 | 한 줄 요약 |
|---|---|---|
| AC-CI-001 | MUTANT-WRITABLE | 12칸이 디스크와 참 상태가 일치하는 배치뿐이라 분류기 **자체 재구현** 이 그대로 통과 |
| AC-CI-002 | MUTANT-WRITABLE | 주입 상태 × 디스크 2축만 곱하고 **동사 축이 없어** `app` 만 디스크 재분류하는 구현이 통과 |
| AC-CI-003 | MUTANT-WRITABLE | 스냅샷 범위가 프로젝트 트리뿐이라 거절 시 **트리 밖 쓰기** 를 관측하지 못함 (mtime 축도 내용 변조에 무력) |
| AC-CI-004 | MUTANT-WRITABLE | 수락 경로에서 **계약 단계 실행 여부** 를 세는 단언이 SPEC 전체에 없음 — 계약 코드가 죽어 있어도 통과 |
| AC-CI-005 | MUTANT-WRITABLE | "줄 1건" 은 줄 수만 세므로 **코드펜스 안의 비활성 import** 와 0바이트 `AGENTS.md` 가 통과 |
| AC-CI-006 | MUTANT-WRITABLE | I4·I5 가 줄 **위치·개행 코드를 미규정** 이라 접미사 매칭(끝줄 전용) 구현이 멱등처럼 보임 |
| AC-CI-007 | MUTANT-WRITABLE | 참조 계수 범위가 **두 진입 파일뿐** 이라 전이 폐포의 두 번째 참조(이중 로드)를 못 봄 |
| AC-CI-008 | MUTANT-WRITABLE | "트리 안에만 쓴다" 는 전칭 주장 vs **두 디렉터리** 관측 — 부모 디렉터리 쓰기가 통과 |
| AC-CI-009 | MUTANT-WRITABLE | `-run 'Codex'` 는 이름 정규식이라 **0 케이스 실행 rc 0**; `GOOS=windows go vet` 은 컴파일만 증명 |
| AC-CI-010 | MUTANT-WRITABLE | 실패 경로에 **지시 파일 보존** 단언이 없어 truncate-then-fail 이 사용자 파일을 0바이트로 만들고 통과 |
| AC-CI-011 | MUTANT-WRITABLE | 비정규 파일 종류를 **네 개만 열거**(소켓·디바이스 누락) + 12칸이 전부 leaf 조작이라 **부모 성분 심링크 탈출** 미포착 |
| (전 AC 공통) | M-SPAWN | 음성 판정이 `exec 0회` 뿐인데 기동 seam 은 `exec` 과 `spawnLaunch` **둘** — `--spawn` 경로가 게이트 밖 |
