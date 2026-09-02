# t238 — CodeRabbit code findings K1-K8, round 4 (PR #1606, SPEC-CODEX-SESSION-MSG-001)

> **Redaction notice (pre-merge).** Command output in this report is verbatim except for
> workstation-specific values, replaced with placeholders so the committed evidence
> discloses no developer or host identity: the absolute worktree path → `<repo>/.claude/worktrees/t187`,
> the process id → `<pid>`, and the hostname → `<host>`. Per-run `t.TempDir()` roots are
> shown as `<tmp>/...`. Every observed VALUE — SHAs, counts, line numbers, message ids,
> timestamps, and exit codes — is as measured. Path elisions are never elisions of a
> result: no PASS/FAIL verdict, count, or error string was shortened.

**Baseline-attribution (global).** All measurements were taken in the worktree
`<repo>/.claude/worktrees/t187`, branch `WT-codex-session-msg`, base commit
`3486bd61666f28788cbe6d8a1719187702207da0`. Confirmed at session start:

```text
$ pwd && git rev-parse --show-toplevel && git rev-parse HEAD && git branch --show-current
<repo>/.claude/worktrees/t187
<repo>/.claude/worktrees/t187
3486bd61666f28788cbe6d8a1719187702207da0
WT-codex-session-msg
```

The tree was left **uncommitted**; the orchestrator owns the commit. Files touched are
confined to `internal/` (6 modified, 2 new). The concurrently-modified files under
`.moai/specs/**`, `.moai/reports/**`, and `.claude/rules/**` belong to a parallel lane and
were not touched by this round — see § Scope verification.

---

## Claim (summary)

| Finding | Disposition |
|---|---|
| K1 — Poll/Send discard a committed result on heartbeat failure | FIXED (option (a), justified below) |
| K2 — claim order effectively random | FIXED (FIFO by `sentAt`, `messageId` tiebreak) |
| K3 — `data` parts unbounded | FIXED (`DefaultSessionMsgMaxDataBytes`, anchored to the text ceiling) |
| K4 — `os.IsNotExist` does not unwrap | FIXED (13 call sites converted: 8 non-test, 5 test) |
| K5 — unresolvable project dir handled implicitly | FIXED (explicit guard; reachability argued, not assumed) |
| K6 — package global mutated without `defer` restore | FIXED (closure + deferred restore; table grouping declined, reason below) |
| K7 — unchecked type assertion in a test | FIXED (comma-ok + `t.Fatalf`) |
| K8 — `data` argument had no handler-level coverage | FIXED (2 new handler tests; **no defect found** — coverage gap only) |

---

## Evidence (final tree)

Every command below was run against the final state of the tree, after all eight findings
were addressed.

```text
$ go build ./...
(no output; exit 0)

$ go vet ./internal/sessionmsg/ ./internal/cli/
(no output; exit 0)

$ GOOS=windows go vet ./internal/sessionmsg/
(no output; exit 0)

$ go test ./internal/sessionmsg/ -race -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	5.398s

$ go test ./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer' -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	2.806s

$ golangci-lint run --timeout=3m ./internal/sessionmsg/...
0 issues.
```

One command outside the assigned batch was added because the change touches
`internal/config/defaults.go` (K3), and that package has its own suite:

```text
$ go test ./internal/config/ -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/config	15.594s
```

`go test ./...` was **not** run (forbidden — shared machine).

New-test roll call, from the verbose run of the `internal/cli` selector:

```text
--- PASS: TestSessionMsgStoreRootRefusesUnresolvableProjectDir (0.00s)
--- PASS: TestSessionMsgSendHandlerCarriesDataArgument (0.01s)
--- PASS: TestSessionMsgSendHandlerRejectsUnencodableData (0.01s)
```

### RED evidence (before the fixes)

Each behavioural fix was driven test-first. The four new `internal/sessionmsg` tests were
run against the unfixed implementation and failed for their intended reasons — not merely
"failed":

