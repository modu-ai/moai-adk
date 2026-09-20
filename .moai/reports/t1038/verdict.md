# t1038 — 워크트리 소유자를 세션에게 직접 묻는 경로

card: t1038
측정 시점: 2026-09-20T06:43Z ~ 06:55Z (UTC)
측정 트리: `.claude/worktrees/t1038` (branch `WT-owner-probe`), 탐침 대상은 이 머신의 전 세션

---

## 1. Claim

**「지금 누가 어느 트리에 앉아 있는가」는 세션에게 메시지로 물을 것이 아니라 OS 에게 그 세션의 프로세스를 물어야 한다.** 배차문은 교차 세션 메시지를 유력한 표면으로 지목했으나, 측정 결과 그 표면의 주소층(`ListAgents`)이 이미 17건 중 2건에서 죽은 이름을 살아 있다고 보고한다. 메시지를 보내기 전에 주소가 틀려 있다.

대신 상태 파일을 한 줄도 읽지 않는 라이브 경로가 존재하고, 그것이 카드가 요구한 다섯 성질을 전부 만족한다:

```
/tmp/cc-socks/<pid>.sock  →  kill -0  →  ps(--name|-n)  →  lsof -d cwd  →  git rev-parse --show-toplevel
   도달 가능 세션 집합       살아 있는가    세션 이름        현재 cwd          cwd → 트리 정규화
```

구현: `.moai/reports/t1038/who-sits-where.sh` (33행 POSIX sh). 전수 17세션에 **1.84초**.

부수적으로 **인계 §7 의 판독 하나를 정정한다.** 「`lsof`+`ps` 는 `agent-17`(이름 아님)」이라고 적혀 있으나 `agent-17` 은 **실재하는 세션 이름**(pid 51444)이고, 그 세션의 라이브 cwd 는 `.claude/worktrees/t1020` 이다. 즉 t1020 에서 갈린 세 경로 중 **`lsof`+`ps` 가 맞았다.** 틀린 것은 그 출력을 이름이 아니라고 읽은 판독이다.

## 2. Evidence

### E1 — 레지스트리 경로가 33% 틀린다 (15건 중 5건)

`moai session list --json` 은 `cwd` 를 담지만 **세션 시작마다 행을 덧붙이고, cwd 는 그 시점 값이다.** pid 16개에 행 143개.

```
$ moai session list --json | jq '[.[].pid]|unique|length'   → 16
$ moai session list --json | jq '[.[].session_id]|unique|length' → 143
$ moai session list --json | jq -r '.[]|select(.pid==48873)|"\(.started_at)  \(.cwd)"'
2026-09-17T16:09:16Z  .../worktrees/t854
2026-09-17T18:37:49Z  .../worktrees/t739
2026-09-17T23:27:22Z  .../worktrees/t905
2026-09-18T00:15:22Z  /Users/goos/MoAI/moai-adk-go
2026-09-18T03:14:23Z  .../worktrees/t915
2026-09-18T05:05:33Z  .../worktrees/t915
2026-09-18T12:25:58Z  .../worktrees/t915
2026-09-18T12:28:46Z  /Users/goos/MoAI/moai-adk-go
2026-09-20T00:54:07Z  /Users/goos/MoAI/moai-adk-go
```

pid 48873 하나에 서로 다른 cwd 가 아홉 줄. 최신 행만 골라 라이브 값과 대조하면:

| pid | name | 라이브 cwd (`lsof`) | 레지스트리 최신 행 | 일치 |
|---|---|---|---|---|
| 34357 | agent-11 | t810 | t810 | ✓ |
| 34481 | agent-12 | t1032 | t1032 | ✓ |
| 34873 | agent-13 | primary | **t1019** | ✗ |
| 35258 | agent-14 | t958 | t958 | ✓ |
| 35685 | agent-15 | t1025 | **`.claude/rules/moai/workflow`** | ✗ |
| 35942 | agent-16 | primary | primary | ✓ |
| 39066 | agent-2 | t1031 | t1031 | ✓ |
| 40001 | agent-4 | t1029 | **primary** | ✗ |
| 45693 | agent-5 | t1036 | t1036 | ✓ |
| 47166 | agent-6 | t1035 | **primary** | ✗ |
| 47786 | agent-7 | t1031 | **`.moai/reports/t1031`** | ✗ |
| 48258 | agent-8 | t1037 | t1037 | ✓ |
| 48438 | agent-9 | t1034 | t1034 | ✓ |
| 48873 | agent-10 | primary | primary | ✓ |
| 51444 | agent-17 | t1020 | t1020 | ✓ |

