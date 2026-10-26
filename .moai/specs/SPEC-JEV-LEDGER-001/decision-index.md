---
id: SPEC-JEV-LEDGER-001
title: "Decision index — SPEC-JEV-LEDGER-001"
created: 2026-10-26
updated: 2026-10-26
---

# Decision index — SPEC-JEV-LEDGER-001

`interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). One row per decision the plan surfaced that the card did not settle. At 0.2.0 (split "a" ruling) every row is kept; rows whose subject migrated say where it went.

### Q1: Does the ledger write rely on the SPEC-JEV-CORE-001 REQ-JEVC-011 inert-sibling-record carve-out by name, or does it require a CORE amendment?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The carve-out text (`.moai/specs/SPEC-JEV-CORE-001/spec.md`, REQ-JEVC-011, "The carve-out: an inert sibling record") delegates authorisation "by name in a downstream SPEC". No prior HISTORY or Amendments row decides this identical question for a ledger file.
Why unresolved: the carve-out was written about a record appended alongside backlog cards. Applying it to a hook-event ledger is a reading, not a decision already on record.
Default: rely on the carve-out by name, no CORE amendment (rule: the option that preserves current behavior, since CORE text stays unchanged)
Alternate: amend SPEC-JEV-CORE-001 REQ-JEVC-011 to name the ledger explicitly
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)
0.2.0 note: still binding. The reduction strengthens condition (i): the recorder never opens the ledger for reading (REQ-JEVL-011, AC-JEVL-014).

### Q2: How is the hook adapter's Jev deadline derived inside the 5 s (Claude Code) / 10 s (Codex) budget, and what is the safety margin?
Label: EVIDENCE-NEEDED
Authority anchor: none. The card states the 5 s / 10 s budgets. The measured host timeouts are PreToolUse 10 s / Stop 5 s / SubagentStop 5 s (Claude Code template) and 10 s (Codex, `internal/codexwiring/codexwiring.go:60`). No measurement exists of the wrapper-start-to-exit wall clock.
Why unresolved: the margin that keeps the whole process under the host kill depends on time outside the `moai` process clock, which has not been measured.
Operator verdict: LANE-DECIDED 2026-10-26T10:41:41Z lane t1762 (decision D1, recorded by manager-spec). Budget = 5 s, margin = 1 s, offset = 4 s, one value for every event and harness. The margin is a decision, not a measurement.
Measurement rule (iter-3 repair, superseded the verdict line's blocking formula): worst-case wrapper-start-to-exit p99 > 5 s − S, S = 0.5 s, S a decision.
0.2.0 split: the two constants and their derivation stay here (REQ-JEVL-013, AC-JEVL-013, plan.md §B). The deadline anchor, the measurement, its sample-validity precondition (audit R3), and S migrated to SPEC-JEV-LEDGER-002 (card t1970), which carries this row's open evidence need forward. The constants remain decisions until -002 measures.

### Q3: What unit does "cost per correct decision" use?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. `internal/jev` records only `Usage.InputTokens`, a byte-length estimate (`internal/jev/jev.go:595`, `:366-368`).
Why unresolved: the card names the metric but not its unit.
Default: estimated input tokens (`Usage.InputTokens`) per correct decision, labelled "estimated" in the report (rule: the option with the smaller user-visible surface)
Alternate: currency, via a compiled price table
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)
0.2.0 split: migrated with `moai jev cost` to SPEC-JEV-LEDGER-003 (card t1971). This SPEC stores `input_tokens` as supplied and prints no figure.

### Q4: Do `advise` / `steer` / `confirm` execute as `shadow` in this SPEC?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The card lists the four-value enum but no behaviour for the non-shadow values. Executing them would reach REQ-JEVC-012 surfaces.
Why unresolved: the card does not say whether this SPEC activates them.
Default: accept and record all four, execute as shadow (rule: the option that preserves current behavior)
Alternate: implement advise output in this SPEC, with a CORE amendment
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)
0.2.0 note: still binding at the resolver (REQ-JEVL-006, AC-JEVL-006). The call-site half migrated to -002.

### Q5: Does the new `moai jev cost` subcommand count as a "user-visible surface change" under card acceptance item 5?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. Card t1762 acceptance item 5 is card text, not a committed authority artifact.
Why unresolved: item 5 asks to confirm that no user-visible surface changes, while the card scope itself added a command.
Default: no (rule: the option that preserves current behavior, since no existing surface changes)
Alternate: treat it as a surface change and amend SPEC-JEV-CORE-001 §D
Operator verdict: DEFAULT-APPLIED 2026-10-26T10:21:34Z manager-spec (plan-phase owner)
0.2.0 split: migrated with the command to SPEC-JEV-LEDGER-003 (card t1971). This SPEC adds no user-visible surface.

### Q6: Land as one card, or split?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The ~10 files / ~400 lines ceiling is stated in the dispatch prompt, not in a committed config.
Why unresolved (at 0.1.x): the estimate exceeded the stated ceiling, and splitting changes the card set.
Default: split into three slices (rule: the option whose undo is a single revert of this SPEC's own commits)
Alternate: land as one card
Operator verdict: LANE-DECIDED 2026-10-26T10:41:41Z lane t1762 (decision D3): one card, serial M1 → M4. Withdrawn by the lane's hold record `.moai/reports/t1762/hold-and-split.md` after the Tier M audit ceiling ended in FAIL 0.88 at `0c2d2aa4c2`.
Operator verdict (supersedes the line above): LEADER-RULED 2026-10-26 factory leader, operator-delegated — split. t1762 is split "a" (M1 + M2, this SPEC at 0.2.0); hook adapter, M3 measurement, and credential wiring → SPEC-JEV-LEDGER-002 (card t1970); `moai jev cost` → SPEC-JEV-LEDGER-003 (card t1971); both land after t1762. Estimate after the split: 11 files, ~770 lines (plan.md §E).

### Q7: Does the ledger recorder live in `internal/jev` or in a package of its own?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. No committed SPEC or config names the recorder's package. `internal/jev/display_only_test.go:15-19` states the package "cannot mutate ... a file", which bears on the question but does not decide it.
Why unresolved: the 0.1.x plan placed the recorder in `internal/jev`; the split ruling asks that `internal/jev/jev.go` be left to card t1968 and prefers no `internal/jev` edit.
Default: a new package `internal/jevledger` (rule: the option that preserves current behavior, since `internal/jev` and its "cannot mutate a file" property stay exactly as they are)
Alternate: new files inside `internal/jev` (`ledger.go` and its platform pair), leaving `jev.go` untouched
Operator verdict: DEFAULT-APPLIED 2026-10-26 manager-spec (plan-phase owner, at plan close of 0.2.0)
