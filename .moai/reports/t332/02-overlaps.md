# M4 — Overlap detection and relation recording (card t332)

Cross-comparison of the 62 in-scope cards against one another, performed over the one-sentence
premise restatements produced at M3 rather than over raw card text.

## §1 Result: zero relations recorded, and that is the finding

**No in-scope card duplicates, contains, absorbs, replaces, or conflicts with another.**

The queue's relation vocabulary is exactly four tokens — `contains` / `absorbs` / `replaces` /
`conflicts` — and every candidate the sweep surfaced turned out to be one of three *weaker* things:

| Shape found | Example | Why it is not one of the four |
|---|---|---|
| shared **origin** investigation | t242 / t243 / t244, each titled "[t216 조사 분리]" | they were **split out of** t216 precisely so they would be decided separately; `contains` would assert that doing t216 does them, which is the opposite of why they exist |
| shared **file or mechanism** | t260 ↔ t280, both on `.moai/lessons-inbox.jsonl` | t260 is about collection *scope*, t280 about drain *deployment*. Same file, disjoint work |
| shared **precondition** | t90 ↔ t196, both downstream of t88/M4 Codex wiring | t90 packages a plugin, t196 rewrites skill content for harness neutrality. Neither absorbs the other |

Recording any of these as `contains` or `absorbs` would have been a **false finding the operator
would then act on** — worse than recording nothing, because a wrong relation is indistinguishable
from a measured one once it is in the store. §A of the SPEC names "two cards describe one defect" as
one of three silent failure shapes; measured across all 62, **that shape is not currently present in
this queue.** That is a real answer, not an empty one.

## §2 Every overlap candidate, with the shared artifact named (AC-BH-014)

Reciprocal pairs (both cards independently named the other):

