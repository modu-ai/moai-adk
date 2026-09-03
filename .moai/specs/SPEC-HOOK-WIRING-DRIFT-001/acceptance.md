# SPEC-HOOK-WIRING-DRIFT-001 — Acceptance Criteria

Every criterion carries four things beyond its Given-When-Then:

- **`Pre-impl observed:`** — the value the criterion's command actually printed,
  before any implementation, tagged with the HEAD it was measured at:
  `@950cb4399` (authoring), `@4842760a7` (audit iteration 1), `@6331d505c` (audit
  iteration 2). **No criterion defers its baseline** (spec.md §C-5 is absolute;
  the three former deferrals were measured in iteration 1 and re-verified by the
  auditor in iteration 2).
- **`Mutant:`** — an implementation attempted that would pass the criterion while
  violating its requirement. Where one is constructible the criterion is shallow
  and was rewritten; where none is, the *attempted* mutants are named with why
  each fails. "None constructible" is never asserted without the attempts.
- **`Fresh mutant attempt (iter N):`** — a **new** mutant attacked against the
  **new** text, recorded per revision. A rewrite that was not re-attacked has not
  been shown to be deeper.

  > **Scope correction, iteration 2.** v0.2.0 applied this discipline only to
  > criteria it *rewrote*. AC-HWD-009 was therefore never re-attacked, and it was
  > the criterion the auditor defeated (N3). The discipline now reads: a
  > criterion carries a fresh attempt when it is rewritten **or** when a revision
  > changes something it depends on — and an unattacked criterion is a known gap,
  > not a passing one.
- **`Harness correction:`** — for every new check or gate, the input constructed
  to make it fail and the failure that must be **observed**. A gate seen only
  passing has not been shown to be a gate.

All commands run from the worktree root. `moai` means a binary built from this
tree (`make build` → `bin/moai`); the globally-installed `moai` is v3.1.2 and
stale relative to this tree, and must not be used for verification.

---

## §D AC Matrix

16 live criteria across 14 requirements. Tier M ceiling is 16 requirements AND
16 criteria, independently — both are at or under budget.

| AC | REQ | Milestone | Severity |
|---|---|---|---|
| ~~AC-HWD-001~~ | — | — | **retired iter 1** — folded into AC-HWD-003 (a) |
| ~~AC-HWD-002~~ | — | — | **retired iter 1** — folded into AC-HWD-003 (b) |
| AC-HWD-003 | REQ-HWD-001 | M1 | MUST |
| AC-HWD-004 | REQ-HWD-002 | M1 | MUST |
| AC-HWD-005 | REQ-HWD-003 | M2 | MUST |
| AC-HWD-006 | REQ-HWD-003 | M2 | MUST |
| AC-HWD-017 | REQ-HWD-003 | M2 | MUST |
| AC-HWD-007 | REQ-HWD-004 | M2 | MUST |
| AC-HWD-008 | REQ-HWD-005 | M2 | MUST |
| AC-HWD-009 | REQ-HWD-006 | M3 | MUST |
| AC-HWD-010 | REQ-HWD-007 | M3 | MUST |
| AC-HWD-018 | REQ-HWD-012 | M3 | MUST |
| AC-HWD-011 | REQ-HWD-008 | M3 | MUST |
| AC-HWD-012 | REQ-HWD-009 | M4 | MUST |
| AC-HWD-013 | REQ-HWD-010 | M4 | MUST |
| AC-HWD-014 | REQ-HWD-011 | M4 | MUST |
| AC-HWD-015 | REQ-HWD-013 | M3 | MUST |
| AC-HWD-016 | REQ-HWD-014 | M3 | MUST |

The two retired rows are kept so that audit-iteration-1 references still
resolve. Their content was not dropped — AC-HWD-003 absorbed both as named
sub-assertions, because a tuple-keyed set-diff strictly subsumes a
script-name grep and an `if`-value list. Retiring them removed redundancy, not
depth, and freed the two slots AC-HWD-017 and AC-HWD-018 now occupy.

Every criterion traces to a `REQ-XXX`. The two [HARD] constraint obligations
that previously traced only to §C-1 and §C-4 were promoted to REQ-HWD-013 and
REQ-HWD-014 in this iteration, so no criterion is requirement-orphaned.

---

## M1 — close the local drift

### AC-HWD-003 — full hook-entry parity, enforced by a test observed failing

