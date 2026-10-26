---
id: SPEC-JEV-LEDGER-001
title: "Plan — Jev decision ledger, shadow-mode hook adapter, cost report"
version: "0.1.2"
created: 2026-10-26
updated: 2026-10-26
---

# Plan — SPEC-JEV-LEDGER-001

Tree measured: `cb30e05b3e5f64774b8dabd87e80c24332f4551d` (card worktree `t1762`, branch `WT-jev-internal-jev`). Every claim about existing code below comes from a command run against that tree. The commands are listed in §I. The iter-2 repair additions (§B budget, §D decode path, §E closure map) were read against `d1aa4ed656af948c95c2567718c4f9a7be54e191`. The files they cite are unchanged between the two commits: `d1aa4ed656` only adds this SPEC's plan artifacts. The iter-3 repair additions (§A unit, §B budget measurement and Ask seam, §E file count, §F write-side open) were read against `82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0` (§I).

The sections run from the decisions most likely to change (data model, interfaces, the authority boundary) to the mechanical steps.

---

## §A. Data model — the ledger row (highest change likelihood)

One JSON object per line in `.moai/state/jev/decisions.jsonl`. Field names are snake_case and stable. Later fields are additive only.

| Field | Type | Source | Notes |
|---|---|---|---|
| `ts` | RFC 3339 UTC string | adapter clock | |
| `consumer` | string | registration name | |
| `event` | `pre-tool` \| `stop` \| `subagent-stop` | dispatcher subcommand | |
| `harness` | `claude` \| `codex` | `--harness` flag | |
| `question_id` | string | `jev.Question.ID` | one row per question |
| `rubric_version` | string | compiled with the consumer | a question-set version, not a model version |
| `model_id` | string | `jev.Answer.Model` / `Usage.Model` | the interpreted model id (`jev-1.13.0` shape). Empty when unanswered. Never `jev-latest`. |
| `request_model` | string | `Usage.RequestModel` | always `jev.ModelID` |
| `availability` | string | `jev.Result.Availability` | `available`, `unreachable`, ... |
| `deadline_exceeded` | bool | adapter (`ctx.Err() == context.DeadlineExceeded`) | |
| `probability` | number \| null | `jev.Answer.Probability` | raw, unthresholded |
| `choice` / `score` | optional | `jev.Answer` | kind-dependent |
| `input_tokens` | int | `Usage.InputTokens` | estimated input tokens for the request: a byte-length estimate, not a billed count (`internal/jev/jev.go:595` sets it from `requestTokens(req)`; `:366-368` documents `EstimateTokens` as "an ESTIMATE, not a tokenizer"). The response's own `usage.input_tokens` (`:405`) is not read (spec §D). Repeated on every row of one call, so the cost report de-duplicates by `call_id`. |
| `call_id` | string | adapter | groups the rows of one request |
| `mode` | string | effective mode | always `shadow` in this SPEC |
| `configured_mode` | string | config | the raw configured value |
| `threshold` | number \| null | config | recorded only |
| `action` | string | adapter | `none` in this SPEC |
| `outcome_label` | string | recorder | `unlabeled` at write time |
| `config_warning` | optional string | adapter | REQ-JEVL-008 |

The card requires seven fields: `model_id`, `rubric_version`, `question_id`, raw probability, `mode`, `action`, and `outcome_label`. All seven appear above. The rest exist because REQ-JEVL-015 needs `input_tokens` and `call_id`, REQ-JEVL-013 needs `deadline_exceeded`, and REQ-JEVL-006/007/008 need the config fields.

`model_id` follows `SPEC-JEV-RESOLVED-MODEL-001` REQ-JRM-005: it never carries the request alias. Card t1768 **has landed**. `internal/jev/jev.go:72` declares `const ModelID = "jev-latest"`, `Usage` separates `Model` from `RequestModel`, and the SPEC reads `status: completed`. Most recent commit on the file: `09140cc18e feat(SPEC-JEV-RESOLVED-MODEL-001): M3 change notice (card t1768)`. The ledger therefore records the response `model` field and never assumes it.

