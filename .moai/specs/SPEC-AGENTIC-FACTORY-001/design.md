# SPEC-AGENTIC-FACTORY-001 — 설계 (0.6.7)

이 문서는 `spec.md`의 표 요구사항(REQ-AF-016~025)이 가리키는 표와, 묶음 요구사항(REQ-AF-001~015)의 세부 계약을 담는다. 코드 구조와 함수 이름은 정하지 않는다. 다만 이미 다른 레인이 만든 이름(예: lane-18 `core.Collaboration.Authorize`, lane-17 `NormalizeRepositoryPath`)은 사실로 인용한다.

정본은 Issue #1948 본문과 댓글 11개이며 가장 새 revision이 이긴다(spec.md §A.1). 정본을 그대로 옮긴 표에는 출처 번호([0]~[10])를 붙였다. 출처가 없는 행은 이 SPEC의 결정이고, 그 결정 번호는 spec.md §C.7에 있다.

## §1 전체 구조

```
GitHub Issue (업무 필드의 권위: 제목·수용 기준·우선순위·의존성, canonical context block)
      ▲ 읽기(관찰)                    ▲ 보고 쓰기(요약·링크·안전 알림, ControlService만)
      │                               │
 관찰 보기 ─────────────── ControlService (결정적 제어·복구, LLM 없음)
                                      │ CAS 쓰기 / 읽기
                                      ▼
                     원격 control ref  refs/heads/moai-control  (기술 상태의 권위)
                       claim · attempt · epoch · effect intent/receipt
                       준비 receipt · checkpoint manifest · 역할 부여 기록
                                      │ 재구축
                                      ▼
                     로컬 캐시 (claim store 구현 = 캐시·시험 구현, 권위 아님)
                                      │
     Coordinator (내부 이름, 표시 leader) ── ControlService 명령으로만 요청
                                      │
     agent ── 자기 attempt 안에서만 쓰기 (Issue 하나, branch 하나, worktree 하나)
                                      │
     GitHub 원격 (병합·검사·보호 — 매번 직접 조회)
```

- 권위가 둘이다. 기술 상태는 control ref, 업무 필드는 Issue([3] DELTA1, [8] S51, DEC-AF-06).
- control ref는 원격에 있다: `ae3a4edcaa529bc927357b345051b96cd04605a1`(이 실행 `git ls-remote`). genesis 커밋 하나, 파일 하나(`control/schema.json`). 운영 쓰기 연결은 lane-18의 몫이며 아직 확인하지 않았다.
- 로컬 캐시는 control ref와 승인된 immutable 원본으로 재구축할 수 있어야 하고, 혼자서는 claim·부여·완료를 만들지 않는다(SYNC12).

## §2 상태 모형 (REQ-AF-016)

### 2.1 상태

| 상태 | 하위 단계 | 뜻 | 출처 |
|---|---|---|---|
| `READY` | — | 실행 대상, 활성 claim 없음 | [0]§8 |
| `BUILDING` | `SPEC` / `IMPLEMENT` / `FIX` | agent가 attempt 안에서 일함. SPEC은 E23 감사 결과로 IMPLEMENT로 감 | [0]§8, [7] S46 |
| `REVIEWING` | — | PR 인계 receipt 뒤. 책임은 배정된 agent에 남아 CI 실패 수리·리뷰 대응·병합·Issue 종료 확인까지 간다. leader(Coordinator)는 인계를 수락하고 변경 알림·실패 복구·배정 조정·통합을 지원한다(0.6.7 리더 결정. S46 문언 "종결 책임 leader"와의 차이는 spec.md §H 0.6.7) | [7] S46 E03, DEC-AF-23 |
| `MERGING` | `remote_integrated` 표시 가능 | durable effect intent와 함께 병합 요청 | [0]§8 E08, E12 |
| `DONE` | — | E13 의무 전부 충족(Issue 종료 포함) | [0]§8 E13 |
| `CANCELLED` | — | 사용자 취소 후 안전 정리 | [0]§8 E15/E16 |
| `PARKED` | (overlay) | 위 단계 위에 덮이는 정지 표시. 단계와 owner를 바꾸지 않음. 사유 하나 + 증거 + 재개 조건 | [0]§8, E20 |

- `PARKED`는 단계가 아니다. E20의 재개는 overlay만 지우고 단계와 owner를 임의로 바꾸지 않는다.
- 0.5.0의 `queued`·`claimed`·`implementing`·`verifying`·`ready`·`landing`·`done`·`blocked`·`on-hold`는 §2.3 대응표로만 남는다.

### 2.1a 닫힌 사유 집합

`PARKED` overlay, 거절 기록, 게이트 기록은 모두 이 표의 값 하나만 쓴다. 표 밖의 값은 거절한다. "쓰는 곳" 열: P = PARKED overlay, R = 동작 거절(상태 불변), G = 병합 게이트 기록.

| 값 | 뜻 | 쓰는 곳 | 기록에 필요한 근거 | 정상 재개 조건 |
|---|---|---|---|---|
| `not-installed` | 선택한 도구·브라우저·환경이 없음 | P, R | 설치 감지 명령의 종료 코드와 출력 | 같은 감지 명령 성공 |
| `connection-failed` | 설치됐지만 기능 점검 실패 | P, R | 감지 성공 + 기능 점검 출력 일부 | 같은 점검 성공 |
| `not-logged-in` | 로그인 필요 | P, R | 도구의 로그아웃 경고 원문 | 운영자 로그인 뒤 점검 성공 |
| `not-configured` | 명시 설정이 필요한 기능·키가 미설정 | P, R, G | 키 부재 또는 `false` 관측 | 운영자가 명시 설정 |
| `auth-failed` | 401 | P, R | 응답 코드와 공급자(비밀 가림) | 운영자가 자격 정상화. 자동 재시도 없음 |
| `permission-denied` | 권한 거부로 분류된 403(§4.5 G5·L3) 또는 정책·가드 거부 | P, R | 응답 문구 또는 가드 이름, 받은 권한 헤더 | 정상 경로로 권한 부여. 우회 없음 |
| `rate-limited` | §4.5 G1~G4·L1·L2 | P, R | 응답 코드, 헤더가 준 재개 시각(없으면 `UNKNOWN`) | 그 시각 이후 한 번 재개 |
| `unclassified-403` | 신호 없는 403(§4.5 G6). 권한이라고도 rate limit이라고도 단정하지 않음. `forge.KindUnclassified`와 같은 문자열 | P, R | 응답 코드와 헤더 부재 기록 | 다음 정상 관측 또는 운영자 판단. 재시도 0 |
| `capacity-short` | 용량 부족(디스크, slot, 유료 한도, checkpoint 상한) | P, R | 측정값 | 용량 확보. 추가 지출은 운영자 승인 |
| `budget-exhausted` | 감사 라운드·API 예산 소진 | P, G | 라운드·예산 기록과 상한 | 운영자 승인 또는 예산 창 갱신 |
| `check-not-passed` | 필수 검사가 skipped·누락·실패·취소·다른 HEAD | G | 검사 이름, 결과, 결과가 가리킨 HEAD | 최신 HEAD에서 다시 실행해 성공 |
| `debt-present` | 감사 판정 `PASS-WITH-DEBT` | G | 판정 파일과 HEAD | 무부채 PASS |
| `verdict-not-pass` | 증거가 없거나, 판정값이 `FAIL`·`BLOCKED`·읽을 수 없음, 또는 다른 HEAD의 판정 | G | 증거 경로(없으면 "없음"), 판정 HEAD, 판정값 원문 | 같은 HEAD의 새 판정 `PASS` |
| `producer-not-allowed` | 귀속 조회 결과가 그 증거 종류에 필요한 역할의 `attributed`가 아님(§2.1b) | G | 증거 경로와 귀속 조회 결과 원문 | 맞는 역할로 귀속되는 생산자가 같은 HEAD로 새 증거 등록. 잠금 중에는 M7 해제가 먼저 |
| `ownership-unverified` | 소유권 확인 실패·불가 | R | 실패한 확인 단계 | 확인 성공 |
| `store-unavailable` | control ref 읽기 실패 또는 CAS 실패 | P, R | 읽기·CAS 오류 원문 | control ref 읽기·CAS 성공. 로컬 캐시만으로 진행하지 않음 |
| `user-control` | 사용자가 제어권을 가져감 | P | 도구 표식 | 사용자가 돌려줌 |
| `operator-ack-missing` | 운영자 keep-set 동작 또는 기존 writer의 안전 정지 ACK가 없음. **역할 부여 부재에는 쓰지 않는다**(그것은 G의 `producer-not-allowed`) | P, R | 확인 대기 기록 | 해당 ACK 수신 |
| `operator-hold` | 운영자(또는 운영자가 승인한 시험 하네스)가 멈춤. **이 값을 PARKED 사유로 둘지는 운영자 결정 대기 OD-12**(spec.md §G.1). 정본 S46에는 이에 대응하는 PARKED 사유가 없다 | P | 요청자와 사유 원문 | 운영자 재개 기록 |
| `dependency` | 의존 Issue 미완료(E02) | P | 원격 의존 증거 | 의존 DONE |
| `stop_unknown` | heartbeat·stop 불명, quiescence 부족(E04) | P | incarnation 확인 기록 | quiescence 확인 또는 안전 정지 기록 |
| `unknown-effect` | 외부 효과 결과 불명(E11, E16) | P | effect intent와 sender 확인 기록 | 원격 사실 확정 |
| `pending-effect` | 병합 뒤 sync·close 실패(E14) | P | merge 사실과 실패한 의무 | 실패한 의무만 다시 수행해 성공 |
| `contract-change` | 계약·설정·정책 변경(E18, E24) | P | proposed/adopted 차이 | 새 attempt에 승인 revision 적용 |
| `start_unknown` | supervisor 기동 불명(E26) | P | launch key 생존 불명 기록 | 기동 사실 확정. 새 launch 금지 |
| `not-issue-author` | 자기 Issue 전용 동작을 남의 Issue에 요청(본문 편집·종료, 협업 허락 없음) | R | 인증 사용자 숫자 id, 작성자 숫자 id | 해당 Issue에 맞는 경로(댓글, 협업 옵션) |
| `identity-unknown` | 인증 사용자 또는 작성자의 숫자 id를 얻지 못함 | R | 실패한 조회 | 두 id 모두 관측 |
| `repository-not-allowed` | 저장소가 허용 목록에 없음, 또는 요청한 저장소와 응답한 저장소가 다름, 또는 301 등 어떤 redirect든 | R | 요청 저장소, 응답 저장소·상태 코드 | 운영자가 허용 목록을 바꿈. redirect는 허용으로 보지 않음 |
| `permission-not-granted` | 협업은 켜졌지만 그 동작의 권한 키가 꺼짐, 또는 자기 Issue의 레이블·담당자 변경(권한 키가 `true`여도) | R | 동작 이름과 키 값 | 남의 Issue: 운영자가 해당 키를 켬. 자기 Issue: 없음(정의 안 됨, §H) |
| `owner-unconfirmed` | 남의 Issue 인수에 기존 owner가 없거나 인계가 확인되지 않음 | R | owner 조회 결과 | 기존 owner의 인계 확인 |
| `foreign-closing-keyword` | PR 본문이나 커밋이 닫아서는 안 되는 Issue를 가리키는 종결 키워드를 가짐 | R | 키워드, 대상 Issue, 위치 | 키워드 제거 |
| `op-not-defined` | 정의되지 않은 Issue 동작(자기 Issue의 제목 편집·마일스톤 변경·다시 열기) | R | 동작 이름 | 없음(정의 안 됨, §H) |

**PARKED 전용 사유의 철자는 정본을 따른다(0.6.2).** 정본 S46(댓글 6489610367) 본문 48·50·57·60·70·72행은 `dependency`, `stop_unknown`, `unknown-effect`, `pending-effect`, `contract-change`, `start_unknown`으로 적는다. `stop_unknown`·`start_unknown` 두 값만 밑줄이고 나머지 넷은 하이픈이다. 0.6.1까지 이 표는 두 값을 하이픈으로 적었고 0.6.2에서 정본 철자로 바꿨다. 같은 두 철자는 정본 댓글 6487471042(167·189행)에도 있다.

`unclassified-403`과 표 끝 7행, 모두 8개 문자열은 lane-18이 보고하고 리더가 확인한 `core/state.go`의 임시 상수다(spec.md §B.2, DEC-AF-16).

**바이트 대조 (0.6.1, 이 SPEC 트리 HEAD `89f4e812a3`)**. 대조 대상은 lane-18 worktree `af-core-claim-runner`의 체크아웃 파일(읽기만 함, 그 트리에 git 실행 없음, 브랜치 ref 파일 값 `fa781ef0e9`)과, 이 트리에서 `git show f9f44e5ecc:<경로>`로 읽은 같은 파일이다. 두 파일은 두 커밋 사이에 바뀌지 않았다(`git diff --stat f9f44e5ecc fa781ef0e9 -- <두 파일>` 출력 없음). 명령은 각 원천에 `grep -n -o -F -e '"<값>"' …`(8개 값을 한 번에), SPEC 쪽은 같은 명령에 따옴표 대신 백틱을 썼다. "따옴표 붙은 값의 개수"는 Go 문자열 리터럴로 정확히 같은 바이트가 나온 횟수다.

| 값 | `state.go` (`f9f44e5ecc`와 체크아웃 같음) | `authz.go` (같음) | SPEC 문자열(이 표, design.md) | 판정 |
|---|---|---|---|---|
| `unclassified-403` | 1개, 75행 `ReasonUnclassified` | 0개 | 64행 | 일치 |
| `not-issue-author` | 1개, 78행 `ReasonNotIssueAuthor` | 1개, 91행 `DenyNotIssueAuthor`(판정 값) | 82행 | 일치 |
| `identity-unknown` | 1개, 79행 `ReasonIdentityUnknown` | 1개, 92행 `DenyIdentityUnknown` | 83행 | 일치 |
| `repository-not-allowed` | 1개, 80행 `ReasonRepoNotAllowed` | 1개, 93행 `DenyRepoNotAllowed` | 84행 | 일치 |
| `permission-not-granted` | 1개, 81행 `ReasonPermissionNotGranted` | 1개, 94행 `DenyPermissionNotGranted` | 85행 | 일치 |
| `owner-unconfirmed` | 1개, 82행 `ReasonOwnerUnconfirmed` | 1개, 95행 `DenyOwnerUnconfirmed` | 86행 | 일치 |
| `foreign-closing-keyword` | 1개, 83행 `ReasonForeignClosing` | 0개 | 87행 | 일치 |
| `op-not-defined` | 1개, 86행 `ReasonOpNotDefined` | 0개(따옴표 문자열 없음. 208행은 상수 `ReasonOpNotDefined`를 반환할 뿐이다) | 88행 | 일치 |

- 8개 모두 SPEC 문자열과 Go 상수가 바이트 단위로 같다. 그래서 SPEC 문자열을 고치지 않았다.
- `op-not-defined`의 위치(리더 정정 반영): `authz.go:208`은 상수 `ReasonOpNotDefined`를 **반환할 뿐이다**. 따옴표 붙은 문자열 `"op-not-defined"`는 `state.go:86`에만 있다.
- `authz.go:96`의 `"unknown-operation"`과 `authz.go:90`의 `"allow"`는 `AuthzVerdict` 값이며 `BlockReason`이 아니다. `BlockReason` 상수 26개는 모두 이 표 안에 있다. 관찰 기록은 spec.md §H(공백).

