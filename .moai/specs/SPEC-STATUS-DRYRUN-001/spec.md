---
id: SPEC-STATUS-DRYRUN-001
title: "Anchor SPEC status parse/write to YAML frontmatter and honor --dry-run on spec status --sync-git (issues #1693 + #1692)"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/spec;internal/cli"
lifecycle: spec-anchored
tags: "spec-status, frontmatter, dry-run, sync-git, drift, data-corruption, bugfix"
tier: M
issue_number: 1693
---

# SPEC-STATUS-DRYRUN-001

## §A. Overview

Two linked external defect reports (GitHub #1693, then its downstream #1692) form one
failure chain against `moai spec status`:

1. **Read side (#1693)** — `ParseStatus` (`internal/spec/status.go` ~L244) tries the
   body status-table format BEFORE the YAML frontmatter. The `statusTableEnPattern`
   (~L19, `\|\s*Status\s*\|\s*([^\||]+)\s*\|`) matches the substring `| Status | Notes |`
   inside a conventional version-history header `| Version | Date | Status | Notes |`,
   and also inside backticked code spans in prose. Result: a SPEC whose body merely
   CONTAINS that header shape reports `Notes` as its status. The drift engine
   (`internal/spec/drift.go:155`) consumes the misparse and records false
   `Drifted: true` entries in `.moai/state/drift-cache.json`.
2. **Write side (#1692)** — the operator, seeing the false DRIFT, runs
   `moai spec status --sync-git` to reconcile. The `syncGit` branch
   (`internal/cli/spec_status.go` L39-41) never passes or checks `dryRun`, so with
   `--yes` every eligible SPEC is written unconditionally — even though `--dry-run`
   is documented on the subcommand as "Preview change without writing".
3. **Corruption amplifier** — `updateStatusInContent` (`internal/spec/status.go` ~L89)
   has the same misdetection order (table before YAML), so the "status update" rewrites
   the table cell after a `Status` header: on a normal frontmattered SPEC this
   CLOBBERS the `Notes` header cell of the version-history table, or the backticked
   prose mentioning it, while leaving the frontmatter `status:` line untouched. The
   operator neither gets the requested update nor an intact document.

Reproduced verbatim (evidence: `.moai/reports/t513/repro.sh`,
`.moai/reports/t513/repro-output-buggy.txt`):

- `--list` reports `SPEC-DEMO-001  Notes` and `SPEC-PROSE-001  Notes` (real
  frontmatter: `completed` / `draft`).
- drift cache holds `"FrontmatterStatus": "Notes", "Drifted": true` twice.
- `spec status --sync-git --dry-run --yes` prints `Summary: updated 2` and WRITES:
  `| Version | Date | Status | Notes |` became
  `| Version | Date | Status | implemented |`; the backticked prose line lost its
  `Notes` token the same way. Two `M` lines appear in `git status`.

This SPEC fixes BOTH sides of the chain as one unit — fixing one side leaves the
other half of the chain live.

## §B. Requirements (GEARS)

### REQ-001 — Frontmatter-anchored status read

**When** a SPEC document carries a YAML frontmatter block containing a `status:` key,
`ParseStatus` (`internal/spec/status.go`) **shall** return the frontmatter value and
**shall not** consult the document body for a status.

### REQ-002 — Legacy fallback read

**While** a SPEC document has no frontmatter block, or its frontmatter block carries
no `status:` key, `ParseStatus` **shall** fall back to the legacy body formats
(body status table `| Status | <value> |` row, `- **Status**:` list item) so legacy
Format D/E SPECs continue to resolve.

### REQ-003 — Frontmatter-anchored status write

**When** `updateStatusInContent` updates the status of a SPEC document that carries a
frontmatter block containing `status:`, it **shall** update that frontmatter value
(and insert a `status:` key only into an existing frontmatter block that lacks one)
and **shall not** modify any body content — version-history table cells, backticked
prose, or any other body line.

### REQ-004 — Legacy fallback write

**While** a SPEC document has no frontmatter block, `updateStatusInContent` **shall**
retain the legacy table/list update paths so status reconciliation remains possible
for frontmatter-less documents.

### REQ-005 — `--dry-run` honored on `--sync-git`

**When** `moai spec status --sync-git` runs with `--dry-run`, the command
(`internal/cli/spec_status.go`) **shall** compute the full reconciliation, print
exactly what WOULD change (per-SPEC current → target lines and the summary), and
write nothing. The command **shall not** modify any file under `.moai/specs/` on
that path.

### REQ-006 — Write path survives

**When** `moai spec status --sync-git` runs WITHOUT `--dry-run` against a
legitimately eligible SPEC, the command **shall** still perform the frontmatter
status update. This SPEC removes a defect, not the reconciliation capability.

### REQ-007 — Loud skip on unparseable status

**When** `syncGitSpecStatuses` parses a SPEC whose current status is not a member of
`spec.ValidStatuses`, the command **shall** skip that SPEC without writing and
**shall** emit a warning to stderr naming the SPEC and the offending parsed value —
a future misparse must fail loud, never feed the writer.

### REQ-008 — Accurate help text

**When** the `spec status` help is rendered, it **shall** describe `--dry-run`
accurately, including its behavior on the `--sync-git` path.

## §C. Acceptance Criteria (summary — full matrix in acceptance.md)

| AC | Verifies | Method |
|----|----------|--------|
| AC-001 | REQ-001 | Go unit test: frontmatter status reads back through `ParseStatus` |
| AC-002 | REQ-001 (negative direction) | Go unit test: body `Status`/`Notes` table cells — header row AND backticked-prose variant — are NEVER returned as the SPEC status |
| AC-003 | REQ-001 (consumer) | `spec drift` shows real frontmatter values; no false DRIFT records |
| AC-004 | REQ-003 | Go unit test: write changes only the frontmatter line; body bytes identical |
| AC-005 | REQ-002 / REQ-004 | Go unit tests: legacy no-frontmatter SPEC still reads and updates via table path |
| AC-006 | REQ-005 | `--sync-git --dry-run --yes` leaves the working tree byte-identical (hashes + `git status`) |
| AC-007 | REQ-006 | Real run (no `--dry-run`) still updates the frontmatter on an eligible SPEC |
| AC-008 | REQ-007 | Invalid parsed status → skip + stderr warning, no write |
| AC-009 | End-to-end | Post-fix rerun of `.moai/reports/t513/repro.sh` flips all buggy verdicts |
| AC-010 | REQ-008 | help-text accuracy assertion (`go test ./internal/cli/...`) — SHOULD |

## §D. Constraints

- Status enum is UNCHANGED (8 canonical values). No new statuses, no renames.
- Kanban's own parser `internal/kanban/status_read.go` is already frontmatter-anchored
  and correct — it is NOT touched.
- Tests and the fix NEVER touch this repository's own `.moai/specs/`; all fixtures
  are built under `t.TempDir()` (`/tmp`).
- All tests use `t.TempDir()` fixtures; no OTEL env vars in tests; affected-package
  test scope is `./internal/spec/... ./internal/cli/...`.
- The fix MUST flip the checked-in repro evidence: a post-fix rerun of
  `.moai/reports/t513/repro.sh` is the end-to-end AC harness (AC-009).

## §E. Investigated, NOT reproduced (record only)

The reporter's secondary hypothesis — an em-dash U+2014 separator plus non-squash
merge breaking git-implied status inference — did NOT reproduce on develop. A
fixture with close subject `docs(SPEC-EMDASH-001): sync-phase artifacts — 3-phase
close` behind a `--no-ff` merge classifies `completed` correctly
(`closeInfixMatch` is a plain contains-check on the lowercased title;
`gitLogAllFullMessage` uses default traversal). Fixture evidence:
`.moai/reports/t513/repro.sh` (SPEC-EMDASH-001 fixture; buggy-run output line 13
shows `SPEC-EMDASH-001 completed / aligned`). Nothing is fixed on that axis.

## §F. Out of Scope

### Out of Scope — git-implied status inference (em-dash axis)

- No change to `closeInfixMatch`, `gitLogAllFullMessage`, or any git-log traversal
  logic; the reported em-dash/non-squash hypothesis was investigated and did not
  reproduce (§E).

### Out of Scope — status model

- The 8-value status enum, `ValidStatuses` contents, and status transition rules
  are unchanged.
- No migration of legacy Format D/E SPECs to frontmatter (fallback support keeps
  them working as-is).

### Out of Scope — adjacent parsers and surfaces

- `internal/kanban/status_read.go` is already correct and is not modified.
- No change to non-`sync-git` `spec status` subcommand semantics beyond the shared
  `ParseStatus` heal.

### Out of Scope — documentation beyond help text

- No docs-site or README rewrite; only the `spec status` help text accuracy
  (REQ-008) and, if touched, incidental comment accuracy.

## §H. Cross-References

- GitHub issues #1693 (read side), #1692 (write side).
- Reproduction: `.moai/reports/t513/repro.sh`, `.moai/reports/t513/repro-output-buggy.txt`.
- Consumers healed by the `ParseStatus` fix: `internal/spec/drift.go`,
  `internal/spec/archive.go`, `internal/cli/spec_status.go`,
  `internal/hook/spec_status.go`.
