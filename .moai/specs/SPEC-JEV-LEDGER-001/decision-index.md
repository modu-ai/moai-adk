---
id: SPEC-JEV-LEDGER-001
title: "Decision index — SPEC-JEV-LEDGER-001"
created: 2026-10-26
---

# Decision index — SPEC-JEV-LEDGER-001

`interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). One row per decision the plan surfaced that the card did not settle.

### Q1: Does the ledger write rely on the SPEC-JEV-CORE-001 REQ-JEVC-011 inert-sibling-record carve-out by name, or does it require a CORE amendment?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The carve-out text (`.moai/specs/SPEC-JEV-CORE-001/spec.md`, REQ-JEVC-011, "The carve-out: an inert sibling record") delegates authorisation "by name in a downstream SPEC". No prior HISTORY or Amendments row decides this identical question for a ledger file.
Why unresolved: the carve-out was written about a record appended alongside backlog cards. Applying it to a hook-event ledger is a reading, not a decision already on record.
Default: rely on the carve-out by name, no CORE amendment (rule: the option that preserves current behavior, since CORE text stays unchanged)
Alternate: amend SPEC-JEV-CORE-001 REQ-JEVC-011 to name the ledger explicitly
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)

### Q2: How is the hook adapter's Jev deadline derived inside the 5 s (Claude Code) / 10 s (Codex) budget, and what is the safety margin?
Label: EVIDENCE-NEEDED
Authority anchor: none. The card states the 5 s / 10 s budgets. The measured host timeouts are PreToolUse 10 s / Stop 5 s / SubagentStop 5 s (Claude Code template) and 10 s (Codex, `internal/codexwiring/codexwiring.go:60`). No measurement exists of the hook's own pre-adapter elapsed time.
Why unresolved: the margin that keeps the whole process under the host kill depends on the elapsed time of the existing handlers, which has not been measured.
Operator verdict: LANE-DECIDED 2026-10-26T10:41:41Z lane t1762 (decision D1, recorded by manager-spec). Budget = 5 s, the tightest host timeout (Claude Code Stop/SubagentStop, and the single shared constant). Margin = 1 s. Effective Jev deadline = 4 s from hook process start, the same for pre-tool, stop, and subagent-stop under both harnesses. Gap: handler elapsed time before the adapter runs was NOT measured. The margin is a decision, not a measurement. Implementation M3 measures it per event and harness, records it in progress.md §E.2, and returns a blocker if measured elapsed time plus 4 s would exceed 5 s.

### Q3: What unit does "cost per correct decision" use?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. `internal/jev` records only `Usage.InputTokens`, and no price constant exists (`grep -rn -i 'price\|usd\|cost' internal/jev/jev.go` shows comments only).
Why unresolved: the card names the metric but not its unit.
Default: charged input tokens per correct decision (rule: the option with the smaller user-visible surface, since it adds no price data)
Alternate: currency, via a compiled price table
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)

### Q4: Do `advise` / `steer` / `confirm` execute as `shadow` in this SPEC?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The card lists the four-value enum but no behaviour for the non-shadow values. Executing them would reach REQ-JEVC-012 surfaces.
Why unresolved: the card does not say whether this SPEC activates them.
Default: accept and record all four, execute as shadow (rule: the option that preserves current behavior)
Alternate: implement advise output in this SPEC, with a CORE amendment
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)

### Q5: Does the new `moai jev cost` subcommand count as a "user-visible surface change" under card acceptance item 5?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. Card t1762 acceptance item 5 is card text, not a committed authority artifact.
Why unresolved: item 5 asks to confirm that no user-visible surface changes, while the card scope itself adds a command.
Default: no. It is a new, explicitly invoked, read-only report that consults no Jev, and existing surfaces are byte-identical (rule: the option that preserves current behavior, since no existing surface changes)
Alternate: treat it as a surface change and amend SPEC-JEV-CORE-001 §D "Out of Scope — user-facing surfaces"
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)

### Q6: Land as one card, or split into t1762-a / -b / -c per plan.md §E?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The ~10 files / ~400 lines ceiling is stated in the dispatch prompt, not in a committed config.
Why unresolved: the estimate (13 files, ~930 lines including tests) exceeds the stated ceiling, and splitting changes the card set.
Default: split into three slices (rule: the option whose undo is a single revert of this SPEC's own commits, since each slice reverts alone)
Alternate: land as one card
Operator verdict: LANE-DECIDED 2026-10-26T10:41:41Z lane t1762 (decision D3, recorded by manager-spec). Alternate taken: one card, carried serially M1 -> M2 -> M3 -> M4 with one commit per milestone. The ~13 files / ~980 lines estimate exceeds the illustrative ceiling, and that is recorded in plan.md §E. The split was rejected because its AC partition closed ACs in slices that could not satisfy their green paths (AC-012/016 in slice a, AC-014 in slice b). Per-milestone commits keep per-slice revertibility.
