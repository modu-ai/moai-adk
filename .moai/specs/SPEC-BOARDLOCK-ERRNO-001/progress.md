# progress.md — SPEC-BOARDLOCK-ERRNO-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-31
card: t379
worktree: .claude/worktrees/t379
branch: WT-boardlock-errno
base: 3f03d9c36 (= origin/develop)
tier: S
plan_phase_evidence: ".moai/reports/t379/errno-probe-source.go.txt, .moai/reports/t379/errno-probe-output.txt (darwin/arm64; CI is ubuntu)"
measured_misclassification_count_at_call_site: 0

## §E.2 Run-phase Evidence

**측정 환경**: darwin/arm64, worktree `.claude/worktrees/t379`, 브랜치 `WT-boardlock-errno`.
run 진입 HEAD `9328a5242`(`origin/develop` 로 fast-forward, divergence `0 0`). M1 커밋 `364bc332f`.
아래 모든 관측은 **이 실행, 이 트리**에서 났다. 리눅스 판정은 CI(ubuntu) 몫이며 여기에 없다.

### E.2.0 이 카드가 한 일 — 그리고 하지 않은 일

**실측 오분류는 0건이다.** 계획 단계 프로브가 유도한 세 케이스 중 실호출 경로로 도달한 errno 는
`EWOULDBLOCK`/`EAGAIN` 하나뿐이었고 그것은 이미 올바르게 분류되고 있었다. `EBADF` 는 오분류되지만
호출 지점이 방금 성공한 `unix.Open` 의 디스크립터를 flock 하므로 도달 불가이고, `EINVAL` 은
darwin/arm64 에서 오류 자체를 내지 않았다.

따라서 이 카드는 **방어적 좁히기 + 회귀 잠금**이며, **오늘 관측 가능한 행동 변화는 0** 이다(측정된
도달 가능 입력 한정). 살아 있는 결함의 수리가 아니다. 이 문서·코드 주석·커밋 메시지 어디에도
그 반대로 읽히는 문장을 쓰지 않았다.

### E.2.1 M1 설계 결정 — 옵션 A 채택

`plan.md §F M1` 이 남겨 둔 유일한 열린 결정. **A(순수 분류 함수)를 채택했다.**

```go
func classifyBoardFlockErr(err error, lockPath string) error
```

`acquireBoardLockImpl` 이 실패 분기에서 이 함수의 결과를 반환한다.

**A 를 고른 이유** — B(`boardFlock = unix.Flock` 패키지 변수 seam)와 견주어:

1. **합성 errno 를 술어에 직접 먹일 수 있다.** AC-BLE-001b·002 가 `unix.Flock` 을 가로채지 않고도
   `ENOLCK`/`EBADF`/`EOPNOTSUPP`/`EINTR` 네 입력을 그대로 분류시킨다. B 는 같은 것을 재기 위해
   syscall 호출 전체를 우회해야 한다.
2. **프로덕션 표면이 늘지 않는다.** B 는 테스트 전용 간접층을 잠금 축(lock axis) 한복판에 남긴다 —
   교체 가능한 패키지 변수는 병렬 테스트에서 그 자체가 경합 대상이 된다.
3. **A 의 유일한 위험(호출 지점 미배선)에 검출기가 있다.** 분류 함수는 옳은데 호출 지점이 그것을
   쓰지 않아도 AC-BLE-001a 는 통과할 수 있다 — 경합은 어느 쪽이든 held 이므로. 이 구멍을
   **M-narrow 뮤턴트가 실제로 잡는다**(E.2.5 에서 관측). 위험이 이론상 남는 것이 아니라 잰 것이다.

**결정 순서를 지켰다** — 이 결정을 먼저 확정하고 그 위에서 테스트를 썼다. 감사가 지적한
iter-2 회귀는 순서를 뒤집어 생겼다.

### E.2.2 부채 2 종결 — AC-BLE-003 의 `/dev/fd` 기제를 교체했다

`plan.md §B.1` 부채-2 의 세 구멍을 **함께** 닫았다. 셋 다 같은 뿌리에서 나왔으므로 기제 자체를
바꾸는 것이 개별 땜질보다 짧다.

| 구멍 | 종전 상태 | 처리 |
|---|---|---|
| (a) "고정 여유"에 값이 없다 | 임계값 미정 — 구현자가 정하게 됨 | `slack = 16` 으로 못박음(`attempts = 200` 의 8%). 누수가 시도의 10분의 1(20)만 일어나도 잡힌다 |
| (b) 플랫폼 주장이 미측정 | "darwin·리눅스 모두 `/dev/fd` 가 프로세스 자신의 fd 를 노출한다" — **어느 쪽에서도 측정 안 됨**, 리눅스는 `/proc` 경유라 환경 의존, **CI 는 ubuntu** | **기제를 교체했다.** 측정할 수 없는 주장을 재는 대신, 잴 필요가 없는 계약으로 갈아탔다 |
| (c) 빈 스윕 절이 없다 | `/dev/fd` 를 못 읽는 러너에서 조용히 `ok` | 유도 실패 횟수를 세어 `induced != attempts` 면 실패. 프로브 open 실패도 `t.Fatalf` |