### 2.1b 귀속 조회 결과 → 게이트 사유 (단일 표, REQ-AF-007·011)

모든 AC가 이 표 하나를 쓴다. 조회 결과 하나는 게이트 사유 하나로만 간다.

| 귀속 조회 결과(하위 SPEC REQ-AFI-007) | 게이트 기록 |
|---|---|
| 증거가 등록되지 않음 | `verdict-not-pass` |
| `attributed`, 역할이 그 종류에 필요한 역할과 같음, 판정 HEAD = 후보 HEAD, 판정값 `PASS` | 통과로 셈 |
| `attributed`, 역할 같음, 판정값 `FAIL`·`BLOCKED`·읽을 수 없음 또는 판정 HEAD ≠ 후보 HEAD | `verdict-not-pass` |
| `attributed`, 역할 같음, 판정값 `PASS-WITH-DEBT` | `debt-present` |
| `attributed`, 역할이 다름 | `producer-not-allowed` |
| `unattributable{identity-spec-incomplete}`(잠금) | `producer-not-allowed` |
| `unattributable{not-registered}` / `{identity-mismatch}` / `{grant-not-in-control-ref}` / `{grant-reverify-failed}` / `{role-kind-mismatch}` / `{was-writer}` | `producer-not-allowed` |
| 귀속 원천 오류·미구성 | `producer-not-allowed` |

- 0.5.0 하위 SPEC의 `no-grant`는 하위 0.2.0에서 `grant-not-in-control-ref`로 이름이 바뀌었고 게이트에서는 `producer-not-allowed`다(iteration-5 D4 해소).
- 귀속 판정 순서: 판정값·HEAD 검사보다 귀속 검사가 먼저다. 귀속이 실패한 증거는 판정값이 `PASS`여도 `producer-not-allowed` 하나로만 기록한다.
- writer 인계(`writer-handoff`)는 두 단계다. 먼저 fencing(REQ-AF-001): 옛 epoch의 등록은 귀속 조회에 닿기 전에 epoch 거절(종료 코드 15)이다. fencing을 통과한 등록만 귀속 조회로 간다.

### 2.1c 대기 기록 필수 필드 (REQ-AF-016, DEC-AF-23, 0.6.7)

출처: 리더 결정(리더 세션 경유). 대기 기록은 `PARKED` overlay와, Issue나 agent가 무언가를 기다린다고 남기는 그 밖의 모든 기록이다(예: CI 결과 대기, 리뷰 대기, 인계 대기). 아래 여섯 필드가 모두 있어야 하며, 하나라도 없으면 그 기록은 **무효**다. 상태 엔진은 무효 기록을 거절하고 이전 상태를 그대로 둔다. 거절에 쓸 새 사유 값은 만들지 않았다(§2.1a 닫힌 집합 불변).

| 필드 | 뜻 |
|---|---|
| 단일 책임자 | 이 대기를 풀 책임이 있는 주체 하나(agent·leader·ControlService·운영자 중 하나, 이름까지) |
| 다음 자동 동작 | 아무도 개입하지 않을 때 다음에 자동으로 일어나는 일 |
| 재확인 시각 | 대기 상태를 다시 확인할 시각(마감) |
| 진단 | 지금 무엇을 왜 기다리는지에 대한 관측 근거 |
| 재시도 상한 | 자동 동작을 되풀이할 최대 횟수 |
| 상승 대상 | 상한에 닿거나 재확인 시각을 넘겼을 때 넘길 대상 |

- 값(시각, 횟수, 상승 대상의 기본값)은 이 SPEC이 정하지 않는다. 필드의 존재만 정한다(spec.md §H 0.6.7).
- "다음 자동 동작"의 형태(리더 결정, 리더 세션 경유, DEC-AF-27): 상대 시간(지금 + N)으로 시각을 밝힌 한 번의 자기 페이스 깨어남이며, 남은 일이 없으면 다시 걸지 않는다. 고정 간격으로 되풀이되는 작업이면 그 기록은 무효다(REQ-AF-016). 반복 cron은 쓰지 않는다 — 상대 시간과 반복 cron 배제는 사용자 결정(user decision via leader AskUserQuestion, transcribed by the leader in that file; original not read by this author — `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-lane-wake-20261030.md`, 0.6.7 서른다섯 번째 후속)이고, "한 번에 하나, 남은 일이 없으면 걸지 않음"과 "절대 시각 한 번짜리 금지"는 그 파일의 리더 잠정 규칙이다. N 값은 여전히 이 SPEC이 정하지 않는다.
- lane-18 구현 관찰(0.6.7, 커밋 `deb7551aac`, 리더 전달·이 작성자는 다시 읽지 않음): 새 `PARKED` 상태로 들어가는 쓰기는 Wait 기록을 요구한다. 그 Wait 기록이 위 여섯 필드를 모두 갖는지는 확인하지 않았다(알려진 것은 `MaxRetries`·`escalate_to` 요구뿐, §6.1). spec.md §H 0.6.7.
- `PARKED` overlay의 기존 세 요소(사유, 근거, 정상 재개 조건)는 그대로이며, 여섯 필드가 함께 있어야 한다. "진단"이 근거를, "다음 자동 동작"·"재확인 시각"이 재개 조건을 대신하지 않는다.

### 2.2 전이표 E01~E27

정본 [0]§8 표에 S46([7]) 개정을 반영했다. "쓰는 역할"은 상태를 실제로 쓰는 쪽이며 모두 ControlService다. 괄호는 요청자다.

| 간선 | 출발 | 사건 | guard(필요한 증거) | 도착 | 쓰는 역할(요청자) |
|---|---|---|---|---|---|
| E01 | READY | claim | 채택 계약·의존·scope·예산, 활성 owner 없음, 준비 receipt와 READY ACK(§4.2) | BUILDING/SPEC. epoch·attempt·예약을 원자적으로 확정 | ControlService(dispatcher) |
| E02 | READY | 의존 미완료 | 원격 의존 증거 | READY + PARKED `dependency` | ControlService |
| E03 | BUILDING | PR 인계 | PR 원격 HEAD/base, 계약, SPEC, task 결과, 검사 상태, 원 writer, leader 수락 receipt, quiescence, 보존 확인 — 새 인계 쓰기에만 모든 필드 요구, 원 writer 출처는 claim 획득 때 Issue 작성자 검사가 충족(§4.8 PRHandoff 대응.  lane-18 필드 이름: `PR URL`·`base`·`adopted SPEC`·`task result`·`check status`·`handoff id`·`LeaderAccepted`·`PR number`·`HEAD`·`branch`·`worktree`·`WriterStopped`·`BoundaryHead` — §4.8) | REVIEWING. 책임은 agent(Holder)에 남음. 새 Issue를 받을지는 acquire 검사가 판정 — 책임과 실행 자리는 따로 있는 두 자원이 아니며 효과로 읽는다(0.6.7 리더 판정, 리더 세션 경유). S46 문언: "종결 책임 leader로, agent claim·slot 반환" — spec.md §H 0.6.7 | ControlService(agent, leader) |
| E04 | BUILDING | heartbeat·stop 불명 | incarnation 확인, quiescence 부족 | 같은 단계 + PARKED `stop_unknown` | ControlService |
| E05 | BUILDING/REVIEWING | 안전 인계 | process·tool·원격 quiescence, 잔여 send·write 정리, 원본 보존 | 새 epoch·attempt로 BUILDING, 보존 후 재검증 | ControlService |
| E06 | REVIEWING | 수리 결정 | finding·예산, 원 Issue·branch의 배타 owner, scope 재예약, 기존 writer 종료 | BUILDING/FIX, 새 attempt. 그 Issue의 책임 agent에 배차(0.6.7 리더 결정. S46 문언은 "가용 agent에 배차" — spec.md §H 0.6.7), 그 agent의 다른 Issue를 덮어쓰지 않음. lane-18 구현 관찰: 수리는 원래 책임 세션에만 허용(`67defdeb1f`), 수리 attempt 이름은 `repair-<n>`(`06f8ccc3c7`) — 리더 전달, 이 작성자는 다시 읽지 않음. 세션 기준과 이 칸의 agent 기준의 차이는 spec.md §H 0.6.7 | ControlService(leader) |
| E07 | REVIEWING | advisory 무응답 | 정책 fallback 허용 | 단계 유지, leader 판단 receipt. 감사 게이트 불변 | ControlService |
| E08 | REVIEWING | 병합 허용 | §11 병합 게이트 전부 | MERGING + durable effect intent | ControlService |
| E09 | REVIEWING | HEAD·base·환경 변경 | provider 사실·manifest 차이 | REVIEWING + 관련 증거 무효. 변경이 필요하면 E06 | ControlService |
| E10 | MERGING | HEAD·보호 변경 | 원격 미병합, sender 잔여 권한·queue 효과 정리 | REVIEWING(불명이면 MERGING + PARKED `unknown-effect`). 이어 수리가 필요하면 E06 | ControlService |
| E11 | MERGING | send 응답 불명 | effect intent와 sender 확인 | MERGING + PARKED `unknown-effect`. 사실 조회만 | ControlService |
| E12 | MERGING | 원격 병합됨 | PR HEAD ↔ 병합 커밋 ↔ 승인 target 포함 | MERGING `remote_integrated`, 후속 의무 등록 | ControlService |
| E13 | MERGING | 의무 모두 충족 | 수용 증거, 병합 커밋, 승인된 로컬 동기화(§4 `local_sync`), 배포 기록 또는 `deploy: N/A`+사유, **Issue 종료**, 정리 | DONE. 그 Issue·epoch 자원만 해제 | ControlService |
| E14 | MERGING | sync·close 실패 | 병합 사실 보존 | MERGING + PARKED `pending-effect`. 무단 revert 없음 | ControlService |
| E15 | READY/BUILDING/REVIEWING | 사용자 취소 | 새 허가 즉시 차단, 모든 writer·effect 안전 정리 | CANCELLED(정리 불명이면 현 단계 + PARKED) | ControlService(사용자) |
| E16 | MERGING | 취소 요청 | sender 잔여 send 제거 + queue·원격 효과 확인 | CANCELLED 또는 MERGING + PARKED `unknown-effect` | ControlService(사용자) |
| E17 | MERGING/DONE | 취소 뒤 병합 관측 | provider 병합 커밋 확인 | 통합 사실 보존, 후속 결정 별도. 자동 revert 없음 | ControlService |
| E18 | 비종결 | 계약·설정·정책 변경. 이어 승인된 revision 채택(adopt) | proposed/adopted 차이, revocation 범위. 채택 간선은 그 revision이 승인되었다는 증거가 있어야만 일어남 | 현 단계 + PARKED `contract-change`. 채택: BUILDING/FIX + PARKED `contract-change`에서 승인된 revision을 채택하면 BUILDING/SPEC, owner·receipt·epoch 유지(0.6.7 리더 결정, 리더 세션 경유. 정본 대조 안 함. 구현 lane-18 `core/transition.go` 137~138행, runner `AdoptRevision` — 리더 전달, 이 작성자는 다시 읽지 않음. spec.md §G.1 OD-16) | ControlService |
| E19 | 비종결 | quota·권한·예산 차단 | §4.5 분류 결과 | 현 단계 + PARKED(분류 값), 재개 조건 기록 | ControlService |
| E20 | PARKED | 재개 | 그 사유의 재개 조건 전부 | overlay 해제, 단계·owner 불변 | ControlService |
| E21 | DONE/CANCELLED | 늦은 명령·결과 | id·epoch·receipt 확인 | no-op 또는 거절 | ControlService |
| E22 | DONE/CANCELLED | 새 업무 의도 | 새 사용자 의도·계약 | 별도 Issue 검토. 기존 DONE claim 재사용 금지 | ControlService |
| E23 | BUILDING/SPEC | SPEC 감사 결과 | 같은 SPEC revision, 선택된 감사자, 요구 범위 | PASS → IMPLEMENT, FAIL → SPEC 수정, INCONCLUSIVE → PARKED | ControlService |
| E24 | BUILDING/FIX | SPEC 변경 제안 | 실제 계약 영향 | PARKED `contract-change`, E18 뒤 SPEC 단계 | ControlService |
| E25 | BUILDING | scope 확대 충돌 | 전체 확대 원자 실패 | checkpoint·quiescence·기존 scope 반환 후 전체 재예약. Issue 독점 유지 | ControlService |
| E26 | 비종결 | supervisor 기동 불명 | launch key 생존 불명 | PARKED `start_unknown`. 새 launch 금지 | ControlService |
| E27 | 아무 상태 | 관찰 사실 수집 | 서명·권한·생산자 검증 | receipt 기록. 사실만으로 허가를 주지 않음 | ControlService |

- 표에 없는 간선과 요청자가 맞지 않는 요청은 거절하고 상태를 바꾸지 않는다.
- E13의 순서: Issue 종료가 DONE 기록보다 먼저다(DEC-AF-07, 시험 M34). 다음 dispatch는 DONE 뒤.
- 0.5.0에 없던 간선 `landing → verifying`은 E10이 메운다. 수리는 E10 → E06이다(DEC-AF-07). 옛 이름이 어느 정본 단계에 놓이는지는 이름이 아니라 증거로 정한다(§2.3).

### 2.3 옛 상태 → 정본 값 대응 (0.4.0·0.5.0 → 정본, 증거 기준)

출처: 'leader decision, relayed via the leader session'(0.6.3 정정). 대응은 **옛 상태 이름이 아니라 증거로** 정하며, 정본 E03(S46 49행, PR 인계 → REVIEWING)을 따른다. 오른쪽 값은 정본 값(§2.1의 단계·하위 단계, §2.1a 사유)만 쓴다.

**증거 규칙** — 옛 상태 이름이 `verifying`·`ready`·`landing` 가운데 무엇이든 아래 순서로 읽는다.

| 관측한 증거 | 정본 값 |
|---|---|
| writer 인계 receipt **와** leader의 인계 수락이 **둘 다** 있음 | `REVIEWING` |
| 병합 의무가 진행 중(`merged=true`) | `MERGING` |
| 그 밖의 모든 경우 — 인계 증거 둘 가운데 하나라도 없는 `landing`, 그리고 `verifying`·`ready` 포함 | `BUILDING/IMPLEMENT` + 검사 증거. 새 하위 단계는 만들지 않는다 |

**이름으로 정하는 나머지 행**

| 옛 상태 | 정본 값 | 출처 |
|---|---|---|
| `queued` | `READY` | 0.6.0 |
| `claimed` | (E01 안의 원자 단계, 별도 상태 없음) | 0.6.0 |
| `implementing` | `BUILDING/IMPLEMENT` 또는 `BUILDING/FIX` | 0.6.0 |
| `done` | `DONE` | 0.6.0 |
| `blocked` | 현 단계 + `PARKED`(사유 하나, §2.1a) | 리더 결정(0.6.2, 0.6.3에서 그대로) |
| `on-hold` | 정하지 않음. `operator-hold`를 PARKED 사유로 둘지와 함께 운영자 결정 대기 OD-12(spec.md §G.1) | — |

