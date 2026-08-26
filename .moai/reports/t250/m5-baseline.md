# M5 Baseline — Grep/Read tool-use counts (AC-GF-022)

- **Measurement date**: 2026-08-25 (authored before the M5 implementation
  within the working session; committed together with it in the same commit
  — `7f2e9e77d` — so ordering rests on the authoring session record, NOT on
  git history, which cannot prove it for a same-commit pair).
- **Measuring session**: the t250 run-phase delegation session (this agent's
  session; `CLAUDE_SESSION_ID` is not exported to agent environments — recorded
  as a gap below).
- **Method**: acceptance.md specifies per-task counts of Grep and Read
  tool-use events "counted from the executing session's transcript, only
  those two tools". The sections below record what WAS measured, why a full
  per-task baseline was NOT obtainable, and the fixed task set the post-run
  (m5-post.md) will measure against.

## What WAS measured (session-transcript scan, Grep/Read only)

Mechanical `grep -c '"name":"<tool>"'` over session transcript JSONL files in the
developer-local Claude Code session store for this repository (machine-specific
path outside the repository; redacted from the committed report) — the 8 most
recent sessions at measurement time:

| Transcript (session id) | Date | Grep tool-use | Read tool-use |
|---|---|---|---|
| db215737-e6da-416f-bf0b-07fcb8a67fbb | 2026-08-23 | 0 | 0 (Bash/WebFetch-only investigation session) |
| 44400e2f-bb9f-4a89-9831-2686c5b61050 | 2026-08-16 | 0 | 12 |
| 0907b906-b37d-4468-85a5-4a79c07c71ea | 2026-08-16 | 0 | 7 |
| 1190d23b-afb5-49a7-9dd1-4799696da747 | 2026-08-16 | 1 | 20 |
| f2ff588f-da89-45e6-b682-d20bf5ba0291 | 2026-08-16 | 0 | 25 |
| ba7aff28-5e7b-483e-b4f9-0c7709857034 | 2026-08-16 | 0 | 0 |
| 7a43e2d9-30e4-4742-b817-5043767fbb51 | 2026-08-16 | 0 | 0 (4KB stub) |
| 072fa7bb-4d0f-427b-8980-f3a32cb94d86 | 2026-08-16 | 0 | 0 (queue/attachment-only) |

Headline observation (real, measured): this repository's agent sessions
perform code search via **Bash grep, not the Grep tool** — the Grep tool-use
count is 0 or 1 per session across the sample; Read runs 0–25 per session.
The search-cost surface the M5 tools compete against is therefore mostly
invisible to a Grep/Read-only counter.

## Fixed task set (defined here, used unchanged by m5-post.md)

1. **T1 file-api**: list the exported API (functions/methods, with signatures
   and files) of `internal/graph`.
2. **T2 callers**: find every caller of `CheckFreshness` and where the call
   sites live (file:line).
3. **T3 trace**: from `refreshEdgesArtifact`, trace which callers reach it
   and which functions it calls (one hop each direction).
4. **T4 find-code**: locate every definition and use of the resolution-grade
   constants (`full` / `name-based` / `none`) across the codebase.
5. **T5 file-api**: list the exported API of `internal/mx` (top-level
   surface only).

## Per-task baseline counts — NOT MEASURED (recorded as a gap, not fabricated)

A compliant baseline would count Grep/Read events per task above from a real
session that performed these tasks WITHOUT the M5 tools. No such session
exists in the transcript corpus (the tasks target code this SPEC itself
introduced; the Grep-tool baseline population is zero across the sample).
Manufacturing a "standard exploration routine" run by this agent — which
already knows every answer — would be a procedure simulation, not a
measurement, and is rejected per the delegation instruction ("do not
fabricate counts").

**Gap statement**: per-task Grep/Read baseline counts are unobtainable from
real prior sessions; the measurable baseline is the aggregate session-scan
table above. The post-run will therefore report (a) per-task M5-tool and
Grep/Read counts from a REAL post-implementation run (measurable — the tools
did not exist when the baseline corpus was produced), and (b) the same
aggregate session-scan shape for comparison. The reduction claim AC-GF-022
can support is bounded accordingly and stated honestly in m5-post.md.