**교체한 기제**: POSIX `open(2)` 의 **"현재 열려 있지 않은 가장 낮은 번호의 디스크립터를 반환한다"**
계약. 실패 획득 N회 전후로 프로브 디스크립터를 하나 열어 그 **번호**를 비교한다. 누수가 없으면
같은 번호가 나오고, 시도마다 하나씩 샜다면 약 N 만큼 큰 번호가 나온다.

이 교체가 (b)를 닫는 방식이 요점이다 — `/dev/fd` 가용성은 **환경에 대한 추론**이라 측정 없이는
주장할 수 없지만, `open(2)` 의 최저번호 할당은 **문서화된 syscall 계약**이고 두 플랫폼에서 동일하며
읽을 수 있는 파일시스템을 요구하지 않는다. 추론을 측정으로 바꾼 것이 아니라, **추론이 필요 없는
근거로 옮긴 것이다.**

그리고 이 기제는 실제로 작동한다 — E.2.5 의 M-leak 관측에서 `fd 6 → 206`, 정확히 200개 누수를
잡았다. 기제가 무언가를 보고 있다는 것이 추정이 아니라 관측이다.

### E.2.3 부채 3 종결 — M-leak 을 실제로 지은 모양에 맞춰 다시 규정했다

`plan.md §B.1` 부채-1: M-leak 이 "**비경합 분기**에서 `unix.Close(fd)` 제거"로 정의돼 있는데
`§F M1` 은 close 를 분기 앞의 **공유 1회 호출**로 규정했으므로 심을 자리가 없다 — 맞는 지적이다.

**M1 에 위임된 선택지 (a)를 택했다: M-leak 을 공유 close 대상으로 다시 규정한다.**

- **재규정**: `acquireBoardLockImpl` 의 실패 분기에서 **그 단일 `unix.Close(fd)` 를 제거**한다.
- **왜 이것이 발화하는가**: 그 close 는 경합 경로와 비경합 경로가 **함께 쓰는 유일한 호출**이다.
  따라서 경합만 유도하는 AC-BLE-003 의 인프로세스 루프로도 그것을 지운 뮤턴트가 잡힌다.
- **관측**: E.2.5 — AC-BLE-003 RED, `probe fd 6 before, 206 after 200 failed acquisitions`.

**정확히 말해 이 가드가 덮는 범위**: 공유 close **지점**이다. 비경합 반환 경로를 독립적으로
지나가지는 않는다 — 그 경로는 실호출로 도달 불가이므로 유도할 수 없다. 그러나 REQ-BLE-004 가
"어떤 errno 든 반환 전에 닫는다"를 요구하고 구현이 **한 지점**으로 그것을 만족시키므로, 그 지점을
지우는 뮤턴트가 잡히면 요구는 비공허하게 잠긴다. 두 분기를 각각 유도해야만 가드가 성립한다는
읽기는 구현이 분기별 close 를 가질 때만 참이고, 이 구현은 그렇지 않다.

**결과**: AC-BLE-003 은 더 이상 "비공허성 가드 없음"이 아니다. `verification-completeness.md` §2 의
mutant probe — 요구를 위반하면서 기준을 만족하는 뮤턴트를 쓸 수 있는가 — 에 대해, 재규정된 M-leak 이
**기준을 RED 로 만들므로 그런 뮤턴트가 아니다.** 감사가 P8 에서 "채택 불가(too shallow to adopt)"로
격상하라던 상태는 해소됐다. 다만 그 해소는 **재규정 뒤**에 성립하는 것이고, 종전 정의 아래에서
감사의 판정이 옳았다는 사실은 남는다.

### E.2.4 EINTR — 유도하지 않았다 (명시)

**커널 유도를 시도하지 않았다.** `spec.md §4` 가 `ENOLCK`/`EOPNOTSUPP`/`EINTR` 의 실측 유도를 명시적
범위 밖으로 두었고, 유도용 시그널 구성은 그 범위를 넘는다.