- 판 이력: 0.6.1까지는 이름 기준(`verifying`·`ready` → `REVIEWING`, `landing` → `MERGING`)이었다. 0.6.2도 이름 기준(`verifying`·`ready` → `BUILDING/IMPLEMENT`, `landing` → receipt에 따라 `REVIEWING`/`MERGING`)이었다. 0.6.3에서 위 증거 규칙으로 바꿨다.
- 이 규칙으로 0.6.2의 두 공백이 닫힌다: 옛 `ready`·`verifying` 정의와 E03의 어긋남, 그리고 정해지지 않았던 `landing` 경우(spec.md §H (i)·(j)).
- 인계 증거 둘과 `merged=true`가 함께 관측될 때 어느 행이 먼저인지는 리더 결정 문언에 없다(spec.md §H (k)).
- 옛 상태 키의 이전은 세 단계로 나란히 진행된다고 기록한다: 새 키 추가 → 전환(옛 값을 새 값에서 파생) → 옛 키 제거. 출처: 'leader decision / lane-18 implementation in progress'. 관찰로만 기록하며 이 SPEC은 코드에 무엇도 요구하지 않는다(spec.md §B.2).

## §3 전달 상태 네 단계 (REQ-AF-005)

| 상태 | 뜻 | 증거 원천 | 증거가 아닌 것 |
|---|---|---|---|
| `sent` | 명령을 보냄 | 보낸 쪽의 전송 결과 기록 | — |
| `received` | 받는 쪽이 명령을 꺼내 읽음 | 받는 쪽의 수신 확인 행(명령 id 포함) | 전송 성공, idle 알림 |
| `executed` | 받는 쪽이 실행을 시작함 | attempt 시작 행 + 작업 트리 HEAD 관측 | 수신 확인, 답장 |
| `completed` | 실행이 끝났고 결과가 control ref 또는 증거 위치에 있음 | ControlService가 읽은 결과 증거(전체 HEAD SHA) | 답장 "완료", idle 알림, 전송 성공 |

ACK가 없으면 다시 보내지 않고 받는 쪽 상태를 먼저 조회한다. 조회 결과가 `received` 이상이면 재전송하지 않는다. 조회 불가면 PARKED `unknown-effect`.

## §4 record와 필드 (REQ-AF-002·017·019)

### 4.0 권위 배치

| 필드·record | 위치(권위) | 쓰기 권한 |
|---|---|---|
| `task_name`, `acceptance_criteria`, `priority`, `depends_on` | Issue 본문(업무 권위) | 사용자. AgenticFactory는 자기 Issue에서만 §4.1 규칙으로 |
| claim, attempt, epoch, owner, writer | control ref | ControlService |
| 상태, 하위 단계, PARKED 사유·재개 조건 | control ref | ControlService |
| effect intent·receipt(§4.8) | control ref | ControlService |
| 준비 receipt(§4.2) | control ref | ControlService(준비 subagent 결과를 받아) |
| checkpoint manifest(§4.6) | control ref | ControlService(lane-18) |
| 역할 부여 기록 | control ref | ControlService(하위 SPEC) |
| 증거 등록(경로, 판정 HEAD, 판정값, 내용 digest, 귀속 조회 결과) | control ref | ControlService |
| `local_sync` | control ref(요약만 Issue) | ControlService. 동기화 전 수정·스테이징·미추적 경로마다 `{path, before:{content_sha256, index_entry, mode}, after:{…}}`. 하나라도 다르면 E13 금지 |
| `deploy` | control ref | ControlService. 배포 기록 경로 + HEAD, 또는 `N/A` + 사유 |
| Issue canonical context block(§4.3 장부 색인) | Issue 본문 블록 | ControlService(보고 쓰기) |
| 로컬 claim store | 로컬 캐시 | 재구축 전용. 권위 아님 |

관찰은 읽기만 한다. 읽기 권한은 쓰기 권한을 주지 않는다.

### 4.1 Issue 쓰기 판정 (REQ-AF-003, DEC-AF-17~19)

판정 순서(모든 Issue 쓰기에 앞서 한 번):

1. 인증된 사용자의 숫자 id와 Issue 작성자의 숫자 id를 얻는다. 하나라도 못 얻으면 `identity-unknown`. 로그인 이름은 비교하지 않는다.
2. 요청 저장소와 응답 저장소를 비교한다. 다르거나 응답이 redirect(301 포함)면 `repository-not-allowed`.
3. 동작이 정의되었는지 본다. 제목 편집·마일스톤 변경·다시 열기는 `op-not-defined`(자기 Issue든 남의 Issue든).
4. 자기 Issue(두 숫자 id가 같음):
   - 본문 편집·댓글·종료(E13/E14): 허용.
   - 레이블·담당자 변경: `permission-not-granted`. 어떤 `permissions.*` 키 값과도 무관.
5. 남의 Issue:
   - 관련 댓글: 허용.
   - claim(인수): `collaboration.enabled = true` 그리고 저장소가 허용 목록에 있음 그리고 기존 owner의 인계 확인. 앞의 둘 중 하나라도 아니면 `repository-not-allowed`(목록 밖) 또는 `not-issue-author`(옵션 꺼짐). 인계 미확인이면 `owner-unconfirmed`.
   - 본문 편집·종료·레이블·담당자: claim 조건 + 그 동작의 권한 키가 `true`. 옵션이 꺼져 있으면 `not-issue-author`, 목록 밖이면 `repository-not-allowed`, 키가 꺼져 있으면 `permission-not-granted`.
6. PR 본문이나 커밋 메시지가 닫아서는 안 되는 Issue를 종결 키워드로 가리키면 그 PR·커밋 effect를 보내기 전에 `foreign-closing-keyword`.
7. 같은 업무의 기존 Issue가 있으면(남의 작업 가로채기용 자기 Issue 중복 발행) 새 Issue를 만들지 않는다.

종결 키워드: GitHub 문서의 아홉 낱말 `close`, `closes`, `closed`, `fix`, `fixes`, `fixed`, `resolve`, `resolves`, `resolved`. 대소문자 혼합과 URL 형태(`<owner>/<repo>#N`, 전체 Issue URL)는 문서에 없지만 안전 쪽으로 포함한다(lane-18 보고, uncollated).

저장소 이름 비교는 공급자별이다: GitHub은 ASCII 소문자 비교(문서 "The name is not case sensitive.", uncollated), GitLab과 알 수 없는 공급자는 정확히 일치. 정규화는 lane-17 `NormalizeRepositoryPath(provider, s)`가 한다.

**읽기 코드와 판정 코드의 책임(두 열)**

| 설정 읽기(lane-17, `internal/config/agentic_factory.go`) | 판정(lane-18, `core.Collaboration.Authorize`) |
|---|---|
| 키 `collaboration.enabled`(bool, 기본 false), `collaboration.allowed_repositories`(문자열 목록, 기본 빈 목록), `collaboration.permissions.{edit_body,close,labels,assignees}`(bool, 기본 false)를 읽는다 | 위 1~7 판정 규칙을 한 곳에서 수행한다 |
| 허용 목록 항목을 검증한다: `/`로 나눈 비지 않은 조각 2개 이상, 와일드카드·`://`·공백·빈 항목·앞뒤 `/`·문자열 아닌 값 거절 | 숫자 id 비교, 저장소 응답 대조, redirect 거절, 종결 키워드 검사 |
| 공급자별 정규화(`NormalizeRepositoryPath`) | 사유 문자열 선택(§2.1a 마지막 8행) |
| 판정하지 않는다. 판정 도우미를 두지 않는다(`ForeignIssueAllows` 같은 이름의 API를 만들지 않는다) | 설정 값을 직접 읽지 않고 주입받는다(`Provider`, `Enabled`, `AllowedRepositories`, `EditBody`, `Close`, `Labels`, `Assignees`) |

연결(읽은 값 → `core.Collaboration` 주입)은 아직 없다. 담당은 lane-18이며 lane-17 읽기 코드를 쓴다(spec.md §G.1, §H).

### 4.2 접수·준비 상태표 (REQ-AF-012·018, [8] S51)

| 행 | 상태 | 확인 조건과 다음 행동 |
|---|---|---|
| 1 | `ACCEPTED` | 요청 intent·승인 범위·request_id·정규 payload hash를 control ref에 기록한 뒤에만 수락 응답. 제품 FSM의 BUILDING 아님 |
| 2 | `ISSUE_CONFIRMED` | 실제 Issue id와 요청 결속 확인(재조회). 응답 유실 뒤 생성 여부 불명이면 재발행하지 않음 |
| 3 | `BRANCH_REGISTERED` | 원격 필요 여부·ref·base OID·기존 branch 대조. 빈 커밋 강제·기존 branch 덮어쓰기 없음 |
| 4 | `AGENT_RESERVED` | 한 Issue 한 agent와 slot·WIP 예약. 아직 개발 실행 금지 |
| 5 | `WT_PREPARED` | 실행 host의 로컬 worktree·dirty·branch·설정 검증, 준비 writer의 quiescence |
| 6 | `READY_ACKED` | 준비 receipt를 예약 agent가 명시 수락. E01 guard와 launch 허가를 충족할 때만 BUILDING |
| 7 | `BLOCKED` 또는 `EFFECT_UNKNOWN` | 실패 단계·잔여 효과·재개 조건 보존. 자동으로 새 Issue·branch·agent를 만들지 않음 |

- 7행이다. 이름으로는 8개(`BLOCKED`와 `EFFECT_UNKNOWN`이 한 행). lane-18이 이 7행을 구현했다고 보고했다(커밋 `da7f3a4290`, 미검증). lane-20의 표시 이름 6개는 이 표에 맞추는 중인 표시 이름이며, 코드 단계 수를 뜻하지 않는다.
- 준비 receipt 필드: request_id, Issue id, branch/base/remote OID, host와 worktree 식별, 준비 grant와 종료 증거, 예약 agent·session, 설정 revision, 점검 결과, READY ACK.
- 실행 규칙: 큐는 동시 준비 수 `intake.max_concurrent`(1 이상. 미설정일 때의 동작은 운영자 결정 대기 OD-9), 단계 기한 `intake.stage_timeout`, 우선순위 후 접수 시각 순서, 요청당 준비자 하나, 기한이 지나면 예약 해제. 재시작 때는 기록된 요청만 재조정한다. 포화 때 수락 후 대기할지 거절할지는 OQ-21.
- 금지: 새 원장 저장소, 무한 하위 작업, 상시 레인.

### 4.3 Issue 맥락 장부 (REQ-AF-019, [8] S51)

| 필드 | 기록 계약 |
|---|---|
| `artifact_id`, `kind`, `title` | Issue 안의 자료 식별. 종류: SPEC·계획·설계 결정·진척·실패/재현·감사·시험·PR 인계·완료 보고 |
| `path`, `link` | 실제 저장소 상대 경로와 **전체 커밋 SHA 고정 URL**. 존재·추적·권한 확인 뒤 기록. 로컬 file URL·비공개 링크만 남기지 않음 |
| `artifact_commit`, `artifact_sha256` | 생성 커밋과 파일 바이트 hash. hash만으로 내용 정확성 PASS를 주장하지 않음 |
| `target_head`, `spec_revision` | 무엇을 설명·검증하는지 결속(생성 커밋과 별개 필드) |
| `verification_state` | `PENDING` / `NOT_RUN` / 실제 결과 / `STALE` / `LINK_BLOCKED` / `UNKNOWN` |
| `owner`, `verified_at` | 작성·검증 담당과 확인 시각(RFC 3339 UTC) |
| `supersedes`, `superseded_by` | 이전 revision 보존과 대체 자료 연결 |
| `accessibility` | 아래 분류 |
| `exposure_review` | `public-ok` / `restricted`. `public-ok`가 아니면 링크를 게시하지 않고 `LINK_BLOCKED` |

**STALE 계기**(하나라도 참이면 `STALE`, 원 판정값은 `stale_of`로 보존, 항목 삭제 금지): ① `target_head`가 현재 후보 HEAD와 다름 ② 링크한 blob의 sha256이 기록값과 다름 ③ `spec_revision`이 현재와 다름 ④ 원격 조회에서 링크 커밋이 사라짐. 원격 조회 자체가 실패하면 `STALE`이 아니라 `UNKNOWN`이며 PASS로 올리지 않는다.

**접근성 분류**

| 값 | 결정 조건 |
|---|---|
| `accessible` | 원격 커밋 조회 200, 그 커밋에 경로 존재, 바이트 hash 일치, 저장소 공개 또는 수신 agent 읽기 권한 확인. 증거 단계 `REMOTE_VERIFIED` + `AVAILABLE`과 같다 |
| `inaccessible-local` | 경로만 있고 커밋·링크 없음, 미추적, ignore 대상, 또는 `file://`·절대 경로. 원본이 `.moai/reports`에만 있는 상태 |
| `inaccessible-unpushed` | 커밋이 로컬에만 있거나 원격 도달성이 확인되지 않음(단계 `COMMITTED`, 또는 `PUSHED`인데 원격 조회 미확인). 오프라인 판정은 이 값까지만 내고 `accessible`은 내지 않음 |
| `stale` | 표시 축으로 `verification_state = STALE`(접근성과 별개 축이지만 관찰 보기에는 같이 보임) |

### 4.4 정제 범주 C1~C5 (REQ-AF-013, DEC-AF-12)

| 범주 | 대상 | 처리 |
|---|---|---|
| C1 자격 증명 | 토큰, API 키, 비밀번호, 쿠키, 개인 키, 서명 secret | `[REDACTED:C1]` |
| C2 개인 식별자 | 이메일, 개인 이름·사용자명이 든 식별자, 전화번호 | `[REDACTED:C2]` |
| C3 절대 홈 경로 | macOS·Linux·Windows 사용자 홈 아래 절대 경로, `~` 전개값 | 저장소 상대 경로로 바꿀 수 있으면 바꾸고, 아니면 `[REDACTED:C3]` |
| C4 원시 로그·스택 | 원시 실행 로그, 패닉·예외 스택, 환경 변수 덤프, 프로세스 목록 | 블록 전체 `[REDACTED:C4]`, 요약 한 줄 |
| C5 비공개 원격 내용 | 비공개 저장소·링크 본문, 접근 제한 artifact 내용 | `[REDACTED:C5]`, 항목 `LINK_BLOCKED` |

