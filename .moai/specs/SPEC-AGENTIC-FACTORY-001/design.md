# SPEC-AGENTIC-FACTORY-001 — 설계

이 문서는 `spec.md`의 요구사항이 가리키는 표와 계약을 담는다. 코드 구조나 함수 이름은 정하지 않는다. §7~§9의 명칭·명령·호환 순서는 모두 **미확정(provisional)**이다. 상위 설계 권고(§13)는 도착했고 PROPOSAL로 반영했지만, 리더·사용자 확인 전까지 확정이 아니다.

입력 보고서 세 건(lane-22 `a-merge-gate-gap-matrix.md`·`c-factory-migration-map.md`, lane-21 `21-browser-contracts.md`)은 커밋되지 않는 로컬 파일이다. 살아남아야 할 내용은 이 문서와 `spec.md`에 옮겨 적었고, 보고서 경로는 출처 표시로만 남긴다. 보고서를 읽은 것이지 그 안의 측정을 이 세션이 다시 실행한 것은 아니다(차이는 `spec.md` §H).

## §1 전체 구조

```
GitHub Issue (사람 소유 필드의 권위 + 효과 상태의 거울 댓글 1개)
        ▲  관찰(읽기)                ▲  거울 쓰기(ControlService만)
        │                            │
 observe 명령 ──────────────  ControlService (결정적 제어·복구, LLM 없음)
                                     │  claim / lease / attempt / delivery 기록
                                     ▼
                            로컬 claim 저장소 (효과 상태의 권위)
                                     │
             coordinator(LLM 조율) ──┤ 요청만 (ControlService 명령을 통해)
                                     │
                         worker 세션 ── 자기 attempt 범위에서만 쓰기
                                     │
                         GitHub 원격 (병합·검사·main 반영 — 매번 직접 조회)
```

## §2 상태 모델과 전이표 (REQ-AF-001, REQ-AF-002)

### 2.1 상태 열거값

| 상태 (영문 식별자) | 표시 이름 | 뜻 |
|---|---|---|
| `queued` | 대기 | Issue가 실행 대상이지만 소유자가 없다 |
| `claimed` | 소유권 확보 | claim 기록이 생겼고 owner·writer가 정해졌다 |
| `implementing` | 구현 | worker가 attempt 안에서 작업 중이다 |
| `verifying` | 동일 HEAD 검증 | 후보 HEAD 하나에 대해 검사·리뷰·감사·QA를 모은다 |
| `ready` | 준비완료 인계 | writer 인계와 coordinator 인계가 모두 있다 |
| `landing` | 병합/배포 확인 | 병합을 요청했고 원격 반영을 확인하는 중이다 |
| `done` | 완료 | 원격 `main` 반영, 로컬 동기화 기록, 배포 기록(또는 사유가 적힌 `deploy: N/A`) 세 가지가 모두 있다(REQ-AF-002). Issue 종료는 `done` 기록 **뒤에** 한다 |
| `blocked` | 차단 | 정상 진행이 불가능하다(인증 거부, 429, 예산 소진, 미설치, 저장소 불일치 등) |
| `on-hold` | 보류 | 운영자 또는 정책이 의도적으로 멈췄다 |

`blocked`와 `on-hold`는 진행 단계가 아니다. 들어갈 때 떠난 단계(`return_stage`)와 사유, 정상 재개 조건을 기록하고, 나올 때는 `return_stage`로만 돌아간다.

### 2.1a 차단 사유 열거값 (`block_reason`)

`blocked`는 어떤 사유로도 성공이 아니다. 사유는 아래 집합 중 하나만 쓰며, 집합 밖 값은 거절한다. "근거"는 사유를 기록할 때 함께 남겨야 하는 관측이다.

| 값 | 뜻 | 기록에 필요한 근거 | 정상 재개 조건 |
|---|---|---|---|
| `not-installed` | 선택한 도구·브라우저·환경이 설치돼 있지 않다 | 설치 감지 명령의 종료 코드와 출력(egolite `command -v ego-browser`, aside `aside --version`) | 같은 감지 명령이 성공 |
| `connection-failed` | 설치는 됐지만 연결·기능 점검이 실패했다 | 설치 감지 성공 + 기능 점검 명령의 종료 코드와 출력 일부(egolite는 `ego-browser nodejs` 한 줄 스크립트) | 같은 기능 점검이 성공 |
| `not-logged-in` | 도구가 로그인을 요구하는데 로그인 상태가 아니다 | 도구가 낸 로그아웃 경고 원문 | 운영자가 로그인한 뒤 같은 점검이 성공. 감지 명령은 미확정(§5.2 Gap) |
| `not-configured` | 명시 설정이 필요한 기능이 미설정이다(예: Codex) | 설정 키 부재 또는 `false` 관측 | 운영자가 명시 설정 |
| `auth-failed` | 인증 실패(401 등) | 응답 코드와 공급자 이름(비밀은 가림) | 운영자가 자격을 정상화. 자동 재시도 없음 |
| `permission-denied` | 권한 거부(403, 정책 거부, 도구 거부) | 거부 응답 또는 거부한 가드 이름 | 권한이 정상 경로로 부여됨. 우회 없음 |
| `rate-limited` | 429 | 응답 코드, 응답이 준 재시도 시점이 있으면 그 값 | 그 시점 이후 운영자 또는 ControlService가 한 번만 재개. 반복 호출 없음 |
| `capacity-short` | 용량·자원 부족(디스크, 동시 실행 자리, 유료 한도) | 측정값 | 용량 확보. 추가 지출은 운영자 승인 |
| `budget-exhausted` | 감사 라운드 상한 소진 | 라운드 기록과 상한 값 | 운영자가 새 라운드를 승인하거나 범위를 나눔 |
| `check-not-passed` | 필수 검사가 skipped·누락·실패·취소·stale | 검사 이름, 결과, 그 결과가 가리킨 HEAD | 최신 HEAD에서 다시 실행해 성공 |
| `debt-present` | 감사 판정이 `PASS-WITH-DEBT` | 판정 파일 경로와 HEAD | 부채를 해소한 무부채 PASS |
| `verdict-not-pass` | 리뷰·감사·QA 증거의 판정값이 `FAIL`, `BLOCKED`, 또는 읽을 수 없음 | 증거 경로, 판정 HEAD, 판정값 원문 | 같은 HEAD의 새 판정이 `PASS` |
| `producer-not-allowed` | 증거의 생산자가 그 증거에 필요한 역할이 아니거나, 리뷰·감사·QA 증거를 writer attempt의 세션이 만들었다 | 증거 경로와 ControlService가 등록 때 관측한 생산자 결속(역할·세션·attempt·epoch) | 허용된 생산자가 같은 HEAD로 새 증거를 등록 |
| `ownership-unverified` | 소유권 확인 실패 또는 불가 | 실패한 확인 단계(claim 없음, holder 불일치, 저장소 읽기 실패) | 소유권 확인 성공 |
| `store-unavailable` | 로컬 claim 저장소가 없거나 읽을 수 없다 | 읽기 오류 원문 | 저장소 복구. 거울에서 재구성하지 않음 |
| `user-control` | 사용자가 제어권을 가져갔다(예: egolite `task.handOff()`, `EGO_TASK_SPACE_NAME_IN_USE`) | 도구가 낸 해당 표식 | 사용자가 제어권을 돌려줌 |
| `operator-ack-missing` | 기존 리더·writer의 안전 중단 ACK 같은 운영자 측 확인이 없다 | 확인 대기 기록 | 해당 ACK 수신 |

`landing`에서 병합 결과를 모르는 상황은 차단이 아니라 `landing` 자기 전이(원격 재조회)다. 원격 조회 자체가 실패하면 그때 위 사유 중 해당하는 값으로 `blocked`가 된다.

**인증·권한·429의 단일 규칙(REQ-AF-014)**: 401·403·429는 언제나 **실패한 그 동작**에 차단 기록을 남기고, 재시도 고리는 없다. 실패한 동작이 Issue 원장 거울 쓰기뿐이면 차단은 거울에 남는다(§4 `mirror_state: blocked(<사유>)`). Issue의 실행 상태(`state`)는 바뀌지 않는다. 권위는 로컬 claim 저장소이고 거울은 그것을 비출 뿐이기 때문이다(REQ-AF-011). 거울은 같은 사유의 정상 재개 조건이 충족된 뒤 ControlService가 한 번 다시 쓴다. 실행 경로의 동작(병합 요청, 검사 조회, 공급자 호출 등)이 실패하면 Issue가 `blocked`로 간다.

