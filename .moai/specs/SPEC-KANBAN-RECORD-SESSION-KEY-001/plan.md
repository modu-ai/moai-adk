# SPEC-KANBAN-RECORD-SESSION-KEY-001 — Implementation Plan

## §A Context

Producer moved: `internal/cli` (the launcher) hands the write to `internal/hook` (the session's own
SessionStart). Schema widened: `internal/kanban`. Transport for the facts the session cannot see:
`internal/config/envkeys.go`.

**No consumer file is edited — but a consumer exists, and this plan is shaped around it.** Version
0.1.0 of this plan asserted there was none; that was wrong. Re-measured:

```
$ grep -rn 'internal/kanban"' internal --include='*.go' | grep -v _test
internal/web/viewmodel_ops.go:23     internal/cli/kanban.go:25       internal/cli/cc.go:11
internal/cli/glm.go:20               internal/cli/factory.go:37      internal/cli/graph.go:14
internal/cli/todo.go:36              internal/cli/todo_analysis.go:26 internal/cli/todo_relate.go:28
internal/cli/todo_drop.go:43         internal/cli/todo_edit_move.go:31
internal/hook/session_start_kanban.go:30  internal/hook/session_start_factory.go:35
```

Thirteen importers. Most touch the queue rather than the record, but **`internal/web` reads the
records themselves**: `loadKanbanRecords` (`viewmodel_ops.go:439-461`) reads every `*.json` under
`.moai/state/kanban/` and unmarshals each into `kanban.Record`, and `buildChain`
(`viewmodel_ops.go:227-233`) then collapses the set with `byRole[role] = r` — **last write wins per
role**, over `os.ReadDir` order. Two records carrying the same role would therefore render whichever
one sorts later, arbitrarily. That measurement is what forces the milestone shape below: the write
must not be duplicated at any shippable point, so the move and the removal are one milestone rather
than two (§D M2).

`internal/web` and `internal/statusline` remain out of scope as **files to edit** (spec.md §D) —
they are consumers this SPEC must not break, not surfaces it changes.

Tier **M**, justified in §B.

Milestones are ordered by **decision reversibility**. M1 fixes the record's schema — the shape every
reader binds to, and the one thing spec.md §A.6 says cannot be renamed later. M2 moves the writer in
one step: it relocates a producer, changes what the launch environment must carry, and removes the
launcher's write in the same change. Read both with care; there is no mechanical tail.

## §B Tier justification — M

| Signal | Measurement |
|---|---|
| Packages touched | 4 (`internal/kanban`, `internal/config`, `internal/hook`, `internal/cli`) |
| Files | 6-8 — enumerated in §C; an enumeration, not a measured diff |
| Schema change | additive only (`Record` gains two `omitempty` fields; spec.md §A.6) |
| Always-loaded doctrine touched | none |
| Published documentation touched | none |
| Template mirrors touched | none (spec.md §C.3) |
| Milestones | 2, each independently shippable without leaving a duplicate writer (§A) |

Files land in the 5-15 band, the schema grows without renaming, and nothing an agent reads at
session start changes. That is Tier M: three artifacts, threshold 0.80.

This SPEC exists because its parent exceeded a Tier L budget. Its own budget — 8 requirements and 9
criteria against Tier M ceilings of 16 and 16 — leaves room for a re-audit to add without hitting a
ceiling. If scope grows past that, the answer is to cut scope, not to raise the ceiling.

## §C File enumeration