대신 **합성 errno 로 분류만 잠갔다** — `unix.EINTR` 을 AC-BLE-001b·002 의 입력 표에 한 행으로 넣었다.
이것은 유도가 아니라 분류 술어에 대한 단언이며, `spec.md §1.3.1` 이 받아들이는 변화로 기록한 결정
(EINTR = 비경합 = 즉시 하드 오류)을 **기계적으로 고정**한다. 이 행이 없으면 그 결정은 나중에 조용히
뒤집힐 수 있다.

**행동 변화의 성격**: 오늘 EINTR 은 `ErrBoardLockHeld` 로 매핑돼 `acquireBoardLockSerialized`
(`board_store.go`)의 경합 재시도 예산에 흡수된다. 좁힌 뒤에는 즉시 하드 오류다. **EINTR 이 실제로
도착하는 배포에서는 행동이 바뀐다.** 도달 가능성 논증(`LOCK_NB` 는 대기하지 않고 flock(2) 의 EINTR 은
대기 중 중단에 대한 것)은 **syscall 계약을 읽은 추론이지 측정이 아니다.**

따라서 이 문서의 "오늘 행동 변화 0" 은 **측정된 도달 가능 입력**(`EWOULDBLOCK`/`EAGAIN`) 한정이다.
미측정 errno 는 REQ-BLE-005 의 보호 범위 밖이다.

### E.2.5 AC 매트릭스 — 등급별 (E1)

**등급 구분은 `acceptance.md §D.1` 을 따른다**: 착지 차단 셋(001b·002·005)만 PASS 로 세고,
회귀-가드 셋(001a·003·004)은 `verification-completeness.md` §2.1 에 따라 **PASS 로 기록하지 않고
등급과 관측 결과로 보고한다**. 여섯 모두에 대해 같은 증거(명령 + 출력 전문 + 종료 코드 + 선택자
매치 수)를 제출한다.

#### RED-now 를 잰 트리

`acceptance.md §D.0` 이 M1 소관으로 남긴 "잴 수 있는 트리"를 실제로 지었다:
HEAD `9328a5242` + 새 계약 테스트 + **분류 seam 스텁(수리 이전 = 넓은 매핑)**.

| 항목 | 값 |
|---|---|
| 트리 SHA | **`9c196204c76b8f7ff2cba3873c7d21ca7c128017`** (`git write-tree`) |
| 부모 HEAD | `9328a5242` |
| 재현 diff | `.moai/reports/t379/red-now-tree.diff` |
| 재현 절차 | `9328a5242` 체크아웃 → `git apply` 위 diff → `git add` 두 파일 → `git write-tree` 가 같은 SHA |

**커밋 SHA 가 아니라 트리 SHA 인 이유**: `verification-completeness.md` §2.1 이 요구하는 네 번째
요소는 "the tree SHA the measurement was taken on" 이다. 알면서 적색인 커밋을 브랜치 이력에 남기는
것은 이 리포의 커밋 전 규율(`CLAUDE.local.md` §4)과 어긋나므로, `git write-tree` 로 **실재하고
주소지정 가능한 트리 객체**를 만들어 그것을 pin 했다. `git ls-tree 9c196204c` 로 검증 가능하다.
**부채**: 이 트리 객체는 어떤 ref 에서도 도달 불가라 gc 대상이다 — 그래서 diff 를 함께 남겼고,
diff 가 실제 복구 수단이다.

#### 착지 차단 셋

| AC | 등급 | 판정 | 명령 | 선택자 매치 | 출력 |
|---|---|---|---|---|---|
| AC-BLE-001b | 착지 차단 | **PASS** | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoNonContentionIsNotHeld$' -count=1 -v` | 1 test / 4 subtest | RED@`9c196204c` rc=1 → GREEN@`364bc332f` rc=0 |
| AC-BLE-002 | 착지 차단 | **PASS** | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoPreservesErrnoAndPath$' -count=1 -v` | 1 test / 4 subtest | RED@`9c196204c` rc=1 → GREEN@`364bc332f` rc=0 |
| AC-BLE-005 | 착지 차단 | **PASS** | 아래 뮤턴트 3종 표 | 3 뮤턴트 | 각 RED + 되돌림 후 GREEN |

RED-now 전문 (트리 `9c196204c`, rc=1):

- **AC-BLE-001b** — 네 하위 케이스 전부
  `board_lock_errno_test.go:80: IsBoardLockHeld(kanban board lock held) = true, want false`
  (`no_locks_available` / `bad_file_descriptor` / `operation_not_supported_on_socket` /
  `interrupted_system_call`), `--- FAIL` ×5, `FAIL github.com/modu-ai/moai-adk/internal/kanban 0.422s`
- **AC-BLE-002** — 네 하위 케이스 전부
  `board_lock_errno_test.go:100: errors.Is(kanban board lock held, <errno>) = false, want true`,
  `--- FAIL` ×5, `FAIL github.com/modu-ai/moai-adk/internal/kanban 0.247s`

