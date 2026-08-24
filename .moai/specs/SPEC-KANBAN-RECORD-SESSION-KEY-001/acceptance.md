# SPEC-KANBAN-RECORD-SESSION-KEY-001 — Acceptance Criteria

Nine criteria against eight requirements. Every criterion names a command and an expected result;
none is satisfied by reading a file and forming a judgement.

Where a criterion is satisfied by an **absence**, the pre-change baseline is stated, re-measured in
this tree at `3c3a6fbf8`, so a zero-hit result is new information rather than a vacuous pass. A
criterion that already passes on the untouched tree is a defect, not a criterion — and where one
here comes close to that shape, its baseline says exactly what makes the post-change pass new
(AC-KRS-008).

Two measurement conventions bind every criterion below:

- **Runtime state lives under the project root, not this worktree.** Every `.moai/state/…` path is
  read under `/Users/goos/MoAI/moai-adk-go/`; the worktree carries no `.moai/state/kanban/` at all.
- **`*.json` in the record directory is not a record glob.** `backlog.json` and `leads.json` sit
  there and carry no `session_id` (`grep -L '"session_id"' .moai/state/kanban/*.json` returns
  exactly those two). No criterion is stated as a file count of that directory: the directory grows
  while this SPEC is open, so a count is not re-measurable and a criterion pinned to one cannot
  distinguish drift from a violation.

## §A The key, the writer, and the role

**AC-KRS-001** (REQ-KRS-001, REQ-KRS-003) — the load-bearing criterion, and the one that is
measurably **false** today. Given a launch of a companion or lane session under role `R`, and given
that the launched session's runtime delivers identifier `S` while the project's single identifier
slot holds a **different** identifier `P`, When that session starts, Then
`.moai/state/kanban/S.json` exists, its `session_id` equals `S`, its `role` equals `R`, and no file
named `P.json` was created by this launch.

Baseline, re-measured 2026-08-24 under `/Users/goos/MoAI/moai-adk-go/` — both halves fail today, and
they fail in opposite directions:

```
$ ls .moai/state/kanban/5d3be9b8-….json .moai/state/kanban/e46fcfef-….json
No such file or directory        ×2     ← both live, registered sessions; neither has a record

$ cat .moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json
{ "session_id": "d281730e-…", "spec_id": "", "role": "lane", "backend": "claude",
  "entered_at": "2026-08-23T17:47:22Z", "deepscan_dir": "", "verify_reentries": 0 }
                                        ← d281730e… is a LEAD session; its record's role reads "lane"

$ cat .moai/state/current-session-id.txt
e46fcfef-1f5c-4f9c-beff-2ada72e26eb5    ← the slot names a session that has no record of its own
```

So the "record exists under its own identifier" half is false for every live session, and the "role
matches the session" half is false for the one record examined. Neither half can pass before the
change, because the pre-change writer resolves its key before the described session exists.

The identifiers will have aged out by merge. The criterion is written against the **property**, and
the merge-time re-measurement (§F) re-establishes it with whatever sessions are live then: *every
registered session lacks a record under its own identifier, and the records that exist are keyed by
a session other than the one they describe.*

**AC-KRS-002** (REQ-KRS-001) — Given a session whose runtime delivers identifier `S`, and a
`.moai/state/current-session-id.txt` containing a different identifier `T`, When the session starts,
Then the record written is `…/kanban/S.json` and no record named for `T` exists; and When the
**record's key-resolution surface** is grepped for the sidecar —
`grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID\|resolveCurrentSessionID' internal/kanban/ internal/cli/kanban.go`
— Then it returns zero hits.

Baseline, measured on the untouched tree:

```
$ grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID\|resolveCurrentSessionID' internal/kanban/ internal/cli/kanban.go
internal/cli/kanban.go:474:	sessionID := resolveLaunchSessionID("")
```

**One hit today, zero required after** — so the grep half observes the removal rather than asserting
a preservation. Both halves are new information.

