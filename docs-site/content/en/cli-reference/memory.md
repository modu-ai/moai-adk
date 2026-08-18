---
title: moai memory
weight: 18
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

A **health-check and archive** tool for the Lessons Protocol memory store. MoAI lessons live as a `MEMORY.md` index plus topic files, and what a session reads at start is the index, not the directory. A lesson written as a file with no index line is therefore **stored but never recalled**.

{{< callout type="info" >}}
**One-line summary**: `moai memory doctor` diagnoses index↔file mismatches (orphan files, dead links) and the topic-file count against its cap; `moai memory archive` folds named files into the archive and takes them off the index.
{{< /callout >}}

## moai memory doctor

```bash
$ moai memory doctor            # human-readable report
$ moai memory doctor --json     # structured output
$ moai memory doctor --dir <path>   # diagnose a different store
$ moai memory doctor --cap 80   # check against a different cap
```

What it diagnoses:

| Item | Meaning |
|------|-----|
| Orphan topic files | Topic files with no index line — lessons that are never recalled |
| Dead index links | Index lines pointing at a missing file — lines that fail on read |
| Topic-file count | Current count against the per-project cap (default 50) |

`--dir` and `--cap` only change the target and the yardstick of the check — they fix nothing. doctor diagnoses.

## moai memory archive

```bash
$ moai memory archive feedback-old-lesson.md
```

Moves the named topic file into `memory/_archive/` and takes its index line down. **Not a delete** — archiving preserves the audit trail. What counts as stale is a judgment call, so the operator names the target by file directly; there is no automatic selection. Its main use is keeping the index light: when the cap (default 50) is exceeded, fold the excess into the archive.

## Related docs

- [Context and memory](/en/claude-code/context-memory/memory) — how Claude Code memory works
- [Self-evolution](/en/advanced/self-evolving) — the flow by which lessons rise into rule-promotion proposals
- [Decision memory](/en/advanced/decision-memory) — the other memory where routing decisions accumulate