**RED 의 이유가 명시적이다**: 넓은 매핑 스텁이 모든 errno 를 sentinel 로 보내기 때문이며,
`verification-completeness.md` §2 가 구분하라는 세 방향 중 **vacuous 도 impossible 도 wrong-reason 도
아니다** — M1 의 좁히기가 정확히 이것을 뒤집는다(GREEN 관측이 그것을 보인다).

#### 회귀-가드 셋 (PASS 로 세지 않는다)

| AC | 등급 | 관측 | 명령 | 선택자 매치 | 비공허성 공급원 |
|---|---|---|---|---|---|
| AC-BLE-001a | 회귀-가드 | **GREEN 유지** (수리 전·후 모두 초록) | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoContentionRemainsHeld$' -count=1 -v` | 1 test | **M-narrow** (관측됨 — 아래) |
| AC-BLE-003 | 회귀-가드 | **GREEN 유지** | `go test ./internal/kanban/ -run '^TestBoardFlockErrnoFailurePathClosesDescriptor$' -count=1 -v` | 1 test | **M-leak(재규정)** + M-narrow(전제 가드) — 부채 3 종결 |
| AC-BLE-004 | 회귀-가드 | **실패 집합 불변** | `go test ./internal/kanban/... -count=1` 수리 전/후 | 패키지 전체 | 뮤턴트 가드 없음 — 구조상 대조 게이트 |

- **AC-BLE-001a** `9c196204c`: `--- PASS: TestBoardFlockErrnoContentionRemainsHeld (0.00s)` / `ok ... 0.268s` (rc=0)
  `364bc332f`: `--- PASS ... (0.00s)` / `ok ... 0.361s` (rc=0)
- **AC-BLE-003** `9c196204c`: `--- PASS: TestBoardFlockErrnoFailurePathClosesDescriptor (0.01s)` / `ok ... 0.303s` (rc=0)
  `364bc332f`: `--- PASS ... (0.01s)` / `ok ... 0.361s` (rc=0)

**AC-BLE-004 — 수리 전/후 실패 집합 대조** (이 AC 의 전부):

| 시점 | 명령 | 출력 | rc |
|---|---|---|---|
| **수리 전** (HEAD `9328a5242`, 어떤 편집도 하기 전) | `go test ./internal/kanban/... -count=1` | `ok  github.com/modu-ai/moai-adk/internal/kanban 17.413s` | 0 |
| **수리 후** (M1 착지 + 뮤턴트 되돌림) | `go test ./internal/kanban/... -count=1` | `ok  github.com/modu-ai/moai-adk/internal/kanban 18.923s` | 0 |

실패 집합: 전후 **모두 공집합**. 달라지지 않았다. 소비자 3곳(`board_store.go:173`,
`integration_lock_mutation.go:103`, `backlog_store.go:736`)을 지나는 기존 경합 테스트들이 수정 없이
계속 통과한다. 기지 플레이크(`TestConcurrencyStress` 계열)는 이번 실행에서 **발화하지 않았다**.

**이 AC 의 공허화 조건**(감사 P3 이 "유일한 조건"이라 적힌 것을 지적했다 — 열거를 넓힌다):
(a) 수리 이전 기준선이 없으면 대조할 것이 없다 — 위에서 편집 이전에 쟀다. (b) **비교되는 두 실행이
소비자 3곳을 지나는 테스트를 하나도 스윕하지 않으면 공집합 대 공집합이라 역시 공허하다.** 이 실행은
패키지 전체(17-19초, 락 계열 테스트 다수 포함)를 돌렸으므로 (b)에 해당하지 않지만, **전후가 모두
`ok` 인 이상 이 대조가 실제로 무엇을 배제했는지는 뮤턴트가 아니라 스윕 범위에 달려 있다** — 그것이
이 AC 가 회귀-가드이지 착지 차단이 아닌 이유다. (c) 기지 플레이크가 집합을 움직이면 판독 불가가 된다.

#### 뮤턴트 3종 (AC-BLE-005) — 정의된 둘 + 재규정된 하나

**세 뮤턴트 모두 컴파일된다.** 첫 M-broad 시도는 `errors` 미사용으로 **빌드 실패**했고, 빌드 실패는
RED 가 아니므로 심는 방식을 고쳐 다시 쟀다. 그 사실을 여기 남긴다 — 빌드 실패를 RED 로 보고했다면
그 자체가 이 SPEC 이 경계하는 공허한 관측이다.

