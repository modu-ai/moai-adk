# t238 — CodeRabbit code findings C1-C9 (PR #1606, SPEC-CODEX-SESSION-MSG-001)

> **Redaction notice (pre-merge).** Command output in this report is verbatim except for
> three workstation-specific values, replaced with placeholders so the committed evidence
> discloses no developer or host identity: the absolute worktree path → `<repo>/.claude/worktrees/t187`,
> the process id → `<pid>`, and the hostname → `<host>`. Every observed VALUE — SHAs,
> timestamps, counts, line numbers, and exit codes — is as measured.
> Scope note: this notice covers the workstation-value substitutions only. Some quoted
> output additionally elides a long path segment for width, shown as `...` (per-run
> `t.TempDir()` roots such as `/var/folders/.../001/victim.json`, and long subtest paths
> such as `--- PASS: .../user_role`). Those are elisions of a PATH, never of a result:
> no PASS/FAIL verdict, count, or error string was shortened.


**Baseline-attribution (global).** All measurements were taken in the worktree
`<repo>/.claude/worktrees/t187`, branch
`WT-codex-session-msg`, base commit `f33cd05649f27f6ba0c44db95505c3e303283a52`.
Confirmed at session start:

```text
$ pwd && git rev-parse --show-toplevel && git rev-parse HEAD && git branch --show-current
<repo>/.claude/worktrees/t187
<repo>/.claude/worktrees/t187
f33cd05649f27f6ba0c44db95505c3e303283a52
WT-codex-session-msg
```

The tree is left **uncommitted**. No branch was created, nothing was pushed.

## Files changed

Modified:

- `internal/sessionmsg/store.go` — C1 (validation at Send/Poll entry), C1 comment
- `internal/sessionmsg/lock_unix.go` — C2 (fd sentinel)
- `internal/sessionmsg/lock.go` — C3 (Abs fallback), C4 (release-error propagation)
- `internal/sessionmsg/lock_windows.go` — C4 (handle retained on CloseHandle failure)
- `internal/sessionmsg/envelope.go` — C5 (data part rejects text payload)
- `internal/sessionmsg/envelope_test.go` — C5 reciprocal table case, C9 serialized-role assertions
- `internal/sessionmsg/store_test.go` — C6 poll batch ceiling test
- `internal/sessionmsg/edge_test.go` — fixture correction (see C1 note)
- `internal/cli/mcp_session_msg_test.go` — C1 tool-surface boundary test, C7 required-schema assertions
- `internal/web/assets/i18n.js` — C8 (English locale only)

Added:

- `internal/sessionmsg/ids.go` — shared id-shape validation
- `internal/sessionmsg/ids_test.go` — C1 core boundary tests
- `internal/sessionmsg/lock_unix_fd_test.go` — C2 regression guard

## Verification (mandated batch)

```text
$ go build ./...
exit=0

$ go vet ./internal/sessionmsg/ ./internal/cli/
exit=0

$ GOOS=windows go vet ./internal/sessionmsg/
exit=0

$ go test ./internal/sessionmsg/ -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	0.590s
exit=0

$ go test ./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer' -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	1.618s
exit=0
```

Re-run in full after the C9 addendum landed, on the final tree:

```text
$ gofmt -l internal/sessionmsg/ internal/cli/mcp_session_msg_test.go
(no output)

$ go build ./...
exit=0

$ go vet ./internal/sessionmsg/ ./internal/cli/
exit=0

$ GOOS=windows go vet ./internal/sessionmsg/
exit=0

$ go test ./internal/sessionmsg/ -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	0.561s
exit=0

$ go test ./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer' -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	1.171s
exit=0
```

Additional (not mandated, cheap, and relevant because the lock implementation changed):

```text
$ go test ./internal/sessionmsg/ -race -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	1.827s
exit=0

$ gofmt -l internal/sessionmsg/ internal/cli/mcp_session_msg_test.go internal/cli/mcp_server.go internal/cli/mcp_session_msg.go
(no output)
```

