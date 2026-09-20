# t965 — `ExitWorktree` 복귀 지점 실측 기록 (plan 단계 조사)

- card: t965
- branch: `WT-exitworktree-return` (base develop `4a97dd2bf`)
- 측정 세션: agent-13, Claude Code, 2026-09-19
- 측정 방식: 이 세션 자신의 `EnterWorktree` / `ExitWorktree` 호출과 그 환경 갱신 보고

> [HARD] 카드의 물음은 「복귀 지점이 틀렸는가」가 아니라 **「A → B → C → Exit 에서 복귀가 B 인가
> A 인가」**다. 아래는 그 물음에 답하기 위한 관측이며, 결함을 전제하지 않는다.

---

## 실행 불가였던 지시 한 줄 — 격리된 임시 저장소

배차는 「재현은 격리된 임시 저장소에서」라고 지시했다. **그대로는 실행할 수 없다.**
`EnterWorktree` 는 현재 저장소에 묶여 있다 — `name` 형은 현재 저장소의
`.claude/worktrees/<name>/` 에 만들고, `path` 형은 **현재 저장소의** `git worktree list` 에
있는 경로만 받는다. 세션 밖의 임시 저장소를 겨눌 수단이 도구 표면에 없다.

가장 가까운 실행 가능한 형태로 대체했다: **이 저장소 안에, 이 카드가 소유한 버리는 워크트리**
(`t965-probe-b`)를 만들어 썼다. 살아 있는 남의 카드 트리는 건드리지 않았다(지시의 취지는 지켰다).

---

## 관측 1 — 단일 진입 (대조군)

```
A = /Users/goos/MoAI/moai-adk-go                     (primary checkout)
  EnterWorktree(name: "t965")        → .claude/worktrees/t965
  ExitWorktree(keep)
```

도구 응답: `Session is now back in /Users/goos/MoAI/moai-adk-go.`
환경 갱신: `Primary working directory: /Users/goos/MoAI/moai-adk-go (was .../t965)`

→ **A 로 복귀.** 단일 진입에서는 문서 서술과 일치한다.

## 관측 2 — 연쇄 진입 (본 실험)

```
A = /Users/goos/MoAI/moai-adk-go                     (primary checkout)
  EnterWorktree(name: "t965-probe-b")  → B = .claude/worktrees/t965-probe-b
  EnterWorktree(path: ".../t965")      → C = .claude/worktrees/t965
  ExitWorktree(keep)
```

도구 응답: `Your work is preserved at /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t965 ...
Session is now back in /Users/goos/MoAI/moai-adk-go.`
환경 갱신: `Primary working directory: /Users/goos/MoAI/moai-adk-go (was .../t965)`

→ **A 로 복귀했다. B(`t965-probe-b`) 가 아니다.**

**판별력 확인**: A 와 B 는 서로 다른 경로 문자열이므로 두 결과는 관측으로 구별된다. 구현이
「직전 워크트리로 복귀」였다면 응답은 `.../t965-probe-b` 를 지목했을 것이다. 이 픽스처는
공허하지 않다.

## 관측 3 — Exit 메시지가 지목하는 트리와 복귀 트리가 다르다

관측 2 의 한 응답 안에서:

- `Your work is preserved at ...` → **C** (방금 떠난 트리)
- `Session is now back in ...` → **A**

읽는 사람이 앞 문장만 보면 「C 에 있다」로, 뒤 문장만 보면 「A 로 갔다」로 읽는다. 원 사례에서
복귀 지점을 혼동한 모양과 정확히 일치한다.

## 관측 4 — 세션 출발 디렉터리조차 복귀 지점이 아니다

이 세션은 **워크트리 안에서 시작**했다(launch cwd = `.claude/worktrees/t788`, 운영자가
`moai cc -w` 로 띄움). 그 상태에서:

```
launch cwd = .claude/worktrees/t788        ← EnterWorktree 로 들어간 것이 아님
  EnterWorktree(path: ".../develop")       ← 이 세션의 첫 EnterWorktree
  ExitWorktree(keep)
```

→ 복귀 지점은 `t788` 이 아니라 **primary checkout** 이었다.

즉 관측 1~4 를 합치면 복귀 지점은 **연쇄 깊이와도 무관하고 세션 출발 디렉터리와도 무관하게
primary checkout 으로 고정**되는 것으로 보인다. 카드가 제시한 「B 인가 A 인가」 이분법에서
답은 A 쪽이지만, A 의 정체가 「직전 디렉터리」가 아니라 **「primary checkout」**이라는 점은
카드도 문서도 말하지 않는다.

