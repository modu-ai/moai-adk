# plan.md — SPEC-BOARDLOCK-ERRNO-001

## §A Context

- **카드/워크트리/브랜치**: t379 · `.claude/worktrees/t379` · `WT-boardlock-errno` · base `3f03d9c36` (= `origin/develop`).
- **산출물**: `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/{spec,plan,acceptance,progress}.md`.
- **결함 한 줄**: `board_lock_unix.go:43` 이 `unix.Flock` 실패의 errno 를 버리고 전부 `ErrBoardLockHeld` 로 매핑한다.
- **그러나 실측 오분류는 0 건**이다 — 프로브가 유도한 세 케이스(A·B·C) 중 실호출 경로로 실제 산출된 errno 는 `EWOULDBLOCK`/`EAGAIN` 하나였고, 그것은 올바르게 분류된다(`spec.md §1.3-4`). **이것은 잰 것에 대한 진술이지 이 호출 지점이 낼 수 있는 errno 에 대한 전수 주장이 아니다** — 유도하지 않은 errno 에 대해 이 측정은 아무 말도 하지 않는다. 계획은 카드 문구가 아니라 이 측정 위에 서 있다.
- **승인된 방향**: 옵션 A, 방어적 좁히기(리드 재량). 기각: 현행 술어 고정 / 카드 종결.
- **기준면**: Windows substrate(`board_lock_windows.go:69`)가 이미 좁다. Unix 를 그 모양에 맞춘다. Windows 는 변경 대상이 아니다.
- **계획 단계 증거**: `.moai/reports/t379/errno-probe-source.go.txt`, `.moai/reports/t379/errno-probe-output.txt` (프로브 2개를 심고 돌린 뒤 되돌림 — 패키지는 현재 깨끗하다). 플랫폼 darwin/arm64, CI 는 ubuntu.

### §A.1 Tier 판정 — S

측정 결과에서 다시 도출했다(가정을 물려받지 않았다).

| 축 | 값 |
|---|---|
| 패키지 | 1 (`internal/kanban`) |
| 예상 영향 파일 | 2 (`board_lock_unix.go` + 테스트 1) |
| 오늘의 행동 변화 | 0 — **측정된 도달 가능 errno(`EWOULDBLOCK`/`EAGAIN`) 한정**. `EINTR` 재분류는 행동 중립이 아니며 받아들이는 변화로 기록돼 있다(`spec.md §1.3.1`) |
| 설계 결정 | 1 (분류 seam 형태) — 되돌리기 쉬움 |
| 외부 계약 변경 | 없음 (`IsBoardLockHeld` 시그니처·sentinel 불변) |
| 마이그레이션/데이터 | 없음 |

방어적 좁히기 + 회귀 잠금 하나. **Tier S** 로 판정한다.

Tier S 는 통상 AC 를 `spec.md §3` 에 인라인하고 `acceptance.md` 를 두지 않는다. 이 카드는 배차문이 `acceptance.md` 를 명시적으로 요구했으므로 **AC 본문의 단일 보유처를 `acceptance.md` 로 두고**, `spec.md §3` 에는 REQ→AC 대조표만 남겼다 — 두 곳에 본문을 적으면 갈리기 때문이다. 이 편차를 여기 기록해 둔다.

### §A.2 PRESERVE 목록 (범위 절제)

`internal/kanban/board_lock_unix.go` 와 그 계약 테스트 **이외 전부**. 특히:
`board_lock_windows.go`, `board_lock.go`(sentinel·술어 정의), `board_store.go`, `backlog_store.go`, `integration_lock_mutation.go`, `board_recover.go`, `internal/spec/**`, `.github/workflows/**`.

## §B Known Issues

- **B-도달불가 검사**: 이 SPEC 이 새로 잠그는 분기는 오늘 실호출로 유도할 수 없다. 그래서 초록이 실제로 무언가를 보고 있는지는 mutation(AC-BLE-005) 말고는 보일 방법이 없다. 뮤테이션을 선택이 아니라 필수로 둔 이유다.
- **B-기지 플레이크**: `internal/kanban` 에는 `TestConcurrencyStress` 계열의 기지 플레이크 이력이 있다(t372/t306 계보). 검증 중 나오면 계열명과 함께 **별도로** 적고 이 카드에 귀속하지 않는다.
- **B-부하 축**: 이 패키지는 잠금 축이라 경합 검증이 부하를 만든다. 배경 프로세스 금지, 정리는 `t.Cleanup`/`timeout` 으로 보장. 후행 `kill` 은 정리가 아니다.
- **B-빌드태그**: 수정 파일이 `//go:build !windows` 라 로컬 darwin 판정만으로 Windows 컴파일을 보증하지 못한다. `GOOS=windows` 빌드가 필수다.
- **B-미측정**: `ENOLCK`/`EOPNOTSUPP`/`EINTR` 는 유도하지 않는다(`spec.md §4` 의 미측정 목록과 같은 셋). 합성 errno 로 분류만 잠그고, 실동작은 미측정으로 남긴다(0 이 아니다).

