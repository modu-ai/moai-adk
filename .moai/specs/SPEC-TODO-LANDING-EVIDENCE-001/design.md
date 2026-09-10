# Design — SPEC-TODO-LANDING-EVIDENCE-001

The shapes the requirements imply, stated once so run-phase does not re-derive them. This document
is subordinate to `spec.md` §B: where the two disagree, §B rules.

---

## 1. The stored value

`items.landing` and `archived_items.landing` are nullable `TEXT`. The value, when present, is a
JSON object:

```json
{
  "ref": "origin/develop",
  "ref_head": "e50964ad3f...",
  "observed_at": "2026-09-03T10:14:22Z",
  "sha": "c9f712232a...",
  "sha_source": "operator",
  "spec_status": "completed"
}
```

Field rules:

| Key | Presence | Rule |
|---|---|---|
| `ref` | always | the ref the observation was made against, as resolved by half A's `LandedRefFor` |
| `ref_head` | always | that ref's head SHA at `observed_at`. **A ref position, never a delivery claim** (REQ-TLE-013) |
| `observed_at` | always | RFC 3339 UTC, the same instant shape the queue already uses for `added_at` |
| `sha` | only when the operator supplied one | the delivering commit, on the operator's authority |
| `sha_source` | present iff `sha` is present | the constant `operator`. The only value this SPEC defines; the machine has no other lawful source (REQ-TLE-012) |
| `spec_status` | present when the card carries a `spec_id` | the frontmatter `status` read at record time, or the unknown marker |

The unknown marker is an explicit value (`"unknown"`), not an omitted key — REQ-TLE-010 distinguishes
*read and found nothing* from *never asked*, and an omitted key cannot carry that distinction.

**Absence is SQL `NULL`**, never `{}` and never `""` (REQ-TLE-006). A card that has had evidence
cleared returns to `NULL`, indistinguishable from one that never had any — deliberately: the queue
records what the operator asserts now, not an audit of what they once asserted.

## 2. Where the ALTER runs

At engine open, inside the same function that executes `backlogDDL`
(`internal/kanban/backlog_sqlite.go:278-299`), **after** the DDL and **before** the `schema_version`
switch. Ordering matters in both directions:

- After the DDL, because on a brand-new database the tables do not exist until the DDL has run.
- Before the version switch, because the switch is the point at which the engine declares the
  database usable; a column added after it would be absent for any caller that short-circuits there.

Idempotence is decided by reading the table's own column metadata
(`SELECT name FROM pragma_table_info('items')`) and issuing the `ALTER` only when the column is
missing. Catching SQLite's `duplicate column name` error instead is rejected: it makes a normal
control path an error path, and it swallows a genuine failure that happens to share the message.

The DDL const itself is **not** edited to include `landing`. Two reasons: `CREATE TABLE IF NOT
EXISTS` does not add columns to an existing table, so the DDL alone cannot serve a database in the
field; and keeping the column in exactly one place (the ALTER) means a new and an upgraded database
converge on the same statement rather than on two that must be kept in step.

## 3. The verb

```
moai todo landed <id> [--sha <sha>] [--ref <ref>]
moai todo landed <id> --clear
```

- `<id>` accepts the bare `<n>` form, normalized to `t<n>`, matching `done` / `undrop`
  (`REQ-TODO-004`).
- `--ref` overrides the resolved ref; absent, the ref resolves exactly as `todo pr` resolves it, so
  the two surfaces agree by construction rather than by convention.
- `--sha` is the only path by which a delivering SHA enters the store, and it is **validated before
  it is stored** (REQ-TLE-020):

  ```
  git rev-parse --verify <sha>^{commit}          # existence; yields the full SHA
  git merge-base --is-ancestor <resolved> <ref>  # reachability from the record's ref
  ```

  The **resolved full SHA** is stored, never the supplied abbreviated form, so a stored value cannot
  become ambiguous as the repository grows. Either check failing is exit 1 with nothing written and
  a stderr line naming which one failed. Neither command takes the card id as an input — that is the
  structural reason this is a referential-integrity check and not the attribution REQ-1.10 forbids
  (`spec.md` §B.3.1). A check that cannot be run (no git, unresolvable ref) is also exit 1: the
  record needs `ref_head` from git anyway (REQ-TLE-005), so an unanswerable git is already fatal to
  forming a record, and no new permissive-vs-refusing policy is introduced.
- `--clear` is mutually exclusive with `--sha` / `--ref`; combining them is a usage error (exit 2).
- Exit codes follow `internal/cli` convention: 0 written; 1 card not found, ref unresolvable, or a
  `--sha` validation failure (existence or reachability); 2 usage.
- The write is one `Mutate` under the queue lock, touching one row's one column.

`--clear` exists because §B.5 requires a correction path; without it a mistyped `--sha` is permanent
short of hand-editing the database, which is precisely the "stored observations cost more to
correct" hazard half A used to justify splitting this axis out.

## 4. The render

Row, seven tab-separated fields:

```
<card id>\t<outcome>\t<pull requests>\t<confidence>\t<queue state>\t<evidence>\t<card text>
```

The evidence cell is empty when the column is NULL. When present it is compact and machine-keyable:

```
landed@origin/develop:c9f7122(operator)     # operator-asserted delivering commit
landed@origin/develop:e50964a(ref-head)     # observed ref position only
```

The parenthesized token is the marker AC-TLE-016 keys on. It is present regardless of the SHA value,
which is why substituting one card's SHA for the other's leaves the cells distinguishable.

`--json` gains the record under a `landing` key on the render-time object. The resolver's own
`PRLinkOutcome` is **not** widened: it is the type REQ-TLE-011 keeps free of stored data, and adding
a stored-evidence field to it would put a delivering SHA inside the resolver's own output shape —
the exact adjacency REQ-1.10 exists to prevent.

## 5. What deliberately does not change

- `GitLandedQuerier`, `ResolveCardPRLink`, `LandingAnswer`, `LandedRefFor` — half A's surfaces,
  consumed unchanged.
- `items.state`, its CHECK, and `schema_version`.
- The `gh` query budget on `todo pr`.
- The five-field `BacklogItem`'s existing fields. `BacklogItem` gains one optional field carrying the
  record; the five that exist keep their names, types, and JSON tags, which is what makes the
  addition additive in `REQ-TODO-013`'s sense.