## §B. Interfaces — the consumer seam and the adapter

**Consumer registration (compiled, empty in this SPEC).** A consumer is a compiled value with: a name, a `rubric_version`, the set of events it accepts, and a pure function that builds a `jev.Request` from the parsed hook input. The function returns "skip" when it has nothing to ask. The registry is a package-level slice that ships empty (REQ-JEVL-002). Tests register a fake through a test-only seam.

**Adapter call site.** `runHookEvent` in `internal/cli/hook.go` is the single dispatcher for every hook subcommand (`hookEventSubcommands` rows (declared at `internal/cli/hook.go:41`) `pre-tool`, `stop`, `subagent-stop` at `internal/cli/hook.go:47,50,59`). Both harnesses pass through it: `harnessModeIsCodex(cmd)` selects the Codex output adapter, and `internal/codexadapter/events.go:68,72,75` maps the same three dispatcher args for Codex. The adapter is called **after** `writeHookOutput` / `writeHookOutputCodex` has returned and **after** the exit-code decision exists (REQ-JEVL-004). It takes the event, the harness, the parsed input, and the process start time. It returns nothing. It uses no shared writer: no `cmd.OutOrStdout()`, no `os.Stderr`.

**Budget.** The hook process has a wall-clock budget: 5 s for Claude Code and 10 s for Codex, the card's figures. Measured host timeouts are recorded for honesty. In `internal/template/templates/.claude/settings.json.tmpl`, the `handle-pre-tool.sh` entries carry `"timeout": 10`, `handle-stop.sh` carries 5, and `handle-subagent-stop.sh` carries 5. For Codex, `internal/codexwiring/codexwiring.go:60` declares `defaultHandlerTimeout = 10`. The tightest host timeout is therefore 5 s (Claude Code `Stop` and `SubagentStop`), so one budget covers all three events and both harnesses. The Codex 10 s is looser and is not used as a separate value.

**Decided (lane decision D1, decision-index Q2):** budget 5 s, safety margin 1 s. The adapter's Jev deadline is `process start + 5 s − 1 s`, which is at most 4 s of process time, identical on `pre-tool`, `stop`, and `subagent-stop` under both harnesses. The `context.WithDeadline` passed to `Ask` is built from that instant, so time already spent by the hook handler shrinks the Jev window. When the deadline has already passed at adapter entry, the call is skipped and no row is written, because nothing was attempted. Two named constants live in `internal/config/defaults.go` per the repository rule that thresholds live there: the 5 s budget and the 1 s margin. The 4 s figure is derived from them, never stored separately.

**Gap, carried to M3:** the 1 s margin is a decision, not a measurement, and it stays a decision until M3 measures it. The deadline is an absolute instant from process start, so handler time is already inside the 4 s and must not be added to it again. What the margin has to absorb lies outside the `moai` process clock or after the deadline:

- **Before process start.** The host runs a wrapper script, not `moai` directly. `internal/template/templates/.claude/hooks/moai/handle-stop.sh.tmpl` starts bash (`:1`), runs `stat` on the hook stderr log (`:18`), checks `command -v moai` (`:25`), and only then runs `exec moai hook stop` (`:26`). `handle-subagent-stop.sh.tmpl:25-26` has the same shape, and `handle-pre-tool.sh.tmpl:70` has the same `command -v moai` check.
- **After the deadline.** The ledger write, the deferred `rs.Shutdown()` trace flush (`internal/cli/hook.go:409`), and process exit.

