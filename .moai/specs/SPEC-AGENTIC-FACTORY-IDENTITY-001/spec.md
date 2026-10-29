---
id: SPEC-AGENTIC-FACTORY-IDENTITY-001
title: "AgenticFactory 역할 부여·세션 신원·운영자 인증 — 하위 계약"
version: "0.2.1"
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
| 0.2.0 | 2026-10-29 | manager-spec | plan-audit iteration-1 FAIL(`.moai/reports/SPEC-AGENTIC-FACTORY-IDENTITY-001/plan-audit-iter1.md`)의 D1~D11과 정본 정합을 고치는 판. 범위는 넓히지 않았다(상위 DEC-AF-05): 서버 쪽 요청자 확인으로 iteration-4 N1·N2·N4를 다루고, D2 완화 구현·여러 기준점(D8)·응답자 인증은 범위 밖이다. 정본과 부딪힌 구체화 세 건을 정본 쪽으로 고쳤다(상위 DEC-AF-21, 리더 결정): 역할 배제를 run 단위에서 Issue 단위로(REQ-AFI-005), Coordinator를 부여 대상에서 뺌(REQ-AFI-001), 부여 기록을 온라인 control ref에 두고 재시작 뒤 재구축·재검증(REQ-AFI-001). DEC-AF-01 ②(요청별 운영자 인증)는 문언을 그대로 두고 충돌을 좁게 기록했다(§F.1). D1: 중계 경로 위험을 OQ-AFI-1 잔여 위험으로 명시하고 요청자 연결로 온 응답을 거절(REQ-AFI-002). D2: OQ-AFI-1 운영자 위험 수용 기록을 완료 조건으로(§F.4, 상위 OD-4). D3: PID 재사용 한계를 정직하게 정의(REQ-AFI-004, design §3.3). D4: 사유 코드를 상위 단일 표와 맞춤(design §6, `no-grant` → `grant-not-in-control-ref`). D5: 어떤 AC도 그것을 만드는 마일스톤 전에 초록이 되지 않게 함(plan). D6: 요청 본문의 pid는 쓰지 않음(REQ-AFI-003). D7: 생산자 7 × 증거 종류 6 = 42칸 + 미등록 6 = 48사례의 생성 행렬 시험(AC-AFI-008). D8: 기준점 하나로 결정. D9: 운영자 기록 리뷰의 귀속 유지(REQ-AFI-006). D10: 부모 pid가 살아 있는 프로세스를 가리키지 않으면 모든 플랫폼에서 `requester-outside-operator-anchor`(design §3.2). D11: 잠금 해제는 상위 M7의 코드 변경 하나이며 lane-18 소유. 명칭 Worker → agent(상위 DEC-AF-08). 이 판은 감사받지 않았다. |
| 0.2.1 | 2026-10-29 | manager-spec | 상위 0.6.1과 맞춘 기록 판. 요구사항·수용 기준·설계 본문은 바뀌지 않았다. 확인한 것: 이 SPEC에는 상위에서 지운 예약 결정 행, 옛 OQ-22·OQ-24, 새 OD-7·OD-8에 해당하는 내용이 없다. 이 SPEC의 사유 코드(design §6)는 상위가 바이트 대조한 사유 문자열 8개와 겹치지 않는다(`grep -rn -F` 8개 값, 출력 없음). progress §E.1의 상위 판 표기를 0.6.1로 고쳤다. 이 판도 감사받지 않았다. |

## §A 범위와 관계

### A.1 규범 원천

- 운영자 결정 DEC-AF-01(2026-10-29, 리더 경유, 상위 0.4.0 §C.5.1에 처음 기록): ① 리뷰·감사 역할은 ControlService만 부여한다. ② 리뷰·감사 역할 등록에는 운영자 인증이 필요하다. ③ agent는 자기에게도 다른 세션에게도 역할을 줄 수 없다. ④ 한 세션이 agent이면서 감사자일 수 없다.
- 운영자 재결정(2026-10-29, 리더 경유, 상위 0.5.0 HISTORY): 역할 부여·세션 신원·운영자 인증을 이 하위 SPEC으로 떼어 내고, 서버 쪽 요청자 확인을 넣는다.
- 상위 규범 원천 Issue #1948(본문과 댓글 11개)의 "같은 HEAD 코드·보안 리뷰 및 필수 독립 감사 무부채 PASS"와 정본 [0]§1·[2]§21의 writer·auditor 독립성이 이 계약의 목적이다.