- 원값과 원값의 hash는 공개 쪽 어디에도 남기지 않는다(`RESTRICTED`의 SHA-256도 원값의 hash가 아니다, 아래 줄).
- 마지막 관문은 LLM이 아닌 결정적 검사다. 범주 하나라도 남으면 게시 전체를 거절한다.
- 게시할 수 없는 자료는 `RESTRICTED`: 상태, 짧은 요약, SHA-256 세 필드만. SHA-256은 그 항목의 가린 장부 형태에 대해 계산하며, 원값에 대해 계산하지 않는다(DEC-AF-28).
- 적용 대상: 장부 행, 댓글, checkpoint manifest, Issue 요약, 정제 증거 파일, 관찰 출력.
- 알려진 적용 범위 부채(relayed, DEC-AF-29): 정제기가 아직 잡지 못하는 네 갈래 — (i) 한국어 이름 + 전화번호(C2), (ii) 패닉·goroutine 스택(C4), (iii) 물결표 `~/...` 경로(C3 문언이 "`~` 전개값"이라 드는지 모호, OD-23의 질문), (iv) 낱말과 숫자를 대시로 이은 passphrase 형태(C1). 측정은 lane-19가 core 커밋 `9e34d6bb87`에서 했다(리더 전달, 이 SPEC 작성자는 읽지 않음). 위 표와 범주는 바뀌지 않는다.
  - on a lane branch at writing; later also in integration (commit `6e44b1b2aa`) — C2: `redact.go:22` `phone`(한국·+82 전화번호), `:23` `person-name`(전화번호 옆 한국어 이름). 이름만 있을 때 가리지 않는다는 것은 lane-18 통지(리더 전달)이고 이 작성자는 시험을 돌리지 않았다. 부채 (i)은 닫지 않는다.
  - on a lane branch at writing; later also in integration (commit `6e44b1b2aa`) — C4: `redact.go:24` `stack`(패닉 줄·goroutine 스택). 같은 커밋이 `sanitize_test.go`의 "보통 글" 목록에서 goroutine 스택 한 줄을 뺐다. 부채 (ii)는 닫지 않는다.
  - on a lane branch at writing; later also in integration (commit `6e44b1b2aa`) — C1: `redact.go:69-70` credential-assignment 규칙에 `pass(?:phrase|w(?:or)?d)`가 들어 `passphrase: …`·`PASSPHRASE=…`·JSON `"passphrase"`·`ssh_passphrase = '…'` 형태를 잡는다(시험 입력 `sanitize_s2_test.go:164-167`). 이름 없는 대시 연결 낱말열은 구현하지 않았다(lane-18 통지, 리더 전달; 이 작성자의 grep도 그런 규칙을 찾지 못함) — 부채 (iv)의 그 부분은 OD-30. 부채 (iv)는 닫지 않는다.
  - 리더가 전한 커밋 `1aa99d4504`(S1 후속)는 위 범주를 담지 않는다. 그 커밋은 통합 브랜치에 든다. `6e44b1b2aa`는 작성 때(HEAD `0659c71de4`, 통합 끝 `a6f1e93c4f`) 들지 않았고, 2026-10-30T01:50Z(HEAD `937a60ee95`, 통합 끝 `b8a90f33e2`; 다시 2026-10-30T02:27Z, HEAD `ed15692414`, 통합 끝 `a0a7121868`에서도 같은 출력) 든다 — 두 번 모두 `git branch --contains 6e44b1b2aa --list WT-agentic-factory-integration`. 그 뒤 `6ee1da00f4`(통합에 듦)가 범주 이름은 두고 규칙을 넓혔다(Python·Java/Node 스택, hex 안 전화번호 제외, 직함 붙은 이름; spec DEC-AF-29). 시험은 돌리지 않았고 부채 (i)·(ii)·(iv)는 닫지 않는다. C3 물결표는 OD-23 그대로.

**증거 위치와 단계(DEC-AF-13)**: 원본은 `.moai/reports`(ignore 규칙 그대로). 공개 안전 정제본만 추적 디렉터리 `.moai/evidence/issue-<N>/`. 단계 `PREPARED → MASKED → COMMITTED → PUSHED → REMOTE_VERIFIED`. `AVAILABLE`은 원격 조회가 확인한 뒤에만 붙는다. 링크는 전체 커밋 SHA 고정.

### 4.5 GitHub·GitLab 응답 분류 (REQ-AF-014, [10] S53)

**GitHub**

| 행 | 상태 | 헤더 | 본문 | 분류 | 재시도 |
|---|---|---|---|---|---|
| G1 | 403 또는 429 | `retry-after` 있음 | 무관 | `rate-limited` (secondary) | 그 초가 지나기 전 0 |
| G2 | 403 또는 429 | `x-ratelimit-remaining: 0` | 무관 | `rate-limited` (primary, `x-ratelimit-resource` 기록) | `x-ratelimit-reset` 전 0 |
| G3 | 403 또는 429 | 위 둘 없음 | secondary 한도 초과 문구 | `rate-limited` (secondary) | 최소 대기 후 지수 backoff, 정해진 횟수 뒤 오류. 대기 값은 설정 |
| G4 | 429 | 위 헤더 없음, 문구 불명 | 무관 | `rate-limited` (unclassified) | G3과 같음. 권한으로 분류하지 않음 |
| G5 | 403 | `x-ratelimit-remaining` > 0, `retry-after` 없음 | "Resource not accessible by …" | `permission-denied` (`X-Accepted-GitHub-Permissions` 기록) | 0. 다른 자격으로 바꾸지 않음 |
| G6 | 403 | G1~G5와 맞지 않음 | 무관 | `unclassified-403` | 0. 우회 0 |
| G7 | 401 | — | — | `auth-failed` | 0 |
| G8 | 404 | — | — | 원천 없음(삭제·권한의 증거로 쓰지 않음) | DELTA1 단일 404 규칙 |

**GitLab**(범위 포함 여부는 OQ-23)

| 행 | 상태 | 헤더 | 분류 | 재시도 |
|---|---|---|---|---|
| L1 | 429 | `Retry-After`, `RateLimit-ResetTime` | `rate-limited` (`RateLimit-Name` 기록) | `Retry-After` 뒤. 계속이면 지수 backoff |
| L2 | 429 | `Retry-After` 없음 | `rate-limited` (unclassified) | 설정값 backoff |
| L3 | 403 | — | `permission-denied` | 0 |
| L4 | 404 | — | 원천 없음 | 삭제·권한 증거 아님 |
| L5 | 정상 | `RateLimit-Remaining`·`RateLimit-Reset` | 남은 몫 관측 | 설정 경계 아래면 미리 늦춤 |
| L6 | 정상 | 헤더 없음(Projects·Groups·Users API) | 남은 몫 `UNKNOWN` | 남은 몫이 있다고 보지 않음 |

**출처**(2026-10-29 WebFetch 요약, 원문 대조 안 함 — uncollated)

| 표기 | URL |
|---|---|
| U1 | https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api |
| U2 | https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api |
| U3 | https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api |
| U4 | https://docs.gitlab.com/administration/settings/user_and_ip_rate_limits/ |
| U5 | https://docs.gitlab.com/user/gitlab_com/rate_limits/ |
| U6 | https://docs.gitlab.com/api/rest/troubleshooting/ |
| U7 | https://docs.gitlab.com/api/rest/ |
| U8 | https://docs.gitlab.com/rate_limits/ |

**예산 규칙**

- 수집기는 저장소마다 활성 하나. webhook이 있으면 우선, 없으면 조건부 요청(ETag/`If-None-Match`)으로 polling. agent는 자기 polling 경로가 없다(REQ-AF-017).
- 변경 요청은 직렬. 공유 예산 아래 우선순위: 상태 안전 쓰기 > 안전 알림 > 진척·보고 댓글(SYNC15).
- 한도·남은 몫·재개 시각은 응답 헤더에서만 얻는다. 없으면 `UNKNOWN`이며 보수적으로(남은 몫이 없다고) 다룬다.
- 문서에 적힌 값은 공급자 설정 스키마의 기본값(§6)으로만 둔다. 출처 URL·조회 날짜·`uncollated` 표시를 함께 둔다. 코드에 숫자를 두지 않는다(DEC-AF-14).
- 같은 데이터·효과를 얻으려고 키·토큰·계정·앱 설치·endpoint를 바꾸지 않는다. 서버·자격·scope·권한·비용을 늘리는 동작은 운영자 별도 승인.

### 4.6 checkpoint 복구 (REQ-AF-002 (c)~(e), DEC-AF-20)

출처: '상위 대화에서 사용자가 승인, 리더 세션 경유'. 소유: 상태·git·manifest = lane-18, Issue 링크 = lane-20, 충돌 회귀·측정 = lane-22.

**네 단계**(각각 독립된 증거와 실패 처리)

| 단계 | 하는 일 | 증거 | 실패 시 |
|---|---|---|---|
| K1 커밋 | attempt가 스스로 스테이징한 경로만 Issue branch에 커밋 | 커밋 전체 SHA, 커밋한 경로 목록 | 상태 `local`로 남지 않음(커밋 없음). attempt 계속 |
| K2 원격 ACK | 그 SHA를 Issue branch에 push(강제 없음) | push 결과 + `ls-remote` readback: ref = SHA | `committed-unacked`. push 전 원격 ref 조회 |
| K3 manifest 검증 | control ref에 manifest를 CAS로 쓰고 readback | readback digest = 쓴 digest, manifest의 SHA = K2 SHA | `acked-manifest-unverified` |
| K4 Issue 링크 | Issue canonical context block에 요약과 고정 링크 | 댓글 effect receipt + 원격 readback | `durable-link-pending`. checkpoint는 `remote-durable` 유지 |

**상태와 재개 규칙**

| 상태 | 뜻 | 상위 상태 | 재개 규칙 |
|---|---|---|---|
| `local` | K1만 끝남, push 의도 없음 | — | K2부터 |
| `pending` | push 의도 기록, ACK 미관측 | — | 아래 `committed-unacked`와 같음 |
| `committed-unacked` | K1 뒤 ACK 전에 충돌·실패 | `pending` | push 전에 원격 ref를 조회. ref = SHA면 ACK 관측으로 보고 K3, 아니면 push. push 횟수는 운영자 결정 대기 OD-11(0.6.1까지의 문언은 "한 번"). 결과 불명이면 PARKED `unknown-effect` |
| `acked-manifest-unverified` | ACK는 있으나 manifest CAS 실패·불일치 | `pending` | `remote-durable` 아님, `recovered` 불가. control ref를 다시 읽고 현재 epoch로 CAS 재기록 |
| `remote-durable` | K2 ACK와 K3 manifest 검증 둘 다 관측 | — | 재사용 가능 |
| `durable-link-pending` | K4 실패 | `remote-durable` | 복구를 취소하지 않음. 링크를 보고 등급 쓰기로 예산 규칙 아래 재시도 |
| `recovered` | 새 writer가 소유 확인 뒤 이 checkpoint에서 재개 | — | — |

**중복 push 금지**: push 전에 같은 ref·SHA의 ACK가 이미 있는지 control ref와 원격 ref로 확인하고, 있으면 push하지 않는다.

**재사용·다른 host 복구**: 이전 writer가 `homestate.ClassifyOwnerWith`로 `dead`이거나 안전 정지를 기록했을 때만, epoch를 올리고(fencing) 재개한다. `live`·`indeterminate`면 거절(`ownership-unverified`). 복구는 두 번째 writer를 시작하지 않는다.

**manifest 필드**(전부 C1~C5 정제 대상)

| 필드 | 내용 |
|---|---|
| `branch` | Issue branch ref |
| `sha` | 커밋 전체 SHA |
| `phase` | 상태와 하위 단계(§2.1) |
| `next_action` | 다음 행동 한 줄 |
| `tests` | 시험 이름, 결과, 실행 기록 위치 |
| `owner` | agent session id |
| `epoch` | 배정 epoch |
| `unresolved_effects` | 확정되지 않은 effect receipt id 목록 |

**금지**: 사용자의 커밋되지 않은 변경을 자동 스테이징·커밋, 미완성 작업을 `main`에 push(대상 ref가 기본 branch면 거절), 강제 push·이력 재작성(non-fast-forward면 거절하고 PARKED), 비밀 스냅숏(C1~C5 값이 남으면 manifest 기록 거절).

**의미 단위·저장 상한**: checkpoint는 시간 간격이 아니라 의미 단위(마일스톤·task 완료, PR 인계 직전, 되돌리기 어려운 동작 직전)마다 만든다. 저장은 아래 매개변수로 묶는다. 숫자 기본값은 모두 **OPEN**이며 이 SPEC은 숫자를 정하지 않는다(OQ-22).

| 키 | 뜻 | 기본값 |
|---|---|---|
| `workflow.agentic_factory.checkpoint.max_per_issue` | Issue당 manifest 색인에 남기는 checkpoint 수 | OPEN |
| `workflow.agentic_factory.checkpoint.retention` | 보존 기간 | OPEN |
| `workflow.agentic_factory.checkpoint.max_manifest_bytes` | manifest 하나의 크기 상한 | OPEN |

상한을 넘을 때: 가장 오래된 checkpoint 가운데 "가장 최근 `remote-durable`이 아니고, 확정되지 않은 effect receipt가 참조하지 않는" 것을 manifest 색인에서 뺀다(커밋과 이력은 지우지 않는다). 뺄 수 있는 것이 없으면 새 checkpoint를 `capacity-short`로 거절하고 기존 것은 그대로 둔다. 키가 미설정일 때의 동작은 정하지 않았다(운영자 결정 대기 OD-8, spec.md §G.1).

**구현 보고(미검증)**: lane-18은 커밋 `e6499b731a`에 상태 `local`·`pending`·`remote-durable`·`recovered`를 구현했다고 보고했다('reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file'). 위 표의 `committed-unacked`·`acked-manifest-unverified`·`durable-link-pending`은 그 보고에 없다. 관찰로만 기록한다(spec.md §H 공백 (d)).

**정본과의 관계(0.6.2, 관찰)**: 위 네 단계 K1~K4와 상태 일곱 개는 이 SPEC이 DEC-AF-20을 풀어 쓴 것이며, 정본 댓글에는 이 이름들이 없다(spec.md §H 공백 (f)). lane-18은 코드의 상태 넷과 필드로 같은 정보를 나타낼 수 있다고 보고했다(`checkpoint.go:125-146`, spec.md §B.2). 이 SPEC은 어느 쪽에 맞추지도 않는다.

**측정**: checkpoint마다 API 호출 수, 토큰, 걸린 시간을 측정값으로 기록한다. 측정할 수 없는 값(현재 토큰 receipt 경로 없음)은 Gap으로 기록하고 추정값을 쓰지 않는다(REQ-AF-025).

### 4.7 효율 비교 기록 어휘 (REQ-AF-025, [9] S52)

| 필드 | 내용 |
|---|---|
| `axis` | 정본 S52의 측정 축 하나: 토큰·모델 비용, 시간(실제 수행과 큐·리뷰·권한 대기 분리), API·반복, 실패·복구(분모와 관측 기간), 정보 이용, 품질·안전·정확성 |
| `baseline_source` | 구 Factory 기준 원천(보존 로그 또는 승인된 격리 fixture만. 없으면 `NOT_RUN` 또는 비교 불가. 기준을 얻으려고 GLM이나 보류된 구 작업을 재개하지 않음) |
| `comparability` | 과제 크기·난도, 모델·effort, 권한, 코드 revision, 환경, 측정 구간의 일치 여부. 맞지 않으면 비교 한계 표시 |
| `value_kind` | `measured` / `estimated` / `UNKNOWN` / `NOT_RUN` |
| `value`, `unit` | 값과 단위 |
| `command` | 측정 명령 |
| `tree` | 측정한 트리 SHA |
| `quality_guard` | 같은 비교의 품질·안전·정확성 지표와 결과. 회귀가 있으면 효율 향상을 보고하지 않음 |
| `threshold_state` | 임계값·표본 설계가 `proposed`인지 `approved`인지(정본 요구, OD-1) |

### 4.8 record (REQ-AF-017, [3] DELTA1, [7] S46)

