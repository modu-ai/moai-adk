# acceptance.md — SPEC-PROJECT-NAVIGATOR-004

> Verification layer. Each AC is `Given-When-Then`, binary-testable. The GEARS obligation lives in `spec.md` (REQ-PN-004-001 .. 006); this file owns the Given-When-Then evidence.

## §D. AC Matrix

| AC | Requirement | Severity | Verification |
|----|-------------|----------|--------------|
| AC-001 | REQ-PN-004-001 | MUST | regression test (Go) |
| AC-002 | REQ-PN-004-002 | MUST | regression test (Go) |
| AC-003 | REQ-PN-004-003 | MUST | regression test (Go) — frontier list still contains implemented |
| AC-004 | REQ-PN-004-004 | MUST | `cmp` hook file vs origin/main HEAD (no diff) |
| AC-005 | REQ-PN-004-005 | MUST | `cmp` template vs mirror |
| AC-006 | REQ-PN-004-006 | MUST | test red→green evidence |
| AC-007 | §25 neutrality constraint | MUST | 5-item pre-commit self-check |

### §D.1 AC-001 — implemented SPEC excluded from "Next task"

**Given** a fixture project with three SPECs: `SPEC-A-001` (status `implemented`, alphabetically first), `SPEC-B-001` (status `in-progress`), `SPEC-C-001` (status `draft`):

**When** the regeneration script is run via `go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1`:

**Then** the `navigator.md` "Next task" line does NOT name `SPEC-A-001`, and the test passes.

### §D.2 AC-002 — in-progress preferred over draft

**Given** the same fixture as AC-001:

**When** the regeneration script is run:

**Then** the `navigator.md` "Next task" line names `SPEC-B-001` (the `in-progress` SPEC), NOT `SPEC-C-001` (the `draft` SPEC).

### §D.3 AC-003 — Current frontier still lists implemented SPEC

**Given** the same fixture as AC-001:

**When** the regeneration script is run:

**Then** the `navigator.md` "Current frontier" list contains a bullet for `SPEC-A-001` (status `implemented`), preserving the inclusive display semantics (REQ-PN-004-003).

### §D.4 AC-004 — SessionStart hook unchanged

**Given** the run-phase branch with the fix applied:

**When** `git diff origin/main..HEAD -- .claude/hooks/moai/handle-session-start-navigator.sh internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh` is run:

**Then** the diff is empty (the hook file is byte-identical to origin/main), confirming the fix propagates via the regenerated `navigator.md` alone.

### §D.5 AC-005 — template/mirror byte-parity

**Given** the run-phase branch with the script fix applied:

**When** `cmp internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh .claude/skills/moai-workflow-project/scripts/navigator-regen.sh` is run:

**Then** `cmp` exits 0 (byte-identical), AND `internal/template/rule_template_mirror_test.go` passes at CI.

### §D.6 AC-006 — regression test red before fix, green after

**Given** the regression test from M3 exists and the script is at pre-fix logic:

**When** `go test ./internal/template/ -run TestNavigatorRegen_NextTask -count=1` is run on the pre-fix HEAD:

**Then** the test FAILS (the pre-fix logic picks `SPEC-A-001`).

**Given** the script fix from M1 is applied:

**When** the same test is run on the post-fix HEAD:

**Then** the test PASSES.

The run-phase §E.2 evidence block cites both the red and the green outputs verbatim.

### §D.7 AC-007 — §25 neutrality pre-commit self-check clean

**Given** the script fix is staged:

**When** the contributor runs the 5-item pre-commit self-check from `.moai/docs/template-internal-isolation-doctrine.md` §25.3:

**Then** all 5 items pass (no SPEC IDs, REQ tokens, dates, commit SHAs, or moai-adk-internal paths leaked into the template surface), AND `internal/template/internal_content_leak_test.go` passes at CI.

## §D.8 Edge cases

- **EC-001 — No in-progress, no draft, only implemented/completed SPECs**: the "Next task" line falls through to the existing "No active SPEC. Consider opening a new SPEC via `/moai plan`." branch. The fix does not alter this fallback.
- **EC-002 — Multiple in-progress SPECs**: alphabetical sort (`sort -k1`) tiebreaks within the `in-progress` tier; the alphabetically-first `in-progress` SPEC wins.
- **EC-003 — SPEC with an unrecognized status** (e.g. a future enum addition): the positive predicate (`$3 == "in-progress"`, `$3 == "draft"`) excludes it from "Next task" by default — safe.
- **EC-004 — Empty SPEC registry**: unchanged — the `ROW_COUNT -eq 0` branch handles it before reaching "Next task".

## §D.9 Quality gate / Definition of Done

- All MUST ACs (AC-001 .. AC-007) PASS with attributed evidence.
- `go test ./internal/template/ -count=1` green.
- `golangci-lint run ./internal/template/` clean.
- `make build` clean.
- `cmp` template vs mirror exits 0.
- One atomic commit on the feature branch (template source + mirror + catalog regen + test).
- Frontmatter `status` transitions `draft → in-progress` at run-phase entry (owned by manager-develop), `in-progress → implemented → completed` later (owned by manager-docs). This plan-phase leaves `status: draft`.

## §D.10 Forward-looking checks (non-blocking)

- After merge, an end-to-end manual smoke (run `/moai project` against a registry with a known in-progress SPEC) confirms the on-disk `navigator.md` carries the expected "Next task" line. Not a MUST AC — the regression test is the binding verification.
- The autonomy-epic memory (`project_autonomy_workflow_epic.md`) is updated at sync-phase to mark this item closed.
