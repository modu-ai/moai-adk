# t5 — `pre_tool.go` askPatterns does not fire: root-cause investigation

Read-only investigation. No source file was edited.

- Repo: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main`, HEAD `3b9b3bf9959669c4bfc43da313e25bca61f910a2`
- Installed binary: `~/go/bin/moai` = `v3.1.0-rc.2`, commit `a84e48961`, built `2026-08-15T02:38:17Z`
- Date of investigation: 2026-08-15

---

## Claim

**The defect is real and reproducible, and the root cause is NOT in `pre_tool.go`.**

`askPatterns` compiles correctly, matches correctly, and `preToolHandler.Handle` returns
`permissionDecision: "ask"` exactly as designed. The `"ask"` verdict is then **discarded by the
registry's multi-handler merge** in `internal/hook/registry.go`, which pre-seeds the merged output
with the event default `allow` and never copies the `PermissionDecision` field out of a handler's
output. Only `deny` survives, because `deny` short-circuits the merge entirely.

Precise location: `internal/hook/registry.go:97` (pre-seed) + `internal/hook/registry.go:177-336`
(`mergeHandlerOutput`, which copies `SystemMessage`, `Continue`, `StopReason`, `Retry`,
`AdditionalContext`, `UpdatedInput`, `UpdatedToolOutput`, `UpdatedMCPToolOutput`, `SessionTitle`,
`MxTags` — and **not** `PermissionDecision` or `PermissionDecisionReason`).

The blast radius is wider than the card reports: **every** `ask` verdict in the PreToolUse hook is
lost, not just file `askPatterns`. `askBashPatterns` (e.g. `git reset --hard`, `git push --force`,
`prisma migrate reset`) is silently downgraded to `allow` by the identical mechanism.

Binary staleness is **ruled out** — see Baseline-attribution.

---

## Evidence

### E1 — Reproduction of the reported behaviour (verbatim)

```bash
echo '{"hook_event_name":"PreToolUse","session_id":"probe-t5","cwd":"/Users/goos/MoAI/moai-adk-go","tool_name":"Edit","tool_input":{"file_path":"/Users/goos/MoAI/moai-adk-go/.github/workflows/ci.yaml","old_string":"a","new_string":"b"}}' | ~/go/bin/moai hook pre-tool; echo "EXIT=$?"
```

```
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
EXIT=0
```

```bash
echo '{"hook_event_name":"PreToolUse","session_id":"probe-t5b","cwd":"/Users/goos/MoAI/moai-adk-go","tool_name":"Edit","tool_input":{"file_path":"/Users/goos/MoAI/moai-adk-go/package.json","old_string":"a","new_string":"b"}}' | ~/go/bin/moai hook pre-tool; echo "EXIT=$?"
```

```
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
EXIT=0
```

### E2 — Contrast probes: `deny` survives, `ask` does not (verbatim, one batch)

```bash
echo "=== deny probe (.git/config) ==="
echo '{"hook_event_name":"PreToolUse","session_id":"p","cwd":"/Users/goos/MoAI/moai-adk-go","tool_name":"Edit","tool_input":{"file_path":"/Users/goos/MoAI/moai-adk-go/.git/config","old_string":"a","new_string":"b"}}' | ~/go/bin/moai hook pre-tool; echo "EXIT=$?"
echo "=== ask-bash probe (git reset --hard) ==="
echo '{"hook_event_name":"PreToolUse","session_id":"p","cwd":"/Users/goos/MoAI/moai-adk-go","tool_name":"Bash","tool_input":{"command":"git reset --hard HEAD~1"}}' | ~/go/bin/moai hook pre-tool; echo "EXIT=$?"
echo "=== Dockerfile ==="
echo '{"hook_event_name":"PreToolUse","session_id":"p","cwd":"/Users/goos/MoAI/moai-adk-go","tool_name":"Write","tool_input":{"file_path":"/Users/goos/MoAI/moai-adk-go/Dockerfile","content":"FROM x"}}' | ~/go/bin/moai hook pre-tool; echo "EXIT=$?"
```

```
=== deny probe (.git/config) ===
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Protected file: access denied for security reasons"}}
EXIT=0
=== ask-bash probe (git reset --hard) ===
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
EXIT=0
=== Dockerfile ===
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}
EXIT=0
```

This is the discriminating observation. `deny` reaches the wire; `ask` does not — through **both**
the file path (`AskPatterns`) and the Bash path (`AskBashPatterns`). Any hypothesis local to
`checkFileAccess` is falsified by the Bash probe, and any hypothesis about pattern compilation is
falsified by `deny` working through the same `compilePatterns` helper on the same struct.

### E3 — The handler itself DOES return `ask` (isolates the registry as the culprit)

`internal/hook/coverage_boost_test.go:308` (`TestPreToolHandler_Handle_EditTool_AskPattern`) calls
`preToolHandler.Handle` **directly**, bypassing the registry, and asserts `DecisionAsk`.

```bash
go test ./internal/hook/ -run 'TestPreToolHandler_Handle_EditTool_AskPattern' -count=1 -v 2>&1 | tail -8
```

```
=== RUN   TestPreToolHandler_Handle_EditTool_AskPattern
=== PAUSE TestPreToolHandler_Handle_EditTool_AskPattern
=== CONT  TestPreToolHandler_Handle_EditTool_AskPattern
2026/08/15 14:03:53 WARN file access security check tool_name=Edit decision=ask reason="Critical config file: package.json"
--- PASS: TestPreToolHandler_Handle_EditTool_AskPattern (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.651s
```

The handler emits `decision=ask` for the *same* `package.json` input that the CLI answers `allow`
for (E1). The divergence therefore lies strictly between `Handle`'s return and the CLI's stdout —
i.e. in `registry.Dispatch`.

### E4 — The code path that loses the verdict

The CLI dispatches through the registry (`internal/cli/hook.go:312`):

```
internal/cli/hook.go:312:	output, err := deps.HookRegistry.Dispatch(ctx, event, input)
```

`Dispatch` pre-seeds the accumulator with the event default, which for PreToolUse is
`NewSafeDefaultOutput(...)` — an output already carrying `permissionDecision: "allow"`
(`internal/hook/registry.go:307`):

```go
// internal/hook/registry.go:95-97
	// Start with the default output for this event type and accumulate
	// non-blocking fields (e.g. SystemMessage) from each handler.
	merged := r.defaultOutputForEvent(event, input)
```

`deny` short-circuits before any merge (`internal/hook/registry.go:140-148`):

```go
		if output != nil && isBlockDecision(output) {
			reason := getBlockReason(output)
			...
			return output, nil
		}
```

`isBlockDecision` recognises only `deny`/`block` — never `ask` (`internal/hook/registry.go:238-259`).
So an `ask` output falls through to `mergeHandlerOutput(merged, output)`
(`internal/hook/registry.go:164`), where the pre-seeded `merged.HookSpecificOutput` is non-nil, so
the early `merged.HookSpecificOutput = output.HookSpecificOutput` assignment is skipped and the
field-by-field copy begins:

```go
	if merged.HookSpecificOutput == nil {
		merged.HookSpecificOutput = output.HookSpecificOutput
		return
	}

	src := output.HookSpecificOutput
	dst := merged.HookSpecificOutput

	// additionalContext from every hook is kept (official multi-hook semantics).
	if src.AdditionalContext != "" { ... }
	if len(src.UpdatedInput) > 0 { ... }
	if src.UpdatedToolOutput != "" { ... }
	if src.UpdatedMCPToolOutput != "" { ... }
	if src.SessionTitle != "" && dst.SessionTitle == "" { ... }
	if len(src.MxTags) > 0 { ... }
```

`src.PermissionDecision` and `src.PermissionDecisionReason` are never read. The handler's
`"ask"` + its reason are dropped on the floor, and the pre-seeded `"allow"` is what gets marshalled
to stdout — exactly matching E1/E2.

Confirmed by exhaustive grep that only one handler is registered for the event, so this is not a
"second handler overwrote the first" case:

```bash
grep -rn "EventPreToolUse" internal/hook/*.go | grep -v _test
```

```
internal/hook/pre_tool.go:361:// EventType returns EventPreToolUse.
internal/hook/pre_tool.go:363:	return EventPreToolUse
internal/hook/registry.go:305:	case EventPreToolUse:
internal/hook/types.go:24:	// EventPreToolUse is triggered before a tool is executed.
internal/hook/types.go:25:	EventPreToolUse EventType = "PreToolUse"
internal/hook/types.go:140:		EventPreToolUse,
```

### E5 — Ruling out the plausible-but-wrong hypotheses

| Hypothesis | Verdict | Evidence |
|---|---|---|
| `askPatterns` regexes don't match the probe paths | **Falsified** | E3: handler emits `decision=ask` for `package.json` |
| Policy is nil, so `Handle` returns early at `pre_tool.go:374` | **Falsified** | E2 deny probe fires from the *same* policy struct; `internal/cli/deps.go:236-239` builds `DefaultSecurityPolicy()` + `MergeExtraPatterns` and registers it |
| An earlier `allow` in `Handle` short-circuits before the ask branch | **Falsified** | `Handle` has no early `allow` return before `checkFileAccess`; the only pre-checks are blocked-tools, `--no-verify`, quality gate, and `checkBashCommand`, all of which return deny or fall through |
| `NewAskOutput` is broken | **Falsified** | `internal/hook/types.go:466-476` sets `PermissionDecision: DecisionAsk` correctly; E3's passing test consumes it |
| The defect is confined to `AskPatterns` (file paths) | **Falsified — it is broader** | E2's `git reset --hard` probe shows `AskBashPatterns` is downgraded identically |
| Installed binary is stale | **Ruled out** | See Baseline-attribution |

### E6 — Test-coverage divergence (why this shipped green)

Every `ask` assertion in the suite is **handler-level**, calling `Handle` directly and never
crossing `registry.Dispatch`:

```bash
grep -rln "DecisionAsk" internal/hook/*_test.go internal/cli/*_test.go
```

```
internal/hook/agent_model_guard_test.go
internal/hook/coverage_boost_test.go
internal/hook/pre_tool_test.go
internal/hook/pretool_deny_infinite_test.go
```

```bash
grep -rn "Ask" internal/hook/registry_test.go
```

```
(no output)
```

There is **no registry-level or CLI-level test that asserts an `ask` outcome**. The tests pass while
the binary does not ask because the tests exercise the producer in isolation and the defect lives in
the consumer. `deny` has both handler-level and end-to-end coverage (`internal/cli/hook_protocol_fix_test.go`),
which is precisely why the `deny` path stayed correct while `ask` rotted unnoticed.

```bash
go test ./internal/hook/ -run 'Ask|AskPattern|PermissionDecision' -count=1
```

```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.979s
```

### E7 — Age of the defect

```bash
git log --oneline -3 -L 176,182:internal/hook/registry.go
```

The current `mergeHandlerOutput` was introduced by `6a3603274`
(`fix(hook): Go hook 프로토콜 공식 스펙 정합 …`). Its diff shows the **predecessor** code merged only
`AdditionalContext`:

```
-		if output != nil && output.HookSpecificOutput != nil {
-			if merged.HookSpecificOutput == nil {
-				merged.HookSpecificOutput = output.HookSpecificOutput
-			} else if output.HookSpecificOutput.AdditionalContext != "" && merged.HookSpecificOutput.AdditionalContext == "" {
-				merged.HookSpecificOutput.AdditionalContext = output.HookSpecificOutput.AdditionalContext
-			}
```

So the omission **predates** that refactor; `6a3603274` broadened the merge without adding the
permission fields. This is a long-standing latent defect, not a recent regression.

---

## Baseline-attribution

- **Binary staleness — ruled out.** The installed binary reports commit `a84e48961`; git HEAD is
  `3b9b3bf99`. The binary is behind HEAD by some commits, so this had to be checked rather than
  assumed. Command and output:

  ```bash
  git diff --stat a84e48961 3b9b3bf99 -- internal/hook/ internal/config/
  ```

  ```
  (empty — no differences)
  ```

  ```bash
  git log --oneline a84e48961..3b9b3bf99 -- internal/hook/
  ```

  ```
  (empty — no commits)
  ```

  Neither `internal/hook` nor `internal/config` changed between the binary's commit and HEAD, so the
  probed binary's behaviour is attributable to the HEAD source I read. A stale binary is **not** the
  root cause.

- **Working tree.** `git status --short` showed only untracked report/screenshot artefacts
  (`.moai/reports/diagram-*.html`, `.moai/reports/webredesign/`, `.playwright-mcp/`,
  `settings-identity.png`) — no modified tracked source, so the tree matches HEAD for all files read.

- **Version banner.** `~/go/bin/moai version` → `v3.1.0-rc.2  a84e48961  built 2026-08-15T02:38:17Z`,
  `exit=0`.

- **Probe attribution.** All probes in E1/E2 were run in this session against
  `~/go/bin/moai` (the installed binary, not `bin/moai`), with `cwd` set to the project root.

---

## Gaps (what I did NOT verify)

1. **No live Claude Code round-trip.** I probed `moai hook pre-tool` directly via stdin. I did not
   observe Claude Code actually consuming the hook output and failing to display a permission
   dialog. The wire format is the contract, and `allow` on the wire cannot produce a dialog — but
   the end-to-end runtime behaviour is inferred, not observed.
2. **No handler-vs-registry differential test was written.** I isolated the registry using an
   *existing* handler-level test (E3) plus code reading. I did not write a new registry-level test
   asserting `ask`, because the task forbade source edits. The causal chain in E4 is read from code,
   so the exact merge-drop is a **code-reading conclusion corroborated by black-box behaviour** (E1,
   E2, E3), not a directly instrumented observation.
3. **`permissionMode` interaction unprobed.** All probes omitted `permission_mode`.
   `NewSafeDefaultOutput(permissionModeOf(input))` varies its output by mode, and I did not check
   whether `bypassPermissions` / `plan` / `acceptEdits` change the pre-seeded value or the symptom.
4. **Other events unprobed.** I did not check whether the same merge omission damages any other
   event's decision fields (e.g. `PermissionRequest`'s `decision.behavior`, which `mergeHandlerOutput`
   also never copies). The `Decision` field is likewise absent from the merge — **hypothesis**: the
   same class of loss may affect PermissionRequest deny/allow when a default is pre-seeded, but
   `defaultOutputForEvent` returns a bare `&HookOutput{}` for that event, so the nil-`HookSpecificOutput`
   early-assign branch probably saves it. **Unverified.**
5. **`security.yaml` extra patterns unexercised.** `MergeExtraPatterns` appends user-supplied
   `extra_ask_patterns`; I did not verify whether this project has any configured, nor probe one.
   They would be lost by the same mechanism regardless.
6. **Full suite not run.** I ran only the targeted `-run 'Ask|AskPattern|PermissionDecision'` and
   single-test selections, not `go test ./...`.

---

## Residual-risk

- **Security posture is materially weaker than the policy declares, silently.** Twenty-seven file
  patterns (lock files, `package.json`, `tsconfig.json`, `Dockerfile`, `docker-compose`,
  `.github/workflows/*`, `terraform/*.tf`, `k8s/*`, `Jenkinsfile`, …) and ten Bash patterns
  (`git push --force`, `git reset --hard`, `git clean -fd`, `prisma migrate reset`,
  `drizzle-kit push`, npm/yarn/pnpm cache clears, …) are documented as requiring confirmation and in
  fact require none. Under `defaultMode: bypassPermissions` — the mode this project's doctrine
  assumes for the PreToolUse deny to be "the SOLE blocking mechanism" (`pre_tool.go:412-417`) — an
  unattended agent can force-push or hard-reset with no gate at all.
- **The fix is a two-line-shaped change in a hot merge path shared by every hook event.** Copying
  `PermissionDecision` unconditionally would let a later handler's `allow` clobber an earlier
  handler's `ask`, or let the pre-seeded default be overwritten in events where that is wrong. The
  precedence rule (deny > ask > allow) has to be written deliberately, not as a blind field copy.
- **`ask` may be masked again by mode-awareness.** If Claude Code is running under
  `bypassPermissions`, `permission_request.go:50` documents a deliberate decision that returning
  `"ask"` would *override* bypass mode and always show the dialog. Fixing the registry will start
  surfacing dialogs that have been suppressed for a long time; that is the correct behaviour but it
  is a **behavioural change users will notice**, and it may be mistaken for a new bug.
- **Time-of-check window.** My probes and the source read are from HEAD `3b9b3bf99` at
  2026-08-15. Other sessions are active on this checkout (per the session context), so the tree
  could move under this report.

---

## Recommended fix (described, NOT applied)

**Primary — teach the merge about permission precedence.**

In `internal/hook/registry.go`'s `mergeHandlerOutput`, add explicit handling for
`PermissionDecision` / `PermissionDecisionReason` with a precedence ladder rather than a
last-writer-wins copy:

- `deny` beats everything (already short-circuited upstream, so this is belt-and-braces);
- `ask` beats `allow` and beats the pre-seeded default — when `src.PermissionDecision == DecisionAsk`
  and `dst` is not already `deny`, copy both the decision and its reason;
- `allow` never overwrites an existing `ask` or `deny`.

The reason string must travel with the decision — an `ask` with an empty
`permissionDecisionReason` gives the user a dialog with no explanation of why.

**Alternative (simpler, narrower) — short-circuit `ask` like `deny`.**

Extend the `Dispatch` short-circuit so a non-empty, non-`allow` `PermissionDecision` returns the
handler's output directly, the way `isBlockDecision` already does for `deny`. This is a smaller
diff and matches the existing structure, but it changes multi-handler semantics: an `ask` from
handler 0 would suppress handler 1's `additionalContext`. Given exactly one PreToolUse handler is
registered today (E4), the practical difference is nil — but the primary fix is the more honest one
if a second handler is ever registered.

**Mandatory accompanying test (this is the real fix).**

The defect shipped because the producer was tested and the consumer was not (E6). Any patch should
add:

1. a registry-level test — `registry.Dispatch(EventPreToolUse, …)` with an `askPatterns`-matching
   `file_path`, asserting the dispatched output carries `permissionDecision == "ask"` **and** a
   non-empty reason;
2. the same assertion for a `AskBashPatterns` command (`git reset --hard`), since that path is
   equally broken and equally untested;
3. a CLI-level assertion in the style of `internal/cli/hook_protocol_fix_test.go`, so the wire
   format is pinned end-to-end the way `deny` already is.

**Re-verification after the fix.** Re-run the exact E1 and E2 probe commands against a freshly
installed binary and expect `permissionDecision: "ask"` with a populated reason. Note the local
reinstall discipline: `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai` (or `make install`) —
a plain `cp` over the existing binary is known to produce a SIGKILL-on-run failure in this project.

**Scope note.** Gap 4 (whether `PermissionRequest`'s `decision.behavior` suffers the same loss)
should be checked as part of the fix, since it lives in the same function and the same omission
pattern.