| record | 필드 |
|---|---|
| `CollectorCheckpoint` | 저장소, 수집 cursor, ETag, 마지막 성공 시각, 수집기 id |
| `SourceRevision` | 원천 객체 id, revision, 관측 시각, 내용 digest |
| `ConsumerDelivery` | 소비자 id, 전달한 revision, 전달 시각 |
| `ConsumerACK` | 소비자 id, ACK한 revision, ACK 시각 |
| `SourceTombstone` | 사라진 원천 id, 관측 근거(단일 404는 근거 아님) |
| `PRQueue` | 저장소, 열린 PR 인계 수, 수리 적체 수, `wip_limit`, 계산 시각 |
| `PRHandoff` | PR URL, 원격 HEAD/base, 계약·SPEC revision, task 결과, 검사 상태, 원 writer, leader 수락 receipt, quiescence 증거, 보존한 worktree·branch·변경. **lane-18 필드 이름(0.6.7, 커밋 `2251e77393`, 리더 전달·이 작성자는 다시 읽지 않음)**: 새 필수 필드 `PR URL`(http/https), `base`, `adopted SPEC`, `task result`, `check status`. 유지 필드 `handoff id`, `LeaderAccepted`, `PR number`, `HEAD`, `branch`, `worktree`, `WriterStopped`, `BoundaryHead`. **대응(리더 판정, 리더 세션 경유)**: "계약" = `adopted SPEC`(AdoptedSpec). "보존한 변경과 quiescence 증거" = `WriterStopped` + `BoundaryHead`==`HEAD` + `branch` + `worktree`(예약된 실제 경로). "원 writer의 출처"는 receipt 필드가 아니라 claim 획득 때의 Issue 작성자 검사(`ErrNotIssueAuthor`)가 충족한다. receipt가 담는 책임 주체 출처는 `Session`/`Agent`다. receipt 필드는 새로 더하지 않는다. 옛 receipt는 이미 기록된 과거 상태로 읽기만 하며 다시 검증하지 않는다(읽기 호환). 모든 필드가 있어야 한다는 조건은 새 인계(E03 쓰기)에만 적용된다 |
| `EffectIntent` / `EffectReceipt` | effect id, dedup key, 종류, 대상, 보내기 전 기록 시각, 결과·원격 readback |

## §5 정책

| 항목 | 정책 | 검증 목표 |
|---|---|---|
| GLM managed | 구현과 로컬 환경 검증 모두에서 제외 | GLM 백엔드 호출 0 |
| Codex | `workflow.agentic_factory.codex.enabled: true`일 때만. 미설정·`false`면 호출 0 | 미설정 상태 Codex 호출 0 |
| 인증·권한·rate limit | §4.5 분류 뒤에만 재시도 판단. 우회 금지 | 분류 전 재시도 0 |
| 새 키·유료 용량·추가 지출 | 운영자 별도 승인 | 승인 없이 0 |
| 브라우저 | `egolite`/`aside`/`chrome` 중 하나. 미선택일 때만 `chrome`. 선택 후 미설치·연결 실패는 PARKED, 자동 대체 없음 | 선택하지 않은 브라우저 호출 0 |

### 5.1 로컬 환경 행렬 (REQ-AF-024)

| 열 | 뜻 |
|---|---|
| `environment` | 환경 이름 |
| `installable` | 설치 관측(`yes`/`no`/`denied`) |
| `version` | 실제 실행한 버전 명령 출력 |
| `start` / `receive` / `test` / `hold_resume` / `collect` | 기동, 작업 수신, 시험, 안전 보류·재개, 결과 회수의 실제 실행 결과 |
| `verdict` | `PASS` 또는 `BLOCKED(<사유>)` |

각 단계 칸은 실제로 실행한 명령, 종료 코드, 출력 일부를 담는다. 실행하지 않은 단계는 `not-run`이며 성공이 아니다. `PASS`는 여섯 단계가 모두 실행되고 성공했을 때만이다. 적어도 하나의 환경이 여섯 단계를 실제로 실행한 기록이 있어야 한다. Codex 환경은 명시 설정이 없으면 `BLOCKED(not-configured)`로 남고, 행렬을 채우려고 Codex를 켜지 않는다.

### 5.2 브라우저 어댑터 계약 (lane-21 보고 옮김, 0.5.0 그대로)

아래 사실은 lane-21이 2026-10-29 로컬 macOS 한 대(ego-browser 0.6.1.13, aside 1.26.1023.1115)에서 도움말·가이드 출력과 웹 문서 요약으로 확인한 것이다. 이 SPEC 작성 세션은 그 명령을 다시 실행하지 않았다.

| 항목 | egolite (ego-lite의 ego-browser) | aside |
|---|---|---|
| 정본 출처 | `citrolabs/ego-lite` 저장소의 `skills/ego-browser/SKILL.md`, `references/install.md` | `docs.aside.com/help/developers` |
| 설치 감지 | `command -v ego-browser` | `aside --version` |
| 기능(연결) 점검 | `ego-browser nodejs`로 `console.log('ego-browser ready')` 실행 → `ego-browser ready` 출력 확인. 전용 `status` 명령 없음 | 공식 문서에 전용 점검 명령 없음. 로그인 감지 후보 `aside account status`(출력 형식 미확인) |
| 사용 전 선행 | `ego-browser skill` 실행 후 출력을 따름. 끝 줄 `[ego-browser:skill] end of guide`가 없으면 잘린 것 | `ASIDE_CLI_AUTH_PATH`를 비공개 절대 경로로 지정 후 `aside login --email ...` |
| 실행 형태 | `ego-browser nodejs <<'EOF' … EOF`, `-e '<script>'`, `< script.js`. 스크립트는 셸 환경변수를 상속하지 않음 | `aside "<prompt>"` 또는 `aside exec`, 이어 하기 `aside --session <id> "Continue"`, `aside repl`, `aside mcp` |
| 권한 | 해당 없음 | `--permission ask\|guard\|full-access`(생략 시 Guard). `aside repl`에는 `--permission` 없음 |
| 작업 단위·종료 | 목표당 TaskSpace 하나. 사용자 소유 공간이면 `EGO_TASK_SPACE_NAME_IN_USE`. 성공 시 `task.finish({ keep: [] })`. 사용자 제어 필요 시 `task.handOff()` | 세션 id로 이어 하기 |
| 업그레이드 | `[ego-browser:notice]`가 보이면 작업을 끝내고 알림, `ego-browser upgrade`는 사용자 승인 뒤 | `aside --update` |
| 비밀 | — | `ASIDE_CLI_AUTH_PATH`가 가리키는 자격 증명 파일을 읽거나 출력하거나 커밋하지 않음 |

사유 근거: `not-installed` = 설치 감지 실패, `connection-failed` = 감지 성공 + 기능 점검 실패, `not-logged-in` = aside 로그아웃 경고(감지 명령 미확정), `user-control` = `task.handOff()` 또는 `EGO_TASK_SPACE_NAME_IN_USE`.

SPEC-ASIDE-BROWSER-001은 aside 부재 시 조용히 기본 도구로 진행(REQ-ASB-012)하고 `CI=true`에서 aside를 제외(REQ-ASB-011)한다. 앞의 것은 #1948의 "명시 선택 후 실패하면 차단"과 반대다. 이 SPEC은 #1948을 따르며 SPEC-ASIDE-BROWSER-001 개정은 미룬다(OQ-13). chrome(claude-in-chrome) 계약은 조사하지 않았다. lane-21이 남긴 확인 공백 ①~⑧(clearing-state 미열람, 저장소 루트 미열람, 연결 실패 문구 미측정, aside 공식 정의 부재, MCP 도구 목록은 로컬 도움말뿐, chrome 미조사, WebFetch 요약, 한 시점·한 기기)은 그대로다.

## §6 설정 키 (REQ-AF-021)

키는 lane-17이 읽기 코드로 만든 것(✓)과 이 SPEC이 제안하는 것으로 나뉜다. 기존 설정 파일은 이 SPEC이 고치지 않는다. 범위 밖 값은 키 경로·허용 범위·파일을 밝히고 거절하며, 값을 깎아 맞추지 않는다. 읽기 오류는 미설정으로 보지 않는다.

| 키 | 형식 | 기본값 | 검증 | 상태 |
|---|---|---|---|---|
| `workflow.agentic_factory.enabled` | bool | — | 필수. bool 외 거절. 키가 없거나 `false`이면 AF는 설정되지 않은 것으로 다룬다(오류가 아니며 이유를 보인다). `true`인데 다른 필수 키가 없거나 범위를 어기면 검증 오류(키 이름 포함)이며 그 결과도 AF 미설정으로 표시(§6.1 두 단계 판정) | 리더 결정 (리더 세션 경유, 0.6.7; §6.1·DEC-AF-25 D3) |
| `workflow.agentic_factory.codex.enabled` | bool, 미설정 허용 | 미설정(=꺼짐) | bool 외 거절. 미설정·`false`는 호출 0 | ✓ `CodexOptIn` |
| `workflow.agentic_factory.browser.selected` | `egolite` \| `aside` \| `chrome`, 미설정 허용 | 미설정(=사용 시 `chrome`) | 열거 외 거절. 미설정은 "선택 없음"이지 `chrome` 선택이 아님 | ✓ `ResolveAgenticFactoryBrowser` |
| `workflow.agentic_factory.pr_queue.wip_limit` | 정수 | 미설정 | 1 이상. `0`·음수·정수 아님·범위 밖 거절. `0`을 "끔"으로 읽지 않음. 미설정일 때의 배차는 운영자 결정 대기(OD-7) | ✓ `WIPLimit` |
| `workflow.agentic_factory.collaboration.enabled` | bool | `false` | bool 외 거절 | ✓ |
| `workflow.agentic_factory.collaboration.allowed_repositories` | 문자열 목록 | 빈 목록 | §4.1 항목 검증 | ✓ |
| `workflow.agentic_factory.collaboration.permissions.edit_body` / `.close` / `.labels` / `.assignees` | bool | 각 `false` | bool 외 거절. **남의 Issue에만 적용** | ✓ |
| `workflow.agentic_factory.intake.max_concurrent` | 정수 | 미설정 | 1 이상. 미설정일 때의 동작은 운영자 결정 대기(OD-9) | 제안. 숫자 OPEN |
| `workflow.agentic_factory.intake.stage_timeout` | duration | 미설정 | 0보다 큼 | 제안. 숫자 OPEN |
| `workflow.agentic_factory.checkpoint.*` | §4.6 | OPEN | §4.6 | 제안 |
| `workflow.agentic_factory.provider.<github\|gitlab>.rate_limits.*` | 공급자 문서값 | 문서값 + `source_url`, `fetched_at`, `collation: uncollated` | 값이 없으면 `UNKNOWN` | 제안 |
| `workflow.agentic_factory.provider.<…>.backoff.*` | 대기·배수·횟수 | OPEN | 0보다 큼 | 제안 |
| `workflow.agentic_factory.budget.reserve` | 안전 쓰기 예비 몫 | OPEN | 0 이상 | 제안 |
| `workflow.agentic_factory.comments.batch_window` | duration | OPEN | 0보다 큼 | 제안 |

아직 이름이 없던 키(0.6.7): lane-20 코드에만 있는 `MaxChunks`·`SeenLimit` 등의 키 이름과 "AF가 설정되었다" 판정 키는 이제 아래 §6.1 표에 있다(DEC-AF-25). 이 표에는 행을 더하지 않았다. 새 키는 새 파일 `.moai/config/sections/agentic-factory.yaml`에만 넣고 기존 설정 파일은 바꾸지 않는다(리더 결정).

Template-First 연동: 실제 삽입 때 로컬 `.moai/config/sections/*.yaml`과 `internal/template/templates/.moai/config/sections/*.yaml`을 함께 바꾸고 `internal/config` 타입·기본값과 `internal/settings` 스키마를 맞춘다. 템플릿에 `codex.enabled: true`를 넣지 않는다.

### 6.1 설정 키 표 (DEC-AF-25로 대부분 확정, 남은 행은 리더 결정 대기)

OD-13·OD-14·OD-17(spec.md §G.1)을 한 표로 묶는다. 출처: 'leader decision, relayed via the leader session'. 행의 코드 사실은 lane-20의 사실 통지(커밋 `efeb3dc87c`, 브랜치 `WT-agentic-monitor-effect-reconcile`)를 리더가 전달한 것이며, **이 작성자는 그 커밋을 다시 읽지 않았다**. 확정은 DEC-AF-25(spec.md §C.7, 'leader decision (Jev consultation; relayed via the leader session)')이다. Jev 점수는 확인 id가 다른 빌드에서 나온 참고 신호(advisory signal)일 뿐 결정이 아니다.

규칙:
- **기본값은 하나도 정하지 않는다.** 모든 행의 기본값 칸은 `—`다. 코드가 `0`에 주는 뜻은 "검증 규칙" 칸에 코드 사실로만 적는다.
- **필수(DEC-AF-25 D5)**: 상태가 "confirmed (leader decision)"인 키는 모두 필수이며, 없으면 거절한다. 이 규칙은 키가 **있어야 한다**는 것이며, 값의 허용 범위는 "검증 규칙" 칸 그대로다(예: `ghidentity.GitHub.MaxAge`의 `0` 허용). 이름이 정해지지 않은 행(상태 "리더 결정 대기")에는 아직 적용하지 않는다.
- **키 이름**: 기존 접두사 `workflow.agentic_factory.*` 아래에 둔다. 상태가 "confirmed (leader decision)"인 행의 이름은 확정이다(DEC-AF-25 D5). 상태가 "리더 결정 대기"인 행(webhook 비밀 참조 `secret_ref`, `Tick` 주기 `loop.tick_period`)의 이름만 아직 제안(provisional)이다(OD-17). lane-17 읽기 코드에 이미 있는 키(browser·codex 키, `pr_queue.wip_limit`, `collaboration.*`)는 이 표에 다시 넣지 않았다. 위 §6 표의 기존 제안 행 `provider.<…>.backoff.*`·`budget.reserve`와 이름을 맞췄다.
- **위치(DEC-AF-25 D1)**: 이 표의 키는 모두 **새 파일 `.moai/config/sections/agentic-factory.yaml`에만** 둔다. 기존 설정 파일, `defaults.go`, `closed_sets.go`, `session_defaults.go`, `schema_sections.go`, `i18n.js`는 바꾸지 않는다. 그래서 바로 위 Template-First 문단의 "기존 sections 파일을 함께 바꾼다"는 이 표의 키에 적용하지 않는다. 이 파일을 template으로 배포할지는 따로 열린 결정이다(OD-19). 결정 전에는 template 배포 파일을 만들지 않는다 — 만들면 `moai update`가 사용자 프로젝트에 그 파일을 쓴다.
- **"AF가 설정되었다"(DEC-AF-25 D3)**: 새 파일이 있고, `workflow.agentic_factory.enabled: true`이고, 필수 키가 모두 검증을 통과할 때만 설정된 것으로 본다. 하나라도 어긋나면 설정되지 않은 것으로 다루고 실패 이유를 보여 준다. 판정은 두 단계다(리더 결정, 리더 세션 경유): enabled가 없거나 false → AF 미설정(오류 아님). enabled가 true인데 다른 필수 키가 없거나 범위 위반 → 검증 오류(키 이름 포함)이며 그 결과도 AF 미설정으로 표시.
- **inbox 키 분리(DEC-AF-25 D4)**: webhook 큐와 agent inbox는 크기 요구가 다른 별개 인스턴스다. 그래서 `notify.inbox.{max_doc_bytes,seen_limit,max_id_bytes}`는 따로 두고, `queue.webhook.*`은 webhook 큐에만 쓴다.
- **비밀**: webhook 경로 서명 비밀은 값이 아니라 **참조**(어디서 읽을지의 이름)로만 둔다. 커밋되는 설정에 비밀 값을 넣지 않는다. 비밀을 어디에 둘지는 이 SPEC 소유자의 결정이며 열려 있다(OD-17).
- **위 §6 표와의 차이(해결)**: 위 §6 표의 `enabled` 행이 기본값 `false`(0.6.0 제안)로 이 표와 어긋났던 것을 리더 결정(리더 세션 경유)에 따라 고쳤다 — 이제 두 표 모두 기본값 `—`·필수이며, 키가 없거나 `false`이면 오류가 아니라 설정되지 않음으로 다룬다.
- **기준 목록**: lane-20은 자기 코드를 다시 읽은 뒤 처음 통지를 스스로 정정했고, 그 정정에 lane-20의 새 설정 필드 두 개(`ghintake.GH.ListLimit`, `adapt.GHReader.PRListLimit`; 커밋 `af72c90458`, 브랜치 `WT-agentic-monitor-effect-reconcile`)를 더한 것이 아래 표의 기준이다(리더 전달, 이 작성자는 다시 읽지 않음). 이로써 예전 하드코딩 값 14a·14b는 이 두 필드로 바뀌었고, 14c(`docstore.modifyRetries = 10`)는 사라졌다. 그래서 lane-20 범위에는 하드코딩된 한도가 남지 않는다. 이 밖에 lane-20 코드에 남은 숫자는 forge host가 `Retry-After`·reset 헤더로 주는 값이며, host가 준 값이지 정책 상수가 아니다.

