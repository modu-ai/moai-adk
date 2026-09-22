# acceptance.md — SPEC-CODEX-LOCALMD-001

모든 카드 수용 항목을 AC-LMD-XXX Given/When/Then으로 사상한다. 코드 레벨 AC는 `codex_local_instructions_test.go` characterization suite로, LIVE AC(AC-LMD-012)는 실제 별도 codex 세션으로 검증한다.

## D. AC Matrix

| AC | 요구 | 시나리오 | 검증 수단 |
|---|---|---|---|
| AC-LMD-001 | REQ-LMD-001/008 | 3 verb 동일 페이로드 | capture 테스트 |
| AC-LMD-002 | REQ-LMD-008 | spawn/-w/factory 경로 동일성 | capture 테스트 |
| AC-LMD-003 | REQ-LMD-003 | precedence 순서 fixture | JSON round-trip 단언 |
| AC-LMD-004 | REQ-LMD-004 | absent/empty 매트릭스 | 테이블 테스트 |
| AC-LMD-005 | REQ-LMD-002 | 파일별 거부 | 거부 테스트 |
| AC-LMD-006 | REQ-LMD-006 | 61KiB 바이트 보존 | round-trip 단언 |
| AC-LMD-007 | REQ-LMD-005 | `-c` 충돌 거부 | 오류 단언 |
| AC-LMD-008 | REQ-LMD-007 | argv 초과 fail-closed | 오류 + launches==0 |
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

## D.5 — AC-LMD-005: 파일별 구조 거부

- **Given** 각 케이스: `CLAUDE.local.md`가 symlink / FIFO / directory / socket / character-device인 경우, `AGENTS.local.md`가 symlink / FIFO / directory / socket / character-device인 경우, 읽기 오류(mode 000)를 내는 경우
- **When** 런치하면
- **Then** 각각 `codexPathGuardError` 어휘의 명명된 오류(예: `"not a regular file (symlink)"`, `"named pipe"`, `"directory"`, `"socket"`, `"character device"`)로 실패하고 `launches == 0`이다 (기존 `_SymlinkIsRefused` 패턴을 2파일×모드로 확장; 오류 메시지가 어느 파일이 거부됐는지 식별한다).
- **Unix-gate note**: symlink/FIFO/socket/device/mode-000 fixture 생성은 Unix-only다. FIFO·socket·character-device·mode-000 셀은 `codex_contract_fixture_unix_test.go` / `codex_contract_fixture_windows_test.go`의 기존 빌드태그 분리 패턴을 따라 Unix 게이트로 작성하고, Windows 빌드에서는 대응 셀을 스킵한다(파일 존재는 이번 실행에서 확인 — research.md). symlink·directory 셀은 양 플랫폼에서 실행한다.

## D.6 — AC-LMD-006: 61,360-byte UTF-8 무절단 round-trip

- **Given** 61,360 bytes(현재 리포 루트 `CLAUDE.local.md` 실측 크기)의 UTF-8 바디(한글·이모지·`<`/`>`/`&`·quotes·개행 포함)를 fixture로 쓰고
- **When** 런치를 캡처해 `json.Unmarshal`로 복원하면
- **Then** 복원 바이트가 입력 바이트와 정확히 동일하고 (sha256 일치), 토큰 길이가 인코딩 확장분만큼만 커지며 truncation 흔적이 없다.

## D.7 — AC-LMD-007: 연산자 `-c` 충돌 명시적 거부

- **Given** 두 로컬 파일이 존재하고 연산자 tail이 `-- -c developer_instructions="operator"`를 포함한다
- **When** 런치하면
- **Then** 충돌을 이름 지어 보고하는 오류로 실패하고(launch 시작 전), child argv가 생성되지 않으며, 어느 값도 무음으로 버려지지 않는다. 음성 대조 2셀: (a) tail에 `developer_instructions` 오버라이드가 없으면 정상 런치다; (b) 두 로컬 파일이 모두 부재(또는 빈 파일)인 상태에서 tail이 `-c developer_instructions=...`를 포함하면 충돌 없이 정상 런치다 — 런처가 합성값을 내지 않으므로 REQ-LMD-005의 "A tail token with NO synthesized launcher value is not a collision" 절의 관측 증거다.

## D.8 — AC-LMD-008: argv 상한 초과 fail-closed

- **Given** 합성 오버라이드의 최종 인코딩 길이(JSON escape 완료 후 exec에 전달되는 `developer_instructions=<value>` 토큰의 바이트 길이)가 선언 상한을 초과하도록 큰 fixture를 넣는다 — 상한 앵커 재기술: 신규 상수(default `internal/config/defaults.go` 단일 정의), 기본값 = Linux `MAX_ARG_STRLEN` 131,072 bytes − 안전 마진 4,096 bytes = 126,976 bytes
- **When** 런치하면 (direct와 `--spawn` 양쪽)
- **Then** 두 경로 모두 launch 이전에 진단형 오류(측정 길이·상한값 포함)로 실패하고, tmux window가 열리지 않으며 `launches == 0`이다.

