# SPEC-HARNESS-LEARNING-EVO-001 — Implementation Plan (L1)

> Derived from `spec.md`. Ordered by decision-reversibility: the two contested design decisions come first (§E), then the seam semantics and store API, then the mechanical wiring.

## §A. Context and measured baseline

### A.1 Baseline commands (re-runnable)

All figures were measured against the **primary checkout** `/Users/goos/MoAI/moai-adk-go` on 2026-08-09. They are not reproducible from this worktree: `.moai/state/`, `.moai/harness/`, and `.moai/evolution/` are gitignored runtime state and are absent here. Re-run from the primary checkout; the figures drift upward as the observer keeps writing, and small drift is expected rather than a discrepancy.

| Claim | Command | Observed |
|---|---|---|
| Ledger effectively empty | `wc -l < .moai/state/routing-ledger.jsonl` | 4 rows |
| No delegation ever recorded | `jq -c '.delegations \| length' .moai/state/routing-ledger.jsonl \| sort \| uniq -c` | `4  0` |
| Outcomes never terminal | `jq -r .outcome .moai/state/routing-ledger.jsonl \| sort \| uniq -c` | `3 abort`, `1 reroute` |
| Sibling observer is not starved | `wc -l < .moai/harness/usage-log.jsonl` | 109,236 rows |
| Nothing writes the map | `grep -rn 'delegation.yaml' internal/ --include='*.go' \| grep -v _test` | 0 matches |
| Proposal ladder works | `ls .moai/harness/proposals \| wc -l` ; `wc -l < .moai/harness/learning-history/tier-promotions.jsonl` | 73 dirs ; 175 rows |
| Agent identity is observable | `grep '"subagent_stop"' .moai/harness/usage-log.jsonl \| jq -r '.agent_type // "(none)"' \| sort \| uniq -c \| sort -rn` | 2,783 rows; 1,941 (69.7%) attributed; 842 `(none)`; retained-catalog names lead the distribution |
| Terminal signal population is empty | `cat .moai/evolution/telemetry/usage-*.jsonl \| jq -s '[.[] \| select(.is_test_pass == true or .is_test_fail == true)] \| length'` | `0` |
| Root cause is hook registration | `jq -r '.hooks.PostToolUse[] \| {matcher, cmds:[.hooks[].command]}' .claude/settings.json` | `handle-post-tool.sh` registered for `Write\|Edit\|MultiEdit` only; `handle-harness-observe.sh` for `matcher: null` |

### A.2 Existing plumbing this SPEC builds on

| Surface | Path | Role here |
|---|---|---|
| Ledger schema + finalize | `internal/harness/routing/types.go` | `Row`, `PendingRow`, `Delegation`, `EvidenceRef`, `PendingRow.Finalize` — consumed unchanged |
| Pending store | `internal/harness/routing/pending.go` | `Record` (sweeps + reroutes self), `sweepStale` (age + liveness guards), `AppendEvidence`, `AppendDelegation`, `FinalizeOnStop` |
| Reader | `internal/harness/routing/reader.go` | `NewReader(path).Read(filter)` — not used by L1; the L2 consumer's entry point |
| CLI ledger verbs | `internal/cli/harness_ledger.go` | `record` / `evidence` / `list`; registered by `newHarnessLedgerCmd()` at `internal/cli/harness_route.go:144` |
| Stop finalizer | `internal/cli/hook.go` `finalizeRoutingLedgerOnStop` | gate 0 (`isHookOptInEnabled`, fail-closed) → gate 1 (`isHarnessLearningEnabled`, fail-open) → finalize |
| Evidence writer | `internal/hook/evidence_writer.go` | `buildEvidenceRecord` / `buildBashRecord` / `classifyTestCommand`; reached only from `logEvidence` (`internal/hook/post_tool.go:224`) |
| Session evidence load | `internal/telemetry/recorder.go` `LoadBySession` | reads exactly two whole day-files (today + yesterday) |
| Observe channel | `internal/cli/hook.go` `runHarnessObserve` | PostToolUse handler behind `handle-harness-observe.sh` (`matcher: null`); today records only `Subject = ToolName` |
| Hook input | `internal/hook/types.go` | `Prompt`, `SessionID`, `AgentType`, `AgentName`, `AgentID`, `ToolInput`, `ToolResponse` |