**5/15 불일치.** 그리고 「일치」쪽도 그냥 믿을 수 없다 — 레지스트리 cwd 는 트리가 아니라 그 순간의 셸 위치라서, `agent-7` 의 `.moai/reports/t1031` 은 이름에 `t1031` 이 들어 있지만 **primary 체크아웃 안의 경로**다. 정규화(`git rev-parse --show-toplevel`) 없이 문자열로 비교하면 정반대 결론이 난다.

### E2 — `ps` 경로의 구멍과 그 수리

처음 만든 표는 `--name` 만 봤고 `lead` 를 놓쳤다. `lead` 는 단축 플래그를 쓴다:

```
$ ps -o command= -p 87645
claude --permission-mode bypassPermissions --no-chrome --model claude-opus-5[1m] -n lead -r lead --settings ...
```

`(--name|-n)` 로 넓혀 재측정하니 `ps` 가 보는 이름 16개:

```
agent-10 agent-11 agent-12 agent-13 agent-14 agent-15 agent-16 agent-17
agent-2 agent-4 agent-5 agent-6 agent-7 agent-8 agent-9 lead
```

### E3 — `ListAgents` 가 죽은 이름을 살아 있다고 보고한다

같은 순간 `ListAgents` 는 `agent-1 [2421a3] · interactive · busy · started 1d ago` 를 출력했다. 그런데:

```
$ ps -eo pid,command | grep -cE '[c]laude .*--name agent-1( |$)'   → 0
$ ps -eo pid,command | grep -cE '[c]laude .*--name agent-10( |$)'  → 1   ← 양성 대조
$ ps -eo pid,command | grep -cE '[c]laude .*--name agent-3( |$)'   → 0
```

양성 대조(`agent-10`, 내 세션)가 1을 내므로 패턴 자체는 작동한다. `agent-1`·`agent-3` 은 **살아 있는 프로세스가 없다.** 소켓 쪽도 같은 말을 한다 — 31개 중 살아 있는 pid 는 17개.

이것이 메시지 표면을 기각하는 근거다. 도달 보장이 없다는 성질 이전에, **주소층 자체가 2/17 틀려 있다.** 죽은 이름에 보낸 질의는 답이 오지 않고, 답이 안 오는 것은 「그 트리에 없다」와 구별되지 않는다.

### E4 — 카드의 표본에 대한 소급 검사: **탐침이 잡는다**

카드는 2026-09-20 의 t1031 중복 배차를 표본으로 주고 「판별식이 이 상황을 사전에 잡을 수 있었는지 재라」고 했다. 탐침 출력 (`.moai/reports/t1038/live-probe.log`):

```
39066   agent-2      /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1031
47786   agent-7      /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1031
```

**같은 toplevel 이 두 줄.** 두 세션이 지금도 t1031 에 함께 앉아 있다. 배차 직전에 이 한 줄을 돌렸다면 `agent-7` 에게 t1031 을 주기 전에 `agent-2` 가 이미 거기 있다는 것이 보였다. 판정: **소급 검사 통과.**

반대로 레지스트리 경로는 같은 순간 `agent-7` 을 `.moai/reports/t1031`(primary)에 놓았으므로 **충돌을 보지 못했다.**

### E5 — 변이 대조: 탐침이 이동을 따라오는가

이 세션(`agent-10`, pid 48873)을 primary → `.claude/worktrees/t1038` 로 옮기고 같은 탐침을 다시 돌렸다.

```
이동 전:  48873   agent-10     /Users/goos/MoAI/moai-adk-go
이동 후:  48873   agent-10     /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1038
```

파일 쓰기는 한 번도 없었다. `EnterWorktree` 는 프로세스의 cwd 를 바꾸고, `lsof` 는 그것을 즉시 읽는다. **인계 §7 이 기술한 기전 — 「락은 풀리지만 프로세스는 살아남아 pid→트리 매핑이 이동을 따라오지 못한다」 — 은 파일을 읽는 경로에만 해당한다.** 프로세스에 직접 물으면 이동이 곧 답이다.

## 3. Baseline-attribution

