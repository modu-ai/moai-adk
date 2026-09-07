# SPEC-STATUS-DRYRUN-001 — Acceptance Criteria

All fixtures live under `t.TempDir()` (`/tmp`). No test or fix touches this
repository's own `.moai/specs/`. Baseline evidence:
`.moai/reports/t513/repro-output-buggy.txt` (buggy run), script
`.moai/reports/t513/repro.sh` (re-runnable end-to-end harness).

## §D. AC Matrix

| AC | Requirement | Verdict method | Severity |
|----|-------------|----------------|----------|
| AC-001 | REQ-001 | `go test ./internal/spec/...` | MUST |
| AC-002 | REQ-001 (negative) | `go test ./internal/spec/...` | MUST |
| AC-003 | REQ-001 (consumer) | repro script step [2] | MUST |
| AC-004 | REQ-003 | `go test ./internal/spec/...` | MUST |
| AC-005 | REQ-002 / REQ-004 | `go test ./internal/spec/...` | MUST |
| AC-006 | REQ-005 | `go test ./internal/cli/...` + repro step [4]-[5] | MUST |
| AC-007 | REQ-006 | `go test ./internal/cli/...` | MUST |
| AC-008 | REQ-007 | `go test ./internal/cli/...` | MUST |
| AC-009 | End-to-end chain | repro script full run | MUST |
| AC-010 | REQ-008 | `go test ./internal/cli/...` (help text assertion) | SHOULD |

## §D.1 — AC-001: Frontmatter status reads back (direction a)

- **Given** a fixture SPEC with frontmatter `status: completed` and a body
  containing a `| Version | Date | Status | Notes |` history table
- **When** `ParseStatus` is called on the document
- **Then** it returns `completed`, and the body is never consulted

## §D.2 — AC-002: Body table cells are NEVER the status (direction b — REQUIRED)

- **Given** fixture SPECs of both defective shapes: (i) frontmatter `status: draft`
  with a version-history table header `| Version | Date | Status | Notes |`;
  (ii) frontmatter `status: draft` whose body merely mentions the header inside a
  backticked code span in prose
- **When** `ParseStatus` is called on each
- **Then** the returned value is `draft` in both cases — never `Notes`, never any
  body-table cell value
- **Guard against vacuous green**: this AC fails if the implementation reads ANY
  body content for a frontmattered document; a mutant that deletes the frontmatter
  branch must turn AC-002 RED (mutation-proven during M1)

## §D.3 — AC-003: Drift consumer healed

- **Given** the repro fixture (3 SPECs: completed / completed-via-git / draft)
- **When** `moai spec drift` runs
- **Then** the FrontmatterStatus column reads the real frontmatter values
  (`completed`, `completed`, `draft`); zero records in
  `.moai/state/drift-cache.json` carry `FrontmatterStatus: "Notes"`; every
  remaining `Drifted: true` record carries the SPEC's true frontmatter status
  against its git-implied status (true positives — SPEC-DEMO-001 and
  SPEC-PROSE-001 have no close commit in the fixture history, so their drift
  against git-implied `implemented` is correct-by-design); expected Summary:
  `2/3 SPECs have status drift`. The fixtures are NOT changed to force alignment.

## §D.4 — AC-004: Write touches only frontmatter

- **Given** a frontmattered fixture SPEC containing a version-history table and a
  backticked-prose line mentioning `| Version | Date | Status | Notes |`
- **When** `updateStatusInContent(content, "implemented")` runs
- **Then** the frontmatter `status:` line is updated to `implemented`; the history
  table rows, the header cell `Notes`, and the backticked prose are byte-identical
  to the input

## §D.5 — AC-005: Legacy fallback survives (read AND write)

- **Given** fixture SPECs with NO frontmatter block: one with a body status table
  row `| Status | draft |`, one with `- **Status**: draft`
- **When** `ParseStatus` / `updateStatusInContent` run on them
- **Then** the legacy value is read and the legacy in-body status is updated
  (REQ-002 / REQ-004 backward compat holds)

## §D.6 — AC-006: `--dry-run` on `--sync-git` writes nothing

- **Given** an eligible fixture project (a SPEC whose frontmatter status differs
  from git-implied) with file hashes recorded before
- **When** `moai spec status --sync-git --dry-run --yes` runs
- **Then** the command prints the would-change plan (per-SPEC current → target
  lines and a summary) AND every file hash is unchanged AND `git status` is empty
  — the working tree is byte-identical

## §D.7 — AC-007: The write path survives

- **Given** the same eligible fixture project
- **When** `moai spec status --sync-git --yes` runs WITHOUT `--dry-run`
- **Then** the eligible SPEC's frontmatter status IS updated to the git-implied
  value (the reconciliation capability is preserved, not deleted)

## §D.8 — AC-008: Invalid parsed status → loud skip

- **Given** a fixture project containing a SPEC whose parsed current status is not
  a member of `spec.ValidStatuses` (e.g. a legacy document that parses to a stray
  token)
- **When** `moai spec status --sync-git --yes` runs
- **Then** that SPEC is skipped with a stderr warning naming the SPEC and the
  offending value, and its file is not written

## §D.9 — AC-009: End-to-end repro verdict flip

- **Given** the fixed binary built from this branch
- **When** `.moai/reports/t513/repro.sh` runs (it rebuilds the /tmp/t513-repro
  fixture from scratch)
- **Then**: step [1] `--list` shows `completed` / `completed` / `draft` (no
  `Notes`); step [2] drift shows real frontmatter values with no false DRIFT;
  step [4] `--sync-git --dry-run --yes` reports the plan with no write; step [5]
  hashes identical to step [3] and `git status` empty; step [6] diff section is
  empty. Output is saved under `.moai/reports/t513/` as the post-fix evidence.

## §D.10 — AC-010: Help text accurate

- **Given** the `spec status` command help output
- **When** it is rendered
- **Then** `--dry-run` is documented as "Preview change without writing" AND its
  `--sync-git` behavior is consistent with that description (no contradictory text)

## Edge Cases

- Frontmatter block exists but carries no `status:` key → body fallback read; write
  inserts `status:` INTO the existing frontmatter block (never creates a second
  block, never appends a stray key to the body).
- Malformed frontmatter (unterminated `---`) → treated as no frontmatter; legacy
  fallback applies; must not panic.
- Status table inside a fenced code block in a legacy (no-frontmatter) document →
  the fallback reader MUST NOT treat fenced content as a live status table (the
  first-30-lines gating must not regress into all-lines scanning for the fallback).
- SPEC file unchanged when status already equals the target → no write, counted as
  skipped (already done).

## Quality Gate Criteria

- `go vet ./internal/spec/... ./internal/cli/...` clean
- `golangci-lint run ./internal/spec/... ./internal/cli/...` clean
- Affected-package tests: `go test ./internal/spec/... ./internal/cli/...` all PASS
  (full-suite verdict delegated to CI on the PR head)
- No test writes outside `t.TempDir()`

## Definition of Done

- All MUST ACs pass with verbatim command output captured under
  `.moai/reports/t513/`
- AC-002 mutation check performed (deleting the frontmatter branch turns it RED)
- Post-fix repro output committed as evidence alongside the fix
- File scope held to `internal/spec/status.go`(+tests) and
  `internal/cli/spec_status.go`(+tests), or deviations justified in plan.md §D
