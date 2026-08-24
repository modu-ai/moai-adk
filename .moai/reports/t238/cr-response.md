# t238 — CodeRabbit response record (PR #1606)

Card: t238 · SPEC-CODEX-SESSION-MSG-001 · branch `WT-codex-session-msg`
Review baseline: PR head `f33cd0564` (review of 2026-08-24 22:07 KST), 17 actionable
inline comments.

## Verdict summary

| # | Site | Severity | Disposition |
|---|------|----------|-------------|
| C1 | `store.go:329-343` ack/agent ids joined into paths | Critical | **ACCEPTED — fixed** (reproduced first) |
| C2 | `lock_unix.go:17-24` fd 0 sentinel | Major | **ACCEPTED — fixed** (reproduced first) |
| C3 | `lock.go:76` `filepath.Abs` error discarded | Trivial | ACCEPTED — fixed |
| C4 | `lock.go:98` / `lock_windows.go` CloseHandle failure | Minor | ACCEPTED — fixed (compile-verified only) |
| C5 | `envelope_test.go:235` data part accepts text | Major | ACCEPTED — fixed |
| C6 | `store_test.go:390` poll batch ceiling uncovered | Trivial | ACCEPTED — test added |
| C7 | `mcp_session_msg_test.go:295` required-schema unasserted | Minor | ACCEPTED — fixed |
| C8 | `i18n.js:627` grammar | Minor | ACCEPTED — fixed |
| — | `store.go:251` pending-mailbox depth cap | Trivial | **REJECTED — out of scope** |
| D1 | `verdict.md:99` + `progress.md` local-env values | Major | ACCEPTED — redacted with disclosure |
| D2 | `acceptance.md:18` shell pipelines | Major | **ACCEPTED, DIRECTION CORRECTED** |
| D3 | `acceptance.md:23` AC-CSM-015 too weak | Minor | ACCEPTED — replaced with ListTools test |
| D4 | `acceptance.md:55` traceability claim | Minor | ACCEPTED — reworded to match matrix |
| D5 | `acceptance.md:64` DoD checklist contradiction | Major | ACCEPTED — 8 checked, 2 left with cause |
| D6 | `plan.md:53` threshold inventory short by one | Minor | ACCEPTED — sixth added |
| D7 | `progress.md:95` `-run` pattern | Major | ACCEPTED — replaced (see below) |
| D8 | `design.md:58` A2A role enum | Major | **ACCEPTED via the documentation option** |

## The two findings that were reproduced before being accepted

Neither was taken on the reviewer's word. Both were reproduced as observed
failures on `f33cd0564` (evidence: `.moai/state/verify/t238/pre-fix-repro.md`)
and re-verified as blocked after the fix.

**C1 — traversal.** `Poll(<registered agent>, ackIDs=["<relative path escaping the
root>/victim"])` returned `err=<nil> ackedCount=1` and deleted a `.json` file
outside the state root. After the fix the same probe returns
`sessionmsg: ack_ids: malformed messageId "../../../../002/victim" (want msg-<hex16>)`
and the file survives.

**C2 — fd 0.** With stdin closed, `acquire()` obtained `fd=0`; `release()`
treated 0 as "unacquired", returned nil without closing, and the next
`acquire()` failed with `resource temporarily unavailable`. After the fix the
same probe re-acquires cleanly.

## D2/D7 — accepted, but the reviewer's prescribed direction was wrong

CodeRabbit asked us to remove the backslash before each pipeline separator while
**keeping** `\|` inside quoted grep patterns as regex alternation. Measured
against GitHub's own renderer (`gh api markdown -f mode=gfm`) on this file's real
header + delimiter + row, GFM unescapes `\|` inside table cells for **both**
positions:

```text
source:   grep -rn "exec.Command\|codex-jobs\|..." ... \| grep -v _test
rendered: grep -rn "exec.Command|codex-jobs|..."  ...  | grep -v _test
```

So the pipeline separator never needed fixing (removing its backslash would break
the table), and the alternation the reviewer told us to preserve is exactly what
breaks — BRE `grep` receives a literal `|`.

**The consequence is worse than a broken command: the ACs were vacuous.**
Control measurement on `internal/cli/`, a directory with 67 genuine matches:

```text
form                                          matches
--------------------------------------------  -------
source   BRE   grep    "a\|b"                      67
rendered BRE   grep    "a|b"      (GFM unescaped)    0
source   ERE   grep -E "a\|b"                        0
rendered ERE   grep -E "a|b"      (GFM unescaped)   67
```

AC-CSM-008 and AC-CSM-010 assert *absence* (0 rows), so the broken form returns 0
and passes silently — a forbidden token could be added and the AC would still
pass. `-E` fixes the rendered reader and breaks the source reader; neither
polarity works both ways.

**Adopted fix: remove the pipe entirely.** Every alternation-bearing check now
uses repeated `-e` patterns (`grep -rn -e "a" -e "b"`), which has nothing for GFM
to unescape, so the source and rendered forms are byte-identical. The `-run`
alternation became two `&&`-joined invocations for the same reason.

**Mutant test (t241).** A forbidden token was appended to `store.go`:

```text
old rendered BRE form : 0   ← would have passed
new -e form           : 1   ← fails, correctly
```

The tree was restored and re-verified at 0.

The `-run` form deserves separate mention because its failure is **green**:
`-run 'TestSessionMsg\|TestMoai...'` selects **zero** tests and still prints
`ok ... [no tests to run]` with exit 0. The replacement selects and runs tests.

## What was rejected, and why

**Pending-mailbox depth cap (`store.go:251`).** The observation is sound — a
looping sender can fill a mailbox until the 24h TTL elapses, and `Poll` reads
every pending file each call. But a depth ceiling is a new policy absent from the
SPEC's requirements, and it adds a rejection path to the MCP contract. That is a
behaviour change, not a review response. Recorded as a follow-up candidate.

## What was accepted in a different form

**D8 — A2A `role` enum.** The premise is correct: proto3 JSON serializes an enum
by NAME, so A2A v1 emits `"ROLE_USER"` / `"ROLE_AGENT"` and this broker's
lowercase `"user"` / `"agent"` are not "the JSON form of" that enum. Changing the
values would re-address every envelope already on disk — an on-disk contract
change, out of scope here. CodeRabbit offered "or document an explicit
conversion"; that option was taken. `design.md` §3.1/§3.2 now state the
divergence and the conversion A2A interop requires, and new test assertions pin
the serialized values so they cannot drift silently.

**D1 — local-environment values.** Redacted to placeholders. One site sits inside
a verbatim e2e evidence block; silently rewriting evidence would make the record
dishonest, so the redaction is disclosed in an adjacent note naming exactly which
three values were replaced and stating nothing else changed. `~/go/bin/moai` was
deliberately kept — it is home-relative, carries no account name, and the
surrounding procedure needs it.

**D5 — DoD checklist.** 8 items checked against cited evidence in `progress.md`;
2 left unchecked with the cause stated (MX tags and the 5-section report have no
quoted command/output in `progress.md`). No item was checked to tidy the list.

## Verification on the final tree

```text
go build ./...                                              exit 0
go vet ./internal/sessionmsg/ ./internal/cli/               exit 0
GOOS=windows go vet ./internal/sessionmsg/                  exit 0
go test ./internal/sessionmsg/ -race -count=1               ok  2.480s
go test ./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer' -count=1  ok  1.974s
golangci-lint run ./internal/sessionmsg/...                 0 issues.
golangci-lint run ./internal/cli/...                        0 issues.
local-env leak scan over PR-changed docs                    NONE
```

`go test ./...` was not run (repo rule — shared machine); CI on the PR head is
the authority for the full suite.

## Gaps — not observed

1. **Windows runtime behaviour of C4.** `GOOS=windows go vet` proves compilation
   only. `CloseHandle` failure handling was never executed. No Windows test.
2. **C3 has no test.** `filepath.Abs` fails only when `os.Getwd()` fails, which
   cannot be provoked in-process without destabilizing the test binary.
3. **Unfiltered `internal/cli` suite not run** (filtered only, per the repo's
   load rule).
4. **No cross-process flock contention run.** The lock fixes were exercised
   in-process, plus `-race`.
5. **AC-CSM-015's mutant was proved structurally, not executed** — the mutant
   requires editing `internal/cli/mcp_server.go`, outside the doc lane's scope.

## Residual risk