| 뮤턴트 | 심은 내용 | RED 가 된 AC | 되돌린 뒤 | 무엇을 쟀나 |
|---|---|---|---|---|
| **M-broad** | `\|\| true` 를 경합 조건에 붙여 모든 errno → `ErrBoardLockHeld` | **001b, 002** (001a·003 은 GREEN) | GREEN | 분류 함수의 **넓은 방향** |
| **M-narrow** | 조건을 `&& ... && false` 로 뒤집어 모든 errno → 비경합 감싼 오류 | **001a, 003** (001b·002 는 GREEN) | GREEN | **호출 지점 배선** |
| **M-leak** (재규정) | 실패 분기의 공유 `unix.Close(fd)` 제거 | **003** (001a·001b·002 는 GREEN) | GREEN | **fd 위생** |

RED 전문 (핵심 줄):

- **M-broad** rc=1 — `board_lock_errno_test.go:80: IsBoardLockHeld(kanban board lock held) = true, want false` ×4 +
  `board_lock_errno_test.go:100: errors.Is(kanban board lock held, <errno>) = false, want true` ×4;
  `--- PASS: TestBoardFlockErrnoContentionRemainsHeld`, `--- PASS: TestBoardFlockErrnoFailurePathClosesDescriptor`;
  `FAIL github.com/modu-ai/moai-adk/internal/kanban 0.550s`
- **M-narrow** rc=1 — `board_lock_errno_test.go:58: IsBoardLockHeld(lock board lock /var/folders/.../board.lock: resource temporarily unavailable) = false, want true`;
  `board_lock_errno_test.go:167: attempt 0: expected contention sentinel, got lock board lock .../board.lock: resource temporarily unavailable`;
  001b·002 는 `--- PASS`; `FAIL ... 0.530s`
- **M-leak** rc=1 — `board_lock_errno_test.go:178: descriptor leak: probe fd 6 before, 206 after 200 failed acquisitions (slack 16)`;
  001a·001b·002 는 `--- PASS`; `FAIL ... 0.493s`

**되돌림 확인**: `git status --short` → `internal/kanban/board_lock_unix.go` 가 목록에 없다(추적 파일
변경 0). 되돌림 후 `go test ./internal/kanban/ -run '^TestBoardFlockErrno' -count=1 -v` rc=0,
네 테스트 + 여덟 하위 케이스 전부 `--- PASS`.

**"서로 다른 AC 를 RED 로" 판정**: M-broad → {001b, 002}, M-narrow → {001a, 003}, M-leak → {003}.
**M-broad 와 M-narrow 의 RED 집합은 서로소이므로 `plan.md §F M3` 의 조건을 만족한다** — 짝이 방향
하나만 재고 있지 않다.

**정직하게 적을 것 하나**: 003 이 M-narrow 에도 반응한다. 003 의 유도가 "경합 sentinel 이 나온다"를
전제로 하기 때문이며, 분류가 뒤집히면 그 전제가 깨져 fd 측정 이전에 멈춘다. 이것은 결함이 아니라
**의도한 전제 가드**다 — 전제가 깨진 채로 fd 를 재면 엉뚱한 것을 재게 된다. 그러나 그 결과 003 은
순수한 fd-위생 가드가 아니라 분류에 결합돼 있으며, **003 의 fd 단언 자체를 재는 뮤턴트는 M-leak
하나다.**

### E.2.6 게이트 (E2 · E3 · E5)

| 항목 | 명령 | 출력 | rc |
|---|---|---|---|
| 호스트 빌드 | `go build ./...` | (출력 없음) | 0 |
| **크로스 플랫폼** | `GOOS=windows GOARCH=amd64 go build ./...` | (출력 없음) | 0 |
| vet | `go vet ./internal/kanban/` | (출력 없음) | 0 |
| **lint** | `golangci-lint run --timeout=5m ./internal/kanban/...` | `0 issues.` | 0 |
| **coverage** | `go test -cover ./internal/kanban/... -count=1` | `ok ... 18.246s coverage: 86.5% of statements` | 0 |

**lint 신규 vs baseline**: baseline(HEAD `9328a5242`, 편집 이전) = `0 issues.`,
수리 후 = `0 issues.` → **신규 지적 0**.

**Windows**: `board_lock_windows.go` 를 **읽기만 했다**(`os.IsExist` 판별이 기준면). 수정 0.
크로스 빌드가 rc=0 이라는 것은 빌드 태그 분기가 유지됐다는 뜻이지, windows substrate 의 **테스트가
돌았다는 뜻이 아니다** — 이 머신에서 windows 테스트는 실행되지 않았다.

### E.2.7 배경 부하 (§D.2 체크리스트)

**배경 프로세스를 하나도 띄우지 않았다.** 모든 Bash 호출은 전경 실행이며 `&` 도 백그라운드 모드도
쓰지 않았다. AC-BLE-003 의 경합 반복은 **인프로세스 루프**(200회)이고 잠금 보유자는 `t.Cleanup` 에
등록돼 해제된다 — 뒤따르는 `kill` 에 의존하는 정리가 없다.

