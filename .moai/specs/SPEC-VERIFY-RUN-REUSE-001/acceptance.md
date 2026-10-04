# SPEC-VERIFY-RUN-REUSE-001 — 인수 기준 (v0.2.0)

검증 계층이므로 Given-When-Then 형식이다. 요구사항 본문(GEARS)은 spec.md §B 가 원본이다. (Tier S 는 AC 를 spec.md §3 에 인라인으로 두는 것이 기본이나, 리더 요청에 따라 시나리오 본문을 이 파일에 두고 spec.md §3 은 색인만 둔다.)

## 공통 설정

- 저장소: `internal/cli/verify_test.go` 의 `initVerifyTestRepo(t)` (임시 git 저장소, `.moai/` gitignore) 재사용.
- **포터블 헬퍼 (Windows CI 포함)**: 셸·`echo`·`sh -c` 를 쓰지 않는다. 명령은 테스트 바이너리 자신을 다시 실행하는 헬퍼 프로세스다 — `-- <os.Args[0]> -test.run=^TestVerifyRunHelperProcess$ -- <mode> [arg]`, 모드는 환경변수 `MOAI_VERIFY_RUN_HELPER` 가 켜졌을 때만 동작한다(꺼져 있으면 `t.Skip`; 일반 테스트 실행에서는 SKIP 으로 조용히 끝난다). 모드: `count <file>`(파일에 한 줄 덧붙임 = 실행 횟수), `exit <N>`, `sleep <d>`, `spawn-sleep <d>`(손자 프로세스), `mutate <file>`(추적 파일 수정), `print <text>`. 카운터 파일은 저장소 밖 `t.TempDir()`.
- Unix 전용 단언(프로세스 그룹 종료)은 `runtime.GOOS == "windows"` 일 때 사유를 밝힌 `t.Skip`("process-group termination is Unix only; residual risk spec §D-5").
- **빈 선택 방지 (verification-completeness.md §1.1)**: 모든 go test AC 는 `-v` 로 실행하고 아래 세 가지를 함께 요구한다 — (i) 열거된 이름 각각에 대해 정확히 한 줄 `--- PASS: <Name>` (이름 수 N 은 각 AC 에 적는다), (ii) 출력에 `[no tests to run]` 이 없음, (iii) exit 0. 이름이 바뀌거나 삭제되면 해당 `--- PASS` 줄이 빠져 N 이 맞지 않으므로 적색이다 (선택자 자체는 계속 exit 0 이므로 이 개수 단언이 검증이다). 13개 CLI 이름 전체의 존재는 `go test ./internal/cli -list '^(...)$'` 가 이름 13줄을 출력하는지로도 확인한다.
- **RED-now 칸** (각 AC 에 4요소: 명령 / 원문 stdout / exit code / 트리 SHA 핀 `2b9e4a4d0`): 이 파일을 쓸 때 현재 트리(`git rev-parse --short HEAD` = `2b9e4a4d0`)에서 실행해 관측한 값이다. 모든 go test AC 의 관측은 같은 모양이며 적색 이유는 "실행된 테스트 수 0 ≠ N" 이다:

  ```
  testing: warning: no tests to run
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/<pkg>	<t>s [no tests to run]
  ```
  exit code 는 0 으로 관측되었다 (명령 자체는 초록이므로 판정은 `--- PASS` 줄 수 0 ≠ N 과 `[no tests to run]` 존재가 적색이다). exit code 는 `go test` 호출 결과이며 이 환경의 도구가 직접 출력하지 않아 `ok`/`PASS` 출력 모양에서 읽은 값이다 — Gap 으로 남긴다.
- **green 경로 칸**: 해당 AC 의 구현 마일스톤(plan.md M2/M3/M4)이 끝나면 `--- PASS` 가 N 줄이 된다.

## AC-VRR-001 — 적중과 정량 단언 (maps REQ-VRR-001, REQ-VRR-002)

