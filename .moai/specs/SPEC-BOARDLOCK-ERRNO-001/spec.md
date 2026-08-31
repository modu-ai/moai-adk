---
id: SPEC-BOARDLOCK-ERRNO-001
title: "Unix board-lock 획득의 errno 보존 — 경합만 ErrBoardLockHeld 로 분류"
version: "0.4.0"
status: in-progress
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P3
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "board-lock, flock, errno, sentinel, defensive-narrowing, regression-pin"
tier: S
era: V3R6
related_specs: [SPEC-STRESS-INVARIANT-VERDICT-001, SPEC-KANBAN-BOARD-001]
---

# SPEC: Unix board-lock 획득의 errno 보존

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-31 | manager-spec | 최초 작성 — 카드 t379. 계측 프로브(2회, 심은 뒤 되돌림) 결과를 전제로 구성. 카드 문구가 아니라 **측정된 결과** 위에 씀 |
| 0.2.0 | 2026-08-31 | manager-spec | plan-audit iter-1(`plan-audit-iter1.md`, PASS-WITH-DEBT 0.85) 지적 반영. §1.3-4 를 전수 주장에서 **측정 진술**로 좁힘, `EINTR` 을 미측정 목록에 추가하고 §1.3.1 로 그 재분류가 행동 중립이 **아님**을 받아들이는 변화로 기록, **REQ-BLE-005 의 범위를 "측정된 도달 가능"으로 좁힘**(D1). §3 대조표에서 AC-BLE-005 의 피복 REQ 에 004 추가(D4). `board_lock_unix.go` 인용 2건 교정(D5) |
| 0.3.0 | 2026-08-31 | manager-spec | plan-audit iter-2(`plan-audit-iter2.md`, 0.83 — iter-1 의 0.85 에서 하락) 반영. iter-2 진단: *"iter-2 가 누락을 확언으로 바꿨고 그중 셋이 성립하지 않는다"*. 따라서 이 pass 는 **뺄셈 전용**이다 — 새 보증·새 가드·새 주장을 더하지 않고 철회·축소·부채 기록만 한다. `plan.md` 에 살아남아 있던 전수 주장 3곳(§A, §A.1 표, §D) 철회(N1), `acceptance.md` 의 `3f03d9c36` RED-now pin 철회(잴 수 없는 트리를 가리키고 있었다 — 측정 트리 구성은 M1 소관), AC-BLE-004 를 덮는다고 적혀 있던 일괄 비공허성 주장 축소(N5), `§E` E8 의 "RED-now 셀 6개" 요구를 3개+등급보고 3개로 축소(N7). N2(M-leak 발화 불가)·N6(`/dev/fd` 기제)는 **닫지 않고** `plan.md §B.1` 에 M1 부채로 기록 |
| 0.4.0 | 2026-08-31 | manager-spec | plan-audit iter-3 반영. **이 pass 도 뺄셈 전용이다.** (P1) `plan.md §F M2` 에 살아 있던 "M2 착수 전 `3f03d9c36` 에서 RED-now 를 재라"는 지시를 철회된 pin 에 맞춰 교정 — 측정 트리는 M1 이 만들고 SHA 는 실측 시점에 기록한다. (P2) M-leak 정의 철회 뒤에도 남아 있던 **뮤턴트 3종 요구 5곳**을 정의된 2종(M-broad·M-narrow)으로 축소, M-leak 행은 철회 상태로 보존. (P3) 이 표 아래 자기 철회 기록에 있던 **"이 배치에서 세 번째 사례"** 라는 배치 전역 집계 삭제 — 이 SPEC 이 검증할 수 없는 전언이었다. 추가로 AC-BLE-004 등급 칸의 "유일한 조건" 보편 주장 축소, `acceptance.md §D.0` 의 "측정 트리는 X **이다**"를 미래형으로 축소, `verification-completeness.md` §2.1 인용과 충돌하던 **"6개 전부 PASS" 착지 조건을 착지 차단 셋(001b·002·005)으로 축소**(회귀-가드 셋은 등급 보고 대상이며 PASS 로 세지 않는다) |

### 감사 자기 철회 기록 (iter-1 → iter-2)

iter-1 감사는 AC 마다 RED-now 셀을 요구했다. 수리 pass 는 그 요구를 되받았다 — AC 셋은 수리 이전 트리에서 **이미 초록**이라 정직한 RED-now 가 존재하지 않고, 억지로 만들어 붙이는 것은 `verification-completeness.md` §2 가 금하는 방향이라는 것이다. **iter-2 에서 감사가 자기 요구가 틀렸음을 받아들였고**, 동시에 자기 표적도 스스로 교정했다 — 실제로 가드가 비어 있는 것은 AC-BLE-004 가 아니라 **AC-BLE-003** 이었다. 이 SPEC 의 감사에서 일어난 일을 다음 라운드를 위해 기록으로 남긴다. 이것은 일어난 일의 기록일 뿐 코드에 대해 아무것도 주장하지 않는다.

