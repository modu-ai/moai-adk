auditor-model: claude-opus-5-5

# t2129 sync-audit — RegisterAgent returns a duplicate number to the authority (Class B)

verdict: PASS-WITH-DEBT
audited_sha: 1a1bdc491138d917e66fbb0050e8f97c6e0f116f

Score: 0.90 (Functionality 92 / Security 95 / Craft 85 / Consistency 92; harmonic mean ≈ 0.91, rounded down for the debt carried). Must-pass firewall: Functionality PASS, Security PASS.
Range audited: 04d8c27a7e..1a1bdc4911 (code commit 86ce32665c; 1a1bdc4911 adds only the card-review record). Tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2129, branch WT-pr-codex-repair-5. Iteration: 1 (first sync audit of this card).

## Evaluation Report

SPEC: none (Class B card t2129, PR #1969 Codex P2 on internal/agenticfactory/claim/identity.go:114)
Overall: PASS-WITH-DEBT — the card's requirement (return the duplicate number through a conditional delete on its revision in ControlRef, reproduced first on a shared ControlRef) is met and witnessed RED→GREEN; two optional in-diff findings about the new read-then-delete window and one surviving mutant are carried as debt.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | RED on base via overlay: `identity_duplicate_number_t2129_test.go:48: authority keys after the race = [agent/1 agent/2], want only agent/1 (the duplicate's number leaked)` / `FAIL`; HEAD: `--- PASS: TestRegisterAgentReturnsADuplicateNumberToTheAuthority (0.00s)`; whole package: `ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	64.842s` |
| Security (25%) | 95/100 | PASS | secret grep over the two changed Go files: no match (`secret_grep_exit=1`); no input surface added; the only new trust-relevant behaviour is F1 (deletes an authority key without re-checking its owner) — optional, see Findings |
| Craft (20%) | 85/100 | PASS | `coverage: 86.0% of statements`; `identity.go:126: giveBackAgent 72.7%`; `golangci-lint run ./internal/agenticfactory/claim/` → `0 issues.`; mutant M3 (cache removal dropped) survives the whole package |
| Consistency (15%) | 92/100 | PASS | `gofmt -l` empty (`gofmt_exit=0`); `go vet` silent (`vet_exit=0`); giveBackAgent mirrors releaseAgent's authority-then-cache order, `s.announce()` before the write, `refErr` wrapping |

### Findings (structured defect-list)

- F1 [Low-Medium, confidence high] [optional] [in-diff] internal/agenticfactory/claim/identity.go:128-134 — giveBackAgent reads `agent/<n>` and deletes it on the revision read, but never checks that the record it read is still this session's number. If `agent/<n>` is released and re-taken by another session between this call's link and its give-back read, the give-back deletes the other session's live registration. Reproduced with an artificial interleaving (probe P2, Evidence §E5): HEAD ends with `keys=[agent/1] agent/2= rev=0 readErr=agenticfactory: no claim held for this issue` after session 9 had taken agent/2; the base left session 9's record in place. The same window with an ABA revision reset (MemoryRef revisions restart at 1 after a delete; probe Q2) shows the same deletion even if an owner check compared only revisions. Reachability: requires a concurrent `ReleaseAgent(S, n)` naming the duplicate number, which RegisterAgent never hands out to S's callers; no non-test caller of `Attach`, `RegisterAgent` or `ReleaseAgent` exists in the tree today (Evidence §E8). Required fix (follow-up card): in giveBackAgent, decode the record read and skip the delete unless it equals `core.Agent{SessionID: sessionID, Number: n}` (pass the session id in); the ABA residue is a property of the revision scheme, shared with releaseAgent, and belongs with the ControlRef revision contract.
- F2 [Low, confidence high] [optional] [in-diff] internal/agenticfactory/claim/identity.go:134 — the `ErrNotHeld` tolerance on Delete never fires with the only ControlRef in the tree: `MemoryRef.Delete` on a key that vanished between the read and the delete returns `ErrRevisionConflict` (`agent/2 is at revision 0, not 1`), so RegisterAgent now returns an error where the base succeeded, although nothing is left to give back (probe Q1, Evidence §E6). The caller's session is in fact registered (a retry returns the low number via agentOf). The conflict case the progress record names as a Gap is real and surfaces as this error. Required fix (follow-up): on `ErrRevisionConflict`, re-read the key and treat `ErrNotHeld` (or a record that is no longer ours, per F1) as given back.
- F3 [Low, confidence high] [optional] [in-diff] internal/agenticfactory/claim/identity.go:139 — mutant M3 (cache-file removal replaced by a no-op) survives the card test and the whole package (Evidence §E7), so the cache half of the give-back is untested. With an authority attached the stale `agent-<n>.json` is not read (agents() reads the authority), so the impact is limited to a later detach. Required fix (optional): assert `os.Stat(racer.agentPath(2))` is not-exist in the card test.
- F4 [Info] [optional] [in-diff] internal/agenticfactory/claim/identity.go:139-141 — no-authority path behaviour change: a failing `os.Remove` (other than not-exist) is now returned instead of swallowed. Judged acceptable: the session is registered under the low number and a retry returns it; the previous silent swallow leaked the number on disk with no signal.
- F5 [Info] [optional] [outside] internal/agenticfactory/claim/controlref.go:109-190 — Rebuild only writes cache entries for keys present in the authority and never removes cache `agent-<n>.json` files for numbers the authority no longer holds. Not a cache-only give-back, and outside this card; listed for the leader only.
- Cross-check of the other agent paths for the cache-only-removal pattern: `grep -n 'os.Remove(s.agentPath'` finds only identity.go:139 (giveBackAgent) and identity.go:290 (releaseAgent, after the authority Delete). releaseAgent deletes on the revision of the record whose session it just checked — no cache-only removal. ResumeAgent and verifyAgent are read-only. No further instance of the pattern found.

### Recommendations

- Open one follow-up card for F1+F2 together (owner check before the delete, re-read on conflict); both are one-function changes in giveBackAgent with a probe-style test already sketched in this audit.
- Add the cache-file assertion (F3) in the same follow-up.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. The card test fails on the base (04d8c27a7e identity.go) with the leaked-key message and passes at HEAD.
2. The whole claim package passes at HEAD; gofmt, go vet and golangci-lint are clean.
3. The fix deletes the duplicate on its revision in the authority; mutants that skip the delete or use a wrong revision are killed by the card test; a mutant dropping the cache removal is not.
4. The read-then-delete window has two adverse interleavings (F1, F2), both requiring a concurrent ReleaseAgent of the duplicate number.
5. `agentFile` with an attached ref maps a missing key to `os.ErrNotExist`.
6. The progress record's Gaps are honest: it states the race is simulated deterministically and not reproduced concurrently, and names the revision-conflict window as unexercised.

### Evidence

E1 — RED on base (overlay replaces identity.go with `git show 04d8c27a7e:internal/agenticfactory/claim/identity.go`, tree untouched):
```
$ go test -count=1 -overlay .../scratchpad/overlay_base.json -run 'TestRegisterAgentReturnsADuplicateNumberToTheAuthority' ./internal/agenticfactory/claim/
--- FAIL: TestRegisterAgentReturnsADuplicateNumberToTheAuthority (0.02s)
    identity_duplicate_number_t2129_test.go:48: authority keys after the race = [agent/1 agent/2], want only agent/1 (the duplicate's number leaked)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.437s
FAIL
```
E2 — GREEN at HEAD:
```
$ go test -count=1 -run 'TestRegisterAgentReturnsADuplicateNumberToTheAuthority' -v ./internal/agenticfactory/claim/
=== RUN   TestRegisterAgentReturnsADuplicateNumberToTheAuthority
--- PASS: TestRegisterAgentReturnsADuplicateNumberToTheAuthority (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.224s
```
E3 — whole package, format, vet, lint, coverage at HEAD:
```
$ go test -count=1 -timeout 150s ./internal/agenticfactory/claim/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	64.842s
exit=0
$ gofmt -l internal/agenticfactory/claim/; echo "gofmt_exit=$?"; go vet ./internal/agenticfactory/claim/; echo "vet_exit=$?"
gofmt_exit=0
vet_exit=0
$ golangci-lint run ./internal/agenticfactory/claim/
0 issues.
$ go test -count=1 -timeout 150s -coverprofile=.../cover.out ./internal/agenticfactory/claim/ ; go tool cover -func=...
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	56.271s	coverage: 86.0% of statements
github.com/modu-ai/moai-adk/internal/agenticfactory/claim/identity.go:87:	RegisterAgent		77.3%
github.com/modu-ai/moai-adk/internal/agenticfactory/claim/identity.go:126:	giveBackAgent		72.7%
github.com/modu-ai/moai-adk/internal/agenticfactory/claim/identity.go:264:	releaseAgent		80.0%
total:										(statements)		86.0%
```
E4 — mutants (overlay of identity.go; card test only unless stated):
```
M1 skip delete (line 134 → `_ = rev; if err := error(nil); err != nil {`)
    identity_duplicate_number_t2129_test.go:48: authority keys after the race = [agent/1 agent/2], want only agent/1 (the duplicate's number leaked)
FAIL   → killed
M2 wrong revision (rev → rev+1)
    identity_duplicate_number_t2129_test.go:38: agenticfactory: the authoritative copy moved on, write refused: agent/2 is at revision 1, not 2
FAIL   → killed
M4 wrong revision + error swallowed (`rev+1); err != nil && false`)
    identity_duplicate_number_t2129_test.go:48: authority keys after the race = [agent/1 agent/2], want only agent/1 (the duplicate's number leaked)
FAIL   → killed
M3 cache removal dropped (line 139 → `if err := error(nil); err != nil && ...`)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.185s   → survives card test
```
E5 — probe P2 (hook fires on the first read of agent/2 after this call's own link, i.e. before giveBackAgent's read; it releases agent/2 and lets session 9 take it):
```
HEAD: zz_audit_probe_test.go:71: P2 retaken-between: got={SessionID:11111111-2222-3333-4444-000000000001 Number:1} err=<nil> keys=[agent/1] agent/2= rev=0 readErr=agenticfactory: no claim held for this issue
base: zz_audit_probe_test.go:71: P2 retaken-between: got={SessionID:11111111-2222-3333-4444-000000000001 Number:1} err=<nil> keys=[agent/1 agent/2] agent/2={"session_id":"11111111-2222-3333-4444-000000000009","number":2} rev=1 readErr=<nil>
```
E6 — probes Q1–Q3 at HEAD (hook fires on the second read of agent/2 after the link — exactly between giveBackAgent's read and its delete):
```
Q1 deleted-between-read-and-delete: got={SessionID: Number:0} err=agenticfactory: the authoritative copy moved on, write refused: agent/2 is at revision 0, not 1 keys=[agent/1]
Q2 retaken-between-read-and-delete: got={SessionID:11111111-2222-3333-4444-000000000001 Number:1} err=<nil> keys=[agent/1] agent/2= rev=0 readErr=agenticfactory: no claim held for this issue
Q3 bumped-between-read-and-delete: got={SessionID: Number:0} err=agenticfactory: the authoritative copy moved on, write refused: agent/2 is at revision 2, not 1 keys=[agent/1 agent/2]
```
(Q3's Bump has no production analogue: the only CompareAndSwap on `agent/` keys is linkAgent's create on revision 0.) Probe P5 at HEAD: `P5 agentFile(42) on empty ref: rev=0 err=file does not exist isNotExist=true`; probe P4 at HEAD: `P4 cache agent-2.json stat err=stat .../agents/agent-2.json: no such file or directory`.
E7 — M3 against the whole package:
```
$ go test -count=1 -timeout 150s -overlay .../ov_m3.json ./internal/agenticfactory/claim/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	46.803s
```
E8 — production reachability:
```
$ grep -rn "\.Attach(" internal/ cmd/ | grep -v _test.go
internal/lsp/core/client.go:254:	c.router.Attach(c.tr)
$ grep -rn "RegisterAgent(\|ReleaseAgent(" internal/ cmd/ | grep -v _test.go | grep -v "func (s \*Store)"
(no output)
```
The only ControlRef implementation in the tree is MemoryRef (`grep -rn "func (.*) Delete(" internal/agenticfactory/ | grep -v _test` → `controlref.go:236: func (m *MemoryRef) Delete`).
E9 — audit plan and codex leg:
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2129
"config_status": "ok", "model": "codex", backends: claude gate "off", codex gate "required", glm gate "off", "cross_model_active": true, "enforced_required": ["codex"]
mcp__moai__codex_audit(project_root=<this tree>, target=baseBranch, delta_base=04d8c27a7ed77afc09217f8623519806d48df3c4, mode=native)
→ "verdict":"pass", "findings":[], "next_steps":[], "resolved_model":"gpt-6.1-sol", "audit_receipt":"rcpt-2d27668354aedef9ffa05a5b", "review_base":"delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)", "transport":"codex app-server"; summary states a static read only, no tests run.
```
E10 — tree left clean:
```
$ git status --short; echo "status_exit=$?"; git rev-parse HEAD
status_exit=0
1a1bdc491138d917e66fbb0050e8f97c6e0f116f
```
(Before this verdict file was written.)

### Baseline-attribution

All Go measurements were taken in this run against the card worktree at HEAD 1a1bdc4911 (E10) with the toolchain on PATH; the base comparison replaces only identity.go with its 04d8c27a7e content through `-overlay`, so the test file and every other file are HEAD's. Probe and mutant sources live in the session scratchpad and entered the build only through `-overlay`. The audit-plan read and the codex leg ran through the installed moai build at commit 04d8c27a7e; the codex result reports `build_lag: ... built from commit 04d8c27a7, an ancestor of this tree's HEAD 1a1bdc491` — the lag concerns the moai tooling, which carries none of this card's change; the codex reviewer read the delta diff itself.

### Gaps

- Refused commands (worktree-isolation guard), verbatim, each replaced by a split command without shell variables: (1) `...this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it...` on the combined `git show ... && go test -overlay` line; (2) `...runs go with a value computed at runtime (the variable S)...` ; (3) `...runs sed with a value computed at runtime (the variable SRC)...` ; (4) `...runs go with a value computed at runtime (the variable f)...`. The replacements ran the same measurements.
- The race is simulated deterministically (stale first `Keys("agent/")` listing; read hooks for the probes); no real concurrent interleaving, no `-race` run, no other OS.
- The audit plan prescribes `audit_multi` for a cross-model plan; per the lead's instruction this audit ran a single `codex_audit` (native, delta base) instead and no `audit_multi`, so no convergence digest exists for the orchestrator's `--result-file` check. The codex gate is the only required backend in the plan and its leg returned pass with receipt rcpt-2d27668354aedef9ffa05a5b. The MCP tool's own transport is reported as `codex app-server`; no app-server was driven directly.
- The codex leg was a static read only (its own statement).
- Coverage was measured for this package only; F1/F2 have no committed test (the probes were overlay-only).

### Residual-risk

- F1: under a concurrent `ReleaseAgent(S, n)` that names the duplicate number, the give-back can delete another session's live number — a new failure mode the base did not have (the base leaked instead). Today no non-test code attaches an authority or registers agents, so the window has no production caller; it becomes live the moment a real ControlRef is attached.
- F2: a spurious RegisterAgent error in the same window; recoverable by retry.
- A future ControlRef implementation whose Delete returns a different error for a vanished key would change F2's shape; the interface comment does not specify it.
- Probe and mutant outputs live in the machine-local scratchpad; the deciding lines are quoted above, the rest is a known loss.

## Iteration history

- Iteration 1 (this file): PASS-WITH-DEBT 0.90 on 1a1bdc4911; debt F1, F2, F3 (optional, follow-up card recommended).

AUDIT-VERDICT receipts: rcpt-2d27668354aedef9ffa05a5b