**M3 measurement (redefined at iter-3).** M3 measures the worst-case wall-clock time from wrapper-script start to process exit. The measurement runs each event's wrapper script with a fake transport that blocks until its context deadline, so Jev holds the process to the 4 s deadline every time. It takes at least 20 runs per event under the Claude Code wrapper and records each sample and the p99 in `progress.md` §E.2. Blocking condition: **when the measured p99 exceeds 5 s − S, with S = 0.5 s, M3 returns a blocker report instead of shipping.** S is a decision recorded in decision-index Q2, not a measurement. The Codex harness command line was not read in this plan (it is registered by `internal/codexwiring`); M3 measures it the same way or records it as a gap.

**Ask call seam.** The adapter calls `Ask` through a one-method interface that `*jev.Client` satisfies, obtained from a package-level factory in `hook_jev.go`. Production returns `jev.New(enabled)`. Tests substitute a wrapper that counts `Ask` calls and delegates to a real client. AC-JEVL-001 counts both the consumer request-builder calls and these `Ask` calls.

**Fail-open.** `jev.Client.Ask` already returns a typed `Result` with no error (`internal/jev/jev.go`, `func (c *Client) Ask`). A deadline surfaces as `Unreachable` from `send` ("transport failed: ..."). The adapter adds `deadline_exceeded` from `ctx.Err()`. A panic inside a consumer's request builder is recovered and dropped.

**Gate placement.** The adapter checks `workflow.jev.enabled` first and returns before touching the registry, the config, or the filesystem (REQ-JEVL-001). The client is constructed with `jev.New(enabled)`, so the existing gate inside `Ask` still applies as a second layer. Because that second layer returns `Disabled` before building a request (`internal/jev/jev.go:538-544`), a transport count alone cannot detect an adapter that skips the first check. AC-JEVL-001 therefore asserts on the request-builder count and the `Ask` count.

## §C. The authority boundary, and why SPEC-JEV-CORE-001 needs no amendment (card item 5)

Evidence read for this conclusion: `.moai/specs/SPEC-JEV-CORE-001/spec.md` lines 94-133 (REQ-JEVC-011 with its two exceptions and its carve-out, REQ-JEVC-012, REQ-JEVC-013, REQ-JEVC-014) and lines 160-178 (the Out of Scope sections), plus `.claude/skills/moai-jev-skill-suggestion/SKILL.md` and `.claude/skills/moai-ref-jev-question-design/SKILL.md`.

| CORE clause | What it requires | How this SPEC meets it without amending |
|---|---|---|
| REQ-JEVC-011, "a Jev answer shall not mutate ... a file" | display-only, with two named exceptions and the inert-sibling-record carve-out | The ledger row sits **alongside** the hook event it names and changes no object. Carve-out condition (i): no code path reads a row back to write a field of anything it names. The only reader is `moai jev cost`, which prints aggregates. Condition (ii): the row carries its own `request_model` / `model_id` fields and renders only through the labelled cost report, never as a mechanical measurement. Condition (iii): no decision, gate, verdict, mark, or disposition selects on it (REQ-JEVL-014). The carve-out says a consumer relying on it is "authorised by name in a downstream SPEC". This SPEC is that downstream SPEC and names itself. The carve-out delegates authorisation downstream, so CORE's text needs no change. Its parenthetical "(today: ...)" goes stale as an illustration, and that is recorded in §H as a residual risk, not fixed. |
| REQ-JEVC-012, no consultation *for* a listed act | the answer must not reach a verdict, gate, queue mutation, or user-surface behaviour change | REQ-JEVL-004 runs the adapter after the hook's output and exit decision are fixed. The "determined first / unread after" discriminator CORE uses holds structurally. Non-shadow modes, the only paths that could change a user surface, are executed as shadow (REQ-JEVL-006) and left to the card that activates them. |
| REQ-JEVC-013, labelling | any surfaced answer is labelled a model signal | No individual answer is surfaced. The cost report labels its aggregates (REQ-JEVL-015). |
| REQ-JEVC-014, identical output when absent | pre-Jev output when disabled or unreachable | REQ-JEVL-001 (gate off) and REQ-JEVL-005 (shadow) go further: byte-identical even when enabled and answered. |
| REQ-JEVC-015/016/017 | gate inside `workflow.yaml`, default false, no request while off | The `consumers` block lives inside the existing `workflow.jev` key. No new section file. `enabled: false` stays the template value (`internal/template/templates/.moai/config/sections/workflow.yaml:350-351`) and the compiled default (`internal/config/defaults.go:1659-1661`). |
| REQ-JEVC-018..021 | credential, secret screening | Untouched. All requests go through `Ask`, which runs `ScreenPayload` before sending. |