---

## 1. 문제 — 형태는 참, 오분류 실측치는 0

### 1.1 먼저, 숫자부터

**이 호출 지점에서 실측된 오분류는 0건이다.** 오늘 깨져 있는 것은 없고, 이 SPEC 이 사라지게 만드는 실패도 없다.

카드 t379 는 t372 verdict `.moai/reports/t372/verdict.md` §9.3 후보 B 에서 왔고, "Unix `ErrBoardLockHeld` 술어가 지나치게 넓다"고 적는다. 그 문장은 **형태(shape)** 에 대한 주장으로는 참이다. 그러나 형태가 참인 것과 그 형태가 실제로 오분류를 만드는 것은 다른 명제이며, 후자를 재기 전에 좁히면 한 방향의 실패를 다른 방향의 실패와 맞바꾸게 된다. 그래서 계획 이전에 집합을 쟀다.

### 1.2 형태 — 확인된 사실

`internal/kanban/board_lock_unix.go:43`:

```go
if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
    _ = unix.Close(fd)
    return nil, ErrBoardLockHeld
}
```

`err` 가 버려진다. 어떤 errno 가 오든 결과는 경합 sentinel 하나다. `IsBoardLockHeld`(`board_lock.go:28`)는 `errors.Is` 로 그 sentinel 을 볼 뿐이므로, 호출자는 "잠겨 있음"과 "flock 이 다른 이유로 실패함"을 구별할 수단이 없다.

### 1.3 집합 — 어떻게 쟀고 무엇이 나왔나

`internal/kanban` 에 프로브 2개를 심고, 돌리고, 되돌렸다(패키지는 지금 깨끗하다). 증거 전문:

- `.moai/reports/t379/errno-probe-source.go.txt` — 심었던 프로브 원문
- `.moai/reports/t379/errno-probe-output.txt` — 그 실행의 출력 전문

플랫폼: **darwin/arm64** (이 머신). CI 는 **ubuntu** 이므로, 아래는 리눅스 커널의 관측이 아니다.

```
A contention   errno=resource temporarily unavailable EWOULDBLOCK=true EAGAIN=true
B closed fd    errno=bad file descriptor              EBADF=true
C invalid how  errno=<nil>                            EINVAL=false
mapping A contention   -> IsBoardLockHeld=true
mapping B closed fd    -> IsBoardLockHeld=true
mapping C invalid how  -> no error produced; nothing to map
second AcquireBoardLock -> err=kanban board lock held IsBoardLockHeld=true
```

읽히는 것:

1. **형태 주장은 참이다** — B(EBADF)가 경합으로 분류된다.
2. **그러나 B 는 이 호출 지점에서 도달 불가다.** `acquireBoardLockImpl` 은 성공한 `unix.Open` 바로 다음 줄에서 `unix.Flock` 을 부른다(`board_lock_unix.go:37-41`). 닫힌·낡은 디스크립터를 들고 있을 수 없다.
3. **C(EINVAL) 도 도달 불가다.** 호출 지점은 컴파일 타임 상수 `unix.LOCK_EX|unix.LOCK_NB` 를 넘긴다. 게다가 darwin/arm64 에서 엉뚱한 `how` 는 **오류 자체를 내지 않았다**.
4. **여기서 유도한 세 케이스(A·B·C) 중 실호출 경로를 통해 실제로 산출된 errno 는 EWOULDBLOCK/EAGAIN 하나였고, 그것은 올바르게 분류된다.** 이 문장은 **잰 것에 대한 진술이지 errno 집합에 대한 전수 주장이 아니다** — 프로브는 세 케이스를 심었을 뿐 이 호출 지점이 낼 수 있는 errno 를 열거하지 않았다. 유도하지 않은 errno(아래 5·6)에 대해서는 이 측정이 아무 말도 하지 않는다.
5. `ENOLCK`(잠금 자원 고갈)·`EOPNOTSUPP`(flock 미지원 파일시스템 — NFS 일부, 일부 컨테이너 오버레이)는 다른 파일시스템·커널에서 여전히 그럴듯하다. **이 머신에서 유도하지 않았다 — 0 이 아니라 미측정이다.**
6. **`EINTR` 도 같은 미측정 목록에 들어간다.** `unix.Flock`(x/sys v0.47.0, `zsyscall_darwin_arm64.go:1337`)은 EINTR 재시도가 없는 생 syscall 래퍼이므로, 커널이 EINTR 을 내면 그대로 호출자에게 도달한다. 다만 **도달 가능성 논증은 "그럴듯하지 않다" 쪽으로 기운다**: `LOCK_NB` 는 대기하지 않는 호출이고 flock(2) 이 EINTR 을 문서화하는 것은 *대기 중 중단*에 대해서다. **이것은 syscall 계약을 읽은 추론이지 측정이 아니다 — 유도하지 않았다.**