### §B.1 M1 로 넘기는 부채 (iter-3 에서 닫지 않는다 — 명시)

두 항목은 M1 의 A/B 설계 결정과 얽혀 있어 그 결정 이전에 확정할 수 없다. **이 절은 부채 기록이지 가드가 아니다** — 여기 적혀 있다는 사실이 아래 성질을 보증하지 않는다.

- **부채-1 (M-leak 이 발화하지 못한다).** `§F M1` 은 `unix.Close(fd)` 를 분류 분기 앞의 **공유 1회 호출**로 규정한다(`양쪽 → 반환 전에 unix.Close(fd)`). 그런데 M-leak 은 "**비경합 분기**에서 `unix.Close(fd)` 제거"로 정의돼 있어, 그 모양에는 제거할 분기가 없다 — 뮤턴트를 적어 넣을 자리가 없다. 게다가 `acceptance.md` AC-BLE-003 의 유도는 같은 root 반복 획득, 즉 **경합만** 만든다. 따라서 **AC-BLE-003 의 비공허성 가드가 실제로 작동하는지는 미확정이며, 현재로선 AC-BLE-003 을 비공허성 근거 없는 AC 로 읽어야 한다.** M1 에서 A/B 결정과 함께 (a) M-leak 을 공유 close 대상으로 다시 규정할지, (b) 그 경우 REQ-BLE-004 의 경합 절반만 덮인다는 사실을 어떻게 적을지 결정한다.
- **부채-2 (`/dev/fd` 기제).** `acceptance.md` AC-BLE-003 의 판정 기제에 세 구멍이 있다. (a) "고정 여유"에 값이 없어 임계값을 구현자가 정하게 된다. (b) "darwin·리눅스 모두 이 경로가 프로세스 자신의 열린 디스크립터를 노출한다"는 **어느 플랫폼에서도 측정되지 않은 추론**인데 사실처럼 적혀 있다(이 문서가 다른 곳에서는 추론/측정을 구분해 적는 것과 어긋난다). 리눅스에서 `/dev/fd` 는 `/proc` 경유이고 그 가용성은 환경 의존이며, **CI 러너는 ubuntu 다**. (c) AC-BLE-001a 가 가진 선택자-매치-수 절이 여기엔 없어, `/dev/fd` 를 못 읽는 러너에서 **조용히 `ok` 로 통과하고 이 AC 는 아무것도 단언하지 않는다**(빈 스윕). M1 에서 플랫폼 라벨과 빈-스윕 절을 **함께** 정한다.

## §C Pre-flight

```bash
git -C .claude/worktrees/t379 branch --show-current
git -C .claude/worktrees/t379 rev-parse --short HEAD
git -C .claude/worktrees/t379 status --short
go test ./internal/kanban/... -count=1
go vet ./internal/kanban/
```

네 번째 명령의 출력이 **수리 전 기준선**이며 AC-BLE-004 가 그것에 귀속된다. 구현 전에 재고 기록한다.

## §D Constraints

- §A.2 PRESERVE 밖 쓰기 금지.
- 검증 범위는 `./internal/kanban/...` — 로컬 전체 스위트(`go test ./...`) 금지, 전 패키지 판정은 CI 몫.
- 배경 부하 금지. 반복 경합은 인프로세스 루프 + `t.Cleanup`.
- 커밋 전 `git rev-parse --short HEAD` + `git branch --show-current` 재판독(AGENTS.md §2), 명시 pathspec 스테이징, sweep 금지.
- Conventional Commits + 카드 id 병기 `(t379)`.
- REQ-BLE-005 는 협상 대상이 아니다 — 다만 그 범위는 **측정된 도달 가능 입력**(`EWOULDBLOCK`/`EAGAIN`)이다(`spec.md §2 REQ-BLE-005`). 미측정 errno 는 이 조항의 보호 범위 밖이며, `EINTR` 의 행동 변화는 `spec.md §1.3.1` 이 받아들이는 것으로 기록한다. "오늘 행동 불변"이라는 넓은 읽기는 철회됐다.

## §E Self-Verification (manager-develop 보고 양식)