- **Given** 깨끗한 임시 저장소와 `count` 헬퍼 명령, **When** `verify run --env FOO -- <헬퍼 count>` 를 5회 연속 실행, **Then** 카운터 파일은 정확히 1줄이고(첫 실행만 실행, 이후 4회 실행 0회), 2~5회차는 exit 0, stderr 에 `reuse` 통지(키·`recorded_at`·`duration_ms` 포함), stdout 은 첫 실행의 명령 출력만 담는다.
- 검증 (N=1): `go test ./internal/cli -run '^TestVerifyRunHitExecutesZeroTimes$' -count=1 -v` — 요구: `--- PASS: TestVerifyRunHitExecutesZeroTimes ` (이름 뒤 공백 포함) 1줄. RED-now: 위 공통 관측(stdout `ok ... [no tests to run]`, exit 0, SHA 2b9e4a4d0).
- 단위 (N=1): `go test ./internal/verify -run '^TestDecideReuseHit$' -count=1 -v` — 요구: `--- PASS: TestDecideReuseHit ` (이름 뒤 공백 포함) 1줄. RED-now: 동일 관측.

## AC-VRR-002 — 트리 변경·명령 바이트 차이 시 미스 (maps REQ-VRR-002, REQ-VRR-005)

- **Given** AC-VRR-001 이 적중하는 상태, **When** (a) 추적 파일 수정, (b) 새 untracked 파일, (c) 새 커밋, (d) 명령 인수에 공백 2개·후행 공백·인수 순서 변경·한 요소 `"a b"` vs 두 요소 `a`,`b` 각각 후 실행, **Then** 모두 미스(카운터 증가). `CanonicalCommand` 는 한 요소 `"a b"` 를 따옴표로 렌더해 두 요소 `a b` 와 구별한다.
- 단위 (N=3): `go test ./internal/verify -run '^(TestCanonicalCommand|TestDecideReuseKeyMismatch|TestDecideReuseCommandBytes)$' -count=1 -v` — 요구: 세 이름 각 `--- PASS` 1줄. RED-now: 공통 관측.
- CLI (N=2): `go test ./internal/cli -run '^(TestVerifyRunMissOnTreeChange|TestVerifyRunMissOnCommandBytes)$' -count=1 -v` — 요구: 두 이름 각 `--- PASS` 1줄. RED-now: 공통 관측.

## AC-VRR-003 — 이전 비정상 종료·TTL 초과 시 미스 (maps REQ-VRR-002)

- **Given** (a) 같은 트리에서 종료 코드 1 로 실행·기록된 명령, (b) 기록 후 `--ttl` 보다 오래된 `recorded_at`, **When** 같은 명령 실행, **Then** 각각 명령이 다시 실행된다. (a)는 새 결과로 기록이 교체되고, 이어서 통과하면 그 다음 실행은 적중한다. 단위 테스트는 `now` 를 주입하고 CLI 테스트는 `--ttl 1ns` 로 즉시 미스를 보인다(기본 TTL 10분 안에서는 적중). TTL 은 `recorded_at`(명령 완료 시각)부터 잰다.
- 단위 (N=2): `go test ./internal/verify -run '^(TestDecideReuseNonzeroExit|TestDecideReuseTTL)$' -count=1 -v` — 요구: 두 이름 각 `--- PASS` 1줄. RED-now: 공통 관측.
- CLI (N=2): `go test ./internal/cli -run '^(TestVerifyRunNeverReusesFailure|TestVerifyRunMissOnTTL)$' -count=1 -v` — 요구: 두 이름 각 `--- PASS` 1줄. RED-now: 공통 관측.

## AC-VRR-004 — env 다이제스트·도구 식별 변경 시 미스 (maps REQ-VRR-002, REQ-VRR-006)

- **Given** `--env GOFLAGS`(값 `-a`) 와 `--tool-version-cmd <헬퍼> --tool-version-cmd -test.run=^TestVerifyRunHelperProcess$ --tool-version-cmd -- --tool-version-cmd print --tool-version-cmd v1`(반복 플래그 = argv 원소 순서) 로 기록, **When** (a) GOFLAGS 값을 `-b` 로, (b) unset 으로, (c) 빈 문자열로 바꾸거나 (d) 도구 식별 출력을 `v2` 로 바꾸고 같은 명령 실행, **Then** 모두 미스. unset 과 빈 문자열은 서로 다른 다이제스트다. 같은 값이면 적중하고, `--env` 에 나열하지 않은 변수의 변화는 적중을 깨지 않는다(§D 의 문서화된 동작). 도구 식별 명령이 시작 불가·비정상 종료·`--tool-version-timeout`(테스트는 200ms) 초과·빈 출력이면 unbound 로 재사용·기록 없이 실행하고 stderr 에 원인을 남기며 종료 코드는 명령의 것이다. 플래그가 없으면 두 실행은 `unversioned` 로 같다. 도구 식별은 stdout 만 쓴다(헬퍼가 stderr 에만 출력하면 빈 출력으로 취급).
- 단위 (N=1): `go test ./internal/verify -run '^TestEnvDigest$' -count=1 -v` — 요구: `--- PASS: TestEnvDigest ` (이름 뒤 공백 포함) 1줄. RED-now: 공통 관측.
- CLI (N=3): `go test ./internal/cli -run '^(TestVerifyRunMissOnEnvChange|TestVerifyRunToolVersionBinding|TestVerifyRunToolVersionTimeout)$' -count=1 -v` — 요구: 세 이름 각 `--- PASS` 1줄. RED-now: 공통 관측.