**로컬 전체 스위트(`go test ./...`)를 돌리지 않았다.** 범위는 `./internal/kanban/...` 한정이고
전 패키지 판정은 CI 몫이다.

### E.2.8 기록해 두는 위험 패턴 — **수리는 형제 표면까지 훑어야 끝난다**

이 카드의 계획 단계에서 같은 형태가 **세 번** 났다: `spec.md` 에서 도달 가능성 주장을 철회했는데
`plan.md` 는 그대로 단언하고 있었고, `acceptance.md` 에서 측정 pin 을 철회했는데 `plan.md:111` 은
여전히 그 SHA 에서 **재라고 지시**하고 있었다. 리드는 같은 배치의 다른 레인에서 같은 형태를
네 번째로 보고했다.

**패턴**: 한 파일에서 수정한 사실이 다른 파일에 복제돼 있으면, 그 파일을 함께 고치기 전까지 수정은
끝나지 않았다. 살아 있는 지시문에 남은 잔재가 낡은 라벨보다 비싸다 — 라벨은 오해를 낳지만 지시문은
**잴 수 없는 것을 재라고 시킨다.**

**이 실행에서 적용한 방식**: 뮤턴트 3종을 각각 심고 되돌린 뒤 `git status --short` 로 추적 파일
변경 0 을 확인했고, `/dev/fd` 기제를 교체할 때 그 기제를 언급하는 표면을 함께 확인했다 —
`acceptance.md:89`(기제 본문), `:91`(부채 기록), `plan.md:48`(미러), `plan.md:109`(M2 지시).
넷 다 SPEC 본문이라 **수정 권한이 없으므로** 고치지 않고, 교체 사실과 그 근거를 여기 한 곳에
적어 §E.2 가 단일 보유처가 되게 했다. 형제 표면을 훑되 권한 경계는 넘지 않았다.

### E.2.9 변경 파일

| 파일 | 성격 | 규모 |
|---|---|---|
| `internal/kanban/board_lock_unix.go` | 분류 seam 추가 + 호출 지점 배선 | +26 / −1 |
| `internal/kanban/board_lock_errno_test.go` | 신규 계약 테스트 (`//go:build !windows`) | +181 |
| `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/spec.md` | frontmatter `status: draft → in-progress` | +1 / −1 |

`plan.md §5` 의 "영향 파일 2개 예상" 안에 있다(SPEC 아티팩트 제외). `§A.2 PRESERVE` 목록 밖의 쓰기 0.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: 364bc332f            # M1; progress/evidence commit은 이 뒤에 이어짐
run_status: complete
tree_at_measurement: 364bc332f       # 모든 GREEN·게이트 관측의 귀속 트리
red_now_tree_sha: 9c196204c76b8f7ff2cba3873c7d21ca7c128017   # git write-tree, 부모 HEAD 9328a5242
run_entry_head: 9328a5242

ac_landing_blocking_pass_count: 3    # 001b, 002, 005
ac_landing_blocking_fail_count: 0
ac_regression_guard_count: 3         # 001a, 003, 004 — PASS로 세지 않고 등급 보고 (verification-completeness §2.1)
ac_regression_guard_observed_green: 3
ac_fail_count: 0

red_now_cells_measured: 3            # 001b, 002, 005 — 명령+출력전문+rc+트리SHA 전부 기록
mutants_defined: 3                   # M-broad, M-narrow, M-leak(재규정)
mutants_red_observed: 3
mutants_reverted_clean: true         # git status --short 추적 변경 0
mutant_red_sets_disjoint: true       # M-broad{001b,002} vs M-narrow{001a,003}

preserve_list_post_run_count: 0      # §A.2 PRESERVE 밖 쓰기 0
total_run_phase_files: 3
m1_to_mN_commit_strategy: "M1 단일 구현 커밋 + 증거 커밋 1 (Tier S)"

pre_repair_baseline_measured: true   # 편집 이전 HEAD 9328a5242 에서
pre_repair_failure_set: "empty (ok, 17.413s, rc=0)"
post_repair_failure_set: "empty (ok, 18.923s, rc=0)"
failure_set_delta: none

cross_platform_build:
  host_go_build: pass                # rc=0
  goos_windows_amd64_build: pass     # rc=0
  windows_tests_executed: false      # 이 머신에서 실행되지 않음 — CI 몫
go_vet: pass
lint_baseline: "0 issues."
lint_post_repair: "0 issues."
new_warnings_or_lints_introduced: 0
coverage: "86.5%"                    # 임계 85% 충족
coverage_scope: "internal/kanban only"

