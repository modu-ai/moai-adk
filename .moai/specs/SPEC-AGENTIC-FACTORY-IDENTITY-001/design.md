# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 설계 (0.2.0)

> Tier M SPEC이지만 검사 절차가 보안 계약의 본체라 별도 설계 문서를 둔다. 명령 이름과 연결 방식은 미확정이다(OQ-AFI-2).

## §1 구성

- 위치(제안): `internal/agenticfactory/identity/`. 소유는 상위 design §10의 lane-18(core)이며, 잠금 해제 연결도 lane-18이 상위 M7에서 한다.
- ControlService 안의 네 부분: 등록기(세션 등록·신원 결속), 부여기(역할 부여·운영자 인증), 계보 검사기(요청자 확인), 귀속기(증거 등록 때 귀속 판정과 읽기 전용 조회).
- 부여 기록과 등록 결속은 **온라인 control ref**에 ControlService가 CAS로 쓴다(REQ-AFI-001, 상위 DEC-AF-06). 로컬 캐시는 재구축 전용이며 캐시만으로 역할을 주지 않는다.

## §2 등록과 부여 절차

### 2.1 세션 등록 (REQ-AFI-004)

1. 연결 상대 pid를 OS에서 얻는다(OQ-AFI-2). 요청 본문의 pid는 읽지 않는다.
2. §3 계보 검사(부정 조건만). 통과하지 못하면 등록하지 않는다.
3. 그 pid의 시작 fingerprint를 `ProbeProcessIdentity`로 읽는다. 읽지 못하면 등록하지 않는다.
4. session id를 발급하고 `{session_id, kind(agent|coordinator|review-session), pid, process_start, run_id, registered_at}`를 control ref에 쓴다. 살아 있는 프로세스에 이미 묶인 session id로 온 요청과, 이미 session id를 가진 프로세스의 두 번째 등록은 `session-already-bound`로 거절.
5. 증거 등록마다 1을 다시 하고, 관측 pid가 결속 pid와 같고 `ClassifyOwnerWith(결속 pid, 결속 fingerprint)`가 `live`일 때만 받는다. 아니면 `identity-mismatch`.

등록만으로는 아무 역할도 생기지 않는다. agent 역할은 상위가 attempt를 발급할 때 그 attempt에 묶인다.

### 2.2 역할 부여 (REQ-AFI-001·002·005)

1. 연결 상대 pid로 §3 계보 검사(부정 + 긍정 조건).
2. 요청자의 제어 단말을 프로세스 표에서 직접 읽는다(Linux `tty_nr`, darwin `e_tdev`, Windows 콘솔 연결 — 설계 후보). 없으면 `no-controlling-terminal`.
3. §4 운영자 확인.
4. Issue 단위 배제(REQ-AFI-005): 대상 세션이 그 Issue의 어느 attempt writer였으면 `was-writer`, 같은 Issue에서 이미 다른 역할을 가졌으면 `role-already-held`.
5. 부여 기록 `{grant_id, role, issue, run_id, target_session_id, target_pid, target_process_start, auth_binding, granted_at}`를 control ref에 CAS로 쓰고 readback 확인. CAS 실패면 부여 없음.

## §3 계보 검사 (REQ-AFI-003)

### 3.1 운영자 기준점 — 하나 (D8 결정)

- ControlService 기동 명령은 자기 부모 프로세스(운영자 셸)의 pid와 시작 fingerprint를 기준점으로 기록한다. 기준점은 **하나**이며 재기동 때만 바뀐다.
- 기동 시점에는 등록 세션이 없으므로 부정 조건은 적용되지 않는다. 등록 세션이 기동하는 경우는 잔여 위험이다(spec.md §F.3).
- 여러 기준점·tmux 서버 기준점은 범위 밖이다.

### 3.2 사슬 걷기 — 순서와 결과

요청자 pid `p`에서 시작한다. 고리마다 아래 순서로 본다. 먼저 걸리는 규칙이 결과다.

| 순서 | 검사 | 결과 |
|---|---|---|
| 1 | `p`의 fingerprint를 읽지 못함(사라짐, 권한 거부, 플랫폼 구현 없음) | `requester-ancestry-indeterminate` |
| 2 | `(p.pid, p.start)`가 등록 세션 결속과 같음 | `requester-is-session-descendant` |
| 3 | `(p.pid, p.start)`가 기준점과 같음 | 기준점 도달(통과) |
| 4 | `p`의 부모 pid가 살아 있는 프로세스를 가리키지 않음: 부모 pid 0, pid 1(Linux init/systemd, darwin launchd)로 입양됨, 또는 Windows `ParentProcessID`가 가리키는 프로세스가 없음 | `requester-outside-operator-anchor` (모든 플랫폼, 1보다 늦게·5보다 먼저) |
| 5 | 부모의 fingerprint를 읽지 못함 | `requester-ancestry-indeterminate` |
| 6 | 순서 비교가 가능한 플랫폼(§3.3)에서 부모의 시작 시각이 `p`보다 늦음 — 진짜 부모는 끝났고 pid가 재사용됨 | `requester-outside-operator-anchor` |
| 7 | 깊이 상한(제안 64)을 넘음 | `requester-ancestry-indeterminate` |
| 8 | 그 밖 | 부모로 올라가 1부터 반복 |