E1 AC 매트릭스(AC-BLE-001a·001b·002·003·004·005 — 명령 + 출력 전문 + 선택자 매치 수) / E2 `GOOS=windows GOARCH=amd64 go build ./...` / E3 `go test -cover ./internal/kanban/...` / E4 서브에이전트 경계(해당 시) / E5 `golangci-lint run --timeout=5m ./internal/kanban/...` (신규 vs baseline 구분) / E6 커밋 SHA + push 상태 / E7 blocker / E8 수리 전 기준선 전문 + `acceptance.md §D.0` **RED-now 셀 3개(001b·002·005)의 실측 전문 + 회귀-가드 3개(001a·003·004)의 등급 보고**(RED-now 를 가진 척하지 않는다) + **정의된 뮤턴트 2종(M-broad·M-narrow)** 의 RED 전문과 되돌린 뒤 GREEN. **M-leak 은 정의가 철회돼 있어 보고 대상이 아니다**(§B.1 부채-1) — 3종을 요구하면 정의되지 않은 뮤턴트를 지어내거나 허수로 보고하게 된다. **"RED-now 셀 6개" 요구는 철회됐다** — `§D.0` 6행 중 시작 색이 RED 인 것은 3개뿐이므로 6개를 요구하면 존재하지 않는 셋을 과대 보고하게 된다.

## §F Milestones — 결정 가역성 순 (바뀔 가능성이 큰 결정이 앞선다)

### M1 — 분류 seam 과 오류 형태 (Priority High · 되돌릴 가능성 최대 · REQ-BLE-001·002·003·004)

이 카드에서 **유일하게 열려 있는 설계 결정**이다. 먼저 놓고 리뷰를 받는다.

결정할 것 (A 를 권한다):

- **A — 순수 분류 함수**: `classifyBoardFlockErr(err error, lockPath string) error` 를 뽑아 `acquireBoardLockImpl` 이 그 결과를 반환한다. 합성 errno 로 직접 단위 테스트가 되고, syscall 경로에 테스트 전용 seam 이 남지 않는다.
- **B — 호출 가능 var seam**: `boardFlock = unix.Flock` 을 패키지 변수로 두고 테스트가 교체. 실호출 경로 전체를 지나가지만 프로덕션 코드에 테스트 전용 간접층이 남는다.

A 를 고르면 **공허성 위험**이 하나 생긴다 — 분류 함수는 옳은데 호출 지점이 그것을 쓰지 않아도 AC-BLE-001a 는 통과한다(경합은 어느 쪽이든 held 다). **이 구멍을 겨냥하는 것은 M3 의 M-narrow 뮤턴트다.** B 를 고르면 이 위험은 줄지만 프로덕션 표면이 늘어난다.

**왜 M-narrow 이고 M-broad 가 아닌가** (감사 D2 교정):

- AC-BLE-001b 는 합성 errno 를 **분류 함수에 직접** 먹인다(`acceptance.md` AC-BLE-001b — "substrate 가 그 실패를 분류하면"). `acquireBoardLockImpl` 을 지나가지 않는다. 따라서 분류를 되돌리는 M-broad 는 **호출 지점 배선 여부와 무관하게** 001b 를 RED 로 만든다 — 배선에 대해 아무것도 말해 주지 않는다.
- AC-BLE-001a 는 `AcquireBoardLock` 을 두 번 부르는 **실호출 경로**다. M-narrow(모든 errno 를 비경합으로)를 심었을 때 001a 가 RED 가 되는 것은, `acquireBoardLockImpl` 이 하드코딩된 sentinel 이 아니라 **분류 함수의 결과를 반환하고 있다는 증거**다.
- 뒤집어 말하면: 배선이 안 된 구현에서는 M-narrow 를 심어도 001a 가 **초록으로 남는다**(호출 지점이 여전히 `ErrBoardLockHeld` 를 직접 반환하므로). **그 생존이 곧 신호다** — AC-BLE-005 가 잡아내야 하는 실패가 정확히 이것이다.

오류 형태(기존 관용에 맞춘다 — `board_lock_unix.go:39` 의 `open board lock %s: %w`):

```
경합      → ErrBoardLockHeld            (변경 없음)
비경합    → fmt.Errorf("lock board lock %s: %w", lockPath, err)
양쪽      → 반환 전에 unix.Close(fd)     (기존 성질 보존)
```

경합 판별은 `errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)`. 두 상수는 리눅스에서 같은 값이고 darwin 에서도 프로브가 둘 다 `true` 로 관측했으므로(§1.3 케이스 A) 둘을 함께 적는 것은 중복이 아니라 이식성 표기다.

### M2 — 양방향 계약 테스트 (Priority High · REQ-BLE-001·002·003·005)