**User-visible surfaces.** No existing user-visible surface changes. The hook output is byte-identical (AC-JEVL-004). `moai doctor`, `moai init`, `moai web`, `jev_ask`, and `moai jev --key` are untouched. The one new surface is `moai jev cost`: explicitly invoked, read-only, and consulting no Jev. It is new, not changed, and REQ-JEVC-012 does not cover it because no Jev answer drives its behaviour. Whether card item 5 counts a new subcommand as a "change" is decision-index Q5.

**A code comment that becomes stale.** The `WorkflowJevConfig` doc comment says the struct "carries a bare Enabled flag and nothing else" (`internal/config/types.go:1052-1058`). Adding `Consumers` falsifies that sentence. M1 rewrites the comment. Its point (no knob can change the alias or the endpoint) still holds and is kept.

**The guard that must move with the code.** `TestJevCallPath_HasExactlyTheDeclaredConsumers` (`internal/cli/doctor_jev_test.go`) states the exact set of `internal/cli` files allowed to import `internal/jev`. The new `hook_jev.go` and `jev_cost.go` must be added to that map with this SPEC's citation. They must not be routed through an intermediate package so that the scan stays green. That would evade the guard rather than satisfy it, so the adapter file imports `internal/jev` directly. `TestJevCallPath_UnreachableFromDecisionSurfaces` names `gtd.go`, `todo_analysis.go`, `todo_autodone.go`, `integration.go`, and `integration_settings_drift.go`. None of them is touched, and that test stays unchanged. `grep -rn 'internal/jev"' internal/hook` returned no match, so the `internal/hook` package gains no import either.

## §D. Config surface

```yaml
workflow:
    jev:
        enabled: false
        # consumers: per-consumer mode and threshold. No consumer is compiled
        # in this release, so an entry here has no effect until one ships.
        # mode: shadow | advise | steer | confirm  (all run as shadow today)
        consumers: {}
```

**Why a typed field cannot carry this (observed at `d1aa4ed656`).** `internal/config/loader.go` imports `gopkg.in/yaml.v3` (`:11`). `loadYAMLFile` runs `yaml.Unmarshal(data, target)` and turns any decode error into `parse workflow.yaml: ErrInvalidYAML` (`:494-495`). `loadWorkflowSection` (`:236-251`) seeds a wrapper with the defaults, calls `loadYAMLFile`, and on any error logs `slog.Warn("failed to load workflow config, using defaults", ...)` and returns nil (`:239-241`). In that case `cfg.Workflow = wrapper.Workflow` at `:247` never runs, so the **whole** `workflow` section stays at its defaults. A plain `Threshold *float64` field therefore fails twice on `threshold: "high"`. The decode error throws away every other workflow setting, not just that consumer's. It also emits a `slog.Warn`, which is printed output and breaks REQ-JEVL-008. The value never reaches the resolver at all.

**Decode path (D6).** Two custom `UnmarshalYAML(*yaml.Node) error` methods make the decode tolerant at consumer granularity, so neither can fail the section:

- `JevConsumers` (the map type behind `WorkflowJevConfig.Consumers`). When the node is a mapping, each value node is decoded into a `JevConsumerConfig` on its own. When the node is anything else (a scalar or a sequence), the result is an empty map plus one recorded section-level warning. It returns nil in both cases.
- `JevConsumerConfig`. It keeps the raw `mode` node and the raw `threshold` node as `yaml.Node` values instead of typed fields, and returns nil for any node shape.

