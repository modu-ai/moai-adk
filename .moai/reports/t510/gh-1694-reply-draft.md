# GH #1694 회신 초안 (t510)

> 게시 조건: 수리가 develop에 착지하고 릴리스 일정이 확인된 뒤, 리드/운영자 확인을 거쳐 게시.
> 이 파일은 초안이며 게시된 상태가 아니다. (2026-09-07 작성, 카드 t510 sync 단계 의무)

---

Thanks for the exceptionally detailed report — the file-count table made the diagnosis
direct. This is confirmed as a defect, not intentional behavior, and the root cause is
now fixed in development.

**Root cause.** The statusline (and a small family of state writers) resolved its state
anchor from the session's current working directory — the statusline receives the
session's `current_dir` — instead of resolving up to the project root. The resolver
never walked upward, so every directory a session touched gained its own `.moai/state/`
on the next render. Your counts map to the four writers exactly:

| Your count | Component |
|---|---|
| `context-usage` 213 | statusline session-telemetry write (fires on every status render) |
| `config-cache` 150 | config loader cache write |
| `github/counts` 133 | statusline GitHub counts refresh |
| `session-memo` 7 | compact-hook session memo |

Your observation that only `state/` artifacts appear — never `config/` — was the key
confirmation: cache writes were anchored to the working directory, while genuine project
initialization always lands at the project root.

**The fix.** All four writers now share one state-anchor resolution: project directory
from the runtime payload → worktree original directory → git root walk-up — and when no
project can be resolved, the write is skipped instead of landing in the current
directory. The statusline display itself is unchanged. New stray directories stop
appearing with the release that carries this fix.

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