### 2.2 전이표

"쓰는 역할"에 없는 역할이 요청하면 거절한다. 표에 없는 간선(단계 건너뛰기 포함)도 거절한다.

| 출발 | 도착 | 가드(필요한 증거) | 쓰는 역할 |
|---|---|---|---|
| `queued` | `claimed` | 원자적 claim 성공, 같은 Issue에 살아 있는 claim 없음 | ControlService |
| `claimed` | `implementing` | attempt 발급, 작업 트리·브랜치·HEAD 기록, delivery `executed` 확인 | ControlService |
| `implementing` | `verifying` | worker가 후보 HEAD(전체 SHA)를 기록, 작업 트리 dirty 없음 | ControlService (worker 요청) |
| `verifying` | `implementing` | 같은 HEAD 검증 중 하나라도 FAIL → 수리 attempt 계속 | ControlService |
| `verifying` | `ready` | §11의 병합 수용 조건 전부 + writer 인계 + coordinator 인계 | ControlService |
| `ready` | `landing` | 병합 요청 직전 HEAD 재확인, 증거가 같은 HEAD를 가리킴 | ControlService |
| `landing` | `done` | 원격 `main`에 병합 반영 확인(GitHub 직접 조회) + 로컬 동기화 기록(동기화 전 수정·스테이징·미추적 경로마다 작업 트리 내용 digest·인덱스 항목·파일 모드가 동기화 뒤에도 같음, §4 `local_sync`) + 배포 기록 또는 사유가 적힌 `deploy: N/A`(REQ-AF-002) | ControlService |
| `landing` | `landing` | 병합 결과 미상 → 원격 재조회(재병합 없음) | ControlService |
| 진행 상태 아무거나 | `blocked` | 차단 사유 + 정상 재개 조건 기록 | ControlService |
| 진행 상태 아무거나 | `on-hold` | 보류 사유 + 재개 조건, 요청자 기록 | ControlService (운영자·coordinator 요청) |
| `blocked` / `on-hold` | `return_stage` | 재개 조건 충족 증거 | ControlService |

- 사람 소유 필드(업무명·수용 기준·우선순위·의존성)는 상태 전이와 별개로 운영자가 Issue에서 고친다.
- `done`으로 간 뒤에야 같은 실행 자리(slot)에 다음 dispatch를 받는다(REQ-AF-021).
- 카드 이관(Phase 2)은 어떤 상태에도 `done`을 만들지 않는다.

## §3 전달 상태 네 단계 (REQ-AF-003)

| 상태 | 뜻 | 증거 원천 | 증거가 아닌 것 |
|---|---|---|---|
| `sent` | 명령을 보냈다 | 보낸 쪽의 전송 결과 기록(큐에 넣은 행, 전송 응답 원문) | — |
| `received` | 받는 쪽이 명령을 꺼내 읽었다 | 받는 쪽이 남긴 수신 확인 행(명령 id 포함) | 전송 성공, idle 알림 |
| `executed` | 받는 쪽이 실행을 시작했다 | attempt 기록의 시작 행 + 작업 트리 HEAD 관측 | 수신 확인, 답장 메시지 |
| `completed` | 실행이 끝났고 결과가 디스크에 있다 | 결과 증거 파일(전체 HEAD SHA 포함)을 ControlService가 읽어 확인 | 답장 "완료", idle 알림, 전송 성공 |

ACK가 오지 않으면 다시 보내지 않고 먼저 받는 쪽 상태를 조회한다. 조회 결과가 `received` 이상이면 재전송하지 않는다. 조회 자체가 불가능하면 `blocked`로 둔다.

## §4 Issue 원장 필드 (REQ-AF-004, REQ-AF-002, REQ-AF-014)

원장은 Issue 본문(사람 소유 필드)과 고정 원장 댓글 하나(효과 상태 거울)로 나뉜다.

| 필드 | 형식 | 위치 | 쓰기 권한 |
|---|---|---|---|
| `task_name` | 문자열 | Issue 제목 | 운영자 |
| `acceptance_criteria` | 목록(문자열) | Issue 본문 | 운영자 |
| `priority` | 열거 `P0`~`P3` | Issue 본문(또는 라벨) | 운영자 |
| `depends_on` | 목록(Issue 번호) | Issue 본문 | 운영자 |
| `owner` | 문자열(세션 식별자 + 표시 이름) | 원장 댓글 | ControlService |
| `writer` | 문자열(세션 식별자) | 원장 댓글 | ControlService |
| `state` | 열거(§2.1) | 원장 댓글 | ControlService |
| `return_stage` | 열거(§2.1) 또는 비어 있음 | 원장 댓글 | ControlService |
| `run_id` / `attempt_id` | 문자열 | 원장 댓글 | ControlService |
| `head_sha` | 40자 16진수 | 원장 댓글 | ControlService (worker 보고를 확인한 뒤) |
| `check_links` | 목록(URL) | 원장 댓글 | ControlService (GitHub 조회 결과) |
| `review_evidence` / `audit_evidence` / `qa_evidence` | 목록(증거 경로 + 판정 HEAD SHA + 판정값 + 생산자 결속{역할, 세션, attempt, epoch}) | 원장 댓글 | ControlService (증거 저장소에서 읽은 값. 생산자 결속은 등록 때 ControlService가 관측한 값, §11) |
| `handoff_writer` / `handoff_coordinator` | 증거 경로 + HEAD SHA + 생산자 결속 | 원장 댓글 | ControlService (각 인계자가 등록한 파일을 읽은 값. writer 인계는 Worker, coordinator 인계는 Coordinator만, §11) |
| `next_action` | 문자열 | 원장 댓글 | ControlService (coordinator 제안을 받아 기록) |
| `block_reason` / `resume_condition` | 문자열 | 원장 댓글 | ControlService |
| `store_seq` | 정수(단조 증가) | 원장 댓글 | ControlService |
| `mirror_state` | 열거 `current` \| `stale` \| `blocked(<block_reason>)` | 로컬 claim 저장소(거울 쓰기 결과 기록) | ControlService. 거울 쓰기가 401·403·429로 실패하면 `blocked(<사유>)`. Issue `state`와 별개(§2.1a 단일 규칙) |
| `local_sync` | 동기화한 원격 SHA + 동기화 전 수정·스테이징·미추적 경로마다 `{path, before: {content_sha256, index_entry, mode}, after: {content_sha256, index_entry, mode}}`. 경로 목록만으로는 판정하지 않음(같은 경로가 다른 내용으로 덮여도 목록은 같기 때문) | 로컬 claim 저장소(원장 댓글에는 경로 수와 전체 digest만) | ControlService(REQ-AF-002 (b)). 경로 하나라도 before ≠ after면 `done` 금지 |
| `deploy` | 배포 기록 경로 + HEAD SHA, 또는 `N/A` + 사유 | 원장 댓글 | ControlService(REQ-AF-002 (c)) |
| `last_checked_at` | RFC 3339 UTC | 원장 댓글 | ControlService |

- 관찰 명령은 이 필드를 읽기만 하고, 읽기 권한이 실행 제어 권한을 주지 않는다.
- 쓰기 경로는 비밀 패턴(토큰, API 키, 자격 증명, 쿠키)을 가린 뒤 기록하고, 가린 사실을 남긴다.

## §5 정책표 (REQ-AF-012 ~ REQ-AF-016)

| 항목 | 정책 | 검증 목표 |
|---|---|---|
| GLM managed | 신규 구현과 로컬 환경 검증 모두에서 제외 | GLM 백엔드 호출 0건 |
| Codex | 운영자가 `workflow.agentic_factory.codex.enabled: true`를 명시했을 때만 사용. 미설정이면 자동 호출·기본 활성화·대체 호출 모두 금지 | 미설정 상태 Codex 호출 0건 |
| 인증 실패·권한 거부·429 | 우회 금지, 반복 호출 금지. 상태 보존, `blocked` + 정상 재개 조건 기록 | 같은 요청 재시도 0회 |
| 새 키·유료 용량·추가 지출 | 운영자 별도 승인 대상 | 승인 기록 없이 0건 |
| 브라우저 | `egolite` / `aside` / `chrome` 중 하나. 미선택일 때만 `chrome`. 명시 선택이 항상 우선. 선택 후 미설치·연결 실패는 `blocked`, 자동 대체 없음. 기존 선택 보존 | 선택하지 않은 브라우저 호출 0건 |

### 5.1 로컬 환경 매트릭스 (GLM 제외)