`go test ./...` was NOT run (repo rule; other lanes share this machine).

---

## C1 — externally supplied ids joined into filesystem paths — FIXED

**Claim.** Caller-supplied `agentId` and `ack_ids` entries are now shape-checked
before any path is built, at both public entry points and therefore at the MCP
tool surface. A malformed id deletes nothing and returns a structured error.

**Fix.** New `internal/sessionmsg/ids.go` holds two package-level compiled
patterns and their helpers:

- `validAgentID` — `^(?:claude|codex)-[0-9a-f]{8}$`, matching the shape
  `newAgentID` mints (`agent.go:63`, `fmt.Sprintf("%s-%x", kind, b)` over 4
  random bytes → 8 lowercase hex). Read, not guessed.
- `validMessageID` — `^msg-[0-9a-f]{16}$`, matching `newMessageID`
  (`store.go`, 8 random bytes → 16 lowercase hex).

`Send` validates `from_agent_id` and `to_agent_id` before `readAgent`; `Poll`
validates `agent_id` and every `ack_ids` entry before `readAgent`. Rejection is
an `*InvalidIDError` carrying the offending field, id class, and value.

**Choice: reject the whole call, not skip the bad entry** — as requested, stated
explicitly. Two reasons. First, silently skipping returns an `ackedCount` the
caller cannot reconcile against what it sent, which hides a caller bug behind a
success result. Second, a traversal id is not a typo — it is either a broken
caller or an attack, and both deserve a signal. Validation runs before any
mutation, so a rejected call leaves the mailbox untouched (no partial effect on
a batch where only one entry is bad).

**MCP surface.** No change was needed in `internal/cli/mcp_session_msg.go`:
`handleSessionMsgPoll` and `handleSessionMsgSend` already route errors through
`sessionMsgToolErr`, which falls back to `toolErr` for any non-`UnknownAgentError`
— producing `IsError: true`. Verified by the new tool-surface test below rather
than by reading alone.

The stale comment at `store.go:94` now says the shape is ENFORCED at the public
entry points, and points at `ids.go`.

**Evidence — RED on pre-fix code.** The three validation call blocks were
temporarily stripped from `store.go` and the suite re-run:

```text
$ go test ./internal/sessionmsg/ -count=1 -run 'TestPollRejectsTraversalAckIDs|TestPollRejectsMalformedAgentID|TestSendRejectsMalformedAgentIDs'
--- FAIL: TestPollRejectsTraversalAckIDs (0.00s)
    ids_test.go:43: Poll accepted traversal ack id "../../../../victim" (result {Messages:[] Remaining:0 ExpiredCount:0 AckedCount:1})
--- FAIL: TestPollRejectsMalformedAgentID (0.00s)
    ids_test.go:91: Poll("../../etc/passwd") rejected with *sessionmsg.UnknownAgentError (...), want *InvalidIDError — the id reached path construction
    [... 6 more ids, same shape ...]
--- FAIL: TestSendRejectsMalformedAgentIDs (0.00s)
    ids_test.go:131: traversal from_agent_id: rejected with *sessionmsg.UnknownAgentError (...), want *InvalidIDError
    ids_test.go:131: traversal to_agent_id: rejected with *sessionmsg.UnknownAgentError (...), want *InvalidIDError
FAIL
```

At the tool surface, same strip:

```text
$ go test ./internal/cli/ -run 'TestSessionMsgPollHandlerRejectsTraversalIDs' -count=1
--- FAIL: TestSessionMsgPollHandlerRejectsTraversalIDs (0.00s)
    mcp_session_msg_test.go:282: traversal ack_ids entry "../../../../../../victim" accepted: map[ackedCount:1 expiredCount:0 messages:[] remaining:0]
    mcp_session_msg_test.go:285: file outside the broker state root was deleted via ack_ids: stat /var/folders/.../001/victim.json: no such file or directory
FAIL
```

That reproduces the reported CRITICAL exactly: `err=nil`, `ackedCount=1`, victim
deleted outside the state root.

