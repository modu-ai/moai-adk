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

## moai graph query

One call takes **exactly one** selector.

| Selector | Question | Answer |
|--------|------|-----|
| `--callers <node>` | What depends directly on this package/SPEC? | Reverse neighbors — importing packages, depending SPECs, and code files carrying the `@MX:SPEC` tag |
| `--blast <node>` | If I edit here, how far does the shake travel? | The blast radius, swept wide over reverse edges (BFS). `@MX:SPEC` edges propagate in both directions, reaching the SPEC a code file implements |
| `--fanin [--limit N]` | Which packages are used the most? | Import fan-in ranking — a stand-in for the @MX:DEBT fan-in query (per-tag-kind edges do not exist yet) |
| `--specs-no-code` | Which SPECs are not connected to code? | SPECs with zero `@MX:SPEC` edges in edges.jsonl |
| `--milestones-no-card` | Which milestones passed without a card? | Milestones whose card cross-check row claims no card, or whose claimed card is absent from the live backlog queue |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

`--edges <path>` points at a different edges.jsonl, and a positional root argument selects a different project root.

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
