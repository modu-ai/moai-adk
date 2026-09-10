# SPEC-STATUS-DRYRUN-001 — Implementation Plan

Tier: M | Mode: TDD (`quality.yaml` `constitution.development_mode: tdd`)
Targets: GitHub #1693 (read) + #1692 (write) — one linked chain, one SPEC.

## §A. Context

- Diagnosis is CONFIRMED and reproduced; the lane owns the evidence under
  `.moai/reports/t513/`. Run-phase does NOT re-derive the diagnosis — it implements
  the accepted fix direction (R1-R4) below and proves it by flipping the repro
  script's verdicts.
- Development mode is TDD: each milestone lands RED tests first, then the fix.

## §B. Known Issues (from the confirmed diagnosis)

1. `parseStatusFromContent` (~L244, `internal/spec/status.go`) attempts the body
   table format before YAML frontmatter, and gates only the ATTEMPT on the
   first-30-lines sample while `parseStatusFromTable` (~L304) scans ALL lines.
2. `statusTableEnPattern` (~L19) matches `| Status | Notes |` inside
   `| Version | Date | Status | Notes |` headers and inside backticked code spans.
3. `internal/cli/spec_status.go` L39-41: `syncGit` branch returns
   `syncGitSpecStatuses(cmd, syncYes)` — `dryRun` is neither passed nor checked.
4. `updateStatusInContent` (~L89) has the same table-first order, so
   `updateStatusInTable` (~L181) clobbers the `Notes` header cell / backticked
   prose instead of the frontmatter (reproduced diff in spec.md §A).

## §C. Accepted Fix Direction (plan to this)

- **R1 (read)** — `ParseStatus` anchors to YAML frontmatter: frontmatter with
  `status:` key → return it, never consult the body. Legacy body formats
  (status table, `- **Status**:` list) remain ONLY as fallback when no
  frontmatter status exists.
- **R2 (write)** — `updateStatusInContent` prefers frontmatter the same way:
  update (or insert into) the existing frontmatter block whenever one exists;
  legacy table/list update paths only for documents with no frontmatter block.
- **R3 (CLI)** — `--sync-git` honors `--dry-run`: compute the reconciliation,
  print exactly what WOULD change, write nothing (reporter's preferred option 1;
  NOT the reject-the-combination option 2). Keep help text accurate.
- **R4 (guard)** — `syncGitSpecStatuses` skips any SPEC whose parsed current
  status is not in `spec.ValidStatuses`, with a loud stderr warning.
- Status enum unchanged; `internal/kanban/status_read.go` untouched.

Objection log: none. The direction is sound; R4 in particular converts any future
misparse from silent corruption into a loud skip.

## §D. Constraints

- Expected file scope: `internal/spec/status.go` (+ `status_test.go`) and
  `internal/cli/spec_status.go` (+ `spec_status_test.go`). ANY other file change
  requires explicit justification here before implementation.
- Never touch this repo's own `.moai/specs/` from tests or the fix; fixtures under
  `t.TempDir()`.
- No OTEL env vars in tests (`t.Setenv` prohibition). No `go test ./...` locally —
  affected packages only (`./internal/spec/... ./internal/cli/...`).

## §E. Self-Verification (run-phase obligations)

- E1: AC matrix PASS/FAIL with verbatim `go test` output for
  `./internal/spec/... ./internal/cli/...`.
- E2: post-fix rerun of `.moai/reports/t513/repro.sh` — capture output and diff it
  against `.moai/reports/t513/repro-output-buggy.txt`. MUST FLIP — every line
  carrying a `Notes` token as a parsed status (lines 12, 14, 19, 21, 32, 34, 44,
  46, 58, 59), plus the no-write evidence: line 61 summary becomes the dry-run
  form (`Summary: ... ` with no `updated 2` write count), lines 64/66 hashes are
  byte-identical to their step-[3] pre-values, lines 68-69 `M` lines are gone,
  lines 80-93 diff section is empty. MUST PERSIST (expected-unchanged, by design
  — a post-fix run that changes these is itself a defect): line 23
  (`Summary: 2/3 SPECs have status drift` — true positives, see acceptance.md
  §D.3) and line 45 (`"GitImpliedStatus": "implemented"` — the git inference is
  correct and untouched by this SPEC).
- E3: `golangci-lint run ./internal/spec/... ./internal/cli/...` clean.
- E4: confirm `git status` clean inside the repro fixture after the dry-run step.

## §F. Milestones (priority order — highest-change-likelihood decisions first)

### M1 (High) — Frontmatter-anchored read (R1) — the semantic core

RED: unit tests in `internal/spec/status_test.go` —
(a) frontmattered SPEC returns the frontmatter value (direction a);
(b) body `| Version | Date | Status | Notes |` header row and the backticked-prose
variant NEVER yield `Notes` (direction b — the vacuous-green guard);
(c) legacy no-frontmatter SPEC still resolves via table/list fallback (REQ-002).
Then fix `parseStatusFromContent` / `ParseStatus`: frontmatter-first, body formats
as fallback only. Keep the attempt/sample gating consistent with the fallback scan
(the 30-line sample must not silently hide a legacy fallback either).

### M2 (High) — Frontmatter-anchored write (R2)

RED: unit tests — `updateStatusInContent` on a frontmattered SPEC changes ONLY the
frontmatter `status:` line (body bytes byte-identical, including the history table
and backticked prose); legacy no-frontmatter SPEC still updates via the table path
(REQ-004); insertion into an existing frontmatter block lacking `status:` works.
Then reorder `updateStatusInContent` to frontmatter-first.

### M3 (High) — CLI: `--dry-run` on `--sync-git` (R3) + valid-status guard (R4)

RED: CLI tests with `t.TempDir()` projects —
(a) `--sync-git --dry-run --yes` writes NOTHING (byte hashes unchanged) while
printing the would-change plan;
(b) `--sync-git --yes` (no dry-run) still performs the frontmatter update on an
eligible SPEC (REQ-006 — the write path survives);
(c) a SPEC with an unparseable status is skipped with a stderr warning and no write
(REQ-007).
Then: thread `dryRun` into `syncGitSpecStatuses`, add the `ValidStatuses` skip,
update help text (REQ-008).

### M4 (Medium) — End-to-end proof + consumers sweep

Re-run `.moai/reports/t513/repro.sh` against the fixed binary; verify AC-009
verdicts (`--list` → completed/draft/completed; drift → real frontmatter values,
0 false DRIFT; dry-run leaves hashes and `git status` untouched). Confirm the
downstream consumers healed by M1 with no per-consumer change needed:
`internal/spec/drift.go`, `internal/spec/archive.go`,
`internal/cli/spec_status.go`, `internal/hook/spec_status.go`.

M1 → M2 → M3 → M4, in order (M2 and M3 are separable but M4 depends on all).

## §G. Anti-Patterns

- Do NOT delete or weaken the legacy fallback paths (REQ-002/REQ-004) — backward
  compat with Format D/E SPECs is a requirement, not incidental behavior.
- Do NOT "fix" the drift cache contents by hand or add cache-migration code — the
  cache regenerates from `ParseStatus`.
- Do NOT take the reporter's option 2 (reject `--dry-run` + `--sync-git`
  combination) — option 1 (honor dry-run) was accepted.
- Do NOT touch `internal/kanban/status_read.go` or the status enum.
- Do NOT add fixture SPECs anywhere under this repository's own `.moai/specs/`.

## §H. Cross-References

- spec.md §A (reproduced evidence), §B (requirements), §E (investigated-not-reproduced).
- acceptance.md — full AC matrix and Given-When-Then scenarios.