| 열 | 뜻 |
|---|---|
| `environment` | 환경 이름(후보 목록은 OQ-10) |
| `installable` | 설치 여부 관측 결과(`yes` / `no` / `denied`) |
| `version` | 실제로 실행한 버전 명령의 출력 |
| `start` / `receive` / `test` / `hold_resume` / `collect` | 기동, 작업 수신, 시험, 안전 보류·재개, 결과 회수 각각의 실제 실행 결과 |
| `verdict` | `PASS` 또는 `BLOCKED(<원인>)`. 미설치·권한 거부·용량 부족·미설정은 `BLOCKED`이며 성공으로 쓰지 않는다 |

각 단계 칸(`installable`부터 `collect`까지)은 실제로 실행한 명령, 종료 코드, 출력 일부 세 가지를 담는다(REQ-AF-015). 실행하지 않은 단계는 `not-run`이며 성공이 아니다. `PASS`는 여섯 단계가 모두 실행되고 성공했을 때만이다. `BLOCKED`는 멈춘 단계와 그 관측을 함께 적는다. 매트릭스에는 적어도 하나의 환경이 여섯 단계를 실제로 실행한 기록이 있어야 한다(전부 `BLOCKED`인 매트릭스는 매트릭스 시험을 통과하지 못한다, AC-AF-014).

Codex 환경은 명시 설정이 없으면 `BLOCKED(not-configured)`로 남고, 매트릭스를 채우려고 Codex를 켜지 않는다.

### 5.2 브라우저 어댑터 계약 (lane-21 `21-browser-contracts.md` 옮김)

아래 사실은 lane-21이 2026-10-29 로컬 macOS 한 대(ego-browser 0.6.1.13, aside 1.26.1023.1115)에서 도움말·가이드 출력과 웹 문서 요약으로 확인한 것이다. 이 세션은 그 명령을 다시 실행하지 않았다(spec.md §H). 구현 시 원문을 다시 확인해야 한다.

| 항목 | egolite (ego-lite의 ego-browser) | aside |
|---|---|---|
| 정본 출처 | `citrolabs/ego-lite` 저장소의 `skills/ego-browser/SKILL.md`, `references/install.md` | `docs.aside.com/help/developers` |
| 설치 감지 | `command -v ego-browser` | `aside --version` |
| 기능(연결) 점검 | `ego-browser nodejs`로 `console.log('ego-browser ready')` 실행 → `ego-browser ready` 출력 확인. 전용 `status` 명령은 없음(`Unknown command: status`) | 공식 문서에 전용 점검 명령 없음. 로그인 상태 감지 후보는 `aside account status`(출력 형식 미확인) |
| 사용 전 선행 | `ego-browser skill`을 먼저 실행하고 출력을 따름. 끝 줄 `[ego-browser:skill] end of guide`가 없으면 잘린 것 | 로그인: `ASIDE_CLI_AUTH_PATH`를 비공개 절대 경로로 지정 후 `aside login --email ...`, 이후 모든 명령에 같은 변수 |
| 실행 형태 | `ego-browser nodejs <<'EOF' … EOF`, `-e '<script>'`, `< script.js`. 스크립트는 셸 환경변수를 상속하지 않고 작업 디렉터리가 다르므로 절대 경로만 씀 | `aside "<prompt>"` 또는 `aside exec`, 이어 하기 `aside --session <id> "Continue"`, REPL `aside repl "<code>"`, MCP `aside mcp` |
| 권한 | 해당 없음 | `--permission ask\|guard\|full-access`(생략 시 Guard). `aside repl`에는 `--permission`이 없음(로컬 도움말 근거) |
| 작업 단위·종료 | 목표당 TaskSpace 하나. 이름이 같으면 재개. 사용자 소유 공간이면 `EGO_TASK_SPACE_NAME_IN_USE` → 사용자에게 물음. 성공 시 `task.finish({ keep: [] })` 한 번. 사용자 제어 필요 시 `task.handOff()` | 세션 id로 이어 하기 |
| 업그레이드 | `[ego-browser:notice]`가 보이면 작업을 끝내고 알림, `ego-browser upgrade`는 사용자 승인 뒤에만 | `aside --update` |
| 프로필 | 사용자가 명시하지 않으면 조회·선택하지 않음 | `--profile` 옵션 존재(도움말) |
| 비밀 | — | `ASIDE_CLI_AUTH_PATH`가 가리키는 자격 증명 파일(기본 `~/.aside/cli/auth.json`)을 읽거나 출력하거나 커밋하지 않음 |

차단 사유 구분(§2.1a)에 쓰는 근거:

- `not-installed`: 설치 감지 실패.
- `connection-failed`: 설치 감지 성공 + 기능 점검 실패. egolite는 두 단계를 따로 둬야 구분이 된다(기능 점검이 연결을 시도하기 때문).
- `not-logged-in`: aside 로그아웃 경고. 감지 명령은 미확정.
- `user-control`: egolite `task.handOff()` 또는 `EGO_TASK_SPACE_NAME_IN_USE`.

정책 대조와 남은 결정:

- egolite 스킬은 "사용자가 명시하면 실제 브라우저 명령을 먼저 실행하고 실패하면 그때 진단"을 요구하며, #1948의 "자동 대체 금지"와 충돌하지 않는다(lane-21 판단).
- SPEC-ASIDE-BROWSER-001은 aside 부재 시 **조용히 기본 도구로 진행**(REQ-ASB-012)하고 `CI=true`에서 aside를 제외(REQ-ASB-011)한다. 앞의 것은 #1948의 "명시 선택 후 실패하면 차단"과 정반대다. 이 SPEC은 #1948을 따르되, SPEC-ASIDE-BROWSER-001 개정은 대체 동작 검증 뒤로 미룬다. `CI=true` 처리는 OQ-13이다.
- aside `exec`의 권한 기본값(Guard 유지, `full-access` 금지 여부)과 `repl` 사용 범위(읽기·탐색·스크린샷만)는 OQ-13에 함께 둔다. 이 SPEC의 제안은 기존 정책 유지(`--permission` 생략 = Guard, `full-access` 금지, repl은 읽기 위주)다.
- chrome 쪽 계약(claude-in-chrome)은 조사되지 않았다.

lane-21 보고서가 남긴 확인 공백(그대로 옮김): ① ego-lite `references/clearing-state.md` 미열람 ② ego-lite 저장소 루트와 `lite.ego.app` 직접 열람 안 함 ③ ego-browser 연결 실패·앱 미실행 시 오류 문구·종료 코드 미측정 ④ aside 앱 실행 필요 여부, `--permission` 공식 정의, `exec` 전체 문법, REPL 120초 제한의 공식 출처 없음 ⑤ aside MCP 도구 목록은 로컬 도움말로만 확인 ⑥ chrome 계약 미조사 ⑦ WebFetch 요약을 거쳐 원문 대조가 아님 ⑧ 한 시점·한 기기 관측.

## §6 설정 스키마 초안 (REQ-AF-017, 실제 삽입은 보류)

새 섹션 `workflow.agentic_factory.*`를 제안한다. 키 이름은 미확정이다. 기존 설정 파일은 고치지 않으며, 실제 삽입은 계약 감사 통과와 lane-15 t1984·lane-14/16 체크포인트 ACK 뒤로 미룬다.

| 키 | 형식 | 기본값 | 검증 |
|---|---|---|---|
| `workflow.agentic_factory.enabled` | bool | `false` | bool 외 거절 |
| `workflow.agentic_factory.codex.enabled` | bool, 미설정 허용 | 미설정(=꺼짐) | bool 외 거절. 미설정·`false`는 호출 0건 |
| `workflow.agentic_factory.browser.selected` | 열거 `egolite` \| `aside` \| `chrome`, 미설정 허용 | 미설정(=사용 시 `chrome`) | 열거 외 거절. 미설정은 "선택 없음"이지 `chrome` 선택이 아니다(기존 선택 보존을 위해 둘을 구분). 선택 모델: 세 선택지 중 운영자가 고른 하나를 이 키에 저장하고, 설정을 다시 읽어도 같은 값이 나와야 한다. 어떤 경로도 운영자 대신 고르지 않는다(REQ-AF-016). 화면 표시는 범위 밖(OQ-18) |
| `workflow.agentic_factory.slots` | 정수 | `1` | 1 이상. 상한은 미정(OQ-16). 하한 위반·형식 위반은 거절 |
| `workflow.agentic_factory.lease.duration` | duration | `15m` | 0보다 커야 함, 범위 밖은 거절 |
| `workflow.agentic_factory.lease.renew_interval` | duration | `5m` | 1~59분의 정수 분, duration의 절반 이하. 범위 밖은 거절(현행 `workflow.factory_lease` 규칙과 같음) |
| `workflow.agentic_factory.audit.max_rounds` | 정수 | `harness.plan_audit_tier_ceilings`를 따름 | 1 이상. 소진 시 `blocked` |
| `workflow.agentic_factory.ledger.repo` | 문자열 `owner/name`, 미설정 허용 | 미설정(=현재 저장소) | 형식 외 거절 |

