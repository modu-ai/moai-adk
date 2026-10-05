# acceptance.md — SPEC-CODEX-LOCALMD-001

모든 카드 수용 항목을 AC-LMD-XXX Given/When/Then으로 사상한다. 코드 레벨 AC는 `codex_local_instructions_test.go` characterization suite로, LIVE AC(AC-LMD-012)는 실제 별도 codex 세션으로 검증한다.

## D. AC Matrix

| AC | 요구 | 시나리오 | 검증 수단 |
|---|---|---|---|
| AC-LMD-001 | REQ-LMD-001/008 | 3 verb 동일 페이로드 | capture 테스트 |
| AC-LMD-002 | REQ-LMD-008 | spawn/-w/factory 경로 동일성 | capture 테스트 |
| AC-LMD-003 | REQ-LMD-003 | precedence 순서 fixture | JSON round-trip 단언 |
| AC-LMD-004 | REQ-LMD-004 | absent/empty 매트릭스 | 테이블 테스트 |
| AC-LMD-005 | REQ-LMD-002 | 양쪽 파일 안전 open/fstat | 거부·경합 테스트 |
| AC-LMD-006 | REQ-LMD-006 | 61KiB 바이트 보존 | round-trip 단언 |
| AC-LMD-007 | REQ-LMD-005 | `-c`/`--config` 충돌 거부 | parser probe + 오류 단언 |
| AC-LMD-008 | REQ-LMD-007 | direct·spawn 최종 표현 초과 차단 | 길이·오류 + launches==0 |
| AC-LMD-009 | REQ-LMD-010 | 입력 비가공 | byte-identity + rename==0 |
| AC-LMD-010 | REQ-LMD-006 | 재시작 재판독 | 2회 런치 테스트 |
| AC-LMD-011 | REQ-LMD-009 | 헬프·핀 정합 | 핀 테스트 갱신 |
| AC-LMD-012 | REQ-LMD-001/003 | LIVE 실세션 검증 | 실측 증거 |

## D.1 — AC-LMD-001: 세 verb에서 동일한 합성 페이로드

- **Given** 프로젝트 루트에 비어있지 않은 `CLAUDE.local.md`와 `AGENTS.local.md`가 있고
- **When** `moai codex`(bare), `moai codex cli`, `moai codex app`을 각각 캡처 하네스로 런치하면
- **Then** 세 런치의 `-c developer_instructions=` 토큰이 바이트 동일하고, 정확히 한 쌍의 토큰만 존재하며 (기존 `TestCodexLocalInstructions_DirectSpawnAndAppSharePrefix` 확장), 캡처된 argv 어디에도 두 번째 `developer_instructions` 토큰이 없다.

## D.2 — AC-LMD-002: spawn / -w / factory 경로 동일성

- **Given** 동일한 두 로컬 파일 fixture가 있고
- **When** `--spawn` 경로(`buildCodexSpawnCommand` shell-quote 전의 args 비교), `-w <worktree>` 경로, `-f`(lead), `-f agent`, `-f agent-2`, `-f lane-3` 경로로 각각 런치하면
- **Then** 모든 경로의 `developer_instructions` 페이로드가 D.1의 bare-form 페이로드와 바이트 동일하다 (funnel(`codex_launcher.go:495`) 단일 통과의 구조적 균일성 재확인).

## D.3 — AC-LMD-003: 출처 순서·구분자 (precedence fixture)

- **Given** `CLAUDE.local.md`는 식별 가능한 마커 `CLAUDE_MARKER`를, `AGENTS.local.md`는 `AGENTS_MARKER`를 담고 있고
- **When** 런치를 캡처해 `json.Unmarshal`로 페이로드를 복원하면
- **Then** 복원 문자열에서 `CLAUDE_MARKER`의 오프셋이 `AGENTS_MARKER`보다 앞서고, REQ-LMD-003이 핀한 구분자 형태(`<!-- source: CLAUDE.local.md -->`, `<!-- source: AGENTS.local.md -->` — 각 파일명을 명명한 헤더 라인)가 각 바디 직전에 존재하며, 연결 순서는 CLAUDE first / AGENTS later로 고정돼 있다.

## D.4 — AC-LMD-004: 부재/빈 파일 매트릭스 (8셀)

- **Given / When / Then** 테이블 — REQ-LMD-004의 빈-파일-무시(empty-as-absent) 규칙에 따라 각 셀의 기대 argv:

| 셀 | CLAUDE.local.md | AGENTS.local.md | 기대 argv |
|---|---|---|---|
| 1 | 부재 | 부재 | `-c` 오버라이드 없음 (argv 불변 — 기존 `_AbsentLeavesArgvUnchanged` 확장) |
| 2 | 존재·비어있지 않음 | 부재 | CLAUDE 내용만 주입 |
| 3 | 부재 | 존재·비어있지 않음 | AGENTS 내용만 주입 |
| 4 | 빈 파일(0 byte) | 존재·비어있지 않음 | AGENTS 내용 주입 (빈 CLAUDE는 무시) |
| 5 | 존재·비어있지 않음 | 빈 파일(0 byte) | CLAUDE 내용 주입 (빈 AGENTS는 무시) |
| 6 | 빈 파일 | 빈 파일 | `-c` 오버라이드 없음 (argv 불변) |
| 7 | 부재 | 빈 파일 | `-c` 오버라이드 없음 (argv 불변) |
| 8 | 빈 파일 | 부재 | `-c` 오버라이드 없음 (argv 불변) |

- 셀 1~6은 캡처된 argv가 정확히 한 개의 `developer_instructions` 오버라이드(또는 0개)를 담는다는 점도 함께 단언한다.

## D.5 — AC-LMD-005: 양쪽 파일의 안전 open/fstat

- **Given** 각 케이스: `CLAUDE.local.md`가 symlink / FIFO / directory / socket / character-device인 경우, `AGENTS.local.md`가 symlink / FIFO / directory / socket / character-device인 경우, 읽기 오류(mode 000)를 내는 경우
- **When** 런치하면
- **Then** 각 비정규 입력은 symlink-follow 없는 안전 open 단계에서 거부되거나, open에 성공한 descriptor의 fstat에서 거부되고, `codexPathGuardError` 어휘의 명명된 오류와 `launches == 0`이 관측된다. 각 파일에 대해 검사와 read 사이에 경로를 다른 파일 또는 symlink로 교체하는 경합 fixture에서는 검증한 descriptor의 원래 바이트만 읽거나 launch 전 fail closed하며, 교체된 경로의 바이트는 절대 주입되지 않는다.
- **Platform note**: FIFO·Unix-domain socket·character-device·mode-000 셀은 해당 개념이 있는 Unix 플랫폼에서 실행한다. directory 셀은 모든 지원 플랫폼에서 실행한다. Windows symlink 셀은 symlink 생성 권한이 있는 Windows 환경에서 실행하며, 권한이 없어 fixture를 만들 수 없으면 PASS로 세지 않고 명시적 GAP 또는 사전 승인된 platform gate로 기록한다.

## D.6 — AC-LMD-006: 61,360-byte UTF-8 무절단 round-trip

- **Given** 61,360 bytes(현재 리포 루트 `CLAUDE.local.md` 실측 크기)의 UTF-8 바디(한글·이모지·`<`/`>`/`&`·quotes·개행 포함)를 fixture로 쓰고
- **When** 런치를 캡처해 `json.Unmarshal`로 복원하면
- **Then** JSON 복원 후 provenance 헤더 경계로 각 body slice를 추출했을 때 각 slice의 bytes와 sha256가 해당 원본 fixture와 정확히 일치한다. provenance와 framing bytes는 전체 payload에는 존재하지만 어느 body hash에도 포함하지 않으며 truncation이 없다.

## D.7 — AC-LMD-007: 연산자 config override 충돌 명시적 거부

- **Given** 두 로컬 파일이 존재하고, Codex 0.155.1 parser probe가 수용한 다섯 표기 `--config developer_instructions=operator`, `--config=developer_instructions=operator`, `-c developer_instructions=operator`, `-c=developer_instructions=operator`, `-cdeveloper_instructions=operator`가 각각 fixture 셀로 있다
- **When** 각 표기를 연산자 `--` tail로 전달하면
- **Then** 모든 수용 표기가 같은 명명형 duplicate-override 오류로 child 시작 전에 실패한다. 음성 대조도 통과한다: 다른 config key와 override 없는 tail은 정상이고, 두 로컬 파일이 모두 부재 또는 빈 파일이면 operator의 `developer_instructions` override가 그대로 전달된다.

## D.8 — AC-LMD-008: direct·spawn 최종 표현별 상한 초과 fail-closed

- **Given** direct와 spawn에 각각 독립 적용되는 기본 상한 126,976 bytes가 있고, (a) direct 최종 토큰의 상한 바로 이하/초과 경계쌍, (b) 작은따옴표 40,000개의 헤더 없는 측정 문자열에서 40,025→160,027 bytes 팽창이 관측된 반례 fixture, (c) spawn 최종 명령의 상한 바로 이하/초과 경계쌍이 있으며, 테스트는 provenance 헤더와 명령 prefix까지 포함한 실제 생성 함수 반환값의 byte length를 측정한다
- **When** direct와 `--spawn`을 각각 런치하면
- **Then** direct는 최종 exec 토큰을, spawn은 완전히 조립·인용된 최종 명령 문자열을 각자의 상한으로 판정한다. 초과 셀은 측정 길이·상한을 포함한 오류로 child/tmux 생성 전에 실패하고 `launches == 0`이며, 이하 셀은 크기 가드 때문에 거부되지 않는다. 40,000-quote 반례는 pre-quote 길이가 작다는 이유로 spawn을 허용하지 않는다.

