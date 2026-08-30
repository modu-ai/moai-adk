---
id: SPEC-BINLAG-INVOCATION-001
title: 트리를 읽는 CLI 호출이 자기 지연을 밝힌다 — PATH 바이너리의 초록을 비실행과 구별하기
version: "0.1.0"
status: draft
created: 2026-08-30
updated: 2026-08-30
author: manager-spec
priority: Medium
phase: "v3.1.4 target"
module: internal/cli, internal/binlag
lifecycle: spec-anchored
tags: "binary-lag, cli-invocation, verification-evidence, advisory, fail-open, local-only"
tier: S
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-08-30 | manager-spec | 최초 작성(카드 t366, plan-phase). `.moai/reports/t366/discovery.md`와 `census.md`의 실측 위에 세움. 수리 방식은 **정하지 않는다** — 세 안(A/B/C)을 §6에 병기하고 Implementation Kickoff Approval 게이트가 고른다. 요구 8 / 수락 8 |

---

## §1 문제 — 초록이 통과가 아니라 비실행일 수 있다

셸에서 부르는 `moai` 하위 명령은 PATH를 통해 **설치본**으로 해석된다. 이 트리에서 실측한 값:

    $ which -a moai
    /Users/goos/go/bin/moai

    $ ~/go/bin/moai version
    v3.1.2  343399d2f  built 2026-08-27T14:07:38Z

트리는 `d7010f86a`다. 즉 `moai spec lint`처럼 **트리를 읽되 판정 규칙은 컴파일되어 있는** 하위 명령은 `343399d2f` 시점의 규칙으로 답한다. 그 결과가 초록이면 두 가지 중 하나인데, 출력만으로는 갈리지 않는다 — 규칙이 전부 돌아 통과했거나, `343399d2f` 이후에 착지한 규칙이 **애초에 실행되지 않았거나**.

이것이 이 카드가 닫으려는 결함이다. 결함은 판정의 **정확성**이 아니라 판정의 **귀속 가능성**에 있다: 초록을 증거로 인용하는 순간, 인용자는 자기가 무엇을 재지 않았는지 알 수 없다.

### §1.1 t326은 결함이 아니다 — 부트스트랩 역설이다

지연 판정 자체는 이미 존재하고 이미 정확하다.

    $ grep -rn checkBinaryFreshness --include='*.go' internal/ | grep -v _test.go
    internal/cli/doctor.go:201:  {"Binary Freshness", checkBinaryFreshness}
    internal/cli/doctor.go:518:  func checkBinaryFreshness(verbose bool) DiagnosticCheck

    $ ~/go/bin/moai doctor | grep Freshness
    warn  Binary Freshness  binary is behind source tree (binary: 343399d2f, HEAD: d7010f86a)

카드 t326(SPEC-BINARY-LAG-VISIBILITY-001, 착지 `c70c6aed9`)이 더한 것은 판정이 아니라 **도달**이다 — `internal/binlag` 패키지와, 묻지 않아도 말하는 세션 시작 권고. 그 권고가 지금 발화하지 않는 이유는 설계 결손이 아니라 설치 시점이다:

    $ git merge-base --is-ancestor 343399d2f c70c6aed9; echo $?
    0                       # 설치본 커밋이 t326 착지의 진조상

    $ strings ~/go/bin/moai | grep -ci binlag
    0                       # 설치본에 binlag 패키지가 없다

SessionStart 래퍼는 PATH의 `moai`를 실행하므로, 지연을 알릴 코드가 **지연된 바이너리 안에** 갇혀 있다. 이 역설은 **1회성**이다. `make build && make install` 한 번이면 이후 모든 세션에서 권고가 뜬다.

[HARD] 이 SPEC은 t326을 결함으로 규정하지 않는다. t326의 표면은 **세션 시작**이고, 이 SPEC의 표면은 **호출 시점**이다 — 서로 다른 축이며, t326이 설치된 뒤에도 남는 간극은 「세션 중간에 트리가 앞서 나간 뒤 호출되는 명령」이다.

