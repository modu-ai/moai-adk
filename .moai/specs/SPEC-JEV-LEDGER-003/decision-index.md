---
id: SPEC-JEV-LEDGER-003
title: "Decision index — SPEC-JEV-LEDGER-003"
created: 2026-11-06
updated: 2026-11-06
---

# Decision index — SPEC-JEV-LEDGER-003

`interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). One row per decision the plan took while grounding (mirrors the SPEC-JEV-LEDGER-001 index format). Each was DEFAULT-APPLIED by manager-spec (plan-phase owner) at authoring; none required an operator round — every default preserves current behavior or takes the smallest surface, and the card dispatch named the four headline questions (allow-list necessity, subcommand placement, `--json` shape, duplicate-row resolution) for plan-phase resolution.

### D1: Does `jev_cost.go` need an entry in the `TestJevCallPath_HasExactlyTheDeclaredConsumers` allow-list (the AC-JEVL-014 part migrated as "by-name allow-list entry")?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The migrated text (SPEC-JEV-LEDGER-001 §D.1, AC-JEVL-014 row) described the entry at split time; the live allow-list (`internal/cli/doctor_jev_test.go:240-249`) names files permitted to import `internal/jev`, and no committed SPEC decides the question for a non-importing reader.
Why unresolved at authoring: the premise under the migrated wording — that the report, like the adapter, would import `internal/jev` — is false on the landed code. The reader needs `internal/jevledger` and the standard library only: it constructs no client, sends no request, and reads no credential (REQ-JEVL-014), so it has no reason to import `internal/jev` at all.
Default: no allow-list entry; the boundary is asserted as the non-importer proof — the package guard passing with `jev_cost.go` present and unlisted, plus a source scan of the file (rule: the option that preserves current behavior — the guard and its allow-list are unchanged — with the smallest surface)
Alternate: add a dead allow-list row for byte-parity with the migrated §D.1 wording
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)
Consequence: `internal/cli/doctor_jev_test.go` is not edited by this card; the deviation from the migrated wording is recorded in acceptance.md AC-JEVL-014 (Carried-text deviation).

### D2: Where does the command attach?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The carried REQ text fixes the invocation spelling `moai jev cost`; no committed text names the registration site.
Default: a `cost` child of the package-level `jevCmd`, registered from `jev_cost.go`'s own `init()`; `internal/cli/jev.go` is not edited (rule: the option that preserves current behavior — the landed credential surface, including its flag-error handling, stays byte-identical)
Alternate: a top-level `moai jev-cost` (breaks the carried spelling), or editing `jev.go` (wider blast radius on a landed credential command)
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)

### D3: What shape does `--json` take?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The card names the flag without a shape; the repository's other `--json` flags are per-command shapes with no shared schema to inherit.
Default: one flat object with `calls`, `estimated_input_tokens`, `correct`, `estimated_input_tokens_per_correct_decision` (carrying `null` at zero correct, mirroring text `undefined`), and `skipped` (always present, `0` included); the token keys carry `estimated` so no figure loses its estimate label in the machine form; plain text stays the default, with `skipped: N` printed only when N ≥ 1, per the carried REQ-JEVL-016 conditioning (rule: the smallest machine-readable surface that carries the same aggregates and keeps every label)
Alternate: mirror the text exactly (always print a `skipped: 0` line), or nest the figures under a heading key
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)

### D4: Which row wins on duplicate `call_id`?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The carried AC fixture makes the rows of one call identical by construction (LEDGER-001 plan §F: `input_tokens` repeats on every row of one request), and no committed text names a winner.
Default: the first row encountered in file order is kept per `call_id` (rule: deterministic scan order — the rule a reader can reason about without re-deriving row equality)
Alternate: last-wins, or refusing duplicate `call_id` groups
Note: the choice cannot change output today — the recorder repeats one call's `input_tokens` across its rows, so any winner yields the same aggregates; the AC's per-row-sum mutant (1200 vs 600) catches the real hazard, de-dup direction aside.
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)

### D5: Decode each ledger line into `jevledger.Row` directly, or into a local parallel struct?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The card binds the row shape as consumed-as-is; no text names the decode target.
Default: decode into `jevledger.Row` (rule: the option with fewer abstractions — a parallel struct is a second copy of the shape that can drift)
Alternate: a local struct carrying only `call_id`, `input_tokens`, `outcome_label`
Note: the report writes nothing, so decoding into the recorder's type cannot re-enter the recorder.
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)

### D6: What does the report do on a read failure other than "absent" (a permission denial, a directory at the ledger path)?
Label: FOUNDER
Class: implementation-level
Authority anchor: none verified. The carried REQ-JEVL-016 names absent, empty, and malformed only.
Default: the same fail-open path — zero calls, exit 0 (rule: the option that preserves current behavior; the ledger is advisory local state and the report is display-only, so a report command must not fail on it)
Alternate: exit non-zero on an unreadable ledger
Residual recorded: an unreadable ledger is indistinguishable from an empty one in the output (plan.md §H).
Operator verdict: DEFAULT-APPLIED 2026-11-06 manager-spec (plan-phase owner)