```text
$ go test ./internal/sessionmsg/ -run 'TestPollReturnsClaimWhenHeartbeatFails|TestSendReturnsMessageIDWhenHeartbeatFails|TestPollClaimsOldestFirst|TestDataPartSizeCeiling' -count=1 -timeout 120s
--- FAIL: TestPollReturnsClaimWhenHeartbeatFails (0.91s)
    round4_test.go:83: poll failed on a heartbeat error: sessionmsg: create temp in <tmp>/agents: open <tmp>/agents/.codex-98d6c52f.json-1867612835.tmp: permission denied
--- FAIL: TestSendReturnsMessageIDWhenHeartbeatFails (0.83s)
    round4_test.go:124: send failed on a heartbeat error: sessionmsg: create temp in <tmp>/agents: open <tmp>/agents/.claude-95bd9ffd.json-3571219146.tmp: permission denied
--- FAIL: TestPollClaimsOldestFirst (0.49s)
    round4_test.go:186: claim[0] = "msg-039cb140242d098b", want "msg-66eec2e7ea6353ee" (oldest-first)
    round4_test.go:186: claim[1] = "msg-1b461157d781a21d", want "msg-669f4df9fb603cc5" (oldest-first)
    round4_test.go:186: claim[2] = "msg-631d1c4e7c2b7d33", want "msg-deb0ff8d922d0f78" (oldest-first)
    round4_test.go:189: claim[2] sentAt 2026-08-23 12:05:00 +0000 UTC precedes claim[1] 2026-08-23 12:07:00 +0000 UTC — not ascending
--- FAIL: TestDataPartSizeCeiling (0.00s)
    round4_test.go:234: data payload one byte over the ceiling accepted
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/sessionmsg	3.020s
```

The `TestPollClaimsOldestFirst` output is worth keeping as the direct observation behind
K2: with `DefaultSessionMsgPollBatch` shortened to 3 and eight messages queued a minute
apart, the pre-fix claim returned three messages that were neither the oldest three nor in
any time order — the third claimed envelope was stamped two minutes *before* the second.
That is the random-subset behaviour, measured rather than inferred.

---

## Per-finding detail

### K1 — heartbeat failure destroyed a committed result — FIXED

**What was wrong.** In both `Send` and `Poll`, the heartbeat runs *after* the mailbox lock
scope closes, by which point the durable effect is already on disk: `Send` has written the
envelope into the recipient's `pending/`, and `Poll` has committed the pending→claimed
move with its `ClaimedAt` stamp. Returning `PollResult{}, err` / `"", err` therefore threw
away a result that had already happened.

**Option chosen: (a) — return the result, drop the heartbeat error.** Applied to both
`Send` and `Poll`, so the two paths stay symmetric. The justification:

- The heartbeat's only observable product is `AgentRecord.LastHeartbeat`, which feeds one
  thing: the `online` flag `ListAgents` computes fresh on every call. A missed refresh
  degrades a liveness *hint*; it corrupts no delivery state.
- It self-heals. The next successful `Send` or `Poll` by that agent rewrites the field,
  and `heartbeat` is already idempotent on a missing record.
- Option (b) — a warning field on `PollResult` — was rejected on asymmetry: `Send` returns
  `(string, error)` and has nowhere to put such a field, so (b) would either fix only half
  the finding or force a wider signature change than the defect warrants.
- Returning `(result, err)` with both non-nil was rejected as ineffective: the MCP handler
  discards the result on any non-nil error (`if err != nil { return toolErr(...) }`), so
  the messages would still be lost. This is asserted, not assumed — the K1 tests require
  `err == nil`, so that variant fails them.

The trade-off accepted and recorded: a *persistently* failing heartbeat (an agents
directory unwritable for a long time) now leaves the agent reporting `online: false` with
no error surfaced anywhere in this package. That is stated under Residual-risk below.

**Mutant a shallower test would pass.** An implementation that keeps
`return PollResult{}, err` and merely *logs* the heartbeat failure passes any test that
asserts only "Poll returns no error" — because it does not; it still returns the error.
The regression lives in the returned *batch*, so both tests assert `len(res.Messages)` /
the returned `messageId`, and additionally stat the on-disk `claimed/` and `pending/`
files to pin that the effect really was committed before the heartbeat ran. A second
mutant — returning the result *and* a non-nil error — is caught by the `err == nil`
assertion, for the handler reason above.