---

## 문서 층 — 무엇이 비어 있나

`ExitWorktree` 도구 설명 두 문장이 서로 다른 방향을 가리킨다:

1. "return the session to the **original working directory**" — A 쪽을 시사한다.
2. "Restores the session's working directory to **where it was before EnterWorktree**" —
   **어느 `EnterWorktree`** 인지 말하지 않는다. 연쇄에서는 이 문장이 B 로도 읽힌다.

`EnterWorktree` 설명은 **정리 대상**에 대해서만 명시적이다: "the previous worktree is left on
disk, untouched, and **only the new one is tracked for exit-time cleanup**". 정리 대상은
한 문장으로 못 박아 두고, **복귀 지점은 같은 수준으로 못 박지 않았다** — 이것이 공백의 정확한
위치다.

또한 관측 4 가 보여 주듯, (2) 는 세션이 워크트리에서 시작한 경우에도 거짓으로 읽힌다.

---

## Gaps (측정하지 않은 것)

- **원 사례 재현 안 됨.** agent-5 는 다른 **워크트리**(`t912`)로 복귀했다고 보고했으나, 이
  세션의 관측은 모두 primary checkout 으로 복귀했다. 두 결과가 다르다. 원 사례는 `/clear`
  이전 구간을 읽을 수 없어 그 세션의 출발 디렉터리를 확인할 수 없고, 따라서 「그 세션의
  primary 가 무엇이었나」를 이 조사로는 채울 수 없다. **추정하지 않는다.**
- **관측 4 는 재실행하지 못한다.** 워크트리에서 시작하는 세션을 이 세션 안에서 다시 만들 수
  없으므로, 세션 이력의 1회 관측이다. 관측 1~3 은 이 조사에서 의도적으로 실행했다.
- `EnterWorktree(name:)` 가 워크트리 안에서 거부되는지는 **시험하지 않았다**(문서에 그렇게
  적혀 있으나 실행으로 확인하지 않음). 실험 설계는 primary 에서 생성하는 경로로 우회했다.
- 런타임 버전은 관측 **도중**에 재지 않았다. 리드 지시로 사후 기록했고(아래 「측정 환경」),
  그래서 이 항목은 닫힌 것이 아니라 **사후 귀속으로 약화된 채 남는다** — 특히 관측 4 는
  그 귀속조차 공유하지 않는다.

## Residual-risk

- 관측은 **도구 응답 문자열과 환경 갱신 보고**에 의존한다. 그 둘이 실제 프로세스 cwd 와
  어긋날 가능성은 이 조사에서 배제하지 않았다(배제하려면 Exit 직후 `pwd` 를 재야 한다 —
  다음 회차 설계에 넣을 것).
- 버리는 워크트리 `t965-probe-b` 가 디스크에 남아 있다. 이 카드의 실험 잔재이며, 정리
  대상이다.

---

## 측정 환경 (Gap 정정 — 리드 지시로 사후 기록)

| 항목 | 값 | 어떻게 얻었나 |
|---|---|---|
| Claude Code | `2.1.278` | `claude --version` |
| OS | macOS `27.0` / `Darwin 27.0.0` | `sw_vers -productVersion`, `uname -sr` |

**이 값은 관측 1~3 을 실행한 뒤 같은 세션에서 잰 것이다.** 관측 도중에 버전이 바뀌지 않았다는
것은 세션이 같다는 사실로부터 따르지만, 관측 **직전**에 재지 않았다는 점은 남긴다 — 엄밀하게는
사후 귀속이다.

**관측 4 는 이 귀속을 공유하지 않는다.** 그 관측은 같은 세션의 이력에서 나왔고 같은 런타임에서
일어났지만, 그 시점에 버전을 재지 않았다 — 위 값이 그때도 참이었다는 것은 세션 연속성에 기댄
추론이지 측정이 아니다.

**왜 적어 두나**: 복귀 지점 동작이 런타임 버전에 걸려 있다면, 이 관측들의 귀속은 여기서
갈린다. 다음 사람이 다른 버전에서 다른 결과를 얻었을 때 「구현이 바뀌었다」와 「측정이
틀렸다」를 구별할 수 있는 좌표가 이 표뿐이다.