### A.2 상위 SPEC과의 관계

- 상위: SPEC-AGENTIC-FACTORY-001. 이 SPEC은 상위 병합 게이트(상위 REQ-AF-007·011)가 리뷰·보안 리뷰·감사·QA 증거, writer 인계, leader 수락의 **생산자를 믿을 근거**를 준다.
- 프론트매터에는 `related_specs`만 적었다. `depends_on`은 실행 전 점검이 완료를 요구하는 차단 관계라 쓰지 않는다.
- 상위가 받는 인터페이스(REQ-AFI-007): 증거 하나마다 `attributed{kind, role, session_id, attempt, epoch}` 또는 `unattributable{reason}`.
- 상위 잠금(상위 DEC-AF-02·04): 상위의 제품 귀속 원천은 컴파일 시점 상수 하나로 고정되어 `unattributable{identity-spec-incomplete}`만 돌려준다. 이 SPEC의 조회는 상위 M7의 코드 변경 하나로 연결되며, 그 착지는 이 SPEC `completed`, 이 SPEC의 AC 전부 PASS, OQ-AFI-1 운영자 위험 수용 기록과 같아야 한다. 이 SPEC은 상위 문서를 고치지 않는다.
- 게이트 사유는 상위 design §2.1b 단일 표가 정한다. 이 SPEC의 귀속 사유는 모두 게이트에서 `producer-not-allowed`가 된다. 역할 부재에 `operator-ack-missing`을 쓰지 않는다(0.1.0 §A.2 문장 폐기).

### A.3 명칭

상위 §A.4를 따른다: 내부 `Coordinator`(표시 `leader`), 실행자 `agent`, 결정적 제어 `ControlService`. 0.1.0의 `Worker`는 모두 `agent`로 읽는다.

## §B 이 트리에서 읽은 사실

- 대화형 서명 방식 상수: `internal/contract/seal.go:47` `MethodInteractiveTTY = "interactive-tty"`(0.1.0, 트리 `8dc835756b`). 판정은 클라이언트가 자기 표준입력을 검사하는 방식이라는 iteration-4 감사자 측정(`internal/cli/contract.go:58`, `internal/contract/sign/defaults.go:76-77`)은 다시 열지 않았다.
- 활성 판정: `homestate.ClassifyOwnerWith`(`internal/homestate/factory_run_retire.go:152-168`)는 fingerprint와 기록값의 **문자열 동등**만 본다. Linux 등의 fingerprint는 `ps -o lstart= -p <pid>` 출력 문자열(초 단위, 순서 비교 정의 없음, `process_fingerprint_unix.go`, 이 실행에서 읽음). darwin은 `"%d.%06d"`, Windows는 생성 시각(0.1.0에서 위치 확인).
- 부모 pid 조회: `internal/session/proc_info_linux.go:17-40`, `proc_info_bsd.go:10-26`, `proc_info_windows.go:30-48`, 모두 `platformProcInfo(pid) (ppid, comm, ok)`(0.1.0 확인).
- 연결 상대 pid를 OS에서 얻는 코드(`SO_PEERCRED`, `LOCAL_PEERPID`, `GetNamedPipeClientProcessId`)는 이 트리에 없다(0.1.0 grep 0건).
- 원격 control ref `refs/heads/moai-control` = `ae3a4edcaa529bc927357b345051b96cd04605a1`(이 실행 `git ls-remote`). 부여 기록을 둘 곳이다.

## §C 요구사항 (GEARS)