The tests also carry an explicit precondition (`s.heartbeat(...)` must fail under the
read-only mode) so a green result cannot come from the induced failure silently not
happening — a test that passes because its fault injection stopped working is the failure
mode these two tests are most exposed to.

### K2 — claim order was effectively random — FIXED

`live` comes from `listEnvelopes(pendingDir)` → `os.ReadDir`, i.e. lexical filename order,
and filenames are `msg-<random hex16>`. The claim loop is now preceded by a sort on
`Delivery.SentAt` ascending, with `Message.MessageID` as a stable tiebreak for envelopes
sharing a clock tick.

**On the SPEC.** Stated plainly, as instructed: **REQ-CSM-006 specifies the batch ceiling
and is SILENT on ordering.** FIFO is a defensible *interpretation*, not a stated
requirement — the argument for it is that it makes the ceiling a delivery *delay* rather
than a lottery, which is what a caller draining a mailbox would expect. It is not a
correction of a spec violation, and this report does not claim it as one. If the SPEC
owner prefers a different order, the sort comparator is the single place to change.

**Mutant a shallower test would pass.** A DESCENDING sort still returns exactly `batch`
messages and still drains the mailbox over repeated polls, so a count-only assertion
passes it. The test pins the *identity* of the claimed subset (the oldest N, compared
against the recorded send order) and its ascending order, so both a reversed sort and a
no-sort implementation fail.

### K3 — `data` parts had no size ceiling — FIXED

`Message.Validate` accumulated `totalText` and checked it, but never looked at
`len(p.Data)`. Added `config.DefaultSessionMsgMaxDataBytes` to the existing session-msg
`var` block in `internal/config/defaults.go`, and a `totalData` accumulator + check in
`Message.Validate`, following the existing error-message style
(`"...: data size %d exceeds ceiling %d"`).

**Value: 65536, anchored to `DefaultSessionMsgMaxTextBytes`.** The anchor, stated rather
than picked: REQ-CSM-005 names ONE body-size ceiling, and a data part is body content, so
a *different* number would be a second policy with nothing behind it. Mirroring the text
ceiling means the requirement continues to have one value; if that value is retuned, both
payload kinds move together. The two are accumulated and bounded *independently* rather
than as a joint sum, matching how `Validate` already treats text — a joint sum would be a
behaviour change to the existing, already-shipped text bound.

**Mutant a shallower test would pass.** An off-by-one guard (`>=` instead of `>`) rejects
an exactly-at-ceiling payload while still rejecting oversize ones, so an over-ceiling-only
test passes it. The test asserts both sides of the boundary — exactly-at-ceiling accepted,
one byte over rejected — and checks the rejection reason names `data size`, so a payload
rejected for an unrelated reason (bad JSON, say) does not count as a pass.

### K4 — `os.IsNotExist` → `errors.Is(err, fs.ErrNotExist)` — FIXED, 13 sites

**Count changed: 13** — 8 in non-test code, 5 in tests:

| File | Sites | Where |
|---|---|---|
| `internal/sessionmsg/store.go` | 6 | `Send` sender, `Send` receiver, `Poll` agent, `listEnvelopes`, `removeIfExists`, `removeFirstOf` |
| `internal/sessionmsg/agent.go` | 2 | `heartbeat`, `readAllAgents` |
| `internal/sessionmsg/store_test.go` | 4 | `os.Stat` absence assertions |
| `internal/sessionmsg/edge_test.go` | 1 | `os.Stat` absence assertion |

Counted from the diff, which removes 14 `os.IsNotExist` lines — 13 call sites plus one doc
comment:

```text
$ git diff -U0 internal/sessionmsg/ | grep -c '^-.*os\.IsNotExist'
14
```

Zero remain anywhere in the touched surface:

```text
$ grep -rn 'os.IsNotExist' internal/sessionmsg/ internal/cli/mcp_session_msg.go
(no output — zero remaining)
```

The `agent.go` doc comment on `readAgent` that described the returned error as
"`os.IsNotExist`-able" was updated to name `fs.ErrNotExist` / `errors.Is`, so the comment
does not outlive the predicate it documents.