background_processes_spawned: 0
local_full_suite_run: false          # 금지 규율 준수
known_flake_fired: false             # TestConcurrencyStress 계열 미발화

l44_pre_commit_fetch: not_performed  # 아래 Gaps 참조
l44_post_push_fetch: n/a             # push 없음 — 통합은 리드 소관
pushed: false
pr_opened: false

debt_closed:
  - "부채-2 (/dev/fd 기제 3구멍): 기제를 POSIX open(2) 최저번호 계약으로 교체, slack=16 고정, 빈 스윕 실패 처리"
  - "부채-3 (M-leak 발화 불가): 공유 close 대상으로 재규정, AC-BLE-003 RED 관측으로 확인"
debt_remaining:
  - "AC-BLE-004: 뮤턴트 가드 없음 (구조상 대조 게이트) — 스윕 범위에 의존"
  - "AC-BLE-003 의 fd 단언은 M-leak 단일 뮤턴트가 잰다"
  - "EINTR / ENOLCK / EOPNOTSUPP 실동작 미측정 — 분류만 잠금"
  - "리눅스 관측 없음 — CI(ubuntu) 몫"
  - "RED-now 트리 객체 9c196204c 는 ref 미도달 (gc 대상); 복구 수단은 red-now-tree.diff"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: 9174f4dd7               # 백필됨 — docs(SPEC-BOARDLOCK-ERRNO-001): sync-phase close
sync_status: complete
changelog_entry: added                   # CHANGELOG.md [Unreleased] → ### Added (파일 자체 관행을 따름; 이 SPEC 은 새 기능이 아니라 방어적 좁히기이지만, 이웃 SPEC-close 항목 전부가 ### Added 아래 있어 그 관행을 따랐다)
changelog_dedup_check: "grep -c -i 'BOARDLOCK-ERRNO' CHANGELOG.md → 0 (편집 전, 이 세션이 직접 실행)"

frontmatter_status_transitions:
  spec_md: "in-progress → completed"     # 이 세션이 수행
  plan_md: "N/A — status: 필드 없음 (ArtifactStatusFieldForbidden, 카드 t357)"
  acceptance_md: "N/A — status: 필드 없음 (ArtifactStatusFieldForbidden, 카드 t357)"
  progress_md: "N/A — status: 필드 없음"

docs_surfaces_checked:
  - target: "README.md, docs-site/content/**"
    command: "grep -rli 'boardlock\\|board.lock\\|flock' README.md docs-site/content"
    result: "4 hits, 전부 docs-site/content/{en,ko,ja,zh}/guides/mcp-server.md 의 'flock' — mcp.json atomic-RWM seam 을 가리키는 무관한 문맥(board-lock 서브시스템과 무관)"
    control: "같은 grep 이 무관 문맥에서 0 이 아닌 결과를 냈으므로, board-lock errno 분류에 대한 0건은 범위를 잘못 잡은 스캔이 아니라 실측 0 이다"
  - target: "이 SPEC 이 소유한 docs 표면"
    result: "없음 — internal/kanban 은 사용자 대면 문서 표면이 없는 내부 패키지"

kickoff_approval_relay: >
  이 레인(카드 t379 워크트리)은 이 토폴로지에서 운영자 채널을 리드 세션이 쥐고 있고 레인은
  운영자에게 직접 물을 수 없다. run-phase 착수 승인이 이 레인에 도달한 경로는 **리드로부터의
  중계**이며, 이 sync 세션이 직접 관측한 승인이 아니다. 이것은 이 토폴로지가 설계한 경로이지만,
  1차 관측이 아니라 중계로 기록한다.

plan_audit_verdict_carried_into_run:
  final_verdict: "PASS-WITH-DEBT 0.83 (iter-3, Tier S 문턱 0.75)"
  trajectory: "iter1 0.85 → iter2 0.83 → iter3 0.83 (iter-2 회귀는 회복되지 않음; iter2·iter3 는 뺄셈 전용)"
  debts_carried_at_run_entry: 6   # plan.md §B.1 이 M1 로 넘긴 것
  debts_closed_by_run: 2          # 부채-2(/dev/fd 기제 교체), 부채-3(M-leak 재규정) — progress.md §E.2.2/§E.2.3
  debts_still_standing: 4         # progress.md §E.3 debt_remaining 4항목 그대로 유지 — 이 sync 세션은 추가로 닫은 것이 없다