요구사항 7개(Tier M 상한 16).

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-AFI-001 | Compound (Ubiquitous + Unwanted) | The ControlService SHALL be the only grantor of the evidence-producing roles — code review, security review, independent audit, and QA — and SHALL decide every grant and every attribution only from grant records that it wrote itself to the online control ref through its own grant procedure and that carry the authentication binding of REQ-AFI-002; after a restart it SHALL rebuild the grant list from those control-ref records and SHALL re-verify each holder's binding and liveness (REQ-AFI-004) before any attribution; a grant record not written by the ControlService, present only in a local cache, or whose binding fails re-verification SHALL NOT confer a role, and no path SHALL grant a role by default. Agent and Coordinator sessions SHALL be registered, not granted. |
| REQ-AFI-002 | Event-driven | When a role grant is requested, the ControlService SHALL grant it only on an operator authentication that it observed itself — the requesting process, verified by REQ-AFI-003, holds a controlling terminal that the ControlService read from the process table, and the operator answered the ControlService's confirmation naming the target run, Issue, role, session id, pid, and process-start fingerprint — that is bound to exactly that target, unused, and unexpired; the answer SHALL count only when it is read on the ControlService's own controlling terminal, and an answer delivered over the requester's connection, a method value, or a terminal state asserted by the client SHALL NOT count; a missing, expired, reused, or mismatched authentication SHALL leave the role absent. |
| REQ-AFI-003 | Compound (Event-driven + Unwanted) | When a role-grant or session-registration request arrives, the ControlService SHALL take the requesting pid only from the connection as observed by the operating system, SHALL ignore any pid stated in the request body, and SHALL walk that process's ancestry verifying each link by pid and process-start fingerprint; it SHALL refuse either request when the requester is a registered session's process or a descendant of one or when any link cannot be established, SHALL additionally refuse a grant request whose chain does not reach the single operator anchor recorded at ControlService start, and SHALL NOT treat an indeterminate ancestry as acceptable. |
| REQ-AFI-004 | Ubiquitous | The ControlService SHALL issue each session id at registration, bind it to the pid and process-start fingerprint it observed, and on every evidence registration SHALL re-observe the registering process and admit the registration only when the observed pid equals the bound pid and `ClassifyOwnerWith` returns `live`; a copied session id, a reused pid whose start fingerprint differs from the bound one, and a second registration for a session id already bound to a live process SHALL be refused. |
| REQ-AFI-005 | Ubiquitous | No session SHALL produce review, security-review, audit, or QA evidence for a change it wrote: the ControlService SHALL refuse such a grant for an Issue to a session that has been the writer (agent) of any attempt on that Issue, and SHALL refuse a grant to a session that already holds another role for the same Issue, even when the operator authentication is valid. |
| REQ-AFI-006 | Event-driven | When an evidence item is registered, the ControlService SHALL attribute it only when the producer's role or registration matches the role that evidence kind requires in `design.md` §5 — including an operator record only for code-review and security-review evidence, the current attempt's agent only for writer-handoff evidence at the current epoch, and a registered Coordinator only for coordinator-acceptance evidence on a change it did not implement — and, for review, audit, and QA evidence, the producer has not been the writer of any attempt on that Issue; every other combination SHALL be recorded as `unattributable` with exactly one reason of `design.md` §6. |
| REQ-AFI-007 | Ubiquitous | The ControlService SHALL expose to the parent merge gate (SPEC-AGENTIC-FACTORY-001 REQ-AF-011) a read-only producer-attribution query that returns, per evidence item, either `attributed` with kind, role, session id, attempt, and epoch, or `unattributable` with one reason of the closed set of `design.md` §6, computed by REQ-AFI-001 through REQ-AFI-006; the query SHALL NOT grant, register, or change any state. |

DEC-AF-01 ②와 정본의 좁은 충돌 기록(REQ-AFI-002 주석): Where the canon allows more broadly than REQ-AFI-002 (no new per-card human approval), the broader allowance SHALL NOT be applied; REQ-AFI-002 stays as written until the operator's original wording of DEC-AF-01 ② is confirmed.

## §D 수용 기준 요약

