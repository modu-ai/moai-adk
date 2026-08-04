# Project Navigator — Regeneration Reference

> Level 3 reference for the Project Navigator living-documents layer. Covers
> the three-file schema, the regeneration procedure (skill-driven + script), the
> atomic-rename strategy, the `--brief` reorientation mode, and the cross-
> subcommand consultation points.

## §1. Artifact set (three files, no more)

The Navigator produces exactly three markdown files under
`.moai/project/navigator/`:

| File | Role | Regenerated when |
|------|------|------------------|
| `navigator.md` | Single entry point — current frontier, next task, link to full file | every sync + on-demand `/moai project` |
| `capability-map.md` | Feature inventory rows (spec-id, title, status, path, provenance) | every sync + on-demand |
| `progress-map.md` | Per-SPEC progress rollup (spec-id, phase, last-commit, frontier milestone) | every sync + on-demand |

A fourth runtime state file lives at `.moai/state/navigator/last-regen-commit.txt`
(the HEAD SHA at last regeneration). It is a staleness sentinel, not a document.
Warnings land at `.moai/logs/navigator-warnings.log`.

The directory is **generated, not scaffolded** — the template ships no fixed
content there (it depends on the user's SPEC registry).

## §2. Schema

### `navigator.md` (entry point)

```markdown
# Project Navigator

> Living project reorientation brief. Regenerated from the SPEC registry + git log.
> Last regeneration: commit `<sha>` at <iso8601>

## Current frontier
- **SPEC-X-001** (in-progress) — <title> — module `<path>` — [<sha-short> @ <iso8601>]

## Next task
Advance **SPEC-X-001** toward its next milestone. See `progress-map.md`.
```

### `capability-map.md` (feature inventory)

```markdown
# Capability Map

| spec-id | title | status | implementation-path | commit-sha | captured-at |
|---------|-------|--------|----------------------|------------|-------------|
| SPEC-X-001 | <title> | <status> | <module> | <40hex> | <iso8601> |
```

### `progress-map.md` (per-SPEC rollup)

```markdown
# Progress Map

| spec-id | status | phase | last-commit-sha | last-commit-at | frontier-milestone |
|---------|--------|-------|-----------------|----------------|--------------------|
| SPEC-X-001 | <status> | <phase> | <40hex> | <iso8601> | M3 |
```

Every row in `capability-map.md` and `progress-map.md` carries a 40-char
`commit-sha` and an ISO-8601 `captured-at`/`last-commit-at`, both drawn from
`git log` for the owning file's last commit. Rows are references (spec-id,
status, frontier, provenance) — never copies of SPEC body content.

## §3. Regeneration procedure

### Inputs (read-only — the Navigator consumes, never owns)

- `.moai/specs/SPEC-*/spec.md` — frontmatter (`id`, `title`, `status`, `phase`, `module`)
- `.moai/specs/SPEC-*/progress.md` — frontier-milestone derivation (best-effort)
- `git log` — provenance (commit-sha + committer date ISO-8601)