- 2가 3보다 먼저다. 기준점 아래에 있는 등록 세션의 자손이 기준점 도달로 통과하지 않게 한다.
- 4·6은 "사슬이 끊겼다"는 같은 뜻이므로 같은 코드를 쓴다(iteration-1 D10: Windows 고아의 코드가 다른 플랫폼과 같음).
- 등록 요청은 긍정 조건을 쓰지 않는다. 등록 요청에서 4·6에 걸리면 "등록 세션 자손 아님"으로 보고 통과한다. 고아를 만들어 등록한 세션은 등록만 될 뿐 역할을 받을 수 없다(부여는 긍정 조건이 필요).
- 부여 요청은 3에 닿아야만 통과한다. 1·2·4·5·6·7은 모두 거절이다.

### 3.3 시작 시각 원천과 PID 재사용의 한계 (iteration-1 D3, 정직한 정의)

| 플랫폼 | fingerprint 원천 | 해상도 | 순서 비교 |
|---|---|---|---|
| Linux 등(darwin·Windows 아닌 unix) | `ps -o lstart= -p <pid>` 출력 문자열 | 1초 | **하지 않음**. 문자열이 순서를 정의하지 않는다 |
| darwin | `"%d.%06d"` 초.마이크로초 | 마이크로초 | 숫자로 파싱해 비교 |
| Windows | 생성 시각 | 100나노초 단위 | 숫자로 비교 |

- `ClassifyOwnerWith`는 동등 비교만 한다. "PID 재사용 탐지"의 뜻: 기록된 fingerprint와 다른 fingerprint의 같은 pid는 `dead`로 본다.
- **Linux 한계**: 같은 초 안에 pid가 재사용되면 fingerprint 문자열이 같아 구분하지 못한다. 사슬 걷기의 6(순서 비교)도 Linux에서는 하지 않는다. 그래서 Linux에서는 끝난 부모의 pid를 같은 초에 다른 프로세스가 받으면 그 프로세스를 부모로 오인할 수 있다.
- 보완: 기준점 긍정 조건(오인한 부모의 사슬도 기준점에 닿아야 함), 상위 epoch fencing(옛 attempt의 쓰기 거절), 증거 등록 때마다 결속 재관측(REQ-AFI-004). 더 정밀한 원천 요구는 OQ-AFI-4.

## §4 운영자 확인 (REQ-AFI-002)

1. ControlService는 **자기 제어 단말**(기준점 운영자 셸의 단말)에 대상 run·Issue·역할·session id·pid·시작 fingerprint와 일회성 확인 값을 보인다.
2. 운영자는 그 단말에서 대상 session id와 확인 값을 입력한다. 답은 그 단말에서 읽은 것만 센다. 요청자 연결로 온 답은 `operator-auth-missing`(iteration-1 D1 (다)).
3. ControlService가 `auth_binding{method: interactive-tty(서버 관측), run_id, issue, role, target_session_id, target_pid, target_process_start, requester_pid, requester_process_start, issued_at, used}`를 만든다.
4. 부여 직전 검사: 대상 필드가 대상 결속과 모두 같음, 아직 쓰지 않음(binding 하나에 부여 하나), 만료 아님(대상 세션 `dead` 또는 run 종료면 만료). 하나라도 어기면 `operator-auth-missing`.
5. 기본 부여는 없다.

응답자 인증(그 단말에 실제로 누가 입력했는지)은 범위 밖이다(spec.md §F.3, OQ-AFI-1).

## §5 생산자 × 증거 종류 (REQ-AFI-006, iteration-1 D7)

| 생산자 \ 종류 | code-review | security-review | audit | qa | writer-handoff | coordinator-acceptance |
|---|---|---|---|---|---|---|
| 코드 리뷰 역할 세션 | 허용 | mismatch | mismatch | mismatch | mismatch | mismatch |
| 보안 리뷰 역할 세션 | mismatch | 허용 | mismatch | mismatch | mismatch | mismatch |
| 감사 역할 세션 | mismatch | mismatch | 허용 | mismatch | mismatch | mismatch |
| QA 역할 세션 | mismatch | mismatch | mismatch | 허용 | mismatch | mismatch |
| 현재 attempt의 agent | mismatch | mismatch | mismatch | mismatch | 허용 | mismatch |
| 등록된 Coordinator | mismatch | mismatch | mismatch | mismatch | mismatch | 허용 |
| 운영자 기록(§4 확인을 거친 운영자 증거, 역할 표시 `operator`) | 허용 | 허용 | mismatch | mismatch | mismatch | mismatch |