| 코드 필드(소관 lane/패키지) | 키 이름 | 위치 | 단위 | 검증 규칙(코드가 현재 하는 것) | 기본값 | 상태 |
|---|---|---|---|---|---|---|
| (이 SPEC, "AF가 설정되었다" 판정 입력, DEC-AF-25 D3) | `workflow.agentic_factory.enabled` | `.moai/config/sections/agentic-factory.yaml` | bool | 코드에 읽기 없음(lane-17 읽기 코드에 없음, OD-14 기록). 필수. bool 외 거절. 판정은 파일 존재 + 이 키 `true` + 필수 키 전부 검증 통과. 어긋나면 설정되지 않음으로 다루고 실패 이유를 보임. 키가 없거나 `false`이면 오류가 아니라 설정되지 않음. `true`인데 다른 필수 키가 없거나 범위를 어기면 검증 오류(키 이름 포함)이며 결과는 설정되지 않음으로 표시(두 단계 판정, §6 표와 같은 문언) | — | confirmed (leader decision) |
| (이 SPEC, claim 저장소 디렉터리, DEC-AF-25 D2) | `workflow.agentic_factory.claim_store.dir` | `.moai/config/sections/agentic-factory.yaml` | 경로 | 코드에 읽기 없음(lane-21 보고: `claim.Open`에 시험 밖 호출자 없음, OD-13 기록). 필수 | — | confirmed (leader decision) |
| `forge.Backoff.MinSecondaryWait` (lane-20 `forge`) | `workflow.agentic_factory.provider.github.backoff.min_secondary_wait` | `.moai/config/sections/agentic-factory.yaml` | duration | 기본값 없음. `Validate`가 `<=0` 거절, `Max < MinSecondaryWait`면 거절 | — | confirmed (leader decision) |
| `forge.Backoff.Base` (lane-20 `forge`) | `workflow.agentic_factory.provider.github.backoff.base` | `.moai/config/sections/agentic-factory.yaml` | duration | 기본값 없음. `<=0` 거절, `Max < Base`면 거절 | — | confirmed (leader decision) |
| `forge.Backoff.Max` (lane-20 `forge`) | `workflow.agentic_factory.provider.github.backoff.max` | `.moai/config/sections/agentic-factory.yaml` | duration | 기본값 없음. `<=0` 거절, `MinSecondaryWait`·`Base`보다 작으면 거절 | — | confirmed (leader decision) |
| `forge.Backoff.MaxAttempts` (lane-20 `forge`) | `workflow.agentic_factory.provider.github.backoff.max_attempts` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 기본값 없음. `<=0` 거절 | — | confirmed (leader decision) |
| `forge.Adaptive.Min` (lane-20 `forge`, 수집 polling) | `workflow.agentic_factory.collector.poll.min` | `.moai/config/sections/agentic-factory.yaml` | duration | 간격은 Min~Max 사이에서 반으로 줄거나 두 배가 된다. 기본값 없음. **검증 코드 없음**(`0`이면 0만큼 기다림). 필요한 규칙 `Min>0`은 lane-20이 아직 집행하지 않음 | — | confirmed (leader decision) |
| `forge.Adaptive.Max` (lane-20 `forge`) | `workflow.agentic_factory.collector.poll.max` | `.moai/config/sections/agentic-factory.yaml` | duration | 기본값 없음. 검증 코드 없음. 필요한 규칙 `Max>=Min`은 아직 집행 안 함 | — | confirmed (leader decision) |
| `forge.Adaptive.Safety` (lane-20 `forge`) | `workflow.agentic_factory.collector.poll.safety` | `.moai/config/sections/agentic-factory.yaml` | duration | push 경로가 정상인 동안의 안전망 polling 간격. 기본값 없음. 검증 코드 없음. 필요한 규칙 `Safety>0`은 아직 집행 안 함 | — | confirmed (leader decision) |
| `forge.Budget.ReserveCalls` (lane-20 `forge`) | `workflow.agentic_factory.budget.reserve` | `.moai/config/sections/agentic-factory.yaml` | 정수(호출 수) | 급한 쓰기를 위해 남겨 두는 호출 수. 검증 없음. `0` = 예비 몫 없음 | — | confirmed (leader decision) |
| `forge.Budget.ReportResource` (lane-20 `forge`) | `workflow.agentic_factory.budget.report_resource` | `.moai/config/sections/agentic-factory.yaml` | 문자열 | 보고 쓰기가 쓰는 host 자원 이름. `''` = 기본 버킷 | — | confirmed (leader decision) |
| `ghidentity.GitHub.MaxAge` (lane-20 `ghidentity`) | `workflow.agentic_factory.identity.max_age` | `.moai/config/sections/agentic-factory.yaml` | duration | 조회한 사실의 재사용 시간. `0` 허용 = 매번 조회 | — | confirmed (leader decision) |
| `labels.GitHubHost.PerPage` (lane-20 `labels`) | `workflow.agentic_factory.labels.per_page` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 목록 페이지 크기. 필수, `0`은 오류 | — | confirmed (leader decision) |
| `ghintake.Publisher.Limit` (lane-20 `ghintake`) | `workflow.agentic_factory.intake.duplicate_search_limit` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 중복 후보 검색 상한. `0`이면 `FindDuplicates`가 거절 | — | confirmed (leader decision) |
| `docstore.Cursors.MaxDocBytes` (lane-20 `docstore`) | `workflow.agentic_factory.cursor.max_doc_bytes` | `.moai/config/sections/agentic-factory.yaml` | bytes | 필수, 정수 >= 1. `0`은 거절한다(`0`은 "상한 없음"과 구별되지 않음, 리더 결정). claim 한도를 넘으면 저장 때 거절. 256 KiB claim 문서 한도의 사본은 lane-20 내부 상수이며 설정 값이 아니다(시험이 실제 저장소 값에 고정) | — | confirmed (leader decision) |
| `docstore.Cursors.MaxChunks` (lane-20 `docstore`) | `workflow.agentic_factory.cursor.max_chunks` | `.moai/config/sections/agentic-factory.yaml` | 정수 | `0` = 나누기 끔(한도를 넘는 cursor는 `ErrCursorTooLarge`로 거절). 양수 = 그 수까지 chunk로 나눔, 넘으면 거절 | — | confirmed (leader decision) |
| `docstore.Queue.MaxDocBytes` (lane-20 `docstore`, webhook 큐) | `workflow.agentic_factory.queue.webhook.max_doc_bytes` | `.moai/config/sections/agentic-factory.yaml` | bytes | 필수, 정수 >= 1(`0`·음수 거절, 리더 결정). 그래서 `seen_limit` 부등식은 언제나 검사된다 | — | confirmed (leader decision) |
| `docstore.Queue.SeenLimit` (lane-20 `docstore`, webhook 큐) | `workflow.agentic_factory.queue.webhook.seen_limit` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 기억하는 소비된 전달 id 수. `0` = 기억하지 않음. 음수 거절. `SeenLimit*(MaxIDBytes+3)*2 <= MaxDocBytes`가 아니면 `ErrQueueConfig` | — | confirmed (leader decision) |
| `docstore.Queue.MaxIDBytes` (lane-20 `docstore`, webhook 큐) | `workflow.agentic_factory.queue.webhook.max_id_bytes` | `.moai/config/sections/agentic-factory.yaml` | 정수(bytes) | id 최대 길이. 언제나 필수(`seen_limit`이 `0`이어도, DEC-AF-25 D5, 리더 결정). 음수 거절. 참고 길이: webhook 전달 id = UUID 36 | — | confirmed (leader decision) |
| `notify.Inbox.MaxDocBytes` (lane-20 `notify`, agent inbox) | `workflow.agentic_factory.notify.inbox.max_doc_bytes` | `.moai/config/sections/agentic-factory.yaml` | bytes | 필수, 정수 >= 1(`0`·음수 거절, 리더 결정). 그래서 `seen_limit` 부등식은 언제나 검사된다. webhook 큐와 다른 인스턴스(DEC-AF-25 D4) | — | confirmed (leader decision) |
| `notify.Inbox.SeenLimit` (lane-20 `notify`, agent inbox) | `workflow.agentic_factory.notify.inbox.seen_limit` | `.moai/config/sections/agentic-factory.yaml` | 정수 | webhook 큐와 같은 규칙. `0` = 기억하지 않음. 음수 거절. `SeenLimit*(MaxIDBytes+3)*2 <= MaxDocBytes`가 아니면 `ErrQueueConfig` | — | confirmed (leader decision) |
| `notify.Inbox.MaxIDBytes` (lane-20 `notify`, agent inbox) | `workflow.agentic_factory.notify.inbox.max_id_bytes` | `.moai/config/sections/agentic-factory.yaml` | 정수(bytes) | id 최대 길이. 언제나 필수(`seen_limit`이 `0`이어도, DEC-AF-25 D5, 리더 결정). 음수 거절. 참고 길이: inbox 메시지 id = 32 | — | confirmed (leader decision) |
| `webhook.Handler.MaxBody` (lane-20 `webhook`) | `workflow.agentic_factory.webhook.max_body` | `.moai/config/sections/agentic-factory.yaml` | bytes | `0`이면 모든 전달이 503 | — | confirmed (leader decision) |
| `webhook.Handler.Provider` (lane-20 `webhook`) | `workflow.agentic_factory.webhook.provider` | `.moai/config/sections/agentic-factory.yaml` | 문자열 | `github`면 저장소 이름을 대소문자 무시로 비교, 그 밖은 정확히 일치 | — | confirmed (leader decision) |
| `webhook.Handler.Routes[].Repo` (lane-20 `webhook`) | `workflow.agentic_factory.webhook.routes[].repository` | `.moai/config/sections/agentic-factory.yaml` | `owner/name` | 경로마다 저장소 하나 | — | confirmed (leader decision) |
| `webhook.Handler.Routes[].Secret` (lane-20 `webhook`) | `workflow.agentic_factory.webhook.routes[].secret_ref` (provisional) | `.moai/config/sections/agentic-factory.yaml`. **값이 아닌 참조만** | 참조(이름) | 전달마다 서명 비밀을 읽는 함수. 비어 있으면 500 거절. lane-20은 값을 어떤 파일에도 쓰지 않음. 비밀의 위치는 열림(OD-17) | — | 리더 결정 대기 |
| `notify.Service.ReNotifyAfter` (lane-20 `notify`) | `workflow.agentic_factory.notify.renotify_after` | `.moai/config/sections/agentic-factory.yaml` | duration | 다시 알린 뒤 다음 재확인까지의 시간. 필수, `<=0`이면 `Tick`이 거절 | — | confirmed (leader decision) |
| 대기 기록 `MaxRetries` (lane-18 `core.Wait`가 요구, 기록하는 쪽이 설정에서 읽어야 함) | `workflow.agentic_factory.wait.max_retries` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 기본값 없음. `<1` 거절. design §2.1c "재시도 상한" 필드의 값 | — | confirmed (leader decision) |
| 대기 기록 `escalate_to` (lane-18 `core.Wait`가 요구) | `workflow.agentic_factory.wait.escalate_to` | `.moai/config/sections/agentic-factory.yaml` | 문자열 | 필수. design §2.1c "상승 대상" 필드의 값 | — | confirmed (leader decision) |
| `ghpr.Source.Repo` (lane-20 `ghpr`) | `workflow.agentic_factory.pr_source.repository` | `.moai/config/sections/agentic-factory.yaml` | `owner/name` | 필수 | — | confirmed (leader decision) |
| (아직 코드 없음) 상주 루프의 `Tick` 주기 | `workflow.agentic_factory.loop.tick_period` (provisional) | `.moai/config/sections/agentic-factory.yaml` | duration | 코드 없음. 키가 정해질 때까지 루프를 쓰지 않음 | — | 리더 결정 대기 |
| `ctxledger` 요약 `MaxBytes` (lane-20 `ctxledger`) | `workflow.agentic_factory.context.summary_max_bytes` | `.moai/config/sections/agentic-factory.yaml` | bytes | Issue 맥락 요약 본문의 상한. `<=0` = 상한 없음 | — | confirmed (leader decision) |
| `ghintake.GH.ListLimit` (lane-20 `ghintake`, 커밋 `af72c90458`; 예전 14a를 대신함) | `workflow.agentic_factory.intake.list_limit` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 필수, 정수 >= 1(리더 결정; lane-20 코드가 `0`을 거절하므로 읽기 코드가 읽어야 함). Issue 목록 호출 한 번이 읽는 최대 항목 수(조회·검색·중복 후보 모두 사용). `<=0`이면 호출을 거절(오류). 목록이 이 수에 닿으면 잘렸을 수 있다고 보고 오류로 보고한다("없음"·"중복 없음"으로 읽지 않음) → 운영 값은 실제 Issue 수보다 충분히 커야 한다. 페이지 나누기는 구현되지 않음 | — | confirmed (leader decision) |
| `adapt.GHReader.PRListLimit` (lane-20 `adapt`, 커밋 `af72c90458`; 예전 14b를 대신함) | `workflow.agentic_factory.monitor.pr_list_limit` | `.moai/config/sections/agentic-factory.yaml` | 정수 | 필수, 정수 >= 1(리더 결정; lane-20 코드가 `0`을 거절하므로 읽기 코드가 읽어야 함). PR 목록 호출 한 번이 읽는 최대 항목 수(열린 PR 수 세기와 `ReadPR`이 사용). `<=0`이면 거절. 이 수에 닿으면 오류 | — | confirmed (leader decision) |