The resolver, a pure function in `internal/config/jev_consumers.go`, then interprets the raw nodes. `mode` becomes the scalar's value when the node is a string scalar and an unrecognized value otherwise. `threshold` becomes a `*float64` when the node is an `!!int` or `!!float` scalar inside [0, 1], and absent plus a warning for any other tag, a non-scalar, or an out-of-range value. Absent stays absent with no warning, and `0` stays distinguishable from absent. Warnings are returned as data and never logged (REQ-JEVL-008). One bad consumer is thereby quarantined, and the rest of the `workflow` section, other consumers included, loads unchanged. `AutonomyKickoffConfig.JevMinConfidence *float64` (`internal/config/types.go:1085`) stays as it is. It is outside this SPEC and is not the pattern to copy here, because its typed decode carries exactly the failure described above.

**Template-first.** The distributed change is `internal/template/templates/.moai/config/sections/workflow.yaml`, edited in the template first, followed by `make build` to regenerate the embedded template. The local `.moai/config/sections/workflow.yaml` is this repository's dogfood config (it carries `enabled: true`). It may gain the same commented `consumers: {}` line, but nothing requires it.

## §E. Milestones (priority-ordered)

Expected sizes are estimates for the card-size ceiling, not measurements.

| M | Priority | Content | Files (prod / test) | Expected changed lines |
|---|---|---|---|---|
| M1 | High | Config: `Consumers` map with the per-consumer tolerant decode (§D), resolution function, defaults constants (5 s budget + 1 s margin), template block, `make build`, plus a loader-level test that loads a real `workflow.yaml` through `internal/config/loader.go` | `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/jev_consumers.go` (new), `internal/config/jev_consumers_test.go` (new), `internal/template/templates/.moai/config/sections/workflow.yaml` | ~80 prod + ~120 test + ~8 yaml |
| M2 | High | Ledger recorder: row type, append (0700/0600, single write), the write-side regular-file open (§F), fail-open drop | `internal/jev/ledger.go` (new), `internal/jev/ledger_test.go` (new), `internal/jev/ledger_open_unix.go` (new), `internal/jev/ledger_open_windows.go` (new) | ~130 prod + ~180 test |
| M3 | High | Hook adapter: empty registry, gate-first, `Ask` seam, budget/deadline, post-output call site; byte-identity and timeout tests on both harnesses; allow-list update; the M3 wall-clock measurement (§B) | `internal/cli/hook_jev.go` (new), `internal/cli/hook_jev_test.go` (new), `internal/cli/hook.go` (call site, ~6 lines), `internal/cli/doctor_jev_test.go` (allow-list, 2 lines) | ~130 prod + ~250 test |
| M4 | Medium | `moai jev cost`: reader, de-dup by `call_id`, undefined-when-zero, malformed-line skip, "estimated" labels | `internal/cli/jev_cost.go` (new), `internal/cli/jev_cost_test.go` (new), `internal/cli/doctor_jev_test.go` (allow-list, 1 line; already counted in M3) | ~80 prod + ~100 test |

**Total estimate: 15 distinct files (M1 5, M2 4, M3 4, M4 2 new), ~420 production lines + ~650 test lines + ~8 yaml ≈ 1080 changed lines.** Iter-3 changes: the iter-2 table listed 14 distinct paths (the iter-2 text said 13, a miscount). `internal/jev/jev_test.go` leaves the set, because the AC-JEVL-001 M2 partial was removed (acceptance.md), and the two platform open files join it (§F). The line estimate rises by about 100 for the write-side open and the added AC-012/AC-013 cases. That is above the Tier M ~1000-line guidance by under 10%. The file count is at the Tier M 15-file upper bound.

