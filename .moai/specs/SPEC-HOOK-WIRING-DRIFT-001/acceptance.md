# SPEC-HOOK-WIRING-DRIFT-001 — Acceptance Criteria

Every criterion carries three things beyond its Given-When-Then:

- **`Pre-impl observed:`** — the value the criterion's command actually printed on
  this tree, at HEAD `950cb4399`, worktree `.claude/worktrees/t216`, before any
  implementation. Measured, not asserted. A criterion whose command already
  passes today observes nothing and would be rejected.
- **`Mutant:`** — an implementation attempted that would pass the criterion while
  violating its requirement. Where one is constructible the criterion is shallow
  and was rewritten; where none is, that is stated with the reason.
- **`Harness correction:`** — for every new check or gate, the input constructed
  to make it fail and the failure that must be **observed**. A gate seen only
  passing has not been shown to be a gate.

All commands run from the worktree root. `moai` means a binary built from this
tree (`make build` → `bin/moai`); the globally-installed `moai` is stale relative
to it and must not be used for verification.

---

## §D AC Matrix

| AC | REQ | Milestone | Severity |
|---|---|---|---|
| AC-HWD-001 | REQ-HWD-001 | M1 | MUST |
| AC-HWD-002 | REQ-HWD-001 | M1 | MUST |
| AC-HWD-003 | REQ-HWD-001 | M1 | MUST |
| AC-HWD-004 | REQ-HWD-002 | M1 | MUST |
| AC-HWD-005 | REQ-HWD-003 | M2 | MUST |
| AC-HWD-006 | REQ-HWD-003 | M2 | MUST |
| AC-HWD-007 | REQ-HWD-004 | M2 | MUST |
| AC-HWD-008 | REQ-HWD-005 | M2 | MUST |
| AC-HWD-009 | REQ-HWD-006 | M3 | MUST |
| AC-HWD-010 | REQ-HWD-007 | M3 | MUST |
| AC-HWD-011 | REQ-HWD-008 | M3 | MUST |
| AC-HWD-012 | REQ-HWD-009 | M4 | MUST |
| AC-HWD-013 | REQ-HWD-010 | M4 | MUST |
| AC-HWD-014 | REQ-HWD-011 | M4 | MUST |
| AC-HWD-015 | §C-1 Template-First | M3 | MUST |
| AC-HWD-016 | §C-4 neutrality | M3 | MUST |

16 criteria across 11 requirements.

---

## M1 — close the local drift

### AC-HWD-001 — the `chain-event.sh` SubagentStop entry exists locally

**Given** `.claude/settings.json` in this project,
**When** `grep -c 'chain-event.sh' .claude/settings.json` runs,
**Then** it prints `1`.

- `Pre-impl observed:` `0`
- `Mutant:` a comment or an unrelated string containing `chain-event.sh` anywhere
  in the file passes this grep while wiring nothing. **Constructible — so this
  criterion is not sufficient alone** and is paired with AC-HWD-003, which
  compares the parsed entry as a keyed tuple under `SubagentStop`. AC-HWD-001 is
  retained as the cheap, readable smoke check, not as the proof.
- `Harness correction:` n/a — this is a measurement of an existing file, not a
  new gate.

### AC-HWD-002 — the `status-transition-ownership.sh` entries match the template's shape

**Given** `.claude/settings.json`,
**When** the parsed `PostToolUse` entries naming `status-transition-ownership.sh`
are counted by total, by `async: true`, and by presence of an `if` predicate,
**Then** the counts are `entries 3 async 3 if-scoped 3`, and the three `if`
values are exactly `Write(**/.moai/specs/**)`, `Edit(**/.moai/specs/**)`,
`MultiEdit(**/.moai/specs/**)`.

Command:

```bash
python3 -c "
import json
d=json.load(open('.claude/settings.json'))
n=a=s=0; ifs=[]
for g in d['hooks'].get('PostToolUse',[]):
    for h in g.get('hooks',[]):
        if 'status-transition-ownership.sh' in json.dumps(h):
            n+=1
            if h.get('async'): a+=1
            if h.get('if'): s+=1; ifs.append(h['if'])
print('entries',n,'async',a,'if-scoped',s); print(sorted(ifs))
"
```