공통 규칙: 범위 밖 값은 키 경로·허용 범위·파일을 밝히고 거절한다. 값을 깎아 맞추지 않는다.

**Template-First 연동(하류 영향)**: 이 저장소의 `.moai/config/sections/*.yaml`은 `internal/template/templates/.moai/config/sections/*.yaml`과 짝이다. 실제 삽입 시 두 쪽을 함께 바꾸고, `internal/config`의 타입·기본값과 `internal/settings`의 스키마도 맞춰야 한다. 템플릿에는 `codex.enabled: true`를 넣지 않는다(현행 템플릿 중립성 규칙: 배포 템플릿에 `allow_write: true`를 두지 않는 것과 같은 이유).

## §7 세션 구조와 역할 — PROPOSAL, 미확정(provisional)

> 아래 명칭과 경계는 상위 설계 권고(§13)를 반영한 **제안**이다. 리더·사용자 확인 전까지 확정이 아니며, 어떤 구현의 근거도 되지 않는다.

### 7.1 제안 명칭표 (상위 설계 권고 반영)

| 개념 | 정식 영문 식별자(CLI·설정) | 한국어 표시 이름 | 출처 |
|---|---|---|---|
| 제품 | Agentic Factory (`agentic_factory`) | 에이전틱 팩토리 | 상위 설계 권고 |
| 상주 결정적 제어·복구(LLM 없음) | `ControlService` | 상주 제어 서비스 | 상위 설계 권고(리더 초안의 `controller`를 대체) |
| LLM 조율 | `Coordinator` (`coordinator`) | 조율자 | 상위 설계 권고 |
| 작업 세션 | `Worker` (`worker`) | 작업자 | 상위 설계 권고 |
| 실행 자리 | `slot` | 실행 자리 | 리더 초안. 기존 자원 `slot` lease와 이름이 겹침(§9.1) |
| 실행 단위 | `run` | 실행 | 리더 초안. 범위 정의 두 갈래(§7.3) |
| 시도 | `attempt` | 시도 | 리더 초안 |

역할로서 사용자에게 보이는 것은 Coordinator와 Worker 둘이다. ControlService는 사람이 직접 대화하는 역할이 아니라 상주 서비스다.

### 7.2 역할 정의 (제안)

| 역할 | 하는 일 | 쓸 수 있는 것 | 쓸 수 없는 것 | 수명 |
|---|---|---|---|---|
| `ControlService` | claim, lease·활성 확인, 배정 epoch, 조정(reconcile), 재개 판단을 규칙대로 결정. 판단에 LLM을 쓰지 않음. 활성 확인은 `homestate.ClassifyOwnerWith` 정본(spec.md §F.2)으로 하고, claim 때 pid와 프로세스 시작 시각을 함께 기록 | claim 저장소, Issue 원장 댓글, 전이 | 코드, SPEC 본문, 사람 소유 필드 | `start`로 떠서 정상 정지까지 상주. 재시작하면 저장소에서 상태를 이어받음 |
| `Coordinator` | 우선순위 판단, 다음 행동 제안, 운영자 대화 | 자기 판단 기록(제안·인계 파일) | claim·lease·attempt·epoch·전이를 직접 쓰기 | 운영자가 띄운 세션 수명. 끊겨도 실행은 계속 |
| `Worker` | attempt 하나의 구현·시험·인계 | 자기 attempt의 작업 트리, 시험 출력, **writer 인계** 증거(자기 epoch로만) | 다른 attempt, claim 저장소, 원장 댓글, 그리고 **코드 리뷰·보안 리뷰·독립 감사·QA 판정·coordinator 인계 증거**(§11 생산자 결속. Worker가 같은 SHA로 만들어도 게이트가 `producer-not-allowed`로 막음) | attempt 하나 동안. 끊기면 attempt는 `lapsed` 후보 |
| 리뷰·감사·QA 세션 | 같은 HEAD의 코드 리뷰·보안 리뷰(§11.1), 독립 감사, 공식 QA 스킬 실행 | 자기 판정 증거 | 코드, claim 저장소, 다른 역할의 증거 | 판정 하나 동안. writer attempt의 세션이 아니어야 함 |
| `slot` | Worker가 차지하는 용량 단위 | (주체가 아님) | — | ControlService가 설정값 `slots`만큼 둔다 |

### 7.3 run과 attempt — 두 정의, 결정 보류

- 리더 초안: `run`은 팩토리 실행 단위 하나(여러 Issue를 담을 수 있음), `attempt`는 특정 Issue + HEAD에 대한 Worker 하나의 시도.
- 조정자 둘째 요구: `run`은 Issue 하나의 처음부터 끝까지의 실행, `attempt`는 그 안의 시도 하나.
- 상위 설계 권고는 이 차이를 정하지 않았다. 이 SPEC도 확정하지 않고 OQ-2에 묶는다. 어느 정의든 아래 불변식은 유지한다.
  1. Issue 하나에는 claim 기록이 하나뿐이다(spec.md §F.1).
  2. Issue 하나에 살아 있는 attempt는 하나뿐이다.
  3. attempt 식별자는 Issue 번호와 순번을 담는다(예: `<repo>#<issue>/a<N>`). 인계 때마다 순번과 배정 epoch가 하나씩 오르고, 옛 attempt는 `lapsed` → `superseded`로 닫힌다(§14).
  4. 살아 있는 소유자의 attempt는 인계하지 않는다(spec.md §F.2).

## §8 명령 체계 — PROPOSAL, 미확정(provisional)

상위 설계 권고의 정식 명령은 `moai factory start`, `moai factory status`, `moai factory attach` 셋이다. 나머지(재개·보류·정지·관찰)는 이 SPEC의 이전 초안을 이어서 둔다. 모두 미확정이다.

**충돌 사실(이 트리에서 확인)**: `moai factory status`는 **이미 존재한다**(`internal/cli/factory_card.go:4175` `Use: "status"`). 제안대로 같은 이름을 쓰면 기존 동작을 바꾸는 것이 되어 REQ-AF-025(기존 명령 불변)와 부딪힌다. 이 SPEC은 "기존 `status` 출력을 그대로 두고 새 정보는 추가 필드나 별도 플래그로만 더한다"를 제안하되, 이름 결정은 OQ-12로 남긴다. `start`, `attach`는 현재 factory 하위 명령에 없다(같은 grep에서 0건).

| 목적 | 명령(제안) | 쓰기 여부 | 동작 계약 |
|---|---|---|---|
| 시작 | `moai factory start` | 쓰기 | ControlService를 띄운다. 이미 떠 있으면 새로 띄우지 않고 기존 것을 알린다(종료 코드 0, 기존 식별자 출력) |
| 상태·목록 | `moai factory status [--json] [--run <id>]` | 읽기 전용 | 실제 실행 상태를 조회해 표시. run이 여럿이면 목록 |
| 관찰 전용 모니터 | `moai factory status --watch` (제안) | 읽기 전용 | 아무것도 바꾸지 않는다. 실행 전후 상태 키가 같아야 한다 |
| 다시 붙기 | `moai factory attach --run <id> [--control]` | 기본 읽기 전용, `--control`일 때만 제어 | 실제 실행을 먼저 조회. 권한은 호출자의 기존 권한을 넘지 않는다 |
| 여러 run 중 선택 | `moai factory status` 후 `attach --run <id>` | 읽기 전용 | 후보가 둘 이상이면 id 없이 붙지 않는다(추측 금지, 종료 코드 12) |
| 안전 재개 | `moai factory attach --run <id> --resume --attempt <id>` (제안) | 쓰기 | 실제 실행 조회 → 소유자 활성 판정이 `live`·`indeterminate`면 새 attempt 없이 알림 → `dead`(spec.md §F.2) 또는 안전 정지 기록 시에만 epoch+1로 새 attempt |
| 보류 | `moai factory hold --run <id> --reason <text>` (제안) | 쓰기 | `on-hold` + 재개 조건 기록 |
| 정상 정지 | `moai factory stop [--run <id>]` (제안) | 쓰기 | 안전 지점 기록 후 정지. 강제 종료 없음 |