| File | Milestone | Change |
|---|---|---|
| `internal/kanban/record.go` | M1 | two additive `omitempty` fields; the constructor and role setter carry them |
| `internal/kanban/record_test.go` | M1 | round-trip, omitempty, and pre-change-bytes compatibility |
| `internal/config/envkeys.go` | M2 | the launch-fact keys the session cannot otherwise observe |
| `internal/cli/kanban.go` | M2 | `recordKanbanSession` becomes the export step and drops its write, in one change |
| `internal/hook/session_start.go` (or a new kanban-record sibling in the same package) | M2 | the record write, keyed by `input.SessionID` |
| `internal/hook/session_start_test.go` (or the new sibling's test) | M2 | key correctness, role/lane/card derivation, fail-open |
| `internal/cli/kanban_test.go` | M2 | the launcher-writes-no-record assertion |

Seven entries; `internal/cli/cc.go` and `internal/cli/glm.go` join at M2 if the eight call sites'
signatures change, bringing it to nine at the upper bound. Six to eight is the working estimate.

## §D Milestones

Each milestone leaves the tree green and adds no half-state.

### M1 — Widen the record schema (highest reversibility cost)

Two additive fields on `Record`, both `omitempty`, no existing key renamed (spec.md §A.6 — the
`@MX:ANCHOR` at `record.go:45` states why).

- **Lane number** — a non-pointer `int`, zero meaning "not a lane". Factory lanes number from 1, so
  0 is unreachable by legitimate data and a pointer would buy a distinction nothing can use. This is
  decision G-6, inherited from the parent and restated in §F.
- **Card identifier** — a `string`, distinct from `SpecID`, empty when neither source yields a value.

**The lane number does not pass through `WithRole`.** `WithRole` (`record.go:116-130`) admits only
the known role set and drops everything else, and its doc comment says why: a consumer must never
have to defend against arbitrary launch-label text arriving from a label. Widening it to
pattern-match `lane-<n>` reopens exactly that. The lane number arrives as its own datum, set by its
own path; the role value stays `lane`, unchanged.

Ships alone: a schema widening with round-trip and compatibility tests, readable by every existing
consumer — `internal/web`'s `loadKanbanRecords` included — and by every record file already on disk.
Nothing writes the new fields yet, so the shipped behaviour is unchanged.

### M2 — Move the write into the session, carry the facts it cannot see, and remove the launcher's write

The record is written from the session's own SessionStart, keyed by `input.SessionID` — the first
point at which the described session's identifier exists at all. The hook already reads the launch
environment for its notices (`session_start_kanban.go:47-56`, `session_start_factory.go:48-54`), so
role, lane number, and the run id are already in reach.

**Three facts are not, and carrying them is part of this milestone rather than an assumption under
it** (spec.md §A.5, measured):

- **Backend** — exists today only as a literal argument at the eight `recordKanbanSession` call
  sites; `grep -rn 'BACKEND' internal/config/envkeys.go` returns nothing.
- **SPEC identifier** — `MOAI_KANBAN_SPEC` is set only for the kanban lead (`kanban.go:174-176`);
  neither `enterKanbanCompanionMode` nor `enterFactoryWorkerMode` sets it.
- **Card-identifier override** — no key exists.

The card identifier itself is derived inside the session from the basename of
`git rev-parse --show-toplevel`, with the override preferred when set and the field left empty when
neither yields a value. An override set to the empty string is treated as unset (acceptance.md §E).

**Fail-open is the constraint that shapes this milestone**, not a detail of it. `WriteBestEffort`
discards every failure by design; the session start must too. A record that cannot be written is a
session that starts normally with no record — the same degradation the launcher path has today
(spec.md §C.2).

**The launcher's write goes in this same change**, not a later one. `recordKanbanSession` stops
calling `kanban.WriteBestEffort` and becomes the step that exports the launch facts; the eight call
sites keep whatever role and backend they pass, now as exported values rather than as arguments to a
write. AC-KRS-003 is the pin.

**Why this is one milestone and not two.** Version 0.1.0 of this plan split the move from the
removal, reasoning that "no consumer reads either yet, so a duplicate is inert". That reasoning was
false, and its premise is measured in §A: `internal/web` reads every record file and collapses the
set by role with last-write-wins over directory order. A shippable state with both writers would
produce two records per launch carrying the same role — the launcher's under the parent's
identifier, the session's under its own — and the console would render whichever sorted later. That
would replace a consistently wrong row with an arbitrary one, which is harder to diagnose than the
defect this SPEC exists to remove. Splitting here buys a smaller diff and costs correctness at the
one point a reviewer might ship; the split is therefore rejected.

Ships alone: at the end of M2 there is exactly one writer, it is the session, and the record is
keyed by the session it describes.

## §E Dependency graph

```
M1 ──> M2
```

Strictly serial, and the edge is real rather than conventional: M2 writes the fields M1 adds, so
landing M2 first would write into a schema that does not carry them. There is no parallel branch —
this SPEC has one thread, which is part of why it is Tier M.

## §F Resolved decisions

Recorded as answers, not as open questions.

**D-6 (the split decision). The writer moves to the session; the launcher does not stay.** The
launcher cannot key the record correctly under any implementation, because the identifier it would
need belongs to a process that does not exist when it runs (spec.md §A.1). Moving the write to the
first actor that holds the child's identifier makes the key correct **by construction** rather than
by a resolution step that could be got wrong again.

*Rejected:* passing the child's identifier down from the launcher. There is nothing to pass — the
backend generates the session identifier after exec, and the launcher never sees it.

*Rejected:* keeping the launcher as the writer "for the fields it already knows" and letting the
session fill in the rest. That produces two writers of one record, with the launcher's write landing
under the wrong key first and the session's landing under the right one — the misattribution this
SPEC exists to remove, now permanent and with a second file per launch. It is the tempting shape
precisely because backend and SPEC identifier are the facts the launcher holds and the session does
not, which is why §G names it explicitly and why M2 carries those facts through the environment
instead.

**G-6 (inherited). The lane number is a non-pointer integer whose zero means "not a lane".** Lanes
number from 1, so the zero value is unreachable by legitimate data and carries the distinction a
pointer would otherwise be needed for. It is data beside the role, never a role value.

**G-1 (inherited). The card identifier is derived from the worktree basename, with an explicit
override preferred — and the derivation is constrained by a containment test.** The dispatch protocol
fixes a card's worktree directory at `.claude/worktrees/<card-id>`, so the value already exists
wherever a card-carrying session stands; measured in this tree, `git rev-parse --show-toplevel`
returns `…/.claude/worktrees/t207`.

The basename is taken **only where that root's parent directory is named `worktrees`**. Without that
test the derivation always yields something inside any checkout, so REQ-KRS-005's "rather than
guessed" clause would be unreachable and the field would carry a checkout name for every session
working outside a card worktree — a live case, not a hypothetical: session `e46fcfef…` stands in
`/Users/goos/moai/moai-adk-go`, basename `moai-adk-go`, which `SPEC-WEB-CONSOLE-015` REQ-WC15-044
would render as a card identifier. The alternative — renaming the field to "worktree basename" and
handing card resolution to the consumer — was rejected: it is the larger change, it moves work into
a SPEC that has already declared this surface out of its scope, and REQ-WC15-044 asks for a card
identifier, not for a directory name to resolve.

`MOAI_KANBAN_ID` is explicitly not the source — `envkeys.go:167-173` documents it as the per-run
identifier (`tk4ntu`), set once per run.

**The sidecar is not fixed here, and this is a scope decision rather than an oversight.**
`.moai/state/current-session-id.txt` carries the same single-slot last-writer-wins shape, and this
SPEC stops reading it — but every other consumer keeps reading it exactly as before. Card **t221**
owns that surface. Merging the two would put a fix for `moai session current` and handoff
attribution inside a SPEC whose acceptance criteria are all about `kanban.Record`, and would make
neither reviewable on its own terms.

**Every existing record file is left alone.** Stated as a property rather than a count — the
directory is live state that grows while this SPEC is open, so a count would be a different number
by the time anyone checked it (spec.md §C.4; the merge-time check is a before/after listing
comparison, acceptance.md §F). They cannot be repaired in any case: the parent identifier they are
named for does not encode which child they described, so any migration would guess. They are
gitignored runtime state, they age out with their sessions, and REQ-KRS-007 keeps them readable
meanwhile.

## §G Anti-patterns to avoid

- **Widening `WithRole` to pattern-match a lane label.** Its drop behaviour is the guard that keeps
  arbitrary launch-label text out of the role field; pattern-matching `lane-<n>` there reopens it.
  The lane number is its own datum (M1).
- **Keeping the launcher as the writer "just for the fields it does know."** Two writers of one
  record, one of them under the wrong key. The whole point of D-6 is that there is exactly one
  writer, and it is the session (§F).
- **Treating the sidecar as fixable here.** It is t221's surface. This SPEC stops reading it and
  fixes nothing about it (§F).
- **Reading the sidecar from the new write path.** It would reproduce the defect inside the fix.
  AC-KRS-002's two-identifier fixture is the pin — a write path that read the sidecar would produce
  `T.json` and fail, whatever that path is named. Its companion grep covers the key-resolution
  surface (`internal/kanban/` and `internal/cli/kanban.go`) and deliberately does **not** grep
  `internal/hook/session_start.go`, which writes the sidecar itself and must stay untouched.
- **Letting a record failure fail the session start.** Fail-open is a design guarantee with an
  `@MX:NOTE` explaining that the absent error return **is** the guarantee. A hook that returns an
  error on an unwritable state directory turns a silent degradation into a broken session
  (AC-KRS-008).
- **Migrating or deleting the existing record files.** They cannot be repaired without guessing
  (§F).
- **Deriving the backend from `ANTHROPIC_BASE_URL` instead of carrying it.** That variable is set by
  the GLM path but is also settable by anyone, so inferring from it is a guess dressed as a
  measurement. REQ-KRS-006 carries the fact explicitly.
- **Splitting M2 into "move the write" and "remove the launcher's write".** It creates a shippable
  state with two writers producing two same-role records per launch, which `internal/web`'s
  `buildChain` resolves by arbitrary directory order (§A, §D M2). Landing the removal first is worse
  still: it leaves no writer at all.
- **Taking the worktree basename unconditionally as the card identifier.** Outside a card worktree
  it yields a checkout name — measured live, `moai-adk-go` — which the console would render as a
  card. The parent-directory containment test is the guard (REQ-KRS-005, AC-KRS-005(c)).
- **Running the full Go suite locally.** Target the affected packages and read CI for the
  full-suite verdict.

## §H Open verification items

Recorded as items to confirm during run, **not** as blockers on this plan.

- **Whether the console renders misattributed rows today.** The on-disk join was measured to be
  broken (spec.md §A.2, §A.3); the rendered result was **not** observed — `moai web` was not
  started. The corrected join is this SPEC's deliverable either way, so nothing here waits on the
  answer; but the parent SPEC's view work should know whether it is fixing a wrong row or an absent
  one.
- **Whether the hook's environment reliably carries the launcher's exports in every launch shape.**
  Measured for the shapes the hook already reads (`session_start_kanban.go`,
  `session_start_factory.go` both read them today, which is strong evidence), but not exercised
  end-to-end for a `moai glm` factory lane. Confirm in M2 rather than assume.

## §I Cross-references

Sibling SPECs are cited by **REQ id**, never by section number: both were rewritten during this
card's plan phase and their sections moved, while their REQ ids did not.

- `SPEC-WEB-CONSOLE-015` **REQ-WC15-043** (the lane join) and **REQ-WC15-044** (the lane number and
  card identifier it delivers) — the consumer that depends on this SPEC landing. That SPEC's version
  0.1.0 asserted the join closes on today's data; it has since withdrawn the claim, and this SPEC's
  §A.2 is the measurement behind the withdrawal.
- `SPEC-SESSION-TELEMETRY-001` **REQ-ST-002** — the sibling requirement that keys its own record by
  the identifier the session runtime delivered, and which for the same reason declines to read the
  single-slot sidecar. This SPEC makes `kanban.Record` agree with it.
- `.moai/reports/t207/spec-split-design.md` Appendix 2 — the measurement that produced decision D-6
  and this SPEC.
- Card **t221** — the single-slot sidecar surface this SPEC deliberately does not touch.
