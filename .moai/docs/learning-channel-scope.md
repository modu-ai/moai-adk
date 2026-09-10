# Learning Channel Scope — the lessons-inbox boundary and the human loop

> Anchor document for the bounded capture-scope claim (SPEC-LEARN-CHANNEL-SCOPE-001).
> Prose surfaces describing `.moai/lessons-inbox.jsonl` state the bounded claim in brief or
> point here; the measured numbers live in this document only (dated baseline — re-verify,
> do not trust). Card: t260.

## The bounded claim (capability + composition)

The lessons-inbox (`.moai/lessons-inbox.jsonl`) records **failure-event stubs only** — the two
wired families `tool_failure:<tool>:<sig>` and `test_fail:<pkg>:`, produced by the stub
appenders at `internal/hook/failure_observer.go:77` (`recordToolFailureEvent`) and `:111`
(`recordTestFailEvent`), both appending through the same inbox append path. The `test_fail`
family has a live call path: the evidence writer's `rec.IsTestFail` branch at
`internal/hook/evidence_writer.go:583-591`.

**Capability and composition are different statements.** The wiring above is capability (what
the channel can record). The composition — what the live file actually contains — is a dated
baseline measured below, not a structural property, and it may drift as events arrive. Neither
statement alone licenses the other: a 100%-`tool_failure` composition does not mean the writer
emits only `tool_failure`, and the two wired families do not mean tool failures dominate the
live file.

The inbox **shall not be presented as capturing defect families that produce neither a tool
failure nor a test failure.** The expensive families observed in this repository — vacuous
green checks, skip-before-verdict, empty-result-set pass conditions, sibling repair misses,
moving-ref invariant assertions, stale-value quotations — surface as neither, and therefore
have zero rows in the inbox by construction.

## The human-mediated loop — the channel for tool-invisible families

The **human-mediated loop** — lane discovery → lead judgment → auto-memory `feedback_*.md`
topic file + `MEMORY.md` index record — is the learning channel that carries the defect
families invisible to both wired families. This is a documentation of the flow that already
exists, not a new mechanism:

- The loop's upstream evidence sources are card progress records
  (`.moai/reports/<card-id>/`, `.moai/specs/<SPEC-ID>/progress.md`) and verdict reports
  (plan-auditor / sync-auditor outputs). They remain what they already are — evidence a human
  reads when writing a lesson; nothing extracts from them automatically.
- The recording target is the existing auto-memory convention (`feedback_*.md` + `MEMORY.md`
  index); no new format, schema, or writer is introduced.

## Dated baseline (re-verify, don't trust)

Measured **2026-09-02** on worktree `WT-learn-channel-gap` @
`d7ce6c6bd8dcc5f48a9ab46555f52d14e68540d9`. The inbox itself is primary-checkout live runtime
state — it does not exist in worktrees and is never copied into one.

| Observation | Command (re-runnable) | Result at baseline |
|---|---|---|
| Composition (whole-file tally) | `jq -r '.event_key // "NO_EVENT_KEY"' /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl \| cut -d: -f1-2 \| sed 's/:$//' \| sort -u` | every bucket `tool_failure:*` — zero rows outside the `{tool_failure, test_fail}` family set |
| File size at measurement | `wc -l < .moai/lessons-inbox.jsonl` (primary path) | 5,958 rows |
| `test_fail` rows | `grep -c 'test_fail:' .moai/lessons-inbox.jsonl` (primary path) | `0` (exit 1 — wired capability, unobserved in this file) |
| Human-channel yield | `find <memory-store> -maxdepth 1 -name 'feedback_*.md' \| wc -l` | 165 files; 146 created since 2026-08-25 |

The durable form of this claim is the **re-verifiability pair** — the tally command plus the
family-set predicate (zero rows outside `{tool_failure, test_fail}`) — not the count. The
inbox grows append-only, so a larger row count with the same zero-outside-families result
*confirms* the claim; only a row outside the family set falsifies it. Baseline staleness is
explained by the measurement date above, not by a count comparison.

## Non-goals (boundary, REQ-LCS-005/006)

- No third stub family, no new hook or writer, no automated extraction pipeline from verdict
  reports or cards, no new record format. If machine extraction ever measures as necessary,
  that is a separate SPEC — the boundary recorded here is what it would amend.
- The inbox stays append-only; the drain pipeline (wrapper → `drain.sh` → `clusters.json`
  staging) is behaviorally unchanged by the scope declaration.

## Claim surfaces (pointer list)

Surfaces describing the inbox's capture scope carry the bounded claim in brief or point here.
Numbers live only in this document:

1. `.claude/rules/moai/core/moai-constitution-detail.md` § Lessons Protocol — and its template
   mirror under `internal/template/templates/` (principle-level line only; the mirror omits
   dev-local measurement values per template neutrality).
2. `.claude/skills/hns-lsel-curator/SKILL.md` — bounded scope note + anchor pointer
   (dev-only skill, no mirror).
3. `CLAUDE.local.md` §28 (LSEL operating section) — bounded inbox-scope clause + anchor
   pointer (local-only file, never mirrored).
