# Acceptance: SPEC-TODO-LAND-AUTO-DONE-001

Every criterion below is machine-verifiable: a Go test against a fixture git repository
built with `internal/kanban/temp_origin.go`, or a CLI invocation whose stdout/stderr and
queue-record bytes are asserted. No criterion rests on visual inspection.

## §D AC Matrix

### AC-AD-001 — Trigger: scan evaluates only live queued/picked items (REQ-AD-001)

**Given** a fixture queue holding one `queued` card, one `picked` card, one `dropped`
card, and one archived card, all with landing-eligible commits on the fixture
`origin/develop`
**When** `moai todo auto-done` runs
**Then** the `queued` and `picked` cards are archived, the `dropped` card's record is
byte-identical before and after, and the archive contains exactly the two closed ids.

### AC-AD-002 — Evidence form 1: recorded delivering SHA (REQ-AD-004.1)

**Given** a `queued` card whose `Landing.SHA` was recorded via `moai todo landed <id>
--sha <c>` (provenance `operator`), where `<c>` is reachable from the fixture
`origin/develop`
**When** the scan runs
**Then** the card is archived, its close line is
`done <id> landing=landed source=auto-land ref=origin/develop form=sha-recorded`, and the
log row carries `form=sha-recorded` and the full resolved SHA.

### AC-AD-003 — Evidence form 2: attributed subject (REQ-AD-004.2)

**Given** a `queued` card with no recorded landing evidence and a commit on the fixture
`origin/develop` whose subject attributes the card through one of the six positional
shapes (e.g. `Merge WT-fixture into develop (card t901)`)
**When** the scan runs
**Then** the card is archived with `form=subject-attribution` and the log row carries the
attributed subject line and that commit's SHA.

### AC-AD-004 — Guard M1: reissued-id collision fails closed (REQ-AD-006)

**Given** an id `t902` that has carried two DISTINCT texts — a predecessor (archived,
whose `fix(t902): ...` commit IS on the fixture `origin/develop`) and the current live
card with NO recorded SHA
**When** the scan runs
**Then** the current card is NOT archived, stdout carries
`skip t902 reason=ambiguous-id`, and the queue record shows the card still live.

### AC-AD-005 — Guard M1 override: recorded SHA disambiguates a reissued id (REQ-AD-006)

**Given** the AC-AD-004 queue, plus a `moai todo landed t902 --sha <c2>` record where
`<c2>` is the CURRENT card's own commit, reachable from `origin/develop`
**When** the scan runs
**Then** the card IS archived with `form=sha-recorded` — the recorded SHA is the one
evidence form that names this card rather than its predecessor.

### AC-AD-006 — Guard M2: run-landed, sync-pending does not close (REQ-AD-007)

**Given** a `picked` card whose `SpecID` points at a fixture SPEC directory whose
frontmatter `status:` reads `implemented` (sync not finished), with an attributing
subject on `origin/develop`
**When** the scan runs
**Then** the card is NOT archived, stdout carries
`skip <id> reason=spec-not-completed`, and a control card whose SPEC reads `completed`
with the same setup IS archived.

### AC-AD-007 — Guard M3: non-landing declaration attributes nothing (REQ-AD-008)

**Given** commits on the fixture `origin/develop` with subjects
`fix(t903): attempt (not merged)` and `docs(t904): notes (NOT landed)` and a live card
for each id with no other attributing subject
**When** the scan runs
**Then** neither card is archived (skip `reason=not-landed` — i.e. the plain three-valued
answer is not-landed), and no subject carrying the negation marker contributes
attribution to any id.

### AC-AD-008 — Inconclusive never closes (REQ-AD-005)

**Given** a card eligible on every other axis, with the landed ref unresolvable in the
fixture repository
**When** the scan runs
**Then** the card is NOT archived, stdout carries
`skip <id> reason=query-inconclusive` (the outcome is ALWAYS the skip line — pinned,
never an exit-code-only signal), the scan exits 0 (an inconclusive CARD is a skip
outcome, not a command failure; exit 1 is reserved for the scan itself being unable to
run, e.g. the queue store unreadable), and no close line for that id appears.

