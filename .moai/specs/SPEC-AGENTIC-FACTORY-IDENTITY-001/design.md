# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 설계

> Tier M SPEC이지만 검사 절차가 보안 계약의 본체라 별도 설계 문서를 둔다. 명령 이름과 연결 방식은 모두 미확정이다.

## §1 구성

- 위치(제안): `internal/agenticfactory/identity/`. 상위 design §10의 범위 표에 이 범위를 더하는 일은 상위의 다음 개정에서 한다(이 SPEC은 상위 문서를 고치지 않음).
- ControlService 안에서 동작하는 네 부분: 등록기(세션 등록·신원 결속), 부여기(역할 부여·운영자 인증), 계보 검사기(요청자 확인), 귀속기(증거 등록 때 귀속 판정, 상위로 내보내는 조회).
- 판정에 쓰는 부여 목록은 ControlService 프로세스 안에만 있다. claim 저장소에는 감사·관찰용 사본(`role_grant`, `operator_ack`, 거절 기록)을 남긴다. 사본은 다시 읽어 권한으로 쓰지 않는다(REQ-AFI-001).

## §2 세션 등록과 신원 (REQ-AFI-004)

1. 요청이 오면 ControlService는 연결 상대 pid를 OS에서 직접 얻는다(OQ-AFI-2). 요청 본문에 적힌 pid는 쓰지 않는다.
2. §3의 계보 검사를 먼저 한다. 통과하지 못하면 등록하지 않는다.
3. 그 pid의 시작 시각을 `ProbeProcessIdentity`로 읽는다. 읽지 못하면 등록하지 않는다.
4. session id를 발급하고 `{session_id, pid, process_start, run_id, registered_at}`를 기록한다. 살아 있는 프로세스에 이미 묶인 session id로 들어온 요청, 이미 session id를 가진 프로세스의 두 번째 등록은 거절한다.
5. 이후 증거 등록마다 1을 다시 하고, 관측 pid가 결속 pid와 같으며 `ClassifyOwnerWith(결속 pid, 결속 시작 시각)`가 `live`일 때만 받는다.

## §3 계보 검사 (REQ-AFI-003)

### 3.1 운영자 기준점

- ControlService 기동 명령(제안 `moai factory start`)은 자기 부모 프로세스(운영자 셸)의 pid와 시작 시각을 **운영자 기준점**으로 기록한다. 기준점은 재기동 때만 바뀐다.
- 기동 시점에 등록 세션은 없으므로 기준점 기록에 부정 조건은 적용되지 않는다. 기동은 운영자 행위로 두며, 등록 세션이 기동하는 경우는 잔여 위험이다(spec.md §F.3).

### 3.2 사슬 걷기

요청자 pid에서 시작해 부모 쪽으로 올라간다. 고리 하나마다:

1. 현재 프로세스 `p`의 부모 pid를 읽는다. 원천: Linux `/proc/<pid>/stat`, darwin·BSD `kern.proc.pid` sysctl의 `Eproc.Ppid`, Windows Toolhelp 스냅숏의 `ParentProcessID`. 이 트리의 `internal/session/proc_info_{linux,bsd,windows}.go` `platformProcInfo`가 같은 원천을 쓴다(재사용 후보).
2. `p`와 부모의 시작 시각을 `ProbeProcessIdentity`로 읽는다.
3. 부모의 시작 시각이 `p`의 시작 시각보다 늦으면 PID 재사용으로 보고 **판정 불가**다. Windows는 부모가 끝난 뒤 그 pid가 재사용될 수 있으므로 이 검사가 특히 필요하다.
4. `p`가 등록 세션의 결속 `(pid, 시작 시각)`과 같으면 **자손(또는 본인)**으로 판정하고 멈춘다.
5. `p`가 운영자 기준점의 `(pid, 시작 시각)`과 같으면 **기준점 도달**로 판정하고 멈춘다.
6. 부모가 pid 1(Linux init/systemd, darwin launchd)이거나 Windows에서 부모 프로세스가 없으면, 기준점에 닿기 전에 사슬이 끊긴 것이므로 **기준점 미도달**이다.
7. 프로세스가 사라졌거나, 읽기가 권한 때문에 거부되거나, 플랫폼 구현이 없으면 **판정 불가**다.
8. 걷는 깊이에 상한을 둔다(제안 64). 넘으면 판정 불가다.

결과와 처리:

