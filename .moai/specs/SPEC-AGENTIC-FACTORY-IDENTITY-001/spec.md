---
id: SPEC-AGENTIC-FACTORY-IDENTITY-001
title: "AgenticFactory 역할 부여·세션 신원·운영자 인증 — 하위 계약"
version: "0.1.0"
status: draft
created: 2026-10-29
updated: 2026-10-29
author: GOOS
priority: P0
phase: "v3.3.0 target"
module: "internal/agenticfactory/identity"
lifecycle: spec-anchored
tags: "agentic-factory, identity, role-grant, operator-authentication, process-ancestry, security"
tier: M
issue_number: 1948
related_specs: [SPEC-AGENTIC-FACTORY-001]
---

# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 역할 부여·세션 신원·운영자 인증

## HISTORY

| 버전 | 날짜 | 작성 | 내용 |
|---|---|---|---|
| 0.1.0 | 2026-10-29 | manager-spec | 상위 SPEC-AGENTIC-FACTORY-001 0.4.0에서 분할해 새로 만든 하위 SPEC(draft). plan-audit iteration-4(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter4.md`)의 분할 제안과 운영자 재결정(2026-10-29, 리더 경유)에 따라 상위의 결정 DEC-AF-01, REQ-AF-018의 역할 부여·신원 조항, design §7.2·§11·§11.1·§11.2의 해당 내용, AC-AF-018의 역할 부여 사례 R1~R10, OQ-19를 옮겼다. iteration-4 N1(요청자 확인 부재)은 서버 쪽 요청자 계보 검사(REQ-AFI-003)로, N2(증거 종류와 역할의 일치)는 REQ-AFI-006·AC-AFI-008로, N4(Worker 요청 금지 조항의 단독 시험)는 AC-AFI-003으로 다룬다. |

## §A 범위와 관계

### A.1 규범 원천

- 운영자 결정 DEC-AF-01(2026-10-29, `AskUserQuestion` 답변을 리더가 전달, 상위 SPEC 0.4.0 §C.5.1에 처음 기록). 운영자가 정한 내용은 네 가지다. ① 리뷰·감사 역할은 ControlService만 부여한다. ② 리뷰·감사 역할 등록에는 운영자 인증이 필요하다. ③ Worker는 자기에게도 다른 세션에게도 역할을 줄 수 없다. ④ 한 세션이 Worker이면서 감사자일 수 없다.
- 운영자 재결정(2026-10-29, 리더 경유): 역할 부여·세션 신원·운영자 인증을 이 하위 SPEC으로 떼어 내고, 서버 쪽 요청자 확인(등록된 세션의 자손 프로세스가 낸 요청 거절, 계보 판정 불가 시 거절)을 넣는다.
- 상위 규범 원천인 GitHub Issue #1948의 "같은 HEAD 코드/보안 리뷰 및 필수 독립감사 무부채 PASS" 조건이 이 계약의 목적이다.

### A.2 상위 SPEC과의 관계

- 상위: SPEC-AGENTIC-FACTORY-001(AgenticFactory 실행 계약). 이 SPEC은 상위의 병합 게이트(상위 REQ-AF-018)가 리뷰·보안 리뷰·감사·QA·coordinator 인계 증거의 **생산자를 믿을 근거**를 제공한다.
- 프론트매터에는 `related_specs`만 적었다. `spec-frontmatter-schema.md`에는 부모·자식 관계를 나타내는 정식 필드가 없고, `depends_on`은 실행 전 점검이 완료를 요구하는 차단 관계라 쓰지 않았다. 관계의 방향은 이 절과 상위 §C.5.1이 정한다.
- **상위가 이 SPEC에서 받는 인터페이스**(REQ-AFI-007, design §6):
  1. 생산자 귀속 조회: 증거 항목 하나에 대해 `attributed{kind, role, session_id, attempt, epoch}` 또는 `unattributable{reason}` 중 하나를 돌려준다.
  2. 귀속은 증거 등록 때 ControlService가 관측·검사한 결과이며, 증거 파일 안의 자기 서술은 쓰지 않는다.
  3. 이 SPEC이 `completed`가 되기 전까지 상위의 잠금 조항(상위 §C.5.1 DEC-AF-02)이 적용되어, 상위 게이트는 이 종류의 증거를 모두 귀속 불가로 본다.
- 이 SPEC은 상위의 상태 모델·`block_reason` 집합(상위 design §2.1a)을 바꾸지 않는다. 역할이 없을 때 게이트가 쓰는 사유는 기존 값 `operator-ack-missing`, 귀속이 맞지 않을 때는 `producer-not-allowed`다.

### A.3 명칭

`ControlService`, `Coordinator`, `Worker`, `run`, `attempt`, 명령 이름은 상위와 같이 모두 **미확정(provisional)**이다. 요구사항은 이름이 아니라 권한 경계와 검사 동작을 묶는다.

## §B 이 트리에서 읽은 사실 (트리 `8dc835756b`)

- 대화형 서명 방식 상수: `internal/contract/seal.go:47` `MethodInteractiveTTY = "interactive-tty"`. 그 판정은 **클라이언트가 자기 표준입력을 검사**하는 방식이다: `internal/cli/contract.go:58` `contractStdinIsTerminalFn = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }`, `internal/contract/sign/defaults.go:76-77`(iteration-4 감사자 측정을 옮김, 이 세션은 `seal.go:47`만 다시 확인). 그래서 이 방식을 그대로 가져오면 서버는 클라이언트의 주장만 받게 된다(iteration-4 N1).
- 활성 판정 정본: `internal/homestate/factory_run_retire.go:152-168` `ClassifyOwnerWith`, 기본 구현 `ProbeProcessIdentity`(`internal/homestate/profile_lease.go:210-220`). 플랫폼별 시작 시각 원천은 `internal/homestate/process_fingerprint_unix.go:11`, `process_fingerprint_darwin.go:11`, `process_fingerprint_windows.go:10`의 `platformProcessFingerprint`(이번 실행에서 함수 위치만 grep으로 확인).
- 부모 pid 조회는 이미 있다: `internal/session/proc_info_linux.go:17-40`(`/proc/<pid>/stat`을 읽어 부모 pid를 돌려줌), `internal/session/proc_info_bsd.go:10-26`(darwin·BSD, `kern.proc.pid` sysctl의 `kp.Eproc.Ppid`), `internal/session/proc_info_windows.go:30-48`(Toolhelp 스냅숏의 `ParentProcessID`). 세 함수 모두 `platformProcInfo(pid) (ppid, comm, ok)` 모양이다(이번 실행에서 grep·head로 확인).
- 연결 상대 pid를 OS에서 얻는 코드(`SO_PEERCRED`, `LOCAL_PEERPID`, `GetNamedPipeClientProcessId`)는 이 트리에 없다(이번 실행 grep 0건).
- `internal/agenticfactory/`는 없다(상위 §B).

## §C 요구사항 (GEARS)

요구사항 7개(Tier M 상한 16).

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AFI-001 | Compound (Ubiquitous + Unwanted) | The ControlService SHALL be the only grantor of the evidence-producing roles — code review, security review, independent audit, QA, and Coordinator — and SHALL decide every grant and every attribution only from the grant list it created through its own grant procedure during the current start; a grant record found in the claim store, written by any other writer, or carried over from a previous start SHALL NOT confer a role, and no path SHALL grant a role by default. |
| REQ-AFI-002 | Event-driven | When a role grant is requested, the ControlService SHALL grant it only on an operator authentication that it observed itself — the requesting process, verified by REQ-AFI-003, holds a controlling terminal that the ControlService read from the process table, and the operator answered the ControlService's confirmation naming the target run, role, session id, pid, and process-start fingerprint — that is bound to exactly that target, unused, and unexpired; a method value or terminal state asserted by the client SHALL NOT count, and a missing, expired, reused, or mismatched authentication SHALL leave the role absent and the dependent gate `blocked(operator-ack-missing)`. |
| REQ-AFI-003 | Compound (Event-driven + Unwanted) | When a role-grant or session-registration request arrives, the ControlService SHALL take the requesting pid from the connection as observed by the operating system and SHALL walk that process's ancestry, verifying each link by pid and process-start fingerprint; it SHALL refuse the request when the requester is a registered session's process or a descendant of one, when the chain does not reach the operator anchor recorded at ControlService start, or when any link cannot be established, and SHALL NOT treat an indeterminate ancestry as acceptable. |
| REQ-AFI-004 | Ubiquitous | The ControlService SHALL issue each session id at registration, bind it to the pid and process-start fingerprint it observed, and on every evidence registration SHALL re-observe the registering process and admit the registration only when the observed pid equals the bound pid and `ClassifyOwnerWith` returns `live`; a copied session id, a reused pid with a different start fingerprint, and a second registration for a session id already bound to a live process SHALL be refused. |
| REQ-AFI-005 | Ubiquitous | No session SHALL hold more than one role in a run: the ControlService SHALL refuse a grant to a session that has been the Worker of any attempt in that run, to a Coordinator session, and to a session that already holds another role, even when the operator authentication is valid. |
| REQ-AFI-006 | Event-driven | When an evidence item is registered, the ControlService SHALL attribute it only when the producer's granted role equals the role that evidence kind requires (`design.md` §5) and, for code review, security review, audit, and QA evidence, the producer has not been the Worker of any attempt in the run; any other combination SHALL be recorded as `unattributable` and the dependent gate SHALL record `blocked(producer-not-allowed)`. |
| REQ-AFI-007 | Ubiquitous | The ControlService SHALL expose to the parent merge gate (SPEC-AGENTIC-FACTORY-001 REQ-AF-018) a read-only producer-attribution query that returns, per evidence item, either `attributed` with kind, role, session id, attempt, and epoch, or `unattributable` with its reason, computed by REQ-AFI-001 through REQ-AFI-006; the query SHALL NOT grant, register, or change any state. |

## §D 수용 기준 요약

`acceptance.md`에 9개(AC-AFI-001 ~ AC-AFI-009, Tier M 상한 16). 모두 run 단계가 이름 붙은 시험에서 `EXPECTED_RED`를 기록하기 전까지 채택 전이며 release-blocking이 아니다. 변형 점검표(acceptance §D)는 문서 위의 추론이다.

## §E 추적표 (REQ → AC)

| REQ | AC |
|---|---|
| REQ-AFI-001 | AC-AFI-001 |
| REQ-AFI-002 | AC-AFI-002 |
| REQ-AFI-003 | AC-AFI-003, AC-AFI-004, AC-AFI-005 |
| REQ-AFI-004 | AC-AFI-006 |
| REQ-AFI-005 | AC-AFI-007 |
| REQ-AFI-006 | AC-AFI-008 |
| REQ-AFI-007 | AC-AFI-009 |

## §F 결정

### F.1 DEC-AF-01 (운영자 결정, 상위 SPEC에서 이관)

- 운영자가 정한 네 가지(§A.1 ①~④)를 그대로 따른다.
- 이 SPEC이 구체화한 범위: 대상 역할은 증거를 내는 다섯 역할 전부다(QA와 Coordinator도 게이트가 세는 증거를 내므로 포함). Worker 역할은 ControlService가 attempt를 발급할 때 운영자 인증 없이 준다. "Worker이면서 감사자일 수 없음"은 "한 세션은 run 안에서 역할 하나, 그 run에서 Worker였던 세션은 다른 역할을 받지 못함"(REQ-AFI-005)으로 넓혔다. 리뷰어와 감사자를 겸하는 것도 같은 이유로 막는다.

### F.2 서버 쪽 요청자 확인 — 결정: 계보 검사 + 운영자 기준점

- **결정**: 부여·등록 요청마다 ControlService가 요청 프로세스의 계보를 직접 확인한다(REQ-AFI-003). 두 조건을 함께 건다. (가) **부정 조건**: 요청자가 등록된 어느 세션 프로세스이거나 그 자손이면 거절한다. (나) **긍정 조건**: 요청자의 조상 사슬이 ControlService 기동 때 기록한 **운영자 기준점**(기동 명령을 실행한 운영자 셸 프로세스)에 닿아야 한다. 어느 고리든 판정할 수 없으면 거절한다.
- **근거**: (가)만 두면 Worker가 자식 프로세스를 두 번 fork해 부모를 init/launchd로 바꾸면(고아 입양) 계보가 끊겨 "자손 아님"으로 읽힌다. (나)가 그 경로를 막는다. 입양된 프로세스의 사슬은 운영자 기준점에 닿지 않기 때문이다. 반대로 (나)만 두면 운영자 셸 아래에서 띄운 Worker의 자식도 기준점에 닿으므로 (가)가 필요하다.
- **판정 불가의 처리**: 고리 중 프로세스가 사라졌거나, 시작 시각을 읽지 못했거나, 부모의 시작 시각이 자식보다 늦거나(PID 재사용), 권한 때문에 다른 프로세스를 읽지 못하거나, 플랫폼 구현이 없으면 그 요청은 거절이다. 기본 부여는 없다.
- **클라이언트 주장의 처리**: `method: interactive-tty` 같은 값을 클라이언트가 보내도 판정에 쓰지 않는다. 제어 단말 여부는 ControlService가 프로세스 표에서 직접 읽는다(REQ-AFI-002).

### F.3 이 결정이 닫는 것과 남기는 것

- **닫는 것**: (1) Worker나 다른 등록 세션이 직접 부여를 요청하는 경로. (2) Worker가 등록되지 않은 자식 프로세스를 띄우고 pty를 붙여 부여를 요청하는 경로. 자손이라 거절된다. (3) 자식을 고아로 만들어 계보를 끊는 경로. 운영자 기준점에 닿지 않아 거절된다. (4) 클라이언트가 방식 값만 주장하는 경로. 서버가 그 값을 읽지 않는다.
- **남기는 것(OQ-AFI-1)**: 운영자 기준점의 후손이면서 등록 세션의 후손이 아닌 프로세스를 Worker가 만들어 낼 수 있는 경로다. 대표적으로 운영자 터미널에 키 입력을 주입하는 경우가 있다(예: tmux `send-keys`로 운영자 창에 명령을 넣음, 운영체제 자동화로 키 입력 전송). 그 요청은 운영자 셸에서 나온 정상 계보를 갖는다. 사람이 실제로 확인했는지를 구분하는 일은 이 결정으로 닫히지 않는다. ControlService를 등록 세션이 먼저 띄우는 경우(기준점이 Worker 쪽으로 잡힘)도 남는다. 기동은 운영자 행위로 두지만 이를 기계로 강제하지는 않는다.

## §G 미결 질문

| ID | 질문 | 막는 것 |
|---|---|---|
| OQ-AFI-1 (상위 OQ-19 이관) | 운영자 터미널 입력 주입(§F.3)과 pty를 조작하는 프로그램은 대화형 확인을 흉내 낼 수 있다. OS 사용자 분리(ControlService와 Worker를 다른 uid로 실행하고 연결 상대 uid 비교), 하드웨어 키, 별도 장치 확인 중 하나를 요구할 것인가? 아니면 이 위험을 수용 기록으로 남기는가? 또 운영자 확인을 기다리는 대기 한도와 시험 하네스의 대기 방식은 무엇인가? | REQ-AFI-002 인증 강도, 상위 시나리오 05·06 실행 |
| OQ-AFI-2 | 연결 상대 pid를 OS에서 얻는 방법(Linux `SO_PEERCRED`, macOS `LOCAL_PEERPID`, Windows 이름 있는 파이프의 클라이언트 pid)은 이 트리에 없다. 연결 방식(Unix 도메인 소켓, 이름 있는 파이프)을 무엇으로 정하는가? | REQ-AFI-003·004 구현 |
| OQ-AFI-3 | 운영자 기준점을 ControlService 기동 명령의 부모 셸로 잡는 것이 맞는가? 운영자가 여러 터미널에서 일할 때 기준점을 여러 개 둘 것인가? 기준점을 바꾸는 방법은 재기동뿐인가? | REQ-AFI-003 긍정 조건 |

## §H 확인 공백

- iteration-4 감사자가 인용한 `internal/cli/contract.go:58`, `internal/contract/sign/defaults.go:76-77`은 이 세션이 다시 열지 않았다.
- `internal/session/proc_info_*.go`의 부모 pid 함수는 위치와 모양만 확인했다. Windows의 `ParentProcessID`는 부모가 끝난 뒤 재사용될 수 있다고 알려져 있으나(그래서 시작 시각 순서 검사를 둔다), 이 트리에서 측정하지 않았다.
- 프로세스 표에서 제어 단말을 읽는 플랫폼별 방법(Linux `/proc/<pid>/stat`의 `tty_nr`, macOS `kern.proc.pid`의 `e_tdev`, Windows의 콘솔 연결 여부)은 설계 후보이며 코드로 확인하지 않았다.
- Linux의 시작 시각 원천은 현행 정본이 `ps -o lstart=`(초 단위)라 같은 초 안의 PID 재사용을 구분하지 못할 수 있다(상위 OQ-17). 계보 검사의 고리 판정도 같은 한계를 갖는다.

## §I 범위 밖

### Out of Scope — 상위 계약 본체

- 상태 모델, 전이표, `block_reason` 집합, 병합 게이트의 검사·판정값·증거 저장 분리는 상위 SPEC-AGENTIC-FACTORY-001이 정한다. 이 SPEC은 생산자 귀속만 다룬다.
- 상위 잠금 조항(상위 §C.5.1 DEC-AF-02)의 해제는 이 SPEC의 `completed` 전이로 생기며, 이 SPEC이 상위 문서를 고치지 않는다.

### Out of Scope — 더 강한 인증 수단

- OS 사용자 분리, 하드웨어 키, 별도 장치 확인의 구현은 OQ-AFI-1의 답이 나오기 전까지 다루지 않는다.
- 운영자 터미널 입력 주입을 막는 장치(터미널 다중화기 권한 제한 등)는 다루지 않는다.

### Out of Scope — 기존 코드 수정

- `internal/contract`, `internal/session`, `internal/homestate`, `internal/discovery`의 기존 함수는 고치지 않는다. 재사용 후보로만 둔다.
- 설정 파일과 템플릿은 고치지 않는다.