The test-side 5 operate on unwrapped `os.Stat` results, so their behaviour is unchanged;
they were converted for consistency, since the sweep was specified as covering all of
`internal/sessionmsg`.

Behavioural significance of the non-test 8: `readAgent` returns
`atomicfile.ReadFile`'s error. If that helper ever wraps (today it does not, which is why
this was latent rather than live), `os.IsNotExist` would stop matching and an unregistered
agent would surface as an opaque read error instead of the structured `UnknownAgentError`
carrying the known-agent list. `errors.Is` is robust to that change.

### K5 — unresolvable project directory — FIXED (reachability argued)

The instruction permitted a reachability rebuttal. I did not take it, because the concern
is reachable — but the reachable failure is narrower than "operates on the wrong
directory", and the fix is scoped to what is actually demonstrable.

`newSessionMsgStore()` built `filepath.Join(resolveProjectDir(), DefaultStateRoot)`.
`resolveProjectDir()` returns `""` when `$CLAUDE_PROJECT_DIR` is unset AND `os.Getwd()`
fails. `filepath.Join("", ".moai/state/session-msg")` yields the **relative**
`".moai/state/session-msg"`, silently re-anchoring the broker on whatever the process CWD
resolves to.

Why it is reachable rather than theoretical: `os.Getwd` fails when the working directory
has been removed underneath the process, and the moai MCP server is explicitly long-lived
(the documented consequence being that it does not see tools added after it started). A
worktree disposed while a server is anchored inside it is exactly that shape, and this
project disposes worktrees routinely.

There is precedent for treating the empty string as a real case rather than an
impossibility: `internal/cli/session.go:224` already guards
`if projectDir := resolveProjectDir(); projectDir != ""`.

**Fix.** A pure `sessionMsgStoreRoot(projectDir string) (string, error)` refuses an empty
project dir with an error naming the cause; `newSessionMsgStore()` now returns
`(*Store, error)` and all four handlers surface it via `toolErr`, i.e. as a structured
tool error and not a Go error. Extracting the guard as a pure function is what makes it
testable without breaking the test process's own CWD.

**Scope restraint (deliberate).** I did **not** change the shared resolution itself.
`$CLAUDE_PROJECT_DIR` set to a *relative* path would also anchor the store relatively, but
that is the behaviour of `resolveProjectDir` shared by every moai MCP tool; making this one
tool diverge would be a scope violation, and the file's own comment states the tools
resolve state "the same way every other moai MCP tool" does. Flagged here, not fixed.

**Mutant a shallower test would pass.** A guard that returns
`filepath.Join("", DefaultStateRoot)` with a nil error still behaves correctly for every
caller whose CWD happens to be the project root, so a happy-path-only test passes it. The
test asserts the empty input is an ERROR *and* that a resolvable input still joins
correctly, so a guard that refuses everything fails too.

### K6 — package global mutated without `defer` restore — FIXED

`TestSendPollAck` set `config.DefaultSessionMsgMaxTextBytes = 8` and restored it on the
next line. Any early `t.Fatal` between the two would leak the shortened ceiling into every
later test in the package.

The override now lives in an immediately-invoked closure with a deferred restore. The
closure — rather than a plain `defer` at test scope — is load-bearing in both directions:
`t.Fatal` calls `runtime.Goexit`, which **does** run deferred functions, so the restore is
safe against an early exit; and closing the scope immediately means the assertions that
follow in the same test (the empty-message and invalid-JSON rejections) still run against
the production ceiling rather than against 8. A function-scoped `defer` would have fixed
the leak while silently changing what those later cases exercise.

**Table grouping: declined.** The reviewer's suggestion was conditional on it not
obscuring which case failed. The four rejection cases here are heterogeneous — one needs a
threshold override and a restore, one passes empty arguments, one passes malformed JSON —
so a table would need a per-row setup hook, and the shared assertion message would lose
the specific wording each case currently carries. The cases already report distinctly
("oversize text accepted", "empty message accepted", "invalid JSON data accepted").