The Navigator does **not** read product/structure/tech.md (those feed
`--audit`, owned by a follow-up SPEC), does **not** read the `@MX` corpus
(owned by a tree-sitter follow-up SPEC), and does **not** read any LSEL
surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`).

### Mechanism

Regeneration is driven by `scripts/navigator-regen.sh` — a self-contained
bash script (no `jq`, no `moai` binary, only `git` + `awk` + `sed` + `grep`).
Invoke it from the project root:

```bash
bash "${CLAUDE_SKILL_DIR}/scripts/navigator-regen.sh"
```

The script resolves the project root via `CLAUDE_PROJECT_DIR` (falling back to
`$PWD`), discovers SPECs under `.moai/specs/SPEC-*/spec.md`, extracts the
required frontmatter fields, sources provenance from `git log`, and writes the
three files atomically. It always exits 0 (fail-open).

### When regeneration runs

1. **`/moai project`** (on-demand): the skill body invokes the script alongside
   `product.md` / `structure.md` / `tech.md` regeneration, so a single invocation
   leaves the full project-context surface current.
2. **`/moai sync`** (chained): the sync-phase workflow invokes the script
   **before** the sync commit lands, so the Navigator's staleness window never
   exceeds one sync cycle.

### Idempotence contract

Two regenerations on the same commit with no intervening SPEC change produce
byte-identical output. This holds because the provenance timestamp is sourced
from `git log --format=%cI` (the HEAD commit's committer date), NOT from
wall-clock. A no-op regeneration is a safe operation.

### Malformed-frontmatter tolerance

When a SPEC's frontmatter is unparseable (missing closing `---`, missing the
required `id` field), the script skips that SPEC row, appends a warning line to
`.moai/logs/navigator-warnings.log`, and continues regenerating the remaining
rows. Regeneration never aborts on a single malformed SPEC.

### Empty-project form

When the project has zero SPECs OR zero commits, the script still exits 0 and
emits a minimal `navigator.md` carrying the literal `no features tracked yet`
placeholder. `capability-map.md` and `progress-map.md` carry only their header
rows.

## §4. Atomic-rename write strategy

Each of the four outputs (three documents + the last-regen-commit sentinel) is
written `<file>.tmp` then `mv`'d into place. A concurrent reader therefore
observes either the previous version or the new version, never a partial
document. The `last-regen-commit.txt` is written LAST — its SHA is a sentinel
for "all three documents reflect this commit".

For deterministic concurrency testing, the script honors
`NAVIGATOR_PRE_RENAME_BARRIER` (a path). When set and a companion
`<path>.armed` sentinel exists, the script writes `"ready"` to the barrier path
on the first atomic_write and blocks until the barrier is removed — letting a
test driver read the target mid-flight without millisecond polling.

## §5. `--brief` reorientation mode

### On-demand full brief (`/moai project --brief`)

Loads the full `navigator.md` entry brief PLUS the current-frontier section of
`progress-map.md` into the active context as a structured reorientation brief.
Use this for mid-session deep re-orientation.

### Ambient auto-brief (SessionStart hook)

`handle-session-start-navigator.sh` reads `navigator.md` at session start and
emits a bounded `additionalContext` (≤500 tokens: current frontier + next task
+ one link to the full file). Fail-open on missing files, unreadable files, or
a generation deadline exceed. See the SessionStart hook registration in
`.claude/settings.json`.

### Staleness signal

When the Navigator's `last-regen-commit.txt` is more than **3 sync cycles**
behind `HEAD` (hard-coded default; overridable via the
`navigator.staleness_cycles` config key), the SessionStart hint includes a
staleness advisory naming the gap.

## §6. Cross-subcommand consultation (Navigator as shared read primitive)

The Navigator is a project-level shared read primitive, not a siloed
`/moai project` feature. `/moai project` owns maintenance (write/regenerate);
other subcommands consume (read) at defined orientation phases. The consultation
points are opt-in per subcommand — they fire only at the named phase.

| Subcommand | Phase | Reads | Gains |
|------------|-------|-------|-------|
| `/moai project` | maintenance | (writes) | owns regeneration + `--brief` + `--audit` |
| `/moai sync` | sync tail | (writes — regenerates) | staleness window ≤ 1 sync cycle |
| `/moai plan` | Phase 1 context-load | `navigator.md` brief | scopes new SPEC against current frontier; prevents duplicate SPECs |
| `/moai run` | start-of-run | brief + owning SPEC's `progress-map.md` row | implementing agent oriented before first action — no re-derivation |
| SessionStart hook | every session | ambient auto-brief (≤500 tokens) | zero-touch re-orientation at session start |

### `/moai plan` Phase 1 consultation

Before drafting a new SPEC, the plan workflow consults `.moai/project/navigator/
navigator.md` (if present) and `capability-map.md`. If the candidate feature
overlaps an existing capability already tracked in `capability-map.md`, prefer
amending the existing SPEC over creating a duplicate. This draws the new SPEC's
boundary against the real project state.

### `/moai run` start-of-run consultation

At run-phase entry, the implementing agent reads the Navigator brief (current
frontier + the owning SPEC's `progress-map.md` row) before its first
implementation action, so it is oriented to what is already done and what is
next without re-deriving project state from many files.

## §7. Boundary vs LSEL

The Navigator answers **"what does the project look like now"** (present-tense,
project-scoped). LSEL answers **"how should the harness itself evolve"**
(past→future, harness-scoped). The two systems do not duplicate responsibility:

- Navigator read-set: `.moai/specs/`, `.moai/project/navigator/`, `git log`
- Navigator write-set: `.moai/project/navigator/`, `.moai/state/navigator/`, `.moai/logs/navigator-warnings.log`
- Navigator never reads/writes: `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`

Orientation ≠ harness self-evolution.
