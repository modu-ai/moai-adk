# SPEC-KANBAN-RECORD-SESSION-KEY-001 — Acceptance Criteria

Nine criteria against eight requirements. Every criterion names a command and an expected result;
none is satisfied by reading a file and forming a judgement.

Where a criterion is satisfied by an **absence**, the pre-change baseline is stated, measured in
this tree at `dfbf828a6`, so a zero-hit result is new information rather than a vacuous pass. A
criterion that already passes on the untouched tree is a defect, not a criterion.

## §A The key, the writer, and the role

**AC-KRS-001** (REQ-KRS-001, REQ-KRS-003) — the load-bearing criterion, and the one that is
measurably **false** today. Given a launch of a companion or lane session under role `R`, and given
that the launched session's runtime delivers identifier `S` while the project's single identifier
slot holds a **different** identifier `P`, When that session starts, Then
`.moai/state/kanban/S.json` exists, its `session_id` equals `S`, its `role` equals `R`, and no file
named `P.json` was created by this launch.

Baseline, measured in this tree on 2026-08-24 — both halves fail today, and they fail in opposite
directions:

```
$ ls .moai/state/kanban/2beac221-….json .moai/state/kanban/c15d8434-….json .moai/state/kanban/3db058e1-….json
No such file or directory        ×3     ← all three live, registered sessions; none has a record

$ cat .moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json
{ "session_id": "d281730e-…", "spec_id": "", "role": "lane", "backend": "claude",
  "entered_at": "2026-08-23T17:47:22Z", "deepscan_dir": "", "verify_reentries": 0 }
                                        ← d281730e… is a LEAD session; its record's role reads "lane"
```

So the "record exists under its own identifier" half is false for every live session, and the "role
matches the session" half is false for the one record that exists. Neither half can pass before the
change, because the pre-change writer resolves its key before the described session exists.

**AC-KRS-002** (REQ-KRS-001) — Given a session whose runtime delivers identifier `S`, and a
`.moai/state/current-session-id.txt` containing a different identifier `T`, When the session starts,
Then the record written is `…/kanban/S.json` and no record named for `T` exists; and When the record
write path is grepped for the sidecar — `grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID' internal/kanban/ internal/hook/session_start*.go` — Then it returns zero
hits. Baseline: the same grep returns zero on the pre-change tree, so the grep half asserts
preservation; the load-bearing half is the two-identifier fixture, which cannot pass before the
change because the pre-change writer produces `T.json` by construction — that is exactly what
`resolveLaunchSessionID` (`internal/cli/launcher_blockcap_infinite.go:126-134`) reads.

**AC-KRS-003** (REQ-KRS-002) — Given the merged tree, two halves. **(a)** When
`grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/` is run, Then it returns zero hits.
**(b)** Given a launcher invocation executed with a temporary project root and a populated
`.moai/state/current-session-id.txt`, When the launcher returns, Then the count of files under
`<root>/.moai/state/kanban/` is unchanged from before the invocation. Baseline, measured:

```
$ grep -rn 'recordKanbanSession(' internal/cli/
internal/cli/cc.go:161   cc.go:175   cc.go:192   cc.go:208
internal/cli/glm.go:224  glm.go:237  glm.go:250  glm.go:264
internal/cli/kanban.go:472    (the definition)
```

Nine hits — eight call sites and one definition — and `kanban.go:478` is the single
`WriteBestEffort` call they all reach. Half (a) therefore goes from one hit to zero, and half (b)
from "one file created" to "no file created", so both are new.

## §B Lane and card identity

**AC-KRS-004** (REQ-KRS-004) — Given a session launched with the lane label `lane-3`, When its
record is written and read back, Then its lane-number field equals `3` and its `role` equals
`lane`; and given a session launched as a kanban lead, Then its lane-number field equals `0`.
Baseline: no lane-number field exists — `grep -n 'Lane\|Card' internal/kanban/record.go` returns
only two hits, both inside the role setter's known-set check (`record.go:119`, `:126`), and
`grep -l '"lane"\s*:' .moai/state/kanban/*.json` returns **0** of 84 files. The `lane-3` half
additionally guards the drop hazard: `WithRole` (`record.go:116-130`) admits only the known role
set, so an implementation that routes the lane number through it loses the `3` silently, and this
criterion fails.

**AC-KRS-005** (REQ-KRS-005) — Three halves, each a distinct source. **(a)** Given a session whose
worktree root is `<…>/.claude/worktrees/t207` and no override in its environment, When its record
is written, Then its card field equals `t207`. **(b)** Given the same session with the override
variable set to `t999`, Then its card field equals `t999`. **(c)** Given a session whose worktree
root cannot be resolved and no override set, Then the card field is absent from the encoded record
and the write still succeeds. Baseline: no card field exists (the same `grep -n 'Lane\|Card'` above
returns no struct field), and no environment key names one — `grep -rn 'CARD' internal/config/envkeys.go`
returns zero lines — so all three halves are new. Half (a) is verified against a real worktree
basename: `git rev-parse --show-toplevel` in this tree returns
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`.

**AC-KRS-006** (REQ-KRS-006) — Given the merged tree, two halves. **(a)** When
`grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go` is run, Then the
backend value no longer appears as a literal argument to a record write, and a backend environment
key is exported on both launcher paths. **(b)** Given a session launched through the GLM path, When
its record is read, Then its `backend` equals `glm`; and given one launched through the Claude path,
`claude`. Baseline, measured: the backend exists **only** as a literal at the eight call sites —
`grep -rn "BackendGLM\|BackendClaude" internal/ | grep -v _test` shows `cc.go:161,175,192,208` and
`glm.go:224,237,250,264` plus the two constant declarations at `record.go:23-24` — and
`grep -rn 'BACKEND' internal/config/envkeys.go` returns **zero** lines, so no environment key names
it today. Half (b) cannot pass before the change: a session-written record has no way to learn its
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

**AC-KRS-008** (REQ-KRS-008) — Given a session start in which the state directory cannot be written
— made unwritable for the duration of the fixture — When the session starts, Then (a) the session
start completes with a zero exit status and its normal output, (b) no error is surfaced to the
session, and (c) no record file exists. Baseline: the pre-change writer already discards every
failure (`WriteBestEffort`, whose `@MX:NOTE` states the absent error return **is** the fail-open
guarantee), so this criterion asserts **preservation across the move** — it is the one criterion
here that is expected to hold before and after, and it is included because moving the write into
the session start is exactly what could turn a silent record failure into a failed session.

**AC-KRS-009** (REQ-KRS-001) — the consumer-facing closure, stated as the parent SPEC states it.
Given a factory run with a registered lane whose `workers.json` entry carries PID `N`, and given the
session bearing PID `N` in `.moai/state/active-sessions.json`, When the chain
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban record` is resolved, Then it
returns a record whose `session_id` equals that entry's `session_id` and whose lane number equals
that lane's number. Baseline: the third hop returns nothing today — all three registered sessions
measured above (`2beac221…`, `c15d8434…`, `3db058e1…`) have no record file, so the resolution is
empty for every one of them, and the one record that exists belongs to a session absent from the
registry. This is the criterion that discharges `SPEC-WEB-CONSOLE-015` §A.5's claim that the join
"closes on today's data"; it does not, and this makes it.

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
- **A lane whose worktree basename is not a card identifier** (a lane started in the primary
  checkout, say). The basename is recorded as-is; the field says what the session was standing in,
  and no attempt is made to validate it against a card queue. Validating would require a queue read
  this SPEC does not take.
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
- [ ] The 84 existing files under `.moai/state/kanban/` neither migrated nor deleted.