### §1.2 범위는 `spec lint` 하나가 아니다

    $ grep -rln "findProjectRootFn\|os.Getwd()" --include='*.go' internal/cli/ | grep -v _test.go | wc -l
    68
    $ ls internal/cli/*.go | grep -v _test.go | wc -l
    199

199개 비테스트 파일 중 68개가 프로젝트 루트나 cwd를 읽는다. 파일 수 대리지표이지 **명령 수 census가 아니다**(`census.md` 귀속 주석). 이 숫자가 세우는 것은 오직 방향 하나다: 계층이 넓으므로 명령별 패치는 형태가 틀렸다.

명령 단위 census — 「트리를 읽고 그 답이 컴파일된 규칙에 의존하는 명령의 목록」 — 은 **존재하지 않는다**. 그 목록에 기대는 수리(§6 Option C)를 고르면, 목록을 만드는 것이 그 수리의 첫 비용이다.

### §1.3 CI는 이 결함을 드러낼 수 없다

    $ grep -rn 'moai ' .github/workflows/
    spec-lint.yml:40               go run ./cmd/moai spec lint --strict
    ci.yml:367,464                 go build -o ./bin/moai ./cmd/moai/ ; ./bin/moai ...
    graph-freshness.yml:78         go build -o ./bin/moai ./cmd/moai ; ./bin/moai ...
    spec-status-auto-sync.yml:85   moai spec status ...        (맨 PATH 호출)

    $ grep -n 'install' .github/workflows/spec-status-auto-sync.yml
    35:        run: make build && make install

모든 워크플로가 체크아웃한 소스에서 빌드하며, 유일한 맨 PATH 호출은 **같은 job의 35행**에서 `make build && make install`을 선행한다. 결함은 **로컬 전용**이다. 따라서 수리는 정확성 축이 아니라 **경고·가시성 축**에 놓이고, 심각도는 그만큼 중간이다.

### §1.4 뜨거운 경로 — 호출 계층의 대부분은 사람이 타이핑하지 않는다

    $ grep -rlc "exec moai hook\|moai hook " .claude/hooks/moai/*.sh | wc -l
    39                      # 43개 훅 래퍼 중 39개가 `moai hook <event>`를 exec 한다
    $ grep -n '"hook"' internal/cli/root.go
    (출력 없음, exit 1)     # `hook` 은 trivialCommands 건너뛰기 목록에 없다

즉 `moai`의 지배적 호출자는 사람이 아니라 훅이며, 훅 경로는 초기화 건너뛰기 대상도 아니다. 호출마다 저장소 비교를 붙이는 수리는 **턴마다 수십 번** git 하위 프로세스를 두 번씩 낳는다. 이 실측이 §6 Option A의 건너뛰기 집합을 요구 수준(REQ-BLI-006)으로 끌어올린 이유다.

---

## §2 요구 (GEARS)

**REQ-BLI-001** — The project shall not accept a tree-reading subcommand's output as verification evidence unless that output was produced by a binary built from the tree under measurement.

**REQ-BLI-002** — **When** a tree-reading subcommand is invoked from a binary whose build commit is a strict ancestor of the tree HEAD, the invocation shall make that lag observable to the invoker in the same invocation, without the invoker having typed a separate diagnostic command.

**REQ-BLI-003** — **Where** the executing tree provides no repository HEAD to compare against, the invocation shall emit nothing, and shall leave the subcommand's stdout and exit status byte-identical to the no-lag case. The not-applicable leniency of `internal/binlag` shall be preserved verbatim rather than re-derived.

**REQ-BLI-004** — The lag verdict shall have exactly one implementation, `internal/binlag`. The invocation-time notice shall consume that seam and shall not re-derive an ancestry comparison of its own.

**REQ-BLI-005** — The notice shall be written to stderr. The notice shall not alter stdout, shall not alter the exit status, and shall not be parseable as part of any subcommand's machine-readable output.

**REQ-BLI-006** — **While** an invocation belongs to the per-event hook class or the launcher class, the repair shall not add a per-invocation repository comparison to it. The existing `trivialCommands` list (`internal/cli/root.go:46`) shall be the starting point of the exempt set, and the `hook` command group shall be added to it rather than a new parallel list being invented.

**REQ-BLI-007** — Any criterion asserting that a class of subcommands emits the notice shall name its exceptions explicitly. **Where** cobra's `PersistentPreRunE` non-chaining behavior causes a subtree with its own handler to bypass a root-level handler, the SPEC shall either name that subtree as an exception or require the chain be made explicit; a universally-quantified criterion that is false for a known subtree shall not be written.

**REQ-BLI-008** — The verification shall build its own binary under a session-private temporary directory. The verification shall not modify, replace, or reinstall the shared installed binary at the user's `GOBIN` path.

### §2.1 요구 ↔ 수락 추적

| 요구 | 수락 |
|---|---|
| REQ-BLI-001 | AC-BLI-001 |
| REQ-BLI-002 | AC-BLI-002, AC-BLI-003 |
| REQ-BLI-003 | AC-BLI-004 |
| REQ-BLI-004 | AC-BLI-005 |
| REQ-BLI-005 | AC-BLI-006 |
| REQ-BLI-006 | AC-BLI-007 |
| REQ-BLI-007 | AC-BLI-008 |
| REQ-BLI-008 | AC-BLI-003 (재현 절차의 구속 조항) |

---

## §3 수락 기준

Tier S는 수락 기준을 `spec.md` §3에 인라인하는 것이 기본이지만, 이 SPEC은 **`acceptance.md`를 수락 기준의 정본으로 둔다**. 사유: REQ-BLI-002의 판정이 **양방향 재현**(경고가 뜨는 쪽 / 뜨지 않는 쪽)을 요구하므로, 한 criterion이 Given-When-Then 두 벌과 각각의 RED-now 셀을 갖는다. 인라인하면 §3이 문서 절반을 차지한다.

수락 기준 8건 전문: [`acceptance.md`](acceptance.md) §D. 이 절은 사본을 두지 않는다(드리프트 방지).

---

## §4 재현 설계 — 반드시 양방향

[HARD] 한 방향 재현은 **지어진 경고**와 **꺼진 경고**를 구별하지 못한다. 「경고가 뜬다」만 관측하면 경고가 항상 뜨는 구현도 통과하고, 「경고가 안 뜬다」만 관측하면 경고가 영영 뜨지 않는 구현도 통과한다. 이 카드가 다루는 결함이 정확히 **초록의 두 가지 의미가 갈리지 않는 것**이므로, 그 결함을 재현하는 절차가 같은 모양을 반복해서는 안 된다.

설계(실행은 run-phase 몫):

1. **다른 쪽** — 트리에는 있고 비교 대상 바이너리에는 없는 규칙을 하나 심는다. 그 바이너리로 트리를 읽는 하위 명령을 부른다. 경고는 **반드시 발화**해야 한다.
2. **같은 쪽** — 두 집합을 일치시킨다(바이너리를 그 트리에서 빌드). 같은 명령을 부른다. 경고는 **반드시 침묵**해야 한다.

[HARD] **재현은 자기 바이너리를 `/tmp` 아래에 빌드한다.** 세션은 공용 설치본 `/Users/goos/go/bin/moai`를 건드리지 않는다 — 아홉 개 레인이 그 경로를 공유하며, 조용히 갱신하면 다른 레인의 측정이 전부 움직인다(REQ-BLI-008).

---

## §5 제약

**제약 C-1** — 비교 로직을 새로 쓰지 않는다. `internal/binlag`의 `Evaluate` / `Verdict` / `Advisory`가 정본이며, 이 SPEC의 코드 변경은 **소비 지점의 추가**에 한정된다(REQ-BLI-004).

**제약 C-2** — 적용 불가(not-applicable) 관용을 **문구 그대로** 보존한다. 일치하는 빌드, 브랜치 빌드, 저장소 없음, 비교 초과, 비교 내부 패닉 — 전부 침묵이다. 배포 사용자 프로젝트의 모든 명령이 경고를 뱉기 시작하는 것이 이 SPEC이 만들 수 있는 최악의 회귀다(REQ-BLI-003).

**제약 C-3** — cobra의 `PersistentPreRunE`는 **연쇄하지 않는다**. 하위 명령이 자기 것을 정의하면 부모 것을 **대체**한다. `internal/cli/root.go:127`이 이미 `worktree.WorktreeCmd`에 하나를 설치했으므로, 루트 수준 권고는 `moai worktree ...`에서 **발화하지 않는다**. 이 사실은 각주가 아니라 수락 기준의 구속 조항이다(REQ-BLI-007 / AC-BLI-008).

**제약 C-4** — 발화 표면은 stderr이며, 어떤 하위 명령의 stdout도 오염시키지 않는다. `moai spec status`, `moai todo`, MCP 경로처럼 출력이 기계 판독되는 명령이 존재하기 때문이다(REQ-BLI-005).

---

## §6 수리 안 — 게이트가 고른다

[HARD] plan-phase는 **고르지 않는다**. 세 안을 비용과 폭발 반경과 함께 병기하고, Implementation Kickoff Approval 게이트가 결정한다. 셋 다 §1.3의 「로컬 전용·경고 축」 판정 아래 있으므로, 어느 것도 정확성을 주장하지 않는다.

### Option A — 호출 시점의 일반 seam

`rootCmd`에 `PersistentPreRunE`를 달아, 트리를 읽는 **어떤** 하위 명령이든 stderr로 지연을 알린다.

- **좋은 점**: seam이 비어 있다. `rootCmd`(`internal/cli/root.go:18`)는 `Use` / `Short` / `Long` / `Version` / `Run`만 정의하고 `PersistentPreRun*`가 없으므로, 합성 문제가 아니라 **빈 슬롯**이다. 계층 전체에 한 번에 닿는다.
- **비용 1 — 뜨거운 경로**: §1.4. 43개 훅 래퍼 중 39개가 `moai hook <event>`를 exec 하고, `hook`은 `trivialCommands`에 없다. 건너뛰기 집합을 정하지 않으면 턴마다 git 하위 프로세스가 수십 쌍 늘어난다. 선례 시간 상자는 `binaryLagJoinBound = 250 * time.Millisecond`(`internal/hook/session_start_binary_lag.go:29`)이지만, 그것은 **형태의 선례이지 비용의 측정이 아니다**.
- **비용 2 — 소음**: 배포 사용자에게 모든 명령이 시끄러워질 위험. C-2의 관용 보존이 이 비용을 막는 유일한 장치다.
- **비용 3 — 거짓 보편**: C-3. `worktree` 하위 트리는 발화하지 않는다. 「모든 하위 명령이 알린다」는 문장은 **쓴 대로 거짓**이다.
- **건너뛰기 집합의 출발점**: 새 목록을 발명하지 않고 `trivialCommands`(`root.go:46` — `--version` / `version` / `-v` / `help` / `--help` / `-h` / `completion` / `cc` / `cg` / `glm`)에서 시작한다. 세 런처는 비용이 아니라 **의미** 때문에 제외된다 — `syscall.Exec`으로 프로세스가 교체되므로, 거기 찍는 권고는 곧 버려질 프로세스가 쓰는 것이다. 여기에 `hook`을 더한다(REQ-BLI-006).

### Option B — 절차만, 코드 무변경

「증거로 인용하는 측정은 그 측정 대상 트리에서 빌드한 바이너리가 낸 것이어야 한다」를 규율로 못박고, 재현하는 레인은 자기 바이너리를 빌드한다(lane-8이 쓴 `/tmp` 패턴).

- **좋은 점**: 가장 싸다. 기계적으로 아무것도 바꾸지 않으므로 회귀 위험 0, 배포 사용자 영향 0.
- **비용**: 강제력이 없다. 이 카드를 낳은 세 번의 조우(t343 · t362 · t357 — 리드 배차문에서 온 것이며 이 세션의 측정이 아니다)가 전부 규율은 이미 존재하는데 발생했다면, 규율 추가는 같은 결과를 낳는다. 다만 그 전제는 **미검증**이다.
- 세 안 중 유일하게 다른 둘과 **배타적이지 않다** — A나 C를 고르더라도 함께 채택할 수 있다.

### Option C — 좁은 발화

증거로 인용되는 명령에만 권고를 붙인다(`spec lint` / `spec audit` / `gate` / `verify`).

- **좋은 점**: 뜨거운 경로를 건드리지 않는다. 소음도 최소.
- **비용 1 — census 부채**: §1.2. 「트리를 읽고 답이 컴파일된 규칙에 의존하는 명령」의 목록이 없다. 목록을 만드는 것이 이 안의 **첫 비용**이며, 223개 `Use:` 정의(107개 파일, 중첩 하위 명령 과다 계상)에서 사람이 타이핑하는 명령을 골라내는 작업이 선행한다.
- **비용 2 — 부식**: 목록에 기대는 수리는 새 명령이 추가될 때 조용히 낡는다. 그 낡음이 만드는 실패 모양은 이 카드가 닫으려는 것과 **같은 모양**이다 — 아무 신호 없는 비발화.

---

## §7 범위 밖

### Out of Scope — t326의 재조사
- `internal/binlag`의 판정 로직, 세션 시작 권고의 설계, 250 ms 조인 경계는 카드 t326이 이미 닫았다. 다시 열지 않는다.
- t326을 결함으로 규정하지 않는다(§1.1). 지금 권고가 뜨지 않는 것은 설치 시점의 1회성 역설이다.

### Out of Scope — 바이너리 재설치의 자동화
- `make build && make install`을 자동으로 수행하는 기능은 만들지 않는다. 판정과 발화까지가 범위이며, 수리는 사람의 몫이다.
- 공용 설치본(`/Users/goos/go/bin/moai`) 갱신은 이 SPEC의 어떤 절차에도 포함되지 않는다(REQ-BLI-008).

### Out of Scope — CI 축
- CI는 이 결함을 드러낼 수 없다(§1.3). 워크플로 파일을 이 SPEC이 수정하지 않는다.
- `spec-status-auto-sync.yml`의 맨 PATH 호출은 같은 job의 `make build && make install`이 선행하므로 결함이 아니다. 「방어적으로」 고치지 않는다.

### Out of Scope — `spec lint` 자체의 성능
- `~/go/bin/moai spec lint`가 이 트리에서 120초 안에 끝나지 않은 관측(`discovery.md` § 부수 관측)은 기록만 하고 다루지 않는다. 지연 축과 다른 축이며, 더 재지 않았다.

### Out of Scope — 명령별 census의 완성 자체
- 게이트가 Option C를 고르지 않는 한, 223개 `Use:` 정의를 사람이 타이핑하는 명령 목록으로 환원하는 작업은 수행하지 않는다. Option C가 살아남으면 그것이 그 안의 첫 마일스톤이다.

### Out of Scope — 실행 중 MCP 서버 프로세스
- 실행 중 `moai mcp-server` 인스턴스와 설치본의 대조는 `checkMCPServerVersion`이 이미 닫았다. 재구현하지 않는다.

---

## §8 잔여 위험

- **R-1** — 세 번의 조우(t343 · t362 · t357)는 리드의 배차문에서 왔고, 이 세션이 측정하지 않았다. 「이 결함이 실제로 몇 번 판정을 오염시켰는가」는 미검증이며, 셋 중 어느 것도 이 SPEC의 근거로 인용되지 않는다.
- **R-2** — 뜨거운 경로 비용(§1.4)은 **호출 빈도**로 세웠지 **호출당 비용**으로 재지 않았다. Option A가 선택되면 그 측정이 run-phase의 첫 항목이다.
- **R-3** — Option B의 「규율은 이미 있는데 발생했다」는 전제가 미검증이다(§6 Option B). 게이트가 B를 고른다면 그 전제를 먼저 확인하는 편이 낫다.
