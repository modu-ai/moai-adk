---
id: SPEC-JEV-LEDGER-001
title: "Plan — Jev decision ledger, shadow-mode hook adapter, cost report"
version: "0.1.0"
created: 2026-10-26
updated: 2026-10-26
---

# Plan — SPEC-JEV-LEDGER-001

Tree measured: `cb30e05b3e5f64774b8dabd87e80c24332f4551d` (card worktree `t1762`, branch `WT-jev-internal-jev`). Every claim about existing code below comes from a command run against that tree. The commands are listed in §I.

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
| `input_tokens` | int | `Usage.InputTokens` | charged tokens for the request. Repeated on every row of one call, so the cost report de-duplicates by `call_id`. |
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

**Adapter call site.** `runHookEvent` in `internal/cli/hook.go` is the single dispatcher for every hook subcommand (`hookSubcommands` rows `pre-tool`, `stop`, `subagent-stop` at `internal/cli/hook.go:47,50,59`). Both harnesses pass through it: `harnessModeIsCodex(cmd)` selects the Codex output adapter, and `internal/codexadapter/events.go:68,72,75` maps the same three dispatcher args for Codex. The adapter is called **after** `writeHookOutput` / `writeHookOutputCodex` has returned and **after** the exit-code decision exists (REQ-JEVL-004). It takes the event, the harness, the parsed input, and the process start time. It returns nothing. It uses no shared writer: no `cmd.OutOrStdout()`, no `os.Stderr`.

**Budget.** The hook process has a wall-clock budget: 5 s for Claude Code and 10 s for Codex, the card's figures. Measured host timeouts are recorded for honesty. In `internal/template/templates/.claude/settings.json.tmpl`, the `handle-pre-tool.sh` entries carry `"timeout": 10`, `handle-stop.sh` carries 5, and `handle-subagent-stop.sh` carries 5. For Codex, `internal/codexwiring/codexwiring.go:60` declares `defaultHandlerTimeout = 10`. The Claude Code budget of 5 s is therefore the smallest host timeout across the three events, so one constant covers all three. The Jev deadline is `budget − elapsed − margin`. When that is ≤ 0 the call is skipped and no row is written (nothing was attempted). Both the budgets and the margin live in `internal/config/defaults.go` as named constants, per the repository rule that thresholds live there. [NEEDS CLARIFICATION: safety margin value — plan default 1 s; decision-index Q2]

**Fail-open.** `jev.Client.Ask` already returns a typed `Result` with no error (`internal/jev/jev.go`, `func (c *Client) Ask`). A deadline surfaces as `Unreachable` from `send` ("transport failed: ..."). The adapter adds `deadline_exceeded` from `ctx.Err()`. A panic inside a consumer's request builder is recovered and dropped.

**Gate placement.** The adapter checks `workflow.jev.enabled` first and returns before touching the registry, the config, or the filesystem (REQ-JEVL-001). The client is constructed with `jev.New(enabled)`, so the existing gate inside `Ask` still applies as a second layer.

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

Go side: `WorkflowJevConfig` gains `Consumers map[string]JevConsumerConfig` with `Mode string` and `Threshold *float64`. The pointer distinguishes absent from 0, following `AutonomyKickoffConfig.JevMinConfidence *float64` at `internal/config/types.go:1085`. Resolution (absent → shadow, invalid → shadow + warning text; out-of-range threshold → absent + warning text) is a pure function that returns the warning instead of printing it (REQ-JEVL-008).

**Template-first.** The distributed change is `internal/template/templates/.moai/config/sections/workflow.yaml`, edited in the template first, followed by `make build` to regenerate the embedded template. The local `.moai/config/sections/workflow.yaml` is this repository's dogfood config (it carries `enabled: true`). It may gain the same commented `consumers: {}` line, but nothing requires it.

## §E. Milestones (priority-ordered)

Expected sizes are estimates for the card-size ceiling, not measurements.

| M | Priority | Content | Files (prod / test) | Expected changed lines |
|---|---|---|---|---|
| M1 | High | Config: `Consumers` map, resolution function, defaults constants (budgets + margin), template block, `make build` | `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/jev_consumers.go` (new), `internal/config/jev_consumers_test.go` (new), `internal/template/templates/.moai/config/sections/workflow.yaml` | ~60 prod + ~90 test + ~8 yaml |
| M2 | High | Ledger recorder: row type, append (0700/0600, single write), fail-open drop; extend `TestDisabled_ConstructsNoRequest` so gate off also leaves no ledger file | `internal/jev/ledger.go` (new), `internal/jev/ledger_test.go` (new), `internal/jev/jev_test.go` | ~90 prod + ~140 test |
| M3 | High | Hook adapter: empty registry, gate-first, budget/deadline, post-output call site; byte-identity and timeout tests on both harnesses; allow-list update | `internal/cli/hook_jev.go` (new), `internal/cli/hook_jev_test.go` (new), `internal/cli/hook.go` (call site, ~6 lines), `internal/cli/doctor_jev_test.go` (allow-list, 2 lines) | ~130 prod + ~220 test |
| M4 | Medium | `moai jev cost`: reader, de-dup by `call_id`, undefined-when-zero, malformed-line skip | `internal/cli/jev_cost.go` (new), `internal/cli/jev_cost_test.go` (new), `internal/cli/doctor_jev_test.go` (allow-list, 1 line) | ~80 prod + ~100 test |

**Total estimate: 13 distinct files, ~370 production lines + ~550 test lines + ~8 yaml ≈ 930 changed lines.**

### Card-size ceiling: exceeded. Proposed split

The scope **exceeds** the ~10 files / ~400 changed lines guidance (13 files, ~930 lines including tests). I am not widening it silently. The proposed split is three landing slices under this one SPEC. Each slice is a separate card with its own commit set, and the SPEC's ACs are partitioned among them:

| Slice | Milestones | ACs it closes | Files | Est. lines |
|---|---|---|---|---|
| t1762-a — ledger + config | M1, M2 | AC-JEVL-006, -007, -009, -010, -011, -012, -016 | 8 | ~390 |
| t1762-b — hook adapter | M3 | AC-JEVL-001, -002, -003, -004, -005, -008, -013, -014 | 4 | ~360 |
| t1762-c — cost report | M4 | AC-JEVL-015 | 3 | ~180 |

Order: a → b → c. Slice b needs the row type and the resolution function from a. Slice c needs only the row type. The alternative is to land as one card and accept the overrun. That is the lane's or the leader's call, and it is recorded here so it is made consciously.

## §F. Technical approach notes

- **Concurrency.** Each row is marshalled to one byte slice ending in `\n` and written with one `Write` on an `O_APPEND|O_CREATE|O_WRONLY` file. Rows of a few hundred bytes stay well below the practical atomic-append size on local filesystems. A test with parallel goroutines asserts every line parses. Multi-process atomicity on network filesystems is not claimed (§H).
- **Project root.** Resolve through the existing `resolveHookProjectRoot()` (env-first: `CLAUDE_PROJECT_DIR`, then `Getwd`), already used in `runHookEvent` for the HOI gate. This avoids the `internal/hook/.moai/` leak class (B7). No new resolver.
- **Windows.** Modes 0700/0600 are POSIX-only. The append path is the same. Cross-build with `GOOS=windows`.
- **Cost report de-duplication.** `input_tokens` repeats on each question row of one request, so the report sums once per `call_id`. A call refused before the send (oversize, secret-detected) is not charged by `Ask` (`Usage` is recorded only after `send`), so the row carries `input_tokens: 0` for those.
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