- 42칸: 허용 8, `role-kind-mismatch` 34. 미등록 생산자 × 6종류 = 6사례(`not-registered`). 생성 행렬 시험은 이 48사례를 모두 돌린다(AC-AFI-008).
- 허용 칸의 추가 조건(행렬과 별도 사례): 리뷰·감사·QA 생산자가 그 Issue의 writer였음 → `was-writer`. agent의 epoch가 현재가 아님 → 상위 fencing이 먼저 거절(귀속 조회 호출 0). Coordinator가 그 변경을 직접 구현함 → `was-writer`.
- 운영자 기록 리뷰는 상위 design §11.1 (b)를 유지한다(iteration-1 D9). 잠금 중에는 상위 원천이 고정값이므로 이 경로도 효력이 없다.

## §6 닫힌 사유 집합

### 6.1 요청 거절 사유 (부여·등록)

| 값 | 뜻 |
|---|---|
| `requester-is-session-descendant` | §3.2 순서 2 |
| `requester-outside-operator-anchor` | §3.2 순서 4·6(부여 요청만 거절로 쓰임) |
| `requester-ancestry-indeterminate` | §3.2 순서 1·5·7 |
| `no-controlling-terminal` | 요청자에 제어 단말 없음 |
| `operator-auth-missing` | 확인 없음·만료·재사용·불일치·다른 경로의 답 |
| `was-writer` | 그 Issue의 writer였던 세션에 리뷰·감사·QA 부여 |
| `role-already-held` | 같은 Issue에서 이미 다른 역할 |
| `session-already-bound` | §2.1 순서 4 |

### 6.2 귀속 사유 (REQ-AFI-007이 돌려주는 값)

| 값 | 뜻 | 상위 게이트(상위 design §2.1b) |
|---|---|---|
| `not-registered` | 생산자가 등록 세션도 운영자 기록도 아님 | `producer-not-allowed` |
| `identity-mismatch` | 등록 때 관측 pid·fingerprint가 결속과 다르거나 `live`가 아님 | `producer-not-allowed` |
| `grant-not-in-control-ref` | control ref에 ControlService가 쓴 유효한 부여 기록이 없음(0.1.0 `no-grant`, `grant-from-previous-start`를 대체) | `producer-not-allowed` |
| `grant-reverify-failed` | 재시작 뒤 결속·활성 재검증 실패 | `producer-not-allowed` |
| `role-kind-mismatch` | §5 mismatch 칸 | `producer-not-allowed` |
| `was-writer` | §5 추가 조건 | `producer-not-allowed` |
| `identity-spec-incomplete` | 상위 잠금의 고정값(값의 소유는 상위). 이 SPEC의 조회는 이 값을 스스로 내지 않는다 | `producer-not-allowed` |

0.1.0의 `was-worker`는 `was-writer`로 이름이 바뀌었다(명칭 agent, Issue 단위).

## §7 재시작과 재구축 (REQ-AFI-001)

1. 기동 때 control ref에서 등록 결속과 부여 기록을 읽는다.
2. 부여 기록마다 대상 세션의 결속을 재관측한다: `ClassifyOwnerWith`가 `live`이고 pid·fingerprint가 같아야 한다. 아니면 그 부여는 `grant-reverify-failed`로 표시하고 쓰지 않는다(지우지 않는다).
3. ControlService가 쓰지 않은 기록(서명·쓴 주체 불일치), 로컬 캐시에만 있는 기록은 읽지 않는다.
4. 재검증이 끝나기 전에는 귀속 조회가 모든 증거에 `grant-reverify-failed`를 돌려준다.
5. 기준점은 새 기동의 부모 셸로 다시 기록된다. 이전 기동의 기준점에 묶인 부여는 2의 재검증만 통과하면 유효하다(부여 시점의 확인은 이미 이루어졌으므로).

## §8 상위로 내보내는 조회 (REQ-AFI-007)

- 모양(제안): `Attribute(evidence_id) → attributed{kind, role, session_id, attempt, epoch} | unattributable{reason}`.
- 읽기 전용: 부여·등록·control ref 순번을 바꾸지 않는다.
- 상위 게이트 단위 시험의 대역은 이 모양을 따른다. 상위 제품 원천은 M7 전까지 잠금 고정값이다.