## D.9 — AC-LMD-009: 입력 비가공

- **Given** 두 로컬 파일이 존재하고
- **When** 정상 런치가 완료되면
- **Then** 두 파일의 내용이 런치 전후 바이트 동일하고, rename/link 카운트가 0이며 (기존 `codex_contract_link_test.go:173` 패턴 준용), `AGENTS.md`/`CLAUDE.md`에 어떤 변경도 없다.

## D.10 — AC-LMD-010: 재시작 재판독

- **Given** 첫 런치 후 `CLAUDE.local.md` 내용을 다른 바이트로 교체하면
- **When** 다시 런치하면
- **Then** 두 번째 캡처의 페이로드가 교체된 바이트를 반영한다 (캐시 없음, 세션마다 최신 바이트 판독).

## D.11 — AC-LMD-011: 헬프·핀 테스트 정합

- **Given** 갱신된 `codexCmd.Long`이 있다
- **When** `TestCodexLocalInstructions_DocumentedInLauncherHelp`를 실행하면
- **Then** 헬프가 두 파일 이름과 dual-file 주입 사실을 기술하고, 갱신된 기대 토큰 집합을 통과하며, "Codex-only"처럼 부정확해진 문구가 헬프에 남아 있지 않다.

## D.12 — AC-LMD-012: LIVE 수용 — 실제 별도 codex 세션 검증 [LIVE]

- **Given** 실제 리포 루트의 `CLAUDE.local.md`에 고유 nonce(예: `LMD-LIVE-<date>-<rand>`)를 담은 줄이 추가되어 있고 (검증 후 원복 — 원본 61,360 bytes와 sha256 일치로 증명), 별도 codex 세션 실행 환경(바이너리·auth)이 준비되어 있다
- **When** `moai codex`로 실제 codex 세션을 열어 nonce 관련 요약 응답을 요청하면
- **Then** (1) 세션 응답이 nonce를 확인하고, (2) 주입 소스·페이로드 증거가 남는다 — 명명된 로그 표면: **codex CLI의 세션 로그**(`CODEX_HOME` 아래 codex CLI가 기록하는 세션 기록 파일; 정확한 경로는 실행 환경의 codex 버전이 결정하는 외부 산출물이다). 증거 기록(실행 명령, 세션 응답 발췌, codex 세션 로그 파일의 **절대 경로**와 관련 발췌)은 `.moai/specs/SPEC-CODEX-LOCALMD-001/` 아래 evidence 파일로 남긴다 — codex 세션 로그 자체는 외부 산출물이므로, 그 경로와 발췌를 담은 이 evidence 파일이 본 SPEC의 감사 가능한 증거다.
- **[HARD]** AC-LMD-012가 `NOT_RUN`이면 SPEC 전체 판정은 PASS 불가다 — 코드 레벨 AC 전체 통과와 무관하게 un-PASS로 기록한다. 환경 부재 시 운영자에게 실행 주체·시점을 상신한다(decision-index R4).

## D.5.x — 간접 검증

- `grep -c "developer_instructions" internal/cli/*.go` — 런처의 argv 생산 지점이 여전히 1곳임을 확인 (토큰 분열 회귀 감지).
- `grep -rn "CLAUDE.local.md" internal/cli/` — 이름 상수 정의 1곳뿐, 하드코딩 0건.
- **plan이 신설하는 exactly-3 오류 경로 테스트**(plan.md M1의 RED 과제 — 2-또는-4-entry 변형 테이블을 `codexInstructionRelPathsFn`에 주입해 `secureCodexInstructionContract`가 `"instruction path table must name exactly three paths"`(`codex_contract.go:225-227`)로 실패하는 것을 단언)가 존재하고 통과한다 — 이 분기는 현재 커버리지 0이다(음성 grep 확인, 이번 실행; 기존 변형 테스트 `codex_contract_test.go:350-370`은 in-place 치환으로 항상 길이 3을 유지한다). 테이블 불변 회귀 감지의 유일한 실행 가능 계기는 이 신설 테스트다.

## 품질 게이트 (Definition of Done)

- E1~E7 (plan.md §E) 전부 관측 출력과 함께 통과.
- AC-LMD-001~011 코드 레벨 전부 통과 + AC-LMD-012 LIVE 증거 착지 (NOT_RUN 시 un-PASS).
- `internal/cli` 커버리지 85% 이상, `golangci-lint run` 0 error.
- decision-index R1~R4가 운영자 결정을 마친 상태로 폐쇄.