- AC-BLE-001a: 실호출 경로(`AcquireBoardLock` 두 번)로 경합 → held=true.
- AC-BLE-001b: 합성 errno 2종 이상(`unix.ENOLCK`, `unix.EBADF`) → held=false, `err != nil`.
- AC-BLE-002: `errors.Is(err, <errno>)` + 메시지에 lock 경로.
- AC-BLE-003: 실패 경로 fd 위생 — **판정 기제는 하나로 고정한다**(감사 D4 교정): 같은 root 에 대해 반복 획득 실패를 N회 유도한 전후로 `/dev/fd` 하위 엔트리 수를 세고 단조 증가하지 않음을 단언한다. "코드 경로 단언" 같은 기제 없는 대안은 삭제됐다.
- 새 테스트 파일은 `//go:build !windows` 를 단다.
- **각 AC 의 시작 색(RED-now / 회귀-가드)은 `acceptance.md §D.0` 이 단일 보유처다.** RED-now 셀을 잴 수 있는 트리는 `3f03d9c36` 자체가 아니라 거기에 새 계약 테스트(분류 수리 이전 상태)를 얹은 트리이며, **그 트리를 만드는 것이 M1 의 일이다**(`acceptance.md §D.0` 의 pin 철회). 따라서 RED-now 실측은 그 트리가 생긴 뒤에 하고, **SHA 는 실측 시점에 그때의 트리 SHA 로 기록한다** — 미리 어떤 SHA 도 적지 않고, 사후에 만들어 붙이지도 않는다.

### M3 — 비공허성 + 게이트 (Priority High · AC-BLE-004·005)

- 뮤턴트 M-broad(전부 held) → AC-BLE-001b 가 RED. 되돌림 후 GREEN. **분류 함수의 넓은 방향**을 잰다.
- 뮤턴트 M-narrow(전부 비경합) → AC-BLE-001a 가 RED. 되돌림 후 GREEN. **호출 지점 배선**을 잰다(M1 §A 위험의 유일한 검출기 — 배선이 없으면 001a 가 초록으로 살아남고, 그 생존이 신호다).
- 뮤턴트 M-leak — **현 정의(비경합 분기에서 `unix.Close(fd)` 제거)로는 심을 자리도, 도달할 유도도 없다(§B.1 부채-1).** M1 이 그 정의를 다시 놓기 전까지 이 항목은 미확정이며, **AC-BLE-003 이 M-leak 으로 보증된다고 적지 않는다**. 그때까지 REQ-BLE-004 는 비공허성 가드가 없는 요구다.
- **정의된 두 뮤턴트**(M-broad·M-narrow)가 **서로 다른** AC 를 RED 로 만드는지 확인 — 같은 AC 가 둘 다에 반응하면 짝이 방향 하나만 재고 있는 것이다.
- 되돌림은 `git status --short` 로 확인.
- 게이트: `go vet` · `golangci-lint` · `go test ./internal/kanban/... -count=1 -cover` · `GOOS=windows GOARCH=amd64 go build ./...`.
- AC-BLE-004: M1 이전 기준선과 실패 집합 대조.

### M4 — 증거 정리 (Priority Medium · 기계적)

- `.moai/reports/t379/` 에 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) 증거.
- 워크트리에서 primary 로 반출할 증거는 반출 후 `cmp` 로 대조(gitignore 되는 산출물은 워크트리 폐기 시 유실된다).
- progress.md `§E.2`/`§E.3` 채움 — manager-develop 소관.

## §G Anti-Patterns (이 카드에서 특히)

- **"오분류 N건 제거"로 성과를 적기.** N 은 0 이다. 그렇게 적으면 측정하지 않은 것을 주장하게 된다.
- **한 방향만 잠그기.** AC-BLE-001a 만 통과하는 구현은 현행 결함 그 자체다.
- **도달 불가를 도달 가능으로 승격시키기.** `EBADF`/`EINVAL` 은 분류 술어의 입력일 뿐, 실호출에서 나온다고 적지 않는다.
- **`ENOLCK` 실동작을 잰 것처럼 적기.** 합성 errno 로 분류만 쟀다.
- **darwin 측정을 리눅스 판정으로 쓰기.** 리눅스는 CI 가 낸다.
- **범위 확장.** `internal/spec` 의 자매 substrate, 재시도 예산, wedge-clear 경로는 만지지 않는다.
- **`t372 §9.3` 의 `errors.Join` 축 재조사.** 도달 불가로 확인됐다(`spec.md §1.4`).

## §H Cross-References

- `internal/kanban/board_lock_unix.go:43` — 매핑 지점
- `internal/kanban/board_lock.go:28` — `IsBoardLockHeld`
- `internal/kanban/board_lock_windows.go:69` — 좁은 기준면
- 소비자 3곳: `board_store.go:173` · `integration_lock_mutation.go:103` · `backlog_store.go:736`
- 폴드(도달 불가 확인): `board_store.go:289` / `backlog_store.go:677`, 호출 지점 `board_store.go:237` · `board_recover.go:77` · `backlog_store.go:623` · `backlog_store.go:644`
- `.moai/reports/t372/verdict.md` §9.3 후보 B — 카드 기원 (`git show origin/develop:` 로만 접근 가능)
- `.moai/reports/t379/errno-probe-{source.go,output}.txt` — 계획 단계 실측