## AC-VRR-005 — 종료 코드 패스스루·타임아웃 (maps REQ-VRR-003)

- **Given** 종료 코드 `0`, `3`, `1` 을 내는 `exit` 헬퍼, **When** 미스 경로로 실행, **Then** `verify run` 의 종료 코드가 각각 같다(`exitCodeError.ExitCode()`). 시작 불가 명령은 127, `--timeout` 초과는 124 이며 둘 다 기록이 남지 않는다. Unix 에서는 `spawn-sleep` 으로 만든 손자 프로세스도 타임아웃 뒤 남지 않는다 (Windows 는 위 사유로 skip). `TestVerifyRunTimeoutAndNotFound` 는 서브테스트 둘로 나뉜다: `ExitCodes124And127`(포터블 124/127 단언)와 `GrandchildKilled`(Unix 전용, Windows 에서는 사유를 밝힌 skip).
- CLI (N=4 PASS 줄): `go test ./internal/cli -run '^(TestVerifyRunExitCodePassthrough|TestVerifyRunTimeoutAndNotFound)$' -count=1 -v` — 요구: `--- PASS: TestVerifyRunExitCodePassthrough `, `--- PASS: TestVerifyRunTimeoutAndNotFound `, `--- PASS: TestVerifyRunTimeoutAndNotFound/ExitCodes124And127 `, `--- PASS: TestVerifyRunTimeoutAndNotFound/GrandchildKilled ` 각 1줄 (서브테스트 줄은 들여쓰기됨; 이름 뒤 공백 포함). Windows 에서는 마지막 줄이 `--- SKIP: TestVerifyRunTimeoutAndNotFound/GrandchildKilled` + 사유 로 대체되는 경우만 허용. RED-now: 공통 관측.

## AC-VRR-006 — 기록 규칙과 fail-open (maps REQ-VRR-004, REQ-VRR-007)

- **Given** (a) `mutate` 헬퍼가 실행 중 추적 파일을 수정, (b) 스토어 디렉터리 쓰기 불가(Windows 에서 권한 모델이 달라 불가능하면 사유를 밝힌 skip), (c) `moai verify record` 로만 기록된 항목, **When** `verify run`, **Then** (a) 결과는 기록되지 않고 stderr 에 트리 이동 통지, (b) 명령은 그대로 실행되고 종료 코드는 명령의 것이며 stderr 에 저장 오류, (c) 그 항목은 `config_digest`/`tool_version` 이 없어 적중하지 않는다. 어느 경우에도 새 스키마 필드·새 저장 경로가 생기지 않는다.
- CLI (N=3): `go test ./internal/cli -run '^(TestVerifyRunTreeMovedNotRecorded|TestVerifyRunStoreFailOpen|TestVerifyRunIgnoresHandRecordedEntry)$' -count=1 -v` — 요구: 세 이름 각 `--- PASS` 1줄 (Windows 의 `StoreFailOpen` 이 skip 이면 해당 줄은 `--- SKIP: ... <사유>` 로 대체되며 이 경우만 허용). RED-now: 공통 관측.

## AC-VRR-007 — 교리 문장 (maps REQ-VRR-008)