The file set is the surface the requirement governs — where the record's key is *resolved* — and it
deliberately excludes `internal/hook/session_start.go`. That file writes the
`.moai/state/current-session-id.txt` sidecar itself (`session_start.go:313`), which §D's t221
exclusion and §F both require to remain untouched; a grep whose file set included it could return
zero only by deleting a writer this SPEC forbids touching. Version 0.1.0 of this criterion made
exactly that mistake (its file set returned six hits, three of them in the sidecar writer).

The prohibition on the *new* write path reading the sidecar is not carried by this grep — the new
file does not exist to grep on the pre-change tree (`grep -rln 'kanban\.Write' internal/hook/`
returns no match, `rc=1`). It is carried by the load-bearing half, the two-identifier fixture: a
write path that read the sidecar would produce `T.json` and fail the criterion, whatever it is
named. That fixture cannot pass before the change, because the pre-change writer produces `T.json`
by construction — which is precisely what `resolveLaunchSessionID`
(`internal/cli/launcher_blockcap_infinite.go:126-134`) reads.

**AC-KRS-003** (REQ-KRS-002) — Given the merged tree, two halves. **(a)** When
`grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/` is run, Then it returns zero hits.
**(b)** Given a launcher invocation executed with a temporary project root and a populated
`.moai/state/current-session-id.txt`, When the launcher returns, Then the listing of
`<root>/.moai/state/kanban/` is identical to the listing taken immediately before the invocation —
same names, no addition. (A fixture root, created by the test and populated by it, so this listing
is reproducible in a way the live directory is not.)

Baseline for half (a), the criterion's own command:

```
$ grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/
internal/cli/kanban.go:478:	kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend).WithRole(role))
```

**One hit today, zero required after.** Half (b) goes from "one file created by the invocation" to
"no file created". Both are new.

Supporting context, not the criterion's own evidence — the call sites that reach that single write:

```
$ grep -rn 'recordKanbanSession(' internal/cli/
internal/cli/glm.go:224   glm.go:237   glm.go:250   glm.go:264
internal/cli/cc.go:161    cc.go:175    cc.go:192    cc.go:208
internal/cli/kanban.go:472    (the definition)
```

Nine hits: eight call sites and one definition, all converging on `kanban.go:478`.

## §B Lane and card identity

**AC-KRS-004** (REQ-KRS-004) — Given a session launched with the lane label `lane-3`, When its
record is written and read back, Then its lane-number field equals `3` and its `role` equals
`lane`; and given a session launched as a kanban lead, Then its lane-number field equals `0`.
Baseline: no lane-number field and no card field exist on the struct —
`grep -n 'Lane\|Card' internal/kanban/record.go` returns exactly two hits, **neither of them a
field**: `record.go:119` is a line of the role setter's doc comment ("the kanban roles (lead + the
three companions) plus RoleLane, which a") and `record.go:126` is the setter's known-set check
itself (`if role == RoleLead || role == RoleLane || isCompanionRole(role) {`). No record on disk
carries a lane number either — `grep -l '"lane"[[:space:]]*:' .moai/state/kanban/*.json` returns
**0** files (measured 2026-08-24 under the project root; the property, not the file count, is what
this baseline asserts). The `lane-3` half
additionally guards the drop hazard: `WithRole` (`record.go:116-130`) admits only the known role
set, so an implementation that routes the lane number through it loses the `3` silently, and this
criterion fails.

**AC-KRS-005** (REQ-KRS-005) — Three halves, each a distinct source, and the third is the one that
pins the anti-guess clause. **(a)** Given a session whose worktree root is
`<…>/.claude/worktrees/t207` — parent directory `worktrees` — and no override in its environment,
When its record is written, Then its card field equals `t207`. **(b)** Given the same session with
the override variable set to `t999`, Then its card field equals `t999`. **(c)** Given a session
whose worktree root is a **primary checkout** — a root whose parent directory is not named
`worktrees`, such as `/Users/goos/moai/moai-adk-go` — and no override set, Then the card field is
absent from the encoded record and the write still succeeds; in particular the field does **not**
equal the root's basename `moai-adk-go`.

