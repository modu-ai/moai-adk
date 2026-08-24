# SPEC Review Report: SPEC-KANBAN-RECORD-SESSION-KEY-001

Iteration: **2/2** (Tier M ceiling — there is no iteration 3)
Verdict: **PASS**
Overall Score: **0.857** harmonic mean (arithmetic mean 0.875); Tier M threshold **0.80**
Blocking findings: **3** — Optional findings: **4**
Iteration-1 findings that did NOT close: **0 of 11**
New findings from the full pass: **7** (3 blocking, 4 optional)

Reasoning context ignored per M1 Context Isolation. This audit reads only the three SPEC artifacts
at version 0.2.0, the code and live state they cite, the two sibling SPECs named for boundary
checking, and the iteration-1 report as the delta baseline. Nothing was repaired or edited.

Audited tree: `<worktree>`, HEAD **`00f207ad9`**, branch
`WT-web-live-todo`. The SPEC attributes its own re-measurement to `3c3a6fbf8`, the immediate parent
of the revision commit:

```
$ git log --oneline -5
00f207ad9 plan(t207): revise SPEC-KANBAN-RECORD-SESSION-KEY-001 after the iteration-1 FAIL
3c3a6fbf8 plan(t207): close the blocking findings from the A and B plan audits
ee039da30 plan(t207): split SPEC-WEB-CONSOLE-015 into four SPECs      <- iteration-1 audit tree
dfbf828a6 plan(t207): measure the session-id slots and find the kanban record key defect (t207)
```

Runtime state (`.moai/state/**`) is read under the **project root** `<project-root>/`,
which is what version 0.2.0 now correctly states. Verified: this worktree carries no
`.moai/state/kanban/` at all (`ls .moai/state/` -> `.gitkeep`, `config-cache.json`,
`context-usage.json`).

---

## Verdict summary

Iteration 1 FAILed at 0.750 with Testability at 0.50 — three verification statements that did not
verify what they claimed. **All eleven iteration-1 findings closed.** Every baseline the SPEC now
states was independently re-measured in this audit and **every one reproduced exactly**, including
the three that iteration 1 found broken. The revision is not a prose rewrite: the criteria are
materially more observable than they were.

The full pass then found seven things the delta could not have seen, three of them blocking. They
are cheaper than iteration 1's and none is a must-pass failure, so the aggregate carries a PASS —
but they are real, and one of them (F-1) is a **recurrence of the exact class** iteration 1 named as
D11, in a criterion iteration 1 did not examine for it.

Because Tier M's ceiling is 2 and this is iteration 2, the three blocking findings route as
**pre-run repairs against a PASSing SPEC**, not as a third audit round.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-KRS-[0-9]*' spec.md | sort -u` ->
  `REQ-KRS-001 … REQ-KRS-008`; `grep -c '^- \*\*REQ-KRS-' spec.md` -> `8`. Eight ids, sequential,
  zero-padded, no gap, no duplicate, each defined exactly once.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md` §B) only, per M3 § Scope. All eight match a GEARS pattern: 001/002/003 ubiquitous
  (`… shall be keyed by … and shall not be keyed by …`); **004 now reads `**While** a session is a
  factory lane`** — the state-driven pattern, which closes iteration-1 D10, where it read `Where`;
  005 ubiquitous with a trailing `**When** neither source yields a value … shall be left empty`;
  006 ubiquitous (`shall be observable … shall not be inferred`); 007 ubiquitous + `**When** a
  record written by a build predating this schema is read`; 008 `**When** a launch fact … is
  unavailable`. No Given-When-Then entry appears in the requirement layer. The Given-When-Then form
  of every `AC-KRS-xxx` is the correct **verification-layer** format and is graded under Group 4,
  not here.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present with correct types:
  `id`, `title`, `version: "0.2.0"` (quoted, bumped), `status: draft` (in the 8-value enum),
  `created: 2026-08-24`, `updated: 2026-08-24` (both ISO), `author`, `priority: P2`, `phase`,
  `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case
  alias. Extra `era`, `tier: M`, `related_specs` are additive.
- **[N/A -> PASS] MP-4 language neutrality.** Single-language scope (`internal/kanban`,
  `internal/cli`, `internal/config`, `internal/hook`); no template mirror exists for those paths.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Referenced SPECs are exactly
  `SPEC-WEB-CONSOLE-015` and `SPEC-SESSION-TELEMETRY-001`; both directories exist and both carry
  `status: draft` — neither retired/superseded/archived. No BLOCKING D7 finding.