### Card-size ceiling: exceeded, carried as serial milestones (lane decision D3)

The scope **exceeds** the illustrative ~10 files / ~400 changed lines ceiling (15 files, ~1080 lines including tests at iter-3; 14 files, ~980 lines when the lane decided). That is recorded here, not hidden. The lane decided to keep **one card**, carried serially M1 → M2 → M3 → M4 inside the lane, with **one commit per milestone**. Each commit names the card id and the milestone, so each milestone reverts on its own. Decision-index Q6 records the decision.

### AC closure map (each AC closes only where its green path is satisfiable)

An AC is **final** at the earliest milestone after which every test in its green path can pass. Where part of its green path is satisfiable earlier, that earlier milestone records a **partial** result: those tests' `--- PASS` lines go into `progress.md` §E.2, and the AC is not marked PASS. An AC is reported PASS only at its final milestone, and the release-blocking set is judged only after M4.

| AC | Partial at | Final at | Why |
|---|---|---|---|
| AC-JEVL-006 | M1 (resolver subtests) | M3 | The call-site mutant probe needs the adapter. |
| AC-JEVL-007 | — | M1 | The resolver and the loader-level test both live in `internal/config`. |
| AC-JEVL-013 | M1 (constants test) | M3 | The deadline behaviour needs the adapter. |
| AC-JEVL-016 | M1 (template test) | M4 | The malformed-ledger behaviour needs the reader. |
| AC-JEVL-009, -010 | — | M2 | Recorder only. |
| AC-JEVL-011 | — | M2 self-verification | Regression guard, not counted as a pass. |
| AC-JEVL-001 | — | M3 | Every assertion (builder calls, `Ask` calls, no ledger directory) needs the adapter. The iter-2 M2 partial was removed: it extended `TestDisabled_ConstructsNoRequest`, which calls only `Ask`, and `Ask` never writes a ledger, so the no-ledger assertion could not fail. |
| AC-JEVL-012 | M2 (`TestLedger_WriteFailureDropped`, `TestLedger_NonRegularPathDropped`, both calling the recorder directly) | M3 | The hook-silent half needs the adapter. |
| AC-JEVL-002, -003, -004, -005, -008 | — | M3 | Adapter only. |
| AC-JEVL-014 | M3 (`hook_jev.go` in the allow-list) | M4 | `jev_cost.go` exists only after M4. |
| AC-JEVL-015 | — | M4 | Cost report only. |

Order: M1 → M2 → M3 → M4. M2 needs nothing from M1. M3 needs the resolver and constants from M1 and the row type from M2. M4 needs the row type from M2 and the template block from M1.

**Rejected alternative: split into t1762-a (M1+M2) / t1762-b (M3) / t1762-c (M4).** It was rejected for two reasons. First, the slicing assigned ACs to slices whose milestones could not satisfy their green paths: AC-012 and AC-016 in slice a needed M3 and M4, and AC-014 in slice b needed M4. Three cards would therefore carry partial ACs across card boundaries, where only a single SPEC record can follow them. Second, the lane decided to keep the card whole. One commit per milestone gives the same per-slice revertibility the split was meant to buy.

## §F. Technical approach notes