Baseline: no card field exists (the same `grep -n 'Lane\|Card' internal/kanban/record.go` above
returns two hits, neither a struct field), and no environment key names one —
`grep -rn 'CARD' internal/config/envkeys.go` returns zero lines, `rc=1` — so all three halves are
new. Half (a) is verified against a real worktree basename: `git rev-parse --show-toplevel` in this
tree returns `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`, whose parent is
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees`. Half (c) is verified against a real live session
rather than a hypothetical: `active-sessions.json` carries `e46fcfef-1f5c-4f9c-beff-2ada72e26eb5`
with `cwd: /Users/goos/moai/moai-adk-go`, whose parent is `/Users/goos/moai`. Half (c) is what
fails an implementation that takes the basename unconditionally — the shape REQ-KRS-005 forbids,
and the shape that would hand `SPEC-WEB-CONSOLE-015` REQ-WC15-044 a checkout name to render as a
card.

**AC-KRS-006** (REQ-KRS-006) — Given the merged tree, two halves. **(a)** When
`grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go` is run, Then the
backend value no longer appears as a literal argument to a record write, and a backend environment
key is exported on both launcher paths. **(b)** Given a session launched through the GLM path, When
its record is read, Then its `backend` equals `glm`; and given one launched through the Claude path,
`claude`. Baseline, measured with the criterion's own scoped command:

```
$ grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go
internal/cli/cc.go:161   cc.go:175   cc.go:192   cc.go:208     (kanban.BackendClaude)
internal/cli/glm.go:224  glm.go:237  glm.go:250  glm.go:264    (kanban.BackendGLM)
```

Eight lines, every one of them the backend argument to `recordKanbanSession`. The command is scoped
to the two launcher files deliberately: a tree-wide grep also returns the constant declarations and
their doc comment (`record.go:23,24,75`) and nine hits on an unrelated same-named constant pair in
`internal/cli/mcp_convergence.go`, none of which is the record's backend. And
`grep -rn 'BACKEND' internal/config/envkeys.go` returns **zero** lines (`rc=1`), so no environment
key names it today. Half (b) cannot pass before the change: a session-written record has no way to learn its
backend, which is precisely why REQ-KRS-006 exists rather than being assumed.

## §C Compatibility and failure

**AC-KRS-007** (REQ-KRS-007) — Two halves. **(a)** Given a record file **produced by the pre-change
writer** — generated in the test by marshalling the pre-change struct, not hand-authored — When it
is read by the post-change reader, Then the read succeeds, every pre-change key is present and
unchanged, and the lane number and card identifier are reported as not recorded; and When that
record is rewritten, Then its pre-existing keys are byte-identical to their input form. **(b)** Given
a record for a non-lane session carrying no card identifier, When it is encoded, Then neither the
lane-number key nor the card key appears in the JSON. The input to half (a) is constrained to
writer-produced bytes so key order and indentation match the marshaller's own output; a
hand-indented fixture would fail a correct implementation for reasons the requirement does not care
about.

**AC-KRS-008** (REQ-KRS-008) — Given a session start whose launch environment carries the facts a
record needs, and whose state directory cannot be written — made unwritable for the duration of the
fixture — When the session starts, Then (a) the **session-start record-write path is reached** and
its write attempt fails, (b) the session start completes with a zero exit status and its normal
output, (c) no error is surfaced to the session, and (d) no record file exists.

Baseline: **there is no session-start record-write path today.**

```
$ grep -rln 'kanban\.Write' internal/hook/
(no match, rc=1)
```

Half (a) is what makes the criterion non-vacuous: it cannot be satisfied on the untouched tree,
because the path whose fail-open behaviour it asserts does not exist there. Without it the remaining
halves would pass on any tree at all — a session that never attempts a record trivially starts
cleanly — which is the defect this document's preamble names. What the criterion buys post-change is
the property that survives the move: `WriteBestEffort` (`internal/kanban/record.go:174`) discards
every failure by design, its `@MX:NOTE` stating that the absent error return **is** the guarantee
rather than an oversight, and moving the write into the session start is exactly what could turn a
silent record failure into a failed session.

**AC-KRS-009** (REQ-KRS-001) — the consumer-facing closure, stated as the parent SPEC states it.
Given a factory run with a registered lane whose `workers.json` entry carries PID `N`, and given the
session bearing PID `N` in `.moai/state/active-sessions.json`, When the chain
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban record` is resolved, Then it
returns a record whose `session_id` equals that entry's `session_id` and whose lane number equals
that lane's number. Baseline: the third hop returns nothing today — both registered sessions
measured above (`5d3be9b8…`, `e46fcfef…`) have no record file, so the resolution is empty for every
one of them, and the records that do exist belong to sessions absent from the registry. This is the
criterion that discharges the claim `SPEC-WEB-CONSOLE-015` made in its version 0.1.0 — that the join
"closes on today's data" — and which that SPEC has since withdrawn: it does not close, and this
criterion is what makes it close. The consumer requirements it unblocks are REQ-WC15-043 (the join)
and REQ-WC15-044 (the lane number and card identifier the join delivers).