1. **`withAgentLock`'s error surface changed.** If `fn()` succeeds but
   `release()` fails, the operation completed on disk yet returns an error — so
   `Send` could report failure for a message that was written, and a retrying
   caller could double-send. The broker is at-least-once by construction
   (claim/ack), so a duplicate is not a contract violation; and on unix `close()`
   on a valid fd essentially cannot fail, leaving the exposure Windows-side and
   unobserved (Gap 1). Surfacing a leaked lock was judged worse to hide.
2. **Windows retry after a failed `CloseHandle`** calls `UnlockFileEx` on an
   already-unlocked range, which will itself fail. Retaining the handle still
   beats leaking it, but the retry is not clean. Unobserved.
3. **The same escaping hazard remains in two unflagged files** — `research.md:32`
   and `spec.md:55` both carry `grep -rn "session_msg\|session-msg"` as
   plan-time absence records. They were not among the 17 comments and were left
   untouched for scope discipline. Recorded here as a follow-up candidate: the
   defect class is identical and their recorded 0-hit results are equally
   unfalsifiable when read from the rendered page.

---

## Round 2 — review of `6a99ef262`

The re-review produced three new findings, all on artifacts this card itself
added. All three were verified against the tree and accepted.

**N1 (Major) — the t238 reports leaked what D1 had just removed.** Measured: 13
occurrences of an absolute worktree path, hostname, or pid across
`code-fixes.md` and `doc-fixes.md`. The lane reports quote the pre-fix values as
"before" evidence, which is a real reason to name them — but the point of D1 is
that no committed artifact carries workstation identity, and quoting it back is
still carrying it. Redacted to the same placeholders (`<repo>/…`, `<pid>`,
`<host>`), with a redaction notice at the top of each report so the edit to
verbatim command output is disclosed rather than silent. `cr-response.md` was
already clean (0 occurrences).

**N2 (Minor) — markdown lint.** 27 opening fences across the three reports were
unlabeled; each is command output, so each is now `text`. One substantive item
hid inside this "cosmetic" finding: the four-row comparison table in this file
carried raw `|` characters inside inline code, which GFM reads as cell
separators — the table rendered with the wrong column count. Rewritten as a
code block, which removes the escaping hazard rather than working around it.
Verified via `gh api markdown -f mode=gfm`: every table in all three reports now
renders with a consistent cell count (verdict table 18 rows × 4 cells).

**N3 (Minor) — a contradiction this card introduced.** Adding the §3.1 role
divergence note left §7 still saying the `message` block "maps directly" to A2A
`Message`. A future implementer following §7 would emit `"user"`/`"agent"` where
A2A requires `"ROLE_USER"`/`"ROLE_AGENT"`. §7 now states the required conversion
at the boundary and points at the §3.1 footnote.

### Round-2 gap

CodeRabbit's combined status on `6a99ef262` reads `state=success` with
description `Review rate limited`, NOT `Review completed`. Per the repo's
CodeRabbit rule that is a gap, not a pass — the status prints identically either
way. It is recorded as a gap despite a `Merge Risk: 🟡 Moderate · up to 6a99e`
line existing whose prefix does match the head, and despite the review having
demonstrably analyzed the new files: the two signals disagree, and the rule's
first condition is the one that failed.

---

## Round 3 — review of `3486bd616`

The gate condition that failed in round 2 is now met: CodeRabbit's combined
status on `3486bd616` reads `state=success` with description `Review completed`,
and the `Merge Risk: 🟡 Moderate · up to 3486b` prefix matches the head. Both
conditions of the repo's CodeRabbit rule hold for the first time on this PR.

One new finding, accepted.

**N4 (Minor) — the redaction notice over-claimed.** The notice added in round 2
said command output was verbatim "except for three values" and that "no other
byte was altered". That is false: some quoted blocks also abbreviate a long
argument or path for width — the SPEC directory shown as `.moai/specs/.../`, the
multi-file `grep` argument list shown as `<5개 문서>`, per-run `t.TempDir()` roots
shown as `/var/folders/.../`, and long subtest paths shown as `--- PASS: .../`.
An evidence notice that overstates its own fidelity is the same defect class as
an unobserved claim, so it is worth fixing rather than waving through as
cosmetic.

Fixed at the root where possible and by disclosure where not:

- `.moai/specs/.../progress.md` was restored to the full
  `.moai/specs/SPEC-CODEX-SESSION-MSG-001/progress.md` (4 occurrences). That
  path carries no workstation identity, so abbreviating it bought nothing.