- **Write-side regular-file open (REQ-JEVL-011/012).** A plain `O_APPEND|O_CREATE|O_WRONLY` open blocks forever on a FIFO with no reader and appends to the target of a symlink. The repository already documents both hazards for the read side: `internal/runtime/regular_file.go:4-8` says that a path-based check followed by an open by name "makes its target the evidence and a FIFO blocks the open". `internal/runtime/regular_open_nofollow.go:12-19` opens with `O_NOFOLLOW` and `O_NONBLOCK`, and `:29-37` then requires the **opened** descriptor to be a regular file and closes it otherwise. The recorder uses the write-side counterpart of that pattern in its own helper, because `openRegularFile` is unexported, read-only, and in another package. On POSIX (`ledger_open_unix.go`, the same build tags as `regular_open_nofollow.go`), the helper opens with `O_WRONLY|O_APPEND|O_CREATE|O_NOFOLLOW|O_NONBLOCK` at mode 0600, calls `Stat` on the opened handle, and closes the handle and drops the row unless the mode is regular. With no reader, a FIFO fails the non-blocking write open at once (`ENXIO`). With a reader, it opens and is refused by the handle check. A symlink fails the open (`ELOOP`) and its target is never touched. On Windows (`ledger_open_windows.go`, mirroring `regular_open_stat.go`), the helper refuses a path that `Lstat` reports as a link, then applies the same handle check. No later step resolves the path again. Every refusal drops the row silently (REQ-JEVL-012).
- **Concurrency.** Each row is marshalled to one byte slice ending in `\n` and written with one `Write` on the handle the write-side open returned (`O_APPEND`). Rows of a few hundred bytes stay well below the practical atomic-append size on local filesystems. A test with parallel goroutines asserts every line parses. Multi-process atomicity on network filesystems is not claimed (§H).
- **Project root.** Resolve through the existing `resolveHookProjectRoot()` (env-first: `CLAUDE_PROJECT_DIR`, then `Getwd`), already used in `runHookEvent` for the HOI gate. This avoids the `internal/hook/.moai/` leak class (B7). No new resolver.
- **Windows.** Modes 0700/0600 are POSIX-only. The append path is the same. Cross-build with `GOOS=windows`.
- **Cost report de-duplication.** `input_tokens` repeats on each question row of one request, so the report sums once per `call_id`. For a call refused before the send (oversize, secret-detected), `Ask` records no `Usage` (it is recorded only after `send`), so the row carries `input_tokens: 0`. The report prints every token figure under the literal label `estimated input tokens` and heads the aggregate block with the literal `model signal aggregate` (AC-JEVL-015).
- **No new MX ANCHOR burden.** `runHookEvent` already carries `@MX:ANCHOR`. The adapter is a single new callee, and the anchor's fan-in is unchanged.

## §G. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| A synchronous Jev call adds latency to every tool call once a consumer exists and the gate is on | High (later cards) | Bounded by the deadline. This SPEC ships no consumer, so no latency today. Later cards own their latency budgets. |
| The adapter writes to stdout/stderr by accident (a log line, a warning) and breaks shadow byte identity | Medium | AC-JEVL-004/005 compare full captured stdout and stderr against the gate-off baseline on both harnesses. Warnings are returned as data (REQ-JEVL-008). |
| Routing the import through a helper package to keep the allow-list test green | Low | §C forbids it explicitly. AC-JEVL-014 checks the allow-list contains the new files by name. |
| Concurrent SPEC on `internal/cli/jev.go` (`SPEC-GLM-JEV-KEY-001`, status `in-progress`) | Medium | `jev cost` lives in its own file and attaches to `jevCmd` from its own `init()`. `jev.go` is not edited. |
| Ledger growth without rotation | Low | Out of scope. Rows are small, and with no consumer nothing is written. |

## §H. Residual risks recorded, not fixed

- CORE REQ-JEVC-011's "(today: `SPEC-JEV-CONSUMERS-001` REQ-JEVN-001 ...)" parenthetical no longer lists every carve-out user once this SPEC lands. It is illustrative, so it is not amended here (decision-index Q1).
- Single-write append atomicity across processes is measured only on the developer's local filesystem.

## §I. Commands that grounded this plan (this run, tree `cb30e05b3e`)