## §D Traceability

| Requirement | Criteria |
|---|---|
| REQ-KRS-001 | AC-KRS-001, AC-KRS-002, AC-KRS-009 |
| REQ-KRS-002 | AC-KRS-003 |
| REQ-KRS-003 | AC-KRS-001 |
| REQ-KRS-004 | AC-KRS-004 |
| REQ-KRS-005 | AC-KRS-005 |
| REQ-KRS-006 | AC-KRS-006 |
| REQ-KRS-007 | AC-KRS-007 |
| REQ-KRS-008 | AC-KRS-008 |

Eight requirements, nine criteria. Every requirement carries at least one criterion; every criterion
names its parent requirements and no criterion is orphaned.

## §E Edge cases

- **A session that is neither a kanban nor a factory session.** No launch facts in its environment,
  so no record is written — unchanged from today, where no launcher ran. The absence is the correct
  answer, not a degraded one.
- **A session standing outside a card worktree** (a lane started in the primary checkout, say). The
  card field is left **empty**: the basename is taken only where the worktree root's parent
  directory is named `worktrees`, so a checkout name never enters the field (REQ-KRS-005,
  AC-KRS-005(c)). The containment test is structural and cheap; validating the resulting value
  against the card queue would be a different check, requiring a queue read this SPEC does not
  take, and is not performed. A card worktree whose directory name is not a real card id therefore
  still records that name — the test constrains where the value may come from, not whether the card
  exists.
- **Two sessions starting concurrently.** Distinct runtime identifiers, distinct files, no shared
  slot — the case AC-KRS-001 makes structurally impossible to collide. No locking is introduced.
- **A record for a session that later dies.** Left in place; §C.4 of spec.md records the disposition
  for the whole directory, and this SPEC adds no reaper.
- **The override variable set to an empty string.** Treated as unset, so the worktree basename is
  used; an empty override must not blank a derivable value.

## §F Definition of Done

- [ ] All nine criteria pass, each with its cited command's verbatim output.
- [ ] Every absence-criterion's baseline re-measured at merge, not carried from this document.
- [ ] `go test ./internal/kanban/... ./internal/cli/... ./internal/hook/... ./internal/config/...`
      passes; the full-suite verdict is read from CI on the pull-request head.
- [ ] `go vet ./...` and `golangci-lint run` clean on the touched packages.
- [ ] No file under `internal/web` modified, and no file under `internal/statusline` modified — the
      two exclusion boundaries this SPEC shares with its siblings.
- [ ] `.moai/state/current-session-id.txt` and its writer unchanged (the t221 boundary).
- [ ] No pre-existing file under `.moai/state/kanban/` is modified, renamed, or removed by this
      change. Checked as a property rather than a count, because the directory is live state that
      grows on its own while the SPEC is open: take `ls -la .moai/state/kanban/` before the change
      and again after, and confirm the "after" listing **contains every "before" entry with an
      unchanged size and mtime**. New entries appearing between the two listings are sessions doing
      their normal work and are not a violation; a missing, renamed, or altered entry is.