### 8.1 관찰 모니터 표시 항목 (REQ-AF-005)

관찰 모니터(`moai factory status` / `--watch`, 이름 미확정)는 프로젝트·Issue·owner마다 아래 항목을 보인다. 항목마다 읽은 원천과 읽은 시각을 함께 표시한다.

| 항목 | 원천 |
|---|---|
| 실행 상태(`state`, `return_stage`) | 로컬 claim 저장소 |
| 마지막 확인 시각(`last_checked_at`) | 로컬 claim 저장소 |
| HEAD SHA | 로컬 claim 저장소(attempt 기록) |
| 필수 검사 결과 | GitHub 직접 조회 |
| 진행(현재 단계, attempt 순번, 배정 epoch) | 로컬 claim 저장소 |
| 차단 사유와 정상 재개 조건 | 로컬 claim 저장소 |
| 다음 행동(`next_action`) | Issue 원장 댓글 |
| 거울 상태(`mirror_state`) | 로컬 claim 저장소 |

관찰과 제어의 분리: 모니터의 읽기 경로는 쓰기 능력을 갖지 않고, 출력에 제어 동작(재개·보류·정지 실행)을 담지 않는다. 제어는 §8 표의 쓰기 명령으로만 한다. 모니터 실행 전후로 저장소 순번과 트리 키가 같아야 한다(AC-AF-022).

종료 코드(계약 수준, 미확정):

| 코드 | 뜻 |
|---|---|
| 0 | 성공 |
| 2 | 사용법 오류 |
| 3 | 대상 없음 |
| 10 | 다른 소유자가 claim을 가짐(거절) |
| 11 | 소유권 확인 실패(쓰기 거절) |
| 12 | 후보가 둘 이상 — 명시 선택 필요 |
| 13 | 차단됨 — 출력에 `block_reason`과 정상 재개 조건 포함 |
| 14 | 원격 상태 미상 — 조회 필요, 재병합 없음 |
| 15 | 오래된 epoch로 쓰려 함(fencing 거절, §14) |
| 20 | ControlService에 닿을 수 없음 |

## §9 기존 → 새 대응과 호환 계약 — PROPOSAL, 미확정(provisional)

### 9.1 현행 진입 명령 — 이 트리에서 확인한 사실

| 사실 | 근거(파일:줄, 직접 읽음) | 확인 |
|---|---|---|
| `-f` / `--factory`는 factory run의 LEADER로 진입 | `internal/cli/cc.go:58` 도움말 | 확인 |
| `-l` / `--lane`은 실행 중인 factory에 다음 빈 `lane-<n>`으로 합류, 인자 없음 | `internal/cli/cc.go:65-66`, `:106`; 인자 거절 문구 `internal/cli/factory.go:98` | 확인 |
| `-f`와 `-l`을 함께 주면 거절("two entry tokens; a launch carries at most one") | `internal/cli/factory.go:100`(문구), `:271-272`(검사) | 확인 |
| `--leader`와 `--factory-run`을 함께 주면 거절("two different selectors") | `internal/cli/factory.go:280-282` | 확인 |
| `--leader`는 `-l`과만 함께 쓸 수 있음 | `internal/cli/factory.go:102`(문구), `:284-286`(검사: `if !p.LaneRole { return p, errors.New(leaderNeedsLaneEntry) }`) | 확인(0.2.0에서 검사 줄 읽음) |
| 같은 `--leader`나 같은 `--factory-run`을 **두 번** 주면 거절 | `internal/cli/factory.go:153-165`, `:174-189`를 읽었으나 반복 검사는 없고 값을 다시 대입함 | **확인 못 함(Gap)** — 읽은 범위에서는 나중 값이 이긴다. 다른 곳에 검사가 있을 수 있음 |
| `worker-<n>` / `agent-<n>`(및 bare `worker`/`agent`)는 레거시로 거절 | `--name` 경로 `internal/cli/factory.go:332-347`(`refuseLegacyEntryNames`), 플래그 값 경로 `:259-263`, 판정 함수 `internal/factory/bootstrap.go:237-242`(`IsLegacyFactoryRoleValue`) | 확인 |
| `lead` / `lead-<suffix>` 리더 이름도 레거시로 거절 | `internal/cli/factory.go:184-186`, `:333-336` | 확인 |
| 파서 진입 순서: 폐기된 `-k` 거절 → 레거시 이름 거절 → `parseFactoryFlag` | `internal/cli/factory.go:366-375`(`parseLauncherEntry`), `:142`(`parseFactoryFlag`) | 확인 |

### 9.2 용어 대응

| 기존 | 근거 | 새 용어(제안) | 비고 |
|---|---|---|---|
| leader (`moai cc -f`) | `internal/cli/cc.go:58` | `Coordinator` (+ 결정적 상태 권한은 `ControlService`로 분리) | 지금 leader는 판단과 상태 권한을 함께 가진다 |
| `lane-<n>` (`moai cc -l`) | `internal/cli/cc.go:65` | `Worker` (+ 차지한 `slot`) | **기존 lane 주소·세션은 즉시 이름을 바꾸지 않는다** |
| slot (자원 slot lease, `moai slot`) | `internal/factory/slot_lease.go` | 그대로(자원 잠금) | 새 `slot`(실행 자리)과 이름이 겹친다 — OQ-2 |
| card (`moai factory next/stage/complete`) | `internal/cli/factory_card.go` | Issue | Phase 2 이관 전까지 카드는 기존대로 동작 |
| run (팩토리 run id, 예: `tnhicy`) | #1938 본문, `internal/factorymsg/factory_run_*.go` | `run`(범위는 §7.3) | |
| lease (카드 임대, `moai factory renew`) | `internal/homestate/factory_lease.go`, `internal/cli/factory_card.go` | attempt의 lease(활성 힌트) + 배정 epoch(권위) | §14 |

### 9.3 명령 이관표 (lane-22 `c-factory-migration-map.md` §5의 C1~C11 옮김)

새 명령 열은 모두 미확정이다. "lane-22 판단"은 보고서 문구를 옮긴 것이고, 이 트리 확인 열은 이 세션이 직접 본 것이다.