- The remaining elisions are genuinely width-driven, so each report's notice now
  states its scope explicitly: the workstation substitutions are what the notice
  covers, and the `...` elisions are of an INVOCATION or a PATH, never of a
  result. No PASS/FAIL verdict, count, line number, or error string was
  shortened anywhere.

---

## Round 4 — the 2026-08-23 batch, never addressed

Triage of all 34 inline comments by timestamp showed the card's original scope
(17 comments, 2026-08-24 13:06) was only part of the picture: a batch of **13
from 2026-08-23 was still open**, verified one by one against the current tree.
Eight were code, five were documents; one was rejected.

### Three substantive defects, each reproduced before acceptance

**K1 — `Poll` discarded a committed claim.** After the mailbox lock released,
a heartbeat write failure returned `PollResult{}, err` — but the pending→claimed
move was already on disk with its `ClaimedAt` stamp. The caller got an empty
result and an error while real messages sat invisible until the 10-minute claim
TTL redeemed them. `Send` had the same shape: the envelope was written, yet the
caller got `("", err)` and a retry would double-send. The heartbeat is a
best-effort side effect, not part of the claim transaction; both now return the
committed result and let a stale `lastHeartbeat` self-heal on the next call.

**K2 — claim order was effectively random.** `listEnvelopes` reads with
`os.ReadDir`, i.e. lexical filename order, and filenames are
`msg-<random hex16>`. With `DefaultSessionMsgPollBatch = 16`, a mailbox holding
more than 16 pending messages returned a random subset rather than the oldest.
Now sorted by `SentAt` with a `MessageID` tiebreak before the ceiling applies.
REQ-CSM-006 specifies the ceiling but is silent on ordering, so FIFO is a
defensible interpretation rather than a stated requirement — recorded as such.

**K3 — `data` parts had no size ceiling.** Text was bounded at 65536 bytes;
`len(p.Data)` was never checked, so an arbitrarily large payload validated and
persisted. This is in scope where the separately-rejected pending-depth cap was
not: REQ-CSM-005 already requires the broker to validate a body-size ceiling, so
this is an unimplemented part of an existing requirement, not new policy.

Also fixed: `os.IsNotExist` → `errors.Is(err, fs.ErrNotExist)` at 13 sites
(the old form does not unwrap), a package global mutated without `defer`
restore, an unchecked type assertion that would panic and abort a whole package
run, and missing handler coverage for the optional `data` argument.

### Two findings where following the review would have made things worse

**"Change the catalogue count from 21 to 25."** Counted directly: the families
table lists exactly 21 tools, the session-messaging subsection adds 4, total 25
— matching both the document header and `catalog.go`. Changing 21 to 25 would
make the heading wrong about the table it labels. The number was right; the
label was ambiguous, so only the label changed.

**"Copy the provenance line into the template mirror."** Rejected. That line
carries a SPEC ID, and SPEC IDs are a forbidden content class in
`internal/template/templates/**` (C1). The mirror's omission is deliberate
neutralization, not drift. Side observation worth recording: the leak guard's
pattern matches specific prefixes (`SPEC-V3R6-` and siblings) and would NOT have
caught `SPEC-CODEX-SESSION-MSG-001` — the doctrine is broader than its guard.

Documents: `research.md`'s comparison table had 3 cells in two rows against a
4-column header (axis and content were merged); split so all five rows carry 4.
`spec.md` REQ-CSM-008 claimed the lazy sweep runs "at any broker call" while
only `Poll` sweeps — reworded to name `Poll` as the sweep point and to state the
consequence honestly (a non-polling receiver's expired files persist until its
next poll, which is a delay bounded by that poll, not indefinite retention).

### Round-4 gap — one unreproduced failure

The first `go test ./internal/sessionmsg/ -race` run after the lane finished
reported FAIL, and its output was NOT captured. Fifteen subsequent runs (one
`-count=10` and five separate processes) all passed, and `go vet`, which
compiles tests, is clean.

The timeline offers an explanation: that run started BEFORE the lane's idle
signal, so it likely observed a half-written tree mid-edit — the lane's own
report carries RED evidence for exactly the tests that would have been failing
at that moment. **That is a hypothesis, not a demonstration.** The failure was
not reproduced, and a not-reproduced concurrency failure is not a fixed one.
Recorded here so a later recurrence is read against this note rather than as a
first sighting.