### 1.3.1 EINTR 의 재분류는 행동 중립이 **아니다** — 받아들이는 변화로 기록한다

§1.3-5 의 `ENOLCK`/`EOPNOTSUPP` 과 `EINTR` 은 "미측정"이라는 점에서는 같지만, **좁히기가 그것들에 미치는 영향의 성격은 다르다**. 이 절은 그 차이를 명시한다 — 적지 않으면 나중에 생긴 변화가 이 카드의 회귀로 오귀속된다.

오늘 `EINTR` 은 다른 모든 errno 와 함께 `ErrBoardLockHeld` 로 매핑되고, 그래서 소비자 3곳에서 **경합 재시도 예산에 흡수된다**(`acquireBoardLockSerialized`, `board_store.go:165-181`). 좁힌 뒤에는 즉시 하드 오류가 된다. 즉 `EINTR` 이 실제로 도착하는 배포에서는 **행동이 바뀐다**.

**결정: `EINTR` 을 경합 등가로 취급하지 않는다.** REQ-BLE-002 의 "EWOULDBLOCK/EAGAIN 이외 전부"에 `EINTR` 이 포함되며, 재시도 대상이 아니라 하드 오류가 된다. 근거 셋:

- 재시도 정책과 대기 예산은 §4 에서 명시적으로 범위 밖이다. `EINTR` 을 경합으로 넣는 것은 분류 변경이 아니라 재시도 정책 확장이다.
- 이 SPEC 의 실체는 errno 를 보존하는 것이다. `EINTR` 이 하드 오류로 드러나면 `errors.Is(err, unix.EINTR)` 로 판별 가능해지고, 오늘처럼 예산 안에서 조용히 삼켜지지 않는다.
- 도달 가능성 논증이 "그럴듯하지 않다" 쪽이므로(§1.3-6), 유도 불가한 입력을 위해 재시도 표면을 넓히는 것은 근거 없는 확장이다.

**이 결정은 받아들이는 행동 변화이며 미측정이다.** `EINTR` 이 실제로 관측되는 날 재시도 등가로 되돌릴지는 그때의 별도 카드 소관이고, 이 SPEC 은 그 판단을 선점하지 않는다. REQ-BLE-005 의 범위가 이에 맞춰 좁혀져 있다(§2 REQ-BLE-005 주석 참조).

### 1.4 t372 §9.3 후보 B 의 `errors.Join` 쪽은 오늘 도달 불가 — 재조사 금지

t372 후보 B 는 `errors.Join` 이 sentinel 을 실어 나를 가능성도 함께 적었고, 그 자신이 "미래에 생기면"으로 한정했다. **본 트리에서 확인했다**: 폴드 함수 `joinBoardReleaseErr`(`board_store.go:289`)와 그 backlog 짝 `joinBacklogReleaseErr`(`backlog_store.go:677`)의 호출 지점 4곳(`board_store.go:237`, `board_recover.go:77`, `backlog_store.go:623`, `backlog_store.go:644`)은 **모두 획득이 성공한 뒤에만** 설치·실행된다 — 획득 실패는 그 앞에서 반환된다. 따라서 `mutErr`·`relErr` 중 어느 쪽도 `ErrBoardLockHeld` 를 실을 수 없다. t372 의 한정은 옳았고, **이 축은 다시 열 것이 없다.**

### 1.5 그런데도 왜 고치는가 — 승인된 방향

리드 판정(옵션 A, 방어적 좁히기). 오늘 유도할 수 없더라도 errno 를 버리는 매핑은 이 저장소가 반복해서 잡아 온 **"실패를 성공으로 읽는"** 형태 그 자체이며, 지금 닫아 두면 `ENOLCK`/`EOPNOTSUPP` 가 실제로 도착했을 때 같은 조사를 다시 하지 않아도 된다.

기각된 선택지 둘: **현재 술어를 그대로 고정**(틀린 매핑을 규범으로 박제한다), **카드 종결**(형태가 실재하므로 조사가 나중에 되풀이된다).