### K7 — unchecked type assertion in a test — FIXED

`kinds[am["kind"].(string)] = true` panicked on a missing or non-string `kind`, aborting
the entire `internal/cli` package run and hiding the assertion message that would have
explained it. Converted to comma-ok with `t.Fatalf`, matching every neighbouring read in
the same loop.

### K8 — `data` argument had no handler-level coverage — FIXED (no defect found)

Verified: zero occurrences of `data` in `internal/cli/mcp_session_msg_test.go` before this
round, making `sessionMsgDataArg` the only argument-parsing branch in the handler layer
with no test.

**The behaviour was already correct.** The new test passed on first run, before any
production change — recorded here rather than presented as a fix:

```text
$ go test ./internal/cli/ -run 'TestSessionMsgSendHandlerCarriesDataArgument|TestSessionMsgSendHandlerRejectsUnencodableData' -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	2.968s
```

Two tests were added: a structured payload sent through the handler and asserted intact
out of the polled envelope, and an unencodable `data` value (a channel) asserted to yield
a structured tool error rather than a Go error or a silent drop.

**Mutant a shallower test would pass.** An implementation that drops the data part
entirely still returns a `messageId` and still delivers one message, so "send succeeded"
or "poll returned 1 message" passes it. The test walks into the polled envelope's `parts`,
finds the `data` part, and compares field values — so a dropped payload fails. The payload
deliberately carries a nested object and a boolean, because a *stringifying*
implementation (`fmt.Sprint` of the argument) survives a flat all-strings payload; the
assertion requires `passed` to be the boolean `true` and `scores.craft` to be a nested
numeric.

---

## Scope verification

```text
$ git status --porcelain
 M .claude/rules/moai/core/moai-mcp-tools.md
 M .moai/reports/t238/code-fixes.md
 M .moai/reports/t238/cr-response.md
 M .moai/reports/t238/doc-fixes.md
 M .moai/specs/SPEC-CODEX-SESSION-MSG-001/research.md
 M .moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md
 M internal/cli/mcp_session_msg.go
 M internal/cli/mcp_session_msg_test.go
 M internal/config/defaults.go
 M internal/sessionmsg/agent.go
 M internal/sessionmsg/edge_test.go
 M internal/sessionmsg/envelope.go
 M internal/sessionmsg/store.go
 M internal/sessionmsg/store_test.go
 M internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md
?? .moai/reports/t187/codex-support-audit.md
?? .moai/reports/t187/e2e/
?? .moai/reports/t187/pr-body.md
?? internal/cli/mcp_session_msg_data_test.go
?? internal/sessionmsg/round4_test.go
```

This round's changes are the eight `internal/` entries: six modified
(`internal/cli/mcp_session_msg.go`, `internal/cli/mcp_session_msg_test.go`,
`internal/config/defaults.go`, `internal/sessionmsg/{agent,envelope,store}.go`) plus
`internal/sessionmsg/{edge,store}_test.go`, and two new files
(`internal/cli/mcp_session_msg_data_test.go`, `internal/sessionmsg/round4_test.go`).

Everything under `.moai/**` and `.claude/**` in that listing belongs to the parallel lane
and was **not** touched by this round; the sole exception is this report, written to the
assigned deliverable path.

Explicitly not done, per the scope instruction: no pending-mailbox depth cap was added
(previously rejected, still rejected); no on-disk `role` value was changed; nothing outside
`internal/` was edited.

---

## Gaps (what was NOT observed)

- **`go test ./...` was not run.** Forbidden for this round; the machine is shared. Only
  `internal/sessionmsg`, `internal/cli` (selector-scoped to `TestSessionMsg|TestMoaiMCPServer`),
  and `internal/config` were executed. Other packages that import `internal/config` were
  not run, so a consumer broken by the new `defaults.go` symbol would not have been seen —
  though the change is purely additive (a new `var` in an existing block), and
  `go build ./...` covers compilation across the whole module.
- **`internal/cli` was run under a selector, not in full.** Non-matching tests in that
  package were not executed. `newSessionMsgStore()`'s signature change (K5) is compile-time
  visible and `go build ./...` is clean, but any *runtime* interaction in an unselected
  `internal/cli` test was not observed.