`acceptance.md`에 9개(Tier M 상한 16). 모두 run 단계가 이름 붙은 시험에서 `EXPECTED_RED`를 기록하기 전까지 채택 전이다. 변형 점검표(acceptance §D)는 문서 추론이다.

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

### F.1 DEC-AF-01 (운영자 결정)

- 운영자가 정한 네 가지(§A.1 ①~④)를 따른다.
- 구체화(0.2.0, 정본에 맞춤, 상위 DEC-AF-21 리더 결정): 부여 대상 역할은 리뷰·보안 리뷰·감사·QA 네 역할이다. Coordinator(표시 `leader`)는 정본 S46에 따라 기존 진입 동작을 유지하며 부여 대상이 아니고 등록만 된다. Coordinator가 한 변경을 직접 구현했으면 그 변경의 증거는 귀속하지 않는다([2]§21). agent 역할은 attempt 발급 때 등록되며 운영자 인증이 없다. "agent이면서 감사자일 수 없음"은 "한 세션은 같은 Issue에서 역할 하나, 그 Issue의 writer였던 세션은 그 Issue의 리뷰·감사·QA 역할을 받지 못함"(REQ-AFI-005)으로 구체화한다. 다른 Issue의 단계별 역할은 정본 [2]§21·S46에 따라 허용한다.
- **② 요청별 운영자 인증 — 좁은 기록**: 정본([0]§1·§2·§5 R2)이 현재 문언보다 더 넓게 허용하는 부분(카드마다 새 사람 승인을 두지 않음)은 적용하지 않는다. REQ-AFI-002는 운영자 원문 확인 전까지 그대로다. 잔여 위험: 실제 운용에서는 리뷰·감사가 Issue마다 새 세션이므로 Issue마다 사람 확인이 생긴다. 이것은 정본 [0]§2와의 해소되지 않은 충돌이며 운영자 열린 항목이다(상위 OD-3).

### F.2 서버 쪽 요청자 확인 — 계보 검사 + 기준점 하나

- 부여·등록 요청마다 ControlService가 OS가 알려 준 연결 상대 pid에서 시작해 계보를 확인한다(REQ-AFI-003). 요청 본문의 pid는 읽지 않는다(D6).
- (가) 부정 조건(부여·등록 모두): 요청자가 등록된 세션 프로세스이거나 그 자손이면 거절.
- (나) 긍정 조건(부여만): 사슬이 기준점에 닿아야 함. 기준점은 **하나**이며 ControlService 기동 명령을 실행한 운영자 셸 프로세스다(D8 결정). 여러 기준점·tmux 서버 기준점 모형은 범위 밖이다. 기준점을 바꾸는 방법은 재기동뿐이다.
- 근거: (가)만 두면 고아 입양으로 계보가 끊겨 "자손 아님"으로 읽힌다. (나)가 막는다. (나)만 두면 운영자 셸 아래 띄운 agent의 자식도 기준점에 닿으므로 (가)가 필요하다. 등록에 (나)를 걸지 않는 이유: agent 세션은 ControlService가 띄우거나 lane이 띄우며 운영자 셸 아래에 있다는 보장이 없다.
- 판정 순서와 사유는 design §3.2.

### F.3 이 결정이 닫는 것과 남기는 것

- 닫는 것: (1) agent나 다른 등록 세션의 직접 부여 요청. (2) agent가 등록되지 않은 자식에 pty를 붙여 요청. (3) 고아로 만들어 계보를 끊는 경로. (4) 클라이언트가 방식 값·단말 상태·pid를 주장하는 경로. (5) 요청자 연결로 확인 응답을 보내는 경로(REQ-AFI-002).
- **남기는 것(OQ-AFI-1)**: 기준점의 후손이면서 등록 세션의 후손이 아닌 프로세스를 agent가 만들어 낼 수 있는 경로. 운영자 터미널에 키 입력을 주입하는 경우(예: tmux `send-keys`, 운영체제 자동화), 그리고 **비자손 중계자**: 등록 세션의 후손이 아닌 다른 프로세스(예: 운영자 셸에서 이미 떠 있던 도우미)가 agent의 요청을 받아 대신 보내거나, ControlService 단말의 확인 응답을 대신 입력하는 경우. 그 요청과 응답은 정상 계보와 정상 단말을 갖는다. 응답자 인증(누가 실제로 응답했는지 확인)은 이 SPEC의 범위 밖이다. ControlService를 등록 세션이 먼저 띄우는 경우(기준점이 agent 쪽으로 잡힘)도 남는다.

