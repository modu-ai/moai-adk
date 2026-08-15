---
title: moai epic Epic Progress
weight: 100
draft: false
description: "The moai epic status command, which computes milestone progress for a multi-SPEC epic from disk."
---

`moai epic status` works out how far an epic has got by **reading only what is on disk**. It keeps no separate progress store; it walks the SPEC documents each time and rebuilds the milestone map. It is read-only and modifies no files.

Once an epic is split across several SPECs, "how many are done" becomes hard to keep straight. Instead of opening each SPEC to check its status, this command answers in one line.

## Overview

```bash
moai epic status <prefix> [OPTIONS]
```

`<prefix>` is a required argument: the SPEC-ID prefix that identifies the epic. Passing `KANBAN`, for instance, targets `.moai/specs/SPEC-KANBAN-*/spec.md`.

## What it reads

Three things go into the milestone map.

1. **SPEC frontmatter** — the `status` value in each `spec.md`. This is the basis for deciding whether a milestone is finished.
2. **Milestone markers in titles** — the `(TOKEN Mx)` marker written in a SPEC title. This is what ties a SPEC to a milestone.
3. **A design report (optional)** — used as the source of the canonical milestone list when one exists. It is discovered automatically, and `--design-report` overrides that.

SPECs without a marker are not discarded; they are reported separately as `untracked_specs` — dropping them quietly would read as "that SPEC does not exist".

## Flags

| Flag | Description |
|------|-------------|
| `--json` | Emit the frozen-shape JSON document on stdout |
| `--design-report <path>` | Override design-report auto-discovery with an explicit path |
| `--marker <token>` | Override the inferred epic token (for example `BAS`) |
| `--base-dir <path>` | Project root (default: the current working directory) |

## Examples

The default output is a human-readable progress board.

```bash
$ moai epic status KANBAN
🎯 KANBAN ▓▓▓▓▓░░░░░ 2/4 (50%)
Epic progress:   KANBAN
  🟢 M0 M0                            SPEC-KANBAN-RENAME-001 (completed)
  ⬜ M1 M1                             SPEC-KANBAN-BOOTSTRAP-001 (draft)
  ⬜ M2 M2                             SPEC-KANBAN-WORKTREE-001 (draft)
  🟢 M3 M3                            SPEC-KANBAN-BOARD-001 (completed)
```

When there is no marker at all, it says so and lists the SPECs that matched instead.

```bash
$ moai epic status DESIGN-DOCS
🎯 DESIGN-DOCS — 2 SPEC(s) matched, none carrying a (TOKEN Mx) milestone marker
Epic progress:   DESIGN-DOCS
untracked_specs: SPEC-DESIGN-DOCS-001, SPEC-DESIGN-DOCS-V31-001
```

`--json` emits a frozen shape that scripts can rely on.

```bash
$ moai epic status KANBAN --json
{
  "epic": "KANBAN",
  "epic_token": "KANBAN",
  "milestones": [
    {
      "id": "M0",
      "label": "M0",
      "status": "done",
      "covered": true,
      "spec_id": "SPEC-KANBAN-RENAME-001",
      "spec_status": "completed",
      "sync_commit_sha": "144573336d07da19f4b8a50aa26c38db2704afb5"
    }
  ],
  "done": 2,
  "total": 4,
  "pct": 50,
  "extra_mx": [],
  "untracked_specs": ["SPEC-KANBAN-TODO-CLI-001"],
  "baseline_attribution": "3b9b3bf9959669c4bfc43da313e25bca61f910a2"
}
```

## JSON fields

| Field | Description |
|-------|-------------|
| `epic` | The prefix given as the argument |
| `epic_token` | The epic token found in title markers; an empty string when none was found |
| `milestones` | The milestone array. Each entry carries id, label, state, the owning SPEC, that SPEC's status, and the sync commit SHA |
| `done` / `total` / `pct` | Completed count, total count, percentage |
| `extra_mx` | Milestones that carry a marker but are absent from the canonical list |
| `untracked_specs` | SPECs that matched the prefix but carry no milestone marker |
| `baseline_attribution` | The git commit SHA at the time of computation |

`baseline_attribution` comes along so the result records **which tree it was read from**. A progress figure on its own says nothing about when it was measured.

## Read-only

The command only observes. It does not change any SPEC's status, stores the progress nowhere, and recomputes from disk on every run. That is why its result can never drift from the actual files.

The computed result also appears as the epic panel in the web console's Monitor area. See [MoAI Web Console](/en/advanced/moai-web-console/).

---

Related: [moai spec Document Management](/en/cli-reference/spec/) · [MoAI Web Console](/en/advanced/moai-web-console/)