| # | 구명령 | 현행 동작 | 새 명령(제안, 미확정) | lane-22 판단·비고 | 이 트리 확인 |
|---|---|---|---|---|---|
| C1 | `moai cc -f` / `moai glm -f` | 리더로 factory run 진입 | Coordinator 진입의 별칭으로 유지 | GLM 경로는 범위 제외 — `glm -f`는 변경 대상에서 뺌 | `cc.go:58` 확인, `glm` 미확인 |
| C2 | `moai cc -l` / `moai glm -l` (`--leader <name>`) | 실행 중인 factory의 레인으로 합류 | Worker 합류의 별칭으로 유지 | `lane-<n>` 레이블·기존 합류 세션 보존 | `cc.go:65-77` 확인 |
| C3 | `moai cc -w [name]` | 격리 작업 트리로 기동 | 유지(역할 명칭과 별개) | 폐기 대상 아님. #1847·#1813과 얽힘 | 도움말 미확인 |
| C4 | `moai factory next` | 다음 카드 임대 | Issue claim(이름 미정) | #1938·#1926·#1924와 함께 정의 | `factory_card.go:2324` |
| C5 | `moai factory stage <card> <state>` | 단계 전이 기록 | ControlService 전이(이름 미정) | 전달 상태 네 단계와 함께 | `factory_card.go:2728` |
| C6 | `moai factory complete` | 통합 창 + 착지 기록 | 병합 게이트 + 원격 확인(이름 미정) | gap 문서 G8과 연결 | `factory_card.go:2938` |
| C7 | `moai factory renew` | 임대 갱신 | attempt 활성 힌트 갱신(§14) | | `factory_card.go:2840` |
| C8 | `moai factory approve` / `decide` | 리더 수령증 / 운영자 결정 | Coordinator 인계 증거 / 운영자 결정 증거 | 권한 경계(#1924)는 이 SPEC이 정함 | `decide`는 `factory_card.go:4335`; `approve`는 `internal/cli/leader_approval_gate.go:286`에 `Use: "approve <card-id> --run <run-id>"` 존재 |
| C9 | `moai factory assign` / `bundle` / `status` / `runs` / `relaunch` / `handoff` / `messaging` / `fallback` / `pickup` / `merge` | 배정·묶음·조회·재기동·인계·메시징·폴백·픽업·병합 | 미정 | 조회 계열(`status`, `runs`)은 읽기 전용이라 제거 순서 마지막 | `status` 존재(`factory_card.go:4175`) — 제안 정식 명령과 이름 충돌(OQ-12) |
| C10 | `moai integration status/acquire/release/policy/remeasure/merge/preflight/resync/candidate` | 통합 창 | 미정 | lane-22 단독 등록 지점. 대체 검증 전 등록 해제 금지 | 미확인 |
| C11 | `moai todo …` (`moai gtd` 호환 별칭) | 카드 큐 | Issue 원장(Phase 2) | 이관은 Phase 2, 지금은 손대지 않음 | 미확인 |

### 9.4 단계 폐기 경로 (lane-22 대응표와 정렬: S1~S4)

이 SPEC의 이전 초안(S0~S3)은 lane-22의 네 단계로 바꿔 맞췄다(REQ-AF-025의 정렬 요구). 순서대로만 진행한다.

| 단계 | 의미 | 다음 단계로 가는 조건 | 증거 |
|---|---|---|---|
| S1 별칭 유지 | 구명령이 새 명령과 나란히, 같은 동작으로 계속 동작. 도움말·인자 해석·거절 문구·stdout·stderr·종료 코드가 §9.6 기준 golden 파일과 바이트 단위로 같음 | 새 명령 구현·등록 | AC-AF-024 호환 시험 PASS |
| S2 폐기 예고 | 구명령 실행 시 대체 명령을 알리는 stderr 한 줄만 추가. 그 한 줄을 뺀 나머지 출력과 종료 코드는 §9.6 기준과 바이트 단위로 같음 | 예고가 실제로 출력됨을 시험으로 확인 + 시나리오 AC 7개 PASS | 예고 문구 시험(AC-AF-024의 S2 사례) + AC-AF-001~007 PASS 기록 |
| S3 대체 검증 | 새 명령이 구명령의 모든 동작·거부 조건을 대체함을 같은 HEAD 시험과 독립 감사로 확인. 구명령 출력은 S2와 같은 기준(예고 한 줄 외 바이트 동일) | 감사 무부채 PASS + writer·Coordinator 인계 | 판정 파일 |
| S4 제거 | 구명령 등록 해제 | S3 증거가 디스크에 있는 부분만 + 운영자 승인 | 운영자 승인 기록 |

공통 불변 조건(lane-22 표와 이 SPEC 공통):

- 기존 세션 ID, 권한, lane 레이블(주소), owner, handle, 열린 임대, 리더 선출 상태를 바꾸지 않는다. 이 값들은 새 명령이 생기기 전과 뒤에 바이트 단위로 같아야 한다(AC-AF-024).
- 구명령을 즉시 제거하지 않는다.
- 실행 중인 세션을 재시작하지 않는다. S4 전에 구명령을 쓰는 세션이 끝났는지 확인하고, 강제 종료하지 않는다.
- GLM은 신규 구현에서 제외되므로 `glm` 쪽은 별칭 유지 대상일 뿐 새 기능을 붙이지 않는다. `moai codex -l`은 구현을 읽지 못했다.

### 9.5 열린 PR 착지 순서 후보 (lane-22 분석, 결정은 리더)

현재 병합은 전체 보류다(Issue "기존 병합 보류 유지"). 아래는 보류 해제 뒤의 **후보**이고, 파일 겹침만으로 도출됐다(lane-22 §2). 이 SPEC은 PR 상태를 다시 측정하지 않았다(제목·OPEN 상태만 확인).

1. #1944 — 규칙·지침 114파일 흡수, 충돌 범위가 가장 넓음
2. #1937 — `internal/config` 정리(설정 단독 writer 시작 전)
3. #1943 — 임대 복구(#1938 관련). claim 첫 조각이 그 위에서 시작하거나 그 뒤에 시작
4. #1916 ↔ #1921 — 둘 다 `launch_view.go`를 고쳐 직렬이어야 함. 순서 무관, 뒤쪽이 재흡수

충돌 배정 금지(lane-22 §3, 보고서 그대로): lane-14(t1932, t1979), lane-15(t1984 Codex 철회), lane-16(t1961)은 checkpoint ACK 전까지 새 writer·새 범위에 배정하지 않는다.

### 9.6 호환 기준 golden 파일 (S1 이전에 따로 커밋)

- 위치(제안): `internal/cli/testdata/agentic_factory_compat/`. 0.2.0 시점 이 경로는 없다(`ls internal/cli/testdata/agentic_factory_compat` → `No such file or directory`, exit 1, 트리 `e14b5bbac2`).
- 내용: 변경 전 빌드로 기록한 ① `moai cc --help`의 `-f`·`-l`·`--leader`·`--factory-run` 도움말, ② 거절 문구와 종료 코드(`-f`+`-l`, `--leader`+`--factory-run`, `--leader` 단독, `worker-<n>`·`agent-<n>`·`lead-<suffix>` 이름), ③ 기존 factory 기록의 lane 레이블·세션 ID·owner·handle·권한 모드 스냅숏.
- 순서(VCI §2.3): M6의 **첫 커밋**이 이 golden 파일만 담는다. S1 명령 등록은 그 다음 커밋부터다. 기준과 변경이 한 커밋에 섞이면 "변경 전에 측정했다"는 주장을 커밋 그래프가 증명하지 못한다.
- 현재 소스 문자열 근거(이 세션 측정): `grep -n -F 'are two entry tokens; a launch carries at most one' internal/cli/factory.go` → `100:	entryTokenConflict    = "%s and %s are two entry tokens; a launch carries at most one"`, exit 0, 트리 `e14b5bbac2`.

### 9.7 Codex 기본값 전환 단계 기록 (AC-AF-020의 운영 기록)

run 단계는 `progress.md`에 `## §J Codex Migration Stage Record` 절을 만들고, 절 첫 줄에 `run_base_sha: <run 단계 브랜치가 갈라져 나온 main 커밋의 전체 SHA>`를 적은 뒤, plan.md §F.3의 세 단계를 실행할 때마다 한 행씩 적는다. 행은 **백틱 없이** `| CX-1 | <stage_sha> | <evidence> |` 모양으로 적는다(AC-AF-020의 `grep -n '^| CX-'`가 걸리도록).

| 행 | 단계 | `stage_sha`(필수) | 필수 증거 |
|---|---|---|---|
| CX-1 | 관문 도입 | 관문을 넣은 커밋의 전체 SHA | — |
| CX-2 | 미설정 0건 검증 | AC-AF-013 PASS를 기록한 커밋의 전체 SHA | AC-AF-013 PASS 출력이 있는 증거 경로와 판정 HEAD |
| CX-3 | 기본값 변경 제안 | 제안을 담은 커밋의 전체 SHA | 제안 문서 경로 |

**순서는 줄 순서가 아니라 커밋 그래프로 증명한다(VCI §2.3)**: 세 단계는 각각 **별도 커밋**이며, CX-1의 커밋은 CX-2 커밋의 엄격한 조상, CX-2의 커밋은 CX-3 커밋의 엄격한 조상이어야 한다. 또 CX-2 커밋 이전(그 커밋 포함)에 Codex 기본값 파일(`.moai/config/sections/workflow.yaml`, `internal/config/defaults.go`, `internal/template/templates/.moai/config/sections/workflow.yaml`)의 Codex 키를 바꾼 커밋이 이 SPEC 브랜치에 없어야 한다. 표의 행 기록 자체는 단계 커밋과 같은 커밋이 아니어도 된다 — 판정은 행에 적힌 `stage_sha`로 한다.

이 절은 `§E.*` 이름 공간을 쓰지 않는다(era 분류 충돌 방지). 기존 Codex 기능의 제거는 이 세 행 뒤의 별도 작업이며 이 표에 넣지 않는다.

## §10 코드 경계 — 범위별 새 경로 하나 (제안)

각 범위는 새 경로 `internal/agenticfactory/<scope>/` 하나를 갖고, 공유 접점마다 writer 레인은 하나다. 조정자가 말한 "여섯 범위"의 정식 목록을 찾지 못해 아래를 제안으로 둔다(OQ-4).

| 범위 | 경로 | 맡는 요구사항 |
|---|---|---|
| `state` | `internal/agenticfactory/state/` | REQ-AF-001, 005, 023, 024 |
| `delivery` | `internal/agenticfactory/delivery/` | REQ-AF-003 |
| `ledger` | `internal/agenticfactory/ledger/` | REQ-AF-004, 011(거울) |
| `claim` | `internal/agenticfactory/claim/` | REQ-AF-006~011 |
| `policy` | `internal/agenticfactory/policy/` | REQ-AF-012~017, 022 |
| `gate` | `internal/agenticfactory/gate/` | REQ-AF-002, 018~021 |

공유 접점과 단일 writer(제안):

| 접점 | 단일 writer | 이유 |
|---|---|---|
| `internal/config`, `internal/settings`, `.moai/config/**`, 템플릿 설정 | 설정 담당 레인 1개 (현재 배정 없음 — t1984·lane-14/16 ACK 뒤) | PR #1937·#1944와 충돌 |
| `internal/cli/launcher.go`, `internal/cli/cc.go` | launcher 담당 레인 1개 | PR #1916·#1921과 충돌 |
| `internal/cli/factory_card.go`, `internal/factory/slot_lease.go`, `internal/homestate/factory_lease.go` | factory 담당 레인 1개 | PR #1943과 충돌, 첫 writer 범위 밖 |
| 명령 계층(`internal/cli`의 새 `factory` 하위 명령) | `state` 범위 레인 | S1 단계에서만. 명령의 판단(후보 선택, 관찰 전용 보장, 종료 코드 결정)은 `state` 범위에 두고 `internal/cli`는 인자를 넘기고 종료 코드를 그대로 돌려주는 얇은 연결만 갖는다. 그래서 AC-AF-022·023 시험이 `internal/agenticfactory/state/`에 있다 |

## §11 병합·완료 수용 규칙과 증거 저장 분리 (REQ-AF-018 ~ REQ-AF-021)

병합 게이트가 통과로 세는 것:

1. 최신 HEAD에서 필수 검사(현재 `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)` — 결정 D5)가 **실제로 실행되어** 성공. 보호 규칙의 `strict: false` 때문에 base에 뒤처진 head도 보호 규칙은 통과시키므로(lane-22 G1), 병합 게이트는 후보 HEAD가 그 검사를 받은 바로 그 커밋인지 따로 대조한다.
2. 같은 HEAD의 코드 리뷰와 보안 리뷰. Codex 미설정 프로젝트에서는 Codex가 아닌 리뷰 경로가 필요하다(결정 D2).
3. 필수 독립 감사의 무부채 PASS.
4. 같은 HEAD의 공식 QA 스킬 판정 PASS(spec.md §F.5).
5. writer 인계와 coordinator 인계가 모두 있음.

통과로 세지 않는 것: `skipped`, 누락, 실패, 취소, stale(다른 HEAD), 판정값 `FAIL`·`BLOCKED`·읽을 수 없음, `PASS-WITH-DEBT`, 허용되지 않은 생산자의 증거, 감사 예산 소진. 모두 `blocked`로 기록한다.

증거 생산자 결속(REQ-AF-018·020):

- 증거는 생산자가 ControlService의 등록 명령으로 올린다. ControlService는 **호출한 세션**의 등록 정보(역할, 세션 식별자, attempt, epoch)를 자기가 관측한 값으로 증거 항목에 붙인다. 증거 파일 안에 생산자가 스스로 적은 역할·세션 값은 판정에 쓰지 않는다.
- 역할별 허용 생산자:

| 증거 | 허용 생산자 | 추가 조건 |
|---|---|---|
| 코드 리뷰·보안 리뷰 | 리뷰 세션(§11.1 (a)) 또는 운영자(§11.1 (b)), Codex 명시 설정 시 Codex 리뷰 | writer attempt의 세션이 아님. GLM 아님 |
| 독립 감사 | 감사 세션 | writer attempt의 세션이 아님 |
| QA 판정 | QA 세션(공식 QA 스킬 실행) | writer attempt의 세션이 아님 |
| writer 인계 | 그 attempt의 Worker | 현재 epoch |
| coordinator 인계 | Coordinator 세션 | Worker 세션이 아님 |

- 결속이 표와 맞지 않으면 그 증거는 없는 것으로 보고 `blocked(producer-not-allowed)`로 기록한다.

### 11.1 공급자 중립 리뷰 경로 (결정 D2 개정, M4)

Codex가 명시 설정되지 않은 프로젝트에서 "같은 HEAD 코드·보안 리뷰"는 아래 둘 중 하나로 채운다. 어느 쪽도 Codex를 부르지 않는다.

- (a) **리뷰 세션**: ControlService에 리뷰 역할로 등록된 세션이, writer attempt의 세션과 다르고 GLM도 Codex도 아닌 공급자로 후보 HEAD를 리뷰해 판정(`PASS` / `FAIL` / `BLOCKED`)을 등록한다. 어떤 공급자·모델을 쓸지는 운영자 설정이다(명령 이름 미확정).
- (b) **운영자 기록 리뷰**: 운영자가 ControlService의 운영자 증거 경로로 같은 HEAD의 리뷰 판정을 기록한다. 이 경로는 #1924 대체 동작(운영자 결정 권한은 운영자 증거로만)과 같은 증거 원천을 쓴다.

둘 다 없으면 병합은 `blocked(not-configured)`다. Codex가 명시 설정된 프로젝트는 Codex 리뷰도 (a)와 같은 결속 조건으로 쓸 수 있다.

증거 저장 분리:

- 코드 후보는 card/Issue 브랜치의 커밋이다.
- 판정 증거(감사·리뷰·QA·인계)는 코드 후보 브랜치 밖에 둔다(spec.md 결정 D4). 제안 위치는 Issue별 증거 전용 ref(`refs/agentic-factory/evidence/<issue>`)이며 파일 배치는 현행 관례(`.moai/reports/<id>/`)를 따른다. 원장 댓글에는 증거 경로·판정 HEAD SHA·내용 digest만 비춘다. 이 ref를 원격에 올리는 것은 외부 공유 동작이라 운영자 확인 대상이다(OQ-14).
- 현행 사례: 판정 파일이 PR 브랜치 안에 커밋되고, 새 HEAD에 다시 묶는 PR이 실제로 있다(lane-22 G7: #1939 "re-pin the card t1988 verdict to the absorbed head" — PR 제목만 근거).
- 증거 항목마다 판정한 HEAD 전체 SHA와 생산자 결속을 적는다. 병합 게이트는 읽기 직전에 후보 HEAD와 대조하며, 다르면 stale로 본다. 증거를 따로 두는 것이 검증 강도를 낮추지 않는다는 뜻이 이것이다.
- 이렇게 하면 감사 결과를 기록하느라 코드 HEAD가 바뀌고, 바뀐 HEAD를 다시 감사해야 하는 순환이 생기지 않는다.

병합 결과 미상: 재병합하지 않고 원격 상태(PR 상태, `main`의 커밋 포함 여부)를 먼저 조회한다. 원격 반영이 확인되면 `done` → Issue 종료 → 다음 dispatch 순서로만 진행한다.

## §12 설계 입력 다섯 건 → 대체 동작 → 재현 시험

이 SPEC은 아래 Issue를 닫지 않는다. 처분은 대체 동작이 검증된 뒤 별도로 한다.

| Issue | 관측한 내용(직접 읽음) | 대체 동작 | 재현 시험(이름) |
|---|---|---|---|
| #1938 (OPEN) | 죽은 run에 `leased`·만료 상태로 남은 카드를 살아 있는 레인에 재발급할 정식 명령이 없음. 우회로 `assign`을 다른 run에 넣으면 카드 하나가 run 셋에 행 셋을 가짐 | claim은 Issue당 하나. 죽은 소유자의 만료 attempt는 같은 claim 아래 새 attempt로 재발급, 옛 attempt는 `superseded` | `TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim` (`internal/agenticfactory/claim/`) |
| #1926 (OPEN) | `--spec` 없이 발급된 카드가 kickoff에서 거절됨(`card carries no valid SPEC id`). 둘째 댓글: 감사자가 쓴 해시(`find|sha256sum`)와 kickoff의 `ComputeHash`가 달라 SPEC을 붙여도 풀리지 않음(측정됨) | 전이 가드는 SPEC 유무가 아니라 §2.2의 증거로 판단. SPEC 없는 Issue는 SPEC 해시 가드를 갖지 않는 경로로 전이하고, 해시를 쓰는 경로에서는 한 가지 해시 정의만 씀 | `TestState_SpeclessIssue_TransitionsWithoutSpecHash` (`internal/agenticfactory/state/`) |
| #1924 (OPEN) | `moai factory decide`가 요구하는 `MOAI_FACTORY_ROLE=operator`를 찍는 코드가 없음. 운영자 행동 카드는 plan을 벗어날 경로가 없음. 작성자는 게이트를 느슨하게 하자는 것이 아니라고 명시 | 운영자 결정 권한은 ControlService가 기록하는 운영자 승인 증거로만 생김. LLM coordinator 세션에 운영자 표식을 자동으로 찍지 않음 | `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused` (`internal/agenticfactory/state/`) |
| #1847 (OPEN) | launcher가 `-w -`의 단독 `-`와 `-w feature/foo`(실제 대상 `.claude/worktrees/feature+foo`)의 실제 대상을 놓침. Codex 연결이 수리 커밋 `e8c42a231c`를 만들었으나 원격 미확인 | attempt의 작업 트리는 정규화된 절대 경로로 claim에 기록되고, 진입은 그 경로로만 판단 | `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch` (`internal/agenticfactory/claim/`) |
| #1813 (OPEN) | 같은 이름의 `.moai/worktrees/<n>`가 있는데 `moai cc -w <n>`이 `.claude/worktrees/<n>`을 새로 만듦. 댓글: 이전 dirty 작업본 승계·폐기 사례. 기존 수정 t1791 / PR #1921 존재 | 새 attempt는 claim이 가리키는 트리를 재사용하고, 다른 트리를 만들지 않음. 기존 dirty 경로는 보존·참조 | `TestClaim_SameNameTree_NoDivergentTreeCreated` (`internal/agenticfactory/claim/`) |

## §13 상위 설계 권고 (도착함, PROPOSAL로 반영)

> 상위 설계 권고는 측정된 구현이 아니라 **설계 권고**다. 리더·사용자가 확인하기 전까지 확정이 아니다.

받은 내용:

- 제품 이름 Agentic Factory. 사용자에게 보이는 역할은 Coordinator와 Worker. LLM이 아닌 상주 결정적 제어는 ControlService(LLM 조율 역할인 Coordinator와 구분).
- 정식 CLI 제안: `moai factory start`, `moai factory status`, `moai factory attach`. 재연결·여러 run 선택·안전 재개 계약은 그대로.
- 호환: 기존 lane 주소와 세션 이름을 즉시 바꾸지 않는다. 바뀌지 않는 ID, owner, handle을 보존한다.
- 새 교차 점검: 카드 TTL/renew 사용자 결정 대 "지속 배정 + epoch" 제안(§14).

반영한 곳: §7(명칭·역할), §8(명령·종료 코드), §9(대응·호환), §14(교차 점검), spec.md REQ-AF-023~025, acceptance.md AC-AF-024.

이 권고가 정하지 않은 것: `run`의 범위(§7.3), 새 실행 자리 `slot`과 기존 자원 `slot`의 이름 충돌, `moai factory status`의 기존 명령 충돌(§8, OQ-12). 이 셋은 미결로 남는다.

## §14 교차 점검 — 카드 TTL/renew(기존 사용자 결정) 대 지속 배정 + epoch(제안)

### 14.1 비교 대상

- **기존 사용자 결정(바꾸지 않음)**: `workflow.factory_lease.duration: 15m`, `renew_interval: 5m`(`.moai/config/sections/workflow.yaml` 47-49행). 범위 검사는 `internal/homestate/factory_lease.go`가 하며 범위 밖 값은 팩토리 개방을 거절하고 깎지 않는다. 기한이 지난 뒤의 `renew`는 `refused-expired`(`internal/cli/factory_card.go:2817`, #1938에서 관측). 별도로 자원 slot lease(`internal/factory/slot_lease.go`)는 기한이 지나면 소유자가 살아 있어도 `--force` 없이 인계된다(REQ-RSL-006).
- **제안**: 배정은 지속되고(시간이 지나도 저절로 사라지지 않음), 배정마다 단조 증가하는 epoch를 붙인다. 모든 쓰기는 자기 epoch를 싣고, 저장소는 현재 epoch보다 작은 epoch의 쓰기를 거절한다(fencing). TTL은 활성 힌트로만 쓴다.
- **이 SPEC의 만료 claim 규칙 초안**(spec.md §F.2): 기한은 지났지만 소유자가 살아 있으면 인계하지 않고 `lapsed`로 표시.

### 14.2 항목별 비교

| 항목 | 기존 카드 TTL/renew | 기존 자원 slot lease | 제안(배정 + epoch) | 충돌 여부 |
|---|---|---|---|---|
| 만료의 뜻 | 기한 경과 → 임대 행이 `leased`·만료 상태로 남음. 자동 인계는 관측되지 않음(#1938: 행이 그대로 남음) | 기한 경과 → 소유자가 살아 있어도 인계 가능 | 기한 경과는 "확인하라"는 신호일 뿐, 배정은 남음 | slot lease와는 **충돌**. 카드 임대와는 직접 충돌 없음 |
| 갱신 | 5분마다 renew, 기한이 지나면 `refused-expired` | 해당 없음(보유자 선언 기한) | renew는 활성 힌트를 새로 고침. epoch는 바뀌지 않음 | 기한이 지난 뒤 **살아 있는 소유자의 renew**를 받을지가 갈림(아래 14.3) |
| 누가 인계하나 | 정식 인계 명령 없음(#1938). 우회로 `assign`을 다른 run에 넣음. 카드 임대의 생존 확인은 PID만 보는 `FactoryProcessAlive`(`factory_card.go:401`) | 다른 세션이 acquire로 바로 인계. `Stale()`은 `FactoryProcessAlive`(`slot_lease.go:183`) | ControlService만. `ClassifyOwnerWith` 판정 `dead` 또는 안전 정지 기록 뒤 epoch+1 | 카드 쪽은 "경로 없음"을 메우는 것이라 충돌 아님 |
| 이중 writer 방지 | 임대 행 기준. run이 다르면 행이 늘어남(#1938의 행 셋) | 보장 안 함(설계상 감수) | epoch fencing으로 옛 writer의 늦은 쓰기를 거절 | 제안이 더 강함 |
| 소유자 사망 시 남는 것 | 임대 행, 카드 트리 | 기록 파일 | 배정 기록, epoch, 트리·브랜치·SHA·dirty 목록 | 충돌 없음 |
| #1938 재발급 경로 | 없음 | 해당 없음 | 같은 claim 아래 epoch+1, 옛 attempt `superseded` | 충돌 없음(새 경로) |

### 14.3 판정

- **카드 TTL/renew와는 원칙상 충돌하지 않는다.** 15분·5분 값과 범위 검사는 그대로 둘 수 있다. 그 값을 "인계 권한을 주는 시계"가 아니라 "활성 확인을 시작하는 시계"로 읽으면 된다.
- **자원 slot lease의 인계 규칙과는 충돌한다.** 그래서 claim을 slot lease로 만들지 않는다(spec.md §F.1). slot lease 자체는 자원 잠금으로 그대로 둔다.
- **한 군데는 기존 결정을 건드릴 수 있다.** 현재 `renew`는 기한이 지나면 `refused-expired`다. 제안을 끝까지 따르면 "기한은 지났지만 살아 있고 epoch가 그대로인 소유자"의 renew를 받아 활성 힌트를 되살리는 편이 자연스럽다. 이는 현행 `refused-expired` 동작을 바꾸는 일이므로 이 SPEC은 정하지 않고 사용자에게 묻는다(OQ-11). 그 결정 전까지 새 흐름도 기한이 지난 renew를 거절하고, 대신 ControlService가 활성 확인 → `lapsed` 표시 → (사망 확정 시) epoch+1 재발급 경로로만 처리한다.

### 14.4 조정안 (기존 사용자 결정은 바꾸지 않음)

1. TTL(15m)과 renew(5m)는 활성 힌트다. 기한이 지나면 ControlService가 활성 확인을 시작한다.
2. 권위는 배정 epoch다. 모든 쓰기는 epoch를 싣고, 현재 epoch보다 작으면 거절한다(종료 코드 15).
3. 인계는 ControlService만, 활성 판정 `dead`(spec.md §F.2의 `homestate.ClassifyOwnerWith` 정본: PID 부재 또는 시작 시각 불일치) 또는 안전 정지 기록 뒤에만, epoch+1로 한다. `indeterminate`는 인계하지 않는다.
4. 기한이 지난 renew 처리 방식(현행 거절 유지 또는 같은 epoch의 살아 있는 소유자에게 허용)은 OQ-11 답을 따른다. 답이 오기 전에는 현행 거절을 유지한다.
5. 자원 slot lease와 `workflow.factory_lease` 키는 이 SPEC에서 고치지 않는다.
