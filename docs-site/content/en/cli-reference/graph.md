---
title: moai graph
weight: 16
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

A tool that gathers the codebase's relationships into one artifact and answers **reverse questions**. "If I touch this package, how far does the shake travel?" "Is this SPEC actually connected to code?" — questions grep cannot answer; they need the relationships gathered in one place.

{{< callout type="info" >}}
**One-line summary**: `moai graph build` gathers the relationships scattered across codemaps, @MX tags, SPECs, and reports into one file — `.moai/project/graph/edges.jsonl` — and `moai graph query` runs reverse queries against that file.
{{< /callout >}}

## Why it exists

MoAI-ADK already holds relationship information in several layers — the import graph in codemaps, the `@MX:SPEC` tags in code, the dependency declarations in SPEC documents, and the milestone records in reports. The problem is that these layers live scattered across different files. To ask "which SPECs are affected if I change this code," you must trace the import direction (files importing the file carrying the `@MX:SPEC` tag) and the SPEC-dependency direction backwards **in the same graph**. edges.jsonl is that one graph.

## moai graph build

```bash
$ moai graph build
```

Gathers import edges, `@MX:SPEC` links, and inter-SPEC dependencies into `.moai/project/graph/edges.jsonl`. It works deterministically — run it twice from the same git HEAD and you get the same content. Queries always read this artifact, so **run build first, before querying**.

On top of the document layers, the build derives edges straight from code — function-call edges (code-call) and import edges (code-import) — and every document edge survives unchanged. Import targets are normalized to repository-local package paths (the go.mod module prefix stripped) so code imports point into the same package domain as the codemaps import graph, and a 16-language grade matrix publishes, per language, the resolution level behind its call captures. Where the two layers disagree about the same relationship, neither is dropped: the doc edge stays with a `disagrees_with` marker naming the code layer's contrary observation, and `--all-disagreements` surfaces the suppressed direction too — local dependencies the code found and the docs stay silent about.

## moai graph query

One call takes **exactly one** selector.

| Selector | Question | Answer |
|--------|------|-----|
| `--callers <node>` | What depends directly on this package/SPEC? | Reverse neighbors — importing packages, depending SPECs, and code files carrying the `@MX:SPEC` tag |
| `--blast <node>` | If I edit here, how far does the shake travel? | The blast radius, swept wide over reverse edges (BFS). `@MX:SPEC` edges propagate in both directions, reaching the SPEC a code file implements |
| `--fanin [--limit N]` | Which packages are used the most? | Import fan-in ranking |
| `--debt-fanin [--limit N]` | How called-into are `@MX:DEBT` targets? | `@MX:DEBT` tag targets ranked descending by evidence-backed call fan-in — a file-scope DEBT is listed at fan-in 0 with a `(self)` marker |
| `--specs-no-code` | Which SPECs are not connected to code? | SPECs with zero `@MX:SPEC` edges in edges.jsonl |
| `--milestones-no-card` | Which milestones passed without a card? | Milestones whose card cross-check row claims no card, or whose claimed card is absent from the live backlog queue |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --debt-fanin
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

`--edges <path>` points at a different edges.jsonl, and a positional root argument selects a different project root.

Before answering, a query refreshes the mechanical layers (the @MX index and edges.jsonl) when they have gone stale. Only files whose content hash moved are re-parsed, so uncommitted edits are reflected in the answer without a rescan; a refresh whose measured cost exceeds gate.yaml's `update_budget_ms` (default 2000ms) warns and still answers. Every answer prints the tree root and commit (or dirty fingerprint) it was computed from, so there is no mistaking which tree answered.

## moai graph check

```bash
$ moai graph check
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
```

Measures how far the graph's three layers — codemaps, the @MX index, edges.jsonl — have fallen behind the code, each by its own metric, and returns a `fresh` / `stale` / `absent` verdict per layer. Codemaps is judged by described-source files changed since the stamped generation commit (reverted churn counts zero), the @MX index by files whose content hash moved, edges.jsonl by source-fingerprint mismatch.