- **[N/A -> PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall' spec.md` -> `0`.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory ->
  `rc=1`, no match. `research.md` correctly absent at Tier M. `plan.md` §H declares its two open
  items "**not** blockers on this plan", the correct disposition.

**No must-pass criterion fails.**

---

## Part 1 — The delta: did the eleven iteration-1 findings close?

**All eleven closed.** Each verified by re-running the measurement, not by reading the revised prose.

### D1 — `AC-KRS-002` grep contradiction — **CLOSED**

Iteration 1: the criterion's grep spanned `internal/kanban/ internal/hook/session_start*.go` and
returned six hits, three of them in the sidecar writer the SPEC forbids touching, so the criterion
was unsatisfiable without violating §D's t221 exclusion.

Version 0.2.0 re-scoped the file set to the **key-resolution surface** and widened the pattern:

```
$ grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID\|resolveCurrentSessionID' internal/kanban/ internal/cli/kanban.go
internal/cli/kanban.go:474:	sessionID := resolveLaunchSessionID("")
rc=0
```

**One hit, matching the SPEC's stated baseline exactly.** Zero after the change is now reachable —
it requires removing `kanban.go:474`, which is precisely what the SPEC intends — and observes a
removal rather than asserting a preservation. The criterion additionally explains *why*
`session_start.go` is excluded and states, with its own measurement, that the prohibition on the new
write path reading the sidecar is carried by the two-identifier fixture rather than by the grep
(`grep -rln 'kanban\.Write' internal/hook/` -> `rc=1`, no match — re-verified). That is the correct
repair, and it names the flaw it replaced.

### D2 — false no-consumer premise — **CLOSED, including the "ships alone" reasoning**

The brief asked specifically whether the milestone's reasoning survived the correction. It did not
survive — it was **removed**, which is the right outcome.

Re-measured:

```
$ grep -rn 'internal/kanban"' internal --include='*.go' | grep -v _test
internal/web/viewmodel_ops.go:23     internal/cli/kanban.go:25   …   (13 importers)
```

`plan.md` §A now opens "**No consumer file is edited — but a consumer exists, and this plan is
shaped around it.** Version 0.1.0 of this plan asserted there was none; that was wrong", reproduces
the thirteen-importer measurement, and names the mechanism (`loadKanbanRecords`
`viewmodel_ops.go:439-461`; `buildChain` `:227-233` collapsing with `byRole[role] = r`, last-write-
wins over `os.ReadDir` order).

The consequence is carried through into the milestone shape rather than left as a note. §D M2 is
now "Move the write into the session, carry the facts it cannot see, **and remove the launcher's
write**" — one milestone — with an explicit paragraph "Why this is one milestone and not two" that
names the old reasoning as false and its premise as measured. §G adds the anti-pattern "Splitting M2
into 'move the write' and 'remove the launcher's write'". M1's surviving "Ships alone" rests on a
different and correct claim ("Nothing writes the new fields yet, so the shipped behaviour is
unchanged") and explicitly names `internal/web`'s `loadKanbanRecords` among the readers that stay
compatible.

No sentence anywhere in the three artifacts still rests on there being no consumer.

### D3 — card identifier prescribing a guess — **CLOSED**

REQ-KRS-005 now reads: "otherwise from the basename of the session's worktree root **only where that
root's parent directory is named `worktrees`**. **When** neither source yields a value — an absent
override together with a worktree root that fails that containment test, or no resolvable worktree
root at all — the field shall be left empty rather than guessed."

Checked against the brief's three questions:

1. **Is the anti-guess clause now reachable?** Yes, and its reachability is demonstrated on live
   data rather than argued. Measured in this tree, `git rev-parse --show-toplevel` ->
   `<worktree>`, parent `…/.claude/worktrees` -> passes,
   yields `t207`. The live session `e46fcfef-1f5c-4f9c-beff-2ada72e26eb5` stands in the primary
   checkout (`cwd` per `active-sessions.json`), parent `<checkouts-parent>` -> **fails the test, field
   left empty**. The empty branch is entered by a session that exists right now.
2. **Is the containment test applicable by an implementation?** Yes. It is a structural test on one
   path component (`filepath.Base(filepath.Dir(root)) == "worktrees"`), needs no queue read, and the
   SPEC says so: `acceptance.md` §E — "The containment test is structural and cheap; validating the
   resulting value against the card queue would be a different check … and is not performed. A card
   worktree whose directory name is not a real card id therefore still records that name — the test
   constrains where the value may come from, not whether the card exists." The residual case is
   named rather than hidden.
3. **Is the `SPEC-WEB-CONSOLE-015` consumer contract still satisfied?** Yes. REQ-WC15-044 asks the
   console to present "the lane number, the card identifier, the SPEC identifier where one exists,
   the session state, and the stage"; REQ-WC15-043 already requires that a lane resolving to no
   session **or a resolved session with no record** is still presented "carrying its lane number and
   an explicit unresolved marker". An empty card field is inside that contract. Both REQ ids were
   verified to exist (`SPEC-WEB-CONSOLE-015/spec.md:171`, `:176`).

**What the constrained derivation yields for a primary checkout: empty** — verified against the one
live session standing in one, and pinned by a new criterion half, AC-KRS-005(c), which states the
value must *not* equal the basename `moai-adk-go`. Both L1 (`.claude/worktrees/<card>`) and L2
(`~/.moai/worktrees/<name>`) card trees pass the test, so it is not over-narrow.

### D4 — stale parent-section citations — **CLOSED, and the sweep is complete**

The brief asked for a sweep beyond the three places iteration 1 named. Executed:

```
$ grep -n 'SPEC-WEB-CONSOLE-015\|SPEC-SESSION-TELEMETRY-001' spec.md plan.md acceptance.md
spec.md:16, 26, 84, 143, 264, 274   acceptance.md:153, 220   plan.md:186, 263, 267
```

**Not one of the ten citations names a sibling section number.** They name REQ ids
(`REQ-WC15-043`, `REQ-WC15-044`, `REQ-ST-002`) or the SPEC as a whole. Every cited id exists:
`SPEC-WEB-CONSOLE-015/spec.md:171` (043), `:176` (044), `SPEC-SESSION-TELEMETRY-001/spec.md:131`
(ST-002). The withdrawal is now described in the past tense with the withdrawal noted (`spec.md:84`
onward: "Version 0.1.0 of that SPEC **asserted** … That assertion has since been **withdrawn
there**"), which matches the parent verbatim (`SPEC-WEB-CONSOLE-015/spec.md:112-113`).

The SPEC also states the rule that prevents recurrence, in two places: `spec.md` §A.3 ("Sibling
citations in this SPEC name REQ ids rather than section numbers: the parent was rewritten and its
sections moved; its REQ ids did not") and `plan.md` §I opening. Every remaining `§X.Y` token in the
three artifacts is a **self**-reference to this SPEC's own sections; all resolve.

### D5 — mutable baseline counts — **CLOSED; the replacement is checkable and smuggles nothing back**

The brief asked three things of the replacement. All three hold.

**(a) Is the property checkable?** Yes. The Definition of Done item is now: "take
`ls -la .moai/state/kanban/` before the change and again after, and confirm the 'after' listing
**contains every 'before' entry with an unchanged size and mtime**. New entries appearing between
the two listings are sessions doing their normal work and are not a violation; a missing, renamed,
or altered entry is." That is a mechanical comparison with an explicit drift/violation
discriminator — exactly the distinction iteration 1 said a reviewer could not make.

**(b) Does the Definition of Done still pin a count?** No. `grep -n '84' acceptance.md` returns
nothing; the fixed count is gone from all three artifacts. Every surviving number is attributed to a
measurement instant and framed as illustrating a property ("79 record files carried a `session_id`
when this was measured on 2026-08-24 … A count quoted here would be a different number by the time
anyone checked it").

**(c) Do the surviving numbers still reproduce?** Re-measured under the project root, right now:

```
$ ls .moai/state/kanban/*.json | wc -l                              ->  81
$ grep -l '"session_id"' .moai/state/kanban/*.json | wc -l          ->  79   <- SPEC says 79 OK
$ grep -L '"session_id"' .moai/state/kanban/*.json
.moai/state/kanban/backlog.json
.moai/state/kanban/leads.json                                              <- SPEC names exactly these two OK
$ grep -h '"role"' .moai/state/kanban/*.json | sort | uniq -c
  34   "role": "lane",   18 lead,  5 run,  4 review,  4 sync,  3 plan      <- SPEC says 34 lane OK
$ grep -l '"lane"[[:space:]]*:' .moai/state/kanban/*.json | wc -l   ->   0   <- SPEC says 0 OK
```

Four for four. The "`*.json` is not a record glob" convention is now stated in `spec.md` §C.4 **and**
in the `acceptance.md` preamble as a binding measurement convention, and it is correct.

### D6 — backend-only claim overstated — **CLOSED**

The claim is now scoped to the command that supports it, and the excluded sets are named:

```
$ grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go
cc.go:161,175,192,208    glm.go:224,237,250,264        <- eight lines, every one the record's backend argument OK
$ grep -rn 'BACKEND' internal/config/envkeys.go   ->   rc=1, zero lines OK
```

`spec.md` §A.5 and `acceptance.md` AC-KRS-006 both now add the parenthetical naming what a tree-wide
grep would additionally return (`record.go:23,24,75` plus nine hits on the unrelated same-named pair
in `internal/cli/mcp_convergence.go`) and why the scope is the two launcher files. The overstated
"only" is gone.

### D7 — `AC-KRS-003` evidence mismatch — **CLOSED**

The criterion now displays its **own** command's output as the baseline:

```
$ grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/
internal/cli/kanban.go:478:	kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend).WithRole(role))
```

reproduces verbatim — one hit. The nine-hit call-site listing is retained under the explicit heading
"Supporting context, **not** the criterion's own evidence", and re-measured: nine hits, eight call
sites plus the definition at `kanban.go:472`. Exactly as stated.

### D8 — comment cited as code — **CLOSED**

AC-KRS-004's baseline now reads "returns exactly two hits, **neither of them a field**:
`record.go:119` is a line of the role setter's doc comment … and `record.go:126` is the setter's
known-set check itself". Re-measured — `:119` is `// the kanban roles (lead + the three companions)
plus RoleLane, which a`, `:126` is `if role == RoleLead || role == RoleLane || isCompanionRole(role)
{`. The characterisation is now exact, and it strengthens the baseline (the point is that no *field*
exists, which the corrected wording states directly).

### D9 — REQ-KRS-006 prescribes transport — **CLOSED as a disclosed decision**

REQ-KRS-006 now says so on purpose: "The means is **deliberately constrained** rather than left to
the implementation: the fact shall be conveyed by the launcher through the launch environment, and
shall not be inferred from any other signal the session can see. (The rejected inference, and why it
is a guess rather than a measurement, is recorded in `plan.md` §F.)" A constraint that names itself
a constraint and cites the rejected alternative is a requirement, not leaked implementation. The
cited rejection exists (`plan.md` §G — deriving the backend from `ANTHROPIC_BASE_URL`).

### D10 — GEARS `Where` semantics — **CLOSED.** REQ-KRS-004 now uses `While`, the state-driven pattern.

### D11 — a criterion that passes on the untouched tree — **CLOSED for AC-KRS-008**

AC-KRS-008 gained an explicit half (a), "the **session-start record-write path is reached** and its
write attempt fails", with the baseline that makes it non-vacuous:

```
$ grep -rln 'kanban\.Write' internal/hook/    ->   (no match, rc=1)
```

re-verified. The criterion now states the reasoning outright: "Without it the remaining halves would
pass on any tree at all — a session that never attempts a record trivially starts cleanly — which is
the defect this document's preamble names." That is the correct repair.

**But the class recurred elsewhere — see F-1 below.**

---

## Part 2 — The full pass: findings the delta could not see

### F-1 (blocking) — `AC-KRS-007(b)` passes on the untouched tree, and is the one criterion with no baseline at all

The brief asked me to test **every** criterion against the pre-change tree. One fails that test.

**AC-KRS-007(b)**: "Given a record for a non-lane session carrying no card identifier, When it is
encoded, Then neither the lane-number key nor the card key appears in the JSON."

On the untouched tree the struct declares neither field:

```
$ grep -n 'Lane\|Card' internal/kanban/record.go
119:// the kanban roles (lead + the three companions) plus RoleLane, which a
126:	if role == RoleLead || role == RoleLane || isCompanionRole(role) {
```

Neither hit is a struct field. An encode of any `Record` today therefore produces JSON in which
neither key appears — **half (b) is satisfied on the pre-change tree, observing nothing.** Post-change
it is non-vacuous (it exercises `omitempty`), which is the identical shape as AC-KRS-008 — and
AC-KRS-008 discloses that shape and justifies it in four sentences. AC-KRS-007 does not.

AC-KRS-007 is moreover **the only criterion in the document with no baseline block**:

```
$ sed -n '/AC-KRS-007/,/AC-KRS-008/p' acceptance.md | grep -ic 'baseline'
0
```

against a preamble that commits "Where a criterion is satisfied by an **absence**, the pre-change
baseline is stated … A criterion that already passes on the untouched tree is a defect, not a
criterion." Half (b) is satisfied by an absence — two absent keys — and states no baseline. The
document's own standard is unmet in exactly one place, and it is the place iteration 1 did not look.

— Severity: **minor** — Class: **blocking** — Required fix: add the baseline AC-KRS-008 already
models: state the `grep -n 'Lane\|Card' internal/kanban/record.go` two-hit / neither-a-field result,
state that half (b) is trivially true pre-change because the fields do not exist, and state what it
buys post-change (that `omitempty` is actually set on both new fields, which a struct without the
tags would fail).

### F-2 (blocking) — `d281730e…` is asserted to be a **lead** session with no cited evidence, and a baseline half rests on it

`spec.md` §A.2 row 3 and `acceptance.md` AC-KRS-001's baseline both turn on the sentence
"`d281730e…` is a **lead** session; the record's `role` reads `lane`". The evidence cell shows only
a `cat` of the record file, which carries `"role": "lane"` and nothing about what the session
actually was. No command anywhere in the three artifacts establishes the session's true role.

I could neither confirm nor refute it. What I could measure:

```
$ cat .moai/state/kanban/leads.json
{"lead": {"pid": 58131, "registered_at": "2026-08-21T16:15:30Z"}}      <- carries no session id
$ grep -rl 'd281730e' .moai/ | head
.moai/state/kanban/backlog.json   .moai/logs/trace-d281730e-….jsonl   .moai/reports/session-d281730e-….md   …
```

The record file itself reproduces verbatim as the SPEC quotes it (`"role": "lane"`,
`"entered_at": "2026-08-23T17:47:22Z"`, mtime `2026-08-24T02:47:22` local = the same instant), so the
*record* half is attributed. The *session* half is not: the claim that this identifier belonged to a
lead is an unattributed premise under `verification-claim-integrity.md` §2, load-bearing for
AC-KRS-001's second baseline half ("the 'role matches the session' half is false for the one record
examined").

This matters because AC-KRS-001 is the SPEC's own "load-bearing criterion". Its **first** baseline
half is solid and independently re-verified — both live registered sessions lack a record:

```
$ ls .moai/state/kanban/5d3be9b8-….json .moai/state/kanban/e46fcfef-….json
No such file or directory   x2                    <- reproduces exactly as the SPEC states OK
$ cat .moai/state/current-session-id.txt
e46fcfef-1f5c-4f9c-beff-2ada72e26eb5              <- reproduces exactly OK
```

So the criterion survives on its first half whatever happens to the second. But as written, half of
the stated baseline cannot be checked by a reader.

— Severity: **major** — Class: **blocking** — Required fix: either cite the command that establishes
`d281730e…`'s role (a trace or session-report line naming it), or drop the role-mismatch sentence
and let the criterion rest on the half that is attributed — the registered-session-has-no-record
property, which is the stronger evidence anyway.

### F-3 (blocking) — `§A.3` claims "the first two hops hold"; measured live, the **second** hop is one-to-many

`spec.md` §A.3: "The first two hops hold — `FactoryWorkerEntry.PID` (`internal/kanban/factory_slots.go:38`)
and `Entry.PID` (`internal/session/registry.go:92`) both exist and the registry carries the runtime
identifier. The **third hop is where it breaks**."

The cited evidence establishes only that the two **fields exist**. The claim made is that the two
**hops hold** — that a PID resolves to *an* active-sessions entry. Measured on live data right now,
it does not:

```
$ python3 -c "import json; [print(e['pid'], e['session_id'], e['cwd'], e['started_at']) for e in json.load(open('.moai/state/active-sessions.json'))]"
51045 5d3be9b8-be19-42ab-8be1-7cb40b29c456 <project-root>/.claude/worktrees/t210 2026-08-24T08:36:43.405154Z
51045 e46fcfef-1f5c-4f9c-beff-2ada72e26eb5 <project-root>                        2026-08-24T09:02:55.161312Z

$ ps -p 51045 -o pid,command
51045 claude … --name lane-9 --settings /var/…/moai-kanban-51045-….json
```

**Two registry entries carry the same live PID.** This is structural, not a fluke —
`Registry.Register` (`internal/session/registry.go:166-199`) deduplicates by `SessionID` only and
appends a new entry with `PID: resolveSessionPID()` for every new session id, with no PID
uniqueness constraint; a long-lived process that starts a second session (a `/clear`) therefore
produces a second entry with the same PID, and both persist until a heartbeat-threshold `Purge`.
Here the two entries are 26 minutes apart on a still-running process.

Two consequences:

1. **§A.3's claim is stronger than its evidence** — the same overstatement shape as iteration-1 D6,
   in a load-bearing sentence rather than a narrative aside.
2. **AC-KRS-009's stated purpose does not follow.** The criterion says "this is the criterion that
   discharges the claim `SPEC-WEB-CONSOLE-015` made in its version 0.1.0 … it does not close, and
   **this criterion is what makes it close**." Fixing the record's key repairs the third hop only;
   with two entries sharing a PID the second hop still returns an ambiguous set, so the join does
   not close on this SPEC's deliverable alone. The criterion itself remains *evaluable* — it is
   stated as a constructed fixture ("Given a factory run with a registered lane whose `workers.json`
   entry carries PID `N`, and given the session bearing PID `N`"), and a test builds one entry — so
   this is not a vacuity defect. It is a scope-of-closure claim that measurement contradicts.

Note the parent already legislates the *other* half of this ambiguity: `SPEC-WEB-CONSOLE-015`
REQ-WC15-047 covers "two or more registered **lanes** carry the same process identifier". Neither
SPEC covers two active-sessions **entries** carrying one PID, which is the case live on this machine.

— Severity: **major** — Class: **blocking** — Required fix: narrow §A.3 to what is measured ("both
PID fields exist"; the second hop's uniqueness is not established), and either soften AC-KRS-009's
"what makes it close" to "what makes the third hop close" or add the PID-collision case to the
handover so `SPEC-WEB-CONSOLE-015` REQ-WC15-047 can be widened to cover the registry side. The
handover is the cheaper of the two and the one that keeps the boundary clean.

### F-4 (optional) — `spec.md` §C.1 says "three packages"; the SPEC's own table and `plan.md` say four

```
$ grep -n 'three packages' spec.md
211:Six to eight files across three packages, no always-loaded doctrine, no published documentation:
$ grep -n 'Packages touched' plan.md
44:| Packages touched | 4 (`internal/kanban`, `internal/config`, `internal/hook`, `internal/cli`) |
```

The §C.1 blast-radius table immediately above line 211 lists `internal/kanban`, `internal/cli`,
`internal/config`, and `internal/hook` — four. `acceptance.md` §F's Definition of Done runs
`go test` over the same four. The Tier M conclusion is unaffected either way.
— Severity: **minor** — Class: **optional** — Required fix: "four packages".

### F-5 (optional) — `plan.md` §C's upper bound (9) exceeds §B's stated file band (6-8)

§B: "Files | **6-8** — enumerated in §C". §C enumerates seven, then: "`internal/cli/cc.go` and
`internal/cli/glm.go` join at M2 if the eight call sites' signatures change, bringing it to **nine**
at the upper bound. Six to eight is the working estimate." The band and the upper bound disagree by
one file. Immaterial to Tier M (both sit inside the 5-15 band), but a reviewer checking the
enumeration against the band finds a mismatch.
— Severity: **minor** — Class: **optional** — Required fix: state §B as "6-9 (working estimate 6-8)".

### F-6 (optional) — a live `cwd` is quoted in a spelling the file does not contain

`spec.md` §A.5, `acceptance.md` AC-KRS-005(c), and `plan.md` §F all present
`<other-checkout>` as session `e46fcfef…`'s "`cwd` in `active-sessions.json`". The file
records `<project-root>`:

```
$ python3 -c "…" active-sessions.json   ->   e46fcfef-… cwd <project-root>
$ ls -di <project-root> <other-checkout>
253706617 …/MoAI/moai-adk-go
253706617 …/moai/moai-adk-go            <- same inode: case-insensitive filesystem, one directory
```

Same directory, so no conclusion changes — the parent is `MoAI`/`moai`, not `worktrees`, and the
containment test still yields empty. But the string is presented as a quotation from a file that
contains a different string, and the lowercase spelling is the one the **iteration-1 report** used
for a *different, now-gone* session (`2e3ace62…`). That is the signature of a carry-over rather than
a re-measurement, in a HISTORY entry claiming "Every measurement re-taken at `3c3a6fbf8`".
— Severity: **minor** — Class: **optional** — Required fix: quote the path as the registry records
it, and note the case-insensitive equivalence if the lowercase form is wanted in prose.

### F-7 (optional) — `plan.md` §D M2 restates the card derivation without its containment test

§D M2: "The card identifier itself is derived inside the session from the basename of
`git rev-parse --show-toplevel`, with the override preferred when set and the field left empty when
neither yields a value." The containment test — the whole substance of the D3 repair — is absent
from this sentence. It is stated correctly three other places in the same file (§F G-1, §G's
"Taking the worktree basename unconditionally" anti-pattern, and §F's `e46fcfef…` worked case), so
the plan as a whole is right; but the milestone body, which is what an implementer reads while
implementing M2, describes the unconstrained derivation the SPEC rejects.
— Severity: **minor** — Class: **optional** — Required fix: add "only where that root's parent
directory is named `worktrees`" to the M2 sentence, or point it at §F G-1.

---

## Premise Verification

**Claim under test** — `kanban.Record` is keyed by the launching session's identifier, because
`recordKanbanSession` resolves the key from a single project-wide slot and the launcher runs before
the session it launches exists.

**The mechanism reproduces exactly, every link, at every cited line.** Re-read at HEAD `00f207ad9`:

```
internal/cli/kanban.go:472   func recordKanbanSession(specID, backend, role string) {
                      :474     sessionID := resolveLaunchSessionID("")
                      :478     kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend).WithRole(role))
internal/cli/launcher_blockcap_infinite.go:126  func resolveLaunchSessionID(sessionOverride string) string {
                      :130     if id, _, ok := resolveCurrentSessionID(); ok { return id }
internal/session/registry.go:45   // … is overwritten on every SessionStart.
                      :52   const CurrentSideChannelFile = ".moai/state/current-session-id.txt"
internal/hook/session_start.go:313  sidecar := filepath.Join(input.ProjectDir, session.CurrentSideChannelFile)
                      :314    if writeErr := os.WriteFile(sidecar, []byte(input.SessionID), 0o600); …
```

All five §A.1 table rows verify at the line numbers cited. Version 0.2.0 also corrected the citation
from `:313` to `:314` — `:313` computes the path, `:314` performs the write with `input.SessionID`.
The correction is right.

**The consequence reproduces, on the SPEC's own freshly-cited data.** The brief asked whether the
SPEC's cited instance still reproduces or has aged again. **It reproduces** — every row of §A.2 was
re-measured and matched, hours after the SPEC was written:

| §A.2 claim | Re-measured now |
|---|---|
| `5d3be9b8…` and `e46fcfef…` live and registered | both present in `active-sessions.json` OK |
| neither has a record of its own | `ls` -> `No such file or directory` x2 OK |
| `d281730e…`'s record carries `"role": "lane"` | `cat` reproduces the quoted JSON verbatim OK |
| the slot names `e46fcfef…` | `cat .moai/state/current-session-id.txt` -> `e46fcfef-…` OK |

The one part that did **not** verify is the assertion that `d281730e…` was a *lead* session (F-2) —
which is a missing attribution, not a contradicted observation.

Compared with iteration 1, where the SPEC's three cited identifiers had all aged out, version 0.2.0
is a **stale-free, currently-reproducing citation**, and it now says so about itself: "The
identifiers above will age out as version 0.1.0's did. What does not age out is the property they
demonstrate."

**Verdict on the premise: sound.** Both faults — single slot, launcher-before-child — are verified
in code and observed in live data. What §A.3 overstates is not the premise but the *reach of the
fix* (F-3).

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity a reasonable engineer resolves consistently | The requirement layer itself is now unambiguous: REQ-KRS-005's containment test removes the two-implementation reading that scored this 0.75 in iteration 1 (D3 closed), and REQ-KRS-006 declares its constraint deliberately. What holds it below 1.0 is document-level: §C.1 says "three packages" where its own table, `plan.md` §B, and the DoD's `go test` line all say four (F-4); §B's 6-8 file band contradicts §C's nine upper bound (F-5); and `plan.md` §D M2 restates the card derivation without the containment test that is the point of it (F-7). None changes what an implementer builds; all three are contradictions a reader must resolve. |
| Completeness | 1.00 | 1.0 — all sections + frontmatter + Out of Scope | HISTORY (two rows, the 0.2.0 row naming the audit report and enumerating the repairs), §A Background (six subsections), §B Requirements, §C Constraints, §D Exclusions. `grep -c '^### Out of Scope — ' spec.md` -> **6**, each with specific `-` bullets. Frontmatter complete (MP-3). The three-artifact Tier M set is complete; `design.md`/`research.md` correctly absent. Version 0.2.0 additionally added the scoping fact iteration 1 did not ask for — that `.moai/state/` belongs to the project root, not the worktree — with the `ls` that proves it. |
| Testability | 0.75 | 0.75 — one criterion not precisely binary-testable as baselined; the rest are | The axis that failed. Materially repaired: AC-KRS-002's grep went from 6 hits (unsatisfiable without violating §D) to 1 hit (an observable removal); AC-KRS-003 now shows its own command's output; AC-KRS-004's baseline characterises both hits correctly; AC-KRS-008 gained the half that makes it non-vacuous; the count-pinned DoD became a before/after listing with an explicit drift-vs-violation rule. **Every one of the eight command-stated baselines was re-run in this audit and every one reproduced exactly** (AC-002 -> 1 hit; AC-003 -> 1 hit + 9 call sites; AC-004 -> 2 non-field hits, 0 lane-key files, 34 lane rows; AC-005 -> `rc=1` on CARD, real worktree root; AC-006 -> 8 lines, `rc=1` on BACKEND; AC-008 -> `rc=1`). Held below 1.0 by three: AC-KRS-007(b) passes on the untouched tree and is the only criterion with no baseline at all (F-1); AC-KRS-001's second baseline half rests on an unattributed session-role claim (F-2); AC-KRS-009's "what makes it close" is broader than the deliverable (F-3). Not 0.50 — no criterion contains a weasel word, every one names a command, and none is unsatisfiable. |
| Traceability | 1.00 | 1.0 — every REQ covered, every AC parented, no orphan | §D table maps 8 REQ -> 9 AC. Verified both directions: `grep -o 'AC-KRS-[0-9]*' acceptance.md \| sort -u` -> 001…009, all nine appear in the table; `grep -o 'REQ-KRS-[0-9]*' spec.md \| sort -u` -> 001…008, all eight appear with at least one criterion. REQ-KRS-001 carries three (001, 002, 009). No AC names a non-existent REQ. Sibling traceability is also clean: every cross-SPEC citation is a REQ id and every cited id resolves. |

**Aggregate = harmonic mean** (per `agent-common-protocol.md` § Skeptical Evaluation Stance — "score
quality as the harmonic mean of dimensions, not the average"):

```
4 / (1/0.75 + 1/1.00 + 1/0.75 + 1/1.00)
  = 4 / (1.33333 + 1.00000 + 1.33333 + 1.00000)
  = 4 / 4.66667
  = 0.857
```

**Harmonic mean 0.857. Arithmetic mean 0.875.** Both clear the Tier M threshold of 0.80, so unlike
iteration 1 the two means do not disagree across the threshold and the choice of mean does not
decide the verdict. The harmonic figure is the binding one.

**Budget** — 8 requirements, 9 criteria, against Tier M ceilings of **16 and 16** applied
independently. Both well inside; `plan.md` §B states the same numbers and adds the right discipline
("If scope grows past that, the answer is to cut scope, not to raise the ceiling").

---

## Defects Found (structured defect-list)

Delta findings (Part 1): **none open — all eleven iteration-1 findings closed.**

Full-pass findings (Part 2):

F1. `AC-KRS-007-VACUOUS-HALF-NO-BASELINE` — `acceptance.md` §C (AC-KRS-007(b)) — Half (b) is
satisfied on the untouched tree (neither key exists to appear), and AC-KRS-007 is the only criterion
in the document carrying no baseline block, against a preamble that requires one for every
absence-satisfied criterion. Recurrence of the iteration-1 D11 class in a criterion iteration 1 did
not test for it. — Severity: **minor** — Class: **blocking** — Required fix: add the baseline
AC-KRS-008 models — state the pre-change `grep -n 'Lane\|Card' internal/kanban/record.go` result,
state that half (b) is trivially true pre-change, and state what it buys post-change (that
`omitempty` is actually set on both new fields).

F2. `AC-KRS-001-UNATTRIBUTED-ROLE-CLAIM` — `spec.md` §A.2 row 3, `acceptance.md` §A (AC-KRS-001
baseline) — "`d281730e…` is a **lead** session" is asserted with no cited command; `leads.json`
carries no session id and I could neither confirm nor refute it. A baseline half rests on it.
— Severity: **major** — Class: **blocking** — Required fix: cite the command that establishes the
session's role, or drop the role-mismatch half and rest the baseline on the
registered-session-has-no-record property, which is attributed and stronger.

F3. `A3-SECOND-HOP-OVERSTATED` — `spec.md` §A.3, with `acceptance.md` §C (AC-KRS-009) — "The first
two hops hold" is stronger than the cited evidence (that two PID fields exist), and live data
contradicts it: two `active-sessions.json` entries carry the same live PID 51045, because
`Registry.Register` (`internal/session/registry.go:166-199`) deduplicates by session id only.
AC-KRS-009's "this criterion is what makes it close" therefore overstates the deliverable's reach.
— Severity: **major** — Class: **blocking** — Required fix: narrow §A.3 to "both PID fields exist";
soften AC-KRS-009 to the third hop; hand the registry-side PID collision to
`SPEC-WEB-CONSOLE-015` REQ-WC15-047, which already covers the workers.json side of the same shape.

F4. `PACKAGE-COUNT-CONTRADICTION` — `spec.md:211` — "three packages" against the four its own table,
`plan.md:44`, and the DoD's `go test` line all name. — Severity: **minor** — Class: **optional** —
Required fix: "four packages".

F5. `FILE-BAND-VS-UPPER-BOUND` — `plan.md` §B vs §C — the 6-8 band and the nine-file upper bound
disagree by one. — Severity: **minor** — Class: **optional** — Required fix: state §B as "6-9
(working estimate 6-8)".

F6. `CWD-QUOTED-IN-A-SPELLING-THE-FILE-LACKS` — `spec.md` §A.5, `acceptance.md` AC-KRS-005(c),
`plan.md` §F — `<other-checkout>` presented as the registry's recorded `cwd`; the file
records `<project-root>`. Same inode on this case-insensitive filesystem, so no
conclusion changes. — Severity: **minor** — Class: **optional** — Required fix: quote the recorded
spelling.

F7. `M2-RESTATES-DERIVATION-WITHOUT-THE-GUARD` — `plan.md` §D M2 — the milestone body describes the
unconstrained basename derivation, omitting the containment test stated correctly in §F G-1 and §G.
— Severity: **minor** — Class: **optional** — Required fix: add the containment clause to the M2
sentence or point it at §F G-1.

---

## Regression Check (iteration-1 defect disposition)

| # | Iteration-1 finding | Class then | Disposition |
|---|---|---|---|
| D1 | `AC-KRS-002` grep contradiction | blocking | **RESOLVED** — file set re-scoped, pattern widened, baseline 1 hit re-measured and reproduces; the replaced flaw is named in the criterion |
| D2 | M2 no-consumer premise false | blocking | **RESOLVED** — 13-importer measurement added, `internal/web` mechanism named, M2+M3 collapsed into one milestone, split added to §G anti-patterns; no sentence still rests on the false premise |
| D3 | card identifier prescribes a guess | blocking | **RESOLVED** — containment test added to REQ-KRS-005, pinned by new AC-KRS-005(c), demonstrated on a live primary-checkout session; consumer contract still satisfied |
| D4 | stale parent-section citations | blocking | **RESOLVED** — sweep complete, zero sibling section-number citations remain, all cited REQ ids exist, tense corrected, recurrence rule stated in two artifacts |
| D5 | mutable baseline counts | blocking | **RESOLVED** — "84" gone from all three artifacts, DoD is a before/after listing with a drift-vs-violation rule, surviving numbers all re-measured and reproduce (79 / 34 / 0 / two non-record files) |
| D6 | backend-only claim overstated | optional | **RESOLVED** — scoped to the two launcher files, excluded sets named |
| D7 | AC-KRS-003 evidence mismatch | optional | **RESOLVED** — criterion's own command shown; call-site listing relabelled supporting context |
| D8 | comment cited as code | optional | **RESOLVED** — `:119` described as the setter's doc comment |
| D9 | REQ-KRS-006 prescribes transport | optional | **RESOLVED** — declared a deliberate constraint with the rejected alternative cited |
| D10 | GEARS `Where` semantics | optional | **RESOLVED** — now `While` |
| D11 | AC-KRS-008 passes on the untouched tree | optional | **RESOLVED for AC-KRS-008** — half (a) added with its `rc=1` baseline. Class recurs in AC-KRS-007(b), reported as **F-1** |

**11 of 11 closed. No stagnation.** No defect appeared unchanged across iterations.

---

## Gaps — what was NOT observed

- **No Go test, build, or vet was run.** Per the brief, all code claims rest on reading source at
  cited lines and on targeted greps. `Registry.Register`'s PID behaviour (F-3) was read at
  `registry.go:166-199` and corroborated by live registry data; it was **not** exercised by
  starting a second session in one process and observing the append.
- **`d281730e…`'s true session role could not be established** (F-2). `leads.json` carries no
  session id; `backlog.json`, the trace log, and the session report all mention the identifier but
  none that I read states its kanban role. The finding is that the claim is unattributed, **not**
  that it is false — I did not refute it.
- **The `moai web` console was not started.** Whether misattributed rows render today is unobserved
  — the same gap `plan.md` §H records honestly for itself.
- **No factory run was launched.** AC-KRS-009's end-to-end join was verified structurally and by the
  absence of records for registered sessions, not by executing a factory run.
- **Process liveness was probed for PID 51045 only** (`ps -p 51045` -> a live `claude … --name
  lane-9`). `kill -0` was not run per registry entry, so "two live entries" is a registry claim
  plus one process observation, not two.
- **The record directory was read once per measurement**, not sampled over time. Counts above are
  true of an instant.
- **Cross-model audit not invoked.** No `audit_multi` / codex / GLM second opinion; single-auditor
  verdict.
- **The full `git diff ee039da30..00f207ad9` was not read line by line.** The delta judgments above
  come from re-running each finding's measurement against version 0.2.0's text, not from reading
  the patch.

---

## Residual Risk

- **The record directory and the session registry are live and moved during this audit.** Every
  count (79 / 81 / 34 / 0), every session identifier, and every "no record exists for S" observation
  is true of an instant. Version 0.2.0 now says this about itself, which is the correct handling —
  but it means a merge-time re-measurement (`acceptance.md` §F requires one) will see different
  numbers and must re-establish the *property*, not the figures.
- **F-3's practical severity depends on how often two sessions share a PID.** Measured: two of two
  live entries do, right now, on this machine. Whether that is typical or an artifact of this
  machine's kanban/factory usage is unmeasured. If typical, `SPEC-WEB-CONSOLE-015`'s lane view needs
  REQ-WC15-047 widened to the registry side before it can render a correct lane row even after this
  SPEC lands.
- **F-1 is cosmetic post-change and real pre-change.** After M1 lands, AC-KRS-007(b) genuinely
  exercises `omitempty`; the defect is that the document does not say so, against its own stated
  standard. An implementer who reads it as already-satisfied could ship the fields without the
  `omitempty` tags and still believe the criterion passed.
- **F-6 suggests one measurement was carried rather than re-taken**, in a revision whose HISTORY
  claims every measurement was re-taken. Every *other* measurement I checked did reproduce, so the
  claim is substantially true; but one carried string is enough to warrant checking the rest at
  merge rather than trusting the HISTORY line.
- **This audit read code at `00f207ad9`; the SPEC measured at `3c3a6fbf8`.** Every citation checked
  still resolves, so no drift was detected between those commits on the cited lines — only the cited
  lines were checked, not the whole diff.
- **A PASS with three blocking findings is a deliberate reading of the contract**, not an oversight.
  The verdict is anchored to the M5 must-pass firewall (none failed) and the rubric scores (0.857 >=
  0.80), per M6's rule that a list of findings must not be used to manufacture a verdict the scores
  do not support. The three blocking findings route as repairs, below.

---

## Recommendation

**PASS at 0.857** (harmonic; arithmetic 0.875) against a 0.80 Tier M threshold, no must-pass failure.

The iteration-1 FAIL is genuinely closed. The axis that carried it — Testability — was repaired by
making criteria observable rather than by rewriting prose around them: the unsatisfiable grep now
returns one hit and demands zero, the count-pinned Definition of Done became a before/after listing
with an explicit drift rule, and **every command-stated baseline in the document was re-run in this
audit and reproduced exactly**. The premise is not merely still true; its cited instance reproduces
today, where iteration 1 found the cited instance had aged out entirely.

Tier M's ceiling is 2 iterations and this is iteration 2, so the three blocking findings are routed
as **pre-run repairs against a PASSing SPEC**, not as a third audit round. In priority order:

1. **F-2 — attribute or drop the `d281730e…` lead-role claim.** It sits inside the SPEC's own
   load-bearing criterion's baseline. Dropping it costs nothing: the registered-session-has-no-record
   half is stronger, is attributed, and reproduces.
2. **F-3 — narrow §A.3 to what was measured and hand the registry-side PID collision to the parent.**
   Two live entries share PID 51045 and `Registry.Register` dedups by session id only, so the second
   hop is one-to-many by construction. `SPEC-WEB-CONSOLE-015` REQ-WC15-047 already covers the
   workers.json side of exactly this shape and is the natural home for the registry side.
3. **F-1 — give AC-KRS-007 the baseline every other criterion has.** Model it on AC-KRS-008, which
   handles the identical shape correctly four paragraphs later.

The four optional findings (F-4 … F-7) are surfaced for the orchestrator's discretion and are not
routed. Per M6 they were not used to manufacture a verdict: the harmonic mean and the must-pass
firewall carry the PASS on their own, and F-7 is the only one that could reach an implementer's
hands — the plan says the right thing about the card derivation in three places and the wrong thing
in the one place M2's implementer reads.

One thing worth carrying into run: **the SPEC's evidence discipline is now better than the code's.**
F-3 was found because §A.3 made a claim one degree stronger than its own citation, in a SPEC that
otherwise attributes everything. That is a good failure mode to have.