## §B. Known issues the plan must not reintroduce

- **B1 — reroute flood.** `Store.Record` finalizes the session's own prior pending row as `reroute` (`rerouteSelf`). A hook calling `Record` on every user prompt would close a pending row every turn, reproducing exactly the `reroute`-only ledger observed today. This is why REQ-HLE-003/004 exist.
- **B2 — sweep on the hot path.** `sweepStale` does an `os.ReadDir` plus one `loadPending` read per foreign pending file plus a liveness lookup. Under `Record` it runs once per dispatch; wiring it into a per-prompt `RecordIfAbsent` would convert it into a per-prompt cost. REQ-HLE-015 forbids that.
- **B3 — frozen/template surface.** `.claude/settings.json` is rendered from `internal/template/templates/.claude/settings.json.tmpl`, and `.claude/skills/moai/SKILL.md` is both a frozen pattern and template-managed. Editing either drags in the Template-First cycle (`make build` + mirror) and the §25 internal-content isolation doctrine. §E D2 chooses a path that avoids both.
- **B4 — inert by default.** Gate 0 is fail-closed and default OFF. Every L1 path ships inert for distributed users; the repository must opt in to gather data. This is correct behavior, not a defect — but it means "no rows appeared" is ambiguous unless the gate state is checked first.
- **B5 — reading the wrong identity field.** The derived `subject` field is `unknown` on all 2,783 observed `subagent_stop` rows. `agent_type` is the populated field. A seam reading `subject` would produce a ledger full of `unknown` delegations that looks structurally healthy.

## §C. Pre-flight

