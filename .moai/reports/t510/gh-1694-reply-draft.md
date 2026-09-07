# GH #1694 회신 초안 (t510) — sync-phase 확정본

> 게시 조건: 수리가 develop에 착지하고 릴리스 일정이 확인된 뒤, 리드/운영자 확인을 거쳐 게시.
> 이 파일은 초안이며 게시된 상태가 아니다. (2026-09-07 sync-phase 확정, 카드 t510)

> **pre-run 초안 대비 변경점 (sync 확정 시)**:
> 1. `session-memo` 7건의 귀속을 **정정**했다 — pre-run 초안은 제보자 4개 계수를 "four writers"에
>    일치시켰으나, run-phase B7 범위 재판정이 session-memo는 이 결함의 방문-디렉터 오염이 아니라
>    **프로젝트당 compact 시점 기록**(훅 사슬의 project-dir env 앵커)임을 확인했다. 표에서 분리해
>    별도 문단으로 옮겼다. 수리 대상 4가족 = 세션 텔레메트리 / landed·github 카운트(같은 board root)
>    / goal 읽기 / config 캐시다.
> 2. 수리의 착지 형태를 명명했다 — 단일 시접 `internal/stateanchor`(고정 우선순위 체인).
> 3. run-phase에서 확정된 신규 사실 2건을 반영했다 — (a) 무프로젝트에서는 상태 쓰기를 생략한다
>    (no project, no state), (b) goal 상태 읽기도 앵커에서 하게 되어 cd한 세션이 프로젝트의 armed
>    goal을 다시 본다(가시성 수리).
> 4. 정리 레시피와 "display unchanged"는 pre-run 초안에서 불변이다.

---

Thanks for the exceptionally detailed report — the file-count table made the diagnosis
direct. This is confirmed as a defect, not intentional behavior, and the root cause is
now fixed in development.

**Root cause.** The statusline (and a small family of state writers) resolved its state
anchor from the session's current working directory — the statusline receives the
session's `current_dir` — instead of resolving up to the project root. The resolver
never walked upward, so every directory a session touched gained its own `.moai/state/`
on the next render. Your counts map to the defect exactly:

| Your count | Component |
|---|---|
| `context-usage` 213 | statusline session-telemetry write (fires on every status render — one record per visited directory) |
| `config-cache` 150 | config loader cache write (it only lands where a `.moai` already exists, so these followed the telemetry strays) |
| `github/counts` 133 | statusline GitHub counts refresh (shares one board root with the landed-counts cache) |

Your observation that only `state/` artifacts appear — never `config/` — was the key
confirmation: cache writes were anchored to the working directory, while genuine project
initialization always lands at the project root.

One entry in your table is *not* part of this defect: `session-memo` (7). Those are
per-project records written once per project at compaction time through the hook chain,
which resolves its own project directory correctly — they don't grow per visited
directory, and we left that writer untouched.

**The fix.** All of the affected writers now resolve their state anchor through one
shared resolver (`internal/stateanchor`) with a fixed precedence: the project directory
from the runtime payload → the worktree's original directory → a git root resolution
(one root shared by every checkout and worktree of the repository). The families routed
through it: the session-telemetry write, the board root feeding both the landed counts
and the GitHub counts cache, the armed-goal state read, and the config cache. When no
project can be resolved at all, the state write is skipped instead of landing in the
current directory. Two visible improvements ride along: a session that has cd'd
elsewhere now still sees the project's armed goal, and no new stray `.moai` directories
appear. The statusline display itself is unchanged. New strays stop appearing with the
release that carries this fix.

**Your questions:**

1. CWD-relative state was unintentional — a resolution bug, not per-directory caching.
   Answered by the fix above.
2. Cleanup: your instinct is right that wholesale deletion needs care. A safe recipe:
   delete any `.moai` directory whose contents are ONLY `state/` (no `config/` inside);
   keep any `.moai` that contains `config/` — that is a real initialized root. A
   dedicated cleanup command is being considered separately.
3. Template gitignore hardening: your observation is correct — the shipped pattern is
   `.moai/state/`, which (containing an inner slash) is anchored to the repository root
   and does not cover nested strays. With the fix, new strays stop appearing at the
   source, so we would rather not ship a broad `**/.moai/state/` pattern that could mask
   real state in legitimate nested roots. If strays from pre-fix versions are a concern,
   the cleanup recipe in (2) handles them explicitly.

We'll follow up here once the fix ships in a release.