- `Pre-impl observed:` `entries 1 async 0 if-scoped 0` / `[]`
- `Mutant:` three copies of the same `if` predicate would satisfy a count-only
  check while leaving two of the three tool paths unscoped. **Constructible —
  hence the criterion asserts the three `if` values, not only the count.**
- `Harness correction:` n/a — measurement of an existing file.

### AC-HWD-003 — full hook-entry parity, both directions, enforced by a test that has been observed failing

**Given** a Go test that renders `templates/.claude/settings.json.tmpl` in memory
(`template.EmbeddedTemplates()` + `template.NewRenderer`, `HookOptIn.Enabled`
matching the project's `system.yaml` value) and set-diffs its hook entries
against the parsed `.claude/settings.json`, keyed on
`(event, matcher, script, if, timeout, async)`,
**When** the test runs against this project,
**Then** both the template-only set and the project-only set are empty, and the
failure message on a non-empty set names each divergent script and its direction.

- `Pre-impl observed:` no such test exists — `grep -rc 'HookEntryParity'
  internal/template` → `0` (no matching files); and the parity condition itself
  is **false today**, established independently by AC-HWD-001 (`0`, expected `1`)
  and AC-HWD-002 (`entries 1 async 0 if-scoped 0`, expected `3/3/3`). The
  measured set-diff at HEAD is 4 template-only entries and 1 project-only entry
  (d1 §E1).
- `Mutant:` a test that compares only **script names** as a set would pass with
  the single unscoped synchronous `status-transition-ownership.sh` entry in place
  (the name is present either way), while three `if`-scoped async entries are
  missing. **Constructible — hence the key is the full tuple including `if`,
  `timeout`, and `async`, not the script name.** A second mutant — comparing only
  template-minus-project — would miss the degenerate project-only entry;
  **hence both directions are asserted.**
- `Harness correction:` [HARD] before this criterion may be marked PASS, the test
  MUST be run against a settings.json copy with **one** entry removed and the
  observed failure recorded verbatim, including the named script and direction. A
  parity test seen only green proves nothing.

### AC-HWD-004 — the inertness of the chain-event entry is stated, not glossed

**Given** the SPEC record and any code comment or doc line this SPEC adds
referring to `chain-event.sh`,
**When** they are read,
**Then** each states that the entry produces no ledger event until the
node-population gap is closed, and **no** line in the change claims the entry
makes chain completion edges record.

- `Pre-impl observed:` `.moai/state/chain/events.jsonl` does not exist in this
  worktree or the primary checkout (`ls .moai/state/chain/` → `.gitkeep` only,
  0 bytes); `grep -rn "CreateNodeAtSpawn" internal | grep -v _test` → one hit, the
  definition at `internal/chain/populate.go:53`, and no caller.
- `Mutant:` a change that adds the entry and a comment reading "restores the
  completion edge" passes AC-HWD-001 and AC-HWD-003 while asserting something
  false. **Constructible — that is precisely the mutant this criterion exists to
  catch**, which is why it is a criterion rather than a note.
- `Harness correction:` n/a — this is a review criterion over prose, verified by
  reading the diff.

---

## M2 — make the drift detectable

### AC-HWD-005 — the diagnostic reports a template-only entry (missing registration)

**Given** a temporary project copy whose `.claude/settings.json` has one hook
entry removed (the `chain-event.sh` SubagentStop entry),
**When** `moai doctor` runs against that copy,
**Then** its output contains a hook-wiring drift line naming `chain-event.sh` and
identifying it as present in the template and absent from the project.

- `Pre-impl observed:` `moai doctor 2>&1 | grep -ci 'hook wiring'` → `0`; and
  `grep -c 'Hook Wiring' internal/cli/doctor.go` → `0`. The only hook-related
  doctor output today is `Hooks Config  hook handlers directory found` and
  `Hook opt-in:  enabled`.
- `Mutant:` a check that reports drift whenever the two files are not
  byte-identical would pass this criterion while flagging every whitespace or
  key-order difference as drift, making it unusable. **Constructible — hence the
  criterion requires the drift line to name the affected script, which only a
  parsed entry-level comparison can produce.**
- `Harness correction:` [HARD] the failing input is the constructed copy above;
  the printed drift line must be recorded verbatim. A diagnostic that has only
  been seen printing "no drift" has not been shown to detect anything.

### AC-HWD-006 — the diagnostic reports a project-only entry (extra registration)

**Given** a temporary project copy whose `.claude/settings.json` carries one hook
entry the rendered template does not (e.g. the degenerate unscoped synchronous
`status-transition-ownership.sh` entry that exists at HEAD today),
**When** `moai doctor` runs against that copy,
**Then** its output names that entry as project-only.

- `Pre-impl observed:` no diagnostic exists (`0`, as AC-HWD-005). The
  project-only entry itself is measured and real at HEAD: the local
  `status-transition-ownership.sh` entry has no `if` and no `async`
  (AC-HWD-002's pre-impl value).
- `Mutant:` a one-directional check (template-minus-project only) passes
  AC-HWD-005 and silently ignores every extra local registration.
  **Constructible — hence this is a separate criterion rather than a clause of
  AC-HWD-005.**
- `Harness correction:` [HARD] failing input constructed as above; the
  project-only line recorded verbatim.

### AC-HWD-007 — the diagnostic changes nothing on disk

**Given** the project tree with a recorded `sha256` and `mtime` for
`.claude/settings.json` and a recursive checksum of `.claude/`,
**When** `moai doctor` runs (both on a drift-free copy and on a drifting copy),
**Then** the `sha256` and the `mtime` of `.claude/settings.json` are unchanged,
and the recursive checksum of `.claude/` is unchanged.

Command sketch:

```bash
shasum -a 256 .claude/settings.json > /tmp/before.sha
stat -f '%m' .claude/settings.json > /tmp/before.mtime
moai doctor > /dev/null 2>&1
shasum -a 256 .claude/settings.json | diff - /tmp/before.sha
stat -f '%m' .claude/settings.json | diff - /tmp/before.mtime
```

- `Pre-impl observed:` `sha256(.claude/settings.json)` =
  `57fc6d11506a4cfd198dc4de1ecea27baa23bea9087a862adaa90e5008a7324e`. No
  diagnostic exists yet, so this criterion measures the invariant the new check
  must preserve rather than a behaviour it currently violates.
- `Mutant:` a check that repairs the drift and then reports "no drift" passes
  AC-HWD-005 on its **first** run against a drifting copy and fails it on the
  second. **Constructible — and it is the exact failure §C-2 forbids**, which is
  why byte-identity AND mtime are both asserted: a repair that produced identical
  bytes (e.g. rewriting the file with the same content) would still move the
  mtime.
- `Harness correction:` [HARD] run the diagnostic twice in a row against the same
  drifting copy and record that the second run reports the same drift as the
  first. A self-repairing check is detectable only by the second run.

### AC-HWD-008 — the diagnostic fails open

**Given** a temporary project copy whose `.claude/settings.json` is truncated to
invalid JSON, and a second copy where the file is absent,
**When** `moai doctor` runs against each,
**Then** the hook-wiring check reports a warn status naming the cause, and the
`moai doctor` exit status equals the exit status of the same command on the
unmodified copy.

- `Pre-impl observed:` `moai doctor; echo $?` on this tree → the baseline exit
  status to be recorded at run time on the freshly built binary; the check does
  not exist (`0` occurrences), so no warn line is emitted today.
- `Mutant:` a check that swallows the error and reports OK passes any
  exit-status assertion while hiding a genuinely broken settings file.
  **Constructible — hence the criterion requires a warn status naming the cause,
  not merely a non-fatal outcome.**
- `Harness correction:` [HARD] both failing inputs (corrupt, absent) constructed;
  the warn line and the exit status recorded verbatim for each.

---

## M3 — record the disposition of the 11

### AC-HWD-009 — every one of the 11 carries a disposition

**Given** `.claude/rules/moai/development/hook-independence.md` and its template
twin,
**When** each of the 11 script names is grepped,
**Then** each appears at least once, on a line carrying one of the five
disposition classes.

Names: `chain-event.sh`, `handle-agent-hook.sh`, `handle-elicitation.sh`,
`handle-elicitation-result.sh`, `handle-notification.sh`,
`handle-session-start-compact.sh`, `handle-session-start-navigator.sh`,
`handle-task-created.sh`, `handle-worktree-create.sh`,
`handle-worktree-remove.sh`, `team-ac-verify.sh`.

- `Pre-impl observed:` `grep -c 'chain-event'` → `0`; `grep -c
  'handle-agent-hook'` → `0`. The four RETIRE-OBS-ONLY names and the two worktree
  names are already present in the existing dormant-surfaces block; the three
  newly-classified names (`chain-event.sh`, `handle-agent-hook.sh`,
  `handle-session-start-compact.sh`, `handle-session-start-navigator.sh`) are
  absent.
- `Mutant:` listing all 11 names under a single blanket heading ("dormant") would
  satisfy a name-presence grep while erasing the distinction the investigation
  established — that 2 are reachable and 3 are open questions.
  **Constructible — hence the criterion requires a disposition class per name,
  drawn from the five-class set, on the same line.**
- `Harness correction:` n/a — a documentation-content criterion; verified by
  reading the rendered rows.

### AC-HWD-010 — the two audit corrections are recorded

**Given** the same rule surface,
**When** it is read,
**Then** it states (a) 33 hook entries across 20 events with the 34th
`"type": "command"` occurrence being `statusLine`, and (b) that
`handle-agent-hook.sh` is registered via agent frontmatter, not settings.

- `Pre-impl observed:` `grep -c 'statusLine' .claude/rules/moai/development/hook-independence.md`
  → to be recorded at run time; `grep -c 'handle-agent-hook'` → `0`. The
  independent measurement backing the claim: `grep -c '"type": "command"'
  .claude/settings.json` → `34`, while a JSON walk of `d['hooks']` counts
  `local hook entries 33` across `events 20`.
- `Mutant:` writing "33 entries" without naming the `statusLine` cause leaves the
  next reader to re-derive 34 from the same grep and conclude the doc is wrong.
  **Constructible — hence the criterion requires the cause, not the number.**
- `Harness correction:` n/a.

### AC-HWD-011 — nothing was deleted

**Given** the project and template hook directories,
**When** their contents are counted and compared to the pre-change inventory,
**Then** the counts and the name sets are unchanged.

- `Pre-impl observed:` `ls .claude/hooks/moai/*.sh | wc -l` → `43`;
  `ls internal/template/templates/.claude/hooks/moai/ | wc -l` → `50`
  (12 `.sh` + 35 `.sh.tmpl` + 3 other).
- `Mutant:` none constructible — the criterion is a direct inventory comparison
  with no interpretive gap.
- `Harness correction:` n/a.

### AC-HWD-015 — Template-First order was followed

**Given** the commit(s) delivering M3,
**When** the local mirror and the template source are diffed,
**Then** `diff .claude/rules/moai/development/hook-independence.md
internal/template/templates/.claude/rules/moai/development/hook-independence.md`
reports no difference, and `make build` was run between the template edit and the
mirror.

- `Pre-impl observed:` the two files are currently **IDENTICAL** (`diff -q` →
  no output), so this criterion asserts that the property is preserved, and its
  falsifying case is an edit landing in only one of the two.
- `Mutant:` editing only the local mirror leaves the diff non-empty and fails —
  no mutant passes while violating the requirement.
- `Harness correction:` [HARD] deliberately verify by checking that an
  intentionally template-only edit **fails** the mirror diff before the mirror is
  written.

### AC-HWD-016 — the template edit is neutrality-clean

**Given** the template-side rule file,
**When** it is scanned for the forbidden content classes,
**Then** it contains no SPEC ID (`SPEC-[A-Z]`), no REQ token (`REQ-[A-Z]`), no
internal card number (`\bt[0-9]{2,3}\b`), no internal date, and no commit SHA;
and `go test ./internal/template/...` neutrality guards pass.

- `Pre-impl observed:` `grep -cE 'SPEC-[A-Z]|REQ-[A-Z]'
  internal/template/templates/.claude/rules/moai/development/hook-independence.md`
  → to be recorded at run time on the unmodified file, establishing the baseline
  the edit must not raise.
- `Mutant:` writing "open question — see card t244" in the template passes
  AC-HWD-009 (the name and a disposition class are present) while leaking
  internal state into 16-language distribution. **Constructible — that is the
  specific leak §C-4 forbids**, which is why the card numbers live in §G of the
  SPEC and the template rows name the pending decision instead.
- `Harness correction:` [HARD] construct a copy of the template file with a card
  number inserted, run the neutrality guard, and record the observed failure.
  A neutrality guard seen only passing has not been shown to be a guard.

---

## M4 — stop the MX dead work

### AC-HWD-012 — `moai mx query` builds the index when it is unavailable

**Given** a project directory containing `.moai/state/` with no `mx-index.json`
and at least one `@MX:DEBT` tag in a source file,
**When** `moai mx query --kind DEBT` runs there,
**Then** it exits `0`, returns the tag, and `.moai/state/mx-index.json` exists
afterwards.

- `Pre-impl observed:` measured on a constructed temp project at HEAD —
  stderr `SidecarUnavailable: sidecar index does not exist — run 'moai mx scan'
  to build the index`, TUI `ERROR Sidecarunavailable: no sidecar index.`,
  `EXIT=1`, and no `mx-index.json` written.
- `Mutant:` an implementation that catches the error and returns an **empty
  result set** with exit `0` passes an exit-code-only assertion while serving a
  wrong (empty) answer. **Constructible — hence the criterion asserts the tag is
  returned AND the index file exists afterwards, not merely the exit code.**
  A second mutant — auto-building only on *absent* and not on *stale* — is
  covered by the stale case below.
- `Harness correction:` [HARD] two failing inputs must be constructed and their
  behaviour observed: (a) an index whose `scanned_at` is older than the freshness
  threshold, which must also trigger a rebuild; (b) a corrupt (non-JSON)
  `mx-index.json`, which must trigger a rebuild rather than an error. Record both
  observations.

### AC-HWD-013 — the SessionStart cold-start scan is gone

**Given** `internal/hook/session_start.go`,
**When** the scan symbols are grepped,
**Then** `grep -c 'runMXColdStartScan'` → `0`, `grep -c 'mxScanNeeded'` → `0`,
and `grep -c 'mxIndexNeedsRebuild'` → `0`.

- `Pre-impl observed:` `runMXColdStartScan` → `5`; `mxScanNeeded` → `6`;
  `mxIndexNeedsRebuild` → `4`.
- `Mutant:` renaming the function while keeping the dispatch passes a
  name-grep-only check. **Constructible — hence the criterion is paired with
  AC-HWD-014's behavioural sibling and with a review requirement that the
  `spawnDeferredAdvisoryScans` signature no longer carries a scan parameter**;
  a rename that preserved the dispatch would leave that parameter in place.
- `Harness correction:` n/a for the grep itself. The removal's *safety* is
  covered by the affected-package test run: `go test ./internal/hook/...` must
  pass, and any test that asserted the scan's presence must be removed in the
  same change rather than skipped.

### AC-HWD-014 — the false goroutine-survival comment is gone

**Given** `internal/hook/session_start.go`,
**When** `grep -c 'durable side effects still land'` runs,
**Then** it prints `0`, and no remaining comment in the file asserts that a
goroutine spawned in the hook process continues after the process returns.

- `Pre-impl observed:` `1` — the single occurrence is at
  `internal/hook/session_start.go:253`, inside the join-bound timeout branch,
  reading *"The goroutine continues to completion in the background (durable side
  effects still land)"*. It contradicts the accurate comment at lines 1531-1533,
  which states the opposite: *"the SessionStart process may exit and kill the
  helper goroutine … the scan only lands if it finishes before the process
  exits"*.
- `Mutant:` deleting the phrase while leaving an equivalent claim in different
  words passes the grep. **Constructible — hence the criterion carries the
  second, review-verified clause** covering any remaining assertion of the same
  idea.
- `Harness correction:` n/a — a documentation-content criterion over a source
  comment.

---

## §D.5 Definition of Done

- All 16 criteria PASS, each with its command output recorded and its
  `Pre-impl observed:` value cited alongside.
- Every new gate (AC-HWD-003, 005, 006, 007, 008, 012, 016) carries an
  **observed failure** on a constructed failing input. A gate with no recorded
  failure observation blocks closure regardless of its passing run.
- `go test ./internal/cli/... -timeout 600s`, `go test ./internal/hook/...`,
  `go test ./internal/template/...` pass; `go vet ./...` clean;
  `golangci-lint run` clean on changed packages.
- `make build` run after the template edit; the mirror diff is empty
  (AC-HWD-015).
- No file under `.claude/hooks/moai/` or
  `internal/template/templates/.claude/hooks/moai/` was added or removed
  (AC-HWD-011).
- The three deferred decisions (§G-1/2/3), the unnumbered fourth defect (§G-4),
  and the deferred snapshot fix (§G-5) are present in the SPEC and were not acted
  on.
- Push, then read CI for the full-suite verdict. The local affected-package runs
  are an early signal, never the verdict.