모든 수치는 이 실행, 이 머신, 2026-09-20T06:43Z~06:55Z 사이에 측정했다. 명령은 §2 에 전문으로 적혀 있다. 전수 관측 한 벌은 `.moai/reports/t1038/live-probe.log` 에 반출했고, 탐침 자체는 `.moai/reports/t1038/who-sits-where.sh` 다. 인용한 인계 §7 의 문장은 `.moai/reports/lead/handoff-20260920.md` 의 것이다.

carry-over 한 값은 없다. 표의 「라이브 cwd」와 「레지스트리 최신 행」은 같은 턴에 나란히 잰 값이다.

## 4. Gaps

관측하지 **않은** 것:

1. **다른 호스트.** 소켓도 `ps` 도 이 머신 안에서만 답한다. 레지스트리는 `host` 필드를 갖고 있으므로, 원격 세션이 있는 배치에서는 탐침이 원리상 못 본다. 이 저장소는 단일 머신이라 지금은 문제가 되지 않는다.
2. **Linux / WSL.** `lsof -a -d cwd` 와 `ps -o command=` 는 darwin 에서만 쟀다. Linux 는 `/proc/<pid>/cwd` 가 더 싸고 확실하지만 **재지 않았다.**
3. **cwd ≠ 쓰기 대상.** 탐침은 세션이 **앉아 있는** 트리를 답한다. 절대경로로 다른 트리에 쓰는 세션은 잡지 못한다. 카드가 「지금 누가 어느 트리에 앉아 있는가」로 범위를 좁혔으므로 범위 안이지만, 이것을 「쓰기 충돌 탐지」로 확대 인용하면 틀린다.
4. **제품 편입.** 이 카드는 설계 판단이다. `moai` CLI 에 서브커맨드로 넣을지, 훅에서 자동 발화할지, 리드 배차 절차에만 둘지는 **결정하지 않았고 코드도 건드리지 않았다.**
5. **`agent-1`·`agent-3` 이 왜 `ListAgents` 에 남았는가.** 죽은 프로세스라는 것만 확인했고, 런타임이 왜 그 이름을 busy 로 보고하는지는 재지 않았다.

## 5. Residual-risk

- **한 번의 판독에는 유효기간이 있다.** 측정 중 `agent-12` 가 `t1032` → `.claude/worktrees/develop` → primary 로 두 번 움직였다. 탐침이 틀린 것이 아니라 세계가 움직인 것이다. 그러나 이는 **배차 직전에 돌려야 한다**는 뜻이다 — 10분 전 출력은 근거가 아니다. 레지스트리 경로의 결함과 정도의 차이일 뿐 종류가 다르지는 않다는 반론이 가능하고, 그 반론은 부분적으로 옳다. 차이는 **비용**이다: 1.84초짜리 재측정은 배차마다 돌릴 수 있고, 파일 기반 경로는 갱신 시점을 고를 수 없다.
- **`lsof` 가 느리고 권한에 민감하다.** 전수 1.84초의 대부분이 `lsof` 다. 세션 수가 늘면 선형으로 늘고, 다른 사용자 소유 프로세스는 읽지 못한다(이 머신은 전부 동일 사용자라 관측되지 않았다).
- **`ps` 명령줄 파싱은 런처 플래그 철자에 묶여 있다.** `--name` 만 보다 `lead` 를 놓친 것이 그 증거다. 런처가 세 번째 철자를 쓰기 시작하면 같은 방식으로 조용히 빠진다 — 그리고 **빠진 세션은 「트리에 아무도 없다」로 읽힌다.** 이 실패 방향이 위험한 쪽이다. 완화책은 소켓 집합을 모수로 쓰는 현재 구조다: 이름을 못 읽어도 `<unnamed>` 으로 행은 남으므로 트리 점유 자체는 보인다.
- **`<unnamed>` 이 실재한다.** pid 50574 가 이름 없이 `/Users/goos/MoAI/mo.ai.kr` 에 앉아 있다. 다른 저장소라 이 배치와 무관하지만, 이름 없는 세션이 이 저장소의 트리에 앉는 경우 배차 대상을 특정할 수 없다.

---

## 판정

**PASS** — 설계 판단 완료. 「세션에게 직접 묻는다」의 올바른 표면은 메시지가 아니라 OS 이며, 카드가 준 표본(t1031 중복 배차)에 대해 소급 검사를 통과했다. 코드 변경 0, 제품 편입은 범위 밖.
