# acceptance.md — SPEC-TODO-QUEUE-HOME-CANON-001

## D. AC Matrix

| AC | Requirement | Verdict basis |
|---|---|---|
| AC-001 | REQ-001 | grep + guard test |
| AC-002 | REQ-002 | diff + catalog hash check |
| AC-003 | REQ-003 | new guard tests (positive + mutation control) |
| AC-004 | REQ-004 | disposition ledger completeness |
| AC-005 | REQ-005 | code observation + existing temp-guard tests still pass |
| AC-006 | REQ-006 | ledger records for any newly found statement |

## AC Scenarios (Given-When-Then)

### AC-001 — Foreman skill names the canonical location

**Given** the template source tree at the run-phase HEAD,
**When** the foreman skill's queue-watch snippet is read
(`internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md`),
**Then** `grep -n 'state/todo' <file>` returns 0 hits, and the snippet's directory
expression names the home-scoped canonical location (`~/.moai/db/<project-key>/todo`
or an equivalent resolver-derived form), verified by the docs-match-resolution test
(AC-003a).

### AC-002 — Template-first parity holds

**Given** the skill edit landed,
**When** `diff -q .claude/skills/moai-kanban-foreman/SKILL.md internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` runs,
**Then** exit 0 (byte-identical), and `git diff` on the same commit shows
`internal/template/catalog.yaml` hash refresh (via
`go run ./internal/template/scripts/gen-catalog-hashes.go --all`) alongside the
skill edit.

### AC-003 — Machine-verifiable guards

**AC-003a** **Given** a git-repository fixture under `t.TempDir()`, **When** the
docs-match-resolution test runs, **Then** the foreman skill's watch path equals
`kanban.StateDirForRoot(<fixture root>)` and differs from the project-local
`<fixture>/.moai/state/todo`.
**AC-003b** **Given** the repo-scope seam test, **When** it scans `internal/` +
`pkg/` non-test Go files, **Then** it finds zero `backlog.json` queue-path
constructions outside `internal/kanban`.
**AC-003c** **Given** either test's input mutated (restore the old `d=.moai/state/todo`
line, or inject a foreign literal `filepath.Join(x, "backlog.json")` into a
non-kanban package), **When** the tests run, **Then** they FAIL — mutation control
executed once and captured in §E.2 evidence.

### AC-004 — Disposition ledger complete

**Given** the survey table (spec.md §1.2, B1-B9), **When** the run phase closes,
**Then** every row carries a final disposition (`converge` / `keep-read-only` /
`keep` with justification / `boundary-exception` naming the owning card), zero rows
are blank, and any reader discovered beyond B1-B9 has its own ledger row under
REQ-004/REQ-006 adjudication.

### AC-005 — Retained fallbacks unchanged

**Given** the run-phase diff, **When** `internal/kanban/state_dir.go` `StateDirForRoot`
is inspected, **Then** the temporary-origin branch (line ~29) and the
home-unresolvable branch (line ~33) are byte-unchanged, and the existing
SPEC-TODO-HOME-TEMP-GUARD-001 tests in `internal/kanban` still pass
(`go test ./internal/kanban/ -run TempOrigin` — output captured).

### AC-006 — New false statements adjudicated

**Given** `grep -rn "state/todo\|state/kanban" internal/template/templates/` at
run-phase close, **When** hits other than t704's owned file and legacy-naming prose
(e.g. "the former `.moai/state/kanban/` source") appear, **Then** each is either
corrected under REQ-001/REQ-002 parity rules and recorded in the ledger, or recorded
as justified legacy NAMING (not a canonical claim) with a one-line reason.

## Edge cases

- `MOAI` home override env set to an absolute path: `StateDirForRoot` honors the
  override (state_dir.go:37); the docs-match-resolution test asserts against the
  DEFAULT resolution only (no env in the fixture).
- Temporary-origin launch base: resolution stays project-local by design — the
  guard tests must NOT assert home resolution for a temp-dir fixture (they use a
  git-repo fixture under `t.TempDir()`? NO — a `t.TempDir()` fixture IS a
  temporary origin. The fixture must be a git-repo directory OUTSIDE
  `os.TempDir()`, or the test must set the home override to a fixture-local
  absolute path, mirroring how existing SPEC-TODO-HOME-TEMP-GUARD-001 tests
  control this). Implementation note: reuse the existing temp-guard tests'
  fixture pattern.
- Skill description line also names the watched path — correct it in the same edit
  (parity with the snippet).
- `backlog.json.lock` (`legacyBacklogLockFileName`, backlog_store.go:47): the seam
  test's literal scan must not confuse lock-file names with queue-path
  constructions; scope the pattern to the queue-path join shape.

## Quality gate criteria

- Tested: new guard tests green + mutation control red-once; affected packages pass.
- Readable/Unified: skill edit matches existing file style; no reformatting.
- Secured: n/a (no trust-boundary change; path statement only).
- Trackable: Conventional commits referencing card t658 and this SPEC id.

## Definition of Done

All six ACs PASS with captured evidence in `progress.md` §E.2; disposition ledger
complete; template/local parity and catalog hashes verified; zero resolution-
behavior change confirmed (A1-A6 return values untouched).