- **Given** run 완료 후 트리, **When** 두 파일(`AGENTS.md`, `internal/template/templates/AGENTS.md.tmpl`) 각각에서 아래 다섯 구절을 고정 문자열로 찾는다, **Then** 다섯 구절 모두 각 파일에서 exit 0 이고 정확히 1줄에 매치되며, 같은 문장 안에 있다(구절 간 줄 간격 ≤ 6줄로 확인): `moai verify run`, `--env`, `recorded_at`, `output not re-observed`, `runs the command directly`. 문장은 TTL 경계(`TTL`)와 재사용=이전 관측(`prior observation`)도 담는다.
- 검증 (파일마다 구절마다 한 호출, 각 exit 0 + 매치 정확히 1줄): `grep -n -F -e "moai verify run" AGENTS.md`, `grep -n -F -e "output not re-observed" AGENTS.md`, `grep -n -F -e "recorded_at" AGENTS.md`, `grep -n -F -e "runs the command directly" AGENTS.md`, `grep -n -F -e "--env" AGENTS.md`; 같은 다섯 호출을 `internal/template/templates/AGENTS.md.tmpl` 에도 실행. 두 파일의 줄 번호 차이는 허용하고 문구는 동일해야 한다. 바이트 한도: `go test ./internal/config -run '^TestCodexContractByteCeiling$' -count=1 -v` — 요구: `--- PASS: TestCodexContractByteCeiling ` (이름 뒤 공백 포함) 1줄 (관측 여유: AGENTS.md 2996 B / 템플릿 3284 B, plan-audit iter1).
- 패턴이 `-` 로 시작하면 grep 이 옵션으로 읽으므로(`grep -F "--env"` → `invalid option`, exit 2) 모든 패턴은 `-e` 로 넘긴다.
- RED-now (핀 `2b9e4a4d0`, 현재 트리에서 관측): `grep -n -F -e "--env" AGENTS.md; echo "exit=$?"` → stdout 비어 있음, `exit=1` (매치 없음 = 정당한 적색). 양성 대조: 같은 형태를 `--env` 가 들어 있는 `.moai/specs/SPEC-VERIFY-RUN-REUSE-001/spec.md` 에 실행하면 `exit=0`.

## AC-VRR-008 — 파일 범위 (spec.md §C 제약; REQ 에 매핑되지 않는 제약 AC)

- **Given** run 브랜치, **When** `git diff --name-only 2b9e4a4d0 HEAD` 실행(카드 기점 SHA 고정), **Then** 출력 경로는 plan.md §B 목록과 `.moai/specs/SPEC-VERIFY-RUN-REUSE-001/`, `.moai/reports/t1452/` 안에만 있다. 금지 목록은 spec.md §C 와 같다: 경로가 `internal/kanban/`, `internal/homestate/`, `internal/factorylane/` 로 시작하거나, `internal/cli/integration`·`internal/cli/factory_complete`·`internal/cli/factory_merge`·`internal/cli/factory_card` 로 시작하거나, `AGENTS.local.md` 이거나, 기본 파일명이 `gitflow-lane-protocol.md`·`kanban-dispatch-mechanics.md` 인 항목이 하나도 없다.
- 검증: `git diff --name-only 2b9e4a4d0 HEAD`. 대조군: 출력이 1줄 이상이어야 하며(0줄이면 측정 불가) 기점 고정 측정이다 — develop 흡수 후에는 merge-base 로 다시 구한다 (gitflow-lane-protocol §8 범위 판정식 규율).

## 엣지 케이스

- 키 계산이 타임아웃(30초)이면 REQ-VRR-007 로 직접 실행.
- 두 레인이 같은 키에 서로 다른 명령을 동시 기록해도 `RecordCheck` 키별 잠금으로 서로 덮지 않는다(기존 보장 — 신규 테스트 불필요).
- 인수 없이 `--` 만 주면 사용법 오류로 exit 2 (기록·실행 없음).

## 완료 정의 (Definition of Done)

- AC-VRR-001~008 의 위 검증 명령이 모두 통과했고(go test 는 N 줄 `--- PASS`, `[no tests to run]` 없음) 출력이 판정서에 인용되어 있다.
- `GOOS=windows GOARCH=amd64 go build ./...` 가 통과한다 (프로세스 그룹 코드는 빌드 태그로 분리).
- `go vet ./internal/verify ./internal/cli` 와 CI 판 golangci-lint 가 0 건이다.
- `go run ./cmd/moai spec lint .moai/specs/SPEC-VERIFY-RUN-REUSE-001/spec.md` 오류 0 건이다.
- 로컬 `AGENTS.md` 와 템플릿 미러의 §4 문장이 같다.
- spec.md §D 잔여 위험이 §4 문장에 경계로 반영되어 있다.