**Mutants named.**

- `TestPollRejectsTraversalAckIDs` — a version asserting only `err != nil` is
  passed by an implementation that validates *after* the delete loop: it errors,
  but the victim is already gone. Both the error type AND the victim's survival
  are asserted, so the ordering is pinned.
- `TestPollRejectsMalformedAgentID` / `TestSendRejectsMalformedAgentIDs` — this
  one was caught **during** authoring, not theorized. My first draft asserted
  only `err != nil` and it **PASSED on the broken code**, because an
  unregistered traversal path already produced `UnknownAgentError` from
  `readAgent`. The tests were strengthened to `errors.As(err, &*InvalidIDError)`,
  which is what distinguishes "rejected before touching the filesystem" from
  "the traversal happened and found nothing". The RED output above is from the
  strengthened version.
- `TestSendRejectsMalformedAgentIDs` also pins `InvalidIDError.Field` per case,
  so validating only `from_agent_id` (read first) while leaving `to_agent_id` —
  the id that names the *written* mailbox path — unchecked still fails.
- `TestIDShapeHelpers` pins the patterns directly, catching an anchor-less or
  case-insensitive regex regression.

**One pre-existing test fixture was corrected** (`edge_test.go:88`): it acked
`"msg-doesnotexist"` to assert "an unknown ack id is tolerated". That id is not
well-formed, so it is now rejected outright. The fixture became
`"msg-0123456789abcdef"` — a valid `msg-<hex16>` that was simply never minted —
which preserves the test's actual intent. This is a fixture correction, not a
weakening: the assertion (`AckedCount == 0`, no error) is unchanged.

## C2 — fd 0 used as the "unacquired" sentinel — FIXED

**Claim.** `agentLock.fd` is initialized to `-1`, only negative descriptors are
treated as unacquired, and release resets to `-1`. A lock acquired while stdin
is closed is now released.

**Fix.** `internal/sessionmsg/lock_unix.go`: added `const unacquiredFD = -1`,
`newAgentLock` returns `&agentLock{fd: unacquiredFD}`, `release()` guards on
`l.fd < 0` and resets to `unacquiredFD`. The Windows implementation's existing
`windows.InvalidHandle` sentinel is untouched — it was already correct.

**Evidence — RED on pre-fix code.** The three lines were temporarily reverted
(`&agentLock{}`, `if l.fd == 0`, `l.fd = 0`):

```text
$ go test ./internal/sessionmsg/ -count=1 -run TestAgentLockReleasesDescriptorZero -v
=== RUN   TestAgentLockReleasesDescriptorZero
    lock_unix_fd_test.go:62: second acquire after releasing fd 0 failed — the flock leaked: sessionmsg lock flock /var/folders/.../fdzero.lock: resource temporarily unavailable
--- FAIL: TestAgentLockReleasesDescriptorZero (0.00s)
FAIL
```

Same symptom as the orchestrator's probe. Post-fix the test PASSES and does not
skip:

```text
$ go test ./internal/sessionmsg/ -count=1 -run TestAgentLockReleasesDescriptorZero -v
=== RUN   TestAgentLockReleasesDescriptorZero
--- PASS: TestAgentLockReleasesDescriptorZero (0.00s)
```

**Test shape.** `internal/sessionmsg/lock_unix_fd_test.go`, `//go:build !windows`,
NOT parallel (it mutates process-global descriptor 0). It dups stdin, restores
it via `t.Cleanup` + `dup2`, and `t.Skip`s if the acquired descriptor is not 0 —
staying honest on a machine where the condition cannot be reproduced.

**Mutant named.** A version asserting only `release() == nil` passes on the
broken code, because the broken `release()` returns `nil` *precisely by doing
nothing*. The load-bearing assertion is that a SECOND `acquire()` on the same
path succeeds.

## C3 — `filepath.Abs` error discarded — FIXED

**Claim.** A failed `Abs` no longer produces an empty map key.