**Given** a Go test that renders `templates/.claude/settings.json.tmpl` in memory
(`template.EmbeddedTemplates()` + `template.NewRenderer`, `HookOptIn.Enabled`
read from the project's `system.yaml`) and set-diffs its hook entries against the
parsed `.claude/settings.json`, keyed on
`(event, matcher, script, if, timeout, async)`,
**When** the test runs against this project,
**Then** all three hold:

- **(a)** the parsed `SubagentStop` group contains an entry whose script is
  `chain-event.sh`, with `timeout: 5`, `type: command`, and no `async` key;
- **(b)** the parsed `PostToolUse` entries naming `status-transition-ownership.sh`
  number exactly **3**, each with `async: true` and `timeout: 5`, and their `if`
  values are exactly `Write(**/.moai/specs/**)`, `Edit(**/.moai/specs/**)`,
  `MultiEdit(**/.moai/specs/**)`;
- **(c)** the template-only set and the project-only set are both **empty**, and
  the failure message on a non-empty set names each divergent script and its
  direction.

- `Pre-impl observed:` no such test exists — `grep -rl 'HookEntryParity'
  internal/` → rc=1, no match `@4842760a7`. All three sub-assertions fail today:
  (a) `grep -c 'chain-event.sh' .claude/settings.json` → **`0`** `@950cb4399`,
  re-confirmed `0` `@4842760a7`; (b) the JSON walk prints `entries 1 async 0
  if-scoped 0` and `[]` `@950cb4399`, re-confirmed byte-for-byte `@4842760a7`;
  (c) the measured set-diff is 4 template-only entries and 1 project-only entry
  (d1 §E1). The target state was independently confirmed in the template at
  `settings.json.tmpl:198` (a) and `:78/:86/:94` (b).
- `Mutant:` a **script-name-only** set comparison passes with the single unscoped
  synchronous entry in place, since the name is present either way — closed by
  keying on the full tuple. A **one-directional** comparison
  (template-minus-project) misses the degenerate project-only entry — closed by
  asserting both directions. Both were constructible against a weaker draft;
  neither passes the text above.
- `Harness correction:` [HARD] before PASS, run the test against a
  `settings.json` copy with **one** entry removed and record the observed failure
  verbatim, including the named script and direction. A parity test seen only
  green proves nothing.

### AC-HWD-004 — the inertness of the chain-event entry is stated, not glossed

**Given** the SPEC record and every code comment, doc line, and commit message
this SPEC adds referring to `chain-event.sh`,
**When** they are read,
**Then** each states that the entry produces no ledger event until the
node-population gap is closed, and **no** line in the change claims the entry
makes chain completion edges record.

- `Pre-impl observed:` `@4842760a7`, measured in this worktree —
  `ls .moai/state/chain/` → `ls: .moai/state/chain/: No such file or directory`
  (the directory does not exist here at all; the `.gitkeep`-only listing quoted
  in v0.1.0 was taken in the **primary checkout** and is not reproducible in this
  tree — corrected per audit D10). The load-bearing premise is independent of
  which tree: `grep -rn 'CreateNodeAtSpawn' internal | grep -v _test` → **1 hit**,
  `internal/chain/populate.go:53`, the definition, **0 callers**.
- `Mutant:` a change that adds the entry with a comment reading "restores the
  completion edge" passes AC-HWD-003 entirely while asserting something false.
  Constructible — and it is precisely the mutant this criterion exists to catch,
  which is why it is a criterion rather than a note.
- `Harness correction:` n/a — a review criterion over prose, verified by reading
  the diff.

---

## M2 — make the drift detectable

> **[HARD] Fixture construction — read before building any M2 failing input**
> (audit D3). M2 is scheduled **before** M1, so the project's live
> `.claude/settings.json` is **drifting** at M2 time and a drift-free copy does
> not exist in the tree. Every M2 failing input is therefore built **from the
> rendered template**, never by copying the project:
>
> 1. **drift-free fixture** — render `settings.json.tmpl` and write the result as
>    the fixture's `.claude/settings.json`. By construction its hook entries equal
>    the template's, so the diagnostic must report no drift.
> 2. **template-only fixture** (AC-HWD-005) — take the drift-free fixture and
>    **remove** the `chain-event.sh` `SubagentStop` entry. This is why the input
>    is expressed as "remove from the rendered fixture" and not "remove from the
>    project": at M2 time `grep -c 'chain-event.sh' .claude/settings.json` → `0`,
>    so there is nothing in the project to remove.
> 3. **project-only fixture** (AC-HWD-006) — take the drift-free fixture and
>    **add** an entry the template does not carry. The live project's unscoped
>    synchronous `status-transition-ownership.sh` entry is a real specimen to copy
>    the shape from, but the fixture is still built from the render.
> 4. **[HARD] Warm every fixture before snapshotting it** — run `bin/moai doctor`
>    **once** against the fixture and discard that run's output, then take the
>    AC-HWD-007 pre-run snapshot. This step exists because `moai doctor` writes
>    `.moai/state/config-cache.json` on its **first** run in any project, as
>    pre-existing machinery unrelated to the hook-wiring check. Without warming,
>    AC-HWD-007's snapshot (iii) changes on the measured run no matter how the
>    check is implemented, and a correct implementation cannot pass. Measured
>    here: on a fresh fixture `find .moai/state -type f` → 0 files; after run 1 →
>    `.moai/state/config-cache.json`; run 2 leaves it byte- and mtime-identical
>    (`8334 1787583098` both times); a third run over a `.moai/state` +
>    `.moai/logs` listing showed no delta at all. Warming is therefore sufficient,
>    not merely a mitigation.

### AC-HWD-005 — the diagnostic reports a template-only entry (missing registration)

**Given** the template-only fixture built per the construction note above,
**When** `bin/moai doctor` runs against that fixture project,
**Then** its output contains a hook-wiring drift line naming `chain-event.sh` and
identifying it as present in the template and absent from the project.

- `Pre-impl observed:` `bin/moai doctor 2>&1 | grep -ci 'hook wiring'` → **`0`**
  `@4842760a7` (measured on the tree-built binary, `moai-adk v3.1.2`, after
  `make build`); `grep -c 'Hook Wiring' internal/cli/doctor.go` → **`0`**.
- `Mutant:` a check that reports drift whenever the two files are not
  byte-identical passes this while flagging every whitespace or key-order
  difference — closed by requiring the drift line to **name the affected script**,
  which only a parsed entry-level comparison produces. A second and more
  dangerous mutant — a **hardcoded expected-entry list** that never renders the
  template — passes this criterion and AC-HWD-006 together; it is closed by
  **AC-HWD-017**, not here. This pair binds the *output*; AC-HWD-017 binds the
  *mechanism*; REQ-HWD-003 requires both, so all three are needed.
- `Harness correction:` [HARD] the failing input is fixture (2) above; the
  printed drift line must be recorded verbatim.

### AC-HWD-006 — the diagnostic reports a project-only entry (extra registration)

**Given** the project-only fixture built per the construction note above,
**When** `bin/moai doctor` runs against it,
**Then** its output names that entry as project-only.

- `Pre-impl observed:` no diagnostic exists — `0` `@4842760a7`, as AC-HWD-005.
  The project-only *shape* is real and measured: the live
  `status-transition-ownership.sh` entry has no `if` and no `async`
  (AC-HWD-003 (b) pre-impl value).
- `Mutant:` a one-directional check (template-minus-project only) passes
  AC-HWD-005 and silently ignores every extra local registration — which is why
  this is a separate criterion rather than a clause of AC-HWD-005.
- `Harness correction:` [HARD] failing input is fixture (3); the project-only line
  recorded verbatim.

### AC-HWD-017 — the diagnostic actually renders the template (mechanism-bound)

**Given** the hook-wiring check implemented so that its template source is an
**injectable parameter** (an `fs.FS`, or template bytes, threaded through the same
seam the production check uses — see plan.md §F M2), and a fixture template whose
`settings.json.tmpl` carries a hook entry for a script name that appears nowhere
in the repository, `zz-fixture-only.sh`,
**When** the check function is invoked with that fixture template source against a
project whose `settings.json` does not carry that entry,
**Then** the returned diagnostic message names `zz-fixture-only.sh` as
template-only.

- `Pre-impl observed:` neither the check nor a shared keyed-entry helper exists —
  `grep -c 'Hook Wiring' internal/cli/doctor.go` → **`0`** and
  `grep -rl 'HookEntryParity' internal/` → rc=1, no match, re-confirmed
  `@6331d505c`. The fixture name is confirmed absent from **source**:
  `grep -rl 'zz-fixture-only' internal/ .claude/hooks/` → **rc=1, no match**
  `@6331d505c`.
  **Corrected scope** (audit N5): v0.2.0 recorded the unscoped
  `grep -rl 'zz-fixture-only' .` → rc=1, and **writing this criterion invalidated
  it** — the command now returns rc=0 and names `acceptance.md` and the audit
  report, because the string is in SPEC prose. The property that matters is that
  no *script* by that name exists, so the command is scoped to where a script
  could live. Same defect class as iteration 1's D10: a cited command whose output
  no longer reproduces is an unattributed baseline, even when the conclusion is
  sound.
- `Mutant:` this criterion **is** the closure of mutant M-2 (a hardcoded
  expected-entry list, which passes AC-HWD-005 and AC-HWD-006 while never
  rendering the template, and rots against the template into the very drift class
  this SPEC exists to close). A hardcoded implementation cannot name a script it
  has never seen.
- `Fresh mutant attempt (iter 1):` **two attempted, both fail.**
  (i) *Parse the embedded template once into a package-level `var` at init.* This
  is still a real render, but it ignores the injected fixture source, so the check
  never sees `zz-fixture-only.sh` → **fails**. (ii) *Implement the shared helper
  correctly but have the doctor check ignore it and use a hardcoded list.* This
  is M-2 resurfacing one level up, and it is why the Given requires the fixture to
  be injected **through the seam the production check uses** and the Then observes
  the **check's own returned message** rather than the helper's return value — a
  check that ignores its template parameter cannot name the fixture script →
  **fails**. Attempt (ii) forced a design constraint now recorded in plan.md §F
  M2: the check must accept its template source as a parameter.
- `Harness correction:` [HARD] the fixture template IS the failing input; record
  the returned message verbatim. Additionally record one negative run — the same
  check with the **unmodified** embedded template against the drift-free fixture,
  which must report no drift — so the criterion is not satisfied by a check that
  names every script unconditionally.

### AC-HWD-007 — the diagnostic changes nothing under the project root

**Given** a fixture **warmed** per step 4 of the construction note (one discarded
`bin/moai doctor` run), and then a recorded pre-run snapshot consisting of
(i) whole-worktree `git status --porcelain`, (ii) `sha256` and `mtime` of
`.claude/settings.json`, and (iii) a listing of name+size+mtime for every file
under `.moai/logs/` and `.moai/state/` — the two gitignored paths a diagnostic
would most plausibly write to, which `git status --porcelain` cannot see —
**excluding `.moai/state/config-cache.json`**,
**When** `bin/moai doctor` runs — both against the drift-free fixture and against
a drifting one —
**Then** all three snapshots are unchanged.

> **Why `config-cache.json` is excluded, and why the exclusion is narrow.**
> REQ-HWD-004 constrains **the diagnostic**; this criterion observes **the whole
> doctor run**, which also carries pre-existing machinery. `config-cache.json` is
> that machinery's first-run write and belongs to no part of this SPEC's change.
> The exclusion is one named file, not a directory — a check writing any *other*
> file under `.moai/state/` still fails, which is the mutant the clause exists to
> catch. Warming makes the exclusion belt-and-braces rather than load-bearing:
> after warming, that file is stable across further runs, so the criterion would
> hold even without naming it. Both are stated so the boundary is explicit.

- `Pre-impl observed:` `@6331d505c` — `sha256(.claude/settings.json)` =
  `57fc6d11506a4cfd198dc4de1ecea27baa23bea9087a862adaa90e5008a7324e` (unchanged
  across all three iterations). Fixture behaviour measured this iteration on a
  fresh fixture: `find .moai/state -type f | wc -l` → **0** before any run;
  after `bin/moai doctor` run 1 (`rc=0`) → `.moai/state/config-cache.json`; run 2
  leaves it identical (`stat` → `8334 1787583098` both times); run 3 over a
  combined `.moai/state` + `.moai/logs` listing → **no delta**. No diagnostic
  exists yet, so this criterion records the invariant the new check must preserve,
  and the warmed-fixture baseline it must be measured against.
- `Mutant:` a check that repairs the drift and then reports "no drift" passes
  AC-HWD-005 on its **first** run and fails it on the second — closed by the
  harness correction below, and by asserting `mtime` as well as bytes (a repair
  that rewrote identical content would still move the mtime).
- `Fresh mutant attempt (iter 1):` **three attempted; two fail, one is an
  acknowledged boundary.** (i) *Write a drift report to
  `.moai/logs/doctor-drift.log`* — this was audit mutant M-3, which defeated the
  v0.1.0 `.claude/`-only observable; it **fails** on snapshot (iii). (ii)
  *Write a cache under `.moai/state/`* — **fails** on snapshot (iii), except for
  the one named exclusion. (iii) *Write outside the project root, e.g.
  `~/.moai/cache/`* — **passes**, and is not observed by any criterion. That is a
  deliberate boundary, not an oversight: REQ-HWD-004 was narrowed in iteration 1
  to "under the project root" so the requirement and its observable agree (audit
  D7), and the out-of-tree case is recorded as residual risk in §D.6 rather than
  papered over. Iteration 2 ruled this narrowing legitimate scoping rather than a
  shrunk requirement, on the ground that it *added* coverage of the motivating
  mutant (i) that v0.1.0 let through.
- `Fresh mutant attempt (iter 2):` **the criterion was rewritten because it was
  not defeated by a mutant but by the tool itself, and the fix was re-attacked.**
  The audit demonstrated that v0.2.0's clause was **unsatisfiable**: `moai doctor`
  writes `.moai/state/config-cache.json` on first run, so snapshot (iii) changed
  against the freshly-built fixture regardless of implementation — a correct
  implementation could not pass, and the failure would have presented in the run
  phase as the diagnostic misbehaving. **Reproduced independently here** on a
  fresh fixture: 0 files under `.moai/state` before, `config-cache.json` after
  run 1 (`rc=0`). Note the irony the audit pinned: fresh-mutant attempt (ii)
  above designs the clause to catch "a cache under `.moai/state/`", and doctor's
  own baseline behaviour *is* that write. Three attacks on the **new** text:
  (a) *a check that writes `config-cache.json` itself to hide behind the
  exclusion* — the file is written by config machinery before the check runs and
  is byte-stable afterwards (measured: `8334 1787583098` identical across runs
  2 and 3), so a check mutating it changes size or mtime → **fails**;
  (b) *a check that writes any other file under `.moai/state/`* → **fails**, the
  exclusion is one named file, not the directory; (c) *a check that repairs the
  drift on run 1* → **fails** the twice-run harness correction, unchanged.
- `Harness correction:` [HARD] run the diagnostic **twice in a row** against the
  same drifting fixture and record that the second run reports the same drift as
  the first. A self-repairing check is detectable only by the second run.

### AC-HWD-008 — the diagnostic fails open

**Given** a fixture project whose `.claude/settings.json` is truncated to invalid
JSON, and a second fixture where the file is absent,
**When** `bin/moai doctor` runs against each,
**Then** the hook-wiring check reports a warn status naming the cause, and
`bin/moai doctor; echo $?` prints the same integer as the baseline below.

- `Pre-impl observed:` `@4842760a7`, measured — `make build` then
  `./bin/moai doctor > /dev/null 2>&1; echo $?` → **`EXIT=0`**. Baseline exit
  status is **0**. The check does not exist:
  `./bin/moai doctor 2>&1 | grep -ci 'hook wiring'` → `0`. (This value was
  deferred in v0.1.0 in violation of §C-5; audit D4. It is now measured.)
- `Mutant:` a check that swallows the error and reports OK passes any
  exit-status-only assertion while hiding a genuinely broken settings file —
  closed by requiring a **warn status naming the cause**, not merely a non-fatal
  outcome.
- `Harness correction:` [HARD] both failing inputs (corrupt, absent) constructed;
  the warn line and the exit status recorded verbatim for each.

---

## M3 — record the disposition of the 11

### AC-HWD-009 — every one of the 11 carries a disposition

**Given** `.claude/rules/moai/development/hook-independence.md` and its template
twin,
**When** each of the 11 script names is looked up,
**Then** each has **its own row** — a line naming **that script and no other** of
the 11 — and that row carries **exactly one** of the five disposition classes,
**and that class is the one expected for that script** per the mapping below.

| Script | Expected class |
|---|---|
| `chain-event.sh` | `reachable-via-template-settings` |
| `handle-agent-hook.sh` | `reachable-via-agent-frontmatter` |
| `handle-session-start-compact.sh` | `reachable-via-in-binary-registry` |
| `handle-elicitation.sh` | `dead-by-decision` |
| `handle-elicitation-result.sh` | `dead-by-decision` |
| `handle-notification.sh` | `dead-by-decision` |
| `handle-task-created.sh` | `dead-by-decision` |
| `handle-worktree-create.sh` | `dead-by-decision` |
| `handle-worktree-remove.sh` | `dead-by-decision` |
| `handle-session-start-navigator.sh` | `open-question` |
| `team-ac-verify.sh` | `open-question` |

The mapping is not new judgment — it is plan.md §F M3's disposition table, moved
into the criterion so the gate can decide it. Sole-name matching treats a name
that is a substring of another (`handle-elicitation` inside
`handle-elicitation-result`) as the same script, not a second one.

- `Pre-impl observed:` `@6331d505c`, re-measured name-by-name over the rule file
  as **matching-line counts with the lines named**, since `grep -c` counts lines
  and several of these names share one:
  `chain-event` **0**, `handle-agent-hook` **0**, `handle-session-start-compact`
  **0**, `handle-session-start-navigator` **0** (absent);
  `handle-elicitation` **1** (line 102), `handle-elicitation-result` **1** (102),
  `handle-task-created` **1** (102), `handle-notification` **1** (101),
  `handle-worktree-create` **1** (105), `handle-worktree-remove` **1** (105),
  `team-ac-verify` **7** (33, 69, 71, 79, 89, 90, 97). **Split: 7 present,
  4 absent.**
  Two corrections carried here. (a) v0.1.0 said "the three newly-classified
  names" and then listed four, and omitted `team-ac-verify.sh` from the present
  set — audit D8; corrected in v0.2.0 to **four** newly-classified names.
  (b) v0.2.0 recorded `handle-elicitation` as **2**, "the count includes
  `handle-elicitation-result` as a substring" — **that value does not reproduce
  and the explanation was a misdiagnosis** (audit N4). The measured value is
  **1**: `grep -c` counts matching *lines*, not occurrences, so two names on one
  line inflate nothing. The real effect runs the other way — three of the 11
  names share line 102 and two more share line 105, so per-name line counts
  **understate** how crowded the file is. That crowding is not a footnote: it is
  the file's existing authoring style, which is precisely why the blanket-line
  mutant below is a realistic authoring shortcut rather than a contrived attack.
- `Mutant:` **v0.2.0's Then clause was defeated** (audit N3, constructed and
  executed by the auditor): "each appears at least once, on a line carrying one of
  the five disposition classes" is satisfied by **a single line carrying all 11
  names and one class** — 11 distinct names present, one disposition-class line,
  criterion passes while assigning `dead-by-decision` to `chain-event.sh`
  (reachable), to `handle-agent-hook.sh` (reachable — §A.2's own correction), and
  to both open questions. The "per name" mitigation lived in this `Mutant:` prose
  and not in the Then clause — the same shape as iteration 1's D1 and D2, and it
  survived because AC-HWD-009 was not rewritten in iteration 1 and so was never
  re-attacked.
- `Fresh mutant attempt (iter 2):` **two constructed and executed against the new
  text; both fail it, both pass the old one.** Fixtures were built and both clause
  forms evaluated mechanically.
  **(A) blanket line** — all 11 names on one `dead-by-decision` row.
  OLD clause → `PASS`; NEW clause → `FAIL ('chain-event.sh: no sole-name row')`.
  **(B) eleven sole-name rows, every one classed `dead-by-decision`** — the mutant
  that matters, because it satisfies the audit's *literally prescribed* fix ("one
  row per name … carrying exactly one of the five classes"): measured against that
  prescribed wording it returns **`PASS`**. OLD clause → `PASS`; NEW clause →
  `FAIL ('chain-event.sh: wrong class, expected reachable-via-template-settings')`.
  **This is why the rewrite goes beyond the prescription**: per-name rows alone
  bind structure, not content, and a uniformly-misclassed table is exactly the
  error the investigation's whole disposition analysis exists to prevent. Binding
  the expected class per name is what closes it.
- `Harness correction:` [HARD] before PASS, evaluate the criterion against fixture
  (B) above — eleven sole-name rows, all classed `dead-by-decision` — and record
  the observed failure naming the first mis-classed script. A disposition
  criterion seen only against the correct table has not been shown to check the
  classes.

### AC-HWD-010 — the two audit corrections are recorded

**Given** the same rule surface,
**When** it is read,
**Then** it states (a) 33 hook entries across 20 events with the 34th
`"type": "command"` occurrence being `statusLine`, and (b) that
`handle-agent-hook.sh` is registered via agent frontmatter, not settings.

- `Pre-impl observed:` `@4842760a7`, measured —
  `grep -c 'statusLine' .claude/rules/moai/development/hook-independence.md` →
  **`0`**; `grep -c 'handle-agent-hook' <same>` → **`0`**. (The `statusLine` value
  was deferred in v0.1.0 in violation of §C-5; audit D4. It is now measured.) The
  independent basis for the claim, unchanged: `grep -c '"type": "command"'
  .claude/settings.json` → **34** while a JSON walk of `d['hooks']` counts
  **33** entries across **20** events.
- `Mutant:` writing "33 entries" without naming the `statusLine` cause leaves the
  next reader to re-derive 34 from the same grep and conclude the doc is wrong —
  closed by requiring the **cause**, not the number.
- `Harness correction:` n/a.

### AC-HWD-018 — the `team-ac-verify.sh` correction is consistent across every surface that states it

**Given** all **six** surfaces REQ-HWD-012 names — three files plus their three
template twins under `internal/template/templates/`:
`.claude/rules/moai/development/hook-independence.md`,
`.claude/rules/moai/core/agent-common-protocol.md` (**always-loaded**), and
`.claude/rules/moai/core/agent-common-protocol-reference.md` —
**When** each is read after M3,
**Then** none of the six describes `team-ac-verify.sh` as "dormant" in the
registered-but-gated sense, and each carries the corrected reading — that it is
**not registered in any settings surface**, so no configuration flag activates it
— on the same line as the script name.

> **[HARD] The edit is line-specific, not a blanket removal.**
> `hook-independence.md` carries **five** occurrences of "dormant" and only two
> are wrong. Correct `:87` (*"10s (TaskCompleted) — **dormant / not wired**"*) and
> the `:89-93` caveat block (*"Row (g) caveat — `team-ac-verify.sh` is dormant …
> forward-looking (activates only under harness `thorough` + team mode
> prerequisites)"*), which is the registered-but-gated reading the wiring
> contradicts. **Leave `:96`, `:106`, and `:107` alone** — they use "dormant"
> correctly, about *other* surfaces (the worktree lifecycle wrappers and the
> `moai hook spec-status` subcommand), and those readings are accurate. A
> `sed`-style sweep over the word would corrupt three correct statements to fix
> two wrong ones.

- `Pre-impl observed:` `@6331d505c` — `grep -c 'dormant'`: **1** in
  `agent-common-protocol.md`, **1** in `agent-common-protocol-reference.md`, **1**
  in each of their two template twins, and **5** in `hook-independence.md` (lines
  87, 89, 96, 106, 107) with **5** in its twin. The wrong lines, verbatim:
  `agent-common-protocol.md:38` — *"`team-ac-verify.sh` (TaskCompleted in team
  mode, dormant)"*; `agent-common-protocol-reference.md:291` — *"TaskCompleted in
  team mode (dormant — harness `thorough` + team prerequisites)"*;
  `hook-independence.md:87` and its `:89-93` caveat. All read as *registered but
  gated*, which the wiring contradicts. The correct survivors are `:96`, `:106`,
  `:107`.
- `Mutant:` deleting the word "dormant" everywhere without stating the correct
  reading passes a count-only check while leaving the reader with no description
  at all. Closed by requiring the **corrected phrase on the same line**, not the
  absence of the old one.
- `Fresh mutant attempt (iter 1):` **RETRACTED — the claim was false.** v0.2.0
  recorded: *"(i) Correct all four files but leave `hook-independence.md` on the
  old wording — fails AC-HWD-009 + AC-HWD-010, which bind the same correction in
  `hook-independence.md`."* **Neither criterion binds that wording** (audit N2):
  AC-HWD-009 binds disposition classes, AC-HWD-010 binds the 33-entry and
  agent-frontmatter statements, and neither touches "dormant". So the mutant —
  correct the four, add a disposition row to satisfy AC-009, leave lines 87-93
  untouched — passed all 16 criteria while violating REQ-HWD-012 on 2 of its 6
  named surfaces, one of them distributed to 16 languages. This is the same
  overclaim class as v0.1.0's "none constructible" (audit D5): an unverified
  closure recorded as a closure. Recording the retraction rather than quietly
  editing it, because the pattern is the finding.
- `Fresh mutant attempt (iter 2):` **two attempted against the six-surface text,
  both fail.** (i) *The N2 mutant above* — correct four surfaces, leave
  `hook-independence.md:87` and `:89-93` — now **fails**, because those two
  surfaces are inside the Given. (ii) *Correct all six by deleting every
  occurrence of "dormant" in `hook-independence.md`* — **fails** the Then's
  "carries the corrected reading" clause for the two target lines, and separately
  destroys three accurate statements at `:96`, `:106`, `:107`, which is why the
  [HARD] caution above is part of the criterion rather than a note beside it.
- `Harness correction:` n/a — documentation content. Note that this criterion is
  a **factual correction, not a decision**: whether team mode should fire the hook
  remains §G-3 / card t244 and is untouched.

### AC-HWD-011 — nothing was deleted, in git terms

**Given** the project and template hook directories,
**When** their **tracked** inventory and working-tree state are measured,
**Then** all three hold: `git ls-files .claude/hooks/moai/ | wc -l` is unchanged;
`git ls-files internal/template/templates/.claude/hooks/moai/ | wc -l` is
unchanged; `git status --porcelain` for both paths is **empty**; and no tracked
wrapper is 0 bytes (`find .claude/hooks/moai -name '*.sh' -size 0 | wc -l` → 0).

- `Pre-impl observed:` `@4842760a7`, measured — `git ls-files .claude/hooks/moai/
  | wc -l` → **43**; `git ls-files internal/template/templates/.claude/hooks/moai/
  | wc -l` → **47**; `git status --porcelain` for both paths → **0 lines**;
  zero-byte wrappers → **0**.
  **Correction to v0.1.0:** it recorded the template count as `50` from
  `ls … | wc -l`. That was wrong — `ls` is aliased to a long listing in this
  environment, so the count included the `total` line and the `.`/`..` entries.
  The tracked count is **47**. Moving to `git ls-files` fixed a latent miscount as
  well as the mutant.
- `Mutant:` **v0.1.0 asserted "none constructible". That claim was false**
  (audit D5, mutant M-4). The interpretive gap: "not deleted" ≠ "name still
  returned by `ls`". Two mutants passed the old text — (a) `git rm --cached
  <wrapper>.sh`, which removes it from the repository (the distributed act
  REQ-HWD-008 forbids) while the working-tree file and the `ls` name set are
  unchanged; (b) truncating a wrapper to 0 bytes, leaving name and count intact
  while the script is functionally gone. Both are closed above: (a) by the
  `git ls-files` count plus porcelain emptiness, (b) by the `-size 0` clause.
  Practical risk was always low — both require deliberate perversity — but the
  overclaimed note is what stops the next reader from strengthening the criterion,
  which is the real defect.
- `Fresh mutant attempt (iter 1):` **three attempted, all fail.** (i) *Replace a
  wrapper's body with `exit 0`* — non-zero size and still tracked, but
  `git status --porcelain` shows it modified → **fails**. (ii) *`git rm --cached`*
  — porcelain shows ` D` for the path → **fails**. (iii) *Add a new wrapper*
  (43→44) — the count clause → **fails**. The count and the porcelain clause are
  not redundant: a commit that deleted and re-added a file would leave porcelain
  clean at check time, and only the count catches it.
- `Harness correction:` n/a — a direct inventory comparison. The clauses were
  derived from the mutants above rather than from a constructed failing run.

### AC-HWD-015 — Template-First order was followed

**Given** the commit(s) delivering M3,
**When** the local mirror and the template source are diffed,
**Then** `diff` reports no difference for **every** file M3 touches —
`hook-independence.md`, `agent-common-protocol.md`, and
`agent-common-protocol-reference.md` — and `make build` was run between the
template edit and the mirror.

- `Pre-impl observed:` `@4842760a7` — `diff -q` local mirror vs template source
  for `hook-independence.md` → rc=0, **IDENTICAL**. The criterion asserts the
  property is preserved across three files now, not one; its falsifying case is an
  edit landing in only one side of any pair.
- `Mutant:` **three attempted, all fail.** (i) *Edit only the local mirror* —
  the diff is non-empty → fails. (ii) *Edit only the template, skip the mirror* —
  the same diff fails in the other direction. (iii) *Edit both but skip
  `make build`* — the diff passes, since it compares source to mirror and not
  either to the binary; this is closed by the second Then clause, and by
  AC-HWD-012's and AC-HWD-005's use of `bin/moai`, which is stale unless the build
  ran. No claim is made that the space is exhausted — these are the three
  attempted.
- `Harness correction:` [HARD] verify by checking that an intentionally
  template-only edit **fails** the mirror diff before the mirror is written.

### AC-HWD-016 — the template edits are neutrality-clean

**Given** every template-side file M3 touches,
**When** each is scanned for the forbidden content classes,
**Then** none contains a SPEC ID (`SPEC-[A-Z]`), a REQ token (`REQ-[A-Z]`), an
internal card number (`\bt[0-9]{2,3}\b`), an internal date, or a commit SHA; and
`go test ./internal/template/...` neutrality guards pass.

- `Pre-impl observed:` `@4842760a7`, measured on the unmodified template file —
  `grep -cE 'SPEC-[A-Z]|REQ-[A-Z]' internal/template/templates/.claude/rules/moai/development/hook-independence.md`
  → **`0`**; `grep -cE '\bt[0-9]{2,3}\b' <same>` → **`0`**. Baseline is zero on
  both classes, so any non-zero count after the edit is the edit's doing. (This
  value was deferred in v0.1.0 in violation of §C-5; audit D4. It is now
  measured.)
- `Mutant:` writing "open question — see card t244" in the template passes
  AC-HWD-009 (name and disposition class both present) while leaking internal
  state into 16-language distribution. Constructible — it is the specific leak
  §C-4 forbids, which is why the card numbers stay out of
  `internal/template/templates/**` and the template rows name the pending decision
  instead.
- `Harness correction:` [HARD] construct a copy of the template file with a card
  number inserted, run the neutrality guard, and record the observed failure. A
  neutrality guard seen only passing has not been shown to be a guard.

---

## M4 — stop the MX dead work

### AC-HWD-012 — `moai mx query` builds the index when it is unavailable

**Given** a project directory containing `.moai/state/` with no `mx-index.json`
and at least one `@MX:DEBT` tag in a source file,
**When** `bin/moai mx query --kind DEBT` runs there,
**Then** it exits `0`, returns the tag, and `.moai/state/mx-index.json` exists
afterwards.

- `Pre-impl observed:` measured on a constructed temp project — stderr
  `SidecarUnavailable: sidecar index does not exist — run 'moai mx scan' to build
  the index`, TUI `ERROR Sidecarunavailable: no sidecar index.`, **`EXIT=1`**, and
  no `mx-index.json` written. Measured `@950cb4399` with the installed v3.1.2
  binary; the code path is unchanged at `@4842760a7`
  (`internal/cli/mx_query.go:97-103`), and the re-run on the tree-built binary is
  a named gap in §D.6.
- `Mutant:` an implementation that catches the error and returns an **empty result
  set** with exit `0` passes an exit-code-only assertion while serving a wrong
  answer — closed by asserting the tag is returned AND the index file exists.
  A second — auto-building only on *absent* and not on *stale* — is closed by the
  harness inputs below.
- `Harness correction:` [HARD] two further failing inputs constructed and observed:
  (a) an index whose `scanned_at` is older than the freshness threshold, which must
  also trigger a rebuild; (b) a corrupt (non-JSON) `mx-index.json`, which must
  trigger a rebuild rather than an error.

### AC-HWD-013 — the SessionStart path carries no MX cold-start scan

**Given** the `internal/hook` package,
**When** the three assertions below are evaluated,
**Then** all three hold:

- **(a) package-scoped absence** — `grep -r 'runMXColdStartScan' internal/hook/ |
  wc -l` → `0`, and the same for `mxIndexNeedsRebuild` and `mxScanNeeded`. The
  scope is the **package**, not one file.
- **(b) arity** — `spawnDeferredAdvisoryScans` declares exactly **4** parameters
  (`projectDir`, `driftFn`, `driftTimeout`, `completed`); no scan-gating boolean
  is threaded through it.
- **(c) behaviour** — `TestSessionStartMXColdStartIntegration`
  (`internal/hook/session_start_mx_test.go:159`) is **inverted, not deleted**: in
  a `t.TempDir()` project carrying an `@MX` tag and no `mx-index.json`, after
  `Handle` returns, `.moai/state/mx-index.json` **does not exist**.

- `Pre-impl observed:` `@4842760a7`, all three fail today —
  **(a)** package-scoped: `runMXColdStartScan` → **8** occurrences
  (`session_start.go` 5, `session_start_mx_test.go` 3), `mxIndexNeedsRebuild` →
  **12** (`session_start.go` 4, `session_start_mx_test.go` 8), `mxScanNeeded` →
  **6** (all in `session_start.go`). *(v0.1.0 measured the single-file counts 5 /
  4 / 6, which is what made mutant M-1 possible.)*
  **(b)** `internal/hook/session_start.go:554-560` declares **5** parameters, the
  fifth being `mxScanNeeded bool`.
  **(c)** measured by running the existing test —
  `go test ./internal/hook/ -run TestSessionStartMXColdStartIntegration -v
  -count=1` → `INFO session start (deferred): MX index built via cold-start scan
  project_dir=/var/folders/…/001 tags=1` and `--- PASS (0.00s)`. The index **is**
  created today, so the inverted assertion fails pre-implementation by
  construction.
- `Mutant:` **v0.1.0's version of this criterion was three greps against one
  file, and it was defeated** (audit D1, mutant M-1): relocate
  `runMXColdStartScan` / `mxIndexNeedsRebuild` into a sibling file in the same
  package, rename them and the `mxScanNeeded` parameter, and all three
  single-file greps go to `0` while the scan still fires and
  `go test ./internal/hook/...` stays green. The stated mitigation lived in the
  `Mutant:` prose — which is not a gate — and the mutant renamed the very
  parameter that prose relied on. Closed by promoting the arity check into the
  Then clause as (b), scoping the greps to the package in (a), and adding the
  behavioural assertion (c).
- `Fresh mutant attempt (iter 1):` **two attempted, both fail.** (i) *Move the
  scan to a different package* (`internal/mx/coldstart`) and call it from
  `session_start.go` — (a) passes, since the old identifiers are gone from
  `internal/hook/`; but the caller still needs the gate boolean, so arity stays 5
  → **(b) fails**; and the index is still created → **(c) fails**. (ii) *Inline
  the scan body at the call site*, computing the gate inside the goroutine so no
  named function and no parameter remain — (a) and (b) both pass → **(c) fails**,
  because the index is still written. Clause (c) is the load-bearing one, and it
  is behavioural rather than lexical, which is exactly why it was added.
- `Harness correction:` the inverted (c) test IS the correction — it fails on the
  pre-implementation tree (measured above: index created, `PASS` on the
  un-inverted assertion) and passes only once the scan is gone. Additionally:
  the tests asserting the scan's presence
  (`session_start_mx_test.go:83,121,135,159`) must be **retired or inverted in the
  same change, never skipped**, and `go test ./internal/hook/...` must pass
  afterwards.

### AC-HWD-014 — the false goroutine-survival comment is gone

**Given** `internal/hook/session_start.go`,
**When** `grep -c 'durable side effects still land'` runs,
**Then** it prints `0`, and no remaining comment in the file asserts that a
goroutine spawned in the hook process continues after the process returns.

- `Pre-impl observed:` **`1`** `@950cb4399`, re-confirmed `@4842760a7` at
  `internal/hook/session_start.go:253`, inside the `case <-joinTimer.C:` branch,
  reading *"The goroutine continues to completion in the background (durable side
  effects still land)"*. It contradicts the accurate comment at lines 1529-1531,
  which states the opposite.
- `Mutant:` deleting the phrase while leaving an equivalent claim in different
  words passes the grep — closed by the second, review-verified clause covering
  any remaining assertion of the same idea.
- `Harness correction:` n/a — a documentation-content criterion over a source
  comment.

---

## §D.5 Definition of Done

- All 16 live criteria PASS, each with its command output recorded and its
  `Pre-impl observed:` value cited alongside.
- Every new gate (AC-HWD-003, 005, 006, 017, 007, 008, 012, 013, 016) carries an
  **observed failure** on a constructed failing input. A gate with no recorded
  failure observation blocks closure regardless of its passing run.
- `go test ./internal/cli/... -timeout 600s`, `go test ./internal/hook/...`,
  `go test ./internal/template/...` pass; `go vet ./...` clean;
  `golangci-lint run` clean on changed packages.
- `make build` run after the template edits; the mirror diff is empty for all
  three M3 files (AC-HWD-015).
- **No wrapper was added or removed, in git terms** — `git ls-files` counts hold
  at 43 / 47 and `git status --porcelain` is empty for both hook paths
  (AC-HWD-011). This is stated in git terms deliberately: a working-tree `ls`
  cannot support an added-or-removed claim.
- The three deferred decisions (§G-1/2/3), the unnumbered fourth defect (§G-4),
  and the deferred snapshot fix (§G-5) are present in the SPEC and were not acted
  on.
- Push, then read CI for the full-suite verdict. The local affected-package runs
  are an early signal, never the verdict.

## §D.6 Residual risk — what these criteria do not observe

1. **Writes outside the project root.** AC-HWD-007 observes the worktree only. A
   diagnostic writing to `~/.moai/` would pass. REQ-HWD-004 was narrowed to match
   the observable rather than left overclaiming; this is the residual gap that
   narrowing leaves, stated rather than hidden.
2. **AC-HWD-012's baseline was measured with the installed v3.1.2 binary**, not
   one built from this tree. The code path (`internal/cli/mx_query.go:97-103`) is
   unchanged at this HEAD, but the re-run on `bin/moai` was not performed.
3. **Most mutants were reasoned, not implemented — but not all.** Mutants in this
   file are generally constructed by reading the criteria against the code, with
   nothing compiled. Mutant (i) under AC-HWD-013's fresh attempt in particular
   reasons from Go package scoping and the **tests** asserting the scan's presence
   at `session_start_mx_test.go:83,121,135,159` (those four lines are `func Test…`
   declarations; the `runMXColdStartScan` call sites are 90/129/148 — corrected
   per audit N7), not from a build. **Three exceptions were executed rather than
   reasoned**: AC-HWD-013 (c)'s behavioural baseline (the integration test was
   run), AC-HWD-009's two iteration-2 mutants (fixtures built and both clause
   forms evaluated mechanically), and AC-HWD-007's unsatisfiability (a fixture
   project was built and `bin/moai doctor` run against it three times).
4. **AC-HWD-009's per-name integers are line counts, and they understate.**
   v0.2.0 claimed `handle-elicitation` counts **2** "because
   `handle-elicitation-result` contains it" — **both the value and the reasoning
   were wrong** (audit N4). Measured: **1**. `grep -c` counts matching *lines*,
   not occurrences, so two names sharing a line inflate nothing. The real
   distortion runs the opposite way: three of the 11 names share line 102 and two
   more share line 105, so per-name line counts understate the file's crowding.
   The present/absent split (7 / 4) is unaffected and was re-verified name-by-name
   with line numbers.