- `grep -n '^func \|^type \|^const\|^var' internal/jev/jev.go`: the `ModelID`, `Usage`, `Result`, `Ask`, and `send` declarations.
- `sed -n 538,610p internal/jev/jev.go`: `Ask` checks the gate first, records `Usage` only after `send`, and returns no error.
- `sed -n 625,668p internal/jev/jev.go`: a transport error, deadline included, returns `Unreachable`.
- `sed -n 320,460p internal/cli/hook.go`: `runHookEvent` dispatch order, the Codex branch, and the output write before the exit-2 branch.
- `grep -n '"pre-tool"\|"stop"\|"subagent-stop"' internal/codexadapter/events.go`: lines 68, 72, 75.
- `grep -rn -i 'timeout' internal/codexwiring/*.go`: `defaultHandlerTimeout = 10`.
- The `settings.json.tmpl` timeout extraction: PreToolUse 10, Stop `handle-stop.sh` 5, SubagentStop 5.
- `sed -n 1048,1075p internal/config/types.go` and `sed -n 1650,1665p internal/config/defaults.go`: `WorkflowJevConfig{Enabled bool}` and default false.
- `grep -n -B2 -A3 '^    jev:' internal/template/templates/.moai/config/sections/workflow.yaml`: `enabled: false` at line 351.
- `git check-ignore -v .moai/state/jev/decisions.jsonl`: `.gitignore:401:.moai/state/`.
- `sed -n 215,340p internal/cli/doctor_jev_test.go`: both guard tests and their allow-lists.
- `git log --oneline -1 -- internal/jev/jev.go`: `09140cc18e`, card t1768 landed.
- `go test ./internal/jev/ -run '^TestDisabled_ConstructsNoRequest$' -count=1 -v`: `--- PASS: TestDisabled_ConstructsNoRequest (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/jev 0.143s`.

Iter-2 additions (tree `d1aa4ed656af948c95c2567718c4f9a7be54e191`):

- Read `internal/config/loader.go` lines 220-259: `loadWorkflowSection` at `:236-251`, the warn-and-default branch at `:239-241`, and the commit of the decoded section at `:247`.
- `grep -n "func loadYAMLFile" -A 25 internal/config/loader.go`: `yaml.Unmarshal(data, target)` at `:494`, error wrapped as `ErrInvalidYAML` at `:495`.
- `grep -n 'yaml' internal/config/loader.go`: `"gopkg.in/yaml.v3"` at `:11`.
- `grep -n "JevMinConfidence\|type WorkflowJevConfig" internal/config/types.go`: `:1059` and `:1085`.

Iter-3 additions (tree `82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0`):

- `grep -n "InputTokens\|func requestTokens\|ESTIMATE\|bytesPerTokenEstimate =" internal/jev/jev.go`: `:116` `bytesPerTokenEstimate = 4`, `:263` `Usage.InputTokens`, `:367` "ESTIMATE, not a tokenizer", `:405` the response `InputTokens` field, `:595` `usage := Usage{InputTokens: requestTokens(req), ...}`, `:750` `func requestTokens`.
- `sed -n 536,546p internal/jev/jev.go`: `Ask` returns `Disabled` before any request when `!c.Enabled`.
- `grep -n "" internal/runtime/regular_file.go` and `grep -n "" internal/runtime/regular_open_nofollow.go`: the read-side pattern cited in §F.
- `grep -n "rs.Shutdown\|exitCodeError\|hookEventSubcommands = \|writeHookOutput" internal/cli/hook.go`: `:41` `hookEventSubcommands`, `:409` `defer rs.Shutdown()`, `:439`/`:442` output writes, `:450` the exit-2 return.
- `grep -n "bash\|stat\|command -v moai\|exec moai" internal/template/templates/.claude/hooks/moai/handle-stop.sh.tmpl`: `:1`, `:18`, `:25`, `:26`. `grep -n "command -v moai\|exec moai"` on `handle-subagent-stop.sh.tmpl` and `handle-pre-tool.sh.tmpl`: `:25-26` and `:70`.
- `sed -n 170,200p internal/cli/hook_fault_injection_test.go`: the exit-2 fault fixture used by AC-JEVL-013 (`:181-189`).