**Fix.** `lock.go`: `absLockPath, absErr := filepath.Abs(lockPath)`; on error,
fall back to `lockPath` as given. Previously the discarded error left
`absLockPath == ""`, which would collapse every distinct lock in the process
onto one shared mutex for the process lifetime — silently, and permanently
(the bad entry stays in the package-global `sync.Map`).

**Evidence.** No dedicated test. `filepath.Abs` fails only when `os.Getwd()`
fails (deleted or unreadable CWD), which cannot be provoked from inside a Go
test without destabilizing the whole process — and the fallback is a
two-line total-function change with no branch that a test could distinguish
short of that. This is recorded as a **Gap** below rather than claimed as
covered.

## C4 — Windows handle lost on CloseHandle failure — FIXED (compile-verified only)

**Claim (code).** `release()` invalidates `l.handle` only when `CloseHandle`
succeeds, and returns the error otherwise. `withAgentLock` propagates a release
error through a named return without overwriting a non-nil error from `fn()`.

**Fix.**

- `lock_windows.go`: `if closeErr == nil { l.handle = windows.InvalidHandle }`.
  Per MS docs the handle stays OPEN on failure, so unconditionally invalidating
  the field dropped the only reference to it.
- `lock.go`: `func (s *Store) withAgentLock(...) (err error)` and

  ```go
  defer func() {
      if relErr := lock.release(); relErr != nil && err == nil {
          err = fmt.Errorf("sessionmsg: release lock %s: %w", lockPath, relErr)
      }
  }()
  ```

  The `err == nil` guard is the load-bearing part: `fn()`'s result is the
  caller's real outcome and must never be masked by a release failure.

**Evidence.**

```text
$ GOOS=windows go vet ./internal/sessionmsg/
exit=0
```

**NOT OBSERVED — stated plainly.** This lane runs on darwin. The Windows
**runtime** behaviour of this change was NOT executed and NOT observed. Only
cross-compilation and vet were verified. I did not add a Windows-only test,
because one I cannot run would be an unverified artifact; the propagation half
of C4 (`lock.go`) is platform-neutral and is exercised by the existing suite on
darwin.

## C5 — `Message.Validate` payload asymmetry — FIXED

**Claim.** A data part carrying a non-empty `Text` is now rejected, mirroring
the existing text-part check.

**Fix.** `envelope.go`, in `case PartKindData`: reject `p.Text != ""` with
`"part[%d] kind %q must not carry a text payload"`.

**Evidence — RED on pre-fix code.** The check was temporarily removed:

```text
$ go test ./internal/sessionmsg/ -count=1 -run 'TestEnvelopeA2AAlignment/validation_rejects_invalid_messages/data_part_carrying_text_payload' -v
    envelope_test.go:256: expected validation error containing "must not carry a text payload", got nil
--- FAIL: TestEnvelopeA2AAlignment/validation_rejects_invalid_messages/data_part_carrying_text_payload
FAIL
```

**Mutant named.** The obvious weak version sets `wantSub: "text"` (mirroring
the sibling text-part case). That substring **also matches the text-part
error**, so a table case wired to the wrong branch would still pass. The case
pins `wantSub: "must not carry a text payload"`, which only the data-part
branch produces.

## C6 — poll batch ceiling uncovered — FIXED (new coverage)

**Claim.** `config.DefaultSessionMsgPollBatch` (16) is now pinned as the
per-poll claim ceiling from both sides.

**Test.** `TestPollBatchCeiling` in `store_test.go` sends `batch+3` messages to
one receiver and asserts: first poll claims **exactly** `batch` with
`Remaining == 3`; second poll claims **exactly** the 3-message overflow with
`Remaining == 0`; the union across both polls is `batch+3` distinct ids, none
delivered twice, none unknown.

**Evidence — mutation check.** This is new coverage of already-correct
behaviour, so there is no pre-fix RED; instead the ceiling was mutated to 1000:

```text
$ go test ./internal/sessionmsg/ -count=1 -run TestPollBatchCeiling
--- FAIL: TestPollBatchCeiling (0.02s)
    store_test.go:432: first poll claimed 19 messages, want exactly the ceiling 16
    store_test.go:435: first poll Remaining = 0, want 3
    store_test.go:445: second poll claimed 0 messages, want the 3-message overflow
FAIL
```

**Mutant named.** An implementation that claims everything but *reports*
`Remaining` correctly is passed by any assertion made only on
`len(res.Messages) <= batch`; one that claims correctly but reports `Remaining`
as 0 is passed by any assertion made only on the claimed count. Both are pinned
exactly, on both polls, plus the no-duplicate-delivery invariant.

## C7 — schema test asserted properties but not `required` — FIXED

**Claim.** `kind`, `name`, `from_agent_id`, `to_agent_id`, `text`, and
`agent_id` are now asserted to appear in each tool schema's `required` list;
`ack_ids` is asserted NOT required.

**Fix.** `mcp_session_msg_test.go`: a `wantRequired` map alongside the existing
`wantArgs`, checked against `tool.InputSchema.Required`.

**Evidence — mutation check.** `mcp.Required()` was dropped from `agent_id` in
`internal/cli/mcp_server.go`:

```text
$ go test ./internal/cli/ -run 'TestSessionMsgToolsRegisteredWithHintsAndDiscipline' -count=1
--- FAIL: TestSessionMsgToolsRegisteredWithHintsAndDiscipline (0.00s)
    mcp_session_msg_test.go:371: session_msg_poll inputSchema does not mark "agent_id" required (got required=[])
FAIL
```

`mcp_server.go` was restored from a backup afterwards and is confirmed
unmodified (`git diff --stat` lists it nowhere).

**Mutant named.** The pre-existing assertion (Properties only) is passed by
exactly this mutant — the property still appears in the schema, so the
caller-contract regression is invisible. The negative assertion on `ack_ids`
catches the opposite mutant: someone "fixing" the test by marking everything
required.

## C8 — i18n grammar — FIXED (English only)

**Claim.** `internal/web/assets/i18n.js:627` reads
`"Claim the messages waiting in this agent's inbox."`

**Other locales checked, not touched.** All three parallel strings were read and
are already grammatical possessives in their own languages:

```text
$ grep -n "session_msg_poll.enabled.desc" internal/web/assets/i18n.js
627:    ... "Claim the messages waiting in this agent's inbox.",
818:    ... "이 에이전트의 받은 편지함에 쌓인 메시지를 가져옵니다.",
1447:   ... "このエージェントの受信箱で待機中のメッセージを取得します。",
2076:   ... "领取此代理收件箱中待处理的消息。",
```

No template mirror exists (`find internal/template/templates -name i18n.js`
returns nothing), so the Template-First obligation does not apply here.

## C9 — serialized role value unpinned — FIXED (new coverage)

**Claim.** `TestEnvelopeA2AAlignment` now pins the on-disk role value for BOTH
roles: an envelope marshalled with `RoleAgent` carries exactly `"role":"agent"`
and one with `RoleUser` carries exactly `"role":"user"`. The deliberate
divergence from A2A v1 ProtoJSON — which renders an enum as its NAME, so a
conformant encoder would emit `ROLE_AGENT` / `ROLE_USER` — is asserted as an
explicit negative.

**Test.** New subtest `role serializes as the lowercase short form` in
`envelope_test.go`, table-driven over both roles. Each case makes three
assertions:

1. Byte-level `strings.Contains(data, `"role":"agent"`)` — the exact on-disk
   pair, not a decoded equivalent.
2. Decoded `msg["role"] == want` — pins the value itself.
3. The ProtoJSON enum name (`"ROLE_" + strings.ToUpper(want)`) must NOT appear
   anywhere in the serialized envelope.

**Scope.** Test half only, as instructed. `design.md` was NOT edited (separate
lane owns the documentation half) and the `Role` constants were NOT changed —
verified: `git diff internal/sessionmsg/envelope.go` shows only the C5
validation hunk; lines 27-28 still read `RoleUser Role = "user"` /
`RoleAgent Role = "agent"`.