- **`golangci-lint` was run on `./internal/sessionmsg/...` only**, per the assigned batch.
  `internal/cli` and `internal/config` were not linted.
- **The `os.Getwd`-failure branch itself is not exercised.** The K5 test covers the guard
  (`sessionMsgStoreRoot("")`), not the path by which `resolveProjectDir()` comes to return
  `""`. Forcing that would require running the test process with a deleted CWD, which is
  not portable and would affect the whole process. The reachability argument in K5 is
  reasoning from `os.Getwd`'s documented failure mode plus this project's worktree
  lifecycle — it is **not** a measured reproduction, and is labelled as such.
- **The K1 tests skip on Windows and as root.** Directory mode bits do not gate file
  creation on Windows, and root bypasses them, so on those platforms the heartbeat-failure
  behaviour is unverified. `GOOS=windows go vet` confirms compilation only, never runtime
  behaviour.
- **No race run of `internal/cli`.** `-race` was run for `internal/sessionmsg` as
  specified; the `internal/cli` selector was run without it.
- **Concurrency under the new sort was not separately stress-tested.** The existing
  concurrency tests in the package passed under `-race`, but no new test exercises many
  concurrent pollers against a deep mailbox with the sort in place.
- **No cross-platform runtime verification.** Only `darwin/arm64` was executed. Windows
  and Linux behaviour rests on CI.

---

## Residual-risk (what could still be wrong despite what was observed)

- **K1 makes a persistent heartbeat failure silent.** This is the accepted cost of option
  (a): an agent whose record is permanently unwritable will report `online: false`
  indefinitely with no error surfaced by `Send` or `Poll`. If that liveness signal is ever
  load-bearing for a caller, the failure will present as "the peer looks dead" with nothing
  in the error path pointing at the cause. A future warning channel on the tool result
  would close this; it was judged out of proportion to the defect here.
- **K2 changes an observable ordering that nothing in the SPEC pinned.** Any consumer that
  had (accidentally) come to depend on the previous arbitrary order would now see different
  behaviour. No such consumer is known, and the previous order was random rather than
  stable, so depending on it was never sound — but the change is behavioural, not merely
  internal.
- **K2's sort is O(n log n) on every poll**, over every pending envelope, all of which are
  already read and unmarshalled by `listEnvelopes`. At the mailbox depths this broker is
  designed for the cost is negligible; at pathological depth the unbounded *read* (which
  predates this change, and which the rejected depth cap would have addressed) dominates
  long before the sort does.
- **K3's ceiling is enforced at validation, on the in-memory envelope.** It bounds what
  `Send` will accept; it does not bound what an already-persisted envelope may contain, so
  a file written by an older binary — or by hand — is still read back without a size check
  in `listEnvelopes`. No sweep or repair path enforces the new bound retroactively.
- **K3 bounds text and data independently, so the worst-case envelope is now the sum of
  both ceilings** (~128 KiB) rather than a single 64 KiB. That is a deliberate consequence
  of matching the existing text-accumulation shape; if REQ-CSM-005's "body size" was meant
  as a joint total, the check needs to move to `totalText + totalData` and the text bound's
  existing meaning changes with it.
- **K5's guard covers the empty-string case only.** A `$CLAUDE_PROJECT_DIR` holding a
  relative path, a non-existent path, or a path outside the project still resolves without
  complaint — shared `resolveProjectDir` behaviour, deliberately left alone. The broker
  would operate on a wrong-but-resolvable root in those cases exactly as it did before.
- **The K1 fault injection is coarse.** Making `agents/` read-only fails the heartbeat via
  `CreateTemp`, which is one specific failure mode. A heartbeat that failed later (during
  `atomicfile.Replace`, say) is not covered, though it reaches the same `_ =` discard.
- **`internal/config/defaults.go` values are `var`, not `const`**, and the new one is no
  exception. Tests that shorten it must restore it — the K3 test does so with `defer`, and
  K6 hardened the neighbouring text-ceiling override for the same reason, but the pattern
  remains available to be got wrong in future tests.