- 예전 14c(`docstore.modifyRetries = 10`) 행은 지웠다: docstore 재시도 상수가 제거됐고, CAS 재시도 한도는 `claim.UpdateDoc`의 것이다(lane-20 `af72c90458`, 리더 전달, 이 작성자는 다시 읽지 않음).
- **구현 관찰(요구사항 아님)**: 이 표의 읽기 코드 `internal/config/agentic_factory_settings.go`(시험 포함, 작성자 보고로 mutant 20개 제거)가 커밋 `dc1bff9643`(브랜치 `WT-agentic-factory-settings-keys`)에 있다. 리더 전달이며 이 SPEC 작성자는 다시 읽지 않았다. SPEC이 정하지 않은 동작을 구현이 고른 곳은 spec.md §H 0.6.7에 열린 관찰로 적었다.

- effect 층 쓰기 effect 시간 제한 여섯 개(`IssueCreate` 등)의 키는 이 표에 없으며 열린 결정이다(spec.md §G.1 OD-20). 그 값은 상주 연결의 코드 상수로 정해졌다(spec.md §C.7 DEC-AF-26) — 설정 키가 아니므로 이 표에 행을 더하지 않는다.
- 상대 경로 `claim_store.dir`의 기준(spec.md OD-21)과 webhook 수신기 listen 주소 키(OD-22)는 열린 결정이며, 이 표에 행을 더하지 않는다.
- "AF가 설정되었다" 판정(OD-14)과 claim 저장소 디렉터리(OD-13)는 DEC-AF-25 D3·D2로 정해졌다(위 규칙, 표 첫 두 행). 남은 것은 이 파일의 template 배포(OD-19)다.
- 범위 검사가 코드에 없는 칸(`forge.Adaptive` 세 필드)은 "필요한 규칙"을 적었을 뿐 이 SPEC이 그 규칙을 채택한 것은 아니다. 키 이름은 확정됐지만, 그 검증 규칙의 채택은 리더 결정을 기다린다(OD-17).

## §7 역할 (REQ-AF-010)

| 역할 | 하는 일 | 쓸 수 있는 것 | 쓸 수 없는 것 |
|---|---|---|---|
| `ControlService` | claim, lease·활성 확인, epoch, reconcile, 재개 판단, 역할 부여(하위 SPEC). LLM 없음 | control ref, Issue 보고 쓰기(§4.1), 전이 | 코드, SPEC 본문 |
| `Coordinator`(표시 `leader`) | 우선순위·다음 행동 제안, 사용자 대화, PR 인계 수락, 변경 알림, 실패 복구, 배정 조정, 통합 지원(0.6.7 리더 결정, DEC-AF-23) | 자기 판단 기록 | 상태 직접 쓰기. 자기가 직접 구현한 변경의 리뷰·감사 증거. 배정된 agent의 구현·수리·리뷰 대응을 대신하는 것 |
| `agent` | Issue 하나의 plan·SPEC·run·sync·PR, 수리. PR 인계 뒤에도 CI 실패 수리·리뷰 대응·병합·Issue 종료 확인까지 책임(0.6.7) | 자기 attempt의 worktree, 시험 출력, writer 인계 증거(자기 epoch) | 다른 attempt, control ref, 역할 부여 요청, 같은 Issue의 리뷰·감사·QA 증거 |
| 리뷰·감사·QA 세션 | 같은 HEAD의 판정 | 부여받은 역할 하나의 증거 | 코드, 다른 역할 증거 |
| 준비 subagent | 접수·조회·준비 결과를 영속 receipt로 반환하고 종료 | 준비 grant 범위(branch·worktree 준비) | 제품 코드·SPEC, READY 전 개발 launch |

독립성(정본 [0]§1, [2]§21): writer와 auditor는 다른 세션이다. 한 Issue의 writer였던 세션은 그 Issue의 리뷰·감사·QA 증거를 낼 수 없다. 다른 Issue의 단계별 역할은 허용한다(하위 SPEC REQ-AFI-005).

## §8 명령과 관찰 (REQ-AF-020·022)

| 목적 | 명령(제안) | 쓰기 | 계약 |
|---|---|---|---|
| 시작 | `moai factory start` | 쓰기 | ControlService를 띄움. 이미 떠 있으면 기존 것을 알림(0) |
| 상태 | `moai factory status [--json]` | 읽기 | 기존 `status` 출력은 그대로, 새 정보는 추가 필드로만(OQ-12) |
| 관찰 | `moai factory status --watch` | 읽기 | 실행 전후 control ref 순번과 트리 키 불변 |
| 다시 붙기 | `moai factory attach --run <id> [--control]` | 기본 읽기 | 후보 둘 이상이면 id 없이 붙지 않음(12) |
| 안전 재개 | `moai factory attach --run <id> --resume` | 쓰기 | 실제 실행 조회 → `dead` 또는 안전 정지만 epoch+1 |
| 보류·정지 | `moai factory hold` / `stop` | 쓰기 | PARKED `operator-hold`(이 사유의 채택 여부는 OD-12 결정 대기) / 안전 지점 기록 후 정지 |

### 8.1 관찰 보기 항목

| 항목 | 원천 |
|---|---|
| 상태·하위 단계·PARKED 사유와 재개 조건 | control ref |
| HEAD SHA, attempt 순번, epoch | control ref |
| 필수 검사 결과 | GitHub 직접 조회 |
| 다음 행동 | control ref |
| 준비 상태(§4.2 7행)와 준비 큐 포화 | control ref |
| PR 적체(열린 인계 수 + 수리 적체 수, `wip_limit`) | control ref `PRQueue` |
| checkpoint 상태(§4.6) | control ref |
| 장부 접근성·STALE 수 | Issue context block |

항목마다 원천과 읽은 시각을 함께 보인다. 준비 상태와 PR 적체는 시작한 작업(BUILDING 이후)과 구분해서 보인다. 출력은 C1~C5 정제를 거친다.

**종료 코드**: 0 성공, 2 사용법 오류, 3 대상 없음, 10 다른 owner의 claim, 11 소유권 확인 실패, 12 후보 둘 이상, 13 PARKED(출력에 사유와 재개 조건), 14 원격 상태 미상, 15 오래된 epoch(fencing), 20 ControlService에 닿지 않음.

## §9 CLI (REQ-AF-022)

### 9.0 결정된 명령 표 (DEC-AF-09)

| 형태 | 처분 |
|---|---|
| `-f` / `--factory` | 유지. leader(Coordinator) 진입 |
| `--leader <name>`(인자형) | 유지 |
| `-a` / `--agent` | agent 진입. lane-22 특성화 시험이 argv 수준 전달을 확인(커밋 `3c7db03d7b`, 리더 보고. Claude 쪽 효과는 단언하지 않음) |
| `-w` / `--worktree` | 유지 |
| `-l/--leader` 제안 | 폐기 |
| `-l/--lane` | 장기 처분 미결(OD-2) |

**충돌 미해결(OD-2)**: 버전 선언 방식, 새 문법의 대상 leader 플래그 이름, `-l/--lane` 장기 처분, 기본 문법 전환 시점. 어느 답이든 다음을 갖춰야 한다:

1. 새 문법은 명시적 버전 선언이 있을 때만 해석한다.
2. 선언 없이 새·옛 문법을 섞으면 호환 오류(종료 코드 2)로 거절한다.
3. 회귀 시험: 어떤 인자도 세션 역할을 뒤바꾸지 않음(역할 뒤바뀜 0), 모르는 인자를 조용히 흡수하지 않음(조용한 흡수 0).
4. `--` 뒤의 인자는 launcher가 해석하지 않고 그대로 넘긴다.

### 9.1 현행 진입 명령 — 확인한 사실 (0.5.0에서 이어짐)

| 사실 | 근거(파일:줄) |
|---|---|
| `-f`/`--factory`는 factory run의 LEADER로 진입 | `internal/cli/cc.go:58` |
| `-l`/`--lane`은 다음 빈 `lane-<n>`으로 합류, 인자 없음 | `internal/cli/cc.go:65-66`, `:106`; `internal/cli/factory.go:98` |
| `-f`와 `-l`을 함께 주면 거절 | `internal/cli/factory.go:100`, `:271-272` |
| `--leader`와 `--factory-run`을 함께 주면 거절 | `internal/cli/factory.go:280-282` |
| `--leader`는 `-l`과만 함께 | `internal/cli/factory.go:102`, `:284-286` |
| 같은 `--leader`를 두 번 주면 나중 값이 이김(읽은 범위) | `internal/cli/factory.go:153-165`, `:174-189` — 반복 검사 확인 못 함(Gap) |
| `worker-<n>`·`agent-<n>`·bare `worker`/`agent`는 레거시로 거절 | `internal/cli/factory.go:332-347`, `:259-263`; `internal/factory/bootstrap.go:237-242` |
| `lead`·`lead-<suffix>` 리더 이름도 레거시로 거절 | `internal/cli/factory.go:184-186`, `:333-336` |
| `cc -l` 리더 검증 위치 | `internal/discovery/factory_discovery.go` `verifyCandidate` |

`-a/--agent` 도입은 `agent-<n>` 레거시 거절과 부딪힌다. 이름 값 `agent-<n>` 거절은 유지하고 플래그 `-a`는 별개 경로로 둔다(OD-2와 함께 확인).

### 9.2 용어 대응

| 기존 | 새 용어 |
|---|---|
| leader (`moai cc -f`) | `Coordinator`(표시 `leader`) + 상태 권한은 `ControlService` |
| `lane-<n>` (`moai cc -l`) | `agent`. 기존 lane 주소·세션은 즉시 바꾸지 않음 |
| 자원 slot lease (`moai slot`) | 그대로(자원 잠금). claim과 다름 |
| card (`moai factory next/stage/complete`) | Issue. Phase 2 이관 전까지 카드는 기존대로 |
| 카드 lease (`moai factory renew`) | attempt의 활성 힌트 + epoch(권위) |

### 9.3 명령 이관표 (lane-22 C1~C11, 0.5.0 그대로 요약)

| # | 구명령 | 새 명령(미확정) | 비고 |
|---|---|---|---|
| C1 | `moai cc -f` / `moai glm -f` | leader 진입 별칭 유지 | GLM은 변경 대상에서 뺌 |
| C2 | `moai cc -l` / `moai glm -l` | agent 합류 별칭 유지 | lane 레이블·기존 세션 보존 |
| C3 | `moai cc -w [name]` | 유지 | #1847·#1813과 얽힘 |
| C4 | `moai factory next` | Issue claim | `factory_card.go:2324` |
| C5 | `moai factory stage` | ControlService 전이 | `factory_card.go:2728` |
| C6 | `moai factory complete` | 병합 게이트 + 원격 확인 | `factory_card.go:2938` |
| C7 | `moai factory renew` | attempt 활성 힌트 갱신 | `factory_card.go:2840` |
| C8 | `moai factory approve` / `decide` | leader 수락 receipt / 운영자 증거 | `leader_approval_gate.go:286`, `factory_card.go:4335` |
| C9 | `status` / `runs` 등 | 미정 | 조회 계열은 제거 순서 마지막. `status` 이름 충돌(OQ-12) |
| C10 | `moai integration …` | 미정 | 대체 검증 전 등록 해제 금지 |
| C11 | `moai todo …` | Issue 원장(Phase 2) | 지금 손대지 않음 |

### 9.4 단계 폐기 경로

| 단계 | 의미 | 다음 단계 조건 | 증거 |
|---|---|---|---|
| S1 별칭 유지 | 구명령이 새 명령과 나란히 같은 동작. 도움말·거절 문구·출력·종료 코드가 §9.6 golden과 바이트 동일 | 새 명령 등록 | AC-AF-021 호환 시험 |
| S2 폐기 예고 | stderr 한 줄만 추가. 나머지 바이트 동일 | 예고 출력 시험 | AC-AF-021 S2 사례 |
| S3 대체 검증 | 새 명령이 모든 동작·거절 조건을 대체함을 같은 HEAD 시험과 독립 감사로 확인 | 무부채 PASS + 인계 | 판정 파일 |
| S4 제거 | 구명령 등록 해제 | S3 증거 + 운영자 승인 | 운영자 승인 기록(형식 OQ-20) |

불변: 기존 세션 id, 권한, lane 레이블, owner, handle, 열린 임대, 리더 선출 상태를 바꾸지 않는다. 구명령을 즉시 제거하지 않는다. 실행 중인 세션을 재시작하지 않는다.

### 9.5 열린 PR 착지 순서 후보 (lane-22 분석, 결정은 리더)

#1944 → #1937 → #1943 → #1916 ↔ #1921(직렬). 파일 겹침만으로 도출한 후보이며 이 SPEC은 PR 상태를 다시 재지 않았다.

### 9.6 호환 기준 golden 파일

- 위치(제안): `internal/cli/testdata/agentic_factory_compat/`(0.2.0 시점 없음).
- 내용: 변경 전 빌드로 기록한 `moai cc --help`의 진입 플래그 도움말, 거절 문구와 종료 코드, 기존 factory 기록의 lane 레이블·세션 id·owner·handle·권한 모드 스냅숏.
- 순서: M6의 **첫 커밋**이 golden 파일만 담는다. S1 등록은 그 다음 커밋부터(VCI §2.3, 커밋 그래프로 증명).

### 9.7 Codex 기본값 전환 단계 기록

run 단계는 `progress.md`에 `## §J Codex Migration Stage Record` 절을 만들고, 첫 줄에 `run_base_sha: <전체 SHA>`를 적은 뒤 plan.md §F.3의 세 단계마다 한 행(`| CX-1 | <stage_sha> | <evidence> |`, 백틱 없이)을 적는다.

| 행 | 단계 | `stage_sha` | 필수 증거 |
|---|---|---|---|
| CX-1 | 관문 도입 | 관문 커밋 | — |
| CX-2 | 미설정 0건 검증 | AC-AF-015 PASS 기록 커밋 | PASS 출력 증거 경로와 HEAD |
| CX-3 | 기본값 변경 제안 | 제안 커밋 | 제안 문서 경로 |

순서는 커밋 그래프로 증명한다: CX-1 ⊂ CX-2 ⊂ CX-3(엄격한 조상). `run_base_sha` 다음 커밋부터 CX-2 커밋까지 모든 커밋에서 Codex 기본값의 (있음/없음, 값) 짝이 `run_base_sha`와 같아야 한다. 대상: 로컬·템플릿 `workflow.yaml`의 `codex.review_gate.enabled`·`codex.task.allow_write`·`audit.model`·`audit.gates.codex`, `internal/config/defaults.go`의 `Codex.ReviewGate.Enabled`·`Codex.Task.AllowWrite`, 로컬·템플릿 `llm.yaml`의 `llm.defaults.sessions.codex.leader`·`.lane`의 `model`·`effort`. 파싱 실패는 FAIL이다. 이 절은 `§E.*` 이름 공간을 쓰지 않는다.

## §10 레인 소유 (정본 [6] DELTA4, DEC-AF-22)