**Evidence — GREEN, and no skip:**

```text
$ go test ./internal/sessionmsg/ -count=1 -timeout 300s -run 'TestEnvelopeA2AAlignment/role_serializes' -v
=== RUN   TestEnvelopeA2AAlignment/role_serializes_as_the_lowercase_short_form
=== RUN   TestEnvelopeA2AAlignment/role_serializes_as_the_lowercase_short_form/agent_role
=== RUN   TestEnvelopeA2AAlignment/role_serializes_as_the_lowercase_short_form/user_role
--- PASS: TestEnvelopeA2AAlignment (0.00s)
    --- PASS: .../role_serializes_as_the_lowercase_short_form (0.00s)
        --- PASS: .../agent_role (0.00s)
        --- PASS: .../user_role (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	0.525s
```

**Evidence — mutation checks.** This is new coverage of already-correct
behaviour, so there is no pre-fix RED. TWO mutants were executed instead, and
the constants restored from a backup after each.

Mutant 1 — the constants flipped to the A2A ProtoJSON enum names (the exact
change CodeRabbit's design.md:58 finding describes as the conformant
alternative):

```text
$ go test ./internal/sessionmsg/ -count=1 -run 'TestEnvelopeA2AAlignment/role_serializes'
--- FAIL: .../agent_role (0.00s)
    envelope_test.go:120: envelope JSON does not carry "role":"agent": {"message":{"messageId":"msg-role","role":"ROLE_AGENT",...}}
    envelope_test.go:134: serialized role = "ROLE_AGENT", want "agent" (A2A ProtoJSON would say "ROLE_AGENT" — the divergence is deliberate)
    envelope_test.go:139: envelope JSON carries the A2A ProtoJSON enum name "ROLE_AGENT"; the broker's on-disk contract is the lowercase short form: ...
--- FAIL: .../user_role (0.00s)
    [same three, for ROLE_USER]
FAIL
```

Mutant 2 — the two values swapped (`RoleUser = "agent"`, `RoleAgent = "user"`).
This is the subtler one: both values stay lowercase and both stay inside the
accepted set, so `Message.Validate` still passes every existing test:

```text
$ go test ./internal/sessionmsg/ -count=1 -run 'TestEnvelopeA2AAlignment/role_serializes'
--- FAIL: .../agent_role (0.00s)
    envelope_test.go:120: envelope JSON does not carry "role":"agent": {"message":{...,"role":"user",...}}
    envelope_test.go:134: serialized role = "user", want "agent" ...
--- FAIL: .../user_role (0.00s)
    envelope_test.go:120: envelope JSON does not carry "role":"user": {"message":{...,"role":"agent",...}}
    envelope_test.go:134: serialized role = "agent", want "user" ...
FAIL
```

**Mutants named.**

- A one-role version (asserting only `RoleAgent`, the role `Send` actually
  sets) is passed by mutant 2's swap in one direction and by any edit to
  `RoleUser` alone — `RoleUser` is never written by `Send`, so nothing else in
  the suite would notice. Both roles are asserted for exactly this reason.
- A version asserting only the decoded `msg["role"]` is passed by a JSON *key*
  rename (`role` → something else) combined with a correct value, because the
  decoded lookup would return `nil` on the old key only if the test also
  checked presence. The byte-level `"role":"user"` pair check pins key and
  value together.
- A version asserting only the byte-level substring is passed by a payload that
  happens to contain that substring elsewhere (e.g. inside a text part). The
  decoded check on the message block rules that out. The two assertions are
  complementary, not redundant.
- Asserting the positive value alone would still pass a hypothetical encoder
  that emitted BOTH forms; the explicit `ROLE_*`-absent assertion closes that.

**Baseline re-verification after mutation.** `git diff` on `envelope.go` was
re-read after restoring, confirming only the C5 hunk is present and the two
`Role` constants are untouched. The full batch was then re-run green (see
below).