| Pair | Shared artifact / mechanism | Relation | Verdict |
|---|---|---|---|
| t90 ↔ t196 | the t88/M4 Codex wiring both are gated on | shared precondition | reading only |
| t237 ↔ t239 | `moai update` overwrite class (self-cross-referenced as "같은 결함형") | shared defect class | reading only |
| t242 ↔ t243 ↔ t244 (all → t216) | the t216 lane-6 hook-audit investigation they were split from | shared origin | reading only |
| t253 ↔ t254 ↔ t262 | `SPEC-CODEX-SESSION-MSG-001` as subject/worked example | shared subject dir | reading only |
| t260 ↔ t280 | `.moai/lessons-inbox.jsonl` | shared file, disjoint concern | reading only |
| t286 ↔ t287 | command/worktree guard regexes (issues #1658 / #1659) | adjacent defects, different artifacts | reading only |
| t295 ↔ t297 (both → t231) | the worktree/launcher subsystem; both cite t231 | shared subsystem | reading only |
| t313 ↔ t319 | claimed "same config surface" | **weak** — t313 is worktree baseRef, t319 is an orphaned `tab_schema.json`. Different work | reading only |
| t324 ↔ t325 | the `develop`→`main` promotion boundary under git-flow | shared boundary, different mechanisms (branch protection vs a workflow's push path) | reading only |
| t344 ↔ t345 | `t241` verdict.md's follow-up candidate table | shared source | reading only |

One-directional candidates naming an in-scope counterpart: t223 → t313, t255 → t237, t263 → t216.
Candidates naming an **out-of-scope** counterpart, reported and carried no further: t154 → t155,
t281 → t310, t300 / t304 → t291, t323 → t317, t327 → t322, t361 → t352, t363 → t294 (`picked`).

Entries reporting no overlap: t288, t296, t302, t305.

**No row above proposes a fold.** Every one is a reading; the fold is the operator's act.

## §3 The pre-existing `t313 contains t295` relation — its substance is now falsified

The queue already carries two relations (`sqlite3 backlog.db "select ... from findings;"` → 2 rows):

| subject | related | relation | recorded |
|---|---|---|---|
| t313 | t295 | `contains` | 2026-08-27T09:48:03Z |
| t318 | t256 | `absorbs` | 2026-08-28T13:36:27Z |

**`t318 absorbs t256`** is out of the disposition proposal entirely on both sides: `t318` is now
`picked` (a live lane owns it) and `t256` is `dropped` (already decided). Reported as a reading;
no un-drop is proposed and no fold across the boundary is suggested.

**`t313 contains t295`** is the one relation between two in-scope cards, and the sweep measured
whether it still delivers. It does not:

```
$ git merge-base --is-ancestor 62ff3c2e6 ee50984abe4f11ac337382b48a26328f091e200a ; echo $?
0                       # t313 LANDED on pinned develop (SPEC-WORKTREE-BASEREF-001)

$ sed -n '248,254p' internal/cli/session_worktree.go
func gitWorktreeAddArgs(destDir, branch, base string) []string {
	args := []string{"git", "worktree", "add", "-b", branch, destDir}
	if base != "" {
		args = append(args, base)
	}
	return args
}
```

The landed work added a **base operand**; the `-b <branch>` that forces creation of a *new* branch
is still unconditional, so there is still no path that checks out an **existing** branch — which is
exactly t295's stated gap. **t313 landed and did not resolve t295.**

Shared artifact: `internal/cli/session_worktree.go` → `gitWorktreeAddArgs`.

### §3.1 The annotation was attempted and refused — a finding in its own right

The sweep tried to append that measurement to the existing relation:

```
$ moai todo relate t313 t295 --relation contains --note "[t332 스윕 실측 2026-08-30] …"
ERROR
Mutate backlog /Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.json:
mutation refused: todo relate: t313 contains t295 is already recorded.
```

No write occurred. Two observations follow, both reported rather than acted on:

1. **A relation, once recorded, cannot be annotated with later evidence.** The duplicate guard is
   keyed on (subject, related, relation) and refuses the pair regardless of the note. So a relation
   recorded on 2026-08-27 stays frozen even after measurement undermines it — **the same rot this
   card exists to catch, one layer down, in the mechanism the card was told to use for recording.**
   The refusal is not a bug in the sense of being wrong to refuse a duplicate; the gap is that there
   is no way to say "still recorded, and here is what changed."
2. **The error message names `backlog.json`** while the store this project actually writes is
   `backlog.db` (SQLite, since `3cb258d62`). A cosmetic path label, but it points a reader at the
   dead file — the same file that made AC-BH-006 vacuous. Reported, not fixed (out of scope, §C).

## §4 Evidence sections

**Claim.** After comparing all 62 in-scope cards pairwise, no pair qualifies for any of the four
relation tokens; zero relations were recorded. The one pre-existing in-scope relation
(`t313 contains t295`) has landed on its subject side without resolving its object side, and the
attempt to annotate it was refused by the CLI.

**Evidence.** The candidate table of §2, derived from the 62 M3 entries' `Overlap candidates` lines
(`awk` extraction over `input/all-cards.md`); the card texts of t90/t196, t216/t242/t243/t244,
t260/t280, t324/t325, t344/t345, t253/t254/t262, t286/t287, t295/t297/t231, t313/t319 read verbatim
from the run-phase snapshot; the two `--is-ancestor` and `sed` outputs of §3; the verbatim CLI
refusal of §3.1.

**Baseline-attribution.** All git queries against the refs pinned in `00-tooling-baseline.md`
(`origin/develop` `ee50984ab…`, `origin/main` `48239c7dc…`, fetched once at 2026-08-30T11:16:22Z);
source read at worktree HEAD `6165f9f5e`; card text from `queue-snapshot-run.tsv` captured
2026-08-30T11:16:45Z; existing relations read from the live SQLite store in this run.

**Gaps.**
- The pairwise comparison was over the M3 **premise restatements** plus targeted reads of the card
  texts named in §2 — it was not an all-pairs read of all 62 full card texts (1,891 pairs). A
  duplicate pair that neither worker noticed and whose restatements do not resemble each other would
  be missed.
- t223 → t313, t255 → t237 and t263 → t216 were accepted as one-directional readings without
  re-reading the counterpart card's full text to test reciprocity.
- Whether the `relate` duplicate guard also refuses a *different* relation token on the same ordered
  pair was not probed — doing so would have written to the store for no purpose.

**Residual-risk.** The zero-relation result is a statement about the four-token vocabulary, not
about the queue's redundancy in a looser sense: several cards share a subsystem closely enough that
dispatching them in parallel could still collide (t295/t297 in the worktree launcher, t324/t325 on
the promotion boundary). The sweep reports those adjacencies; scheduling them apart is the
operator's call, and nothing here records it as a relation because none of the four tokens says
"schedule these apart."