1. Confirm the worktree is `feat/harness-learning-evo` and `git status` is clean of foreign edits.
2. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` green before the first edit (baseline attribution).
3. `golangci-lint run --timeout=2m 2>&1 | tail -5` — record the pre-existing baseline so NEW findings are separable.
4. Record the primary checkout's current ledger row count and telemetry test-signal count so a post-change delta is attributable.

## §D. Constraints

- Additive only within ledger schema v1 (REQ-HLE-014).
- Fail-open on every hook seam; gates re-checked at each seam, matching `finalizeRoutingLedgerOnStop`'s existing pattern.
- No `AskUserQuestion` in any new file (subagent boundary; `internal/harness/routing/subagent_boundary_test.go` already guards the routing package).
- No template edits, no `.claude/settings.json` edit, no `delegation.yaml` read or write, no frozen-allowlist edit.
- Do not touch runtime-managed files (`.moai/harness/*`, `.moai/state/*`, `.moai/evolution/*`).

## §E. Decisions carried into audit

Two decisions are contested and are stated here rather than buried in a milestone.

### D1 — `matched_subcommand` write policy: **first-writer-wins**

A session running `/moai plan` and then `/moai run` produces both literals against one pending row, because the row only finalizes on a terminal signal. `Store.Annotate`'s stated semantics (REQ-HLE-005) is that a non-empty value overwrites, so the seam must state its own policy explicitly.

**Chosen: first-writer-wins.** Delegations accumulated before the second literal arrived were spawned under the first subcommand; overwriting the label retroactively re-attributes observations that were never made under the new label. It also keeps the seam idempotent, which is the same invariant REQ-HLE-003 protects. The cost is that a plan-then-run session attributes its run-phase delegations to `plan` — recorded as residual risk R5 and as the driver for the deferred per-subcommand row split (`spec.md` §G). The downstream consumer must therefore treat one row as one subcommand observation, which `SPEC-HARNESS-LEARNING-EVO-002` states explicitly.

### D2 — terminal-signal source: extend the observe handler, do not edit settings

`spec.md` §A.6 establishes that `buildBashRecord` is unreachable because `handle-post-tool.sh` is registered for `Write|Edit|MultiEdit` only. Three paths were considered.

| Option | Verdict |
|---|---|
| **(a)** Add `Bash` to the `handle-post-tool.sh` matcher | Rejected as first choice. `.claude/settings.json` is rendered from `internal/template/templates/.claude/settings.json.tmpl`, so this triggers the Template-First cycle and the §25 neutrality doctrine — exactly what REQ-HLE-011's boundary exists to avoid. It also switches the whole Stop evidence gate on for every distributed user, a blast radius far beyond this SPEC. |
| **(b)** Derive the signal from the harness-observe channel as it stands | **Refuted by inspection.** `runHarnessObserve` (`internal/cli/hook.go`) builds `harness.Event{EventType: agent_invocation, Subject: hookInput.ToolName}` and discards `ToolInput` and `ToolResponse` entirely. The channel *is* registered for Bash (`matcher: null`), but the handler carries neither the command text nor the outcome. Option (b) as originally framed does not work. |
| **(c)** Extend `runHarnessObserve` to also assemble the evidence record, scoped to Bash | **Chosen.** The wrapper already receives the full Bash PostToolUse payload; only the handler discards it. Calling the existing `buildEvidenceRecord` path for `ToolName == "Bash"` is a pure Go change — no `settings.json` edit, no template edit, so REQ-HLE-011's boundary holds. Scoping to Bash also avoids a double write, since `Write`/`Edit` are already covered by `handle-post-tool.sh`. Requires exporting a small entry point from package `hook`, because `buildEvidenceRecord` and `logEvidence` are unexported. |

**Unverified premise, gated by M0.** Option (c) assumes the Bash PostToolUse payload delivered on the matcher-null wrapper carries `tool_input.command` and `tool_response`. Registration for Bash is verified; the payload content on that channel is a documented Claude Code contract but was not runtime-observed here. M0 settles it before M4 is written. **If M0 falsifies the premise, fall back to option (a)** and pay the Template-First cost explicitly: edit `internal/template/templates/.claude/settings.json.tmpl`, run `make build`, mirror to `.claude/settings.json`, and re-check §25 neutrality — and state in the run-phase report that REQ-HLE-011's no-template boundary was traded away deliberately, not overlooked.

## §F. Milestones

Ordered by decision-reversibility. M0 and M1 carry the decisions most likely to change; M4 is mechanical.

### M0 — Probe the terminal-signal premise (gate on §E D2)

- Enable the local hook opt-in gate, run a single Bash test command in a live session, and capture what the harness-observe wrapper actually receives (the wrapper's stdin, or an equivalent instrumented capture in the handler).
- **Decision gate**: does the captured payload carry a non-empty `tool_input.command` and a `tool_response`?
  - **Yes** → proceed with option (c); M4 seam C is a Go-only change.
  - **No** → proceed with option (a); add the Template-First steps to M4 and record the boundary trade in `progress.md`.
- Do not begin M4 before this gate resolves. M1-M3 are independent of it and may proceed in parallel.

### M1 — Seam semantics and store API (highest change-likelihood)

The genuinely contested design decision: what a pending row's lifecycle is under mechanical emission.

- Add `Store.RecordIfAbsent(PendingRow) error` — creates a pending row only when the session has none; a no-op otherwise. **Does not run `sweepStale`** (REQ-HLE-015, B2) and never calls `rerouteSelf` (REQ-HLE-004).
- Add `Store.Annotate(sessionID string, patch RoutingPatch) error` — patches routing metadata onto an existing pending row; no create, no finalize; empty fields leave existing values untouched (REQ-HLE-005).
- `RoutingPatch` is a new value type in `routing/types.go` carrying the six annotatable fields as optional values.
- `Store.Record` is left untouched — sweep, reroute-on-redispatch, and both sweep guards keep their current behavior, which AC-HLE-001 and AC-HLE-003 pin.
- Add `moai harness ledger annotate` to `internal/cli/harness_ledger.go`, flag set mirroring `record` minus stdin.

**Why first:** every later milestone binds to this API, and a wrong lifecycle choice here reproduces B1 invisibly.

### M2 — Delegation identity and outcome contract

- Define the absent-identity marker as an exported constant in `routing/types.go` — a non-empty sentinel distinct from every retained-catalog agent name (REQ-HLE-008). It must be recognizable downstream, so it is a declared constant, not an ad-hoc string at the call site.
- Define the unknown-outcome marker likewise (REQ-HLE-009), reusing the existing outcome vocabulary where one already fits.
- Record in the package doc comment that `agent` holds `agent_type` verbatim and may be a spawn name rather than an agent type (`spec.md` §A.5 caveat 2), so a future reader does not mistake the field for a validated enum.

### M3 — Bash evidence reachability

Chosen path per §E D2 option (c), subject to M0.

- Export a minimal entry point from package `hook` (e.g. `hook.LogEvidenceFor(input)`) wrapping the existing `logEvidence`, so `internal/cli` can call the assembled-record path without duplicating classification logic.
- In `runHarnessObserve`, after the existing gate checks, call it **only when `hookInput.ToolName == "Bash"`** — the Write/Edit/MultiEdit cases stay owned by `handle-post-tool.sh`, so no record is written twice (REQ-HLE-011, AC-HLE-013).
- Leave the existing `harness.Event` recording untouched; this is additive.

### M4 — L1 wiring: the three mechanical seams

New file `internal/hook/routing_ledger.go` holding the seam helpers, so the three handler files take minimal diffs.

- **Seam A — UserPromptSubmit.** Gate 0 → gate 1 → `RecordIfAbsent` with digest/class derived from `input.Prompt`; `matched_subcommand` set from a literal `/moai <sub>` prefix only when currently empty (REQ-HLE-002/003/006, §E D1).
- **Seam B — SubagentStop.** Gate 0 → gate 1 → `AppendDelegation{Agent: agentTypeOrMarker(input), Outcome: observedOutcomeOrUnknown(input), Blocker: nil}` (REQ-HLE-007/008/009).
- **Seam C — Stop.** In `internal/cli/hook.go`, immediately before `finalizeRoutingLedgerOnStop`, append a terminal `gate_exit` evidence ref when `telemetry.LoadBySession` reports an observed test pass or fail for the session (REQ-HLE-010). Absence of a signal appends nothing.
- All three swallow errors into the diagnostic sink (REQ-HLE-012) and no-op when either gate is closed (REQ-HLE-013).

### M5 — Verification and dogfood

- Full-suite run plus the §D quality gate of `acceptance.md`.
- Manual dogfood check for R1: enable the opt-in, run one real `/moai` dispatch with at least one subagent, read the resulting ledger row, and record the observed row in `progress.md` §E.2. This is the only evidence that the live hook dispatch reaches the seams, and it is manual by construction.

## §G. Anti-patterns

- **AP-1 — reroute-per-prompt.** Calling `Store.Record` from the UserPromptSubmit seam. Reproduces the observed failure exactly and looks like success (rows appear).
- **AP-2 — sweep-per-prompt.** Calling `sweepStale` from `RecordIfAbsent` "for symmetry with `Record`". Converts a once-per-dispatch directory scan into a once-per-prompt one (B2).
- **AP-3 — inferred success.** Defaulting a stopping subagent's delegation outcome to `success` because no failure was seen. Absence of a failure signal is not evidence of success.
- **AP-4 — instruction patch.** "Fixing" emission by strengthening the wording in `.claude/skills/moai/SKILL.md`. That is the mechanism that already failed (`spec.md` §A.2), and it drags in the frozen/template surface.
- **AP-5 — live-data AC.** Writing an acceptance criterion that reads `.moai/state/routing-ledger.jsonl` or `.moai/evolution/telemetry/*.jsonl` and asserts content. Both are gitignored runtime state, gate-dependent, and empty in CI.
- **AP-6 — reading `subject`.** Taking the delegation's agent from the derived `subject` field, which is `unknown` on 100% of observed rows (B5).
- **AP-7 — silent identity normalization.** Mapping an unrecognized `agent_type` onto a catalog name, or dropping the entry, at the seam. The seam records verbatim; discrimination is the consumer's job.
- **AP-8 — double evidence write.** Calling the evidence path from the observe handler for `Write`/`Edit` as well as `Bash`, producing two telemetry records for one tool call.

## §H. Deferred / follow-up (not in this SPEC)

- The L2 analyzer — `SPEC-HARNESS-LEARNING-EVO-002`.
- Ledger schema v2 with an injected-skills field, which would unlock skill-level proposals.
- Per-subcommand row splitting, which would retire the first-writer-wins approximation (§E D1).
- Widening the Bash evidence path beyond the observe-channel carve-out (i.e. option (a) on its own merits), which would switch the Stop evidence gate on for distributed users.

## §I. File inventory (run phase)

| Path | Change |
|---|---|
| `internal/harness/routing/types.go` | edit — `RoutingPatch`, absent-identity + unknown-outcome constants |
| `internal/harness/routing/pending.go` | edit — `RecordIfAbsent`, `Annotate` |
| `internal/harness/routing/pending_test.go` | edit — lifecycle, sweep-guard, and no-sweep-on-create tests |
| `internal/cli/harness_ledger.go` | edit — `annotate` verb |
| `internal/cli/harness_route.go` | (no change — `newHarnessLedgerCmd()` is already registered at line 144; the new verb attaches inside the ledger command) |
| `internal/hook/evidence_writer.go` | edit — exported entry point wrapping `logEvidence` |
| `internal/hook/routing_ledger.go` | new — three seam helpers + gates |
| `internal/hook/routing_ledger_test.go` | new |
| `internal/hook/user_prompt_submit.go` | edit — seam A call |
| `internal/hook/subagent_stop.go` | edit — seam B call |
| `internal/cli/hook.go` | edit — Bash carve-out in `runHarnessObserve`; seam C before the finalizer |
| `internal/cli/hook_test.go` | edit — observe-channel Bash evidence + end-to-end terminal close |
| `internal/telemetry/recorder_test.go` | edit — two-day window assertion |

Roughly 13 files, of which 9 are edits. No template files, no `.claude/` files (unless M0 forces option (a), which adds `internal/template/templates/.claude/settings.json.tmpl` and its mirror).

## §J. Tier classification

**Tier M.** Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, the scope-based table places Tier M at 300-1000 LOC and 5-15 files affected; this SPEC's inventory is ~13 files and the change is well under 1000 LOC of production code. The requirement and acceptance-criterion counts (16 and 16) sit exactly at the Tier M ceiling, which is the binding constraint that motivated the split from the original 33/36 SPEC. Tier M carries the 3-artifact set (spec.md + plan.md + acceptance.md) plus `progress.md`, and a plan-auditor PASS threshold of 0.80.

## §K. Cross-references

- `SPEC-HARNESS-LEARNING-EVO-002` — the L2 consumer of the rows this SPEC produces
- `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §2.5, §5 row P3
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the REQ/AC budget that forced the split
- `.claude/rules/moai/core/verification-claim-integrity.md` — the §E falsification discipline and the M0 probe gate
- `CLAUDE.local.md` §2 (Template-First), §25 (template isolation) — the cost §E D2 option (c) avoids