| 결과 | 처리 | 거절 기록 코드(내부) |
|---|---|---|
| 자손(또는 본인) | 거절 | `requester-is-session-descendant` |
| 기준점 미도달 | 거절 | `requester-outside-operator-anchor` |
| 판정 불가 | 거절 | `requester-ancestry-indeterminate` |
| 기준점 도달, 등록 세션을 지나지 않음 | 다음 검사(§4)로 | — |

거절 기록 코드는 이 SPEC 안의 기록이며 상위의 `block_reason` 값이 아니다. 거절 결과로 필요한 역할이 없으면 상위 게이트는 `blocked(operator-ack-missing)`를 기록한다.

### 3.3 검사 순서

4(자손 판정)는 5(기준점 도달)보다 먼저다. 기준점 아래에 있는 등록 세션의 자손이 기준점 도달로 통과하지 않게 하려는 것이다.

## §4 운영자 인증 (REQ-AFI-002)

1. 부여 요청자가 §3을 통과하면, ControlService는 프로세스 표에서 요청자의 제어 단말을 직접 읽는다(설계 후보: Linux `tty_nr`, darwin `e_tdev`, Windows 콘솔 연결. spec.md §H). 제어 단말이 없으면 거절한다. 요청 본문의 `method` 값이나 단말 상태 주장은 읽지 않는다.
2. ControlService는 대상 run·역할·session id·pid·시작 시각과 일회성 확인 값을 요청자에게 보이고, 운영자가 대상 session id와 확인 값을 입력해 답한다.
3. ControlService는 `operator_ack{operator, method: interactive-tty(서버 관측), run_id, role, target_session_id, target_pid, target_process_start, requester_pid, requester_process_start, issued_at, used}`를 만든다.
4. 부여 직전 검사: 대상 필드가 부여 대상의 결속 값과 모두 같음, 아직 쓰지 않음(ack 하나에 부여 하나), 만료 아님(대상 세션의 활성 판정이 `dead`이거나 run이 끝났으면 만료).
5. 하나라도 어기면 부여하지 않는다. 기본 부여는 없다.

## §5 역할과 증거 종류 (REQ-AFI-005, REQ-AFI-006)

| 증거 종류 | 필요한 역할 | 추가 조건 |
|---|---|---|
| 코드 리뷰 | 코드 리뷰 | 그 run에서 Worker였던 세션이 아님 |
| 보안 리뷰 | 보안 리뷰 | 같음 |
| 독립 감사 | 감사 | 같음 |
| QA 판정 | QA | 같음 |
| coordinator 인계 | Coordinator | Worker였던 세션이 아님 |
| writer 인계 | 그 attempt의 Worker(현재 epoch) | — |

- 역할 분리: 한 세션은 run 안에서 역할 하나다. Worker였던 세션, Coordinator 세션, 이미 다른 역할을 가진 세션에는 인증이 유효해도 부여하지 않는다.
- 증거 종류와 역할이 맞지 않으면(예: 코드 리뷰 역할 세션이 보안 리뷰를 등록) `unattributable{reason: role-kind-mismatch}`다.
- 운영자 기록 리뷰(상위 design §11.1 (b))는 운영자가 §4와 같은 확인을 거쳐 등록하는 증거이며, 귀속 결과의 `role`은 `operator`다. 상위 게이트는 코드 리뷰·보안 리뷰 종류에서만 `operator` 귀속을 허용한다.

## §6 상위로 내보내는 조회 (REQ-AFI-007)

- 모양(제안): `Attribute(evidence_id) → attributed{kind, role, session_id, attempt, epoch} | unattributable{reason}`.
- `reason` 값: `not-registered`, `identity-mismatch`, `no-grant`, `role-kind-mismatch`, `was-worker`, `grant-from-previous-start`, `identity-spec-incomplete`(이 SPEC 완료 전 실제 구현 자리가 돌려주는 값).
- 읽기 전용이다. 조회는 부여·등록·저장소 순번을 바꾸지 않는다.
- 상위가 시험에 쓰는 대역(stub)은 이 모양을 그대로 따른다. 상위 AC-AF-018의 게이트 사례는 대역을 쓰고, 잠금 사례는 실제 구현 자리(이 SPEC 완료 전에는 언제나 `unattributable{reason: identity-spec-incomplete}`)를 쓴다(상위 §C.5.1 DEC-AF-02).