sync_session_gaps: >
  이 sync 세션은 run-phase 이후 새로 실행한 검증이 없다 — CHANGELOG 중복 grep(§changelog_dedup_check) 과
  docs 표면 grep(§docs_surfaces_checked) 두 건 외에는 progress.md §E.2/§E.3 이 이미 기록한 관측을
  재인용했을 뿐, 재측정하지 않았다. golangci-lint·go test·coverage·크로스빌드는 재실행하지 않았다.
  plan-audit 재실행도 없다(§E.3 이 이미 "감사 재실행 없음"으로 기록). 리눅스·Windows 런타임 관측은
  여전히 없다 — CI(ubuntu) 몫이다.

ordering_record_body_correction_after_close: >
  This record exists because the close commit and the body correction landed out of the order the
  Status Transition Ownership Matrix assumes, and it states why that was accepted rather than
  reverted.

  What happened, in order: the sync commit `9174f4dd7` carried `status: completed` (the
  `in-progress → completed` transition, per the matrix) BEFORE the AC-BLE-003 body correction —
  replacing the `/dev/fd` judgment-mechanism description with the POSIX `open(2)`
  lowest-numbered-descriptor contract the run phase actually built, plus converting present-tense
  "this debt is still open" language to past tense across spec.md/plan.md/acceptance.md (spec.md
  bumped to 0.5.0 with a HISTORY row) — was committed. The hold instruction that would have
  sequenced the correction ahead of the close reached this agent only after `9174f4dd7` had already
  landed.

  Why it was not reverted, all three reasons together:
  1. The branch (`WT-boardlock-errno`) is unpushed, so the lifecycle event that matters is the
     eventual `develop` merge, not any individual local commit on this branch — the sequence a
     drift detector or a later auditor reads is the branch's state at integration, not the
     intermediate commit order inside it.
  2. What a drift detector or auditor actually reads is the branch's FINAL state at integration
     time. By the time this branch merges, `status: completed` and the corrected body co-exist on
     the same tree — indistinguishable from having been authored in the "correct" order from the
     start.
  3. The Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md`)
     defines no row for a plain `completed → in-progress` walk-back. The only documented path back
     from `completed` is the full amendment procedure — `amendment_of:` frontmatter + a HISTORY
     `## Amendments` section + plan-audit cache invalidation — which exists for substantive
     amendments, not for reordering two commits that both landed in the same session. Invoking that
     procedure here would invent an owner and a transition the matrix does not define, which is a
     worse departure from the matrix than the ordering issue it would fix.

  This is NOT the normal path. It happened because the hold instruction arrived after the sync
  agent had already committed the close. Repeating this pattern on a future card pays the full
  revert-and-redo cost instead — the correct sequencing is: hold the body-correction instruction
  BEFORE the sync commit lands, not after.

  status: is unchanged by this correction commit — it still reads `completed`, matching the matrix's
  scope (`updated:` refresh only; no other frontmatter field, no body content is claimed as owned by
  this note).

  correction_commit_sha: be812f401   # backfilled — docs(SPEC-BOARDLOCK-ERRNO-001): correct AC-BLE-003 mechanism description + record close ordering (t379)
```

## t396 vacuity-condition amendment

카드 t396(lane-10, 브랜치 WT-ble4-vacuity-count)이 AC-BLE-004("실패 집합 불변")의 두 번째 공허화 조건을 acceptance.md `§AC-BLE-004` 등급 칸에 추가했다. 처우는 **조건 추가**다 — AC 재배치(신규 테스트 작성)가 아니다. 근거: mutation 락의 darwin 경합 테스트는 프로세스 재실행 plumbing 설계를 요하는 별개 작업이라 별도 카드 축으로 남긴다.

Census (판정 플랫폼 기준, 축별 경합 테스트 계수 — this tree): board 락 축 4 · backlog 락 축 3 · mutation 락 축 windows 2 / darwin-linux 0. 세 축 중 mutation 락 축만 darwin/linux 스윕을 지나지 않으므로, AC가 단언하는 "세 소비자를 지나는" 실패 집합 기준으로 이 축에서 조건이 성립한다.

판정 근거: `.moai/reports/t396/verdict.md` — 기록 시점 untracked(리드 반출본)로, 오케스트레이터 커밋이 이 브랜치에 함께 싣는 것을 전제로 인용한다. 선행 근거: t379의 AC-BLE-004 측정 방법은 패키지 전체 스윕(`go test ./internal/kanban/... -count=1`)이며, darwin에서 이 스윕은 mutation 락 축을 아무것도 지나지 않는다.

첫 번째 조건(기준선 부재)은 그대로 두고 헤지 문구만 조정했다 — "지금 식별된 공허화 조건은 이것이다(전수 열거는 아니다)"에서 "공허화 조건은 둘이다(전수 열거는 아니다)"로. 조건이 2개가 된 시점부터 단수 진술은 더 이상 정확하지 않다.