### F.4 완료 조건과 잠금 해제

- 이 SPEC은 OQ-AFI-1에 대한 운영자 위험 수용 기록(상위 OD-4) 없이 `completed`가 될 수 없다.
- 잠금 해제는 상위 M7의 코드 변경 하나이며 lane-18이 소유한다(D11). 그 변경은 이 SPEC `completed`, 이 SPEC AC 전부 PASS, 위 수용 기록과 같은 착지다. 이 SPEC은 해제를 스스로 일으키지 않는다.

## §G 미결 질문

| ID | 질문 | 막는 것 |
|---|---|---|
| OQ-AFI-1 | 키 입력 주입과 비자손 중계자(§F.3)의 잔여 위험을 운영자가 수용하는가? 수용하지 않는다면 OS 사용자 분리·하드웨어 키·별도 장치 확인 중 무엇을 별도 SPEC으로 요구하는가? 운영자 확인 대기 한도는? | 이 SPEC `completed`, 상위 M7 |
| OQ-AFI-2 | 연결 방식(Unix 도메인 소켓, 이름 있는 파이프)과 OS별 연결 상대 pid 획득 방법 | REQ-AFI-003·004 구현 |
| OQ-AFI-3 | (0.2.0 해소) 기준점은 하나, 재기동으로만 바뀜. 여러 기준점은 범위 밖 | — |
| OQ-AFI-4 | Linux에서 같은 초 PID 재사용을 구분하는 더 정밀한 시작 시각 원천을 요구할지(상위 OQ-17) | REQ-AFI-004 강도 |

## §H 확인 공백

- iteration-4 감사자 인용(`internal/cli/contract.go:58`, `internal/contract/sign/defaults.go:76-77`)은 다시 열지 않았다.
- 부모 pid 함수는 위치와 모양만 확인했다. Windows `ParentProcessID`가 부모 종료 뒤 재사용된 pid를 가리킬 수 있다는 것은 알려진 동작이며 이 트리에서 측정하지 않았다.
- 프로세스 표에서 제어 단말을 읽는 방법(Linux `/proc/<pid>/stat`의 `tty_nr`, macOS `kern.proc.pid`의 `e_tdev`, Windows 콘솔 연결)은 설계 후보이며 코드로 확인하지 않았다.
- control ref 운영 쓰기(CAS)는 상위 lane-18 구현을 기다린다.
- 하위 iteration-1 판정 파일의 D1~D11 본문은 리더 전달 사본으로 읽었다.

## §I 범위 밖

### Out of Scope — 상위 계약 본체

- 상태 모형, 전이표, 사유 집합, 병합 게이트, 잠금 상수는 상위 SPEC-AGENTIC-FACTORY-001이 정한다. 이 SPEC은 생산자 귀속만 다룬다.
- 잠금 해제 변경은 상위 M7(lane-18)이며 이 SPEC이 상위 문서를 고치지 않는다.

### Out of Scope — 더 강한 인증 수단과 범위 확대

- 응답자 인증, OS 사용자 분리, 하드웨어 키, 별도 장치 확인의 구현(iteration-1 D1 (가), D2 완화 구현)은 다루지 않는다.
- 여러 운영자 기준점, tmux 서버 기준점 모형(iteration-1 D8 확대)은 다루지 않는다.
- 운영자 터미널 입력 주입을 막는 장치는 다루지 않는다.

### Out of Scope — 기존 코드 수정

- `internal/contract`, `internal/session`, `internal/homestate`, `internal/discovery`의 기존 함수는 고치지 않는다. 재사용 후보로만 둔다.
- 설정 파일과 템플릿은 고치지 않는다.
