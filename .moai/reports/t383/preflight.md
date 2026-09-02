# t383 M0 — pre-flight (run-phase)

tree: `297a21ea7`, branch `WT-memory-index-budget`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t383`.

`$D` = `$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory`
`$L` = `$HOME/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory`

## The store moved again between plan-phase and run-phase entry

Every size figure moved; every defect figure held. This is the fourth independent
reading of the same asymmetry and it is now the best-evidenced claim in the card.

| Metric | M1 (measurements.md) | iter-2 | iter-3 | **M0 run-phase** |
|---|---|---|---|---|
| bytes | 26,280 | 26,290 | 26,577 | **28,009** |
| lines (`wc -l`) | 163 | — | — | **171** |
| chars (`wc -m`) | 18,463 | — | — | **19,556** |
| entry lines (`^- \[`) | 123 | 123 | 124 | **130** |
| unique targets file-wide | 189 | — | 190 | **196** |
| active topic files | 177 | — | 178 | **185** |
| **dangling (`.Code`)** | **58** | **58** | **58** | **58** |
| **orphans (`.Code`)** | **46** | — | — | **46** |

Nobody in this session touched the store before these readings.

## Commands and verbatim output

### Index size

```
$ wc -c < "$D/MEMORY.md"   ->    28009
$ wc -m < "$D/MEMORY.md"   ->    19556
$ wc -l < "$D/MEMORY.md"   ->      171
$ grep -c '^- \[' "$D/MEMORY.md"                                  -> 130
$ grep -o '](\([^)]*\.md\))' "$D/MEMORY.md" | sort -u | wc -l      -> 196
$ grep -c '(session: ' "$D/MEMORY.md"                             -> 7
```

`(session: …)` markers: **7** (plan-phase dated reference was 6; a concurrent session
added one). The command decides; the figure is a dated reference, 2026-08-31.

### Topic-file counts

```
$ find "$D" -maxdepth 1 -mindepth 1 -name '*.md' | wc -l   -> 185
$ find "$L" -maxdepth 1 -mindepth 1 -name '*.md' | wc -l   -> 1098
```

`find` is used rather than `ls | wc -l` deliberately: `ls` is aliased in this
environment and inflates the count.

The active count of 185 includes `MEMORY.md` itself; `moai memory doctor` reports
`topic_files: 184`, which excludes it. Both are recorded so the one-file difference is
not later read as drift.

### `moai memory doctor` — with `--dir`, gate asserted before any count

```
$ moai memory doctor --json --dir "$D" > /tmp/t383-doctor-preflight.json ; echo $?
0
$ jq -e '.[0] | select(.exists == true and .index_lines > 0)' ... >/dev/null && echo "GATE PASS"
GATE PASS
$ jq '.[0] | {dir: .store.dir, exists, topic_files, index_lines}' ...
{
  "dir": "/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory",
  "exists": true,
  "topic_files": 184,
  "index_lines": 171
}
$ jq '[.[0].findings[]? | select(.Code=="MEMORY_DANGLING_INDEX_LINK")] | length' ...   -> 58
$ jq '[.[0].findings[]? | select(.Code=="MEMORY_ORPHAN_NOT_INDEXED")] | length' ...    -> 46
```

Capital `.Code` per AC-MSR-009's case trap. The `exists`/`index_lines` gate passed
before any count was read.

## The occurrence sweep (plan.md §B.4, broadened form)

`grep -rn '25KB\|25 \* 1024\|25600\|200-line\|200 lines'` over
`.claude/rules .claude/output-styles CLAUDE.md AGENTS.md internal/config internal/template/templates/.claude`.

In-scope occurrences (the §A.5 enumerated set), all confirmed present at run-phase entry:

| File | Lines |
|---|---|
| `.claude/rules/moai/workflow/moai-memory.md` | 29, 117, 175 (`25KB`) |
| `.claude/output-styles/moai/moai.md` | 165 (`200-line/25KB`) |
| `internal/config/token_budget_guard.go` | 95 (comment), 98 (`25 * 1024`) |
| mirrors of the two `.claude/` files | same lines |

**Out-of-scope matches, recorded rather than acted on** (see verdict.md gap G6):

- `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-memory-official.md:125`
  — asserts "The first 200 lines of `MEMORY.md` (or the first 25KB, whichever comes first)
  are loaded at the start of every session." This is the **same unconfirmed claim** this card
  corrects, on a surface `spec.md` §A.5 does not enumerate and `REQ-MSR-001` does not reach.
  It is NOT edited here. Enumerated so it is not mistaken for handled.
- `moai-memory.md:156` — `MEMORY_INDEX_OVERFLOW | MEMORY.md exceeds 200 lines`. This describes
  a **lint rule's own threshold**, not the loader's cut, and carries no `25KB`. Left standing;
  it does not affect AC-MSR-001, whose `moai-memory.md` grep is for `25KB` only.
- `coding-standards.md:25`, `skill-writing-craft.md:133`, and several skill files match
  `200 lines` about CLAUDE.md / skill-file length — a different subject entirely. Not in scope.

## Gate outcome

M0 sampling gate: **PROCEED** — 0 of 12 superseded against a stop threshold of ≥ 4.
Detail in `m0-sample.md`.