| 레인 | 소유 | REQ |
|---|---|---|
| lane-17 | 설정 읽기·검증(`internal/config`) | REQ-AF-021, REQ-AF-003의 읽기 쪽 |
| lane-18 | core: 상태 엔진, claim·fencing, control ref git 구현, 접수 상태, 사유 타입, `core.Collaboration.Authorize`, checkpoint 상태·git·manifest, 잠금 상수와 해제(M7) | REQ-AF-001·002·004·006·009·011·012·016·017·018 |
| lane-19 | provider 템플릿(참조만, 다른 worktree) | REQ-AF-023 |
| lane-20 | effect 경계: 외부 효과 실행, 보호 약화 거절 집행, Issue 쓰기, checkpoint Issue 링크, 관찰 보기 | REQ-AF-005·008·013·014·019·020 |
| lane-21 | 브라우저 어댑터 | REQ-AF-021 브라우저 쪽 |
| lane-22 | 시험: e2e, 전용 시험, 시나리오 묶음, 충돌 회귀, 측정 | REQ-AF-024·025, 모든 AC의 시험 |
| launcher 담당 | CLI | REQ-AF-022 |
| (공통) | 병합 게이트 | REQ-AF-007·010은 lane-18(판정) + lane-20(effect) |

REQ-AF-015(공급자 관문)는 lane-17(키 읽기)과 lane-20(호출 경계)이 나눠 갖는다.

## §11 병합 게이트 (REQ-AF-007·010·011)

통과로 세는 것(모두 같은 후보 HEAD):

1. 필수 검사(`Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`)가 그 HEAD에서 실제로 실행되어 성공. 보호 규칙의 `strict: false` 때문에 게이트가 따로 HEAD를 대조한다.
2. 코드 리뷰와 보안 리뷰(§11.1) — 각각 §2.1b에서 "통과로 셈".
3. 필수 독립 감사의 무부채 PASS.
4. 공식 QA 스킬 PASS(spec.md §F.5).
5. 유효한 PR 인계 receipt(E03): writer 인계(현재 epoch의 agent) + leader 수락 — 둘 다 귀속 조회를 거친다.

**증거 종류별 필요한 역할**

| 증거 종류 | 필요한 역할(귀속 조회 결과) | 추가 조건 |
|---|---|---|
| `code-review` | 코드 리뷰 역할, 또는 운영자 기록(§11.1 (b)) | 그 Issue의 writer였던 세션이 아님. GLM 아님 |
| `security-review` | 보안 리뷰 역할, 또는 운영자 기록 | 위와 같음 |
| `audit` | 감사 역할 | 그 Issue의 writer·리뷰·QA 세션이 아님 |
| `qa` | QA 역할 | 그 Issue의 writer였던 세션이 아님 |
| `writer-handoff` | 그 attempt의 agent | 현재 epoch(fencing이 먼저) |
| `coordinator-acceptance` | 등록된 Coordinator 세션(부여 대상 아님, 하위 SPEC) | 그 변경을 직접 구현하지 않았음 |

### 11.1 공급자 중립 리뷰 경로 (D2)

- (a) 리뷰 세션: 하위 SPEC 부여로 코드 리뷰 또는 보안 리뷰 역할을 받은 세션이, writer와 다르고 GLM·Codex가 아닌 공급자로 리뷰. 공급자 선택은 운영자 설정.
- (b) 운영자 기록 리뷰: 운영자가 ControlService의 운영자 증거 경로로 기록. 하위 SPEC의 운영자 기록 확인을 거쳐야 한다. 잠금 중에는 이 경로도 효력이 없다.
- 둘 다 없으면 병합은 `not-configured`. Codex 명시 설정 프로젝트는 Codex 리뷰를 (a)와 같은 결속 조건으로 쓸 수 있다.

### 11.2 잠금 (DEC-AF-02, DEC-AF-04)

| 항목 | 계약 |
|---|---|
| 위치 | 제품 코드의 컴파일 시점 상수 하나(이름은 lane-18이 정함). 귀속 원천은 그 상수 하나로만 동작을 고른다 |
| 입력 | 설정 키, 환경 변수, 파일, 플래그 어느 것도 읽지 않는다 |
| 잠긴 동작 | 모든 증거 종류(§11 표 6종)에 `unattributable{identity-spec-incomplete}` → 게이트 `producer-not-allowed` |
| 덮는 범위 | 시나리오 하네스 포함 모든 호출자. 귀속 대역은 게이트 단위 시험 안에서만 |
| 시나리오 하네스 | 시작 때 귀속 원천 종류를 기록. 제품 원천이 아니면 FAIL(대역 사용 금지) |
| 해제 | 그 상수 한 곳을 바꾸는 코드 변경 하나. 하위 SPEC `completed` + 하위 AC 전부 PASS + 하위 OQ-AFI-1 운영자 위험 수용 기록과 같은 착지(plan M7, lane-18) |
| 해제 때 함께 | AC-AF-011 L1·L2를 "제품 원천이 등록된 증거를 귀속한다"로 다시 쓴다 |

### 11.3 보호 약화 거절 (REQ-AF-008, 0.5.0 REQ-AF-019, DEC-AF-22)

effect 경계(lane-20)가 요청을 보내기 전에 거절한다. 사유 타입은 lane-18, 시험은 lane-22.

| # | 약화 연산 | 거절 사유 |
|---|---|---|
| W1 | 필수 검사 삭제 | `permission-denied` |
| W2 | 브랜치 보호 약화(`enforce_admins` 끄기, 필수 리뷰 줄이기 등) | `permission-denied` |
| W3 | 선형 이력 설정 변경 | `permission-denied` |
| W4 | `main` 강제 push | `permission-denied` |
| W5 | 보호 브랜치 삭제 | `permission-denied` |
| W6 | 관리자 우회 병합 | `permission-denied` |
| W7 | 레이블·설정 endpoint로 보호 규칙 변경 | `permission-denied` |

거절은 상태를 바꾸지 않고 전송 0이다. 일곱 연산은 각각 사유 하나로만 기록된다.

## §12 설계 입력 다섯 건 → 대체 동작 → 재현 시험 (0.5.0 그대로)

| Issue | 관측 | 대체 동작 | 재현 시험 |
|---|---|---|---|
| #1938 | 죽은 run에 남은 카드를 재발급할 정식 명령 없음 | Issue당 claim 하나, 죽은 owner의 만료 attempt는 같은 claim 아래 새 attempt | `TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim` |
| #1926 | `--spec` 없는 카드가 kickoff에서 거절, 해시 정의 불일치 | 전이 guard는 §2.2 증거로 판단. 해시는 한 정의만 | `TestState_SpeclessIssue_TransitionsWithoutSpecHash` |
| #1924 | 운영자 표식을 찍는 코드 없음 | 운영자 결정 권한은 운영자 증거로만. LLM 세션에 자동 표식 없음 | `TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused` |
| #1847 | launcher가 `-w -`와 `-w feature/foo`의 실제 대상을 놓침 | worktree는 정규화된 절대 경로로 claim에 기록 | `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch` |
| #1813 | 같은 이름의 다른 트리 생성 | 새 attempt는 claim이 가리키는 트리 재사용 | `TestClaim_SameNameTree_NoDivergentTreeCreated` |

이 Issue들은 이 SPEC이 닫지 않는다.

## §13 (정리됨)

0.5.0의 "상위 설계 권고" 절은 정본과 리더 결정으로 대체되었다(§A.4 명칭, §8 명령). 옛 내용은 커밋 `3275706cc9`의 이 파일 §13에 있다.

## §14 카드 TTL/renew 대 지속 배정 + epoch (0.5.0 판정 유지)

- 카드 TTL(15m)·renew(5m)는 활성 힌트로 읽는다. 기한이 지나면 ControlService가 활성 확인을 시작한다. 값과 범위 검사는 그대로.
- 권위는 배정 epoch. 모든 쓰기는 epoch를 싣고, 현재보다 작으면 거절(종료 코드 15).
- 인계는 ControlService만, `ClassifyOwnerWith`가 `dead`이거나 안전 정지 기록 뒤, epoch+1로. `indeterminate`는 인계하지 않음.
- 자원 slot lease의 "기한 지나면 인계" 규칙과는 충돌하므로 claim을 slot lease로 만들지 않는다.
- 기한이 지난 뒤 살아 있는 owner의 renew 처리는 OQ-11. 답이 오기 전에는 현행 거절 유지.

## §15 시험 manifest (DEC-AF-11)

규칙: 시험마다 **주 AC 하나**. 보조 근거는 별도 열이며 그 AC의 충족 셈에 넣지 않는다. 정본 시험은 manifest 실행기의 하위 시험 `TestAcceptanceManifest/<정본 ID>`로 돌리는 것을 제안한다(이름은 lane-22가 정함). AC 판정은 하위 시험마다 `--- PASS: TestAcceptanceManifest/<ID>` 줄이 있고 `[no tests to run]`이 없으며, PASS 줄 수가 아래 "건수"와 같을 때다.

### 15.1 정본 시험 배정 (165건)

| AC | 주 시험(정본 ID) | 건수 | 보조 근거(셈 제외) |
|---|---|---|---|
| AC-AF-001 | M01, M02, M03, M04, M05 | 5 | SYNC03 |
| AC-AF-002 | SYNC12, M47 | 2 | M03, M09 |
| AC-AF-003 | M13, M15 | 2 | M23 |
| AC-AF-004 | M06, M07, M08, M09, M10, M11, M12 | 7 | M29 |
| AC-AF-005 | M18, M19, M20, M29, M30, M31, SYNC03, SYNC04, SYNC05, SYNC11 | 10 | PREP02, CONTEXT05 |
| AC-AF-006 | M33, M34, M35, M36 | 4 | — |
| AC-AF-007 | M14, M24, M25, M26, M27, M32, M37, M55 | 8 | HANDOFF01~08 |
| AC-AF-008 | (없음) | 0 | M25 |
| AC-AF-009 | HANDOFF01~08 | 8 | M33 |
| AC-AF-010 | M22, M23 | 2 | M55 |
| AC-AF-011 | (없음) | 0 | M23 |
| AC-AF-012 | PREP01~10 | 10 | M01 |
| AC-AF-013 | M39, M40, CONTEXT04 | 3 | — |
| AC-AF-014 | M17, M28, M43, M44, GH05, GH06, SYNC09, SYNC10, WT15, SYNC13, SYNC14, SYNC15, SYNC16, SYNC17, SYNC18 | 15 | EFF06 |
| AC-AF-015 | M21 | 1 | M22 |
| AC-AF-016 | M16, M41, M42, M48, M54, M56 | 6 | 셀이 붙은 M 시험 전부 |
| AC-AF-017 | M45, M49, M50, M51, M52, M53, SYNC01, SYNC02, SYNC06, SYNC07, SYNC08 | 11 | M47 |
| AC-AF-018 | CONTEXT01, CONTEXT02, CONTEXT03, CONTEXT05, CONTEXT06 | 5 | PREP 전부 |
| AC-AF-019 | (없음) | 0 | PREP09, M28 |
| AC-AF-020 | M38, M46, M46a~M46g | 9 | — |
| AC-AF-021 | CLI01~CLI23 | 23 | — |
| AC-AF-022 | WT01~WT14, WT16, GH01~GH04, GH07~GH15 | 28 | — |
| AC-AF-023 | (없음) | 0 | — |
| AC-AF-024 | EFF01~EFF06 | 6 | — |
| AC-AF-025 | (없음) | 0 | M48 |
| | **합계** | **165** | |

- 셈: M01~M56(56) + M46a~g(7) + CLI01~23(23) + SYNC01~18(18) + WT01~16(16) + GH01~15(15) + HANDOFF01~08(8) + PREP01~10(10) + CONTEXT01~06(6) + EFF01~06(6) = 165. 손으로 대조했다(추정이 아닌 표 합계이지만 기계 대조는 하지 않았다).
- 정본 주 시험이 0인 AC: AC-AF-008, 011, 019, 023, 025. 이 다섯은 아래 전용 시험이 주 시험이다.

### 15.2 전용 시험 (정본 밖, 주 AC 하나씩)

| AC | 전용 시험 | 상태 |
|---|---|---|
| AC-AF-008 | `TestProtectionGuard_*` 3개(lane-22 작성 보고) + 약화 연산 7개 거절 시험(이름 lane-22 결정, 제안: `TestProtectionGuard_RefusesEveryWeakeningOperation`) | 앞 3개 작성됨(미검증), 거절 시험 NOT_RUN(거절 경로 없음) |
| AC-AF-011 | 잠금 L1~L4, P4a·P4b(제안: `TestGate_RoleEvidence_RequiresChildAttribution_AndLockHolds`) | NOT_RUN |
| AC-AF-019 | 관찰 보기 준비 상태·PR 적체 구분(제안: `TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly`) | NOT_RUN |
| AC-AF-023 | 환경 행렬(제안: `TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess`) | NOT_RUN |
| AC-AF-025 | `TestScenarioSuite`(af_e2e, 실제 Issue). 하위 `07_RemoteIntegrated_Close_Then_Done_NextDispatch` 등 | NOT_RUN(운영자 승인 필요) |
| AC-AF-003 | Issue 쓰기 판정 사례 IA1~IA16(제안: `TestIssueAuthority_*`) | NOT_RUN |
| AC-AF-002 | checkpoint 사례 CK1~CK10(제안: `TestCheckpoint_*`) | NOT_RUN |
| AC-AF-013 | `TestRedact_*` 13개(design-impact L.9 목록) | NOT_RUN |
| AC-AF-014 | 응답 분류 G1~G8·L1~L6(제안: `TestForgeClassify_*`) | NOT_RUN |
| AC-AF-006 | §12 재현 시험 중 해당 없음. 시험 이름 미전달인 lane-18 종결 순서 시험은 보조 | — |
| AC-AF-001·016 | §12 재현 시험 5개는 각각 AC-AF-001(#1938·#1847·#1813) 또는 AC-AF-016(#1926·#1924)의 전용 시험 | NOT_RUN |

전용 시험 이름은 lane-22가 정한다. 위의 "제안"은 이름이 정해지기 전까지의 자리 표시이며, 정해진 이름이 오면 이 표를 고친다(lane-22는 제안 이름을 쓰지 않았다고 보고했다).

## §16 provider template 참조 (REQ-AF-023, [4] DELTA2, [5] DELTA3)

| 항목 | 이 SPEC의 처분 |
|---|---|
| Issue·PR template | 다른 worktree(`af-templates`)의 작업. 이 SPEC은 참조만 |
| 기존 template 파일과 `moai update` (알려진 동작, DEC-AF-24) | 프로젝트에 provider template 파일이 이미 있는데 manifest에 없으면 `moai update`는 그 파일을 `user_created`로 기록하고 건너뛴다. 그 파일은 뒤의 template 개정을 받지 못한다. 사용자 파일 보존이 우선하며 결함이 아니다. 근거: 브랜치 `WT-agentic-factory-templates` 커밋 `8ebbacad30`(`internal/template/user_owned_templates.go` `isUserOwnedTemplatePath`, `deployer.go` 한 줄), 보호 시험 승격 `47d8b782b7`. 출처: 'leader decision, relayed via the leader session'. 이 실행은 두 커밋을 읽지 않았다 |
| 계층·분류(tier·epic·sub-issue) | 공급자가 지원하지 않으면 `not supported`로 보고. 흉내 내지 않음 |
| 초기 setup | DELTA3 setup 적용은 운영자 결정 |
| 댓글 묶음 | 급하지 않은 진척 댓글은 `comments.batch_window` 안에서 묶음. 보안 실패·승인·실제 차단은 지연 없이 |
| Pele SVG | 생성하지 않았다(기록된 공백) |