### AC-AD-009 — Close-line contract (REQ-AD-010, C4)

**Given** any successful close
**When** the scan's stdout is read
**Then** the close line starts with the canonical `done <id> landing=landed` prefix,
carries `source=auto-land`, one skip line exists per skipped card, the last line is a
summary, and `recordFactoryCardState` fired for the closed card (factory state
observable as for a manual `done`).

### AC-AD-010 — Execution log (REQ-AD-011)

**Given** any scan that closed at least one card
**When** the log file under `RuntimeStateDirForRoot` is read
**Then** it is valid JSONL, one row per closed card (plus skip rows), each row carrying
card id, RFC 3339 UTC instant, evidence form, landed ref and ref head at scan time, and
the subject+SHA (form 2) or recorded SHA (form 1).

### AC-AD-011 — Reversibility (REQ-AD-012)

**Given** a card the scan closed
**When** `moai todo undone <id>` runs
**Then** the card and its findings return to the queue at the held position, the log
gains a reversal row naming the original closure row, and a second `undone` is refused
(`no archived backlog item`).

### AC-AD-012 — Dry-run byte identity (REQ-AD-013)

**Given** a queue with a mix of closeable and skippable cards
**When** `moai todo auto-done --dry-run` runs
**Then** the queue record's bytes are identical before and after (whole-record
comparison, not a field sample), while stdout carries exactly the close and skip lines
the non-dry run produces.

### AC-AD-013 — Idempotence (NFR-4)

**Given** a scan that just closed its eligible cards
**When** the scan runs a second time with no intervening change
**Then** it closes nothing and reports zero closes.

### AC-AD-014 — No landing-column writes (REQ-AD-009)

**Given** any scan run, closing or skipping
**When** the queue record is diffed before/after
**Then** no item's `Landing` field changed, and `kanban.LandingSHASourceOperator`
remains the only provenance value in the package (grep-equivalent assertion on the
closed constant set).

### AC-AD-015 — Fetch boundary (REQ-AD-003) [delta D1]

**Given** a fixture repository whose remote observation is a counting script-git (a
stub `git` on PATH that increments a counter file on every `fetch` invocation and then
delegates to the real binary, or fails the fetch question per fixture)
**When** the scan runs WITH `--fetch`
**Then** the counter file shows exactly ONE fetch invocation for the scan, and the scan
evaluates against the post-fetch ref position
**When** the scan runs WITHOUT `--fetch` on the same fixture
**Then** the counter file shows ZERO fetch invocations and the evaluation uses the
local remote-tracking ref as it stood before the run.

### AC-AD-016 — Close-surface exclusivity (REQ-AD-002) [delta D2]