대조 기준면은 **Windows 다**. `board_lock_windows.go:69` 는 이미 좁다 — `os.IsExist` 로 경합을 판별하고, `os.ErrPermission` 은 짧은 재시도 예산으로, 나머지는 감싼 하드 오류로 보낸다. Windows 를 바꾸는 것이 아니라 Unix 를 그 모양에 맞추는 것이다. CI 의 ubuntu 러너가 넓은 쪽(Unix substrate)을 돌린다.

---

## 2. 요구사항 (GEARS)

**REQ-BLE-001** — When the non-blocking exclusive `flock` at the board-lock acquisition site fails with `EWOULDBLOCK` or `EAGAIN`, the Unix board-lock substrate shall return an error that `IsBoardLockHeld` reports true for.

> 이 호출 지점에서 **측정된** 유일한 errno 이자(§1.3-4) 유일하게 올바른 분류다. 좁히기가 이 방향을 건드리면 그것은 수리가 아니라 규칙 비활성화다.

**REQ-BLE-002** — When that `flock` fails with any errno other than `EWOULDBLOCK`/`EAGAIN`, the substrate shall return an error that `IsBoardLockHeld` reports false for.

> REQ-BLE-001 과 짝을 이루는 반대 방향. 둘 중 하나만 잠그면 "올바른 좁히기"와 "그냥 술어를 껐음"이 구별되지 않는다.
>
> **`EINTR` 은 이 "이외 전부"에 포함된다** — 경합 등가로 예외 처리하지 않는다. 근거와 그 결과(재시도 예산 흡수 → 즉시 하드 오류)는 §1.3.1 이 소유한다.

**REQ-BLE-003** — The non-contention error of REQ-BLE-002 shall preserve the underlying errno for `errors.Is` inspection and shall name the lock path in its message.

> errno 를 버리지 않는 것이 이 SPEC 의 실체다. `ENOLCK` 이 실제로 도착한 날 운영자가 `errors.Is(err, unix.ENOLCK)` 로 판별할 수 있어야 하고, 메시지만으로도 어느 아티팩트인지 읽혀야 한다.

**REQ-BLE-004** — While the `flock` fails on any errno, contention or otherwise, the substrate shall close the descriptor opened for that attempt before returning.

> 현재 코드가 이미 지키는 성질이며(`board_lock_unix.go:42`), 분기가 둘로 갈리는 변경에서 가장 흔하게 깨지는 성질이다. 새 분기가 fd 를 새게 하지 않는다는 것을 요구로 못박는다.

**REQ-BLE-005** — The observable behaviour of `AcquireBoardLock` and of every `IsBoardLockHeld` consumer shall be unchanged on every **measured-reachable** input — that is, on `EWOULDBLOCK`/`EAGAIN`, the only errno this call site was observed to produce (§1.3-4).

> §1.3 의 측정이 곧 이 조항의 근거다. 측정된 입력에서 관측 가능한 차이는 **0** 이다. 이 문장이 없으면 나중에 생긴 변화가 이 카드의 회귀로 오귀속된다.
>
> **범위가 "오늘 도달 가능한 전부"가 아니라 "측정된 도달 가능"인 이유**: errno 집합은 전수 열거되지 않았다(§1.3-4). 미측정 errno — `ENOLCK`·`EOPNOTSUPP`·`EINTR` — 는 이 조항의 보호 범위 **밖**이며, 그 중 `EINTR` 은 재분류가 행동 중립이 아님이 §1.3.1 에 명시적으로 기록된 **받아들이는 변화**다. 이 조항을 "어떤 입력에서도 행동 불변"으로 읽으면 §1.3.1 과 정면으로 모순된다.

### 2.1 소비자 쪽 파급 — 의도된 것, 그리고 오늘의 크기

`IsBoardLockHeld` 소비자는 셋이며 모두 같은 모양이다 — `if !IsBoardLockHeld(err) { return ... }`:

| 위치 | 넓은 매핑에서의 현재 처리 | 좁힌 뒤 |
|---|---|---|
| `internal/kanban/board_store.go:173` (`acquireBoardLockSerialized`) | 비경합 errno 를 경합으로 보고 대기 예산만큼 재시도 | 즉시 하드 오류로 반환 |
| `internal/kanban/integration_lock_mutation.go:103` | 같음 (예산 소진 후 wedge-clear 경로까지 감) | 즉시 감싼 오류로 반환 |
| `internal/kanban/backlog_store.go:736` | 같음 | 즉시 감싼 오류로 반환 |