## D.9 — AC-LMD-009: 소스 트리 입력 비가공

- **Given** 두 로컬 파일이 존재하고
- **When** 정상 런치가 완료되면
- **Then** 사용자 로컬 입력 `CLAUDE.local.md`와 `AGENTS.local.md`는 모든 단계에서 바이트 동일하고 rename/link 카운트가 0이다. 런처와 LIVE 실행은 `AGENTS.md`/`CLAUDE.md`도 수정하지 않는다. 이는 sync-phase에서 manager-docs가 추적되는 리포 `AGENTS.md`의 설명 문구를 갱신하는 것과 구분된다.

## D.10 — AC-LMD-010: 재시작 재판독

- **Given** 첫 런치 후 `CLAUDE.local.md` 내용을 다른 바이트로 교체하면
- **When** 다시 런치하면
- **Then** 두 번째 캡처의 페이로드가 교체된 바이트를 반영한다 (캐시 없음, 세션마다 최신 바이트 판독).

## D.11 — AC-LMD-011: 헬프·핀 테스트 정합

- **Given** 갱신된 `codexCmd.Long`이 있다
- **When** `TestCodexLocalInstructions_DocumentedInLauncherHelp`를 실행하면
- **Then** 헬프가 `CLAUDE.local.md`는 Claude workflow와 공유되는 공통 로컬 입력, `AGENTS.local.md`는 Codex-specific 입력이며 `moai codex`가 둘 다 주입한다고 구분한다. 테스트는 부정확한 exclusivity만 거부하고, `AGENTS.local.md`를 정확히 수식하는 "Codex-only" 자체는 실패 조건으로 삼지 않는다.

## D.12 — AC-LMD-012: LIVE 수용 — 독립 fixture의 Codex 레인 [LIVE]

- **Given** Codex 인증을 사용할 수 있는 Codex 레인이 소스 트리 밖 독립 fixture를 만들고, `CLAUDE.local.md`에는 `CLAUDE_NONCE=<unique-a>`, `AGENTS.local.md`에는 `AGENTS_NONCE=<unique-b>`를 기록하며, 소스 트리 네 입력 파일의 사전 해시를 기록했다
- **When** fixture를 working directory로 하여 120초 hard timeout 아래 정확히 `"$BUILT_MOAI" codex -- exec --json --skip-git-repo-check "$PROMPT"`를 실행한다. `$PROMPT`는 도구 호출 없이 두 nonce의 정확한 값과 각 source filename을 JSON으로 답하도록 요구하고, evidence에는 변수의 실제 전개값을 기록한다.
- **Then** PASS는 bounded command가 exit 0이고, 응답 JSON과 Codex session log 양쪽 모두에서 `(CLAUDE_NONCE, CLAUDE.local.md)`와 `(AGENTS_NONCE, AGENTS.local.md)` 두 쌍이 정확히 확인되며, 사후 소스 트리 네 입력 파일 해시가 사전값과 같을 때만 성립한다. timeout, nonce/source 쌍 하나라도 누락, log 부재, 비영 종료는 PASS가 아니다.
- **[HARD]** AC-LMD-012가 `NOT_RUN`이면 전체 판정은 PASS가 아니다. 인증 또는 Codex 레인 부재는 Gap으로 보고한다.

## D.5.x — 간접 검증

- `grep -c "developer_instructions" internal/cli/*.go` — 런처의 argv 생산 지점이 여전히 1곳임을 확인 (토큰 분열 회귀 감지).
- `grep -rn "CLAUDE.local.md" internal/cli/` — 이름 상수 정의 1곳뿐, 하드코딩 0건.
- **plan이 신설하는 exactly-3 오류 경로 테스트**(plan.md M1의 RED 과제 — 2-또는-4-entry 변형 테이블을 `codexInstructionRelPathsFn`에 주입해 `secureCodexInstructionContract`가 `"instruction path table must name exactly three paths"`(`codex_contract.go:225-227`)로 실패하는 것을 단언)가 존재하고 통과한다 — 이 분기는 현재 커버리지 0이다(음성 grep 확인, 이번 실행; 기존 변형 테스트 `codex_contract_test.go:350-370`은 in-place 치환으로 항상 길이 3을 유지한다). 테이블 불변 회귀 감지의 유일한 실행 가능 계기는 이 신설 테스트다.

## 품질 게이트 (Definition of Done)

- E1~E7 (plan.md §E) 전부 관측 출력과 함께 통과.
- AC-LMD-001~011 코드 레벨 전부 통과 + AC-LMD-012 LIVE 증거 착지 (NOT_RUN 시 un-PASS).
- scoped coverprofile 기준 `internal/cli` 패키지 커버리지는 측정된 변경 전 기준선보다 하락하지 않고, 신규·변경 함수 coverage는 85% 이상이며 요구 분기가 실행되고, `golangci-lint run`은 0 error다.