**Given** the full `moai todo` verb surface in the built binary
**When** a scope test walks every verb's code path to the queued/picked archive
transition (`BacklogRecord.ArchiveCard` and its callers)
**Then** exactly two verbs reach it: `done` (`newTodoDoneCmd`) and `auto-done`
(the scan); every other verb (`drop`, `undrop`, `landed`, `relate`, `edit`, `add`,
`undone`'s restore path, ...) reaches no archive transition — asserted structurally
(grep/call-graph) with the verb allowlist named in the test
**And** a control run executes `moai todo landed <id> --sha <c>` against a live queued
card with landing-eligible commits and proves the card's `State`, queue position, text,
and spec id are unchanged and NO archive entry is added — only the card's `landing`
field may change (recording evidence transitions no card; the invariant inherited from
`runTodoLanded`, which does write the `Landing` column via `store.Mutate` —
`internal/cli/todo_landed.go:127-140`).

### AC-AD-017 — Observed false-negative landing shapes attribute and close (REQ-AD-004) [delta 0.3.0, lead field fixtures]

Provenance: lead field observation 2026-09-13 — two real cards whose work HAD landed on
`origin/develop` but which `moai todo done --require-landed` rejected (false negatives);
both were closed by hand via `moai todo landed --sha`. This AC makes both shapes
first-class fixture cases, IN CONTRAST to the false-positive direction already guarded
by AC-AD-004/005 (reissued-id collisions, t654/t656/t657).

**Given** a fixture repository reproducing both shapes:

- **Shape A (t603, landing commit `d8b7836aa`)**: a live queued card with no recorded
  landing evidence, and a commit on the landed ref whose subject is
  `fix(hooks): sync-phase 게이트의 C++ 검사 복구 (t603, H08)` — the card id inside a
  trailing parenthetical followed by a comma and an extra token.
- **Shape B (t681)**: a live queued card carrying an operator-recorded `Landing.SHA`
  (recorded via `moai todo landed --sha`, validated reachable from the landed ref),
  where the merge commit `ae980ef2d` mentions the id in a NON-attributing subject
  position and the card's actual repair commit `4fb28a5c6`'s subject carries NO id —
  so no subject on the ref attributes the card.

**When** the scan runs
**Then** the Shape-A card IS attributed via its subject and archived with
`form=subject-attribution` — a comma-form trailing parenthetical MUST attribute to
`t603`; narrow subject matching must not reject a group for carrying a non-card token
after the comma
**And** the Shape-B card IS archived with `form=sha-recorded` — the recorded
delivering SHA is exactly the evidence form that survives subject-attribution misses
**And the contrast holds in the same fixture pair**: the reissued-id collision shapes
(AC-AD-004/005) still behave as specified — an id carried by MORE THAN ONE distinct
text still skips with `reason=ambiguous-id` on subject evidence alone and still
requires a recorded SHA to close. Attributing a legitimately-landed comma-form subject
must NOT loosen the collision gate: the two directions are judged by the same gates,
and only the reissue shape (multiple distinct texts for one id) trips `ambiguous-id`.

## §D.1 Edge cases

- Card whose recorded SHA is on a DIFFERENT ref than the resolved landed ref
  (reachability fails) → skip, never close (form-1 validation reuses
  `merge-base --is-ancestor` semantics).
- Card with `SpecID` set but SPEC directory unreadable → `spec_status=unknown` is NOT
  `completed` → skip `spec-not-completed` (unknown is not a pass).
- Card with NO `SpecID` (Class A/B shape) → the sync gate does not apply; landing
  evidence alone decides.
- Negation marker in a commit BODY with a clean attributing SUBJECT → attributes
  normally (the predicate is subject-stream-only; body negation is out of scope and
  recorded as a residual risk below).

## §D.2 Quality gates

- `go test ./internal/kanban/...` and `go test ./internal/cli/...` — all new tests pass;
  no existing test regresses.
- `golangci-lint run` clean on touched packages.
- `TestNew_NoAskUserQuestion`-style subagent-boundary guard extended to the new verb.
- LSP: zero errors at close (baseline captured at M1 start).

## §D.3 Definition of Done

- All AC-AD-001..017 GREEN with their fixture tests committed.
- REQ-AD-014 doc-step (folded back from the retired standalone AC, verified as a DoD
  file-content check): the lead post-push procedure step naming `moai todo auto-done`
  as the step run immediately after the remote-landing confirmation
  (`git fetch origin develop` + `git rev-parse origin/develop`) is present in BOTH
  `internal/template/templates/.claude/skills/moai/workflows/todo.md` AND its local
  mirror `.claude/skills/moai/workflows/todo.md`, in corresponding sections (mirror
  consistency), and neither file names a lane as the scan's runner (lanes never push).
- `moai todo auto-done --help` documents the two evidence forms, the CANONICAL
  FOUR-TOKEN skip-reason set (`ambiguous-id`, `spec-not-completed`, `not-landed`,
  `query-inconclusive` — the closed vocabulary enumerated in REQ-AD-010), the
  exit-code policy (0 for skip outcomes, 1 only when the scan itself cannot run), and
  the dry-run contract.