Every generated artifact declares, in a provenance block, which tree and commit it describes. An artifact without one is reported `absent` — unjudgeable, never silently fresh — and absent fails the check too: a fresh worktree holds none of these untracked artifacts, and the check says so instead of passing. Exit codes: 0 all fresh · 1 stale or absent · 2 system error. The pre-commit quality gate's graph-freshness step and the CI graph-freshness job consume this value directly. Thresholds are tuned in gate.yaml's `graph_freshness` section.

No filesystem mtime is read anywhere. A fresh checkout resets every mtime, which an mtime-based metric would misread as freshly regenerated — so every metric here is a content hash, a git diff, or a fingerprint.

A stale verdict now carries its own attribution. When the `codemaps` layer goes stale, stderr prints how many of the drifting files this change itself contributed (`contribution`) and the commit it was measured against (`contribution_base`, typically `HEAD^1`), followed by up to 10 driving paths — anything past that is summarized as `... and N more`. `--json` exposes the same data as the `contribution` / `contribution_base` / `driving_paths` / `driving_paths_omitted` fields. A single stderr line now tells a lane whether it merely inherited the red (contribution 0) or caused it (contribution > 0).

## moai graph stamp codemaps

```bash
$ moai graph stamp codemaps
OK: stamped .moai/project/codemaps/provenance.json
provenance: tree=/path/to/project commit=1a2b3c4d5e6
```

Run as the last step after regenerating codemaps. The content is curated by `/moai codemaps`; this command records **which tree state that content describes**, in `provenance.json` — the anchor `moai graph check` judges the codemaps layer against.

### Naming a merge-surviving commit (`--commit`)

```bash
$ moai graph stamp codemaps --commit "$(git merge-base HEAD origin/main)"
OK: stamped .moai/project/codemaps/provenance.json
provenance: tree=/path/to/project commit=1a2b3c4d5e6
```

By default the stamp records the checked-out HEAD. On a feature branch that is a trap: this repository merges pull requests with **squash merges**, so the branch's commits — including its HEAD — never enter main's history. A stamp naming a branch-local HEAD is orphaned the moment the squash lands, and every later pull request inherits a red graph-freshness check (`not comparable`, exit 2). That exact failure shipped once and was traced in the `0d15864ae90b` incident.

`--commit <rev>` accepts any `git rev-parse` expression (a full sha, a short rev, a ref), resolves it to the full sha, and records that sha verbatim. The merge-base recipe above is the safe form: `git merge-base HEAD origin/main` is an ancestor of main (so it survives the squash) **and** content-equal to your branch's described sources at the branch point (so the check does not count other PRs' merged churn as your drift). Never restamp against a branch-local HEAD — `--commit` with a dirty described-source tree is rejected outright, because a named commit and a content fingerprint are two different honesty claims and the schema carries only one anchor.

CI backs this discipline mechanically: the graph-freshness workflow verifies the tracked stamp's commit is an ancestor of the pull request's base branch before reporting any freshness verdict, so an orphan-bound stamp fails the check by name instead of surfacing as a generic exit 2 after the merge.

## Caveats for two selectors

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--specs-no-code`**: "unconnected" is not "unimplemented." Most SPECs deliver docs, rules, or harness work and are complete with no code. Read this list as a coverage map, not a defect list.
{{< /callout >}}

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--milestones-no-card`**: the backlog queue deletes a row when its card is `done`. So "not in the queue" lumps together **finished cards and cards never issued at all**. Adjudicate each item with `git log --oneline --grep 'merge: tNN'` — a commit means it passed; no commit means it is a new-card candidate. A zero-hit grep still does not mean "the work never happened." The card may have been reissued under a new id, so check the lineage before cutting a new card.
{{< /callout >}}

## Related docs

- [Kanban Mode](/en/advanced/kanban-mode) — the card flow the milestone-card cross-check watches over
- [`/moai mx`](/en/utility-commands/moai-mx) — the origin of @MX tags and `@MX:SPEC` links
- [Navigator](/en/core-concepts/navigator) — the other graph binding design decisions, SPECs, and symbols (nav-graph.json)