**이것이 의도다** — 지금 경합으로 삼켜지는 것을 하드 오류로 드러낸다. 그리고 **측정된 입력에서 관측 가능한 차이는 0 이다**: 프로브가 유도한 비경합 errno 들이 이 호출 지점에서 도달 불가이기 때문이다(§1.3). 두 문장이 함께 있어야 한다.

세 번째 문장이 필요하다: **미측정 errno 에서는 차이가 0 이 아닐 수 있다.** 특히 `EINTR` 이 도착하는 배포에서는 위 표의 "재시도 예산만큼 재시도 → 즉시 하드 오류" 전이가 실제로 일어난다(§1.3.1). 그 변화는 받아들이는 것으로 기록됐지 측정된 것이 아니다.

---

## 3. 수락 기준 — REQ 피복 대조표

AC 본문은 `acceptance.md` 가 단일 보유처다(중복 기재하지 않는다 — 두 곳에 적으면 갈린다).

| REQ | 피복 AC |
|---|---|
| REQ-BLE-001 | AC-BLE-001a |
| REQ-BLE-002 | AC-BLE-001b |
| REQ-BLE-003 | AC-BLE-002 |
| REQ-BLE-004 | AC-BLE-003 |
| REQ-BLE-005 | AC-BLE-004 |
| (비공허성 — REQ-BLE-001·002 공동. **004 는 M-leak 정의 철회로 현재 미피복** — `acceptance.md §D.0`, `plan.md §B.1` 부채-1) | AC-BLE-005 |

미피복 REQ 0건, 고아 AC 0건. **양방향 짝은 AC-BLE-001a ↔ AC-BLE-001b** 이며, 둘은 함께여야만 판정을 이룬다.

**측정 불가 형태의 AC 는 쓰지 않는다.** "오분류 N건이 사라진다"로 쓸 수 없다 — N 은 0 이다.

---

## 4. 범위 밖 (exclusions)

### Out of Scope — Windows substrate
- `board_lock_windows.go` 는 건드리지 않는다. 이미 좁고(`os.IsExist` 판별), 이 SPEC 의 기준면이지 변경 대상이 아니다.

### Out of Scope — t372 §9.3 의 `errors.Join` 축
- §1.4 에서 도달 불가로 확인됐다. 폴드 함수 2개와 그 호출 지점 4곳을 수정하지 않는다.

### Out of Scope — `ENOLCK` / `EOPNOTSUPP` / `EINTR` 의 실측 유도
- 이 머신(darwin/arm64)에서 유도하지 않으며, 유도용 파일시스템·컨테이너·시그널을 구성하지 않는다. 셋 다 미측정으로 남기고 그 사실을 명시한다(§1.3-5·§1.3-6).

### Out of Scope — 재시도 정책과 대기 예산
- `boardLockWaitBudget`·`boardLockRetryWait`·wedge-clear 경로는 손대지 않는다. 이 SPEC 은 분류만 바꾼다.

### Out of Scope — `internal/spec` 의 자매 lock substrate
- `internal/spec/lock_unix.go` 가 같은 패턴을 갖더라도 본 카드 범위 밖이다. 그쪽 조사는 별도 카드 소관.

### Out of Scope — EINTR 재시도 도입
- `unix.Flock` 이 EINTR 재시도를 갖지 않음을 §1.3-6 에 기록만 하고, 재시도를 도입하지 않는다. `EINTR` 은 REQ-BLE-002 의 비경합 쪽으로 떨어지며, 그 결과가 행동 중립이 아님은 §1.3.1 이 받아들이는 변화로 기록한다.

---

## 5. 제약

- **영향 파일 2개 예상**: `internal/kanban/board_lock_unix.go`(분류 분기) + 신규/기존 `internal/kanban` 테스트 1개. 이 범위를 넘으면 설계를 다시 본다.
- **검증 범위**: `./internal/kanban/...` 한정. 이 저장소에서 로컬 전체 스위트(`go test ./...`)는 금지다.
- **배경 부하 금지**: 이 패키지는 잠금 축이다. 경합을 만드는 검증은 `t.Cleanup` 등록 또는 `timeout` 래핑으로 정리가 보장돼야 하며, 뒤따르는 `kill` 은 정리가 아니다.
- **Windows 빌드 보존**: `GOOS=windows go build ./...` 가 계속 통과해야 한다(빌드 태그 분기 파일 수정).
- REQ-BLE-005 는 협상 대상이 아니다 — **측정된 도달 가능 입력**(`EWOULDBLOCK`/`EAGAIN`)에서 행동이 바뀌면 그것은 수리 실패다. 미측정 errno 는 이 조항의 범위 밖이며, `EINTR` 의 행동 변화는 §1.3.1 이 받아들이는 것으로 기록한다.