---

## Scope respected

- No pending-mailbox depth cap was added (`store.go:251` untouched).
- No file under `.moai/specs/**` was touched. The only `.moai/reports/**` write
  is this report, at the path the task mandated.
- The on-disk `role` values (`"user"` / `"agent"`) are unchanged — C9 pins them,
  it does not alter them.
- `design.md` was NOT edited (C9 documentation half belongs to another lane).

## Gaps (NOT observed)

1. **Windows runtime behaviour of C4.** Not executed. `GOOS=windows go vet`
   proves it compiles; it proves nothing about `CloseHandle` failure handling at
   runtime. No Windows test was added.
2. **C3 has no test.** `filepath.Abs` fails only when `os.Getwd()` fails, which
   cannot be provoked in-process without destabilizing the test binary. The fix
   is compile- and review-verified only.
3. **Full test suites not run.** Only `./internal/sessionmsg/` in full and
   `./internal/cli/ -run 'TestSessionMsg|TestMoaiMCPServer'` filtered, per the
   task's explicit instruction. Whether the *unfiltered* `internal/cli` suite
   still passes was NOT observed. CI on the PR head is the authority.
4. **`golangci-lint` not run.** Only `go vet` and `gofmt -l`.
5. **No e2e / cross-process run.** The flock fixes were exercised only
   in-process (plus `-race`); no two-process contention scenario was executed.
6. **C1 coverage is entry-point-scoped.** `Register` mints its own ids and
   `heartbeat`/`readAgent` are internal callers reached only through validated
   entry points; I did not add validation there and did not test those paths for
   traversal.

## Residual risk

1. **C4 changes `withAgentLock`'s error surface.** If `fn()` succeeds but
   `release()` fails, the operation *did* complete on disk yet now returns an
   error — e.g. `Send` would return `("", err)` for a message that was actually
   written. I judge surfacing the failure to be right (a failed release means a
   leaked lock, which is worse than a false-negative return), and it is what the
   review asked for, but it is a behaviour change and a caller that retries on
   error could double-send. On unix `close()` on a valid fd essentially does not
   fail, so the practical exposure is Windows-side and unobserved (Gap 1).
2. **Windows retry path after a failed CloseHandle.** A later `release()` on a
   retained handle calls `UnlockFileEx` again on an already-unlocked range,
   which will itself fail. Retaining the handle is still strictly better than
   leaking it, but the retry is not clean. Unobserved (Gap 1).
3. **C1 pattern rigidity.** `validAgentID` hard-codes the `claude|codex` kind
   set via the `KindClaude`/`KindCodex` constants. Adding a third kind requires
   no code change (the pattern is built from the constants) — but adding a kind
   whose name is not `[a-z]+` would need the pattern revisited. Similarly, any
   future change to id minting (longer hex, different prefix) must update
   `ids.go` in lockstep or every existing on-disk mailbox becomes unaddressable.
4. **Existing on-disk state.** If any mailbox in a live `.moai/state/session-msg`
   tree carries an id that does not match the enforced shape (hand-edited, or
   minted by an older build), `Poll`/`Send` on it now hard-fail. I found no such
   producer in the code, but I did not scan any live state tree.
5. **C9 pins the encoder, not the decoder.** The assertions cover Go →
   JSON. Nothing added here pins the reverse direction: an envelope on disk
   carrying an unrecognized role string still unmarshals into `Role` without
   complaint, and is only caught later by `Message.Validate` if that envelope
   is validated. Reading existing state is out of the assertion's reach.
6. **The A2A divergence is now pinned in the test but still undocumented in
   `design.md`.** Until the documentation lane lands its half, a reader of the
   test learns the divergence is deliberate only from the test's own comment.
7. **`TestAgentLockReleasesDescriptorZero` may skip on a loaded machine.** If
   another goroutine grabs descriptor 0 between the close and the acquire, the
   test skips rather than fails. That is deliberate honesty, but it means the
   guard can go quiet without anyone noticing.
